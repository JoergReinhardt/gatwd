package evaluator

import (
	"fmt"
	"iter"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/joergreinhardt/gatwd/arrows"
	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/comonads"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/optics"
	"github.com/joergreinhardt/gatwd/parser"
)

var ansiRegex = regexp.MustCompile("\033\\[[0-9;]*m")

// PipelineStep is one stage of a lazy stream transformation: a function that
// maps an input sequence of ASTNode[T] to an output sequence of the same type.
// The type parameter T is the payload carried by each node (commonly
// functions.Either[error, any]). Because steps compose as plain function
// composition over iter.Seq, an entire pipeline is itself a PipelineStep and
// nothing is evaluated until the final sequence is ranged over.
type PipelineStep[T any] func(iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]]

// EvaluatePipeline threads the initial stream through each step in order,
// returning the composed output stream. The type parameter T is the node
// payload type. Composition is left-to-right: steps[0] sees the raw input and
// each subsequent step consumes the previous step's output. Evaluation stays
// fully lazy — this method only wires the stages together; no element is pulled
// until the returned iter.Seq is iterated.
func (e *Engine[T]) EvaluatePipeline(initial iter.Seq[dsltypes.ASTNode[T]], steps []PipelineStep[T]) iter.Seq[dsltypes.ASTNode[T]] {
	stream := initial
	for _, step := range steps {
		stream = step(stream)
	}
	return stream
}

// StepAdapter lifts an Adapter into a PipelineStep by monadically binding it
// over the stream. The type parameter T is the node payload. Because an Adapter
// maps one node to a sequence of nodes, binding it flattens those per-node
// sequences into a single continuous output stream (a Kleisli fan-out followed
// by a flatten).
func StepAdapter[T any](adapter Adapter[T]) PipelineStep[T] {
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		return functions.Bind(stream, adapter)
	}
}

// StepPar runs step over every incoming node concurrently on a pool of workers and
// flattens the results into one stream (the `||>` operator). Each node is handed to
// step as a one-element stream, so step must be a per-element transformation (map,
// filter, an adapter, a command); a whole-stream step such as fold or sort would see
// only singletons. Output order is non-deterministic, as for functions.ParBind.
func StepPar[T any](step PipelineStep[T], workers int) PipelineStep[T] {
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		return functions.ParBind(stream, workers, func(node dsltypes.ASTNode[T]) iter.Seq[dsltypes.ASTNode[T]] {
			return step(func(yield func(dsltypes.ASTNode[T]) bool) { yield(node) })
		})
	}
}

// ComposeSteps chains steps left to right into one step (the empty chain is identity).
func ComposeSteps[T any](steps ...PipelineStep[T]) PipelineStep[T] {
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		for _, s := range steps {
			stream = s(stream)
		}
		return stream
	}
}

// StepTake returns a PipelineStep that yields at most the first n nodes of the
// stream and then stops, truncating (potentially infinite) sequences. The type
// parameter T is the node payload.
func StepTake[T any](n int) PipelineStep[T] {
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		return functions.Take(stream, n)
	}
}

// StepDrop returns a PipelineStep that discards the first n nodes of the stream
// and yields every node after them. The type parameter T is the node payload.
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
					if !yield(newNode) {
						return
					}
				}
			}
		})
	}
}

// StepPlus (<+>) merges several sources into one stream: every incoming node is run
// through each branch adapter in turn and the outputs are concatenated, branch by
// branch. It is the pipeline form of arrows.Plus (ArrowPlus / MonadPlus choice) —
// the generic, domain-free way to feed a single downstream pipeline from many
// producers, e.g. `spawn StrVal: ns |> (crawl_a <+> crawl_b <+> crawl_c) |> build`.
// Left (error) nodes pass through untouched; with no branches the step is the zero
// arrow and yields nothing for Right nodes.
func StepPlus(branches ...Adapter[functions.Either[error, any]]) PipelineStep[functions.Either[error, any]] {
	type node = dsltypes.ASTNode[functions.Either[error, any]]
	ks := make([]arrows.Kleisli[node, node], len(branches))
	for idx, b := range branches {
		ks[idx] = arrows.Kleisli[node, node](b)
	}
	plus := arrows.Plus(ks...)
	return func(stream iter.Seq[node]) iter.Seq[node] {
		return func(yield func(node) bool) {
			for n := range stream {
				if n.Value.IsLeft() {
					if !yield(n) {
						return
					}
					continue
				}
				for out := range plus(n) {
					if !yield(out) {
						return
					}
				}
			}
		}
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
					if !yield(node) {
						return
					}
					continue
				}

				val, _ := node.Value.UnwrapRight()
				hash := fmt.Sprintf("%v", val)

				if !visited.Contains(hash) {
					visited = visited.Add(hash) // Immutably create new Set
					if !yield(node) {
						return
					}
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
					if !yield(node) {
						return
					}
					continue
				}
				val, _ := node.Value.UnwrapRight()
				if tuple, ok := val.(functions.Pair[any, any]); ok {
					formatted := fmt.Sprintf(formatStr, tuple.Fst, tuple.Snd)
					newNode := dsltypes.NewASTNode(functions.Right[error, any](any(formatted)), node.DSLType)
					if !yield(newNode) {
						return
					}
				} else {
					errNode := InjectError[any](fmt.Errorf("Type Error: format_tuple expected a Tuple (Pair), but received %T", val), node.DSLType)
					if !yield(errNode) {
						return
					}
				}
			}
		}
	}
}

