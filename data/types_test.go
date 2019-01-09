package data

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)

func TestMutability(t *testing.T) {
	a := Con(true).(BoolVal)
	b := Con(false).(BoolVal)
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
		Con(true),
		Con(1),
		Con(1, 2, 3, 4, 5, 6, 7),
		Con(int8(8)),
		Con(int16(16)),
		Con(int32(32)),
		Con(float32(32.16)),
		Con(float64(64.64)),
		Con(complex64(float32(32))),
		Con(complex128(float64(1.6))),
		Con(byte(3)),
		Con(time.Now()),
		Con(rune('ö')),
		Con(big.NewInt(23)),
		Con(big.NewFloat(23.42)),
		Con(big.NewRat(23, 42)),
		Con([]byte("test")),
		Con("test"))

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
	var ds = Con(0, 7, 45,
		134, 4, 465, 3, 645,
		2452, 34, 45, 3535,
		24, 4, 24, 2245,
		24, 42, 4, 24)

	var ns = ds.(Chain).NativeSlice()

	fmt.Println(ns)
	fmt.Println(ds.Flag())
}
