package evaluator_test

import (
	"context"
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func ExampleEngine_EvaluatePipeline() {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)
	dsltypes.RegisterTypeFlag(flag, "Metric")

	// Source generator: yields numbers
	source := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		yield(dsltypes.NewASTNode(functions.Right[error, any](any(1)), flag))
	}

	// Pipeline: add 10 to element, visualize output
	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepMathAdd("10"),
		visualizer.StepVisualize("--curved --border"),
	}

	stream := engine.EvaluatePipeline(source, pipeline)
	for node := range stream {
		val, _ := node.Value.UnwrapRight()
		fmt.Print(val.(string))
	}

	// Output:
	// ╭───────────────╮
	// │ ╰─ Metric: 11 │
	// ╰───────────────╯
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

	type CheckpointReport struct {
		Job       string
		Watermark uint64
	}
	report := CheckpointReport{Job: "job_alpha", Watermark: watermark}
	fmt.Print(visualizer.Format(report, visualizer.ParseOptions("--curved --border")))

	// Output:
	// ╭───────────────────────╮
	// │ ╰─ CheckpointReport   │
	// │     ├─ Job: job_alpha │
	// │     ╰─ Watermark: 4   │
	// ╰───────────────────────╯
}

func ExampleStepChunk() {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)

	source := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		yield(dsltypes.NewASTNode(functions.Right[error, any](any(10)), flag))
		yield(dsltypes.NewASTNode(functions.Right[error, any](any(20)), flag))
	}

	// Batch items into chunks of 2
	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepChunk(2),
	}

	stream := engine.EvaluatePipeline(source, pipeline)
	opts := visualizer.ParseOptions("--curved --border")
	for node := range stream {
		val, _ := node.Value.UnwrapRight()
		chunk := val.(*collections.Chunk[dsltypes.ASTNode[functions.Either[error, any]]])
		fmt.Print(visualizer.Format(chunk, opts))
		break
	}

	// Output:
	// ╭────────────────╮
	// │ ╰─ List        │
	// │     ├─ [0]: 10 │
	// │     ╰─ [1]: 20 │
	// ╰────────────────╯
}

func ExampleGenerateMockUsers() {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)

	// Generate mock users and filter for active
	stream := evaluator.GenerateMockUsers(2, flag)
	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepFilter(func(val any) bool {
			return val.(evaluator.MockUser).Active
		}),
	}

	opts := visualizer.ParseOptions("--curved --border")
	for node := range engine.EvaluatePipeline(stream, pipeline) {
		val, _ := node.Value.UnwrapRight()
		fmt.Print(visualizer.Format(val, opts))
	}

	// Output:
	// ╭──────────────────────╮
	// │ ╰─ MockUser          │
	// │     ├─ ID: 0         │
	// │     ├─ Name: admin_A │
	// │     ├─ Role: admin   │
	// │     ├─ Score: 0      │
	// │     ╰─ Active: true  │
	// ╰──────────────────────╯
}

func ExampleRunExample() {
	// Programmatic example test execution
	ex := evaluator.Example{
		Name: "DemoExample",
		F: func() {
			fmt.Println("foo")
		},
		Output: "foo",
	}

	passed, _, _ := evaluator.RunExample(ex)

	type SuiteSummary struct {
		Suite  string
		Passed bool
	}
	fmt.Print(visualizer.Format(SuiteSummary{Suite: ex.Name, Passed: passed}, visualizer.ParseOptions("--curved --border")))

	// Output:
	// ╭───────────────────────────╮
	// │ ╰─ SuiteSummary           │
	// │     ├─ Suite: DemoExample │
	// │     ╰─ Passed: true       │
	// ╰───────────────────────────╯
}
