package transducers

// Reducer is a generic fold step: it takes an accumulated state R and a new element A,
// and returns the new accumulated state R.
type Reducer[A, R any] func(R, A) R

// Transducer is a function that transforms one Reducer into another.
// 
// In Functional Programming, Transducers allow you to mathematically compose 
// a massive pipeline of Maps and Filters independent of the data structure 
// they operate on. This completely eliminates intermediate allocations.
type Transducer[A, B, R any] func(Reducer[B, R]) Reducer[A, R]

// Map returns a Transducer that applies the function `fn` to the input element 
// before passing it to the inner Reducer.
func Map[A, B, R any](fn func(A) B) Transducer[A, B, R] {
	return func(next Reducer[B, R]) Reducer[A, R] {
		return func(acc R, item A) R {
			return next(acc, fn(item))
		}
	}
}

// Filter returns a Transducer that only passes elements to the inner Reducer 
// if they satisfy the predicate.
func Filter[A, R any](predicate func(A) bool) Transducer[A, A, R] {
	return func(next Reducer[A, R]) Reducer[A, R] {
		return func(acc R, item A) R {
			if predicate(item) {
				return next(acc, item)
			}
			return acc
		}
	}
}

// Sequence is a utility to execute a Transducer pipeline over a slice.
func Sequence[A, B any](input []A, xf Transducer[A, B, []B]) []B {
	// The ultimate inner reducer simply appends to the slice
	appendReducer := func(acc []B, item B) []B {
		return append(acc, item)
	}

	// Apply the transducer to the append reducer
	reducer := xf(appendReducer)

	// Execute the fold
	var acc []B
	for _, item := range input {
		acc = reducer(acc, item)
	}
	return acc
}
