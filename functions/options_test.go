package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var testIsZero = DeclareTest(func(args ...Expression) bool {
	for _, arg := range args {
		if arg.(Native).Eval().(d.Numeral).GoInt() != 0 {
			return false
		}
	}
	return true
})

func TestTestable(t *testing.T) {

	fmt.Printf("test zero is zero (true): %t\n", testIsZero(DeclareNative(0)))
	if !testIsZero(DeclareNative(0)) {
		t.Fail()
	}
	fmt.Printf("test one is zero (false): %t\n", testIsZero(DeclareNative(1)))
	if testIsZero(DeclareNative(1)) {
		t.Fail()
	}
	fmt.Printf("test three zeros are zero (true): %t\n",
		testIsZero(DeclareNative(0), DeclareNative(0), DeclareNative(0)))
	if !testIsZero(DeclareNative(0), DeclareNative(0), DeclareNative(0)) {
		t.Fail()
	}
}

var compZero = DeclareComparator(func(args ...Expression) int {
	switch args[0].(Native).Eval().(d.Numeral).GoInt() {
	case -1:
		return -1
	case 0:
		return 0
	}
	return 1
})

func TestCompareable(t *testing.T) {
	fmt.Printf("zero equals zero (0): %d\n", compZero(DeclareNative(0)))
	fmt.Printf("minus one lesser zero (-1): %d\n", compZero(DeclareNative(-1)))
	fmt.Printf("one greater zero (1): %d\n", compZero(DeclareNative(1)))
}

var caseZero = DeclareCase(testIsZero, DeclareNative("this is indeed zero"))

func TestCase(t *testing.T) {
	var result = caseZero.Call(DeclareNative(0))
	fmt.Printf("case zero: %s\n", result)
	if result.String() != "this is indeed zero" {
		t.Fail()
	}
	fmt.Printf("case none zero: %s\n", result)
	if result.Type().Match(None) {
		t.Fail()
	}
}

var isInt = DeclareTest(func(args ...Expression) bool {
	return args[0].(Native).TypeNat().Match(d.Int)
})
var caseInt = DeclareCase(isInt, DeclareNative("this is an int"))

var isUint = DeclareTest(func(args ...Expression) bool {
	return args[0].(Native).TypeNat().Match(d.Uint)
})
var caseUint = DeclareCase(isUint, DeclareNative("this is a uint"))

var isFloat = DeclareTest(func(args ...Expression) bool {
	return args[0].(Native).TypeNat().Match(d.Float)
})
var caseFloat = DeclareCase(isFloat, DeclareNative("this is a float"))

var swi = DeclareSwitch(caseFloat, caseUint, caseInt)

func TestSwitch(t *testing.T) {
	fmt.Printf("switch: %s\n\n", swi)

	fmt.Printf("types:\nint: %s uint: %s float: %s\n\n",
		DeclareNative(1).Type(), DeclareNative(uint(1)).Type(), DeclareNative(1.1).Type())

	fmt.Printf("type matches:\nint: %t uint: %t float: %t\n\n",
		DeclareNative(1).TypeNat().Match(d.Int),
		DeclareNative(uint(1)).TypeNat().Match(d.Uint),
		DeclareNative(1.1).TypeNat().Match(d.Float))

	var result, cases = swi(DeclareNative(1))
	fmt.Printf("result: %s, cases: %s\n",
		result, cases)
	for len(cases) > 0 {
		result, cases = swi(DeclareNative(1))
		fmt.Printf("result: %s, cases: %s\n",
			result, cases)
	}

	swi = swi.Reload()
	result, cases = swi(DeclareNative(true))
	fmt.Printf("result: %s, cases: %s\n",
		result, cases)
	for len(cases) > 0 {
		result, cases = swi(DeclareNative(true))
		fmt.Printf("result: %s, cases: %s\n",
			result, cases)
	}

	swi = swi.Reload()
	fmt.Printf("\nis int: %t\n", isInt(DeclareNative(1)))
	fmt.Printf("case int: %s\n", caseInt.Call(DeclareNative(1)))
	result = swi.Call(DeclareNative(1))
	fmt.Printf("switch return: %s\n", result)
	fmt.Printf("switch return type: %s\n\n", result.Type())
	if result.String() != "this is an int" {
		t.Fail()
	}

	swi = swi.Reload()
	fmt.Printf("is float: %t\n", isFloat(DeclareNative(1.1)))
	fmt.Printf("case float: %s\n", caseFloat.Call(DeclareNative(1.1)))
	result = swi.Call(DeclareNative(1.1))
	fmt.Printf("switch return: %s\n", result)
	fmt.Printf("switch return type: %s\n\n", result.Type())
	if result.TypeFnc().Match(Def(d.String)) {
		t.Fail()
	}

	swi = swi.Reload()
	fmt.Printf("is uint: %t\n", isUint(DeclareNative(uint(1))))
	fmt.Printf("case uint: %s\n", caseUint.Call(DeclareNative(uint(11))))
	result = swi.Call(DeclareNative(uint(1)))
	fmt.Printf("switch return: %s\n", result)
	fmt.Printf("switch return type: %s\n\n", result.Type())
	if result.TypeFnc().Match(Def(d.String)) {
		t.Fail()
	}

	swi = swi.Reload()
	result = swi.Call(DeclareNative(true))
	fmt.Printf("switch return call with bool: %s\n", result)
	fmt.Printf("switch return type: %s\n\n", result.Type())
	if result.TypeFnc().Match(Def(None)) {
		t.Fail()
	}
}

