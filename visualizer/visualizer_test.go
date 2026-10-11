package visualizer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/visualizer"
)

type customStringer struct {
	msg string
}

func (c customStringer) String() string {
	return "CustomStringer(" + c.msg + ")"
}

type customErr struct {
	detail string
}

func (e customErr) Error() string {
	return "CustomErrorDetail: " + e.detail
}

func TestVisualizer_StringerAndError(t *testing.T) {
	opts := visualizer.Options{}

	// 1. Stringer formatted directly
	s := customStringer{msg: "hello-world"}
	formattedStringer := visualizer.Format(s, opts)
	if !strings.Contains(formattedStringer, "CustomStringer(hello-world)") {
		t.Fatalf("expected stringer output, got %q", formattedStringer)
	}

	// 2. Error formatted directly (should call .Error() and not expose raw reflect unexported fields)
	err := customErr{detail: "database timeout"}
	formattedErr := visualizer.Format(err, opts)
	if !strings.Contains(formattedErr, "CustomErrorDetail: database timeout") {
		t.Fatalf("expected error message in visualizer, got %q", formattedErr)
	}

	// 3. Standard errors.New
	stdErr := errors.New("sentinel failure")
	formattedStdErr := visualizer.Format(stdErr, opts)
	if !strings.Contains(formattedStdErr, "sentinel failure") {
		t.Fatalf("expected sentinel failure, got %q", formattedStdErr)
	}
}

func TestVisualizer_StepVisualize_ErrorNode(t *testing.T) {
	step := visualizer.StepVisualize("--border")
	engine := evaluator.NewEngine[functions.Either[error, any]]()

	trigger := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		errNode := dsltypes.NewASTNode(
			functions.Left[error, any](errors.New("upstream service unavailable")),
			dsltypes.BitFlag(1),
		)
		yield(errNode)
	}

	stream := engine.EvaluatePipeline(trigger, []evaluator.PipelineStep[functions.Either[error, any]]{step})

	count := 0
	for node := range stream {
		count++
		if !node.Value.IsLeft() {
			t.Fatalf("expected node to remain Left, got Right")
		}
		err, _ := node.Value.UnwrapLeft()
		errMsg := err.Error()
		if !strings.Contains(errMsg, "upstream service unavailable") {
			t.Fatalf("expected formatted error box containing 'upstream service unavailable', got %q", errMsg)
		}
		// Check that border was applied
		if !strings.Contains(errMsg, "╭") && !strings.Contains(errMsg, "┌") && !strings.Contains(errMsg, "+") {
			t.Fatalf("expected visual box border in error message, got %q", errMsg)
		}
	}
	if count != 1 {
		t.Fatalf("expected 1 visualized error node, got %d", count)
	}
}

func TestASTNode_Stringer(t *testing.T) {
	// Right node with int
	nodeRight := dsltypes.NewASTNode(
		functions.Right[error, any](any(42)),
		dsltypes.BitFlag(1),
	)
	if nodeRight.String() != "42" {
		t.Fatalf("expected '42', got %q", nodeRight.String())
	}

	// Left node with error
	nodeLeft := dsltypes.NewASTNode(
		functions.Left[error, any](errors.New("invalid parameter")),
		dsltypes.BitFlag(1),
	)
	if !strings.Contains(nodeLeft.String(), "[ERROR] invalid parameter") {
		t.Fatalf("expected '[ERROR] invalid parameter', got %q", nodeLeft.String())
	}
}

func TestVisualizer_CollectionsGraphable(t *testing.T) {
	opts := visualizer.Options{}

	// HAMT
	hamt := collections.NewHAMT[string, int]().Set("alpha", 1).Set("beta", 2)
	hamtVis := visualizer.Format(hamt, opts)
	if !strings.Contains(hamtVis, "alpha: 1") || !strings.Contains(hamtVis, "beta: 2") {
		t.Fatalf("expected HAMT visualization to contain keys, got:\n%s", hamtVis)
	}

	// Chunk
	chunk := collections.NewChunk[string](3)
	chunk.Append("first")
	chunk.Append("second")
	chunkVis := visualizer.Format(chunk, opts)
	if !strings.Contains(chunkVis, "first") || !strings.Contains(chunkVis, "second") {
		t.Fatalf("expected Chunk visualization to contain items, got:\n%s", chunkVis)
	}

	// Set
	set := collections.NewSet("foo", "bar")
	setVis := visualizer.Format(set, opts)
	if !strings.Contains(setVis, "foo") || !strings.Contains(setVis, "bar") {
		t.Fatalf("expected Set visualization to contain elements, got:\n%s", setVis)
	}
}
