package evaluator

import (
	"fmt"
	"iter"
	"strconv"
	"strings"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// This file wires the generic persistent Graph collection (collections/graph.go,
// Layer 1) and its traversal algorithms (collections/graph_walk.go) through to the
// evaluator's streaming pipeline (Layer 3) as a family of PipelineSteps. Each step
// is the Layer-3 half of a DSL builtin; the Layer-5 string dispatch that names them
// lives in pkg/cli/cli.go.
//
// # The DSL graph instantiation
//
// The host engine's Graph is fully generic — Graph[K, C, N, E] over identity,
// color, node payload, and edge weight. The dynamically-typed DSL pins all four to
// concrete carriers so a single value type flows through the Either[error, any]
// trampoline:
//
//   - K (identity) = string   — node names are strings in the DSL
//   - C (color)    = string   — a color is a sum-type constructor *name*; the
//                               parser already lexes `type EdgeColor = Selects | …`
//                               into tags, and the DSL passes the tag as a string
//   - N (payload)  = any      — a node's associated data (a HAMT after ingest, or
//                               just its id for a bare topological graph)
//   - E (weight)   = any      — numeric by default (int 1), but kept open so richer
//                               weights can flow; ShortestPath reads it via ToFloat64
//
// dslGraph names that single instantiation so the steps below read cleanly.
type dslGraph = *collections.Graph[string, string, any, any]

// emptyDSLGraph is the shared empty value every construction pipeline grows from.
func emptyDSLGraph() dslGraph {
	return collections.NewGraph[string, string, any, any]()
}

// asGraph extracts a dslGraph from a node's Right payload, reporting whether the
// payload actually is one. A nil (never-seeded) graph is normalised to an empty
// graph so downstream steps need not nil-check.
func asGraph(val any) (dslGraph, bool) {
	if g, ok := val.(dslGraph); ok {
		if g == nil {
			return emptyDSLGraph(), true
		}
		return g, true
	}
	return nil, false
}

// parseWeight interprets a DSL weight token as the most specific numeric type it
// parses to (int, then float64), falling back to the raw string. Keeping ints as
// ints means unweighted graphs round-trip cleanly, while floats enable fractional
// trust weights on inferred edges (e.g. 0.5); ShortestPath coerces either via
// ToFloat64.
func parseWeight(tok string) any {
	if i, err := strconv.Atoi(tok); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(tok, 64); err == nil {
		return f
	}
	return tok
}

// formatEdge renders an edge as "from-Color->to", a deterministic, assertable
// projection that preserves the edge's color (unlike a bare neighbour id).
func formatEdge(e functions.Edge[string, string, any]) string {
	return fmt.Sprintf("%s-%s->%s", e.From, e.Color, e.To)
}

// graphMap is the common shape of every construction step: it threads the focused
// graph through a pure transformation f, leaving Left nodes and non-graph payloads
// untouched so a graph step is a no-op on a mis-typed stream rather than a crash.
func graphMap(f func(dslGraph) dslGraph) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) {
						return
					}
					continue
				}
				val, _ := node.Value.UnwrapRight()
				if g, ok := asGraph(val); ok {
					res := any(f(g))
					if !yield(dsltypes.NewASTNode(functions.Right[error, any](res), node.DSLType)) {
						return
					}
					continue
				}
				if !yield(node) {
					return
				}
			}
		}
	}
}

// graphUnfold is the common shape of every projection/walk step: it reads the
// focused graph and emits a *stream* of derived values (ids, edge strings, counts),
// again passing Left nodes and non-graph payloads through unchanged.
func graphUnfold(f func(dslGraph, dsltypes.BitFlag, func(dsltypes.ASTNode[functions.Either[error, any]]) bool) bool) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) {
						return
					}
					continue
				}
				val, _ := node.Value.UnwrapRight()
				if g, ok := asGraph(val); ok {
					if !f(g, node.DSLType, yield) {
						return
					}
					continue
				}
				if !yield(node) {
					return
				}
			}
		}
	}
}

