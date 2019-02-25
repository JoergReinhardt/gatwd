package functions

import (
	"fmt"
	"testing"

	d "github.com/JoergReinhardt/gatwd/data"
)

var pred = NewPredicate(func(scrut ...Functional) bool {
	if scrut[0].TypeNat().Flag().Match(d.Int) {
		return true
	}
	return false
})

func TestOptionalTypes(t *testing.T) {

	var result = pred(New(12))

	fmt.Printf("test predicate with int instance: %t\n", result)
	if !result {
		t.Fail()
	}

	result = pred(New("stringy"))
	fmt.Printf("test predicate with string instance: %t\n", result)
	if result {
		t.Fail()
	}
}

func TestMaybeTypes(t *testing.T) {

	var m = NewMaybe(pred, New("this is the expression to return"))

	m = NewMaybe(pred, New(12))

	var result = m(New(1))
	fmt.Printf(
		"applying int to a maybe instance expecting an int: %t %s\n",
		result.Maybe(), result.Value())
	if !result.Maybe() {
		t.Fail()
	}

	result = m(New("not int"))
	fmt.Printf(
		"applying int to a maybe instance expecting an int: %t %s\n",
		result.Maybe(), result.Value())
	if result.Maybe() {
		t.Fail()
	}
}

func TestCaseFncTypes(t *testing.T) {

	var cf = NewCaseFnc(
		NewCaseExpr(pred, New("expression 0")),
		NewCaseExpr(pred, New("expression 1")),
		NewCaseExpr(pred, New("expression 2")),
		NewCaseExpr(pred, New("expression 3")),
		NewCaseExpr(pred, New("expression 4")),
	)

	var result = cf(New("fick dir ins knie"))

	fmt.Printf("case expression result: %s\n", result)

	if result.Maybe() {
		t.Fail()
	}

	var pred0 = NewPredicate(func(scrut ...Functional) bool {
		if scrut[0].TypeNat().Flag().Match(d.Int) {
			return true
		}
		return false
	})
	var pred1 = NewPredicate(func(scrut ...Functional) bool {
		if scrut[0].TypeNat().Flag().Match(d.Uint) {
			return true
		}
		return false
	})
	var pred2 = NewPredicate(func(scrut ...Functional) bool {
		if scrut[0].TypeNat().Flag().Match(d.Byte) {
			return true
		}
		return false
	})
	var pred3 = NewPredicate(func(scrut ...Functional) bool {
		if scrut[0].TypeNat().Flag().Match(d.Bytes) {
			return true
		}
		return false
	})
	var pred4 = NewPredicate(func(scrut ...Functional) bool {
		if scrut[0].TypeNat().Flag().Match(d.Float) {
			return true
		}
		return false
	})

	cf = NewCaseFnc(
		NewCaseExpr(pred0, New("Int 0")),
		NewCaseExpr(pred1, New("Uint 1")),
		NewCaseExpr(pred2, New("Byte 2")),
		NewCaseExpr(pred3, New("Bytes 3")),
		NewCaseExpr(pred4, New("Float 4")),
	)

	result = cf(New(12.22))

	fmt.Printf("banöbelörbelpröbenrörp: %s\n", result.String())

	if result.String() == "Float 4" {
		t.Fail()
	}
}
