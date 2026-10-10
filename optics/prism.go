package optics

import "github.com/joergreinhardt/gatwd/functions"

// Prism is an optic for handling Algebraic Sum Types (variants).
//
// While a Lens focuses on a field that MUST exist (like a field in a struct),
// a Prism focuses on a branch of a Sum Type that MIGHT exist (e.g., extracting
// the String value from an ASTNode IF it is a StringNode).
//
// 1. Preview: Attempts to extract the value A from S, returning Maybe[A].
// 2. Review: Wraps the value A back into the variant S.
type Prism[S, A any] struct {
	Preview func(S) functions.Maybe[A]
	Review  func(A) S
}

// ComposePrisms mathematically composes two Prisms.
func ComposePrisms[S, A, B any](p1 Prism[S, A], p2 Prism[A, B]) Prism[S, B] {
	return Prism[S, B]{
		Preview: func(s S) functions.Maybe[B] {
			// Try to get A from S. If it succeeds, try to get B from A.
			return functions.BindMaybe(p1.Preview(s), p2.Preview)
		},
		Review: func(b B) S {
			return p1.Review(p2.Review(b))
		},
	}
}
