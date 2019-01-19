package functions

import (
	"fmt"
	d "github.com/JoergReinhardt/godeep/data"
	"testing"
)

func TestDefinition(t *testing.T) {
	def := newFuncDef(
		12,
		"TestFuncName",
		d.Record.Flag(),
		Value,
		PreFix, Lazy, Right_Bound, Imutable,
		newConstant(d.New("test string Constant")),
		newFlag(11, Record, d.String.Flag()),
		newPair(d.New("first key"), newFlag(12, Value, d.String.Flag())),
		newPair(d.New("second key"), newFlag(12, Value, d.String.Flag())),
		newPair(d.New("third key"), newFlag(12, Value, d.String.Flag())),
	)

	fmt.Println(def)

	pos := newFuncDef(
		12,
		"TestFuncName",
		d.Record.Flag(),
		Value,
		PreFix, Lazy, Right_Bound, Imutable,
		newConstant(d.New("test string Constant")),
		newFlag(11, UnaryFnc, d.Function.Flag()),
		newFlag(11, Value, d.String.Flag()),
		newFlag(11, Value, d.String.Flag()),
		newFlag(11, Value, d.String.Flag()),
	)

	fmt.Println(pos)
}
