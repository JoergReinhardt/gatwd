package functions

import (
	"fmt"
	"math/rand"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var (
	intEq = NewTest(d.Int, func(a, b Functor) bool {
		return a.(Evaluable).Eval().(d.IntVal) == b.(Evaluable).Eval().(d.IntVal)
	})
	rndms = func() VecVal {
		var rs = NewVector()
		for i := 0; i < 10; i++ {
			rs = rs.ConsVec(Box(d.IntVal(rand.Intn(10))))
		}
		return rs
	}()
)

func TestTestable(t *testing.T) {

	fmt.Printf("test: %s\n", intEq)
	fmt.Printf("test type: %s\n", intEq.Type())

	fmt.Printf("test zero is zero (true): %t\n", intEq.Test(Dat(0), Dat(0)))
	if !intEq.Test(Dat(0), Dat(0)) {
		t.Fail()
	}

	fmt.Printf("test one is zero (false): %t\n", intEq.Test(Dat(1), Dat(0)))
	if intEq.Test(Dat(1), Dat(0)) {
		t.Fail()
	}

	var eq = intEq.Equal()
	fmt.Printf("cast to type equal: %s\n", eq)
}

var compZero = NewComparator(d.Int, func(a, b Functor) int {
	var l = a.(Atom)().(d.IntVal)
	var r = b.(Atom)().(d.IntVal)
	switch {
	case l < r:
		return -1
	case l == r:
		return 0
	}
	return 1
})

func TestCompareable(t *testing.T) {
	fmt.Printf("compareable: %s\n", compZero)
	fmt.Printf("zero equals zero (0): %s\n", compZero(Dat(0), Dat(0)).String())
	fmt.Printf("minus one lesser zero (-1): %s\n", compZero(Dat(-1), Dat(0)))
	fmt.Printf("one greater zero (1): %s\n", compZero(Dat(1), Dat(0)))
	//
	//	fmt.Printf("0 == 0: %t\n", compZero.Equal(Dat(0), Dat(0)))
	//	fmt.Printf("0 == 1: %t\n", compZero.Equal(Dat(0), Dat(1)))

}

func TestCase(t *testing.T) {
	var (
		lesser = func(a, b Functor) bool {
			return a.(Atom)().(d.IntVal) <= b.(Atom)().(d.IntVal)
		}
		ascending = func(init, arg Functor) Functor {
			var (
				pair = init.(ValPair)
				vec  = pair.Left().(VecVal)
			)
			if IsNone(arg) {
				return NewPair(NewNone(), vec)
			}
			if vec.Empty() || lesser(vec.Last(), arg) {
				return NewPair(vec.Cons(arg), NewNone())
			}
			return NewPair(NewVector(arg), vec)
		}
		merge2 func(a, b Functor) Functor
	)
	merge2 = func(l, r Functor) Functor {
		var a, b = l.(VecVal), r.(VecVal)
		if a.Empty() {
			return b
		}
		if b.Empty() {
			return a
		}
		if lesser(a.Head(), b.Head()) {
			var head, tail = a.Continue()
			return GenVal(func() (Functor, GenVal) {
				return head, merge2(tail, b).(GenVal)
			})
		}
		var head, tail = b.Continue()
		return GenVal(func() (Functor, GenVal) {
			return head, merge2(tail, a).(GenVal)
		})
	}
	fmt.Printf("\nrandom numbers\n%s\n", rndms)

	var pairs = Fold(rndms, NewPair(NewVector(), NewNone()), ascending)
	fmt.Printf("\nfolded to pairs:\n%s\n", pairs)

	var ascends = Fold(pairs, NewVector(), func(init, arg Functor) Functor {
		return arg.(ValPair).Right()
	})
	fmt.Printf("\nreduced to left values:\n%s\n", ascends)

	var head Functor
	head, ascends = ascends()
	for !ascends.Empty() {
		fmt.Println(head)
		head, ascends = ascends()
	}
}

func TestSwitch(t *testing.T) {
}

func TestMaybe(t *testing.T) {
	var (
		intType = Dat(0).Type()
		def     = Define(Dat(func(args ...d.Native) d.Native {
			fmt.Println(args)
			return args[0].(d.IntVal) + args[1].(d.IntVal)
		}),
			DecSym("MaybeInt"),
			Declare(intType),
			Declare(intType, intType),
		)
		maybeInt = NewMaybe(def)
	)
	fmt.Println(def)
	fmt.Println(def.Type().TypeArgs())
	fmt.Println(def.Type().TypeId())
	fmt.Println(def.Type().TypeRet())
	var res = maybeInt.Call(Dat(2))
	fmt.Println(res)
	res = res.Call(Dat(20))
	fmt.Println(res.(Def).Unbox())
	fmt.Println(res.Type())

	res = maybeInt.Call(Dat("not an int"))
	fmt.Println(res)
	fmt.Println(res.Type())
}

func TestOption(t *testing.T) {
}

//func TestEnum(t *testing.T) {
//	var enumtype EnumCon
//	var weekdays = NewVector(
//		Dat("Monday"),
//		Dat("Tuesday"),
//		Dat("Wednesday"),
//		Dat("Thursday"),
//		Dat("Friday"),
//		Dat("Saturday"),
//		Dat("Sunday"),
//	)
//	enumtype = NewEnumType(func(day d.Numeral) Expression {
//		var idx = day.GoInt()
//		if idx > 6 {
//			idx = idx%6 - 1
//		}
//		return weekdays()[idx]
//	})
//
//	fmt.Printf("enum type days of the week: %s type: %s\n", enumtype, enumtype.Type().TypeName())
//	var enum = enumtype(d.IntVal(8))
//	fmt.Printf("wednesday eum: %s\n", enum)
//	fmt.Printf("eum expr: %s\n", enum.Type())
//	var val, idx, _ = enum()
//	fmt.Printf("enum value val %s, index: %s\n",
//		val, idx)
//}
