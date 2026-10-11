package evaluator

import (
	"fmt"
	"iter"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/joergreinhardt/gatwd/comonads"

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
					errNode := InjectError[any](fmt.Errorf("Type Error: format_tuple expected a Tuple (Pair), but received %T", val), node.DSLType)
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
						errNode := InjectError[any](fmt.Errorf("Lens Error: Cannot focus on field '%s' because it does not exist in this Record", field), node.DSLType)
						if !yield(errNode) { return }
					}
				} else {
					errNode := InjectError[any](fmt.Errorf("Lens Error: Cannot focus because Node is not a Record (Map)"), node.DSLType)
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
					errNode := InjectError[any](fmt.Errorf("Lens Error: Cannot update because Node is not a Record (Map)"), node.DSLType)
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
					errNode := InjectError[any](fmt.Errorf("Applicative Error: Split [***] expected a Tuple (Pair), but received a scalar value"), node.DSLType)
					if !yield(errNode) { return }
				}
			}
		}
	}
}

// StepCollectList aggregates the entire stream into a single collections.List
func StepCollectList() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			var elements []any
			var lastFlag dsltypes.BitFlag
			for node := range stream {
				if node.Value.IsLeft() { continue }
				val, _ := node.Value.UnwrapRight()
				elements = append(elements, val)
				lastFlag = node.DSLType
			}
			
			// Build immutable list backwards
			var list *collections.List[any] = collections.Nil[any]()
			for i := len(elements) - 1; i >= 0; i-- {
				list = collections.Cons(elements[i], list)
			}
			
			newNode := dsltypes.NewASTNode(functions.Right[error, any](any(list)), lastFlag)
			yield(newNode)
		}
	}
}

// StepBuildTree takes a stream (or a list) and builds a simple Tree with a Root node and the stream elements as children.
func StepBuildTree(rootName string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			var elements []any
			var lastFlag dsltypes.BitFlag
			for node := range stream {
				if node.Value.IsLeft() { continue }
				val, _ := node.Value.UnwrapRight()
				elements = append(elements, val)
				lastFlag = node.DSLType
			}
			
			// Build children list
			var children *collections.List[*collections.Tree[any]] = collections.Nil[*collections.Tree[any]]()
			for i := len(elements) - 1; i >= 0; i-- {
				childTree := collections.Leaf[any](elements[i])
				children = collections.Cons(childTree, children)
			}
			
			tree := collections.NewTree[any](any(rootName), children)
			newNode := dsltypes.NewASTNode(functions.Right[error, any](any(tree)), lastFlag)
			yield(newNode)
		}
	}
}


// StepZipper converts a Tree into a Zipper, or extracts from a Zipper.
func StepZipper(op string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				
				val, _ := node.Value.UnwrapRight()
				var zipper *comonads.TreeZipper[any]
				
				// Initialize Zipper if it's a Tree
				if tree, ok := val.(*collections.Tree[any]); ok {
					zipper = comonads.NewZipper(tree)
				} else if z, ok := val.(*comonads.TreeZipper[any]); ok {
					zipper = z
				} else {
					errNode := InjectError[any](fmt.Errorf("Comonad Error: Cannot navigate because Node is neither a Tree nor a Zipper Cursor"), node.DSLType)
					if !yield(errNode) { return }
					continue
				}
				
				// Apply Operation
				if op == "into_zipper" {
					newNode := dsltypes.NewASTNode(functions.Right[error, any](any(zipper)), node.DSLType)
					if !yield(newNode) { return }
				} else if op == "down" {
					nextZipper := zipper.Down()
					if nextZipper != nil {
						newNode := dsltypes.NewASTNode(functions.Right[error, any](any(nextZipper)), node.DSLType)
						if !yield(newNode) { return }
					} else {
						errNode := InjectError[any](fmt.Errorf("Cursor Error: Cannot move Zipper DOWN because the current Tree node has no children"), node.DSLType)
						if !yield(errNode) { return }
					}
				} else if op == "right" {
						nextZipper := zipper.Right()
						if nextZipper != nil {
							newNode := dsltypes.NewASTNode(functions.Right[error, any](any(nextZipper)), node.DSLType)
							if !yield(newNode) { return }
						} else {
							errNode := InjectError[any](fmt.Errorf("Cursor Error: Cannot move Zipper RIGHT because there is no sibling"), node.DSLType)
							if !yield(errNode) { return }
						}
					} else if op == "left" {
						nextZipper := zipper.Left()
						if nextZipper != nil {
							newNode := dsltypes.NewASTNode(functions.Right[error, any](any(nextZipper)), node.DSLType)
							if !yield(newNode) { return }
						} else {
							errNode := InjectError[any](fmt.Errorf("Cursor Error: Cannot move Zipper LEFT because there is no sibling"), node.DSLType)
							if !yield(errNode) { return }
						}
					} else if op == "up" {
					nextZipper := zipper.Up()
					if nextZipper != nil {
						newNode := dsltypes.NewASTNode(functions.Right[error, any](any(nextZipper)), node.DSLType)
						if !yield(newNode) { return }
					} else {
						errNode := InjectError[any](fmt.Errorf("Cursor Error: Cannot move Zipper UP because the Cursor is already at the Root of the Tree"), node.DSLType)
						if !yield(errNode) { return }
					}
				} else if op == "extract" {
					extracted := zipper.Extract()
					newNode := dsltypes.NewASTNode(functions.Right[error, any](extracted), node.DSLType)
					if !yield(newNode) { return }
				}
			}
		}
	}
}

