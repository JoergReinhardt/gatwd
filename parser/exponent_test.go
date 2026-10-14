package parser_test

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/parser"
)

// F6: an `e` exponent makes a Float literal; `5em`-style words and malformed tails stay errors.
func TestExponentLiterals(t *testing.T) {
	for src, want := range map[string]float64{"1e3": 1000, "1.5e-3": 0.0015, "2E10": 2e10, "1e+2": 100, "2 * 1e1": 20} {
		got, err := parser.EvalInfixString(src, nil)
		if f, ok := got.(float64); err != nil || !ok || f != want {
			t.Errorf("%s = %#v, %v; want float64 %v", src, got, err, want)
		}
	}
	for _, tc := range []struct{ src, want string }{
		{"1e400", "out of range"},
		{"1e3x", "after the exponent"},
		{"1e3.5", "after the exponent"},
		{"1e5q", "after the exponent"},
	} {
		_, err := parser.EvalInfixString(tc.src, nil)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error = %v, want one containing %q", tc.src, err, tc.want)
		}
	}
	// Not an exponent: `2e` has no digits, so `e` stays an identifier (an application).
	if _, err := parser.EvalInfixString("2e", func(string) (any, bool) { return nil, false }); err == nil {
		t.Error("`2e` must not lex as an exponent literal")
	}
}
