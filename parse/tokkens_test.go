package parse

import (
	"fmt"
	"testing"

	d "github.com/JoergReinhardt/gatwd/data"
	f "github.com/JoergReinhardt/gatwd/functions"
	l "github.com/JoergReinhardt/gatwd/lex"
)

func TestTokenTypes(t *testing.T) {
	kind := NewFncTypeToken(0, f.Vector)
	arg := NewArgumentToken(0, "test argument")
	parm := NewParameterToken(0, "test parameter key", "test parameter value")
	typ := NewNatTypeToken(0, d.String)
	dat := NewDataValueToken(0, "test data value")
	pair := NewPairToken(0, "test pair key", "test pair value")
	col := NewTokenCollection(0,
		NewSyntaxToken(1, l.RightArrow),
		NewSyntaxToken(2, l.FatRArrow),
		NewSyntaxToken(3, l.FatLArrow),
		NewSyntaxToken(4, l.LeftArrow))

	fmt.Printf(
		"kind: %s\narg: %s\nparm: %s\ntype: %s\ndata: %s\npair: %s\ncol: %s\n",
		kind, arg, parm, typ, dat, pair, col)

}
func TestTokenTree(t *testing.T) {
	root := NewTokenCollection(0,
		NewDataValueToken(1, "this is a test leave value of data.StrVal type"),
		NewDataValueToken(2, "this is another test leave value of data.StrVal type"),
		NewPairToken(3, "key of test parameter", "value of test parameter"),
		NewTokenCollection(3,
			NewKeyValToken(5,
				f.New("  first key of second level parameter\n"),
				f.New("  first value of second level parameter\n"),
			),
			NewKeyValToken(6,
				f.New("  second key of second level parameter\n"),
				f.New("  second value of second level parameter\n"),
			),
			NewKeyValToken(7,
				f.New("  third key of second level parameter\n"),
				f.New("  third value of second level parameter\n"),
			),
			NewKeyValToken(8,
				f.New("  four key of second level parameter\n"),
				f.New("  four value of second level parameter\n"),
			),
			NewKeyValToken(9,
				f.New("  this parameter contains another nested layer\n"),
				NewTokenCollection(10,
					NewKeyValToken(11,
						f.New("    second layer first key of second level parameter\n"),
						f.New("    second layer first value of second level parameter\n"),
					),
					NewKeyValToken(12,
						f.New("     second layer second key of second level parameter\n"),
						f.New("    second layer second value of second level parameter\n"),
					),
					NewKeyValToken(13,
						f.New("    second layer third key of second level parameter\n"),
						f.New("    second layer third value of second level parameter\n"),
					),
					NewKeyValToken(14,
						f.New("    second layer four key of second level parameter\n"),
						f.New("    second layer four value of second level parameter\n"),
					),
				),
			),
		),
	)
	fmt.Println(root)
}
