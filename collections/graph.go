package collections

import (
	"fmt"
	"iter"
	"sort"

	"github.com/joergreinhardt/gatwd/functions"
)

// Graph is a fully persistent, immutable, directed, typed ("colored"), weighted
// multigraph.
//
// # What it is, mathematically
//
// A directed multigraph is a set of vertices together with a bag of directed
// edges — "multi" because two vertices may be joined by more than one edge (a
// Service may both "selects" and "owns" the same Pod, and both facts are kept).
// "Typed/colored" means every vertex and edge carries a discriminator (its
// color), so a single Graph can hold heterogeneous relations and be queried by
// type. "Weighted" means each edge carries a payload used by weighted walks.
// The graph is *cyclic by design*: unlike the acyclic single-root Tree, a Graph
// may contain cycles, so every traversal is made cycle-safe by a visited set.
//
// # How it is represented (dogfooding the persistent collections)
//
// Rather than a raw Go map, the adjacency is three HAMTs — giving O(log32 N)
// updates with structural sharing, so an incremental change (an operator's watch
// event adding one node) copies a handful of trie nodes, not the whole graph:
//
//   - nodes: HAMT[K, Node]        — the declared vertices and their payloads
//   - out:   HAMT[K, List[Edge]]  — forward adjacency (successor edges)
//   - in:    HAMT[K, List[Edge]]  — reverse adjacency (predecessor edges)
//
// Keeping the reverse index makes predecessor queries and node removal cheap.
//
// # Type parameters
//
//   - K is the vertex identity (comparable, used as the HAMT key).
//   - C is the vertex/edge color (comparable, used for by-color indexing).
//   - N is the free-form node payload.
//   - E is the edge weight/payload.
//
// All operations are immutable and return a new Graph; the receiver is never
// mutated, so a Graph is safe to share across goroutines.
type Graph[K comparable, C comparable, N any, E any] struct {
	nodes *HAMT[K, functions.Node[K, C, N]]
	out   *HAMT[K, *List[functions.Edge[K, C, E]]]
	in    *HAMT[K, *List[functions.Edge[K, C, E]]]
}

