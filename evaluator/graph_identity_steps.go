package evaluator

import (
	"iter"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// StepGraphInsertNode is the strict twin of StepGraphAddNode: it declares a node
// with the given color and identity (payload = the identity), but a graph that
// already holds that identity yields a Left error node — a
// collections.DuplicateNodeError — instead of silently overwriting the earlier
// node (Mandate 6, fail loud).
//
// Use it where a second declaration of an id is a mistake in the data, e.g. when a
// stream of records is lifted into a graph; use StepGraphAddNode where the later
// declaration is meant to update the earlier one. After an error the stream
// carries only the Left node for that graph; the graph is not modified.
func StepGraphInsertNode(color, id string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) {
						return
					}
					continue
				}
				val, _ := node.Value.UnwrapRight()
				g, ok := asGraph(val)
				if !ok {
					if !yield(node) {
						return
					}
					continue
				}
				res, err := g.InsertNode(id, color, any(id))
				var out functions.Either[error, any]
				if err != nil {
					out = functions.Left[error, any](err)
				} else {
					out = functions.Right[error, any](any(res))
				}
				if !yield(dsltypes.NewASTNode(out, node.DSLType)) {
					return
				}
			}
		}
	}
}
