package collections

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

// buildWalkGraph constructs a small weighted digraph used across the walk/algorithm
// tests:
//
//	a →1→ b →1→ d
//	a →4→ c →1→ d
//	(plus, where noted, a back edge forming a cycle)
//
// Node payloads are ints so FoldWalk accumulation is observable.
func buildWalkGraph() *Graph[string, string, int, int] {
	g := NewGraph[string, string, int, int]()
	g = g.AddNode("a", "n", 1).AddNode("b", "n", 2).AddNode("c", "n", 3).AddNode("d", "n", 4)
	g = g.Connect("a", "b", "e", 1)
	g = g.Connect("b", "d", "e", 1)
	g = g.Connect("a", "c", "e", 4)
	g = g.Connect("c", "d", "e", 1)
	return g
}

func TestFoldWalkVisitsEachNodeOnceOnCycle(t *testing.T) {
	g := buildWalkGraph().Connect("d", "a", "e", 1) // introduce a cycle d→a
	// Sum every reachable payload; the cycle must not cause infinite accumulation.
	sum := g.FoldWalk("a", WalkBFS, 0, func(acc any, n functions.Node[string, string, int]) any {
		return acc.(int) + n.Payload
	})
	if sum != 1+2+3+4 {
		t.Fatalf("FoldWalk sum = %v, want %d", sum, 1+2+3+4)
	}
}

func TestShortestPathPrefersCheaperRoute(t *testing.T) {
	g := buildWalkGraph()
	path, cost, ok := g.ShortestPath("a", "d")
	if !ok {
		t.Fatal("ShortestPath a→d: ok = false, want true")
	}
	wantPath := []string{"a", "b", "d"} // 1+1=2 beats a→c→d = 4+1=5
	if !reflect.DeepEqual(path, wantPath) {
		t.Fatalf("path = %v, want %v", path, wantPath)
	}
	if cost != 2 {
		t.Fatalf("cost = %v, want 2", cost)
	}
}

func TestShortestPathSameNodeIsZero(t *testing.T) {
	g := buildWalkGraph()
	path, cost, ok := g.ShortestPath("a", "a")
	if !ok || cost != 0 || !reflect.DeepEqual(path, []string{"a"}) {
		t.Fatalf("ShortestPath a→a = (%v, %v, %v), want ([a], 0, true)", path, cost, ok)
	}
}

func TestShortestPathUnreachable(t *testing.T) {
	g := buildWalkGraph().AddNode("z", "n", 9) // z has no incoming edges
	if _, _, ok := g.ShortestPath("a", "z"); ok {
		t.Fatal("ShortestPath a→z: ok = true, want false (unreachable)")
	}
}

func TestTopoSortAcyclicYieldsOrder(t *testing.T) {
	g := buildWalkGraph()
	res := g.TopoSort()
	if !res.IsRight() {
		t.Fatalf("TopoSort on DAG returned Left, want Right")
	}
	order, _ := res.UnwrapRight()
	pos := make(map[string]int, len(order))
	for i, id := range order {
		pos[id] = i
	}
	// Every edge must point forwards in the ordering.
	for _, e := range g.Edges() {
		if pos[e.From] >= pos[e.To] {
			t.Fatalf("edge %s→%s violates order %v", e.From, e.To, order)
		}
	}
}

func TestTopoSortCycleYieldsRemainder(t *testing.T) {
	// b↔c mutual cycle; a and d are acyclic around it.
	g := NewGraph[string, string, int, int]()
	g = g.AddNode("a", "n", 0).AddNode("b", "n", 0).AddNode("c", "n", 0).AddNode("d", "n", 0)
	g = g.Connect("a", "b", "e", 1).Connect("b", "c", "e", 1).Connect("c", "b", "e", 1).Connect("c", "d", "e", 1)
	res := g.TopoSort()
	if !res.IsLeft() {
		t.Fatalf("TopoSort on cyclic graph returned Right, want Left")
	}
	rem, _ := res.UnwrapLeft()
	// b↔c are the cycle; d is unorderable because it is downstream of c, which
	// never becomes a source. Kahn's remainder is everything left unordered.
	if !reflect.DeepEqual(rem, []string{"b", "c", "d"}) {
		t.Fatalf("cyclic remainder = %v, want [b c d]", rem)
	}
}

