package k8sgraph_test

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
)

const (
	collisionsPath = "testdata/collisions.yaml"
	scopesPath     = "testdata/scopes.yaml"
	duplicatesPath = "testdata/duplicates.yaml"
)

func offlineGraph(t *testing.T, path, step string) []string {
	t.Helper()
	return edgesOf(t, newStackInterp(), `spawn StrVal: `+path+` |> manifests |> k8s_graph |> `+step+` |> collect_list`)
}

// TestQualifiedIdsKeepCollidingNamesApart (D6): the fixture holds a Service, a Pod, a
// ConfigMap and an Ingress all called "web", and "web" Services, Pods and ConfigMaps in
// two namespaces. With metadata.name as the id they were one node each (the last one
// read); every one is now its own node, and references resolve inside their namespace.
func TestQualifiedIdsKeepCollidingNamesApart(t *testing.T) {
	nodes := offlineGraph(t, collisionsPath, "graph_nodes")
	wantNodes := []string{
		"ConfigMap/default/web", "ConfigMap/prod/web", "Ingress/default/web",
		"Pod/default/web", "Pod/prod/web", "Role/prod/web", "RoleBinding/prod/web",
		"Service/default/web", "Service/prod/web", "ServiceAccount/default/web",
	}
	if !reflect.DeepEqual(nodes, wantNodes) {
		t.Fatalf("nodes = %v\nwant   %v", nodes, wantNodes)
	}

	edges := offlineGraph(t, collisionsPath, "graph_edges")
	wantEdges := []string{
		"ConfigMap/default/web-Configures->Pod/default/web",
		"ConfigMap/prod/web-Configures->Pod/prod/web",
		"Ingress/default/web-RoutesTo->Service/default/web",
		"Pod/default/web-Calls->Service/default/web", // the env var "web" names default's Service
		"RoleBinding/prod/web-Binds->Role/prod/web",
		"RoleBinding/prod/web-Binds->ServiceAccount/default/web", // the subject names its own namespace
		"Service/default/web-Selects->Pod/default/web",
		"Service/prod/web-Selects->Pod/prod/web",
	}
	if !reflect.DeepEqual(edges, wantEdges) {
		t.Fatalf("edges = %v\nwant   %v", edges, wantEdges)
	}
	for _, e := range edges {
		// a relation never crosses a namespace except through a reference that names one
		if strings.Contains(e, "Selects") && strings.Count(e, "/default/")+strings.Count(e, "/prod/") != 2 {
			t.Errorf("malformed edge %s", e)
		}
		if e == "Service/default/web-Selects->Pod/prod/web" || e == "Service/prod/web-Selects->Pod/default/web" {
			t.Errorf("a Service selected a Pod of another namespace: %s", e)
		}
	}
}

// TestUnscopedResourcesUseTheEmptyScope: a resource without a namespace is Kind/_/name,
// and a reference from a namespaced resource falls back to that node when the namespaced
// one does not exist (a cluster-scoped target, or manifests that name no namespace).
func TestUnscopedResourcesUseTheEmptyScope(t *testing.T) {
	if got, want := offlineGraph(t, scopesPath, "graph_nodes"), []string{"ConfigMap/_/shared", "Pod/prod/user", "Role/_/shared", "RoleBinding/prod/grant"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("nodes = %v, want %v", got, want)
	}
	want := []string{"ConfigMap/_/shared-Configures->Pod/prod/user", "RoleBinding/prod/grant-Binds->Role/_/shared"}
	if got := offlineGraph(t, scopesPath, "graph_edges"); !reflect.DeepEqual(got, want) {
		t.Fatalf("edges = %v, want %v", got, want)
	}
}

// TestDuplicateIdIsAnErrorNamingBoth is the NEGATIVE fixture of D6: the same
// kind/namespace/name defined twice is an [ERROR] that names the id and both
// declarations, on every way of reading the graph — never one node holding whichever
// document came last. The third document (same name, another namespace) is not a duplicate.
func TestDuplicateIdIsAnErrorNamingBoth(t *testing.T) {
	for _, step := range []string{"graph_nodes |> collect_list", "graph_edges |> collect_list", "graph_order"} {
		out := newStackInterp().EvaluateToValues(`spawn StrVal: ` + duplicatesPath + ` |> manifests |> k8s_graph |> ` + step)
		if len(out) == 0 {
			t.Fatalf("%s: a graph with a duplicate id produced no output at all", step)
		}
		got := ""
		for _, v := range out {
			if _, isErr := v.(error); !isErr {
				t.Errorf("%s: got a %T (%v), want an error value", step, v, v)
			}
			got += fmt.Sprint(v) + "\n"
		}
		for _, want := range []string{`duplicate node id "Pod/default/dup"`, `"app":"first"`, `"app":"second"`} {
			if !strings.Contains(got, want) {
				t.Errorf("%s: output does not contain %q:\n%s", step, want, got)
			}
		}
		if strings.Contains(got, "Pod/other/dup") {
			t.Errorf("%s: the same name in another namespace was reported as a duplicate:\n%s", step, got)
		}
	}
}

// TestSelectsIsNamespaceLocal: a Service selects Pods of its own namespace only, also for
// the empty selector that falls back to "every Pod".
func TestSelectsIsNamespaceLocal(t *testing.T) {
	manifest := `
apiVersion: v1
kind: Pod
metadata: {name: p, namespace: a, labels: {app: x}}
---
apiVersion: v1
kind: Pod
metadata: {name: p, namespace: b, labels: {app: x}}
---
apiVersion: v1
kind: Service
metadata: {name: s, namespace: a}
spec:
  selector: {app: x}
---
apiVersion: v1
kind: Service
metadata: {name: all, namespace: b}
spec:
  selector: {}
`
	path := t.TempDir() + "/ns.yaml"
	if err := os.WriteFile(path, []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	want := []string{"Service/a/s-Selects->Pod/a/p", "Service/b/all-Selects->Pod/b/p"}
	if got := offlineGraph(t, path, "graph_edges_of Selects"); !reflect.DeepEqual(got, want) {
		t.Fatalf("Selects = %v, want %v", got, want)
	}
}

// TestIdHelpers pins the id vocabulary of lib/k8s/ids.gatwd.
func TestIdHelpers(t *testing.T) {
	in := k8sgraph.ApplyEdition(cli.NewEdition("ids"), nil).BuildInterpreter()
	for stmt, want := range map[string]string{
		`k8s_id "Pod" "default" "web"`:          "Pod/default/web",
		`k8s_short "Service/prod/web"`:          "web",
		`k8s_id_scope "Service/prod/web"`:       "prod",
		`k8s_id_kind "Service/prod/web"`:        "Service",
		`k8s_short (k8s_id "Role" "_" "admin")`: "admin",
		`k8s_ref "Pod" "a" "x" graph_empty`:     "Pod/a/x", // a dangling reference reads as the id it looked for

	} {
		got := in.EvaluateToValues(stmt)
		if len(got) != 1 || fmt.Sprint(got[0]) != want {
			t.Errorf("%s = %v, want %s", stmt, got, want)
		}
	}
}
