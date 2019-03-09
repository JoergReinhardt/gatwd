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
