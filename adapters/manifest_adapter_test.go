package adapters_test

import (
	"os"
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

func collectPayloads(t *testing.T, path string) []any {
	t.Helper()
	adapter := adapters.ManifestAdapter()
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
