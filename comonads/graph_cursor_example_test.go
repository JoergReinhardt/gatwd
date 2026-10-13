package comonads_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/comonads"
)

// ExampleExtendGraph shows the comonadic cobind recomputing every node's payload
// from its local neighbourhood — here each node's new payload is its out-degree.
func ExampleExtendGraph() {
	g := collections.NewGraph[string, string, string, int]()
	g = g.AddNode("a", "n", "").AddNode("b", "n", "").AddNode("c", "n", "")
	g = g.Connect("a", "b", "e", 1).Connect("a", "c", "e", 1)

	degrees := comonads.ExtendGraph(g, func(c *comonads.GraphCursor[string, string, string, int]) int {
		return len(c.Successors())
	})

	a, _ := degrees.GetNode("a")
	b, _ := degrees.GetNode("b")
	fmt.Println("a out-degree:", a.Payload)
	fmt.Println("b out-degree:", b.Payload)
	// Output:
	// a out-degree: 2
	// b out-degree: 0
}
