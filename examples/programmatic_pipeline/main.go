package main

import (
	"fmt"
	"iter"
	"strings"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func main() {
	header := visualizer.Format("gatwd Programmatic Pipeline: EitherT Auditing & Security Gate", visualizer.ParseOptions("--heavy --border --color=cyan"))
	fmt.Print(header)

	// 1. Instantiate an Engine capable of handling Errors (EitherT)
	engine := evaluator.NewEngine[functions.Either[error, string]]()

	// 2. Define a Mock Adapter: Returns a stream of Pod names
	podFlag := dsltypes.BitFlag(1 << 2)
	dsltypes.RegisterTypeFlag(podFlag, "KubernetesPod")

	getPodsAdapter := func(node dsltypes.ASTNode[string]) iter.Seq[dsltypes.ASTNode[string]] {
		return func(yield func(dsltypes.ASTNode[string]) bool) {
			yield(dsltypes.NewASTNode("nginx-pod-01", podFlag))
			yield(dsltypes.NewASTNode("redis-pod-02", podFlag))
			yield(dsltypes.NewASTNode("root-shell-pod-03", podFlag)) // A bad pod!
			yield(dsltypes.NewASTNode("nginx-pod-04", podFlag))
		}
	}

	// 3. Wrap it in EitherT to make it safe for the pipeline
	safeAdapter := evaluator.EitherT(getPodsAdapter)

	// 4. Create an Auditing Gate (expect)
	isSafe := func(s string) bool { return !strings.Contains(s, "root") }
	securityGate := evaluator.StepExpect(isSafe, fmt.Errorf("CRITICAL SECURITY VIOLATION: Root pod detected!"))

	// 5. Build the Pipeline: spawn |> getPods |> expect |> take 3
	pipeline := []evaluator.PipelineStep[functions.Either[error, string]]{
		evaluator.StepAdapter(safeAdapter),
		securityGate,
		evaluator.StepTake[functions.Either[error, string]](3),
	}

	// 6. Spawn the Root Node
	rootFlag := dsltypes.BitFlag(1)
	dsltypes.RegisterTypeFlag(rootFlag, "Namespace")
	rootValue := functions.Right[error, string]("default-namespace")

	initialStream := func(yield func(dsltypes.ASTNode[functions.Either[error, string]]) bool) {
		yield(dsltypes.NewASTNode(rootValue, rootFlag))
	}

	// 7. Evaluate and pretty-print stream events
	terminalStream := engine.EvaluatePipeline(initialStream, pipeline)

	fmt.Println("\n--- Pipeline Execution Stream ---")
	optsPass := visualizer.ParseOptions("--curved --border --color=green")
	optsFail := visualizer.ParseOptions("--heavy --border --color=red")

	for node := range terminalStream {
		if node.Value.IsLeft() {
			err, _ := node.Value.UnwrapLeft()
			fmt.Print(visualizer.FormatAST(node.DSLType, err, optsFail))
		} else {
			val, _ := node.Value.UnwrapRight()
			type PodStatus struct {
				Pod    string
				Status string
			}
			rec := PodStatus{Pod: val, Status: "PASSED_AUDIT"}
			fmt.Print(visualizer.FormatAST(node.DSLType, rec, optsPass))
		}
	}
}
