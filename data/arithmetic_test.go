package data

import (
	"fmt"
	"testing"
)

func TestArithmetic(t *testing.T) {

	var i = arithmetics(IntVal(42), IntVal(23), Substract).(Numeral)
	fmt.Printf("result of substracting int from int: %s\n", i)
	fmt.Printf("result type of substracting int from int: %s\n", i.TypeNat())
	if i.Int() != 19 {
		fmt.Println("compare int result failed")
		t.Fail()
	}

	var f = arithmetics(IntVal(42), FltVal(23.42), Add).(Numeral)
	fmt.Printf("result of adding int to float: %s\n", f)
	fmt.Printf("result type of adding int to float: %s\n", f.TypeNat())
	if f.Float() != 65.42 {
		fmt.Println("compare float result failed")
		t.Fail()
	}

	var q = arithmetics(IntVal(42), UintVal(23), Divide).(Numeral)
	fmt.Printf("result of dividing int by uint: %s\n", q)
	fmt.Printf("result type of dividing int by uint: %s\n", q.TypeNat())
	fmt.Printf("result cast to float: %f\n", q.Float())
	if q.Float() != 1.826086956521739 {
		fmt.Printf("compare ratio result failed: %f\n", q.Float())
		t.Fail()
	}
}
