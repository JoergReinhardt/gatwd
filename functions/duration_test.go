package functions_test

import (
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

func dur(t *testing.T, text string) functions.Duration {
	t.Helper()
	d, err := functions.ParseDuration(text)
	if err != nil {
		t.Fatalf("ParseDuration(%q): %v", text, err)
	}
	return d
}

const (
	sec  = functions.Duration(1e9)
	min_ = 60 * sec
	hour = 60 * min_
	day  = 24 * hour
)

func TestParseDuration(t *testing.T) {
	for _, tc := range []struct {
		text string
		want functions.Duration
	}{
		{"5s", 5 * sec},
		{"250ms", 250e6},
		{"1h30m", hour + 30*min_},
		{"1h30m15s", hour + 30*min_ + 15*sec},
		{"2d", 2 * day},
		{"1w", 7 * day},
		{"1w2d3h", 9*day + 3*hour},
		{"1.5h", hour + 30*min_},
		{"0.5s", 500e6},
		{"1.5us", 1500},
		{"3µs", 3000}, // micro sign
		{"3μs", 3000}, // Greek mu
		{"7ns", 7},
		{"0s", 0},
		{"-5s", -5 * sec},
		{"+5s", 5 * sec},
		{"30m1h", hour + 30*min_}, // any order, a unit once
		{"9223372036854775807ns", functions.Duration(math.MaxInt64)},
		{"-9223372036854775808ns", functions.Duration(math.MinInt64)},
	} {
		if got := dur(t, tc.text); got != tc.want {
			t.Errorf("ParseDuration(%q) = %d, want %d", tc.text, got, tc.want)
		}
	}
}

// Every malformed literal is an error that names the problem, never a silent zero.
func TestParseDurationErrors(t *testing.T) {
	for _, tc := range []struct{ text, want string }{
		{"", "empty"},
		{"5", "no unit"},
		{"5q", "unknown duration unit"},
		{"5min", "unknown duration unit"},
		{"1h1h", "used twice"},
		{"1h30", "no unit"},
		{"h", "expected a number"},
		{"-", "sign without a number"},
		{"1h 30m", "unknown duration unit"},
		{" 5s", "expected a number"},
		{"1h0.5m", "only the first segment"},
		{"1.s", "nothing after the decimal point"},
		{"0.1ns", "not a whole number of nanoseconds"},
		{"9223372036854775808ns", "out of range"},
		{"106752d", "out of range"}, // 106752 days > the int64 nanoseconds
		{"99999999999999999999h", "out of range"},
		{strings.Repeat("9", 200) + "s", "longer than"},
	} {
		_, err := functions.ParseDuration(tc.text)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("ParseDuration(%q) error = %v, want one containing %q", tc.text, err, tc.want)
		}
	}
}

func TestDurationString(t *testing.T) {
	for _, tc := range []struct {
		d    functions.Duration
		want string
	}{
		{0, "0s"},
		{5 * sec, "5s"},
		{-5 * sec, "-5s"},
		{hour + 30*min_, "1h30m"},
		{250e6, "250ms"},
		{sec + 500e6, "1s500ms"},
		{1500, "1us500ns"},
		{7, "7ns"},
		{2 * day, "48h"},
		{30 * day, "720h"},
		{functions.Duration(math.MaxInt64), "2562047h47m16s854ms775us807ns"},
		{functions.Duration(math.MinInt64), "-2562047h47m16s854ms775us808ns"},
	} {
		if got := tc.d.String(); got != tc.want {
			t.Errorf("Duration(%d).String() = %q, want %q", int64(tc.d), got, tc.want)
		}
	}
}

// Property: what a Duration prints re-parses to the same value.
func TestDurationPrintParseRoundTrip(t *testing.T) {
	vals := []int64{0, 1, -1, 999, 1000, 1e6, 1e9, 61e9, 3600e9, 3601e9, 86400e9, math.MaxInt64, math.MinInt64, 123456789012345, -987654321098765}
	for _, v := range vals {
		d := functions.Duration(v)
		back, err := functions.ParseDuration(d.String())
		if err != nil || back != d {
			t.Errorf("round trip of %d via %q = %d, %v", v, d.String(), int64(back), err)
		}
	}
}

