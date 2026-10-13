package adapters

import (
	"fmt"
	"io"
	"io/fs"
	"iter"
	"os"
	"path/filepath"
	"strings"

	k8syaml "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
)

// ManifestAdapter reads Kubernetes-style YAML/JSON manifests from the filesystem
// (a single file, or every *.yaml/*.yml/*.json under a directory tree) and yields
// each document as the SAME nested-HAMT payload the live
// NativeKubernetesResourceAdapter produces. That shared payload shape is the whole
// point: the exact same graph-building DSL — relation rules, walks, folds — runs
// unchanged over offline manifests as over a live cluster, giving true live/offline
// parity (design doc §4.1) and enabling the live→manifest/kustomize transform
// (§6 #10), since both sides speak nested HAMT.
//
// Why manifests rather than an in-process `kustomize build`? `kustomize build`
// *output* is itself a stream of YAML documents, so `kustomize build overlay/ |
// … > out.yaml` followed by this adapter already yields offline parity with zero
// extra dependency. A native in-process Kustomize wrapper (pulling the
// sigs.k8s.io/kustomize API) is a thin follow-up that an edition can add when it
// needs build-on-read; it would reuse unstructuredToValue identically.
//
// The input node's value is the path to read (a file or directory). Multi-document
// files (`---`-separated) are fully supported; empty documents are skipped. A read
// or parse failure is yielded as an error-valued node so evaluator.EitherT lifts it
// to a Left — the branch fails loudly instead of silently vanishing. Documents are
// visited in the lexical order filepath.WalkDir guarantees, so a crawl is
// deterministic.
func ManifestAdapter() evaluator.Adapter[any] {
	return func(node dsltypes.ASTNode[any]) iter.Seq[dsltypes.ASTNode[any]] {
		return func(yield func(dsltypes.ASTNode[any]) bool) {
			path, ok := node.Value.(string)
			if !ok || path == "" {
				yield(dsltypes.NewASTNode[any](
					fmt.Errorf("ManifestAdapter: expected a file or directory path, got %T", node.Value),
					node.DSLType))
				return
			}

			files, err := manifestFiles(path)
			if err != nil {
				yield(dsltypes.NewASTNode[any](err, node.DSLType))
				return
			}

			for _, file := range files {
				docs, err := decodeManifestFile(file)
				if err != nil {
					if !yield(dsltypes.NewASTNode[any](fmt.Errorf("%s: %w", file, err), node.DSLType)) {
						return
					}
					continue
				}
				for _, doc := range docs {
					if !yield(dsltypes.NewASTNode[any](unstructuredToValue(doc), node.DSLType)) {
						return
					}
				}
			}
		}
	}
}

// manifestFiles resolves a path to the ordered set of manifest files it denotes:
// the file itself if path is a file, or every *.yaml/*.yml/*.json beneath it (in
// filepath.WalkDir's lexical order) if path is a directory.
func manifestFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}

	var files []string
	walkErr := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(p)) {
		case ".yaml", ".yml", ".json":
			files = append(files, p)
		}
		return nil
	})
	return files, walkErr
}

// decodeManifestFile decodes every YAML/JSON document in a manifest file into the
// generic map[string]any shape (matching the live adapter's unstructured content),
// skipping empty documents. It streams documents via the apimachinery YAML-or-JSON
// decoder so multi-document files and JSON manifests are handled uniformly.
func decodeManifestFile(path string) ([]map[string]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := k8syaml.NewYAMLOrJSONDecoder(f, 4096)
	var docs []map[string]any
	for {
		var obj map[string]any
		decErr := decoder.Decode(&obj)
		if decErr == io.EOF {
			break
		}
		if decErr != nil {
			return docs, decErr
		}
		if len(obj) == 0 {
			continue // empty document (e.g. a trailing `---`)
		}
		docs = append(docs, obj)
	}
	return docs, nil
}
