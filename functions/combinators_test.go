package functions_test

import (
	"strconv"
	"testing"
	"github.com/joergreinhardt/gatwd/functions"
)

func TestCompose(t *testing.T) {
	double := func(i int) int { return i * 2 }
	toString := func(i int) string { return strconv.Itoa(i) }

	doubleThenString := functions.Compose(toString, double)
	
	res := doubleThenString(5)
	if res != "10" {
		t.Fatalf("Expected '10', got '%s'", res)
	}
}

func TestCurry(t *testing.T) {
	add := func(a, b int) int { return a + b }
	
	curriedAdd := functions.Curry2(add)
	addFive := curriedAdd(5)
	
	res := addFive(10)
	if res != 15 {
		t.Fatalf("Expected 15, got %d", res)
	}
}
