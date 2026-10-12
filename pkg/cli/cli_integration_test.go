package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
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

	t.Run("Pure DSL Conway Step", func(t *testing.T) {
		interp := NewInterpreter()
		interp.EvaluateLine("let blinker = [[2, 1], [2, 2], [2, 3]]")
		// Test frequencies of neighbors
		resBirths := interp.EvaluateToValues("let nbrs = [[x + dx, y + dy] | [x, y] <- blinker, dx <- [-1, 0, 1], dy <- [-1, 0, 1], (dx != 0 || dy != 0)] in let pairs = nbrs |> unchunk |> frequencies |> hamt_pairs |> collect_list in [pt | [pt, cnt] <- pairs, cnt == 3]")
		if len(resBirths) != 1 {
			t.Fatalf("Expected 1 result for births, got %d", len(resBirths))
		}
		t.Logf("Birth cells: %v", resBirths[0])

		// Test survivors (cnt == 2 and was in blinker)
		resNextGen := interp.EvaluateToValues("let nbrs = [[x + dx, y + dy] | [x, y] <- blinker, dx <- [-1, 0, 1], dy <- [-1, 0, 1], (dx != 0 || dy != 0)] in let pairs = nbrs |> unchunk |> frequencies |> hamt_pairs |> collect_list in let births = [pt | [pt, cnt] <- pairs, cnt == 3] in let survivors = [pt | [pt, cnt] <- pairs, cnt == 2, (pt == [2, 1] || pt == [2, 2] || pt == [2, 3])] in births ++ survivors")
		if len(resNextGen) != 1 {
			t.Fatalf("Expected 1 result for next gen, got %d", len(resNextGen))
		}
		t.Logf("Next generation cells: %v", resNextGen[0])

		// Test defining def conway_step = match { live -> ... }
		interp.EvaluateLine("def test_nbrs = match { live -> [[x + dx, y + dy] | [x, y] <- live, dx <- [-1, 0, 1], dy <- [-1, 0, 1], (dx != 0 || dy != 0)] }")
		rNbrs := interp.EvaluateToValues("spawn Life: [[2, 1], [2, 2], [2, 3]] |> test_nbrs")
		t.Logf("rNbrs: %v", rNbrs)

		interp.EvaluateLine("def test_pairs = match { live -> let nbrs = [[x + dx, y + dy] | [x, y] <- live, dx <- [-1, 0, 1], dy <- [-1, 0, 1], (dx != 0 || dy != 0)] in nbrs |> frequencies |> hamt_pairs |> collect_list }")
		rPairs := interp.EvaluateToValues("spawn Life: [[2, 1], [2, 2], [2, 3]] |> test_pairs")
		t.Logf("rPairs: %v", rPairs)

		interp.EvaluateLine("def conway_step = match { live -> let nbrs = [[x + dx, y + dy] | [x, y] <- live, dx <- [-1, 0, 1], dy <- [-1, 0, 1], (dx != 0 || dy != 0)] in let pairs = nbrs |> frequencies |> hamt_pairs |> collect_list in let births = [pt | [pt, cnt] <- pairs, cnt == 3] in let survivors = [p | [pt, cnt] <- pairs, p <- live, cnt == 2, p == pt] in births ++ survivors }")
		resGen1 := interp.EvaluateToValues("let live = [[2, 1], [2, 2], [2, 3]] in live |> conway_step")
		if len(resGen1) != 1 {
			t.Fatalf("Expected 1 result for gen 1, got %d", len(resGen1))
		}
		strGen1 := fmt.Sprintf("%v", resGen1[0])
		if !strings.Contains(strGen1, "[1 2]") || !strings.Contains(strGen1, "[2 2]") || !strings.Contains(strGen1, "[3 2]") {
			t.Errorf("Expected vertical blinker cells in Gen 1, got %s", strGen1)
		}
		if strings.Contains(strGen1, "[2 1]") || strings.Contains(strGen1, "[2 3]") {
			t.Errorf("Horizontal-only cells must be dead in Gen 1, got %s", strGen1)
		}

		resGen2 := interp.EvaluateToValues("let live = [[2, 1], [2, 2], [2, 3]] in live |> conway_step |> conway_step")
		if len(resGen2) != 1 {
			t.Fatalf("Expected 1 result for gen 2, got %d", len(resGen2))
		}
		strGen2 := fmt.Sprintf("%v", resGen2[0])
		if !strings.Contains(strGen2, "[2 1]") || !strings.Contains(strGen2, "[2 2]") || !strings.Contains(strGen2, "[2 3]") {
			t.Errorf("Expected horizontal blinker cells in Gen 2, got %s", strGen2)
		}
		if strings.Contains(strGen2, "[1 2]") || strings.Contains(strGen2, "[3 2]") {
			t.Errorf("Vertical-only cells must be dead in Gen 2, got %s", strGen2)
		}

		resGlider4 := interp.EvaluateToValues("let g = [[2, 1], [3, 2], [1, 3], [2, 3], [3, 3]] in g |> conway_step |> conway_step |> conway_step |> conway_step")
		if len(resGlider4) != 1 {
			t.Fatalf("Expected 1 result for glider gen 4, got %d", len(resGlider4))
		}
		strGlider4 := fmt.Sprintf("%v", resGlider4[0])
		for _, expectedCoord := range []string{"[3 2]", "[4 3]", "[2 4]", "[3 4]", "[4 4]"} {
			if !strings.Contains(strGlider4, expectedCoord) {
				t.Errorf("Expected glider coordinate %s in Gen 4, got %s", expectedCoord, strGlider4)
			}
		}
		if strings.Contains(strGlider4, "[2 1]") {
			t.Errorf("Origin cell [2 1] must be dead in Glider Gen 4, got %s", strGlider4)
		}
	})
}
