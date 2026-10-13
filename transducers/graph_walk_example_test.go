package transducers_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/transducers"
)

// ExampleTransduceWalk drives a transducer over a cycle-safe breadth-first walk,
// projecting each visited node's payload in one fused pass.
func ExampleTransduceWalk() {
	g := collections.NewGraph[string, string, string, int]()
	g = g.AddNode("ingress", "Ingress", "ing").
		AddNode("svc", "Service", "my-svc").
		AddNode("pod", "Pod", "my-pod")
	g = g.Connect("ingress", "svc", "routes", 1).
		Connect("svc", "pod", "selects", 1).
		Connect("pod", "ingress", "back", 1) // cycle: the walk still terminates

	names := transducers.TransduceWalk(g, "ingress", collections.WalkBFS,
		transducers.Map[functions.Node[string, string, string], string, []string](
			func(n functions.Node[string, string, string]) string { return n.Payload },
		),
	)
	fmt.Println(names)
	// Output: [ing my-svc my-pod]
}
