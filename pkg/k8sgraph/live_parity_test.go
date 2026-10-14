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
	// ns is what the live crawl asks for: "default" for the single-namespace stacks, all
	// namespaces ("") for the fixture whose point is the same name in two of them.
	for _, fx := range []struct{ fixture, ns string }{{stackPath, `default`}, {fullStackPath, `default`}, {collisionsPath, `""`}} {
		fixture := fx.fixture
		t.Run(fixture, func(t *testing.T) {
			offline := k8sgraph.ApplyEdition(cli.NewEdition("parity-offline"), nil).BuildInterpreter()
			live := k8sgraph.ApplyEdition(cli.NewEdition("parity-live"), k8sgraphtest.FakeClusterFromFile(t, fixture)).BuildInterpreter()

			offSrc := `spawn StrVal: ` + fixture + ` |> manifests |> k8s_graph`
			liveSrc := `spawn StrVal: ` + fx.ns + ` |> k8s_resources |> k8s_graph`
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
	want := []string{"ConfigMap", "Deployment", "PersistentVolumeClaim", "Pod", "Role", "RoleBinding", "Service", "Service", "ServiceAccount", "StatefulSet"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merged kinds = %v, want %v", got, want)
	}
}

// TestLiveCrawlWithoutSecrets backs the how-to's security recipe (§2.4): the standard crawl
// includes Secrets (their payload, `data` included, enters the graph), and an override that
// redefines `k8s_resources` without `k8s_secrets` leaves them out while every other kind is
// still crawled. Both halves are asserted, so neither "Secrets are never read" nor "the
// recipe does nothing" can pass unnoticed.
func TestLiveCrawlWithoutSecrets(t *testing.T) {
	fixture := "testdata/secrets.yaml"
	live := k8sgraph.ApplyEdition(cli.NewEdition("secrets"), k8sgraphtest.FakeClusterFromFile(t, fixture)).BuildInterpreter()
	kinds := `spawn StrVal: default |> k8s_resources |> hamt_get kind |> collect_list`
	want := []string{"Pod", "Secret", "Secret"}
	if got := edgesOf(t, live, kinds); !reflect.DeepEqual(got, want) {
		t.Fatalf("standard crawl kinds = %v, want %v", got, want)
	}

	// The recipe, as the how-to prints it.
	live.EvaluateScript(`def k8s_resources = (k8s_pods <+> k8s_services <+> k8s_configmaps <+> k8s_endpoints <+> k8s_deployments <+> k8s_replicasets <+> k8s_ingresses <+> k8s_pvcs <+> k8s_serviceaccounts <+> k8s_roles <+> k8s_rolebindings)`)
	if got, want := edgesOf(t, live, kinds), []string{"Pod"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("crawl without k8s_secrets: kinds = %v, want %v", got, want)
	}
}