// NewGraph returns an empty graph.
func NewGraph[K comparable, C comparable, N any, E any]() *Graph[K, C, N, E] {
	return &Graph[K, C, N, E]{
		nodes: NewHAMT[K, functions.Node[K, C, N]](),
		out:   NewHAMT[K, *List[functions.Edge[K, C, E]]](),
		in:    NewHAMT[K, *List[functions.Edge[K, C, E]]](),
	}
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (g *Graph[K, C, N, E]) norm() *Graph[K, C, N, E] {
	if g == nil {
		return NewGraph[K, C, N, E]()
	}
	res := *g
	if res.nodes == nil {
		res.nodes = NewHAMT[K, functions.Node[K, C, N]]()
	}
	if res.out == nil {
		res.out = NewHAMT[K, *List[functions.Edge[K, C, E]]]()
	}
	if res.in == nil {
		res.in = NewHAMT[K, *List[functions.Edge[K, C, E]]]()
	}
	return &res
}

// edgesOf collects a persistent edge list into a plain slice.
func edgesOf[K comparable, C comparable, E any](l *List[functions.Edge[K, C, E]]) []functions.Edge[K, C, E] {
	if l == nil {
		return nil
	}
	var out []functions.Edge[K, C, E]
	for e := range l.Seq() {
		out = append(out, e)
	}
	return out
}

// listOfEdges builds a persistent edge list from a slice (preserving order).
func listOfEdges[K comparable, C comparable, E any](edges []functions.Edge[K, C, E]) *List[functions.Edge[K, C, E]] {
	res := Nil[functions.Edge[K, C, E]]()
	for i := len(edges) - 1; i >= 0; i-- {
		res = Cons(edges[i], res)
	}
	return res
}

// sortEdges returns edges in the deterministic lexicographic order defined by
// Edge.Compare, so projections and walks are reproducible across runs.
func sortEdges[K comparable, C comparable, E any](edges []functions.Edge[K, C, E]) []functions.Edge[K, C, E] {
	sort.SliceStable(edges, func(i, j int) bool {
		return edges[i].Compare(edges[j]) < 0
	})
	return edges
}

// indexAppend returns a new adjacency HAMT with edge e appended under key.
func indexAppend[K comparable, C comparable, E any](
	idx *HAMT[K, *List[functions.Edge[K, C, E]]], key K, e functions.Edge[K, C, E],
) *HAMT[K, *List[functions.Edge[K, C, E]]] {
	existing, _ := idx.Get(key)
	return idx.Set(key, Cons(e, existing))
}

// indexRemove returns a new adjacency HAMT with the first edge Equal to e removed
// from the list under key. The key is dropped entirely when its list empties.
func indexRemove[K comparable, C comparable, E any](
	idx *HAMT[K, *List[functions.Edge[K, C, E]]], key K, e functions.Edge[K, C, E],
) *HAMT[K, *List[functions.Edge[K, C, E]]] {
	existing, ok := idx.Get(key)
	if !ok {
		return idx
	}
	kept := make([]functions.Edge[K, C, E], 0)
	removed := false
	for cur := range existing.Seq() {
		if !removed && cur.Equal(e) {
			removed = true
			continue
		}
		kept = append(kept, cur)
	}
	if !removed {
		return idx
	}
	if len(kept) == 0 {
		return idx.Delete(key)
	}
	return idx.Set(key, listOfEdges(kept))
}

// ---------------------------------------------------------------------------
// Construction
// ---------------------------------------------------------------------------

// WithNode returns a new graph with the node added, or its payload/color updated
// if a node with the same identity already exists.
func (g *Graph[K, C, N, E]) WithNode(node functions.Node[K, C, N]) *Graph[K, C, N, E] {
	res := g.norm()
	res.nodes = res.nodes.Set(node.ID, node)
	return res
}

// AddNode is sugar for WithNode constructing the node from its parts.
func (g *Graph[K, C, N, E]) AddNode(id K, color C, payload N) *Graph[K, C, N, E] {
	return g.WithNode(functions.NewNode(id, color, payload))
}

// AddEdge returns a new graph with the directed edge inserted into both the
// forward and reverse adjacency indexes. Parallel edges are permitted (this is a
// multigraph); use Union for de-duplicating merges.
func (g *Graph[K, C, N, E]) AddEdge(e functions.Edge[K, C, E]) *Graph[K, C, N, E] {
	res := g.norm()
	res.out = indexAppend(res.out, e.From, e)
	res.in = indexAppend(res.in, e.To, e)
	return res
}

// Connect is sugar for AddEdge building the edge from its parts (default,
// ungrouped, not inferred).
func (g *Graph[K, C, N, E]) Connect(from, to K, color C, weight E) *Graph[K, C, N, E] {
	return g.AddEdge(functions.NewEdge(from, to, color, weight))
}

// AddInferredEdge returns a new graph with an inferred/assumed edge (see
// functions.Edge.Inferred) inserted.
func (g *Graph[K, C, N, E]) AddInferredEdge(from, to K, color C, weight E) *Graph[K, C, N, E] {
	return g.AddEdge(functions.NewInferredEdge(from, to, color, weight))
}

// AddFanOut models one logical 1:n relation as n grouped binary edges sharing
// the given group id and color (e.g. one Service that "selects" several Pods).
// EdgesByGroup fetches the whole relation back as a unit.
func (g *Graph[K, C, N, E]) AddFanOut(color C, weight E, group string, from K, tos []K) *Graph[K, C, N, E] {
	res := g.norm()
	for _, to := range tos {
		res = res.AddEdge(functions.NewGroupedEdge(from, to, color, weight, group))
	}
	return res
}

// AddFanIn models one logical n:1 relation as n grouped binary edges sharing the
// given group id and color (e.g. several Pods that are all "owned by" one
// ReplicaSet).
func (g *Graph[K, C, N, E]) AddFanIn(color C, weight E, group string, froms []K, to K) *Graph[K, C, N, E] {
	res := g.norm()
	for _, from := range froms {
		res = res.AddEdge(functions.NewGroupedEdge(from, to, color, weight, group))
	}
	return res
}

// RemoveEdge returns a new graph with the first edge Equal to e removed from both
// adjacency indexes.
func (g *Graph[K, C, N, E]) RemoveEdge(e functions.Edge[K, C, E]) *Graph[K, C, N, E] {
	res := g.norm()
	res.out = indexRemove(res.out, e.From, e)
	res.in = indexRemove(res.in, e.To, e)
	return res
}

// RemoveNode returns a new graph with the node and every edge incident to it
// (both its out-edges and the edges pointing at it) removed. The reverse index
// makes locating incoming edges cheap.
func (g *Graph[K, C, N, E]) RemoveNode(id K) *Graph[K, C, N, E] {
	res := g.norm()
	// Detach incoming edges from their sources' out-lists.
	for _, e := range edgesOf(mustList(res.in, id)) {
		res.out = indexRemove(res.out, e.From, e)
	}
	// Detach outgoing edges from their targets' in-lists.
	for _, e := range edgesOf(mustList(res.out, id)) {
		res.in = indexRemove(res.in, e.To, e)
	}
	res.out = res.out.Delete(id)
	res.in = res.in.Delete(id)
	res.nodes = res.nodes.Delete(id)
	return res
}

// mustList fetches an adjacency list, returning an empty list when absent.
func mustList[K comparable, C comparable, E any](
	idx *HAMT[K, *List[functions.Edge[K, C, E]]], key K,
) *List[functions.Edge[K, C, E]] {
	if l, ok := idx.Get(key); ok {
		return l
	}
	return Nil[functions.Edge[K, C, E]]()
}

// ---------------------------------------------------------------------------
// Projection
// ---------------------------------------------------------------------------

// Order returns the number of declared nodes (the graph's order |V|).
func (g *Graph[K, C, N, E]) Order() int {
	if g == nil || g.nodes == nil {
		return 0
	}
	return g.nodes.Len()
}

// Size returns the number of edges (the graph's size |E|).
func (g *Graph[K, C, N, E]) Size() int {
	if g == nil || g.out == nil {
		return 0
	}
	total := 0
	for _, l := range g.out.Seq() {
		total += l.Len()
	}
	return total
}

// IsEmpty reports whether the graph has no nodes and no edges.
func (g *Graph[K, C, N, E]) IsEmpty() bool {
	return g.Order() == 0 && g.Size() == 0
}

// HasNode reports whether a node with the given identity has been declared.
func (g *Graph[K, C, N, E]) HasNode(id K) bool {
	if g == nil || g.nodes == nil {
		return false
	}
	_, ok := g.nodes.Get(id)
	return ok
}

// GetNode returns the declared node for an identity, and whether it was present.
func (g *Graph[K, C, N, E]) GetNode(id K) (functions.Node[K, C, N], bool) {
	if g == nil || g.nodes == nil {
		var zero functions.Node[K, C, N]
		return zero, false
	}
	return g.nodes.Get(id)
}

// Nodes returns all declared nodes, ordered deterministically by identity.
func (g *Graph[K, C, N, E]) Nodes() []functions.Node[K, C, N] {
	if g == nil || g.nodes == nil {
		return nil
	}
	var out []functions.Node[K, C, N]
	for _, n := range g.nodes.Seq() {
		out = append(out, n)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return compareIDs(out[i].ID, out[j].ID) < 0
	})
	return out
}

