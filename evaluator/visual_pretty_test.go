package evaluator_test

import (
	"context"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func TestVisualPretty_MockUsersStream(t *testing.T) {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)

	// 1. Generate lazy stream of mock users
	stream := evaluator.GenerateMockUsers(4, flag)

	// 2. Filter for active users
	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepFilter(func(val any) bool {
			user := val.(evaluator.MockUser)
			return user.Active
		}),
	}

	opts := visualizer.ParseOptions("--curved --border --color-type=MockUser:cyan")
	t.Log("=== PRETTY-PRINTED ACTIVE MOCK USERS ===")
	for node := range engine.EvaluatePipeline(stream, pipeline) {
		val, _ := node.Value.UnwrapRight()
		card := visualizer.Format(val, opts)
		t.Logf("\n%s", card)
	}
}

func TestVisualPretty_MockUserChunking(t *testing.T) {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)

	// Stream 5 mock users through a batching transducer (Chunk of 2)
	stream := evaluator.GenerateMockUsers(5, flag)
	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepChunk(2),
	}

	opts := visualizer.ParseOptions("--curved --border --box-all")
	t.Log("=== PRETTY-PRINTED CHUNKS OF MOCK DATA ===")
	chunkIdx := 0
	for node := range engine.EvaluatePipeline(stream, pipeline) {
		val, _ := node.Value.UnwrapRight()
		chunk := val.(*collections.Chunk[dsltypes.ASTNode[functions.Either[error, any]]])

		type ChunkSummary struct {
			BatchNumber int
			Capacity    int
			ItemCount   int
			Items       []evaluator.MockUser
		}

		var users []evaluator.MockUser
		for i := 0; i < chunk.Len(); i++ {
			itemNode, _ := chunk.Get(i)
			user, _ := itemNode.Value.UnwrapRight()
			users = append(users, user.(evaluator.MockUser))
		}

		summary := ChunkSummary{
			BatchNumber: chunkIdx,
			Capacity:    chunk.Capacity(),
			ItemCount:   chunk.Len(),
			Items:       users,
		}
		chunkIdx++

		card := visualizer.Format(summary, opts)
		t.Logf("\n%s", card)
	}
}

func TestVisualPretty_PipelineStepVisualize(t *testing.T) {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)

	// Stream mock users through the visualizer step transducer
	stream := evaluator.GenerateMockUsers(3, flag)
	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		visualizer.StepVisualize("--curved --border"),
	}

	t.Log("=== STREAM PROCESSED THROUGH StepVisualize TRANSDUCER ===")
	for node := range engine.EvaluatePipeline(stream, pipeline) {
		val, _ := node.Value.UnwrapRight()
		t.Logf("Evaluator Stream Node:%s", val.(string))
	}
}

func TestVisualPretty_TelemetryDashboard(t *testing.T) {
	type TelemetryDashboard struct {
		PipelineName     string
		TotalEvents      int
		ProcessedBatches int
		AvgLatencyMs     float64
		Status           string
	}

	dashboard := TelemetryDashboard{
		PipelineName:     "gatwd-stream-analytics",
		TotalEvents:      50000,
		ProcessedBatches: 2500,
		AvgLatencyMs:     0.18,
		Status:           "HEALTHY",
	}

	opts := visualizer.ParseOptions("--heavy --border --color=green")
	card := visualizer.Format(dashboard, opts)
	t.Logf("=== PIPELINE TELEMETRY DASHBOARD ===\n%s", card)
}

func TestVisualPretty_CheckpointedWorkflow(t *testing.T) {
	ctx := context.Background()
	store := evaluator.NewMemoryCheckpointStore()
	flag := dsltypes.BitFlag(1)

	// Stream 4 mock users with checkpointing at batch interval 2
	stream := evaluator.GenerateMockUsers(4, flag)
	stepCheck := evaluator.StepCheckpoint(ctx, store, "visual_job", 2)

	engine := evaluator.NewEngine[functions.Either[error, any]]()
	for range engine.EvaluatePipeline(stream, []evaluator.PipelineStep[functions.Either[error, any]]{stepCheck}) {
	}

	watermark, _ := store.Load(ctx, "visual_job")

	type CheckpointStatus struct {
		JobID       string
		Watermark   uint64
		Persistence string
		Committed   bool
	}

	status := CheckpointStatus{
		JobID:       "visual_job",
		Watermark:   watermark,
		Persistence: "In-Memory Atomic Store",
		Committed:   true,
	}

	opts := visualizer.ParseOptions("--curved --border")
	t.Logf("=== RESUMABLE CHECKPOINT METADATA ===\n%s", visualizer.Format(status, opts))
}
