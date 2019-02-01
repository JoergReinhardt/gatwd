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
	arg := NewArgumentToken(f.NewArgument(d.New("test argument")))
	parm := NewParameterToken(f.NewKeyValueParm(d.New("test parameter key"), d.New("test parameter value")))
	typ := NewDataTypeToken(d.String)
	dat := NewDataValueToken(d.New("test data value"))
	pair := NewPairValueToken(f.NewPair(d.New("test pair key"), d.New("test pair value")))
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
		NewDataValueToken(d.New("this is a test leave value of data.StrVal type")),
		NewDataValueToken(d.New("this is another test leave value of data.StrVal type")),
		NewDataValueToken(f.NewKeyValueParm(d.New("key of test parameter"),
			d.New("value of test parameter"))),
		NewTokenCollection(
			NewKeyValToken(
				f.NewValue(d.New(
					"  first key of second level parameter\n")),
				f.NewValue(d.New(
					"  first value of second level parameter\n")),
			),
			NewKeyValToken(
				f.NewValue(d.New(
					"  second key of second level parameter\n")),
				f.NewValue(d.New(
					"  second value of second level parameter\n")),
			),
			NewKeyValToken(
				f.NewValue(d.New(
					"  third key of second level parameter\n")),
				f.NewValue(d.New(
					"  third value of second level parameter\n")),
			),
			NewKeyValToken(
				f.NewValue(d.New(
					"  four key of second level parameter\n")),
				f.NewValue(d.New(
					"  four value of second level parameter\n")),
			),
			NewKeyValToken(
				f.NewValue(d.New(
					"  this parameter contains another nested layer\n")),
				NewTokenCollection(
					NewKeyValToken(
						f.NewValue(d.New(
							"    second layer first key of second level parameter\n")),
						f.NewValue(d.New(
							"    second layer first value of second level parameter\n")),
					),
					NewKeyValToken(
						f.NewValue(d.New(
							"     second layer second key of second level parameter\n")),
						f.NewValue(d.New(
							"    second layer second value of second level parameter\n")),
					),
					NewKeyValToken(
						f.NewValue(d.New(
							"    second layer third key of second level parameter\n")),
						f.NewValue(d.New(
							"    second layer third value of second level parameter\n")),
					),
					NewKeyValToken(
						f.NewValue(d.New(
							"    second layer four key of second level parameter\n")),
						f.NewValue(d.New(
							"    second layer four value of second level parameter\n")),
					),
				),
			),
		),
	)
	fmt.Println(root)
}
