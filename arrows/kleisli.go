package arrows

import (
	"iter"
	"github.com/joergreinhardt/gatwd/functions"
)

// Kleisli represents an Arrow that encapsulates a monadic computation (A -> iter.Seq[B]).
// 
// In the gatwd architecture, this mathematically formalizes our Evaluator `Adapter`.
// By representing the Adapter as a Kleisli Arrow, we gain the ability to use 
// combinators like Split (***) and FanOut (&&&) to build complex routing logic.
type Kleisli[A, B any] func(A) iter.Seq[B]

// Compose combines two Kleisli Arrows: f(A) -> Seq[B] and g(B) -> Seq[C] 
// into a new Arrow h(A) -> Seq[C] by using Monadic Bind.
func Compose[A, B, C any](f Kleisli[A, B], g Kleisli[B, C]) Kleisli[A, C] {
	return func(a A) iter.Seq[C] {
		// Evaluate f(a) to get a sequence of B, then Bind g to that sequence.
		return functions.Bind(f(a), g)
	}
}

// FanOut (&&&) takes a single input A, routes it to both f and g independently,
// and zips their output streams together into a sequence of Pairs.
func FanOut[A, B, C any](f Kleisli[A, B], g Kleisli[A, C]) Kleisli[A, functions.Pair[B, C]] {
	return func(a A) iter.Seq[functions.Pair[B, C]] {
		return functions.Zip(f(a), g(a))
	}
}
