package k8sgraph_test

import (
	"os"
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
		"Owns":     {"graph_edges_of Owns", []string{"Deployment/default/nginx-deploy-Owns->ReplicaSet/default/nginx-rs", "ReplicaSet/default/nginx-rs-Owns->Pod/default/nginx-pod"}},
		"Selects":  {"graph_edges_of Selects", []string{"Service/default/nginx-svc-Selects->Pod/default/nginx-pod"}},
		"RoutesTo": {"graph_edges_of RoutesTo", []string{"Ingress/default/web-RoutesTo->Service/default/nginx-svc"}},
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
			if e == "Service/default/nginx-svc-Selects->Pod/default/other-pod" {
				t.Fatalf("selector app=nginx must not select other-pod")
			}
		}
	})

	t.Run("transitive reachability from the Ingress", func(t *testing.T) {
		got := edgesOf(t, interp, src+` |> graph_reachable Ingress/default/web |> collect_list`)
		want := []string{"Ingress/default/web", "Pod/default/nginx-pod", "Service/default/nginx-svc"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("reachable from Ingress/default/web = %v, want %v", got, want)
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
	got := ordered(t, interp, `spawn StrVal: `+stackPath+` |> manifests |> k8s_graph |> graph_shortest_path Ingress/default/web Pod/default/nginx-pod |> collect_list`)
	want := []string{"Ingress/default/web", "Service/default/nginx-svc", "Pod/default/nginx-pod"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("trace Ingress/default/web -> Pod/default/nginx-pod = %v, want %v", got, want)
	}
	if none := ordered(t, interp, `spawn StrVal: `+stackPath+` |> manifests |> k8s_graph |> graph_shortest_path Ingress/default/web Pod/default/other-pod |> collect_list`); len(none) != 0 {
		t.Fatalf("the Ingress must not reach other-pod, got trace %v", none)
	}
}

// TestE2EDependencyResolve is the "dependency resolve" use case: everything an
// owner transitively manages, and an apply order that puts owners first.
func TestE2EDependencyResolve(t *testing.T) {
	interp := newStackInterp()
	src := `spawn StrVal: ` + stackPath + ` |> manifests |> k8s_graph`

	got := edgesOf(t, interp, src+` |> graph_reachable Deployment/default/nginx-deploy |> collect_list`)
	want := []string{"Deployment/default/nginx-deploy", "Pod/default/nginx-pod", "ReplicaSet/default/nginx-rs"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reachable from Deployment/default/nginx-deploy = %v, want %v", got, want)
	}

	order := ordered(t, interp, src+` |> k8s_apply_order |> collect_list`)
	pos := map[string]int{}
	for i, id := range order {
		pos[id] = i
	}
	for _, e := range [][2]string{{"Deployment/default/nginx-deploy", "ReplicaSet/default/nginx-rs"}, {"ReplicaSet/default/nginx-rs", "Pod/default/nginx-pod"}, {"Ingress/default/web", "Service/default/nginx-svc"}, {"Service/default/nginx-svc", "Pod/default/nginx-pod"}} {
		if from, ok := pos[e[0]]; !ok || from >= pos[e[1]] {
			t.Errorf("apply order %v must place %s before %s", order, e[0], e[1])
		}
	}
}

const fullStackPath = "testdata/stack_full.yaml"

// TestPodSpecRelations proves the Phase 8 relation matrix — Configures, Mounts,
// Binds and the INFERRED Calls — derived purely in the DSL prelude from pod specs,
// Deployment pod templates, RBAC bindings and plain data values.
func TestPodSpecRelations(t *testing.T) {
	interp := newStackInterp()
	src := `spawn StrVal: ` + fullStackPath + ` |> manifests |> k8s_graph`

	cases := map[string]struct {
		step string
		want []string
	}{
		// volume + env reference from app-pod are ONE fact (graph_add_edges adds an identical
		// spec once), plus the Deployment and StatefulSet templates' envFrom; missing-config
		// has no node, so no edge.
		"Configures": {"graph_edges_of Configures", []string{"ConfigMap/default/app-config-Configures->Deployment/default/worker", "ConfigMap/default/app-config-Configures->Pod/default/app-pod", "ConfigMap/default/app-config-Configures->StatefulSet/default/db"}},
		"Mounts":     {"graph_edges_of Mounts", []string{"PersistentVolumeClaim/default/app-data-Mounts->Pod/default/app-pod"}},
		// the User subject is not a ServiceAccount, so only app-sa (and the Role) bind.
		"Binds":    {"graph_edges_of Binds", []string{"RoleBinding/default/app-binding-Binds->Role/default/app-role", "RoleBinding/default/app-binding-Binds->ServiceAccount/default/app-sa"}},
		"Selects":  {"graph_edges_of Selects", []string{"Service/default/nginx-svc-Selects->Pod/default/app-pod"}},
		"Inferred": {"graph_edges_inferred", []string{"ConfigMap/default/app-config-Calls->Service/default/nginx-svc", "Pod/default/app-pod-Calls->Service/default/backend-svc"}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := edgesOf(t, interp, src+` |> `+c.step+` |> collect_list`)
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("%s = %v, want %v", name, got, c.want)
			}
		})
	}

	t.Run("Calls edges are inferred, declared relations are not", func(t *testing.T) {
		declared := edgesOf(t, interp, src+` |> graph_edges_of Calls |> collect_list`)
		inferred := edgesOf(t, interp, src+` |> graph_edges_inferred |> collect_list`)
		if len(inferred) == 0 || len(declared) != len(inferred) {
			t.Fatalf("Calls must all be inferred: declared=%v inferred=%v", declared, inferred)
		}
		for _, e := range edgesOf(t, interp, src+` |> graph_edges_inferred |> collect_list`) {
			if e != "ConfigMap/default/app-config-Calls->Service/default/nginx-svc" && e != "Pod/default/app-pod-Calls->Service/default/backend-svc" {
				t.Errorf("unexpected inferred edge %s", e)
			}
		}
	})
}