func eval(t *testing.T, op functions.OpKind, l, r any) (any, error) {
	t.Helper()
	return functions.EvalBinaryOp(op, l, r)
}

func mustEval(t *testing.T, op functions.OpKind, l, r any) any {
	t.Helper()
	v, err := eval(t, op, l, r)
	if err != nil {
		t.Fatalf("%v %s %v: %v", l, op, r, err)
	}
	return v
}

func TestDurationArithmetic(t *testing.T) {
	// Duration ± Duration is a Duration (structural: the raw Go type).
	if got := mustEval(t, functions.OpAdd, hour, 30*min_); got != functions.Duration(hour+30*min_) {
		t.Errorf("1h + 30m = %#v", got)
	}
	if got := mustEval(t, functions.OpSub, 30*min_, hour); got != functions.Duration(-30*min_) {
		t.Errorf("30m - 1h = %#v", got)
	}
	// Duration * number, number * Duration, Duration / number: Int, Float and Ratio.
	half, _ := functions.NewRatio(big.NewInt(1), big.NewInt(2))
	for _, tc := range []struct {
		op   functions.OpKind
		l, r any
		want functions.Duration
	}{
		{functions.OpMul, 5 * sec, 3, 15 * sec},
		{functions.OpMul, 3, 5 * sec, 15 * sec},
		{functions.OpMul, 5 * sec, 1.5, 7*sec + 500e6},
		{functions.OpMul, 1.5, 5 * sec, 7*sec + 500e6},
		{functions.OpMul, 5 * sec, half, 2*sec + 500e6},
		{functions.OpMul, 5 * sec, big.NewInt(2), 10 * sec},
		{functions.OpMul, 5 * sec, -2, -10 * sec},
		{functions.OpDiv, 10 * sec, 4, 2*sec + 500e6},
		{functions.OpDiv, 10 * sec, 4.0, 2*sec + 500e6},
		{functions.OpDiv, 10 * sec, half, 20 * sec},
	} {
		if got := mustEval(t, tc.op, tc.l, tc.r); got != tc.want {
			t.Errorf("%v %s %v = %#v, want %v", tc.l, tc.op, tc.r, got, tc.want)
		}
	}
	// truncation toward zero for both signs
	if got := mustEval(t, functions.OpDiv, functions.Duration(7), 2); got != functions.Duration(3) {
		t.Errorf("7ns / 2 = %v, want 3ns", got)
	}
	if got := mustEval(t, functions.OpDiv, functions.Duration(-7), 2); got != functions.Duration(-3) {
		t.Errorf("-7ns / 2 = %v, want -3ns (toward zero)", got)
	}
	if got := mustEval(t, functions.OpMul, functions.Duration(3), 0.5); got != functions.Duration(1) {
		t.Errorf("3ns * 0.5 = %v, want 1ns", got)
	}
}

// Duration / Duration is the exact ratio: a Ratio, demoted to an Int when whole.
func TestDurationRatio(t *testing.T) {
	got := mustEval(t, functions.OpDiv, 90*min_, hour)
	r, ok := got.(*functions.Ratio)
	if !ok || r.String() != "3r2" {
		t.Errorf("90m / 1h = %#v (%T), want the Ratio 3r2", got, got)
	}
	whole := mustEval(t, functions.OpDiv, 2*hour, hour)
	if !isInt(whole) || !functions.StructuralEqual(whole, 2) {
		t.Errorf("2h / 1h = %#v (%T), want the Int 2", whole, whole)
	}
	if neg := mustEval(t, functions.OpDiv, -90*min_, hour); neg.(*functions.Ratio).String() != "-3r2" {
		t.Errorf("-90m / 1h = %v", neg)
	}
	// floor quotient and remainder, the sign of the divisor
	if q := mustEval(t, functions.OpIDiv, 100*min_, hour); !isInt(q) || !functions.StructuralEqual(q, 1) {
		t.Errorf("100m // 1h = %#v, want 1", q)
	}
	if q := mustEval(t, functions.OpIDiv, -100*min_, hour); !functions.StructuralEqual(q, -2) {
		t.Errorf("-100m // 1h = %#v, want -2 (floor)", q)
	}
	if m := mustEval(t, functions.OpMod, 100*min_, hour); m != functions.Duration(40*min_) {
		t.Errorf("100m %% 1h = %#v, want 40m", m)
	}
	if m := mustEval(t, functions.OpMod, -100*min_, hour); m != functions.Duration(20*min_) {
		t.Errorf("-100m %% 1h = %#v, want 20m (the divisor's sign)", m)
	}
	if m := mustEval(t, functions.OpMod, 100*min_, -hour); m != functions.Duration(-20*min_) {
		t.Errorf("100m %% -1h = %#v, want -20m", m)
	}
}

