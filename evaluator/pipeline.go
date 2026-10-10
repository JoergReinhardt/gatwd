package evaluator

import (
	"fmt"
	"iter"
	"strings"

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

// StepFold collapses a stream of nodes into a single scalar value.
// E.g., counting elements, summing values.
func StepFold(operation string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			count := 0
			var lastFlag dsltypes.BitFlag
			for node := range stream {
				if node.Value.IsLeft() { continue } // Skip errors in fold
				count++
				lastFlag = node.DSLType
			}
			
			if operation == "count" {
				val := functions.Right[error, any](fmt.Sprintf("%d", count))
				yield(dsltypes.NewASTNode(val, lastFlag))
			} else {
				errNode := InjectError[any](fmt.Errorf("Unknown fold operation: %s", operation), lastFlag)
				yield(errNode)
			}
		}
	}
}

// StepMatch routes a node to different adapters based on a regex/string match.
// It acts like an exhaustive 'case ... of' in Haskell.
func StepMatch(cases map[string]Adapter[functions.Either[error, any]], defaultAdapter Adapter[functions.Either[error, any]]) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				val, _ := node.Value.UnwrapRight()
				strVal := fmt.Sprintf("%v", val)
				
				matched := false
				for pattern, adapter := range cases {
					if strings.Contains(strVal, pattern) {
						for res := range adapter(node) {
							if !yield(res) { return }
						}
						matched = true
						break
					}
				}
				
				if !matched && defaultAdapter != nil {
					for res := range defaultAdapter(node) {
						if !yield(res) { return }
					}
				} else if !matched {
					// Discard if no match and no default
				}
			}
		}
	}
}

// StepFilter acts exactly like a MaybeT Monad Transformer in a graph.
// If the predicate fails (yielding 'Nothing'), the node is silently dropped 
// rather than being converted into an Either.Left Error.
func StepFilter(predicate func(any) bool) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}

				val, _ := node.Value.UnwrapRight()
				if predicate(val) {
					if !yield(node) { return }
				}
				// If it fails the predicate, it evaluates to Nothing and is silently dropped!
			}
		}
	}
}

// StepSplit (***) routes the Left side of a Tuple to Adapter1, and the Right side to Adapter2 concurrently.
// It acts as the Applicative Functor (<*>) over Product Types.
func StepSplit(adapter1, adapter2 Adapter[functions.Either[error, any]]) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		// For simplicity in the DSL engine wrapper, we evaluate sequentially or concurrently
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				val, _ := node.Value.UnwrapRight()
				
				if tuple, ok := val.(functions.Pair[any, any]); ok {
					// We mock the concurrent Applicative application here
					leftNode := dsltypes.NewASTNode(functions.Right[error, any](tuple.Fst), node.DSLType)
					rightNode := dsltypes.NewASTNode(functions.Right[error, any](tuple.Snd), node.DSLType)
					
					var res1, res2 any
					for r1 := range adapter1(leftNode) {
						if !r1.Value.IsLeft() { res1, _ = r1.Value.UnwrapRight(); break }
					}
					for r2 := range adapter2(rightNode) {
						if !r2.Value.IsLeft() { res2, _ = r2.Value.UnwrapRight(); break }
					}
					
					newTuple := functions.Pair[any, any]{Fst: res1, Snd: res2}
					newNode := dsltypes.NewASTNode(functions.Right[error, any](any(newTuple)), node.DSLType)
					if !yield(newNode) { return }
				} else {
					errNode := InjectError[any](fmt.Errorf("Split (***) failed: Node is not a Tuple"), node.DSLType)
					if !yield(errNode) { return }
				}
			}
		}
	}
}
