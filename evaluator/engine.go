package evaluator

import (
	"iter"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// Adapter is a monadic function that takes an ASTNode and yields a sequence of new ASTNodes.
//
// In our architecture, Adapters act as Type Classes. They encapsulate the IO Monad 
// (side effects like SSH, file parsing) safely away from the evaluation loop.
type Adapter[T any] func(dsltypes.ASTNode[T]) iter.Seq[dsltypes.ASTNode[T]]

// Engine executes the interactive graph traversal DSL.
type Engine[T any] struct {
	adapters map[dsltypes.BitFlag]Adapter[T]

	// OnMissingAdapter is called when a node matches no registered adapter.
	// This fulfills Phase 5 (The Interactive REPL Pause & Resume).
	// Because iterators pause the execution safely, this callback can block, 
	// prompt a user for input, dynamically construct a new Adapter, and return it.
	OnMissingAdapter func(dsltypes.BitFlag) Adapter[T]
}

// NewEngine initializes an empty evaluator engine.
func NewEngine[T any]() *Engine[T] {
	return &Engine[T]{
		adapters: make(map[dsltypes.BitFlag]Adapter[T]),
	}
}

// Register maps an Adapter to a specific BitFlag quasi-type.
func (e *Engine[T]) Register(flag dsltypes.BitFlag, adapter Adapter[T]) {
	e.adapters[flag] = adapter
}

// Evaluate sequentially traverses the graph.
func (e *Engine[T]) Evaluate(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
	return functions.Bind(stream, e.applyAdapter)
}

// ParEvaluate concurrently traverses the graph using a worker pool.
func (e *Engine[T]) ParEvaluate(stream iter.Seq[dsltypes.ASTNode[T]], workers int) iter.Seq[dsltypes.ASTNode[T]] {
	return functions.ParBind(stream, workers, e.applyAdapter)
}

// applyAdapter is the core pattern-matching router.
func (e *Engine[T]) applyAdapter(node dsltypes.ASTNode[T]) iter.Seq[dsltypes.ASTNode[T]] {
	for flag, adapter := range e.adapters {
		if node.Match(flag) {
			return adapter(node)
		}
	}
	
	// Phase 5: REPL Pause concept.
	// If a node has no adapter, check if we have an interactive callback.
	if e.OnMissingAdapter != nil {
		newAdapter := e.OnMissingAdapter(node.DSLType)
		if newAdapter != nil {
			// Register it for future use
			e.Register(node.DSLType, newAdapter)
			// Execute it immediately on the current node
			return newAdapter(node)
		}
	}

	// Yield nothing if unhandled
	return func(yield func(dsltypes.ASTNode[T]) bool) {}
}
