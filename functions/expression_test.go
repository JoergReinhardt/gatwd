package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var addInts = ConstructExpressionType(DecNative(func(args ...d.Native) d.Native {
	var a, b = args[0].(d.IntVal), args[1].(d.IntVal)
	return a + b
}), Def(
	DecNative(0).Type(),
	DecNative(0).Type(),
),
	DecNative(0).Type(),
	DefSym("AddInts"))

func TestExpression(t *testing.T) {

	fmt.Printf("addInts: %s argtype : %s identype: %s, retype: %s\n",
		addInts, addInts.Type().TypeArguments(),
		addInts.Type().TypeIdent(),
		addInts.Type().TypeReturn())

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
	fmt.Printf("result2: %s argtype : %s identype: %s, retype: %s\n",
		result2, result2.Type().TypeArguments(),
		result2.Type().TypeIdent(),
		result2.Type().TypeReturn())
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

func TestTuple(t *testing.T) {
	var tup = ConstructTupleType(
		DecNative(0).Type(),
		DecNative(0.0).Type(),
		DecNative("").Type(),
	)
	fmt.Printf("tuple: %s\n", tup)
	var tupval = tup(DecNative(23), DecNative(42.23), DecNative("string"))
	fmt.Printf("tuple value: %s type: %s\n", tupval, tupval.Type())

}

func TestNamedTuple(t *testing.T) {
	var ntup = ConstructTupleType(
		DefSym("Named Tuple"),
		DecNative(0).Type(),
		DecNative(0.0).Type(),
		DecNative("").Type(),
	)
	fmt.Printf("named tuple: %s name: %s\n", ntup, ntup.Symbol())
	var tupval = ntup(DecNative(23), DecNative(42.23), DecNative("string"))
	fmt.Printf("tuple value: %s type: %s\n", tupval, tupval.Type())
}

func TestDeclaredTuple(t *testing.T) {
	var tup = ConstructTupleType(
		DecNative(0).Type(),
		DecNative(0.0).Type(),
		DecNative("").Type(),
	).Declare()
	fmt.Printf("tuple: %s\n", tup)
	var partial = tup.Call(DecNative(23))
	fmt.Printf("partial: %s\n", partial)
	partial = partial.Call(DecNative(42.23))
	fmt.Printf("partial: %s\n", partial)
	partial = partial.Call(DecNative("string"))
	fmt.Printf("partial: %s\n", partial)
}
