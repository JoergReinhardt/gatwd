package functions

import (
	"fmt"
	"testing"
)

var listA = NewVector(New(0), New(1), New(2), New(3),
	New(4), New(5), New(6), New(7), New(8), New(9))

var listB = NewVector(New(10), New(11), New(12), New(13),
	New(14), New(15), New(16), New(17), New(18), New(19))

func conList(args ...Callable) Consumeable {
	return NewList(args...)
}
func printCons(cons Consumeable) {
	var head, tail = cons.DeCap()
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
	var head, tail = alist()

	for i := 0; i < 5; i++ {
		head, tail = tail()
	}

	printCons(tail)
	fmt.Println("")

	head, tail = tail(listB()...)

	head, tail = tail()

	fmt.Println(head)
	fmt.Println("")

	printCons(tail)
}
