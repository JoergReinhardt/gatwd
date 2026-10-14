package main

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

const stack = "testdata/stack.yaml"

func strs(t *testing.T, pipeline string) []string {
	t.Helper()
	vals := edition().BuildInterpreter().EvaluateToValues(pipeline)
	if len(vals) != 1 {
		t.Fatalf("%s: want one value, got %d: %v", pipeline, len(vals), vals)
	}
	items, ok := functions.ToSliceAny(vals[0])
	if !ok {
		t.Fatalf("%s: want a list, got %T (%v)", pipeline, vals[0], vals[0])
	}
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = fmt.Sprint(it)
	}
	sort.Strings(out)
	return out
}

// gp_roots requires -D PATTERN (same convention as -D OVERLAY): "**" matches every project,
// a pattern naming one narrows the scan to it — this is what keeps a `visualize --dot` render
// small enough to read without editing tasks.gatwd (see relations.gatwd's gp_roots doc).
func TestGpRootsPatternNarrowsProjects(t *testing.T) {
	roots := func(pattern string) []string {
		t.Helper()
		interp := edition().BuildInterpreter()
		interp.Constants["OVERLAY"] = "staging"
		interp.Constants["PATTERN"] = pattern
		vals := interp.EvaluateToValues(`spawn StrVal: testdata/roots |> walk_fs |> gp_roots |> collect_list`)
		if len(vals) != 1 {
			t.Fatalf("pattern %q: want one value, got %d: %v", pattern, len(vals), vals)
		}
		items, ok := functions.ToSliceAny(vals[0])
		if !ok {
			t.Fatalf("pattern %q: want a list, got %T (%v)", pattern, vals[0], vals[0])
		}
		out := make([]string, len(items))
		for i, it := range items {
			out[i] = fmt.Sprint(it)
		}
		sort.Strings(out)
		return out
	}

	all := roots("**")
	wantAll := []string{"testdata/roots/proj-a/overlays/staging", "testdata/roots/proj-b/overlays/staging"}
	if strings.Join(all, "|") != strings.Join(wantAll, "|") {
		t.Fatalf(`PATTERN="**" = %v, want %v (proj-a's production overlay must stay excluded by OVERLAY)`, all, wantAll)
	}

	narrowed := roots("**/proj-a/**")
	wantNarrowed := []string{"testdata/roots/proj-a/overlays/staging"}
	if strings.Join(narrowed, "|") != strings.Join(wantNarrowed, "|") {
		t.Fatalf("PATTERN=\"**/proj-a/**\" = %v, want %v", narrowed, wantNarrowed)
	}

	// Negative control: a pattern matching nothing must yield an empty list, not an error or
	// the whole unfiltered set.
	none := roots("**/no-such-project/**")
	if len(none) != 0 {
		t.Fatalf(`PATTERN="**/no-such-project/**" = %v, want an empty list`, none)
	}
}

func TestExposesEdges(t *testing.T) {
	got := strs(t, `spawn StrVal: `+stack+` |> manifests |> gitops_graph |> graph_edges_of Exposes |> collect_list`)
	want := []string{
		"Service/demo/app-Exposes->Deployment/demo/app",
		"Service/demo/db-Exposes->StatefulSet/demo/db",
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("Exposes = %v, want %v (the orphan Service must expose nothing; a StatefulSet counts as a workload too)", got, want)
	}
}

func TestRoutesToEdges(t *testing.T) {
	got := strs(t, `spawn StrVal: `+stack+` |> manifests |> gitops_graph |> graph_edges_of RoutesTo |> collect_list`)
	want := []string{
		"Ingress/demo/ext-RoutesTo->Service/demo/app",
		"Ingress/demo/web-RoutesTo->Service/demo/app",
		"Ingress/demo/web-RoutesTo->Service/demo/db",
		"Ingress/demo/web-RoutesTo->Service/demo/orphan",
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("RoutesTo = %v, want %v", got, want)
	}
}

// DependsOn, derived from a URL-shaped ConfigMap value whose host is this cluster's internal
// "<svc>.<ns>.svc" DNS form: it resolves straight to the Service. A real but unrelated external
// URL (UNRELATED_URL) produces no edge at all.
func TestDependsOnEdges(t *testing.T) {
	got := strs(t, `spawn StrVal: `+stack+` |> manifests |> gitops_graph |> graph_edges_of DependsOn |> collect_list`)
	want := []string{"Deployment/demo/app-DependsOn->Service/other/backend"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("DependsOn = %v, want %v", got, want)
	}
}