// compareIDs orders two node identities deterministically for reproducible
// projections, degrading to a formatted-string comparison for exotic key types.
func compareIDs(a, b any) int {
	if c, err := functions.Compare(a, b); err == nil {
		return c
	}
	sa, sb := fmt.Sprintf("%v", a), fmt.Sprintf("%v", b)
	switch {
	case sa < sb:
		return -1
	case sa > sb:
		return 1
	}
	return 0
}

// NodeIDs returns all declared node identities in deterministic order.
func (g *Graph[K, C, N, E]) NodeIDs() []K {
	nodes := g.Nodes()
	ids := make([]K, 0, len(nodes))
	for _, n := range nodes {
		ids = append(ids, n.ID)
	}
	return ids
}

// Edges returns every edge in the graph, in deterministic lexicographic order.
func (g *Graph[K, C, N, E]) Edges() []functions.Edge[K, C, E] {
	if g == nil || g.out == nil {
		return nil
	}
	var all []functions.Edge[K, C, E]
	for _, l := range g.out.Seq() {
		all = append(all, edgesOf(l)...)
	}
	return sortEdges(all)
}

// OutEdges returns the edges leaving a node, in deterministic order.
func (g *Graph[K, C, N, E]) OutEdges(id K) []functions.Edge[K, C, E] {
	if g == nil || g.out == nil {
		return nil
	}
	l, _ := g.out.Get(id)
	return sortEdges(edgesOf(l))
}

