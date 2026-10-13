package evaluator

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

var noSeed = func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {}

func TestStepAssertEmptyStreamIsAnError(t *testing.T) {
	engine := NewEngine[functions.Either[error, any]]()
	empty := StepEmit() // yields nothing
	nodes := engine.EvaluatePipeline(noSeed, []PipelineStep[functions.Either[error, any]]{empty, StepAssert("A")})
	var errs []error
	for n := range nodes {
		if n.Value.IsRight() {
			t.Fatalf("assert passed vacuously on an empty stream")
		}
		e, _ := n.Value.UnwrapLeft()
		errs = append(errs, e)
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "empty stream") {
		t.Fatalf("want exactly one 'empty stream' error, got %v", errs)
	}
}

func TestStepAssertNonEmptyStillPassesAndFails(t *testing.T) {
	engine := NewEngine[functions.Either[error, any]]()
	run := func(expected string) []bool {
		var rights []bool
		for n := range engine.EvaluatePipeline(noSeed, []PipelineStep[functions.Either[error, any]]{StepEmit("A"), StepAssert(expected)}) {
			rights = append(rights, n.Value.IsRight())
		}
		return rights
	}
	if got := run("A"); len(got) != 1 || !got[0] {
		t.Fatalf("matching assert: got %v, want one Right", got)
	}
	if got := run("B"); len(got) != 1 || got[0] {
		t.Fatalf("mismatching assert: got %v, want one Left", got)
	}
}
