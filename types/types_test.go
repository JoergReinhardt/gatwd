package types

import (
	"fmt"
	"testing"
)

func TestMutability(t *testing.T) {
	a := New(true).(*BoolVal)
	b := Make(false).(BoolVal)
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
		New(true),
		New(1),
		New(int8(8)),
		New(int16(16)),
		New(int32(32)),
		New(float32(32.16)),
		New(float64(64.64)),
		New(complex64(float32(32))),
		New(complex128(float64(1.6))),
		New(byte(3)),
		New([]byte("test")),
		New("test"))

	s1 := newSlice()

	fmt.Printf("List-1: %s\tList-2: %s\n", s0, s1)
}

func TestCellType(t *testing.T) {
}
