package arrows_test

import (
	"iter"
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
