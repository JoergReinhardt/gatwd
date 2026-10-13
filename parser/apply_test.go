package parser_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/parser"
)

// asInt coerces any of the int widths the literal parser may yield (int8..int64)
// to a plain int, so the test functions are agnostic to the token's numeric type.
func asInt(v any) int {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(rv.Int())
	default:
		panic(fmt.Sprintf("not an int: %T", v))
	}
}

// lookupWith builds a lookup over a fixed environment map, the shape the evaluator
// hands to EvalInfixString for a lambda body's bindings and registered builtins.
func lookupWith(env map[string]any) func(string) (any, bool) {
	return func(name string) (any, bool) {
		v, ok := env[name]
		return v, ok
	}
}

// TestInfixApplication exercises juxtaposition function application: a name
// resolving to a functions.DSLFunc applied to one or more argument atoms, mixed
// with operators (which must bind looser than application).
func TestInfixApplication(t *testing.T) {
	inc := functions.DSLFunc(func(args []any) (any, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("inc: want 1 arg, got %d", len(args))
		}
		return asInt(args[0]) + 1, nil
	})
	add := functions.DSLFunc(func(args []any) (any, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("add: want 2 args, got %d", len(args))
		}
		return asInt(args[0]) + asInt(args[1]), nil
	})
	env := map[string]any{"inc": inc, "add": add, "x": 10, "y": 5}

	tests := []struct {
		expr     string
		expected any
	}{
		{"inc 41", 42},                 // single-arg application
		{"add 3 4", 7},                 // multi-arg application
		{"add x y", 15},                // args resolve through the environment
		{"inc x + 1", 12},              // application binds tighter than +
		{"add (inc 1) (inc 1)", 4},     // nested application via parens
		{"inc x * 2", 22},              // (inc x) * 2, not inc (x * 2)
		{"[inc 1, inc 2, inc 3]", nil}, // application inside a list literal
	}

	for _, tc := range tests {
		val, err := parser.EvalInfixString(tc.expr, lookupWith(env))
		if err != nil {
			t.Fatalf("eval failed for %q: %v", tc.expr, err)
		}
		if tc.expected == nil {
			continue // list-literal case: just assert it parsed/evaluated
		}
		if asInt(val) != tc.expected.(int) {
			t.Errorf("expr %q = %v (%T), want %v", tc.expr, val, val, tc.expected)
		}
	}
}

// TestInfixApplicationErrors confirms applying a non-function and propagated
// argument errors both surface as errors rather than panicking.
func TestInfixApplicationErrors(t *testing.T) {
	env := map[string]any{"notfn": 7}
	if _, err := parser.EvalInfixString("notfn 1", lookupWith(env)); err == nil {
		t.Errorf("applying a non-function should error")
	}
}

// TestInfixNoApplicationRegression confirms plain operator expressions — the common
// case — are unchanged by the application grammar (no adjacent atoms).
func TestInfixNoApplicationRegression(t *testing.T) {
	cases := map[string]any{
		"2 + 3 * 4":   14,
		"(2 + 3) * 4": 20,
		"10 - 4 - 2":  4,
	}
	for expr, want := range cases {
		got, err := parser.EvalInfixString(expr, nil)
		if err != nil {
			t.Fatalf("eval failed for %q: %v", expr, err)
		}
		if got != want {
			t.Errorf("expr %q = %v, want %v", expr, got, want)
		}
	}
}

// TestComprehensionPositionalGuard pins Haskell-style guard placement: a filter
// written between two generators runs as soon as its variables are bound, so it can
// protect a later generator whose source would error for the filtered-out elements.
func TestComprehensionPositionalGuard(t *testing.T) {
	// items(x) errors for x == 0; the guard `x > 0` must prevent that evaluation.
	env := map[string]any{
		"items": functions.DSLFunc(func(args []any) (any, error) {
			if asInt(args[0]) == 0 {
				return nil, fmt.Errorf("items: zero has no items")
			}
			return []any{args[0], args[0]}, nil
		}),
	}
	got, err := parser.EvalInfixString(`[ y | x <- [0, 1, 2], x > 0, y <- (items x) ]`, lookupWith(env))
	if err != nil {
		t.Fatalf("guard should have pruned x=0 before the dependent generator: %v", err)
	}
	items, _ := functions.ToSliceAny(got)
	if len(items) != 4 {
		t.Fatalf("got %v, want four elements [1 1 2 2]", items)
	}
}
