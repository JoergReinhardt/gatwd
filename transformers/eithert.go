package transformers

import (
	"iter"
	"github.com/joergreinhardt/gatwd/functions"
)

// EitherT is a Monad Transformer that stacks the Either (Error) Monad 
// on top of the Iterator (Seq) Monad.
//
// Without transformers, dealing with a stream of operations that could fail
// requires unpacking `iter.Seq[Either[L, R]]` at every single step.
// EitherT cleanly abstracts this away, allowing you to map and bind over
// the success values (R) while short-circuiting on errors (L).
type EitherT[L, R any] struct {
	Stream iter.Seq[functions.Either[L, R]]
}

// Map applies a pure function to the Right values in the stream, ignoring the Lefts.
func (e EitherT[L, R]) Map(fn func(R) R) EitherT[L, R] {
	mapped := func(yield func(functions.Either[L, R]) bool) {
		for val := range e.Stream {
			if val.IsLeft() {
				if !yield(val) {
					return
				}
			} else {
				// Safely apply function to the Right value
				r, _ := val.UnwrapRight()
				newRight := functions.Right[L, R](fn(r))
				if !yield(newRight) {
					return
				}
			}
		}
	}
	return EitherT[L, R]{Stream: mapped}
}
