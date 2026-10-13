package optics

import (
	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// This file exposes the three optics of a Graph (collections/graph.go) so the DSL
// (and Go callers) can perform focused, immutable rewrites of a graph without
// hand-threading the persistent HAMT adjacency:
//
//   - NodePayloadLens  — a Lens onto one node's payload (a field that may be absent).
//   - ColorNodePrism   — a Prism onto a node that exists AND has a given color.
//   - AllEdgeWeights / OutEdgeWeights — Traversals onto many edge weights at once.
//
// Each returns a new graph via the collection's own immutable operations, so the
// structural sharing of the underlying HAMTs is preserved.

// NodePayloadLens focuses on the payload of the node identified by id.
//
// Because a node may be absent, the focused value is a Maybe[N]: Get returns
// Just(payload) when the node exists and Nothing otherwise. Set with Just updates
// the payload while preserving the node's color (an existing node is required to
// know its color, so setting a payload on an absent node is a no-op); Set with
// Nothing removes the node entirely (and, as in Graph.RemoveNode, its incident
// edges).
func NodePayloadLens[K comparable, C comparable, N any, E any](id K) Lens[*collections.Graph[K, C, N, E], functions.Maybe[N]] {
	return Lens[*collections.Graph[K, C, N, E], functions.Maybe[N]]{
		Get: func(g *collections.Graph[K, C, N, E]) functions.Maybe[N] {
			if g == nil {
				return functions.Nothing[N]()
			}
			if n, ok := g.GetNode(id); ok {
				return functions.Just(n.Payload)
			}
			return functions.Nothing[N]()
		},
		Set: func(g *collections.Graph[K, C, N, E], mv functions.Maybe[N]) *collections.Graph[K, C, N, E] {
			if g == nil {
				g = collections.NewGraph[K, C, N, E]()
			}
			if mv.IsNothing() {
				return g.RemoveNode(id)
			}
			v, _ := mv.Unwrap()
			if n, ok := g.GetNode(id); ok {
				return g.AddNode(id, n.Color, v)
			}
			return g // cannot invent a color for an undeclared node
		},
	}
}

// ColorNodePrism focuses on the payload of the node id *only when* that node exists
// and carries the given color — the optic for "this node, if it is a Pod".
//
// Preview returns Just(payload) on a color match and Nothing otherwise. Review, as
// a Prism must, reconstructs a minimal structure from the branch: a fresh graph
// containing exactly that one colored node. (Use NodePayloadLens to edit a node's
// payload *within* an existing graph.)
func ColorNodePrism[K comparable, C comparable, N any, E any](id K, color C) Prism[*collections.Graph[K, C, N, E], N] {
	return Prism[*collections.Graph[K, C, N, E], N]{
		Preview: func(g *collections.Graph[K, C, N, E]) functions.Maybe[N] {
			if g == nil {
				return functions.Nothing[N]()
			}
			if n, ok := g.GetNode(id); ok && n.Color == color {
				return functions.Just(n.Payload)
			}
			return functions.Nothing[N]()
		},
		Review: func(payload N) *collections.Graph[K, C, N, E] {
			return collections.NewGraph[K, C, N, E]().AddNode(id, color, payload)
		},
	}
}

// AllEdgeWeights is a Traversal onto every edge weight in the graph. Modify is the
// Functor map over edge weights (Graph.MapEdges); GetAll lists the weights in the
// graph's deterministic edge order.
func AllEdgeWeights[K comparable, C comparable, N any, E any]() Traversal[*collections.Graph[K, C, N, E], E] {
	return Traversal[*collections.Graph[K, C, N, E], E]{
		GetAll: func(g *collections.Graph[K, C, N, E]) []E {
			if g == nil {
				return nil
			}
			var out []E
			for _, e := range g.Edges() {
				out = append(out, e.Weight)
			}
			return out
		},
		Modify: func(g *collections.Graph[K, C, N, E], f func(E) E) *collections.Graph[K, C, N, E] {
			if g == nil {
				return collections.NewGraph[K, C, N, E]()
			}
			return g.MapEdges(f)
		},
	}
}

// OutEdgeWeights is a Traversal onto the weights of the edges *leaving* a single
// node — the focused bulk edit "re-cost everything this node points at". Modify
// rewrites each matching edge in place (remove-then-add the weight-adjusted edge),
// leaving all other edges and every node untouched.
func OutEdgeWeights[K comparable, C comparable, N any, E any](id K) Traversal[*collections.Graph[K, C, N, E], E] {
	return Traversal[*collections.Graph[K, C, N, E], E]{
		GetAll: func(g *collections.Graph[K, C, N, E]) []E {
			if g == nil {
				return nil
			}
			var out []E
			for _, e := range g.OutEdges(id) {
				out = append(out, e.Weight)
			}
			return out
		},
		Modify: func(g *collections.Graph[K, C, N, E], f func(E) E) *collections.Graph[K, C, N, E] {
			if g == nil {
				return collections.NewGraph[K, C, N, E]()
			}
			res := g
			for _, e := range g.OutEdges(id) {
				ne := e
				ne.Weight = f(e.Weight)
				res = res.RemoveEdge(e).AddEdge(ne)
			}
			return res
		},
	}
}
