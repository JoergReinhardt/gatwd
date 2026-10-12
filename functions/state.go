package functions

import (
	"context"
)

// State represents a pure stateful computation: S -> (A, S).
// It encapsulates the canonical Haskell State Monad.
type State[S, A any] func(S) (A, S)

// Run evaluates the state computation with initial state s, returning both value and final state.
func (st State[S, A]) Run(s S) (A, S) {
	return st(s)
}

// Eval runs the computation and returns only the resulting value A.
func (st State[S, A]) Eval(s S) A {
	val, _ := st(s)
	return val
}

// Exec runs the computation and returns only the final state S.
func (st State[S, A]) Exec(s S) S {
	_, nextS := st(s)
	return nextS
}

// ReturnState lifts a pure value into a State computation without modifying state.
func ReturnState[S, A any](val A) State[S, A] {
	return func(s S) (A, S) {
		return val, s
	}
}

// BindState chains two stateful computations (monadic bind >>=).
func BindState[S, A, B any](m State[S, A], f func(A) State[S, B]) State[S, B] {
	return func(s S) (B, S) {
		a, nextS := m(s)
		return f(a)(nextS)
	}
}

// MapState transforms the result of a State computation.
func MapState[S, A, B any](m State[S, A], f func(A) B) State[S, B] {
	return func(s S) (B, S) {
		a, nextS := m(s)
		return f(a), nextS
	}
}

// Get fetches the current state as the result value.
func Get[S any]() State[S, S] {
	return func(s S) (S, S) {
		return s, s
	}
}

// Put replaces the current state with newState.
func Put[S any](newState S) State[S, struct{}] {
	return func(_ S) (struct{}, S) {
		return struct{}{}, newState
	}
}

// Modify transforms the state with a pure function.
func Modify[S any](f func(S) S) State[S, struct{}] {
	return func(s S) (struct{}, S) {
		return struct{}{}, f(s)
	}
}

// StateContext is a context-aware stateful computation: (Context, S) -> (A, S, error).
// It reconciles the pure functional State Monad with Go's standard library
// context.Context (for deadlines, timeouts, cancellation, and trace spans).
type StateContext[S, A any] func(ctx context.Context, s S) (A, S, error)

// Run executes the StateContext computation.
func (sc StateContext[S, A]) Run(ctx context.Context, s S) (A, S, error) {
	if err := ctx.Err(); err != nil {
		var zero A
		return zero, s, err
	}
	return sc(ctx, s)
}

// BindStateContext chains two context-aware state computations with early cancellation check.
func BindStateContext[S, A, B any](sc StateContext[S, A], f func(A) StateContext[S, B]) StateContext[S, B] {
	return func(ctx context.Context, s S) (B, S, error) {
		if err := ctx.Err(); err != nil {
			var zero B
			return zero, s, err
		}
		a, nextS, err := sc(ctx, s)
		if err != nil {
			var zero B
			return zero, nextS, err
		}
		return f(a)(ctx, nextS)
	}
}

// String implements fmt.Stringer for State.
func (st State[S, A]) String() string {
	return "<State Monad S -> (A, S)>"
}
