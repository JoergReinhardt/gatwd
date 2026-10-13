package transducers_test

import (
	"reflect"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/transducers"
)

// walkGraph: ingress → svc → pod, with a cycle pod → ingress to prove cycle-safety.
func walkGraph() *collections.Graph[string, string, string, int] {
	g := collections.NewGraph[string, string, string, int]()
	g = g.AddNode("ingress", "Ingress", "ing").
		AddNode("svc", "Service", "my-svc").
		AddNode("pod", "Pod", "my-pod")
	g = g.Connect("ingress", "svc", "routes", 1).
		Connect("svc", "pod", "selects", 1).
		Connect("pod", "ingress", "back", 1) // cycle
	return g
}

func TestWalkNodesCycleSafeBFSOrder(t *testing.T) {
	g := walkGraph()
	nodes := transducers.WalkNodes(g, "ingress", collections.WalkBFS)
	var order []string
	for _, n := range nodes {
		order = append(order, n.ID)
	}
	if !reflect.DeepEqual(order, []string{"ingress", "svc", "pod"}) {
		t.Fatalf("BFS walk order = %v, want [ingress svc pod]", order)
	}
}

// TestTransduceWalkFusesMapFilter composes a filter (keep Service nodes) and a map
// (project the payload) into one pass over the cycle-safe walk.
func TestTransduceWalkFusesMapFilter(t *testing.T) {
	g := walkGraph()
	keepServices := transducers.Filter[functions.Node[string, string, string], []string](
		func(n functions.Node[string, string, string]) bool { return n.Color == "Service" },
	)
	projectName := transducers.Map[functions.Node[string, string, string], string, []string](
		func(n functions.Node[string, string, string]) string { return n.Payload },
	)
	// Compose: Filter then Map. A Transducer composes outer(inner).
	pipeline := func(next transducers.Reducer[string, []string]) transducers.Reducer[functions.Node[string, string, string], []string] {
		return keepServices(projectName(next))
	}

	got := transducers.TransduceWalk(g, "ingress", collections.WalkBFS, pipeline)
	if !reflect.DeepEqual(got, []string{"my-svc"}) {
		t.Fatalf("TransduceWalk = %v, want [my-svc]", got)
	}
}
