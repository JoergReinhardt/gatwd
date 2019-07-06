package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var truth = NewTestTruth("x = 0", func(args ...Expression) bool {
	var num = args[0].Eval().(d.IntVal)
	return num == 0
})

var trinary = NewTestTrinary("x > 0", func(args ...Expression) int {
	var num = args[0].Eval().(d.IntVal)
	if num < 0 {
		return -1
	}
	if num > 0 {
		return 1
	}
	return 0
})

var compare = NewTestCompare("x = 0", func(args ...Expression) int {
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

	fmt.Printf("test truth bool false, true, false: %t, %t, %t\n",
		truth.Test(New(1)), truth.Test(New(0)), truth.Test(New(-1)))

	if !truth.Test(New(0)) {
		t.Fail()
	}
	if truth.Test(New(1)) || truth.Test(New(-1)) {
		t.Fail()
	}

	fmt.Printf("compare truth int -1, 0, -1: %d, %d, %d\n",
		truth.Compare(New(-1)), truth.Compare(New(0)), truth.Compare(New(1)))

	if truth.Compare(New(0)) != 0 {
		t.Fail()
	}
	if truth.Compare(New(1)) != -1 || truth.Compare(New(-1)) != -1 {
		t.Fail()
	}

	fmt.Printf("trinary truth truth type: False, Undecided, True: %s %s %s\n",
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

	fmt.Printf("test trinary truth bool type: false, false, true: %t %t %t\n",
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

	fmt.Printf("compare order type: -1, 0, 1: %d %d %d\n",
		compare.Compare(New(-1)), compare.Compare(New(0)), compare.Compare(New(1)))
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

	fmt.Printf("test compare bool type: false, true, false: %t %t %t\n",
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

	fmt.Printf("test compare Lesser, Equal, Greater all true: %t %t %t\n",
		compare.Lesser(New(-1)), compare.Equal(New(0)), compare.Greater(New(1)))
	if !compare.Lesser(New(-1)) {
		t.Fail()
	}
	if !compare.Equal(New(0)) {
		t.Fail()
	}
	if !compare.Greater(New(1)) {
		t.Fail()
	}

	fmt.Printf("test compare Lesser, Equal, Greater all false: %t %t %t\n",
		compare.Lesser(New(1)), compare.Equal(New(1)), compare.Greater(New(-1)))
	if compare.Lesser(New(1)) {
		t.Fail()
	}
	if compare.Equal(New(1)) {
		t.Fail()
	}
	if compare.Greater(New(-1)) {
		t.Fail()
	}

	fmt.Printf("test trinary Truth, True, Undecided, False all true: %t %t %t\n",
		trinary.True(New(1)), trinary.Undecided(New(0)), trinary.False(New(-1)))
	if !trinary.True(New(1)) {
		t.Fail()
	}
	if !trinary.Undecided(New(0)) {
		t.Fail()
	}
	if !trinary.False(New(-1)) {
		t.Fail()
	}

	fmt.Printf("test trinary Truth, Undecided, False all false: %t %t %t\n",
		trinary.True(New(-1)), trinary.Undecided(New(11)), trinary.False(New(1)))
	if trinary.True(New(-1)) {
		t.Fail()
	}
	if trinary.Undecided(New(1)) {
		t.Fail()
	}
	if trinary.False(New(1)) {
		t.Fail()
	}

	fmt.Printf("test Truth,True, False all true: %t %t\n",
		truth.True(New(0)), truth.False(New(1)))
	if !truth.True(New(0)) {
		t.Fail()
	}
	if !truth.False(New(1)) {
		t.Fail()
	}

	fmt.Printf("test Truth, True, False all false: %t %t\n",
		truth.True(New(1)), truth.False(New(0)))
	if truth.True(New(1)) {
		t.Fail()
	}
	if truth.False(New(0)) {
		t.Fail()
	}
}

var test = NewTestTruth("x = String|Integers|Float", func(args ...Expression) bool {
	for _, arg := range args {
		if !arg.(NativeConst).TypeNat().Match(
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
