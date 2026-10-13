package evaluator_test

import (
	"fmt"
	"iter"
	"testing"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

type enode = dsltypes.ASTNode[functions.Either[error, any]]

func intNode(v any) enode {
	return dsltypes.NewASTNode(functions.Right[error, any](v), 0)
}

func emitAdapter(vals ...any) evaluator.Adapter[functions.Either[error, any]] {
	return func(enode) iter.Seq[enode] {
		return func(yield func(enode) bool) {
			for _, v := range vals {
				if !yield(intNode(v)) {
					return
				}
			}
		}
	}
}

// TestStepPlusMergesSourcesPerNode asserts on the raw yielded values: each incoming
// node is expanded by every branch in order, Left nodes pass through, and an empty
// plus is the zero arrow.
func TestStepPlusMergesSourcesPerNode(t *testing.T) {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	source := func(yield func(enode) bool) {
		yield(intNode("a"))
		yield(dsltypes.NewASTNode(functions.Left[error, any](fmt.Errorf("boom")), 0))
		yield(intNode("b"))
	}
	step := evaluator.StepPlus(emitAdapter(1, 2), emitAdapter(), emitAdapter("x"))
	var got []any
	var errs int
	for n := range engine.EvaluatePipeline(source, []evaluator.PipelineStep[functions.Either[error, any]]{step}) {
		if n.Value.IsLeft() {
			errs++
			continue
		}
		v, _ := n.Value.UnwrapRight()
		got = append(got, v)
	}
	want := []any{1, 2, "x", 1, 2, "x"}
	if fmt.Sprint(got) != fmt.Sprint(want) || errs != 1 {
		t.Fatalf("got %v (errors %d), want %v and 1 error", got, errs, want)
	}

	zero := evaluator.StepPlus()
	n := 0
	for range engine.EvaluatePipeline(func(yield func(enode) bool) { yield(intNode(1)) }, []evaluator.PipelineStep[functions.Either[error, any]]{zero}) {
		n++
	}
	if n != 0 {
		t.Fatalf("StepPlus() must be the zero arrow, yielded %d", n)
	}
}
