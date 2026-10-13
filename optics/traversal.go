package optics

// Traversal is the optic for focusing on *zero or more* targets at once.
//
// A Lens focuses exactly one field that always exists; a Prism focuses one branch
// that may or may not be present; a Traversal generalises both to focus on an
// arbitrary number of targets inside a structure — every element of a collection,
// every edge of a graph, every leaf of a tree.
//
// It provides:
//  1. GetAll: extract all focused values A from S, in a deterministic order.
//  2. Modify: return a BRAND NEW S with f applied to every focused value.
//
// A Lens is the degenerate Traversal with exactly one target; a Prism is the
// degenerate Traversal with at most one. Keeping Traversal a standalone optic lets
// bulk immutable rewrites ("add one to every edge weight") compose with the same
// vocabulary as single-target Lens/Prism edits.
type Traversal[S, A any] struct {
	GetAll func(S) []A
	Modify func(S, func(A) A) S
}

// ComposeTraversals composes a Traversal from S to A with one from A to B into a
// Traversal from S to B: it focuses every B reachable through every A.
func ComposeTraversals[S, A, B any](t1 Traversal[S, A], t2 Traversal[A, B]) Traversal[S, B] {
	return Traversal[S, B]{
		GetAll: func(s S) []B {
			var out []B
			for _, a := range t1.GetAll(s) {
				out = append(out, t2.GetAll(a)...)
			}
			return out
		},
		Modify: func(s S, f func(B) B) S {
			return t1.Modify(s, func(a A) A {
				return t2.Modify(a, f)
			})
		},
	}
}

// Set is sugar for Modify with a constant function, replacing every focused value
// with the single value v.
func (t Traversal[S, A]) Set(s S, v A) S {
	return t.Modify(s, func(A) A { return v })
}
