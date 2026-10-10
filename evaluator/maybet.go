package evaluator

import (
	"iter"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// MaybeT is a Monad Transformer that safely wraps an Adapter.
// If the Node contains 'Nothing', it silently drops it from the stream,
// preventing panics without propagating an Error.
func MaybeT[T any](adapter Adapter[T]) Adapter[functions.Maybe[T]] {
	return func(node dsltypes.ASTNode[functions.Maybe[T]]) iter.Seq[dsltypes.ASTNode[functions.Maybe[T]]] {
		return func(yield func(dsltypes.ASTNode[functions.Maybe[T]]) bool) {
			if node.Value.IsNothing() {
				// Silently bypass (yield nothing)
				return
			}
			
			val, _ := node.Value.Unwrap()
			innerNode := dsltypes.NewASTNode(val, node.DSLType)
			
			for result := range adapter(innerNode) {
				wrappedValue := functions.Just[T](result.Value)
				wrappedNode := dsltypes.NewASTNode(wrappedValue, result.DSLType)
				if !yield(wrappedNode) { return }
			}
		}
	}
}
