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

var maybeInt = NewMaybeTypeConstructor(NewPredicate(func(args ...Callable) bool {
	if args[0] != nil {
		return args[0].TypeNat().Match(d.Numbers)
	}
	return false
}))

var add = func(l, r d.Native) MaybeVal {
	var left, right d.Numeral
	if l.TypeNat() != d.Nil {
		left = l.Eval().(d.Numeral)
	}
	if r.TypeNat() != d.Nil {
		right = r.Eval().(d.Numeral)
	}
	return NewMaybeValue(DataVal(d.IntVal(left.Int() + right.Int()).Eval))
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

func TestTupleConstruction(t *testing.T) {
	var tupleCons = TupleTypeConstructor(
		DataVal(d.StrVal("field one").Eval),
		DataVal(d.StrVal("field two").Eval),
		DataVal(d.IntVal(42).Eval),
		DataVal(d.FltVal(23.42).Eval),
	)
	fmt.Printf("tuple type constructor: %s\n", tupleCons)
	var tupleVal = tupleCons(
		DataVal(d.StrVal("field one altered").Eval),
		DataVal(d.StrVal("field two altered").Eval),
		DataVal(d.IntVal(23).Eval),
		DataVal(d.FltVal(42.23).Eval),
	)
	fmt.Printf("altered tuple type fields: %s\n", tupleVal())
}

func TestRecordTypeConstruction(t *testing.T) {
	var recordType = NewRecordType(
		NewPair(DataVal(d.StrVal("arschloch").Eval), DataVal(d.StrVal("wichskrepel").Eval)),
		NewPair(DataVal(d.StrVal("kackscheisse").Eval), DataVal(d.StrVal("dreckskacke").Eval)),
	)
	fmt.Printf("record type: %s\n", recordType())
	var recordVal = recordType(
		NewPair(DataVal(d.StrVal("arschloch").Eval), DataVal(d.StrVal("fuckscheisse").Eval)),
		NewPair(DataVal(d.StrVal("kackscheisse").Eval), DataVal(d.StrVal("dreckskacke").Eval)),
	)
	fmt.Printf("record type: %s\n", recordVal())
}