// StepRoute delegates evaluation to the Engine's internal Type-Based routing registry.
// This allows the DSL to defer to natively registered Go adapters (e.g., engine.Register(Flag, Adapter))
// based strictly on the ASTNode's Algebraic Type (BitFlag).
func StepRoute[T any](engine *Engine[T]) PipelineStep[T] {
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		return engine.Evaluate(stream)
	}
}

// StepBuildBST aggregates a stream of comparable strings into a Binary Search Tree
func StepBuildBST() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			var bst *collections.BST[string]
			var lastFlag dsltypes.BitFlag
			for node := range stream {
				if node.Value.IsLeft() { continue }
				val, _ := node.Value.UnwrapRight()
				strVal := fmt.Sprintf("%v", val)
				lastFlag = node.DSLType
				
				cmp := func(a, b string) int { return strings.Compare(a, b) }
				bst = bst.Insert(strVal, cmp)
			}
			newNode := dsltypes.NewASTNode(functions.Right[error, any](any(bst)), lastFlag)
			yield(newNode)
		}
	}
}

// StepBuildBTree aggregates a stream into a mock B-Tree with 3 keys per node
func StepBuildBTree() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			var elements []any
			var lastFlag dsltypes.BitFlag
			for node := range stream {
				if node.Value.IsLeft() { continue }
				val, _ := node.Value.UnwrapRight()
				elements = append(elements, val)
				lastFlag = node.DSLType
			}
			
			// Mock a B-Tree structure for visualizer testing
			var btree *collections.BTree[any]
			if len(elements) > 3 {
				c1 := collections.NewBTreeNode(elements[0:2], nil)
				c2 := collections.NewBTreeNode(elements[2:4], nil)
				btree = collections.NewBTreeNode([]any{"RootKey"}, []*collections.BTree[any]{c1, c2})
			} else {
				btree = collections.NewBTreeNode(elements, nil)
			}
			
			newNode := dsltypes.NewASTNode(functions.Right[error, any](any(btree)), lastFlag)
			yield(newNode)
		}
	}
}

// StepBuildRBTree aggregates a stream into a Red-Black Tree
func StepBuildRBTree() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			var rbt *collections.RBTree[string]
			var lastFlag dsltypes.BitFlag
			for node := range stream {
				if node.Value.IsLeft() { continue }
				val, _ := node.Value.UnwrapRight()
				strVal := fmt.Sprintf("%v", val)
				lastFlag = node.DSLType
				
				cmp := func(a, b string) int { return strings.Compare(a, b) }
				if rbt == nil {
					rbt = collections.LeafRB(strVal)
				} else {
					rbt = rbt.Insert(strVal, cmp)
				}
			}
			newNode := dsltypes.NewASTNode(functions.Right[error, any](any(rbt)), lastFlag)
			yield(newNode)
		}
	}
}

