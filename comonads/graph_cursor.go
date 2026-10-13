package comonads

import (
	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// GraphCursor is a Comonad acting as a cursor over a Graph: a whole graph together
// with a *focused* node identity.
//
// # Comonad, informally
//
// Where a Monad lets you *build up* a computation by injecting a value and
// sequencing dependent steps (Bind: W a → (a → W b) → W b), a Comonad lets you
// *query in context*: it holds a value embedded in a surrounding structure and
// lets you recompute, at every position, a new value derived from that position's
// local neighbourhood. Its two operations mirror the monad's, dualised:
//
//   - Extract (W a → a): read the value at the focus — here, the focused node.
//   - Extend  ((W a → b) → W a → W b): run a "local view" function at *every*
//     position and collect the results back into the same shape — here, recompute
//     every node's payload from a cursor focused on it, yielding a new graph.
//
// This is the natural way to express *propagation* rules over an infrastructure
// graph: "each node's new state is a function of its own state and its
// neighbours'" is literally one Extend. Iterating Extend to a fixpoint gives
// reachability-style compliance checks, trust/score diffusion, and taint
// propagation — all without a hand-rolled traversal.
//
// The TreeZipper in this package is the acyclic single-root analogue; GraphCursor
// is its cyclic, multi-neighbour generalisation. Navigation here is by identity
// (To/Successors/Predecessors) rather than by Up/Down/Left/Right, because a graph
// node has a *set* of neighbours in each direction rather than an ordered spine.
//
// Type parameters mirror collections.Graph: K identity, C color, N node payload,
// E edge weight.
type GraphCursor[K comparable, C comparable, N any, E any] struct {
	Graph *collections.Graph[K, C, N, E]
	Focus K
}

// NewGraphCursor creates a cursor over g focused on the node identity focus. The
// focus need not (yet) be a declared node; Extract reports a zero node until one
// exists.
func NewGraphCursor[K comparable, C comparable, N any, E any](
	g *collections.Graph[K, C, N, E], focus K,
) *GraphCursor[K, C, N, E] {
	return &GraphCursor[K, C, N, E]{Graph: g, Focus: focus}
}

// Extract fulfils the Comonad interface (W[A] → A): it returns the node currently
// under focus. The boolean second return reports whether a node with the focused
// identity is actually declared in the graph.
func (c *GraphCursor[K, C, N, E]) Extract() (functions.Node[K, C, N], bool) {
	return c.Graph.GetNode(c.Focus)
}

// FocusID returns the identity currently under focus.
func (c *GraphCursor[K, C, N, E]) FocusID() K { return c.Focus }

// To moves the focus to another identity in the same graph, returning a new
// cursor. It returns nil when no node with that identity is declared, so callers
// can chain navigation and stop at a dead end (mirroring the nil-on-edge
// convention of TreeZipper).
func (c *GraphCursor[K, C, N, E]) To(id K) *GraphCursor[K, C, N, E] {
	if !c.Graph.HasNode(id) {
		return nil
	}
	return &GraphCursor[K, C, N, E]{Graph: c.Graph, Focus: id}
}

// Successors returns one cursor per distinct out-neighbour of the focus, in the
// graph's deterministic neighbour order — the "down/forward" move of the cursor.
func (c *GraphCursor[K, C, N, E]) Successors() []*GraphCursor[K, C, N, E] {
	return c.cursorsAt(c.Graph.Neighbors(c.Focus))
}

// Predecessors returns one cursor per distinct in-neighbour of the focus — the
// reverse move, cheap thanks to the graph's maintained reverse index.
func (c *GraphCursor[K, C, N, E]) Predecessors() []*GraphCursor[K, C, N, E] {
	return c.cursorsAt(c.Graph.Predecessors(c.Focus))
}

// cursorsAt lifts a slice of identities into cursors over the same graph.
func (c *GraphCursor[K, C, N, E]) cursorsAt(ids []K) []*GraphCursor[K, C, N, E] {
	out := make([]*GraphCursor[K, C, N, E], 0, len(ids))
	for _, id := range ids {
		out = append(out, &GraphCursor[K, C, N, E]{Graph: c.Graph, Focus: id})
	}
	return out
}

// OutEdges returns the edges leaving the focus (with weights and colors), for rules
// that discriminate on edge type or weight rather than just adjacency.
func (c *GraphCursor[K, C, N, E]) OutEdges() []functions.Edge[K, C, E] {
	return c.Graph.OutEdges(c.Focus)
}

// InEdges returns the edges entering the focus.
func (c *GraphCursor[K, C, N, E]) InEdges() []functions.Edge[K, C, E] {
	return c.Graph.InEdges(c.Focus)
}

// Extend is the comonadic cobind restricted to payload-preserving rules
// (W a → a) → W a → W a: it recomputes *every* node's payload by focusing a cursor
// on it and applying f, then returns a cursor on the resulting graph at the same
// focus. Because the payload type is unchanged, Extend can be iterated in place to
// a fixpoint — the usual shape of a propagation/compliance pass.
//
// For a rule that changes the payload type (N → R), use the package-level
// ExtendGraph, which Go's method rules cannot express as a method.
func (c *GraphCursor[K, C, N, E]) Extend(f func(*GraphCursor[K, C, N, E]) N) *GraphCursor[K, C, N, E] {
	ng := ExtendGraph(c.Graph, f)
	return &GraphCursor[K, C, N, E]{Graph: ng, Focus: c.Focus}
}

// ExtendGraph is the full comonadic cobind over a graph:
// (GraphCursor → R) → Graph → Graph-with-payload-R. It evaluates the local-view
// function f at every declared node (with a cursor focused there) and rebuilds the
// graph with each node's payload replaced by f's result, preserving every
// identity, color, and edge unchanged.
//
// It is a free function rather than a method because it changes the node-payload
// type parameter from N to R, which a Go method may not introduce.
func ExtendGraph[K comparable, C comparable, N any, R any, E any](
	g *collections.Graph[K, C, N, E],
	f func(*GraphCursor[K, C, N, E]) R,
) *collections.Graph[K, C, R, E] {
	res := collections.NewGraph[K, C, R, E]()
	for _, n := range g.Nodes() {
		cur := &GraphCursor[K, C, N, E]{Graph: g, Focus: n.ID}
		res = res.AddNode(n.ID, n.Color, f(cur))
	}
	for _, e := range g.Edges() {
		res = res.AddEdge(e)
	}
	return res
}

// DuplicateGraph is the comonadic W a → W (W a): it replaces every node's payload
// with a cursor focused on that node, so the resulting graph's payloads are
// themselves positioned views of the original. Extend f is always equal to Map f
// over DuplicateGraph; it is provided for completeness and for rules expressed as a
// plain Map over the context-carrying graph. (It is named DuplicateGraph rather
// than Duplicate to avoid colliding with the Grid comonad's Duplicate.)
func DuplicateGraph[K comparable, C comparable, N any, E any](
	g *collections.Graph[K, C, N, E],
) *collections.Graph[K, C, *GraphCursor[K, C, N, E], E] {
	return ExtendGraph(g, func(c *GraphCursor[K, C, N, E]) *GraphCursor[K, C, N, E] { return c })
}
