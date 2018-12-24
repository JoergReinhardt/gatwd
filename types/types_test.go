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
	a := Make(true).(boolVal)
	b := Make(false).(boolVal)
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
		Make(true),
		Make(1),
		Make(1, 2, 3, 4, 5, 6, 7),
		Make(int8(8)),
		Make(int16(16)),
		Make(int32(32)),
		Make(float32(32.16)),
		Make(float64(64.64)),
		Make(complex64(float32(32))),
		Make(complex128(float64(1.6))),
		Make(byte(3)),
		Make(time.Now()),
		Make(rune('ö')),
		Make(big.NewInt(23)),
		Make(big.NewFloat(23.42)),
		Make(big.NewRat(23, 42)),
		Make([]byte("test")),
		Make("test"))

	s1 := newSlice()
	//s1 := []Evaluable{}
	//s1 := []int{}

	fmt.Printf("List-0: %s\n", s0)

	for i := 0; i < 1000000000; i++ {
		s1 = sliceAppend(s1, Make(i))
		//s1 = append(s1, i)
	}

	fmt.Printf("List-1 len: %d\t\n", len(s1))
}
func TestTimeType(t *testing.T) {
	ts := time.Now()
	v := timeVal(ts)
	fmt.Printf("time stamp: %s\n", v)
}
func TestTokenTypes(t *testing.T) {
	var i uint
	var typ TokenType = 1
	for i = 0; i < uint(len(syntax))-1; i++ {
		typ = 1 << i
		fmt.Printf("index:\t%d\tString:\t%s\t\tSyntax:\t%s\n", i, typ.String(), typ.Syntax())
	}
}
