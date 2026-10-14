package cli_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/pkg/cli"
)

func TestErgonomicSteps(t *testing.T) {
	interp := cli.NewInterpreter()

	// 1. Test MapStep
	interp.RegisterStep(cli.MapStep("shout", func(val any) any {
		return strings.ToUpper(fmt.Sprint(val))
	}))

	// 2. Test FilterStep
	interp.RegisterStep(cli.FilterStep("non_empty", func(val any) bool {
		return strings.TrimSpace(fmt.Sprint(val)) != ""
	}))

	// 3. Test ArgStep
	interp.RegisterStep(cli.ArgStep("prefix", func(arg string, val any) (any, error) {
		return fmt.Sprintf("%s:%v", arg, val), nil
	}))

	// 4. Test SinkStep
	var captured []any
	interp.RegisterStep(cli.SinkStep("record", func(val any) {
		captured = append(captured, val)
	}))

	interp.EvaluateLine("type StrVal = String")
	res := interp.EvaluateToValues("spawn StrVal: hello |> shout |> prefix GATWD |> record |> assert GATWD:HELLO")
	if len(res) == 0 || res[0] != "GATWD:HELLO" {
		t.Fatalf("expected pipeline to yield 'GATWD:HELLO', got %v", res)
	}

	if len(captured) == 0 || captured[0] != "GATWD:HELLO" {
		t.Fatalf("expected SinkStep to capture 'GATWD:HELLO', got %v", captured)
	}
}

func TestEditionBuilder(t *testing.T) {
	edition := cli.NewEdition("gatwd-test").
		WithType("Word", func(s string) any { return s }).
		WithConstant("APP_ENV", "staging").
		WithStep(cli.MapStep("double_str", func(val any) any {
			return fmt.Sprintf("%v%v", val, val)
		}))

	interp := edition.BuildInterpreter()

	env, ok := interp.Constants["APP_ENV"]
	if !ok || env != "staging" {
		t.Fatalf("expected constant APP_ENV to be 'staging', got %v", env)
	}

	res := interp.EvaluateToValues("spawn Word: ab |> double_str |> assert abab")
	if len(res) == 0 || res[0] != "abab" {
		t.Fatalf("expected double_str to yield 'abab', got %v", res)
	}
}

func TestFirstClassAppliance(t *testing.T) {
	tmpDir := t.TempDir()
	overrideFile := filepath.Join(tmpDir, "override.gatwd")

	defaultScript := `
type Status = OK | FAIL
const RETRY_LIMIT = 3
def check = assert OK
`
	app, err := cli.NewAppliance(
		cli.WithEmbeddedScript(defaultScript),
		cli.WithConfigMapOverride(overrideFile),
		cli.WithStep(cli.MapStep("tag_prod", func(val any) any {
			return fmt.Sprintf("prod_%v", val)
		})),
	)
	if err != nil {
		t.Fatalf("failed to create appliance: %v", err)
	}

	limit, ok := app.GetConstant("RETRY_LIMIT")
	if !ok || !functions.StructuralEqual(limit, 3) {
		t.Fatalf("expected embedded default RETRY_LIMIT '3', got %v", limit)
	}

	res := app.Execute("spawn Status: OK |> check |> tag_prod |> assert prod_OK")
	if len(res) == 0 || res[0] != "prod_OK" {
		t.Fatalf("expected result 'prod_OK', got %v", res)
	}

	// Write override to simulate live ConfigMap update
	updatedScript := `
type Status = OK | FAIL
const RETRY_LIMIT = 99
def check = assert FAIL
`
	if err := os.WriteFile(overrideFile, []byte(updatedScript), 0644); err != nil {
		t.Fatalf("failed to write override file: %v", err)
	}

	if err := app.Reload(); err != nil {
		t.Fatalf("failed to reload appliance: %v", err)
	}

	limitUpdated, ok := app.GetConstant("RETRY_LIMIT")
	if !ok || !functions.StructuralEqual(limitUpdated, 99) {
		t.Fatalf("expected updated RETRY_LIMIT '99', got %v", limitUpdated)
	}
}

