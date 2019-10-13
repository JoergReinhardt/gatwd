package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var addInts = Define(Dat(func(args ...d.Native) d.Native {
	var a, b = args[0].(d.IntVal), args[1].(d.IntVal)
	return a + b
}), DefSym("AddInts"), Def(Dat(0).Type()), Def(Dat(0).Type(), Dat(0).Type()))

func TestExpression(t *testing.T) {

	fmt.Printf("addInts: %s argtype : %s identype: %s, retype: %s\n",
		addInts,
		addInts.Type().TypeArguments(),
		addInts.Type().TypeIdent(),
		addInts.Type().TypeReturn())

	var wrong = addInts.Call(Dat("string one"), Dat(true))
	fmt.Printf("called with argument of wrong type: %s\n", wrong)
	if !wrong.Type().Match(None) {
		t.Fail()
	}

	var partial = addInts.Call(Dat(23))
	fmt.Printf("partial: %s argtype : %s identype: %s, retype: %s\n",
		partial,
		partial.Type().TypeArguments(),
		partial.Type().TypeIdent(),
		partial.Type().TypeReturn())
	if !partial.Type().TypeReturn().Match(DefSym("AddInts")) {
		t.Fail()
	}

	var wrongpart = partial.Call(Dat("string"))
	fmt.Printf("partial called with argument of wrong type: %s\n", wrongpart)
	if !wrongpart.Type().Match(None) {
		t.Fail()
	}

	var complete = partial.Call(Dat(42))
	fmt.Printf("complete: %s\n", complete)
	if data, ok := complete.(NatEval); ok {
		if num, ok := data.Eval().(d.IntVal); ok {
			if num.Int() != 65 {
				t.Fail()
			}
		}
	}

	var result2 = addInts.Call(Dat(23), Dat(42))
	fmt.Printf("result2: %s argtype : %s identype: %s, retype: %s\n",
		result2, result2.Type().TypeArguments(),
		result2.Type().TypeIdent(),
		result2.Type().TypeReturn())
	fmt.Printf("result2: %s\n", result2)
	if vec, ok := result2.(VecVal); ok {
		if vec.Len() != 2 {
			t.Fail()
		}
	}

	var result3 = addInts.Call(Dat(23), Dat(42), Dat(23))
	fmt.Printf("result3: %s\n", result3)
	if vec, ok := result3.(VecVal); ok {
		if !vec()[1].Type().TypeReturn().Match(DefSym("AddInts")) {
			t.Fail()
		}
	}

	var result4 = addInts.Call(Dat(23), Dat(42),
		Dat(23), Dat(42))
	fmt.Printf("result4: %s\n", result4)
	if vec, ok := result4.(VecVal); ok {
		if !vec.Head().Type().MatchArgs(Dat(0)) {
			t.Fail()
		}
	}
}

func TestTuple(t *testing.T) {
	var con = NewTuple(
		Def(Def(Data, Constant), d.Int),
		Def(Def(Data, Constant), d.Float),
		Def(Def(Data, Constant), d.Bool))
	fmt.Printf("tuple constructor %s\n", con)

	var tup = con.Call(Dat(23), Dat(42.23), Dat(true))
	fmt.Printf("tuple %s\n", tup)
	if tup.(TupleVal).Value()[0].(NatEval).Eval() != d.IntVal(23) ||
		tup.(TupleVal).Value()[1].(NatEval).Eval() != d.FltVal(42.23) ||
		tup.(TupleVal).Value()[2].(NatEval).Eval() != d.BoolVal(true) {
		t.Fail()
	}

	tup = con.Call(Dat(23), Dat(42.23), Dat(true))
	fmt.Printf("tuple second call %s\n", tup)
	if tup.(TupleVal).Value()[0].(NatEval).Eval() != d.IntVal(23) ||
		tup.(TupleVal).Value()[1].(NatEval).Eval() != d.FltVal(42.23) ||
		tup.(TupleVal).Value()[2].(NatEval).Eval() != d.BoolVal(true) {
		t.Fail()
	}

	var partial = con.Call(Dat(23))
	fmt.Printf("partial0: %s partial0 type-fnc: %s\n", partial, partial.TypeFnc().TypeName())
	fmt.Printf("partial0 type: %s ident: %s\n", partial.Type(), partial.Type().TypeIdent())

	var partial1 = partial.Call(Dat(42.23), Dat(false))
	fmt.Printf("partial1: %s partial1 type-fnc: %s\n", partial1, partial1.TypeFnc().TypeName())
	fmt.Printf("partial1 type: %s\n", partial1.Type())

	tup = partial1.Call(Dat(true))
	fmt.Printf("result: %s\n", tup)
	if tup.(TupleVal).Value()[0].(NatEval).Eval() != d.IntVal(23) ||
		tup.(TupleVal).Value()[1].(NatEval).Eval() != d.FltVal(42.23) ||
		tup.(TupleVal).Value()[2].(NatEval).Eval() != d.BoolVal(true) {
		t.Fail()
	}
}
