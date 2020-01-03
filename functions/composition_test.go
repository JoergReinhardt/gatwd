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
		head, tail = pair.Left().Call(Dat(13)), pair.Right().(ListVal)
		fmt.Printf("list called with 13: %s\n", head)
	}
}

func TestFoldSequence(t *testing.T) {
	var f = Fold(intsA, Dat(0),
		func(init, head Expression) Expression {
			return addInts(init, head)
		}).(ListVal)
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

	if ohead.(ListVal).Head().(DatConst)().(d.IntVal) != 7 ||
		ehead.(ListVal).Head().(DatConst)().(d.IntVal) != 6 {
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

	head, tail = head.(ListVal).Continue()
	for !tail.Empty() {
		head, tail = tail.(ListVal).Continue()
	}

	fmt.Printf("last elements head: %s\n", head.(VecVal).Head())
	if head.(VecVal).Head().(DatConst)().(d.IntVal) != 0 {
		t.Fail()
	}

	token = TakeN(intsA, 5)
	fmt.Printf("take five: %s\n", token)
	fmt.Printf("take five type: %s\n", token.Type())
	fmt.Printf("take five matches sequences: %t\n", token.Type().Match(Sequences))
}

func TestFlatttenSequence(t *testing.T) {
	fmt.Printf("take two: %s\n", token)
	var flat = Flatten(token)
	fmt.Printf("flattened list of lists: %s\n", flat)
	var head, tail = flat.Continue()
	if head.(DatConst)().(d.IntVal) != 0 {
		t.Fail()
	}
	for head, tail = tail.Continue(); tail.Empty(); {
	}
	if head.(DatConst)().(d.IntVal) != 1 {
		t.Fail()
	}
}

var zipped Grouped = Zip(abc, intsA, func(l, r Expression) Expression {
	return NewKeyPair(string(l.(DatConst)().(d.StrVal)), r)
})

func TestZipSequence(t *testing.T) {
	fmt.Printf("zipped: %s\nhead: %s\n", zipped, zipped.Head())
	if zipped.Head().(Paired).Key().String() != "a" {
		t.Fail()
	}
	for i := 0; i < 25; i++ {
		zipped = zipped.Tail().(Grouped)
	}
	if zipped.Head().(Paired).Key().String() != "z" {
		t.Fail()
	}
}

func TestSplitSequence(t *testing.T) {
	var splitted = Split(zipped, NewPair(NewVector(), NewVector()),
		func(pair Paired, head Expression) Paired {
			var (
				right = pair.Right().(VecVal)
				left  = pair.Left().(VecVal)
				val   = head.(KeyPair).Value()
				key   = head.(KeyPair).Key()
			)
			left = left.Cons(key).(VecVal)
			right = right.Cons(val).(VecVal)
			return NewPair(left, right)
		})
	fmt.Printf("splitted: %s\n", splitted)
	if splitted.Head().(Paired).Left().(VecVal).Head().String() != "z" {
		t.Fail()
	}
}

func TestBindSequence(t *testing.T) {

	var bound = Bind(intsA, intsB, func(l, r Expression, args ...Expression) Expression {
		if len(args) > 0 {
			return addInts(append([]Expression{
				addInts(l, r),
			}, args...)...)
		}
		return addInts(l, r)
	})
	fmt.Printf("bound without passing arguments: %s\n", bound)

	if bound.Head().(DatConst)().(d.IntVal) != 10 {
		t.Fail()
	}

	bound = Bind(bound, intsB, func(l, r Expression, args ...Expression) Expression {
		if len(args) > 0 {
			return addInts(append([]Expression{
				addInts(l, r),
			}, args...)...)
		}
		return addInts(l, r)
	})
	fmt.Printf("bound after binding ints-b again: %s\n", bound)

	if bound.Head().(DatConst)().(d.IntVal) != 20 {
		t.Fail()
	}
}

func TestSortSequence(t *testing.T) {
	var rndm = NewVector(randInts(20)...)
	fmt.Printf("random: %s\n", rndm)

	var sorted = Sort(rndm,
		func(l, r Expression) bool {
			return l.(DatConst)().(d.IntVal) <
				r.(DatConst)().(d.IntVal)
		})
	fmt.Printf("sorted: %s\n", sorted)
	fmt.Printf("concat a & b: %s\n", NewList(intsA()...).Concat(NewList(intsB()...)))

}
