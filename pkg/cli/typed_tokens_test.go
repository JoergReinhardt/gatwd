package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

// The unquoted tokens of `const`, `emit` and `-D` are typed by one rule (T1): a canonical
// number, Duration, Time or Date literal is that value; anything else, including a
// non-canonical spelling of one (`007s`, `1.50h`, `+5s`), a quoted `const` value and a
// token the lexer would not read as a literal, is a String.
func TestTemporalLiteralTokensAreTyped(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateScript(strings.Join([]string{
		"const AGE = 30d",
		"const HOUR = 1h30m",
		"const DAY = @2026-10-13",
		"const WHEN = @2026-10-13T10:00:00+02:00",
		"const SUM = 1h + 30m",
		"const QDUR = \"30d\"",
		"const QDAY = \"@2026-10-13\"",
		"const QSUM = \"1h + 30m\"",
		"const PAD = 007s",
		"const FRAC = 1.50h",
		"const PLUS = +5s",
		"const ZONE0 = @2026-10-13T10:00:00+00:00",
		"const BADDAY = @2026-02-30",
	}, "\n"))
	for name, kind := range map[string]string{
		"AGE": "Duration", "HOUR": "Duration", "DAY": "Date", "WHEN": "Time", "SUM": "Duration",
		"QDUR": "String", "QDAY": "String", "QSUM": "String",
		"PAD": "String", "FRAC": "String", "BADDAY": "String",
		// a const value with an operator character is an expression (`const P = +5` is 5): it is evaluated
		"PLUS": "Duration", "ZONE0": "Time",
	} {
		if got := functions.KindName(in.Constants[name]); got != kind {
			t.Errorf("const %s is a %s (%#v), want %s", name, got, in.Constants[name], kind)
		}
	}
	// a quoted value is exactly the text written, `@` and all
	if got := in.Constants["QDAY"]; got != "@2026-10-13" {
		t.Errorf(`const QDAY = "@2026-10-13" is %#v`, got)
	}
	if got := in.Constants["WHEN"].(functions.Time).String(); got != "2026-10-13T10:00:00+02:00" {
		t.Errorf("the zone of a Time constant was lost: %s", got)
	}

	// used inside a lambda: the value, not text
	for src, want := range map[string]any{
		`head (map (\x -> x + AGE) [1d, 2d])`:     functions.LiteralToken("744h"),
		`DAY |> map (\d -> d + 1d)`:               functions.LiteralToken("@2026-10-14"),
		`WHEN |> map (\t -> t + 1h)`:              functions.LiteralToken("@2026-10-13T11:00:00+02:00"),
		`head (filter (\x -> x > AGE) [5d, 40d])`: functions.LiteralToken("960h"),
		`AGE + 1d`:    functions.LiteralToken("744h"),
		`DAY + 1d`:    functions.LiteralToken("@2026-10-14"),
		`WHEN - 1h`:   functions.LiteralToken("@2026-10-13T09:00:00+02:00"),
		`SUM > AGE`:   false,
		`AGE > SUM`:   true,
		`QDUR ++ "!"`: "30d!",
	} {
		vals := in.EvaluateToValues(src)
		if len(vals) != 1 {
			t.Errorf("%s = %v", src, vals)
			continue
		}
		got := vals[0]
		if items, ok := functions.ToSliceAny(got); ok && len(items) == 1 {
			got = items[0]
		}
		if functions.KindName(got) != functions.KindName(want) || !functions.StructuralEqual(got, want) {
			t.Errorf("%s = %#v (%s), want %#v (%s)", src, got, functions.KindName(got), want, functions.KindName(want))
		}
	}
	// a text constant still refuses to be added to a Duration (the failure the typing fixes)
	if vals := in.EvaluateToValues(`[1d] |> map (\x -> x + QDUR)`); len(vals) != 1 || !strings.Contains(fmt.Sprint(vals[0]), "String") {
		t.Errorf("a quoted constant must stay text, got %v", vals)
	}
}

// The statement path (a script line) splices a constant into a lambda step as text; a quoted
// String constant must be spliced quoted, or `30d` would lex as a Duration inside the lambda.
func TestQuotedConstantStaysStringInStatementLambda(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateScript("const QDUR = \"30d\"\nconst QNUM = \"5\"\nconst NAME = alice\nconst AGE = 30d")
	for src, want := range map[string]string{
		`[1d] |> map (\n -> QDUR == "30d")`: "[true]",
		`[1d] |> map (\n -> QNUM == "5")`:   "[true]",
		`[1d] |> map (\n -> NAME)`:          "[alice]",
		`[1d] |> map (\n -> n < AGE)`:       "[true]",
	} {
		got := in.evaluatePipelineStr(src)
		if len(got) != 1 || got[0] != want {
			t.Errorf("%s = %v, want %s", src, got, want)
		}
	}
	if got := in.evaluatePipelineStr(`[1d] |> map (\n -> n + QDUR)`); len(got) != 1 || !strings.Contains(got[0], "Duration and String") {
		t.Errorf("a quoted constant is text and does not add to a Duration, got %v", got)
	}
}