// StepExpect returns a PipelineStep that validates each Right value against
// predicate, acting as a guard in the Either error-handling railway. The type
// parameter T is the success payload. Nodes already carrying a Left are passed
// through untouched; for a Right value the predicate is applied and, when it
// returns false, the node is replaced by an error node wrapping err (via
// InjectError). Values that satisfy the predicate are forwarded unchanged.
func StepExpect[T any](predicate func(T) bool, err error) PipelineStep[functions.Either[error, T]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, T]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, T]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, T]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) {
						return
					}
					continue
				}

				val, _ := node.Value.UnwrapRight()
				if !predicate(val) {
					errNode := InjectError[T](err, node.DSLType)
					if !yield(errNode) {
						return
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

// StepZip returns a PipelineStep that expands each incoming node through the
// secondary Adapter, flattening the Adapter's per-node output into the stream.
// The type parameter T is the node payload. For every source node it iterates
// the sequence produced by secondary(node) and yields each result in turn, so a
// one-to-many secondary adapter interleaves its outputs directly into the main
// stream.
func StepZip[T any](secondary Adapter[T]) PipelineStep[T] {
	return func(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
		return func(yield func(dsltypes.ASTNode[T]) bool) {
			for node := range stream {
				secondaryStream := secondary(node)
				for secNode := range secondaryStream {
					if !yield(secNode) {
						return
					}
				}
			}
		}
	}
}

// StepFocus returns a PipelineStep that narrows each record-like node down to a
// single named field, using a functional lens as the composable getter. It
// handles both map[string]any records and *collections.HAMT values. Left nodes
// pass through unchanged; for a Right record it yields a node holding just the
// focused field's value. If the field is absent, or the node is neither a
// Record nor a HAMT, it emits a descriptive lens error node (via InjectError)
// instead of panicking — the optics-based, total alternative to raw field
// access.
func StepFocus(field string) PipelineStep[functions.Either[error, any]] {
	lens := optics.RecordLens(field)
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

				if record, ok := val.(map[string]any); ok {
					if _, exists := record[field]; exists {
						fieldVal := lens.Get(record)
						newNode := dsltypes.NewASTNode(functions.Right[error, any](fieldVal), node.DSLType)
						if !yield(newNode) {
							return
						}
					} else {
						errNode := InjectError[any](fmt.Errorf("Lens Error: Cannot focus on field '%s' because it does not exist in this Record", field), node.DSLType)
						if !yield(errNode) {
							return
						}
					}
				} else if hamt, ok := val.(*collections.HAMT[string, any]); ok {
					hamtLens := optics.HAMTLens[string, any](field)
					mVal := hamtLens.Get(hamt)
					if mVal.IsJust() {
						fieldVal, _ := mVal.Unwrap()
						newNode := dsltypes.NewASTNode(functions.Right[error, any](fieldVal), node.DSLType)
						if !yield(newNode) {
							return
						}
					} else {
						errNode := InjectError[any](fmt.Errorf("Lens Error: Cannot focus on field '%s' because it does not exist in this HAMT", field), node.DSLType)
						if !yield(errNode) {
							return
						}
					}
				} else {
					errNode := InjectError[any](fmt.Errorf("Lens Error: Cannot focus because Node is not a Record or HAMT"), node.DSLType)
					if !yield(errNode) {
						return
					}
				}
			}
		}
	}
}

// StepUpdate returns a PipelineStep that immutably sets a named field to
// newValue on each record-like node, using a functional lens as the composable
// setter. It handles both map[string]any records and *collections.HAMT values,
// producing a new structure rather than mutating the original. Left nodes pass
// through unchanged; nodes that are neither a Record nor a HAMT become a
// descriptive lens error node (via InjectError). This is the set counterpart to
// StepFocus's get.
func StepUpdate(field string, newValue any) PipelineStep[functions.Either[error, any]] {
	lens := optics.RecordLens(field)
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

				if record, ok := val.(map[string]any); ok {
					newRecord := lens.Set(record, newValue)
					newNode := dsltypes.NewASTNode(functions.Right[error, any](any(newRecord)), node.DSLType)
					if !yield(newNode) {
						return
					}
				} else if hamt, ok := val.(*collections.HAMT[string, any]); ok {
					hamtLens := optics.HAMTLens[string, any](field)
					newHamt := hamtLens.Set(hamt, functions.Just(newValue))
					newNode := dsltypes.NewASTNode(functions.Right[error, any](any(newHamt)), node.DSLType)
					if !yield(newNode) {
						return
					}
				} else {
					errNode := InjectError[any](fmt.Errorf("Lens Error: Cannot update because Node is not a Record or HAMT"), node.DSLType)
					if !yield(errNode) {
						return
					}
				}
			}
		}
	}
}

