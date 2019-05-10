package functions

import (
	"fmt"
	d "github.com/joergreinhardt/gatwd/data"
	"testing"
)

var intslice = []Callable{
	New(0), New(1), New(2), New(3), New(4), New(5), New(6), New(7), New(8),
	New(9), New(0), New(134), New(8566735), New(4534), New(3445),
	New(76575), New(2234), New(45), New(7646), New(64), New(3), New(314),
}

var intkeys = []Callable{New("zero"), New("one"), New("two"), New("three"),
	New("four"), New("five"), New("six"), New("seven"), New("eight"), New("nine"),
	New("ten"), New("eleven"), New("twelve"), New("thirteen"), New("fourteen"),
	New("fifteen"), New("sixteen"), New("seventeen"), New("eighteen"),
	New("nineteen"), New("twenty"), New("twentyone"),
}

var maybeInt = NewMaybeConstructor(NewPredicate(func(arg Callable) bool {
	if arg != nil {
		return arg.TypeNat().Match(d.Numbers)
	}
	return false
}))

var add = func(l, r d.Native) JustVal {
	var left, right d.Numeral
	if l.TypeNat() != d.Nil {
		left = l.Eval().(d.Numeral)
	}
	if r.TypeNat() != d.Nil {
		right = r.Eval().(d.Numeral)
	}
	return maybeInt(DataVal(d.IntVal(left.Int() + right.Int()).Eval))().(JustVal)
}

func TestMaybe(t *testing.T) {
	fmt.Printf("maybeInt: %s\n", maybeInt)
	var intVal = maybeInt(DataVal(d.IntVal(42).Eval))
	var strVal = maybeInt(DataVal(d.StrVal("test string").Eval))
	var fltVal = maybeInt(DataVal(d.FltVal(42.23).Eval))
	var imagVal = maybeInt(DataVal(d.ImagVal(complex(7.121, 12.34)).Eval))
	fmt.Printf("maybe number int: %s\n", intVal())
	if intVal().Eval().(d.IntVal) != 42 {
		t.Fail()
	}
	fmt.Printf("maybe number string: %s\n", strVal())
	if !strVal().TypeFnc().Match(None) {
		t.Fail()
	}
	fmt.Printf("maybe number float: %s\n", fltVal())
	if fltVal().Eval().(d.FltVal) != 42.23 {
		t.Fail()
	}
	fmt.Printf("maybe number imaginary: %s\n", imagVal())
}

func TestMapMaybe(t *testing.T) {
	var mappedL = MapL(
		NewList(intslice...),
		Map(func(args ...Callable) Callable {
			if len(args) > 0 {
				return maybeInt(args[0])()
			}
			return maybeInt(NewNone())()
		}))

	var mappedR = MapL(
		NewList(intslice...),
		Map(func(args ...Callable) Callable {
			if len(args) > 0 {
				return maybeInt(args[0])()
			}
			return maybeInt(NewNone())()
		}))

	fmt.Printf("mapped maybe integers: %s\n", mappedL)

	var added = MapL(
		mappedL,
		Map(func(args ...Callable) Callable {
			var left, right Callable
			right, mappedR = mappedR()
			if len(args) > 0 {
				left = args[0]
			}
			if right != nil {
				return add(left.(JustVal), right.(JustVal))
			}
			return NewNone()
		}))

	_, _ = mappedR()

	fmt.Printf("operator added maybe integers: %s\n", added)
}
