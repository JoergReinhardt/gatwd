package functions

// Maybe represents an Algebraic Sum Type of Just(T) | None.
//
// In functional programming, Maybe is a safe alternative to nil pointers.
// It explicitly models the potential absence of a value.
type Maybe[T any] struct {
	value T
	valid bool
}

// Just constructs a Maybe containing a value.
func Just[T any](v T) Maybe[T] {
	return Maybe[T]{value: v, valid: true}
}

// None constructs an empty Maybe.
func None[T any]() Maybe[T] {
	return Maybe[T]{valid: false}
}

// IsJust returns true if the Maybe contains a value.
func (m Maybe[T]) IsJust() bool {
	return m.valid
}

// Unwrap returns the value and a boolean indicating if it was present.
func (m Maybe[T]) Unwrap() (T, bool) {
	return m.value, m.valid
}

// MapMaybe applies a function to the value inside a Maybe, returning a new Maybe.
// This makes Maybe a Functor.
func MapMaybe[T, U any](m Maybe[T], fn func(T) U) Maybe[U] {
	if !m.valid {
		return None[U]()
	}
	return Just(fn(m.value))
}

// BindMaybe chains a function that returns a Maybe onto an existing Maybe.
// This makes Maybe a Monad.
func BindMaybe[T, U any](m Maybe[T], fn func(T) Maybe[U]) Maybe[U] {
	if !m.valid {
		return None[U]()
	}
	return fn(m.value)
}

// Either represents an Algebraic Sum Type of Left(L) | Right(R).
//
// It is often used for error handling where Right is the success value
// and Left is the error.
type Either[L, R any] struct {
	left  *L
	right *R
}

// Left constructs an Either containing the left value.
func Left[L, R any](l L) Either[L, R] {
	return Either[L, R]{left: &l}
}

// Right constructs an Either containing the right value.
func Right[L, R any](r R) Either[L, R] {
	return Either[L, R]{right: &r}
}

// IsLeft returns true if the Either contains the left value.
func (e Either[L, R]) IsLeft() bool {
	return e.left != nil
}

// IsRight returns true if the Either contains the right value.
func (e Either[L, R]) IsRight() bool {
	return e.right != nil
}

// UnwrapLeft returns the left value and a boolean indicating its presence.
func (e Either[L, R]) UnwrapLeft() (L, bool) {
	if e.left != nil {
		return *e.left, true
	}
	var zero L
	return zero, false
}

// UnwrapRight returns the right value and a boolean indicating its presence.
func (e Either[L, R]) UnwrapRight() (R, bool) {
	if e.right != nil {
		return *e.right, true
	}
	var zero R
	return zero, false
}
