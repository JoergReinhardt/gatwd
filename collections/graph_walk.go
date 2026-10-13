package collections

import (
	"math"
	"sort"

	"github.com/joergreinhardt/gatwd/functions"
)

// This file adds the traversal *algorithms* over the persistent Graph collection
// (collections/graph.go): a traversal-fold, weighted shortest paths, topological
// ordering, and strongly-connected-component decomposition. They are kept beside
// the collection (Layer 1) because they are pure functions of a Graph value and
// introduce no dependency on the higher Layer-2 traversal structures (the comonad
// zipper, optics, and transducers), which are *built on top of* these.
//
// Every algorithm restricts itself to the graph's declared nodes (those added via
// AddNode/WithNode). An edge may legitimately point at an identity that was never
// declared as a node (a dangling target); such targets are visited by the raw
// reachability walks (WalkFrom/BFS/DFS) but are ignored by TopoSort and SCC, whose
// results are only meaningful over the declared vertex set.
//
// Internal scratch state uses ordinary Go maps and slices: the *graph itself*
// stays persistent and immutable, but a single algorithm invocation is free to
// use transient local bookkeeping — the dogfooding rule concerns the stored data
// structure, not an algorithm's stack frame.

// FoldWalk reduces the nodes reachable from start, in the chosen walk order, with
// the accumulator f — a catamorphism *along a traversal* rather than over the
// unordered vertex set (FoldNodes).
//
// The walk is the same cycle-safe traversal as WalkFrom, so FoldWalk terminates on
// cyclic graphs, visiting each reachable node exactly once. Reachable identities
// that have no declared node (dangling edge targets) are skipped, so f only ever
// sees real nodes — matching the element type of FoldNodes.
//
// Example: summing an int payload in breadth-first order accumulates start's
// payload first, then its neighbours', and so on.
func (g *Graph[K, C, N, E]) FoldWalk(
	start K, strategy WalkStrategy, initial any,
	f func(acc any, n functions.Node[K, C, N]) any,
) any {
	src := g.norm()
	acc := initial
	for k := range src.WalkFrom(start, strategy) {
		if n, ok := src.GetNode(k); ok {
			acc = f(acc, n)
		}
	}
	return acc
}

// edgeCost extracts an edge's numeric weight for the weighted algorithms. Any
// weight convertible to a float64 is used directly; a non-numeric weight defaults
// to 1, so an "unweighted" graph (the default weight is int 1) behaves like a
// hop-count graph and ShortestPath degenerates to a fewest-edges search.
func edgeCost[K comparable, C comparable, E any](e functions.Edge[K, C, E]) float64 {
	if f, ok := functions.ToFloat64(e.Weight); ok {
		return f
	}
	return 1
}

// ShortestPath computes a minimum-weight directed path from one node to another
// using Dijkstra's algorithm, returning the node identities along the path
// (inclusive of both endpoints), the path's total weight, and whether any path
// exists.
//
// # The algorithm
//
// Dijkstra grows a frontier outward from the source, always finalising the
// not-yet-finalised node of least known distance. Because edge weights are taken
// to be non-negative (edgeCost), once a node is finalised its distance cannot
// improve, which is what makes the greedy choice correct. Ties between equal
// distances are broken by node identity so the result is deterministic across
// runs. This implementation is the simple O(V²) array variant — ample for the
// infrastructure graphs this engine targets, and free of any heap dependency.
//
// A zero-length path (from == to) yields the single-element path [from] with
// weight 0. When to is unreachable from from, ok is false.
func (g *Graph[K, C, N, E]) ShortestPath(from, to K) (path []K, cost float64, ok bool) {
	src := g.norm()
	dist := map[K]float64{from: 0}
	prev := map[K]K{}
	visited := map[K]bool{}

	for {
		// Select the unvisited node of least tentative distance, breaking ties
		// deterministically by identity.
		var cur K
		found := false
		best := math.Inf(1)
		for k, d := range dist {
			if visited[k] {
				continue
			}
			if !found || d < best || (d == best && compareIDs(k, cur) < 0) {
				best, cur, found = d, k, true
			}
		}
		if !found || cur == to {
			break
		}
		visited[cur] = true
		for _, e := range src.OutEdges(cur) {
			nd := dist[cur] + edgeCost(e)
			if od, seen := dist[e.To]; !seen || nd < od {
				dist[e.To] = nd
				prev[e.To] = cur
			}
		}
	}

	total, reached := dist[to]
	if !reached {
		return nil, 0, false
	}
	// Reconstruct the path backwards through the predecessor chain.
	path = []K{to}
	for cur := to; cur != from; {
		p, has := prev[cur]
		if !has {
			// to == from (already [to]) or a disconnected sink; the latter is
			// impossible once dist[to] exists via a predecessor, so this guards
			// only the from==to case where the loop never iterates.
			break
		}
		path = append([]K{p}, path...)
		cur = p
	}
	return path, total, true
}

