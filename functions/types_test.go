package functions

import (
	"fmt"
	d "github.com/joergreinhardt/gatwd/data"
	"testing"
)

func TestPattern(t *testing.T) {
	var pat = Def(d.Int, d.Float, Vector, Collections)
	fmt.Printf("pat: %s\n", pat)
	fmt.Printf("pat matches Int, Float, Vector, Consumeables: %t\n",
		pat.MatchTypes(d.Int, d.Float, Vector, Collections))
	if !pat.MatchTypes(d.Int, d.Float, Vector, Collections) {
		t.Fail()
	}
	fmt.Printf("pat matches Numbers, Float, Vector, Consumeables: %t\n",
		pat.MatchTypes(d.Numbers, d.Float, Vector, Collections))
	if !pat.MatchTypes(d.Numbers, d.Float, Vector, Collections) {
		t.Fail()
	}
	fmt.Printf("pat matches Numbers, Numbers, Vector, List: %t\n",
		pat.MatchTypes(d.Numbers, d.Numbers, Vector, List))
	if !pat.MatchTypes(d.Numbers, d.Numbers, Vector, List) {
		t.Fail()
	}
	fmt.Printf("pat matches Boolean, Numbers, Vector, List: %t\n",
		pat.MatchTypes(d.Bool, d.Numbers, Vector, List))
	if pat.MatchTypes(d.Bool, d.Numbers, Vector, List) {
		t.Fail()
	}
}

func TestNestedPattern(t *testing.T) {
}
