package functions

import (
	"fmt"
	"testing"

	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lang"
)

func TestIdGenerator(t *testing.T) {
	ts := newTypeState()
	var id int
	id = ts.NewUid()
	fmt.Println(id)
	id = ts.NewUid()
	fmt.Println(id)
	if id != 1 {
		t.Fail()
	}
	id = ts.NewUid()
	id = ts.NewUid()
	id = ts.NewUid()
	id = ts.NewUid()
	id = ts.NewUid()
	fmt.Println(id)
	if id != 6 {
		t.Fail()
	}
}
func TestSliceMatch(t *testing.T) {
	ts := [][]Token{
		[]Token{
			newToken(Syntax_Token, l.Lambda),
			newToken(Syntax_Token, l.DoubCol),
			newToken(Data_Type_Token, d.Bool),
			newToken(Syntax_Token, l.RightArrow),
			newToken(Data_Type_Token, d.Bool),
		},
		[]Token{
			newToken(Syntax_Token, l.Lambda),
			newToken(Syntax_Token, l.DoubCol),
			newToken(Data_Type_Token, d.Bool),
			newToken(Syntax_Token, l.RightArrow),
			newToken(Data_Type_Token, d.Bool),
		},
		[]Token{
			newToken(Syntax_Token, l.Lambda),
			newToken(Syntax_Token, l.DoubCol),
			newToken(Data_Type_Token, d.Slice),
			newToken(Syntax_Token, l.RightArrow),
			newToken(Data_Type_Token, d.Numeric),
			newToken(Syntax_Token, l.RightArrow),
			newToken(Data_Type_Token, d.Int),
			newToken(Syntax_Token, l.RightArrow),
			newToken(Data_Type_Token, d.Bool),
		},
		[]Token{
			newToken(Syntax_Token, l.Lambda),
			newToken(Syntax_Token, l.DoubCol),
			newToken(Data_Type_Token, d.Slice),
			newToken(Syntax_Token, l.RightArrow),
			newToken(Data_Type_Token, d.Numeric),
			newToken(Syntax_Token, l.RightArrow),
			newToken(Data_Type_Token, d.Int),
			newToken(Syntax_Token, l.RightArrow),
			newToken(Data_Type_Token, d.Bool),
		},
		[]Token{
			newToken(Syntax_Token, l.Lambda),
			newToken(Syntax_Token, l.DoubCol),
			newToken(Data_Type_Token, d.Symbolic),
			newToken(Syntax_Token, l.RightArrow),
			newToken(Data_Type_Token, d.Numeric),
			newToken(Syntax_Token, l.RightArrow),
			newToken(Data_Type_Token, d.Int),
			newToken(Syntax_Token, l.RightArrow),
			newToken(Data_Type_Token, d.Bool),
		},
	}

	sortTokenSlice(ts)

	ok := compareTokenSequence(ts[0], ts[1])
	fmt.Println(tokens(ts[0]))
	fmt.Println(tokens(ts[1]))
	fmt.Println(ok)
	if !ok {
		t.Fail()
	}

	ok = sortSlicePairByLength(ts[0], ts[2])
	fmt.Println(tokens(ts[0]))
	fmt.Println(tokens(ts[2]))
	fmt.Println(ok)
	if ok {
		t.Fail()
	}

	ok = sortSlicePairByLength(ts[2], ts[3])
	fmt.Println(tokens(ts[2]))
	fmt.Println(tokens(ts[3]))
	fmt.Println(ok)
	if !ok {
		t.Fail()
	}

	ok = sortSlicePairByLength(ts[3], ts[4])
	fmt.Println(tokens(ts[3]))
	fmt.Println(tokens(ts[4]))
	fmt.Println(ok)
	if ok {
		t.Fail()
	}
	nomatch := []Token{
		newToken(Syntax_Token, l.Lambda),
		newToken(Syntax_Token, l.DoubCol),
		newToken(Data_Value_Token, d.Symbolic),
		newToken(Syntax_Token, l.LeftArrow),
		newToken(Data_Value_Token, d.Numeric),
		newToken(Syntax_Token, l.FatLArrow),
		newToken(Data_Value_Token, d.Int),
		newToken(Syntax_Token, l.LeftArrow),
		newToken(Data_Value_Token, d.Bool),
	}

	ok = sliceContainsSignature(nomatch, ts)
	fmt.Println(ok)

	ok = sliceContainsSignature(ts[0], ts)
	fmt.Println(ok)
}
func TestDataEnclosures(t *testing.T) {
	data := newData(d.New("this is the testfunction speaking from within enclosure"))
	fmt.Println(data)
	fmt.Println(data.Flag())
}
func TestPairEnclosures(t *testing.T) {
	pair := newPair(d.New("test key:"), d.New("test data in a pair"))
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
	fmt.Println(vec.Type())
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
	fmt.Println(vec.Type())
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
	var dat Data
	var parm Argumented
	parm = newArgument(d.New("test parameter"))
	dat = parm.Arg()
	fmt.Println(dat)
	dat, parm = parm.Set(newData(d.New("changer parameter")))
	fmt.Println(dat)
	dat, parm = parm.Set(newData(d.New("yet another parameter")))
	dat, parm = parm.Set()
	fmt.Println(dat)
	fmt.Println(dat)
	dat, parm = parm.Set(newData(d.New("yup, works just fine ;)")))
	fmt.Println(dat)
	fmt.Println(parm.Type())
	fmt.Println(dat.Flag())
}
func TestAccParamEnclosure(t *testing.T) {
	acc := newPraedicate(newPair(d.New("test-key"), d.New("test value")))
	fmt.Println(acc)
	_, acc = acc.Set(newPair(d.New(12), d.New("one million dollar")))
	fmt.Println(acc)
	if acc.Key() != d.New(12) {
		t.Fail()
	}
	_, acc = acc.Set(newPair(d.New(13), d.New("two million dollar")))
	fmt.Println(acc)
	if acc.Key() != d.New(13) {
		t.Fail()
	}
}
func TestTokenToSignature(t *testing.T) {
	syn := toksS(
		l.RightArrow,
		l.RightArrow,
		l.LeftBra,
		l.Pipe,
		l.Pipe,
		l.RightBra,
		l.RightArrow,
		l.RightArrow,
	)
	typ := toksD(
		d.Int,
		d.Int,
		d.Byte,
		d.Rune,
		d.Int,
		d.Int,
	)
	fmt.Println(syn)
	fmt.Println(typ)

}
func TestApplyArgs(t *testing.T) {
	args := newArguments(d.New(0), d.New(1), d.New(2), d.New(3), d.New(4), d.New(5))

	fmt.Println(args)
	_, args = args.Set(args.Args()...)

	fmt.Println(args)
	if args.Args()[3].Data().(d.IntVal) != 3 {
		t.Fail()
	}

	_, args = args.Set(newArguments(d.New(7), d.New(1), d.New(2), d.New(5), d.New(4), d.New(8)).Args()...)

	fmt.Println(args)
	if args.Args()[3].Data().(d.IntVal) != 5 &&
		args.Args()[0].Data().(d.IntVal) != 7 &&
		args.Args()[5].Data().(d.IntVal) != 8 {
		t.Fail()
	}
}

