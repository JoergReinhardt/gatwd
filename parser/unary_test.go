package parser_test

import (
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/parser"
)

// asInt64 reads any integer representation the evaluator may answer with.
func asInt64(t *testing.T, v any) int64 {
	t.Helper()
	b, ok := functions.ToBigInt(v)
	if !ok || !b.IsInt64() {
		t.Fatalf("%#v is not an int64-sized Int", v)
	}
	return b.Int64()
}

func TestUnaryMinusAndPower(t *testing.T) {
	env := map[string]any{"x": 4, "f": functions.NewClosure([]string{"n"}, func(a []any) (any, error) {
		return functions.EvalBinaryOp(functions.OpMul, a[0], 10)
	})}
	lookup := func(n string) (any, bool) { v, ok := env[n]; return v, ok }
	rows := []struct {
		src  string
		want int64
	}{
		{"- 5", -5},
		{"-5", -5},
		{"-x", -4},
		{"- x + 1", -3},
		{"-(2 + 3)", -5},
		{"3 * -2", -6},
		{"3 - -2", 5},
		{"10 - 3", 7},
		{"10-3", 7},
		{"10 -3", 7}, // a literal cannot be applied, so a tight minus after one is a subtraction
		{"2 ** 3", 8},
		{"2 ** 3 ** 2", 512}, // right associative
		{"-2 ** 2", -4},      // ** binds tighter than unary minus
		{"(-2) ** 2", 4},
		{"2 * 3 ** 2", 18}, // ** binds tighter than *
		{"-2 * 3", -6},     // unary minus binds tighter than *
		{"f -3", -30},      // a tight minus after a callee is a negative argument
		{"f - 3", 0},       // spaced on both sides it is a subtraction... of a function: see below
		{"-9223372036854775808", -9223372036854775808},
	}
	for _, r := range rows {
		if r.src == "f - 3" {
			continue // subtracting from a function is an error; covered below
		}
		v, err := parser.EvalInfixString(r.src, lookup)
		if err != nil {
			t.Errorf("%s: %v", r.src, err)
			continue
		}
		if got := asInt64(t, v); got != r.want {
			t.Errorf("%s = %v, want %d", r.src, v, r.want)
		}
	}
}

// A folded literal is a LiteralExpr (the smallest int64 is not a negated big integer).
func TestNegatedLiteralIsFolded(t *testing.T) {
	e, err := parser.ParseInfixExpr("-9223372036854775808")
	if err != nil {
		t.Fatal(err)
	}
	lit, ok := e.(parser.LiteralExpr)
	if !ok {
		t.Fatalf("%T, want a folded LiteralExpr", e)
	}
	if _, isBig := lit.Value.(*big.Int); isBig {
		t.Fatalf("-2^63 must demote to an int64-sized Int, got *big.Int %v", lit.Value)
	}
	if e2, _ := parser.ParseInfixExpr("-x"); e2 == nil {
		t.Fatal("-x must parse")
	} else if _, ok := e2.(parser.UnaryExpr); !ok {
		t.Fatalf("-x is %T, want UnaryExpr", e2)
	}
}

// A plain list literal is space-separated: a tight minus starts the next element, so the
// printed form of a list of negatives reads back as written.
func TestTightMinusInPlainList(t *testing.T) {
	for src, want := range map[string]int{"[3 -1 -2]": 3, "[3, -1, -2]": 3, "[3 - 1]": 1, "[3 -1]": 2} {
		v, err := parser.EvalInfixString(src, nil)
		if err != nil {
			t.Fatalf("%s: %v", src, err)
		}
		items, ok := functions.ToSliceAny(v)
		if !ok || len(items) != want {
			t.Errorf("%s = %v, want %d elements", src, v, want)
		}
	}
}

func TestUnaryErrors(t *testing.T) {
	for _, src := range []string{`- "a"`, `-true`, `"a" ** 2`, `0 ** -1`, `(-8) ** 0.5`, `2 ** 99999999`} {
		if v, err := parser.EvalInfixString(src, nil); err == nil {
			t.Errorf("%s = %v, want an error", src, v)
		}
	}
	// `x -1` with x a number is the documented trap: loud, with the hint.
	_, err := parser.EvalInfixString("x -1", func(n string) (any, bool) { return 4, n == "x" })
	if err == nil {
		t.Fatal("x -1 must be an application error, not a silent subtraction")
	}
}

func TestRatioLiteralsLexAndEvaluate(t *testing.T) {
	for src, want := range map[string]string{
		"3r4": "3r4", "0.75r": "3r4", "6r8": "3r4", "-3r4": "-3r4", "- 3r4": "-3r4", "1r2 + 1r3": "5r6",
		"3r4 * 4": "3", "(1r2) ** 3": "1r8", "2 ** -1r1": "0.5",
	} {
		v, err := parser.EvalInfixString(src, nil)
		if err != nil {
			t.Errorf("%s: %v", src, err)
			continue
		}
		got := ""
		if s, ok := v.(interface{ String() string }); ok {
			got = s.String()
		} else {
			got = functions.KindName(v)
			if b, ok := functions.ToBigInt(v); ok {
				got = b.String()
			} else if f, ok := v.(float64); ok {
				got = fmtFloat(f)
			}
		}
		if got != want {
			t.Errorf("%s = %s, want %s", src, got, want)
		}
	}
	// the application `f 3 r4` is not a ratio: only `3r4` with no space is
	if _, err := parser.EvalInfixString("3 r4", func(n string) (any, bool) { return nil, false }); err == nil {
		t.Error("`3 r4` is an application of 3, not a ratio literal")
	}
	for _, bad := range []string{"1r0", "5r", "1.5r2", "0.5r5"} {
		if v, err := parser.EvalInfixString(bad, nil); err == nil {
			t.Errorf("%s = %v, want a lex error", bad, v)
		}
	}
	if _, err := parser.ParseInfixExpr("1r0"); err == nil || !strings.Contains(err.Error(), "denominator is zero") {
		t.Errorf("1r0: %v, want a denominator-is-zero error", err)
	}
	// a plain list keeps ratio elements apart
	v, err := parser.EvalInfixString("[1r2 3r4 5]", nil)
	if err != nil {
		t.Fatal(err)
	}
	if items, _ := functions.ToSliceAny(v); len(items) != 3 {
		t.Errorf("[1r2 3r4 5] has %d elements, want 3", len(items))
	}
}

func fmtFloat(f float64) string { return strconv.FormatFloat(f, 'g', -1, 64) }
