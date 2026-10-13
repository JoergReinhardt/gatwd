package adapters_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/joergreinhardt/gatwd/adapters"
	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
)

const multiDocManifest = `apiVersion: v1
kind: Pod
metadata:
  name: nginx-pod
  namespace: default
spec:
  selector:
    app: nginx
---
apiVersion: v1
kind: Service
metadata:
  name: web
spec:
  selector:
    app: nginx
`

func collectPayloads(t *testing.T, path string, opts ...adapters.ManifestOption) []any {
	t.Helper()
	adapter := adapters.ManifestAdapter(opts...)
	var payloads []any
	for node := range adapter(dsltypes.NewASTNode[any](path, dsltypes.BitFlag(0))) {
		payloads = append(payloads, node.Value)
	}
	return payloads
}

func TestManifestAdapterMultiDoc(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "app.yaml")
	if err := os.WriteFile(file, []byte(multiDocManifest), 0o644); err != nil {
		t.Fatal(err)
	}

	payloads := collectPayloads(t, file)
	if len(payloads) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(payloads))
	}

	// Same nested-HAMT shape as the live adapter: deep navigation works, and the
	// payload is a HAMT, not a bare name.
	pod := payloads[0]
	if _, ok := pod.(*collections.HAMT[string, any]); !ok {
		t.Fatalf("expected *HAMT payload, got %T", pod)
	}
	if kind := hamtGet(t, pod, "kind"); kind != "Pod" {
		t.Fatalf("doc0 kind = %v, want Pod", kind)
	}
	meta := hamtGet(t, pod, "metadata")
	if name := hamtGet(t, meta, "name"); name != "nginx-pod" {
		t.Fatalf("doc0 metadata.name = %v, want nginx-pod", name)
	}
	spec := hamtGet(t, pod, "spec")
	selector := hamtGet(t, spec, "selector")
	if app := hamtGet(t, selector, "app"); app != "nginx" {
		t.Fatalf("doc0 spec.selector.app = %v, want nginx", app)
	}

	if kind := hamtGet(t, payloads[1], "kind"); kind != "Service" {
		t.Fatalf("doc1 kind = %v, want Service", kind)
	}
}

func TestManifestAdapterDirectoryWalk(t *testing.T) {
	dir := t.TempDir()
	// Two files; WalkDir visits them in lexical order (a.yaml before b.yaml).
	if err := os.WriteFile(filepath.Join(dir, "a.yaml"), []byte("kind: Pod\nmetadata:\n  name: one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.yml"), []byte("kind: Service\nmetadata:\n  name: two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A non-manifest file must be ignored.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# not a manifest"), 0o644); err != nil {
		t.Fatal(err)
	}

	payloads := collectPayloads(t, dir)
	if len(payloads) != 2 {
		t.Fatalf("expected 2 manifest docs (README.md ignored), got %d", len(payloads))
	}
	if kind := hamtGet(t, payloads[0], "kind"); kind != "Pod" {
		t.Fatalf("lexical order broken: first doc kind = %v, want Pod", kind)
	}
	if kind := hamtGet(t, payloads[1], "kind"); kind != "Service" {
		t.Fatalf("second doc kind = %v, want Service", kind)
	}
}

func TestManifestAdapterMissingPath(t *testing.T) {
	payloads := collectPayloads(t, filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if len(payloads) != 1 {
		t.Fatalf("expected 1 error node, got %d", len(payloads))
	}
	if _, ok := payloads[0].(error); !ok {
		t.Fatalf("expected an error payload for a missing path, got %T", payloads[0])
	}
}

// TestManifestAdapterKustomizeRender proves a kustomization root is rendered (via
// the pluggable renderer) and the rendered YAML is then decoded like any other
// manifest — the adapter itself stays kustomize-agnostic.
func TestManifestAdapterKustomizeRender(t *testing.T) {
	dir := t.TempDir()
	// Any kustomization file makes the dir a render root; content is irrelevant
	// here because the renderer is injected.
	if err := os.WriteFile(filepath.Join(dir, "kustomization.yaml"), []byte("resources: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var renderedDir string
	fakeRender := func(d string) ([]byte, error) {
		renderedDir = d
		return []byte(multiDocManifest), nil
	}

	payloads := collectPayloads(t, dir, adapters.WithRenderer(fakeRender))
	if renderedDir != dir {
		t.Fatalf("renderer called with %q, want %q", renderedDir, dir)
	}
	if len(payloads) != 2 {
		t.Fatalf("expected 2 rendered documents, got %d", len(payloads))
	}
	if kind := hamtGet(t, payloads[0], "kind"); kind != "Pod" {
		t.Fatalf("rendered doc0 kind = %v, want Pod", kind)
	}
	if kind := hamtGet(t, payloads[1], "kind"); kind != "Service" {
		t.Fatalf("rendered doc1 kind = %v, want Service", kind)
	}
}

// TestManifestAdapterKustomizeFallsBackToWalk proves that a Kustomize-enabled
// adapter still treats a *plain* directory (no kustomization file) as raw manifests
// — the renderer must not run.
func TestManifestAdapterKustomizeFallsBackToWalk(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pod.yaml"), []byte("kind: Pod\nmetadata:\n  name: one\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	called := false
	fakeRender := func(string) ([]byte, error) {
		called = true
		return nil, nil
	}

	payloads := collectPayloads(t, dir, adapters.WithRenderer(fakeRender))
	if called {
		t.Fatal("renderer ran on a plain (non-kustomization) directory")
	}
	if len(payloads) != 1 {
		t.Fatalf("expected 1 walked document, got %d", len(payloads))
	}
	if kind := hamtGet(t, payloads[0], "kind"); kind != "Pod" {
		t.Fatalf("walked doc kind = %v, want Pod", kind)
	}
}

// TestManifestAdapterKustomizeEndToEnd renders a real kustomization with the host's
// kustomize renderer (kubectl kustomize), gated on that binary being present.
func TestManifestAdapterKustomizeEndToEnd(t *testing.T) {
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl not installed; skipping real kustomize render")
	}

	dir := t.TempDir()
	resource := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: base-cm\ndata:\n  key: value\n"
	if err := os.WriteFile(filepath.Join(dir, "cm.yaml"), []byte(resource), 0o644); err != nil {
		t.Fatal(err)
	}
	kustomization := "resources:\n  - cm.yaml\nnamePrefix: rendered-\n"
	if err := os.WriteFile(filepath.Join(dir, "kustomization.yaml"), []byte(kustomization), 0o644); err != nil {
		t.Fatal(err)
	}

	payloads := collectPayloads(t, dir, adapters.WithKustomizeCommand("kubectl", "kustomize"))
	if len(payloads) != 1 {
		t.Fatalf("expected 1 rendered resource, got %d: %v", len(payloads), payloads)
	}
	// The namePrefix transformer proves it was genuinely rendered by kustomize,
	// not read raw.
	meta := hamtGet(t, payloads[0], "metadata")
	if name := hamtGet(t, meta, "name"); name != "rendered-base-cm" {
		t.Fatalf("rendered metadata.name = %v, want rendered-base-cm (namePrefix applied)", name)
	}
}
