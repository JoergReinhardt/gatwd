package parse

import (
	"bytes"
	"sort"

	d "github.com/JoergReinhardt/gatwd/data"
)

////////////////////////////////////////////////////
type TokenBuffer d.AsyncVal

func NewTokenBuffer(callbacks ...func()) *TokenBuffer {
	return (*TokenBuffer)(d.NewAsync(callbacks...))
}
func (s TokenBuffer) AsyncVal() *d.AsyncVal { return (*d.AsyncVal)(&s) }
func (s TokenBuffer) String() string {
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
func (s *TokenBuffer) dataSlice() d.DataSlice {
	return s.Native.(d.DataSlice)
}
func (s *TokenBuffer) slice() []d.Native {
	return s.dataSlice().Slice()
}
func (s *TokenBuffer) setDirty() {
	s.Clean = false
}
func (s *TokenBuffer) SetClean() {
	s.Lock()
	defer s.Unlock()
	s.Clean = true
}
func (s *TokenBuffer) Len() int {
	s.Lock()
	defer s.Unlock()
	return s.len()
}
func (s *TokenBuffer) CurrentPos() int {
	s.Lock()
	defer s.Unlock()

	var l = s.len()
	if l > 0 {
		var lt = s.tokens()[l-1]
		// current position is last tokens start position plus it's
		// length in byte
		return lt.Pos() + len([]byte(lt.String()))
	}
	return 0
}
func (s *TokenBuffer) len() int {
	return s.dataSlice().Len()
}
func (s *TokenBuffer) tokens() []Token {
	return toks(s.slice()...)
}
func (s *TokenBuffer) Tokens() []Token {
	s.Lock()
	defer s.Unlock()
	return s.tokens()
}
func (s *TokenBuffer) Get(i int) Token {
	s.Lock()
	defer s.Unlock()
	if s.len() > 0 {
		return s.dataSlice().GetInt(i).(Token)
	}
	return TokVal{}
}
func (s *TokenBuffer) Range(i, j int) []Token {
	s.Lock()
	defer s.Unlock()
	var toks = []Token{}
	for _, dat := range s.dataSlice()[i:j] {
		toks = append(toks, dat.(Token))
	}
	return toks
}
func (s *TokenBuffer) Split(i int) (h, t []Token) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	var head, tail = d.SliceSplit(s.dataSlice(), i)
	return toks(head...), toks(tail...)
}
func (s *TokenBuffer) Set(i int, tok Token) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	if s.len() > i {
		s.dataSlice().SetInt(i, tok)
	}
}
func (s *TokenBuffer) Append(toks ...Token) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	s.Native = d.SliceAppend(s.dataSlice().Slice(), nats(toks...)...)
}
func (s *TokenBuffer) Insert(i int, toks []Token) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	if s.len() > i {
		s.Native = d.SliceInsertVector(s.dataSlice(), i, nats(toks...)...)
	}
}
func (s *TokenBuffer) Delete(i int) {
	s.Lock()
	defer s.Unlock()
	s.setDirty()
	if s.len() > i {
		s.Native = d.SliceDelete(s.dataSlice(), i)
	}
}
func (t *TokenBuffer) Sort() {
	t.Lock()
	defer t.Unlock()
	var ts = tokSort(toks([]d.Native(t.Native.(d.DataSlice))...))
	sort.Sort(ts)
	t.Native = d.DataSlice(nats(ts...))
}
func (t *TokenBuffer) Search(pos int) int {
	if t.len() > pos {
		return sort.Search(t.Len(), func(i int) bool {
			return pos < t.dataSlice().Slice()[i].(Token).Pos()
		})
	}
	return -1
}

//////
func nats(toks ...Token) []d.Native {
	var nats = []d.Native{}
	for _, nat := range toks {
		nats = append(nats, nat)
	}
	return nats
}
func toks(nats ...d.Native) []Token {
	var toks = []Token{}
	for _, nat := range nats {
		toks = append(toks, nat.(Token))
	}
	return toks
}

//////
type tokSort []Token

func (t tokSort) Len() int { return len(t) }
func (t tokSort) Less(i, j int) bool {
	return []Token(t)[i].Pos() <
		[]Token(t)[j].Pos()
}
func (t tokSort) Swap(i, j int) {
	[]Token(t)[i], []Token(t)[j] = []Token(t)[j], []Token(t)[i]
}
