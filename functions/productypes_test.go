package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var testIsZero = NewTest(func(args ...Expression) bool {
	for _, arg := range args {
		if arg.(NatEval).Eval().(d.Numeral).GoInt() != 0 {
			return false
		}
	}
	return true
})

func TestTestable(t *testing.T) {

	fmt.Printf("test: %s\n", testIsZero)

	fmt.Printf("test zero is zero (true): %t\n", testIsZero(Dat(0)))
	if !testIsZero(Dat(0)) {
		t.Fail()
	}

	fmt.Printf("test one is zero (false): %t\n", testIsZero(Dat(1)))
	if testIsZero(Dat(1)) {
		t.Fail()
	}

	fmt.Printf("test three zeros are zero (true): %t\n",
		testIsZero(Dat(0), Dat(0), Dat(0)))
	if !testIsZero(Dat(0), Dat(0), Dat(0)) {
		t.Fail()
	}
}

var compZero = NewComparator(func(args ...Expression) int {
	switch args[0].(NatEval).Eval().(d.Numeral).GoInt() {
	case -1:
		return -1
	case 0:
		return 0
	}
	return 1
})

func TestCompareable(t *testing.T) {
	fmt.Printf("compareable: %s\n", compZero)
	fmt.Printf("zero equals zero (0): %d\n", compZero(Dat(0)))
	fmt.Printf("minus one lesser zero (-1): %d\n", compZero(Dat(-1)))
	fmt.Printf("one greater zero (1): %d\n", compZero(Dat(1)))
}

var caseZero = NewCase(
	testIsZero,
	Dat("this is indeed zero"),
	d.Numbers,
	d.String,
)

func TestCase(t *testing.T) {
	fmt.Printf("case: %s\n", caseZero)

	var result = caseZero.Call(Dat(0))
	fmt.Printf("case zero: %s\n", result)
	if result.String() != "this is indeed zero" {
		t.Fail()
	}

	result = caseZero.Call(Dat(1))
	fmt.Printf("case none zero: %s none zero type: %s\n",
		result, result.Type())
	if !result.Type().Match(None) {
		t.Fail()
	}
}

var isInteger = NewTest(func(args ...Expression) bool {
	for _, arg := range args {
		if arg.TypeFnc().Match(Data) {
			return arg.(Native).TypeNat().Match(d.Int)
		}
	}
	return false
})
var caseInteger = NewCase(isInteger, Dat("this is an int"), d.Int, d.String)
var isUint = NewTest(func(args ...Expression) bool {
	for _, arg := range args {
		if arg.TypeFnc().Match(Data) {
			return arg.(Native).TypeNat().Match(d.Uint)
		}
	}
	return false
})
var caseUint = NewCase(isUint, Dat("this is a uint"), d.Uint, d.String)
var isFloat = NewTest(func(args ...Expression) bool {
	for _, arg := range args {
		if arg.TypeFnc().Match(Data) {
			return arg.(Native).TypeNat().Match(d.Float)
		}
	}
	return false
})
var caseFloat = NewCase(isFloat, Dat("this is a float"), d.Float, d.String)
var swi = NewSwitch(caseFloat, caseUint, caseInteger)

func TestSwitch(t *testing.T) {

	var result = swi.Call(Dat(42))
	fmt.Printf("result from calling switch on 42: %s\n", result)
	if !result.Type().MatchArgs(Dat(0)) {
		t.Fail()
	}

	swi = swi.Reload()
	result = swi.Call(Dat(uint(42)))
	fmt.Printf("result from calling switch on uint 42: %s\n", result)
	if !result.Type().MatchArgs(Dat(uint(0))) {
		t.Fail()
	}

	swi = swi.Reload()
	result = swi.Call(Dat(42.23))
	fmt.Printf("result from calling switch on uint 42.23: %s\n", result)
	if !result.Type().MatchArgs(Dat(uint(0.0))) {
		t.Fail()
	}

	swi = swi.Reload()
	result = swi.Call(Dat(true))
	fmt.Printf("result from calling switch on true: %s\n", result)
	if !result.Type().Match(None) {
		t.Fail()
	}
}

func TestMaybe(t *testing.T) {
	var maybeString = NewMaybe(caseInteger)
	var str = maybeString(Dat(42))
	fmt.Printf("string: %s\n", str)
	if str.Type().TypeReturn().Match(Def(Data, d.String)) {
		t.Fail()
	}
	var none = maybeString(Dat(true))
	fmt.Printf("none type: %s fnctype: %s\n", none.Type(), none.TypeFnc())
	if !none.Type().TypeReturn().Match(None) {
		t.Fail()
	}

	fmt.Printf("maybe string: %s, type: %s\n", maybeString, maybeString.Type())
	fmt.Printf("str: %s str-type: %s, none: %s\n", str, str.Type(), none)
}

func TestOption(t *testing.T) {
	var (
		option   = NewEitherOr(caseInteger, Define(Dat("this is a float"), None, d.String))
		intStr   = option(Dat(23))
		fltStr   = option(Dat(42.23))
		boolNone = option(Dat(true))
	)
	fmt.Printf("option: %s, option type: %s\n",
		option, option.Type())

	fmt.Printf("intStr: %s, fltStr: %s, boolNone: %s\n",
		intStr, fltStr, boolNone)
	fmt.Printf("intStr type: %s, fltStr type: %s, boolNone type: %s\n",
		intStr.Type().TypeReturn(), fltStr.Type().TypeReturn(), boolNone.Type())

	fmt.Printf("type of intStr: %s, fltStr: %s, boolNone: %s\n",
		intStr.Type(), fltStr.Type(), boolNone.Type())

	if !intStr.Type().TypeReturn().Match(Either) {
		t.Fail()
	}
	if !fltStr.Type().TypeReturn().Match(Or) {
		t.Fail()
	}
	if !boolNone.Type().TypeReturn().Match(None) {
		t.Fail()
	}
}

func TestEnum(t *testing.T) {
	var enumtype EnumDef
	var weekdays = NewVector(
		Dat("Monday"),
		Dat("Tuesday"),
		Dat("Wednesday"),
		Dat("Thursday"),
		Dat("Friday"),
		Dat("Saturday"),
		Dat("Sunday"),
	)
	enumtype = NewEnumType(func(day d.Numeral) Expression {
		var idx = day.GoInt()
		if idx > 6 {
			idx = idx%6 - 1
		}
		return weekdays()[idx]
	})

	fmt.Printf("enum type days of the week: %s type: %s\n", enumtype, enumtype.Type().TypeName())
	var enum = enumtype(d.IntVal(8))
	fmt.Printf("wednesday eum: %s\n", enum)
	fmt.Printf("eum expr: %s\n", enum.Type())
	var val, idx, _ = enum()
	fmt.Printf("enum value val %s, index: %s\n",
		val, idx)
}