func StepMathAdd(amount string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() { if !yield(node) { return }; continue }
				
				val, _ := node.Value.UnwrapRight()
				var newVal any
				
				switch v := val.(type) {
				case *big.Int:
					n := new(big.Int)
					n.SetString(amount, 10)
					res := new(big.Int).Add(v, n)
					newVal = res
				case int:
					n, _ := strconv.Atoi(amount)
					newVal = v + n
				case float64:
					n, _ := strconv.ParseFloat(amount, 64)
					newVal = v + n
				case string:
					if iv, err := strconv.Atoi(v); err == nil {
						n, _ := strconv.Atoi(amount)
						newVal = iv + n
					} else if fv, err := strconv.ParseFloat(v, 64); err == nil {
						n, _ := strconv.ParseFloat(amount, 64)
						newVal = fv + n
					} else {
						newVal = v + amount
					}
				default:
					errNode := InjectError[any](fmt.Errorf("math_add failed: unsupported type %T", val), node.DSLType)
					if !yield(errNode) { return }
					continue
				}
				
				newNode := dsltypes.NewASTNode(functions.Right[error, any](newVal), node.DSLType)
				if !yield(newNode) { return }
			}
		}
	}
}

// StepBuildArray aggregates a stream into a native Go slice
func StepBuildArray() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			var arr []any
			var lastFlag dsltypes.BitFlag
			for node := range stream {
				if node.Value.IsLeft() { continue }
				val, _ := node.Value.UnwrapRight()
				arr = append(arr, val)
				lastFlag = node.DSLType
			}
			yield(dsltypes.NewASTNode(functions.Right[error, any](any(arr)), lastFlag))
		}
	}
}

// StepMap applies a sub-pipeline to each individual node in the stream (FlatMap/Monadic Bind)
func StepMap(subPipeline []PipelineStep[functions.Either[error, any]], engine *Engine[functions.Either[error, any]]) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() { 
					if !yield(node) { return }
					continue 
				}

				// Create a singleton stream for this node
				singleton := func(subYield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
					subYield(node)
				}
				
				// Evaluate the sub-pipeline over this single node
				results := engine.EvaluatePipeline(singleton, subPipeline)
				
				// Yield all results from the sub-pipeline downstream
				for resNode := range results {
					if !yield(resNode) { return }
				}
			}
		}
	}
}

// StepSort buffers the entire lazy sequence into a slice, sorts it using a provided comparator, 
// and then yields it back as a lazy sequence. This represents a strict evaluation boundary.
func StepSort(less func(a, b any) bool) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			var buffer []dsltypes.ASTNode[functions.Either[error, any]]
			var errs []dsltypes.ASTNode[functions.Either[error, any]]

			// Drain the stream
			for node := range stream {
				if node.Value.IsLeft() {
					errs = append(errs, node)
					continue
				}
				buffer = append(buffer, node)
			}

			// Yield errors immediately
			for _, errNode := range errs {
				if !yield(errNode) { return }
			}

			// Sort the buffer
			sort.SliceStable(buffer, func(i, j int) bool {
				valI, _ := buffer[i].Value.UnwrapRight()
				valJ, _ := buffer[j].Value.UnwrapRight()
				return less(valI, valJ)
			})

			// Yield the sorted buffer lazily
			for _, node := range buffer {
				if !yield(node) { return }
			}
		}
	}
}

// Default numeric/string sorting comparator
func DefaultLess(a, b any) bool {
	switch va := a.(type) {
	case int:
		if vb, ok := b.(int); ok { return va < vb }
	case float64:
		if vb, ok := b.(float64); ok { return va < vb }
	case string:
		if vb, ok := b.(string); ok { return va < vb }
	}
	// Fallback to string comparison
	return fmt.Sprintf("%v", a) < fmt.Sprintf("%v", b)
}

