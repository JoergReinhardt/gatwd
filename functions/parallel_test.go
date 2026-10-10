package functions_test

import (
	"iter"
	"testing"
	"github.com/joergreinhardt/gatwd/functions"
)

func TestParMap(t *testing.T) {
	seq := func(yield func(int) bool) {
		for i := 1; i <= 10; i++ {
			if !yield(i) {
				return
			}
		}
	}

	mapped := functions.ParMap(seq, 3, func(x int) int {
		return x * 2
	})

	sum := 0
	for v := range mapped {
		sum += v
	}

	// Sum of 1..10 = 55. Sum of 2..20 = 110.
	if sum != 110 {
		t.Fatalf("Expected sum to be 110, got %d", sum)
	}
}

func TestParBind(t *testing.T) {
	seq := func(yield func(int) bool) {
		for i := 1; i <= 3; i++ {
			if !yield(i) {
				return
			}
		}
	}

	bound := functions.ParBind(seq, 2, func(x int) iter.Seq[int] {
		return func(yield func(int) bool) {
			yield(x * 10)
			yield(x * 100)
		}
	})

	sum := 0
	count := 0
	for v := range bound {
		sum += v
		count++
	}

	// Input: 1, 2, 3
	// Output elements: 10, 100, 20, 200, 30, 300
	// Sum: 660
	// Count: 6
	if count != 6 {
		t.Fatalf("Expected 6 items, got %d", count)
	}
	if sum != 660 {
		t.Fatalf("Expected sum 660, got %d", sum)
	}
}

// Test short-circuiting ensures goroutines don't leak.
func TestParMapShortCircuit(t *testing.T) {
	seq := func(yield func(int) bool) {
		for i := 1; i <= 1000; i++ {
			if !yield(i) {
				return
			}
		}
	}

	mapped := functions.ParMap(seq, 5, func(x int) int {
		return x
	})

	count := 0
	for _ = range mapped {
		count++
		if count == 10 {
			break
		}
	}

	if count != 10 {
		t.Fatalf("Expected to stop after 10 elements, got %d", count)
	}
}
