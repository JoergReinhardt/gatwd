// Package parser provides monadic parser combinators following the Wadler "List of Successes" method.
//
// Built natively on Go 1.23 iter.Seq, Combinator[A] composes lazy streams of parsed results:
//   - Monadic Primitives: Bind, Map, Or, Satisfy, MatchString, Token, WhiteSpace.
//   - High-Level Combinators: Number, StringLiteral, ParseTake, ParseDrop, ParseZip, ParseFanOut, ParseSplit, ParseMatchCase.
//   - BitFlag AST Parser: NewParser compiles type definitions and registers BitFlag tags for runtime efficiency.
//go:generate go run github.com/joergreinhardt/gatwd/cmd/gatwd-gen ..
package parser
