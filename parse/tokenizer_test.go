package parse

import (
	"fmt"
	"testing"

	l "github.com/joergreinhardt/gatwd/lex"
)

var line = `
this is a test line withour 123123 anything interesting, exept one comma and a full stop.
data 'this would be the second line, starting with a keyword, followed by quoted text'
now for a little stuff to \cp != replace \y \F === ... >>> why is there a linebreak after each special character!??? ...ARSCHLOCH we'll see, or not \xo .
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
