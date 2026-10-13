package k8sgraph

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/yaml"
)

// renderMu serialises in-process renders. Upstream reported a data race when one
// Kustomizer is used from several goroutines (kubernetes-sigs/kustomize#4824) and
// kustomize keeps process-wide state while building (the OpenAPI schema); v0.21.3
// guards that schema with its own lock, and we still do not let renders overlap,
// because our `||>` parallel steps could otherwise start several at once and a render
// is cheap next to the I/O around it. Every render also gets its own Kustomizer.
var renderMu sync.Mutex

// InProcessRenderer returns a directory→YAML renderer that builds a kustomization
// root inside this process with sigs.k8s.io/kustomize/api — no `kustomize` or
// `kubectl` binary, no subprocess. It is the function adapters.WithRenderer takes,
// and the reason the dependency lives here in the shim and not in the adapter or in
// core.
//
// The output is what `kustomize build dir` prints (documents separated by `---`, in
// kustomize's legacy resource order), so the manifest adapter decodes it exactly
// like any other YAML stream. A failure — a missing resource, an invalid
// kustomization, a refused feature — is returned as an error that names the root
// and carries kustomize's own message; nothing is rendered partially.
//
// Safety defaults, all on:
//
//   - kustomize's default load restriction stays: files outside the kustomization
//     root cannot be read by a patch, a generator or a resource file (a base
//     directory next to the overlay, `../base`, is a base and remains allowed);
//   - exec plugins, KRM functions and Helm stay disabled, so a kustomization cannot
//     run a program (it is refused with kustomize's "plugin disabled" error);
//   - remote bases and files (git and http/https/ssh/file URLs) are refused before
//     kustomize can fetch anything, because the library hard-wires `git clone` and
//     HTTP GET and offers no switch. Enabling them would be a deliberate, separate
//     option; there is none yet.
//
// Concurrent calls are serialised (see renderMu).
func InProcessRenderer() func(dir string) ([]byte, error) {
	return func(dir string) (out []byte, err error) {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("kustomize: cannot render %s: %w", dir, err)
		}
		renderMu.Lock()
		defer renderMu.Unlock()
		defer func() {
			// A library panic must not take the whole process down, nor hide as silence.
			if r := recover(); r != nil {
				out, err = nil, fmt.Errorf("kustomize: rendering %s panicked: %v", dir, r)
			}
		}()

		opts := krusty.MakeDefaultOptions()
		opts.Reorder = krusty.ReorderOptionLegacy // the order `kustomize build` prints
		fsys := &guardedFS{FileSystem: filesys.MakeFsOnDisk()}
		resMap, err := krusty.MakeKustomizer(opts).Run(fsys, abs)
		if refused := fsys.refusal(); refused != nil {
			// kustomize reports a refused read as "unable to find a kustomization"; say why.
			return nil, fmt.Errorf("kustomize: rendering %s refused: %w", dir, refused)
		}
		if err != nil {
			return nil, fmt.Errorf("kustomize: rendering %s failed: %w", dir, err)
		}
		out, err = resMap.AsYaml()
		if err != nil {
			return nil, fmt.Errorf("kustomize: rendering %s failed: %w", dir, err)
		}
		return out, nil
	}
}

// guardedFS is the on-disk filesystem kustomize reads through, with one addition:
// every kustomization file it opens is checked for remote references first, so a
// remote base is refused *before* the library clones or downloads it. All other
// methods pass through.
type guardedFS struct {
	filesys.FileSystem

	mu sync.Mutex
	// configs are the generator/transformer/validator configuration files named by
	// the kustomization files read so far (cleaned absolute paths). They are scanned
	// when read too, because a builtin generator configured in such a file can name
	// remote files.
	configs map[string]bool
	// refused is the first remote reference found; kustomize swallows the read error
	// it caused, so the renderer reports this instead.
	refused error
}

func (g *guardedFS) refuse(err error) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.refused == nil {
		g.refused = err
	}
	return err
}

func (g *guardedFS) refusal() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.refused
}

var kustomizationFileNames = map[string]bool{"kustomization.yaml": true, "kustomization.yml": true, "Kustomization": true}

// ReadFile reads path, refusing a kustomization (or a generator/transformer
// configuration it names) that references a remote base or file.
func (g *guardedFS) ReadFile(path string) ([]byte, error) {
	data, err := g.FileSystem.ReadFile(path)
	if err != nil {
		return data, err
	}
	switch {
	case kustomizationFileNames[filepath.Base(path)]:
		if err := g.guardKustomization(path, data); err != nil {
			return nil, g.refuse(err)
		}
	case g.isConfig(path):
		if err := guardConfig(path, data); err != nil {
			return nil, g.refuse(err)
		}
	}
	return data, nil
}

