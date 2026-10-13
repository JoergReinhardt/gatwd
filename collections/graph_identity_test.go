package collections

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

type testGraph = Graph[string, string, any, any]

func newTestGraph() *testGraph { return NewGraph[string, string, any, any]() }

// TestGraphInsertNode: InsertNode declares an id exactly once. A second insertion
// of the same id is a *DuplicateNodeError naming the id and both colors/payloads,
// never a silent overwrite (which is what AddNode/WithNode do on purpose).
func TestGraphInsertNode(t *testing.T) {
	g, err := newTestGraph().InsertNode("Pod/default/web", "Pod", "first")
	if err != nil {
		t.Fatalf("first insertion failed: %v", err)
	}
	if !g.HasNode("Pod/default/web") {
		t.Fatal("inserted node is missing")
	}

	g2, err := g.InsertNode("Pod/default/web", "Service", "second")
	if err == nil {
		t.Fatal("duplicate id was accepted silently")
	}
	if g2 != nil {
		t.Errorf("a failed insertion must not return a graph, got %v", g2)
	}
	var dup *DuplicateNodeError
	if !errors.As(err, &dup) {
		t.Fatalf("error is %T, want *DuplicateNodeError", err)
	}
	if dup.ID != "Pod/default/web" || dup.ExistingColor != "Pod" || dup.NewColor != "Service" ||
		dup.ExistingPayload != "first" || dup.NewPayload != "second" {
		t.Errorf("DuplicateNodeError lost information: %+v", dup)
	}
	for _, want := range []string{"Pod/default/web", "Pod", "Service", "first", "second"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("message %q does not name %q", err.Error(), want)
		}
	}

	// the receiver is persistent: the failed insertion changed nothing
	if n, _ := g.GetNode("Pod/default/web"); n.Payload != "first" || n.Color != "Pod" {
		t.Errorf("original graph was modified: %v", n)
	}
	// AddNode keeps its documented upsert meaning
	if n, _ := g.AddNode("Pod/default/web", "Service", "second").GetNode("Pod/default/web"); n.Payload != "second" {
		t.Errorf("AddNode no longer upserts: %v", n)
	}
	// same name in another kind or namespace is a different identity
	for _, id := range []string{"Service/default/web", "Pod/prod/web"} {
		if _, err := g.InsertNode(id, "X", nil); err != nil {
			t.Errorf("distinct id %q rejected: %v", id, err)
		}
	}
}

// TestGraphInsertNodeIdenticalDuplicate: even a byte-identical second definition
// is reported, and the message says so (a resource listed twice).
func TestGraphInsertNodeIdenticalDuplicate(t *testing.T) {
	g, _ := newTestGraph().InsertNode("a", "N", "same")
	_, err := g.InsertNode("a", "N", "same")
	if err == nil || !strings.Contains(err.Error(), "identical") {
		t.Fatalf("identical duplicate: want an error saying identical, got %v", err)
	}
}

// TestGraphAdjacencyIgnoresUndeclaredEndpoints (F13): an edge whose far endpoint
// was never declared as a node is kept (Edges, Size — the graph is not rewritten)
// but is not traversed: neighbours, predecessors, walks and shortest paths agree
// with the node listing.
func TestGraphAdjacencyIgnoresUndeclaredEndpoints(t *testing.T) {
	g := newTestGraph().
		AddNode("a", "N", nil).AddNode("b", "N", nil).
		Connect("a", "b", "E", 1).
		Connect("a", "ghost", "E", 1).
		Connect("phantom", "a", "E", 1)

	if got := g.Size(); got != 3 {
		t.Fatalf("stored edges = %d, want 3 (the graph itself must not be rewritten)", got)
	}
	if got, want := g.Neighbors("a"), []string{"b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Neighbors(a) = %v, want %v (ghost is not a node)", got, want)
	}
	if got := g.Predecessors("a"); len(got) != 0 {
		t.Errorf("Predecessors(a) = %v, want none (phantom is not a node)", got)
	}
	if got, want := g.Reachable("a").Len(), 2; got != want {
		t.Errorf("Reachable(a) has %d members, want %d (a, b)", got, want)
	}
	if path, _, ok := g.ShortestPath("a", "ghost"); ok {
		t.Errorf("ShortestPath reached an undeclared node: %v", path)
	}
	if path, _, ok := g.ShortestPath("a", "b"); !ok || !reflect.DeepEqual(path, []string{"a", "b"}) {
		t.Errorf("ShortestPath(a,b) = %v %v", path, ok)
	}
}

