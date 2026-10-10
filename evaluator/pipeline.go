package evaluator

import (
	"iter"

	"github.com/joergreinhardt/gatwd/arrows"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// PipelineStep represents a single transformation step in a graph traversal pipeline.
// It maps a stream of AST Nodes to a new stream of AST Nodes.
type PipelineStep[T any] func(iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]]

// EvaluatePipeline takes an initial stream and pipes it through multiple steps.
// This perfectly represents the `|>` operator in the DSL.
func (e *Engine[T]) EvaluatePipeline(initial iter.Seq[dsltypes.ASTNode[T]], steps []PipelineStep[T]) iter.Seq[dsltypes.ASTNode[T]] {
	stream := initial
	for _, step := range steps {
		stream = step(stream)
	}
	return stream
}

// --- Built-in Pipeline Steps ---

// StepAdapter lifts a standard Node-to-Stream Adapter into a Stream-to-Stream PipelineStep.
// It uses `functions.Bind` to flatmap the adapter across the incoming stream.
func StepAdapter[T any](adapter Adapter[T]) PipelineStep[T] {
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		return functions.Bind(stream, adapter)
	}
}

// StepTake lifts functions.Take into a PipelineStep.
func StepTake[T any](n int) PipelineStep[T] {
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		return functions.Take(stream, n)
	}
}

// StepDrop lifts functions.Drop into a PipelineStep.
func StepDrop[T any](n int) PipelineStep[T] {
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		return functions.Drop(stream, n)
	}
}

// StepFanOut lifts arrows.FanOut (&&&) into a PipelineStep.
// It routes the incoming node into BOTH adapters concurrently, and zips the results.
// The resulting node value is discarded here (or could be concatenated).
func StepFanOut[T any](adapter1, adapter2 Adapter[T]) PipelineStep[T] {
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		// Use the Kleisli Arrow FanOut!
		fanOutArrow := arrows.FanOut(arrows.Kleisli[dsltypes.ASTNode[T], dsltypes.ASTNode[T]](adapter1), arrows.Kleisli[dsltypes.ASTNode[T], dsltypes.ASTNode[T]](adapter2))
		
		// Map the fan-out over the stream
		return functions.Bind(stream, func(node dsltypes.ASTNode[T]) iter.Seq[dsltypes.ASTNode[T]] {
			return func(yield func(dsltypes.ASTNode[T]) bool) {
				for pair := range fanOutArrow(node) {
					// We yield the Left result arbitrarily for this demo, 
					// but in a real system we might construct a Tuple Node.
					if !yield(pair.Fst) {
						return
					}
					if !yield(pair.Snd) {
						return
					}
				}
			}
		})
	}
}
