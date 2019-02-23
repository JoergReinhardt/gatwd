package parse

import (
	"strings"

	f "github.com/JoergReinhardt/gatwd/functions"
	l "github.com/JoergReinhardt/gatwd/lex"
)

type doFnc func(Lexer) f.StateFnc

type Lexer func() (*LineBuffer, *TokenBuffer, string)

func (lex Lexer) Buffer() (*LineBuffer, *TokenBuffer, string) { return lex() }
func (lex Lexer) LineBuffer() *LineBuffer                     { l, _, _ := lex(); return l }
func (lex Lexer) TokenBuffer() *TokenBuffer                   { _, t, _ := lex(); return t }
func (lex Lexer) CurrentLine() string                         { _, _, c := lex(); return c }
func (lex Lexer) nextState(do doFnc) f.StateFnc               { return do(lex) }
func newLexer(l *LineBuffer, t *TokenBuffer, curl string) Lexer {
	return func() (*LineBuffer, *TokenBuffer, string) { return l, t, curl }
}

// lexer gets called, whenever linebuffer changes & runs til buffer is depleted
func NewLexer(lbuf *LineBuffer) *TokenBuffer {
	var bytes = []byte{}
	lbuf.ReadLine(&bytes)
	// allocate new token buffer
	var tbuf = NewTokenBuffer()
	// enclose both buffers in a lexer instance
	var lex = newLexer(lbuf, tbuf, string(bytes))
	// retrieve state function from lexer closure over buffers
	var sf = lex.nextState(func(l Lexer) f.StateFnc {
		return lexer(lex)
	})
	// subscribe state functions run method to be called back once per
	// change in line buffer
	lbuf.Subscribe(sf.Run)
	// return token buffer reference
	return tbuf
}

// case expression matcher
func UtfPrefix(line string) bool {
	return strings.HasPrefix(line, l.UniCharString)
}
func AsciiPrefix(line string) bool {
	return strings.HasPrefix(line, l.AsciiString)
}
func DigitPrefix(line string) bool {
	return strings.HasPrefix(line, l.DigitString)
}
func KeyWordPrefix(line string) bool {
	return strings.HasPrefix(line, l.KeyWordString)
}

func lexer(lex Lexer) f.StateFnc {

	var do doFnc
	var lbuf, tbuf, curl = lex()

	// if currentline is zero
	if len(curl) == 0 {
		// try read a(-nother) line
		var bytes = []byte{}
		i, err := lbuf.ReadLine(&bytes)
		// error or empty read → buffer depleted
		if err != nil || i <= 0 {
			return nil
		}
		// else set new current line
		curl = string(bytes)
	}

	// return next state based on currently trailing characters
	switch {
	case UtfPrefix(curl):
		do = consumeUtf
	case AsciiPrefix(curl):
		do = consumeAscii
	case DigitPrefix(curl):
		do = consumeDigits
	case KeyWordPrefix(curl):
		do = consumeKeyword
	default:
		do = consumeLetters
	}

	// build new lexer closure
	lex = newLexer(lbuf, tbuf, curl)
	// retrieve and return next state fnc
	return lex.nextState(do)
}

func consumeUtf(lex Lexer) f.StateFnc {
	var do doFnc
	var lbuf, tbuf, curl = lex()

	for _, utf := range l.UniChars {
		if strings.HasPrefix(curl, utf) {
			item, _ := l.MatchUtf8(utf)
			tbuf.Append(NewSyntaxToken(tbuf.CurrentPos(), item))
			curl = strings.TrimPrefix(curl, utf)
		}
	}

	do = lexer

	lex = newLexer(lbuf, tbuf, curl)
	return lex.nextState(do)
}

func consumeAscii(lex Lexer) f.StateFnc {
	var do doFnc
	var lbuf, tbuf, curl = lex()

	for _, asc := range l.Ascii {
		if strings.HasPrefix(curl, asc) {
			item, _ := l.MatchItem(asc)
			tbuf.Append(NewSyntaxToken(
				tbuf.CurrentPos(),
				item))
			curl = strings.TrimPrefix(curl, asc)
		}
	}

	do = lexer

	lex = newLexer(lbuf, tbuf, curl)
	return lex.nextState(do)
}

func consumeKeyword(lex Lexer) f.StateFnc {
	var do doFnc
	var lbuf, tbuf, curl = lex()

	for _, keyword := range l.Keywords {
		if strings.HasPrefix(curl, keyword) {
			tbuf.Append(NewKeywordToken(
				tbuf.CurrentPos(),
				keyword))
			curl = strings.TrimPrefix(curl, keyword)
		}
	}

	do = lexer

	lex = newLexer(lbuf, tbuf, curl)
	return lex.nextState(do)
}

func consumeDigits(lex Lexer) f.StateFnc {
	var do doFnc
	var lbuf, tbuf, curl = lex()
	var digits = []byte{}

	for DigitPrefix(curl) {
		digits = append(digits, curl[0])
	}

	tbuf.Append(NewDigitToken(
		tbuf.CurrentPos(),
		string(digits)))

	do = lexer

	lex = newLexer(lbuf, tbuf, curl)
	return lex.nextState(do)
}

func consumeLetters(lex Lexer) f.StateFnc {
	var do doFnc
	var lbuf, tbuf, curl = lex()
	var letters = []rune{}
	var runes = []rune(curl)

	for strings.ContainsAny(string(runes[0]),
		l.LetterString+l.CapitalString) {

		letters = append(letters, runes[0])
		if len(runes) > 1 {
			curl = string(runes[1:])
			runes = []rune(curl)
		} else {
			curl = ""
			runes = runes[:0]
			break
		}
	}

	tbuf.Append(NewWordToken(
		tbuf.CurrentPos(),
		string(letters)))

	do = lexer

	lex = newLexer(lbuf, tbuf, curl)
	return lex.nextState(do)
}