// emitValue is a small helper for graphUnfold bodies: yield one Right value.
func emitValue(flag dsltypes.BitFlag, yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool, v any) bool {
	return yield(dsltypes.NewASTNode(functions.Right[error, any](v), flag))
}

// ---------------------------------------------------------------------------
// Construction steps (graph in → graph out)
// ---------------------------------------------------------------------------

// StepGraphEmpty replaces each incoming seed value with a fresh empty graph,
// the root every construction pipeline grows from (`spawn … |> graph_empty |> …`).
func StepGraphEmpty() PipelineStep[functions.Either[error, any]] {
	return graphMap2(func(_ any) any { return any(emptyDSLGraph()) })
}

// graphMap2 is like graphMap but seeds from *any* payload (not only an existing
// graph), used by graph_empty to turn a bare spawn node into a graph.
func graphMap2(f func(any) any) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) {
						return
					}
					continue
				}
				val, _ := node.Value.UnwrapRight()
				if !yield(dsltypes.NewASTNode(functions.Right[error, any](f(val)), node.DSLType)) {
					return
				}
			}
		}
	}
}

// StepGraphAddNode adds a node with the given color and identity (its payload
// defaults to the identity string; richer payloads arrive via ingest adapters).
func StepGraphAddNode(color, id string) PipelineStep[functions.Either[error, any]] {
	return graphMap(func(g dslGraph) dslGraph { return g.AddNode(id, color, any(id)) })
}

// StepGraphAddEdge connects from→to with the given color and weight.
func StepGraphAddEdge(color string, weight any, from, to string) PipelineStep[functions.Either[error, any]] {
	return graphMap(func(g dslGraph) dslGraph { return g.Connect(from, to, color, weight) })
}

// StepGraphAddInferredEdge connects from→to as an *inferred* edge — one assumed
// rather than read from source data, carried with a (typically lower) trust weight.
func StepGraphAddInferredEdge(color string, weight any, from, to string) PipelineStep[functions.Either[error, any]] {
	return graphMap(func(g dslGraph) dslGraph { return g.AddInferredEdge(from, to, color, weight) })
}

// StepGraphAddFanOut adds one logical 1:n relation as n grouped binary edges
// (from→each of tos), all sharing a color and an auto-derived group id so the whole
// relation can be fetched back with graph_edges_of_group.
func StepGraphAddFanOut(color string, weight any, from string, tos []string) PipelineStep[functions.Either[error, any]] {
	group := color + ":" + from
	return graphMap(func(g dslGraph) dslGraph { return g.AddFanOut(color, weight, group, from, tos) })
}

// StepGraphAddFanIn adds one logical n:1 relation as n grouped binary edges
// (each of froms→to), the dual of StepGraphAddFanOut.
func StepGraphAddFanIn(color string, weight any, to string, froms []string) PipelineStep[functions.Either[error, any]] {
	group := color + ":" + to
	return graphMap(func(g dslGraph) dslGraph { return g.AddFanIn(color, weight, group, froms, to) })
}

// StepGraphRemoveNode deletes a node and all of its incident edges.
func StepGraphRemoveNode(id string) PipelineStep[functions.Either[error, any]] {
	return graphMap(func(g dslGraph) dslGraph { return g.RemoveNode(id) })
}

// StepGraphRemoveEdge removes every from→to edge of the given color (any weight),
// expressed as a Reject so it is weight-agnostic.
func StepGraphRemoveEdge(color, from, to string) PipelineStep[functions.Either[error, any]] {
	return graphMap(func(g dslGraph) dslGraph {
		return g.RejectEdges(func(e functions.Edge[string, string, any]) bool {
			return e.From == from && e.To == to && e.Color == color
		})
	})
}

// ---------------------------------------------------------------------------
// Projection steps (graph in → stream of derived values)
// ---------------------------------------------------------------------------

// StepGraphOrder emits the node count (|V|) as a single int.
func StepGraphOrder() PipelineStep[functions.Either[error, any]] {
	return graphUnfold(func(g dslGraph, flag dsltypes.BitFlag, yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) bool {
		return emitValue(flag, yield, g.Order())
	})
}

