package functions_test

import (
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

// rat builds a Ratio (or the Int it demotes to) from n/d.
func rat(t *testing.T, n, d int64) any {
	t.Helper()
	v, err := functions.NewRatio(big.NewInt(n), big.NewInt(d))
	if err != nil {
		t.Fatalf("NewRatio(%d, %d): %v", n, d, err)
	}
	return v
}

func lit(t *testing.T, text string) any {
	t.Helper()
	v, ok := functions.ParseNumberLiteral(text)
	if !ok {
		t.Fatalf("ParseNumberLiteral(%q) failed", text)
	}
	return v
}

func op(t *testing.T, o functions.OpKind, a, b any) any {
	t.Helper()
	v, err := functions.EvalBinaryOp(o, a, b)
	if err != nil {
		t.Fatalf("%v %s %v: %v", a, o, b, err)
	}
	return v
}

// The canonical form: lowest terms, positive denominator, and a whole fraction is an Int.
func TestRatioCanonicalForm(t *testing.T) {
	r, ok := rat(t, 6, -8).(*functions.Ratio)
	if !ok {
		t.Fatalf("6/-8 is %T, want *functions.Ratio", rat(t, 6, -8))
	}
	if r.Num().Int64() != -3 || r.Denom().Int64() != 4 || r.String() != "-3r4" {
		t.Errorf("6/-8 = %s (num %v, den %v), want -3r4", r, r.Num(), r.Denom())
	}
	if v := rat(t, 4, 2); v != 2 {
		t.Errorf("4/2 = %#v, want the int 2, not a Ratio", v)
	}
	if v := rat(t, 0, 5); v != 0 {
		t.Errorf("0/5 = %#v, want the int 0", v)
	}
	// arithmetic demotes the same way
	if v := op(t, functions.OpAdd, rat(t, 1, 2), rat(t, 1, 2)); v != 1 {
		t.Errorf("1r2 + 1r2 = %#v, want the int 1", v)
	}
	if v := op(t, functions.OpMul, rat(t, 2, 3), rat(t, 3, 2)); v != 1 {
		t.Errorf("2r3 * 3r2 = %#v, want the int 1", v)
	}
	if _, err := functions.NewRatio(big.NewInt(1), big.NewInt(0)); err == nil {
		t.Error("a zero denominator must be an error")
	}
	// the value is never aliased: mutating a returned numerator does not change the Ratio
	r.Num().SetInt64(99)
	if r.String() != "-3r4" {
		t.Errorf("Num() aliased the Ratio: %s", r)
	}
}

func TestRatioLiterals(t *testing.T) {
	for text, want := range map[string]string{
		"3r4": "3r4", "6r8": "3r4", "0.75r": "3r4", "12.5r": "25r2", "0.1r": "1r10", "+3r4": "3r4", "-3r4": "-3r4", "-0.5r": "-1r2",
	} {
		v, ok := functions.ParseNumberLiteral(text)
		if !ok {
			t.Errorf("%q must parse", text)
			continue
		}
		r, isRatio := v.(*functions.Ratio)
		if !isRatio || r.String() != want {
			t.Errorf("%q = %#v, want Ratio %s", text, v, want)
		}
	}
	for text, want := range map[string]int{"4r2": 2, "0r5": 0, "10r5": 2, "2.0r": 2} {
		if v := lit(t, text); !functions.StructuralEqual(v, want) {
			t.Errorf("%q = %#v, want Int %d", text, v, want)
		} else if _, isRatio := v.(*functions.Ratio); isRatio {
			t.Errorf("%q stayed a Ratio", text)
		}
	}
	for _, bad := range []string{"1r0", "3r", "r4", "1.5r2", "3r-4", "--3r4", "1r2r3", "0.5.5r"} {
		if v, ok := functions.ParseNumberLiteral(bad); ok {
			t.Errorf("%q = %#v, want a rejection", bad, v)
		}
	}
	if _, err := functions.ParseRatioLiteral("1r0"); err == nil || !strings.Contains(err.Error(), "denominator is zero") {
		t.Errorf("1r0: %v, want a denominator-is-zero error", err)
	}
	// the printed form round-trips through the lexer's reader
	for _, text := range []string{"3r4", "-7r2", "1r1000", "3602879701896397r36028797018963968"} {
		v := lit(t, text)
		if v.(*functions.Ratio).String() != text {
			t.Errorf("%q printed as %v", text, v)
		}
		if _, ok := functions.ParseCanonicalNumber(text); !ok {
			t.Errorf("%q must be a canonical number token (it prints back unchanged)", text)
		}
	}
	for _, text := range []string{"6r8", "0.75r", "4r2"} {
		if _, ok := functions.ParseCanonicalNumber(text); ok {
			t.Errorf("%q does not print back unchanged, so it is not canonical", text)
		}
	}
}

func TestRatioParseText(t *testing.T) {
	for text, want := range map[string]string{"3/4": "3r4", "3r4": "3r4", "-0.75": "-3r4", "1.5e-3": "3r2000", "12": "12", "6/3": "2", ".5": "1r2"} {
		v, ok := functions.ParseRatioText(text)
		if !ok {
			t.Errorf("ParseRatioText(%q) failed", text)
			continue
		}
		got := ""
		if r, isR := v.(*functions.Ratio); isR {
			got = r.String()
		} else {
			b, _ := functions.ToBigInt(v)
			got = b.String()
		}
		if got != want {
			t.Errorf("ParseRatioText(%q) = %s, want %s", text, got, want)
		}
	}
	for _, bad := range []string{"", " 1/2", "1/2 ", "1/0", "nan", "inf", "0x10", "1e9999", "1/", "/2", "a", "1_000", "1/2/3", strings.Repeat("1", 600)} {
		if v, ok := functions.ParseRatioText(bad); ok {
			t.Errorf("ParseRatioText(%q) = %#v, want a rejection", bad, v)
		}
	}
}

// The field axioms over a table of mixed Int and Ratio values (exactness, not tolerance).
func TestRatioFieldAxioms(t *testing.T) {
	vals := []any{
		0, 1, -1, 2, -7, 100, rat(t, 1, 2), rat(t, -1, 3), rat(t, 22, 7), rat(t, 1, 10), rat(t, -355, 113),
		rat(t, 3602879701896397, 36028797018963968), functions.CanonInt(bigOf("100000000000000000000")), lit(t, "7r1000000000000000000000"),
	}
	eq := func(a, b any) bool { return functions.StructuralEqual(a, b) }
	for _, a := range vals {
		if !eq(op(t, functions.OpSub, a, a), 0) {
			t.Errorf("a - a != 0 for %v", a)
		}
		if !eq(op(t, functions.OpAdd, a, 0), a) || !eq(op(t, functions.OpMul, a, 1), a) {
			t.Errorf("identity elements fail for %v", a)
		}
		if !eq(op(t, functions.OpAdd, a, op(t, functions.OpSub, 0, a)), 0) {
			t.Errorf("a + (-a) != 0 for %v", a)
		}
		if !functions.StructuralEqual(a, 0) {
			inv, err := functions.MakeRatio(1, a)
			if err != nil {
				t.Fatalf("1/%v: %v", a, err)
			}
			if !eq(op(t, functions.OpMul, a, inv), 1) {
				t.Errorf("a * (1/a) != 1 for %v", a)
			}
		}
		for _, b := range vals {
			if !eq(op(t, functions.OpAdd, a, b), op(t, functions.OpAdd, b, a)) || !eq(op(t, functions.OpMul, a, b), op(t, functions.OpMul, b, a)) {
				t.Errorf("commutativity fails for %v, %v", a, b)
			}
			for _, c := range vals {
				if !eq(op(t, functions.OpAdd, op(t, functions.OpAdd, a, b), c), op(t, functions.OpAdd, a, op(t, functions.OpAdd, b, c))) {
					t.Errorf("+ associativity fails for %v, %v, %v", a, b, c)
				}
				if !eq(op(t, functions.OpMul, op(t, functions.OpMul, a, b), c), op(t, functions.OpMul, a, op(t, functions.OpMul, b, c))) {
					t.Errorf("* associativity fails for %v, %v, %v", a, b, c)
				}
				if !eq(op(t, functions.OpMul, a, op(t, functions.OpAdd, b, c)), op(t, functions.OpAdd, op(t, functions.OpMul, a, b), op(t, functions.OpMul, a, c))) {
					t.Errorf("distributivity fails for %v, %v, %v", a, b, c)
				}
			}
			// floor division and the modulo that goes with it, for a non-zero divisor
			if !functions.StructuralEqual(b, 0) {
				q := op(t, functions.OpIDiv, a, b)
				m := op(t, functions.OpMod, a, b)
				if !eq(op(t, functions.OpAdd, op(t, functions.OpMul, b, q), m), a) {
					t.Errorf("a == b*(a//b) + a%%b fails for %v, %v (q=%v m=%v)", a, b, q, m)
				}
			}
		}
	}
	// the headline: Ratio does what Float cannot
	if !functions.StructuralEqual(op(t, functions.OpAdd, lit(t, "0.1r"), lit(t, "0.2r")), lit(t, "0.3r")) {
		t.Error("0.1r + 0.2r must equal 0.3r exactly")
	}
	if got, _ := functions.EvalBinaryOp(functions.OpEq, op(t, functions.OpAdd, 0.1, 0.2), 0.3); got != false {
		t.Error("0.1 + 0.2 == 0.3 must stay false for Float")
	}
}

func TestRatioMixedArithmeticAndPromotion(t *testing.T) {
	rows := []struct {
		o    functions.OpKind
		a, b any
		want any
	}{
		{functions.OpDiv, rat(t, 1, 3), 2, rat(t, 1, 6)}, // Ratio / Int stays exact
		{functions.OpDiv, 1, rat(t, 3, 1), 1.0 / 3},      // 3r1 is the Int 3: Int / Int is unchanged (D11)
		{functions.OpDiv, 1, rat(t, 3, 2), rat(t, 2, 3)}, // Int / Ratio exact
		{functions.OpAdd, rat(t, 1, 2), 0.5, 1.0},        // D12: arithmetic with a Float promotes to Float
		{functions.OpMul, rat(t, 1, 3), 3.0, 1.0},        // the Float result is the correctly rounded 1/3 * 3
		{functions.OpSub, rat(t, 1, 2), complex(0, 1), complex(0.5, -1)},
		{functions.OpIDiv, rat(t, 7, 2), rat(t, 1, 2), 7}, // floor quotient is an Int
		{functions.OpIDiv, rat(t, -7, 2), 2, -2},
		{functions.OpMod, rat(t, 7, 2), rat(t, 1, 3), rat(t, 1, 6)}, // a - b*(a//b)
		{functions.OpMod, rat(t, -7, 2), 2, rat(t, 1, 2)},
		{functions.OpPow, rat(t, 1, 2), 3, rat(t, 1, 8)},
		{functions.OpPow, rat(t, 1, 2), -3, 8},
		{functions.OpPow, rat(t, 2, 3), 0, 1},
		{functions.OpPow, rat(t, 1, 4), 0.5, 0.5}, // a non-integer exponent promotes to Float
		{functions.OpPow, 4, rat(t, 1, 2), 2.0},
	}
	for _, r := range rows {
		got := op(t, r.o, r.a, r.b)
		if !functions.StructuralEqual(got, r.want) {
			t.Errorf("%v %s %v = %#v, want %#v", r.a, r.o, r.b, got, r.want)
		}
	}
	// the promoted result really is a Float, the exact one really is a *Ratio
	if _, isF := op(t, functions.OpAdd, rat(t, 1, 2), 0.25).(float64); !isF {
		t.Error("Ratio + Float must be a float64")
	}
	if _, isR := op(t, functions.OpAdd, rat(t, 1, 2), 1).(*functions.Ratio); !isR {
		t.Error("Ratio + Int must be a *Ratio")
	}
	// errors
	for _, b := range []struct {
		o    functions.OpKind
		a, b any
	}{
		{functions.OpDiv, rat(t, 1, 2), 0}, {functions.OpIDiv, rat(t, 1, 2), 0}, {functions.OpMod, rat(t, 1, 2), 0},
		{functions.OpDiv, 1, rat(t, 0, 1)}, {functions.OpPow, 0, -1}, {functions.OpPow, rat(t, 1, 2), 99999999},
		{functions.OpAdd, rat(t, 1, 2), "a"}, {functions.OpMul, "a", rat(t, 1, 2)}, {functions.OpAdd, rat(t, 1, 2), functions.Char('a')},
		{functions.OpAdd, lit(t, strings.Repeat("9", 400)+"r2"), 1.5}, // a Ratio beyond the double range is not silently Inf
	} {
		if v, err := functions.EvalBinaryOp(b.o, b.a, b.b); err == nil {
			t.Errorf("%v %s %v = %#v, want an error", b.a, b.o, b.b, v)
		}
	}
}

// D12: Ratio against Float compares EXACTLY, through the double's binary value.
func TestRatioComparisonIsExact(t *testing.T) {
	tenth := rat(t, 1, 10)
	rows := []struct {
		a, b    any
		cmp     int
		ordered bool
	}{
		{tenth, 0.1, -1, true}, // 1/10 < the double 0.1 (0.1000000000000000055…)
		{0.1, tenth, 1, true},
		{rat(t, 1, 2), 0.5, 0, true},
		{rat(t, 1, 2), 1, -1, true},
		{2, rat(t, 5, 2), -1, true},
		{rat(t, -1, 3), rat(t, -1, 2), 1, true},
		{rat(t, 7, 2), math.Inf(1), -1, true},
		{rat(t, 7, 2), math.Inf(-1), 1, true},
		{rat(t, 7, 2), math.NaN(), 0, false},
		{math.NaN(), rat(t, 7, 2), 0, false},
		{lit(t, "1r3"), lit(t, "2r6"), 0, true},
	}
	for _, r := range rows {
		cmp, ordered, ok := functions.CompareNumbers(r.a, r.b)
		if !ok || ordered != r.ordered || (ordered && cmp != r.cmp) {
			t.Errorf("CompareNumbers(%v, %v) = %d,%v,%v; want %d,%v", r.a, r.b, cmp, ordered, ok, r.cmp, r.ordered)
		}
	}
	for _, c := range []struct {
		o    functions.OpKind
		a, b any
		want bool
	}{
		{functions.OpEq, tenth, 0.1, false}, // the headline of D12
		{functions.OpNeq, tenth, 0.1, true},
		{functions.OpEq, rat(t, 1, 2), 0.5, true},
		{functions.OpLt, tenth, 0.1, true},
		{functions.OpEq, rat(t, 3, 4), rat(t, 3, 4), true},
		{functions.OpEq, rat(t, 3, 4), "3r4", false},
		{functions.OpEq, rat(t, 3, 4), 1, false},
		{functions.OpLte, rat(t, 3, 4), 1, true},
	} {
		if got := op(t, c.o, c.a, c.b); got != c.want {
			t.Errorf("%v %s %v = %v, want %v", c.a, c.o, c.b, got, c.want)
		}
	}
	// the answer to "is this float the tenth I meant?"
	tr, err := functions.ToRatio(0.1)
	if err != nil || !functions.StructuralEqual(tr, tenth) {
		t.Errorf("to_ratio 0.1 = %v, %v; want 1r10", tr, err)
	}
	// sorting a mixed list of Ints and Ratios by Compare
	if c, _ := functions.Compare(rat(t, 1, 3), rat(t, 1, 2)); c != -1 {
		t.Errorf("Compare(1r3, 1r2) = %d", c)
	}
}

func TestRatioConversions(t *testing.T) {
	// to_ratio: the shortest decimal; to_ratio_exact: the binary value
	for in, want := range map[float64]string{0.1: "1r10", 0.75: "3r4", 0.001: "1r1000", -2.5: "-5r2", 1e21: "1000000000000000000000", 0.3: "3r10"} {
		got, err := functions.ToRatio(in)
		if err != nil || !functions.StructuralEqual(got, lit(t, want)) {
			t.Errorf("ToRatio(%v) = %v, %v; want %s", in, got, err, want)
		}
	}
	if got, _ := functions.ToRatioExact(0.1); got.(*functions.Ratio).String() != "3602879701896397r36028797018963968" {
		t.Errorf("ToRatioExact(0.1) = %v", got)
	}
	if got, _ := functions.ToRatioExact(0.5); !functions.StructuralEqual(got, rat(t, 1, 2)) {
		t.Errorf("ToRatioExact(0.5) = %v", got)
	}
	if got, _ := functions.ToRatio(7); got != 7 {
		t.Errorf("ToRatio(Int) = %#v, want the Int itself", got)
	}
	// to_float is correctly rounded: the round trip through a decimal-representable value
	for _, text := range []string{"1r10", "3r4", "22r7", "-355r113", "1r3", "123456789r1000"} {
		r := lit(t, text)
		f, err := functions.ToFloat(r)
		if err != nil {
			t.Fatal(err)
		}
		// the float must be the nearest double: no other double is closer
		fr, _ := functions.ToRatioExact(f)
		diff := op(t, functions.OpSub, fr, r)
		next, _ := functions.ToRatioExact(math.Nextafter(f.(float64), math.Inf(1)))
		prev, _ := functions.ToRatioExact(math.Nextafter(f.(float64), math.Inf(-1)))
		abs := func(v any) any { a, _ := functions.Abs(v); return a }
		if c, _, _ := functions.CompareNumbers(abs(diff), abs(op(t, functions.OpSub, next, r))); c > 0 {
			t.Errorf("to_float %s = %v is not the nearest double (next is closer)", text, f)
		}
		if c, _, _ := functions.CompareNumbers(abs(diff), abs(op(t, functions.OpSub, prev, r))); c > 0 {
			t.Errorf("to_float %s = %v is not the nearest double (previous is closer)", text, f)
		}
	}
	for _, text := range []string{"1r10", "3r4", "7r8", "-3r1000"} { // decimal-representable
		r := lit(t, text)
		f, _ := functions.ToFloat(r)
		back, err := functions.ToRatio(f)
		if err != nil || !functions.StructuralEqual(back, r) {
			t.Errorf("to_ratio (to_float %s) = %v, %v; want %s", text, back, err, text)
		}
	}
	// underflow is documented: a fraction below the smallest double converts to 0.0, not an error
	if f, err := functions.ToFloat(lit(t, "1r1"+strings.Repeat("0", 400))); err != nil || f != 0.0 {
		t.Errorf("to_float of a tiny Ratio = %v, %v; want 0 (documented underflow)", f, err)
	}
	// negative cases
	for name, f := range map[string]func() (any, error){
		"to_ratio NaN":          func() (any, error) { return functions.ToRatio(math.NaN()) },
		"to_ratio +Inf":         func() (any, error) { return functions.ToRatio(math.Inf(1)) },
		"to_ratio String":       func() (any, error) { return functions.ToRatio("0.5") },
		"to_ratio Complex":      func() (any, error) { return functions.ToRatio(complex(1, 0)) },
		"to_ratio_exact NaN":    func() (any, error) { return functions.ToRatioExact(math.NaN()) },
		"to_ratio_exact String": func() (any, error) { return functions.ToRatioExact("x") },
		"approx max_den 0":      func() (any, error) { return functions.ToRatioApprox(0.5, 0) },
		"approx max_den -3":     func() (any, error) { return functions.ToRatioApprox(0.5, -3) },
		"approx max_den Float":  func() (any, error) { return functions.ToRatioApprox(0.5, 2.5) },
		"approx NaN":            func() (any, error) { return functions.ToRatioApprox(math.NaN(), 10) },
		"to_int 1r2":            func() (any, error) { return functions.ToInt(rat(t, 1, 2)) },
		"to_float huge":         func() (any, error) { return functions.ToFloat(lit(t, strings.Repeat("9", 400)+"r2")) },
		"ratio Float":           func() (any, error) { return functions.MakeRatio(0.5, 2) },
		"ratio String":          func() (any, error) { return functions.MakeRatio("1", 2) },
		"ratio zero":            func() (any, error) { return functions.MakeRatio(1, 0) },
		"ratio zero Ratio":      func() (any, error) { return functions.MakeRatio(1, rat(t, 0, 1)) },
		"numer Float":           func() (any, error) { return functions.Numer(0.5) },
		"denom String":          func() (any, error) { return functions.Denom("1/2") },
		"mediant Float":         func() (any, error) { return functions.RatioMediant(0.5, 1) },
		"from_cf empty":         func() (any, error) { return functions.RatioFromCF(nil) },
		"from_cf zero tail":     func() (any, error) { return functions.RatioFromCF([]any{1, 0}) },
		"from_cf Float":         func() (any, error) { return functions.RatioFromCF([]any{1, 2.5}) },
	} {
		if v, err := f(); err == nil {
			t.Errorf("%s = %#v, want an error", name, v)
		}
	}
	if _, _, err := functions.RatioMixed(0.5); err == nil {
		t.Error("ratio_mixed of a Float must be an error")
	}
	if _, err := functions.RatioCF("x"); err == nil {
		t.Error("ratio_cf of a String must be an error")
	}
}

// to_ratio_approx is the best rational approximation with a bounded denominator: checked
// against brute force, and on the convergents of pi.
func TestRatioApprox(t *testing.T) {
	for md, want := range map[int]string{1: "3", 7: "22r7", 100: "311r99", 113: "355r113", 1000: "355r113"} {
		got, err := functions.ToRatioApprox(3.14159265, md)
		if err != nil || !functions.StructuralEqual(got, lit(t, want)) {
			t.Errorf("approx pi with max_den %d = %v, %v; want %s", md, got, err, want)
		}
	}
	if got, _ := functions.ToRatioApprox(0.3333333, 100); !functions.StructuralEqual(got, rat(t, 1, 3)) {
		t.Errorf("approx 0.3333333 = %v, want 1r3", got)
	}
	// brute force: no fraction with a denominator <= max_den is strictly closer
	for _, x := range []float64{0.618, 1.4142135, -2.7182818, 0.5772156, 12.34567} {
		xr, _ := functions.ToRatio(x)
		for md := 1; md <= 60; md++ {
			got, err := functions.ToRatioApprox(x, md)
			if err != nil {
				t.Fatal(err)
			}
			gr, _ := functions.ToRatioExact(got) // exact for Int and Ratio alike
			errGot, _ := functions.Abs(op(t, functions.OpSub, gr, xr))
			if d, _ := functions.Denom(got); functions.StructuralEqual(d, 0) || mustCmp(t, d, md) > 0 {
				t.Errorf("approx %v max_den %d = %v has denominator %v", x, md, got, d)
			}
			for q := 1; q <= md; q++ {
				p, _ := functions.RoundToInt(functions.RoundHalfAway, op(t, functions.OpMul, xr, q))
				cand := rat2(t, p, q)
				errCand, _ := functions.Abs(op(t, functions.OpSub, cand, xr))
				if mustCmp(t, errCand, errGot) < 0 {
					t.Errorf("approx %v max_den %d = %v, but %v is closer", x, md, got, cand)
				}
			}
		}
	}
	// an exact fraction within the bound is returned unchanged
	if got, _ := functions.ToRatioApprox(rat(t, 3, 8), 100); !functions.StructuralEqual(got, rat(t, 3, 8)) {
		t.Errorf("approx 3r8 = %v", got)
	}
}

func rat2(t *testing.T, n any, d int) any {
	t.Helper()
	v, err := functions.MakeRatio(n, d)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func mustCmp(t *testing.T, a, b any) int {
	t.Helper()
	c, ordered, ok := functions.CompareNumbers(a, b)
	if !ok || !ordered {
		t.Fatalf("cannot compare %v and %v", a, b)
	}
	return c
}

func TestRatioRounding(t *testing.T) {
	type row struct {
		in         any
		floor, cel int
		trunc, rnd int
		even       int
	}
	for _, r := range []row{
		{rat(t, 7, 2), 3, 4, 3, 4, 4},
		{rat(t, -7, 2), -4, -3, -3, -4, -4},
		{rat(t, 5, 2), 2, 3, 2, 3, 2},
		{rat(t, -5, 2), -3, -2, -2, -3, -2},
		{rat(t, 1, 3), 0, 1, 0, 0, 0},
		{rat(t, 2, 3), 0, 1, 0, 1, 1},
		{rat(t, -1, 3), -1, 0, 0, 0, 0},
		{rat(t, -2, 3), -1, 0, 0, -1, -1},
		{rat(t, 22, 7), 3, 4, 3, 3, 3},
	} {
		for mode, want := range map[functions.RoundingMode]int{
			functions.RoundFloor: r.floor, functions.RoundCeil: r.cel, functions.RoundTrunc: r.trunc,
			functions.RoundHalfAway: r.rnd, functions.RoundHalfEven: r.even,
		} {
			got, err := functions.RoundToInt(mode, r.in)
			if err != nil || !functions.StructuralEqual(got, want) {
				t.Errorf("mode %d of %v = %v, %v; want %d", mode, r.in, got, err, want)
			}
		}
	}
}

func TestRatioParts(t *testing.T) {
	r := rat(t, -7, 2)
	if n, _ := functions.Numer(r); n != -7 {
		t.Errorf("numer = %v", n)
	}
	if d, _ := functions.Denom(r); d != 2 {
		t.Errorf("denom = %v", d)
	}
	if n, _ := functions.Numer(5); n != 5 {
		t.Errorf("numer of an Int = %v", n)
	}
	if d, _ := functions.Denom(5); d != 1 {
		t.Errorf("denom of an Int = %v", d)
	}
	w, f, err := functions.RatioMixed(rat(t, 7, 2))
	if err != nil || w != 3 || !functions.StructuralEqual(f, rat(t, 1, 2)) {
		t.Errorf("mixed 7r2 = %v, %v, %v", w, f, err)
	}
	w, f, _ = functions.RatioMixed(rat(t, -7, 2))
	if w != -3 || !functions.StructuralEqual(f, rat(t, -1, 2)) {
		t.Errorf("mixed -7r2 = %v, %v", w, f)
	}
	// continued fractions round-trip
	for _, r := range []any{rat(t, 7, 3), rat(t, 355, 113), rat(t, -22, 7), rat(t, 1, 10), 5, rat(t, 3, 8)} {
		cf, err := functions.RatioCF(r)
		if err != nil {
			t.Fatal(err)
		}
		back, err := functions.RatioFromCF(cf)
		if err != nil || !functions.StructuralEqual(back, r) {
			t.Errorf("from_cf(cf(%v)) = %v (cf %v), %v", r, back, cf, err)
		}
	}
	if cf, _ := functions.RatioCF(rat(t, 7, 3)); len(cf) != 2 || cf[0] != 2 || cf[1] != 3 {
		t.Errorf("cf 7r3 = %v, want [2 3]", cf)
	}
	if m, _ := functions.RatioMediant(rat(t, 1, 2), rat(t, 2, 3)); !functions.StructuralEqual(m, rat(t, 3, 5)) {
		t.Errorf("mediant 1r2 2r3 = %v, want 3r5", m)
	}
}

func TestRatioKindAndNumericness(t *testing.T) {
	r := rat(t, 1, 2)
	if functions.KindName(r) != "Ratio" || !functions.IsNumeric(r) || !functions.IsNumber(r) {
		t.Errorf("Ratio kind = %s numeric=%v", functions.KindName(r), functions.IsNumeric(r))
	}
	if f, ok := functions.ToFloat64(r); !ok || f != 0.5 {
		t.Errorf("ToFloat64(1r2) = %v, %v", f, ok)
	}
	if _, ok := functions.ToBigInt(r); ok {
		t.Error("a Ratio is not an Int: ToBigInt must refuse it")
	}
	if c, ok := functions.ToComplex128(r); !ok || c != complex(0.5, 0) {
		t.Errorf("ToComplex128(1r2) = %v, %v", c, ok)
	}
	if s, _ := functions.Sign(rat(t, -1, 2)); s != -1 {
		t.Errorf("sign -1r2 = %v", s)
	}
	if a, _ := functions.Abs(rat(t, -1, 2)); !functions.StructuralEqual(a, rat(t, 1, 2)) {
		t.Errorf("abs -1r2 = %v", a)
	}
	if n, _ := functions.Negate(r); !functions.StructuralEqual(n, rat(t, -1, 2)) {
		t.Errorf("negate 1r2 = %v", n)
	}
	if s, _ := functions.Sqrt(rat(t, 1, 4)); s != 0.5 {
		t.Errorf("sqrt 1r4 = %v", s)
	}
}
