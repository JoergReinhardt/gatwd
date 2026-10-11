package evaluator_test

import (
	"context"
	"fmt"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

func ExampleEngine_EvaluatePipeline() {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)

	// Source generator: yields numbers 1 to 5
	source := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		for i := 1; i <= 5; i++ {
			if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(i)), flag)) {
				return
			}
		}
	}

	// Pipeline: add 10 to each element, take first 3
	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepMathAdd("10"),
		evaluator.StepTake[functions.Either[error, any]](3),
	}

	stream := engine.EvaluatePipeline(source, pipeline)

	for node := range stream {
		val, _ := node.Value.UnwrapRight()
		fmt.Println(val)
	}

	// Output:
	// 11
	// 12
	// 13
}

func ExampleStepCheckpoint() {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)

	source := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		for i := 1; i <= 4; i++ {
			if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(i)), flag)) {
				return
			}
		}
	}

	store := evaluator.NewMemoryCheckpointStore()
	step := evaluator.StepCheckpoint(context.Background(), store, "job_alpha", 2)

	stream := engine.EvaluatePipeline(source, []evaluator.PipelineStep[functions.Either[error, any]]{step})
	for range stream {
	}

	watermark, _ := store.Load(context.Background(), "job_alpha")
	fmt.Printf("saved watermark: %d\n", watermark)

	// Output:
	// saved watermark: 4
}
