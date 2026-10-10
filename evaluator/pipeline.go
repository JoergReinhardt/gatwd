package evaluator

import (
	"fmt"
	"iter"

	"github.com/joergreinhardt/gatwd/arrows"
	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

type PipelineStep[T any] func(iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]]

func (e *Engine[T]) EvaluatePipeline(initial iter.Seq[dsltypes.ASTNode[T]], steps []PipelineStep[T]) iter.Seq[dsltypes.ASTNode[T]] {
	stream := initial
	for _, step := range steps {
		stream = step(stream)
	}
	return stream
}

func StepAdapter[T any](adapter Adapter[T]) PipelineStep[T] {
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		return functions.Bind(stream, adapter)
	}
}

func StepTake[T any](n int) PipelineStep[T] {
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		return functions.Take(stream, n)
	}
}

func StepDrop[T any](n int) PipelineStep[T] {
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		return functions.Drop(stream, n)
	}
}

// StepFanOut routes the incoming node into BOTH adapters concurrently.
// It yields a single Tuple (functions.Pair) containing both results.
func StepFanOut(adapter1, adapter2 Adapter[functions.Either[error, any]]) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		fanOutArrow := arrows.FanOut(arrows.Kleisli[dsltypes.ASTNode[functions.Either[error, any]], dsltypes.ASTNode[functions.Either[error, any]]](adapter1), arrows.Kleisli[dsltypes.ASTNode[functions.Either[error, any]], dsltypes.ASTNode[functions.Either[error, any]]](adapter2))
		
		return functions.Bind(stream, func(node dsltypes.ASTNode[functions.Either[error, any]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
			return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
				for pair := range fanOutArrow(node) {
					v1, _ := pair.Fst.Value.UnwrapRight()
					v2, _ := pair.Snd.Value.UnwrapRight()
					
					// Package the results into a Tuple
					tuple := functions.Pair[any, any]{Fst: v1, Snd: v2}
					newNode := dsltypes.NewASTNode(functions.Right[error, any](any(tuple)), node.DSLType)
					if !yield(newNode) { return }
				}
			}
		})
	}
}

// StepDistinct deduplicates a stream of nodes using the Immutable Set.
// This is critical for preventing infinite loops in cyclic graphs.
func StepDistinct() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			// Initialize the comonadic/functional state for deduplication
			visited := collections.NewSet[string]()
			
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				
				val, _ := node.Value.UnwrapRight()
				hash := fmt.Sprintf("%v", val)
				
				if !visited.Contains(hash) {
					visited = visited.Add(hash) // Immutably create new Set
					if !yield(node) { return }
				}
			}
		}
	}
}

// StepFormatTuple allows pattern matching/formatting of the Tuples yielded by FanOut.
func StepFormatTuple(formatStr string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				val, _ := node.Value.UnwrapRight()
				if tuple, ok := val.(functions.Pair[any, any]); ok {
					formatted := fmt.Sprintf(formatStr, tuple.Fst, tuple.Snd)
					newNode := dsltypes.NewASTNode(functions.Right[error, any](any(formatted)), node.DSLType)
					if !yield(newNode) { return }
				} else {
					errNode := InjectError[any](fmt.Errorf("FormatTuple failed: type %T", val), node.DSLType)
					if !yield(errNode) { return }
				}
			}
		}
	}
}

func StepExpect[T any](predicate func(T) bool, err error) PipelineStep[functions.Either[error, T]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, T]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, T]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, T]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}

				val, _ := node.Value.UnwrapRight()
				if !predicate(val) {
					errNode := InjectError[T](err, node.DSLType)
					if !yield(errNode) { return }
				} else {
					if !yield(node) { return }
				}
			}
		}
	}
}

func StepZip[T any](secondary Adapter[T]) PipelineStep[T] {
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		return func(yield func(dsltypes.ASTNode[T]) bool) {
			for node := range stream {
				secondaryStream := secondary(node)
				for secNode := range secondaryStream {
					if !yield(secNode) { return }
				}
			}
		}
	}
}

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
					newRecord := make(map[string]any)
					for k, v := range record { newRecord[k] = v }
					newRecord[field] = newValue
					
					newNode := dsltypes.NewASTNode(functions.Right[error, any](any(newRecord)), node.DSLType)
					if !yield(newNode) { return }
				} else {
					errNode := InjectError[any](fmt.Errorf("Update failed: Node is not a Record"), node.DSLType)
					if !yield(errNode) { return }
				}
			}
		}
	}
}
