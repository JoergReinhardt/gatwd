package optics_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/optics"
)

// ExampleNodePayloadLens focuses a node's payload and performs an immutable update
// that preserves the node's color and the rest of the graph.
func ExampleNodePayloadLens() {
	g := collections.NewGraph[string, string, int, int]().AddNode("x", "Pod", 1)
	lens := optics.NodePayloadLens[string, string, int, int]("x")

	updated := lens.Set(g, lens.Get(g).Map(func(v int) int { return v + 10 }))

	v, _ := lens.Get(updated).Unwrap()
	fmt.Println("new payload:", v)
	// original is untouched
	orig, _ := lens.Get(g).Unwrap()
	fmt.Println("original:", orig)
	// Output:
	// new payload: 11
	// original: 1
}
