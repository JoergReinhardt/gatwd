package evaluator_test

import (
	"context"
	"testing"
	"time"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

func TestLargePipeline_ChainingAndPerformance(t *testing.T) {
	const totalItems = 50000
	flag := dsltypes.BitFlag(1 << 5)

	// 1. Generate 50,000 mock users lazily
	userSource := evaluator.GenerateMockUsers(totalItems, flag)

	// 2. Transducer 1: Telemetry item counter
	var telemetryCount uint64
	step1 := evaluator.StepTelemetry(&telemetryCount)

	// 3. Transducer 2: Filter (only active users)
	step2 := evaluator.StepFilter(func(val any) bool {
		user := val.(evaluator.MockUser)
		return user.Active
	})

	// 4. Transducer 3: Chunking (batch into 250-item bounded arrays for cache efficiency)
	step3 := evaluator.StepChunk(250)

	// 5. Transducer 4: Checkpointing (save watermark every 5 chunks)
	store := evaluator.NewMemoryCheckpointStore()
	step4 := evaluator.StepCheckpoint(context.Background(), store, "large_job", 5)

	// 6. Chain and execute the pipeline
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		step1,
		step2,
		step3,
		step4,
	}

	start := time.Now()
	stream := engine.EvaluatePipeline(userSource, pipeline)

	chunkCount := 0
	for range stream {
		chunkCount++
	}
	elapsed := time.Since(start)

	// In MockUser, Active is i%2 == 0, so exactly 25,000 active users
	const expectedActive = totalItems / 2
	const expectedChunks = expectedActive / 250

	if chunkCount != expectedChunks {
		t.Fatalf("expected %d chunks, got %d", expectedChunks, chunkCount)
	}
	if telemetryCount != totalItems {
		t.Fatalf("expected telemetry count %d, got %d", totalItems, telemetryCount)
	}

	watermark, err := store.Load(context.Background(), "large_job")
	if err != nil || watermark == 0 {
		t.Fatalf("expected saved watermark, got %d (err=%v)", watermark, err)
	}

	throughput := float64(totalItems) / elapsed.Seconds()
	t.Logf("Processed %d items into %d chunks in %v (%.2f items/sec)", totalItems, chunkCount, elapsed, throughput)
}
