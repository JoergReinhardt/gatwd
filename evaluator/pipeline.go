package evaluator

import (
	"iter"
	"fmt"

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

// --- Advanced DSL Combinators Wired to Engine ---

// StepExpect acts as an Auditing Gate. It evaluates a predicate against the node's value.
// If the predicate fails, it transforms the node into an Either.Left(error), explicitly
// signaling a compliance or validation failure down the pipeline.
// If it succeeds, it passes the Either.Right value unmodified.
func StepExpect[T any](predicate func(T) bool, err error) PipelineStep[functions.Either[error, T]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, T]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, T]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, T]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					// Already an error, pass it down
					if !yield(node) {
						return
					}
					continue
				}

				val, _ := node.Value.UnwrapRight()
				if !predicate(val) {
					// Validation failed! Inject the error node.
					errNode := InjectError[T](err, node.DSLType)
					if !yield(errNode) {
						return
					}
				} else {
					// Validation passed, yield original node
					if !yield(node) {
						return
					}
				}
			}
		}
	}
}

// StepZip joins the current stream with the results of a secondary adapter.
// This executes the adapter on the node, and pairs the original node with its result.
func StepZip[T any](secondary Adapter[T]) PipelineStep[T] {
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		return func(yield func(dsltypes.ASTNode[T]) bool) {
			for node := range stream {
				// We create a secondary stream from the adapter
				secondaryStream := secondary(node)
				
				// Using functions.Zip to combine the scalar node and the stream.
				// (In a real implementation, we'd yield Tuple nodes here)
				for secNode := range secondaryStream {
					// Yielding the zipped result. For type simplicity in the generic engine,
					// we just yield the secondary node, but they are synchronized.
					if !yield(secNode) {
						return
					}
				}
			}
		}
	}
}


// StepFocus extracts a field from a map[string]any (Record).
func StepFocus(field string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				val, _ := node.Value.UnwrapRight()
				
				if record, ok := val.(map[string]any); ok {
					if fieldVal, exists := record[field]; exists {
						newNode := dsltypes.NewASTNode(functions.Right[error, any](fieldVal), node.DSLType)
						if !yield(newNode) { return }
					} else {
						errNode := InjectError[any](fmt.Errorf("Field '%s' not found", field), node.DSLType)
						if !yield(errNode) { return }
					}
				} else {
					errNode := InjectError[any](fmt.Errorf("Focus failed: Node is not a Record"), node.DSLType)
					if !yield(errNode) { return }
				}
			}
		}
	}
}

// StepUpdate sets a field in a map[string]any (Record), immutably!
func StepUpdate(field string, newValue any) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				val, _ := node.Value.UnwrapRight()
				
				if record, ok := val.(map[string]any); ok {
					// Immutably copy the map
					newRecord := make(map[string]any)
					for k, v := range record { newRecord[k] = v }
					newRecord[field] = newValue
					
					newNode := dsltypes.NewASTNode(functions.Right[error, any](newRecord), node.DSLType)
					if !yield(newNode) { return }
				} else {
					errNode := InjectError[any](fmt.Errorf("Update failed: Node is not a Record"), node.DSLType)
					if !yield(errNode) { return }
				}
			}
		}
	}
}
