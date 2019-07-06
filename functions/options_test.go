package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var truthcase = NewCase(test, nil)

var falsetest = NewTestTruth("false", func(args ...Expression) bool {
	return false
})
var falsecase = NewCase(falsetest, nil)

var generic = NewGeneric(func(args ...Expression) Expression {
	var str string
	for n, arg := range args {
		str = str + arg.String()
		if n < len(args)-1 {
			str = str + " "
		}
	}
	return NewNative(d.StrVal(str))
}, "Generic", NewNative(d.NewNull(d.String)))

var genericcase = NewCase(test, generic)

var partialcase = NewCase(test, DefinePartial("To String",
	generic, NewNative(d.NewNull(d.String)), NewNative(d.TyNat(d.String|d.Int|d.Float))))

//NewNative(d.TyNat(d.String|d.Int|d.Float))
func TestCase(t *testing.T) {

	var result, ok = truthcase(New(42))
	fmt.Printf("truth case type-name: %s result: %s ok: %t\n\n",
		truthcase.TypeName(), result, ok)

	result, ok = genericcase(New(42.23))
	fmt.Printf("generic case type-name: %s result: %s ok: %t\n\n",
		genericcase.TypeName(), result, ok)

	result, ok = partialcase(New("string"))
	fmt.Printf("partial case type-name: %s result: %s ok: %t\n\n",
		partialcase.TypeName(), result, ok)

	result, ok = partialcase(New(true))
	fmt.Printf("partial case type-name: %s result: %s ok: %t\n\n",
		partialcase.TypeName(), result, ok)
}

func TestSwitch(t *testing.T) {

	var swi = NewSwitch(falsecase, truthcase, partialcase, genericcase)

	var result, pair, ok = swi(New(23), New(42.23))

	fmt.Printf("switch int & float result: %s, (current,args): %s, ok: %t\n",
		result, pair, ok)

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

	fmt.Printf("cases: %s\n", swi.Cases())
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
