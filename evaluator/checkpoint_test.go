package evaluator

import (
	"context"
	"testing"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

func TestStepCheckpoint_Continuations(t *testing.T) {
	// 1. Create a data source (0 to 99)
	source := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		for i := 0; i < 100; i++ {
			node := dsltypes.NewASTNode(functions.Right[error, any](any(i)), 0)
			if !yield(node) {
				return
			}
		}
	}

	store := NewMemoryCheckpointStore()
	checkpointID := "pipeline_job_1"
	interval := uint64(10)

	// --- RUN 1: Execute, but cancel context after 35 elements ---
	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	
	step1 := StepCheckpoint(ctx1, store, checkpointID, interval)
	stream1 := step1(source)
	
	var run1Results []int
	for node := range stream1 {
		val, _ := node.Value.UnwrapRight()
		run1Results = append(run1Results, val.(int))
		
		if len(run1Results) == 35 {
			// Trigger a pause! This will be caught at the next checkpoint boundary (which is at 40).
			// Wait, the context check happens AT the checkpoint interval!
			// So it will process up to 40, save 40, and then exit.
			cancel1()
		}
	}
	
	if len(run1Results) != 40 {
		t.Fatalf("Expected Run 1 to stop at 40 elements due to interval cancellation, got %d", len(run1Results))
	}
	
	offset, _ := store.Load(context.Background(), checkpointID)
	if offset != 40 {
		t.Fatalf("Expected saved watermark to be 40, got %d", offset)
	}

	// --- RUN 2: Resume from the watermark (40) ---
	ctx2 := context.Background()
	step2 := StepCheckpoint(ctx2, store, checkpointID, interval)
	stream2 := step2(source) // the source starts from 0 again!
	
	var run2Results []int
	for node := range stream2 {
		val, _ := node.Value.UnwrapRight()
		run2Results = append(run2Results, val.(int))
	}
	
	if len(run2Results) != 60 {
		t.Fatalf("Expected Run 2 to process remaining 60 elements, got %d", len(run2Results))
	}
	
	// The first element of run 2 should be 40!
	if run2Results[0] != 40 {
		t.Fatalf("Expected Run 2 to start at element 40, got %d", run2Results[0])
	}
	
	finalOffset, _ := store.Load(context.Background(), checkpointID)
	if finalOffset != 100 {
		t.Fatalf("Expected final watermark to be 100, got %d", finalOffset)
	}
}
