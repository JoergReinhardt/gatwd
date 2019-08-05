package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var listA = NewVector(DeclareNative(0), DeclareNative(1), DeclareNative(2), DeclareNative(3),
	DeclareNative(4), DeclareNative(5), DeclareNative(6), DeclareNative(7), DeclareNative(8), DeclareNative(9))

var listB = NewVector(DeclareNative(10), DeclareNative(11), DeclareNative(12), DeclareNative(13),
	DeclareNative(14), DeclareNative(15), DeclareNative(16), DeclareNative(17), DeclareNative(18), DeclareNative(19))

func conList(args ...Expression) Consumeable {
	return NewList(args...)
}
func printCons(cons Consumeable) {
	var head, tail = cons.Consume()
	if head != nil {
		fmt.Println(head)
		printCons(tail)
	}
}
func TestEmptyList(t *testing.T) {
	var list = NewList()
	fmt.Printf("empty list pattern length: %d\n", list.Type().Len())
	fmt.Printf("empty list type name: %s\n", list.Type().TypeName())
}
func TestList(t *testing.T) {
	var list = NewList(listA()...)
	fmt.Printf("list type name: %s\n", list.Type().TypeName())
	printCons(list)
}

func TestConList(t *testing.T) {

	var alist = NewList(listA()...)
	var head Expression

	for i := 0; i < 5; i++ {
		head, alist = alist()
		fmt.Println("for loop: " + head.String())
	}

	alist = alist.Con(listB()...)

	printCons(alist)
}

func TestPushList(t *testing.T) {

	var alist = NewList(listA()...)
	var head Expression

	for i := 0; i < 5; i++ {
		head, alist = alist()
		fmt.Println("for loop: " + head.String())
	}

	alist = alist.Push(listB()...)

	printCons(alist)
}

func TestPairVal(t *testing.T) {
	var pair = NewPair(DeclareNone(), DeclareNone())
	fmt.Printf("name of empty pair: %s\n", pair.Type().TypeName())
	pair = NewPair(DeclareNative(12), DeclareNative("string"))
	fmt.Printf("name of (int,string) pair: %s\n", pair.Type().TypeName())
}

var list = NewList(DeclareNative(0), DeclareNative(1), DeclareNative(2), DeclareNative(3))

func TestMapList(t *testing.T) {
	var add = DeclareFunction(func(args ...Expression) Expression {
		if len(args) > 0 {
			return DeclareData(args[0].(DataConst).Eval().(d.IntVal) + d.IntVal(10))
		}
		return DeclareNative(1)
	}, Def(Data, d.Int))
	fmt.Printf("add %s\n", add)
	fmt.Printf("add two %s\n", add.Call(DeclareNative(d.IntVal(2))))
	var list = list.Map(add)
	var head, tail = list.Consume()
	fmt.Printf("head: %s, tail: %s\n", head, tail)
	if head.(DataConst).Eval() != d.IntVal(10) {
		t.Fail()
	}
}

var add = DeclareExpression(DeclareFunction(func(args ...Expression) Expression {
	if len(args) > 0 {
		return DeclareExpression(DeclareFunction(func(args ...Expression) Expression {
			if len(args) > 1 {
				var a, b = args[0].(DataConst).Eval().(d.IntVal),
					args[1].(DataConst).Eval().(d.IntVal)
				return DeclareNative(a + b)
			}
			return DeclareNone()
		}, Def(Data, d.Int)), Def(Data, d.Int), Def(Data, d.Int)).Call(args[0])
	}
	return DeclareNone()
}, Def(Data, d.Int)), Def(Data, d.Int))

