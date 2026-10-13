package cli

import (
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

func closureInterp(t *testing.T) *Interpreter {
	t.Helper()
	i := NewInterpreter()
	i.EvaluateScript("type IntVal = Int\ndef add = \\x y -> x + y\ndef sub = \\x y -> x - y\ndef pos = \\x -> x > 0\ndef mk = \\k -> \\x -> x + k\ndef pipe_macro = map (\\x -> x + 1)")
	return i
}

func intOf(t *testing.T, v any) int {
	t.Helper()
	f, ok := functions.ToFloat64(v)
	if !ok {
		t.Fatalf("not a number: %T %v", v, v)
	}
	return int(f)
}

// TestDefLambdaResolvesToClosure proves a lambda-bodied def is a real *Closure value
// of the declared arity (not re-parsed macro text), and a partial application is a
// Closure of the remaining arity.
func TestDefLambdaResolvesToClosure(t *testing.T) {
	i := closureInterp(t)
	c, ok := i.defClosure("add")
	if !ok || c.Arity() != 2 {
		t.Fatalf("add must be a 2-ary closure, got %v %v", c, ok)
	}
	vals := i.EvaluateToValues("add 1")
	part, ok := vals[0].(*functions.Closure)
	if !ok || part.Arity() != 1 {
		t.Fatalf("add 1 must be a 1-ary *Closure, got %T", vals[0])
	}
	if _, ok := i.defClosure("pipe_macro"); ok {
		t.Fatal("a non-lambda def must stay a pipeline macro")
	}
}

// TestClosureApplicationForms covers expression position, over-application and
// data-last pipeline application, asserting on the raw numeric value.
func TestClosureApplicationForms(t *testing.T) {
	i := closureInterp(t)
	cases := map[string]int{
		"add 1 2":                  3,
		"(add 10) 5":               15,
		"mk 5 10":                  15,
		"spawn IntVal: 5 |> add 3": 8,
		"let f = add 100 in spawn IntVal: 1 |> f": 101,
	}
	for src, want := range cases {
		vals := i.EvaluateToValues(src)
		if len(vals) != 1 || intOf(t, vals[0]) != want {
			t.Errorf("%s = %v, want %d", src, vals, want)
		}
	}
}

// TestClosureHigherOrderSteps proves map/filter/reject/fold accept closure names and
// partial applications, yielding structural collections.
func TestClosureHigherOrderSteps(t *testing.T) {
	i := closureInterp(t)
	asInts := func(src string) []int {
		vals := i.EvaluateToValues(src)
		if len(vals) != 1 {
			t.Fatalf("%s: want 1 value, got %v", src, vals)
		}
		sl, ok := functions.ToSliceAny(vals[0])
		if !ok {
			t.Fatalf("%s: want a collection, got %T", src, vals[0])
		}
		out := make([]int, len(sl))
		for idx, v := range sl {
			out[idx] = intOf(t, v)
		}
		return out
	}
	eq := func(a, b []int) bool {
		if len(a) != len(b) {
			return false
		}
		for k := range a {
			if a[k] != b[k] {
				return false
			}
		}
		return true
	}
	if got := asInts("[1, 2, 3] |> map (add 10)"); !eq(got, []int{11, 12, 13}) {
		t.Errorf("map partial: %v", got)
	}
	if got := asInts("[1, -2, 3] |> filter pos"); !eq(got, []int{1, 3}) {
		t.Errorf("filter closure: %v", got)
	}
	if got := asInts("[1, -2, 3] |> reject pos"); !eq(got, []int{-2}) {
		t.Errorf("reject closure: %v", got)
	}
	if v := i.EvaluateToValues("[1, 2, 3] |> fold add 0"); len(v) != 1 || intOf(t, v[0]) != 6 {
		t.Errorf("fold seeded: %v", v)
	}
	if v := i.EvaluateToValues("[4, 5] |> fold add"); len(v) != 1 || intOf(t, v[0]) != 9 {
		t.Errorf("fold unseeded: %v", v)
	}
}

// TestClosureErrorsAreLoud: applying a closure to an incompatible value surfaces an
// error value instead of vanishing or panicking.
func TestClosureErrorsAreLoud(t *testing.T) {
	i := closureInterp(t)
	out := i.evaluatePipelineStr(`spawn IntVal: 1 |> sub "a"`)
	if len(out) != 1 || len(out[0]) < 7 || out[0][:7] != "[ERROR]" {
		t.Fatalf("want an [ERROR] result, got %v", out)
	}
}

// TestEvaluateToValuesEmptyInput: empty input yields no values instead of panicking.
func TestEvaluateToValuesEmptyInput(t *testing.T) {
	i := NewInterpreter()
	for _, in := range []string{"", "   ", "\n"} {
		if v := i.EvaluateToValues(in); len(v) != 0 {
			t.Errorf("%q: want no values, got %v", in, v)
		}
	}
}
