package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var truth = NewTruthTest(func(args ...Expression) bool {
	var num = args[0].Eval().(d.IntVal)
	return num == 0
})

var trinary = NewTrinaryTest(func(args ...Expression) int {
	var num = args[0].Eval().(d.IntVal)
	if num < 0 {
		return -1
	}
	if num > 0 {
		return 1
	}
	return 0
})

var compare = NewCompareTest(func(args ...Expression) int {
	var num = args[0].Eval().(d.IntVal)
	if num < 0 {
		return -1
	}
	if num > 0 {
		return 1
	}
	return 0
})

func TestTruth(t *testing.T) {
	fmt.Printf("truth truth type 1, 0: %s, %s\n", truth(New(0)), truth(New(1)))

	if truth(New(0)) != True {
		t.Fail()
	}
	if truth(New(1)) != False || truth(New(-1)) != False {
		t.Fail()
	}

	fmt.Printf("test truth bool 1, 0, -1: %t, %t, %t\n",
		truth.Test(New(1)), truth.Test(New(0)), truth.Test(New(-1)))

	if !truth.Test(New(0)) {
		t.Fail()
	}
	if truth.Test(New(1)) || truth.Test(New(-1)) {
		t.Fail()
	}

	fmt.Printf("compare truth int -1, 0, 1: %d, %d, %d\n",
		truth.Compare(New(-1)), truth.Compare(New(0)), truth.Compare(New(1)))

	if truth.Compare(New(0)) != 0 {
		t.Fail()
	}
	if truth.Compare(New(1)) != -1 || truth.Compare(New(-1)) != -1 {
		t.Fail()
	}

	fmt.Printf("trinary truth truth type: -1, 0, 1: %s %s %s\n",
		trinary(New(-1)), trinary(New(0)), trinary(New(1)))

	if trinary(New(-1)) != False {
		t.Fail()
	}
	if trinary(New(0)) != Undecided {
		t.Fail()
	}
	if trinary(New(1)) != True {
		t.Fail()
	}

	fmt.Printf("test trinary truth bool type: -1, 0, 1: %t %t %t\n",
		trinary.Test(New(-1)), trinary.Test(New(0)), trinary.Test(New(1)))

	if trinary.Test(New(-1)) {
		t.Fail()
	}
	if trinary.Test(New(0)) {
		t.Fail()
	}
	if !trinary.Test(New(1)) {
		t.Fail()
	}

	fmt.Printf("compare trinary truth int type: -1, 0, 1: %d %d %d\n",
		trinary.Compare(New(-1)), trinary.Compare(New(0)), trinary.Compare(New(1)))
	if trinary.Compare(New(-1)) != -1 {
		t.Fail()
	}
	if trinary.Compare(New(0)) != 0 {
		t.Fail()
	}
	if trinary.Compare(New(1)) != 1 {
		t.Fail()
	}

	fmt.Printf("compare order type: -1, 0, 1: %s %s %s\n",
		compare(New(-1)), compare(New(0)), compare(New(1)))
	if compare(New(-1)) != Lesser {
		t.Fail()
	}
	if compare(New(0)) != Equal {
		t.Fail()
	}
	if compare(New(1)) != Greater {
		t.Fail()
	}

	fmt.Printf("compare int type: -1, 0, 1: %d %d %d\n",
		compare.Compare(New(-1)), compare.Compare(New(0)), compare.Compare(New(1)))
	if compare.Compare(New(-1)) != -1 {
		t.Fail()
	}
	if compare.Compare(New(0)) != 0 {
		t.Fail()
	}
	if compare.Compare(New(1)) != 1 {
		t.Fail()
	}

	fmt.Printf("test compare bool type: -1, 0, 1: %t %t %t\n",
		compare.Test(New(-1)), compare.Test(New(0)), compare.Test(New(1)))
	if compare.Test(New(-1)) {
		t.Fail()
	}
	if !compare.Test(New(0)) {
		t.Fail()
	}
	if compare.Test(New(1)) {
		t.Fail()
	}
}

var test = NewTruthTest(func(args ...Expression) bool {
	for _, arg := range args {
		if !arg.(NativeExpr).TypeNat().Match(
			d.String | d.Integers | d.Float) {
			return false
		}
	}
	return true
})

