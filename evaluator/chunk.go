package evaluator

import (
	"iter"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// StepChunk is a transducer that takes a lazy iter.Seq and batches the incoming
// nodes into eager Chunks of the specified size.
// It yields a stream of ASTNodes where each successfully batched chunk is
// wrapped in a functions.Either[error, any]. If an error node is encountered,
// the current chunk (if any) is yielded first, followed by the error node.
func StepChunk(size int) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			chunk := collections.NewChunk[dsltypes.ASTNode[functions.Either[error, any]]](size)

			for node := range stream {
				if node.Value.IsLeft() {
					// Yield any accumulated chunk before the error
					if chunk.Len() > 0 {
						newNode := dsltypes.NewASTNode(functions.Right[error, any](any(chunk)), node.DSLType) // Use type from the current node or what makes sense
						if !yield(newNode) {
							return
						}
						chunk = collections.NewChunk[dsltypes.ASTNode[functions.Either[error, any]]](size)
					}
					// Yield the error
					if !yield(node) {
						return
					}
					continue
				}

				chunk.Append(node)
				if chunk.IsFull() {
					newNode := dsltypes.NewASTNode(functions.Right[error, any](any(chunk)), node.DSLType)
					if !yield(newNode) {
						return
					}
					chunk = collections.NewChunk[dsltypes.ASTNode[functions.Either[error, any]]](size)
				}
			}

			// Yield remaining chunk if it's not empty
			if chunk.Len() > 0 {
				// We need a dummy type or type from the first element of chunk.
				// Since we don't track it cleanly, we can use the type of the first node.
				items := chunk.Items()
				firstType := items[0].DSLType
				newNode := dsltypes.NewASTNode(functions.Right[error, any](any(chunk)), firstType)
				yield(newNode)
			}
		}
	}
}

// StepUnchunk unpacks incoming Chunk nodes back into a flat stream of their individual constituent nodes.
// If an element is not a Chunk, it passes through untouched. Error nodes pass through cleanly.
func StepUnchunk() PipelineStep[functions.Either[error, any]] {
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
				if chunk, ok := val.(*collections.Chunk[dsltypes.ASTNode[functions.Either[error, any]]]); ok {
					for _, item := range chunk.Items() {
						if !yield(item) {
							return
						}
					}
				} else {
					if !yield(node) {
						return
					}
				}
			}
		}
	}
}
