package cli

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/dsltypes"
)

func TestBoolFunctions(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct {
		src  string
		want any
	}{
		{`not true`, false},
		{`not false`, true},
		{`and true true`, true},
		{`and true false`, false},
		{`or false false`, false},
		{`or false true`, true},
		{`xor true false`, true},
		{`xor true true`, false},
		{`xor false false`, false},
		{`[true, true, false] |> fold and true`, false},
		{`[false, false, true] |> fold or false`, true},
		{`fold and true [true, true]`, true},
		{`fold or false []`, false},
		{`true |> not`, false},
		{`false |> or true`, true},
		{`true == true`, true},
		{`true != false`, true},
		{`true == 1`, false},
		{`map (\b -> not b) [true, false]`, nil},
	} {
		got := evalOne(t, in, tc.src)
		if tc.want == nil {
			items := itemsOf(t, got)
			if len(items) != 2 || items[0] != false || items[1] != true {
				t.Errorf("%s = %#v", tc.src, items)
			}
			continue
		}
		if got != tc.want {
			t.Errorf("%s = %#v, want %#v", tc.src, got, tc.want)
		}
	}
}

// A number, a string, … is never truthy: the Bool functions fail loudly.
func TestBoolFunctionsFailLoud(t *testing.T) {
	in := NewInterpreter()
	for _, src := range []string{`not 0`, `not "x"`, `and 1 true`, `or true "x"`, `xor [] true`} {
		res := in.evaluatePipelineStr(src)
		if len(res) != 1 || !strings.Contains(res[0], "[ERROR]") || !strings.Contains(res[0], "want a Bool") {
			t.Errorf("%s = %v, want an [ERROR] naming Bool", src, res)
		}
	}
}

// compare answers with the Ordering sum type, as a real Variant.
func TestCompareReturnsOrdering(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct{ src, tag string }{
		{`compare 1 2`, "LT"},
		{`compare 2 2`, "EQ"},
		{`compare 3 2`, "GT"},
		{`compare "b" "a"`, "GT"},
		{`compare 'a' 'b'`, "LT"},
		{`compare [1, 2] [1, 3]`, "LT"},
	} {
		v, ok := evalOne(t, in, tc.src).(dsltypes.Variant)
		if !ok || v.Tag != tc.tag {
			t.Errorf("%s = %#v, want the Variant %s", tc.src, v, tc.tag)
		}
	}
	// The constants are the very same values compare produces.
	for _, src := range []string{`compare 1 2 == LT`, `compare 2 2 == EQ`, `compare 3 2 == GT`, `compare 1 2 != GT`} {
		if got := evalOne(t, in, src); got != true {
			t.Errorf("%s = %#v, want true", src, got)
		}
	}
	if got := evalOne(t, in, `LT == GT`); got != false {
		t.Errorf("LT == GT = %#v", got)
	}
	if res := in.evaluatePipelineStr(`compare (pair 1 2) 3`); len(res) != 1 || !strings.Contains(res[0], "incomparable") {
		t.Errorf("an incomparable pair must still be an error, got %v", res)
	}
}

func TestOrderingFunctions(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct {
		src  string
		want any
	}{
		{`ordering_to_int LT`, -1},
		{`ordering_to_int EQ`, 0},
		{`ordering_to_int (compare 5 1)`, 1},
		{`ordering_to_int (ordering_reverse LT)`, 1},
		{`ordering_to_int (ordering_reverse EQ)`, 0},
		{`ordering_to_int (ordering_from_int 9)`, 1},
		{`ordering_to_int (ordering_from_int 0)`, 0},
		{`ordering_to_int (ordering_then EQ GT)`, 1},
		{`ordering_to_int (ordering_then LT GT)`, -1},
		{`ordering_to_int (ordering_then GT LT)`, 1},
		{`ordering_to_int (ordering_then EQ EQ)`, 0},
	} {
		if got := evalOne(t, in, tc.src); got != tc.want {
			t.Errorf("%s = %#v, want %#v", tc.src, got, tc.want)
		}
	}
	for _, src := range []string{`ordering_to_int 5`, `ordering_reverse "x"`, `ordering_then LT 1`, `ordering_from_int "a"`} {
		res := in.evaluatePipelineStr(src)
		if len(res) != 1 || !strings.Contains(res[0], "[ERROR]") {
			t.Errorf("%s = %v, want an [ERROR]", src, res)
		}
	}
}

// Ordering is an ordinary sum type: match routes on it and --check knows its branches.
func TestOrderingMatchesAndIsChecked(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateScript(`def sign_word = match { GT -> echo bigger | LT -> echo smaller | EQ -> echo same }`)
	for src, want := range map[string]string{
		`compare 3 1 |> sign_word`: "bigger",
		`compare 1 3 |> sign_word`: "smaller",
		`compare 2 2 |> sign_word`: "same",
	} {
		if got := in.evaluatePipelineStr(src); len(got) != 1 || got[0] != want {
			t.Errorf("%s = %v, want [%s]", src, got, want)
		}
	}
	ws := warnings("def partial = match { GT -> echo a | LT -> echo b }")
	if !hasWarning(ws, "non-exhaustive match on Ordering: missing EQ") {
		t.Errorf("--check must report the missing EQ branch, got %v", ws)
	}
}
