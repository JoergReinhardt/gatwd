package functions

import (
	"fmt"
	"testing"
)

var listA = NewVector(New(0), New(1), New(2), New(3),
	New(4), New(5), New(6), New(7), New(8), New(9))

var listB = NewVector(New(10), New(11), New(12), New(13),
	New(14), New(15), New(16), New(17), New(18), New(19))

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
	fmt.Printf("empty list pattern length: %d\n", list.Type().(TyPattern).Len())
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
	pair = NewPair(New(12), New("string"))
	fmt.Printf("name of (int,string) pair: %s\n", pair.Type().TypeName())
}