func TestTruthTest(t *testing.T) {

	fmt.Printf("test name: %s\n", test.TypeName())

	var result = test(New(42))
	fmt.Printf("test integer (expect True): %s\n", result)
	if result != True {
		t.Fail()
	}

	result = test(New(42.23))
	fmt.Printf("test float (expect True): %s\n", result)
	if result != True {
		t.Fail()
	}

	result = test(New("string"))
	fmt.Printf("test string (expect True): %s\n", result)
	if result != True {
		t.Fail()
	}

	result = test(New(true))
	fmt.Printf("test bool (expect False): %s\n", result)
	if result != False {
		t.Fail()
	}
}

var truthcase = NewCase(test)

var generic = NewGeneric(func(args ...Expression) Expression {
	var str string
	for n, arg := range args {
		str = str + arg.String()
		if n < len(args)-1 {
			str = str + " "
		}
	}
	return NewNative(d.StrVal(str))
})

var genericcase = NewCase(test, generic)

var defcase = NewCase(test, DefineExprType("To String",
	generic, New(d.String|d.Int|d.Float), New(d.String)))

func TestCase(t *testing.T) {

	var result, ok = truthcase(New(42))
	fmt.Printf("truth case type-name: %s result: %s ok: %t\n\n",
		truthcase.TypeName(), result, ok)

	result, ok = genericcase(New(42.23))
	fmt.Printf("generic case type-name: %s result: %s ok: %t\n\n",
		genericcase.TypeName(), result, ok)

	result, ok = defcase(New("string"))
	fmt.Printf("defined case type-name: %s result: %s ok: %t\n\n",
		defcase.TypeName(), result, ok)

	result, ok = defcase(New(true))
	fmt.Printf("defined case type-name: %s result: %s ok: %t\n\n",
		defcase.TypeName(), result, ok)
}

func TestSwitch(t *testing.T) {

	var swi = NewSwitch(defcase, truthcase, genericcase)

	fmt.Printf("switch int & float argument: %s\n",
		swi.Call(New(23), New(42.23)))

	if val := swi.Call(New(42), New(42.23)); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("successfull call passing int: %s\n",
		swi.Call(New(42)))

	if val := swi.Call(New(42)); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("successfull call passing float: %s\n",
		swi.Call(New(23.42)))
	if val := swi.Call(New(23.42)); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("successfull call passing string: %s\n",
		swi.Call(New("string")))
	if val := swi.Call(New(23.42)); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("successfull call passing multiple integers: %s\n",
		swi.Call(New(23), New(42), New(65)))
	if val := swi.Call(New(23), New(42),
		New(65)); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("successfull call passing mixed args: %s\n",
		swi.Call(New(23), New(42.23), New("string")))
	if val := swi.Call(New(23), New(42.23),
		New("string")); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("unsuccessfull call to Switch passing boolean: %s\n\n",
		swi.Call(New(true)))
	if val := swi.Call(New(true)); !val.TypeFnc().Match(None) {
		t.Fail()
	}
}

func TestMaybe(t *testing.T) {

	//	var maybe = NewMaybe(NewCase(NewTruthTest(func(args ...Callable) bool {
	//		if len(args) > 0 {
	//			return args[0].TypeNat().Match(d.String)
	//		}
	//		return false
	//	}), NewUnary(func(arg Callable) Callable { return arg })))
	//
	//	fmt.Printf("maybe: %s\n", maybe)
	//
	//	var str = maybe(New("string"))
	//	fmt.Printf("str: %s str type name: %s\n", str, str.TypeName())
	//	if str.String() != "string" {
	//		t.Fail()
	//	}
	//
	//	var none = maybe(New(1))
	//	fmt.Printf("none: %s none type name: %s\n", none, none.TypeName())
	//
	//	if none.TypeFnc() != None {
	//		t.Fail()
	//	}
}

func TestEither(t *testing.T) {
	//
	//	var either = NewEither(NewCase(
	//		NewTruthTest(func(args ...Callable) bool {
	//			if len(args) > 0 {
	//				return args[0].TypeNat().Match(d.String)
	//			}
	//			return false
	//		})))
	//
	//	fmt.Printf("either: %s either type name: %s\n", either, either.TypeName())
	//
	//	var str = either(New("string"))
	//	fmt.Printf("str: %s str type name: %s fnc type: %s nat type: %s\n",
	//		str, str.TypeName(), str.TypeFnc(), str.TypeNat().TypeName())
	//	if str.String() != "string" {
	//		t.Fail()
	//	}
	//
	//	var err = either(New(1))
	//	fmt.Printf("err: %s err type name: %s type nat: %s\n", err, err.TypeName(), err.TypeNat())
	//	if err.Eval().(d.ErrorVal).E.Error() != "error: 1" {
	//		t.Fail()
	//	}
}
