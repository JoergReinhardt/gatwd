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
