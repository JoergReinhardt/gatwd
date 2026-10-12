package parser_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/joergreinhardt/gatwd/parser"
)

func TestInfixParser_Evaluation(t *testing.T) {
	tests := []struct {
		expr     string
		expected any
	}{
		{"2 + 3 * 4", 14},
		{"(2 + 3) * 4", 20},
		{"7 / 2", 3.5},
		{"7 // 2", 3},
		{"10 - 4 - 2", 4},
		{"10 % 3", 1},
		{"\"hello \" ++ \"world\"", "hello world"},
		{"5 < 10 && 3 == 3", true},
		{"5 > 10 || 2 == 3", false},
	}

	for _, tc := range tests {
		val, err := parser.EvalInfixString(tc.expr, nil)
		if err != nil {
			t.Fatalf("eval failed for %q: %v", tc.expr, err)
		}
		switch exp := tc.expected.(type) {
		case int:
			// May be int or int8/int16/int32/int64
			switch v := val.(type) {
			case int:
				if v != exp {
					t.Errorf("expr %q: expected %v, got %v", tc.expr, exp, v)
				}
			case int8:
				if int(v) != exp {
					t.Errorf("expr %q: expected %v, got %v", tc.expr, exp, v)
				}
			case int16:
				if int(v) != exp {
					t.Errorf("expr %q: expected %v, got %v", tc.expr, exp, v)
				}
			case int32:
				if int(v) != exp {
					t.Errorf("expr %q: expected %v, got %v", tc.expr, exp, v)
				}
			case int64:
				if int(v) != exp {
					t.Errorf("expr %q: expected %v, got %v", tc.expr, exp, v)
				}
			default:
				t.Errorf("expr %q: unexpected type %T (%v)", tc.expr, val, val)
			}
		case float64:
			if v, ok := val.(float64); !ok || v != exp {
				t.Errorf("expr %q: expected float %v, got %v (%T)", tc.expr, exp, val, val)
			}
		case string:
			if v, ok := val.(string); !ok || v != exp {
				t.Errorf("expr %q: expected %q, got %q", tc.expr, exp, v)
			}
		case bool:
			if v, ok := val.(bool); !ok || v != exp {
				t.Errorf("expr %q: expected %v, got %v", tc.expr, exp, val)
			}
		}
	}
}

func TestInfixParser_VariableLookup(t *testing.T) {
	env := map[string]any{
		"base":  100,
		"scale": 1.5,
		"bigN":  big.NewInt(999999999999999),
	}
	lookup := func(k string) (any, bool) {
		v, ok := env[k]
		return v, ok
	}

	val, err := parser.EvalInfixString("base * 2", lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, ok := val.(int); !ok || v != 200 {
		t.Fatalf("expected 200, got %v (%T)", val, val)
	}

	val2, err := parser.EvalInfixString("base * scale", lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, ok := val2.(float64); !ok || v != 150.0 {
		t.Fatalf("expected 150.0, got %v (%T)", val2, val2)
	}
}

func TestInfixParser_ListComprehensions(t *testing.T) {
	// [x * 2 | x <- [1, 2, 3, 4], x > 2] should yield [6, 8]
	val, err := parser.EvalInfixString("[x * 2 | x <- [1, 2, 3, 4], x > 2]", nil)
	if err != nil {
		t.Fatalf("comprehension failed: %v", err)
	}
	resStr := fmt.Sprintf("%v", val)
	if resStr != "[6 8]" {
		t.Fatalf("expected '[6 8]', got %v", resStr)
	}
}

func TestInfixParser_MultiGeneratorComprehensions(t *testing.T) {
	expr := "[[x, y] | x <- [1, 2], y <- [10, 20]]"
	val, err := parser.EvalInfixString(expr, nil)
	if err != nil {
		t.Fatalf("multi-generator failed: %v", err)
	}
	resStr := fmt.Sprintf("%v", val)
	expected := "[[1 10] [1 20] [2 10] [2 20]]"
	if resStr != expected {
		t.Fatalf("expected %q, got %q", expected, resStr)
	}

	exprDestruct := "[[x + dx, y + dy] | [x, y] <- [[2, 2]], dx <- [-1, 0, 1], dy <- [-1, 0, 1], (dx != 0 || dy != 0)]"
	valD, err := parser.EvalInfixString(exprDestruct, nil)
	if err != nil {
		t.Fatalf("destructuring comprehension failed: %v", err)
	}
	resDStr := fmt.Sprintf("%v", valD)
	t.Logf("Neighbors of [2, 2]: %s", resDStr)
}

func TestInfixParser_ListConcatenation(t *testing.T) {
	expr := "[1, 2] ++ [3, 4]"
	val, err := parser.EvalInfixString(expr, nil)
	if err != nil {
		t.Fatalf("list concatenation failed: %v", err)
	}
	resStr := fmt.Sprintf("%v", val)
	expected := "[1 2 3 4]"
	if resStr != expected {
		t.Fatalf("expected %q, got %q", expected, resStr)
	}
}