// emit and -D read the same tokens.
func TestTemporalLiteralTokensEmitAndDefine(t *testing.T) {
	in := NewInterpreter()
	vals := in.EvaluateToValues(`spawn S: x |> emit 5s, 1h30m, 30d, @2026-10-13, @2026-10-13T10:00:00Z, 007s, 1.50h, +5s, 5q, 10 |> drop 1 |> collect_list`)
	var kinds []string
	for _, it := range itemsOf(t, vals[0]) {
		kinds = append(kinds, functions.KindName(it))
	}
	if got, want := strings.Join(kinds, ","), "Duration,Duration,Duration,Date,Time,String,String,String,String,Int"; got != want {
		t.Errorf("emit token kinds = %s, want %s", got, want)
	}
	if got := in.EvaluateToValues(`spawn S: x |> emit 5s, 6s |> drop 1 |> fold (\a b -> a + b) 0s`); len(got) != 1 || got[0] != functions.Duration(11e9) {
		t.Errorf("emit 5s, 6s |> fold = %#v, want 11s", got)
	}

	d := NewInterpreter()
	if err := d.setDefines(map[string]string{"X": "5s", "AGE": "30d", "DAY": "@2026-10-13", "T": "@2026-10-13T10:00:00Z", "D": "@2026-10-13", "PAD": "007s", "NAME": "bob"}); err != nil {
		t.Fatal(err)
	}
	for name, kind := range map[string]string{"X": "Duration", "AGE": "Duration", "DAY": "Date", "T": "Time", "PAD": "String", "NAME": "String"} {
		if got := functions.KindName(d.Constants[name]); got != kind {
			t.Errorf("-D %s is a %s (%#v), want %s", name, got, d.Constants[name], kind)
		}
	}
	got := evalOne(t, d, `[1s, 2s] |> map (\x -> x + X)`)
	if items := itemsOf(t, got); len(items) != 2 || items[0] != functions.Duration(6e9) || items[1] != functions.Duration(7e9) {
		t.Errorf("a -D duration must work inside a lambda, got %#v", items)
	}
	// operator sections take the same literals as operands, from text or from a constant
	for src, want := range map[string]string{
		`spawn S: x |> emit 5s |> drop 1 |> + 5s`:          "10s",
		`spawn S: x |> emit 5s |> drop 1 |> + X`:           "10s",
		`spawn S: x |> emit @2026-10-13 |> drop 1 |> + 7d`: "2026-10-20",
		`spawn S: x |> emit 1d |> drop 1 |> + AGE`:         "744h",
		`spawn S: x |> emit @2026-10-20 |> drop 1 |> - D`:  "168h",
	} {
		vals := d.EvaluateToValues(src)
		if len(vals) != 1 || functions.KindName(vals[0]) == "String" || fmt.Sprint(vals[0]) != want {
			t.Errorf("%s = %#v, want %s", src, vals, want)
		}
	}
}

// The advisory checker types a `const` literal the way the interpreter does: a bare
// Duration is a Duration (so `AGE + 5` is a finding), a quoted one is a String.
func TestCheckerTypesTemporalConstants(t *testing.T) {
	var msgs []string
	for _, d := range CheckScript("const AGE = 30d\nconst Q = \"30d\"\nlet a = AGE + 5\nlet b = Q + 1d\nlet c = AGE + 1d\n") {
		msgs = append(msgs, d.Msg)
	}
	got := strings.Join(msgs, "\n")
	for _, want := range []string{"not defined for Duration and Int", "not defined for String and Duration"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing finding %q in:\n%s", want, got)
		}
	}
	if len(msgs) != 2 {
		t.Errorf("want exactly two findings (AGE + 1d is fine), got:\n%s", got)
	}
}

