package optics

import (
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

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

// VariantPrism constructs a Prism focusing on a specific branch tag of a dsltypes.Variant.
func VariantPrism[A any](tag string, flag dsltypes.BitFlag) Prism[dsltypes.Variant, A] {
	return Prism[dsltypes.Variant, A]{
		Preview: func(v dsltypes.Variant) functions.Maybe[A] {
			if v.Is(tag) {
				if cast, ok := v.Value.(A); ok {
					return functions.Just(cast)
				}
			}
			return functions.Nothing[A]()
		},
		Review: func(a A) dsltypes.Variant {
			return dsltypes.NewVariant(tag, flag, a)
		},
	}
}

// EitherRightPrism constructs a Prism focusing on the Right branch of an Either[L, R].
func EitherRightPrism[L, R any]() Prism[functions.Either[L, R], R] {
	return Prism[functions.Either[L, R], R]{
		Preview: func(e functions.Either[L, R]) functions.Maybe[R] {
			if e.IsRight() {
				r, _ := e.UnwrapRight()
				return functions.Just(r)
			}
			return functions.Nothing[R]()
		},
		Review: func(r R) functions.Either[L, R] {
			return functions.Right[L, R](r)
		},
	}
}

// EitherLeftPrism constructs a Prism focusing on the Left branch of an Either[L, R].
func EitherLeftPrism[L, R any]() Prism[functions.Either[L, R], L] {
	return Prism[functions.Either[L, R], L]{
		Preview: func(e functions.Either[L, R]) functions.Maybe[L] {
			if e.IsLeft() {
				l, _ := e.UnwrapLeft()
				return functions.Just(l)
			}
			return functions.Nothing[L]()
		},
		Review: func(l L) functions.Either[L, R] {
			return functions.Left[L, R](l)
		},
	}
}
