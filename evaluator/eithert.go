package evaluator

import (
	"iter"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// EitherT is a Monad Transformer that safely wraps a pure Adapter[T]
// into an error-aware Adapter[Either[error, T]].
//
// This allows the Evaluator Engine to gracefully handle network failures,
// timeouts, or parsing errors by bypassing the broken branch of the graph
// and passing the Error node down the stream to be logged, without crashing
// the program or silently halting execution.
func EitherT[T any](adapter Adapter[T]) Adapter[functions.Either[error, T]] {
	return func(node dsltypes.ASTNode[functions.Either[error, T]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, T]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, T]]) bool) {

			// 1. BYPASS: If the node contains an Error (Left), we do not run the adapter.
			// We instantly yield the error node forward so the user sees the failure!
			if node.Value.IsLeft() {
				yield(node)
				return
			}

			// 2. UNWRAP: The node contains a valid value (Right).
			val, _ := node.Value.UnwrapRight()
			innerNode := dsltypes.NewASTNode(val, node.DSLType)

			// 3. EXECUTE: Run the pure underlying adapter.
			for result := range adapter(innerNode) {
				// 4. RE-WRAP: Package the pure result back into an Either[error, T]
				wrappedValue := functions.Right[error, T](result.Value)
				wrappedNode := dsltypes.NewASTNode(wrappedValue, result.DSLType)

				if !yield(wrappedNode) {
					return
				}
			}
		}
	}
}

// InjectError is a helper to yield a terminating Error node directly into the graph.
// Adapters can use this instead of `return` to emit graceful errors.
func InjectError[T any](err error, flag dsltypes.BitFlag) dsltypes.ASTNode[functions.Either[error, T]] {
	val := functions.Left[error, T](err)
	return dsltypes.NewASTNode(val, flag)
}
