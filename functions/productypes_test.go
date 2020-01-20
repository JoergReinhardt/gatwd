package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var intEq = NewTest(d.Int, func(a, b Functor) bool {
	return a.(Evaluable).Eval().(d.IntVal) == b.(Evaluable).Eval().(d.IntVal)
})

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
	var r = a.(Atom)().(d.IntVal)
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

	var eq = compZero.Equal()
	fmt.Printf("equal: %s\n", eq.String())
	fmt.Printf("equal type args: %s\n", eq.Type().TypeArgs())
	fmt.Printf("0 == 0: %s\n", eq.Call(Dat(0), Dat(0)))
	fmt.Printf("0 == 1: %s\n", eq.Call(Dat(0), Dat(1)))

}

func TestCase(t *testing.T) {
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
