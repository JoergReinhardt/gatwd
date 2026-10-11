package cli

import (
	"fmt"
	"github.com/joergreinhardt/gatwd/functions"
	"testing"
)

func TestStructuralPipelines(t *testing.T) {
	interpreter := NewInterpreter()
	interpreter.EvaluateLine("type IntVal = Int")

	// Register the test type
	interpreter.Parser.RegisterTypeFactory("IntVal", func(val string) any {
		var i int
		fmt.Sscanf(val, "%d", &i)
		return i
	})

	tests := []struct {
		name     string
		pipeline string
		assert   func(t *testing.T, results []any)
	}{
		{
			name:     "Primitive Spawning",
			pipeline: "spawn IntVal: 99",
			assert: func(t *testing.T, results []any) {
				if len(results) != 1 {
					t.Fatalf("Expected 1 result, got %d", len(results))
				}
				if val, ok := results[0].(int); !ok || val != 99 {
					t.Errorf("Expected int(99), got %T(%v)", results[0], results[0])
				}
			},
		},
		{
			name:     "Native Math Addition",
			pipeline: "spawn IntVal: 99 |> math_add 1",
			assert: func(t *testing.T, results []any) {
				if len(results) != 1 {
					t.Fatalf("Expected 1 result, got %d", len(results))
				}
				if val, ok := results[0].(int); !ok || val != 100 {
					t.Errorf("Expected int(100), got %T(%v)", results[0], results[0])
				}
			},
		},
		{
			name:     "Native Pattern Matching",
			pipeline: "spawn IntVal: 99 |> match { 99 -> math_add 1 |> math_add 2 }",
			assert: func(t *testing.T, results []any) {
				if len(results) != 1 {
					t.Fatalf("Expected 1 result, got %d", len(results))
				}
				if val, ok := results[0].(int); !ok || val != 102 {
					t.Errorf("Expected int(102), got %T(%v)", results[0], results[0])
				}
			},
		},
		{
			name:     "Native Kleisli FanOut Arrow",
			pipeline: "spawn IntVal: 1 |> (math_add 100 &&& math_add 200)",
			assert: func(t *testing.T, results []any) {
				if len(results) != 1 {
					t.Fatalf("Expected 1 result, got %d", len(results))
				}
				pair, ok := results[0].(functions.Pair[any, any])
				if !ok {
					t.Fatalf("Expected functions.Pair, got %T", results[0])
				}
				if pair.Fst != 101 || pair.Snd != 201 {
					t.Errorf("Expected {101, 201}, got {%v, %v}", pair.Fst, pair.Snd)
				}
			},
		},
		{
			name:     "Native Applicative Split Arrow",
			pipeline: "spawn IntVal: 1 |> (math_add 100 &&& math_add 200) |> [math_add 100 *** math_add 200]",
			assert: func(t *testing.T, results []any) {
				if len(results) != 1 {
					t.Fatalf("Expected 1 result, got %d", len(results))
				}
				pair, ok := results[0].(functions.Pair[any, any])
				if !ok {
					t.Fatalf("Expected functions.Pair, got %T", results[0])
				}
				if pair.Fst != 201 || pair.Snd != 401 {
					t.Errorf("Expected {201, 401}, got {%v, %v}", pair.Fst, pair.Snd)
				}
			},
		},
		{
			name:     "Functional FlatMap Closure",
			pipeline: "spawn IntVal: 10 |> map (\\x -> math_add 5 |> math_add 5)",
			assert: func(t *testing.T, results []any) {
				if len(results) != 1 {
					t.Fatalf("Expected 1 result, got %d", len(results))
				}
				if val, ok := results[0].(int); !ok || val != 20 {
					t.Errorf("Expected int(20), got %T(%v)", results[0], results[0])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := interpreter.EvaluateToValues(tt.pipeline)
			tt.assert(t, results)
		})
	}

	t.Run("Ackermann Non-Primitive Recursion", func(t *testing.T) {
		interp := NewInterpreter()
		interp.EvaluateLine("def ackermann = match { [0, n] -> n + 1 | [m, 0] -> [m - 1, 1] |> ackermann | [m, n] -> let res = [m, n - 1] |> ackermann in [m - 1, res] |> ackermann }")
		res := interp.EvaluateToValues("[2, 2] |> ackermann")
		if len(res) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(res))
		}
		if fmt.Sprintf("%v", res[0]) != "7" {
			t.Errorf("Expected 7 for A(2, 2), got %v", res[0])
		}
		res29 := interp.EvaluateToValues("[3, 2] |> ackermann")
		if len(res29) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(res29))
		}
		if fmt.Sprintf("%v", res29[0]) != "29" {
			t.Errorf("Expected 29 for A(3, 2), got %v", res29[0])
		}
	})
}
