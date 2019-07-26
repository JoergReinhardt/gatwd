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

	fmt.Printf("test zero is zero (true): %t\n", testIsZero(NewNative(0)))
	if !testIsZero(NewNative(0)) {
		t.Fail()
	}
	fmt.Printf("test one is zero (false): %t\n", testIsZero(NewNative(1)))
	if testIsZero(NewNative(1)) {
		t.Fail()
	}
	fmt.Printf("test three zeros are zero (true): %t\n",
		testIsZero(NewNative(0), NewNative(0), NewNative(0)))
	if !testIsZero(NewNative(0), NewNative(0), NewNative(0)) {
		t.Fail()
	}
}

var compZero = NewCompare(func(args ...Expression) int {
	switch args[0].(Native).Eval().(d.Numeral).GoInt() {
	case -1:
		return -1
	case 0:
		return 0
	}
	return 1
})

func TestCompareable(t *testing.T) {
	fmt.Printf("zero equals zero (0): %d\n", compZero(NewNative(0)))
	fmt.Printf("minus one lesser zero (-1): %d\n", compZero(NewNative(-1)))
	fmt.Printf("one greater zero (1): %d\n", compZero(NewNative(1)))
}

var caseZero = NewCase(testIsZero, NewNative("this is indeed zero"))

func TestCase(t *testing.T) {
	var result = caseZero.Call(NewNative(0))
	fmt.Printf("case zero: %s\n", result)
	if result.String() != "this is indeed zero" {
		t.Fail()
	}
	fmt.Printf("case none zero: %s\n", result)
	if result.Type().Match(None) {
		t.Fail()
	}
}

var isInt = NewTest(func(args ...Expression) bool {
	return args[0].(Native).TypeNat().Match(d.Int)
})
var caseInt = NewCase(isInt, NewNative("this is an int"))

var isUint = NewTest(func(args ...Expression) bool {
	return args[0].(Native).TypeNat().Match(d.Uint)
})
var caseUint = NewCase(isUint, NewNative("this is a uint"))

var isFloat = NewTest(func(args ...Expression) bool {
	return args[0].(Native).TypeNat().Match(d.Float)
})
var caseFloat = NewCase(isFloat, NewNative("this is a float"))

var swi = NewSwitch(caseFloat, caseUint, caseInt)

func TestTypeSwitch(t *testing.T) {
	fmt.Printf("switch: %s\n\n", swi)

	fmt.Printf("types:\nint: %s uint: %s float: %s\n\n",
		NewNative(1).Type(), NewNative(uint(1)).Type(), NewNative(1.1).Type())

	fmt.Printf("type matches:\nint: %t uint: %t float: %t\n\n",
		NewNative(1).TypeNat().Match(d.Int),
		NewNative(uint(1)).TypeNat().Match(d.Uint),
		NewNative(1.1).TypeNat().Match(d.Float))

	var result, cases = swi(NewNative(1))
	fmt.Printf("result: %s, cases: %s\n",
		result, cases)
	for len(cases) > 0 {
		result, cases = swi(NewNative(1))
		fmt.Printf("result: %s, cases: %s\n",
			result, cases)
	}

	swi = swi.Reload()
	result, cases = swi(NewNative(true))
	fmt.Printf("result: %s, cases: %s\n",
		result, cases)
	for len(cases) > 0 {
		result, cases = swi(NewNative(true))
		fmt.Printf("result: %s, cases: %s\n",
			result, cases)
	}

	swi = swi.Reload()
	fmt.Printf("\nis int: %t\n", isInt(NewNative(1)))
	fmt.Printf("case int: %s\n", caseInt.Call(NewNative(1)))
	result = swi.Call(NewNative(1))
	fmt.Printf("switch return: %s\n", result)
	fmt.Printf("switch return type: %s\n\n", result.Type())
	if result.String() != "this is an int" {
		t.Fail()
	}

	swi = swi.Reload()
	fmt.Printf("is float: %t\n", isFloat(NewNative(1.1)))
	fmt.Printf("case float: %s\n", caseFloat.Call(NewNative(1.1)))
	result = swi.Call(NewNative(1.1))
	fmt.Printf("switch return: %s\n", result)
	fmt.Printf("switch return type: %s\n\n", result.Type())
	if result.TypeFnc().Match(Def(Data, d.String)) {
		t.Fail()
	}

	swi = swi.Reload()
	fmt.Printf("is uint: %t\n", isUint(NewNative(uint(1))))
	fmt.Printf("case uint: %s\n", caseUint.Call(NewNative(uint(11))))
	result = swi.Call(NewNative(uint(1)))
	fmt.Printf("switch return: %s\n", result)
	fmt.Printf("switch return type: %s\n\n", result.Type())
	if result.TypeFnc().Match(Def(Data, d.String)) {
		t.Fail()
	}

	swi = swi.Reload()
	result = swi.Call(NewNative(true))
	fmt.Printf("switch return call with bool: %s\n", result)
	fmt.Printf("switch return type: %s\n\n", result.Type())
	if result.TypeFnc().Match(Def(None)) {
		t.Fail()
	}
}

func TestMaybe(t *testing.T) {
	var maybe = NewMaybe(caseInt)
	var result = maybe(NewNative(1))
	fmt.Printf("result of calling maybe int with an int: %s\n", result)
	fmt.Printf("result type: %s\n", result.Type())
	if !result.TypeFnc().Match(Data) {
		t.Fail()
	}
	result = maybe(NewNative(1.1))
	fmt.Printf("result of calling maybe int with a float: %s\n", result)
	fmt.Printf("result type: %s\n", result.Type())
	if !result.TypeFnc().Match(None) {
		t.Fail()
	}
}

func TestOption(t *testing.T) {
	var option = NewOption(caseInt, caseFloat)
	fmt.Printf("option: %s\n", option)
	fmt.Printf("option type: %s\n", option.Type())
	var result = option(NewNative(1))
	fmt.Printf("option called with int: %s\n", result)
	fmt.Printf("result type %s\n", result.Type())

	result = option(NewNative(1.1))
	fmt.Printf("option called with float: %s\n", result)
	fmt.Printf("result type %s\n", result.Type())

	result = option(NewNative(true))
	fmt.Printf("option called with bool: %s\n", result)
	fmt.Printf("result type %s\n", result.Type())
}