// StepFold collapses a stream of nodes into a single scalar value.
// Supports: "count", "len", "length", "sum", "product", "prod", "min", "max", "concat", "any", "or", "all", "and".
func StepFold(operation string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			op := strings.ToLower(strings.TrimSpace(operation))
			var items []any
			var lastFlag dsltypes.BitFlag
			count := 0
			for node := range stream {
				if node.Value.IsLeft() {
					continue
				} // Skip errors in fold
				val, _ := node.Value.UnwrapRight()
				items = append(items, val)
				count++
				lastFlag = node.DSLType
			}

			if len(items) == 1 {
				if colSlice, ok := functions.ToSliceAny(items[0]); ok {
					items = colSlice
					count = len(colSlice)
				}
			}

			switch op {
			case "count", "len", "length":
				val := functions.Right[error, any](fmt.Sprintf("%d", count))
				yield(dsltypes.NewASTNode(val, lastFlag))

			case "sum":
				var acc any = int64(0)
				for _, it := range items {
					var err error
					acc, err = functions.EvalBinaryOp(functions.OpAdd, acc, it)
					if err != nil {
						yield(InjectError[any](fmt.Errorf("fold sum error: %w", err), lastFlag))
						return
					}
				}
				yield(dsltypes.NewASTNode(functions.Right[error, any](acc), lastFlag))

			case "product", "prod":
				var acc any = int64(1)
				for _, it := range items {
					var err error
					acc, err = functions.EvalBinaryOp(functions.OpMul, acc, it)
					if err != nil {
						yield(InjectError[any](fmt.Errorf("fold product error: %w", err), lastFlag))
						return
					}
				}
				yield(dsltypes.NewASTNode(functions.Right[error, any](acc), lastFlag))

			case "min":
				if len(items) == 0 {
					yield(InjectError[any](fmt.Errorf("fold min on empty stream"), lastFlag))
					return
				}
				minVal := items[0]
				for _, it := range items[1:] {
					c, err := functions.Compare(it, minVal)
					if err == nil && c < 0 {
						minVal = it
					}
				}
				yield(dsltypes.NewASTNode(functions.Right[error, any](minVal), lastFlag))

			case "max":
				if len(items) == 0 {
					yield(InjectError[any](fmt.Errorf("fold max on empty stream"), lastFlag))
					return
				}
				maxVal := items[0]
				for _, it := range items[1:] {
					c, err := functions.Compare(it, maxVal)
					if err == nil && c > 0 {
						maxVal = it
					}
				}
				yield(dsltypes.NewASTNode(functions.Right[error, any](maxVal), lastFlag))

			case "concat":
				if len(items) == 0 {
					yield(dsltypes.NewASTNode(functions.Right[error, any](any("")), lastFlag))
					return
				}
				acc := items[0]
				for _, it := range items[1:] {
					var err error
					acc, err = functions.EvalBinaryOp(functions.OpConcat, acc, it)
					if err != nil {
						yield(InjectError[any](fmt.Errorf("fold concat error: %w", err), lastFlag))
						return
					}
				}
				yield(dsltypes.NewASTNode(functions.Right[error, any](acc), lastFlag))

			case "any", "or":
				res := false
				for _, it := range items {
					if b, ok := it.(bool); ok {
						if b {
							res = true
							break
						}
					} else if s, ok := it.(string); ok {
						if s != "" && s != "0" && s != "false" {
							res = true
							break
						}
					} else if functions.IsNumeric(it) {
						if f, ok := functions.ToFloat64(it); ok && f != 0 {
							res = true
							break
						}
					}
				}
				yield(dsltypes.NewASTNode(functions.Right[error, any](any(res)), lastFlag))

			case "all", "and":
				res := true
				for _, it := range items {
					truthy := false
					if b, ok := it.(bool); ok {
						truthy = b
					} else if s, ok := it.(string); ok {
						truthy = s != "" && s != "0" && s != "false"
					} else if functions.IsNumeric(it) {
						if f, ok := functions.ToFloat64(it); ok {
							truthy = f != 0
						}
					}
					if !truthy {
						res = false
						break
					}
				}
				yield(dsltypes.NewASTNode(functions.Right[error, any](any(res)), lastFlag))

			default:
				errNode := InjectError[any](fmt.Errorf("Unknown fold operation: %s", operation), lastFlag)
				yield(errNode)
			}
		}
	}
}

// StepSetUnion unions incoming Set with operand sets or accumulates a stream of sets.
func StepSetUnion(operand ...any) PipelineStep[functions.Either[error, any]] {
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
				if len(operand) > 0 {
					res, err := functions.EvalBinaryOp(functions.OpConcat, val, operand[0])
					if err != nil {
						yield(InjectError[any](fmt.Errorf("set_union error: %w", err), node.DSLType))
						return
					}
					if !yield(dsltypes.NewASTNode(functions.Right[error, any](res), node.DSLType)) {
						return
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

// rangeMapLike invokes fn for every key/value of a product-like payload (a
// persistent *collections.HAMT, a *collections.Map, or a map[string]any Record),
// reporting whether v was map-like at all. It is the iteration dual of HAMTGetKey.
func rangeMapLike(v any, fn func(k string, val any)) bool {
	switch c := v.(type) {
	case *collections.HAMT[string, any]:
		for k, val := range c.Seq() {
			fn(k, val)
		}
		return true
	case *collections.Map[string, any]:
		for k, val := range c.Seq() {
			fn(k, val)
		}
		return true
	case map[string]any:
		for k, val := range c {
			fn(k, val)
		}
		return true
	default:
		return false
	}
}

// IsSubsetValue reports whether sub ⊆ super, generically and domain-free, across the
// two families where containment is well-defined:
//
//   - Record/map-like (HAMT, Map, map[string]any): sub ⊆ super iff every (key,value)
//     of sub is present in super with a structurally-equal value. This is the
//     "selector ⊆ labels" shape — a map is a subset when all its bindings hold.
//   - Set/collection-like (Set, List, Chunk, any ToSliceAny-able value): sub ⊆ super
//     iff every element of sub is a member of super (by structural equality).
//
// The empty subset is ⊆ anything (vacuously true). A type mismatch (one map-like and
// one not, or neither recognizable) is an error, not a silent false, so a mistyped
// operand is debuggable. Equality uses functions.StructuralEqual so nested
// collections and ADTs compare by value.
func IsSubsetValue(sub, super any) (bool, error) {
	// Map-like subset: every binding of sub must hold in super.
	subIsMap := false
	result := true
	ranged := rangeMapLike(sub, func(k string, v any) {
		subIsMap = true
		sv, found, navigable := HAMTGetKey(super, k)
		if !navigable || !found || !functions.StructuralEqual(v, sv) {
			result = false
		}
	})
	if ranged {
		if _, _, superNavigable := HAMTGetKey(super, ""); subIsMap && !superNavigable {
			// super is not map-like at all — a genuine type mismatch.
			return false, fmt.Errorf("subset_of: subset is a Record/HAMT/Map but superset is not (got %T)", super)
		}
		return result, nil
	}

	// Collection-like subset: every element of sub must be a member of super.
	subEls, okA := functions.ToSliceAny(sub)
	superEls, okB := functions.ToSliceAny(super)
	if okA && okB {
		for _, e := range subEls {
			member := false
			for _, s := range superEls {
				if functions.StructuralEqual(e, s) {
					member = true
					break
				}
			}
			if !member {
				return false, nil
			}
		}
		return true, nil
	}

	return false, fmt.Errorf("subset_of: operands must be both Records/HAMTs or both Sets/collections (got %T and %T)", sub, super)
}

// StepSubsetOf tests whether the incoming value (the subject) is a subset of other,
// yielding a bool. `subject |> subset_of other` reads "subject ⊆ other"; it is the
// pipeline-step twin of the subset_of expression-level builtin. See IsSubsetValue
// for the generic Set/Record semantics.
func StepSubsetOf(other any) PipelineStep[functions.Either[error, any]] {
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
				res, err := IsSubsetValue(val, other)
				if err != nil {
					if !yield(InjectError[any](err, node.DSLType)) {
						return
					}
					continue
				}
				if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(res)), node.DSLType)) {
					return
				}
			}
		}
	}
}

