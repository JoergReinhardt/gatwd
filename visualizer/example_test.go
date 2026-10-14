package visualizer_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/visualizer"
)

type Config struct {
	Host string
	Port int
	TLS  bool
}

func ExampleFormat_curved() {
	cfg := Config{
		Host: "localhost",
		Port: 8080,
		TLS:  true,
	}

	opts := visualizer.ParseOptions("--curved")
	fmt.Print(visualizer.Format(cfg, opts))

	// Output:
	// ╰─ Config
	//     ├─ Host: localhost
	//     ├─ Port: 8080
	//     ╰─ TLS: true
}

func ExampleFormat_heavy() {
	cfg := Config{
		Host: "0.0.0.0",
		Port: 443,
		TLS:  true,
	}

	opts := visualizer.ParseOptions("--heavy")
	fmt.Print(visualizer.Format(cfg, opts))

	// Output:
	// ┗━━ Config
	//     ┣━━ Host: 0.0.0.0
	//     ┣━━ Port: 443
	//     ┗━━ TLS: true
}

func ExampleFormat_border() {
	cfg := Config{
		Host: "127.0.0.1",
		Port: 9000,
		TLS:  false,
	}

	opts := visualizer.ParseOptions("--curved --border")
	fmt.Print(visualizer.Format(cfg, opts))

	// Output:
	// ╭────────────────────────╮
	// │ ╰─ Config              │
	// │     ├─ Host: 127.0.0.1 │
	// │     ├─ Port: 9000      │
	// │     ╰─ TLS: false      │
	// ╰────────────────────────╯
}

func ExampleFormat_hamt() {
	hamt := collections.NewHAMT[string, int]().Set("counter", 42)

	opts := visualizer.ParseOptions("--curved --border")
	fmt.Print(visualizer.Format(hamt, opts))

	// Output:
	// ╭────────────────────╮
	// │ ╰─ Record          │
	// │     ╰─ counter: 42 │
	// ╰────────────────────╯
}

func ExampleFormat_chunk() {
	chunk := collections.NewChunk[string](3)
	chunk.Append("alpha")
	chunk.Append("beta")

	opts := visualizer.ParseOptions("--curved --border")
	fmt.Print(visualizer.Format(chunk, opts))

	// Output:
	// ╭───────────────────╮
	// │ ╰─ List           │
	// │     ├─ [0]: alpha │
	// │     ╰─ [1]: beta  │
	// ╰───────────────────╯
}

func ExampleFormat_tree() {
	child1 := collections.Leaf("child_A")
	child2 := collections.Leaf("child_B")
	children := collections.Cons(child1, collections.Cons(child2, collections.Nil[*collections.Tree[string]]()))
	tree := collections.NewTree("root_node", children)

	opts := visualizer.ParseOptions("--curved")
	fmt.Print(visualizer.Format(tree, opts))

	// Output:
	// ╰─ root_node
	//     ├─ child_A
	//     ╰─ child_B
}

// ExampleFormat_graph shows a Graph rendering as a real directed diagram in --dot output (one
// box per node, real arrows for its edges) instead of the generic reflection walk every other
// type gets: a node reached by more than one edge is still drawn once, not once per incoming
// edge.
func ExampleFormat_graph() {
	g := collections.NewGraph[string, string, int, int]().
		AddNode("svc", "service", 0).AddNode("pod", "pod", 0).
		Connect("svc", "pod", "selects", 1)

	opts := visualizer.ParseOptions("--dot")
	fmt.Print(visualizer.Format(g, opts))
	// Output:
	// digraph G {
	//   node [shape=box, style=filled, color=black, fillcolor=white];
	//   "pod" [label="pod"];
	//   "svc" [label="svc"];
	//   "svc" -> "pod" [label="selects"];
	// }
}

func ExampleStepVisualize() {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)
	dsltypes.RegisterTypeFlag(flag, "Config")

	// Stream generating a single test record
	source := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		cfg := Config{Host: "gatwd.io", Port: 80, TLS: false}
		yield(dsltypes.NewASTNode(functions.Right[error, any](any(cfg)), flag))
	}

	// Pipeline with StepVisualize transducer
	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		visualizer.StepVisualize("--curved --border"),
	}

	for node := range engine.EvaluatePipeline(source, pipeline) {
		val, _ := node.Value.UnwrapRight()
		fmt.Print(val.(string))
	}

	// Output:
	// ╭───────────────────────╮
	// │ ╰─ Config             │
	// │     ├─ Host: gatwd.io │
	// │     ├─ Port: 80       │
	// │     ╰─ TLS: false     │
	// ╰───────────────────────╯
}
