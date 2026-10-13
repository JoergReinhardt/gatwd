package adapters

import (
	"iter"
	"time"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// ClockAdapter is the adapter behind the `now` step: it reads a clock and emits the
// reading as a functions.Time.
//
// # Why an adapter
//
// "What time is it?" is the one question a pure function cannot answer: asking twice gives
// two answers, which is a side effect. Gatwd keeps every effect in an adapter at the edge of
// the pipeline (as scan_file does for a file), so the Time, Duration and Date arithmetic
// stays pure and a rule such as `expires - now < 14d` is a pure function of the reading.
//
// # Injection
//
// The clock is a parameter, `func() time.Time`, so a test or an Edition passes a fixed or
// scripted clock and nothing sleeps or reads the wall clock; nil means time.Now. The
// reading is shown in UTC (local time is never implicit) and its monotonic part is dropped
// (functions.NewTime). For every node that reaches the adapter it emits one Time node with
// the incoming node's type flag, or an error node when the reading is outside the
// supported years 1 to 9999.
//
// Example (a fixed clock, in a test):
//
//	fixed := func() time.Time { return time.Date(2026, 10, 13, 9, 30, 0, 0, time.UTC) }
//	step := evaluator.StepAdapter(evaluator.EitherT(adapters.ClockAdapter(fixed)))
func ClockAdapter(clock func() time.Time) evaluator.Adapter[any] {
	if clock == nil {
		clock = time.Now
	}
	return func(node dsltypes.ASTNode[any]) iter.Seq[dsltypes.ASTNode[any]] {
		return func(yield func(dsltypes.ASTNode[any]) bool) {
			t, err := functions.NewTime(clock().UTC())
			if err != nil {
				yield(dsltypes.NewASTNode[any](err, node.DSLType))
				return
			}
			yield(dsltypes.NewASTNode[any](t, node.DSLType))
		}
	}
}
