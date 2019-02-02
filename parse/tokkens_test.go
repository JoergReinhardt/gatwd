package parse

import (
	"fmt"
	"testing"

	d "github.com/JoergReinhardt/godeep/data"
	f "github.com/JoergReinhardt/godeep/functions"
	l "github.com/JoergReinhardt/godeep/lex"
)

func TestTokenTypes(t *testing.T) {
	kind := NewKindToken(f.Vector)
	arg := NewArgumentToken(f.NewArgument(f.New("test argument")))
	parm := NewParameterToken(f.NewKeyValueParm(f.New("test parameter key"), f.New("test parameter value")))
	typ := NewDataTypeToken(d.String)
	dat := NewDataValueToken(f.New("test data value"))
	pair := NewPairValueToken(f.NewPair(f.New("test pair key"), f.New("test pair value")))
	col := NewTokenCollection(
		NewSyntaxToken(l.RightArrow),
		NewSyntaxToken(l.FatRArrow),
		NewSyntaxToken(l.FatLArrow),
		NewSyntaxToken(l.LeftArrow))

	fmt.Printf(
		"kind: %s\narg: %s\nparm: %s\ntype: %s\ndata: %s\npair: %s\ncol: %s\n",
		kind, arg, parm, typ, dat, pair, col)
}
func TestTokenTree(t *testing.T) {
	root := NewTokenCollection(
		NewDataValueToken(f.New("this is a test leave value of data.StrVal type")),
		NewDataValueToken(f.New("this is another test leave value of data.StrVal type")),
		NewDataValueToken(f.NewKeyValueParm(f.New("key of test parameter"),
			f.New("value of test parameter"))),
		NewTokenCollection(
			NewKeyValToken(
				f.New(
					"  first key of second level parameter\n"),
				f.New(
					"  first value of second level parameter\n"),
			),
			NewKeyValToken(
				f.New(
					"  second key of second level parameter\n"),
				f.New(
					"  second value of second level parameter\n"),
			),
			NewKeyValToken(
				f.New(
					"  third key of second level parameter\n"),
				f.New(
					"  third value of second level parameter\n"),
			),
			NewKeyValToken(
				f.New(
					"  four key of second level parameter\n"),
				f.New(
					"  four value of second level parameter\n"),
			),
			NewKeyValToken(
				f.New(
					"  this parameter contains another nested layer\n"),
				NewTokenCollection(
					NewKeyValToken(
						f.New(
							"    second layer first key of second level parameter\n"),
						f.New(
							"    second layer first value of second level parameter\n"),
					),
					NewKeyValToken(
						f.New(
							"     second layer second key of second level parameter\n"),
						f.New(
							"    second layer second value of second level parameter\n"),
					),
					NewKeyValToken(
						f.New(
							"    second layer third key of second level parameter\n"),
						f.New(
							"    second layer third value of second level parameter\n"),
					),
					NewKeyValToken(
						f.New(
							"    second layer four key of second level parameter\n"),
						f.New(
							"    second layer four value of second level parameter\n"),
					),
				),
			),
		),
	)
	fmt.Println(root)
}
