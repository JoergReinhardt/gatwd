package functions

import (
	"fmt"
	"testing"

	d "github.com/JoergReinhardt/godeep/data"
)

func TestFlag(t *testing.T) {
	flag := newFlag(Parameter|Vector, d.Vector.Flag()|d.Function.Flag())
	fmt.Println(flag.String())
}
