package functions_test

import (
	"iter"
	"testing"
	"github.com/joergreinhardt/gatwd/functions"
)

func TestMapFilterFold(t *testing.T) {
	// Create a simple sequence 1..5
	seq := func(yield func(int) bool) {
		for i := 1; i <= 5; i++ {
			if !yield(i) {
				return
			}
		}
	}

	// Filter even numbers
	evens := functions.Filter(seq, func(x int) bool { return x%2 == 0 })

	// Map: multiply by 10
	mapped := functions.Map(evens, func(x int) int { return x * 10 })

	// Fold: sum them up
	// 2*10 + 4*10 = 60
	sum := functions.Fold(mapped, 0, func(acc, val int) int {
		return acc + val
	})

	if sum != 60 {
		t.Fatalf("Expected 60, got %d", sum)
	}
}

func TestBind(t *testing.T) {
	seq := func(yield func(int) bool) {
		yield(1)
		yield(2)
	}

	// Bind function returning two strings for each int
	fn := func(x int) iter.Seq[int] {
		return func(yield func(int) bool) {
			yield(x * 10)
			yield(x * 100)
		}
	}

	bound := functions.Bind(seq, fn)
	
	results := []int{}
	for v := range bound {
		results = append(results, v)
	}

	if len(results) != 4 {
		t.Fatalf("Expected 4 results, got %d", len(results))
	}
	if results[0] != 10 || results[1] != 100 || results[2] != 20 || results[3] != 200 {
		t.Fatalf("Unexpected results sequence: %v", results)
	}
}
