package cli

import (
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

func TestCharFuncsStructural(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct {
		src  string
		want any
	}{
		{`'a'`, functions.Char('a')},
		{`char_code 'a'`, 97},
		{`char_code "é"`, 233}, // a one-character String converts on demand (D1)
		{`char_from_code 233`, functions.Char('é')},
		{`char_to_str 'x'`, "x"},
		{`char_upper 'q'`, functions.Char('Q')},
		{`char_lower 'Q'`, functions.Char('q')},
		{`char_upper "q"`, functions.Char('Q')},
		{`char_is_digit '7'`, true},
		{`char_is_digit 'x'`, false},
		{`char_is_alpha 'x'`, true},
		{`char_is_alpha '1'`, false},
		{`char_is_space ' '`, true},
		{`char_is_space 'x'`, false},
		{`char_is_upper 'Q'`, true},
		{`char_is_upper 'q'`, false},
		{`char_is_lower 'q'`, true},
		{`char_is_lower 'Q'`, false},
		{`'a' == "a"`, true},
		{`'a' < 'b'`, true},
		{`compare 'a' 'b'`, -1},
	} {
		got := evalOne(t, in, tc.src)
		if got != tc.want {
			t.Errorf("%s = %#v (%T), want %#v (%T)", tc.src, got, got, tc.want, tc.want)
		}
	}
}

// char_of_str is Maybe-valued: malformed input is Nothing, never an error.
func TestCharOfStr(t *testing.T) {
	in := NewInterpreter()
	just, ok := evalOne(t, in, `char_of_str "z"`).(functions.Maybe[any])
	if !ok || !just.IsJust() {
		t.Fatalf(`char_of_str "z" must be Just, got %v`, just)
	}
	if v, _ := just.MaybeValue(); v != functions.Char('z') {
		t.Errorf(`char_of_str "z" payload = %#v, want Char z`, v)
	}
	for _, bad := range []string{`""`, `"zz"`} {
		nothing, ok := evalOne(t, in, `char_of_str `+bad).(functions.Maybe[any])
		if !ok || nothing.IsJust() {
			t.Errorf("char_of_str %s must be Nothing, got %v", bad, nothing)
		}
	}
}

// Function form inside a lambda over a list of Chars; the carrier and element type survive.
func TestCharFuncsInsideLambda(t *testing.T) {
	in := NewInterpreter()
	got := in.EvaluateToValues(`['a', 'b'] |> map (\c -> char_upper c)`)
	items, _ := functions.ToSliceAny(got[0])
	if len(items) != 2 || items[0] != functions.Char('A') || items[1] != functions.Char('B') {
		t.Errorf("map char_upper = %#v", items)
	}
}

func TestCharFuncsFailLoud(t *testing.T) {
	in := NewInterpreter()
	for _, src := range []string{`char_code 5`, `char_code "ab"`, `char_code ""`, `char_from_code -1`, `char_from_code 1114112`, `char_from_code "x"`, `char_of_str 5`} {
		res := in.evaluatePipelineStr(src)
		if len(res) != 1 || len(res[0]) < 7 || res[0][:7] != "[ERROR]" {
			t.Errorf("%s = %v, want an [ERROR]", src, res)
		}
	}
}

// A let-bound Char must reach the evaluator as a Char: its text ("x") would re-parse as a
// String (or an undefined name), so it is not a text-substitutable scalar.
func TestLetBoundCharKeepsItsType(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateLine(`let lc = 'x'`)
	if got := evalOne(t, in, `char_code lc`); got != 120 {
		t.Errorf("char_code lc = %#v, want 120", got)
	}
	if got := evalOne(t, in, `lc`); got != functions.Char('x') {
		t.Errorf("lc = %#v (%T), want Char x", got, got)
	}
}
