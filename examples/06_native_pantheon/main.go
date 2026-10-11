package main

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func main() {
	opts := visualizer.Options{Border: true, Style: visualizer.CurvedStyle}

	fmt.Println(visualizer.Format("Native Engine API: The Functional Pantheon", opts))

	// 1. Sets and Tuples
	set := collections.NewSet[string]()
	set = set.Add("Node_A").Add("Node_B").Add("Node_A")

	fmt.Println(visualizer.Format(map[string]any{
		"Feature":       "Set Deduplication",
		"ContainsNodeA": set.Contains("Node_A"),
		"Size":          set.Len(),
		"Items":         set.Elements(),
	}, opts))

	// 2. EitherT and MaybeT Mechanics
	val := functions.Right[error, any]("Healthy Node")
	rightVal, _ := val.UnwrapRight()

	fmt.Println(visualizer.Format(map[string]any{
		"Feature": "Error Monad (Either)",
		"IsRight": val.IsRight(),
		"Value":   rightVal,
	}, opts))

	// 3. StepFold (Catamorphism) via Pipeline
	stream := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		yield(dsltypes.NewASTNode(functions.Right[error, any](any("A")), 0))
		yield(dsltypes.NewASTNode(functions.Right[error, any](any("B")), 0))
		yield(dsltypes.NewASTNode(functions.Right[error, any](any("C")), 0))
	}

	foldStep := evaluator.StepFold("count")
	var foldCount any
	for node := range foldStep(stream) {
		foldCount, _ = node.Value.UnwrapRight()
	}

	fmt.Println(visualizer.Format(map[string]any{
		"Feature":   "Stream Fold (Catamorphism)",
		"Operation": "count",
		"Stream":    []string{"A", "B", "C"},
		"Result":    foldCount,
	}, opts))

	fmt.Println("Notice how the engine natively supports these mathematically!")
}
