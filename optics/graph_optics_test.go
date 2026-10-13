package optics_test

import (
	"reflect"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/optics"
)

func optGraph() *collections.Graph[string, string, int, int] {
	g := collections.NewGraph[string, string, int, int]()
	g = g.AddNode("pod", "Pod", 1).AddNode("svc", "Service", 2)
	g = g.Connect("svc", "pod", "selects", 5).Connect("svc", "pod", "owns", 7)
	return g
}

func TestNodePayloadLensGetSet(t *testing.T) {
	g := optGraph()
	lens := optics.NodePayloadLens[string, string, int, int]("pod")

	got := lens.Get(g)
	if !got.IsJust() {
		t.Fatal("Get(pod) = Nothing, want Just")
	}
	if v, _ := got.Unwrap(); v != 1 {
		t.Fatalf("Get(pod) payload = %d, want 1", v)
	}

	updated := lens.Set(g, got.Map(func(int) int { return 42 }))
	if n, _ := updated.GetNode("pod"); n.Payload != 42 {
		t.Fatalf("Set payload = %d, want 42", n.Payload)
	}
	// Original is untouched (immutability).
	if n, _ := g.GetNode("pod"); n.Payload != 1 {
		t.Fatalf("original mutated: pod = %d, want 1", n.Payload)
	}
	// Color preserved through payload set.
	if n, _ := updated.GetNode("pod"); n.Color != "Pod" {
		t.Fatalf("Set dropped color: %q, want Pod", n.Color)
	}
}

func TestNodePayloadLensSetNothingRemoves(t *testing.T) {
	g := optGraph()
	lens := optics.NodePayloadLens[string, string, int, int]("pod")
	nothing := lens.Set(g, functions.Nothing[int]())
	if nothing.HasNode("pod") {
		t.Fatal("Set(Nothing) should remove the node")
	}
	// Incident edges go with it.
	if len(nothing.OutEdges("svc")) != 0 {
		t.Fatalf("edges to removed node survived: %v", nothing.OutEdges("svc"))
	}
}

func TestColorNodePrismPreview(t *testing.T) {
	g := optGraph()
	asService := optics.ColorNodePrism[string, string, int, int]("svc", "Service")
	if p := asService.Preview(g); !p.IsJust() {
		t.Fatal("Preview svc as Service = Nothing, want Just")
	}
	asPod := optics.ColorNodePrism[string, string, int, int]("svc", "Pod")
	if p := asPod.Preview(g); p.IsJust() {
		t.Fatal("Preview svc as Pod = Just, want Nothing (wrong color)")
	}
	rebuilt := asService.Review(99)
	if n, ok := rebuilt.GetNode("svc"); !ok || n.Color != "Service" || n.Payload != 99 {
		t.Fatalf("Review built %+v, want svc/Service/99", n)
	}
}

func TestAllEdgeWeightsTraversal(t *testing.T) {
	g := optGraph()
	tr := optics.AllEdgeWeights[string, string, int, int]()
	if ws := tr.GetAll(g); !reflect.DeepEqual(ws, []int{7, 5}) {
		// Edges sort lexicographically: (svc,pod,owns,7) < (svc,pod,selects,5)
		t.Fatalf("GetAll weights = %v, want [7 5]", ws)
	}
	doubled := tr.Modify(g, func(w int) int { return w * 2 })
	if ws := tr.GetAll(doubled); !reflect.DeepEqual(ws, []int{14, 10}) {
		t.Fatalf("after Modify weights = %v, want [14 10]", ws)
	}
}

func TestOutEdgeWeightsTraversalScopedToNode(t *testing.T) {
	g := optGraph().AddNode("other", "Pod", 0).Connect("other", "pod", "selects", 1)
	tr := optics.OutEdgeWeights[string, string, int, int]("svc")
	bumped := tr.Modify(g, func(w int) int { return w + 100 })
	// svc's out-edges bumped...
	for _, e := range bumped.OutEdges("svc") {
		if e.Weight < 100 {
			t.Fatalf("svc edge not bumped: %v", e)
		}
	}
	// ...but "other"'s edge untouched.
	oe := bumped.OutEdges("other")
	if len(oe) != 1 || oe[0].Weight != 1 {
		t.Fatalf("other's out-edge changed: %v, want weight 1", oe)
	}
}
