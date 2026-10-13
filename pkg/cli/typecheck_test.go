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
	i.EvaluateScript("type IntVal = Int\ndef inc : Int -> Int = \\x -> x + 1\ndef add : Int -> Int -> Int = \\a b -> a + b")
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
inc 'a'
'a' + 1
def cwrong : Char -> Int = \c -> c
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
		"argument 1 of 'inc' has type Char, expected Int",
		"operator '+' is not defined for Char and Int",
		"def cwrong: body has type Char but the annotation says Int",
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
def isa : Char -> Bool = \c -> c == 'a'
def code : Char -> Int = \c -> char_code c
def asStr : Char -> String = \c -> char_to_str c
isa 'a'
isa "a"
code 'z'
'a' < 'b'
'a' == "a"
['a', 'b'] |> map (\c -> code c)
not true
and true false
[true, false] |> fold or false
compare 1 2 == LT
compare "a" "b" != GT
ordering_to_int (compare 1 2)
def cmp3 : Int -> Int -> Ordering = \a b -> compare a b
true |> not
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

// ---- constraint names (hardening W3): Num, Integral, Fractional, Comparable, Appendable ----

func TestConstraintNamesParse(t *testing.T) {
	for _, src := range []string{"Num", "Num -> Num -> Num", "Integral -> Fractional", "Comparable -> Comparable -> Bool", "List(Num) -> Appendable"} {
		ty, err := ParseTypeAnnotation(src)
		if err != nil || ty.String() != src {
			t.Errorf("%q -> %v, %v", src, ty, err)
		}
	}
	if ty, _ := ParseTypeAnnotation("Num"); !ty.isConstraint() {
		t.Error("Num must be a constraint")
	}
	if ty, _ := ParseTypeAnnotation("List(Num)"); ty.isConstraint() {
		t.Error("List(Num) is a type, not a constraint")
	}
	for _, name := range []string{"Num", "Integral", "Fractional", "Comparable", "Appendable"} {
		if !IsConstraintName(name) {
			t.Errorf("%s must be a constraint name", name)
		}
		if _, err := ParseTypeAnnotation(name + "(Int)"); err == nil {
			t.Errorf("%s(Int) must not parse: a constraint takes no type arguments", name)
		}
	}
	if IsConstraintName("Int") || IsConstraintName("Number") {
		t.Error("Int and Number are not constraint names")
	}
}

// Valid uses of every constraint: the checker must stay silent (zero false positives).
func TestConstraintsValidProgramIsSilent(t *testing.T) {
	src := `
def sum2 : Num -> Num -> Num = \a b -> a + b
def half : Fractional -> Fractional = \x -> x / 2
def idiv2 : Integral -> Integral = \n -> n // 2
def modulo : Integral -> Integral -> Integral = \a b -> a % b
def below : Comparable -> Comparable -> Bool = \a b -> a < b
def glue : Appendable -> Appendable -> Appendable = \a b -> a ++ b
def neg : Num -> Num = \x -> -x
def sq : Num -> Num = \x -> x ** 2
def as_int : Num -> Int = \x -> 1
def mean : List(Num) -> Num = \xs -> sum2 (head xs) 1
def keep : Num -> Num = \n -> sum2 n 1
sum2 1 2
sum2 1.5 2
sum2 (1 + 2i) 1
half 3
half 3.5
idiv2 7
modulo 7 3
below "a" "b"
below 1 2
below 'a' 'b'
below [1] [2]
glue "a" "b"
glue [1] [2]
neg 4
sq 3
as_int 2.5
keep (sum2 1 2)
[1, 2, 3] |> map (\x -> sum2 x 1)
`
	if ws := warnings(src); len(ws) != 0 {
		t.Fatalf("valid constraint program must be silent, got %v", ws)
	}
}

