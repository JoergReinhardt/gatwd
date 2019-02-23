package run

import (
	"io"
	"log"
	"strings"

	f "github.com/JoergReinhardt/gatwd/functions"
	l "github.com/JoergReinhardt/gatwd/lex"
	p "github.com/JoergReinhardt/gatwd/parse"
	"github.com/gohxs/readline"
)

////////////////////////////////////////////////////////////////////////////
//// READLINE MONAD
///
// instanciate readline with a listener that replaces ascii di-, & trigraphs
// against unicode
func NewReadLine() (sf f.StateFnc, linebuf *p.LineBuffer) {

	// create readline config
	var config = &readline.Config{
		Prompt:                 "\033[31m»\033[0m ",
		HistoryFile:            "/tmp/readline-multiline",
		InterruptPrompt:        "^C",
		EOFPrompt:              "exit",
		DisableAutoSaveHistory: true,
	}

	linebuf = p.NewLineBuffer()

	var listener = newListener(linebuf)
	// set listener function
	config.SetListener(listener)

	// allocate readline instance
	var rl, err = readline.NewEx(config)
	if err != nil {
		panic(err)
	}
	rl.Refresh()

	log.SetOutput(rl.Stderr())

	// STATE MONAD
	sf = func() f.StateFnc {

		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				rl.Close()
				return nil
			} else {
				return sf
			}
		} else if err == io.EOF {
			rl.Close()
			return nil
		}

		return func() f.StateFnc { return sf() }
	}
	// return state function & thread-safe line buffer
	return sf, linebuf
}

////////////////////////////////////////////////////////////////////////////
//// LISTENER FUNCTION
///
// the listener replaces di-, & trigraph representations special characters
// with their true unicode representation at every keystroke. the character
// currently under the cursor will stay expadet, until cursor position
// progresses to fully revel it in the part of the line preceeding, or folling
// the cursor. once revealed, it will be replaced instantly.

type listenerFnc func([]rune, int, rune) ([]rune, int, bool)

func newListener(linebuf *p.LineBuffer) listenerFnc {

	// word boundary characters as string
	var boundary = strings.Join(l.UniChars(), "")

	return func(line []rune, pos int, key rune) ([]rune, int, bool) {

		var head, tail []rune

		if line == nil {
			line = []rune(" ")
			return line, 0, true
		}
		if len(line) == 0 {
			line = []rune(" ")
		}

		switch {
		// cursor at start of line
		case pos == 0:
			line = uni(line)
		// cursor at end of line
		case pos >= len(line)-1:
			line = uni(line)
			pos = len(line) - 1
		// cursor somewhere inbetween
		default:
			head = uni(line[:pos])
			tail = uni(line[pos:])
			var runes = asc([]rune{key})
			if len(runes) > 1 {
				head = append(head, runes[0])
				tail = append(runes[1:], tail...)
			}
			pos = len(head)
			line = append(head, tail...)
		}

		// on word-boundary update trail in buffer to trigger lexer
		if strings.ContainsAny(string(key), boundary) {
			linebuf.UpdateTrailing(line)
		}

		return line, pos, true
	}
}

//// PRE DEFINED STRING/RUNE REPLACER
///
// replaces digtaphs with unicode
func uni(runes []rune) []rune { return []rune(acr.Replace(string(runes))) }
func asclen(r []rune) int     { return len(acr.Replace(string(r))) + 1 }

var acr = l.NewAsciiReplacer()

// replaces unicode with digtaphs
func asc(runes []rune) []rune { return []rune(ucr.Replace(string(runes))) }
func unilen(r []rune) int     { return len(ucr.Replace(string(r))) }

var ucr = l.NewUnicodeReplacer()
