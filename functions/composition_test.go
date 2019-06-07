package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var listA = NewVector(New(0), New(1), New(2), New(3),
	New(4), New(5), New(6), New(7), New(8), New(9))

var listB = NewVector(New(10), New(11), New(12), New(13),
	New(14), New(15), New(16), New(17), New(18), New(19))

func conList(args ...Callable) Consumeable {
	return NewList(args...)
}
func printCons(cons Consumeable) {
	var head, tail = cons.Consume()
	if head != nil {
		fmt.Println(head)
		printCons(tail)
	}
}

func TestList(t *testing.T) {
	var list = NewList(listA()...)
	printCons(list)
}

func TestConList(t *testing.T) {

	var alist = NewList(listA()...)
	var head Callable

	for i := 0; i < 5; i++ {
		head, alist = alist()
		fmt.Println("for loop: " + head.String())
	}

	alist = alist.Cons(listB()...)

	printCons(alist)
}

func TestPushList(t *testing.T) {

	var alist = NewList(listA()...)
	var head Callable

	for i := 0; i < 5; i++ {
		head, alist = alist()
		fmt.Println("for loop: " + head.String())
	}

	alist = alist.Push(listB()...)

	printCons(alist)
}

func TestListMapF(t *testing.T) {

	var list = NewList(listA()...)
	var fmap = func(args ...Callable) Callable {
		return New(args[0].Eval().(d.IntVal).Int() * 3)
	}

	var mapped = MapL(list, fmap)

	printCons(mapped)
}

func TestListFoldF(t *testing.T) {

	var list = NewList(listA()...)
	var fold = Fold(func(ilem, head Callable, args ...Callable) Callable {
		return New(ilem.Eval().(d.IntVal) + head.Eval().(d.IntVal))
	})
	var ilem = New(0)

	var folded = FoldL(list, ilem, fold)

	printCons(folded)
}

func TestListFoldAndMap(t *testing.T) {

	var list = NewList(listA()...)
	var elem = New(0)
	var fold = func(elem, head Callable, args ...Callable) Callable {
		return New(elem.Eval().(d.IntVal) + head.Eval().(d.IntVal))
	}
	var fmap = func(args ...Callable) Callable {
		return New(args[0].Eval().(d.IntVal).Int() * 3)
	}

	var mapped = MapL(list, fmap)
	var folded = FoldL(mapped, elem, fold)

	printCons(folded)

	folded = FoldL(list, elem, fold)
	mapped = MapL(folded, fmap)

	var head, result Callable
	head, mapped = mapped()

	for {
		fmt.Println(head)
		head, mapped = mapped()
		if head == nil {
			break
		}
		result = head
	}

	if result.Eval().(d.IntVal) != 135 {
		t.Fail()
	}
}

func TestConsumeableFoldAndMap(t *testing.T) {

	var vec = listA
	var elem = New(0)
	var fold = func(elem, head Callable, args ...Callable) Callable {
		return New(elem.Eval().(d.IntVal) + head.Eval().(d.IntVal))
	}
	var fmap = func(args ...Callable) Callable {
		return New(args[0].Eval().(d.IntVal).Int() * 3)
	}

	var mapped = MapC(vec, fmap)
	var folded = FoldF(mapped, elem, fold)

	folded = FoldF(vec, elem, fold)
	mapped = MapF(folded, fmap)

	var head, result Callable
	head, mapped = mapped()

	for {
		fmt.Println(head)
		head, mapped = mapped()
		if head == nil {
			break
		}
		result = head
	}

	if result.Eval().(d.IntVal) != 135 {
		t.Fail()
	}
}

var keys = []Callable{New("zero"), New("one"), New("two"), New("three"),
	New("four"), New("five"), New("six"), New("seven"), New("eight"), New("nine"),
	New("ten")}

var vals = []Callable{New(0), New(1), New(2), New(3), New(4), New(5), New(6),
	New(7), New(8), New(9), New(10)}

func TestZipLists(t *testing.T) {
	var zipped = ZipL(NewList(keys...), NewList(vals...), func(l, r Callable) Paired { return NewPair(l, r) })
	fmt.Printf("zipped list: %s\n", zipped)
}

func TestZipConsumeable(t *testing.T) {
	var zipped = ZipF(NewList(keys...), NewList(vals...), func(l, r Callable) Paired { return NewPair(l, r) })

	var head, tail = zipped.Consume()
	for head != nil {
		fmt.Printf("%s, ", head)
		head, tail = tail.Consume()
	}
}

func TestFilterList(t *testing.T) {
	var filtered = FilterL(NewList(vals...), Filter(func(head Callable, args ...Callable) bool {
		if (head.Eval().(d.IntVal) % 2) == 0 {
			return true
		}
		return false
	}))

	var head, tail = filtered()
	for head != nil {
		fmt.Printf("filtered element: %s\n", head)
		head, tail = tail()
	}
}

func TestFilterConsumeable(t *testing.T) {
	var filtered = FilterF(NewList(vals...), Filter(func(head Callable, args ...Callable) bool {
		if (head.Eval().(d.IntVal) % 2) == 0 {
			return true
		}
		return false
	}))

	var head, tail = filtered.Consume()
	for head != nil {
		fmt.Printf("filtered element: %s\n", head)
		head, tail = tail.Consume()
	}
}

func TestBindF(t *testing.T) {
	var bind = func(f, g Callable) Callable {
		if nf, ok := f.Eval().(d.Numeral); ok {
			if ng, ok := g.Eval().(d.Numeral); ok {
				return NewData(d.IntVal(nf.Int() * ng.Int()))
			}
		}
		return nil
	}
	var bound = BindF(listA, listB, bind)
	var head Callable
	head, bound = bound()
	if head.Eval().(d.IntVal) != 0 {
		t.Fail()
	}
	for head != nil {
		fmt.Printf("%s\n", head)
		head, bound = bound()
	}
}