// F9e: a match def with constructor arms, called from a lambda body, used to
// panic on a nil Parser in the lambda's child interpreter.
func TestMatchDefFromLambdaBodyKeepsParser(t *testing.T) {
	interp := cli.NewInterpreter()
	interp.EvaluateScript("type F9eColor = F9eRed | F9eGreen\ndef f9e_one = match { F9eRed -> 1 | _ -> 2 }")
	res := interp.EvaluateToValues("[F9eRed, F9eGreen] |> map (\\n -> n |> f9e_one)")
	if len(res) != 1 {
		t.Fatalf("want one result, got %v", res)
	}
	if got := fmt.Sprint(res[0]); got != "[1 2]" {
		t.Fatalf("got %s", got)
	}
}

// F9d: a pattern or parameter list before `=` in a def is a loud error, never a
// silently defined-as-nothing function.
func TestDefWithPatternHeadIsLoudError(t *testing.T) {
	for _, tc := range []struct{ line, want string }{
		{"def 1x = 1", `cannot define "1x"`},
		{"def (a, b) = 1", `cannot define "(a, b)"`},
		{"def broken", "needs `=`"},
	} {
		var buf strings.Builder
		interp := cli.NewInterpreter()
		interp.Out = &buf
		interp.EvaluateScript(tc.line)
		if !strings.Contains(buf.String(), "[ERROR]") || !strings.Contains(buf.String(), tc.want) {
			t.Errorf("%q: got %q, want an [ERROR] containing %q", tc.line, buf.String(), tc.want)
		}
	}
	var buf strings.Builder
	interp := cli.NewInterpreter()
	interp.Out = &buf
	interp.EvaluateScript("def ok : Int -> Int = \\x -> x + 1")
	if strings.Contains(buf.String(), "[ERROR]") {
		t.Errorf("annotated def must still work: %q", buf.String())
	}
}

// F9a: guards (`pat if cond -> body`) fall through when false, see the pattern's
// bindings, and must be Bools.
func TestMatchGuards(t *testing.T) {
	interp := cli.NewInterpreter()
	interp.EvaluateScript("def f9a_sign = match { n if n > 0 -> 1 | 0 -> 0 | _ -> -1 }\ndef f9a_g = match { [x, y] if x > y -> x | [x, y] -> y }\ndef f9a_bad = match { n if n + 1 -> 1 | _ -> 2 }")
	for _, tc := range []struct{ in, want string }{
		{"5 |> f9a_sign", "[1]"}, {"0 |> f9a_sign", "[0]"}, {"-3 |> f9a_sign", "[-1]"},
		{"[3, 1] |> f9a_g", "[3]"}, {"[1, 3] |> f9a_g", "[3]"},
	} {
		if got := fmt.Sprint(interp.EvaluateToValues(tc.in)); got != tc.want {
			t.Errorf("%s: got %s want %s", tc.in, got, tc.want)
		}
	}
	got := fmt.Sprint(interp.EvaluateToValues("5 |> f9a_bad"))
	if !strings.Contains(got, "must be a Bool") {
		t.Errorf("non-Bool guard must be a loud error, got %s", got)
	}
}

