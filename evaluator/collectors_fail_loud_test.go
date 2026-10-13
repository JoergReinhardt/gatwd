package evaluator

import (
	"errors"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// TestCollectingStepsDeliverUpstreamErrors (Mandate 6): a step that gathers a whole
// stream into one value must never hide an upstream error node behind a plausible
// result. Every aggregating step below, fed `1, <error>, 2`, must deliver the error
// (the Left channel) and yield NO aggregate over the failed stream. The streaming
// steps (sort, distinct, chunk, the pivot filters) keep their per-element output but
// must still pass the error on. A clean stream is the control: it must deliver one
// value and no error, so the test can fail in both directions.
func TestCollectingStepsDeliverUpstreamErrors(t *testing.T) {
	ok := func(v any) either { return functions.Right[error, any](v) }
	boom := errors.New("boom: upstream failure")
	pivot := dsltypes.NewASTNode(ok(5), 0)

	rows := []struct {
		name       string
		step       PipelineStep[either]
		aggregates bool // true: the aggregate must be withheld when the stream failed
		clean      []either
	}{
		{"collect_list", StepCollectList(), true, nil},
		{"to_list", StepToList(), true, nil},
		{"to_chunk", StepToChunk(), true, nil},
		{"to_set", StepToSet(), true, nil},
		{"to_hamt", StepToHAMT(), true, []either{ok("a:1"), ok("b:2")}},
		{"frequencies", StepFrequencies(), true, nil},
		{"build_tree", StepBuildTree("root"), true, nil},
		{"build_bst", StepBuildBST(), true, nil},
		{"build_btree", StepBuildBTree(), true, nil},
		{"build_rbtree", StepBuildRBTree(), true, nil},
		{"build_array", StepBuildArray(), true, nil},
		{"fold count", StepFold("count"), true, nil},
		{"fold sum", StepFold("sum"), true, nil},
		{"fold product", StepFold("product"), true, nil},
		{"fold min", StepFold("min"), true, nil},
		{"fold max", StepFold("max"), true, nil},
		{"fold concat", StepFold("concat"), true, []either{ok("a"), ok("b")}},
		{"sort", StepSort(DefaultLess), false, nil},
		{"distinct", StepDistinct(), false, nil},
		{"chunk", StepChunk(2), false, nil},
		{"unchunk", StepUnchunk(), false, nil},
		{"filter_less", StepFilterLess(pivot), false, nil},
		{"filter_greater_equal", StepFilterGreaterEqual(pivot), false, nil},
	}

	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			good := row.clean
			if good == nil {
				good = []either{ok(1), ok(2)}
			}

			// Control: the clean stream yields no error.
			for _, e := range drain(row.step(streamOf(good...))) {
				if e.IsLeft() {
					l, _ := e.UnwrapLeft()
					t.Fatalf("clean stream produced an error: %v", l)
				}
			}

			failedIn := []either{good[0], functions.Left[error, any](boom)}
			failedIn = append(failedIn, good[1:]...)
			out := drain(row.step(streamOf(failedIn...)))

			lefts, rights := 0, 0
			for _, e := range out {
				if e.IsLeft() {
					l, _ := e.UnwrapLeft()
					if !strings.Contains(l.Error(), "boom") {
						t.Errorf("error lost its message: %v", l)
					}
					lefts++
				} else {
					rights++
				}
			}
			if lefts == 0 {
				t.Fatalf("the upstream error was swallowed; output: %v", out)
			}
			if row.aggregates && rights != 0 {
				t.Errorf("an aggregate was built over a failed stream: %v", out)
			}
		})
	}
}
