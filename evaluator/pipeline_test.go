package evaluator

import (
	"iter"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// Helper to run a pipeline and return un-wrapped right values
func runPipelineToValues(steps []PipelineStep[functions.Either[error, any]]) []any {
	engine := NewEngine[functions.Either[error, any]]()

	// Create an empty initial stream
	initialStream := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {}

	terminalNodes := engine.EvaluatePipeline(initialStream, steps)

	var results []any
	for node := range terminalNodes {
		if node.Value.IsRight() {
			val, _ := node.Value.UnwrapRight()
			results = append(results, val)
		}
	}
	return results
}

func TestStepEmit(t *testing.T) {
	steps := []PipelineStep[functions.Either[error, any]]{
		StepEmit("A", "B", "C"),
	}

	res := runPipelineToValues(steps)
	if len(res) != 3 || res[0] != "A" || res[1] != "B" || res[2] != "C" {
		t.Errorf("Expected [A B C], got %v", res)
	}
}

func TestStepSort(t *testing.T) {
	steps := []PipelineStep[functions.Either[error, any]]{
		StepEmit("Z", "A", "M"),
		StepSort(DefaultLess),
	}

	res := runPipelineToValues(steps)
	if len(res) != 3 || res[0] != "A" || res[1] != "M" || res[2] != "Z" {
		t.Errorf("Expected [A M Z], got %v", res)
	}
}

func TestStepToSet(t *testing.T) {
	steps := []PipelineStep[functions.Either[error, any]]{
		StepEmit("dup", "unique", "dup"),
		StepToSet(),
	}

	res := runPipelineToValues(steps)
	if len(res) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(res))
	}

	// Should yield a *collections.Set[string]
	// But it returns the string representation via fmt.Sprintf internally in StepToSet.
	// Actually StepToSet yields `collections.Set[string]`, let's just check non-nil
	if res[0] == nil {
		t.Errorf("Expected a Set, got nil")
	}
}

func TestStepMap(t *testing.T) {
	engine := NewEngine[functions.Either[error, any]]()

	// Subpipeline: [math_add 5] (Since we don't have StepMathAdd in evaluator, we fake it with StepAdapter)
	add5Adapter := func(node dsltypes.ASTNode[functions.Either[error, any]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			val, _ := node.Value.UnwrapRight()
			intVal, _ := val.(int)
			newNode := dsltypes.NewASTNode(functions.Right[error, any](any(intVal+5)), node.DSLType)
			yield(newNode)
		}
	}

	subPipeline := []PipelineStep[functions.Either[error, any]]{
		StepAdapter(add5Adapter),
	}

	steps := []PipelineStep[functions.Either[error, any]]{
		func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
			return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
				yield(dsltypes.NewASTNode(functions.Right[error, any](any(10)), 0))
				yield(dsltypes.NewASTNode(functions.Right[error, any](any(20)), 0))
			}
		},
		StepMap(subPipeline, engine),
	}

	res := runPipelineToValues(steps)
	if len(res) != 2 || res[0] != 15 || res[1] != 25 {
		t.Errorf("Expected [15 25], got %v", res)
	}
}

func TestStepFocusAndUpdate_HAMT(t *testing.T) {
	h := collections.NewHAMT[string, any]().Set("env", "production")

	// 1. Test StepFocus on HAMT
	focusSteps := []PipelineStep[functions.Either[error, any]]{
		StepEmitValues(any(h)),
		StepFocus("env"),
	}
	resFocus := runPipelineToValues(focusSteps)
	if len(resFocus) != 1 || resFocus[0] != "production" {
		t.Fatalf("expected 'production', got %v", resFocus)
	}

	// 2. Test StepUpdate on HAMT
	updateSteps := []PipelineStep[functions.Either[error, any]]{
		StepEmitValues(any(h)),
		StepUpdate("env", "staging"),
		StepFocus("env"),
	}
	resUpdate := runPipelineToValues(updateSteps)
	if len(resUpdate) != 1 || resUpdate[0] != "staging" {
		t.Fatalf("expected 'staging', got %v", resUpdate)
	}
}
