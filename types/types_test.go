package types

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)

func TestTypeFlag(t *testing.T) {
	fmt.Printf("null:\t\t%scomp:\t\t%snat:\t\t%smask:\t\t%s\n",
		fshow(Nullable),
		fshow(Composed),
		fshow(Natives),
		fshow(Mask),
	)
	fmt.Printf("tree:\t\t%stree rotated:\t%stree shifted:\t%s\n",
		fshow(Tree),
		fshow(frot(Tree.Type(), flen(Natives.Type()))),
		fshow(fhigh(Tree)),
	)
	fmt.Printf("test match true: %t, false: %t\n",
		BigInt.Type().Match(BigInt.Type()),
		BigInt.Type().Match(Attr.Type()),
	)
}
func TestMutability(t *testing.T) {
	a := data(true).(boolVal)
	b := data(false).(boolVal)
	if a == b {
		t.Log("freh assigned values should be different", a, b)
	}
	a = b
	if a != b {
		t.Log("value of value has been assigned and should not differ", a, b)
	}
}
func TestTypeAllocation(t *testing.T) {
	s0 := newSlice(
		data(true),
		data(1),
		data(1, 2, 3, 4, 5, 6, 7),
		data(int8(8)),
		data(int16(16)),
		data(int32(32)),
		data(float32(32.16)),
		data(float64(64.64)),
		data(complex64(float32(32))),
		data(complex128(float64(1.6))),
		data(byte(3)),
		data(time.Now()),
		data(rune('ö')),
		data(big.NewInt(23)),
		data(big.NewFloat(23.42)),
		data(big.NewRat(23, 42)),
		data([]byte("test")),
		data("test"))

	s1 := newSlice()
	//s1 := []Evaluable{}
	//s1 := []int{}

	fmt.Printf("List-0: %s\n", s0.String())

	for i := 0; i < 1000; i++ {
		s1 = sliceAppend(s1, data(i))
		//s1 = append(s1, i)
	}

	fmt.Printf("List-1 len: %d\t\n", len(s1))
}
func TestTimeType(t *testing.T) {
	ts := time.Now()
	v := timeVal(ts)
	fmt.Printf("time stamp: %s\n", v.String())
}
func TestTokenTypes(t *testing.T) {
	var i uint
	var typ TokenType = 1
	for i = 0; i < uint(len(syntax))-1; i++ {
		typ = 1 << i
		fmt.Printf("index:\t%d\tString:\t%s\t\tSyntax:\t%s\n", i, typ.String(), typ.Syntax())
	}
}
