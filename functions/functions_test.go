package functions

import (
	"fmt"
	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lang"
	"testing"
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
			newToken(Data_Type_Token, d.Numeral),
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
			newToken(Data_Type_Token, d.Numeral),
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
			newToken(Data_Type_Token, d.Numeral),
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
		newToken(Data_Value_Token, d.Numeral),
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
	fmt.Println(data.Type())
	fmt.Println(data.Flag())
}
func TestPairEnclosures(t *testing.T) {
	pair := newPair(d.New("test key:"), d.New("test data in a pair"))
	a, b := pair()
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
	parm := newArgument(d.New("test parameter"))
	dat, parm = parm()
	fmt.Println(dat)
	dat, parm = parm(newData(d.New("changer parameter")))
	dat, parm = parm()
	fmt.Println(dat)
	dat, parm = parm(newData(d.New("yet another parameter")))
	dat, parm = parm()
	fmt.Println(dat)
	fmt.Println(dat)
	dat, parm = parm(newData(d.New("yup, works just fine ;)")))
	fmt.Println(dat)
	fmt.Println(parm.Type())
	fmt.Println(dat.Flag())
}
func TestAccParamEnclosure(t *testing.T) {
	acc := newAccAttribute(newPair(d.New("test-key"), d.New("test value")))
	fmt.Println(acc)
	_, acc = acc(newPair(d.New(12), d.New("one million dollar")))
	fmt.Println(acc)
	if acc.Key() != d.New(12) {
		t.Fail()
	}
	_, acc = acc(newPair(d.New(13), d.New("two million dollar")))
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
	args := newArgSet(d.New(0), d.New(1), d.New(2), d.New(3), d.New(4), d.New(5))
	fmt.Println(args)
}
