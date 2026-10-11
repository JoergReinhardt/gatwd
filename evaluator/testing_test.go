package evaluator_test

import (
	"fmt"
	"iter"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

func makeIntStream(count int) func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
	return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		for i := 0; i < count; i++ {
			node := dsltypes.NewASTNode(
				functions.Right[error, any](any(i)),
				dsltypes.BitFlag(1),
			)
			if !yield(node) {
				return
			}
		}
	}
}

func TestVerifyExampleOutput(t *testing.T) {
	exampleOutput := `
		10
		20
		30
	`
	actualOutput := []string{"10", "20", "30"}
	evaluator.VerifyExampleOutput(t, exampleOutput, actualOutput)
}

func TestStepTelemetry(t *testing.T) {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	var count uint64 = 0

	telemetryStep := evaluator.StepTelemetry(&count)
	stream := engine.EvaluatePipeline(
		makeIntStream(25),
		[]evaluator.PipelineStep[functions.Either[error, any]]{telemetryStep},
	)

	consumed := 0
	for range stream {
		consumed++
	}

	if consumed != 25 {
		t.Fatalf("expected 25 items consumed, got %d", consumed)
	}
	if count != 25 {
		t.Fatalf("expected telemetry counter to equal 25, got %d", count)
	}
}

func TestExampleRunner_Ordered(t *testing.T) {
	ex := evaluator.Example{
		Name: "OrderedExample",
		F: func() {
			fmt.Println("Alpha")
			fmt.Println("Beta")
			fmt.Println("Gamma")
		},
		Output: `
			Alpha
			Beta
			Gamma
		`,
		Unordered: false,
	}

	pass, captured, diff := evaluator.RunExample(ex)
	if !pass {
		t.Fatalf("expected example to pass, diff: %s", diff)
	}
	if !strings.Contains(captured, "Alpha") {
		t.Fatalf("expected captured stdout to contain Alpha")
	}
}

func TestExampleRunner_Unordered(t *testing.T) {
	ex := evaluator.Example{
		Name: "UnorderedExample",
		F: func() {
			// Emitted in non-alphabetical order
			fmt.Println("Zeta")
			fmt.Println("Alpha")
			fmt.Println("Mu")
		},
		Output: `
			Alpha
			Mu
			Zeta
		`,
		Unordered: true,
	}

	pass, _, diff := evaluator.RunExample(ex)
	if !pass {
		t.Fatalf("expected unordered example to pass, diff: %s", diff)
	}
}

func TestExampleRunner_Mismatch(t *testing.T) {
	ex := evaluator.Example{
		Name: "MismatchExample",
		F: func() {
			fmt.Println("DifferentOutput")
		},
		Output: `
			ExpectedOutput
		`,
		Unordered: false,
	}

	pass, _, diff := evaluator.RunExample(ex)
	if pass {
		t.Fatalf("expected example to fail on output mismatch")
	}
	if !strings.Contains(diff, "EXAMPLE ASSERTION FAILED") {
		t.Fatalf("expected formatted visual diff card, got: %s", diff)
	}
}

func TestRunExamples_Suite(t *testing.T) {
	suite := []evaluator.Example{
		{
			Name: "SuiteEx1",
			F: func() {
				fmt.Println("100")
			},
			Output: "100",
		},
		{
			Name: "SuiteEx2_Unordered",
			F: func() {
				fmt.Println("B")
				fmt.Println("A")
			},
			Output:    "A\nB",
			Unordered: true,
		},
	}

	evaluator.RunExamples(t, suite)
}

func TestCaptureStdout_PanicSafety(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic to propagate")
		}
	}()

	_, _ = evaluator.CaptureStdout(func() {
		fmt.Println("about to panic")
		panic("deliberate panic for safety test")
	})
}

func TestRunInternalExamples_StandardLib(t *testing.T) {
	examples := []evaluator.Example{
		{
			Name: "Internal1",
			F: func() {
				fmt.Println("hello world")
			},
			Output: "hello world\n",
		},
	}

	matchAll := func(pat, str string) (bool, error) {
		return true, nil
	}

	ok := evaluator.RunInternalExamples(matchAll, examples)
	if !ok {
		t.Fatalf("expected testing.RunExamples to succeed")
	}
}

func BenchmarkPipelineExecution(b *testing.B) {
	sourceFactory := func() iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return makeIntStream(100)
	}

	steps := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepMathAdd("5"),
		evaluator.StepTake[functions.Either[error, any]](50),
	}

	report := evaluator.BenchmarkPipeline(b, "IntStreamAddAndTake", sourceFactory, steps)
	if report.ItemsProcessed == 0 {
		b.Fatalf("expected processed items in benchmark report")
	}
	fmt.Println(report.String())
}
