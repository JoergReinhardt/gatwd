package transducers

import (
	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// This file expresses a graph *walk as a transducer source*, so that a cycle-safe
// traversal composes with the same Map/Filter pipeline vocabulary as any other
// sequence — with zero intermediate allocations between the composed steps.
//
// A walk is naturally an unfold (one node expands into its successors); a
// transducer is a composable fold step. Bridging them means: materialise the walk
// order once (WalkNodes), then drive a composed Transducer over it (TransduceWalk).
// The traversal stays cycle-safe (each reachable node visited once) and the
// Map/Filter fusion stays allocation-free, exactly as Sequence provides for plain
// slices.

// WalkNodes returns the declared nodes reachable from start, in the chosen
// cycle-safe walk order (the same order as collections.Graph.WalkFrom). Reachable
// identities with no declared node (dangling edge targets) are skipped, so the
// result is a clean stream of real nodes ready to feed a transducer.
func WalkNodes[K comparable, C comparable, N any, E any](
	g *collections.Graph[K, C, N, E], start K, strategy collections.WalkStrategy,
) []functions.Node[K, C, N] {
	if g == nil {
		return nil
	}
	var out []functions.Node[K, C, N]
	for k := range g.WalkFrom(start, strategy) {
		if n, ok := g.GetNode(k); ok {
			out = append(out, n)
		}
	}
	return out
}

// TransduceWalk drives a composed Transducer over a graph walk: it takes the nodes
// visited from start in the given strategy's order and runs them through xf,
// collecting the transformed results. Because xf is an ordinary Transducer, a whole
// Map/Filter/… pipeline fuses into one pass over the walk — e.g. "breadth-first
// from the ingress, keep Service nodes, project their names" is one composed
// transducer applied here with no intermediate slices.
//
// It is the graph-shaped sibling of Sequence (which transduces a plain slice);
// indeed it is exactly Sequence over WalkNodes.
func TransduceWalk[K comparable, C comparable, N any, E any, B any](
	g *collections.Graph[K, C, N, E], start K, strategy collections.WalkStrategy,
	xf Transducer[functions.Node[K, C, N], B, []B],
) []B {
	return Sequence(WalkNodes(g, start, strategy), xf)
}
