package run

import (
	"fmt"
	"testing"

	d "github.com/JoergReinhardt/gatwd/data"
)

func TestAtomicObjectAllocation(t *testing.T) {
	c := allocateAtomicConstant(d.New("testvalue"))
	fmt.Println(c.Expr.Eval())
}
