package cli

import (
	"strings"
	"testing"
)

func warnings(src string) []string {
	var out []string
	for _, d := range CheckScript(src) {
		out = append(out, d.Severity+": "+d.Msg)
	}
	return out
}

func hasWarning(ws []string, sub string) bool {
	for _, w := range ws {
		if strings.Contains(w, sub) {
			return true
		}
	}
	return false
}

func TestParseTypeAnnotation(t *testing.T) {
	cases := map[string]string{
		"Int":                       "Int",
		"Int -> Int -> Bool":        "Int -> Int -> Bool",
		"(Int -> Int) -> List(Int)": "(Int -> Int) -> List(Int)",
		"List(String) -> Int":       "List(String) -> Int",
		"Graph(String)":             "Graph(String)",
	}
	for src, want := range cases {
		ty, err := ParseTypeAnnotation(src)
		if err != nil || ty.String() != want {
			t.Errorf("%q -> %v %v, want %q", src, ty, err, want)
		}
	}
	if ty, _ := ParseTypeAnnotation("A -> B -> C"); ty.arity() != 2 || !ty.Args[1].isFn() {
		t.Error("-> must be right-associative with arity 2")
	}
	for _, bad := range []string{"", "Int ->", "(Int", "List(Int", "-> Int", "Int Int"} {
		if _, err := ParseTypeAnnotation(bad); err == nil {
			t.Errorf("%q must not parse", bad)
		}
	}
}

// TestAnnotatedDefStillRuns: the annotation is advisory and never alters evaluation.
func TestAnnotatedDefStillRuns(t *testing.T) {
	i := NewInterpreter()
	i.EvaluateScript("def inc : Int -> Int = \\x -> x + 1\ndef add : Int -> Int -> Int = \\a b -> a + b")
	if v := i.EvaluateToValues("inc 41"); len(v) != 1 || intOf(t, v[0]) != 42 {
		t.Fatalf("annotated def must evaluate normally, got %v", v)
	}
	if v := i.EvaluateToValues("spawn IntVal: 1 |> add 2"); len(v) != 1 || intOf(t, v[0]) != 3 {
		t.Fatalf("got %v", v)
	}
	if _, ok := i.Funcs["inc"]; !ok {
		t.Fatalf("the def name must not include the annotation: %v", i.Funcs)
	}
}

func TestCheckFindsDefiniteContradictions(t *testing.T) {
	src := `type IntVal = Int
def inc : Int -> Int = \x -> x + 1
def two : Int -> Int -> Int = \a b -> a * b
def greet : String -> String = \s -> s ++ "!"
def wrong : Int -> String = \x -> x + 1
inc "x"
two 1 2 3
greet 5
[1] * 2
"a" - 1
true + 1
length [1] [2]
map (\x -> x + y) [1, 2]
spawn IntVal: hello |> inc
spawn IntVal: 5 |> two "a"
[ x - 1 | x <- ["a", "b"] ]
let s = "hi" in inc s
`
	ws := warnings(src)
	for _, want := range []string{
		"argument 1 of 'inc' has type String, expected Int",
		"'two' takes 2 argument(s) but is applied to 3",
		"argument 1 of 'greet' has type Int, expected String",
		"operator '*' is not defined for List(Int) and Int",
		"operator '-' is not defined for String and Int",
		"operator '+' is not defined for Bool and Int",
		"'length' takes 1 argument(s) but is applied to 2",
		"unbound name 'y'",
		"seed 'hello' has type String, but IntVal is declared as Int",
		"argument 1 of 'two' has type String, expected Int",
		"def wrong: body has type Int but the annotation says String",
	} {
		if !hasWarning(ws, want) {
			t.Errorf("missing warning %q in %v", want, ws)
		}
	}
	for _, d := range CheckScript(src) {
		if d.Severity != "warning" {
			t.Errorf("typing findings must be advisory, got %s: %s", d.Severity, d.Msg)
		}
	}
}

// TestCheckNoFalsePositives is the gradual-typing contract: anything unknown is Any
// and silent, valid programs produce nothing.
func TestCheckNoFalsePositives(t *testing.T) {
	src := `type IntVal = Int
type Node = Server
def inc : Int -> Int = \x -> x + 1
def add = \a b -> a + b
def flex : Int -> Float = \x -> x * 1.5
const K = 5
let blinker = [[2, 1], [2, 2]]
inc 1
inc K
add 1 2
(add 1) 2
map (\x -> inc x) [1, 2]
[ inc x | x <- [1, 2, 3], x > 1 ]
let n = 5 in inc n
spawn IntVal: 5 |> inc |> inc
spawn Node: anything |> echo hi
spawn IntVal: 3 |> map (\x -> echo hello %s)
fold (\a x -> a + x) 0 [1, 2, 3]
def macro = graph_topo_sort
"a" ++ "b"
[1] ++ [2]
1 + 2.5
`
	if ws := warnings(src); len(ws) != 0 {
		t.Fatalf("valid program must be silent, got %v", ws)
	}
}

func TestCheckStructuralErrorsStillErrors(t *testing.T) {
	ds := CheckScript("let x := 1\ndef broken\nspawn IntVal: 1 |>")
	errs := 0
	for _, d := range ds {
		if d.Severity == "error" {
			errs++
		}
	}
	if errs != 3 {
		t.Fatalf("want 3 structural errors, got %v", ds)
	}
}
