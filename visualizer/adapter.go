package visualizer

import (
	"iter"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// ParseOptions converts a string like "--ascii --collapse --depth=2" into Options.

// StepVisualize is a pipeline step that transforms any complex structural Node 
// into a pretty-printed UTF-8 string graph.
func StepVisualize(args string) evaluator.PipelineStep[functions.Either[error, any]] {
	opts := ParseOptions(args)
	
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				
				val, _ := node.Value.UnwrapRight()
				graphStr := "\n" + Format(val, opts)
				
				newNode := dsltypes.NewASTNode(functions.Right[error, any](any(graphStr)), node.DSLType)
				if !yield(newNode) { return }
			}
		}
	}
}
