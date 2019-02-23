package parse

import (
	"fmt"
	"strconv"
	"strings"

	d "github.com/JoergReinhardt/gatwd/data"
)

//////////////////////////////////////////////////
type LineBuffer d.AsyncVal

func NewLineBuffer(callbacks ...func()) *LineBuffer {
	return (*LineBuffer)(d.NewAsync(callbacks...))
}
func (s LineBuffer) AsynVal() *d.AsyncVal   { return (*d.AsyncVal)(&s) }
func (s *LineBuffer) Subscribe(c ...func()) { (*d.AsyncVal)(s).Subscribe(c...) }
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
func (s *LineBuffer) callBack() {
	for _, call := range s.Calls {
		call()
	}
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
func (s *LineBuffer) peek() byte {
	if len(s.bytes()) > 0 {
		return s.bytes()[0]
	}
	return byte(0)
}
func (s *LineBuffer) Peek() byte {
	s.Lock()
	defer s.Unlock()
	s.setClean()

	return s.peek()
}
func (s *LineBuffer) peekN(n int) []byte {
	if len(s.bytes()) > n {
		return s.bytes()[:n]
	}
	return nil
}
func (s *LineBuffer) PeekN(n int) []byte {
	s.Lock()
	defer s.Unlock()
	s.setClean()

	return s.peekN(n)
}
func (s *LineBuffer) Read(p *[]byte) (int, error) {
	s.Lock()
	defer s.Unlock()
	s.setClean()

	return s.read(p)
}
func (s *LineBuffer) ReadString() (string, error) {
	var buf = make([]byte, 0, s.Len())
	n, err := s.Read(&buf)
	if err != nil {
		return string(buf), fmt.Errorf("error in lexer at position: %d"+
			" while trying to read string from line buffer\n"+
			"buffer content: ",
			strconv.Itoa(n),
			string(*s.Native.(*d.ByteVec)))
	}
	return string(buf), nil
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
