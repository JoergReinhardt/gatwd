package functions_test

import (
	"math/big"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// D2: `+` is numeric addition only. A String or Char operand is a type error that names both
// types and points at ++; ++ still concatenates, and the other operators are untouched.
func TestAddIsNumericOnly(t *testing.T) {
	for _, tc := range []struct {
		l, r any
		want string
	}{
		{"a", "b", "String and String"},
		{"a", 1, "String and Int"},
		{1, "a", "Int and String"},
		{"1", 2, "String and Int"}, // a numeric-looking String is still a String
		{"1", "2", "String and String"},
		{functions.Char('a'), 1, "Char and Int"},
		{"a", functions.Char('b'), "String and Char"},
		{"a", collections.Cons[any](1, collections.Nil[any]()), "String and List"},
	} {
		res, err := functions.EvalBinaryOp(functions.OpAdd, tc.l, tc.r)
		if err == nil || !strings.Contains(err.Error(), tc.want) || !strings.Contains(err.Error(), "++") {
			t.Errorf("%#v + %#v = %#v, %v; want an error containing %q and a ++ hint", tc.l, tc.r, res, err, tc.want)
		}
	}
	for _, tc := range []struct{ l, r, want any }{{1, 2, 3}, {1.5, 1, 2.5}} {
		if got, err := functions.EvalBinaryOp(functions.OpAdd, tc.l, tc.r); err != nil || !functions.StructuralEqual(got, tc.want) {
			t.Errorf("%v + %v = %v, %v; want %v", tc.l, tc.r, got, err, tc.want)
		}
	}
	if got, err := functions.EvalBinaryOp(functions.OpConcat, "a", "b"); err != nil || got != "ab" {
		t.Errorf(`"a" ++ "b" = %v, %v`, got, err)
	}
}

func TestParseNumberLiteral(t *testing.T) {
	for _, tc := range []struct {
		text string
		want any
		ok   bool
	}{
		{"5", 5, true},
		{"-5", -5, true},
		{"+5", 5, true},
		{"007", 7, true},
		{"1.5", 1.5, true},
		{"-0.25", -0.25, true},
		{"2i", complex(0, 2), true},
		{"99999999999999999999", nil, true}, // beyond int64: a *big.Int
		{"", nil, false},
		{"abc", nil, false},
		{"5x", nil, false},
		{" 5", nil, false},
		{"1.2.3", nil, false},
		{"--5", nil, false},
		{"0x10", nil, false},
	} {
		got, ok := functions.ParseNumberLiteral(tc.text)
		if ok != tc.ok {
			t.Errorf("ParseNumberLiteral(%q) ok = %v, want %v (%#v)", tc.text, ok, tc.ok, got)
			continue
		}
		if tc.want != nil && !functions.StructuralEqual(got, tc.want) {
			t.Errorf("ParseNumberLiteral(%q) = %#v, want %#v", tc.text, got, tc.want)
		}
	}
	got, _ := functions.ParseNumberLiteral("99999999999999999999")
	if b, ok := got.(*big.Int); !ok || b.String() != "99999999999999999999" {
		t.Errorf("a literal beyond int64 must be a *big.Int, got %#v", got)
	}
}

func TestParseCanonicalNumber(t *testing.T) {
	for _, text := range []string{"10", "-5", "0", "1.5", "-0.25", "99999999999999999999"} {
		if _, ok := functions.ParseCanonicalNumber(text); !ok {
			t.Errorf("%q is canonical", text)
		}
	}
	for _, text := range []string{"007", "+5", "1.50", "10.0", "2i", "1e3", "abc", "", "-0", "5 "} {
		if got, ok := functions.ParseCanonicalNumber(text); ok {
			t.Errorf("%q must stay text, got %#v", text, got)
		}
	}
}

func TestOperandValue(t *testing.T) {
	if got := functions.OperandValue(functions.OpAdd, "10"); !functions.StructuralEqual(got, 10) || functions.KindName(got) != "Int" {
		t.Errorf("a numeric operand becomes a number, got %#v", got)
	}
	if got := functions.OperandValue(functions.OpMul, "2.5"); got != 2.5 {
		t.Errorf("got %#v", got)
	}
	if got := functions.OperandValue(functions.OpConcat, "10"); got != "10" {
		t.Errorf("++ keeps its operand as text, got %#v", got)
	}
	if got := functions.OperandValue(functions.OpAdd, "abc"); got != "abc" {
		t.Errorf("a non-number stays text (and then fails loudly), got %#v", got)
	}
}

func TestKindName(t *testing.T) {
	for v, want := range map[string]any{
		"String": "s", "Char": functions.Char('a'), "Bool": true, "Float": 1.5, "Complex": complex(1, 1),
		"BigInt": big.NewInt(1), "Int": 3, "List": collections.Nil[any](),
	} {
		if got := functions.KindName(want); got != v {
			t.Errorf("KindName(%#v) = %q, want %q", want, got, v)
		}
	}
}
