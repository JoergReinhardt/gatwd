package functions

import (
	"fmt"
	"testing"
)

var listA = NewVector(DecNative(0), DecNative(1), DecNative(2), DecNative(3),
	DecNative(4), DecNative(5), DecNative(6), DecNative(7), DecNative(8), DecNative(9))

var listB = NewVector(DecNative(10), DecNative(11), DecNative(12), DecNative(13),
	DecNative(14), DecNative(15), DecNative(16), DecNative(17), DecNative(18), DecNative(19))

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
	fmt.Printf("empty list patterns: %d\n", list.Type().Patterns())
	fmt.Printf("empty list arg types: %s\n", list.Type().TypeArguments())
	fmt.Printf("empty list ident types: %s\n", list.Type().TypeIdent())
	fmt.Printf("empty list return types: %s\n", list.Type().TypeReturn())
	fmt.Printf("empty list type name: %s\n", list.Type())
}
func TestList(t *testing.T) {
	var list = NewList(listA()...)
	fmt.Printf("list pattern length: %d\n", list.Type().Len())
	fmt.Printf("list patterns: %d\n", list.Type().Patterns())
	fmt.Printf("list arg types: %s\n", list.Type().ArgumentsName())
	fmt.Printf("list ident types: %s\n", list.Type().IdentName())
	fmt.Printf("list return types: %s\n", list.Type().ReturnName())
	fmt.Printf("list type name: %s\n", list.Type())
	fmt.Printf("list head type: %s\n", list.Head().Type())
	fmt.Printf("list head type: %s\n", list.TypeElem())
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
	fmt.Printf("name of empty pair: %s\n", pair.Type())

	pair = NewPair(DecNative(12), DecNative("string"))
	fmt.Printf("name of (int,string) pair: %s\n", pair.Type())
	fmt.Printf("name of (int,string) pair args: %s\n", pair.Type().TypeArguments())
	fmt.Printf("name of (int,string) pair return: %s\n", pair.Type().TypeReturn())
}

var list = NewList(DecNative(0), DecNative(1), DecNative(2), DecNative(3))

func TestMapList(t *testing.T) {
}

func TestFoldList(t *testing.T) {
}

func TestFilterList(t *testing.T) {
}
