package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var intslice = []int{
	1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 134, 8566735,
	4534, 3445, 76575, 2234, 45, 7646, 64, 3, 314,
}

var parms = func() []Parametric {
	var res = []Parametric{}
	for _, dat := range intslice {
		res = append(res, NewFromData(d.IntVal(dat)))
	}
	return res
}()

func TestRecursiveList(t *testing.T) {
	fmt.Printf("parms: %s\n", parms)
	var list = NewList()
	var head Parametric
	head, list = list(parms...)
	for head != nil {
		fmt.Printf("head: %s\n", head)
		head, list = list()
	}
}
func TestListCon(t *testing.T) {
	var list = NewList(parms...)
	fmt.Printf("list len: %d\n", list.Len())

	if list.Len() != 21 {
		t.Fail()
	}

	var slice = []Parametric{}
	var head Parametric
	head, list = list()
	slice = append(slice, head)

	for i := 0; i < 5; i++ {
		fmt.Printf("head: %s\n", head)
		head, list = list()
		slice = append(slice, head)
	}

	head, list = list(slice...)

	if list.Len() != 20 {
		t.Fail()
	}

	for i := 0; i < 5; i++ {
		head, list = list()
		slice = append(slice, head)
	}
	fmt.Printf("slice: %s\n", slice)
	if slice[5].Eval().(d.IntVal).Int() != 6 {
		t.Fail()
	}

}

func TestListMapF(t *testing.T) {
	var list = NewList(parms...)
	var mapped = LMapE(list, UnaryFnc(func(arg Parametric) Parametric {
		return New(arg.Eval().(d.IntVal).Int() + 42)
	}))
	var head Parametric
	head, mapped = mapped()
	for head != nil {
		fmt.Printf("mapped head: %s\n", head.Call())
		head, mapped = mapped()
	}
}

func TestListReverse(t *testing.T) {
	var list = NewList(parms...)
	var reverse = ReverseList(list)
	var head Parametric
	head, reverse = reverse()
	for head != nil {
		fmt.Printf("reversed head: %s\n", head.Call())
		head, reverse = reverse()
	}
}

func TestListFold(t *testing.T) {
	var list = NewList(parms...)
	var folded = LFoldL(list,
		BinaryFnc(func(accum, arg Parametric) Parametric {
			return New(
				accum.Eval().(d.IntVal).Int() +
					arg.Eval().(d.IntVal).Int())
		}),
		New(0),
	)
	fmt.Printf("folded element: %s\n", folded)
}

func TestCurry(t *testing.T) {
}
