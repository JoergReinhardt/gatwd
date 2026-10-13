package cli

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

// D2: `+` is numeric-only. Every row used to coerce or concatenate silently.
func TestD2PlusIsNumericOnly(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct{ src, want string }{
		{`"a" + "b"`, "use ++"},
		{`"a" + 1`, "not defined for String and Int"},
		{`1 + "a"`, "not defined for Int and String"},
		{`"1" + 2`, "not defined for String and Int"},
		{`"1" + "2"`, "str_to_int"},
		{`[1, 2] |> map (\x -> x + "a")`, "not defined for Int and String"},
		{`fold (\a x -> a + x) 0 ["1", "2"]`, "not defined for Int and String"},
	} {
		res := in.evaluatePipelineStr(tc.src)
		if len(res) != 1 || !strings.Contains(res[0], "[ERROR]") || !strings.Contains(res[0], tc.want) {
			t.Errorf("%s = %v, want an [ERROR] containing %q", tc.src, res, tc.want)
		}
	}
	// What still works: numbers, and ++ for text, and the explicit conversion.
	for src, want := range map[string]any{
		`1 + 2`:                                 3,
		`1.5 + 1`:                               2.5,
		`"a" ++ "b"`:                            "ab",
		`"1" ++ "2"`:                            "12",
		`maybe_default 0 (str_to_int "41") + 1`: 42,
	} {
		if got := evalOne(t, in, src); !functions.StructuralEqual(got, want) || functions.KindName(got) != functions.KindName(want) {
			t.Errorf("%s = %#v, want %#v", src, got, want)
		}
	}
}

// The tokens of `emit`, `const` and `-D` are literals: a canonical number is a number, any
// other text is a String. One rule, so numbers from a script's own literals add up and the
// text of a log line never silently does.
func TestLiteralTokensAreTyped(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateScript("type IntVal = Int\nconst N = 5\nconst F = 1.5\nconst NAME = alice\nconst QUOTED = \"5\"\nconst PADDED = 007\nconst SUM = 2 * 3")
	for name, kind := range map[string]string{
		"N": "Int", "F": "Float", "NAME": "String", "QUOTED": "String", "PADDED": "String", "SUM": "Int",
	} {
		if got := functions.KindName(in.Constants[name]); got != kind {
			t.Errorf("const %s is a %s (%#v), want %s", name, got, in.Constants[name], kind)
		}
	}
	for src, want := range map[string]any{
		`N + 1`:                       6,
		`[1, 2] |> map (\x -> x + N)`: nil,
		`SUM * 2`:                     12,
		`spawn IntVal: 0 |> emit 10, 20 |> drop 1 |> fold sum`: 30,
	} {
		got := in.EvaluateToValues(src)
		if len(got) != 1 {
			t.Errorf("%s = %v", src, got)
			continue
		}
		if want == nil {
			if items := itemsOf(t, got[0]); len(items) != 2 || !functions.StructuralEqual(items[0], 6) || !functions.StructuralEqual(items[1], 7) {
				t.Errorf("%s = %#v", src, items)
			}
			continue
		}
		if !functions.StructuralEqual(got[0], want) {
			t.Errorf("%s = %#v, want %#v", src, got[0], want)
		}
	}
	// emit: numbers are numbers, other tokens stay text (and a non-canonical number too)
	vals := in.EvaluateToValues(`spawn S: x |> emit 10, 007, 1.50, alice, -5, 1.5 |> drop 1 |> collect_list`)
	items := itemsOf(t, vals[0])
	var kinds []string
	for _, it := range items {
		kinds = append(kinds, functions.KindName(it))
	}
	if got := strings.Join(kinds, ","); got != "Int,String,String,String,Int,Float" {
		t.Errorf("emit token kinds = %s, want Int,String,String,String,Int,Float", got)
	}

	// -D uses the same rule
	d := NewInterpreter()
	if err := d.setDefines(map[string]string{"LIMIT": "10", "NAME": "bob", "ZIP": "01234"}); err != nil {
		t.Fatal(err)
	}
	if functions.KindName(d.Constants["LIMIT"]) != "Int" || functions.KindName(d.Constants["NAME"]) != "String" || functions.KindName(d.Constants["ZIP"]) != "String" {
		t.Errorf("-D kinds = %#v", d.Constants)
	}
	if got := evalOne(t, d, `[1, 2] |> map (\x -> x + LIMIT)`); len(itemsOf(t, got)) != 2 {
		t.Errorf("a -D number must work inside a lambda, got %#v", got)
	}
}

// Operator sections and math_* steps take a literal token as the operand: it is a number.
func TestSectionOperandsAreNumbers(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateScript("type IntVal = Int")
	for src, want := range map[string]any{
		`spawn IntVal: 5 |> + 1`:        6,
		`spawn IntVal: 5 |> math_add 1`: 6,
		`spawn IntVal: 5 |> - 2`:        3,
		`spawn IntVal: 5 |> * 3`:        15,
		`spawn IntVal: 5 |> / 2`:        2.5,
	} {
		got := in.EvaluateToValues(src)
		if len(got) != 1 || !functions.StructuralEqual(got[0], want) {
			t.Errorf("%s = %v, want %v", src, got, want)
		}
	}
	// A text node reaching a numeric section is a loud error, not a silent parse.
	res := in.evaluatePipelineStr(`spawn StrVal: 5 |> + 1`)
	if len(res) != 1 || !strings.Contains(res[0], "String and Int") {
		t.Errorf("a String node in an arithmetic section must fail loudly, got %v", res)
	}
	// ... and ++ keeps its operand as text
	if got := in.evaluatePipelineStr(`spawn StrVal: a |> ++ 5`); len(got) != 1 || got[0] != "a5" {
		t.Errorf("spawn StrVal: a |> ++ 5 = %v, want [a5]", got)
	}
}

// The checker already asks the runtime's operator matrix, so it warns about string +.
func TestCheckWarnsAboutStringPlus(t *testing.T) {
	ws := warnings("\"a\" + \"b\"\n\"a\" + 1\n[1] + 2")
	if !hasWarning(ws, "operator '+' is not defined for String and String") || !hasWarning(ws, "operator '+' is not defined for String and Int") {
		t.Errorf("--check must report string +, got %v", ws)
	}
}

// A String constant (`const`, `-D`) is spliced into a pipeline head as its quoted literal: a
// bare word would be read as an undefined name and the statement would print nothing.
func TestStringConstantsSpliceAsLiterals(t *testing.T) {
	in := NewInterpreter()
	if err := in.setDefines(map[string]string{"NAME": "bob", "GREETING": "hello world", "ZIP": "01234"}); err != nil {
		t.Fatal(err)
	}
	in.EvaluateScript("const LABEL = \"a b\"\nconst WORD = gatwd")
	for src, want := range map[string]any{
		`str_upper NAME`:     "BOB",
		`length GREETING`:    11,
		`str_upper GREETING`: "HELLO WORLD",
		`length ZIP`:         5,
		`length LABEL`:       3,
		`str_upper WORD`:     "GATWD",
	} {
		got := in.evaluatePipelineStr(src)
		if len(got) != 1 || got[0] != displayText(want) {
			t.Errorf("%s = %v, want [%v]", src, got, want)
		}
	}
}