// Each contradiction is reported, with the constraint named in the message.
func TestConstraintsReportContradictions(t *testing.T) {
	for src, want := range map[string]string{
		"def sum2 : Num -> Num = \\a -> a\nsum2 \"a\"":                                                   "expected Num",
		"def sum2 : Num -> Num = \\a -> a\nsum2 true":                                                    "expected Num",
		"def idiv : Integral -> Int = \\a -> 1\nidiv \"x\"":                                              "expected Integral",
		"def fr : Fractional -> Int = \\a -> 1\nfr [1]":                                                  "expected Fractional",
		"def cmp : Comparable -> Int = \\a -> 1\ncmp (3 + 4i)":                                           "expected Comparable",
		"def app : Appendable -> Int = \\a -> 1\napp 3":                                                  "expected Appendable",
		"def app : Appendable -> Int = \\a -> 1\napp true":                                               "expected Appendable",
		"def f : Num -> String = \\a -> \"s\"\ndef g : String -> Int = \\s -> 1\ng (f 1)":                "", // fine
		"def bad : Integral -> Fractional = \\a -> a\ndef use : Integral -> Int = \\i -> 1\nuse (bad 1)": "", // Fractional into Integral: meets only through the tower; see below
		"def bad : Integral -> Int = \\x -> -\"a\"":                                                      "operator '-' is not defined for String",
		"def bad : Num(Int) -> Int = \\x -> 1":                                                           "takes no type arguments",
	} {
		ws := warnings(src)
		if want == "" {
			continue
		}
		if !hasWarning(ws, want) {
			t.Errorf("%q: want a warning containing %q, got %v", src, want, ws)
		}
	}
	// a String-typed result where a Num is promised
	if ws := warnings("def f : Num -> Num = \\a -> \"s\""); !hasWarning(ws, "annotation says Num") {
		t.Errorf("a String body under a Num annotation must warn, got %v", ws)
	}
	// two constraints with no member in common contradict each other
	if ws := warnings("def only_int : Integral -> Int = \\a -> 1\ndef frac : Fractional -> Int = \\a -> only_int a"); !hasWarning(ws, "expected Integral") {
		t.Errorf("a Fractional where an Integral is wanted must warn, got %v", ws)
	}
}

// Unary minus is typed through the runtime oracle.
func TestUnaryMinusIsChecked(t *testing.T) {
	if ws := warnings("def n : Int -> Int = \\x -> -x\nn 3\n- 5\n-2.5\n-(1 + 2)"); len(ws) != 0 {
		t.Errorf("valid unary minus must be silent, got %v", ws)
	}
	if ws := warnings(`- "a"`); !hasWarning(ws, "operator '-' is not defined for String") {
		t.Errorf(`- "a" must warn, got %v`, ws)
	}
	if ws := warnings("def s : String -> Int = \\x -> -x"); !hasWarning(ws, "operator '-' is not defined for String") {
		t.Errorf("negating a String parameter must warn, got %v", ws)
	}
}

// Ratio in the checker: a numeric type, a member of Num and Fractional, typed from literals.
func TestRatioIsChecked(t *testing.T) {
	valid := `
def half : Ratio -> Ratio = \x -> x / 2r1
def third : Fractional -> Fractional = \x -> x * 1r3
def total : Num -> Num -> Num = \a b -> a + b
def keep : Ratio -> Num = \x -> x
half 1r3
half 1
third 1r2
total 1r2 3
total 1r2 0.5
[1r2, 1, 2.5] |> map (\x -> total x 1r2)
`
	if ws := warnings(valid); len(ws) != 0 {
		t.Fatalf("valid Ratio program must be silent, got %v", ws)
	}
	for src, want := range map[string]string{
		"def f : Ratio -> Int = \\x -> x\nf \"a\"": "expected Ratio",
		"def f : Num -> Num = \\x -> x - \"a\"":    "operator '-' is not defined for Num and String",
		"def f : Num -> String = \\x -> x + 1":     "annotation says String",
		"def f : Ratio -> String = \\x -> x + 1r2": "annotation says String",
		"1r2 + \"a\"": "operator '+' is not defined for Ratio and String",
		"def f : Integral -> Int = \\x -> 1\nf 1r2 ++ \"a\" ++ true": "", // nothing certain: stays silent or not, but must not panic
	} {
		ws := warnings(src)
		if want != "" && !hasWarning(ws, want) {
			t.Errorf("%q: want a warning containing %q, got %v", src, want, ws)
		}
	}
	if ty := tyOfValue(mustRatio(t)); ty.Name != "Ratio" || !ty.isNumeric() {
		t.Errorf("a Ratio literal must be typed Ratio and numeric, got %v", ty)
	}
}

func mustRatio(t *testing.T) any {
	t.Helper()
	got := NewInterpreter().EvaluateToValues("1r2")
	if len(got) != 1 {
		t.Fatalf("1r2 = %v", got)
	}
	return got[0]
}
