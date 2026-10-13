package arrows_test

import (
	"iter"
	"slices"
	"testing"

	"github.com/joergreinhardt/gatwd/arrows"
	"github.com/joergreinhardt/gatwd/functions"
)

func TestKleisli(t *testing.T) {
	f := func(a int) iter.Seq[int] {
		return func(yield func(int) bool) {
			yield(a * 2)
			yield(a * 3)
		}
	}

	g := func(b int) iter.Seq[int] {
		return func(yield func(int) bool) {
			yield(b + 1)
		}
	}

	h := arrows.Compose(f, g)

	var res []int
	for val := range h(10) {
		res = append(res, val)
	}

	if len(res) != 2 || res[0] != 21 || res[1] != 31 {
		t.Fatalf("Unexpected kleisli composition result: %v", res)
	}
}

func TestFanOut(t *testing.T) {
	f := func(a int) iter.Seq[int] {
		return func(yield func(int) bool) {
			yield(a * 2)
		}
	}

	g := func(a int) iter.Seq[int] {
		return func(yield func(int) bool) {
			yield(a + 10)
		}
	}

	h := arrows.FanOut(f, g)

	var res []functions.Pair[int, int]
	for val := range h(10) {
		res = append(res, val)
	}

	if len(res) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(res))
	}

	if res[0].Fst != 20 || res[0].Snd != 20 {
		t.Errorf("Expected {20, 20}, got {%v, %v}", res[0].Fst, res[0].Snd)
	}
}

func TestSplit(t *testing.T) {
	f := func(a int) iter.Seq[int] {
		return func(yield func(int) bool) {
			yield(a * 2)
		}
	}

	g := func(b int) iter.Seq[int] {
		return func(yield func(int) bool) {
			yield(b + 10)
		}
	}

	h := arrows.Split(f, g)

	var res []functions.Pair[int, int]
	pair := functions.Pair[int, int]{Fst: 10, Snd: 20}
	for val := range h(pair) {
		res = append(res, val)
	}

	if len(res) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(res))
	}

	if res[0].Fst != 20 || res[0].Snd != 30 {
		t.Errorf("Expected {20, 30}, got {%v, %v}", res[0].Fst, res[0].Snd)
	}
}

func TestPlusConcatenatesInOrder(t *testing.T) {
	gen := func(vals ...int) arrows.Kleisli[string, int] {
		return func(string) iter.Seq[int] {
			return func(yield func(int) bool) {
				for _, v := range vals {
					if !yield(v) {
						return
					}
				}
			}
		}
	}
	collect := func(k arrows.Kleisli[string, int]) []int {
		var out []int
		for v := range k("x") {
			out = append(out, v)
		}
		return out
	}
	if got := collect(arrows.Plus(gen(1, 2), gen(3), gen(), gen(4, 5))); !slices.Equal(got, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("Plus = %v", got)
	}
	// monoid laws: zero arrow is the identity, Plus is associative
	if got := collect(arrows.Plus[string, int]()); len(got) != 0 {
		t.Fatalf("arrows.Plus() must be the zero arrow, got %v", got)
	}
	if got := collect(arrows.Plus(arrows.Plus[string, int](), gen(7))); !slices.Equal(got, []int{7}) {
		t.Fatalf("zero is not an identity: %v", got)
	}
	l := collect(arrows.Plus(arrows.Plus(gen(1), gen(2)), gen(3)))
	r := collect(arrows.Plus(gen(1), arrows.Plus(gen(2), gen(3))))
	if !slices.Equal(l, r) {
		t.Fatalf("Plus not associative: %v vs %v", l, r)
	}
}

// TestPlusIsLazy: a consumer that stops after the first element never starts the
// later arrows.
func TestPlusIsLazy(t *testing.T) {
	started := false
	first := arrows.Kleisli[string, int](func(string) iter.Seq[int] {
		return func(yield func(int) bool) { yield(1) }
	})
	second := arrows.Kleisli[string, int](func(string) iter.Seq[int] {
		started = true
		return func(yield func(int) bool) { yield(2) }
	})
	for range arrows.Plus(first, second)("x") {
		break
	}
	if started {
		t.Fatal("the second arrow must not run after an early stop")
	}
}