func isInt(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

// A Duration is never a number: every mixed operation is an error naming both types.
func TestDurationNegatives(t *testing.T) {
	bigDur := functions.Duration(math.MaxInt64)
	for _, tc := range []struct {
		name string
		op   functions.OpKind
		l, r any
		want string
	}{
		{"plus an Int", functions.OpAdd, 5 * sec, 1, "not defined for Duration and Int"},
		{"Int plus", functions.OpAdd, 1, 5 * sec, "not defined for Int and Duration"},
		{"minus an Int", functions.OpSub, 5 * sec, 1, "not defined for Duration and Int"},
		{"plus a Float", functions.OpAdd, 5 * sec, 1.5, "not defined"},
		{"plus a String", functions.OpAdd, 5 * sec, "x", "not defined"},
		{"times a Duration", functions.OpMul, 5 * sec, 5 * sec, "multiply a Duration by a number"},
		{"number over a Duration", functions.OpDiv, 5, 5 * sec, "cannot be divided by a length of time"},
		{"power", functions.OpPow, 5 * sec, 2, "not defined"},
		{"idiv by a number", functions.OpIDiv, 5 * sec, 2, "not defined"},
		{"mod by a number", functions.OpMod, 5 * sec, 2, "not defined"},
		{"div by zero Duration", functions.OpDiv, hour, functions.Duration(0), "division by zero"},
		{"idiv by zero Duration", functions.OpIDiv, hour, functions.Duration(0), "division by zero"},
		{"mod by zero Duration", functions.OpMod, hour, functions.Duration(0), "modulo by zero"},
		{"div by zero Int", functions.OpDiv, hour, 0, "division by zero"},
		{"div by zero Float", functions.OpDiv, hour, 0.0, "division by zero"},
		{"times NaN", functions.OpMul, hour, math.NaN(), "cannot scale"},
		{"times Inf", functions.OpMul, hour, math.Inf(1), "cannot scale"},
		{"times a Complex", functions.OpMul, hour, complex(1, 1), "cannot scale a Duration by Complex"},
		{"times a String", functions.OpMul, hour, "3", "cannot scale a Duration by String"},
		{"order against an Int", functions.OpLt, 5 * sec, 5, "not defined for Duration and Int"},
		{"order against a String", functions.OpGt, "5s", 5 * sec, "not defined for String and Duration"},
		{"overflow add", functions.OpAdd, bigDur, functions.Duration(1), "Duration overflow"},
		{"overflow add negative", functions.OpAdd, functions.Duration(math.MinInt64), functions.Duration(-1), "Duration overflow"},
		{"overflow sub", functions.OpSub, functions.Duration(math.MinInt64), functions.Duration(1), "Duration overflow"},
		{"overflow sub of min", functions.OpSub, functions.Duration(0), functions.Duration(math.MinInt64), "Duration overflow"},
		{"overflow mul", functions.OpMul, bigDur, 2, "Duration overflow"},
		{"overflow mul float", functions.OpMul, bigDur, 1e30, "Duration overflow"},
		{"overflow div by a tiny ratio", functions.OpDiv, bigDur, 0.5, "Duration overflow"},
	} {
		got, err := eval(t, tc.op, tc.l, tc.r)
		if err == nil {
			t.Errorf("%s: %v %s %v = %#v, want an error containing %q", tc.name, tc.l, tc.op, tc.r, got, tc.want)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %q does not contain %q", tc.name, err, tc.want)
		}
	}
}

func TestDurationEqualityAndOrder(t *testing.T) {
	if !functions.StructuralEqual(5*sec, functions.Duration(5000e6)) {
		t.Error("5s must equal 5000ms")
	}
	if functions.StructuralEqual(5*sec, 5) || functions.StructuralEqual(5, 5*sec) || functions.StructuralEqual(5*sec, int64(5e9)) {
		t.Error("a Duration never equals a number")
	}
	if functions.StructuralEqual(5*sec, "5s") {
		t.Error("a Duration never equals the String of its literal")
	}
	if got := mustEval(t, functions.OpEq, 5*sec, 5); got != false {
		t.Errorf("5s == 5 = %v, want false", got)
	}
	if got := mustEval(t, functions.OpNeq, 5*sec, 5); got != true {
		t.Errorf("5s != 5 = %v, want true", got)
	}
	for _, tc := range []struct {
		op   functions.OpKind
		l, r functions.Duration
		want bool
	}{
		{functions.OpLt, sec, min_, true}, {functions.OpLt, min_, sec, false}, {functions.OpLt, sec, sec, false},
		{functions.OpLte, sec, sec, true}, {functions.OpGt, hour, min_, true}, {functions.OpGte, min_, hour, false},
		{functions.OpLt, -sec, sec, true}, {functions.OpEq, sec, sec, true}, {functions.OpNeq, sec, min_, true},
	} {
		if got := mustEval(t, tc.op, tc.l, tc.r); got != tc.want {
			t.Errorf("%v %s %v = %v, want %v", tc.l, tc.op, tc.r, got, tc.want)
		}
	}
	if c, err := functions.Compare(sec, min_); err != nil || c != -1 {
		t.Errorf("Compare(1s, 1m) = %d, %v", c, err)
	}
	if c, err := functions.Compare(min_, sec); err != nil || c != 1 {
		t.Errorf("Compare(1m, 1s) = %d, %v", c, err)
	}
	if c, err := functions.Compare(sec, sec); err != nil || c != 0 {
		t.Errorf("Compare(1s, 1s) = %d, %v", c, err)
	}
	if _, err := functions.Compare(sec, 1); err == nil {
		t.Error("Compare(Duration, Int) must be an error")
	}
	if _, err := functions.Compare("a", sec); err == nil {
		t.Error("Compare(String, Duration) must be an error")
	}
}

func TestDurationNegateAbs(t *testing.T) {
	if got, err := functions.Negate(5 * sec); err != nil || got != functions.Duration(-5*sec) {
		t.Errorf("Negate(5s) = %#v, %v", got, err)
	}
	if got, err := functions.Abs(-5 * sec); err != nil || got != functions.Duration(5*sec) {
		t.Errorf("Abs(-5s) = %#v, %v", got, err)
	}
	if got, err := functions.Abs(5 * sec); err != nil || got != functions.Duration(5*sec) {
		t.Errorf("Abs(5s) = %#v, %v", got, err)
	}
	if _, err := functions.Negate(functions.Duration(math.MinInt64)); err == nil {
		t.Error("negating the smallest Duration must be an overflow error")
	}
	if _, err := functions.Abs(functions.Duration(math.MinInt64)); err == nil {
		t.Error("abs of the smallest Duration must be an overflow error")
	}
}

// The structural contract: a Duration is its own Go type, never one of the numbers.
func TestDurationIsNotANumber(t *testing.T) {
	var v any = 5 * sec
	if functions.IsNumber(v) || functions.IsNumeric(v) {
		t.Error("a Duration must not be a number")
	}
	if !functions.IsDuration(v) || !functions.IsTemporal(v) {
		t.Error("IsDuration/IsTemporal must hold")
	}
	if functions.IsDuration(int64(5)) {
		t.Error("an int64 is not a Duration")
	}
	if functions.KindName(v) != "Duration" {
		t.Errorf("KindName = %q", functions.KindName(v))
	}
}
