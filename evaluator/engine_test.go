package evaluator_test

import (
	"fmt"
	"iter"
	"testing"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
)

func TestEngineEvaluation(t *testing.T) {
	// 1. Define our dynamic BitFlag quasi-types
	const (
		IsHost = dsltypes.BitFlag(1 << iota)
		IsDatabase
		IsUnknown
	)

	engine := evaluator.NewEngine[string]()

	// 2. Register an adapter for Host: When evaluating a Host, it "discovers" two Databases
	engine.Register(IsHost, func(node dsltypes.ASTNode[string]) iter.Seq[dsltypes.ASTNode[string]] {
		return func(yield func(dsltypes.ASTNode[string]) bool) {
			db1 := dsltypes.NewASTNode(node.Value+"_db1", IsDatabase)
			db2 := dsltypes.NewASTNode(node.Value+"_db2", IsDatabase)
			yield(db1)
			yield(db2)
		}
	})

	// 3. Register an adapter for Database: When evaluating a Database, it discovers nothing (halts)
	engine.Register(IsDatabase, func(node dsltypes.ASTNode[string]) iter.Seq[dsltypes.ASTNode[string]] {
		return func(yield func(dsltypes.ASTNode[string]) bool) {
			// e.g. run a query, return results, but here we just yield a terminal string
			terminal := dsltypes.NewASTNode(node.Value+"_data", IsUnknown)
			yield(terminal)
		}
	})

	// 4. Create an initial stream of 2 Hosts
	initialStream := func(yield func(dsltypes.ASTNode[string]) bool) {
		yield(dsltypes.NewASTNode("hostA", IsHost))
		yield(dsltypes.NewASTNode("hostB", IsHost))
	}

	// 5. Evaluate Layer 1 (Hosts -> Databases)
	dbStream := engine.Evaluate(initialStream)
	
	// 6. Evaluate Layer 2 (Databases -> Terminals)
	dataStream := engine.Evaluate(dbStream)

	// Collect the results
	var results []string
	for node := range dataStream {
		results = append(results, node.Value)
	}

	if len(results) != 4 {
		t.Fatalf("Expected 4 terminal nodes, got %d", len(results))
	}
	
	expected := map[string]bool{
		"hostA_db1_data": true,
		"hostA_db2_data": true,
		"hostB_db1_data": true,
		"hostB_db2_data": true,
	}

	for _, res := range results {
		if !expected[res] {
			t.Fatalf("Unexpected terminal result: %s", res)
		}
	}
}

func TestParEngineEvaluation(t *testing.T) {
	const IsHost = dsltypes.BitFlag(1 << 0)
	engine := evaluator.NewEngine[string]()

	engine.Register(IsHost, func(node dsltypes.ASTNode[string]) iter.Seq[dsltypes.ASTNode[string]] {
		return func(yield func(dsltypes.ASTNode[string]) bool) {
			for i := 0; i < 5; i++ {
				child := dsltypes.NewASTNode(fmt.Sprintf("%s_child%d", node.Value, i), 0)
				if !yield(child) {
					return
				}
			}
		}
	})

	initialStream := func(yield func(dsltypes.ASTNode[string]) bool) {
		yield(dsltypes.NewASTNode("A", IsHost))
		yield(dsltypes.NewASTNode("B", IsHost))
	}

	// ParEvaluate (concurrently map Hosts to Children)
	childStream := engine.ParEvaluate(initialStream, 2)

	count := 0
	for range childStream {
		count++
	}

	if count != 10 { // 2 hosts * 5 children
		t.Fatalf("Expected 10 children, got %d", count)
	}
}
