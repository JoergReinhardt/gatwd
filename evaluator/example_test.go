package evaluator_test

import (
	"context"
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
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

func ExampleStepChunk() {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)

	source := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		for i := 1; i <= 5; i++ {
			if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(i)), flag)) {
				return
			}
		}
	}

	// Batch items into chunks of 2
	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepChunk(2),
	}

	stream := engine.EvaluatePipeline(source, pipeline)
	for node := range stream {
		val, _ := node.Value.UnwrapRight()
		chunk := val.(*collections.Chunk[dsltypes.ASTNode[functions.Either[error, any]]])
		fmt.Printf("chunk size %d\n", chunk.Len())
	}

	// Output:
	// chunk size 2
	// chunk size 2
	// chunk size 1
}

func ExampleGenerateMockUsers() {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)

	// Generate 4 mock users
	stream := evaluator.GenerateMockUsers(4, flag)

	// Filter for active users only
	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepFilter(func(val any) bool {
			user := val.(evaluator.MockUser)
			return user.Active
		}),
	}

	for node := range engine.EvaluatePipeline(stream, pipeline) {
		val, _ := node.Value.UnwrapRight()
		user := val.(evaluator.MockUser)
		fmt.Printf("User %d: %s (%s)\n", user.ID, user.Name, user.Role)
	}

	// Output:
	// User 0: admin_A (admin)
	// User 2: viewer_C (viewer)
}

func ExampleRunExample() {
	// Programmatic example test execution
	ex := evaluator.Example{
		Name: "DemoExample",
		F: func() {
			fmt.Println("foo")
			fmt.Println("bar")
		},
		Output: "foo\nbar",
	}

	passed, captured, _ := evaluator.RunExample(ex)
	fmt.Printf("passed: %t, captured lines: %t\n", passed, len(captured) > 0)

	// Output:
	// passed: true, captured lines: true
}

func ExampleRunExample_unordered() {
	// Permutation-invariant matching (Unordered = true)
	ex := evaluator.Example{
		Name: "UnorderedDemo",
		F: func() {
			fmt.Println("z")
			fmt.Println("a")
		},
		Output:    "a\nz",
		Unordered: true,
	}

	passed, _, _ := evaluator.RunExample(ex)
	fmt.Printf("unordered match passed: %t\n", passed)

	// Output:
	// unordered match passed: true
}