// A recursive call inside an infix expression is evaluated once per level, not twice:
// with the old compile-time probe this took 2^n calls and never finished.
func TestClauseRecursionIsLinear(t *testing.T) {
	interp := cli.NewInterpreter()
	interp.EvaluateScript("def lin 0 = 0\ndef lin n = (lin (n - 1)) + 1")
	done := make(chan string, 1)
	go func() { done <- fmt.Sprint(interp.EvaluateToValues("lin 40")) }()
	select {
	case got := <-done:
		if got != "[40]" {
			t.Errorf("lin 40 = %s, want [40]", got)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("clause recursion is exponential again")
	}
}

// F9b: consecutive `def NAME <patterns> = body` lines form one curried function.
func TestClauseDefinitions(t *testing.T) {
	interp := cli.NewInterpreter()
	interp.EvaluateScript("def c_fact 0 = 1\ndef c_fact n = n * (c_fact (n - 1))\ndef c_add 0 y = y\ndef c_add x 0 = x\ndef c_add x y = x + y\ndef c_mx x y if x > y = x\ndef c_mx x y = y")
	for _, tc := range []struct{ in, want string }{
		{"c_fact 5", "[120]"}, {"c_add 3 4", "[7]"}, {"c_add 0 4", "[4]"},
		{"map (c_add 1) [1, 2]", "[[2 3]]"}, {"c_mx 3 1", "[3]"}, {"c_mx 1 3", "[3]"},
	} {
		if got := fmt.Sprint(interp.EvaluateToValues(tc.in)); got != tc.want {
			t.Errorf("%s: got %s want %s", tc.in, got, tc.want)
		}
	}
	for _, tc := range []struct{ script, want string }{
		{"def a 0 = 1\ndef a x y = 2", "earlier clauses"},
		{"def b 0 = 1\n1 + 1\ndef b n = 2", "consecutive"},
		{"def c x =", "needs a body"},
	} {
		var buf strings.Builder
		in := cli.NewInterpreter()
		in.Out = &buf
		in.EvaluateScript(tc.script)
		if !strings.Contains(buf.String(), "[ERROR]") || !strings.Contains(buf.String(), tc.want) {
			t.Errorf("%q: got %q, want [ERROR] with %q", tc.script, buf.String(), tc.want)
		}
	}
}

// F8a: a line that cannot end a statement continues it, and a line that cannot start one continues the previous.
func TestMultiLineContinuation(t *testing.T) {
	var buf strings.Builder
	interp := cli.NewInterpreter()
	interp.Out = &buf
	interp.EvaluateScript("def ml_area = \\w h ->\n  w * h\nml_area 3 4\n[1, 2, 3]\n  |> map (\\x -> x +\n     1)\n  |> length\n1 + 1\n  ++ [2]")
	out := buf.String()
	if strings.Contains(out, "[ERROR]") || !strings.Contains(out, "12") || !strings.Contains(out, "3") {
		t.Errorf("continuation failed: %q", out)
	}
}

// F1: a syntax error in a script names the line and column and points at the token; the old sentence stays first.
func TestSyntaxErrorPosition(t *testing.T) {
	var buf strings.Builder
	interp := cli.NewInterpreter()
	interp.Out = &buf
	interp.EvaluateScript("1 + 1\n\n3 + * 2")
	out := buf.String()
	for _, want := range []string{`cannot evaluate "3 + * 2"`, "at line 3, column 5", "    3 + * 2\n        ^"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %q", want, out)
		}
	}
}

// F1: the pointer comes from the lexer/parser offset, not from searching the message.
func TestSyntaxErrorPositionTable(t *testing.T) {
	cases := []struct{ script, want string }{
		{"let s = \"a + b\" in s + * 2", "at line 1, column 24"},
		{"x = (1 + 2", "at line 1,"},
		{"\n\n1 + 2 3 )", "at line 3, column 9"},
		{"let é = 1 in é + * 2", "at line 1, column 18"},
		{"1 + ?", "at line 1, column 5"},
	}
	for _, c := range cases {
		var buf strings.Builder
		interp := cli.NewInterpreter()
		interp.Out = &buf
		interp.EvaluateScript(c.script)
		if !strings.Contains(buf.String(), c.want) {
			t.Errorf("%q: want %q in %q", c.script, c.want, buf.String())
		}
	}
}

func TestCheckScriptReportsExpressionSyntaxPosition(t *testing.T) {
	diags := cli.CheckScript("let x = 1\ndef f = \\x -> x + * 1\n3 + * 2\n")
	want := map[int]int{2: 19, 3: 5}
	got := map[int]int{}
	for _, d := range diags {
		if d.Severity == "error" {
			got[d.Line] = d.Col
		}
	}
	for l, c := range want {
		if got[l] != c {
			t.Errorf("line %d: column %d, want %d (all: %v)", l, got[l], c, diags)
		}
	}
	if len(got) != 2 {
		t.Errorf("want exactly 2 errors, got %v", diags)
	}
}

