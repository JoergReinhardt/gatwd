package types

import (
	"fmt"
	"testing"
	"time"
)

func TestTypeStringer(t *testing.T) {
	var str string
	var u uint
	var i uint
	for u < uint(NATIVES) {
		if flag(Unary).match(ValType(u)) {
			str = str + ValType(u).String() + "\n"
		}
		i = i + 1
		u = uint(1) << i
	}
	fmt.Println(str)
	fmt.Println(flag(Unary).String())
}
func TestMutability(t *testing.T) {
	a := Make(true).(boolVal).Ref().(*boolVal)
	b := Make(false).(boolVal)
	if *a == b {
		t.Log("freh assigned values should be different", a, b)
	}
	*a = b
	if *a != b {
		t.Log("value of value has been assigned and should not differ", a, b)
	}
}
func TestTypeAllocation(t *testing.T) {
	s0 := newSlice(
		New(true),
		New(1),
		New(1, 2, 3, 4, 5, 6, 7),
		New(int8(8)),
		New(int16(16)),
		New(int32(32)),
		New(float32(32.16)),
		New(float64(64.64)),
		New(complex64(float32(32))),
		New(complex128(float64(1.6))),
		New(byte(3)),
		New(time.Now()),
		New([]byte("test")),
		New("test"))

	s1 := newSlice()

	fmt.Printf("List-1: %s\tList-2: %s\n", s0, s1)
}

func TestCellType(t *testing.T) {
}
func TestTimeType(t *testing.T) {
	ts := time.Now()
	v := timeVal(ts)
	fmt.Printf("time stamp: %s\n", v)
}
