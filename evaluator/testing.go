package evaluator

import (
	"fmt"
	"iter"
	"strings"
	"testing"
	"time"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// BenchmarkReport captures performance telemetry from a pipeline benchmark.
type BenchmarkReport struct {
	Name           string
	Operations     int
	TotalDuration  time.Duration
	OpsPerSec      float64
	NanosecsPerOp  int64
	ItemsProcessed uint64
}

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
	expectedLines := strings.Split(strings.TrimSpace(expectedOutput), "\n")
	for i := range expectedLines {
		expectedLines[i] = strings.TrimSpace(expectedLines[i])
	}

	var trimmedActual []string
	for _, l := range actualOutput {
		tl := strings.TrimSpace(l)
		if tl != "" {
			trimmedActual = append(trimmedActual, tl)
		}
	}

	joinedExpected := strings.Join(expectedLines, "\n")
	joinedActual := strings.Join(trimmedActual, "\n")

	if joinedExpected != joinedActual {
		AssertVisualDiff(t, joinedExpected, joinedActual)
	}
}
