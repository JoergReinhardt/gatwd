package evaluator

import (
	"iter"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// MockGenerator defines a function that generates a mock value given an index.
type MockGenerator[T any] func(index int) T

// GenerateStream creates a lazy iter.Seq of ASTNodes by repeatedly calling the generator function.
//
// This dogfoods our own stream architecture to produce mocked datasets for unit tests,
// benchmarks, and integration pipelines without allocating an entire slice in memory upfront.
func GenerateStream[T any](
	generator MockGenerator[T],
	count int,
	flag dsltypes.BitFlag,
) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
	return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		for i := 0; i < count; i++ {
			val := generator(i)
			node := dsltypes.NewASTNode(
				functions.Right[error, any](any(val)),
				flag,
			)
			if !yield(node) {
				return
			}
		}
	}
}

// MockUser is a structured domain type used to test pipelines handling complex Go structs.
type MockUser struct {
	ID     int
	Name   string
	Role   string
	Score  float64
	Active bool
}

// GenerateMockUsers generates a lazy stream of MockUser records.
func GenerateMockUsers(count int, flag dsltypes.BitFlag) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
	dsltypes.RegisterTypeFlag(flag, "MockUser")
	roles := []string{"admin", "editor", "viewer", "guest"}
	return GenerateStream(func(i int) MockUser {
		return MockUser{
			ID:     i,
			Name:   roles[i%len(roles)] + "_" + string(rune('A'+(i%26))),
			Role:   roles[i%len(roles)],
			Score:  float64(i) * 1.5,
			Active: i%2 == 0,
		}
	}, count, flag)
}
