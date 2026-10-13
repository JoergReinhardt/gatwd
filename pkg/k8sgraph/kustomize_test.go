package k8sgraph_test

import (
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/joergreinhardt/gatwd/adapters"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
)

const kdir = "testdata/kustomize"

// docsOf splits a multi-document YAML stream into resources keyed by kind/namespace/name.
func docsOf(t *testing.T, stream string) map[string]map[string]any {
	t.Helper()
	out := map[string]map[string]any{}
	for _, part := range strings.Split(stream, "\n---\n") {
		var doc map[string]any
		if err := yaml.Unmarshal([]byte(part), &doc); err != nil {
			t.Fatalf("not YAML: %v\n%s", err, part)
		}
		if len(doc) == 0 {
			continue
		}
		meta, _ := doc["metadata"].(map[string]any)
		key := doc["kind"].(string) + "/" + str(meta["namespace"]) + "/" + str(meta["name"])
		if _, dup := out[key]; dup {
			t.Fatalf("duplicate resource %s", key)
		}
		out[key] = doc
	}
	return out
}

func str(v any) string { s, _ := v.(string); return s }

func keys(m map[string]map[string]any) []string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func mustRender(t *testing.T, dir string) string {
	t.Helper()
	out, err := k8sgraph.InProcessRenderer()(dir)
	if err != nil {
		t.Fatalf("render %s: %v", dir, err)
	}
	return string(out)
}

// TestInProcessRendererMatchesExpected renders the overlay in process and compares it,
// resource by resource, with the plain manifests written out by hand from the base and
// the overlay. Every transformer the overlay uses must have left its mark: prefix,
// namespace, label, patch, and the references renamed with their targets.
func TestInProcessRendererMatchesExpected(t *testing.T) {
	got := docsOf(t, mustRender(t, kdir+"/overlays/prod"))
	want := docsOf(t, readFile(t, kdir+"/expected-prod/prod.yaml"))

	if len(got) != 4 {
		t.Fatalf("expected 4 rendered resources (3 from the base, 1 from the overlay), got %v", keys(got))
	}
	if !reflect.DeepEqual(keys(got), keys(want)) {
		t.Fatalf("rendered resources %v, want %v", keys(got), keys(want))
	}
	for k, w := range want {
		if !reflect.DeepEqual(got[k], w) {
			t.Errorf("%s differs\n got: %v\nwant: %v", k, got[k], w)
		}
	}
	// Pin the markers explicitly as well, so "rendered" cannot degrade into "read".
	for _, marker := range []string{"prod-web-pod", "prod-web-svc", "namespace: prod", "value: warn", "name: prod-web-config"} {
		if !strings.Contains(mustRender(t, kdir+"/overlays/prod"), marker) {
			t.Errorf("rendered output lacks %q", marker)
		}
	}
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestInProcessRendererMatchesKubectl cross-checks the library against the kustomize
// bundled in kubectl when it is installed (never required, never networked).
func TestInProcessRendererMatchesKubectl(t *testing.T) {
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl not installed; skipping the cross-check against `kubectl kustomize`")
	}
	out, err := exec.Command("kubectl", "kustomize", kdir+"/overlays/prod").Output()
	if err != nil {
		t.Skipf("kubectl kustomize unavailable: %v", err)
	}
	want := docsOf(t, string(out))
	got := docsOf(t, mustRender(t, kdir+"/overlays/prod"))
	if len(want) == 0 {
		t.Fatal("kubectl rendered nothing")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("in-process render differs from `kubectl kustomize`:\n got: %v\nwant: %v", got, want)
	}
}

// TestInProcessRendererRefusals is the negative fixture of the renderer: every overlay
// here must fail, loudly, with a message that names the root and the cause.
func TestInProcessRendererRefusals(t *testing.T) {
	cases := []struct {
		dir  string
		want []string // every substring must appear
	}{
		{"broken", []string{"overlays/broken", "does-not-exist.yaml"}},
		{"remote", []string{"overlays/remote", "remote reference", "https://example.invalid/org/repo//base?ref=v1"}},
		{"remote-config", []string{"overlays/remote-config", "remote reference", "https://example.invalid/settings.env"}},
		{"exec", []string{"overlays/exec", "plugin"}},
		{"escape", []string{"overlays/escape", "security", "not in or below"}},
	}
	for _, c := range cases {
		t.Run(c.dir, func(t *testing.T) {
			out, err := k8sgraph.InProcessRenderer()(kdir + "/overlays/" + c.dir)
			if err == nil {
				t.Fatalf("expected an error, rendered:\n%s", out)
			}
			if out != nil {
				t.Errorf("a failed render must return no output, got %q", out)
			}
			for _, w := range c.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q lacks %q", err, w)
				}
			}
		})
	}
}

// TestInProcessRendererSerialised renders the same root from many goroutines: the
// results must be identical and, under -race, free of data races. Recorded honestly:
// removing renderMu does NOT make this test fail with kustomize v0.21.3 (30 runs of 16
// goroutines under -race), because the library now locks its global schema itself —
// the mutex is defence in depth, the test pins the behavior we promise.
func TestInProcessRendererSerialised(t *testing.T) {
	render := k8sgraph.InProcessRenderer()
	want := mustRender(t, kdir+"/overlays/prod")
	var wg sync.WaitGroup
	errs := make(chan string, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := render(kdir + "/overlays/prod")
			if err != nil || string(out) != want {
				errs <- "concurrent render differs or failed"
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}

// TestKustomizeStepRendersInProcess runs the adapter the way the `kustomize` step is
// built, with no external binary on the path at all.
func TestKustomizeStepRendersInProcess(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // no kustomize, no kubectl, no git
	adapter := adapters.ManifestAdapter(adapters.WithRenderer(k8sgraph.InProcessRenderer()))
	n := 0
	for node := range adapter(dsltypes.NewASTNode[any](kdir+"/overlays/prod", dsltypes.BitFlag(0))) {
		if e, ok := node.Value.(error); ok {
			t.Fatalf("render failed without external binaries: %v", e)
		}
		n++
	}
	if n != 4 {
		t.Fatalf("expected 4 rendered resources, got %d", n)
	}
}
