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
		fshow(frot(Tree.Flag(), flen(Natives.Flag()))),
		fshow(fhigh(Tree)),
	)
	fmt.Printf("test match true: %t, false: %t\n",
		BigInt.Flag().Match(BigInt.Flag()),
		BigInt.Flag().Match(Attr.Flag()),
	)
}
func TestMutability(t *testing.T) {
	a := conData(true).(boolVal)
	b := conData(false).(boolVal)
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

	s1 := newSlice()
	//s1 := []Evaluable{}
	//s1 := []int{}

	fmt.Printf("List-0: %s\n", s0.String())

	for i := 0; i < 1000; i++ {
		s1 = sliceAppend(s1, conData(i))
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
	var typ TokType = 1
	for i = 0; i < uint(len(syntax))-1; i++ {
		typ = 1 << i
		fmt.Printf("index:\t%d\tString:\t%s\t\tSyntax:\t%s\n", i, typ.String(), typ.Syntax())
	}
}
