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
