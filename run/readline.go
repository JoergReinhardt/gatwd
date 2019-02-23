package run

import (
	"bytes"
	"io"
	"log"
	"sort"
	"strings"
	"sync"

	d "github.com/JoergReinhardt/gatwd/data"
	l "github.com/JoergReinhardt/gatwd/lex"
	p "github.com/JoergReinhardt/gatwd/parse"
	"github.com/gohxs/readline"
)

// QUEUE
type QueueVal d.MutexVal

func (q *QueueVal) HasToken() bool {
	q.Lock()
	defer q.Unlock()
	var length = q.Native.(d.DataSlice).Len()
	if length > 0 {
		return true
	}
	return false
}
func (q *QueueVal) Put(tok p.Token) {
	q.Lock()
	defer q.Unlock()
	(*q).Native = d.SlicePut(q.Native.(d.DataSlice), tok)
}
func (q *QueueVal) Pull() p.Token {
	q.Lock()
	defer q.Unlock()
	var nat d.Native
	nat, (*q).Native = d.SlicePull(q.Native.(d.DataSlice))
	return nat.(p.Token)
}
func (q *QueueVal) Peek() p.Token {
	q.Lock()
	defer q.Unlock()
	var nat d.Native
	nat = q.Native.(d.DataSlice).GetInt(0)
	return nat.(p.Token)
}
func (q *QueueVal) PeekN(n int) p.Token {
	q.Lock()
	defer q.Unlock()
	var nat d.Native
	nat = q.Native.(d.DataSlice).GetInt(n)
	return nat.(p.Token)
}
func NewQueue() *QueueVal {
	return &QueueVal{
		sync.Mutex{},
		d.DataSlice{},
	}
}

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
func (s *LineBuffer) SetClean() {
	s.Lock()
	defer s.Unlock()
	s.Clean = true
}
func (s *LineBuffer) setDirty() {
	s.Clean = false
}
func (s *LineBuffer) byteVec() *d.ByteVec {
	return s.Native.(*d.ByteVec)
}
func (s *LineBuffer) bytes() []byte {
	s.setDirty()
	return []byte(*s.byteVec())
}
func (s *LineBuffer) Bytes() []byte {
	s.Lock()
	defer s.Unlock()
	return s.bytes()
}
func (s *LineBuffer) Len() int {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	return s.byteVec().Len()
}
func (s *LineBuffer) Append(p []byte) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	*(s.byteVec()) = append(s.bytes(), p...)
}
func (s *LineBuffer) WriteRunes(r []rune) {
	s.WriteString(string(r))
}
func (s *LineBuffer) WriteString(str string) {
	s.Append([]byte(str))
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
func (s *LineBuffer) Cut(i, j int) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	(s.byteVec()).Cut(i, j)
}
func (s *LineBuffer) Delete(i int) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	(s.byteVec()).Delete(i)
}
func (s *LineBuffer) Get(i int) byte {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	return byte((s.byteVec()).Get(d.IntVal(i)).(d.ByteVal))
}
func (s *LineBuffer) Range(i, j int) []byte {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	return []byte((s.byteVec()).Range(i, j))
}
func (s *LineBuffer) UpdateTrailing(line []rune) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	var bytes = []byte(string(line))
	var buflen = len(s.bytes())
	var trailen = len(bytes)
	if buflen >= trailen {
		var end = buflen - 1
		var start = end - trailen - 1
		copy(([]byte(*s.Native.(*d.ByteVec)))[start:end], bytes)
	}
}

////////////////////////////////////////////////////
type Tokens d.AsyncVal

func NewTokens() Tokens {
	return Tokens(d.AsyncVal{
		sync.Mutex{},
		true,
		d.DataSlice{},
	})
}
func (s Tokens) String() string {
	s.Lock()
	defer s.Unlock()
	var str = bytes.NewBuffer([]byte{})
	var l = len(s.dataSlice())
	for i, tok := range toks(s.slice()...) {
		str.WriteString(tok.String())
		if i < l-1 {
			str.WriteString("\n")
		}
	}

	return str.String()
}
func (s *Tokens) dataSlice() d.DataSlice {
	return s.Native.(d.DataSlice)
}
func (s *Tokens) slice() []d.Native {
	return s.dataSlice().Slice()
}
func (s *Tokens) setDirty() {
	s.Clean = false
}
func (s *Tokens) SetClean() {
	s.Lock()
	defer s.Unlock()
	s.Clean = true
}
func (s *Tokens) Len() int {
	s.Lock()
	defer s.Unlock()
	return s.dataSlice().Len()
}
func (s *Tokens) Tokens() []p.Token {
	s.Lock()
	defer s.Unlock()
	return toks(s.slice()...)
}
func (s *Tokens) Get(i int) p.Token {
	s.Lock()
	defer s.Unlock()
	return s.dataSlice().GetInt(i).(p.Token)
}
func (s *Tokens) Range(i, j int) []p.Token {
	s.Lock()
	defer s.Unlock()
	var toks = []p.Token{}
	for _, dat := range s.dataSlice()[i:j] {
		toks = append(toks, dat.(p.Token))
	}
	return toks
}
func (s *Tokens) Split(i int) (h, t []p.Token) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	var head, tail = d.SliceSplit(s.dataSlice(), i)
	return toks(head...), toks(tail...)
}
func (s *Tokens) Set(i int, tok p.Token) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	s.dataSlice().SetInt(i, tok)
}
func (s *Tokens) Append(toks ...p.Token) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	s.Native = d.SliceAppend(s.dataSlice().Slice(), nats(toks...)...)
}
func (s *Tokens) Insert(i int, toks []p.Token) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	s.Native = d.SliceInsertVector(s.dataSlice(), i, nats(toks...)...)
}
func (s *Tokens) Delete(i int) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	s.Native = d.SliceDelete(s.dataSlice(), i)
}
func (t *Tokens) Sort() {
	t.Lock()
	defer t.Unlock()
	var ts = tokSort(toks([]d.Native(t.Native.(d.DataSlice))...))
	sort.Sort(ts)
	t.Native = d.DataSlice(nats(ts...))
}
func (t *Tokens) Search(pos int) int {
	return sort.Search(t.Len(), func(i int) bool {
		return pos < t.dataSlice().Slice()[i].(p.Token).Pos()
	})
}

//////
func nats(toks ...p.Token) []d.Native {
	var nats = []d.Native{}
	for _, nat := range toks {
		nats = append(nats, nat)
	}
	return nats
}
func toks(nats ...d.Native) []p.Token {
	var toks = []p.Token{}
	for _, nat := range nats {
		toks = append(toks, nat.(p.Token))
	}
	return toks
}

//////
type tokSort []p.Token

func (t tokSort) Len() int { return len(t) }
func (t tokSort) Less(i, j int) bool {
	return []p.Token(t)[i].Pos() <
		[]p.Token(t)[j].Pos()
}
func (t tokSort) Swap(i, j int) {
	[]p.Token(t)[i], []p.Token(t)[j] = []p.Token(t)[j], []p.Token(t)[i]
}

/////////////////////////////////////////////////////////
func NewReadLine() (sf StateFnc, linebuf *LineBuffer) {

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
	sf = func() StateFnc {

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

		return func() StateFnc { return sf() }
	}
	// return state function & thread-safe line buffer
	return sf, linebuf
}

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
func asclen(r []rune) int     { return len(acr.Replace(string(r))) }

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