// StepSetIntersect computes the intersection of incoming Set with other.
func StepSetIntersect(other any) PipelineStep[functions.Either[error, any]] {
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
				rvVal := reflect.ValueOf(val)
				mIntersect := rvVal.MethodByName("Intersect")
				if mIntersect.IsValid() && other != nil && reflect.TypeOf(other).AssignableTo(mIntersect.Type().In(0)) {
					out := mIntersect.Call([]reflect.Value{reflect.ValueOf(other)})
					if !yield(dsltypes.NewASTNode(functions.Right[error, any](out[0].Interface()), node.DSLType)) {
						return
					}
				} else {
					elemsA, okA := functions.ToSliceAny(val)
					elemsB, okB := functions.ToSliceAny(other)
					if okA && okB {
						resSet := collections.NewSet[any]()
						for _, ea := range elemsA {
							for _, eb := range elemsB {
								if functions.StructuralEqual(ea, eb) {
									resSet = resSet.Add(ea)
									break
								}
							}
						}
						if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(resSet)), node.DSLType)) {
							return
						}
					} else {
						yield(InjectError[any](fmt.Errorf("set_intersect: expected Set, got %T and %T", val, other), node.DSLType))
						return
					}
				}
			}
		}
	}
}

// StepSetDiff computes the relative complement of incoming Set minus other.
func StepSetDiff(other any) PipelineStep[functions.Either[error, any]] {
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
				rvVal := reflect.ValueOf(val)
				mDiff := rvVal.MethodByName("Diff")
				if mDiff.IsValid() && other != nil && reflect.TypeOf(other).AssignableTo(mDiff.Type().In(0)) {
					out := mDiff.Call([]reflect.Value{reflect.ValueOf(other)})
					if !yield(dsltypes.NewASTNode(functions.Right[error, any](out[0].Interface()), node.DSLType)) {
						return
					}
				} else {
					elemsA, okA := functions.ToSliceAny(val)
					elemsB, okB := functions.ToSliceAny(other)
					if okA && okB {
						resSet := collections.NewSet[any]()
						for _, ea := range elemsA {
							found := false
							for _, eb := range elemsB {
								if functions.StructuralEqual(ea, eb) {
									found = true
									break
								}
							}
							if !found {
								resSet = resSet.Add(ea)
							}
						}
						if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(resSet)), node.DSLType)) {
							return
						}
					} else {
						yield(InjectError[any](fmt.Errorf("set_diff: expected Set, got %T and %T", val, other), node.DSLType))
						return
					}
				}
			}
		}
	}
}

