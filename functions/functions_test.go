package functions

import (
	"fmt"
	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lang"
	"testing"
)

func TestIdGenerator(t *testing.T) {
	var id int
	id = conUID()
	fmt.Println(id)
	id = conUID()
	fmt.Println(id)
	if id != 1 {
		t.Fail()
	}
	id = conUID()
	id = conUID()
	id = conUID()
	id = conUID()
	id = conUID()
	fmt.Println(id)
	if id != 6 {
		t.Fail()
	}
}
func TestSliceMatch(t *testing.T) {
	ts := [][]Token{
		[]Token{
			conToken(Syntax_Token, l.Lambda.Flag()),
			conToken(Syntax_Token, l.DoubCol.Flag()),
			conToken(Data_Value_Token, d.Bool.Flag()),
			conToken(Syntax_Token, l.LeftArrow.Flag()),
			conToken(Data_Value_Token, d.Bool.Flag()),
		},
		[]Token{
			conToken(Syntax_Token, l.Lambda.Flag()),
			conToken(Syntax_Token, l.DoubCol.Flag()),
			conToken(Data_Value_Token, d.Bool.Flag()),
			conToken(Syntax_Token, l.LeftArrow.Flag()),
			conToken(Data_Value_Token, d.Bool.Flag()),
		},
		[]Token{
			conToken(Syntax_Token, l.Lambda.Flag()),
			conToken(Syntax_Token, l.DoubCol.Flag()),
			conToken(Data_Value_Token, d.Slice.Flag()),
			conToken(Syntax_Token, l.LeftArrow.Flag()),
			conToken(Data_Value_Token, d.Numeral.Flag()),
			conToken(Syntax_Token, l.LeftArrow.Flag()),
			conToken(Data_Value_Token, d.Int.Flag()),
			conToken(Syntax_Token, l.LeftArrow.Flag()),
			conToken(Data_Value_Token, d.Bool.Flag()),
		},
		[]Token{
			conToken(Syntax_Token, l.Lambda.Flag()),
			conToken(Syntax_Token, l.DoubCol.Flag()),
			conToken(Data_Value_Token, d.Slice.Flag()),
			conToken(Syntax_Token, l.LeftArrow.Flag()),
			conToken(Data_Value_Token, d.Numeral.Flag()),
			conToken(Syntax_Token, l.LeftArrow.Flag()),
			conToken(Data_Value_Token, d.Int.Flag()),
			conToken(Syntax_Token, l.LeftArrow.Flag()),
			conToken(Data_Value_Token, d.Bool.Flag()),
		},
		[]Token{
			conToken(Syntax_Token, l.Lambda.Flag()),
			conToken(Syntax_Token, l.DoubCol.Flag()),
			conToken(Data_Value_Token, d.Symbolic.Flag()),
			conToken(Syntax_Token, l.LeftArrow.Flag()),
			conToken(Data_Value_Token, d.Numeral.Flag()),
			conToken(Syntax_Token, l.LeftArrow.Flag()),
			conToken(Data_Value_Token, d.Int.Flag()),
			conToken(Syntax_Token, l.LeftArrow.Flag()),
			conToken(Data_Value_Token, d.Bool.Flag()),
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
		conToken(Syntax_Token, l.Lambda.Flag()),
		conToken(Syntax_Token, l.DoubCol.Flag()),
		conToken(Data_Value_Token, d.Symbolic.Flag()),
		conToken(Syntax_Token, l.LeftArrow.Flag()),
		conToken(Data_Value_Token, d.Numeral.Flag()),
		conToken(Syntax_Token, l.FatLArrow.Flag()),
		conToken(Data_Value_Token, d.Int.Flag()),
		conToken(Syntax_Token, l.LeftArrow.Flag()),
		conToken(Data_Value_Token, d.Bool.Flag()),
	}

	ok = sliceContainsSignature(nomatch, ts)
	fmt.Println(ok)

	ok = sliceContainsSignature(ts[0], ts)
	fmt.Println(ok)
}
func TestDataEnclosures(t *testing.T) {
	data := con(d.Con("this is the testfunction speaking from within enclosure"))
	fmt.Println(data)
	fmt.Println(data.Type())
	fmt.Println(data.Flag())
}
func TestPairEnclosures(t *testing.T) {
	pair := conPair(d.Con("test key:"), d.Con("test data in a pair"))
	a, b := pair()
	fmt.Println(a)
	fmt.Println(b)
}
func TestVectorEnclosures(t *testing.T) {
	vec := conVector(
		d.Con("first data in slice"),
		d.Con("second data entry in slice"),
		d.Con("third data entry in slice"),
		d.Con("fourth data entry in slice"),
		d.Con("fifth data entry in slice"),
		d.Con("sixt data entry in slice"),
		d.Con("seventh data entry in slice"),
		d.Con("eigth data entry in slice"),
		d.Con("nineth data entry in slice"),
		d.Con("tenth data entry in slice"),
	)
	fmt.Println(vec.Flag())
	fmt.Println(vec.Type())
	fmt.Println(vec.Slice())
	fmt.Println(vec.String())

	vec1 := d.Con(0, 7, 45,
		134, 4, 465, 3, 645,
		2452, 34, 45, 3535,
		24, 4, 24, 2245,
		24, 42, 4, 24)
	fmt.Println(vec1.Flag())
	fmt.Println(vec1.String())
}
func TestParameterEnclosure(t *testing.T) {
	var dat Data
	parm := conParam(d.Con("test parameter"))
	dat, parm = parm()
	fmt.Println(dat)
	dat, parm = parm(con(d.Con("changer parameter")))
	dat, parm = parm()
	fmt.Println(dat)
	dat, parm = parm(con(d.Con("yet another parameter")))
	dat, parm = parm()
	fmt.Println(dat)
	fmt.Println(dat)
	dat, parm = parm(con(d.Con("yup, works just fine ;)")))
	fmt.Println(dat)
	fmt.Println(parm.Type())
	fmt.Println(dat.Flag())
}
