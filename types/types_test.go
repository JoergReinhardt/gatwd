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
	a := Make(true).(boolVal).ref().(*boolVal)
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
		New(rune('ö')),
		New(big.NewInt(23)),
		New(big.NewFloat(23.42)),
		New(big.NewRat(23, 42)),
		New([]byte("test")),
		New("test"))

	s1 := newSlice()

	fmt.Printf("List-1: %s\tList-2: %s\n", s0, s1)
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
