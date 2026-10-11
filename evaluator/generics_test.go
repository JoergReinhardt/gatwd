package evaluator_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

func TestGenerics_DiverseTypes(t *testing.T) {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1 << 2)

	// 1. Testing Float64 Stream
	t.Run("Float64_Stream", func(t *testing.T) {
		floatStream := evaluator.GenerateStream(func(i int) float64 {
			return float64(i) * 0.5
		}, 10, flag)

		var results []float64
		for node := range floatStream {
			val, _ := node.Value.UnwrapRight()
			results = append(results, val.(float64))
		}
		if len(results) != 10 || results[2] != 1.0 {
			t.Fatalf("expected float results, got %v", results)
		}
	})

	// 2. Testing Complex Structs (MockUser)
	t.Run("Complex_Struct_MockUser", func(t *testing.T) {
		userStream := evaluator.GenerateMockUsers(20, flag)

		// Filter for active users with role "admin"
		filterStep := evaluator.StepFilter(func(val any) bool {
			user := val.(evaluator.MockUser)
			return user.Active && user.Role == "admin"
		})

		stream := engine.EvaluatePipeline(userStream, []evaluator.PipelineStep[functions.Either[error, any]]{filterStep})

		count := 0
		for node := range stream {
			count++
			val, _ := node.Value.UnwrapRight()
			user := val.(evaluator.MockUser)
			if !user.Active || user.Role != "admin" {
				t.Fatalf("unexpected user yielded: %+v", user)
			}
		}
		if count == 0 {
			t.Fatalf("expected at least one active admin")
		}
	})

	// 3. Testing Nested Persistent Collections (List and HAMT)
	t.Run("Nested_Collections_Stream", func(t *testing.T) {
		type Document struct {
			Title    string
			Tags     *collections.List[string]
			Metadata *collections.HAMT[string, any]
		}

		docStream := evaluator.GenerateStream(func(i int) Document {
			tags := collections.Cons("urgent", collections.Cons(fmt.Sprintf("v%d", i), collections.Nil[string]()))
			meta := collections.NewHAMT[string, any]().Set("version", i).Set("author", "gatwd")
			return Document{
				Title:    fmt.Sprintf("Doc_%d", i),
				Tags:     tags,
				Metadata: meta,
			}
		}, 5, flag)

		for node := range docStream {
			val, _ := node.Value.UnwrapRight()
			doc := val.(Document)
			h, _ := doc.Tags.Head()
			if h != "urgent" {
				t.Fatalf("expected tag 'urgent', got %q", h)
			}
			author, ok := doc.Metadata.Get("author")
			if !ok || author != "gatwd" {
				t.Fatalf("expected author 'gatwd', got %v", author)
			}
		}
	})

	// 4. Testing Chunking Over Diverse Structs
	t.Run("Chunking_Over_Structs", func(t *testing.T) {
		userStream := evaluator.GenerateMockUsers(25, flag)
		chunkStep := evaluator.StepChunk(10)

		stream := engine.EvaluatePipeline(userStream, []evaluator.PipelineStep[functions.Either[error, any]]{chunkStep})

		var chunkSizes []int
		for node := range stream {
			val, _ := node.Value.UnwrapRight()
			c := val.(*collections.Chunk[dsltypes.ASTNode[functions.Either[error, any]]])
			chunkSizes = append(chunkSizes, c.Len())
		}

		if len(chunkSizes) != 3 || chunkSizes[0] != 10 || chunkSizes[1] != 10 || chunkSizes[2] != 5 {
			t.Fatalf("expected chunk sizes [10 10 5], got %v", chunkSizes)
		}
	})

	// 5. Testing Error Propagation Across Generics
	t.Run("Error_Short_Circuit", func(t *testing.T) {
		streamWithErr := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			yield(dsltypes.NewASTNode(functions.Right[error, any](any(1)), flag))
			yield(evaluator.InjectError[any](fmt.Errorf("synthetic pipeline failure"), flag))
			yield(dsltypes.NewASTNode(functions.Right[error, any](any(3)), flag))
		}

		takeStep := evaluator.StepTake[functions.Either[error, any]](3)
		stream := engine.EvaluatePipeline(streamWithErr, []evaluator.PipelineStep[functions.Either[error, any]]{takeStep})

		var encounteredError bool
		for node := range stream {
			if node.Value.IsLeft() {
				err, _ := node.Value.UnwrapLeft()
				if strings.Contains(err.Error(), "synthetic pipeline failure") {
					encounteredError = true
				}
			}
		}
		if !encounteredError {
			t.Fatalf("expected error node to pass through cleanly")
		}
	})
}
