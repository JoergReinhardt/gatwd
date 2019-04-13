package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var resource = NewVector(New(0), New(1), New(2), New(3),
	New(4), New(5), New(6), New(7), New(8), New(9))

var suspend = UnaryFnc(func(first Callable) Callable {
	return UnaryFnc(func(second Callable) Callable {
		return New(first.Eval().(d.IntVal) +
			second.Eval().(d.IntVal))
	})
})

func TestFold(t *testing.T) {
	var result = Fold(
		NewFunctor(resource),
		BinaryFnc(func(ilem, arg Callable) Callable {
			return NewNative(
				ilem.Eval().(d.IntVal) + arg.Eval().(d.IntVal),
			)
		}),
		NewNative(d.IntVal(0)),
	)
	fmt.Println(result)
	if result.Eval().(d.IntVal) != d.IntVal(45) {
		t.Fail()
	}
}

func TestFoldF(t *testing.T) {
	var result = FoldF(
		NewFunctor(resource),
		BinaryFnc(func(ilem, arg Callable) Callable {
			return NewNative(
				ilem.Eval().(d.IntVal) + arg.Eval().(d.IntVal),
			)
		}),
		NewNative(d.IntVal(0)),
	)

	var val Callable
	val, result = result()

	if val.Eval().(d.IntVal) != d.IntVal(0) {
		t.Fail()
		if val.Eval().(d.IntVal) != d.IntVal(10) {
			t.Fail()
		}
	}

	for result != nil {
		fmt.Println(val)
		val, result = result()
	}
}

func TestMap(t *testing.T) {
	var result = Map(
		NewFunctor(resource),
		UnaryFnc(func(arg Callable) Callable {
			return New(arg.Eval().(d.IntVal) * 10)
		}))

	var head, tail = result()

	if head.Eval().(d.IntVal) != d.IntVal(0) {
		t.Fail()
		if head.Eval().(d.IntVal) != d.IntVal(10) {
			t.Fail()
		}
	}

	for head != nil {
		fmt.Println(head)
		head, tail = tail()
	}

}

func TestMapF(t *testing.T) {
	var result = MapF(
		NewFunctor(resource),
		UnaryFnc(func(arg Callable) Callable {
			return New(arg.Eval().(d.IntVal) * 10)
		}))

	var head, tail = result()

	if head.Eval().(d.IntVal) != d.IntVal(0) {
		t.Fail()
		if head.Eval().(d.IntVal) != d.IntVal(10) {
			t.Fail()
		}
	}

	for head != nil {
		fmt.Println(head)
		head, tail = tail()
	}
}

func TestFilter(t *testing.T) {
	var result = Filter(
		NewFunctor(resource),
		TruthFnc(func(args ...Callable) d.BoolVal {
			var remain = args[0].Eval().(d.IntVal).Int() % 2
			if remain != 0 {
				return d.BoolVal(true)
			}
			return d.BoolVal(false)
		}),
	)

	var head, tail = result()

	for head != nil {
		fmt.Println(head)
		head, tail = tail()
	}
}

func TestFilterF(t *testing.T) {
	var result = FilterF(
		NewFunctor(resource),
		TruthFnc(func(args ...Callable) d.BoolVal {
			var remain = args[0].Eval().(d.IntVal).Int() % 2
			if remain != 0 {
				return d.BoolVal(true)
			}
			return d.BoolVal(false)
		}),
	)

	var head, tail = result()

	for head != nil {
		fmt.Println(head)
		head, tail = tail()
	}
}
