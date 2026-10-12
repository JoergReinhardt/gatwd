package functions_test

import (
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

// TestEitherBind covers the Monad bind added to Either: it chains on Right and
// short-circuits (propagates) on Left — railway-oriented error handling.
func TestEitherBind(t *testing.T) {
	double := func(n int) functions.Either[string, int] {
		return functions.Right[string, int](n * 2)
	}

	out := functions.Right[string, int](5).Bind(double)
	if v, ok := out.UnwrapRight(); !ok || v != 10 {
		t.Errorf("Right(5).Bind(double) = %v (ok=%v), want Right(10)", v, ok)
	}

	left := functions.Left[string, int]("boom").Bind(double)
	if _, ok := left.UnwrapRight(); ok {
		t.Errorf("Left.Bind should short-circuit, not produce a Right")
	}
	if lv, ok := left.UnwrapLeft(); !ok || lv != "boom" {
		t.Errorf("Left.Bind should propagate the original Left, got %v (ok=%v)", lv, ok)
	}
}
