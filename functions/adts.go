package functions

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

// BindMaybe applies a function returning a Maybe to the value inside a Maybe.
func BindMaybe[A, B any](m Maybe[A], f func(A) Maybe[B]) Maybe[B] {
	if m.IsNothing() { return Nothing[B]() }
	val, _ := m.Unwrap()
	return f(val)
}