// A quoted `const` value is a String in every position, whatever its text would evaluate to
// (`const Q = "2+3"` is the text 2+3, never 5). The unquoted forms keep the canonical-literal
// typing: `const N = 2+3` is an expression and is 5.
func TestQuotedConstantIsAlwaysString(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateScript(strings.Join([]string{
		`const Q = "2+3"`, `const R = "5"`, `const M = "-5"`, `const T = "1 < 2"`, `const E = ""`, `const D = "30d"`,
		`const N = 2+3`, `const K = 5`, `const B = 007`, `const J = "a" ++ "b"`,
	}, "\n"))
	// the stored value, by raw Go type
	for name, want := range map[string]string{"Q": "2+3", "R": "5", "M": "-5", "T": "1 < 2", "E": "", "D": "30d"} {
		got, ok := in.Constants[name].(string)
		if !ok || got != want {
			t.Errorf("const %s = %#v, want the string %q", name, in.Constants[name], want)
		}
	}
	// the unquoted cases are pinned as they were: an expression is evaluated, a canonical literal typed
	if got := in.Constants["N"]; functions.KindName(got) != "Int" || got != 5 {
		t.Errorf("const N = 2+3 is %#v, want the Int 5", got)
	}
	if got := in.Constants["K"]; functions.KindName(got) != "Int" {
		t.Errorf("const K = 5 is %#v, want an Int", got)
	}
	if got, ok := in.Constants["B"].(string); !ok || got != "007" {
		t.Errorf("const B = 007 is %#v, want the String 007", in.Constants["B"])
	}
	// a quoted concatenation expression is not one literal: evaluated, as before
	if got, ok := in.Constants["J"].(string); !ok || got != "ab" {
		t.Errorf(`const J = "a" ++ "b" is %#v, want ab`, in.Constants["J"])
	}

	// every position: bare expression, pipeline head, lambda body, step argument, emit, spawn
	for src, want := range map[string]string{
		`Q`:                                   "2+3",
		`R`:                                   "5",
		`Q ++ "!"`:                            "2+3!",
		`R ++ "!"`:                            "5!",
		`Q |> length`:                         "3",
		`R |> ++ "!"`:                         "5!",
		`Q |> ++ R`:                           "2+35",
		`[1, 2] |> map (\x -> Q)`:             "[2+3 2+3]",
		`[1, 2] |> map (\x -> R == "5")`:      "[true true]",
		`[1, 2] |> map (\x -> M == "-5")`:     "[true true]",
		`[1, 2] |> map (\x -> length Q)`:      "[3 3]",
		`[1, 2] |> map (\x -> T == "1 < 2")`:  "[true true]",
		`[1, 2] |> filter (\x -> Q == "2+3")`: "[1 2]",
		`[1, 2] |> map (\x -> N + x)`:         "[6 7]",
		`spawn S: Q |> collect_list`:          "[2+3]",
		`spawn S: x |> emit "2+3", "5", 5 |> drop 1 |> collect_list |> map (\v -> v ++ "!")`: "[2+3! 5! 5!]",
		`spawn S: x |> emit "2+3", "5", 5 |> drop 1 |> collect_list |> map (\v -> v == "5")`: "[false true false]",
		`N`: "5",
	} {
		got := in.evaluatePipelineStr(src)
		if len(got) != 1 || got[0] != want {
			t.Errorf("%s = %q, want %s", src, got, want)
		}
	}
	// raw type on the value path: the String, not a number
	for _, src := range []string{`Q`, `R`, `M`, `[1] |> map (\x -> R)`, `let v = Q in v`} {
		vals := in.EvaluateToValues(src)
		if len(vals) != 1 {
			t.Errorf("%s = %v", src, vals)
			continue
		}
		got := vals[0]
		if items, ok := functions.ToSliceAny(got); ok && len(items) == 1 {
			got = items[0]
		}
		if _, ok := got.(string); !ok {
			t.Errorf("%s is %#v (%s), want a string", src, got, functions.KindName(got))
		}
	}
	// the String does not do arithmetic (Mandate 6: loud, not coerced)
	if got := in.evaluatePipelineStr(`[1] |> map (\x -> x + R)`); len(got) != 1 || !strings.Contains(got[0], "String") {
		t.Errorf("an Int plus a quoted constant must be an error, got %v", got)
	}
	// `-D`: the value as the shell delivers it. Quotes that reach the program (`-D 'Q="5"'`) make a String;
	// without them the canonical-literal rule applies, so `-D Q=5` is an Int and `-D Q=2+3` the text 2+3.
	d := NewInterpreter()
	if err := d.setDefines(map[string]string{"A": `"5"`, "B": "5", "C": "2+3", "E": `"2+3"`, "F": `"30d"`}); err != nil {
		t.Fatal(err)
	}
	for name, kind := range map[string]string{"A": "String", "B": "Int", "C": "String", "E": "String", "F": "String"} {
		if got := functions.KindName(d.Constants[name]); got != kind {
			t.Errorf("-D %s is a %s (%#v), want %s", name, got, d.Constants[name], kind)
		}
	}
	if d.Constants["A"] != "5" || d.Constants["E"] != "2+3" || d.Constants["F"] != "30d" {
		t.Errorf("-D quotes must be removed, got %#v %#v %#v", d.Constants["A"], d.Constants["E"], d.Constants["F"])
	}
	for src, want := range map[string]string{
		`[1] |> map (\x -> A == "5")`: "[true]", `[1] |> map (\x -> B + x)`: "[6]", `[1] |> map (\x -> E ++ "!")`: "[2+3!]",
	} {
		if got := d.evaluatePipelineStr(src); len(got) != 1 || got[0] != want {
			t.Errorf("%s = %q, want %s", src, got, want)
		}
	}
}