// TestGraphAddEdgesOnce (F14): identical edges within one batch, and edges the
// graph already holds, are added once; distinct colors/weights stay distinct.
func TestGraphAddEdgesOnce(t *testing.T) {
	e := func(from, to, color string, w any) functions.Edge[string, string, any] {
		return functions.NewEdge[string, string, any](from, to, color, w)
	}
	g := newTestGraph().AddNode("a", "N", nil).AddNode("b", "N", nil)
	g = g.AddEdgesOnce([]functions.Edge[string, string, any]{
		e("a", "b", "Configures", 1), e("a", "b", "Configures", 1), // the volume AND the env reference
		e("a", "b", "Owns", 1),         // other color: kept
		e("a", "b", "Configures", 2.5), // other weight: kept
	})
	if got := g.Size(); got != 3 {
		t.Fatalf("Size = %d, want 3 (one Configures/1 deduplicated)", got)
	}
	again := g.AddEdgesOnce([]functions.Edge[string, string, any]{e("a", "b", "Owns", 1)})
	if again.Size() != 3 {
		t.Errorf("an edge already in the graph was added again: Size = %d", again.Size())
	}
	// AddEdge keeps its documented multigraph meaning
	if got := g.AddEdge(e("a", "b", "Owns", 1)).Size(); got != 4 {
		t.Errorf("AddEdge must stay a multigraph insert, Size = %d", got)
	}
}

// TestSourceIsMetadataBesideTheValue: a HAMT's Source is invisible to every reader of the
// map (equality, lookup, iteration, size, JSON), survives Set and Delete, and is dropped by
// the operations that build a fresh map. Without it, the duplicate error reads as before;
// with it, the error names where each payload was read.
func TestSourceIsMetadataBesideTheValue(t *testing.T) {
	plain := NewHAMT[string, any]().Set("a", 1).Set("b", 2)
	a := plain.WithSource(Source{Origin: "a.yaml"})
	b := plain.WithSource(Source{Origin: "b.yaml", Document: 2})

	if _, ok := plain.Source(); ok {
		t.Fatal("a HAMT without WithSource reports a source")
	}
	if a.Len() != 2 {
		t.Fatalf("WithSource lost entries: %d", a.Len())
	}
	if src, ok := a.Source(); !ok || src.Origin != "a.yaml" {
		t.Fatalf("source = %+v, %v", src, ok)
	}
	if !a.Equal(b) || !a.Equal(plain) || !plain.Equal(b) {
		t.Error("equality must ignore the source")
	}
	if !functions.StructuralEqual(a, b) {
		t.Error("StructuralEqual must ignore the source")
	}
	ja, _ := a.MarshalJSON()
	jp, _ := plain.MarshalJSON()
	if string(ja) != string(jp) {
		t.Errorf("JSON shows the source: %s vs %s", ja, jp)
	}
	if v, _ := a.Get("b"); v != 2 {
		t.Errorf("Get on a sourced map = %v", v)
	}
	if _, ok := plain.Source(); ok {
		t.Error("WithSource modified its receiver")
	}

	// Set and Delete keep it (the result is still that record, edited); a bulk rebuild does not
	if src, ok := a.Set("c", 3).Source(); !ok || src.Origin != "a.yaml" {
		t.Errorf("Set dropped the source: %+v %v", src, ok)
	}
	if src, ok := a.Delete("a").Source(); !ok || src.Origin != "a.yaml" {
		t.Errorf("Delete dropped the source: %+v %v", src, ok)
	}
	if src, ok := a.Merge(plain.Set("z", 9)).Source(); !ok || src.Origin != "a.yaml" {
		t.Errorf("Merge keeps the receiver's source: %+v %v", src, ok)
	}
	if _, ok := a.Filter(func(string, any) bool { return true }).Source(); ok {
		t.Error("Filter builds a fresh map and must not claim an origin")
	}
	var nilHAMT *HAMT[string, any]
	if src, ok := nilHAMT.WithSource(Source{Origin: "x"}).Source(); !ok || src.Origin != "x" {
		t.Errorf("WithSource on a nil HAMT: %+v %v", src, ok)
	}

	// SourceOf is for any value: a HAMT with a source, nothing for the rest
	if _, ok := SourceOf("text"); ok {
		t.Error("a String has no source")
	}
	if _, ok := SourceOf(nil); ok {
		t.Error("nil has no source")
	}

	// the duplicate error names both origins, in the payloads' own words
	g, err := newTestGraph().InsertNode("Pod/default/dup", "Pod", any(a))
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.InsertNode("Pod/default/dup", "Pod", any(b))
	var dup *DuplicateNodeError
	if !errors.As(err, &dup) {
		t.Fatalf("want a *DuplicateNodeError, got %v", err)
	}
	if dup.ExistingSource != (Source{Origin: "a.yaml"}) || dup.NewSource != (Source{Origin: "b.yaml", Document: 2}) {
		t.Errorf("sources not recorded on the error: %+v / %+v", dup.ExistingSource, dup.NewSource)
	}
	for _, want := range []string{" in a.yaml (", " in b.yaml, document 2 ("} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("message %q lacks %q", err.Error(), want)
		}
	}

	// payloads without a source: the message is exactly what it was before provenance existed
	g2, _ := newTestGraph().InsertNode("x", "A", "first")
	_, err = g2.InsertNode("x", "B", "second")
	want := `duplicate node id "x": declared as A (first) and again as B (second) — the two declarations are different`
	if err == nil || err.Error() != want {
		t.Errorf("message without sources changed:\n got %v\nwant %s", err, want)
	}
}
