package parser

import (
	"iter"
	"strconv"
	"strings"

	"github.com/joergreinhardt/gatwd/functions"
)

// Number parses a sequence of digits into an integer.
func Number() Combinator[int] {
	isDigit := func(b byte) bool { return b >= '0' && b <= '9' }
	return Bind(Many(Satisfy(isDigit)), func(bytes []byte) Combinator[int] {
		return func(input string) iter.Seq[functions.Pair[int, string]] {
			return func(yield func(functions.Pair[int, string]) bool) {
				if len(bytes) == 0 {
					return
				}
				val, err := strconv.Atoi(string(bytes))
				if err == nil {
					yield(functions.Pair[int, string]{Fst: val, Snd: input})
				}
			}
		}
	})
}

// ParseTake parses "take <number>"
func ParseTake() Combinator[int] {
	return Bind(Token("take"), func(_ string) Combinator[int] {
		return Bind(WhiteSpace, func(_ []byte) Combinator[int] {
			return Number()
		})
	})
}

// ParseDrop parses "drop <number>"
func ParseDrop() Combinator[int] {
	return Bind(Token("drop"), func(_ string) Combinator[int] {
		return Bind(WhiteSpace, func(_ []byte) Combinator[int] {
			return Number()
		})
	})
}

// ParseZipperOp parses Comonadic movement commands
func ParseZipperOp() Combinator[string] {
	return Or(
		Token("up"),
		Or(
			Token("down"),
			Or(
				Token("left"),
				Or(
					Token("right"),
					Token("extract"),
				),
			),
		),
	)
}

// ParseFanOut parses the Kleisli arrow "&&&" between two identifiers
func ParseFanOut() Combinator[functions.Pair[string, string]] {
	isAlpha := func(b byte) bool { return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') }
	wordParser := Map(Many(Satisfy(isAlpha)), func(b []byte) string { return string(b) })

	return Bind(Token("("), func(_ string) Combinator[functions.Pair[string, string]] {
		return Bind(Token(""), func(_ string) Combinator[functions.Pair[string, string]] {
			return Bind(wordParser, func(leftAdapter string) Combinator[functions.Pair[string, string]] {
				return Bind(Token("&&&"), func(_ string) Combinator[functions.Pair[string, string]] {
					return Bind(Token(""), func(_ string) Combinator[functions.Pair[string, string]] {
						return Bind(wordParser, func(rightAdapter string) Combinator[functions.Pair[string, string]] {
							return Bind(Token(")"), func(_ string) Combinator[functions.Pair[string, string]] {
								return func(input string) iter.Seq[functions.Pair[functions.Pair[string, string], string]] {
									return func(yield func(functions.Pair[functions.Pair[string, string], string]) bool) {
										yield(functions.Pair[functions.Pair[string, string], string]{
											Fst: functions.Pair[string, string]{Fst: leftAdapter, Snd: rightAdapter},
											Snd: input,
										})
									}
								}
							})
						})
					})
				})
			})
		})
	})
}

// StringLiteral parses a quoted string (e.g. "Security Violation")
func StringLiteral() Combinator[string] {
	return Bind(Token("\""), func(_ string) Combinator[string] {
		return func(input string) iter.Seq[functions.Pair[string, string]] {
			return func(yield func(functions.Pair[string, string]) bool) {
				idx := strings.Index(input, "\"")
				if idx != -1 {
					yield(functions.Pair[string, string]{Fst: input[:idx], Snd: input[idx+1:]})
				}
			}
		}
	})
}

// ParseExpect parses the expect validation gate: expect "must_be_running" "Not Running!"
func ParseExpect() Combinator[functions.Pair[string, string]] {
	return Bind(Token("expect"), func(_ string) Combinator[functions.Pair[string, string]] {
		return Bind(StringLiteral(), func(keyword string) Combinator[functions.Pair[string, string]] {
			return Bind(Token(""), func(_ string) Combinator[functions.Pair[string, string]] { // space
				return Bind(StringLiteral(), func(errMsg string) Combinator[functions.Pair[string, string]] {
					return func(input string) iter.Seq[functions.Pair[functions.Pair[string, string], string]] {
						return func(yield func(functions.Pair[functions.Pair[string, string], string]) bool) {
							yield(functions.Pair[functions.Pair[string, string], string]{
								Fst: functions.Pair[string, string]{Fst: keyword, Snd: errMsg},
								Snd: input,
							})
						}
					}
				})
			})
		})
	})
}

// ParseZip parses "zip <AdapterName>"
func ParseZip() Combinator[string] {
	isAlpha := func(b byte) bool { return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') }
	wordParser := Map(Many(Satisfy(isAlpha)), func(b []byte) string { return string(b) })

	return Bind(Token("zip"), func(_ string) Combinator[string] {
		return Bind(WhiteSpace, func(_ []byte) Combinator[string] {
			return wordParser
		})
	})
}