// Hairpin is the distinct color for the OTHER DependsOn shape: an external-dns hostname this
// cluster answers for resolves to the Ingress declaring that exact host, not to a Service — the
// external load-balancer hop in between is not represented. Keeping it a separate color (not
// just another DependsOn edge) is what lets `graph_edges_of Hairpin` find these on their own.
func TestHairpinEdges(t *testing.T) {
	got := strs(t, `spawn StrVal: `+stack+` |> manifests |> gitops_graph |> graph_edges_of Hairpin |> collect_list`)
	want := []string{"Deployment/demo/app-Hairpin->Ingress/demo/ext"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("Hairpin = %v, want %v", got, want)
	}
}

// The ConfigMap itself is never part of the dependency path: it only gets a Configures edge to
// the workload it feeds, never a DependsOn edge of its own.
func TestConfigMapIsNotInTheDependencyPath(t *testing.T) {
	got := strs(t, `spawn StrVal: `+stack+` |> manifests |> gitops_graph |> graph_edges_of Configures |> collect_list`)
	want := []string{"ConfigMap/demo/app-config-Configures->Deployment/demo/app"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("Configures = %v, want %v", got, want)
	}

	depends := strs(t, `spawn StrVal: `+stack+` |> manifests |> gitops_graph |> graph_edges_of DependsOn |> collect_list`)
	depends = append(depends, strs(t, `spawn StrVal: `+stack+` |> manifests |> gitops_graph |> graph_edges_of Hairpin |> collect_list`)...)
	for _, e := range depends {
		if strings.Contains(e, "ConfigMap") {
			t.Errorf("DependsOn/Hairpin edge %q names a ConfigMap; a ConfigMap must never be part of the dependency path", e)
		}
	}
}

// The view must carry hosts, IP and port names, show the orphan Service with no workload, and
// label a StatefulSet-backed child by its real kind (never hard-coded as "Deployment").
func TestTreeLabels(t *testing.T) {
	vals := edition().BuildInterpreter().EvaluateToValues(`spawn StrVal: ` + stack + ` |> manifests |> gitops_graph |> gp_trees |> visualize`)
	out := fmt.Sprint(vals)
	for _, w := range []string{
		"Ingress web hosts=demo.example.test,orphan.example.test,db.example.test ip=10.0.0.7",
		"Service app via=http ports=http:80",
		"Deployment app containerPorts=http:8080",
		"Service orphan via=http ports=http:80",
		"Service db via=sql ports=sql:5432",
		"StatefulSet db containerPorts=sql:5432",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("tree view lacks %q in %s", w, out)
		}
	}
	// A Tree node must render as its own label with nested branches, never as a generic
	// struct dump: no cons-cell Head/Tail, no synthetic "Struct"/"Value:"/"Children:" layers,
	// no slice-index "[0]:" prefix on what is really a tree child, not an array element.
	for _, noise := range []string{"Head:", "Tail:", "Struct", "Value:", "Children:", "[0]:"} {
		if strings.Contains(out, noise) {
			t.Errorf("tree view leaks internal representation (%q) instead of a plain label tree: %s", noise, out)
		}
	}
}

// Negative control: a missing path must surface as an error, never as an empty success.
func TestMissingPathFailsLoud(t *testing.T) {
	vals := edition().BuildInterpreter().EvaluateToValues(`spawn StrVal: testdata/missing.yaml |> manifests |> gitops_graph`)
	if len(vals) != 1 || !strings.Contains(fmt.Sprint(vals[0]), "no such file") {
		t.Fatalf("a missing manifest must yield an error naming the path, got %v", vals)
	}
}

// A wide scan over several roots must not lose every root's resources because one of them
// cannot be read (a real repo hits this with a broken kustomize component symlink): without
// skip_errors the whole ingest fold withholds its result (pinned here as the negative control),
// with it the good root's graph is still built and the bad one is merely reported.
func TestSkipErrorsIsolatesABrokenRoot(t *testing.T) {
	const scan = `spawn StrVal: x |> emit testdata/stack.yaml, testdata/missing.yaml |> drop 1 |> manifests `

	strict := edition().BuildInterpreter().EvaluateToValues(scan + `|> gitops_graph`)
	if len(strict) != 1 || !strings.Contains(fmt.Sprint(strict[0]), "no such file") {
		t.Fatalf("control: a broken root among several must still fail the whole ingest, got %v", strict)
	}

	got := strs(t, scan+`|> skip_errors |> gitops_graph |> graph_edges_of Exposes |> collect_list`)
	want := []string{
		"Service/demo/app-Exposes->Deployment/demo/app",
		"Service/demo/db-Exposes->StatefulSet/demo/db",
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("skip_errors: Exposes = %v, want %v (the good root's graph, despite the broken one)", got, want)
	}
}
