package transformers

import (
	"context"
	"iter"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// StateT is a Monad Transformer that stacks the State Monad on top of the Stream (Seq) Monad.
// It transforms a stream of items while threading an immutable state S through each step.
func StateT[S, A, B any](initialState S, step func(S, A) (B, S)) func(iter.Seq[A]) iter.Seq[functions.Pair[B, S]] {
	return func(stream iter.Seq[A]) iter.Seq[functions.Pair[B, S]] {
		return func(yield func(functions.Pair[B, S]) bool) {
			current := initialState
			for a := range stream {
				b, nextS := step(current, a)
				current = nextS
				if !yield(functions.Pair[B, S]{Fst: b, Snd: current}) {
					return
				}
			}
		}
	}
}

// StateContextT is an interruptible, context-aware State Transformer.
// It yields (Node, State) while respecting Go context deadlines and cancellations.
func StateContextT[S, A any](
	ctx context.Context,
	initialState S,
	step func(context.Context, S, A) (A, S, error),
) func(iter.Seq[dsltypes.ASTNode[functions.Either[error, A]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, A]]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, A]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, A]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, A]]) bool) {
			current := initialState
			for node := range stream {
				if err := ctx.Err(); err != nil {
					errNode := dsltypes.NewASTNode(functions.Left[error, A](err), node.DSLType)
					yield(errNode)
					return
				}

				if node.Value.IsLeft() {
					if !yield(node) {
						return
					}
					continue
				}

				val, _ := node.Value.UnwrapRight()
				nextVal, nextS, err := step(ctx, current, val)
				if err != nil {
					errNode := dsltypes.NewASTNode(functions.Left[error, A](err), node.DSLType)
					yield(errNode)
					return
				}
				current = nextS
				wrapped := dsltypes.NewASTNode(functions.Right[error, A](nextVal), node.DSLType)
				if !yield(wrapped) {
					return
				}
			}
		}
	}
}
