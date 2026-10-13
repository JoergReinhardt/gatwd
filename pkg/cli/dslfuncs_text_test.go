package cli

import (
	"strings"
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

func itemsOf(t *testing.T, v any) []any {
	t.Helper()
	items, ok := functions.ToSliceAny(v)
	if !ok {
		t.Fatalf("%#v is not a sequence", v)
	}
	return items
}

// String is a collection row of Chars (D1): the closed operations (filter, reject, reverse,
// tail, cons) answer with a String, map and flatmap with a List of whatever f returned.
func TestStringIsACollectionRow(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct {
		src  string
		want any
	}{
		{`length "héllo"`, 5},
		{`reverse "abc"`, "cba"},
		{`filter (\c -> c != 'b') "abc"`, "ac"},
		{`reject (\c -> c == 'b') "abc"`, "ac"},
		{`filter (\c -> c == 'z') "abc"`, ""},
		{`tail "abc"`, "bc"},
		{`tail "a"`, ""},
		{`cons 'x' "abc"`, "xabc"},
		{`cons "x" "abc"`, "xabc"},
		{`head "abc"`, functions.Char('a')},
		{`nth 1 "abc"`, functions.Char('b')},
		{`member 'b' "abc"`, true},
		{`member "b" "abc"`, true},
		{`member 'z' "abc"`, false},
		{`is_empty ""`, true},
		{`is_empty "a"`, false},
		{`fold (\a c -> a + (char_code c)) 0 "ab"`, 195},
		{`"ab" ++ ['c', 'd']`, "abcd"},
		{`['a'] ++ "bc"`, "abc"},
		{`"ab" ++ []`, "ab"},
	} {
		if got := evalOne(t, in, tc.src); got != tc.want {
			t.Errorf("%s = %#v (%T), want %#v (%T)", tc.src, got, got, tc.want, tc.want)
		}
	}
	mapped := itemsOf(t, evalOne(t, in, `map (\c -> char_upper c) "abc"`))
	if len(mapped) != 3 || mapped[0] != functions.Char('A') || mapped[2] != functions.Char('C') {
		t.Errorf("map over a String = %#v, want a list of Chars", mapped)
	}
	codes := itemsOf(t, evalOne(t, in, `map (\c -> char_code c) "ab"`))
	if len(codes) != 2 || codes[0] != 97 {
		t.Errorf("map char_code over a String = %#v", codes)
	}
	fm := itemsOf(t, evalOne(t, in, `flatmap (\c -> [c, c]) "ab"`))
	if len(fm) != 4 {
		t.Errorf("flatmap over a String = %#v", fm)
	}
	if got := itemsOf(t, evalOne(t, in, `to_list "ab"`)); len(got) != 2 || got[0] != functions.Char('a') {
		t.Errorf("to_list \"ab\" = %#v", got)
	}
}

func TestStringRowFailsLoud(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct{ src, want string }{
		{`head ""`, "empty"},
		{`tail ""`, "empty"},
		{`nth 3 "abc"`, "out of range"},
		{`cons 5 "abc"`, "cannot cons onto a String"},
		{`cons "xy" "abc"`, "one-character"},
		{`map (\c -> nosuch c) "ab"`, "nosuch"},
		{`filter (\c -> 1) "ab"`, "want bool"},
	} {
		res := in.evaluatePipelineStr(tc.src)
		if len(res) != 1 || !strings.Contains(res[0], "[ERROR]") || !strings.Contains(res[0], tc.want) {
			t.Errorf("%s = %v, want an [ERROR] containing %q", tc.src, res, tc.want)
		}
	}
}

// asString is the collection-of-Chars → String half of the conversion table: an empty
// collection spells "", a mixed collection is an error naming the element.
func TestAsString(t *testing.T) {
	if s, err := asString(listOf([]any{functions.Char('h'), functions.Char('i')})); err != nil || s != "hi" {
		t.Errorf("asString(list of Chars) = %q, %v", s, err)
	}
	if s, err := asString(listOf(nil)); err != nil || s != "" {
		t.Errorf("asString(empty list) = %q, %v", s, err)
	}
	if s, err := asString("x"); err != nil || s != "x" {
		t.Errorf("asString(String) = %q, %v", s, err)
	}
	_, err := asString(listOf([]any{functions.Char('a'), 1}))
	if err == nil || !strings.Contains(err.Error(), "element 1") {
		t.Errorf("a mixed collection must name the offending element, got %v", err)
	}
	if _, err := asString(5); err == nil {
		t.Error("asString(5) must fail")
	}
}

// In a pipeline a String node is ONE element of the stream (a line, a word); the stream
// steps do not descend into it. The String row belongs to the expression function forms.
func TestStreamStepsKeepStringNodesScalar(t *testing.T) {
	in := NewInterpreter()
	got := in.evaluatePipelineStr(`spawn S: x |> emit ab, cd |> drop 1 |> map (\s -> s ++ "!")`)
	if strings.Join(got, ",") != "ab!,cd!" {
		t.Errorf("map over a stream of Strings = %v, want [ab! cd!]", got)
	}
	got = in.evaluatePipelineStr(`spawn S: x |> emit ab, cd |> drop 1 |> filter (\s -> s == "ab")`)
	if strings.Join(got, ",") != "ab" {
		t.Errorf("filter over a stream of Strings = %v, want [ab]", got)
	}
}
