package arrows

import (
	"github.com/joergreinhardt/gatwd/functions"
	"iter"
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

// Split (***) takes a Pair(A, B), routes A to f and B to g,
// and zips their output streams together into a sequence of Pair(C, D).
func Split[A, B, C, D any](f Kleisli[A, C], g Kleisli[B, D]) Kleisli[functions.Pair[A, B], functions.Pair[C, D]] {
	return func(pair functions.Pair[A, B]) iter.Seq[functions.Pair[C, D]] {
		return functions.Zip(f(pair.Fst), g(pair.Snd))
	}
}

// Plus (<+>) is the ArrowPlus / MonadPlus choice combinator: it feeds one input A to
// every arrow in fs and concatenates their output sequences, in argument order.
//
// Where FanOut (&&&) pairs the outputs of two arrows and so needs exactly two, Plus
// is the monoid of Kleisli arrows under sequence concatenation: with no arrows it is
// the zero arrow (always empty), it is associative, and Plus(f) is f. It is the
// generic way to MERGE several sources into one stream — e.g. the crawl streams of
// many resource kinds feeding one downstream pipeline — without the combinator
// knowing anything about what the arrows produce.
//
// The arrows are evaluated lazily and in order: the outputs of fs[0] are yielded
// completely before fs[1] is started, and a consumer that stops early stops the
// remaining arrows from running at all.
func Plus[A, B any](fs ...Kleisli[A, B]) Kleisli[A, B] {
	return func(a A) iter.Seq[B] {
		return func(yield func(B) bool) {
			for _, f := range fs {
				for b := range f(a) {
					if !yield(b) {
						return
					}
				}
			}
		}
	}
}
