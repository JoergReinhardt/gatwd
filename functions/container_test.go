package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
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

var f = VariadicExpr(func(args ...Callable) Callable {
	var str = "f and "
	str = str + args[0].String()
	return NewNative(d.StrVal(str))
})
var g = VariadicExpr(func(args ...Callable) Callable {
	var str = "g and "
	str = str + args[0].String()
	return NewNative(d.StrVal(str))
})
var h = VariadicExpr(func(args ...Callable) Callable {
	var str = "h and "
	str = str + args[0].String()
	return NewNative(d.StrVal(str))
})
var i = VariadicExpr(func(args ...Callable) Callable {
	var str = "i and "
	str = str + args[0].String()
	return NewNative(d.StrVal(str))
})
var j = VariadicExpr(func(args ...Callable) Callable {
	var str = "j and "
	str = str + args[0].String()
	return NewNative(d.StrVal(str))
})
var k = ConstantExpr(func() Callable {
	return NewNative(d.StrVal("k"))
})

func TestCurry(t *testing.T) {
	var result = Curry(f, g, h, i, j, k)
	fmt.Println(result)
	if result.String() != "f and g and h and i and j and k" {
		t.Fail()
	}
}

func TestNary(t *testing.T) {
	var nary = NewNary(
		VariadicExpr(
			func(args ...Callable) Callable {
				return NewVector(args...)
			}),
		NewComposedType(
			"StringTriple",
			Data,
			d.String,
			Functor, // ← divides arguments from return values
			Vector,
			Data,
			d.String,
		), 3)

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
	var partial = r6.(VecVal)()[r6.(VecVal).Len()-1].Call(
		NewNative(d.StrVal("7")))

	fmt.Printf("partialy applyed narys remaining arity: %d\n",

		partial.(NaryExpr).Arity())
	partial = partial.Call(NewNative(d.StrVal("8")))

	fmt.Println(partial.Call())
}

func TestSwitch(t *testing.T) {
	var swi = NewSwitch(
		// matches return values native types string,integer, and float
		NewCase(NewPredictAll(func(arg Callable) bool {
			return arg.TypeNat().Match(d.String | d.Integers | d.Float)
		})))
	fmt.Printf("switch: %s\n", swi.Call(New(23), New(42, 23)))
	fmt.Printf("successfull call to Switch passing int: %s\n", swi.Call(New(42)))

	if val := swi.Call(New(42)); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("successfull call to Switch passing float: %s\n", swi.Call(New(23.42)))
	if val := swi.Call(New(23.42)); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("successfull call to Switch passing string: %s\n", swi.Call(New("string")))
	if val := swi.Call(New(23.42)); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("successfull call to Switch passing multiple integers: %s\n",
		swi.Call(New(23), New(42), New(65)))
	if val := swi.Call(New(23), New(42), New(65)); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("successfull call to Switch passing mixed args: %s\n",
		swi.Call(New(23), New(42.23), New("string")))
	if val := swi.Call(New(23), New(42.23), New("string")); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("unsuccessfull call to Switch passing boolean: %s\n", swi.Call(New(true)))
	if val := swi.Call(New(true)); !val.TypeFnc().Match(None) {
		t.Fail()
	}
}

func TestMaybeType(t *testing.T) {

	var swi = NewSwitch(NewCase(NewPredictAll(func(arg Callable) bool {
		return arg.TypeNat().Match(d.Numbers)
	})))

	fmt.Printf("test switch passing an int: %s\n", swi.Call(New(17)))
	fmt.Printf("test switch passing a string: %s\n", swi.Call(New("string")))
	fmt.Printf("test switch passing multiple ints: %s\n", swi.Call(New(7), New(17)))

	var maybeNumber = DefineMaybeType(swi)

	fmt.Printf("composed Type: %s TypeName: %s FncType: %s NativeType: %s\n",
		maybeNumber, maybeNumber.TypeName(),
		maybeNumber.TypeFnc().TypeName(),
		maybeNumber.TypeNat().TypeName())

	var number = maybeNumber(New(13))
	var nan = maybeNumber(New("string"))

	fmt.Printf("data instances number: %s, nan: %s\n", number, nan)

	var add = NewNary(func(args ...Callable) Callable {
		var numbers = NewVector()
		for _, arg := range args {
			var maybe = maybeNumber(arg)
			if maybe.TypeFnc().Match(Just) {
				numbers = numbers.Append(maybe.Call())
			}
		}
		return NewNative(args[0].Eval().(d.IntVal) + args[1].Eval().(d.IntVal))
	}, NewComposedType("Add", d.Numbers), 2)

	var result = add(New(3), New(5), New(7), New(11), New(13), New(17))

	fmt.Printf("addition results: %s\n", result)

	if result.(VecVal)()[0].Eval().(d.IntVal) != 8 {
		t.Fail()
	}
}

func TestTupleConstruction(t *testing.T) {
}

func TestRecordTypeConstruction(t *testing.T) {
}