var acc = newPraedicates(
	newPair(
		d.New("first key"),
		d.New("first value"),
	),
	newPair(
		d.New("second key"),
		d.New("second value"),
	),
	newPair(
		d.New("third key"),
		d.New("third value"),
	),
	newPair(
		d.New("fourth key"),
		d.New("fourth value"),
	),
	newPair(
		d.New("fifth key"),
		d.New("fifth value"),
	),
	newPair(
		d.New("sixt key"),
		d.New("sixt value")))
var acc2 = newPraedicates(
	newPair(
		d.New("first key"),
		d.New("changed first value"),
	),
	newPair(
		d.New("six key"),
		d.New("changed sixt value")))

func TestAccAttrs(t *testing.T) {
	fmt.Println(acc)
	p, acc1 := acc.Set(newPraedicates(
		newPair(
			d.New("first key"),
			d.New("first value"),
		),
		newPair(
			d.New("second key"),
			d.New("changed second value"),
		),
		newPair(
			d.New("third key"),
			d.New("third value"),
		),
		newPair(
			d.New("fourth key"),
			d.New("changed fourth value"),
		),
		newPair(
			d.New("fifth key"),
			d.New("fifth value"),
		),
		newPair(
			d.New("sixt key"),
			d.New("sixt value"))).Pairs()...)

	fmt.Println(p)
	fmt.Println(acc1)

	_, acc2 := acc1.Set(newPraedicates(
		newPair(
			d.New("second key"),
			d.New("changed second value again"),
		),
		newPair(
			d.New("fourth key"),
			d.New("changed fourth value again"))).Pairs()...)

	fmt.Println(acc2)
}
func TestSearchAccAttrs(t *testing.T) {
	praed := d.New("fourth key")
	var cha = pairSorter{}
	args, _ := acc.Set()
	for _, c := range args {
		cha = append(cha, c)
	}
	cha.Sort(d.String.Flag())
	fmt.Println(cha)
	idx := cha.Search(praed)
	fmt.Println(cha[idx].Left())
	if cha[idx].Left().String() != praed.String() {
		t.Fail()
	}
}

var macc = newPairSorter(
	newPair(
		d.New("string"),
		d.New("string value"),
	),
	newPair(
		d.New("int"),
		d.New(12),
	),
	newPair(
		d.New("uint"),
		d.New(uint(10)),
	),
	newPair(
		d.New("float"),
		d.New(4.2),
	),
)

