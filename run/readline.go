package run

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	d "github.com/JoergReinhardt/gatwd/data"
	f "github.com/JoergReinhardt/gatwd/functions"
	l "github.com/JoergReinhardt/gatwd/lex"
	"github.com/gohxs/readline"
)

//////////////////////////////////////////////////
type LineBuffer d.AsyncVal

func NewSource() *LineBuffer {
	return (*LineBuffer)(&d.AsyncVal{
		sync.Mutex{},
		true,
		&d.ByteVec{},
	})
}
func (s LineBuffer) String() string {
	(&s).Lock()
	defer (&s).Unlock()
	return s.byteVec().String()
}
func (s LineBuffer) Lines() []string {
	return strings.Split(s.String(), "\n")
}
func (s LineBuffer) Fields() [][]string {
	var fields = [][]string{}
	for _, line := range s.Lines() {
		fields = append(fields, strings.Fields(line))
	}
	return fields
}
func (s *LineBuffer) setClean() {
	s.Clean = true
}
func (s *LineBuffer) SetClean() {
	s.Lock()
	defer s.Unlock()
	s.setClean()
}
func (s *LineBuffer) setDirty() {
	s.Clean = false
}
func (s *LineBuffer) byteVec() *d.ByteVec {
	return s.Native.(*d.ByteVec)
}
func (s *LineBuffer) bytes() []byte {
	return []byte(*s.byteVec())
}
func (s *LineBuffer) string() string {
	return string(s.bytes())
}
func (s *LineBuffer) Bytes() []byte {
	s.Lock()
	defer s.Unlock()
	return s.bytes()
}
func (s *LineBuffer) Runes() []rune {
	return []rune(s.String())
}
func (s *LineBuffer) len() int {
	return s.byteVec().Len()
}
func (s *LineBuffer) Len() int {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	return s.len()
}
func (s *LineBuffer) read(p *[]byte) (int, error) {
	var n = cap(*p)
	if n >= 0 && n < s.len() {
		*p = append(make([]byte, 0, n), s.rang(0, n)...)
		s.cut(0, n)
		return n, nil
	}

	return 0, fmt.Errorf(
		"could not read line from buffer\n"+
			"buffer index position: %d\n"+
			"buffer: %s\n", n, s.string())
}
func (s *LineBuffer) Read(p *[]byte) (int, error) {
	s.Lock()
	defer s.Unlock()
	s.setClean()

	return s.read(p)
}

// read line reads one line from buffer & either replaces p with the bytes read
// from buffer if length of p is zero, or appends bytes read from buffer to p,
// if it's length is greater than zero
func (s *LineBuffer) ReadLine(p *[]byte) (int, error) {
	s.Lock()
	defer s.Unlock()
	s.setClean()

	var lines = strings.Split(s.string(), "\n")
	var length = len([]byte(lines[0]))
	s.cut(0, length)
	*p = []byte(lines[0])

	return length, nil
}

// writes the content of p to the underlying buffer
func (s *LineBuffer) Write(p []byte) (int, error) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	*(s.byteVec()) = append(s.bytes(), p...)
	return len(p), nil
}
func (s *LineBuffer) WriteRunes(r []rune) (int, error) {
	return s.WriteString(string(r))
}
func (s *LineBuffer) WriteString(str string) (int, error) {
	return s.Write([]byte(str))
}
func (s *LineBuffer) Insert(i, j int, b byte) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	(s.byteVec()).Insert(i, j, b)
}
func (s *LineBuffer) InsertSlice(i, j int, p []byte) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	(s.byteVec()).InsertSlice(i, j, p...)
}
func (s *LineBuffer) ReplaceSlice(i, j int, trail []byte) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	copy(([]byte(*s.Native.(*d.ByteVec)))[i:j], trail)
}
func (s *LineBuffer) Split(i int) (h, t []byte) {
	s.Lock()
	defer s.Unlock()
	var head, tail = d.SliceSplit(s.Native.(d.ByteVec).Slice(), i)
	for _, b := range head {
		h = append(h, byte(b.(d.ByteVal)))
	}
	for _, b := range tail {
		t = append(t, byte(b.(d.ByteVal)))
	}
	return h, t
}
func (s *LineBuffer) cut(i, j int) {
	(s.byteVec()).Cut(i, j)
}
func (s *LineBuffer) Cut(i, j int) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()

	s.cut(i, j)
}
func (s *LineBuffer) delete(i int) {
	(s.byteVec()).Delete(i)
}
func (s *LineBuffer) Delete(i int) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	s.delete(i)
}
func (s *LineBuffer) Get(i int) byte {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	return byte((s.byteVec()).Get(d.IntVal(i)).(d.ByteVal))
}
func (s *LineBuffer) rang(i, j int) []byte {
	return []byte((s.byteVec()).Range(i, j))
}
func (s *LineBuffer) Range(i, j int) []byte {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	return s.rang(i, j)
}
func (s *LineBuffer) UpdateTrailing(line []rune) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	var bytes = []byte(string(line))
	var buflen = len(s.bytes())
	var trailen = len(bytes)
	if buflen >= trailen {
		var end = buflen
		var start = end - trailen
		copy(([]byte(*s.Native.(*d.ByteVec)))[start:end], bytes)
	}
}

////////////////////////////////////////////////////////////////////////////
//// READLINE MONAD
///
// instanciate readline with a listener that replaces ascii di-, & trigraphs
// against unicode
func NewReadLine() (sf f.StateFnc, linebuf *LineBuffer) {

	// create readline config
	var config = &readline.Config{
		Prompt:                 "\033[31m»\033[0m ",
		HistoryFile:            "/tmp/readline-multiline",
		InterruptPrompt:        "^C",
		EOFPrompt:              "exit",
		DisableAutoSaveHistory: true,
	}

	linebuf = NewSource()

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

func newListener(linebuf *LineBuffer) listenerFnc {

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

var acr = strings.NewReplacer(digraphReplacementList()...)

func digraphReplacementList() []string {
	var acrl = []string{}
	for _, dig := range l.Digraphs() {
		acrl = append(acrl, dig)
		acrl = append(acrl, l.AsciiToUnicode(dig))
	}
	return acrl
}

// replaces unicode with digtaphs
func asc(runes []rune) []rune { return []rune(ucr.Replace(string(runes))) }
func unilen(r []rune) int     { return len(ucr.Replace(string(r))) }

var ucr = strings.NewReplacer(unicodeReplacementList()...)

func unicodeReplacementList() []string {
	var ucrl = []string{}
	for _, unc := range l.UniChars() {
		ucrl = append(ucrl, unc)
		ucrl = append(ucrl, l.UnicodeToASCII(unc))
	}
	return ucrl
}
