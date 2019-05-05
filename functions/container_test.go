package functions

import (
	"fmt"
	"testing"
)

var intslice = []int{
	1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 134, 8566735,
	4534, 3445, 76575, 2234, 45, 7646, 64, 3, 314,
}

var ints = func() []Callable {
	var res = []Callable{}
	for _, i := range intslice {
		res = append(res, New(i))
	}
	return res
}()

func TestVector(t *testing.T) {
	var tup = NewVector(ints...)
	fmt.Printf("vector: %s\n", tup)
}
