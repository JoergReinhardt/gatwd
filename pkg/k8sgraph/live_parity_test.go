package k8sgraph_test

import (
	"reflect"
	"testing"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph/k8sgraphtest"
)

// TestLiveOfflineParity is design §6 #10: the SAME k8s_graph prelude over the merged
// live crawl streams (k8s_resources, built with <+>) derives exactly the graph the
// offline manifest path derives from the same objects — nodes, declared edges and
// inferred edges alike. Source-agnostic downstream is the point of the shared
// nested-HAMT payload shape.
func TestLiveOfflineParity(t *testing.T) {
	for _, fixture := range []string{stackPath, fullStackPath} {
		t.Run(fixture, func(t *testing.T) {
			offline := k8sgraph.ApplyEdition(cli.NewEdition("parity-offline"), nil).BuildInterpreter()
			live := k8sgraph.ApplyEdition(cli.NewEdition("parity-live"), k8sgraphtest.FakeClusterFromFile(t, fixture)).BuildInterpreter()

			offSrc := `spawn StrVal: ` + fixture + ` |> manifests |> k8s_graph`
			liveSrc := `spawn StrVal: default |> k8s_resources |> k8s_graph`
			for _, q := range []string{"graph_nodes", "graph_edges", "graph_edges_inferred"} {
				off := edgesOf(t, offline, offSrc+` |> `+q+` |> collect_list`)
				on := edgesOf(t, live, liveSrc+` |> `+q+` |> collect_list`)
				if len(off) == 0 && q != "graph_edges_inferred" {
					t.Fatalf("%s: offline result is empty — the comparison would be vacuous", q)
				}
				if !reflect.DeepEqual(off, on) {
					t.Errorf("%s differs:\n offline %v\n live    %v", q, off, on)
				}
			}
		})
	}
}

// TestLiveSourceMergesAllKinds: the merged stream really carries every crawled kind.
func TestLiveSourceMergesAllKinds(t *testing.T) {
	live := k8sgraph.ApplyEdition(cli.NewEdition("merge"), k8sgraphtest.FakeClusterFromFile(t, fullStackPath)).BuildInterpreter()
	got := edgesOf(t, live, `spawn StrVal: default |> k8s_resources |> hamt_get kind |> collect_list`)
	want := []string{"ConfigMap", "Deployment", "PersistentVolumeClaim", "Pod", "Role", "RoleBinding", "Service", "Service", "ServiceAccount"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merged kinds = %v, want %v", got, want)
	}
}