// TopoSort returns a topological ordering of the declared nodes as Right, or — if
// the graph contains a directed cycle, so no such ordering exists — the set of
// nodes that could not be ordered (the cyclic remainder) as Left.
//
// # Why an Either
//
// A topological sort is only defined on a DAG; a graph with a cycle has no linear
// order respecting every edge. Rather than signalling failure out-of-band, the
// result is an Either[cycle, order]: Right carries the ordering (every edge points
// forwards in it), Left carries the identities still entangled in cycles. This is
// the idiomatic total-function encoding — the caller pattern-matches the two
// outcomes.
//
// The algorithm is Kahn's: repeatedly emit a source (a node with no remaining
// incoming edges), always choosing the least identity so the ordering is
// deterministic. Self-loops and mutual cycles leave their nodes with a permanently
// positive in-degree, which is exactly the Left remainder. Edges to undeclared
// targets are ignored.
func (g *Graph[K, C, N, E]) TopoSort() functions.Either[[]K, []K] {
	src := g.norm()
	ids := src.NodeIDs() // deterministic order (sorted by identity)

	indeg := make(map[K]int, len(ids))
	for _, id := range ids {
		indeg[id] = 0
	}
	succ := make(map[K][]K, len(ids))
	for _, id := range ids {
		for _, nb := range src.Neighbors(id) {
			if _, declared := indeg[nb]; !declared {
				continue // ignore edges to undeclared targets
			}
			succ[id] = append(succ[id], nb)
			indeg[nb]++
		}
	}

	order := make([]K, 0, len(ids))
	done := make(map[K]bool, len(ids))
	for len(order) < len(ids) {
		progressed := false
		for _, id := range ids { // ids sorted → smallest eligible source first
			if done[id] || indeg[id] != 0 {
				continue
			}
			done[id] = true
			order = append(order, id)
			for _, nb := range succ[id] {
				indeg[nb]--
			}
			progressed = true
			break
		}
		if !progressed {
			break // only cyclic nodes remain
		}
	}

	if len(order) == len(ids) {
		return functions.Right[[]K, []K](order)
	}
	var remainder []K
	for _, id := range ids {
		if !done[id] {
			remainder = append(remainder, id)
		}
	}
	return functions.Left[[]K, []K](remainder)
}

// SCC decomposes the graph into its strongly connected components using Tarjan's
// algorithm: a maximal set of nodes each reachable from every other in the set.
//
// # What a component is
//
// Two nodes are in the same strongly connected component iff there is a directed
// path each way between them. Every single node is trivially in a component of its
// own unless a larger cycle absorbs it, so the returned slice partitions the whole
// declared vertex set. Condensing each component to a point yields a DAG — SCC is
// therefore the standard pre-processing step that lets cyclic-graph problems reuse
// DAG algorithms (and complements TopoSort, which fails precisely on non-trivial
// components).
//
// # Determinism
//
// Tarjan discovers components in a single depth-first pass. For reproducibility
// each component is returned with its identities sorted, and the components
// themselves are sorted by their least identity.
func (g *Graph[K, C, N, E]) SCC() [][]K {
	src := g.norm()
	ids := src.NodeIDs()
	declared := make(map[K]bool, len(ids))
	for _, id := range ids {
		declared[id] = true
	}

	index := make(map[K]int, len(ids))
	low := make(map[K]int, len(ids))
	onStack := make(map[K]bool, len(ids))
	var stack []K
	counter := 0
	var comps [][]K

	var strongconnect func(v K)
	strongconnect = func(v K) {
		index[v], low[v] = counter, counter
		counter++
		stack = append(stack, v)
		onStack[v] = true
		for _, w := range src.Neighbors(v) {
			if !declared[w] {
				continue
			}
			if _, seen := index[w]; !seen {
				strongconnect(w)
				if low[w] < low[v] {
					low[v] = low[w]
				}
			} else if onStack[w] {
				if index[w] < low[v] {
					low[v] = index[w]
				}
			}
		}
		if low[v] == index[v] {
			var comp []K
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				comp = append(comp, w)
				if w == v {
					break
				}
			}
			sort.SliceStable(comp, func(i, j int) bool {
				return compareIDs(comp[i], comp[j]) < 0
			})
			comps = append(comps, comp)
		}
	}

	for _, id := range ids {
		if _, seen := index[id]; !seen {
			strongconnect(id)
		}
	}
	sort.SliceStable(comps, func(i, j int) bool {
		return compareIDs(comps[i][0], comps[j][0]) < 0
	})
	return comps
}
