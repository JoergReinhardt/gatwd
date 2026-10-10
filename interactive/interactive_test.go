package interactive_test

import (
	"testing"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/interactive"
)

func TestInteractiveREPLPause(t *testing.T) {
	engine := evaluator.NewEngine[string]()

	const IsUnknown = dsltypes.BitFlag(1 << 5)
	const IsDiscovered = dsltypes.BitFlag(1 << 6)

	// Simulate the Phase 5 "OnMissingAdapter" Pause
	// When the engine hits a node it doesn't know, this callback fires.
	engine.OnMissingAdapter = func(flag dsltypes.BitFlag) evaluator.Adapter[string] {
		// In a real REPL, this would block on bufio.NewReader(os.Stdin).ReadString('\n')
		// For the test, we simulate the user typing: "echo mock_child_of_%s"
		simulatedUserInput := "echo mock_child_of_%s"
		
		// The user also assigns a type to the output
		simulatedOutputFlag := IsDiscovered

		// We return the generic exec adapter based on user input
		return interactive.GenericExecAdapter(simulatedUserInput, simulatedOutputFlag)
	}

	// Create initial stream with an Unknown node
	initialStream := func(yield func(dsltypes.ASTNode[string]) bool) {
		yield(dsltypes.NewASTNode("RootNode", IsUnknown))
	}

	// Evaluate!
	// 1. The engine hits "RootNode" (IsUnknown).
	// 2. It finds no adapter.
	// 3. It pauses the iterator, calls OnMissingAdapter.
	// 4. The callback "prompts the user", generates an exec adapter, and registers it.
	// 5. The engine resumes, executing `echo mock_child_of_RootNode`.
	// 6. The exec adapter yields the result tagged with IsDiscovered.
	resultStream := engine.Evaluate(initialStream)

	var results []string
	for node := range resultStream {
		if !node.Match(IsDiscovered) {
			t.Fatalf("Expected output to have IsDiscovered flag")
		}
		results = append(results, node.Value)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result from echo command, got %d", len(results))
	}
	if results[0] != "mock_child_of_RootNode" {
		t.Fatalf("Unexpected execution output: %s", results[0])
	}
}
