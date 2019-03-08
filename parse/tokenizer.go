package parse

import (
	f "github.com/joergreinhardt/gatwd/functions"
)

// lexer error returns an error, reflecting the rune that was lexed and the
// error that was returned.
type LexerState f.StateFnc

func NewLexerState(r rune)
