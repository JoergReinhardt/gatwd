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

func TestStrLibraryValues(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct {
		src  string
		want any
	}{
		{`str_join "-" ["a", "b", 'c']`, "a-b-c"},
		{`str_join "," []`, ""},
		{`str_from_chars ['a', 'b']`, "ab"},
		{`str_from_chars []`, ""},
		{`str_trim "  x "`, "x"},
		{`str_upper "héllo"`, "HÉLLO"},
		{`str_lower "ÀB"`, "àb"},
		{`str_repeat 3 "ab"`, "ababab"},
		{`str_repeat 0 "ab"`, ""},
		{`str_contains "ell" "hello"`, true},
		{`str_contains "z" "hello"`, false},
		{`str_contains "" "hello"`, true},
		{`str_starts_with "he" "hello"`, true},
		{`str_ends_with "he" "hello"`, false},
		{`str_replace "l" "L" "hello"`, "heLLo"},
		{`str_slice 1 3 "héllo"`, "él"},
		{`str_slice 2 2 "abc"`, ""},
		{`str_pad_left 5 '0' "42"`, "00042"},
		{`str_pad_left 1 '0' "42"`, "42"},
		{`str_pad_right 4 "." "ab"`, "ab.."},
		{`str_from [1, 2]`, "[1 2]"},
		{`str_from 'x'`, "x"},
		{`str_from "x"`, "x"},
		{`show "a"`, `"a"`},
		{`show 'a'`, `'a'`},
		{`show '\n'`, `'\n'`},
		{`show ['a', "b", 1]`, `['a', "b", 1]`},
		{`str_format "%s has %d items (%.2f%%)" ["x", 3, 12.345]`, "x has 3 items (12.35%)"},
		{`str_format "%5d|%-5s|%c|%v" [42, "ab", 'z', [1, 2]]`, "   42|ab   |z|[1 2]"},
		{`str_format "%x" [255]`, "ff"},
		{`str_format "%q" ["a b"]`, `"a b"`},
		{`str_format "100%%" []`, "100%"},
		{`str_format "%s" ['c']`, "c"},
		{`str_to_int "-42"`, "Just(-42)"},
		{`str_to_int "+7"`, "Just(7)"},
		{`str_to_int "99999999999999999999"`, "Just(99999999999999999999)"},
		{`str_to_int ""`, "Nothing"},
		{`str_to_int "4x"`, "Nothing"},
		{`str_to_int " 4"`, "Nothing"},
		{`str_to_int "1.5"`, "Nothing"},
		{`str_to_float "1.5e3"`, "Just(1500)"},
		{`str_to_float ".5"`, "Just(0.5)"},
		{`str_to_float "nan"`, "Nothing"},
		{`str_to_float "inf"`, "Nothing"},
		{`str_to_float "0x1p-2"`, "Nothing"},
		{`str_to_float "1e999"`, "Nothing"},
		{`str_to_float "x"`, "Nothing"},
		{`str_index_of "z" "hello"`, "Nothing"},
		{`str_index_of "l" "héllo"`, "Just(2)"},
		{`str_index_of "" "abc"`, "Just(0)"},
	} {
		got := evalOne(t, in, tc.src)
		if s, isStr := tc.want.(string); isStr && strings.HasPrefix(s, "Just(") || tc.want == "Nothing" {
			got = displayText(got) // a Maybe prints as Just(x) / Nothing
		}
		if got != tc.want {
			t.Errorf("%s = %#v, want %#v", tc.src, got, tc.want)
		}
	}
}

