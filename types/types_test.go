package types

import (
	"fmt"
	"testing"
)

func TestTypeStringer(t *testing.T) {
	var str string
	var u uint
	var i uint
	for u < uint(MAX_VALUE_TYPE) {
		if Flag(Unary).Match(ValType(u)) {
			str = str + ValType(u).String() + "\n"
		}
		i = i + 1
		u = uint(1) << i
	}
	fmt.Println(str)
	fmt.Println(Flag(Unary).String())
}
func TestMutability(t *testing.T) {
	a := U(true).(*boolVal)
	b := Val(false).(boolVal)
	if *a == b {
		t.Log("freh assigned values should be different", a, b)
	}
	*a = b
	if *a != b {
		t.Log("value of value has been assigned and should not differ", a, b)
	}
}
func TestTypeAllocation(t *testing.T) {
	//var output = []string{}
	s0 := newSlice(
		U(true),
		U(1),
		U(int8(8)),
		U(int16(16)),
		U(int32(32)),
		U(float32(32.16)),
		U(float64(64.64)),
		U(complex64(float32(32))),
		U(complex128(float64(1.6))),
		U(byte(3)),
		U([]byte("test")),
		U("test"))

	s1 := newSlice()

	fmt.Printf("List-1: %s\tList-2: %s\n", s0, s1)
}

func TestCellType(t *testing.T) {
}