func TestMixedTypeAccessor(t *testing.T) {
	macc.Sort(d.Flag.Flag())
	idx := macc.Search(d.String.Flag())
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
	acc3 := applyPraedicate(acc, acc2.Pairs()...)
	fmt.Println(acc3)
	acc2 = newPraedSet(append(acc2.Pairs(), newPair(d.New("seventh key"), d.New("changed seventh value")))...)
}

var accc = newPraedicates(
	newPair(
		d.New("eigth key"),
		d.New("changed eigth value"),
	),
	newPair(
		d.New("second key"),
		d.New("second value"),
	),
	newPair(
		d.New("thirteenth key"),
		d.New("hirteenth value"),
	),
	newPair(
		d.New("nineth key"),
		d.New("nineth value"),
	))
var accl = newPraedicates(
	append(acc.Pairs(), []Paired{newPair(
		d.New("seventh key"),
		d.New("seventh value"),
	),
		newPair(
			d.New("eigth key"),
			d.New("eigth value"),
		),
		newPair(
			d.New("nineth key"),
			d.New("nineth value"),
		),
		newPair(
			d.New("tenth key"),
			d.New("tenth value"),
		),
		newPair(
			d.New("eleventh key"),
			d.New("eleventh value"),
		),
		newPair(
			d.New("twelveth key"),
			d.New("twelveth value"))}...)...)

func TestFmtAccessorBenchmarkExpression(t *testing.T) {
	fmt.Printf("accessors to replace:\n%s\n", accc)
	fmt.Printf("accessor set to replace accessors in:\n%s\n", accl)
}
func BenchmarkAccessorApply(b *testing.B) {
	//var accn = []Accessable{}
	for i := 0; i < b.N; i++ {
		_ = applyPraedicate(accl, accc.Pairs()...)
	}
}

var accc1 = newPraedicates(
	newPair(
		d.New("fourteenth key"),
		d.New("changed fourteenth value"),
	),
	newPair(
		d.New("third key"),
		d.New("changed third value"),
	),
	newPair(
		d.New("seventh key"),
		d.New("changed seventh value"),
	),
	newPair(
		d.New("first key"),
		d.New("changed first value"),
	),
	newPair(
		d.New("eigth key"),
		d.New("changed changed eigth value"),
	),
	newPair(
		d.New("second key"),
		d.New("changed second value"),
	),
	newPair(
		d.New("thirteenth key"),
		d.New("changed hirteenth value"),
	),
	newPair(
		d.New("nineth key"),
		d.New("changed nineth value"),
	))

func TestFmtMoreAccessorsBenchmarkExpression(t *testing.T) {
	fmt.Printf("more accessors to replace:\n%s\n", accc1)
	fmt.Printf("same accessor set to replace accessors in:\n%s\n", applyPraedicate(accl, accc1.Pairs()...))
}
func BenchmarkMoreAccessorApply(b *testing.B) {
	//var accn = []Accessable{}
	for i := 0; i < b.N; i++ {
		_ = applyPraedicate(accl, accc1.Pairs()...)
	}
}
func TestTuple(t *testing.T) {
	var slice []Data
	for _, dat := range accl.Pairs() {
		slice = append(slice, dat)
	}
	tup := newTuple(slice...)
	fmt.Println(tup)
	for tup.Tail() != nil {
		tup = tup.Shift()
		var head, tail = tup.Head(), tup.Tail()
		fmt.Printf("head: %s\ntail %s\n\n", head, tail)
		fmt.Println(tail)
	}
	dat := d.New("refill data-0")
	tup = conTuple(d.New("refill data-1"), tup)
	fmt.Println(conTuple(dat, tup))
	tup = conTuple(d.New("refill data-2"), tup)
	fmt.Println(conTuple(dat, tup))
	tup = conTuple(d.New("refill data-3"), tup)
	fmt.Println(conTuple(dat, tup))
	tup = conTuple(d.New("refill data-4"), tup)
	fmt.Println(conTuple(dat, tup))
	tup = conTuple(d.New("refill data-5"), tup)
	fmt.Println(conTuple(dat, tup))
	tup = conTuple(d.New("refill data-6"), tup)
	fmt.Println(conTuple(dat, tup))
	tup = conTuple(d.New("refill data-7"), tup)
	fmt.Println(conTuple(dat, tup))
	tup = conTuple(d.New("refill data-8"), tup)
	fmt.Println(conTuple(dat, tup))
	tup = conTuple(d.New("refill data-9"), tup)
	fmt.Println(conTuple(dat, tup))
	tup = conTuple(d.New("refill data-10"), tup)
	fmt.Println(conTuple(dat, tup))
}