func TestCheckScriptClauseDefinitions(t *testing.T) {
	lines := func(src, sev string) map[int]string {
		out := map[int]string{}
		for _, d := range cli.CheckScript(src) {
			if d.Severity == sev {
				out[d.Line] = d.Msg
			}
		}
		return out
	}
	// Compliant clauses (including a guard and a catch-all) are silent.
	if d := cli.CheckScript("def f 0 = 1\ndef f n = n\ndef g x if x > 0 = 1\ndef g x = 0\n"); len(d) != 0 {
		t.Errorf("clauses must be clean, got %v", d)
	}
	errs := lines("def a 0 y = y\ndef a x = x\necho hi\ndef a x y = x\n", "error")
	if len(errs) != 2 || !strings.Contains(errs[2], "1 parameter(s) but earlier clauses") || !strings.Contains(errs[4], "must be consecutive") {
		t.Errorf("arity and gap errors expected on lines 2 and 4, got %v", errs)
	}
	warns := lines("type Color = Red | Green | Blue\ndef n Red = 1\ndef n Green = 2\n", "warning")
	if !strings.Contains(warns[2], "missing Blue") || len(warns) != 1 {
		t.Errorf("one exhaustiveness warning on the first clause expected, got %v", warns)
	}
	if w := lines("type Color = Red | Green\ndef n Red = 1\ndef n Green = 2\n", "warning"); len(w) != 0 {
		t.Errorf("a complete clause set must be silent, got %v", w)
	}
}

// F9c: a lambda parameter may be a pattern; the lambda desugars to a hoisted match.
func TestPatternLambdas(t *testing.T) {
	interp := cli.NewInterpreter()
	interp.EvaluateScript("type Color = Red | Green\ndef pl_swap = \\(a, b) -> [b, a]\ndef pl_z = \\0 y -> y")
	for _, tc := range []struct{ in, want string }{
		{"pl_swap (pair 1 2)", "[[2 1]]"}, {"pl_z 0 5", "[5]"},
		{"map (\\0 -> \"zero\") [0]", "[[zero]]"},
		{"(\\[h, ...t] -> h) [7,8]", "[7]"},
		{"map (\\(a, b) -> a + b) [(pair 1 2), (pair 3 4)]", "[[3 7]]"},
		{"map (\\Red -> 1) [Red]", "[[1]]"},
		{"map (\\_ -> 9) [1, 2]", "[[9 9]]"},
		{"map (\\0 -> 10) [0]", "[[10]]"},
		{"map (\\(a, b) -> (\\(c, d) -> a + c) (pair 1 2)) [(pair 5 6)]", "[[6]]"},
	} {
		if got := fmt.Sprint(interp.EvaluateToValues(tc.in)); got != tc.want {
			t.Errorf("%s: got %s want %s", tc.in, got, tc.want)
		}
		if got := fmt.Sprint(interp.EvaluateToValues(tc.in)); got != tc.want {
			t.Errorf("%s (re-evaluated): got %s want %s", tc.in, got, tc.want)
		}
	}
	for _, tc := range []struct{ in, want string }{
		{"map (\\0 -> 10) [1]", "no arm matches"},
		{"map (\\Red -> 1) [Green]", "no arm matches"},
		{"fold (\\acc (k, v) -> acc) 0 [1]", "among several parameters"},
		{"map (\\0 ->) [0]", "body"},
		{"map (\\[h, ...] -> h) [1]", "lambda"},
	} {
		if got := fmt.Sprint(interp.EvaluateToValues(tc.in)); !strings.Contains(got, tc.want) {
			t.Errorf("%s: got %s, want error with %q", tc.in, got, tc.want)
		}
	}
	if w := cli.CheckScript("type Color = Red | Green\nmap (\\Red -> 1) [Red]\n"); len(w) != 1 || !strings.Contains(w[0].Msg, "missing Green") {
		t.Errorf("non-exhaustive pattern lambda must warn once, got %v", w)
	}
	if d := cli.CheckScript("map (\\(a, b) -> a) [1]\nmap (\\_ -> 1) [1]\n"); len(d) != 0 {
		t.Errorf("exhaustive pattern lambdas must be silent, got %v", d)
	}
}
