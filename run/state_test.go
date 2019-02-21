package run

import (
	"fmt"
	"testing"

	l "github.com/JoergReinhardt/gatwd/lex"
	p "github.com/JoergReinhardt/gatwd/parse"
)

func TestStateFncProgress(t *testing.T) {
	var count int
	var sf StateFnc
	sf = func() StateFnc {
		count = count + 1
		if count == 10 {
			return nil
		}
		return sf
	}
	sf.Run()
	fmt.Println(count)
	if count != 10 {
		t.Fail()
	}
}
func TestQueue(t *testing.T) {
	var q = NewQueue()
	q.Put(p.NewDataValueToken(0, "one"))
	q.Put(p.NewDataValueToken(1, "two"))
	q.Put(p.NewDataValueToken(2, "three"))
	q.Put(p.NewDataValueToken(3, "four"))
	q.Put(p.NewDataValueToken(4, "five"))
	q.Put(p.NewDataValueToken(5, "six"))
	q.Put(p.NewDataValueToken(6, "seven"))

	var str string
	for q.HasToken() {
		tok := q.Pull()
		fmt.Println(tok)
		str = str + " " + tok.String()
	}
	fmt.Println(str)
	if str != ` "one" "two" "three" "four" "five" "six" "seven"` {
		t.Fail()
	}
}

var line = []rune(`\y => -> === :: \n ab\tcd 123 12 data`)

func TestReplaceDigraphs(t *testing.T) {
	l, p := replaceLine(line, 6)
	fmt.Println(string(l))
	fmt.Println(p)
	l, p = replaceLine(line, 0)
	fmt.Println(string(l))
	fmt.Println(p)
	l, p = replaceLine(line, len(line)-1)
	fmt.Println(string(l))
	fmt.Println(p)
	fmt.Println(len(line) - 1)
	fmt.Println(len(l) - 1)

	fmt.Println(string(digraphs(l...)))
}
func TestReplaceUnicode(t *testing.T) {
	uc := l.UniChars()
	fmt.Println(uc)
}
func TestThreadsafeSource(t *testing.T) {
	source := NewSource()
	source.Append([]byte(string(line)))
	fmt.Println(source)
	source.Delete(3)
	fmt.Println(source)
	source.InsertSlice(8, 10, []byte(string(line)))
	fmt.Println(source)
	source.Cut(5, 30)
	fmt.Println(source)
}
func TestThreadsafeTokens(t *testing.T) {
	toks := NewTokens()
	fmt.Println(toks)
	toks.Append(
		p.NewDataValueToken(0, "this"),
		p.NewDataValueToken(4, "is"),
		p.NewDataValueToken(6, "a"),
		p.NewDataValueToken(7, "public"),
		p.NewDataValueToken(13, "service"),
		p.NewDataValueToken(20, "annauncement"),
		p.NewDataValueToken(32, "‥."),
		p.NewDataValueToken(34, "and"),
		p.NewDataValueToken(37, "this"),
		p.NewDataValueToken(41, "is"),
		p.NewDataValueToken(43, "not"),
		p.NewDataValueToken(46, "a"),
		p.NewDataValueToken(47, "test!"),
	)
	fmt.Println(toks)
	toks.Delete(5)
	fmt.Println(toks)
	fmt.Println(toks.Range(4, 10))
	toks.Insert(5, toks.Tokens())
	fmt.Println(toks)
	toks.Sort()
	idx := toks.Search(3)
	fmt.Println(idx)
	fmt.Println(toks.Get(idx))
	fmt.Println(toks.Get(23))
}
