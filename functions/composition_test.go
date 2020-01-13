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

	if m.Call().(Paired).Left().Call(Dat(13)).(AtomVal)().(d.IntVal) != 13 {
		t.Fail()
	}

	var (
		head Expression
		pair Paired
		tail = m
	)
	for !tail.Empty() {
		pair = tail.Call().(Paired)
		head, tail = pair.Left().Call(Dat(13)), pair.Right().(ListVal)
		fmt.Printf("list called with 13: %s\n", head)
	}
}

func TestFoldSequence(t *testing.T) {
	var f = Fold(intsA, Dat(0),
		func(init, head Expression) Expression {
			fmt.Printf("init: %s. head: %s\n", init, head)
			return addInts(init, head)
		})
	fmt.Printf("folded list: %s\n", f)

	var head, tail = f.Continue()
	for i := 0; i < 8; i++ {
		head, tail = tail.Continue()
	}
	fmt.Printf("head after eight continuations: %s\n", head)
	if head.(AtomVal)().(d.IntVal) != 36 {
		t.Fail()
	}
}

func TestFilterPassSequence(t *testing.T) {

	var (
		isEven = func(arg Expression) bool {
			return arg.(AtomVal)().(d.IntVal)%2 == 0
		}
		odd  = Filter(intsA, isEven)
		even = Pass(intsA, isEven)
	)

	var ohead, otail = odd.Continue()
	var ehead, etail = even.Continue()
	fmt.Printf("odd head: %s\neven head: %s\n", ohead, ehead)
	fmt.Printf("odd tail: %s\neven tail: %s\n", otail, etail)
	for i := 0; i < 3; i++ {
		ohead, otail = otail.Continue()
		ehead, etail = etail.Continue()
		fmt.Printf("odd head: %s\neven head: %s\n", ohead, ehead)
	}

	if ohead.(VecVal).Last().(AtomVal)().(d.IntVal) != 7 ||
		ehead.(VecVal).Last().(AtomVal)().(d.IntVal) != 6 {
		t.Fail()
	}
}

var token = TakeN(intsA, 2)

func TestTakeNSequence(t *testing.T) {

	fmt.Printf("take two: %s\n", token)

	var head, tail = token.Continue()
	fmt.Printf("head: %s\n", head)

	for !tail.Empty() {
		head, tail = tail.Continue()
	}
	fmt.Printf("last element: %s\n", head)

	head, tail = head.(Grouped).Continue()
	for !tail.Empty() {
		head, tail = tail.(Grouped).Continue()
	}

	fmt.Printf("last elements head: %s\n", head.(Grouped).Head())
	if head.(Grouped).Head().(AtomVal)().(d.IntVal) != 8 {
		t.Fail()
	}

	token = TakeN(intsA, 5)
	fmt.Printf("take five: %s\n", token)
	fmt.Printf("take five type: %s\n", token.Type())
	fmt.Printf("take five matches sequences: %t\n", token.Type().Match(Collections))

	token = TakeN(intsA, 4)
	fmt.Printf("take four: %s\n", token)
}

func TestFlatttenSequence(t *testing.T) {
	fmt.Printf("take two: %s\n", token)
	var flat = Flatten(token)
	fmt.Printf("flattened list of lists: %s\n", flat)
	var head, tail = flat.Continue()
	if head.(AtomVal)().(d.IntVal) != 0 {
		t.Fail()
	}
	for head, tail = tail.Continue(); tail.Empty(); {
	}
	if head.(AtomVal)().(d.IntVal) != 1 {
		t.Fail()
	}
}

var zipped Grouped = Zip(abc, intsA, func(l, r Expression) Expression {
	return NewKeyPair(string(l.(AtomVal)().(d.StrVal)), r)
})

func TestZipSequence(t *testing.T) {
	fmt.Printf("zipped: %s\nhead: %s\n", zipped, zipped.Head())
	if zipped.Head().(Paired).Key().String() != "a" {
		t.Fail()
	}
	var head, tail = zipped.Continue()
	for !tail.Empty() {
		head, tail = tail.Continue()
	}
	fmt.Printf("tail: %s\nhead: %s\n", tail, head)
	if head.(Paired).Key().String() != "z" {
		t.Fail()
	}
}

func TestSplitSequence(t *testing.T) {
	fmt.Printf("zipped: %s\n", zipped)
	var splitted = Split(zipped, func(arg Expression) Paired {
		return arg.(Paired)
	})
	fmt.Printf("splitted: %s\n", splitted)
}

func TestBindSequence(t *testing.T) {
}
