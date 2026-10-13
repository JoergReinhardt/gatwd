package functions_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

func bigOf(s string) *big.Int { b, _ := new(big.Int).SetString(s, 10); return b }

func TestNegate(t *testing.T) {
	rows := []struct {
		in, want any
	}{
		{int8(5), int(-5)},
		{int(0), int(0)},
		{int64(math.MinInt64), bigOf("9223372036854775808")}, // negating the smallest int64 promotes
		{bigOf("9223372036854775808"), int(math.MinInt64)},   // ... and the way back demotes
		{bigOf("-9223372036854775809"), bigOf("9223372036854775809")},
		{uint64(math.MaxUint64), bigOf("-18446744073709551615")},
		{2.5, -2.5},
		{complex(1, 2), complex(-1, -2)},
	}
	for _, r := range rows {
		got, err := functions.Negate(r.in)
		if err != nil {
			t.Errorf("Negate(%#v): %v", r.in, err)
			continue
		}
		if !functions.StructuralEqual(got, r.want) {
			t.Errorf("Negate(%#v) = %#v, want %#v", r.in, got, r.want)
		}
	}
	// the structural type: an int64-sized result is a Go int, never a *big.Int
	got, _ := functions.Negate(bigOf("9223372036854775808"))
	if _, ok := got.(int); !ok {
		t.Errorf("negated 2^63 is %T, want int", got)
	}
	for _, bad := range []any{"a", true, functions.Char('a'), nil, []any{1}} {
		if v, err := functions.Negate(bad); err == nil {
			t.Errorf("Negate(%#v) = %#v, want an error", bad, v)
		}
	}
}

func TestPow(t *testing.T) {
	rows := []struct {
		base, exp, want any
	}{
		{2, 10, 1024},
		{int8(-2), int8(3), int(-8)},
		{2, 0, 1},
		{0, 0, 1},
		{0, 5, 0},
		{1, 1000000000000, 1},
		{-1, 1000000000001, -1},
		{-1, 1000000000000, 1},
		{2, 64, bigOf("18446744073709551616")},
		{2, -1, 0.5},
		{2.0, 3, 8.0},
		{4, 0.5, 2.0},
		{complex(2, 0), 2, complex(4, 0)},
	}
	for _, r := range rows {
		got, err := functions.Pow(r.base, r.exp)
		if err != nil {
			t.Errorf("Pow(%v, %v): %v", r.base, r.exp, err)
			continue
		}
		if !functions.StructuralEqual(got, r.want) {
			t.Errorf("Pow(%v, %v) = %#v, want %#v", r.base, r.exp, got, r.want)
		}
	}
	// 2 ** 62 fits int64 (a Go int); 2 ** 63 does not (a *big.Int).
	if v, _ := functions.Pow(2, 62); v != int(1)<<62 {
		t.Errorf("2**62 = %#v", v)
	}
	if v, _ := functions.Pow(2, 63); v == nil {
		t.Error("2**63 missing")
	} else if _, ok := v.(*big.Int); !ok {
		t.Errorf("2**63 is %T, want *big.Int", v)
	}
	for _, bad := range [][2]any{{0, -1}, {0.0, -1.0}, {-8.0, 0.5}, {2, 99999999}, {"a", 2}, {2, "a"}, {1e308, 2.0}, {2, bigOf("99999999999999999999")}} {
		if v, err := functions.Pow(bad[0], bad[1]); err == nil {
			t.Errorf("Pow(%v, %v) = %#v, want an error", bad[0], bad[1], v)
		}
	}
}
