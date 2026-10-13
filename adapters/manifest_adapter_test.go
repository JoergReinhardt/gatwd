package adapters_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// TestManifestAdapterKustomizeSwitch proves the render switch is read at read time, not
// when the adapter is built: while it is off a kustomization root is read as the plain
// files it holds (the renderer must not run, and the kustomization file itself is one of
// the documents), once it is on the same adapter renders.
func TestManifestAdapterKustomizeSwitch(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("kustomization.yaml", "resources:\n  - pod.yaml\nnamePrefix: x-\n")
	write("pod.yaml", "kind: Pod\nmetadata:\n  name: raw\n")

	renders := 0
	on := false
	adapter := adapters.WithRenderer(func(string) ([]byte, error) {
		renders++
		return []byte("kind: Pod\nmetadata:\n  name: x-raw\n"), nil
	})

	off := collectPayloads(t, dir, adapter, adapters.WithKustomizeSwitch(func() bool { return on }))
	if renders != 0 {
		t.Fatal("renderer ran while the switch was off")
	}
	if len(off) != 2 { // kustomization.yaml and pod.yaml, in lexical order
		t.Fatalf("switch off: expected the 2 plain files, got %d: %v", len(off), off)
	}
	if name := hamtGet(t, hamtGet(t, off[1], "metadata"), "name"); name != "raw" {
		t.Fatalf("switch off: pod name = %v, want the raw one", name)
	}

	on = true
	rendered := collectPayloads(t, dir, adapter, adapters.WithKustomizeSwitch(func() bool { return on }))
	if renders != 1 || len(rendered) != 1 {
		t.Fatalf("switch on: renders=%d payloads=%d, want 1 and 1", renders, len(rendered))
	}
	if name := hamtGet(t, hamtGet(t, rendered[0], "metadata"), "name"); name != "x-raw" {
		t.Fatalf("switch on: pod name = %v, want x-raw", name)
	}
}

// TestManifestAdapterRenderErrorNamesTheRoot proves a failing renderer is loud: one
// error node, carrying the renderer's message, and no half-rendered resources.
func TestManifestAdapterRenderErrorNamesTheRoot(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "kustomization.yaml"), []byte("resources: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	payloads := collectPayloads(t, dir, adapters.WithRenderer(func(d string) ([]byte, error) {
		return nil, fmt.Errorf("kustomize: rendering %s failed: boom", d)
	}))
	if len(payloads) != 1 {
		t.Fatalf("expected exactly one error node, got %d: %v", len(payloads), payloads)
	}
	err, ok := payloads[0].(error)
	if !ok || !strings.Contains(err.Error(), dir) || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected an error naming %s and the cause, got %v", dir, payloads[0])
	}
}

// sourceOf is the provenance a payload carries ("" when it has none).
func sourceOf(t *testing.T, payload any) (collections.Source, bool) {
	t.Helper()
	return collections.SourceOf(payload)
}

// TestManifestAdapterAttachesSource: every payload remembers its file, and its position
// when the file holds several documents; a rendered kustomization root names the root.
// The source is metadata beside the payload: the payload itself is unchanged.
func TestManifestAdapterAttachesSource(t *testing.T) {
	dir := t.TempDir()
	multi := filepath.Join(dir, "app.yaml")
	single := filepath.Join(dir, "single.yaml")
	for p, content := range map[string]string{
		multi:  "# leading comment\n" + multiDocManifest,
		single: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cfg\n",
	} {
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got := collectPayloads(t, dir)
	want := []collections.Source{
		{Origin: multi, Document: 1},
		{Origin: multi, Document: 2},
		{Origin: single}, // one document: no position
	}
	if len(got) != len(want) {
		t.Fatalf("got %d payloads, want %d", len(got), len(want))
	}
	for i, w := range want {
		if src, ok := sourceOf(t, got[i]); !ok || src != w {
			t.Errorf("payload %d: source %+v (present=%v), want %+v", i, src, ok, w)
		}
	}
	if s := (collections.Source{Origin: multi, Document: 2}).String(); s != multi+", document 2" {
		t.Errorf("Source.String() = %q", s)
	}

	// the payload is what it was without the source: same keys, same values
	if hamtGet(t, got[0], "kind") != "Pod" || got[0].(*collections.HAMT[string, any]).Len() != 4 {
		t.Errorf("attaching a source changed the payload: %v", got[0])
	}

	// a rendered root: the origin is the render, the position counts inside the stream
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "kustomization.yaml"), []byte("resources: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rendered := collectPayloads(t, root, adapters.WithRenderer(func(string) ([]byte, error) { return []byte(multiDocManifest), nil }))
	for i, p := range rendered {
		want := collections.Source{Origin: "kustomize render of " + root, Document: i + 1}
		if src, ok := sourceOf(t, p); !ok || src != want {
			t.Errorf("rendered payload %d: source %+v (present=%v), want %+v", i, src, ok, want)
		}
	}
}
