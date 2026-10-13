package evaluator

import (
	"errors"
	"iter"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

type either = functions.Either[error, any]

func streamOf(vals ...either) iter.Seq[dsltypes.ASTNode[either]] {
	return func(yield func(dsltypes.ASTNode[either]) bool) {
		for _, v := range vals {
			if !yield(dsltypes.NewASTNode(v, 0)) {
				return
			}
		}
	}
}

func drain(seq iter.Seq[dsltypes.ASTNode[either]]) []either {
	var out []either
	for n := range seq {
		out = append(out, n.Value)
	}
	return out
}

// TestStepGraphInsertNode: the strict step declares a new id, and on a repeated id
// yields a Left holding a *collections.DuplicateNodeError rather than a graph.
func TestStepGraphInsertNode(t *testing.T) {
	g := any(emptyDSLGraph())
	step := StepGraphInsertNode("Pod", "Pod/default/web")

	first := drain(step(streamOf(functions.Right[error, any](g))))
	if len(first) != 1 || first[0].IsLeft() {
		t.Fatalf("first insertion: %v", first)
	}
	g1, _ := first[0].UnwrapRight()
	graph, ok := g1.(dslGraph)
	if !ok || !graph.HasNode("Pod/default/web") {
		t.Fatalf("result is %T, want a graph holding the node", g1)
	}

	second := drain(step(streamOf(functions.Right[error, any](g1))))
	if len(second) != 1 || !second[0].IsLeft() {
		t.Fatalf("duplicate insertion must be a Left, got %v", second)
	}
	err, _ := second[0].UnwrapLeft()
	var dup *collections.DuplicateNodeError
	if !errors.As(err, &dup) || dup.ID != "Pod/default/web" {
		t.Fatalf("Left is %v (%T), want a *DuplicateNodeError for the id", err, err)
	}

	// Left nodes and non-graph values pass through untouched, like every graph step
	passed := drain(step(streamOf(functions.Left[error, any](errors.New("upstream")), functions.Right[error, any]("not a graph"))))
	if len(passed) != 2 || !passed[0].IsLeft() || passed[1].IsLeft() {
		t.Errorf("pass-through broke: %v", passed)
	}
}

// TestStepCollectListKeepsErrors (Mandate 6): collect_list used to drop Left nodes,
// so an error upstream (a duplicate id, an unreadable manifest) came out as a
// plausible-looking list. Now the error nodes are passed on instead of the list.
func TestStepCollectListKeepsErrors(t *testing.T) {
	ok := func(v any) either { return functions.Right[error, any](v) }

	clean := drain(StepCollectList()(streamOf(ok(1), ok(2))))
	if len(clean) != 1 || clean[0].IsLeft() {
		t.Fatalf("clean stream: %v", clean)
	}

	boom := errors.New("boom: duplicate node id")
	failed := drain(StepCollectList()(streamOf(ok(1), functions.Left[error, any](boom), ok(2))))
	if len(failed) != 1 || !failed[0].IsLeft() {
		t.Fatalf("a stream with an error must yield the error, not a list: %v", failed)
	}
	if e, _ := failed[0].UnwrapLeft(); !strings.Contains(e.Error(), "duplicate") {
		t.Errorf("lost the error: %v", e)
	}
}
