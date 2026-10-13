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

// ExampleHAMT_WithSource shows provenance: a record remembers the file it was read
// from beside its contents, so the duplicate error can name both files. The source is
// invisible to equality — two records with the same contents are equal wherever they
// were read.
func ExampleHAMT_WithSource() {
	record := collections.NewHAMT[string, any]().Set("name", "web")
	a := record.WithSource(collections.Source{Origin: "base/web.yaml"})
	b := record.WithSource(collections.Source{Origin: "overlay/all.yaml", Document: 2})
	fmt.Println(a.Equal(b))

	g := collections.NewGraph[string, string, any, int]()
	g, _ = g.InsertNode("Pod/default/web", "Pod", any(a))
	_, err := g.InsertNode("Pod/default/web", "Pod", any(b))
	var dup *collections.DuplicateNodeError
	if errors.As(err, &dup) {
		fmt.Println(dup.ExistingSource, "|", dup.NewSource)
	}
	// Output:
	// true
	// base/web.yaml | overlay/all.yaml, document 2
}
