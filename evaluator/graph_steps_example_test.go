package evaluator_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// ExampleStepGraphNodes shows the graph construction/projection steps composing in
// a pipeline: a seed node is turned into an empty graph, two colored nodes and a
// directed edge are added, and the graph is then unfolded back into its node
// identities — the DSL `spawn … |> graph_empty |> graph_add_node … |> graph_nodes`
// surface, exercised directly over the Either[error, any] stream.
func ExampleStepGraphNodes() {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)

	source := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		yield(dsltypes.NewASTNode(functions.Right[error, any](any("seed")), flag))
	}

	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepGraphEmpty(),
		evaluator.StepGraphAddNode("Service", "web"),
		evaluator.StepGraphAddNode("Pod", "nginx"),
		evaluator.StepGraphAddEdge("Selects", 1, "web", "nginx"),
		evaluator.StepGraphNodes(),
	}

	for node := range engine.EvaluatePipeline(source, pipeline) {
		val, _ := node.Value.UnwrapRight()
		fmt.Println(val)
	}

	// Output:
	// nginx
	// web
}

// ExampleStepGraphShortestPath shows a weighted walk: Dijkstra prefers the two-hop
// a→b→c route (cost 1+2=3) over the direct a→c edge (cost 5), so the projected path
// threads through b.
func ExampleStepGraphShortestPath() {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)

	source := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		yield(dsltypes.NewASTNode(functions.Right[error, any](any("seed")), flag))
	}

	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepGraphEmpty(),
		evaluator.StepGraphAddNode("N", "a"),
		evaluator.StepGraphAddNode("N", "b"),
		evaluator.StepGraphAddNode("N", "c"),
		evaluator.StepGraphAddEdge("E", 1, "a", "b"),
		evaluator.StepGraphAddEdge("E", 2, "b", "c"),
		evaluator.StepGraphAddEdge("E", 5, "a", "c"),
		evaluator.StepGraphShortestPath("a", "c"),
	}

	for node := range engine.EvaluatePipeline(source, pipeline) {
		val, _ := node.Value.UnwrapRight()
		fmt.Println(val)
	}

	// Output:
	// a
	// b
	// c
}

// ExampleStepGraphFoldNodes shows the node Functor and Catamorphism composing:
// every node's payload is rewritten to 10 by MapNodes, then FoldNodes sums the
// three payloads to 30 — the DSL `… |> graph_map_nodes (\p -> 10) |>
// graph_fold_nodes (\acc p -> acc + p) 0` surface.
func ExampleStepGraphFoldNodes() {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)

	source := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		yield(dsltypes.NewASTNode(functions.Right[error, any](any("seed")), flag))
	}

	add := func(acc, payload any) any { return acc.(int) + payload.(int) }
	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepGraphEmpty(),
		evaluator.StepGraphAddNode("N", "a"),
		evaluator.StepGraphAddNode("N", "b"),
		evaluator.StepGraphAddNode("N", "c"),
		evaluator.StepGraphMapNodes(func(any) any { return 10 }),
		evaluator.StepGraphFoldNodes(0, add),
	}

	for node := range engine.EvaluatePipeline(source, pipeline) {
		val, _ := node.Value.UnwrapRight()
		fmt.Println(val)
	}

	// Output:
	// 30
}
