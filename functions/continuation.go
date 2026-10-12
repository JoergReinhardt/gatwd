package functions

// Cont represents a continuation-passing computation: (A -> R) -> R.
// In category theory and functional programming, Cont is the "mother of all monads",
// capable of expressing coroutines, exceptions, generators, and non-deterministic backtracking.
type Cont[R, A any] func(func(A) R) R

// Run evaluates the continuation with the final callback.
func (c Cont[R, A]) Run(k func(A) R) R {
	return c(k)
}

// ReturnCont lifts a pure value into a Cont computation.
func ReturnCont[R, A any](val A) Cont[R, A] {
	return func(k func(A) R) R {
		return k(val)
	}
}

// BindCont chains two continuation computations (monadic bind >>=).
func BindCont[R, A, B any](m Cont[R, A], f func(A) Cont[R, B]) Cont[R, B] {
	return func(k func(B) R) R {
		return m(func(a A) R {
			return f(a)(k)
		})
	}
}

// MapCont transforms the value inside a Cont computation.
func MapCont[R, A, B any](m Cont[R, A], f func(A) B) Cont[R, B] {
	return func(k func(B) R) R {
		return m(func(a A) R {
			return k(f(a))
		})
	}
}

// CallCC implements Call-With-Current-Continuation.
// It exposes an early-escape callback 'exit' to the inner function. Calling exit(val)
// instantly short-circuits the remaining pipeline and returns val to the outer boundary.
func CallCC[R, A, B any](f func(func(A) Cont[R, B]) Cont[R, A]) Cont[R, A] {
	return func(k func(A) R) R {
		return f(func(a A) Cont[R, B] {
			return func(_ func(B) R) R {
				return k(a) // bypass downstream, call top-level k directly
			}
		})(k)
	}
}