// StepGraphSize emits the edge count (|E|) as a single int.
func StepGraphSize() PipelineStep[functions.Either[error, any]] {
	return graphUnfold(func(g dslGraph, flag dsltypes.BitFlag, yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) bool {
		return emitValue(flag, yield, g.Size())
	})
}

// StepGraphNodes unfolds the graph into its node identities, in deterministic order.
func StepGraphNodes() PipelineStep[functions.Either[error, any]] {
	return graphUnfold(func(g dslGraph, flag dsltypes.BitFlag, yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) bool {
		for _, id := range g.NodeIDs() {
			if !emitValue(flag, yield, id) {
				return false
			}
		}
		return true
	})
}

// StepGraphEdges unfolds the graph into "from-Color->to" edge strings.
func StepGraphEdges() PipelineStep[functions.Either[error, any]] {
	return graphUnfold(func(g dslGraph, flag dsltypes.BitFlag, yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) bool {
		for _, e := range g.Edges() {
			if !emitValue(flag, yield, formatEdge(e)) {
				return false
			}
		}
		return true
	})
}

// StepGraphOut unfolds the out-neighbour identities of id (forward adjacency).
func StepGraphOut(id string) PipelineStep[functions.Either[error, any]] {
	return idUnfold(func(g dslGraph) []string { return g.Neighbors(id) })
}

// StepGraphIn unfolds the in-neighbour identities of id (reverse adjacency).
func StepGraphIn(id string) PipelineStep[functions.Either[error, any]] {
	return idUnfold(func(g dslGraph) []string { return g.Predecessors(id) })
}

// StepGraphNeighbors is an alias for StepGraphOut (forward neighbours).
func StepGraphNeighbors(id string) PipelineStep[functions.Either[error, any]] {
	return StepGraphOut(id)
}

// idUnfold is the shared body for the neighbour-style projections.
func idUnfold(pick func(dslGraph) []string) PipelineStep[functions.Either[error, any]] {
	return graphUnfold(func(g dslGraph, flag dsltypes.BitFlag, yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) bool {
		for _, id := range pick(g) {
			if !emitValue(flag, yield, id) {
				return false
			}
		}
		return true
	})
}

// StepGraphNodesOf unfolds the identities of nodes whose color is the given one.
func StepGraphNodesOf(color string) PipelineStep[functions.Either[error, any]] {
	return graphUnfold(func(g dslGraph, flag dsltypes.BitFlag, yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) bool {
		for _, n := range g.NodesByColor(color) {
			if !emitValue(flag, yield, n.ID) {
				return false
			}
		}
		return true
	})
}

// StepGraphEdgesOf unfolds the "from-Color->to" strings of edges of the given color.
func StepGraphEdgesOf(color string) PipelineStep[functions.Either[error, any]] {
	return edgeUnfold(func(g dslGraph) []functions.Edge[string, string, any] { return g.EdgesByColor(color) })
}

// StepGraphEdgesOfGroup unfolds the edges forming one grouped n-ary relation.
func StepGraphEdgesOfGroup(group string) PipelineStep[functions.Either[error, any]] {
	return edgeUnfold(func(g dslGraph) []functions.Edge[string, string, any] { return g.EdgesByGroup(group) })
}

// StepGraphEdgesInferred unfolds only the edges marked inferred (assumed).
func StepGraphEdgesInferred() PipelineStep[functions.Either[error, any]] {
	return edgeUnfold(func(g dslGraph) []functions.Edge[string, string, any] { return g.InferredEdges() })
}

// edgeUnfold is the shared body for the edge-string projections.
func edgeUnfold(pick func(dslGraph) []functions.Edge[string, string, any]) PipelineStep[functions.Either[error, any]] {
	return graphUnfold(func(g dslGraph, flag dsltypes.BitFlag, yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) bool {
		for _, e := range pick(g) {
			if !emitValue(flag, yield, formatEdge(e)) {
				return false
			}
		}
		return true
	})
}

