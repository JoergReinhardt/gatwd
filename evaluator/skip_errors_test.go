package evaluator

import (
	"errors"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

// TestSkipErrorsDropsLeftsKeepsRights pins StepSkipErrors against the rule
// TestCollectingStepsDeliverUpstreamErrors pins the other way: an aggregate
// withholds its result over a failed stream (Mandate 6), and StepSkipErrors is
// the explicit opt-in that removes the Lefts first so the aggregate downstream
// sees a clean stream and succeeds. Negative control: without StepSkipErrors,
// the identical stream still fails StepCollectList (asserted first, so this
// test could not pass by accident).
func TestSkipErrorsDropsLeftsKeepsRights(t *testing.T) {
	ok := func(v any) either { return functions.Right[error, any](v) }
	boom := errors.New("boom: upstream failure")
	failed := []either{ok(1), functions.Left[error, any](boom), ok(2)}

	// Negative control: the raw failed stream still fails collect_list.
	raw := drain(StepCollectList()(streamOf(failed...)))
	if len(raw) != 1 || !raw[0].IsLeft() {
		t.Fatalf("control: collect_list over a failed stream should withhold its aggregate, got %v", raw)
	}

	// skip_errors alone: the Left is reported-and-dropped, both Rights survive.
	out := drain(StepSkipErrors()(streamOf(failed...)))
	if len(out) != 2 {
		t.Fatalf("skip_errors: want 2 surviving elements, got %v", out)
	}
	for _, e := range out {
		if e.IsLeft() {
			t.Fatalf("skip_errors: a Left reached downstream: %v", out)
		}
	}

	// Composed with collect_list: the aggregate now succeeds over just the Rights.
	composed := drain(StepCollectList()(StepSkipErrors()(streamOf(failed...))))
	if len(composed) != 1 || composed[0].IsLeft() {
		t.Fatalf("skip_errors |> collect_list: want one successful aggregate, got %v", composed)
	}
	list, ok2 := functions.ToSliceAny(mustRight(t, composed[0]))
	if !ok2 || len(list) != 2 {
		t.Fatalf("skip_errors |> collect_list: want [1 2], got %v", composed[0])
	}
}

func mustRight(t *testing.T, e either) any {
	t.Helper()
	v, ok := e.UnwrapRight()
	if !ok {
		t.Fatalf("expected a Right, got %v", e)
	}
	return v
}
