package k8sgraph_test

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
)

// TestCrawlStep drives the domain shim end-to-end from the DSL: a CrawlStep built
// over a fake dynamic client is registered on a real interpreter, and a pipeline
// line crawls the namespace and navigates the nested-HAMT payload — proving the
// adapter→DSL bridge (P2 structured adapter → P3a cli.AdapterStep → P3b CrawlStep)
// is wired through to Layer 5.
func TestCrawlStep(t *testing.T) {
	pod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "nginx-pod",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{"app": "nginx"},
			},
		},
	}
	client := fake.NewSimpleDynamicClient(runtime.NewScheme(), pod)

	interp := cli.NewInterpreter()
	interp.RegisterStep(k8sgraph.CrawlStep("k8s_pods", client, k8sgraph.PodsGVR))

	t.Run("deep field access over the crawled resource", func(t *testing.T) {
		vals := interp.EvaluateToValues(`spawn StrVal: default |> k8s_pods |> hamt_path spec.selector.app`)
		if len(vals) != 1 {
			t.Fatalf("expected 1 value, got %d: %v", len(vals), vals)
		}
		if vals[0] != "nginx" {
			t.Fatalf("spec.selector.app = %v, want nginx", vals[0])
		}
	})

	t.Run("resource yielded as a navigable payload, not a bare name", func(t *testing.T) {
		vals := interp.EvaluateToValues(`spawn StrVal: default |> k8s_pods |> hamt_get kind`)
		if len(vals) != 1 || vals[0] != "Pod" {
			t.Fatalf("hamt_get kind = %v, want [Pod]", vals)
		}
	})
}
