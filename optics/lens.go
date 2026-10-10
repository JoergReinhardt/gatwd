package optics

// Lens provides functional getters and setters for deep, immutable structures.
//
// In a purely functional architecture (like our immutable AST Tree), you cannot
// mutate a field directly (e.g., node.Value = X). A Lens encapsulates the logic
// to "focus" on a specific field A inside a structure S.
//
// It provides:
// 1. Get: Extract the focused value A from S.
// 2. Set: Return a BRAND NEW copy of S with the focused value updated to A.
type Lens[S, A any] struct {
	Get func(S) A
	Set func(S, A) S
}

// ComposeLenses mathematically composes two lenses.
// If you have a Lens from S to A, and a Lens from A to B, 
// you can compose them into a Lens from S to B.
// This allows you to chain deep modifications safely!
func ComposeLenses[S, A, B any](l1 Lens[S, A], l2 Lens[A, B]) Lens[S, B] {
	return Lens[S, B]{
		Get: func(s S) B {
			return l2.Get(l1.Get(s))
		},
		Set: func(s S, b B) S {
			// Get the inner A, modify it with l2, then set the new A into S with l1
			a := l1.Get(s)
			newA := l2.Set(a, b)
			return l1.Set(s, newA)
		},
	}
}
