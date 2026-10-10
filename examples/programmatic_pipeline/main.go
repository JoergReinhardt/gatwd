package main

import (
	"fmt"
	"iter"
	"strings"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

func main() {
	fmt.Println("=== gatwd Programmatic Pipeline Example ===")
	fmt.Println("Demonstrating EitherT Auditing, StepTake, and StepExpect\n")

	// 1. We instantiate an Engine capable of handling Errors (EitherT)
	engine := evaluator.NewEngine[functions.Either[error, string]]()

	// 2. Define a Mock Adapter: Returns a stream of Pod names
	getPodsAdapter := func(node dsltypes.ASTNode[string]) iter.Seq[dsltypes.ASTNode[string]] {
		return func(yield func(dsltypes.ASTNode[string]) bool) {
			yield(dsltypes.NewASTNode("nginx-pod-01", node.DSLType))
			yield(dsltypes.NewASTNode("redis-pod-02", node.DSLType))
			yield(dsltypes.NewASTNode("root-shell-pod-03", node.DSLType)) // A bad pod!
			yield(dsltypes.NewASTNode("nginx-pod-04", node.DSLType))
		}
	}

	// 3. Wrap it in EitherT to make it safe for the pipeline
	safeAdapter := evaluator.EitherT(getPodsAdapter)
	
	// 4. Create an Auditing Gate (expect)
	// It checks if the string contains "root". If it does, it transforms the node 
	// into an explicit Error!
	isSafe := func(s string) bool { return !strings.Contains(s, "root") }
	securityGate := evaluator.StepExpect(isSafe, fmt.Errorf("CRITICAL SECURITY VIOLATION: Root pod detected!"))

	// 5. Build the Pipeline: spawn |> getPods |> expect |> take 3
	pipeline := []evaluator.PipelineStep[functions.Either[error, string]]{
		evaluator.StepAdapter(safeAdapter),
		securityGate,
		evaluator.StepTake[functions.Either[error, string]](3), // Bounding transducer!
	}

	// 6. Spawn the Root Node (Target Namespace)
	rootFlag := dsltypes.BitFlag(1)
	rootValue := functions.Right[error, string]("default-namespace")
	
	initialStream := func(yield func(dsltypes.ASTNode[functions.Either[error, string]]) bool) {
		yield(dsltypes.NewASTNode(rootValue, rootFlag))
	}

	// 7. Evaluate!
	terminalStream := engine.EvaluatePipeline(initialStream, pipeline)

	fmt.Println("--- Pipeline Execution Output ---")
	for node := range terminalStream {
		if node.Value.IsLeft() {
			err, _ := node.Value.UnwrapLeft()
			fmt.Printf("[AUDIT FAILED] %v\n", err)
		} else {
			val, _ := node.Value.UnwrapRight()
			fmt.Printf("[PASSED] Node -> %s\n", val)
		}
	}
}
