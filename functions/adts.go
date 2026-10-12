package functions

import (
	"fmt"
	"reflect"
)

// Either represents a value of one of two possible types (a disjoint union).
type Either[L, R any] struct {
	left  *L
	right *R
}

// Left constructs an Either holding a value of the left type L. By convention
// the Left branch carries the "error" or alternative case in railway-oriented
// programming, so Bind and Map leave it untouched. The type parameters are L
// (the stored left type) and R (the unused right type, which must still be
// supplied so the returned Either[L, R] is fully typed).
func Left[L, R any](val L) Either[L, R] { return Either[L, R]{left: &val} }

// Right constructs an Either holding a value of the right type R. By convention
// the Right branch carries the "success" case, so it is the value that Map and
// Bind operate on. The type parameters are L (the unused left type, required so
// the result is fully typed) and R (the stored right type).
func Right[L, R any](val R) Either[L, R] { return Either[L, R]{right: &val} }

// IsLeft reports whether e holds a Left (left-typed) value.
func (e Either[L, R]) IsLeft() bool { return e.left != nil }

// IsRight reports whether e holds a Right (right-typed) value.
func (e Either[L, R]) IsRight() bool { return e.right != nil }

// UnwrapLeft extracts the Left value. It returns the stored L and true when e
// is a Left; otherwise it returns the zero value of L and false. The boolean
// lets callers distinguish a genuine Left from a defaulted zero value.
func (e Either[L, R]) UnwrapLeft() (L, bool) {
	if e.IsLeft() {
		return *e.left, true
	}
	var zero L
	return zero, false
}

