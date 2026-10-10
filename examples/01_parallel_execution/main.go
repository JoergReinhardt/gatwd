package main

import (
	"fmt"
	"iter"
	"time"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
)

func main() {
	fmt.Println("=== Operator: ParBind (Parallel Execution / ||>) ===")
	engine := evaluator.NewEngine[string]()

	// An adapter that simulates a slow network request (e.g., pinging a server)
	slowPingAdapter := func(node dsltypes.ASTNode[string]) iter.Seq[dsltypes.ASTNode[string]] {
		return func(yield func(dsltypes.ASTNode[string]) bool) {
			time.Sleep(100 * time.Millisecond) // Simulated latency
			yield(dsltypes.NewASTNode(node.Value+"_pinged", node.DSLType))
		}
	}
	engine.Register(1, slowPingAdapter)

	// Create an initial stream of 50 servers
	initialStream := func(yield func(dsltypes.ASTNode[string]) bool) {
		for i := 1; i <= 50; i++ {
			yield(dsltypes.NewASTNode(fmt.Sprintf("Server_%d", i), 1))
		}
	}

	start := time.Now()
	// Evaluate the stream concurrently using 10 workers!
	fmt.Println("Pinging 50 servers concurrently using ParEvaluate (10 workers)...")
	terminalNodes := engine.ParEvaluate(initialStream, 10)

	count := 0
	for range terminalNodes {
		count++
	}
	
	duration := time.Since(start)
	fmt.Printf("Successfully processed %d nodes in %v\n", count, duration)
	fmt.Println("Notice how it took ~500ms instead of 5,000ms! That is the power of ParBind.")
}
