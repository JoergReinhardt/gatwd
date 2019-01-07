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
func TestEnclosures(t *testing.T) {
	data := con(d.Con("this is the testfunction speaking"))
	fmt.Println(data)
	fmt.Println(data.Type())
	fmt.Println(data.Flag())
}
