package k8sgraph_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
)

// syntheticCluster renders a manifest of the given size: `services` Services, each
// selecting `pods/services` Pods, each Pod owned by a ReplicaSet owned by a
// Deployment (one per service), each Pod referencing a shared ConfigMap, plus one
// Ingress per ten services. It is the workload the O(services x pods) Selects pass
// and the pod-spec passes are measured against.
func syntheticCluster(services, pods int) string {
	var b strings.Builder
	doc := func(s string) { b.WriteString(s); b.WriteString("---\n") }
	doc("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: shared\ndata:\n  k: v\n")
	for s := 0; s < services; s++ {
		doc(fmt.Sprintf("apiVersion: v1\nkind: Service\nmetadata:\n  name: svc-%d\nspec:\n  selector:\n    app: app-%d\n", s, s))
		doc(fmt.Sprintf("apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: dep-%d\n", s))
		doc(fmt.Sprintf("apiVersion: apps/v1\nkind: ReplicaSet\nmetadata:\n  name: rs-%d\n  ownerReferences:\n    - name: dep-%d\n", s, s))
		if s%10 == 0 {
			doc(fmt.Sprintf("apiVersion: networking.k8s.io/v1\nkind: Ingress\nmetadata:\n  name: ing-%d\nspec:\n  rules:\n    - http:\n        paths:\n          - backend:\n              service:\n                name: svc-%d\n", s, s))
		}
	}
	for p := 0; p < pods; p++ {
		s := p % services
		doc(fmt.Sprintf("apiVersion: v1\nkind: Pod\nmetadata:\n  name: pod-%d\n  labels:\n    app: app-%d\n  ownerReferences:\n    - name: rs-%d\nspec:\n  volumes:\n    - name: c\n      configMap:\n        name: shared\n", p, s, s))
	}
	return b.String()
}

// BenchmarkK8sGraph measures the full offline pipeline
//
//	manifests |> k8s_graph |> graph_order
//
// (decode -> ingest -> every relation pass) across cluster sizes. Compare sub-
// benchmarks to read the scaling of the quadratic Selects pass; see
// benchmarks/README.md for the recorded results.
func BenchmarkK8sGraph(b *testing.B) {
	for _, size := range []struct{ services, pods int }{{10, 100}, {25, 250}, {50, 500}, {100, 1000}} {
		b.Run(fmt.Sprintf("svc%d_pods%d", size.services, size.pods), func(b *testing.B) {
			path := filepath.Join(b.TempDir(), "cluster.yaml")
			if err := os.WriteFile(path, []byte(syntheticCluster(size.services, size.pods)), 0o644); err != nil {
				b.Fatal(err)
			}
			interp := k8sgraph.ApplyEdition(cli.NewEdition("bench"), nil).BuildInterpreter()
			pipeline := `spawn StrVal: ` + path + ` |> manifests |> k8s_graph |> graph_order`
			if got := interp.EvaluateToValues(pipeline); len(got) != 1 {
				b.Fatalf("warm-up produced %v", got)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				interp.EvaluateToValues(pipeline)
			}
		})
	}
}
