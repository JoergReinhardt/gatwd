package parser_test

import (
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/parser"
)

// TestLambdaClosures covers lambda abstraction as an expression value: immediate
// application, multi-argument lambdas, currying/partial application, over-application
// through a returned function, lexical capture and higher-order use.
func TestLambdaClosures(t *testing.T) {
	env := map[string]any{"n": 100}
	tests := []struct {
		expr     string
		expected any
	}{
		{`(\x -> x + 1) 41`, 42},
		{`(\x y -> x * y) 6 7`, 42},
		{`(\x y -> x - y) 10`, nil}, // partial application: a closure, checked below
		{`((\x y -> x - y) 10) 3`, 7},
		{`(\x -> \y -> x + y) 1 2`, 3},            // over-application through a returned closure
		{`(\x -> x + n) 1`, 101},                  // free variable resolved in the defining env
		{`(\f x -> f (f x)) (\v -> v * 2) 5`, 20}, // higher-order: functions as arguments
		{`(\x -> x * x) 3 + 1`, 10},               // application binds tighter than +
		{`(\x -> -5 + x) 7`, 2},                  // `->` followed by a signed literal
	}
	for _, tc := range tests {
		got, err := parser.EvalInfixString(tc.expr, lookupWith(env))
		if err != nil {
			t.Errorf("%s: unexpected error %v", tc.expr, err)
			continue
		}
		if tc.expected == nil {
			c, ok := got.(*functions.Closure)
			if !ok || c.Arity() != 1 {
				t.Errorf("%s: want a 1-ary *Closure, got %T %v", tc.expr, got, got)
			}
			continue
		}
		if asInt(got) != tc.expected.(int) {
			t.Errorf("%s: got %v want %v", tc.expr, got, tc.expected)
		}
	}
}

// TestLambdaShadowingAndLexicalScope proves a closure keeps the environment of its
// definition site and that parameters shadow outer names.
func TestLambdaShadowingAndLexicalScope(t *testing.T) {
	mk, err := parser.EvalInfixString(`\k -> \x -> x + k`, lookupWith(nil))
	if err != nil {
		t.Fatal(err)
	}
	add5, err := functions.Apply(mk, []any{5})
	if err != nil {
		t.Fatal(err)
	}
	add7, _ := functions.Apply(mk, []any{7})
	r5, _ := functions.Apply(add5, []any{1})
	r7, _ := functions.Apply(add7, []any{1})
	if asInt(r5) != 6 || asInt(r7) != 8 {
		t.Fatalf("captured k leaked between closures: %v %v", r5, r7)
	}

	got, err := parser.EvalInfixString(`(\x -> x) 1`, lookupWith(map[string]any{"x": 99}))
	if err != nil || asInt(got) != 1 {
		t.Fatalf("param must shadow outer x: %v %v", got, err)
	}
}

// TestLambdaMalformed ensures a lambda without parameters or arrow is a parse error,
// not a silent success.
func TestLambdaMalformed(t *testing.T) {
	for _, src := range []string{`\ -> 1`, `\x 1`, `\x ->`} {
		if _, err := parser.ParseInfixExpr(src); err == nil {
			t.Errorf("%q: expected a parse error", src)
		}
	}
}

// TestApplyNonFunctionStillErrors keeps the non-callable error discipline.
func TestApplyNonFunctionStillErrors(t *testing.T) {
	if _, err := parser.EvalInfixString(`x 1`, lookupWith(map[string]any{"x": 3})); err == nil {
		t.Fatal("applying an int must error")
	}
}
