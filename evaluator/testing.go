package evaluator

import (
	"bytes"
	"fmt"
	"io"
	"iter"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

var stdoutMu sync.Mutex

// Example models Go's testing.InternalExample contract for testable, reproducible examples.
//
// In Go's testing framework, an example function executes and emits output to standard output.
// The output is verified against an expected output block, supporting both strict sequential
// matching (Unordered = false, corresponding to "// Output:") and permutation-invariant line
// matching (Unordered = true, corresponding to "// Unordered output:").
type Example struct {
	Name      string
	F         func()
	Output    string
	Unordered bool
}

// ToInternalExample converts an Example into Go's standard library testing.InternalExample.
func (e Example) ToInternalExample() testing.InternalExample {
	return testing.InternalExample{
		Name:      e.Name,
		F:         e.F,
		Output:    e.Output,
		Unordered: e.Unordered,
	}
}

// FromInternalExample converts a standard library testing.InternalExample into an Example.
func FromInternalExample(ie testing.InternalExample) Example {
	return Example{
		Name:      ie.Name,
		F:         ie.F,
		Output:    ie.Output,
		Unordered: ie.Unordered,
	}
}

// CaptureStdout intercepts and returns standard output produced during execution of fn.
// It synchronizes concurrent calls with a mutex and guarantees that os.Stdout is restored
// even if fn panics.
func CaptureStdout(fn func()) (out string, err error) {
	stdoutMu.Lock()
	defer stdoutMu.Unlock()

	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		return "", pipeErr
	}

	origStdout := os.Stdout
	os.Stdout = w

	outCh := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		_ = r.Close()
		outCh <- buf.String()
	}()

	defer func() {
		_ = w.Close()
		os.Stdout = origStdout
		out = <-outCh
	}()

	fn()
	return
}

// CompareExampleOutput compares expected and actual outputs according to Go's example contract.
// If unordered is true, the relative order of non-empty trimmed lines does not matter.
// If outputs differ, it returns false along with a formatted visual diff card.
func CompareExampleOutput(expected, actual string, unordered bool) (bool, string) {
	expLines := splitAndTrimLines(expected)
	actLines := splitAndTrimLines(actual)

	if unordered {
		sort.Strings(expLines)
		sort.Strings(actLines)
	}

	expJoined := strings.Join(expLines, "\n")
	actJoined := strings.Join(actLines, "\n")

	if expJoined == actJoined {
		return true, ""
	}

	modeStr := "Ordered"
	if unordered {
		modeStr = "Unordered"
	}

	diffCard := fmt.Sprintf("\n╭── EXAMPLE ASSERTION FAILED (%s) ──────────────────╮\n│ Expected:\n%s\n│ Actual:\n%s\n╰─────────────────────────────────────────────────────────────╯",
		modeStr, indentLines(expJoined, "│   "), indentLines(actJoined, "│   "))

	return false, diffCard
}