func TestApplyList(t *testing.T) {
	var list = NewList(DeclareNative(0), DeclareNative(1), DeclareNative(2), DeclareNative(3))
	var first = add(DeclareNative(10))
	fmt.Printf("first: %s\n", first)
	var second = first.Call(DeclareNative(10))
	fmt.Printf("second: %s\n", second)
	var fns = list.Map(add)
	var applyed = NewList(DeclareNative(10), DeclareNative(20), DeclareNative(30), DeclareNative(40), DeclareNative(50), DeclareNative(60), DeclareNative(70)).Apply(fns)
	fmt.Printf("applyed: %s\n", applyed)
	var head, tail = applyed.Head(), applyed.Tail()
	fmt.Printf("head: %s, tail: %s\n", head, tail)
}

func TestFoldList(t *testing.T) {
	var sum = DeclareExpression(DeclareFunction(func(args ...Expression) Expression {
		var sum = DataExpr(func(args ...d.Native) d.Native {
			if len(args) > 1 {
				var a, b = args[0].(d.IntVal),
					args[1].(d.IntVal)
				return a + b
			}
			return d.IntVal(0)
		}).Call(args...)
		return sum
	}, Def(Data, d.Int)), Def(Data, d.Int), Def(Data, d.Int))
	fmt.Printf("sum: %s\n", sum)
	var result = sum.Call(DeclareNative(1))
	fmt.Printf("result: %s\n", result)
	result = result.Call(DeclareNative(3))
	fmt.Printf("result: %s\n", result)

	var fold = NewList(DeclareNative(0), DeclareNative(1), DeclareNative(2), DeclareNative(4), DeclareNative(5), DeclareNative(6), DeclareNative(7)).FoldL(sum, DeclareNative(8))
	fmt.Printf("fold: %s\n", fold)
}

func TestFilterList(t *testing.T) {
	var filter = DeclareTest(func(args ...Expression) bool {
		if nat, ok := args[0].(Native); ok {
			if i, ok := nat.Eval().(d.IntVal); ok {
				return i%2 == 0
			}
		}
		return false
	})
	var list = NewList(DeclareNative(0), DeclareNative(1), DeclareNative(2), DeclareNative(4), DeclareNative(5), DeclareNative(6), DeclareNative(7)).Filter(filter)
	fmt.Printf("list: %s\n", list)

}

func TestTakeN(t *testing.T) {
	var list = NewList(DeclareNative(0), DeclareNative(1), DeclareNative(2), DeclareNative(3), DeclareNative(4),
		DeclareNative(5), DeclareNative(6), DeclareNative(7), DeclareNative(8), DeclareNative(9))

	fmt.Printf("list: %s\n", list)
	var head, tail = list.TakeN(4)
	fmt.Printf("head: %s, list: %s\n", head, tail)
	var vec = NewVector(head)
	head, tail = tail.TakeN(4)
	vec = vec.Con(head)
	fmt.Printf("head: %s, list: %s vec: %s\n", head, tail, vec)
	head, tail = tail.TakeN(4)
	vec = vec.Con(head)
	fmt.Printf("head: %s, list: %s vec: %s\n", head, tail, vec)

	list = NewList(DeclareNative(0), DeclareNative(1), DeclareNative(2), DeclareNative(3), DeclareNative(4),
		DeclareNative(5), DeclareNative(6), DeclareNative(7), DeclareNative(8), DeclareNative(9))

	var take4 = DeclareFunction(func(args ...Expression) Expression {
		var ok bool
		var init ColVec
		if len(args) > 0 {
			if init, ok = args[0].(ColVec); ok {
				var vec ColVec
				if init.Len() == 0 {
					vec = NewVector()
				}
				if init.Len() > 0 {
					vec, init = init()[init.Len()-1].(ColVec), NewVector(init()[:init.Len()-1]...)
				}
				if len(args) > 1 {
					if vec.Len() < 4 {
						vec = vec.Con(args[1])
						return init.Con(vec)
					}
					return init.Con(vec, NewVector(args[1]))
				}
			}
		}
		return nil
	}, Def(Type))

	list = list.FoldL(take4, NewVector())
	fmt.Printf("list: %s\n", list)

	var expr Expression
	expr, tail = list()
	for expr != nil {
		fmt.Printf("head: %s, tail: %s\n", expr, tail)
		expr, tail = tail()
	}
}
