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
func TestFlag(t *testing.T) {
	flag := Flag.Flag()
	ok := FlagMatch(flag, Flag.Flag())
	fmt.Println(ok)
	if !ok {
		t.Fail()
	}
	ok = FlagMatch(flag, Flag.Flag()|Int.Flag())
	fmt.Println(ok)
	if !ok {
		t.Fail()
	}
	ok = FlagMatch(flag, Int.Flag())
	fmt.Println(ok)
	if ok {
		t.Fail()
	}
	count := FlagCount(String | Int | Float)
	fmt.Println(count)
	if count != 3 {
		t.Fail()
	}

	fmt.Println(BitFlag(Int.Flag() | Float.Flag()).Decompose())
	fmt.Println(BitFlag(Int | Float).String())

	fmt.Println(BitFlag(Symbolic))

	if fmt.Sprint(BitFlag(Symbolic)) != "Byte|Rune|Bytes|String|Error" {
		t.Fail()
	}

	fmt.Println(BitFlag(Symbolic))
}

var s0 = NewChain(
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

func TestTypeAllocation(t *testing.T) {

	fmt.Println(s0.ContainedTypes())
	if fmt.Sprint(s0.ContainedTypes()) != "Bool|Int|Int8|Int16|Int32|BigInt|Float|Flt32|BigFlt|Ratio|Imag|Byte|Bytes|String|Time" {
		t.Fail()
	}
	s1 := NewChain()
	//s1 := []Evaluable{}
	//s1 := []int{}

	fmt.Printf("List-0: %s\n", s0.String())
	fmt.Printf("List-0 Length: %d\n", s0.Len())
	if len(s0) != 18 {
		t.Fail()
	}

	for i := 0; i < 1000; i++ {
		s1 = ChainAdd(s1, s0...)
	}
	if len(s1) != 18000 {
		t.Fail()
	}
	fmt.Printf("contained types s1: %s\n", s1.ContainedTypes())

	fmt.Printf("List-1 len: %d\t\n", len(s1))
	fmt.Printf("List-1 type: %s\t\n", s1.Flag().String())
}

func TestLiFo(t *testing.T) {
	//var sr = Chain{}
	var d Data
	var s = NewChain()
	var sr = NewChain()
	for i := 0; i < 10; i++ {
		s = ChainPush(s, New(i))
		fmt.Println(d)
		fmt.Println(s)
		fmt.Println(ChainLen(s))
	}
	fmt.Println(s)
	for ChainLen(s) > 0 {
		d, s = ChainPop(s)
		fmt.Println(d)
		fmt.Println(s)
		fmt.Println(s.Len())
		sr = append(sr, d)
	}
	fmt.Println(sr)
	if sr[0] != New(9) {
		t.Fail()
	}

}
func TestFiFo(t *testing.T) {
	//var sr = Chain{}
	var d Data
	var s = Chain{}
	var sr = Chain{}
	for i := 0; i < 10; i++ {
		s = ChainPut(s, New(i))
		fmt.Println(d)
		fmt.Println(s)
		fmt.Println(ChainLen(s))
	}
	for !ChainEmpty(s) {
		d, s = ChainPull(s)
		sr = append(sr, d)
		fmt.Println(d)
		fmt.Println(s)
		fmt.Println(ChainLen(s))
	}
	fmt.Println(sr)

	if sr[0] != New(0) {
		t.Fail()
	}
}
func TestConDecap(t *testing.T) {
	//var sr = Chain{}
	var d Data
	var s = Chain{}
	var sr = Chain{}
	for i := 0; i < 10; i++ {
		s = ChainCon(s, New(i))
		fmt.Println(d)
		fmt.Println(s)
		fmt.Println(ChainLen(s))
	}
	for !ChainEmpty(s) {
		d, s = ChainDecap(s)
		sr = append(sr, d)
		fmt.Println(d)
		fmt.Println(s)
		fmt.Println(ChainLen(s))
	}
	fmt.Println(sr)

	if sr[0] != New(9) {
		t.Fail()
	}
}
func BenchmarkListAdd(b *testing.B) {
	var s1 = Chain{}
	for i := 0; i < b.N; i++ {
		s1 = ChainAdd(s1, s0...)
	}
}
func BenchmarkListAppend(b *testing.B) {
	var s1 = Chain{}
	for i := 0; i < b.N; i++ {
		s1 = ChainAppend(s1, s0...)
	}
}
func BenchmarkListPushPop(b *testing.B) {
	var s1 = Chain{}
	for i := 0; i < b.N; i++ {
		s1 = ChainPush(s1, s0[0])
	}
	for i := 0; i < b.N; i++ {
		_, s1 = ChainPop(s1)
	}
}
func BenchmarkListPutPull(b *testing.B) {
	var s1 = Chain{}
	for i := 0; i < b.N; i++ {
		s1 = ChainPut(s1, s0[0])
	}
	for i := 0; i < b.N; i++ {
		_, s1 = ChainPull(s1)
	}
}
func BenchmarkConDecap(b *testing.B) {
	var s1 = Chain{}
	for i := 0; i < b.N; i++ {
		s1 = ChainCon(s1, s0[0])
	}
	for i := 0; i < b.N; i++ {
		_, s1 = ChainDecap(s1)
	}
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

	fmt.Println(ds)
	fmt.Println(ns)

}
func TestAllTypes(t *testing.T) {
	fmt.Println(ListAllTypes())
	if fmt.Sprint(ListAllTypes()) != "[Nil Bool Int Int8 Int16 Int32 BigInt Uint Uint8 Uint16 Uint32 Float Flt32 BigFlt Ratio Imag Imag64 Byte Rune Bytes String Time Duration Error Slice Map Function Flag]" {
		t.Fail()
	}
}

func TestSearchChain(t *testing.T) {
	sl := New(1, 11, 45, 324, 2, 35, 3, 435, 4, 3).(Chain)
	fmt.Println(sl)
	sl.Sort(Int)
	fmt.Println(sl)
	comp := New(11)
	dat := sl.Search(comp)
	fmt.Println(dat)
	data := sl.Search(New(11))
	fmt.Println(data)
}
