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
	fmt.Printf("zero equals zero (0): %d\n", compZero(DecNative(0)))
	fmt.Printf("minus one lesser zero (-1): %d\n", compZero(DecNative(-1)))
	fmt.Printf("one greater zero (1): %d\n", compZero(DecNative(1)))
}

var caseZero = DecCase(testIsZero, DecNative("this is indeed zero"))

func TestCase(t *testing.T) {
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
var caseInt = DecCase(isInt, DecNative("this is an int"))

var isUint = DecTest(func(args ...Expression) bool {
	for _, arg := range args {
		if arg.TypeFnc().Match(Data) {
			return arg.(Native).TypeNat().Match(d.Uint)
		}
	}
	return false
})
var caseUint = DecCase(isUint, DecNative("this is a uint"))

var isFloat = DecTest(func(args ...Expression) bool {
	for _, arg := range args {
		if arg.TypeFnc().Match(Data) {
			return arg.(Native).TypeNat().Match(d.Float)
		}
	}
	return false
})
var caseFloat = DecCase(isFloat, DecNative("this is a float"))

var swi = DecSwitch(caseFloat, caseUint, caseInt)

func TestSwitch(t *testing.T) {
	var result, cases = swi(DecNative(1))
	for len(cases) > 0 {
		fmt.Printf("result from calling switch passing int: %s\n", result)
		result, cases = DecSwitch(cases...)(DecNative(1))
	}

	swi = swi.Switch()
	result = swi.Call(DecNative(1))
	fmt.Printf("result from call(1): %s\n", result)
}