// reachableToStrings is a small test helper rendering a *Set[string] as a
// sorted slice, so ReachableTo's set-valued result compares with reflect.DeepEqual.
func reachableToStrings(s *Set[string]) []string {
	out := s.Elements()
	sort.Strings(out)
	return out
}

// ReachableTo is Reachable's mirror image: everyone who can reach d (forward),
// versus only a and b who can reach b.
func TestReachableToMirrorsReachable(t *testing.T) {
	g := buildWalkGraph()
	if got, want := reachableToStrings(g.ReachableTo("d")), []string{"a", "b", "c", "d"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ReachableTo(d) = %v, want %v (everyone depends on d)", got, want)
	}
	if got, want := reachableToStrings(g.ReachableTo("b")), []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ReachableTo(b) = %v, want %v", got, want)
	}
	// Negative control: Reachable(d) (forward) must NOT agree with ReachableTo(d)
	// (backward) on this asymmetric graph — if it ever did, the two would have
	// collapsed into the same (wrong) direction.
	if got := reachableToStrings(g.Reachable("d")); reflect.DeepEqual(got, []string{"a", "b", "c", "d"}) {
		t.Fatalf("control failed: forward Reachable(d) = %v should differ from ReachableTo(d)", got)
	}
}

// A cycle must not make ReachableTo loop forever, exactly as Reachable is
// already pinned against TestFoldWalkVisitsEachNodeOnceOnCycle.
func TestReachableToTerminatesOnCycle(t *testing.T) {
	g := buildWalkGraph().Connect("d", "a", "e", 1) // d→a closes a cycle
	got := reachableToStrings(g.ReachableTo("a"))
	want := []string{"a", "b", "c", "d"} // the whole cycle can now reach a
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReachableTo(a) on cyclic graph = %v, want %v", got, want)
	}
}

// renderTree flattens a Tree into "id(child,child)" so structure (including
// duplication) is a single comparable string in tests.
func renderTree(t *Tree[string]) string {
	kids := t.Children().Elements()
	if len(kids) == 0 {
		return t.Value()
	}
	parts := make([]string, len(kids))
	for i, k := range kids {
		parts[i] = renderTree(k.(*Tree[string]))
	}
	return t.Value() + "(" + strings.Join(parts, ",") + ")"
}

// Unroll must DUPLICATE d (reached via both b and c) rather than collapsing it
// to the single node the graph itself has — the whole point of unrolling.
func TestUnrollDuplicatesSharedDescendant(t *testing.T) {
	g := buildWalkGraph() // a→b→d, a→c→d (diamond)
	tree := g.Unroll("a")
	if got, want := renderTree(tree), "a(b(d),c(d))"; got != want {
		t.Fatalf("Unroll(a) = %q, want %q", got, want)
	}
	// d appears twice in the tree though the graph has only 4 nodes — that
	// extra occurrence IS the duplication; a negative control on this count
	// would catch an accidental collapse back to a DAG-shaped tree (size 4).
	if got, want := tree.Size(), 5; got != want {
		t.Fatalf("Unroll(a).Size() = %d, want %d (b, c and TWO occurrences of d, plus a)", got, want)
	}
}

// A cycle must terminate Unroll (a tree cannot be infinite): the node that
// would re-enter its own ancestor path becomes a leaf instead of recursing.
func TestUnrollTerminatesOnCycleByDuplicatingTheCyclingNode(t *testing.T) {
	g := buildWalkGraph().Connect("d", "a", "e", 1) // d→a closes a cycle through every path
	tree := g.Unroll("a")
	if got, want := renderTree(tree), "a(b(d(a)),c(d(a)))"; got != want {
		t.Fatalf("Unroll(a) on cyclic graph = %q, want %q", got, want)
	}
}

func TestSCCPartitionsVertices(t *testing.T) {
	// One non-trivial component {b,c}, two singletons {a},{d}.
	g := NewGraph[string, string, int, int]()
	g = g.AddNode("a", "n", 0).AddNode("b", "n", 0).AddNode("c", "n", 0).AddNode("d", "n", 0)
	g = g.Connect("a", "b", "e", 1).Connect("b", "c", "e", 1).Connect("c", "b", "e", 1).Connect("c", "d", "e", 1)
	comps := g.SCC()
	want := [][]string{{"a"}, {"b", "c"}, {"d"}}
	if !reflect.DeepEqual(comps, want) {
		t.Fatalf("SCC = %v, want %v", comps, want)
	}
}
