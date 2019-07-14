package functions

import (
	"fmt"
	d "github.com/joergreinhardt/gatwd/data"
	"testing"
)

func TestPattern(t *testing.T) {
	var pat = ConPattern(d.Int, d.Float, Vector, Consumeables)
	fmt.Printf("pat: %s\n", pat)
	fmt.Printf("pat matches Int, Float, Vector, Consumeables: %t\n",
		pat.MatchAll(d.Int, d.Float, Vector, Consumeables))
	if !pat.MatchAll(d.Int, d.Float, Vector, Consumeables) {
		t.Fail()
	}
	fmt.Printf("pat matches Numbers, Float, Vector, Consumeables: %t\n",
		pat.MatchAll(d.Numbers, d.Float, Vector, Consumeables))
	if !pat.MatchAll(d.Numbers, d.Float, Vector, Consumeables) {
		t.Fail()
	}
	fmt.Printf("pat matches Numbers, Numbers, Vector, List: %t\n",
		pat.MatchAll(d.Numbers, d.Numbers, Vector, List))
	if !pat.MatchAll(d.Numbers, d.Numbers, Vector, List) {
		t.Fail()
	}
}

func TestNestedPattern(t *testing.T) {
	var nest = ConPattern(ConPattern(d.Int, d.Float), ConPattern(Vector, List), Consumeables)
	fmt.Printf("nest: %s\n", nest)
	fmt.Printf("nest print '(',' ',')': %s\n", nest.Print("(", " ", ")"))
	fmt.Printf("nest match pattern(d.Int), %t\n", nest.MatchAll(ConPattern(d.Int)))
	if !nest.MatchAll(ConPattern(d.Int)) {
		t.Fail()
	}
	fmt.Printf("nest match d.Int, %t\n", nest.MatchAll(d.Int))
	if nest.MatchAll(d.Int) {
		t.Fail()
	}

	var head, tail = nest.Consume()
	fmt.Printf("consumed nest head: %s, tail: %s\n",
		head.(TyPattern).Print("(", " ", ")"), tail)
	for head != nil {
		fmt.Printf("consumed nest head: %s, tail: %s\n",
			head.Type().TypeName(), tail)
		head, tail = tail.Consume()
	}
}