// InEdges returns the edges entering a node, in deterministic order.
func (g *Graph[K, C, N, E]) InEdges(id K) []functions.Edge[K, C, E] {
	if g == nil || g.in == nil {
		return nil
	}
	l, _ := g.in.Get(id)
	return sortEdges(edgesOf(l))
}

// Neighbors returns the distinct successor identities of a node (via out-edges),
// in deterministic order. An edge whose far endpoint was never declared as a node
// is not followed: adjacency agrees with the node listing (NodeIDs, Nodes), so a
// walk never yields an id that has no node. Such an edge is still stored — Edges
// and Size count it — because Connect does not create endpoints; declare the node
// and the edge starts to count.
func (g *Graph[K, C, N, E]) Neighbors(id K) []K {
	return dedupIDs(g.OutEdges(id), func(e functions.Edge[K, C, E]) K { return e.To }, g.HasNode)
}

// Predecessors returns the distinct predecessor identities of a node (via
// in-edges), in deterministic order, leaving out edges whose source is not a
// declared node (see Neighbors).
func (g *Graph[K, C, N, E]) Predecessors(id K) []K {
	return dedupIDs(g.InEdges(id), func(e functions.Edge[K, C, E]) K { return e.From }, g.HasNode)
}

// dedupIDs projects an endpoint out of each edge, drops endpoints that are not
// declared nodes (see Neighbors) and removes duplicates, preserving the (already
// sorted) edge order.
func dedupIDs[K comparable, C comparable, E any](edges []functions.Edge[K, C, E], pick func(functions.Edge[K, C, E]) K, declared func(K) bool) []K {
	seen := make(map[K]struct{}, len(edges))
	var out []K
	for _, e := range edges {
		k := pick(e)
		if !declared(k) {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	return out
}

// NodesByColor returns all declared nodes whose color equals c.
func (g *Graph[K, C, N, E]) NodesByColor(c C) []functions.Node[K, C, N] {
	var out []functions.Node[K, C, N]
	for _, n := range g.Nodes() {
		if n.Color == c {
			out = append(out, n)
		}
	}
	return out
}

// EdgesByColor returns every edge whose color equals c, in deterministic order.
func (g *Graph[K, C, N, E]) EdgesByColor(c C) []functions.Edge[K, C, E] {
	var out []functions.Edge[K, C, E]
	for _, e := range g.Edges() {
		if e.Color == c {
			out = append(out, e)
		}
	}
	return out
}

// EdgesByGroup returns the binary edges that together form one logical n-ary
// relation (those sharing the given group id), in deterministic order.
func (g *Graph[K, C, N, E]) EdgesByGroup(group string) []functions.Edge[K, C, E] {
	var out []functions.Edge[K, C, E]
	for _, e := range g.Edges() {
		if e.Group == group {
			out = append(out, e)
		}
	}
	return out
}

// InferredEdges returns every edge marked inferred/assumed, in deterministic
// order.
func (g *Graph[K, C, N, E]) InferredEdges() []functions.Edge[K, C, E] {
	var out []functions.Edge[K, C, E]
	for _, e := range g.Edges() {
		if e.Inferred {
			out = append(out, e)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Functor
// ---------------------------------------------------------------------------

// MapNodes returns a new graph with f applied to every node payload, leaving
// identities, colors, and the full edge structure untouched. This is the Functor
// map over node payloads.
func (g *Graph[K, C, N, E]) MapNodes(f func(N) N) *Graph[K, C, N, E] {
	src := g.norm()
	res := NewGraph[K, C, N, E]()
	for _, n := range src.Nodes() {
		res.nodes = res.nodes.Set(n.ID, functions.NewNode(n.ID, n.Color, f(n.Payload)))
	}
	res.out, res.in = src.out, src.in
	return res
}

// MapEdges returns a new graph with f applied to every edge weight, leaving nodes
// and edge topology untouched. This is the Functor map over edge weights.
func (g *Graph[K, C, N, E]) MapEdges(f func(E) E) *Graph[K, C, N, E] {
	return g.rebuild(func(n functions.Node[K, C, N]) functions.Node[K, C, N] { return n },
		func(e functions.Edge[K, C, E]) functions.Edge[K, C, E] {
			e.Weight = f(e.Weight)
			return e
		})
}

// Bimap returns a new graph with fn applied to every node payload and fe applied
// to every edge weight — the Bifunctor map treating nodes and edges as the two
// independent variances of a graph.
func (g *Graph[K, C, N, E]) Bimap(fn func(N) N, fe func(E) E) *Graph[K, C, N, E] {
	return g.rebuild(
		func(n functions.Node[K, C, N]) functions.Node[K, C, N] {
			return functions.NewNode(n.ID, n.Color, fn(n.Payload))
		},
		func(e functions.Edge[K, C, E]) functions.Edge[K, C, E] {
			e.Weight = fe(e.Weight)
			return e
		})
}

// rebuild reconstructs the graph from scratch, transforming each node and edge,
// keeping both adjacency indexes in sync. It underlies MapEdges/Bimap/Filter.
func (g *Graph[K, C, N, E]) rebuild(
	fn func(functions.Node[K, C, N]) functions.Node[K, C, N],
	fe func(functions.Edge[K, C, E]) functions.Edge[K, C, E],
) *Graph[K, C, N, E] {
	src := g.norm()
	res := NewGraph[K, C, N, E]()
	for _, n := range src.Nodes() {
		res = res.WithNode(fn(n))
	}
	for _, e := range src.Edges() {
		res = res.AddEdge(fe(e))
	}
	return res
}

// ---------------------------------------------------------------------------
// Foldable
// ---------------------------------------------------------------------------

// FoldNodes reduces all nodes (in deterministic order) using the accumulator f —
// the catamorphism over the vertex set.
func (g *Graph[K, C, N, E]) FoldNodes(initial any, f func(acc any, n functions.Node[K, C, N]) any) any {
	acc := initial
	for _, n := range g.Nodes() {
		acc = f(acc, n)
	}
	return acc
}

// FoldEdges reduces all edges (in deterministic order) using the accumulator f —
// the catamorphism over the edge set.
func (g *Graph[K, C, N, E]) FoldEdges(initial any, f func(acc any, e functions.Edge[K, C, E]) any) any {
	acc := initial
	for _, e := range g.Edges() {
		acc = f(acc, e)
	}
	return acc
}

// ---------------------------------------------------------------------------
// Filterable
// ---------------------------------------------------------------------------

// FilterNodes returns the induced subgraph over the nodes satisfying predicate:
// rejected nodes are dropped together with every edge incident to them.
func (g *Graph[K, C, N, E]) FilterNodes(predicate func(functions.Node[K, C, N]) bool) *Graph[K, C, N, E] {
	src := g.norm()
	res := NewGraph[K, C, N, E]()
	keep := make(map[K]struct{})
	for _, n := range src.Nodes() {
		if predicate(n) {
			keep[n.ID] = struct{}{}
			res = res.WithNode(n)
		}
	}
	for _, e := range src.Edges() {
		if _, okF := keep[e.From]; !okF {
			continue
		}
		if _, okT := keep[e.To]; !okT {
			continue
		}
		res = res.AddEdge(e)
	}
	return res
}

// FilterEdges returns a new graph keeping every node but only the edges
// satisfying predicate.
func (g *Graph[K, C, N, E]) FilterEdges(predicate func(functions.Edge[K, C, E]) bool) *Graph[K, C, N, E] {
	src := g.norm()
	res := NewGraph[K, C, N, E]()
	for _, n := range src.Nodes() {
		res = res.WithNode(n)
	}
	for _, e := range src.Edges() {
		if predicate(e) {
			res = res.AddEdge(e)
		}
	}
	return res
}

// RejectNodes is the complement of FilterNodes.
func (g *Graph[K, C, N, E]) RejectNodes(predicate func(functions.Node[K, C, N]) bool) *Graph[K, C, N, E] {
	return g.FilterNodes(func(n functions.Node[K, C, N]) bool { return !predicate(n) })
}

// RejectEdges is the complement of FilterEdges.
func (g *Graph[K, C, N, E]) RejectEdges(predicate func(functions.Edge[K, C, E]) bool) *Graph[K, C, N, E] {
	return g.FilterEdges(func(e functions.Edge[K, C, E]) bool { return !predicate(e) })
}

// ---------------------------------------------------------------------------
// Monoid / set algebra
// ---------------------------------------------------------------------------

// Union returns the graph monoid's combination of g and other: the union of
// their node sets (other's payload wins on identity collisions, matching HAMT's
// right-biased Merge) and the union of their edge bags, de-duplicating edges that
// are structurally Equal.
func (g *Graph[K, C, N, E]) Union(other *Graph[K, C, N, E]) *Graph[K, C, N, E] {
	a, b := g.norm(), other.norm()
	res := NewGraph[K, C, N, E]()
	for _, n := range a.Nodes() {
		res = res.WithNode(n)
	}
	for _, n := range b.Nodes() {
		res = res.WithNode(n)
	}
	seen := make(map[string]struct{})
	add := func(e functions.Edge[K, C, E]) {
		k := edgeKey(e)
		if _, ok := seen[k]; ok {
			return
		}
		seen[k] = struct{}{}
		res = res.AddEdge(e)
	}
	for _, e := range a.Edges() {
		add(e)
	}
	for _, e := range b.Edges() {
		add(e)
	}
	return res
}

// Diff returns g with every node and edge that also appears in other removed
// (relative complement on both vertex and edge sets). Removing a node also drops
// its incident edges, as in RemoveNode.
func (g *Graph[K, C, N, E]) Diff(other *Graph[K, C, N, E]) *Graph[K, C, N, E] {
	a, b := g.norm(), other.norm()
	res := a
	otherEdges := make(map[string]struct{})
	for _, e := range b.Edges() {
		otherEdges[edgeKey(e)] = struct{}{}
	}
	res = res.RejectEdges(func(e functions.Edge[K, C, E]) bool {
		_, ok := otherEdges[edgeKey(e)]
		return ok
	})
	for _, n := range b.Nodes() {
		if res.HasNode(n.ID) {
			res = res.RemoveNode(n.ID)
		}
	}
	return res
}

// Intersect returns the graph containing the nodes present in both g and other
// (g's payload kept) and the edges present in both.
func (g *Graph[K, C, N, E]) Intersect(other *Graph[K, C, N, E]) *Graph[K, C, N, E] {
	a, b := g.norm(), other.norm()
	res := NewGraph[K, C, N, E]()
	for _, n := range a.Nodes() {
		if b.HasNode(n.ID) {
			res = res.WithNode(n)
		}
	}
	otherEdges := make(map[string]struct{})
	for _, e := range b.Edges() {
		otherEdges[edgeKey(e)] = struct{}{}
	}
	for _, e := range a.Edges() {
		if _, ok := otherEdges[edgeKey(e)]; ok {
			res = res.AddEdge(e)
		}
	}
	return res
}

// edgeKey renders an edge as a canonical string for set-membership de-duplication
// in Union/Diff/Intersect.
func edgeKey[K comparable, C comparable, E any](e functions.Edge[K, C, E]) string {
	return fmt.Sprintf("%v\x00%v\x00%v\x00%v\x00%s\x00%t", e.From, e.To, e.Color, e.Weight, e.Group, e.Inferred)
}

// ---------------------------------------------------------------------------
// Walks (cycle-safe)
// ---------------------------------------------------------------------------

// WalkStrategy selects the traversal order used by WalkFrom.
type WalkStrategy int

const (
	// WalkBFS visits nodes in breadth-first order (a queue frontier).
	WalkBFS WalkStrategy = iota
	// WalkDFS visits nodes in depth-first order (a stack frontier).
	WalkDFS
)

// WalkFrom returns a lazy, cycle-safe iterator over the identities reachable from
// start (inclusive), visiting each node exactly once in the chosen strategy's
// order. A visited Set guards against the cycles a Graph may contain.
func (g *Graph[K, C, N, E]) WalkFrom(start K, strategy WalkStrategy) iter.Seq[K] {
	return func(yield func(K) bool) {
		src := g.norm()
		visited := NewSet[K]()
		frontier := []K{start}
		for len(frontier) > 0 {
			var cur K
			if strategy == WalkDFS {
				cur = frontier[len(frontier)-1]
				frontier = frontier[:len(frontier)-1]
			} else {
				cur = frontier[0]
				frontier = frontier[1:]
			}
			if visited.Contains(cur) {
				continue
			}
			visited = visited.Add(cur)
			if !yield(cur) {
				return
			}
			neighbors := src.Neighbors(cur)
			if strategy == WalkDFS {
				// Push in reverse so the smallest neighbor is popped first,
				// giving a deterministic left-to-right depth-first order.
				for i := len(neighbors) - 1; i >= 0; i-- {
					if !visited.Contains(neighbors[i]) {
						frontier = append(frontier, neighbors[i])
					}
				}
			} else {
				for _, nb := range neighbors {
					if !visited.Contains(nb) {
						frontier = append(frontier, nb)
					}
				}
			}
		}
	}
}

// BFS returns a cycle-safe breadth-first iterator from start (inclusive).
func (g *Graph[K, C, N, E]) BFS(start K) iter.Seq[K] {
	return g.WalkFrom(start, WalkBFS)
}

// DFS returns a cycle-safe depth-first iterator from start (inclusive).
func (g *Graph[K, C, N, E]) DFS(start K) iter.Seq[K] {
	return g.WalkFrom(start, WalkDFS)
}

// Reachable returns the set of identities reachable from start, including start
// itself. Because the walk is cycle-safe, this terminates on cyclic graphs.
func (g *Graph[K, C, N, E]) Reachable(start K) *Set[K] {
	res := NewSet[K]()
	for k := range g.BFS(start) {
		res = res.Add(k)
	}
	return res
}

// ---------------------------------------------------------------------------
// Equality & introspection
// ---------------------------------------------------------------------------

// Equal reports whether g and other have the same nodes (identities, colors, and
// structurally-equal payloads) and the same edge bag. Edge comparison is
// order-independent: both edge sets are canonically sorted, then compared
// element-by-element, so insertion order never affects equality.
func (g *Graph[K, C, N, E]) Equal(other *Graph[K, C, N, E]) bool {
	if g == other {
		return true
	}
	a, b := g.norm(), other.norm()
	if a.Order() != b.Order() || a.Size() != b.Size() {
		return false
	}
	for _, n := range a.Nodes() {
		on, ok := b.GetNode(n.ID)
		if !ok || !n.Equal(on) {
			return false
		}
	}
	ae, be := a.Edges(), b.Edges()
	for i := range ae {
		if !ae[i].Equal(be[i]) {
			return false
		}
	}
	return true
}

// IsGraph implements a marker method for type introspection by the operator and
// DSL layers.
func (g *Graph[K, C, N, E]) IsGraph() bool { return true }

// GraphEqual reports structural equality against another value, succeeding only
// when other is a *Graph of the same concrete type. It lets the untyped operator
// layer dispatch == without reflection.
func (g *Graph[K, C, N, E]) GraphEqual(other any) bool {
	o, ok := other.(*Graph[K, C, N, E])
	if !ok {
		return false
	}
	return g.Equal(o)
}

// String implements fmt.Stringer with a compact size summary.
func (g *Graph[K, C, N, E]) String() string {
	return fmt.Sprintf("Graph(|V|=%d, |E|=%d)", g.Order(), g.Size())
}

// Graph exposes the node and edge listing for the generic visualizer (tree/DOT
// formatting). The method name matches the visualizer's Graph() convention used
// by the other collections.
func (g *Graph[K, C, N, E]) Graph() any {
	if g == nil || g.IsEmpty() {
		return "<empty Graph>"
	}
	nodes := make([]string, 0, g.Order())
	for _, n := range g.Nodes() {
		nodes = append(nodes, n.String())
	}
	edges := make([]string, 0, g.Size())
	for _, e := range g.Edges() {
		edges = append(edges, e.String())
	}
	return struct {
		Nodes []string
		Edges []string
	}{Nodes: nodes, Edges: edges}
}
