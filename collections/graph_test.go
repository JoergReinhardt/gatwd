package collections

import (
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

// buildSample constructs a small colored, weighted, cyclic graph used across the
// structural tests:
//
//	svc  --selects(1)-->  pod1
//	svc  --selects(1)-->  pod2
//	pod1 --calls(2)----->  svc     (back-edge: makes the graph cyclic)
//
// Nodes are colored "service"/"pod"; the two selects edges share group "g1".
func buildSample() *Graph[string, string, int, int] {
	g := NewGraph[string, string, int, int]()
	g = g.AddNode("svc", "service", 0)
	g = g.AddNode("pod1", "pod", 1)
	g = g.AddNode("pod2", "pod", 2)
	g = g.AddFanOut("selects", 1, "g1", "svc", []string{"pod1", "pod2"})
	g = g.AddEdge(functions.NewEdge("pod1", "svc", "calls", 2))
	return g
}

func TestGraphConstructionAndProjection(t *testing.T) {
	g := buildSample()

	if g.Order() != 3 {
		t.Errorf("Order = %d, want 3", g.Order())
	}
	if g.Size() != 3 {
		t.Errorf("Size = %d, want 3", g.Size())
	}
	if !g.HasNode("svc") || g.HasNode("missing") {
		t.Errorf("HasNode wrong")
	}
	n, ok := g.GetNode("pod1")
	if !ok || n.Color != "pod" || n.Payload != 1 {
		t.Errorf("GetNode(pod1) = %+v, ok=%v", n, ok)
	}

	// Successors of svc are its two selected pods (deterministic, deduped).
	nb := g.Neighbors("svc")
	if len(nb) != 2 || nb[0] != "pod1" || nb[1] != "pod2" {
		t.Errorf("Neighbors(svc) = %v, want [pod1 pod2]", nb)
	}
	// svc has one predecessor (pod1, via the calls back-edge).
	pred := g.Predecessors("svc")
	if len(pred) != 1 || pred[0] != "pod1" {
		t.Errorf("Predecessors(svc) = %v, want [pod1]", pred)
	}

	// By-color / by-group / inferred projections.
	if pods := g.NodesByColor("pod"); len(pods) != 2 {
		t.Errorf("NodesByColor(pod) = %d, want 2", len(pods))
	}
	if sel := g.EdgesByColor("selects"); len(sel) != 2 {
		t.Errorf("EdgesByColor(selects) = %d, want 2", len(sel))
	}
	if grp := g.EdgesByGroup("g1"); len(grp) != 2 {
		t.Errorf("EdgesByGroup(g1) = %d, want 2", len(grp))
	}
	if inf := g.InferredEdges(); len(inf) != 0 {
		t.Errorf("InferredEdges = %d, want 0", len(inf))
	}
}

func TestGraphImmutabilityAndRemoval(t *testing.T) {
	g := buildSample()
	g2 := g.RemoveNode("pod2")

	// Original is untouched (persistent/immutable).
	if g.Order() != 3 || g.Size() != 3 {
		t.Errorf("original mutated: Order=%d Size=%d", g.Order(), g.Size())
	}
	// pod2 and its single incident selects edge are gone.
	if g2.HasNode("pod2") {
		t.Errorf("pod2 still present after RemoveNode")
	}
	if g2.Order() != 2 || g2.Size() != 2 {
		t.Errorf("after RemoveNode: Order=%d Size=%d, want 2,2", g2.Order(), g2.Size())
	}

	// RemoveEdge drops exactly the matching edge.
	g3 := g.RemoveEdge(functions.NewEdge("pod1", "svc", "calls", 2))
	if g3.Size() != 2 {
		t.Errorf("after RemoveEdge: Size=%d, want 2", g3.Size())
	}
	if len(g3.Predecessors("svc")) != 0 {
		t.Errorf("calls back-edge not removed from reverse index")
	}
}

// Degree counts EDGES, not distinct neighbors: svc has two outgoing selects
// edges (to two different pods, so OutDegree and len(Neighbors) agree here) and
// one incoming calls edge from pod1 (the cycle back-edge) — svc's total degree
// is 2+1=3, matching neither in isolation.
func TestGraphDegree(t *testing.T) {
	g := buildSample()
	if got := g.OutDegree("svc"); got != 2 {
		t.Errorf("OutDegree(svc) = %d, want 2", got)
	}
	if got := g.InDegree("svc"); got != 1 {
		t.Errorf("InDegree(svc) = %d, want 1", got)
	}
	if got := g.Degree("svc"); got != 3 {
		t.Errorf("Degree(svc) = %d, want 3", got)
	}
	// pod2 only ever receives: in-degree 1 (svc's selects), out-degree 0.
	if got := g.OutDegree("pod2"); got != 0 {
		t.Errorf("OutDegree(pod2) = %d, want 0", got)
	}
	if got := g.InDegree("pod2"); got != 1 {
		t.Errorf("InDegree(pod2) = %d, want 1", got)
	}
	// A node with no incident edges at all has degree 0 — negative control so
	// this test can fail if Degree ever drifts to counting declared nodes
	// instead of incident edges.
	g2 := g.AddNode("lonely", "pod", 9)
	if got := g2.Degree("lonely"); got != 0 {
		t.Errorf("Degree(lonely) = %d, want 0", got)
	}
}

func TestGraphInferredEdge(t *testing.T) {
	g := NewGraph[string, string, float64, float64]()
	g = g.AddNode("pod", "pod", 0)
	g = g.AddNode("svc", "service", 0)
	g = g.AddInferredEdge("pod", "svc", "consumes", 0.5)

	inf := g.InferredEdges()
	if len(inf) != 1 || !inf[0].Inferred || inf[0].Weight != 0.5 {
		t.Errorf("InferredEdges = %+v, want one inferred consumes edge weight 0.5", inf)
	}
}

func TestGraphFunctor(t *testing.T) {
	g := buildSample()

	// MapNodes doubles payloads; topology untouched.
	mn := g.MapNodes(func(p int) int { return p * 10 })
	if n, _ := mn.GetNode("pod2"); n.Payload != 20 {
		t.Errorf("MapNodes pod2 payload = %d, want 20", n.Payload)
	}
	if mn.Size() != g.Size() {
		t.Errorf("MapNodes changed edge count")
	}

	// MapEdges increments weights; node payloads untouched.
	me := g.MapEdges(func(w int) int { return w + 100 })
	for _, e := range me.OutEdges("svc") {
		if e.Weight != 101 {
			t.Errorf("MapEdges weight = %d, want 101", e.Weight)
		}
	}
	if n, _ := me.GetNode("pod1"); n.Payload != 1 {
		t.Errorf("MapEdges altered node payload")
	}

	// Bimap touches both variances.
	bm := g.Bimap(func(p int) int { return p + 1 }, func(w int) int { return w * 2 })
	if n, _ := bm.GetNode("pod1"); n.Payload != 2 {
		t.Errorf("Bimap node = %d, want 2", n.Payload)
	}
}

func TestGraphFoldable(t *testing.T) {
	g := buildSample()
	sumPayload := g.FoldNodes(0, func(acc any, n functions.Node[string, string, int]) any {
		return acc.(int) + n.Payload
	})
	if sumPayload.(int) != 3 { // 0 + 1 + 2
		t.Errorf("FoldNodes sum = %v, want 3", sumPayload)
	}
	sumWeight := g.FoldEdges(0, func(acc any, e functions.Edge[string, string, int]) any {
		return acc.(int) + e.Weight
	})
	if sumWeight.(int) != 4 { // 1 + 1 + 2
		t.Errorf("FoldEdges sum = %v, want 4", sumWeight)
	}
}

func TestGraphFilterable(t *testing.T) {
	g := buildSample()

	// FilterNodes to just the pods -> drops svc and every edge touching svc,
	// leaving an edgeless two-node induced subgraph.
	pods := g.FilterNodes(func(n functions.Node[string, string, int]) bool {
		return n.Color == "pod"
	})
	if pods.Order() != 2 || pods.Size() != 0 {
		t.Errorf("FilterNodes(pod): Order=%d Size=%d, want 2,0", pods.Order(), pods.Size())
	}

	// FilterEdges keeps all nodes but only the selects edges.
	sel := g.FilterEdges(func(e functions.Edge[string, string, int]) bool {
		return e.Color == "selects"
	})
	if sel.Order() != 3 || sel.Size() != 2 {
		t.Errorf("FilterEdges(selects): Order=%d Size=%d, want 3,2", sel.Order(), sel.Size())
	}

	// Reject is the complement.
	noCalls := g.RejectEdges(func(e functions.Edge[string, string, int]) bool {
		return e.Color == "calls"
	})
	if noCalls.Size() != 2 {
		t.Errorf("RejectEdges(calls): Size=%d, want 2", noCalls.Size())
	}
}

func TestGraphMonoid(t *testing.T) {
	a := NewGraph[string, string, int, int]().
		AddNode("x", "n", 1).AddNode("y", "n", 2).Connect("x", "y", "e", 1)
	b := NewGraph[string, string, int, int]().
		AddNode("y", "n", 2).AddNode("z", "n", 3).Connect("y", "z", "e", 1)

	u := a.Union(b)
	if u.Order() != 3 || u.Size() != 2 {
		t.Errorf("Union: Order=%d Size=%d, want 3,2", u.Order(), u.Size())
	}

	// Union de-duplicates identical edges.
	dup := a.Union(a)
	if dup.Size() != a.Size() {
		t.Errorf("Union(a,a) duplicated edges: Size=%d, want %d", dup.Size(), a.Size())
	}

	inter := a.Intersect(b)
	if inter.Order() != 1 || inter.Size() != 0 {
		t.Errorf("Intersect: Order=%d Size=%d, want 1,0 (shared node y, no shared edge)", inter.Order(), inter.Size())
	}

	diff := u.Diff(b)
	if diff.HasNode("z") {
		t.Errorf("Diff should have removed node z")
	}
}

func TestGraphWalksCycleSafe(t *testing.T) {
	// A 3-cycle a->b->c->a plus a spur c->d.
	g := NewGraph[string, string, int, int]().
		AddNode("a", "n", 0).AddNode("b", "n", 0).
		AddNode("c", "n", 0).AddNode("d", "n", 0).
		Connect("a", "b", "e", 1).Connect("b", "c", "e", 1).
		Connect("c", "a", "e", 1).Connect("c", "d", "e", 1)

	var bfs []string
	for k := range g.BFS("a") {
		bfs = append(bfs, k)
	}
	// BFS from a: a, then b, then c, then d — terminates despite the cycle.
	want := []string{"a", "b", "c", "d"}
	if len(bfs) != 4 {
		t.Fatalf("BFS visited %v, want 4 distinct nodes", bfs)
	}
	for i := range want {
		if bfs[i] != want[i] {
			t.Errorf("BFS order = %v, want %v", bfs, want)
			break
		}
	}

	// DFS also terminates and visits every reachable node exactly once.
	seen := map[string]int{}
	for k := range g.DFS("a") {
		seen[k]++
	}
	if len(seen) != 4 {
		t.Errorf("DFS distinct = %d, want 4", len(seen))
	}
	for k, c := range seen {
		if c != 1 {
			t.Errorf("DFS visited %s %d times, want 1", k, c)
		}
	}

	reach := g.Reachable("a")
	if reach.Len() != 4 {
		t.Errorf("Reachable(a) = %d, want 4", reach.Len())
	}
}

func TestGraphEqualityOrderIndependent(t *testing.T) {
	a := NewGraph[string, string, int, int]().
		AddNode("x", "n", 1).AddNode("y", "n", 2).
		Connect("x", "y", "e", 1).Connect("y", "x", "e", 1)
	// Same graph, edges added in the opposite order.
	b := NewGraph[string, string, int, int]().
		AddNode("y", "n", 2).AddNode("x", "n", 1).
		Connect("y", "x", "e", 1).Connect("x", "y", "e", 1)

	if !a.Equal(b) {
		t.Errorf("Equal should be insertion-order independent")
	}

	// A differing payload breaks equality.
	c := b.AddNode("x", "n", 999)
	if a.Equal(c) {
		t.Errorf("Equal must detect payload difference")
	}
}

// TestGraphOperatorDispatch asserts the ++ and == operators reach the Graph via
// the untyped operator layer (functions.EvalBinaryOp / StructuralEqual), proving
// the matrix row is wired, not just the Go methods.
func TestGraphOperatorDispatch(t *testing.T) {
	a := NewGraph[string, string, int, int]().AddNode("x", "n", 1).Connect("x", "x", "e", 1)
	b := NewGraph[string, string, int, int]().AddNode("y", "n", 2)

	// ++  -> Union
	res, err := functions.EvalBinaryOp(functions.OpConcat, a, b)
	if err != nil {
		t.Fatalf("EvalBinaryOp(++) error: %v", err)
	}
	ug, ok := res.(*Graph[string, string, int, int])
	if !ok {
		t.Fatalf("++ returned %T, want *Graph", res)
	}
	if ug.Order() != 2 || ug.Size() != 1 {
		t.Errorf("a ++ b: Order=%d Size=%d, want 2,1", ug.Order(), ug.Size())
	}

	// ==  -> StructuralEqual
	eq, err := functions.EvalBinaryOp(functions.OpEq, a, a.Union(NewGraph[string, string, int, int]()))
	if err != nil {
		t.Fatalf("EvalBinaryOp(==) error: %v", err)
	}
	if eq != true {
		t.Errorf("a == a(copy) = %v, want true", eq)
	}
	neq, _ := functions.EvalBinaryOp(functions.OpNeq, a, b)
	if neq != true {
		t.Errorf("a != b = %v, want true", neq)
	}
}

// TestGraphDotNodesDrawsFanInOnce is the property a generic struct/slice walk gets wrong: "svc"
// is the target of two edges (from pod1 and pod2's... no, here svc fans OUT to both pods, so the
// shared node is svc itself, reached once as a node but twice as an edge endpoint) — DotNodes
// must still report it exactly once, with DotEdges carrying both connections separately.
func TestGraphDotNodesDrawsFanInOnce(t *testing.T) {
	g := buildSample()

	nodes := g.DotNodes()
	if len(nodes) != g.Order() {
		t.Fatalf("DotNodes() returned %d nodes, want %d (one per node, regardless of edge count)", len(nodes), g.Order())
	}
	seen := map[string]int{}
	for _, n := range nodes {
		if n[0] != n[1] {
			t.Errorf("DotNodes() entry %v: id and label must both be the node's identity", n)
		}
		seen[n[0]]++
	}
	for id, count := range seen {
		if count != 1 {
			t.Errorf("node %q appears %d times in DotNodes(), want exactly once", id, count)
		}
	}

	edges := g.DotEdges()
	if len(edges) != g.Size() {
		t.Fatalf("DotEdges() returned %d edges, want %d", len(edges), g.Size())
	}
	want := map[[3]string]bool{
		{"svc", "pod1", "selects"}: true,
		{"svc", "pod2", "selects"}: true,
		{"pod1", "svc", "calls"}:   true,
	}
	for _, e := range edges {
		if !want[e] {
			t.Errorf("unexpected DotEdges() entry %v", e)
		}
		delete(want, e)
	}
	if len(want) != 0 {
		t.Errorf("DotEdges() is missing %v", want)
	}
}
