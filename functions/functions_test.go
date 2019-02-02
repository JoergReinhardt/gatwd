package functions

import (
	"fmt"
	"testing"

	d "github.com/JoergReinhardt/godeep/data"
)

func TestDataEnclosures(t *testing.T) {
	data := New("this is the testfunction speaking from within enclosure")
	fmt.Println(data)
	fmt.Println(data.Flag())
}
func TestPairEnclosures(t *testing.T) {
	pair := NewPair(New("test key:"), New("test data in a pair"))
	a, b := pair.Left(), pair.Right()
	fmt.Println(a)
	fmt.Println(b)
}
func TestStringVectorEnclosures(t *testing.T) {
	vec := NewVector(
		New("first data in slice"),
		New("second data entry in slice"),
		New("third data entry in slice"),
		New("fourth data entry in slice"),
		New("fifth data entry in slice"),
		New("sixt data entry in slice"),
		New("seventh data entry in slice"),
		New("eigth data entry in slice"),
		New("nineth data entry in slice"),
		New("tenth data entry in slice"),
	)
	fmt.Println(vec.Flag())
	fmt.Println(vec.Slice())
	fmt.Println(vec.String())
}
func TestMixedVectorEnclosures(t *testing.T) {
	vec := NewVector(
		New("this is"),
		New("a vector of"),
		New("mixed type"),
		New(5, 7, 234, 4, 546, 324, 4),
	)
	fmt.Println(vec.Flag())
	fmt.Println(vec.Len())
	fmt.Println(vec.Empty())
}
func TestIntegerVectorEnclosures(t *testing.T) {
	vec := New(0, 7, 45,
		134, 4, 465, 3, 645,
		2452, 34, 45, 3535,
		24, 4, 24, 2245,
		24, 42, 4, 24)
	fmt.Println(vec.Flag())
	fmt.Println(vec.String())
}
func TestParameterEnclosure(t *testing.T) {
	var dat Functional
	var parm Argumented
	parm = NewArgument(New("test parameter"))
	dat = parm.Arg()
	fmt.Println(parm.Arg())
	_, parm = parm.Apply(New("changer parameter"))
	fmt.Println(parm.Arg())
	_, parm = parm.Apply(New("yet another parameter"))
	_, parm = parm.Apply()
	fmt.Println(parm.Arg())
	fmt.Println(parm.Arg())
	_, parm = parm.Apply(New("yup, works just fine ;)"))
	fmt.Println(parm.Arg())
	fmt.Println(dat.Flag())
}
func TestAccParamEnclosure(t *testing.T) {
	acc := NewKeyValueParm(New("test-key"), New("test value"))
	fmt.Println(acc)
	_, acc = acc.Apply(NewKeyValueParm(New(12), New("one million dollar")))
	fmt.Printf("Accessor Type: %s\n", acc.Acc().Flag())
	fmt.Printf("Accessor: %s\n", acc.Acc())
	fmt.Printf("Argument: %s\n", acc.Arg())
	if acc.Arg().Eval() != New("one million dollar").Eval() {
		t.Fail()
	}
	_, acc = acc.Apply(NewKeyValueParm(New(13), New("two million dollar")))
	fmt.Println(acc)
	if acc.Arg().Eval() != New("two million dollar").Eval() {
		t.Fail()
	}
}
func TestApplyArgs(t *testing.T) {
	args := NewwArguments(New(0), New(1), New(2), New(3), New(4), New(5))

	fmt.Println(args)
	var dat []d.Data
	_, args = args.Apply(args.Data()...)

	fmt.Println(args)
	//	if dat[3].(d.IntVal) != 3 {
	//		t.Fail()
	//	}

	_, args = args.Apply(New(7), New(1), New(2), New(5), New(4), New(8))

	fmt.Println(args)
	fmt.Println(dat)
	//	if dat[3].(d.IntVal) != 5 &&
	//		args.Args()[0].Data().(d.IntVal) != 7 &&
	//		args.Args()[5].Data().(d.IntVal) != 8 {
	//		t.Fail()
	//	}
	//	fmt.Println(args.Get(3))
}

var acc = NewParameters(
	NewPair(
		New("first key"),
		New("first value"),
	),
	NewPair(
		New("second key"),
		New("second value"),
	),
	NewPair(
		New("third key"),
		New("third value"),
	),
	NewPair(
		New("fourth key"),
		New("fourth value"),
	),
	NewPair(
		New("fifth key"),
		New("fifth value"),
	),
	NewPair(
		New("sixt key"),
		New("sixt value")))
var acc2 = NewParameters(
	NewPair(
		New("first key"),
		New("changed first value"),
	),
	NewPair(
		New("six key"),
		New("changed sixt value")))

func TestAccAttrs(t *testing.T) {
	fmt.Printf("original list: %s\n", acc)
	fmt.Printf("change set: %s\n", acc2)
	_, acc1 := acc.Apply(acc2.Parms()...)
	fmt.Printf("list after appying change set: %s\n", acc1)
	if acc1.Get(New("first key")).Right().Eval() != New("changed first value").Eval() {
		t.Fail()
	}
	if acc1.Get(New("second key")).Right().Eval() != New("second value").Eval() {
		t.Fail()
	}

}
func TestSearchAccAttrs(t *testing.T) {
	praed := New("fourth key")
	fmt.Printf("why nil: %s\n", praed.Flag())
	var cha = newPairSorter(acc.Pairs()...)
	cha.Sort(d.String)
	fmt.Println(cha)
	idx := cha.Search(praed)
	if idx != -1 {
		fmt.Println(cha[idx].Left())
	}
	if cha[idx].Left().String() != praed.String() {
		t.Fail()
	}

}

