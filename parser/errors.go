package parser

import (
	"fmt"
	"iter"

	"github.com/joergreinhardt/gatwd/functions"
)

// SyntaxErrorTracker records the furthest point the parser reached before failing.
// Because Combinator parsing involves backtracking (Or), the "deepest" failure
// is almost always the actual syntax error the user made.
type SyntaxErrorTracker struct {
	Expected string
	Input    string
}

// Expect wraps a Combinator and records a beautiful Elm-style error if it fails.
// It tracks the shortest remaining input, indicating the deepest parse tree traversal.
func Expect[A any](p Combinator[A], expected string, tracker *SyntaxErrorTracker) Combinator[A] {
	return func(input string) iter.Seq[functions.Pair[A, string]] {
		return func(yield func(functions.Pair[A, string]) bool) {
			matched := false
			for res := range p(input) {
				matched = true
				if !yield(res) {
					return
				}
			}

			// If the parse failed, check if this is the deepest error recorded
			if !matched {
				if tracker.Input == "" || len(input) <= len(tracker.Input) {
					tracker.Input = input
					tracker.Expected = expected
				}
			}
		}
	}
}

// FormatError pretty-prints an Elm/Rust style syntax error pointer!
func FormatError(originalInput string, tracker *SyntaxErrorTracker) string {
	if tracker.Expected == "" {
		return "Unknown Syntax Error"
	}

	consumed := len(originalInput) - len(tracker.Input)
	if consumed < 0 {
		consumed = 0
	}

	out := fmt.Sprintf("\n🔥 Syntax Error:\n\n    %s\n    ", originalInput)
	for i := 0; i < consumed; i++ {
		out += " "
	}
	out += fmt.Sprintf("^-- Expected %s here.\n", tracker.Expected)
	
	return out
}
