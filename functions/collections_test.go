package functions

import (
	"fmt"
	d "github.com/joergreinhardt/gatwd/data"
	"testing"
)

var listA = NewVector(Dat(0), Dat(1), Dat(2), Dat(3),
	Dat(4), Dat(5), Dat(6), Dat(7), Dat(8), Dat(9))

var listB = NewVector(Dat(10), Dat(11), Dat(12), Dat(13),
	Dat(14), Dat(15), Dat(16), Dat(17), Dat(18), Dat(19))

func conList(args ...Expression) Sequential {
	return NewList(args...)
}
func printCons(cons Traversable) {
	var head, tail = cons.Traverse()
	//if !head.Type().Match(None) {
	if !head.Type().Match(None) {
		fmt.Println(head)
		printCons(tail)
	}
}
func TestEmptyList(t *testing.T) {
	var list = NewList()
	fmt.Printf("empty list pattern length: %d\n",
		list.Type().Len())
	fmt.Printf("empty list patterns: %d\n",
		list.Type().Pattern())
	fmt.Printf("empty list arg types: %s\n",
		list.Type().TypeArguments())
	fmt.Printf("empty list ident types: %s\n",
		list.Type().TypeIdent())
	fmt.Printf("empty list return types: %s\n",
		list.Type().TypeReturn())
	fmt.Printf("empty list type name: %s\n",
		list.Type())
}
func TestList(t *testing.T) {
	var list = NewList(listA()...)
	fmt.Printf("list pattern length: %d\n",
		list.Type().Len())
	fmt.Printf("list patterns: %d\n",
		list.Type().Pattern())
	fmt.Printf("list arg types: %s\n",
		list.Type().ArgumentsName())
	fmt.Printf("list ident types: %s\n",
		list.Type().IdentName())
	fmt.Printf("list return types: %s\n",
		list.Type().ReturnName())
	fmt.Printf("list type name: %s\n",
		list.Type())
	printCons(list)
}

func TestConList(t *testing.T) {

	var alist = NewList(listA()...)
	var head Expression

	for i := 0; i < 5; i++ {
		head, alist = alist.ConsumeList()
		fmt.Println("for loop: " + head.String())
	}

	alist = alist.Cons(listB()...).(ListVal)

	printCons(alist)
}

func TestPushList(t *testing.T) {

	var alist = NewList(listA()...)
	var head Expression

	for i := 0; i < 5; i++ {
		head, alist = alist.ConsumeList()
		fmt.Println("for loop: " + head.String())
	}

	printCons(alist)
}

func TestVector(t *testing.T) {

	var vec = NewVector(listA.Slice()...)
	fmt.Printf("vector: %s\n", vec)

	vec = vec.Cons(listB.Slice()...).(VecVal)
	fmt.Printf("vector after cons list-B: %s\n", vec)
	fmt.Printf("vector first: %s last: %s\n", vec.First(), vec.Last())

	var head, tail = vec.Consume()
	for !head.Type().Match(None) {
		fmt.Printf("head: %s\n", head)
		head, tail = tail.Consume()
	}
}

func TestPairVal(t *testing.T) {
	var pair = NewPair(NewNone(), NewNone())
	fmt.Printf("name of empty pair: %s\n", pair.Type())

	pair = NewPair(Dat(12), Dat("string"))
	fmt.Printf("name of (int,string) pair: %s\n",
		pair.Type())
	fmt.Printf("name of (int,string) pair args: %s\n",
		pair.Type().TypeArguments())
	fmt.Printf("name of (int,string) pair return: %s\n",
		pair.Type().TypeReturn())
}

func TestSequence(t *testing.T) {
	var seq = NewSequence(listA)
	var head, tail = seq()
	for !head.Type().Match(None) {
		fmt.Printf("head iteration: %s\n", head)
		head, tail = tail()
	}
	fmt.Printf("sequence: %s\n", seq)
	fmt.Printf("seq head: %s, tail: %s type: %s\n",
		seq.Head(), seq.Tail(), seq.TypeFnc())
}

var (
	mapAddInt = Define(GenericFunc(func(args ...Expression) Expression {
		if args[0].Type().Match(Data) &&
			args[1].Type().Match(Data) {
			if ia, ok := args[0].(NatEval).Eval().(d.Integer); ok {
				if ib, ok := args[1].(NatEval).Eval().(d.Integer); ok {
					return Box(ia.Int() + ib.Int())
				}
			}
		}
		return NewNone()
	}),
		DefSym("+"),
		Def(Def(Data, Constant), d.Int),
		Def(
			Def(Def(Data, Constant), d.Int),
			Def(Def(Data, Constant), d.Int),
		))
)

func TestMapList(t *testing.T) {
	var applyAdd = NewSequence(listA).Apply(mapAddInt)
	fmt.Printf("list-adder head: %s\ntail: %s\ntype: %s\n",
		applyAdd.Head(), applyAdd.Tail(), applyAdd.Type())

	var result = applyAdd.Map(listB)
	fmt.Printf("addition results: %s\n", result)
}

func TestFoldList(t *testing.T) {
}

func TestFilterList(t *testing.T) {
}
