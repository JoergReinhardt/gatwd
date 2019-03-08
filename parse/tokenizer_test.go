package parse

import (
	"fmt"
	l "github.com/joergreinhardt/gatwd/lex"
	"testing"
)

var line = `
this is a test line withour 123123 anything interesting, exept one comma and a full stop.
data 'this would be the second line, starting with a keyword, followed by quoted text'
now for a little stuff to replace \y \F === ... >>> why is there a linebreak after each special character!?... we'll see, or not \xo .
that's enougth for now.
`

func TestNewSyntaxToken(t *testing.T) {
	fmt.Printf("token: %v\n", NewSyntaxToken(l.GetUtf8Item(string([]rune("\n")))))
}
func TestAsciiKeysSortedByLength(t *testing.T) {
	for _, key := range l.AsciiKeysSortedByLength {
		fmt.Printf("ascii keys sorted by length: %s\n", string(key))
	}
}

var buffer, queue, state = NewLexer()

var replacer = l.NewAsciiReplacer()

func writeLine() {
	n, err := buffer.WriteString(line)
	if err != nil {
		fmt.Printf(
			"error while writeing line to buffer: %s\nbytes written: %d\n",
			err, n)
	}
}
func TestPop(t *testing.T) {
	writeLine()
	var ru = pop(buffer, queue)
	var r = string(ru)
	fmt.Printf("popped rune:\n%s\n", r)
	ru = pop(buffer, queue)
	r = string(ru)
	fmt.Printf("popped rune:\n%s\n", r)
}
func TestBackup(t *testing.T) {
	backup(buffer)
	fmt.Printf("buffer:\n%s\n", buffer.BufferVal.String())
}
func TestLexer(t *testing.T) {
	state.Run()
	fmt.Printf("queue after state has been run: %s\n", queue.DataSlice.String())
}