// StepHAMTMerge merges incoming HAMT with other.
func StepHAMTMerge(other any) PipelineStep[functions.Either[error, any]] {
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
				res, err := functions.EvalBinaryOp(functions.OpConcat, val, other)
				if err != nil {
					yield(InjectError[any](fmt.Errorf("hamt_merge error: %w", err), node.DSLType))
					return
				}
				if !yield(dsltypes.NewASTNode(functions.Right[error, any](res), node.DSLType)) {
					return
				}
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
					if !yield(node) {
						return
					}
					continue
				}
				val, _ := node.Value.UnwrapRight()
				strVal := fmt.Sprintf("%v", val)

				matched := false
				for pattern, adapter := range cases {
					if strings.Contains(strVal, pattern) {
						for res := range adapter(node) {
							if !yield(res) {
								return
							}
						}
						matched = true
						break
					}
				}

				if !matched && defaultAdapter != nil {
					for res := range defaultAdapter(node) {
						if !yield(res) {
							return
						}
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
					if !yield(node) {
						return
					}
					continue
				}

				val, _ := node.Value.UnwrapRight()
				if predicate(val) {
					if !yield(node) {
						return
					}
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
					if !yield(node) {
						return
					}
					continue
				}
				val, _ := node.Value.UnwrapRight()

				if tuple, ok := val.(functions.Pair[any, any]); ok {
					// We mock the concurrent Applicative application here
					leftNode := dsltypes.NewASTNode(functions.Right[error, any](tuple.Fst), node.DSLType)
					rightNode := dsltypes.NewASTNode(functions.Right[error, any](tuple.Snd), node.DSLType)

					var res1, res2 any
					for r1 := range adapter1(leftNode) {
						if !r1.Value.IsLeft() {
							res1, _ = r1.Value.UnwrapRight()
							break
						}
					}
					for r2 := range adapter2(rightNode) {
						if !r2.Value.IsLeft() {
							res2, _ = r2.Value.UnwrapRight()
							break
						}
					}

					newTuple := functions.Pair[any, any]{Fst: res1, Snd: res2}
					newNode := dsltypes.NewASTNode(functions.Right[error, any](any(newTuple)), node.DSLType)
					if !yield(newNode) {
						return
					}
				} else {
					errNode := InjectError[any](fmt.Errorf("Applicative Error: Split [***] expected a Tuple (Pair), but received a scalar value"), node.DSLType)
					if !yield(errNode) {
						return
					}
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
				if node.Value.IsLeft() {
					continue
				}
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
				if node.Value.IsLeft() {
					continue
				}
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
					if !yield(node) {
						return
					}
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
					if !yield(errNode) {
						return
					}
					continue
				}

				// Apply Operation
				if op == "into_zipper" {
					newNode := dsltypes.NewASTNode(functions.Right[error, any](any(zipper)), node.DSLType)
					if !yield(newNode) {
						return
					}
				} else if op == "down" {
					nextZipper := zipper.Down()
					if nextZipper != nil {
						newNode := dsltypes.NewASTNode(functions.Right[error, any](any(nextZipper)), node.DSLType)
						if !yield(newNode) {
							return
						}
					} else {
						errNode := InjectError[any](fmt.Errorf("Cursor Error: Cannot move Zipper DOWN because the current Tree node has no children"), node.DSLType)
						if !yield(errNode) {
							return
						}
					}
				} else if op == "right" {
					nextZipper := zipper.Right()
					if nextZipper != nil {
						newNode := dsltypes.NewASTNode(functions.Right[error, any](any(nextZipper)), node.DSLType)
						if !yield(newNode) {
							return
						}
					} else {
						errNode := InjectError[any](fmt.Errorf("Cursor Error: Cannot move Zipper RIGHT because there is no sibling"), node.DSLType)
						if !yield(errNode) {
							return
						}
					}
				} else if op == "left" {
					nextZipper := zipper.Left()
					if nextZipper != nil {
						newNode := dsltypes.NewASTNode(functions.Right[error, any](any(nextZipper)), node.DSLType)
						if !yield(newNode) {
							return
						}
					} else {
						errNode := InjectError[any](fmt.Errorf("Cursor Error: Cannot move Zipper LEFT because there is no sibling"), node.DSLType)
						if !yield(errNode) {
							return
						}
					}
				} else if op == "up" {
					nextZipper := zipper.Up()
					if nextZipper != nil {
						newNode := dsltypes.NewASTNode(functions.Right[error, any](any(nextZipper)), node.DSLType)
						if !yield(newNode) {
							return
						}
					} else {
						errNode := InjectError[any](fmt.Errorf("Cursor Error: Cannot move Zipper UP because the Cursor is already at the Root of the Tree"), node.DSLType)
						if !yield(errNode) {
							return
						}
					}
				} else if op == "extract" {
					extracted := zipper.Extract()
					newNode := dsltypes.NewASTNode(functions.Right[error, any](extracted), node.DSLType)
					if !yield(newNode) {
						return
					}
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
				if node.Value.IsLeft() {
					continue
				}
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
				if node.Value.IsLeft() {
					continue
				}
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
				if node.Value.IsLeft() {
					continue
				}
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

// StepMathAdd returns a PipelineStep that adds amount to each Right value via
// the type-directed functions.EvalBinaryOp with OpAdd, so the operation adapts
// to the operands (numeric addition, string/monoid concatenation, etc.). The
// amount is passed as a string and coerced by EvalBinaryOp. Left nodes pass
// through unchanged; if the addition fails for a given value, that node becomes
// an error node (via InjectError) and the stream continues.
func StepMathAdd(amount string) PipelineStep[functions.Either[error, any]] {
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
				newVal, err := functions.EvalBinaryOp(functions.OpAdd, val, functions.OperandValue(functions.OpAdd, amount))
				if err != nil {
					errNode := InjectError[any](fmt.Errorf("math_add failed: %w", err), node.DSLType)
					if !yield(errNode) {
						return
					}
					continue
				}

				newNode := dsltypes.NewASTNode(functions.Right[error, any](newVal), node.DSLType)
				if !yield(newNode) {
					return
				}
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
				if node.Value.IsLeft() {
					continue
				}
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
					if !yield(node) {
						return
					}
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
					if !yield(resNode) {
						return
					}
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
				if !yield(errNode) {
					return
				}
			}

			// Sort the buffer
			sort.SliceStable(buffer, func(i, j int) bool {
				valI, _ := buffer[i].Value.UnwrapRight()
				valJ, _ := buffer[j].Value.UnwrapRight()
				return less(valI, valJ)
			})

			// Yield the sorted buffer lazily
			for _, node := range buffer {
				if !yield(node) {
					return
				}
			}
		}
	}
}

// DefaultLess is the fallback ordering predicate used by StepSort, reporting
// whether a sorts before b. It compares natively when both operands share a
// known type (int, float64, or string); for mismatched or unrecognized types it
// degrades gracefully to a lexicographic comparison of their fmt "%v"
// renderings, so it is total and never panics.
func DefaultLess(a, b any) bool {
	switch va := a.(type) {
	case int:
		if vb, ok := b.(int); ok {
			return va < vb
		}
	case float64:
		if vb, ok := b.(float64); ok {
			return va < vb
		}
	case string:
		if vb, ok := b.(string); ok {
			return va < vb
		}
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
				if !yield(errNode) {
					return
				}
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
				if !yield(errNode) {
					return
				}
			}

			newNode := dsltypes.NewASTNode(functions.Right[error, any](any(hamt)), lastFlag)
			yield(newNode)
		}
	}
}

