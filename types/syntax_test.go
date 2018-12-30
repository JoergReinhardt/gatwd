package types

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestTokenTypes(t *testing.T) {
	var i uint
	var typ TokType = 1
	for i = 0; i < uint(len(syntax))-1; i++ {
		typ = 1 << i
		fmt.Printf("index:\t%d\tString:\t%s\t\tSyntax:\t%s\n", i, typ.String(), typ.Syntax())
	}
}
func TestTypeRegistry(t *testing.T) {
	initTypeDef()
	var td Nodular = conTypeDef(List.Flag(), "TestType", tok_underscore.Flag(), tok_rightArrow.Flag())
	spew.Dump(td)
	fmt.Printf("%s\n", td)
	td = td.(*typeDef).Next().(*chainSigNode)
	spew.Dump(td)
}