var intCase = DeclareTypeCase(d.Int)
var uintCase = DeclareTypeCase(d.Uint)
var floatCase = DeclareTypeCase(d.Float)

func TestAllTypeCase(t *testing.T) {

	fmt.Printf("all type cases: %s, %s, %s\n",
		intCase.Type(), uintCase.Type(), floatCase.Type())

	fmt.Printf("int value? %s\n", intCase.Call(DeclareNative(1)))
	if !intCase(DeclareNative(1)).Type().Match(Def(d.Int)) {
		t.Fail()
	}

	fmt.Printf("not int value? %s\n", intCase(DeclareNative(1.1)))
	if !intCase(DeclareNative(1.1)).Type().Match(Def(None)) {
		t.Fail()
	}
}

var typeSwitch = NewTypeSwitch(Def(d.Int), Def(d.Uint), Def(d.Float))

func TestTypeSwitch(t *testing.T) {

	fmt.Printf("type switch: %s\n", typeSwitch)

	var result, cases = typeSwitch(DeclareNative(1))
	for len(cases) > 0 {
		fmt.Printf("result: %s cases: %s type: %s\n", result, cases, result.Type())
		result, cases = DeclareSwitch(cases...)(DeclareNative(1))
	}

	typeSwitch = typeSwitch.Reload()
	result = typeSwitch.Call(DeclareNative(1.2))
	fmt.Printf("result from Call %s Type: %s\n", result, result.Type())
	if !result.Type().Match(Def(d.Float)) {
		t.Fail()
	}

	typeSwitch = typeSwitch.Reload()
	result = typeSwitch.Call(DeclareNative(1))
	fmt.Printf("result from Call %s Type: %s\n", result, result.Type())
	if !result.Type().Match(Def(d.Int)) {
		t.Fail()
	}
}

func TestMaybe(t *testing.T) {
	var maybe = DeclareMaybe(caseInt)
	var result = maybe(DeclareNative(1))
	fmt.Printf("result of calling maybe int with an int: %s\n", result)
	fmt.Printf("result type: %s\n", result.Type())
	if !result.TypeFnc().Match(Data) {
		t.Fail()
	}
	result = maybe(DeclareNative(1.1))
	fmt.Printf("result of calling maybe int with a float: %s\n", result)
	fmt.Printf("result type: %s\n", result.Type())
	if !result.TypeFnc().Match(None) {
		t.Fail()
	}
}

func TestOption(t *testing.T) {
	var option = DeclareOption(caseInt, caseFloat)
	fmt.Printf("option: %s\n", option)
	fmt.Printf("option type: %s\n", option.Type())

	var result = option(DeclareNative(1))
	fmt.Printf("option called with int: %s\n", result)
	fmt.Printf("result type %s\n", result.Type())
	if !result.TypeReturn().Match(Def(d.String)) {
		t.Fail()
	}

	result = option(DeclareNative(1.1))
	fmt.Printf("option called with float: %s\n", result)
	fmt.Printf("result type %s\n", result.Type())
	if !result.TypeReturn().Match(Def(d.String)) {
		t.Fail()
	}

	result = option(DeclareNative(true))
	fmt.Printf("option called with bool: %s\n", result)
	fmt.Printf("result type %s\n", result.Type())
	if !result.TypeReturn().Match(None) {
		t.Fail()
	}
}

func TestIf(t *testing.T) {
	var ifs = DeclareBranch(caseInt, caseFloat)
	fmt.Printf("if: %s\n", ifs)
	fmt.Printf("if type: %s\n", ifs.Type())

	var result = ifs(DeclareNative(1))
	fmt.Printf("if called with int: %s\n", result)
	fmt.Printf("result type %s\n", result.Type())
	if !result.TypeReturn().Match(Def(d.String)) {
		t.Fail()
	}

	result = ifs(DeclareNative(1.1))
	fmt.Printf("if called with float: %s\n", result)
	fmt.Printf("result type %s\n", result.Type())
	if !result.TypeReturn().Match(Def(d.String)) {
		t.Fail()
	}

	result = ifs(DeclareNative(true))
	fmt.Printf("if called with bool: %s\n", result)
	fmt.Printf("result type %s\n", result.Type())
	if !result.TypeReturn().Match(None) {
		t.Fail()
	}
}
