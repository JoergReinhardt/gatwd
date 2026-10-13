package collections_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// ExampleGraph demonstrates building a small typed, weighted graph and reading
// its order, size, and a by-color projection back out.
func ExampleGraph() {
	g := collections.NewGraph[string, string, int, int]()
	g = g.AddNode("svc", "service", 0)
	g = g.AddNode("pod1", "pod", 1)
	g = g.AddNode("pod2", "pod", 2)
	// One logical "selects" relation fanning out to both pods (shared group).
	g = g.AddFanOut("selects", 1, "g1", "svc", []string{"pod1", "pod2"})

	fmt.Printf("|V|=%d |E|=%d\n", g.Order(), g.Size())
	fmt.Println("pods:", len(g.NodesByColor("pod")))
	fmt.Println("neighbors of svc:", g.Neighbors("svc"))
	// Output:
	// |V|=3 |E|=2
	// pods: 2
	// neighbors of svc: [pod1 pod2]
}

// ExampleGraph_bfs shows that a cyclic graph can be walked safely: the traversal
// terminates even though a->b->c->a forms a cycle.
func ExampleGraph_bfs() {
	g := collections.NewGraph[string, string, int, int]().
		AddNode("a", "n", 0).AddNode("b", "n", 0).AddNode("c", "n", 0).
		Connect("a", "b", "e", 1).Connect("b", "c", "e", 1).Connect("c", "a", "e", 1)

	var order []string
	for k := range g.BFS("a") {
		order = append(order, k)
	}
	fmt.Println(order)
	// Output:
	// [a b c]
}

// ExampleGraph_union demonstrates the graph monoid: Union merges node and edge
// sets and de-duplicates identical edges.
func ExampleGraph_union() {
	a := collections.NewGraph[string, string, int, int]().
		AddNode("x", "n", 1).AddNode("y", "n", 2).Connect("x", "y", "e", 1)
	b := collections.NewGraph[string, string, int, int]().
		AddNode("y", "n", 2).AddNode("z", "n", 3).Connect("y", "z", "e", 1)

	u := a.Union(b)
	fmt.Printf("|V|=%d |E|=%d\n", u.Order(), u.Size())
	// Output:
	// |V|=3 |E|=2
}

// ExampleGraph_inferredEdge shows marking an assumed edge (one not read from the
// source data) so walks and folds can trust-weight it separately.
func ExampleGraph_inferredEdge() {
	g := collections.NewGraph[string, string, int, float64]().
		AddNode("pod", "pod", 0).AddNode("svc", "service", 0).
		AddInferredEdge("pod", "svc", "consumes", 0.5)

	for _, e := range g.InferredEdges() {
		fmt.Printf("%s weight=%.1f inferred=%t\n", e.Color, e.Weight, e.Inferred)
	}
	// Output:
	// consumes weight=0.5 inferred=true
}

// ExampleEdge shows the Edge ADT's stringer and its deterministic ordering.
func ExampleEdge() {
	e := functions.NewEdge("a", "b", "calls", 3)
	inferred := functions.NewInferredEdge("a", "b", "calls", 3)
	fmt.Println(e.String())
	fmt.Println(inferred.String())
	fmt.Println(e.Compare(inferred) < 0) // non-inferred sorts before inferred
	// Output:
	// a -calls-> b
	// a -calls-> b~
	// true
}
