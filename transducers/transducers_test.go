package transducers_test

import (
	"testing"
	"github.com/joergreinhardt/gatwd/transducers"
)

func TestTransducers(t *testing.T) {
	// 1. Double the number (Map)
	mapper := transducers.Map[int, int, []int](func(x int) int { return x * 2 })
	
	// 2. Keep only numbers > 5 (Filter)
	filter := transducers.Filter[int, []int](func(x int) bool { return x > 5 })

	// Note: We'd normally compose these with a `Compose` function, but in Go 
	// we just nest them. The mapper runs first, then passes to filter.
	pipeline := func(next transducers.Reducer[int, []int]) transducers.Reducer[int, []int] {
		return mapper(filter(next))
	}

	input := []int{1, 2, 3, 4, 5}
	
	// 1*2=2, 2*2=4, 3*2=6, 4*2=8, 5*2=10
	// >5 -> 6, 8, 10
	result := transducers.Sequence(input, pipeline)

	if len(result) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(result))
	}
	if result[0] != 6 || result[1] != 8 || result[2] != 10 {
		t.Fatalf("Unexpected results: %v", result)
	}
}
