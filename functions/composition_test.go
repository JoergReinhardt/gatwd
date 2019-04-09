package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var resource = NewVector(New(0), New(1), New(2), New(3),
	New(4), New(5), New(6), New(7), New(8), New(9))

func TestFold(t *testing.T) {
	var result = Fold(
		NewFunctor(resource),
		BinaryFnc(func(ilem, arg Callable) Callable {
			return Native(func() d.Native {
				return ilem.Eval().(d.IntVal) +
					arg.Eval().(d.IntVal)
			})
		}),
		Native(func() d.Native { return d.IntVal(0) }),
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
			return Native(func() d.Native {
				return ilem.Eval().(d.IntVal) +
					arg.Eval().(d.IntVal)
			})
		}),
		Native(func() d.Native { return d.IntVal(0) }),
	)

	var val Callable
	val, result = result()

	for val != nil {
		fmt.Println(val)
		val, result = result()
	}

	fmt.Println(val)
}
