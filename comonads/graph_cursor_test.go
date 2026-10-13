package comonads

import (
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
)

// taintGraph: a → b → c, plus an isolated node x. Only a starts "tainted".
func taintGraph() *collections.Graph[string, string, bool, int] {
	g := collections.NewGraph[string, string, bool, int]()
	g = g.AddNode("a", "n", true).AddNode("b", "n", false).AddNode("c", "n", false).AddNode("x", "n", false)
	g = g.Connect("a", "b", "e", 1).Connect("b", "c", "e", 1)
	return g
}

func TestGraphCursorExtractAndNavigate(t *testing.T) {
	c := NewGraphCursor(taintGraph(), "a")
	n, ok := c.Extract()
	if !ok || n.ID != "a" || n.Payload != true {
		t.Fatalf("Extract at a = (%+v, %v), want node a payload true", n, ok)
	}
	succ := c.Successors()
	if len(succ) != 1 || succ[0].FocusID() != "b" {
		t.Fatalf("Successors(a) = %v, want [b]", ids(succ))
	}
	if to := c.To("b"); to == nil || to.FocusID() != "b" {
		t.Fatalf("To(b) failed")
	}
	if dead := c.To("nope"); dead != nil {
		t.Fatalf("To(nonexistent) = %v, want nil", dead)
	}
	pred := NewGraphCursor(taintGraph(), "c").Predecessors()
	if len(pred) != 1 || pred[0].FocusID() != "b" {
		t.Fatalf("Predecessors(c) = %v, want [b]", ids(pred))
	}
}

func ids[K comparable, C comparable, N any, E any](cs []*GraphCursor[K, C, N, E]) []K {
	out := make([]K, len(cs))
	for i, c := range cs {
		out[i] = c.FocusID()
	}
	return out
}

// TestGraphCursorExtendPropagates exercises the comonad: one Extend is one hop of
// taint propagation (a node is tainted if it or any predecessor is tainted);
// iterating to a fixpoint taints the whole a→b→c chain but never the isolated x.
func TestGraphCursorExtendPropagates(t *testing.T) {
	spread := func(c *GraphCursor[string, string, bool, int]) bool {
		if n, ok := c.Extract(); ok && n.Payload {
			return true
		}
		for _, p := range c.Predecessors() {
			if n, ok := p.Extract(); ok && n.Payload {
				return true
			}
		}
		return false
	}

	cur := NewGraphCursor(taintGraph(), "a")
	for i := 0; i < 3; i++ { // iterate to fixpoint
		cur = cur.Extend(spread)
	}

	for _, id := range []string{"a", "b", "c"} {
		n, _ := cur.Graph.GetNode(id)
		if !n.Payload {
			t.Fatalf("node %s should be tainted after propagation", id)
		}
	}
	x, _ := cur.Graph.GetNode("x")
	if x.Payload {
		t.Fatalf("isolated node x should never be tainted")
	}
}

// TestExtendGraphChangesPayloadType verifies the full comonadic cobind can change
// the payload type N→R (here bool→int: each node's out-degree).
func TestExtendGraphChangesPayloadType(t *testing.T) {
	g := taintGraph()
	degrees := ExtendGraph(g, func(c *GraphCursor[string, string, bool, int]) int {
		return len(c.Successors())
	})
	want := map[string]int{"a": 1, "b": 1, "c": 0, "x": 0}
	for id, w := range want {
		n, ok := degrees.GetNode(id)
		if !ok || n.Payload != w {
			t.Fatalf("out-degree of %s = (%v, %v), want %d", id, n.Payload, ok, w)
		}
	}
	// Topology must be preserved by the cobind.
	if degrees.Size() != g.Size() || degrees.Order() != g.Order() {
		t.Fatalf("ExtendGraph changed topology: order/size %d/%d, want %d/%d",
			degrees.Order(), degrees.Size(), g.Order(), g.Size())
	}
}

func TestDuplicateGraphRefocusesEachNode(t *testing.T) {
	dup := DuplicateGraph(taintGraph())
	for _, id := range []string{"a", "b", "c", "x"} {
		n, ok := dup.GetNode(id)
		if !ok || n.Payload.FocusID() != id {
			t.Fatalf("DuplicateGraph node %s payload focus = %v, want %s", id, n.Payload, id)
		}
	}
}
