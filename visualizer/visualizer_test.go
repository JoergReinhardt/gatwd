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

func TestVisualizer_ASTNodeBitFlagType(t *testing.T) {
	type RawHostData struct {
		Payload string
		Count   int
	}

	customFlag := dsltypes.BitFlag(1 << 8)
	dsltypes.RegisterTypeFlag(customFlag, "DomainTelemetryRecord")

	node := dsltypes.NewASTNode(RawHostData{
		Payload: "sensor_matrix_active",
		Count:   100,
	}, customFlag)

	opts := visualizer.ParseOptions("--curved --border")
	output := visualizer.Format(node, opts)

	// It MUST use the AST node's BitFlag type ("DomainTelemetryRecord")
	// and NOT the Go reflection struct type ("RawHostData")!
	if !strings.Contains(output, "DomainTelemetryRecord") {
		t.Fatalf("expected visualization to use AST node BitFlag type 'DomainTelemetryRecord', got:\n%s", output)
	}
	if strings.Contains(output, "RawHostData") {
		t.Fatalf("visualization should not use Go reflection type 'RawHostData', got:\n%s", output)
	}

	// Test FormatAST directly
	astOut := visualizer.FormatAST(customFlag, RawHostData{Payload: "direct", Count: 1}, opts)
	if !strings.Contains(astOut, "DomainTelemetryRecord") {
		t.Fatalf("expected FormatAST to use BitFlag type 'DomainTelemetryRecord', got:\n%s", astOut)
	}

	// Test flag color application via BitFlag
	optsColor := visualizer.DefaultOptions()
	optsColor.FlagColors[customFlag] = "cyan"
	colorOut := visualizer.Format(node, optsColor)
	if !strings.Contains(colorOut, "\033[36m") {
		t.Fatalf("expected FlagColors to colorize output via BitFlag, got:\n%s", colorOut)
	}
}

func TestVisualizer_BorderColorAndAlternateBg(t *testing.T) {
	// 1. Border color
	optsBorderColor := visualizer.ParseOptions("--border-color=yellow")
	if optsBorderColor.BorderColor != "yellow" || !optsBorderColor.Border {
		t.Fatalf("expected BorderColor=yellow and Border=true, got %+v", optsBorderColor)
	}
	resBorder := visualizer.Format("hello", optsBorderColor)
	// Yellow ANSI code is 33
	if !strings.Contains(resBorder, "\033[33m") {
		t.Fatalf("expected yellow border ANSI escape sequence, got:\n%s", resBorder)
	}

	// 2. Alternating background colors on lists
	list := []string{"item1", "item2", "item3", "item4"}
	optsAlt := visualizer.ParseOptions("--alternate-bg=blue,magenta")
	if len(optsAlt.AlternateBg) != 2 || optsAlt.AlternateBg[0] != "blue" || optsAlt.AlternateBg[1] != "magenta" {
		t.Fatalf("expected AlternateBg [blue magenta], got %+v", optsAlt.AlternateBg)
	}
	resAlt := visualizer.Format(list, optsAlt)
	// Blue bg code is 44, magenta bg code is 45
	if !strings.Contains(resAlt, "\033[44m") {
		t.Fatalf("expected blue background (44m) in list output, got:\n%s", resAlt)
	}
	if !strings.Contains(resAlt, "\033[45m") {
		t.Fatalf("expected magenta background (45m) in list output, got:\n%s", resAlt)
	}
}

