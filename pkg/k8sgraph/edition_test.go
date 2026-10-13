package k8sgraph_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
)

const stackPath = "testdata/stack.yaml"

func edgesOf(t *testing.T, interp *cli.Interpreter, pipeline string) []string {
	t.Helper()
	vals := interp.EvaluateToValues(pipeline)
	if len(vals) != 1 {
		t.Fatalf("%s: expected one value, got %d: %v", pipeline, len(vals), vals)
	}
	items, ok := functions.ToSliceAny(vals[0])
	if !ok {
		t.Fatalf("%s: expected a list, got %T (%v)", pipeline, vals[0], vals[0])
	}
	out := make([]string, len(items))
	for i, it := range items {
		out[i], _ = it.(string)
	}
	sort.Strings(out)
	return out
}

// TestPreludeRelations proves the embedded pure-DSL prelude derives Owns, Selects
// and RoutesTo over a fixture manifest, driven through an assembled (offline) edition.
func TestPreludeRelations(t *testing.T) {
	interp := k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s-test"), nil).BuildInterpreter()
	src := `spawn StrVal: ` + stackPath + ` |> manifests |> k8s_graph`

	cases := map[string]struct {
		step string
		want []string
	}{
		"Owns":     {"graph_edges_of Owns", []string{"nginx-deploy-Owns->nginx-rs", "nginx-rs-Owns->nginx-pod"}},
		"Selects":  {"graph_edges_of Selects", []string{"nginx-svc-Selects->nginx-pod"}},
		"RoutesTo": {"graph_edges_of RoutesTo", []string{"web-RoutesTo->nginx-svc"}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := edgesOf(t, interp, src+` |> `+c.step+` |> collect_list`)
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("%s edges = %v, want %v", name, got, c.want)
			}
		})
	}

	t.Run("a Pod the selector does not match is not selected", func(t *testing.T) {
		for _, e := range edgesOf(t, interp, src+` |> graph_edges |> collect_list`) {
			if e == "nginx-svc-Selects->other-pod" {
				t.Fatalf("selector app=nginx must not select other-pod")
			}
		}
	})

	t.Run("transitive reachability from the Ingress", func(t *testing.T) {
		got := edgesOf(t, interp, src+` |> graph_reachable web |> collect_list`)
		want := []string{"nginx-pod", "nginx-svc", "web"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("reachable from web = %v, want %v", got, want)
		}
	})
}

// TestPreludeLoadsFromEmbeddedFS pins the multi-file import chain: prelude.gatwd
// imports its siblings from the embed.FS, so every library definition resolves.
func TestPreludeLoadsFromEmbeddedFS(t *testing.T) {
	interp := cli.NewInterpreter()
	if err := interp.LoadScriptFS(k8sgraph.PreludeFS(), k8sgraph.PreludePath); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"k8s_ingest", "rel_owns", "rel_selects", "rel_routes_to", "k8s_relations", "k8s_apply_order", "k8s_cycles", "k8s_graph"} {
		if _, ok := interp.Funcs[name]; !ok {
			t.Errorf("prelude did not define %s", name)
		}
	}
}

func newStackInterp() *cli.Interpreter {
	return k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s-test"), nil).BuildInterpreter()
}

// ordered is edgesOf without the sort: it keeps the pipeline's own ordering.
func ordered(t *testing.T, interp *cli.Interpreter, pipeline string) []string {
	t.Helper()
	vals := interp.EvaluateToValues(pipeline)
	if len(vals) != 1 {
		t.Fatalf("%s: expected one value, got %d: %v", pipeline, len(vals), vals)
	}
	items, ok := functions.ToSliceAny(vals[0])
	if !ok {
		t.Fatalf("%s: expected a list, got %T", pipeline, vals[0])
	}
	out := make([]string, len(items))
	for i, it := range items {
		out[i], _ = it.(string)
	}
	return out
}

// TestE2ERequestTrace is the "request trace" use case: the path a request takes
// from the Ingress to the Pod that serves it, derived purely from the manifests.
func TestE2ERequestTrace(t *testing.T) {
	interp := newStackInterp()
	got := ordered(t, interp, `spawn StrVal: `+stackPath+` |> manifests |> k8s_graph |> graph_shortest_path web nginx-pod |> collect_list`)
	want := []string{"web", "nginx-svc", "nginx-pod"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("trace web -> nginx-pod = %v, want %v", got, want)
	}
	if none := ordered(t, interp, `spawn StrVal: `+stackPath+` |> manifests |> k8s_graph |> graph_shortest_path web other-pod |> collect_list`); len(none) != 0 {
		t.Fatalf("the Ingress must not reach other-pod, got trace %v", none)
	}
}

// TestE2EDependencyResolve is the "dependency resolve" use case: everything an
// owner transitively manages, and an apply order that puts owners first.
func TestE2EDependencyResolve(t *testing.T) {
	interp := newStackInterp()
	src := `spawn StrVal: ` + stackPath + ` |> manifests |> k8s_graph`

	got := edgesOf(t, interp, src+` |> graph_reachable nginx-deploy |> collect_list`)
	want := []string{"nginx-deploy", "nginx-pod", "nginx-rs"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reachable from nginx-deploy = %v, want %v", got, want)
	}

	order := ordered(t, interp, src+` |> k8s_apply_order |> collect_list`)
	pos := map[string]int{}
	for i, id := range order {
		pos[id] = i
	}
	for _, e := range [][2]string{{"nginx-deploy", "nginx-rs"}, {"nginx-rs", "nginx-pod"}, {"web", "nginx-svc"}, {"nginx-svc", "nginx-pod"}} {
		if from, ok := pos[e[0]]; !ok || from >= pos[e[1]] {
			t.Errorf("apply order %v must place %s before %s", order, e[0], e[1])
		}
	}
}
