package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var testIsZero = DecTest(func(args ...Expression) bool {
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

var compZero = DecComparator(func(args ...Expression) int {
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

var caseZero = DecCase(
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

var isInt = DecTest(func(args ...Expression) bool {
	for _, arg := range args {
		if arg.TypeFnc().Match(Data) {
			return arg.(Native).TypeNat().Match(d.Int)
		}
	}
	return false
})
var caseInt = DecCase(isInt, DecNative("this is an int"), d.Int, d.String)
var isUint = DecTest(func(args ...Expression) bool {
	for _, arg := range args {
		if arg.TypeFnc().Match(Data) {
			return arg.(Native).TypeNat().Match(d.Uint)
		}
	}
	return false
})
var caseUint = DecCase(isUint, DecNative("this is a uint"), d.Uint, d.String)
var isFloat = DecTest(func(args ...Expression) bool {
	for _, arg := range args {
		if arg.TypeFnc().Match(Data) {
			return arg.(Native).TypeNat().Match(d.Float)
		}
	}
	return false
})
var caseFloat = DecCase(isFloat, DecNative("this is a float"), d.Float, d.String)
var swi = DecSwitch(caseFloat, caseUint, caseInt)

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
	var maybeString = DecMaybe(caseInt)
	var str = maybeString(DecNative(42))
	if !str.Type().MatchArgs(DecNative("")) {
		t.Fail()
	}
	var none = maybeString(DecNative(true))
	if !none.Type().Match(None) {
		t.Fail()
	}

	fmt.Printf("maybe string: %s\n", maybeString)
	fmt.Printf("str: %s str-type: %s, none: %s\n", str, str.Type(), none)
}

func TestOption(t *testing.T) {
	var (
		option   = DecOption(caseInt, caseFloat)
		intStr   = option(DecNative(23))
		fltStr   = option(DecNative(42.23))
		boolNone = option(DecNative(true))
	)
	if intStr.Type().MatchArgs(DecNative(0)) {
		t.Fail()
	}
	if fltStr.Type().MatchArgs(DecNative(0.0)) {
		t.Fail()
	}
	if boolNone.Type().MatchArgs(DecNative(true)) {
		t.Fail()
	}

	fmt.Printf("ist str: %s, flt str: %s, bool none: %s\n ",
		intStr, fltStr, boolNone)
}
