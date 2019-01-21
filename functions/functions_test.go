package functions

import (
	"fmt"
	"testing"

	d "github.com/JoergReinhardt/godeep/data"
)

func TestDataEnclosures(t *testing.T) {
	data := NewFncData(d.New("this is the testfunction speaking from within enclosure"))
	fmt.Println(data)
	fmt.Println(data.Flag())
}
func TestPairEnclosures(t *testing.T) {
	pair := NewPair(d.New("test key:"), d.New("test data in a pair"))
	a, b := pair.Left(), pair.Right()
	fmt.Println(a)
	fmt.Println(b)
}
func TestStringVectorEnclosures(t *testing.T) {
	vec := newVector(
		d.New("first data in slice"),
		d.New("second data entry in slice"),
		d.New("third data entry in slice"),
		d.New("fourth data entry in slice"),
		d.New("fifth data entry in slice"),
		d.New("sixt data entry in slice"),
		d.New("seventh data entry in slice"),
		d.New("eigth data entry in slice"),
		d.New("nineth data entry in slice"),
		d.New("tenth data entry in slice"),
	)
	fmt.Println(vec.Flag())
	fmt.Println(vec.Slice())
	fmt.Println(vec.String())
}
func TestMixedVectorEnclosures(t *testing.T) {
	vec := newVector(
		d.New("this is"),
		d.New("a vector of"),
		d.New("mixed type"),
		d.New(5, 7, 234, 4, 546, 324, 4),
	)
	fmt.Println(vec.Flag())
	fmt.Println(vec.Len())
	fmt.Println(vec.Empty())
}
func TestIntegerVectorEnclosures(t *testing.T) {
	vec := d.New(0, 7, 45,
		134, 4, 465, 3, 645,
		2452, 34, 45, 3535,
		24, 4, 24, 2245,
		24, 42, 4, 24)
	fmt.Println(vec.Flag())
	fmt.Println(vec.String())
}
func TestParameterEnclosure(t *testing.T) {
	var dat Function
	var parm Argumented
	parm = NewArgument(d.New("test parameter"))
	dat = parm.Arg()
	fmt.Println(dat)
	dat, parm = parm.Apply(NewFncData(d.New("changer parameter")))
	fmt.Println(dat)
	dat, parm = parm.Apply(NewFncData(d.New("yet another parameter")))
	dat, parm = parm.Apply()
	fmt.Println(dat)
	fmt.Println(dat)
	dat, parm = parm.Apply(NewFncData(d.New("yup, works just fine ;)")))
	fmt.Println(dat)
	fmt.Println(dat.Flag())
}
func TestAccParamEnclosure(t *testing.T) {
	acc := NewParameter(NewPair(d.New("test-key"), d.New("test value")))
	fmt.Println(acc)
	_, acc = acc.Apply(NewPair(d.New(12), d.New("one million dollar")))
	fmt.Println(acc)
	if acc.Key() != d.New(12) {
		t.Fail()
	}
	_, acc = acc.Apply(NewPair(d.New(13), d.New("two million dollar")))
	fmt.Println(acc)
	if acc.Key() != d.New(13) {
		t.Fail()
	}
}
func TestApplyArgs(t *testing.T) {
	args := NewwArguments(d.New(0), d.New(1), d.New(2), d.New(3), d.New(4), d.New(5))

	fmt.Println(args)
	var dat []Function
	dat, args = args.Apply(args.Data()...)

	fmt.Println(args)
	if dat[3].(d.IntVal) != 3 {
		t.Fail()
	}

	dat, args = args.Apply(d.New(7), d.New(1), d.New(2), d.New(5), d.New(4), d.New(8))

	fmt.Println(args)
	if dat[3].(d.IntVal) != 5 &&
		args.Args()[0].Data().(d.IntVal) != 7 &&
		args.Args()[5].Data().(d.IntVal) != 8 {
		t.Fail()
	}
	fmt.Println(args.Get(3))
}

var acc = NewParameters(
	NewPair(
		d.New("first key"),
		d.New("first value"),
	),
	NewPair(
		d.New("second key"),
		d.New("second value"),
	),
	NewPair(
		d.New("third key"),
		d.New("third value"),
	),
	NewPair(
		d.New("fourth key"),
		d.New("fourth value"),
	),
	NewPair(
		d.New("fifth key"),
		d.New("fifth value"),
	),
	NewPair(
		d.New("sixt key"),
		d.New("sixt value")))
var acc2 = NewParameters(
	NewPair(
		d.New("first key"),
		d.New("changed first value"),
	),
	NewPair(
		d.New("six key"),
		d.New("changed sixt value")))