// UnwrapRight extracts the Right value. It returns the stored R and true when e
// is a Right; otherwise it returns the zero value of R and false. The boolean
// lets callers distinguish a genuine Right from a defaulted zero value.
func (e Either[L, R]) UnwrapRight() (R, bool) {
	if e.IsRight() {
		return *e.right, true
	}
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

// Bind (monadic flatMap) applies f to the Right value and returns the resulting
// Either, short-circuiting unchanged when the receiver is a Left. This is the
// Monad bind for Either and the basis of railway-oriented error handling:
// a Left propagates through a chain of Binds without invoking f.
func (e Either[L, R]) Bind(f func(R) Either[L, R]) Either[L, R] {
	if e.IsRight() {
		r, _ := e.UnwrapRight()
		return f(r)
	}
	return e
}

// MapLeft applies f to the Left value if present.
func (e Either[L, R]) MapLeft(f func(L) L) Either[L, R] {
	if e.IsLeft() {
		l, _ := e.UnwrapLeft()
		return Left[L, R](f(l))
	}
	return e
}

// Bimap applies fLeft if Left, or fRight if Right.
func (e Either[L, R]) Bimap(fLeft func(L) L, fRight func(R) R) Either[L, R] {
	if e.IsLeft() {
		l, _ := e.UnwrapLeft()
		return Left[L, R](fLeft(l))
	}
	r, _ := e.UnwrapRight()
	return Right[L, R](fRight(r))
}

// Fold evaluates onLeft if e is Left, or onRight if e is Right.
func (e Either[L, R]) Fold(onLeft func(L) any, onRight func(R) any) any {
	if e.IsLeft() {
		l, _ := e.UnwrapLeft()
		return onLeft(l)
	}
	r, _ := e.UnwrapRight()
	return onRight(r)
}

// Equal returns true if both are Left with equal values or both Right with equal values.
func (e Either[L, R]) Equal(other Either[L, R]) bool {
	if e.IsLeft() && other.IsLeft() {
		l1, _ := e.UnwrapLeft()
		l2, _ := other.UnwrapLeft()
		return reflect.DeepEqual(l1, l2)
	}
	if e.IsRight() && other.IsRight() {
		r1, _ := e.UnwrapRight()
		r2, _ := other.UnwrapRight()
		return reflect.DeepEqual(r1, r2)
	}
	return false
}

// FoldEither collapses an Either[L, R] into a single value of type T by applying
// onLeft when e is a Left or onRight when e is a Right. Unlike the Either.Fold
// method, which is constrained to return any, this free function is generic in
// the result type T, so both branch functions must agree on producing a T. This
// is the catamorphism (fold) for Either: it is the canonical way to eliminate
// the sum type and continue with an ordinary value.
func FoldEither[L, R, T any](e Either[L, R], onLeft func(L) T, onRight func(R) T) T {
	if e.IsLeft() {
		l, _ := e.UnwrapLeft()
		return onLeft(l)
	}
	r, _ := e.UnwrapRight()
	return onRight(r)
}

// Swap returns a new Either with the Left and Right types exchanged.
func (e Either[L, R]) Swap() Either[R, L] {
	if e.IsLeft() {
		l, _ := e.UnwrapLeft()
		return Right[R, L](l)
	}
	r, _ := e.UnwrapRight()
	return Left[R, L](r)
}

// EitherValue returns left value, right value, and whether e is Left.
func (e Either[L, R]) EitherValue() (any, any, bool) {
	if e.IsLeft() {
		l, _ := e.UnwrapLeft()
		return l, nil, true
	}
	r, _ := e.UnwrapRight()
	return nil, r, false
}

// String renders e for debugging as "Left(<value>)" or "Right(<value>)",
// satisfying the fmt.Stringer interface.
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

// Just constructs a Maybe[T] that contains the present value val. It is the
// "some" constructor of the Option type and the counterpart to Nothing.
func Just[T any](val T) Maybe[T] { return Maybe[T]{value: &val} }

// Nothing constructs an empty Maybe[T], representing the absence of a value.
// It is the "none" constructor of the Option type and the identity element
// that Map, Bind, and Filter fall through to without invoking their functions.
func Nothing[T any]() Maybe[T] { return Maybe[T]{value: nil} }

// IsJust reports whether m contains a value.
func (m Maybe[T]) IsJust() bool { return m.value != nil }

// IsNothing reports whether m is empty.
func (m Maybe[T]) IsNothing() bool { return m.value == nil }

// Unwrap extracts the contained value. It returns the stored T and true when m
// is Just; otherwise it returns the zero value of T and false. The boolean lets
// callers safely distinguish a present value from the empty case rather than
// risking a nil dereference.
func (m Maybe[T]) Unwrap() (T, bool) {
	if m.IsJust() {
		return *m.value, true
	}
	var zero T
	return zero, false
}

// UnwrapOr returns the contained value or def if Nothing.
func (m Maybe[T]) UnwrapOr(def T) T {
	if m.IsJust() {
		return *m.value
	}
	return def
}

// MaybeValue returns the inner value (if any) and whether m is Just.
func (m Maybe[T]) MaybeValue() (any, bool) {
	if m.IsJust() {
		return *m.value, true
	}
	return nil, false
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
	if m.IsNothing() {
		return Nothing[B]()
	}
	val, _ := m.Unwrap()
	return f(val)
}

// Fold applies f to the value if Just, or returns def if Nothing.
func (m Maybe[T]) Fold(def any, f func(T) any) any {
	if m.IsJust() {
		return f(*m.value)
	}
	return def
}

// Filter returns the Maybe unchanged if predicate returns true on Just; otherwise Nothing.
func (m Maybe[T]) Filter(p func(T) bool) Maybe[T] {
	if m.IsJust() && p(*m.value) {
		return m
	}
	return Nothing[T]()
}

// Equal returns true if both are Nothing or both are Just with equal values.
func (m Maybe[T]) Equal(other Maybe[T]) bool {
	if m.IsNothing() && other.IsNothing() {
		return true
	}
	if m.IsJust() && other.IsJust() {
		return reflect.DeepEqual(*m.value, *other.value)
	}
	return false
}

// String renders m for debugging as "Just(<value>)" or "Nothing", satisfying
// the fmt.Stringer interface.
func (m Maybe[T]) String() string {
	if m.IsNothing() {
		return "Nothing"
	}
	val, _ := m.Unwrap()
	return "Just(" + fmt.Sprintf("%v", val) + ")"
}

// Graph exposes the internal structure for generic UTF-8 Tree visualizations
func (m Maybe[T]) Graph() any {
	if m.IsNothing() {
		return "<Nothing>"
	}
	val, _ := m.Unwrap()
	return struct {
		Just T
	}{Just: val}
}
