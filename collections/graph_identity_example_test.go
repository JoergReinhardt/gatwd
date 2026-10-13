package collections_test

import (
	"errors"
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
)

// ExampleGraph_InsertNode shows ingest that refuses to merge two resources which
// were given the same id: qualified ids keep a Pod and a Service named "web"
// apart, while the same qualified id twice is reported.
func ExampleGraph_InsertNode() {
	g := collections.NewGraph[string, string, string, int]()
	g, _ = g.InsertNode("Pod/default/web", "Pod", "pod web")
	g, _ = g.InsertNode("Service/default/web", "Service", "service web") // same name, other kind: fine

	_, err := g.InsertNode("Pod/default/web", "Pod", "another pod web")
	var dup *collections.DuplicateNodeError
	fmt.Println(g.Order(), errors.As(err, &dup))
	fmt.Println(err)
	// Output:
	// 2 true
	// duplicate node id "Pod/default/web": declared as Pod (pod web) and again as Pod (another pod web) — the two declarations are different
}
