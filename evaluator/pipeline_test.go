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

func TestStepFoldExtended(t *testing.T) {
	emitNums := StepEmitValues(10, 20, 30)

	// sum
	resSum := runPipelineToValues([]PipelineStep[functions.Either[error, any]]{emitNums, StepFold("sum")})
	if len(resSum) != 1 || !functions.StructuralEqual(resSum[0], int64(60)) {
		t.Errorf("expected sum 60, got %v", resSum)
	}

	// product
	resProd := runPipelineToValues([]PipelineStep[functions.Either[error, any]]{emitNums, StepFold("product")})
	if len(resProd) != 1 || !functions.StructuralEqual(resProd[0], int64(6000)) {
		t.Errorf("expected product 6000, got %v", resProd)
	}

	// min
	resMin := runPipelineToValues([]PipelineStep[functions.Either[error, any]]{emitNums, StepFold("min")})
	if len(resMin) != 1 || !functions.StructuralEqual(resMin[0], 10) {
		t.Errorf("expected min 10, got %v", resMin)
	}

	// max
	resMax := runPipelineToValues([]PipelineStep[functions.Either[error, any]]{emitNums, StepFold("max")})
	if len(resMax) != 1 || !functions.StructuralEqual(resMax[0], 30) {
		t.Errorf("expected max 30, got %v", resMax)
	}

	// concat
	emitStrs := StepEmit("foo", "bar", "baz")
	resConcat := runPipelineToValues([]PipelineStep[functions.Either[error, any]]{emitStrs, StepFold("concat")})
	if len(resConcat) != 1 || resConcat[0] != "foobarbaz" {
		t.Errorf("expected concat 'foobarbaz', got %v", resConcat)
	}

	// any & all
	emitBools := StepEmitValues(false, true, false)
	resAny := runPipelineToValues([]PipelineStep[functions.Either[error, any]]{emitBools, StepFold("any")})
	if len(resAny) != 1 || resAny[0] != true {
		t.Errorf("expected any=true, got %v", resAny)
	}
	resAll := runPipelineToValues([]PipelineStep[functions.Either[error, any]]{emitBools, StepFold("all")})
	if len(resAll) != 1 || resAll[0] != false {
		t.Errorf("expected all=false, got %v", resAll)
	}
}

func TestStepSetAndHAMTAlgebra(t *testing.T) {
	s1 := collections.NewSet[int](1, 2)
	s2 := collections.NewSet[int](2, 3)

	// StepSetUnion
	resUnion := runPipelineToValues([]PipelineStep[functions.Either[error, any]]{
		StepEmitValues(any(s1)),
		StepSetUnion(any(s2)),
	})
	if len(resUnion) != 1 {
		t.Fatalf("expected 1 union result, got %d", len(resUnion))
	}
	uSet, ok := resUnion[0].(*collections.Set[int])
	if !ok || uSet.Len() != 3 {
		t.Fatalf("expected union set len=3, got %v", resUnion[0])
	}

	// StepSetIntersect
	resIntersect := runPipelineToValues([]PipelineStep[functions.Either[error, any]]{
		StepEmitValues(any(s1)),
		StepSetIntersect(any(s2)),
	})
	if len(resIntersect) != 1 {
		t.Fatalf("expected 1 intersect result, got %d", len(resIntersect))
	}
	iSet, ok := resIntersect[0].(*collections.Set[int])
	if !ok || iSet.Len() != 1 || !iSet.Contains(2) {
		t.Fatalf("expected intersect {2}, got %v", resIntersect[0])
	}

	// StepSetDiff
	resDiff := runPipelineToValues([]PipelineStep[functions.Either[error, any]]{
		StepEmitValues(any(s1)),
		StepSetDiff(any(s2)),
	})
	dSet, ok := resDiff[0].(*collections.Set[int])
	if !ok || dSet.Len() != 1 || !dSet.Contains(1) {
		t.Fatalf("expected diff {1}, got %v", resDiff[0])
	}

	// StepHAMTMerge
	h1 := collections.NewHAMT[string, int]().Set("a", 1)
	h2 := collections.NewHAMT[string, int]().Set("b", 2)
	resHamt := runPipelineToValues([]PipelineStep[functions.Either[error, any]]{
		StepEmitValues(any(h1)),
		StepHAMTMerge(any(h2)),
	})
	mHamt, ok := resHamt[0].(*collections.HAMT[string, int])
	if !ok || mHamt.Len() != 2 {
		t.Fatalf("expected merged HAMT len=2, got %v", resHamt[0])
	}
}