// StepToList aggregates all elements in the stream into an immutable persistent List.
func StepToList() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			var items []any
			var errs []dsltypes.ASTNode[functions.Either[error, any]]
			var lastFlag dsltypes.BitFlag

			for node := range stream {
				if node.Value.IsLeft() {
					errs = append(errs, node)
					continue
				}
				val, _ := node.Value.UnwrapRight()
				items = append(items, val)
				lastFlag = node.DSLType
			}

			for _, errNode := range errs {
				if !yield(errNode) {
					return
				}
			}

			resList := collections.Nil[any]()
			for i := len(items) - 1; i >= 0; i-- {
				resList = collections.Cons(items[i], resList)
			}
			newNode := dsltypes.NewASTNode(functions.Right[error, any](any(resList)), lastFlag)
			yield(newNode)
		}
	}
}

// StepToChunk aggregates all elements in the stream into a contiguous bounded Chunk.
func StepToChunk() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			var items []any
			var errs []dsltypes.ASTNode[functions.Either[error, any]]
			var lastFlag dsltypes.BitFlag

			for node := range stream {
				if node.Value.IsLeft() {
					errs = append(errs, node)
					continue
				}
				val, _ := node.Value.UnwrapRight()
				items = append(items, val)
				lastFlag = node.DSLType
			}

			for _, errNode := range errs {
				if !yield(errNode) {
					return
				}
			}

			resChunk := collections.NewChunk[any](len(items))
			for _, it := range items {
				resChunk.Append(it)
			}
			newNode := dsltypes.NewASTNode(functions.Right[error, any](any(resChunk)), lastFlag)
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
					if !yield(node) {
						return
					}
					continue
				}
				val, _ := node.Value.UnwrapRight()
				res, found, navigable := HAMTGetKey(val, key)
				if !navigable {
					yield(node) // pass through if not a HAMT/Map
					continue
				}
				if found {
					yield(dsltypes.NewASTNode(functions.Right[error, any](res), node.DSLType))
				} else {
					yield(InjectError[any](fmt.Errorf("key '%s' not found", key), node.DSLType))
				}
			}
		}
	}
}

// HAMTGetKey reads a single key out of any product-like payload the DSL produces —
// a persistent *collections.HAMT, a *collections.Map, or a map[string]any Record —
// returning the value, whether the key was present, and whether the payload was a
// navigable carrier at all. Keeping this dispatch in one place lets both the
// hamt_get pipeline step and the hamt_get expression-level builtin (the function
// form applied inside fold/map lambda bodies) share identical semantics.
func HAMTGetKey(val any, key string) (value any, found bool, navigable bool) {
	switch carrier := val.(type) {
	case *collections.HAMT[string, any]:
		v, ok := carrier.Get(key)
		return v, ok, true
	case *collections.Map[string, any]:
		v, ok := carrier.Get(key)
		return v, ok, true
	case map[string]any:
		v, ok := carrier[key]
		return v, ok, true
	default:
		return nil, false, false
	}
}

// NavigateHAMTPath walks a dotted path ("spec.selector.app") through nested
// product-like payloads, returning the value at the path or a descriptive error
// naming the failing segment. It is the pure core shared by the hamt_path pipeline
// step and the hamt_path expression-level builtin: an empty path returns the input
// unchanged (identity), and because a path over possibly-absent keys is an affine
// traversal (a "preview"), a missing or non-navigable segment is an error rather
// than a silent nil.
func NavigateHAMTPath(cur any, path string) (any, error) {
	if path == "" {
		return cur, nil
	}
	for _, seg := range strings.Split(path, ".") {
		next, found, navigable := HAMTGetKey(cur, seg)
		if !navigable {
			return nil, fmt.Errorf("Path Error: cannot navigate segment '%s' of path '%s': parent is not a Record/HAMT/Map (got %T)", seg, path, cur)
		}
		if !found {
			return nil, fmt.Errorf("Path Error: segment '%s' of path '%s' not found", seg, path)
		}
		cur = next
	}
	return cur, nil
}

// StepHAMTPath navigates a dotted path ("spec.selector.app") through nested
// payloads in a single step — the multi-level generalization of hamt_get, and the
// DSL-facing accessor that gives struct-like deep field access over gatwd's
// dynamically-typed, open-schema data (the shape the Kubernetes adapter yields).
//
// It is the dynamic, multi-carrier sibling of optics.HAMTPathLens: at each path
// segment it dispatches on the carrier so one accessor spans every product-like
// value the DSL produces — a persistent *collections.HAMT, a *collections.Map, and
// a map[string]any Record (from parse_record). Because a path over possibly-absent
// keys is an Optional (affine traversal), not a total Lens, this realizes its
// "preview" direction: it stops and emits a descriptive error (via InjectError,
// naming the failing segment and the full path) the moment a segment is missing or
// its parent is not navigable, so a mistyped config path is debuggable instead of
// silently yielding nil. An empty path passes the node through unchanged.
func StepHAMTPath(path string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) {
						return
					}
					continue
				}
				cur, _ := node.Value.UnwrapRight()
				res, err := NavigateHAMTPath(cur, path)
				if err != nil {
					if !yield(InjectError[any](err, node.DSLType)) {
						return
					}
					continue
				}
				if !yield(dsltypes.NewASTNode(functions.Right[error, any](res), node.DSLType)) {
					return
				}
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
					if !yield(node) {
						return
					}
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
					if !yield(node) {
						return
					}
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
					if !yield(node) {
						return
					}
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
					if !yield(node) {
						return
					}
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

