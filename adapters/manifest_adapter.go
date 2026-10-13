package adapters

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	k8syaml "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
)

// ManifestOption configures ManifestAdapter. Options are the parameterization seam
// through which Kustomize support is enabled *without* the adapter ever importing a
// kustomize library: the adapter only knows "render a directory to a YAML byte
// stream via a pluggable function". The renderer is either the external
// `kustomize build` (WithKustomize, WithKustomizeCommand) or any function the caller
// injects (WithRenderer) — the Kubernetes shim injects an in-process one
// (pkg/k8sgraph.InProcessRenderer). All kustomize-specific behavior lives in the
// renderer, never in the adapter.
type ManifestOption func(*manifestConfig)

type manifestConfig struct {
	kustomize bool
	render    func(dir string) ([]byte, error)
	// enabled is consulted on every read: false reads a kustomization root as the
	// plain files it contains. nil means always enabled.
	enabled func() bool
}

// WithKustomize enables Kustomize rendering. When the adapter is pointed at a
// *kustomization root* (a directory containing kustomization.yaml/.yml or
// Kustomization), it first runs `kustomize build` to render the library into a
// plain multi-document YAML stream, then decodes that stream exactly like any other
// manifest. Plain directories and files are read directly, unchanged — so a mixed
// tree of raw manifests and kustomize overlays "just works". Since Kustomize
// libraries are a daily occurrence, this is the option most editions will enable.
func WithKustomize() ManifestOption {
	return func(c *manifestConfig) { c.kustomize = true }
}

// WithKustomizeCommand overrides the external renderer command (default
// `kustomize build`). For example WithKustomizeCommand("kubectl", "kustomize")
// renders via `kubectl kustomize <dir>`. The target directory is always appended as
// the final argument. Implies WithKustomize.
func WithKustomizeCommand(name string, args ...string) ManifestOption {
	return func(c *manifestConfig) {
		c.kustomize = true
		c.render = execRenderer(name, args...)
	}
}

// WithRenderer injects a custom directory→YAML renderer, bypassing the external
// command. It is the seam used by tests and the one through which the Kubernetes
// shim plugs the in-process kustomize API (sigs.k8s.io/kustomize) without the
// adapter ever importing it. Implies WithKustomize.
func WithRenderer(fn func(dir string) ([]byte, error)) ManifestOption {
	return func(c *manifestConfig) {
		c.kustomize = true
		c.render = fn
	}
}

// WithKustomizeSwitch makes kustomize rendering switchable at read time: on is
// consulted for every path the adapter reads, and while it reports false a
// kustomization root is read as the plain files it contains (exactly what an adapter
// without WithKustomize does). It exists for an off switch the user flips on the
// command line (`--no-kustomize`) *after* the step was registered. Only meaningful
// together with WithKustomize, WithKustomizeCommand or WithRenderer.
func WithKustomizeSwitch(on func() bool) ManifestOption {
	return func(c *manifestConfig) { c.enabled = on }
}

// execRenderer builds the default renderer: it runs `name args... dir` and returns
// the rendered YAML on stdout, surfacing a failed build (binary missing, invalid
// kustomization) as an error including the tool's stderr.
func execRenderer(name string, args ...string) func(string) ([]byte, error) {
	return func(dir string) ([]byte, error) {
		cmdArgs := append(append([]string{}, args...), dir)
		cmd := exec.Command(name, cmdArgs...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("kustomize render failed (%s %s): %w: %s",
				name, strings.Join(cmdArgs, " "), err, strings.TrimSpace(stderr.String()))
		}
		return stdout.Bytes(), nil
	}
}

// ManifestAdapter reads Kubernetes-style YAML/JSON manifests from the filesystem
// (a single file, or every *.yaml/*.yml/*.json under a directory tree) and yields
// each document as the SAME nested-HAMT payload the live
// NativeKubernetesResourceAdapter produces. That shared payload shape is the whole
// point: the exact same graph-building DSL — relation rules, walks, folds — runs
// unchanged over offline manifests as over a live cluster, giving true live/offline
// parity (design doc §4.1) and enabling the live→manifest/kustomize transform
// (§6 #10), since both sides speak nested HAMT.
//
// Kustomize libraries are dealt with by rendering them and then treating the output
// as any other YAML stream: `kustomize build` output is itself a stream of YAML
// documents, so the same decoder (and unstructuredToValue) serves both. Rendering is
// a pluggable function (see ManifestOption); a root is only rendered when an option
// asks for it, so a plain adapter reads every file of the directory, kustomization
// files included, exactly as it always did.
//
// The input node's value is the path to read (a file or directory). Multi-document
// files (`---`-separated) are fully supported; empty documents are skipped. A read
// or parse failure is yielded as an error-valued node so evaluator.EitherT lifts it
// to a Left — the branch fails loudly instead of silently vanishing. Documents are
// visited in the lexical order filepath.WalkDir guarantees, so a crawl is
// deterministic.
//
// With WithKustomize, WithKustomizeCommand or WithRenderer, a path that is a
// *kustomization root* is first rendered and the resulting YAML stream is decoded
// identically; a render error (the root and the tool's message) is yielded as an
// error node. Only the path itself is tested for being a root: a directory tree
// that merely *contains* roots is read as plain files, so point the adapter at the
// root you mean.
func ManifestAdapter(opts ...ManifestOption) evaluator.Adapter[any] {
	cfg := manifestConfig{render: execRenderer("kustomize", "build")}
	for _, o := range opts {
		o(&cfg)
	}

	return func(node dsltypes.ASTNode[any]) iter.Seq[dsltypes.ASTNode[any]] {
		return func(yield func(dsltypes.ASTNode[any]) bool) {
			path, ok := node.Value.(string)
			if !ok || path == "" {
				yield(dsltypes.NewASTNode[any](
					fmt.Errorf("ManifestAdapter: expected a file or directory path, got %T", node.Value),
					node.DSLType))
				return
			}

			// Kustomize libraries: render the root to YAML, then decode as usual.
			if cfg.kustomize && (cfg.enabled == nil || cfg.enabled()) && isKustomizationRoot(path) {
				rendered, err := cfg.render(path)
				if err != nil {
					yield(dsltypes.NewASTNode[any](err, node.DSLType))
					return
				}
				docs, err := decodeManifests(bytes.NewReader(rendered))
				if err != nil {
					if !yield(dsltypes.NewASTNode[any](fmt.Errorf("kustomize output: %w", err), node.DSLType)) {
						return
					}
				}
				for _, doc := range docs {
					if !yield(dsltypes.NewASTNode[any](unstructuredToValue(doc), node.DSLType)) {
						return
					}
				}
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

// isKustomizationRoot reports whether dir directly contains a kustomization file,
// i.e. is the root of a kustomize overlay/base that `kustomize build` can render.
func isKustomizationRoot(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	for _, name := range []string{"kustomization.yaml", "kustomization.yml", "Kustomization"} {
		if fi, err := os.Stat(filepath.Join(dir, name)); err == nil && !fi.IsDir() {
			return true
		}
	}
	return false
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
	return decodeManifests(f)
}

// decodeManifests decodes every YAML/JSON document from an arbitrary reader — a
// manifest file on disk or the stdout of `kustomize build` — into the generic
// map[string]any shape, skipping empty documents.
func decodeManifests(r io.Reader) ([]map[string]any, error) {
	decoder := k8syaml.NewYAMLOrJSONDecoder(r, 4096)
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
