package functions

import "fmt"

// Either represents a value of one of two possible types (a disjoint union).
type Either[L, R any] struct {
	left  *L
	right *R
}

func Left[L, R any](val L) Either[L, R] { return Either[L, R]{left: &val} }
func Right[L, R any](val R) Either[L, R] { return Either[L, R]{right: &val} }

func (e Either[L, R]) IsLeft() bool { return e.left != nil }
func (e Either[L, R]) IsRight() bool { return e.right != nil }

func (e Either[L, R]) UnwrapLeft() (L, bool) {
	if e.IsLeft() { return *e.left, true }
	var zero L
	return zero, false
}

func (e Either[L, R]) UnwrapRight() (R, bool) {
	if e.IsRight() { return *e.right, true }
	var zero R
	return zero, false
}

// Map applies f to the Right value if present.
func (e Either[L, R]) Map(f func(R) R) Either[L, R] {
	if e.IsRight() {
		r, _ := e.UnwrapRight()
		return Right[L, R](f(r))
	}
	return e
}

// Fold evaluates onLeft if e is Left, or onRight if e is Right.
func FoldEither[L, R, T any](e Either[L, R], onLeft func(L) T, onRight func(R) T) T {
	if e.IsLeft() {
		l, _ := e.UnwrapLeft()
		return onLeft(l)
	}
	r, _ := e.UnwrapRight()
	return onRight(r)
}

func (e Either[L, R]) String() string {
	if e.IsLeft() {
		l, _ := e.UnwrapLeft()
		return "Left(" + fmt.Sprintf("%v", l) + ")"
	}
	r, _ := e.UnwrapRight()
	return "Right(" + fmt.Sprintf("%v", r) + ")"
}

// Maybe represents an optional value (Option type).
// It safely encapsulates the concept of "null" in purely functional terms.
type Maybe[T any] struct {
	value *T
}

func Just[T any](val T) Maybe[T] { return Maybe[T]{value: &val} }
func Nothing[T any]() Maybe[T] { return Maybe[T]{value: nil} }

func (m Maybe[T]) IsJust() bool { return m.value != nil }
func (m Maybe[T]) IsNothing() bool { return m.value == nil }

func (m Maybe[T]) Unwrap() (T, bool) {
	if m.IsJust() { return *m.value, true }
	var zero T
	return zero, false
}

// UnwrapOr returns the contained value or def if Nothing.
func (m Maybe[T]) UnwrapOr(def T) T {
	if m.IsJust() { return *m.value }
	return def
}

// Map applies f to the contained value if Just.
func (m Maybe[T]) Map(f func(T) T) Maybe[T] {
	if m.IsJust() {
		return Just(f(*m.value))
	}
	return Nothing[T]()
}

// BindMaybe applies a function returning a Maybe to the value inside a Maybe.
func BindMaybe[A, B any](m Maybe[A], f func(A) Maybe[B]) Maybe[B] {
	if m.IsNothing() { return Nothing[B]() }
	val, _ := m.Unwrap()
	return f(val)
}

func (m Maybe[T]) String() string {
	if m.IsNothing() {
		return "Nothing"
	}
	val, _ := m.Unwrap()
	return "Just(" + fmt.Sprintf("%v", val) + ")"
}

// Graph exposes the internal structure for generic UTF-8 Tree visualizations
func (m Maybe[T]) Graph() any {
	if m.IsNothing() { return "<Nothing>" }
	val, _ := m.Unwrap()
	return struct {
		Just T
	}{ Just: val }
}
