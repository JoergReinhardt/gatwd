// Package arrows implements Category Theory Arrow combinators for functional stream processing.
//
// In functional programming, Arrows generalize monads by modeling computations with both inputs
// and outputs (A -> B). They allow complex parallel and split routing through pipelines without
// shared mutable state.
//
// Key Combinators:
//
//   - Kleisli Arrows: Wraps a monadic function (A -> Either[error, B]) into a composable arrow.
//
//   - FanOut (&&&): Evaluates two independent arrows concurrently on a single input node,
//     yielding a Product Type Tuple (B, C).
//
//   - Split (***): Takes a Product Type Tuple (A, B) and routes the first element to Arrow 1
//     and the second element to Arrow 2, producing (C, D).
package arrows
