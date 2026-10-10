package functions

// Compose performs mathematical function composition: f(g(x)).
// 
// In functional programming, composing small, pure functions into larger pipelines 
// is the primary way to build complex logic without writing monolithic blocks of code.
func Compose[A, B, C any](f func(B) C, g func(A) B) func(A) C {
	return func(x A) C {
		return f(g(x))
	}
}

// Curry2 takes a function of two arguments and returns a function that takes the 
// first argument and returns a function that takes the second.
// 
// This allows for "Partial Application". You can supply one argument now, and 
// supply the second argument later in the evaluation loop.
func Curry2[A, B, C any](f func(A, B) C) func(A) func(B) C {
	return func(a A) func(B) C {
		return func(b B) C {
			return f(a, b)
		}
	}
}

// Uncurry2 reverses the currying process, taking a function that returns a function
// and turning it back into a standard two-argument function.
func Uncurry2[A, B, C any](f func(A) func(B) C) func(A, B) C {
	return func(a A, b B) C {
		return f(a)(b)
	}
}

// Identity simply returns whatever is passed into it.
// 
// While trivial, Identity functions are often required in higher-order 
// functions or monads where a "do nothing" placeholder is needed mathematically.
func Identity[T any](x T) T {
	return x
}

// Const returns a function that always returns a constant value, regardless 
// of the input. Useful for masking inputs in a functional pipeline.
func Const[A, B any](b B) func(A) B {
	return func(_ A) B {
		return b
	}
}
