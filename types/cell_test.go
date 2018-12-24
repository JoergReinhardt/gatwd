package types

import (
	"fmt"
	"testing"
)

func TestCellAllocation(t *testing.T) {
	c1 := constructDataCell(Make("this is string test data"))
	fmt.Printf("data: %s type: %s, string: %s, copy: %s, eval: %s reference: %s\n",
		c1, c1.Type(), c1.String(), c1.Copy(), c1.Eval(), c1.ref())
	c2 := constructDataCell(Make(3, 4, 5, 6, 7, 8, 9, 10, "this", "is", "string", "array"))
	fmt.Printf("data: %s type: %s, string: %s, copy: %s, eval: %s reference: %s\n",
		c2, c2.Type(), c2.String(), c2.Copy(), c2.Eval(), c2.ref())

	eva := stripEvalMethodSet(Make("test string"))
	fmt.Printf("eval method set: %s\n", eva)

	tup := constructTuple(New("one"), New("two"), New("three"), New("four"))
	fmt.Printf("tuple: %s\n", tup)
	h, tu := tup.Decap()
	fmt.Printf("head: %s\ttail: %s\n", h, tu)
}
