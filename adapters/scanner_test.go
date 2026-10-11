package adapters

import (
	"bufio"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// testFlag is a DSL quasi-type used across all scanner tests.
const testFlag dsltypes.BitFlag = 1 << 4

func TestScannerToSeq_BasicLines(t *testing.T) {
	input := "hello\nworld\nfoo"
	scanner := bufio.NewScanner(strings.NewReader(input))

	var got []string
	for node := range ScannerToSeq(scanner, testFlag) {
		if node.Value.IsLeft() {
			err, _ := node.Value.UnwrapLeft()
			t.Fatalf("unexpected error node: %v", err)
		}
		val, _ := node.Value.UnwrapRight()
		got = append(got, string(val.([]byte)))
	}

	expected := []string{"hello", "world", "foo"}
	if len(got) != len(expected) {
		t.Fatalf("expected %d nodes, got %d", len(expected), len(got))
	}
	for i, v := range got {
		if v != expected[i] {
			t.Errorf("node[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestScannerToSeq_EmptyInput(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader(""))

	count := 0
	for range ScannerToSeq(scanner, testFlag) {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 nodes for empty input, got %d", count)
	}
}

func TestScannerToSeq_DSLTypePreserved(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("a\nb"))
	customFlag := dsltypes.BitFlag(1 << 7)

	for node := range ScannerToSeq(scanner, customFlag) {
		if !node.Match(customFlag) {
			t.Errorf("node DSLType = %v, want %v", node.DSLType, customFlag)
		}
	}
}

func TestScannerToSeq_EarlyTermination(t *testing.T) {
	input := "alpha\nbeta\ngamma\ndelta"
	scanner := bufio.NewScanner(strings.NewReader(input))

	var got []string
	seq := ScannerToSeq(scanner, testFlag)
	// Use Take to consume only the first 2 elements.
	for node := range functions.Take(seq, 2) {
		val, _ := node.Value.UnwrapRight()
		got = append(got, string(val.([]byte)))
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 nodes after Take(2), got %d", len(got))
	}
	if got[0] != "alpha" || got[1] != "beta" {
		t.Errorf("expected [alpha beta], got %v", got)
	}
}

func TestScannerAdapter_PropagatesUpstreamError(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("should not scan"))
	adapter := ScannerAdapter(scanner)

	// Feed an error (Left) node into the adapter.
	errNode := evaluator.InjectError[any](
		errTestSentinel,
		testFlag,
	)

	var gotLeft bool
	for node := range adapter(errNode) {
		if node.Value.IsLeft() {
			gotLeft = true
			err, _ := node.Value.UnwrapLeft()
			if err != errTestSentinel {
				t.Errorf("propagated error = %v, want %v", err, errTestSentinel)
			}
		}
	}
	if !gotLeft {
		t.Error("expected upstream error to be propagated, but got Right nodes")
	}
}

var errTestSentinel = errorString("test sentinel error")

// errorString is a minimal error type to avoid importing "errors" or "fmt" just for a sentinel.
type errorString string

func (e errorString) Error() string { return string(e) }

func TestScannerAdapter_YieldsTokens(t *testing.T) {
	input := "line1\nline2\nline3"
	scanner := bufio.NewScanner(strings.NewReader(input))
	adapter := ScannerAdapter(scanner)

	// Feed a dummy Right node to trigger the scanner.
	trigger := dsltypes.NewASTNode(
		functions.Right[error, any](any("trigger")),
		testFlag,
	)

	var got []string
	for node := range adapter(trigger) {
		if node.Value.IsLeft() {
			err, _ := node.Value.UnwrapLeft()
			t.Fatalf("unexpected error: %v", err)
		}
		val, _ := node.Value.UnwrapRight()
		got = append(got, string(val.([]byte)))
	}

	expected := []string{"line1", "line2", "line3"}
	if len(got) != len(expected) {
		t.Fatalf("expected %d nodes, got %d", len(expected), len(got))
	}
	for i, v := range got {
		if v != expected[i] {
			t.Errorf("node[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestScannerAdapter_InheritsDSLType(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("x"))
	adapter := ScannerAdapter(scanner)
	triggerFlag := dsltypes.BitFlag(1 << 9)

	trigger := dsltypes.NewASTNode(
		functions.Right[error, any](any("go")),
		triggerFlag,
	)

	for node := range adapter(trigger) {
		if !node.Match(triggerFlag) {
			t.Errorf("emitted node DSLType = %v, want %v", node.DSLType, triggerFlag)
		}
	}
}

func TestScannerAdapter_PipelineIntegration(t *testing.T) {
	input := "aaa\nbbb\nccc"
	scanner := bufio.NewScanner(strings.NewReader(input))

	engine := evaluator.NewEngine[functions.Either[error, any]]()
	adapter := ScannerAdapter(scanner)

	// Build a pipeline: scanner → take(2)
	steps := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepAdapter(adapter),
		evaluator.StepTake[functions.Either[error, any]](2),
	}

	trigger := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		node := dsltypes.NewASTNode(
			functions.Right[error, any](any("start")),
			testFlag,
		)
		yield(node)
	}

	result := engine.EvaluatePipeline(trigger, steps)

	var got []string
	for node := range result {
		if node.Value.IsLeft() {
			err, _ := node.Value.UnwrapLeft()
			t.Fatalf("unexpected error in pipeline: %v", err)
		}
		val, _ := node.Value.UnwrapRight()
		got = append(got, string(val.([]byte)))
	}

	if len(got) != 2 {
		t.Fatalf("pipeline expected 2 nodes, got %d", len(got))
	}
	if got[0] != "aaa" || got[1] != "bbb" {
		t.Errorf("pipeline got %v, want [aaa bbb]", got)
	}
}