func TestAccAttrs(t *testing.T) {
	fmt.Println(acc)
	p, acc1 := acc.Apply(NewParameters(
		NewPair(
			d.New("first key"),
			d.New("first value"),
		),
		NewPair(
			d.New("second key"),
			d.New("changed second value"),
		),
		NewPair(
			d.New("third key"),
			d.New("third value"),
		),
		NewPair(
			d.New("fourth key"),
			d.New("changed fourth value"),
		),
		NewPair(
			d.New("fifth key"),
			d.New("fifth value"),
		),
		NewPair(
			d.New("sixt key"),
			d.New("sixt value"))).Pairs()...)

	fmt.Println(p)
	fmt.Println(acc1)
	fmt.Printf("get \"second key\" %s\n", acc1.Get(d.New("second key")))
	if acc1.Get(d.New("second key")).Right() != d.New("changed second value") {
		t.Fail()
	}

	_, acc2 := acc1.Apply(NewParameters(
		NewPair(
			d.New("second key"),
			d.New("changed second value again"),
		),
		NewPair(
			d.New("fourth key"),
			d.New("changed fourth value again"))).Pairs()...)

	fmt.Println(acc2)

}
func TestSearchAccAttrs(t *testing.T) {
	praed := d.New("fourth key")
	var cha = pairSorter{}
	args, _ := acc.Apply()
	for _, c := range args {
		cha = append(cha, c)
	}
	cha.Sort(d.String)
	fmt.Println(cha)
	idx := cha.Search(praed)
	fmt.Println(cha[idx].Left())
	if cha[idx].Left().String() != praed.String() {
		t.Fail()
	}

}

var macc = newPairSorter(
	NewPair(
		d.New("string"),
		d.New("string value"),
	),
	NewPair(
		d.New("int"),
		d.New(12),
	),
	NewPair(
		d.New("uint"),
		d.New(uint(10)),
	),
	NewPair(
		d.New("float"),
		d.New(4.2),
	),
)

func TestMixedTypeAccessor(t *testing.T) {
	macc.Sort(d.Flag)
	idx := macc.Search(d.String)
	fmt.Printf("%d\n", idx)
	if idx > 0 {
		found := macc[idx]
		fmt.Println(found.Right())
		if found.Right().String() != "string value" {
			t.Fail()
		}

		idx = macc.Search(d.Int.Flag())
		foundi := macc[idx]
		fmt.Printf("%d\n", foundi.Right())
		if foundi.Right().(Integer).Int() != 12 {
			t.Fail()
		}

		idx = macc.Search(d.Uint.Flag())
		foundu := macc[idx]
		fmt.Printf("%d\n", foundu.Right())
		if foundu.Right().(Unsigned).Uint() != 10 {
			t.Fail()
		}

		idx = macc.Search(d.Float.Flag())
		foundf := macc[idx]
		fmt.Printf("%f\n", foundf.Right())
		if foundf.Right().(Irrational).Float() != 4.2 {
			t.Fail()
		}
	}
}
func TestApplyAccessAttrs(t *testing.T) {
	acc3 := ApplyParams(acc, acc2.Pairs()...)
	fmt.Println(acc3)
	acc2 = NewParameters(append(acc2.Pairs(), NewPair(d.New("seventh key"), d.New("changed seventh value")))...)
}

var accc = NewParameters(
	NewPair(
		d.New("eigth key"),
		d.New("changed eigth value"),
	),
	NewPair(
		d.New("second key"),
		d.New("second value"),
	),
	NewPair(
		d.New("thirteenth key"),
		d.New("hirteenth value"),
	),
	NewPair(
		d.New("nineth key"),
		d.New("nineth value"),
	))
var accl = NewParameters(
	append(acc.Pairs(), []Paired{NewPair(
		d.New("seventh key"),
		d.New("seventh value"),
	),
		NewPair(
			d.New("eigth key"),
			d.New("eigth value"),
		),
		NewPair(
			d.New("nineth key"),
			d.New("nineth value"),
		),
		NewPair(
			d.New("tenth key"),
			d.New("tenth value"),
		),
		NewPair(
			d.New("eleventh key"),
			d.New("eleventh value"),
		),
		NewPair(
			d.New("twelveth key"),
			d.New("twelveth value"))}...)...)

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
		d.New("fourteenth key"),
		d.New("changed fourteenth value"),
	),
	NewPair(
		d.New("third key"),
		d.New("changed third value"),
	),
	NewPair(
		d.New("seventh key"),
		d.New("changed seventh value"),
	),
	NewPair(
		d.New("first key"),
		d.New("changed first value"),
	),
	NewPair(
		d.New("eigth key"),
		d.New("changed changed eigth value"),
	),
	NewPair(
		d.New("second key"),
		d.New("changed second value"),
	),
	NewPair(
		d.New("thirteenth key"),
		d.New("changed hirteenth value"),
	),
	NewPair(
		d.New("nineth key"),
		d.New("changed nineth value"),
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
	l := newRecursive(
		d.New("this"),
		d.New("is"),
		d.New("a"),
		d.New("test"),
		d.New("List"),
	)
	var h Function
	fmt.Println(l.Len())
	fmt.Println(l.Empty())
	for l != nil {
		h, l = l.DeCap()
		fmt.Println(h)
	}
}
func TestTuple(t *testing.T) {
	tup := newTuple(
		d.New("this"),
		d.New("is"),
		d.New("a"),
		d.New("test"),
		d.New("Tuple"),
		d.New(19),
		d.New(23.42),
	)
	fmt.Println(tup)
}
func TestRecord(t *testing.T) {
	rec := newRecord(
		NewPair(d.New("key-0"), d.New("this")),
		NewPair(d.New("key-1"), d.New("is")),
		NewPair(d.New("key-2"), d.New("a")),
		NewPair(d.New("key-3"), d.New("test")),
		NewPair(d.New("key-4"), d.New("Tuple")),
		NewPair(d.New("key-5"), d.New(19)),
		NewPair(d.New("key-6"), d.New(23.42)),
	)
	fmt.Println(rec)
	fmt.Println(rec.ArgSig())
}