// StepGraphReachable unfolds the identities reachable from id (inclusive), sorted.
func StepGraphReachable(id string) PipelineStep[functions.Either[error, any]] {
	return graphUnfold(func(g dslGraph, flag dsltypes.BitFlag, yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) bool {
		set := g.Reachable(id)
		ids := make([]string, 0, set.Len())
		for k := range set.Seq() {
			ids = append(ids, k)
		}
		sortStrings(ids)
		for _, k := range ids {
			if !emitValue(flag, yield, k) {
				return false
			}
		}
		return true
	})
}

// ---------------------------------------------------------------------------
// Walk & algorithm steps
// ---------------------------------------------------------------------------

// StepGraphBFS unfolds the identities visited in a breadth-first walk from id.
func StepGraphBFS(id string) PipelineStep[functions.Either[error, any]] {
	return walkUnfold(func(g dslGraph) iter.Seq[string] { return g.BFS(id) })
}

// StepGraphDFS unfolds the identities visited in a depth-first walk from id.
func StepGraphDFS(id string) PipelineStep[functions.Either[error, any]] {
	return walkUnfold(func(g dslGraph) iter.Seq[string] { return g.DFS(id) })
}

// StepGraphWalkFrom unfolds a walk from id using the named strategy ("bfs"/"dfs").
func StepGraphWalkFrom(id, strategy string) PipelineStep[functions.Either[error, any]] {
	strat := collections.WalkBFS
	if strings.EqualFold(strategy, "dfs") {
		strat = collections.WalkDFS
	}
	return walkUnfold(func(g dslGraph) iter.Seq[string] { return g.WalkFrom(id, strat) })
}

// walkUnfold is the shared body for the lazy walk projections.
func walkUnfold(walk func(dslGraph) iter.Seq[string]) PipelineStep[functions.Either[error, any]] {
	return graphUnfold(func(g dslGraph, flag dsltypes.BitFlag, yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) bool {
		for id := range walk(g) {
			if !emitValue(flag, yield, id) {
				return false
			}
		}
		return true
	})
}

// StepGraphShortestPath unfolds the node identities along a minimum-weight path
// from→to (inclusive). It emits nothing when no path exists.
func StepGraphShortestPath(from, to string) PipelineStep[functions.Either[error, any]] {
	return graphUnfold(func(g dslGraph, flag dsltypes.BitFlag, yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) bool {
		path, _, ok := g.ShortestPath(from, to)
		if !ok {
			return true
		}
		for _, id := range path {
			if !emitValue(flag, yield, id) {
				return false
			}
		}
		return true
	})
}

// StepGraphTopoSort unfolds a topological ordering of the node identities. If the
// graph contains a cycle (no total order exists) it injects an error instead.
func StepGraphTopoSort() PipelineStep[functions.Either[error, any]] {
	return graphUnfold(func(g dslGraph, flag dsltypes.BitFlag, yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) bool {
		res := g.TopoSort()
		if res.IsLeft() {
			rem, _ := res.UnwrapLeft()
			return yield(InjectError[any](fmt.Errorf("graph has a cycle; %d node(s) unordered", len(rem)), flag))
		}
		order, _ := res.UnwrapRight()
		for _, id := range order {
			if !emitValue(flag, yield, id) {
				return false
			}
		}
		return true
	})
}

// StepGraphSCC unfolds the strongly-connected components, one "a,b,c" string each
// (members sorted within a component; components in the algorithm's order).
func StepGraphSCC() PipelineStep[functions.Either[error, any]] {
	return graphUnfold(func(g dslGraph, flag dsltypes.BitFlag, yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) bool {
		for _, comp := range g.SCC() {
			members := make([]string, len(comp))
			copy(members, comp)
			sortStrings(members)
			if !emitValue(flag, yield, strings.Join(members, ",")) {
				return false
			}
		}
		return true
	})
}

// sortStrings is a tiny insertion sort kept local so the graph steps need no extra
// import; the slices it sorts (neighbourhoods, components) are small.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}