// TestSelectsIndexedPassSemantics pins the label-index optimisation to the naive
// definition (selector ⊆ labels): a multi-pair selector needs ALL pairs, an
// unlabeled Pod is never selected, a selector pair no Pod has selects nothing, and
// an empty selector selects every labeled Pod.
func TestSelectsIndexedPassSemantics(t *testing.T) {
	manifest := `
apiVersion: v1
kind: Pod
metadata: {name: both, labels: {app: web, tier: front}}
---
apiVersion: v1
kind: Pod
metadata: {name: app-only, labels: {app: web}}
---
apiVersion: v1
kind: Pod
metadata: {name: bare}
---
apiVersion: v1
kind: Service
metadata: {name: strict}
spec:
  selector: {app: web, tier: front}
---
apiVersion: v1
kind: Service
metadata: {name: loose}
spec:
  selector: {app: web}
---
apiVersion: v1
kind: Service
metadata: {name: nobody}
spec:
  selector: {app: ghost}
---
apiVersion: v1
kind: Service
metadata: {name: everything}
spec:
  selector: {}
`
	path := t.TempDir() + "/sel.yaml"
	if err := os.WriteFile(path, []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	got := edgesOf(t, newStackInterp(), `spawn StrVal: `+path+` |> manifests |> k8s_graph |> graph_edges_of Selects |> collect_list`)
	// no manifest here names a namespace, so every id has the scope "_"
	want := []string{
		"Service/_/everything-Selects->Pod/_/app-only", "Service/_/everything-Selects->Pod/_/both",
		"Service/_/loose-Selects->Pod/_/app-only", "Service/_/loose-Selects->Pod/_/both",
		"Service/_/strict-Selects->Pod/_/both",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Selects = %v\nwant     %v", got, want)
	}
}