// StepToSet aggregates the entire stream into a deduplicated Set and yields the Set downstream.
func StepToSet() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			set := collections.NewSet[string]()
			var errs []dsltypes.ASTNode[functions.Either[error, any]]
			var lastFlag dsltypes.BitFlag

			for node := range stream {
				if node.Value.IsLeft() {
					errs = append(errs, node)
					continue
				}
				val, _ := node.Value.UnwrapRight()
				set = set.Add(fmt.Sprintf("%v", val))
				lastFlag = node.DSLType
			}

			for _, errNode := range errs {
				if !yield(errNode) { return }
			}

			newNode := dsltypes.NewASTNode(functions.Right[error, any](any(set)), lastFlag)
			yield(newNode)
		}
	}
}

// StepToHAMT folds key-value pairs (or "k:v" strings) from the stream into an immutable HAMT map.
func StepToHAMT() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			hamt := collections.NewHAMT[string, any]()
			var errs []dsltypes.ASTNode[functions.Either[error, any]]
			var lastFlag dsltypes.BitFlag

			for node := range stream {
				if node.Value.IsLeft() {
					errs = append(errs, node)
					continue
				}
				val, _ := node.Value.UnwrapRight()
				lastFlag = node.DSLType

				// Check if val is a Pair
				if p, ok := val.(functions.Pair[string, any]); ok {
					hamt = hamt.Set(p.Fst, p.Snd)
				} else if str, ok := val.(string); ok {
					// Split on ":" or "="
					var k, v string
					if strings.Contains(str, ":") {
						parts := strings.SplitN(str, ":", 2)
						k, v = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
					} else if strings.Contains(str, "=") {
						parts := strings.SplitN(str, "=", 2)
						k, v = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
					} else {
						k, v = str, str
					}
					hamt = hamt.Set(k, v)
				} else {
					hamt = hamt.Set(fmt.Sprintf("%v", val), val)
				}
			}

			for _, errNode := range errs {
				if !yield(errNode) { return }
			}

			newNode := dsltypes.NewASTNode(functions.Right[error, any](any(hamt)), lastFlag)
			yield(newNode)
		}
	}
}

// StepHAMTGet extracts a value by key from a flowing HAMT or Map node.
func StepHAMTGet(key string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				val, _ := node.Value.UnwrapRight()
				if h, ok := val.(*collections.HAMT[string, any]); ok {
					if res, found := h.Get(key); found {
						yield(dsltypes.NewASTNode(functions.Right[error, any](res), node.DSLType))
					} else {
						yield(InjectError[any](fmt.Errorf("key '%s' not found in HAMT", key), node.DSLType))
					}
					continue
				}
				if m, ok := val.(*collections.Map[string, any]); ok {
					if res, found := m.Get(key); found {
						yield(dsltypes.NewASTNode(functions.Right[error, any](res), node.DSLType))
					} else {
						yield(InjectError[any](fmt.Errorf("key '%s' not found in Map", key), node.DSLType))
					}
					continue
				}
				// Pass through if not a HAMT/Map
				yield(node)
			}
		}
	}
}

// StepHAMTSet associates key with val in an incoming HAMT or Map node, yielding the updated persistent collection.
func StepHAMTSet(key string, val any) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				v, _ := node.Value.UnwrapRight()
				if h, ok := v.(*collections.HAMT[string, any]); ok {
					yield(dsltypes.NewASTNode(functions.Right[error, any](any(h.Set(key, val))), node.DSLType))
					continue
				}
				if m, ok := v.(*collections.Map[string, any]); ok {
					yield(dsltypes.NewASTNode(functions.Right[error, any](any(m.Set(key, val))), node.DSLType))
					continue
				}
				yield(node)
			}
		}
	}
}

// StepHAMTDel deletes key from an incoming HAMT or Map node, yielding the updated persistent collection.
func StepHAMTDel(key string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				v, _ := node.Value.UnwrapRight()
				if h, ok := v.(*collections.HAMT[string, any]); ok {
					yield(dsltypes.NewASTNode(functions.Right[error, any](any(h.Delete(key))), node.DSLType))
					continue
				}
				if m, ok := v.(*collections.Map[string, any]); ok {
					yield(dsltypes.NewASTNode(functions.Right[error, any](any(m.Delete(key))), node.DSLType))
					continue
				}
				yield(node)
			}
		}
	}
}

