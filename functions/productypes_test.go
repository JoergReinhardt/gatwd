package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var testIsZero = NewTest(func(args ...Expression) bool {
	for _, arg := range args {
		if arg.(Native).Eval().(d.Numeral).GoInt() != 0 {
			return false
		}
	}
	return true
})

func TestTestable(t *testing.T) {

	fmt.Printf("test: %s\n", testIsZero)

	fmt.Printf("test zero is zero (true): %t\n", testIsZero(DecNative(0)))
	if !testIsZero(DecNative(0)) {
		t.Fail()
	}

	fmt.Printf("test one is zero (false): %t\n", testIsZero(DecNative(1)))
	if testIsZero(DecNative(1)) {
		t.Fail()
	}

	fmt.Printf("test three zeros are zero (true): %t\n",
		testIsZero(DecNative(0), DecNative(0), DecNative(0)))
	if !testIsZero(DecNative(0), DecNative(0), DecNative(0)) {
		t.Fail()
	}
}

var compZero = NewComparator(func(args ...Expression) int {
	switch args[0].(Native).Eval().(d.Numeral).GoInt() {
	case -1:
		return -1
	case 0:
		return 0
	}
	return 1
})

func TestCompareable(t *testing.T) {
	fmt.Printf("compareable: %s\n", compZero)
	fmt.Printf("zero equals zero (0): %d\n", compZero(DecNative(0)))
	fmt.Printf("minus one lesser zero (-1): %d\n", compZero(DecNative(-1)))
	fmt.Printf("one greater zero (1): %d\n", compZero(DecNative(1)))
}

var caseZero = NewCase(
	testIsZero,
	DecNative("this is indeed zero"),
	d.Numbers,
	d.String,
)

func TestCase(t *testing.T) {
	fmt.Printf("case: %s\n", caseZero)

	var result = caseZero.Call(DecNative(0))
	fmt.Printf("case zero: %s\n", result)
	if result.String() != "this is indeed zero" {
		t.Fail()
	}

	result = caseZero.Call(DecNative(1))
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
var caseInteger = NewCase(isInteger, DecNative("this is an int"), d.Int, d.String)
var isUint = NewTest(func(args ...Expression) bool {
	for _, arg := range args {
		if arg.TypeFnc().Match(Data) {
			return arg.(Native).TypeNat().Match(d.Uint)
		}
	}
	return false
})
var caseUint = NewCase(isUint, DecNative("this is a uint"), d.Uint, d.String)
var isFloat = NewTest(func(args ...Expression) bool {
	for _, arg := range args {
		if arg.TypeFnc().Match(Data) {
			return arg.(Native).TypeNat().Match(d.Float)
		}
	}
	return false
})
var caseFloat = NewCase(isFloat, DecNative("this is a float"), d.Float, d.String)
var swi = NewSwitch(caseFloat, caseUint, caseInteger)

func TestSwitch(t *testing.T) {

	var result = swi.Call(DecNative(42))
	fmt.Printf("result from calling switch on 42: %s\n", result)
	if !result.Type().MatchArgs(DecNative(0)) {
		t.Fail()
	}

	swi = swi.Reload()
	result = swi.Call(DecNative(uint(42)))
	fmt.Printf("result from calling switch on uint 42: %s\n", result)
	if !result.Type().MatchArgs(DecNative(uint(0))) {
		t.Fail()
	}

	swi = swi.Reload()
	result = swi.Call(DecNative(42.23))
	fmt.Printf("result from calling switch on uint 42.23: %s\n", result)
	if !result.Type().MatchArgs(DecNative(uint(0.0))) {
		t.Fail()
	}

	swi = swi.Reload()
	result = swi.Call(DecNative(true))
	fmt.Printf("result from calling switch on true: %s\n", result)
	if !result.Type().Match(None) {
		t.Fail()
	}
}

func TestMaybe(t *testing.T) {
	var maybeString = NewOptional(caseInteger)
	var str = maybeString(DecNative(42))
	fmt.Printf("string: %s\n", str)
	if str.Type().TypeReturn().Match(Def(Data, d.String)) {
		t.Fail()
	}
	var none = maybeString(DecNative(true))
	fmt.Printf("none type: %s fnctype: %s\n", none.Type(), none.TypeFnc())
	if !none.Type().TypeReturn().Match(None) {
		t.Fail()
	}

	fmt.Printf("maybe string: %s, type: %s\n", maybeString, maybeString.Type())
	fmt.Printf("str: %s str-type: %s, none: %s\n", str, str.Type(), none)
}

func TestOption(t *testing.T) {
	var (
		option   = NewEitherOr(caseInteger, caseFloat)
		intStr   = option(DecNative(23))
		fltStr   = option(DecNative(42.23))
		boolNone = option(DecNative(true))
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
	var enumtype EnumType
	enumtype = NewEnumType(
		func(days ...d.Integer) Expression {
			if len(days) > 0 {
				if len(days) > 1 {
					var vec = NewVector()
					for _, day := range days {
						if (d.IntVal(0) <= day.Int()) &&
							(day.Int() <= d.IntVal(6)) {
							var result, _, _ = enumtype(day)
							vec = vec.Cons(result).(VecVal)
						}
					}
				}
				switch days[0].Int() {
				case 0:
					return DecNative("Monday")
				case 1:
					return DecNative("Tuesday")
				case 2:
					return DecNative("Wednesday")
				case 3:
					return DecNative("Thursday")
				case 4:
					return DecNative("Friday")
				case 5:
					return DecNative("Saturday")
				case 6:
					return DecNative("Sunday")
				}
			}
			return NewNone()
		},
		d.IntVal(0),
		d.IntVal(6),
	)

	fmt.Printf("enum type days of the week: %s\n", enumtype)
	var enum, min, max = enumtype(d.IntVal(2))
	fmt.Printf("wednesday eum: %s, min: %s, max: %s\n",
		enum, min, max)
	var val, idx, typ = enum()
	fmt.Printf("enum value val %s, index: %s, type: %s min %s, max %s\n",
		val, idx, typ, typ.Low(), typ.High())
}
