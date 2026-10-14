package evaluator_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// ExampleStepGraphNodes shows the graph construction/projection steps composing in
// a pipeline: a seed node is turned into an empty graph, two colored nodes and a
// directed edge are added, and the graph is then unfolded back into its node
// identities — the DSL `spawn … |> graph_empty |> graph_add_node … |> graph_nodes`
// surface, exercised directly over the Either[error, any] stream. The ids are qualified
// (Kind/scope/name) so that a Service and a Pod called alike stay two nodes.
func ExampleStepGraphNodes() {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)

	source := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		yield(dsltypes.NewASTNode(functions.Right[error, any](any("seed")), flag))
	}

	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepGraphEmpty(),
		evaluator.StepGraphAddNode("Service", "Service/default/web"),
		evaluator.StepGraphAddNode("Pod", "Pod/default/nginx"),
		evaluator.StepGraphAddEdge("Selects", 1, "Service/default/web", "Pod/default/nginx"),
		evaluator.StepGraphNodes(),
	}

	for node := range engine.EvaluatePipeline(source, pipeline) {
		val, _ := node.Value.UnwrapRight()
		fmt.Println(val)
	}

	// Output:
	// Pod/default/nginx
	// Service/default/web
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

// ExampleStepGraphOver shows the single-node payload Lens: graph_set pins node a's
// payload to 10, graph_over doubles just that node's payload via the lens (b is
// untouched), and graph_focus reads the result back — the focused-rewrite surface
// `… |> graph_set a 10 |> graph_over a (\p -> p * 2) |> graph_focus a`.
func ExampleStepGraphOver() {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)

	source := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		yield(dsltypes.NewASTNode(functions.Right[error, any](any("seed")), flag))
	}

	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepGraphEmpty(),
		evaluator.StepGraphAddNode("N", "a"),
		evaluator.StepGraphAddNode("N", "b"),
		evaluator.StepGraphSet("a", 10),
		evaluator.StepGraphOver("a", func(p any) any { return p.(int) * 2 }),
		evaluator.StepGraphFocus("a"),
	}

	for node := range engine.EvaluatePipeline(source, pipeline) {
		val, _ := node.Value.UnwrapRight()
		fmt.Println(val)
	}

	// Output:
	// 20
}

// ExampleStepGraphDegree shows that degree counts EDGES, not distinct
// neighbors: b has one out-edge (to c) and two in-edges (from a and from the
// parallel a→b edge), so its total degree is 3 though it has only one
// predecessor and one successor.
func ExampleStepGraphDegree() {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)

	source := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		yield(dsltypes.NewASTNode(functions.Right[error, any](any("seed")), flag))
	}

	build := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepGraphEmpty(),
		evaluator.StepGraphAddNode("N", "a"),
		evaluator.StepGraphAddNode("N", "b"),
		evaluator.StepGraphAddNode("N", "c"),
		evaluator.StepGraphAddEdge("E", 1, "a", "b"),
		evaluator.StepGraphAddEdge("E", 1, "a", "b"), // a second, parallel a→b edge
		evaluator.StepGraphAddEdge("E", 1, "b", "c"),
	}

	run := func(extra evaluator.PipelineStep[functions.Either[error, any]]) any {
		pipeline := append(append([]evaluator.PipelineStep[functions.Either[error, any]]{}, build...), extra)
		var out any
		for node := range engine.EvaluatePipeline(source, pipeline) {
			out, _ = node.Value.UnwrapRight()
		}
		return out
	}

	fmt.Println(run(evaluator.StepGraphInDegree("b")))
	fmt.Println(run(evaluator.StepGraphOutDegree("b")))
	fmt.Println(run(evaluator.StepGraphDegree("b")))

	// Output:
	// 2
	// 1
	// 3
}

// ExampleStepGraphReachableTo shows the mirror of StepGraphReachable: everyone
// upstream of c (a and b depend on it, transitively through b), where
// StepGraphReachable from a would instead show everything a depends on.
func ExampleStepGraphReachableTo() {
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
		evaluator.StepGraphAddEdge("E", 1, "b", "c"),
		evaluator.StepGraphReachableTo("c"),
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

// ExampleStepGraphUnroll shows the graph→tree unrolling DUPLICATING d, reached
// via both b and c, instead of collapsing it to the single node the graph has:
// the rendered tree shows two distinct leaves named d, one per path.
func ExampleStepGraphUnroll() {
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
		evaluator.StepGraphAddNode("N", "d"),
		evaluator.StepGraphAddEdge("E", 1, "a", "b"),
		evaluator.StepGraphAddEdge("E", 1, "a", "c"),
		evaluator.StepGraphAddEdge("E", 1, "b", "d"),
		evaluator.StepGraphAddEdge("E", 1, "c", "d"),
		evaluator.StepGraphUnroll("a"),
	}

	var render func(t *collections.Tree[any]) string
	render = func(t *collections.Tree[any]) string {
		kids := t.Children().Elements()
		if len(kids) == 0 {
			return fmt.Sprint(t.Value())
		}
		out := fmt.Sprint(t.Value()) + "("
		for i, k := range kids {
			if i > 0 {
				out += ","
			}
			out += render(k.(*collections.Tree[any]))
		}
		return out + ")"
	}

	for node := range engine.EvaluatePipeline(source, pipeline) {
		val, _ := node.Value.UnwrapRight()
		fmt.Println(render(val.(*collections.Tree[any])))
	}

	// Output:
	// a(b(d),c(d))
}