var macc = newPairSorter(
	NewPair(
		New("string"),
		New("string value"),
	),
	NewPair(
		New("int"),
		New(12),
	),
	NewPair(
		New("uint"),
		New(uint(10)),
	),
	NewPair(
		New("float"),
		New(4.2),
	),
)

func TestMixedTypeAccessor(t *testing.T) {
	macc.Sort(d.Flag)
	idx := macc.Search(New(d.String))
	fmt.Printf("%d\n", idx)
	if idx > 0 {
		found := macc[idx]
		fmt.Println(found.Right())
		if found.Right().String() != "string value" {
			t.Fail()
		}

		idx = macc.Search(New(d.Int.Flag()))
		foundi := macc[idx]
		fmt.Printf("%d\n", foundi.Right())
		if foundi.Right().(Integer).Int() != 12 {
			t.Fail()
		}

		idx = macc.Search(New(d.Uint.Flag()))
		foundu := macc[idx]
		fmt.Printf("%d\n", foundu.Right())
		if foundu.Right().(Unsigned).Uint() != 10 {
			t.Fail()
		}

		idx = macc.Search(New(d.Float.Flag()))
		foundf := macc[idx]
		fmt.Printf("%f\n", foundf.Right())
		if foundf.Right().(Irrational).Float() != 4.2 {
			t.Fail()
		}
	}
}
func TestFlag(t *testing.T) {
	data := New("test string")
	fmt.Printf("test strings flag: %s\n", data.Flag())
}
func TestApplyAccessAttrs(t *testing.T) {
	acc3 := ApplyParams(acc, acc2.Pairs()...)
	fmt.Println(acc3)
	acc2 = NewParameters(append(acc2.Pairs(), NewPair(New("seventh key"), New("changed seventh value")))...)
}

var accc = NewParameters(
	NewPair(
		New("eigth key"),
		New("changed eigth value"),
	),
	NewPair(
		New("second key"),
		New("second value"),
	),
	NewPair(
		New("thirteenth key"),
		New("hirteenth value"),
	),
	NewPair(
		New("nineth key"),
		New("nineth value"),
	))
var accl = NewParameters(
	append(acc.Pairs(), []Paired{NewPair(
		New("seventh key"),
		New("seventh value"),
	),
		NewPair(
			New("eigth key"),
			New("eigth value"),
		),
		NewPair(
			New("nineth key"),
			New("nineth value"),
		),
		NewPair(
			New("tenth key"),
			New("tenth value"),
		),
		NewPair(
			New("eleventh key"),
			New("eleventh value"),
		),
		NewPair(
			New("twelveth key"),
			New("twelveth value"))}...)...)

func TestFmtAccessorBenchmarkExpression(t *testing.T) {
	fmt.Printf("accessors to replace:\n%s\n", accc)
	fmt.Printf("accessor set to replace accessors in:\n%s\n", accl)
}
func BenchmarkAccessorApply(b *testing.B) {
	//var accn = []Accessable{}
	for i := 0; i < b.N; i++ {
		_ = ApplyParams(accl, accc.Pairs()...)
	}
}

var accc1 = NewParameters(
	NewPair(
		New("fourteenth key"),
		New("changed fourteenth value"),
	),
	NewPair(
		New("third key"),
		New("changed third value"),
	),
	NewPair(
		New("seventh key"),
		New("changed seventh value"),
	),
	NewPair(
		New("first key"),
		New("changed first value"),
	),
	NewPair(
		New("eigth key"),
		New("changed changed eigth value"),
	),
	NewPair(
		New("second key"),
		New("changed second value"),
	),
	NewPair(
		New("thirteenth key"),
		New("changed hirteenth value"),
	),
	NewPair(
		New("nineth key"),
		New("changed nineth value"),
	))

func TestFmtMoreAccessorsBenchmarkExpression(t *testing.T) {
	fmt.Printf("more accessors to replace:\n%s\n", accc1)
	fmt.Printf("same accessor set to replace accessors in:\n%s\n", ApplyParams(accl, accc1.Pairs()...))
}
func BenchmarkMoreAccessorApply(b *testing.B) {
	//var accn = []Accessable{}
	for i := 0; i < b.N; i++ {
		_ = ApplyParams(accl, accc1.Pairs()...)
	}
}
func TestRecursive(t *testing.T) {
	l := NewRecursiveList(
		New("this"),
		New("is"),
		New("a"),
		New("test"),
		New("List"),
	)
	var h d.Data
	fmt.Println(l.Len())
	fmt.Println(l.Empty())
	for l != nil {
		h, l = l.DeCap()
		fmt.Println(h)
	}
}
func TestTuple(t *testing.T) {
	tup := NewTuple(
		New("this"),
		New("is"),
		New("a"),
		New("test"),
		New("Tuple"),
		New(19),
		New(23.42),
	)
	fmt.Println(tup)
}
func TestRecord(t *testing.T) {
	rec := NewRecord(
		NewPair(New("key-0"), New("this")),
		NewPair(New("key-1"), New("is")),
		NewPair(New("key-2"), New("a")),
		NewPair(New("key-3"), New("test")),
		NewPair(New("key-4"), New("Tuple")),
		NewPair(New("key-5"), New(19)),
		NewPair(New("key-6"), New(23.42)),
	)
	fmt.Println(rec)
	fmt.Println(rec.ArgSig())
}
