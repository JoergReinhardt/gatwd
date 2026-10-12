// Package transducers implements composable, source-independent transformation
// recipes in the style of Clojure's transducers.
//
// # What is a transducer?
//
// A transducer is a transformation of a reducing function into another reducing
// function. A Reducer[A, R] folds a value of type A into an accumulator of type
// R (R, A) -> R. A Transducer[A, B, R] is a function that takes a Reducer over B
// and returns a Reducer over A — i.e. it describes "how to feed As into a process
// that consumes Bs", independent of where the As come from or how the result is
// built.
//
//	type Reducer[A, R any]       func(R, A) R
//	type Transducer[A, B, R any] func(Reducer[B, R]) Reducer[A, R]
//
// # Why do we need them here?
//
// Ordinary Map/Filter each allocate an intermediate collection. Because
// transducers operate purely at the reducing-function level, they *compose by
// ordinary function composition* and then run in a single pass with no
// intermediate allocation — the same recipe can drive a slice, a channel, or an
// iter.Seq. This is the allocation-free fusion the streaming engine relies on when
// a pipeline chains many Map/Filter stages over large inputs.
//
// # How it is implemented in Go
//
// Map lifts an element function A -> B; Filter lifts a predicate; both return a
// Transducer that wraps the downstream Reducer. Sequence applies a composed
// transducer to a slice and collects the result.
//
// # Quick start
//
//	// Double, then keep the even results, in one fused pass.
//	xf := transducers.Map[int, int, []int](func(x int) int { return x * 2 })
//	out := transducers.Sequence(input, xf) // []int
package transducers
