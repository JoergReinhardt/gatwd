package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var addInts = Define(Dat(func(args ...d.Native) d.Native {
	var a, b = args[0].(d.IntVal), args[1].(d.IntVal)
	return a + b
}),
	DefSym("+"),
	Def(Dat(0).Type()),
	Def(Dat(0).Type(), Dat(0).Type()),
)

func TestExpression(t *testing.T) {
	fmt.Printf(`
 defines expression to perform addition on integers of type data/Native.

 - should not take any arguments except d.IntVal
   - first argument is a symbol definition expressing the name
   - second argument is the return type (derived from instance)
   - third argument is the arguments types (derived from instances) wrapped
     by a composed type
 - should return a partial, when only one argument is passed
 - should return atomic integer result, when two args are passed
 - should return a vector of resulting integers‥.
 - ‥.where the last element might be a partialy applyed addition, if an odd
   number of arguments has been passed` + "\n\n")
	fmt.Printf("add ints expression definition type: %s\n"+
		"type-ident: %s\ntype-args: %s\nreturn type: %s\n",
		addInts.Type(), addInts.TypeId(),
		addInts.TypeArgs(), addInts.TypeRet())

	fmt.Printf("addInts: %s argtype : %s identype: %s, retype: %s\n",
		addInts,
		addInts.Type().TypeArgs(),
		addInts.Type().TypeId(),
		addInts.Type().TypeRet())

	var wrong = addInts.Call(Dat("string one"), Dat(true))
	fmt.Printf("called with argument of wrong type: %s\n", wrong)
	if !wrong.Type().Match(None) {
		t.Fail()
	}

	var partial = addInts.Call(Dat(23))
	fmt.Printf("partial: %s argtype : %s identype: %s, retype: %s\n",
		partial,
		partial.Type().TypeArgs(),
		partial.Type().TypeId(),
		partial.Type().TypeRet())

	fmt.Printf("manno ey: %s\n", partial.Type())
	if !partial.Type().Match(Def(Partial, DefSym("+"))) {
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
		result2, result2.Type().TypeArgs(),
		result2.Type().TypeId(),
		result2.Type().TypeRet())
	fmt.Printf("result2: %s\n", result2)
	if vec, ok := result2.(VecVal); ok {
		if vec.Len() != 2 {
			t.Fail()
		}
	}

	var result3 = addInts.Call(Dat(23), Dat(42), Dat(23))
	fmt.Printf("result3 type: %s type-ident: %s type-return: %s type-arguments: %s\n",
		result3.Type().String(),
		result3.Type().TypeId(),
		result3.Type().TypeRet(),
		result3.Type().TypeArgs(),
	)

	fmt.Printf("result3: %s\n", result3)
	fmt.Printf("result3 element 0 type: %s\n", result3.(VecVal)()[0].Type())
	if vec, ok := result3.(VecVal); ok {
		if !vec()[0].Type().Match(Data) {
			t.Fail()
		}
	}

	complete = result3.(VecVal)()[1].Call(Dat(42))
	fmt.Printf("completed result3[1] partial: %s\n", complete)
	if complete.(DatConst).Eval().(d.Numeral).Int() != 65 {
		t.Fail()
	}

	var result4 = addInts.Call(
		Dat(23), Dat(42), Dat(23), Dat(42),
	)
	fmt.Printf("result4: %s\n", result4)
	if vec, ok := result4.(VecVal); ok {
		if !vec()[0].Type().Match(Dat(0).Type()) {
			t.Fail()
		}
	}
}

func TestTuple(t *testing.T) {
	var con = NewTupleType(
		Def(Def(Data, Constant), d.Int),
		Def(Def(Data, Constant), d.Float),
		Def(Def(Data, Constant), d.Bool),
		Def(Def(Data, Constant), d.Bool),
	)
	fmt.Printf("tuple constructor %s\n", con)

	var tup = con.Call(Dat(23), Dat(42.23), Dat(true), Dat(false))
	fmt.Printf("tuple %s\n", tup)
	if tup.(TupVal)[0].(NatEval).Eval() != d.IntVal(23) ||
		tup.(TupVal)[1].(NatEval).Eval() != d.FltVal(42.23) ||
		tup.(TupVal)[2].(NatEval).Eval() != d.BoolVal(true) {
		t.Fail()
	}

	tup = con.Call(Dat(23), Dat(42.23), Dat(true), Dat(false))
	fmt.Printf("tuple second call %s\n", tup)
	if tup.(TupVal)[0].(NatEval).Eval() != d.IntVal(23) ||
		tup.(TupVal)[1].(NatEval).Eval() != d.FltVal(42.23) ||
		tup.(TupVal)[2].(NatEval).Eval() != d.BoolVal(true) {
		t.Fail()
	}

	var partial = con.Call(Dat(23))
	fmt.Printf("partial0: %s partial0 type-fnc: %s\n", partial, partial.TypeFnc().TypeName())
	fmt.Printf("partial0 type: %s ident: %s\n", partial.Type(), partial.Type().TypeId())

	var partial1 = partial.Call(Dat(42.23), Dat(true), Dat(false))
	fmt.Printf("partial1: %s partial1 type-fnc: %s\n", partial1, partial1.TypeFnc().TypeName())
	fmt.Printf("partial1 type: %s\n", partial1.Type())

}

var (
	recfield = NewRecordType(
		NewKeyPair("zero int", Dat(0).Type()),
		NewKeyPair("one uint", Dat(uint(0)).Type()),
		NewKeyPair("two float", Dat(float64(0.0)).Type()),
		NewKeyPair("three bool", Dat(bool(false)).Type()),
	)

	rec = recfield(
		NewKeyPair("zero int", Dat(1)),
		NewKeyPair("one uint", Dat(uint(2))),
		NewKeyPair("two float", Dat(float64(3.0))),
		NewKeyPair("three bool", Dat(bool(true))),
	)
)

func TestRecord(t *testing.T) {
	fmt.Printf("record definition:%s \n", recfield)
	fmt.Printf("record def type:%s \n", recfield.Type())

	var rec = recfield(
		NewKeyPair("zero int", Dat(1)),
		NewKeyPair("one uint", Dat(uint(2))),
		NewKeyPair("two float", Dat(float64(3.0))),
		NewKeyPair("three bool", Dat(bool(true))),
	)
	fmt.Printf("record value APPLYED CORRECTLY:%s \n", rec)
	fmt.Printf("record val type:%s \n", rec.Type())
	fmt.Printf("record val length:%d \n", rec.Len())

	rec = recfield(
		NewKeyPair("zero int", Dat(1)),
		NewKeyPair("one uint", Dat(uint(2))),
		NewKeyPair("two float", Dat(float64(3.0))),
		NewKeyPair("three bool", Dat(bool(true))),
	)
	fmt.Printf("record value CALLED CORRECTLY:%s \n", rec)
	fmt.Printf("record val type:%s \n", rec.Type())

	rec = recfield(
		NewKeyPair("zero int", Dat(1)),
		NewKeyPair("one uint", Dat(int(2))),
		NewKeyPair("two float", Dat(float64(3.0))),
		NewKeyPair("three bool", Dat(bool(true))),
	)
	fmt.Printf("record value applyed INCORRECTLY:%s \n", rec)
	fmt.Printf("record val type:%s \n", rec.Type())
}

func TestRecordGet(t *testing.T) {
}
