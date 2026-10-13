package functions_test

import (
	"fmt"
	"math"
	"math/big"
	"strings"
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

// ---- the logical Int -----------------------------------------------------------------

// boundary integers: everything around the int64 edge, in every storage the tower uses.
func boundaryInts() []any {
	var out []any
	for _, s := range []string{
		"0", "1", "-1", "2", "-2", "7", "-7", "127", "-128", "32768", "2147483648",
		"9223372036854775806", "9223372036854775807", "9223372036854775808", "9223372036854775809",
		"-9223372036854775807", "-9223372036854775808", "-9223372036854775809",
		"18446744073709551615", "18446744073709551616", "100000000000000000000", "-100000000000000000000",
	} {
		b := bigOf(s)
		out = append(out, functions.CanonInt(b))
		if b.IsInt64() {
			out = append(out, functions.TryNarrowInteger(b.Int64())) // int8..int64 literal storage
		}
		if b.IsUint64() {
			out = append(out, b.Uint64()) // a Uint
		}
	}
	return out
}

func asBigT(t *testing.T, v any) *big.Int {
	t.Helper()
	b, ok := functions.ToBigInt(v)
	if !ok {
		t.Fatalf("%#v is not an integer", v)
	}
	return b
}

// requireCanonical asserts the logical-Int invariant: an int when it fits int64, a
// *big.Int only when it does not — never a *big.Int that merely holds a small number.
func requireCanonical(t *testing.T, ctx string, v any) {
	t.Helper()
	b, ok := functions.ToBigInt(v)
	if !ok {
		t.Fatalf("%s: %#v is not an Int", ctx, v)
	}
	if b.IsInt64() {
		if _, isInt := v.(int); !isInt {
			t.Errorf("%s: %s fits int64 but is %T, want int", ctx, b, v)
		}
		return
	}
	if _, isBig := v.(*big.Int); !isBig {
		t.Errorf("%s: %s does not fit int64 but is %T, want *big.Int", ctx, b, v)
	}
}

func TestIntArithmeticIsExactAndCanonical(t *testing.T) {
	vals := boundaryInts()
	for _, l := range vals {
		for _, r := range vals {
			lb, rb := asBigT(t, l), asBigT(t, r)
			ctx := func(op string) string { return fmt.Sprintf("%v(%T) %s %v(%T)", l, l, op, r, r) }
			for _, c := range []struct {
				op   functions.OpKind
				want *big.Int
			}{
				{functions.OpAdd, new(big.Int).Add(lb, rb)},
				{functions.OpSub, new(big.Int).Sub(lb, rb)},
				{functions.OpMul, new(big.Int).Mul(lb, rb)},
			} {
				got, err := functions.EvalBinaryOp(c.op, l, r)
				if err != nil {
					t.Fatalf("%s: %v", ctx(string(c.op)), err)
				}
				requireCanonical(t, ctx(string(c.op)), got)
				if asBigT(t, got).Cmp(c.want) != 0 {
					t.Errorf("%s = %v, want %v", ctx(string(c.op)), got, c.want)
				}
			}
			if rb.Sign() == 0 {
				for _, op := range []functions.OpKind{functions.OpDiv, functions.OpIDiv, functions.OpMod} {
					if v, err := functions.EvalBinaryOp(op, l, r); err == nil {
						t.Errorf("%s = %v, want a division-by-zero error", ctx(string(op)), v)
					} else if !strings.Contains(err.Error(), "by zero") {
						t.Errorf("%s: error %q does not say 'by zero'", ctx(string(op)), err)
					}
				}
				continue
			}
			// floor division and the sign-of-divisor modulo: a == b*(a//b) + a%b, 0 <= |a%b| < |b|
			q, err := functions.EvalBinaryOp(functions.OpIDiv, l, r)
			if err != nil {
				t.Fatalf("%s: %v", ctx("//"), err)
			}
			m, err := functions.EvalBinaryOp(functions.OpMod, l, r)
			if err != nil {
				t.Fatalf("%s: %v", ctx("%"), err)
			}
			requireCanonical(t, ctx("//"), q)
			requireCanonical(t, ctx("%"), m)
			qb, mb := asBigT(t, q), asBigT(t, m)
			if recomposed := new(big.Int).Add(new(big.Int).Mul(rb, qb), mb); recomposed.Cmp(lb) != 0 {
				t.Errorf("%s: b*(a//b)+a%%b = %v, want %v", ctx("//,%"), recomposed, lb)
			}
			if mb.Sign() != 0 && mb.Sign() != rb.Sign() {
				t.Errorf("%s: modulo %v must take the sign of the divisor %v", ctx("%"), mb, rb)
			}
			if new(big.Int).Abs(mb).Cmp(new(big.Int).Abs(rb)) >= 0 {
				t.Errorf("%s: |modulo| %v must be below |divisor| %v", ctx("%"), mb, rb)
			}
			// `/` is exact when it divides and a Float otherwise
			d, err := functions.EvalBinaryOp(functions.OpDiv, l, r)
			if err != nil {
				t.Fatalf("%s: %v", ctx("/"), err)
			}
			if mb.Sign() == 0 {
				requireCanonical(t, ctx("/"), d)
				if asBigT(t, d).Cmp(qb) != 0 {
					t.Errorf("%s = %v, want %v", ctx("/"), d, qb)
				}
			} else if _, isF := d.(float64); !isF {
				t.Errorf("%s = %#v, want a Float (it does not divide)", ctx("/"), d)
			}
		}
	}
}

// F11 and the demotion rule, spelled out.
func TestBigIntegersDemoteAndPrintWithoutPlus(t *testing.T) {
	big1 := bigOf("9223372036854775808")
	got, err := functions.EvalBinaryOp(functions.OpSub, big1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got != int(math.MaxInt64) {
		t.Errorf("2^63 - 1 = %#v, want the int 9223372036854775807", got)
	}
	got, _ = functions.EvalBinaryOp(functions.OpSub, functions.CanonInt(bigOf("100000000000000000000")), bigOf("100000000000000000000"))
	if got != 0 {
		t.Errorf("x - x = %#v, want the int 0, not a *big.Int holding 0", got)
	}
	got, _ = functions.EvalBinaryOp(functions.OpDiv, bigOf("18446744073709551616"), bigOf("18446744073709551616"))
	if got != 1 {
		t.Errorf("(2^64)/(2^64) = %#v, want the int 1", got)
	}
}

func TestExactNumericComparison(t *testing.T) {
	two53 := bigOf("9007199254740992")
	one := new(big.Int).Add(two53, big.NewInt(1))
	rows := []struct {
		a, b    any
		cmp     int
		ordered bool
	}{
		{1, 1.0, 0, true},
		{one, 9007199254740992.0, 1, true}, // 2^53+1 is above the double 2^53 (a double compare says equal)
		{9007199254740992.0, one, -1, true},
		{int8(3), uint64(3), 0, true},
		{bigOf("18446744073709551616"), 1.8446744073709552e19, 0, true}, // 2^64 is exactly a double
		{bigOf("18446744073709551617"), 1.8446744073709552e19, 1, true},
		{int(5), math.Inf(1), -1, true},
		{int(5), math.Inf(-1), 1, true},
		{bigOf("100000000000000000000"), math.Inf(1), -1, true},
		{math.NaN(), 1, 0, false},
		{1, math.NaN(), 0, false},
		{math.NaN(), math.NaN(), 0, false},
		{-0.0, 0, 0, true},
	}
	for _, r := range rows {
		cmp, ordered, ok := functions.CompareNumbers(r.a, r.b)
		if !ok || ordered != r.ordered || (ordered && cmp != r.cmp) {
			t.Errorf("CompareNumbers(%#v, %#v) = %d,%v,%v; want %d,%v", r.a, r.b, cmp, ordered, ok, r.cmp, r.ordered)
		}
	}
	// NaN: every ordering operator is false, == is false, != is true.
	for _, op := range []functions.OpKind{functions.OpLt, functions.OpLte, functions.OpGt, functions.OpGte, functions.OpEq} {
		if got, err := functions.EvalBinaryOp(op, math.NaN(), 1); err != nil || got != false {
			t.Errorf("NaN %s 1 = %v, %v; want false", op, got, err)
		}
	}
	if got, _ := functions.EvalBinaryOp(functions.OpNeq, math.NaN(), math.NaN()); got != true {
		t.Errorf("NaN != NaN = %v, want true", got)
	}
	// not numbers: ok is false
	for _, v := range []any{"5", true, functions.Char('a'), complex(1, 0), nil} {
		if _, _, ok := functions.CompareNumbers(v, 1); ok {
			t.Errorf("CompareNumbers(%#v, 1) claims both are real numbers", v)
		}
	}
	if _, err := functions.Compare(math.NaN(), 1); err == nil {
		t.Error("Compare with NaN must be an error (a sort over NaN is not a total order)")
	}
}

func TestFloatAndComplexArithmetic(t *testing.T) {
	rows := []struct {
		op   functions.OpKind
		a, b any
		want any
	}{
		{functions.OpAdd, 1, 2.5, 3.5},
		{functions.OpDiv, 7, 2, 3.5},
		{functions.OpIDiv, 7.5, 2, 3},
		{functions.OpIDiv, -7.5, 2, -4},
		{functions.OpMod, -7.5, 2, 0.5},
		{functions.OpMod, 7.5, -2, -0.5},
		{functions.OpMul, complex(0, 1), complex(0, 1), complex(-1, 0)},
		{functions.OpAdd, 1, complex(0, 1), complex(1, 1)},
		{functions.OpDiv, complex(4, 0), 2, complex(2, 0)},
	}
	for _, r := range rows {
		got, err := functions.EvalBinaryOp(r.op, r.a, r.b)
		if err != nil || !functions.StructuralEqual(got, r.want) {
			t.Errorf("%v %s %v = %#v, %v; want %#v", r.a, r.op, r.b, got, err, r.want)
		}
	}
	// the floor quotient of floats is an Int
	if q, _ := functions.EvalBinaryOp(functions.OpIDiv, 7.5, 2); q != 3 {
		t.Errorf("7.5 // 2 = %#v, want the int 3", q)
	}
	bad := []struct {
		op   functions.OpKind
		a, b any
	}{
		{functions.OpDiv, 1, 0}, {functions.OpDiv, 1.5, 0}, {functions.OpDiv, 1, 0.0}, {functions.OpIDiv, 1.5, 0},
		{functions.OpMod, 1.5, 0}, {functions.OpDiv, complex(1, 1), 0},
		{functions.OpIDiv, complex(1, 1), 2}, {functions.OpMod, complex(1, 1), 2}, // no floor on a plane
		{functions.OpAdd, bigOf("1" + strings.Repeat("0", 400)), 1.5}, // an Int beyond the double range is not silently +Inf
		{functions.OpIDiv, math.Inf(1), 1}, {functions.OpIDiv, math.NaN(), 1},
	}
	for _, r := range bad {
		if v, err := functions.EvalBinaryOp(r.op, r.a, r.b); err == nil {
			t.Errorf("%v %s %v = %#v, want an error", r.a, r.op, r.b, v)
		}
	}
}

func TestNumericFunctions(t *testing.T) {
	type fn1 = func(any) (any, error)
	rounds := func(mode functions.RoundingMode) fn1 {
		return func(v any) (any, error) { return functions.RoundToInt(mode, v) }
	}
	rows := []struct {
		name string
		f    fn1
		in   any
		want any
	}{
		{"floor", rounds(functions.RoundFloor), 2.7, 2},
		{"floor", rounds(functions.RoundFloor), -2.1, -3},
		{"ceil", rounds(functions.RoundCeil), 2.1, 3},
		{"ceil", rounds(functions.RoundCeil), -2.9, -2},
		{"trunc", rounds(functions.RoundTrunc), -2.9, -2},
		{"round", rounds(functions.RoundHalfAway), 2.5, 3},
		{"round", rounds(functions.RoundHalfAway), -2.5, -3},
		{"round_even", rounds(functions.RoundHalfEven), 2.5, 2},
		{"round_even", rounds(functions.RoundHalfEven), 3.5, 4},
		{"floor of an Int", rounds(functions.RoundFloor), 7, 7},
		{"floor of a huge Float", rounds(functions.RoundFloor), 1e30, bigOf("1000000000000000019884624838656")},
		{"sign", functions.Sign, -3, -1},
		{"sign", functions.Sign, 0, 0},
		{"sign", functions.Sign, 2.5, 1},
		{"sign of a big Int", functions.Sign, bigOf("-100000000000000000000"), -1},
		{"abs", functions.Abs, -5, 5},
		{"abs", functions.Abs, -2.5, 2.5},
		{"abs of int64 min", functions.Abs, math.MinInt64, bigOf("9223372036854775808")},
		{"sqrt", functions.Sqrt, 16, 4.0},
		{"sqrt", functions.Sqrt, 2, math.Sqrt2},
		{"sqrt complex", functions.Sqrt, complex(-4, 0), complex(0, 2)},
		{"isqrt", functions.ISqrt, 17, 4},
		{"isqrt big", functions.ISqrt, bigOf("100000000000000000000000000000000000000000000000000"), bigOf("10000000000000000000000000")},
		{"to_int exact", functions.ToInt, 3.0, 3},
		{"to_int of an Int", functions.ToInt, 7, 7},
		{"to_float", functions.ToFloat, 3, 3.0},
		{"to_float correctly rounded", functions.ToFloat, bigOf("9007199254740993"), 9007199254740992.0},
		{"to_float of a Float", functions.ToFloat, 2.5, 2.5},
	}
	for _, r := range rows {
		got, err := r.f(r.in)
		if err != nil {
			t.Errorf("%s(%#v): %v", r.name, r.in, err)
			continue
		}
		if !functions.StructuralEqual(got, r.want) {
			t.Errorf("%s(%#v) = %#v, want %#v", r.name, r.in, got, r.want)
		}
	}
	// results that are Ints are canonical
	if v, _ := functions.RoundToInt(functions.RoundFloor, 1e30); v == nil {
		t.Fatal("nil")
	} else if _, ok := v.(*big.Int); !ok {
		t.Errorf("floor 1e30 is %T, want *big.Int", v)
	}
	if v, _ := functions.RoundToInt(functions.RoundFloor, 2.7); v != 2 {
		t.Errorf("floor 2.7 = %#v, want the int 2", v)
	}

	// negative cases: every partial function has one
	negs := []struct {
		name string
		f    fn1
		in   any
	}{
		{"floor NaN", rounds(functions.RoundFloor), math.NaN()},
		{"round +Inf", rounds(functions.RoundHalfAway), math.Inf(1)},
		{"floor String", rounds(functions.RoundFloor), "2.5"},
		{"floor Complex", rounds(functions.RoundFloor), complex(1, 0)},
		{"sign NaN", functions.Sign, math.NaN()},
		{"sign String", functions.Sign, "1"},
		{"sign Complex", functions.Sign, complex(1, 0)},
		{"abs String", functions.Abs, "-1"},
		{"abs Complex", functions.Abs, complex(3, 4)},
		{"sqrt negative", functions.Sqrt, -4},
		{"sqrt String", functions.Sqrt, "4"},
		{"isqrt negative", functions.ISqrt, -1},
		{"isqrt Float", functions.ISqrt, 4.0},
		{"to_int fraction", functions.ToInt, 3.5},
		{"to_int NaN", functions.ToInt, math.NaN()},
		{"to_int String", functions.ToInt, "3"},
		{"to_int Complex", functions.ToInt, complex(1, 0)},
		{"to_float String", functions.ToFloat, "3"},
		{"to_float huge Int", functions.ToFloat, bigOf("1" + strings.Repeat("0", 400))},
		{"to_float Complex", functions.ToFloat, complex(1, 0)},
	}
	for _, r := range negs {
		if v, err := r.f(r.in); err == nil {
			t.Errorf("%s = %#v, want an error", r.name, v)
		}
	}

	// two-argument functions
	for _, r := range []struct{ a, b, gcd, lcm any }{
		{12, 18, 6, 36}, {-12, 18, 6, 36}, {0, 5, 5, 0}, {0, 0, 0, 0}, {7, 13, 1, 91},
		{bigOf("100000000000000000000"), 40, 40, bigOf("100000000000000000000")},
	} {
		g, err := functions.GCD(r.a, r.b)
		l, err2 := functions.LCM(r.a, r.b)
		if err != nil || err2 != nil || !functions.StructuralEqual(g, r.gcd) || !functions.StructuralEqual(l, r.lcm) {
			t.Errorf("gcd/lcm(%v, %v) = %v, %v (%v, %v); want %v, %v", r.a, r.b, g, l, err, err2, r.gcd, r.lcm)
		}
	}
	for _, p := range [][2]any{{1.5, 2}, {2, "4"}, {"a", "b"}} {
		if _, err := functions.GCD(p[0], p[1]); err == nil {
			t.Errorf("gcd(%v, %v) must be an error", p[0], p[1])
		}
		if _, err := functions.LCM(p[0], p[1]); err == nil {
			t.Errorf("lcm(%v, %v) must be an error", p[0], p[1])
		}
	}
	if q, r, err := functions.QuotRem(-7, 2); err != nil || q != -3 || r != -1 {
		t.Errorf("QuotRem(-7, 2) = %v, %v, %v; want -3, -1", q, r, err)
	}
	if _, _, err := functions.QuotRem(1, 0); err == nil {
		t.Error("QuotRem by zero must be an error")
	}
	if _, _, err := functions.QuotRem(1.5, 2); err == nil {
		t.Error("QuotRem on a Float must be an error")
	}
	for _, v := range []any{math.NaN(), math.Inf(1), 1.5, 3} {
		nan, err := functions.IsNaN(v)
		inf, err2 := functions.IsInf(v)
		if err != nil || err2 != nil {
			t.Fatalf("is_nan/is_inf(%v): %v %v", v, err, err2)
		}
		if f, ok := v.(float64); ok && (nan != math.IsNaN(f) || inf != math.IsInf(f, 0)) {
			t.Errorf("is_nan/is_inf(%v) = %v/%v", v, nan, inf)
		}
	}
	if _, err := functions.IsNaN("x"); err == nil {
		t.Error("is_nan of a String must be an error")
	}
	if _, err := functions.IsInf(true); err == nil {
		t.Error("is_inf of a Bool must be an error")
	}
}

// ---- strings are not numbers (the rest of D2) ------------------------------------------

func TestTextIsNeverANumber(t *testing.T) {
	if _, ok := functions.ToFloat64("1.5"); ok {
		t.Error("ToFloat64 must not parse a String")
	}
	if _, ok := functions.ToBigInt("15"); ok {
		t.Error("ToBigInt must not parse a String")
	}
	if _, ok := functions.ToComplex128("1+2i"); ok {
		t.Error("ToComplex128 must not parse a String")
	}
	eq := []struct {
		a, b any
		want bool
	}{
		{"5", 5, false},
		{5, "5", false},
		{"5", 5.0, false},
		{"5.0", 5, false},
		{"10", "10.0", false}, // two Strings compare as text
		{"10", "10", true},
		{"1+2i", complex(1, 2), false},
		{"true", true, false},
		{"a", functions.Char('a'), true}, // D1: the explicit one-character rule
		{functions.Char('a'), "a", true},
		{"ab", functions.Char('a'), false},
		{"", 0, false},
		{"0", false, false},
	}
	for _, r := range eq {
		if got := functions.StructuralEqual(r.a, r.b); got != r.want {
			t.Errorf("StructuralEqual(%#v, %#v) = %v, want %v", r.a, r.b, got, r.want)
		}
		op := functions.OpEq
		if got, _ := functions.EvalBinaryOp(op, r.a, r.b); got != r.want {
			t.Errorf("%#v == %#v = %v, want %v", r.a, r.b, got, r.want)
		}
		if got, _ := functions.EvalBinaryOp(functions.OpNeq, r.a, r.b); got != !r.want {
			t.Errorf("%#v != %#v = %v, want %v", r.a, r.b, got, !r.want)
		}
	}
	for _, p := range [][2]any{{"5", 3}, {3, "5"}, {"5", 3.5}} {
		if c, err := functions.Compare(p[0], p[1]); err == nil {
			t.Errorf("Compare(%#v, %#v) = %d, want an error (a String is not a number)", p[0], p[1], c)
		} else if !strings.Contains(err.Error(), "str_to_int") {
			t.Errorf("Compare(%#v, %#v) error %q does not point to str_to_int", p[0], p[1], err)
		}
	}
}
