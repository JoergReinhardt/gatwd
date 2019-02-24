package parse

import (
	"fmt"
	"testing"
)

var line = `
this is a bit of test text
let's replace something :: \y != \M
did this work out as expected?
  `

func TestTokenize(t *testing.T) {
	lbuf := *NewLineBuffer()
	(&lbuf).Write([]byte(line))
	fmt.Println(lbuf.String())
	tbuf := *NewLexer(&lbuf)

	lexer(newLexer(&lbuf, &tbuf, line))

	fmt.Println(tbuf.Tokens())
}
