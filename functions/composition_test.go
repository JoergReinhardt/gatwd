package functions

import (
	"fmt"
	"testing"
)

//func TestMapF(t *testing.T) {
//
//	var vector = NewVector(listA()...)
//	var fmap = func(args ...Expression) Expression {
//		return New(args[0].Eval().(d.IntVal).Idx() * 3)
//	}
//
//	var mapped = MapF(vector, fmap)
//
//	printCons(mapped)
//}
//
//func TestMapL(t *testing.T) {
//
//	var list = NewList(listA()...)
//	var fmap = func(args ...Expression) Expression {
//		return New(args[0].Eval().(d.IntVal).Idx() * 3)
//	}
//
//	var mapped = MapL(list, fmap)
//
//	printCons(mapped)
//}
//
//func TestFoldL(t *testing.T) {
//
//	var list = NewList(listA()...)
//	var fold = Fold(func(ilem, head Expression, args ...Expression) Expression {
//		return New(ilem.Eval().(d.IntVal) + head.Eval().(d.IntVal))
//	})
//	var ilem = New(0)
//
//	var folded = FoldL(list, ilem, fold)
//
//	printCons(folded)
//}
//
//func TestFoldF(t *testing.T) {
//
//	var vector = NewVector(listA()...)
//	var fold = Fold(func(ilem, head Expression, args ...Expression) Expression {
//		return New(ilem.Eval().(d.IntVal) + head.Eval().(d.IntVal))
//	})
//	var ilem = New(0)
//
//	var folded = FoldF(vector, ilem, fold)
//
//	printCons(folded)
//}
//
//func TestListFoldAndMap(t *testing.T) {
//
//	var list = NewList(listA()...)
//	var elem = New(0)
//	var fold = func(elem, head Expression, args ...Expression) Expression {
//		return New(elem.Eval().(d.IntVal) + head.Eval().(d.IntVal))
//	}
//	var fmap = func(args ...Expression) Expression {
//		return New(args[0].Eval().(d.IntVal).Idx() * 3)
//	}
//
//	var mapped = MapL(list, fmap)
//	var folded = FoldL(mapped, elem, fold)
//
//	printCons(folded)
//
//	folded = FoldL(list, elem, fold)
//	mapped = MapL(folded, fmap)
//
//	var head, result Expression
//	head, mapped = mapped()
//
//	for {
//		fmt.Println(head)
//		head, mapped = mapped()
//		if head == nil {
//			break
//		}
//		result = head
//	}
//
//	if result.Eval().(d.IntVal) != 135 {
//		t.Fail()
//	}
//}
//
//func TestConsumeableFoldAndMap(t *testing.T) {
//
//	var vec = listA
//	var elem = New(0)
//	var fold = func(elem, head Expression, args ...Expression) Expression {
//		return New(elem.Eval().(d.IntVal) + head.Eval().(d.IntVal))
//	}
//	var fmap = func(args ...Expression) Expression {
//		return New(args[0].Eval().(d.IntVal).Idx() * 3)
//	}
//
//	var mapped = MapF(vec, fmap)
//	var folded = FoldF(mapped, elem, fold)
//
//	folded = FoldF(vec, elem, fold)
//	mapped = MapF(folded, fmap)
//
//	var head, result Expression
//	head, mapped = mapped()
//
//	for {
//		fmt.Println(head)
//		head, mapped = mapped()
//		if head == nil {
//			break
//		}
//		result = head
//	}
//
//	if result.Eval().(d.IntVal) != 135 {
//		t.Fail()
//	}
//}
//
var keys = []Expression{New("zero"), New("one"), New("two"), New("three"),
	New("four"), New("five"), New("six"), New("seven"), New("eight"), New("nine"),
	New("ten")}

var vals = []Expression{New(0), New(1), New(2), New(3), New(4), New(5), New(6),
	New(7), New(8), New(9), New(10)}

func TestZipLists(t *testing.T) {
	var zipped = ZipL(NewList(keys...), NewList(vals...), func(l, r Expression) Paired { return NewPair(l, r) })
	fmt.Printf("zipped list: %s\n", zipped)
}

func TestZipConsumeable(t *testing.T) {
	var zipped = ZipF(NewList(keys...), NewList(vals...), func(l, r Expression) Paired { return NewPair(l, r) })

	var head, tail = zipped.Consume()
	for head != nil {
		fmt.Printf("%s,\n ", head)
		head, tail = tail.Consume()
	}
}

//func TestFilterList(t *testing.T) {
//	var filtered = FilterL(NewList(vals...), Filter(func(head Expression, args ...Expression) bool {
//		if (head.Eval().(d.IntVal) % 2) == 0 {
//			return true
//		}
//		return false
//	}))
//
//	var head, tail = filtered()
//	for head != nil {
//		fmt.Printf("filtered element: %s\n", head)
//		head, tail = tail()
//	}
//}

//func TestFilterConsumeable(t *testing.T) {
//	var filtered = FilterF(NewList(vals...), Filter(func(head Expression, args ...Expression) bool {
//		if (head.Eval().(d.IntVal) % 2) == 0 {
//			return true
//		}
//		return false
//	}))
//
//	var head, tail = filtered.Consume()
//	for head != nil {
//		fmt.Printf("filtered element: %s\n", head)
//		head, tail = tail.Consume()
//	}
//}

//func TestBindF(t *testing.T) {
//	// bind function will multiply numerals
//	var bind = func(f, g Expression) Expression {
//		if nf, ok := f.Eval().(d.Numeral); ok {
//			if ng, ok := g.Eval().(d.Numeral); ok {
//				return NewNative(d.IntVal(nf.Int() * ng.Int()))
//			}
//		}
//		return nil
//	}
//	var bound = BindF(listA, listB, bind)
//	var head Expression
//	head, bound = bound()
//	if head.Eval().(d.IntVal) != 0 {
//		t.Fail()
//	}
//	for head != nil {
//		fmt.Printf("%s\n", head)
//		head, bound = bound()
//	}
//}
//
//var f = VariadLambda(func(args ...Expression) Expression {
//	var str = "f and "
//	str = str + args[0].String()
//	return NewNative(d.StrVal(str))
//})
//var df = Define("f expr", f)
//var g = VariadLambda(func(args ...Expression) Expression {
//	var str = "g and "
//	str = str + args[0].String()
//	return NewNative(d.StrVal(str))
//})
//var dg = Define("g expr", g)
//var h = VariadLambda(func(args ...Expression) Expression {
//	var str = "h and "
//	str = str + args[0].String()
//	return NewNative(d.StrVal(str))
//})
//var dh = Define("h expr", h)
//var i = VariadLambda(func(args ...Expression) Expression {
//	var str = "i and "
//	str = str + args[0].String()
//	return NewNative(d.StrVal(str))
//})
//var di = Define("i expr", i)
//var j = VariadLambda(func(args ...Expression) Expression {
//	var str = "j and "
//	str = str + args[0].String()
//	return NewNative(d.StrVal(str))
//})
//var dj = Define("j expr", j)
//var k = ConstLambda(func() Expression {
//	return NewNative(d.StrVal("k"))
//})
//var dk = Define("k expr", k)
//
//func TestCurry(t *testing.T) {
//	var result = Curry(f, g, h, i, j, k)
//	fmt.Println(result)
//	if result.String() != "f and g and h and i and j and k" {
//		t.Fail()
//	}
//	var defresult = Curry(df, dg, dh, di, dj, dk)
//	fmt.Println(defresult)
//	if defresult.String() != "f and g and h and i and j and k" {
//		t.Fail()
//	}
//}