func TestStrLibraryStructural(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct {
		src  string
		want []any
	}{
		{`str_split "," "a,b,,c"`, []any{"a", "b", "", "c"}},
		{`str_split "," ""`, []any{""}},
		{`str_words "  hi   there "`, []any{"hi", "there"}},
		{`str_chars "hé"`, []any{functions.Char('h'), functions.Char('é')}},
		{`str_chars ""`, nil},
		{`str_lines (str_join (char_to_str '\n') ["a", "b"])`, []any{"a", "b"}},
		{`str_lines (str_join (char_to_str '\n') ["a", "b", ""])`, []any{"a", "b"}},
		{`str_lines (str_join (char_to_str '\n') ["a", ""])`, []any{"a"}},
		{`str_lines ""`, nil},
	} {
		got := itemsOf(t, evalOne(t, in, tc.src))
		if len(got) != len(tc.want) {
			t.Errorf("%s = %#v, want %#v", tc.src, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("%s[%d] = %#v, want %#v", tc.src, i, got[i], tc.want[i])
			}
		}
	}
	for _, src := range []string{`str_to_int "5"`, `str_index_of "a" "a"`} {
		if mb, ok := evalOne(t, in, src).(functions.Maybe[any]); !ok || !mb.IsJust() {
			t.Errorf("%s must be a Just Maybe", src)
		}
	}
	if mb, ok := evalOne(t, in, `str_to_int "5"`).(functions.Maybe[any]); ok {
		if v, _ := mb.MaybeValue(); !functions.StructuralEqual(v, 5) || functions.IsChar(v) {
			t.Errorf(`str_to_int "5" payload = %#v (%T), want the narrowed Int 5`, v, v)
		}
	}
}

// A collection of Chars is accepted wherever a String is demanded (the D1 table).
func TestStrLibraryAcceptsCharCollections(t *testing.T) {
	in := NewInterpreter()
	if got := evalOne(t, in, `str_upper (map (\c -> char_lower c) "ABC")`); got != "ABC" {
		t.Errorf("str_upper over a List of Chars = %#v, want ABC", got)
	}
	if got := evalOne(t, in, `str_contains ['b', 'c'] ['a', 'b', 'c']`); got != true {
		t.Errorf("str_contains over Char collections = %#v", got)
	}
}

func TestStrLibraryFailsLoud(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct{ src, want string }{
		{`str_split "" "abc"`, "empty separator"},
		{`str_split "," 5`, "want a String"},
		{`str_upper 5`, "want a String"},
		{`str_upper ['a', 1]`, "element 1"},
		{`str_join "," [1, 2]`, "element 0"},
		{`str_join "," 5`, "collection"},
		{`str_replace "" "x" "abc"`, "empty pattern"},
		{`str_slice 2 1 "abc"`, "out of range"},
		{`str_slice 0 9 "abc"`, "out of range"},
		{`str_slice (0 - 1) 2 "abc"`, "out of range"},
		{`str_slice "0" 1 "abc"`, "want an Int"},
		{`str_repeat (0 - 1) "a"`, "negative"},
		{`str_repeat 99999999 "abcdefghij"`, "longer than"},
		{`str_repeat "3" "a"`, "want an Int"},
		{`str_pad_left 3 "ab" "x"`, "one-character"},
		{`str_format "%d" ["x"]`, "wants an Int"},
		{`str_format "%s" [1]`, "wants a String"},
		{`str_format "%f" ["x"]`, "wants a number"},
		{`str_format "%s" []`, "needs an argument"},
		{`str_format "%s" ["a", "b"]`, "argument(s) given"},
		{`str_format "%z" [1]`, "unsupported verb"},
		{`str_format "abc%" []`, "ends inside"},
		{`str_format "%s" 5`, "must be a list"},
	} {
		res := in.evaluatePipelineStr(tc.src)
		if len(res) != 1 || !strings.Contains(res[0], "[ERROR]") || !strings.Contains(res[0], tc.want) {
			t.Errorf("%s = %v, want an [ERROR] containing %q", tc.src, res, tc.want)
		}
	}
}

// Every str_* function form is also a step (F18), applied data-last to the whole String.
func TestStrLibraryAsSteps(t *testing.T) {
	in := NewInterpreter()
	for _, c := range [][2]string{
		{`"a,b" |> str_split ","`, `str_split "," "a,b"`},
		{`"hello" |> str_upper`, `str_upper "hello"`},
		{`"  x " |> str_trim |> str_upper`, `str_upper (str_trim "  x ")`},
		{`["a", "b"] |> str_join "+"`, `str_join "+" ["a", "b"]`},
		{`"42" |> str_to_int`, `str_to_int "42"`},
		{`"hello" |> str_slice 1 3`, `str_slice 1 3 "hello"`},
		{`"hello" |> reverse`, `reverse "hello"`},
	} {
		step := strings.Join(in.evaluatePipelineStr(c[0]), ";")
		fn := strings.Join(in.evaluatePipelineStr(c[1]), ";")
		if step == "" || step != fn || strings.Contains(step, "ERROR") {
			t.Errorf("%s = %q but %s = %q", c[0], step, c[1], fn)
		}
	}
}