// StepHAMTPairs unfolds an incoming HAMT or Map into a stream of key-value Pairs.
func StepHAMTPairs() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) {
						return
					}
					continue
				}
				v, _ := node.Value.UnwrapRight()
				if h, ok := v.(*collections.HAMT[string, any]); ok {
					for k, val := range h.Seq() {
						pair := functions.Pair[any, any]{Fst: any(k), Snd: val}
						if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(pair)), node.DSLType)) {
							return
						}
					}
					continue
				}
				if m, ok := v.(*collections.Map[string, any]); ok {
					for k, val := range m.Seq() {
						pair := functions.Pair[any, any]{Fst: any(k), Snd: val}
						if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(pair)), node.DSLType)) {
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
					if !yield(node) {
						return
					}
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
					if !yield(node) {
						return
					}
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
					if !yield(node) {
						return
					}
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
					if !yield(node) {
						return
					}
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
	var anyVals []any
	for _, v := range values {
		anyVals = append(anyVals, v)
	}
	return StepEmitValues(anyVals...)
}

// StepEmitValues emits a fixed set of arbitrary typed values into the stream.
func StepEmitValues(values ...any) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			// First, pass through upstream
			for node := range stream {
				if !yield(node) {
					return
				}
			}

			// Then emit our values
			for _, v := range values {
				node := dsltypes.NewASTNode(functions.Right[error, any](v), 0)
				if !yield(node) {
					return
				}
			}
		}
	}
}

// StepAssert checks if the incoming node's string representation matches the expected string.
// If it fails, it yields an error node. If it passes, it yields the original node unaltered.
// An assertion that never sees a node would pass vacuously, so a stream that ends without
// delivering any node yields an error instead (fail loud: a broken pipeline is not a pass).
func StepAssert(expected string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			var seen bool
			for node := range stream {
				seen = true
				if node.Value.IsLeft() {
					if !yield(node) {
						return
					}
					continue
				}

				val, _ := node.Value.UnwrapRight()
				var actual string
				if b, ok := val.([]byte); ok {
					actual = string(b)
				} else {
					actual = fmt.Sprintf("%v", val)
				}

				expClean := strings.ReplaceAll(expected, `\n`, "\n")
				expClean = strings.ReplaceAll(expClean, `\t`, "\t")
				actualClean := ansiRegex.ReplaceAllString(actual, "")
				expClean = ansiRegex.ReplaceAllString(expClean, "")

				if strings.TrimSpace(actualClean) != strings.TrimSpace(expClean) {
					err := fmt.Errorf("Assertion failed: expected '%s', got '%s'", expected, actual)
					errNode := InjectError[any](err, node.DSLType)
					if !yield(errNode) {
						return
					}
				} else {
					if !yield(node) {
						return
					}
				}
			}
			if !seen {
				yield(InjectError[any](fmt.Errorf("assertion on an empty stream: expected '%s' but no value arrived", expected), 0))
			}
		}
	}
}

// StepMathSub returns a PipelineStep that subtracts amount from each Right value
// via functions.EvalBinaryOp with OpSub, so the operation is type-directed over
// the operands. The amount is passed as a string and coerced by EvalBinaryOp.
// Left nodes pass through unchanged; if the subtraction fails for a value, that
// node becomes an error node (via InjectError) and the stream continues.
func StepMathSub(amount string) PipelineStep[functions.Either[error, any]] {
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
				newVal, err := functions.EvalBinaryOp(functions.OpSub, val, functions.OperandValue(functions.OpSub, amount))
				if err != nil {
					errNode := InjectError[any](fmt.Errorf("math_sub failed: %w", err), node.DSLType)
					if !yield(errNode) {
						return
					}
					continue
				}

				newNode := dsltypes.NewASTNode(functions.Right[error, any](newVal), node.DSLType)
				if !yield(newNode) {
					return
				}
			}
		}
	}
}

// StepDebugPrint returns a transparent PipelineStep that prints each node to
// stdout (tagged with prefix and whether it is a LEFT error or RIGHT value) and
// then forwards it unchanged. It is a side-effecting "tap" or probe: the stream
// is observed but never altered, so it can be inserted anywhere for debugging
// without affecting the pipeline's result.
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
				if !yield(node) {
					return
				}
			}
		}
	}
}

// StepYieldRuntimeNode returns a PipelineStep that discards the incoming stream
// and instead emits a single, pre-computed runtime node — a constant generator
// used to splice a value produced elsewhere into a pipeline. As a convenience,
// if that node wraps a *collections.List, the list is flattened and its
// elements are yielded one per node; any other value (or a Left) is yielded
// as-is.
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
						if !yield(newNode) {
							return
						}
						current, _ = current.Tail()
					}
					return
				}
			}
			yield(node)
		}
	}
}

// StepFilterLess returns a PipelineStep that keeps only the nodes whose integer
// value is strictly less than the pivot carried by pivotNode. Both the pivot
// and each element are coerced to int (ints directly, strings via strconv);
// Left nodes, a Left pivot, and non-integer values are dropped. Together with
// StepFilterGreaterEqual this forms the two halves of a quicksort-style
// partition around a pivot.
func StepFilterLess(pivotNode dsltypes.ASTNode[functions.Either[error, any]]) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			if pivotNode.Value.IsLeft() {
				return
			}
			pVal, _ := pivotNode.Value.UnwrapRight()
			var pivotInt int
			switch p := pVal.(type) {
			case int:
				pivotInt = p
			case string:
				pivotInt, _ = strconv.Atoi(p)
			default:
				return
			}

			for node := range stream {
				if node.Value.IsLeft() {
					continue
				}
				val, _ := node.Value.UnwrapRight()
				var vInt int
				switch v := val.(type) {
				case int:
					vInt = v
				case string:
					vInt, _ = strconv.Atoi(v)
				default:
					continue
				}

				if vInt < pivotInt {
					if !yield(node) {
						return
					}
				}
			}
		}
	}
}

