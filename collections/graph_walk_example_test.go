package collections_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
)

// dagExample builds the diamond a→b→d / a→c→d with a cheap b-route and an
// expensive c-route.
func dagExample() *collections.Graph[string, string, int, int] {
	g := collections.NewGraph[string, string, int, int]()
	g = g.AddNode("a", "n", 0).AddNode("b", "n", 0).AddNode("c", "n", 0).AddNode("d", "n", 0)
	return g.Connect("a", "b", "e", 1).Connect("b", "d", "e", 1).
		Connect("a", "c", "e", 4).Connect("c", "d", "e", 1)
}

func ExampleGraph_ShortestPath() {
	path, cost, ok := dagExample().ShortestPath("a", "d")
	fmt.Println(path, cost, ok)
	// Output: [a b d] 2 true
}

func ExampleGraph_TopoSort() {
	res := dagExample().TopoSort()
	order, _ := res.UnwrapRight()
	fmt.Println(order)
	// Output: [a b c d]
}

func ExampleGraph_SCC() {
	g := collections.NewGraph[string, string, int, int]()
	g = g.AddNode("a", "n", 0).AddNode("b", "n", 0).AddNode("c", "n", 0)
	g = g.Connect("a", "b", "e", 1).Connect("b", "c", "e", 1).Connect("c", "b", "e", 1)
	fmt.Println(g.SCC())
	// Output: [[a] [b c]]
}
