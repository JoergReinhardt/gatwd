package evaluator

import (
	"math/big"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

// DefaultLess orders numbers of any kind and the time types by VALUE; the printed text
// ("100" < "1r2" < "2r3", "1m" < "5s") is only the last resort for unrelated types.
func TestDefaultLessOrdersByValue(t *testing.T) {
	half, _ := functions.NewRatio(big.NewInt(1), big.NewInt(2))
	for _, tc := range []struct {
		a, b any
		want bool
	}{
		{functions.Duration(5e9), functions.Duration(60e9), true}, // "5s" > "1m" as text
		{functions.Duration(60e9), functions.Duration(5e9), false},
		{functions.Duration(-5e9), functions.Duration(1e9), true},
		{half, 5, true},    // "1r2" < "5" as text too, but 100 below is the telling case
		{100, half, false}, // "100" < "1r2" as text
		{int8(9), int64(10), true},
		{big.NewInt(9), 10, true},
		{9, 10.5, true},
	} {
		if got := DefaultLess(tc.a, tc.b); got != tc.want {
			t.Errorf("DefaultLess(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
	// unrelated types still fall back to text, and never panic
	if !DefaultLess("a", "b") || DefaultLess("b", "a") {
		t.Error("strings must still order as text")
	}
	_ = DefaultLess(functions.Duration(1), "x")
}
