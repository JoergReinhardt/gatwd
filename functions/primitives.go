package functions

import "iter"

// Map applies a pure function to every element in the sequence lazily.
//
// In Category Theory, this makes iter.Seq a "Functor". 
// Instead of creating a new list in memory, it returns a new sequence
// that applies the function `fn` as values are pulled. This allows for
// safe evaluation of infinite sequences.
//
// Example:
//
//	ints := iter.Seq[int]{...}
//	strs := Map(ints, strconv.Itoa)
func Map[T, U any](seq iter.Seq[T], fn func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			if !yield(fn(v)) {
				return
			}
		}
	}
}

// Fold reduces a sequence into a single accumulated value lazily.
//
// This is the functional "reduce". It applies an accumulator function 
// `fn` to the initial value `acc` and each element of the sequence.
func Fold[T, U any](seq iter.Seq[T], acc U, fn func(U, T) U) U {
	for v := range seq {
		acc = fn(acc, v)
	}
	return acc
}

// Filter lazily yields only the elements of a sequence that satisfy the predicate.
func Filter[T any](seq iter.Seq[T], predicate func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if predicate(v) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Bind (often called FlatMap) chains dependent monadic operations lazily.
//
// In Category Theory, this is the defining operation of a "Monad" (the >>= operator).
// If you map a sequence with a function that returns ANOTHER sequence, you get a
// sequence of sequences (iter.Seq[iter.Seq[U]]). Bind applies the function and
// lazily flattens the result into a single continuous stream.
//
// Example:
//
//	hosts := iter.Seq[Host]{...}
//	getInterfaces := func(h Host) iter.Seq[Interface]
//	allInterfaces := Bind(hosts, getInterfaces)
func Bind[T, U any](seq iter.Seq[T], fn func(T) iter.Seq[U]) iter.Seq[U] {
	return func(yield func(U) bool) {
		for t := range seq {
			for u := range fn(t) {
				if !yield(u) {
					return
				}
			}
		}
	}
}
