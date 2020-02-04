package functions

import (
	"fmt"
	"math/rand"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var (
	intEq = NewTest(d.Int, func(args ...Functor) bool {
		return args[0].(Evaluable).Eval().(d.IntVal) ==
			args[1].(Evaluable).Eval().(d.IntVal)
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

var compZero = NewComparator(Dat(0).Type(), func(a, b Functor) int {
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
	fmt.Printf("zero equals zero (0): %s\n", compZero(Dat(0), Dat(0)))
	if !compZero(Dat(0), Dat(0)).(TyFnc).Match(Equal) {
		t.Fail()
	}
	fmt.Printf("minus one lesser zero (-1): %s\n", compZero(Dat(-1), Dat(0)))
	if !compZero(Dat(-1), Dat(0)).(TyFnc).Match(Lesser) {
		t.Fail()
	}
	fmt.Printf("one greater zero (1): %s\n", compZero(Dat(1), Dat(0)))
	if !compZero(Dat(1), Dat(0)).(TyFnc).Match(Greater) {
		t.Fail()
	}
	fmt.Printf("0 == 0: %t\n", compZero.Equal(Dat(0), Dat(0)))
	if !compZero.Equal(Dat(0), Dat(0)) {
		t.Fail()
	}
	fmt.Printf("0 == 1: %t\n", compZero.Equal(Dat(0), Dat(1)))
	if compZero.Equal(Dat(0), Dat(1)) {
		t.Fail()
	}
	fmt.Printf("0 < 0: %t\n", compZero.Lesser(Dat(0), Dat(0)))
	if compZero.Lesser(Dat(0), Dat(0)) {
		t.Fail()
	}
	fmt.Printf("0 < 1: %t\n", compZero.Lesser(Dat(0), Dat(1)))
	if !compZero.Lesser(Dat(0), Dat(1)) {
		t.Fail()
	}
	fmt.Printf("0 > 0: %t\n", compZero.Greater(Dat(0), Dat(0)))
	if compZero.Greater(Dat(0), Dat(0)) {
		t.Fail()
	}
	fmt.Printf("1 > 0: %t\n", compZero.Greater(Dat(1), Dat(0)))
	if !compZero.Greater(Dat(1), Dat(0)) {
		t.Fail()
	}

}

func TestPolymorph(t *testing.T) {
	fmt.Printf("addFloats type: %s\n", addFloats.Type())
	fmt.Printf("addFloats arg types: %s\n", addFloats.TypeArgs())
	fmt.Printf("addFloats return types: %s\n", addFloats.TypeRet())

	fmt.Printf("addInts type: %s\n", addInts.Type())
	fmt.Printf("addInts arg types: %s\n", addInts.TypeArgs())
	fmt.Printf("addInts return types: %s\n", addInts.TypeRet())

	var poly = NewPolyMorph(DecSym("+"), addInts, addUints, addFloats)

	fmt.Printf("polymorh definition:\n%s\n", poly.Type())
	fmt.Printf("polymorh adding ints 10 + 22: %s\n", poly.Call(Dat(11), Dat(22)))
	if poly.Call(Dat(11), Dat(22)).(Atom)().(d.IntVal) != 33 {
		t.Fail()
	}
	fmt.Printf("polymorh adding uints 11 + 22: %s\n", poly.Call(Dat(uint(11)), Dat(uint(22))))
	if poly.Call(Dat(uint(11)), Dat(uint(22))).(Atom)().(d.UintVal) !=
		Dat(uint(33)).(Atom)().(d.UintVal) {
		t.Fail()
	}
	fmt.Printf("polymorh adding floats 1.1 + 2.2: %s\n", poly.Call(Dat(1.1), Dat(2.2)))
	if poly.Call(Dat(1.1), Dat(2.2)).(Atom)().(d.FltVal) != 3.3000000000000003 {
		t.Fail()
	}

	var part = poly.Call(Dat(11))
	fmt.Printf("partialy applyed polymorph: %s\n", part)
	fmt.Printf("partialy applyed polymorphs type: %s\n", part.Type())
	fmt.Printf("partialy applyed polymorphs type function: %s\n", part.TypeFnc().TypeName())
	if !part.TypeFnc().Match(Type | Partial) {
		t.Fail()
	}

	part = part.Call(Dat(22))
	fmt.Printf("fully applyed polymorph: %s\n", part)
	fmt.Printf("fully applyed polymorphs type: %s\n", part.Type())
	fmt.Printf("fully applyed polymorphs type function: %s\n", part.TypeFnc())
	if !part.TypeFnc().Match(Data) {
		t.Fail()
	}

	var wrong = poly.Call(Dat(10), Dat(1.1))
	fmt.Printf("apply incompatible args int 10 + float 1.1: %s\n", wrong)
	if !poly.Call(Dat(10), Dat(1.1)).Type().Match(None) {
		t.Fail()
	}
}

var opt = NewOption(Define(Lambda(func(args ...Functor) Functor {
	if args[0].(Atom)().(d.IntVal) > 0 {
		return Box(d.StrVal("positive"))
	}
	return NewNone()
}), DecSym("positiveOrNone"), Dat("").Type(), Dat(0).Type()))

func TestOption(t *testing.T) {

	//fmt.Printf("optional: %s\n", opt.TypeName())

	var o = opt.Call(Dat(-1))
	fmt.Printf("negative → none: %s\n", o)
	if !o.Type().Match(None) {
		t.Fail()
	}

	o = opt.Call(Dat(1))
	fmt.Printf("positive → string: %s\n", o)
	fmt.Printf("o type: %s\n", o.Type())
	if !o.Type().Match(Dat("").Type()) {
		t.Fail()
	}
}

var alt = NewAlternative(
	Define(Lambda(func(args ...Functor) Functor {
		if args[0].(Atom)().(d.IntVal) > 0 {
			return Box(d.StrVal("positive"))
		}
		return NewNone()
	}), DecSym("positiveOrNone"), Dat("").Type(), Dat(0).Type()),
	Define(Lambda(func(args ...Functor) Functor {
		if args[0].(Atom)().(d.IntVal) < 0 {
			return Box(d.StrVal("negative"))
		}
		return NewNone()
	}), DecSym("negativeOrNone"), Dat("").Type(), Dat(0).Type()),
)

func TestAlternative(t *testing.T) {

	fmt.Printf("alternative: %s\n", alt)

	var a = alt.Call(Dat(-1))
	fmt.Printf("alt called with -1: %s\n", a)

	a = alt.Call(Dat(1))
	fmt.Printf("alt called with 1: %s\n", a)

	var rec = NewRecordCon(Define(alt), Define(opt))
	fmt.Printf("record??? %s\n", rec.Call(Dat(10), Dat(11)))
}
