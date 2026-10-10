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
}

// NewEngine initializes an empty evaluator engine.
func NewEngine[T any]() *Engine[T] {
	return &Engine[T]{
		adapters: make(map[dsltypes.BitFlag]Adapter[T]),
	}
}

// Register maps an Adapter to a specific BitFlag quasi-type.
//
// If a node matches the BitFlag, this adapter will be executed.
func (e *Engine[T]) Register(flag dsltypes.BitFlag, adapter Adapter[T]) {
	e.adapters[flag] = adapter
}

// Evaluate sequentially traverses the graph.
//
// It consumes a lazy sequence of nodes, pattern-matches their runtime BitFlag 
// to find an Adapter, executes the Adapter, and lazily flattens the resulting
// nodes back into the stream.
//
// This is structurally a continuous FlatMap / Bind over the IO Monad.
func (e *Engine[T]) Evaluate(stream iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]] {
	return functions.Bind(stream, e.applyAdapter)
}

// ParEvaluate concurrently traverses the graph using a worker pool.
//
// Because ASTNodes and BitFlags are immutable, and the Adapters isolate
// side-effects safely in the IO Monad, evaluating the graph in parallel 
// is mathematically pure and requires no mutex locking.
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
	// If a node has no adapter, it yields nothing, gracefully halting this path
	// of the traversal. In a fully interactive version, this could yield a
	// special "UnknownNode" type up to the UI to trigger the interactive prompt.
	return func(yield func(dsltypes.ASTNode[T]) bool) {}
}
