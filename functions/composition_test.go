package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

func TestMapSequence(t *testing.T) {

	var m = Map(intsA, func(arg Expression) Expression {
		return addInts(arg)
	})

	fmt.Printf("list-a mapped to add-ints: %s\n", m)

	var head, tail = m.Continue()
	for !tail.Empty() {
		fmt.Printf("expression called on 10: %s\n", head.Call(Dat(10)))
		head, tail = tail.Continue()
	}

	fmt.Printf("list-a mutated?: %s\n", m)

	m = Map(m, func(arg Expression) Expression {
		var result = arg.Call(Dat(10))
		return result
	})

	fmt.Printf("mapped list-a mapped to add-10: %s\n", m)
}

func TestApplySequence(t *testing.T) {

	var m = Apply(intsA, func(head Expression, args ...Expression) Expression {
		return addInts(append([]Expression{head}, args...)...)
	})

	fmt.Printf("add-ints applyed to list-a: %s\n", m)

	if m.Call().(Paired).Left().Call(Dat(13)).(DatConst)().(d.IntVal) != 13 {
		t.Fail()
	}

	var (
		head Expression
		pair Paired
		tail = m
	)
	for !tail.Empty() {
		pair = tail.Call().(Paired)
		head, tail = pair.Left().Call(Dat(13)), pair.Right().(SeqVal)
		fmt.Printf("list called with 13: %s\n", head)
	}
}

func TestFoldSequence(t *testing.T) {
	var f = Fold(intsA, Dat(0), func(init, head Expression) Expression {
		return addInts(init, head)
	})
	fmt.Printf("folded list: %s\n", f)

	var head, tail = f.Continue()
	for i := 0; i < 8; i++ {
		head, tail = tail.Continue()
	}
	fmt.Printf("head after eight continuations: %s\n", head)
	if head.(DatConst)().(d.IntVal) != 36 {
		t.Fail()
	}
}

func TestFilterPassSequence(t *testing.T) {

	var (
		isEven = func(arg Expression) bool {
			return arg.(DatConst)().(d.IntVal)%2 == 0
		}
		odd  = Filter(intsA, isEven)
		even = Pass(intsA, isEven)
	)
	fmt.Printf("odd: %s\neven: %s\n", odd, even)

	var ohead, otail = odd.Continue()
	var ehead, etail = even.Continue()
	for i := 0; i < 3; i++ {
		ohead, otail = otail.Continue()
		ehead, etail = etail.Continue()
		fmt.Printf("odd head: %s\neven head: %s\n", ohead, ehead)
	}

	if ohead.(SeqVal).Head().(DatConst)().(d.IntVal) != 7 ||
		ehead.(SeqVal).Head().(DatConst)().(d.IntVal) != 6 {
		t.Fail()
	}
}

func TestTakeNSequence(t *testing.T) {

	var token = TakeN(intsA, 2)
	fmt.Printf("take two: %s\n", token)

	var head, tail = token.Continue()
	fmt.Printf("head: %s\n", head)

	for !tail.Empty() {
		head, tail = tail.Continue()
	}
	fmt.Printf("last element: %s\n", head)

	head, tail = head.(SeqVal).Continue()
	for !tail.Empty() {
		head, tail = tail.(SeqVal).Continue()
	}

	fmt.Printf("last elements head: %s\n", head.(VecVal).Head())
	if head.(VecVal).Head().(DatConst)().(d.IntVal) != 0 {
		t.Fail()
	}

	token = TakeN(intsA, 5)
	fmt.Printf("take five: %s\n", token)
}

func TestZipSequence(t *testing.T) {
}

func TestBindSequence(t *testing.T) {
}
