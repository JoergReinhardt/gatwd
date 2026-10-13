package cli

import (
	"math/big"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// matchQuery defines the match-bodied `def`s of setup (EvaluateScript, the way a script
// does) and then runs one query through EvaluateToValues, returning the raw values (an
// error arrives as an `error` element, never as text).
func matchQuery(t *testing.T, setup, query string) []any {
	t.Helper()
	in := NewInterpreter()
	if setup != "" {
		in.EvaluateScript(setup)
	}
	return in.EvaluateToValues(query)
}

// wantValue asserts a query produced exactly one value, equal (by the language's own ==)
// to want, and not an error.
func wantValue(t *testing.T, got []any, want any) {
	t.Helper()
	if len(got) != 1 {
		t.Fatalf("got %d results %#v, want exactly one (%#v)", len(got), got, want)
	}
	if err, isErr := got[0].(error); isErr {
		t.Fatalf("got error %v, want %#v", err, want)
	}
	if !functions.StructuralEqual(got[0], want) {
		t.Fatalf("got %#v (%T), want %#v (%T)", got[0], got[0], want, want)
	}
}

// wantError asserts a query produced exactly one error value whose text contains every
// fragment, so an arm that matches nothing, or a pattern that means nothing, is loud.
func wantError(t *testing.T, got []any, fragments ...string) {
	t.Helper()
	if len(got) != 1 {
		t.Fatalf("got %d results %#v, want exactly one error", len(got), got)
	}
	err, isErr := got[0].(error)
	if !isErr {
		t.Fatalf("got %#v (%T), want an error containing %q", got[0], got[0], fragments)
	}
	for _, f := range fragments {
		if !strings.Contains(err.Error(), f) {
			t.Fatalf("error %q does not contain %q", err.Error(), f)
		}
	}
}

// --- Pinned behaviour: these arms were right before the fix and must stay right. ---

func TestMatchPinnedConstructorRouting(t *testing.T) {
	setup := "type OS = Linux | Windows | Darwin\n" +
		"def route = match { Linux -> echo PENGUIN | Windows -> echo WINDOW | Darwin -> echo APPLE }"
	for tag, want := range map[string]string{"Linux": "PENGUIN", "Windows": "WINDOW", "Darwin": "APPLE"} {
		wantValue(t, matchQuery(t, setup, "spawn OS: "+tag+" |> route"), want)
	}
}

func TestMatchPinnedHeadTailEmptyAndTuple(t *testing.T) {
	setup := "def sum = match { [] -> 0 | [h, ...t] -> h + (t |> sum) }\n" +
		"def ack = match { [0, n] -> n + 1 | [m, 0] -> [m - 1, 1] |> ack | [m, n] -> let r = [m, n - 1] |> ack in [m - 1, r] |> ack }\n" +
		"def swap = match { [a, b] -> [b, a] }\n" +
		"def kind = match { _ -> 7 }\n" +
		"def ident = match { x -> x + 1 }"
	in := NewInterpreter()
	in.EvaluateScript(setup)
	wantValue(t, in.EvaluateToValues("[2, 2] |> ack"), 7)
	wantValue(t, in.EvaluateToValues("[3, 2] |> ack"), 29)
	if got := in.evaluatePipelineStr("[1, 2] |> swap"); len(got) != 1 || got[0] != "[2 1]" {
		t.Fatalf("[1, 2] |> swap = %v, want [[2 1]]", got)
	}
	wantValue(t, in.EvaluateToValues("5 |> kind"), 7)
	wantValue(t, in.EvaluateToValues("5 |> ident"), 6)
	wantValue(t, in.EvaluateToValues("[] |> sum"), 0)
}

// --- (1) A literal pattern is equality, never substring containment. ---

func TestMatchLiteralPatternIsEquality(t *testing.T) {
	setup := "def one = match { 0 -> 10 | 1 -> 11 | n -> 99 }"
	in := NewInterpreter()
	in.EvaluateScript(setup)
	for _, row := range []struct {
		in   string
		want int
	}{{"0", 10}, {"1", 11}, {"20", 99}, {"11", 99}, {"12", 99}, {"5", 99}, {"101", 99}} {
		wantValue(t, in.EvaluateToValues(row.in+" |> one"), row.want)
	}
}

func TestMatchRecursiveFib(t *testing.T) {
	setup := "def fib = match { 0 -> 0 | 1 -> 1 | n -> let a = (n - 1) |> fib in let b = (n - 2) |> fib in a + b }"
	in := NewInterpreter()
	in.EvaluateScript(setup)
	for n, want := range map[string]int{"0": 0, "1": 1, "2": 1, "9": 34, "10": 55, "11": 89} {
		wantValue(t, in.EvaluateToValues(n+" |> fib"), want)
	}
}

func TestMatchLiteralKinds(t *testing.T) {
	setup := "def f = match { 0 -> 10 | -1 -> 11 | 2.5 -> 12 | \"a\" -> 13 | 'b' -> 14 | true -> 15 | 5s -> 16 | @2026-10-13 -> 17 | _ -> 99 }"
	in := NewInterpreter()
	in.EvaluateScript(setup)
	for _, row := range []struct {
		query string
		want  int
	}{
		{"0 |> f", 10},
		{"0.0 |> f", 10}, // 0.0 == 0, exactly as the language's ==
		{"(0 - 1) |> f", 11},
		{"2.5 |> f", 12},
		{"\"a\" |> f", 13},
		{"'b' |> f", 14},
		{"true |> f", 15},
		{"5s |> f", 16},
		{"5 |> f", 99},
		{"false |> f", 99},
		{"\"b\" |> f", 14}, // a one-character String equals the Char (D1), as in `"b" == 'b'`
		{"\"0\" |> f", 99}, // a String is never a number (D2)
		{"\"xay\" |> f", 99},
		{"@2026-10-13 |> f", 17},
		{"@2026-10-14 |> f", 99},
	} {
		wantValue(t, in.EvaluateToValues(row.query), row.want)
	}
}

func TestMatchStringPatternIsEquality(t *testing.T) {
	setup := "def h = match { \"foo\" -> 1 | _ -> 2 }"
	in := NewInterpreter()
	in.EvaluateScript(setup)
	wantValue(t, in.EvaluateToValues(`"foo" |> h`), 1)
	wantValue(t, in.EvaluateToValues(`"xfoox" |> h`), 2) // no implicit substring matching
	wantValue(t, in.EvaluateToValues(`7 |> h`), 2)       // never substring-match a non-String
}

// --- (2) Literal elements of a fixed-length pattern use the same rule. ---

func TestMatchFixedPatternLiteralElements(t *testing.T) {
	setup := "def f = match { [0, y] -> 100 | [\"a\", y] -> 200 | ['c', y] -> 250 | [a, b] -> 300 }"
	in := NewInterpreter()
	in.EvaluateScript(setup)
	wantValue(t, in.EvaluateToValues("[0, 1] |> f"), 100)
	wantValue(t, in.EvaluateToValues("[0.0, 1] |> f"), 100) // exact across Int/Float
	wantValue(t, in.EvaluateToValues(`["0", 1] |> f`), 300) // a String "0" is not the number 0
	wantValue(t, in.EvaluateToValues(`['0', 1] |> f`), 300) // nor is the Char '0'
	wantValue(t, in.EvaluateToValues(`["a", 1] |> f`), 200)
	wantValue(t, in.EvaluateToValues(`['c', 1] |> f`), 250)
	wantValue(t, in.EvaluateToValues("[10, 1] |> f"), 300) // 10 contains "0" as text, but is not 0
}

func TestMatchHeadElementPattern(t *testing.T) {
	setup := "def f = match { [0, ...t] -> 1 | [h, ...t] -> 2 | [] -> 3 }"
	in := NewInterpreter()
	in.EvaluateScript(setup)
	wantValue(t, in.EvaluateToValues("[0, 5] |> f"), 1)
	wantValue(t, in.EvaluateToValues("[7, 5] |> f"), 2)
	wantValue(t, in.EvaluateToValues("[] |> f"), 3)
}

// --- (3) A value that matches no arm is an error, never silence. ---

func TestMatchNoArmIsAnError(t *testing.T) {
	setup := "def g = match { 1 -> 11 | 2 -> 22 }\n" +
		"def e = match { [] -> 0 }\n" +
		"type Sig = SRed | SGreen | SBlue\n" +
		"def sig = match { SRed -> 1 | SGreen -> 2 }"
	in := NewInterpreter()
	in.EvaluateScript(setup)
	wantError(t, in.EvaluateToValues("3 |> g"), "match: no arm matches 3")
	wantError(t, in.EvaluateToValues(`"x" |> g`), "match: no arm matches x")
	wantError(t, in.EvaluateToValues("[1, 2] |> e"), "match: no arm matches")
	wantError(t, in.EvaluateToValues("spawn Sig: SBlue |> sig"), "match: no arm matches")
	// The Int the arms are written for still routes.
	wantValue(t, in.EvaluateToValues("2 |> g"), 22)
}

func TestMatchNoArmErrorDoesNotStopTheStream(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateScript("type IV = Int\ndef g = match { 1 -> 11 | 2 -> 22 }")
	got := in.EvaluateToValues("spawn IV: 1 |> emit 9, 2 |> g")
	if len(got) != 3 {
		t.Fatalf("got %#v, want three results (value, error, value)", got)
	}
	wantValue(t, got[0:1], 11)
	wantError(t, got[1:2], "match: no arm matches 9")
	wantValue(t, got[2:3], 22)
}

// --- A pattern that means nothing is a loud error naming the pattern. ---

func TestMatchBadPatternIsAnError(t *testing.T) {
	for _, row := range []struct {
		name, def string
		want      []string
	}{
		{"guard", "def k = match { n if n > 0 -> 1 | _ -> 2 }", []string{"guards are not supported yet", "n if n > 0"}},
		{"operator expression", "def k = match { 1 + 2 -> 1 | _ -> 2 }", []string{"invalid pattern", "1 + 2"}},
		{"nested list", "def k = match { [[a, b], c] -> 1 | _ -> 2 }", []string{"nested", "[a, b]"}},
		{"arm without arrow", "def k = match { 1 -> 1 | zzz }", []string{"arm", "zzz"}},
		{"empty tuple", "def k = match { () -> 1 | _ -> 2 }", []string{"invalid pattern", "()"}},
		{"bad tail", "def k = match { [h, ...3] -> 1 | _ -> 2 }", []string{"invalid pattern", "...3"}},
		{"bad element", "def k = match { [a + 1, b] -> 1 | _ -> 2 }", []string{"invalid pattern", "a + 1"}},
	} {
		t.Run(row.name, func(t *testing.T) {
			wantError(t, matchQuery(t, row.def, "5 |> k"), row.want...)
		})
	}
}

// --- The arm compiler, table-driven. ---

func TestCompileArmPattern(t *testing.T) {
	for _, row := range []struct {
		text string
		kind armKind
		ok   bool
	}{
		{"_", armWild, true},
		{"x", armIdent, true},
		{"Red", armIdent, true},
		{"[]", armEmpty, true},
		{"0", armLit, true},
		{"-3", armLit, true},
		{"1.5", armLit, true},
		{`"s"`, armLit, true},
		{`'c'`, armLit, true},
		{"true", armLit, true},
		{"false", armLit, true},
		{"5s", armLit, true},
		{"@2026-10-13", armLit, true},
		{"[a, b]", armFixed, true},
		{"(a, 0)", armFixed, true},
		{`["a,b", _]`, armFixed, true},
		{"[h, ...t]", armHeadTail, true},
		{"[0, ...t]", armHeadTail, true},
		{"n if n > 0", 0, false},
		{"1 + 2", 0, false},
		{"[[a], b]", 0, false},
		{"[a, ...b, c]", 0, false},
		{"[a, b, ...t]", 0, false},
		{"", 0, false},
		{"()", 0, false},
	} {
		p, err := compileArmPattern(row.text)
		if row.ok != (err == nil) {
			t.Fatalf("compileArmPattern(%q) error = %v, want ok=%v", row.text, err, row.ok)
		}
		if err == nil && p.kind != row.kind {
			t.Fatalf("compileArmPattern(%q).kind = %v, want %v", row.text, p.kind, row.kind)
		}
	}
}

func TestPatternLiteralValues(t *testing.T) {
	for _, row := range []struct {
		text string
		want any
	}{
		{"0", 0},
		{"-3", -3},
		{"1.5", 1.5},
		{`"a"`, "a"},
		{`'a'`, functions.Char('a')},
		{"true", true},
		{"false", false},
		{"5s", functions.Duration(5_000_000_000)},
		{"99999999999999999999", func() any { n, _ := new(big.Int).SetString("99999999999999999999", 10); return n }()},
	} {
		got, ok := parsePatternLiteral(row.text)
		if !ok || !functions.StructuralEqual(got, row.want) {
			t.Fatalf("parsePatternLiteral(%q) = %#v, %v; want %#v", row.text, got, ok, row.want)
		}
	}
	for _, text := range []string{"x", "[1]", "1 + 2", "- \"a\"", "n if n"} {
		if v, ok := parsePatternLiteral(text); ok {
			t.Fatalf("parsePatternLiteral(%q) = %#v, want not a literal", text, v)
		}
	}
}

// matchArm, table-driven: which interpreter the body runs in and what is bound.
func TestMatchArmBindings(t *testing.T) {
	in := NewInterpreter()
	flag := in.Parser.GetFlag("Variable")
	list := func(xs ...any) any {
		l := collections.Nil[any]()
		for k := len(xs) - 1; k >= 0; k-- {
			l = collections.Cons[any](xs[k], l)
		}
		return l
	}
	for _, row := range []struct {
		pat   string
		val   any
		match bool
		bound map[string]any
	}{
		{"_", 5, true, nil},
		{"x", 5, true, map[string]any{"x": 5}},
		{"5", 5, true, nil},
		{"5", 6, false, nil},
		{"5", "5", false, nil},
		{"[a, 0]", list(7, 0), true, map[string]any{"a": 7}},
		{"[a, 0]", list(7, 1), false, nil},
		{"[a, 0]", list(7), false, nil},
		{"[a, b]", 5, false, nil},
		{"[_, _]", list(1, 2), true, nil},
		{"[h, ...t]", list(), false, nil},
		{"[]", list(), true, nil},
		{"[]", list(1), false, nil},
	} {
		p, err := compileArmPattern(row.pat)
		if err != nil {
			t.Fatalf("%q: %v", row.pat, err)
		}
		scope, ok := in.matchArm(p, row.val, flag)
		if ok != row.match {
			t.Fatalf("matchArm(%q, %#v) = %v, want %v", row.pat, row.val, ok, row.match)
		}
		for name, want := range row.bound {
			rt, has := scope.RuntimeNodes[name]
			if !has {
				t.Fatalf("matchArm(%q, %#v) did not bind %s", row.pat, row.val, name)
			}
			got, _ := rt.Value.UnwrapRight()
			if !functions.StructuralEqual(got, want) {
				t.Fatalf("%s bound to %#v, want %#v", name, got, want)
			}
		}
	}
}