func splitAndTrimLines(s string) []string {
	var lines []string
	for _, l := range strings.Split(strings.TrimSpace(s), "\n") {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

func indentLines(s, prefix string) string {
	if s == "" {
		return prefix + "<empty>"
	}
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}

// RunExample executes a single Example, intercepting stdout and verifying output against ex.Output.
// It returns whether the example passed, the captured output, and any diff error message.
func RunExample(ex Example) (bool, string, string) {
	captured, err := CaptureStdout(ex.F)
	if err != nil {
		return false, "", fmt.Sprintf("failed to intercept stdout: %v", err)
	}
	passed, diff := CompareExampleOutput(ex.Output, captured, ex.Unordered)
	return passed, captured, diff
}

// RunExamples runs a slice of Example tests against a testing.TB harness.
// If t is a *testing.T (supporting subtests), each example executes in its own subtest.
// If any example fails, a formatted failure diff is logged or failed.
func RunExamples(t testing.TB, examples []Example) {
	t.Helper()
	type subtestRunner interface {
		Run(name string, f func(t *testing.T)) bool
	}

	if runner, ok := t.(subtestRunner); ok {
		for _, ex := range examples {
			runner.Run(ex.Name, func(st *testing.T) {
				st.Helper()
				passed, captured, diff := RunExample(ex)
				if !passed {
					st.Fatalf("Example %s failed!\n%s\nCaptured output:\n%s", ex.Name, diff, captured)
				}
			})
		}
		return
	}

	for _, ex := range examples {
		passed, captured, diff := RunExample(ex)
		if !passed {
			t.Fatalf("Example %s failed!\n%s\nCaptured output:\n%s", ex.Name, diff, captured)
		}
	}
}

// RunInternalExamples executes examples via Go's standard library testing.RunExamples function.
func RunInternalExamples(matchString func(pat, str string) (bool, error), examples []Example) bool {
	var stdExamples []testing.InternalExample
	for _, ex := range examples {
		stdExamples = append(stdExamples, ex.ToInternalExample())
	}
	return testing.RunExamples(matchString, stdExamples)
}

// BenchmarkReport captures performance telemetry from a pipeline benchmark.
type BenchmarkReport struct {
	Name           string
	Operations     int
	TotalDuration  time.Duration
	OpsPerSec      float64
	NanosecsPerOp  int64
	ItemsProcessed uint64
}

// String renders the BenchmarkReport as a single human-readable summary line
// (name, operation count, total duration, throughput, ns/op, and items
// processed), satisfying the fmt.Stringer interface.
func (r BenchmarkReport) String() string {
	return fmt.Sprintf("Benchmark %s: %d ops in %v (%.2f ops/sec, %d ns/op, %d items)",
		r.Name, r.Operations, r.TotalDuration, r.OpsPerSec, r.NanosecsPerOp, r.ItemsProcessed)
}

// BenchmarkPipeline runs a pipeline under Go's testing.B benchmark harness,
// providing throughput, allocation profiling, and latency metrics.
func BenchmarkPipeline(
	b *testing.B,
	name string,
	sourceFactory func() iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]],
	steps []PipelineStep[functions.Either[error, any]],
) BenchmarkReport {
	b.Helper()
	b.ReportAllocs()

	engine := NewEngine[functions.Either[error, any]]()

	// Warm-up run
	warmStream := engine.EvaluatePipeline(sourceFactory(), steps)
	for range warmStream {
	}

	b.ResetTimer()
	start := time.Now()
	var totalItems uint64 = 0

	for i := 0; i < b.N; i++ {
		stream := engine.EvaluatePipeline(sourceFactory(), steps)
		for node := range stream {
			if node.Value.IsRight() {
				totalItems++
			}
		}
	}

	duration := time.Since(start)
	b.StopTimer()

	opsPerSec := float64(b.N) / duration.Seconds()
	nsPerOp := duration.Nanoseconds() / int64(b.N)

	report := BenchmarkReport{
		Name:           name,
		Operations:     b.N,
		TotalDuration:  duration,
		OpsPerSec:      opsPerSec,
		NanosecsPerOp:  nsPerOp,
		ItemsProcessed: totalItems,
	}
	b.Log(report.String())
	return report
}

// StepTelemetry is a non-intrusive transducer that tracks item count and elapsed duration
// of stream processing without modifying stream elements.
func StepTelemetry(counter *uint64) PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if counter != nil {
					*counter++
				}
				if !yield(node) {
					return
				}
			}
		}
	}
}

// AssertVisualDiff compares expected against actual. If they differ, it formats
// a visual comparison box and fails the test via testing.TB.
func AssertVisualDiff(t testing.TB, expected string, actual string) {
	t.Helper()
	if expected == actual {
		return
	}

	diffCard := fmt.Sprintf("\n╭── ASSERTION FAILED ─────────────────────────────────╮\n│ Expected: %-42s│\n│ Actual:   %-42s│\n╰─────────────────────────────────────────────────────╯",
		expected, actual)

	t.Fatalf("%s", diffCard)
}

// VerifyExampleOutput tests that actual lines produced by a pipeline match an
// expected example block (similar to Go's standard testable Example // Output: contract).
func VerifyExampleOutput(t testing.TB, expectedOutput string, actualOutput []string) {
	t.Helper()
	actual := strings.Join(actualOutput, "\n")
	passed, diff := CompareExampleOutput(expectedOutput, actual, false)
	if !passed {
		t.Fatalf("%s", diff)
	}
}