// StepHAMTKeys unfolds an incoming HAMT or Map into a stream of its keys.
func StepHAMTKeys() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				v, _ := node.Value.UnwrapRight()
				if h, ok := v.(*collections.HAMT[string, any]); ok {
					for k := range h.Seq() {
						if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(k)), node.DSLType)) {
							return
						}
					}
					continue
				}
				if m, ok := v.(*collections.Map[string, any]); ok {
					for k := range m.Seq() {
						if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(k)), node.DSLType)) {
							return
						}
					}
					continue
				}
				yield(node)
			}
		}
	}
}

// StepHAMTValues unfolds an incoming HAMT or Map into a stream of its values.
func StepHAMTValues() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				v, _ := node.Value.UnwrapRight()
				if h, ok := v.(*collections.HAMT[string, any]); ok {
					for _, val := range h.Seq() {
						if !yield(dsltypes.NewASTNode(functions.Right[error, any](val), node.DSLType)) {
							return
						}
					}
					continue
				}
				if m, ok := v.(*collections.Map[string, any]); ok {
					for _, val := range m.Seq() {
						if !yield(dsltypes.NewASTNode(functions.Right[error, any](val), node.DSLType)) {
							return
						}
					}
					continue
				}
				yield(node)
			}
		}
	}
}

// StepSetContains checks whether an incoming Set contains the specified item, yielding a boolean ASTNode.
func StepSetContains(item string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				v, _ := node.Value.UnwrapRight()
				if s, ok := v.(*collections.Set[string]); ok {
					yield(dsltypes.NewASTNode(functions.Right[error, any](any(s.Contains(item))), node.DSLType))
					continue
				}
				yield(node)
			}
		}
	}
}

// StepSetAdd adds item to an incoming Set node, yielding the updated persistent Set.
func StepSetAdd(item string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				v, _ := node.Value.UnwrapRight()
				if s, ok := v.(*collections.Set[string]); ok {
					yield(dsltypes.NewASTNode(functions.Right[error, any](any(s.Add(item))), node.DSLType))
					continue
				}
				yield(node)
			}
		}
	}
}

// StepSetRemove removes item from an incoming Set node, yielding the updated persistent Set.
func StepSetRemove(item string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				v, _ := node.Value.UnwrapRight()
				if s, ok := v.(*collections.Set[string]); ok {
					yield(dsltypes.NewASTNode(functions.Right[error, any](any(s.Remove(item))), node.DSLType))
					continue
				}
				yield(node)
			}
		}
	}
}

// StepSetElements unfolds an incoming Set into a stream of its constituent elements.
func StepSetElements() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				v, _ := node.Value.UnwrapRight()
				if s, ok := v.(*collections.Set[string]); ok {
					for elem := range s.Seq() {
						if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(elem)), node.DSLType)) {
							return
						}
					}
					continue
				}
				yield(node)
			}
		}
	}
}


// StepEmit emits a fixed set of string values sequentially into the stream.
func StepEmit(values ...string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			// First, pass through upstream
			for node := range stream {
				if !yield(node) { return }
			}
			
			// Then emit our values
			for _, v := range values {
				node := dsltypes.NewASTNode(functions.Right[error, any](any(v)), 0)
				if !yield(node) { return }
			}
		}
	}
}

// StepAssert checks if the incoming node's string representation matches the expected string.
// If it fails, it yields an error node. If it passes, it yields the original node unaltered.
func StepAssert(expected string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) { return }
					continue
				}
				
				val, _ := node.Value.UnwrapRight()
				var actual string
				if b, ok := val.([]byte); ok {
					actual = string(b)
				} else {
					actual = fmt.Sprintf("%v", val)
				}
				
				if actual != expected {
					err := fmt.Errorf("Assertion failed: expected '%s', got '%s'", expected, actual)
					errNode := InjectError[any](err, node.DSLType)
					if !yield(errNode) { return }
				} else {
					if !yield(node) { return }
				}
			}
		}
	}
}

