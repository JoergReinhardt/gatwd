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
				if len(bytes) == 0 { return }
				val, err := strconv.Atoi(string(bytes))
				if err == nil {
					yield(functions.Pair[int, string]{Fst: val, Snd: input})
				}
			}
		}
	})
}

func ParseTake() Combinator[int] {
	return Bind(Token("take"), func(_ string) Combinator[int] {
		return Bind(WhiteSpace, func(_ []byte) Combinator[int] { return Number() })
	})
}

func ParseDrop() Combinator[int] {
	return Bind(Token("drop"), func(_ string) Combinator[int] {
		return Bind(WhiteSpace, func(_ []byte) Combinator[int] { return Number() })
	})
}

func ParseZipperOp() Combinator[string] {
	return Or(Token("up"), Or(Token("down"), Or(Token("left"), Or(Token("right"), Token("extract")))))
}

// ParseFanOut parses the Kleisli arrow "(AdapterA &&& AdapterB)"
func ParseFanOut() Combinator[functions.Pair[string, string]] {
	// Parse everything until &&&
	isContentLeft := func(b byte) bool { return b != '&' }
	wordParserLeft := Map(Many(Satisfy(isContentLeft)), func(b []byte) string { return strings.TrimSpace(string(b)) })
	
	// Parse everything until )
	isContentRight := func(b byte) bool { return b != ')' }
	wordParserRight := Map(Many(Satisfy(isContentRight)), func(b []byte) string { return strings.TrimSpace(string(b)) })

	return Bind(Token("("), func(_ string) Combinator[functions.Pair[string, string]] {
		return Bind(Token(""), func(_ string) Combinator[functions.Pair[string, string]] {
			return Bind(wordParserLeft, func(leftAdapter string) Combinator[functions.Pair[string, string]] {
				return Bind(Token("&&&"), func(_ string) Combinator[functions.Pair[string, string]] {
					return Bind(Token(""), func(_ string) Combinator[functions.Pair[string, string]] {
						return Bind(wordParserRight, func(rightAdapter string) Combinator[functions.Pair[string, string]] {
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

func ParseExpect() Combinator[functions.Pair[string, string]] {
	return Bind(Token("expect"), func(_ string) Combinator[functions.Pair[string, string]] {
		return Bind(StringLiteral(), func(keyword string) Combinator[functions.Pair[string, string]] {
			return Bind(Token(""), func(_ string) Combinator[functions.Pair[string, string]] {
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

func ParseZip() Combinator[string] {
	isAlpha := func(b byte) bool { return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') }
	wordParser := Map(Many(Satisfy(isAlpha)), func(b []byte) string { return string(b) })

	return Bind(Token("zip"), func(_ string) Combinator[string] {
		return Bind(WhiteSpace, func(_ []byte) Combinator[string] {
			return wordParser
		})
	})
}

// ParseMatchCase parses a match branch of the form "Pattern -> TargetPipeline".
func ParseMatchCase() Combinator[functions.Pair[string, string]] {
	isPatternChar := func(b byte) bool { return b != '-' && b != '\n' }
	patternParser := Map(Many(Satisfy(isPatternChar)), func(b []byte) string { return strings.TrimSpace(string(b)) })

	return Bind(patternParser, func(pattern string) Combinator[functions.Pair[string, string]] {
		return Bind(Token("->"), func(_ string) Combinator[functions.Pair[string, string]] {
			return func(input string) iter.Seq[functions.Pair[functions.Pair[string, string], string]] {
				return func(yield func(functions.Pair[functions.Pair[string, string], string]) bool) {
					target := strings.TrimSpace(input)
					yield(functions.Pair[functions.Pair[string, string], string]{
						Fst: functions.Pair[string, string]{Fst: pattern, Snd: target},
						Snd: "",
					})
				}
			}
		})
	})
}

// ParseSplit parses the arrow split combinator "[ Left *** Right ]".
func ParseSplit() Combinator[functions.Pair[string, string]] {
	isLeftChar := func(b byte) bool { return b != '*' }
	leftParser := Map(Many(Satisfy(isLeftChar)), func(b []byte) string { return strings.TrimSpace(string(b)) })

	isRightChar := func(b byte) bool { return b != ']' }
	rightParser := Map(Many(Satisfy(isRightChar)), func(b []byte) string { return strings.TrimSpace(string(b)) })

	return Bind(Token("["), func(_ string) Combinator[functions.Pair[string, string]] {
		return Bind(leftParser, func(left string) Combinator[functions.Pair[string, string]] {
			return Bind(Token("***"), func(_ string) Combinator[functions.Pair[string, string]] {
				return Bind(rightParser, func(right string) Combinator[functions.Pair[string, string]] {
					return Bind(Token("]"), func(_ string) Combinator[functions.Pair[string, string]] {
						return func(input string) iter.Seq[functions.Pair[functions.Pair[string, string], string]] {
							return func(yield func(functions.Pair[functions.Pair[string, string], string]) bool) {
								yield(functions.Pair[functions.Pair[string, string], string]{
									Fst: functions.Pair[string, string]{Fst: left, Snd: right},
									Snd: input,
								})
							}
						}
					})
				})
			})
		})
	})
}