func (g *guardedFS) isConfig(path string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.configs[canonical(path)]
}

func (g *guardedFS) guardKustomization(path string, data []byte) error {
	var k types.Kustomization
	if err := yaml.Unmarshal(data, &k); err != nil {
		return fmt.Errorf("%s: cannot be checked for remote references: %w", path, err)
	}
	entries := pathEntries(&k)
	for _, e := range entries {
		if isRemoteRef(e) {
			return fmt.Errorf("%s: remote reference %q refused (in-process kustomize renders local files only)", path, e)
		}
	}
	dir := filepath.Dir(path)
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.configs == nil {
		g.configs = map[string]bool{}
	}
	for _, c := range append(append(append([]string{}, k.Generators...), k.Transformers...), k.Validators...) {
		g.configs[canonical(filepath.Join(dir, c))] = true
	}
	return nil
}

// guardConfig refuses a generator/transformer configuration file in which any string
// value is a remote reference. Deliberately blunt (every scalar, not just `files`):
// a false positive is a loud, explained refusal, a false negative a network fetch.
func guardConfig(path string, data []byte) error {
	var doc any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil // not one document of YAML: kustomize will report it
	}
	var bad string
	var walk func(v any)
	walk = func(v any) {
		switch t := v.(type) {
		case string:
			if bad == "" && (isRemoteRef(t) || isRemoteRef(afterEquals(t))) {
				bad = t
			}
		case []any:
			for _, e := range t {
				walk(e)
			}
		case map[string]any:
			for _, e := range t {
				walk(e)
			}
		}
	}
	walk(doc)
	if bad != "" {
		return fmt.Errorf("%s: remote reference %q refused (in-process kustomize renders local files only)", path, bad)
	}
	return nil
}

// pathEntries lists every string of a kustomization that kustomize reads as a path
// or a base: resources, components, bases, crds, configurations, generators,
// transformers, validators, patch and replacement files, the OpenAPI schema, and the
// file/env sources of the config-map and secret generators (`key=path` forms count
// by their path).
func pathEntries(k *types.Kustomization) []string {
	var out []string
	add := func(s ...string) { out = append(out, s...) }
	add(k.Resources...)
	add(k.Components...)
	add(k.Bases...)
	add(k.Crds...)
	add(k.Configurations...)
	add(k.Generators...)
	add(k.Transformers...)
	add(k.Validators...)
	for _, p := range k.PatchesStrategicMerge {
		// A strategic-merge entry is a file name or an inline patch (multi-line YAML).
		if s := string(p); !strings.ContainsAny(s, "\n") {
			add(s)
		}
	}
	for _, p := range k.Patches {
		add(p.Path)
	}
	for _, p := range k.PatchesJson6902 {
		add(p.Path)
	}
	for _, r := range k.Replacements {
		add(r.Path)
	}
	if p, ok := k.OpenAPI["path"]; ok {
		add(p)
	}
	for _, g := range k.ConfigMapGenerator {
		add(g.FileSources...)
		add(g.EnvSources...)
		add(g.EnvSource)
	}
	for _, g := range k.SecretGenerator {
		add(g.FileSources...)
		add(g.EnvSources...)
		add(g.EnvSource)
	}
	// Normalise `key=path` file sources to both halves, so either side is checked.
	var norm []string
	for _, e := range out {
		norm = append(norm, e, afterEquals(e))
	}
	return norm
}

func afterEquals(s string) string {
	if _, rest, ok := strings.Cut(s, "="); ok {
		return rest
	}
	return s
}

// userRef is kustomize's username test for SCP-style git URLs (`git@host:org/repo`).
var userRef = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*@`)

// isRemoteRef reports whether s would be fetched by kustomize instead of read from
// disk. It mirrors kustomize's own decision (internal/git NewRepoSpecFromURL and
// loader.IsRemoteFile, both internal and therefore not importable): a scheme of
// ssh://, https://, http:// or file://, the legacy `git::` prefix, a `user@host`
// SCP-style address, or a GitHub shorthand (`github.com/org/repo`). Anything else is
// a local path.
func isRemoteRef(s string) bool {
	l := strings.ToLower(s)
	for _, p := range []string{"ssh://", "https://", "http://", "file://", "git::", "github.com/", "github.com:"} {
		if strings.HasPrefix(l, p) {
			return true
		}
	}
	return userRef.MatchString(s)
}

// canonical resolves symlinks (best effort) so the path kustomize reads and the path
// recorded from a kustomization compare equal.
func canonical(p string) string {
	p = filepath.Clean(p)
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return p
}