func StepMathSub(amount string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() { if !yield(node) { return }; continue }
				
				val, _ := node.Value.UnwrapRight()
				var newVal any
				
				switch v := val.(type) {
				case int:
					n, _ := strconv.Atoi(amount)
					newVal = v - n
				case float64:
					n, _ := strconv.ParseFloat(amount, 64)
					newVal = v - n
				case string:
					if iv, err := strconv.Atoi(v); err == nil {
						n, _ := strconv.Atoi(amount)
						newVal = iv - n
					} else if fv, err := strconv.ParseFloat(v, 64); err == nil {
						n, _ := strconv.ParseFloat(amount, 64)
						newVal = fv - n
					} else {
						errNode := InjectError[any](fmt.Errorf("math_sub failed: cannot subtract from non-numeric string %q", v), node.DSLType)
						if !yield(errNode) { return }
						continue
					}
				default:
					errNode := InjectError[any](fmt.Errorf("math_sub failed: unsupported type %T", val), node.DSLType)
					if !yield(errNode) { return }
					continue
				}
				
				newNode := dsltypes.NewASTNode(functions.Right[error, any](newVal), node.DSLType)
				if !yield(newNode) { return }
			}
		}
	}
}

func StepDebugPrint(prefix string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					err, _ := node.Value.UnwrapLeft()
					fmt.Println(prefix, "LEFT:", err)
				} else {
					val, _ := node.Value.UnwrapRight()
					fmt.Println(prefix, "RIGHT:", val)
				}
				if !yield(node) { return }
			}
		}
	}
}

func StepYieldRuntimeNode(node dsltypes.ASTNode[functions.Either[error, any]]) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			// Ignore the incoming stream, just yield the runtime node.
			// But if it's a List, emit its contents!
			if !node.Value.IsLeft() {
				val, _ := node.Value.UnwrapRight()
				if list, ok := val.(*collections.List[any]); ok {
					current := list
					for !current.IsEmpty() {
						head, _ := current.Head()
						newNode := dsltypes.NewASTNode(functions.Right[error, any](head), node.DSLType)
						if !yield(newNode) { return }
						current, _ = current.Tail()
					}
					return
				}
			}
			yield(node)
		}
	}
}

func StepFilterLess(pivotNode dsltypes.ASTNode[functions.Either[error, any]]) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			if pivotNode.Value.IsLeft() { return }
			pVal, _ := pivotNode.Value.UnwrapRight()
			var pivotInt int
			switch p := pVal.(type) {
			case int: pivotInt = p
			case string: pivotInt, _ = strconv.Atoi(p)
			default: return
			}
			
			for node := range stream {
				if node.Value.IsLeft() { continue }
				val, _ := node.Value.UnwrapRight()
				var vInt int
				switch v := val.(type) {
				case int: vInt = v
				case string: vInt, _ = strconv.Atoi(v)
				default: continue
				}
				
				if vInt < pivotInt {
					if !yield(node) { return }
				}
			}
		}
	}
}

func StepFilterGreaterEqual(pivotNode dsltypes.ASTNode[functions.Either[error, any]]) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			if pivotNode.Value.IsLeft() { return }
			pVal, _ := pivotNode.Value.UnwrapRight()
			var pivotInt int
			switch p := pVal.(type) {
			case int: pivotInt = p
			case string: pivotInt, _ = strconv.Atoi(p)
			default: return
			}
			
			for node := range stream {
				if node.Value.IsLeft() { continue }
				val, _ := node.Value.UnwrapRight()
				var vInt int
				switch v := val.(type) {
				case int: vInt = v
				case string: vInt, _ = strconv.Atoi(v)
				default: continue
				}
				
				if vInt >= pivotInt {
					if !yield(node) { return }
				}
			}
		}
	}
}

func StepConcat(left, right Adapter[functions.Either[error, any]]) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				for r := range left(node) {
					if !yield(r) { return }
				}
				for r := range right(node) {
					if !yield(r) { return }
				}
			}
		}
	}
}

func StepSeq(adapters []Adapter[functions.Either[error, any]]) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				for _, adapter := range adapters {
					for r := range adapter(node) {
						if !yield(r) { return }
					}
				}
			}
		}
	}
}

func StepEmpty() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for range stream {
				// yield nothing
			}
		}
	}
}
