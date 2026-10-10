package main

import (
	"fmt"
	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

func main() {
	fmt.Println("=== Native Engine API: The Functional Pantheon ===")


	// 1. Sets and Tuples
	fmt.Println("\n-- 1. Sets & Deduplication --")
	set := collections.NewSet[string]()
	set = set.Add("Node_A").Add("Node_B").Add("Node_A")
	fmt.Printf("Set contains Node_A: %v\n", set.Contains("Node_A"))

	// 2. EitherT and MaybeT Mechanics
	fmt.Println("\n-- 2. Error Transformers --")
	val := functions.Right[error, any]("Healthy Node")
	fmt.Printf("Either Right: %v\n", val.IsRight())

	// 3. StepFold (Catamorphism) via Pipeline
	fmt.Println("\n-- 3. Folding Streams --")
	stream := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		yield(dsltypes.NewASTNode(functions.Right[error, any](any("A")), 0))
		yield(dsltypes.NewASTNode(functions.Right[error, any](any("B")), 0))
		yield(dsltypes.NewASTNode(functions.Right[error, any](any("C")), 0))
	}
	
	foldStep := evaluator.StepFold("count")
	for node := range foldStep(stream) {
		v, _ := node.Value.UnwrapRight()
		fmt.Printf("Fold Count: %v\n", v)
	}

	fmt.Println("\nNotice how the engine natively supports these mathematically!")
}
