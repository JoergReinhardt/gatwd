package arrows_test

import (
	"iter"
	"testing"
	"github.com/joergreinhardt/gatwd/arrows"
)

func TestKleisli(t *testing.T) {
	// a -> [a*2, a*3]
	f := func(a int) iter.Seq[int] {
		return func(yield func(int) bool) {
			yield(a * 2)
			yield(a * 3)
		}
	}
	
	// b -> [b+1]
	g := func(b int) iter.Seq[int] {
		return func(yield func(int) bool) {
			yield(b + 1)
		}
	}

	h := arrows.Compose(f, g)
	
	// Output of h(10) should be f(10) -> [20, 30] -> g(20), g(30) -> [21, 31]
	var res []int
	for val := range h(10) {
		res = append(res, val)
	}

	if len(res) != 2 || res[0] != 21 || res[1] != 31 {
		t.Fatalf("Unexpected kleisli composition result: %v", res)
	}
}
