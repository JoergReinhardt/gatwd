package cli_test

import (
	"iter"
	"strings"
	"testing"

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
