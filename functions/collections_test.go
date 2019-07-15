package functions

import (
	"fmt"
	"testing"
)

var listA = NewVector(NewNative(0), NewNative(1), NewNative(2), NewNative(3),
	NewNative(4), NewNative(5), NewNative(6), NewNative(7), NewNative(8), NewNative(9))

var listB = NewVector(NewNative(10), NewNative(11), NewNative(12), NewNative(13),
	NewNative(14), NewNative(15), NewNative(16), NewNative(17), NewNative(18), NewNative(19))

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
	var pair = NewPair(NewNone(), NewNone())
	fmt.Printf("name of empty pair: %s\n", pair.Type().TypeName())
	pair = NewPair(NewNative(12), NewNative("string"))
	fmt.Printf("name of (int,string) pair: %s\n", pair.Type().TypeName())
}
