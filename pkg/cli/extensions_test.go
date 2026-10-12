package cli_test

import (
	"iter"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/pkg/cli"
)

func TestBareInterpreter_CoreAndExtension(t *testing.T) {
	// 1. Bare Interpreter with no builtins
	bare := cli.NewBareInterpreter()
	bare.EvaluateLine("type IntVal = Int")

	// Core primitive works
	res := bare.EvaluateToValues("spawn IntVal: 0 |> emit 1, 2, 3 |> drop 1 |> head |> assert 1")
	if len(res) == 0 || res[0] != "1" {
		t.Fatalf("expected bare interpreter to execute core emit/drop/head/assert, got %v", res)
	}

	// 2. Register a third-party custom step handler onto the bare interpreter
	bare.RegisterStep(func(stepStr string, interp *cli.Interpreter) (evaluator.PipelineStep[functions.Either[error, any]], bool) {
		if strings.HasPrefix(stepStr, "custom_multiply ") {
			// Mocking a third-party arithmetic library extension
			return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
				return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
					for node := range stream {
						if node.Value.IsRight() {
							val, _ := node.Value.UnwrapRight()
							strVal := val.(string)
							yield(dsltypes.NewASTNode(functions.Right[error, any](any(strVal+"_custom")), node.DSLType))
						}
					}
				}
			}, true
		}
		return nil, false
	})

	resCustom := bare.EvaluateToValues("spawn IntVal: 0 |> emit hello |> drop 1 |> custom_multiply 2 |> assert hello_custom")
	if len(resCustom) == 0 || resCustom[0] != "hello_custom" {
		t.Fatalf("expected custom step to execute, got %v", resCustom)
	}
}

func TestBuiltins_HAMTAndSetAndChunk(t *testing.T) {
	full := cli.NewInterpreter()
	full.RegisterType("type StrVal = String", func(s string) any { return s })
	full.RegisterType("type IntVal = Int", func(s string) any { return s })

	// HAMT set, get, del
	hamtRes := full.EvaluateToValues("spawn StrVal: Start |> emit user:alice |> to_hamt |> hamt_set role admin |> hamt_get role |> assert admin")
	if len(hamtRes) == 0 || hamtRes[0] != "admin" {
		t.Fatalf("expected HAMT set and get to yield admin, got %v", hamtRes)
	}

	// HAMT delete
	hamtDelRes := full.EvaluateToValues("spawn StrVal: Start |> emit a:1, b:2 |> to_hamt |> hamt_del a |> hamt_get b |> assert 2")
	if len(hamtDelRes) == 0 || hamtDelRes[0] != "2" {
		t.Fatalf("expected HAMT delete to keep remaining key b:2, got %v", hamtDelRes)
	}

	// Set contains, add, remove
	setAddRes := full.EvaluateToValues("spawn StrVal: Start |> emit alpha |> to_set |> set_add beta |> set_contains beta |> assert true")
	if len(setAddRes) == 0 || setAddRes[0] != true {
		t.Fatalf("expected Set add and contains to return true, got %v", setAddRes)
	}

	// Chunk and Unchunk roundtrip
	unchunkRes := full.EvaluateToValues("spawn IntVal: 0 |> emit 10, 20, 30 |> drop 1 |> chunk 2 |> unchunk |> collect_list |> assert [10 20 30]")
	if len(unchunkRes) == 0 {
		t.Fatalf("expected unchunk roundtrip to succeed, got %v", unchunkRes)
	}
}

func TestEmbeddedScripts_EvaluateScriptAndFS(t *testing.T) {
	interp := cli.NewInterpreter()

	// 1. Test EvaluateScript with embedded multi-line script content (simulating //go:embed string)
	embeddedDSL := `
type Metric = Int
const BASE = 50
def bump = math_add 10
spawn Metric: BASE |> bump |> assert 60
`
	interp.EvaluateScript(embeddedDSL)
	val, ok := interp.Constants["BASE"]
	if !ok || val != "50" {
		t.Fatalf("expected constant 'BASE' to be set to '50', got %v", val)
	}

	// 2. Test LoadScriptFS with virtual filesystem (simulating //go:embed embed.FS)
	mockFS := fstest.MapFS{
		"scripts/operator_reconcile.gatwd": &fstest.MapFile{
			Data: []byte(`
type ClusterStatus = OK | DEGRADED
const EXPECTED_STATE = OK
spawn ClusterStatus: EXPECTED_STATE |> assert OK
`),
		},
	}

	err := interp.LoadScriptFS(mockFS, "scripts/operator_reconcile.gatwd")
	if err != nil {
		t.Fatalf("expected LoadScriptFS to succeed, got error: %v", err)
	}

	opState, ok := interp.Constants["EXPECTED_STATE"]
	if !ok || opState != "OK" {
		t.Fatalf("expected constant 'EXPECTED_STATE' to be 'OK', got %v", opState)
	}
}

func TestFPCompletenessAndOperations(t *testing.T) {
	interp := cli.NewInterpreter()

	// 1. Extended Folds
	resSum := interp.EvaluateToValues("spawn IntVal: 0 |> emit 10, 20, 30 |> drop 1 |> fold sum |> assert 60")
	if len(resSum) == 0 {
		t.Fatalf("fold sum failed: %v", resSum)
	}

	resProd := interp.EvaluateToValues("spawn IntVal: 0 |> emit 2, 3, 4 |> drop 1 |> fold product |> assert 24")
	if len(resProd) == 0 {
		t.Fatalf("fold product failed: %v", resProd)
	}

	resMin := interp.EvaluateToValues("spawn IntVal: 0 |> emit 50, 10, 80 |> drop 1 |> fold min |> assert 10")
	if len(resMin) == 0 {
		t.Fatalf("fold min failed: %v", resMin)
	}

	resMax := interp.EvaluateToValues("spawn IntVal: 0 |> emit 50, 10, 80 |> drop 1 |> fold max |> assert 80")
	if len(resMax) == 0 {
		t.Fatalf("fold max failed: %v", resMax)
	}

	resConcat := interp.EvaluateToValues("spawn StrVal: Start |> emit A, B, C |> drop 1 |> fold concat |> assert ABC")
	if len(resConcat) == 0 {
		t.Fatalf("fold concat failed: %v", resConcat)
	}

	// 2. Map Lambdas over Stream
	resMapStream := interp.EvaluateToValues("spawn IntVal: 0 |> emit 1, 2, 3 |> drop 1 |> map (\\x -> x * 10) |> collect_list |> assert [10 20 30]")
	if len(resMapStream) == 0 {
		t.Fatalf("map lambda over stream failed: %v", resMapStream)
	}

	// 3. Map Lambdas over Collection
	resMapColl := interp.EvaluateToValues("let xs = [1, 2, 3] in xs |> map (\\x -> x + 10) |> assert [11 12 13]")
	if len(resMapColl) == 0 {
		t.Fatalf("map lambda over list failed: %v", resMapColl)
	}

	// 4. Infix Concatenation and Lexicographical Compare
	resInfix := interp.EvaluateToValues("let a = [1, 2] in let b = [3, 4] in (a ++ b) |> assert [1 2 3 4]")
	if len(resInfix) == 0 {
		t.Fatalf("list infix concat failed: %v", resInfix)
	}

	resCmp := interp.EvaluateToValues("let a = [1, 2] in let b = [1, 3] in (a < b) |> assert true")
	if len(resCmp) == 0 || resCmp[0] != true {
		t.Fatalf("list lexicographical compare failed: %v", resCmp)
	}
}
