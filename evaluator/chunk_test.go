package evaluator

import (
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

func TestStepChunk(t *testing.T) {
	// Create a simple stream of 5 nodes
	nodes := []dsltypes.ASTNode[functions.Either[error, any]]{
		dsltypes.NewASTNode(functions.Right[error, any](any(1)), dsltypes.BitFlag(1)),
		dsltypes.NewASTNode(functions.Right[error, any](any(2)), dsltypes.BitFlag(1)),
		dsltypes.NewASTNode(functions.Right[error, any](any(3)), dsltypes.BitFlag(1)),
		dsltypes.NewASTNode(functions.Right[error, any](any(4)), dsltypes.BitFlag(1)),
		dsltypes.NewASTNode(functions.Right[error, any](any(5)), dsltypes.BitFlag(1)),
	}

	stream := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		for _, node := range nodes {
			if !yield(node) {
				return
			}
		}
	}

	chunkStep := StepChunk(2)
	chunkedStream := chunkStep(stream)

	var results []dsltypes.ASTNode[functions.Either[error, any]]
	for node := range chunkedStream {
		results = append(results, node)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 chunks, got %d", len(results))
	}

	// First chunk should have 2 nodes
	c1, ok := results[0].Value.UnwrapRight()
	if !ok {
		t.Fatalf("Expected Right value")
	}
	chunk1, ok := c1.(*collections.Chunk[dsltypes.ASTNode[functions.Either[error, any]]])
	if !ok || chunk1.Len() != 2 {
		t.Fatalf("Expected chunk of len 2, got %v", chunk1)
	}

	// Third chunk should have 1 node
	c3, ok := results[2].Value.UnwrapRight()
	if !ok {
		t.Fatalf("Expected Right value")
	}
	chunk3, ok := c3.(*collections.Chunk[dsltypes.ASTNode[functions.Either[error, any]]])
	if !ok || chunk3.Len() != 1 {
		t.Fatalf("Expected chunk of len 1, got %v", chunk3)
	}
}
