package data

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)

func TestMutability(t *testing.T) {
	a := conData(true).(BoolVal)
	b := conData(false).(BoolVal)
	if a == b {
		t.Log("freh assigned values should be different", a, b)
	}
	a = b
	if a != b {
		t.Log("value of value has been assigned and should not differ", a, b)
	}
}
func TestTypeAllocation(t *testing.T) {
	s0 := conChain(
		conData(true),
		conData(1),
		conData(1, 2, 3, 4, 5, 6, 7),
		conData(int8(8)),
		conData(int16(16)),
		conData(int32(32)),
		conData(float32(32.16)),
		conData(float64(64.64)),
		conData(complex64(float32(32))),
		conData(complex128(float64(1.6))),
		conData(byte(3)),
		conData(time.Now()),
		conData(rune('ö')),
		conData(big.NewInt(23)),
		conData(big.NewFloat(23.42)),
		conData(big.NewRat(23, 42)),
		conData([]byte("test")),
		conData("test"))

	s1 := conChain()
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
