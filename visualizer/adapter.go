package visualizer

import (
	"iter"
	"strings"
	"strconv"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// ParseOptions converts a string like "--ascii --collapse --depth=2" into Options.
func ParseOptions(args string) Options {
	opts := DefaultOptions()
	parts := strings.Fields(args)
	
	for _, p := range parts {
		if p == "--ascii" {
			opts.Style = ASCIIStyle
		} else if p == "--collapse" {
			opts.CollapseScalars = true
		} else if strings.HasPrefix(p, "--depth=") {
			dStr := strings.TrimPrefix(p, "--depth=")
			if d, err := strconv.Atoi(dStr); err == nil {
				opts.MaxDepth = d
			}
		}
	}
	return opts
}

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
