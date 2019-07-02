package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var intslice = []Expression{
	New(0), New(1), New(2), New(3), New(4), New(5), New(6), New(7), New(8),
	New(9), New(0), New(134), New(8566735), New(4534), New(3445),
	New(76575), New(2234), New(45), New(7646), New(64), New(3), New(314),
}

var intkeys = []Expression{New("zero"), New("one"), New("two"), New("three"),
	New("four"), New("five"), New("six"), New("seven"), New("eight"), New("nine"),
	New("ten"), New("eleven"), New("twelve"), New("thirteen"), New("fourteen"),
	New("fifteen"), New("sixteen"), New("seventeen"), New("eighteen"),
	New("nineteen"), New("twenty"), New("twentyone"),
}

func TestNary(t *testing.T) {
	var nary = NewNary(
		VariadicExpr(func(args ...Expression) Expression {
			return NewVector(args...)
		}), Type, Type, Type)

	var r0 = nary(NewNative(d.StrVal("0")))

	var r1 = nary(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")))

	var r2 = nary(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")),
		NewNative(d.StrVal("2")))

	var r3 = nary(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")),
		NewNative(d.StrVal("2")), NewNative(d.StrVal("3")))

	var r4 = nary(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")),
		NewNative(d.StrVal("2")), NewNative(d.StrVal("3")),
		NewNative(d.StrVal("4")))

	var r5 = nary(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")),
		NewNative(d.StrVal("2")), NewNative(d.StrVal("3")),
		NewNative(d.StrVal("4")), NewNative(d.StrVal("5")))

	var r6 = nary(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")),
		NewNative(d.StrVal("2")), NewNative(d.StrVal("3")),
		NewNative(d.StrVal("4")), NewNative(d.StrVal("5")),
		NewNative(d.StrVal("6")))

	var r7 = nary(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")),
		NewNative(d.StrVal("2")), NewNative(d.StrVal("3")),
		NewNative(d.StrVal("4")), NewNative(d.StrVal("5")),
		NewNative(d.StrVal("6")), NewNative(d.StrVal("7")))

	var r8 = nary(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")),
		NewNative(d.StrVal("2")), NewNative(d.StrVal("3")),
		NewNative(d.StrVal("4")), NewNative(d.StrVal("5")),
		NewNative(d.StrVal("6")), NewNative(d.StrVal("7")),
		NewNative(d.StrVal("8")))

	fmt.Println(r0.Call())
	fmt.Println(r1.Call())
	fmt.Println(r2.Call())
	fmt.Println(r3.Call())
	fmt.Println(r4.Call())
	fmt.Println(r5.Call())
	fmt.Println(r6.Call())
	fmt.Println(r7.Call())
	fmt.Println(r8.Call())

	// apply additional arguments to partialy applyed expression
}

func TestTuple(t *testing.T) {
}

func TestRecordType(t *testing.T) {
}
