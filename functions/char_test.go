package functions_test

import (
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

func TestCharFromCode(t *testing.T) {
	for _, tc := range []struct {
		code int64
		want functions.Char
		ok   bool
	}{
		{97, 'a', true},
		{233, 'é', true},
		{0x1F600, '😀', true},
		{0, 0, true},
		{-1, 0, false},
		{0xD800, 0, false}, // surrogate half
		{0xDFFF, 0, false},
		{0x110000, 0, false},
	} {
		got, ok := functions.CharFromCode(tc.code)
		if ok != tc.ok || (ok && got != tc.want) {
			t.Errorf("CharFromCode(%#x) = %v, %v; want %v, %v", tc.code, got, ok, tc.want, tc.ok)
		}
	}
}

func TestCharOfString(t *testing.T) {
	if c, ok := functions.CharOfString("é"); !ok || c != 'é' {
		t.Errorf(`CharOfString("é") = %v, %v`, c, ok)
	}
	for _, bad := range []string{"", "ab", "\xff", "é\xff"} {
		if c, ok := functions.CharOfString(bad); ok {
			t.Errorf("CharOfString(%q) = %v, true; want false", bad, c)
		}
	}
}

func TestCharsStringRoundTrip(t *testing.T) {
	in := "aé😀"
	cs, err := functions.CharsOfString(in)
	if err != nil || len(cs) != 3 {
		t.Fatalf("CharsOfString(%q) = %v, %v", in, cs, err)
	}
	if got := functions.StringOfChars(cs); got != in {
		t.Errorf("round trip = %q, want %q", got, in)
	}
	if _, err := functions.CharsOfString("a\xffb"); err == nil {
		t.Error("invalid UTF-8 must be an error, not U+FFFD")
	}
	if got := functions.StringOfChars(nil); got != "" {
		t.Errorf("StringOfChars(nil) = %q, want empty", got)
	}
}

// A Char is its own type: it never passes for a number (the `case int32` arms of the
// numeric tower do not match a named type), and arithmetic on it is an error.
func TestCharIsNotNumeric(t *testing.T) {
	c := functions.Char('a')
	if functions.IsNumeric(c) {
		t.Error("IsNumeric(Char) must be false")
	}
	for _, op := range []functions.OpKind{functions.OpAdd, functions.OpSub, functions.OpMul} {
		if res, err := functions.EvalBinaryOp(op, c, 1); err == nil {
			t.Errorf("Char %s Int = %v, want an error", op, res)
		}
	}
}

func TestCharComparison(t *testing.T) {
	a, b := functions.Char('a'), functions.Char('b')
	for _, tc := range []struct {
		op    functions.OpKind
		l, r  any
		want  bool
		label string
	}{
		{functions.OpEq, a, a, true, "Char == Char"},
		{functions.OpEq, a, b, false, "Char == other Char"},
		{functions.OpNeq, a, b, true, "Char != Char"},
		{functions.OpLt, a, b, true, "a < b"},
		{functions.OpGte, a, b, false, "a >= b"},
		{functions.OpLte, a, a, true, "a <= a"},
		{functions.OpEq, a, "a", true, "Char == one-char String (D1)"},
		{functions.OpEq, "a", a, true, "one-char String == Char (D1)"},
		{functions.OpLt, "a", b, true, "one-char String < Char (D1)"},
		{functions.OpEq, a, "ab", false, "Char == longer String"},
		{functions.OpEq, a, 97, false, "Char is not its code point"},
	} {
		got, err := functions.EvalBinaryOp(tc.op, tc.l, tc.r)
		if err != nil {
			t.Errorf("%s: %v", tc.label, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%s = %v, want %v", tc.label, got, tc.want)
		}
	}
	if _, err := functions.Compare(a, "ab"); err == nil {
		t.Error("Compare(Char, longer String) must be an error")
	}
	if c, err := functions.Compare(functions.Char('é'), functions.Char('z')); err != nil || c <= 0 {
		t.Errorf("é (U+00E9) must order after z (U+007A) by code point, got %d, %v", c, err)
	}
}
