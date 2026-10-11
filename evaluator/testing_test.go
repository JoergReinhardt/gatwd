package evaluator_test

import (
	"fmt"
	"iter"
	"testing"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

func makeIntStream(count int) func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
	return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		for i := 0; i < count; i++ {
			node := dsltypes.NewASTNode(
				functions.Right[error, any](any(i)),
				dsltypes.BitFlag(1),
			)
			if !yield(node) {
				return
			}
		}
	}
}

func TestVerifyExampleOutput(t *testing.T) {
	exampleOutput := `
		10
		20
		30
	`
	actualOutput := []string{"10", "20", "30"}
	evaluator.VerifyExampleOutput(t, exampleOutput, actualOutput)
}

func TestStepTelemetry(t *testing.T) {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	var count uint64 = 0

	telemetryStep := evaluator.StepTelemetry(&count)
	stream := engine.EvaluatePipeline(
		makeIntStream(25),
		[]evaluator.PipelineStep[functions.Either[error, any]]{telemetryStep},
	)

	consumed := 0
	for range stream {
		consumed++
	}

	if consumed != 25 {
		t.Fatalf("expected 25 items consumed, got %d", consumed)
	}
	if count != 25 {
		t.Fatalf("expected telemetry counter to equal 25, got %d", count)
	}
}

func BenchmarkPipelineExecution(b *testing.B) {
	sourceFactory := func() iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return makeIntStream(100)
	}

	steps := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepMathAdd("5"),
		evaluator.StepTake[functions.Either[error, any]](50),
	}

	report := evaluator.BenchmarkPipeline(b, "IntStreamAddAndTake", sourceFactory, steps)
	if report.ItemsProcessed == 0 {
		b.Fatalf("expected processed items in benchmark report")
	}
	fmt.Println(report.String())
}
