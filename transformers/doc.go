// Package transformers implements monad transformers — constructs that stack one
// monad's behaviour on top of another so a single pipeline can combine several
// effects (error handling, threaded state) without nesting them by hand.
//
// # The problem monad transformers solve
//
// Real pipelines need more than one effect at once: a computation may both be
// able to fail (the Either monad) and carry evolving state (the State monad).
// Nesting these directly — Either[error, State[S, A]] — is awkward to thread
// through every step. A monad transformer adds one monad's capability *to an
// existing base monad*, giving a single flattened type to work with.
//
// # What this package provides
//
//   - EitherT[L, R]: the error/short-circuit transformer. It augments a base
//     computation with a Left/Right failure channel; a Left bypasses the
//     remaining steps (railway-oriented error handling) while Map transforms the
//     Right value. (The streaming engine's stream-level counterpart lives in
//     evaluator.EitherT, which lifts an Adapter into an error-aware Adapter.)
//
//   - StateT: the state transformer over an iter.Seq stream. Given an initial
//     state S and a step (S, A) -> (B, S), it threads the state left-to-right
//     through the stream, yielding a Pair[B, S] at each element — e.g. tracking
//     already-visited hosts while crawling a graph.
//
//   - StateContextT: StateT augmented with a context.Context so a stateful stream
//     can be cancelled cooperatively partway through.
//
// # How it is implemented in Go
//
// Each transformer is an ordinary generic function/type over Go iterators and the
// Layer-0 algebraic types (functions.Either, functions.Pair). There is no runtime
// type machinery — the "stacking" is expressed directly with Go generics.
//
// # Quick start
//
//	// Thread a running total through a stream of deltas.
//	sums := transformers.StateT(0, func(s, d int) (int, int) {
//	    return s + d, s + d // (yielded value, next state)
//	})
//	for pair := range sums(deltas) {
//	    // pair.Fst is the running total, pair.Snd is the state
//	}
package transformers
