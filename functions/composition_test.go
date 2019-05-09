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
	var fold = FoldFExpr(func(ilem, head Callable, args ...Callable) Callable {
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

	var mapped = MapF(vec, fmap)
	var folded = FoldF(mapped, elem, fold)

	printCons(folded)

	folded = FoldF(vec, elem, fold)
	mapped = MapF(folded, fmap)

	var head, result Callable
	head, mapped = mapped.Consume()

	for {
		fmt.Println(head)
		head, mapped = mapped.Consume()
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
	var filtered = FilterL(NewList(vals...), FilterFExpr(func(head Callable, args ...Callable) bool {
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
	var filtered = FilterF(NewList(vals...), FilterFExpr(func(head Callable, args ...Callable) bool {
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

var numFunc = NewFunctor(NewFromData(d.IntVal(42)))
var numApp = NewApplicable(numFunc, func(expr NaryExpr, args ...Callable) Callable {
	if len(args) > 0 {
		if len(args) > 1 {
			var numbers = []Callable{}
			for _, arg := range args {
				if arg.TypeNat().Match(d.Numbers) {
					if arg.TypeFnc().Match(Functors) {
						numbers = append(numbers, arg.Call())
					}
					numbers = append(numbers, arg)
				}
			}
			return expr(numbers...)
		}
		if args[0].TypeNat().Match(d.Numbers) {
			return expr(args[0])
		}
	}

	return expr()
})

func TestFunctor(t *testing.T) {
	fmt.Printf("functor: %s, function type: %s, native type: %s, call without args: %s, call with '1' arg: %s call with multiple args: %s\n",
		numFunc,
		numFunc.TypeFnc().TypeName(),
		numFunc.TypeNat().TypeName(),
		numFunc.Call(),
		numFunc.Call(NewFromData(d.IntVal(1))),
		numFunc.Call(NewFromData(d.IntVal(1)),
			NewFromData(d.IntVal(1)),
			NewFromData(d.IntVal(2)),
			NewFromData(d.IntVal(3)),
			NewFromData(d.IntVal(4)),
			NewFromData(d.IntVal(5))))
}

func TestApplicable(t *testing.T) {
	fmt.Printf("applicable: %s, function type: %s, native type: %s, call without args: %s\n",
		numApp,
		numApp.TypeFnc().TypeName(),
		numApp.TypeNat().TypeName(),
		numApp.Call())

	fmt.Printf("function type: %s, native type: %s, call with flat expression argument: %s\n",
		numApp.Call(numFunc).TypeFnc().TypeName(),
		numApp.Call(numFunc).TypeNat().TypeName(),
		numApp.Call(NewFromData(d.IntVal(23))))

	fmt.Printf("function type: %s, native type: %s, call with numeric functor argument: %s\n",
		numApp.Call(numFunc).TypeFnc().TypeName(),
		numApp.Call(numFunc).TypeNat().TypeName(),
		numApp.Call(numFunc))

	fmt.Printf("function type: %s, native type: %s, call with multiple flat arguments: %s\n",
		numApp.Call(numFunc).TypeFnc().TypeName(),
		numApp.Call(numFunc).TypeNat().TypeName(),
		numApp.Call(NewFromData(d.IntVal(1)),
			NewFromData(d.IntVal(1)),
			NewFromData(d.IntVal(2)),
			NewFromData(d.IntVal(3)),
			NewFromData(d.StrVal("str arg")),
			NewFromData(d.IntVal(4)),
			NewFromData(d.IntVal(5))))

	// apply function should deal with non numeric arguments
	fmt.Printf("function type: %s, native type: %s, call with multiple functor arguments: %s\n",
		numApp.Call(numFunc).TypeFnc().TypeName(),
		numApp.Call(numFunc).TypeNat().TypeName(),
		numApp.Call(NewFunctor(NewFromData(d.IntVal(1))),
			NewFunctor(NewFromData(d.IntVal(1))),
			NewFunctor(NewFromData(d.IntVal(2))),
			NewFunctor(NewFromData(d.IntVal(3))),
			NewFunctor(NewFromData(d.StrVal("str arg"))),
			NewFunctor(NewFromData(d.IntVal(4))),
			NewFunctor(NewFromData(d.IntVal(5)))))

}
