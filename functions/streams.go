package functions

import (
	"fmt"
	"iter"
	"reflect"
)

// Pair is an Algebraic Product Type that holds exactly two values.
type Pair[A, B any] struct {
	Fst A
	Snd B
}

// String formats the pair as a standard Gatwd list literal [Fst, Snd].
func (p Pair[A, B]) String() string {
	return "[" + fmt.Sprintf("%v", p.Fst) + ", " + fmt.Sprintf("%v", p.Snd) + "]"
}

// PairElements returns the constituent elements.
func (p Pair[A, B]) PairElements() (any, any) {
	return p.Fst, p.Snd
}

// Swap returns a new Pair with the components exchanged: (a, b) -> (b, a).
func (p Pair[A, B]) Swap() Pair[B, A] {
	return Pair[B, A]{Fst: p.Snd, Snd: p.Fst}
}

// Map applies f to the second element (Snd), returning a new Pair (Functor over 2nd component).
func (p Pair[A, B]) Map(f func(B) B) Pair[A, B] {
	return Pair[A, B]{Fst: p.Fst, Snd: f(p.Snd)}
}

// MapFst applies f to the first element (Fst), returning a new Pair.
func (p Pair[A, B]) MapFst(f func(A) A) Pair[A, B] {
	return Pair[A, B]{Fst: f(p.Fst), Snd: p.Snd}
}

// Bimap applies f to Fst and g to Snd, returning a new Pair (Bifunctor).
func (p Pair[A, B]) Bimap(f func(A) A, g func(B) B) Pair[A, B] {
	return Pair[A, B]{Fst: f(p.Fst), Snd: g(p.Snd)}
}

// Fold reduces the pair by applying f to both components.
func (p Pair[A, B]) Fold(f func(A, B) any) any {
	return f(p.Fst, p.Snd)
}

// Equal returns true if both components are structurally equal.
func (p Pair[A, B]) Equal(other Pair[A, B]) bool {
	return reflect.DeepEqual(p.Fst, other.Fst) && reflect.DeepEqual(p.Snd, other.Snd)
}

// Zip combines two lazy sequences into a single sequence of Pairs.
//
// If one sequence is shorter than the other, Zip stops when the shorter 
// sequence is exhausted. This is mathematically equivalent to a "convolution" 
// of two lists.
func Zip[A, B any](seqA iter.Seq[A], seqB iter.Seq[B]) iter.Seq[Pair[A, B]] {
	return func(yield func(Pair[A, B]) bool) {
		nextA, stopA := iter.Pull(seqA)
		defer stopA()
		nextB, stopB := iter.Pull(seqB)
		defer stopB()

		for {
			a, okA := nextA()
			if !okA {
				return
			}
			b, okB := nextB()
			if !okB {
				return
			}
			if !yield(Pair[A, B]{Fst: a, Snd: b}) {
				return
			}
		}
	}
}

// ZipWith behaves like Zip, but instead of returning Pairs, it applies a 
// combiner function `f(a, b)` to merge the elements into a new value.
func ZipWith[A, B, C any](seqA iter.Seq[A], seqB iter.Seq[B], f func(A, B) C) iter.Seq[C] {
	return func(yield func(C) bool) {
		nextA, stopA := iter.Pull(seqA)
		defer stopA()
		nextB, stopB := iter.Pull(seqB)
		defer stopB()

		for {
			a, okA := nextA()
			if !okA {
				return
			}
			b, okB := nextB()
			if !okB {
				return
			}
			if !yield(f(a, b)) {
				return
			}
		}
	}
}

// Take evaluates only the first `n` elements of a lazy sequence.
//
// This is ABSOLUTELY CRITICAL for infinite sequences. Because `iter.Seq` 
// can loop forever (e.g. an infinite state machine), `Take` provides the 
// boundary to prevent infinite loops.
func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		count := 0
		for v := range seq {
			if count >= n {
				return
			}
			if !yield(v) {
				return
			}
			count++
		}
	}
}

// Drop ignores the first `n` elements of a sequence and yields the rest.
func Drop[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		count := 0
		for v := range seq {
			if count < n {
				count++
				continue
			}
			if !yield(v) {
				return
			}
		}
	}
}
