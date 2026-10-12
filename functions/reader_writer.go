package functions

// Reader represents a computation with access to an immutable environment R: R -> A.
// It encapsulates the canonical Haskell Reader Monad (dependency injection / configuration).
type Reader[R, A any] func(R) A

// Run evaluates the reader computation with environment r.
func (rd Reader[R, A]) Run(r R) A {
	return rd(r)
}

// ReturnReader lifts a pure value into Reader.
func ReturnReader[R, A any](val A) Reader[R, A] {
	return func(_ R) A {
		return val
	}
}

// BindReader chains two reader computations (monadic bind >>=).
func BindReader[R, A, B any](m Reader[R, A], f func(A) Reader[R, B]) Reader[R, B] {
	return func(r R) B {
		a := m(r)
		return f(a)(r)
	}
}

// MapReader transforms the result of a Reader computation.
func MapReader[R, A, B any](m Reader[R, A], f func(A) B) Reader[R, B] {
	return func(r R) B {
		return f(m(r))
	}
}

// Ask retrieves the entire environment R.
func Ask[R any]() Reader[R, R] {
	return func(r R) R {
		return r
	}
}

// Asks retrieves a projection or selector of the environment.
func Asks[R, A any](selector func(R) A) Reader[R, A] {
	return func(r R) A {
		return selector(r)
	}
}

// Local modifies the environment locally for a sub-computation.
func Local[R, A any](f func(R) R, m Reader[R, A]) Reader[R, A] {
	return func(r R) A {
		return m(f(r))
	}
}

// Writer represents a computation that produces an accumulated log W alongside value A: (A, W).
// It encapsulates the canonical Haskell Writer Monad (audit trails, telemetry, side-channel logs).
type Writer[W, A any] struct {
	Value A
	Log   W
}

// Run unwraps the writer value and log.
func (wr Writer[W, A]) Run() (A, W) {
	return wr.Value, wr.Log
}

// ReturnWriter lifts a value into Writer with an empty/initial log.
func ReturnWriter[W, A any](val A, empty W) Writer[W, A] {
	return Writer[W, A]{Value: val, Log: empty}
}

// BindWriter chains two writer computations combining logs with monoid append function.
func BindWriter[W, A, B any](m Writer[W, A], appendLog func(W, W) W, f func(A) Writer[W, B]) Writer[W, B] {
	next := f(m.Value)
	return Writer[W, B]{
		Value: next.Value,
		Log:   appendLog(m.Log, next.Log),
	}
}

// Tell emits an entry to the writer log.
func Tell[W any](log W) Writer[W, struct{}] {
	return Writer[W, struct{}]{
		Value: struct{}{},
		Log:   log,
	}
}

// Listen exposes the accumulated log to the computed value.
func Listen[W, A any](m Writer[W, A]) Writer[W, Pair[A, W]] {
	return Writer[W, Pair[A, W]]{
		Value: Pair[A, W]{Fst: m.Value, Snd: m.Log},
		Log:   m.Log,
	}
}
