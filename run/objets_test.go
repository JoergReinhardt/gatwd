package run

import (
	"fmt"
	"testing"

	d "github.com/JoergReinhardt/godeep/data"
)

func TestAtomicObjectAllocation(t *testing.T) {
	c := allocateAtomicConstant(d.New("testvalue"))
	fmt.Println(c.Closure.Eval())
}
