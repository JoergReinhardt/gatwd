package data

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)

func TestMutability(t *testing.T) {
	a := New(true).(BoolVal)
	b := New(false).(BoolVal)
	if a == b {
		t.Log("freh assigned values should be different", a, b)
	}
	a = b
	if a != b {
		t.Log("value of value has been assigned and should not differ", a, b)
	}
}
func TestTypeAllocation(t *testing.T) {
	s0 := ConChain(
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
		New(rune('ö')),
		New(big.NewInt(23)),
		New(big.NewFloat(23.42)),
		New(big.NewRat(23, 42)),
		New([]byte("test")),
		New("test"))

	s1 := ConChain()
	//s1 := []Evaluable{}
	//s1 := []int{}

	fmt.Printf("List-0: %s\n", s0.String())

	for i := 0; i < 1000; i++ {
		s1 = ChainAdd(s1, s0...)
	}

	fmt.Printf("List-1 len: %d\t\n", len(s1))
}
func TestTimeType(t *testing.T) {
	ts := time.Now()
	v := TimeVal(ts)
	fmt.Printf("time stamp: %s\n", v.String())
}
func TestNativeSlice(t *testing.T) {
	var ds = New(0, 7, 45,
		134, 4, 465, 3, 645,
		2452, 34, 45, 3535,
		24, 4, 24, 2245,
		24, 42, 4, 24)

	var ns = ds.(Chain).NativeSlice()

	fmt.Println(ns)
	fmt.Println(ds.Flag())

	fmt.Println(stringChainTable(ds.(Chain)...))
}
func TestAllTypes(t *testing.T) {
	fmt.Println(AllTypes())
}
