package functions

import (
	"fmt"
	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lang"
	"testing"
)

func TestIdGenerator(t *testing.T) {
	var id int
	id = conId()
	fmt.Println(id)
	id = conId()
	fmt.Println(id)
	if id != 1 {
		t.Fail()
	}
	id = conId()
	id = conId()
	id = conId()
	id = conId()
	id = conId()
	fmt.Println(id)
	if id != 6 {
		t.Fail()
	}
}
func TestSliceMatch(t *testing.T) {
	ts := [][]Token{
		[]Token{
			conToken(Syntax, l.Lambda.Flag()),
			conToken(Syntax, l.DoubCol.Flag()),
			conToken(Data_Type, d.Bool.Flag()),
			conToken(Syntax, l.LeftArrow.Flag()),
			conToken(Data_Type, d.Bool.Flag()),
		},
		[]Token{
			conToken(Syntax, l.Lambda.Flag()),
			conToken(Syntax, l.DoubCol.Flag()),
			conToken(Data_Type, d.Bool.Flag()),
			conToken(Syntax, l.LeftArrow.Flag()),
			conToken(Data_Type, d.Bool.Flag()),
		},
		[]Token{
			conToken(Syntax, l.Lambda.Flag()),
			conToken(Syntax, l.DoubCol.Flag()),
			conToken(Data_Type, d.Slice.Flag()),
			conToken(Syntax, l.LeftArrow.Flag()),
			conToken(Data_Type, d.Numeral.Flag()),
			conToken(Syntax, l.LeftArrow.Flag()),
			conToken(Data_Type, d.Int.Flag()),
			conToken(Syntax, l.LeftArrow.Flag()),
			conToken(Data_Type, d.Bool.Flag()),
		},
		[]Token{
			conToken(Syntax, l.Lambda.Flag()),
			conToken(Syntax, l.DoubCol.Flag()),
			conToken(Data_Type, d.Slice.Flag()),
			conToken(Syntax, l.LeftArrow.Flag()),
			conToken(Data_Type, d.Numeral.Flag()),
			conToken(Syntax, l.LeftArrow.Flag()),
			conToken(Data_Type, d.Int.Flag()),
			conToken(Syntax, l.LeftArrow.Flag()),
			conToken(Data_Type, d.Bool.Flag()),
		},
		[]Token{
			conToken(Syntax, l.Lambda.Flag()),
			conToken(Syntax, l.DoubCol.Flag()),
			conToken(Data_Type, d.Symbolic.Flag()),
			conToken(Syntax, l.LeftArrow.Flag()),
			conToken(Data_Type, d.Numeral.Flag()),
			conToken(Syntax, l.LeftArrow.Flag()),
			conToken(Data_Type, d.Int.Flag()),
			conToken(Syntax, l.LeftArrow.Flag()),
			conToken(Data_Type, d.Bool.Flag()),
		},
	}

	sortTokens(ts)

	ok := smatch(ts[0], ts[1])
	fmt.Println(Tokens(ts[0]))
	fmt.Println(Tokens(ts[1]))
	fmt.Println(ok)
	if !ok {
		t.Fail()
	}

	ok = sigsMatch(ts[0], ts[2])
	fmt.Println(Tokens(ts[0]))
	fmt.Println(Tokens(ts[2]))
	fmt.Println(ok)
	if ok {
		t.Fail()
	}

	ok = sigsMatch(ts[2], ts[3])
	fmt.Println(Tokens(ts[2]))
	fmt.Println(Tokens(ts[3]))
	fmt.Println(ok)
	if !ok {
		t.Fail()
	}

	ok = sigsMatch(ts[3], ts[4])
	fmt.Println(Tokens(ts[3]))
	fmt.Println(Tokens(ts[4]))
	fmt.Println(ok)
	if ok {
		t.Fail()
	}

}
