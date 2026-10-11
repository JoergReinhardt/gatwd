package cli_test

import (
	"testing"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

func TestInterpreter_Constants(t *testing.T) {
	interp := cli.NewInterpreter()
	interp.EvaluateLine("type IntVal = Int")

	// 1. Native Go Constant Registration
	interp.RegisterConstant("STATUS_OK", 200)
	interp.RegisterConstants(map[string]any{
		"BASE_OFFSET": 50,
		"MULTIPLIER":  2,
	})

	// Verify constant used in spawn and pipeline
	results := interp.EvaluateToValues("spawn IntVal: BASE_OFFSET |> math_add 10 |> assert 60")
	if len(results) != 1 || results[0] != 60 {
		t.Fatalf("expected result 60 from constant BASE_OFFSET, got %v", results)
	}

	// 2. DSL Script-defined constants
	interp.EvaluateLine("const TAX_RATE = 15")
	interp.EvaluateLine("const INITIAL_BALANCE = 100")

	results = interp.EvaluateToValues("spawn IntVal: INITIAL_BALANCE |> math_add TAX_RATE |> assert 115")
	if len(results) != 1 || results[0] != 115 {
		t.Fatalf("expected result 115 from DSL constants, got %v", results)
	}
}
