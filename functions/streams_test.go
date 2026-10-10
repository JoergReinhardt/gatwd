package functions_test

import (
	"testing"
	"github.com/joergreinhardt/gatwd/functions"
)

func TestZipAndTake(t *testing.T) {
	// Infinite sequence generator
	infiniteNums := func(yield func(int) bool) {
		i := 1
		for {
			if !yield(i) {
				return
			}
			i++
		}
	}

	infiniteStrs := func(yield func(string) bool) {
		for {
			if !yield("A") {
				return
			}
		}
	}

	// Zip them together
	zipped := functions.Zip(infiniteNums, infiniteStrs)

	// We MUST use Take, otherwise the loop will run forever
	firstThree := functions.Take(zipped, 3)

	count := 0
	for pair := range firstThree {
		count++
		if pair.Snd != "A" {
			t.Fatalf("Expected 'A', got %s", pair.Snd)
		}
	}

	if count != 3 {
		t.Fatalf("Expected 3 items due to Take, got %d", count)
	}
}
