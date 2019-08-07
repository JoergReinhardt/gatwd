package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var addInts = DecExpression(FuncVal(func(args ...Expression) Expression {
	if len(args) > 0 {
		var (
			a = args[0].(Native).Eval().(d.IntVal)
			b = args[1].(Native).Eval().(d.IntVal)
		)
		return DecNative(a + b)
	}
	return Def(Def(
		DecNative(0).Type(),
		DecNative(0).Type(),
	),
		DecNative(0).Type(),
		DefSym("AddInts"),
	)
}),
	Def(
		DecNative(0).Type(),
		DecNative(0).Type(),
	),
	DecNative(0).Type(),
	DefSym("AddInts"),
)

func TestExpression(t *testing.T) {

	var wrong = addInts.Call(DecNative("string one"), DecNative(true))
	fmt.Printf("called with argument of wrong type: %s\n", wrong)
	if !wrong.Type().Match(None) {
		t.Fail()
	}

	var partial = addInts.Call(DecNative(23))
	fmt.Printf("partial: %s argtype : %s identype: %s, retype: %s\n",
		partial, partial.Type().TypeArguments(),
		partial.Type().TypeIdent(),
		partial.Type().TypeReturn())
	if !partial.Type().TypeReturn().Match(DefSym("AddInts")) {
		t.Fail()
	}

	var wrongpart = partial.Call(DecNative("string"))
	fmt.Printf("partial called with argument of wrong type: %s\n", wrongpart)
	if !wrongpart.Type().Match(None) {
		t.Fail()
	}

	var complete = partial.Call(DecNative(42))
	fmt.Printf("complete: %s\n", complete)
	if data, ok := complete.(Native); ok {
		if num, ok := data.Eval().(d.IntVal); ok {
			if num.Int() != 65 {
				t.Fail()
			}
		}
	}

	var result2 = addInts.Call(DecNative(23), DecNative(42))
	fmt.Printf("result2: %s\n", result2)
	if vec, ok := result2.(VecType); ok {
		if vec.Len() != 2 {
			t.Fail()
		}
	}

	var result3 = addInts.Call(DecNative(23), DecNative(42), DecNative(23))
	fmt.Printf("result3: %s\n", result3)
	if vec, ok := result3.(VecType); ok {
		if !vec()[1].Type().TypeReturn().Match(DefSym("AddInts")) {
			t.Fail()
		}
	}

	var result4 = addInts.Call(DecNative(23), DecNative(42),
		DecNative(23), DecNative(42))
	fmt.Printf("result4: %s\n", result4)
	if vec, ok := result4.(VecType); ok {
		if !vec.Head().Type().MatchArgs(DecNative(0)) {
			t.Fail()
		}
	}
}
