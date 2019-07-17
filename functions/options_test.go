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
	fmt.Printf("case zero: %s\n", caseZero(NewNative(0)))
	if caseZero(NewNative(0)).String() != "this is indeed zero" {
		t.Fail()
	}
	fmt.Printf("case none zero: %s\n", caseZero(NewNative(1)))
	if !caseZero(NewNative(1)).Type().Match(None) {
		t.Fail()
	}
}

var isInt = NewTest(func(args ...Expression) bool {
	return args[0].(Native).Eval().Type().Match(d.Int)
})
var isUint = NewTest(func(args ...Expression) bool {
	return args[0].(Native).Eval().Type().Match(d.Uint)
})
var isFloat = NewTest(func(args ...Expression) bool {
	return args[0].(Native).Eval().Type().Match(d.Float)
})
var caseInt = NewCase(isInt, NewNative("this is an int"))
var caseFloat = NewCase(isFloat, NewNative("this is a float"))
var caseUint = NewCase(isUint, NewNative("this is a uint"))
var swi = NewSwitch(caseFloat, caseUint, caseInt)

func TestTypeSwitch(t *testing.T) {
	var result Expression
	fmt.Printf("switch: %s\n", swi)
	result, swi = swi(NewNative(1))
	fmt.Printf("switch return: %s, switch: %s\n", result, swi)
	result, swi = swi(NewNative(1))
	fmt.Printf("switch return: %s, switch: %s\n", result, swi)
	result, swi = swi(NewNative(1))
	fmt.Printf("switch return: %s, switch: %s\n", result, swi)
	result, swi = swi(NewNative(1))
	fmt.Printf("switch return: %s, switch: %s\n", result, swi)
	fmt.Printf("switch return: %s\n", swi.Call(NewNative(1)))
}
