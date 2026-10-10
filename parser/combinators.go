package parser

import (
	"iter"
	"strings"

	"github.com/joergreinhardt/gatwd/functions"
)

// Parser is a monadic function following the classic Wadler "List of Successes" method.
// It takes an input string and returns a lazy iter.Seq of possible successful parses.
// Each success contains the parsed value A, and the remaining unparsed string.
// If it fails, it simply yields nothing (an empty sequence).
//
// By using iter.Seq, we eat our own dogfood! We can use our generic `Bind`, `Map`, 
// and `Filter` primitives to build a world-class Parser Combinator!
type Combinator[A any] func(string) iter.Seq[functions.Pair[A, string]]

// Bind chains two parsers together. If the first succeeds, its result is passed
// to the function `f` to generate the next parser, which runs on the remaining input.
func Bind[A, B any](p Combinator[A], f func(A) Combinator[B]) Combinator[B] {
	return func(input string) iter.Seq[functions.Pair[B, string]] {
		// Use the core functions.Bind over the lazy sequence of successes!
		return functions.Bind(p(input), func(res functions.Pair[A, string]) iter.Seq[functions.Pair[B, string]] {
			return f(res.Fst)(res.Snd)
		})
	}
}

// Or returns a parser that tries p1, and if it fails, yields the results of p2.
func Or[A any](p1, p2 Combinator[A]) Combinator[A] {
	return func(input string) iter.Seq[functions.Pair[A, string]] {
		return func(yield func(functions.Pair[A, string]) bool) {
			matched := false
			for res := range p1(input) {
				matched = true
				if !yield(res) {
					return
				}
			}
			if !matched {
				for res := range p2(input) {
					if !yield(res) {
						return
					}
				}
			}
		}
	}
}

// Map transforms the result of a parser without consuming more input.
func Map[A, B any](p Combinator[A], f func(A) B) Combinator[B] {
	return func(input string) iter.Seq[functions.Pair[B, string]] {
		return func(yield func(functions.Pair[B, string]) bool) {
			for res := range p(input) {
				if !yield(functions.Pair[B, string]{Fst: f(res.Fst), Snd: res.Snd}) {
					return
				}
			}
		}
	}
}

// --- Primitive Combinators ---

// MatchString parses a specific literal string.
func MatchString(s string) Combinator[string] {
	return func(input string) iter.Seq[functions.Pair[string, string]] {
		return func(yield func(functions.Pair[string, string]) bool) {
			if strings.HasPrefix(input, s) {
				yield(functions.Pair[string, string]{Fst: s, Snd: input[len(s):]})
			}
		}
	}
}

// Satisfy parses a single character if it matches a predicate.
func Satisfy(predicate func(byte) bool) Combinator[byte] {
	return func(input string) iter.Seq[functions.Pair[byte, string]] {
		return func(yield func(functions.Pair[byte, string]) bool) {
			if len(input) > 0 && predicate(input[0]) {
				yield(functions.Pair[byte, string]{Fst: input[0], Snd: input[1:]})
			}
		}
	}
}

// Many parses 0 or more occurrences of p.
func Many[A any](p Combinator[A]) Combinator[[]A] {
	return func(input string) iter.Seq[functions.Pair[[]A, string]] {
		return func(yield func(functions.Pair[[]A, string]) bool) {
			var results []A
			rem := input
			
			// Greedy consume
			for {
				matched := false
				for res := range p(rem) {
					results = append(results, res.Fst)
					rem = res.Snd
					matched = true
					break // Only take the first success of p (deterministic)
				}
				if !matched {
					break
				}
			}
			yield(functions.Pair[[]A, string]{Fst: results, Snd: rem})
		}
	}
}

// WhiteSpace consumes spaces and tabs.
var WhiteSpace = Many(Satisfy(func(b byte) bool { return b == ' ' || b == '\t' }))

// Token strips leading whitespace before matching a string.
func Token(s string) Combinator[string] {
	return Bind(WhiteSpace, func(_ []byte) Combinator[string] {
		return MatchString(s)
	})
}