// StepFilterGreaterEqual returns a PipelineStep that keeps only the nodes whose
// integer value is greater than or equal to the pivot carried by pivotNode.
// Both the pivot and each element are coerced to int (ints directly, strings
// via strconv); Left nodes, a Left pivot, and non-integer values are dropped.
// It is the complement of StepFilterLess, together partitioning a stream around
// a pivot.
func StepFilterGreaterEqual(pivotNode dsltypes.ASTNode[functions.Either[error, any]]) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			if pivotNode.Value.IsLeft() {
				return
			}
			pVal, _ := pivotNode.Value.UnwrapRight()
			var pivotInt int
			switch p := pVal.(type) {
			case int:
				pivotInt = p
			case string:
				pivotInt, _ = strconv.Atoi(p)
			default:
				return
			}

			for node := range stream {
				if node.Value.IsLeft() {
					continue
				}
				val, _ := node.Value.UnwrapRight()
				var vInt int
				switch v := val.(type) {
				case int:
					vInt = v
				case string:
					vInt, _ = strconv.Atoi(v)
				default:
					continue
				}

				if vInt >= pivotInt {
					if !yield(node) {
						return
					}
				}
			}
		}
	}
}

// StepConcat returns a PipelineStep that applies two adapters to every incoming
// node and concatenates their outputs, yielding all results from left before
// all results from right for each node. It realizes the monoidal "++"
// (concatenation) of two one-to-many adapters over the stream, preserving order.
func StepConcat(left, right Adapter[functions.Either[error, any]]) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				for r := range left(node) {
					if !yield(r) {
						return
					}
				}
				for r := range right(node) {
					if !yield(r) {
						return
					}
				}
			}
		}
	}
}

// StepSeq returns a PipelineStep that, for each incoming node, runs the given
// adapters in order and yields every result they produce. It generalizes
// StepConcat from two adapters to an ordered slice, sequencing their per-node
// outputs into one flattened stream.
func StepSeq(adapters []Adapter[functions.Either[error, any]]) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				for _, adapter := range adapters {
					for r := range adapter(node) {
						if !yield(r) {
							return
						}
					}
				}
			}
		}
	}
}

// StepMathMul multiplies stream values by amount, with overflow promotion
func StepMathMul(amount string) PipelineStep[functions.Either[error, any]] {
	return StepBinaryOp(functions.OpMul, amount)
}

// StepMathDiv divides stream values by amount, promoting integer division to float64
func StepMathDiv(amount string) PipelineStep[functions.Either[error, any]] {
	return StepBinaryOp(functions.OpDiv, amount)
}

// StepMathIDiv performs integer Euclidean division (//)
func StepMathIDiv(amount string) PipelineStep[functions.Either[error, any]] {
	return StepBinaryOp(functions.OpIDiv, amount)
}

// StepMathMod calculates remainder / modulo (%)
func StepMathMod(amount string) PipelineStep[functions.Either[error, any]] {
	return StepBinaryOp(functions.OpMod, amount)
}

// StepMathPow raises stream values to the power amount (`**`): exact for an Int raised to a
// non-negative Int, a Float otherwise (see functions.Pow).
func StepMathPow(amount string) PipelineStep[functions.Either[error, any]] {
	return StepBinaryOp(functions.OpPow, amount)
}

// StepBinaryOp evaluates a binary operator with the stream element on the left and operand on the right
func StepBinaryOp(op functions.OpKind, operand string) PipelineStep[functions.Either[error, any]] {
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
				newVal, err := functions.EvalBinaryOp(op, val, functions.OperandValue(op, operand))
				if err != nil {
					errNode := InjectError[any](fmt.Errorf("operator %s failed: %w", op, err), node.DSLType)
					if !yield(errNode) {
						return
					}
					continue
				}

				newNode := dsltypes.NewASTNode(functions.Right[error, any](newVal), node.DSLType)
				if !yield(newNode) {
					return
				}
			}
		}
	}
}

// StepEmpty returns a PipelineStep that fully drains its input and yields
// nothing, producing an empty stream. It still iterates the source so any
// upstream side effects run, acting as a terminal "sink" that discards all
// values — the absorbing element for stream composition.
func StepEmpty() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for range stream {
				// yield nothing
			}
		}
	}
}

// StepFrequencies counts occurrences of elements in the stream, producing a persistent HAMT of item -> count.
func StepFrequencies() PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			h := collections.NewHAMT[string, any]()
			var lastFlag dsltypes.BitFlag
			for node := range stream {
				if node.Value.IsLeft() {
					continue
				}
				val, _ := node.Value.UnwrapRight()
				key := fmt.Sprintf("%v", val)
				lastFlag = node.DSLType

				currentCount := 0
				if curVal, exists := h.Get(key); exists {
					if c, ok := curVal.(int); ok {
						currentCount = c
					}
				}
				h = h.Set(key, currentCount+1)
			}
			yield(dsltypes.NewASTNode(functions.Right[error, any](any(h)), lastFlag))
		}
	}
}

// StepEBNFParse compiles an EBNF grammar specification and parses each incoming string node into an AST.
func StepEBNFParse(startRule, ebnfText string) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			grammar, err := parser.ParseEBNFGrammar(startRule, ebnfText)
			if err != nil {
				errNode := InjectError[any](fmt.Errorf("EBNF grammar error: %w", err), 0)
				yield(errNode)
				return
			}

			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) {
						return
					}
					continue
				}

				val, _ := node.Value.UnwrapRight()
				strInput := fmt.Sprintf("%v", val)
				ast, _, pErr := grammar.Parse(strInput)
				if pErr != nil {
					errNode := InjectError[any](fmt.Errorf("EBNF parse failed for %q: %w", strInput, pErr), node.DSLType)
					if !yield(errNode) {
						return
					}
					continue
				}

				newNode := dsltypes.NewASTNode(functions.Right[error, any](any(ast)), node.DSLType)
				if !yield(newNode) {
					return
				}
			}
		}
	}
}
