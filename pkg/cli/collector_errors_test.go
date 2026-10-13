package cli

import (
	"strings"
	"testing"
)

// TestCollectingStepsDeliverUpstreamErrorsInDSL (Mandate 6): a stream whose middle node
// fails (`6 // 0`) used to come out of build_bst as a tree of the surviving values, out of
// `fold sum` as 9, and out of frequencies / to_hamt / to_set as a collection that silently
// lacked the failed element. Every collecting step must now deliver the error and nothing
// that looks like a result. (A DSL `assert` cannot match an [ERROR], so this lives in Go.)
func TestCollectingStepsDeliverUpstreamErrorsInDSL(t *testing.T) {
	const failing = `spawn S: x |> emit 1, 0, 2 |> drop 1 |> map (\x -> 6 // x)`
	for _, step := range []string{
		"collect_list", "to_list", "to_chunk", "to_set", "to_hamt", "frequencies",
		"build_bst", "build_tree", "build_btree", "build_rbtree", "build_array",
		"fold count", "fold sum", "fold max",
	} {
		t.Run(step, func(t *testing.T) {
			in := NewInterpreter()
			got := in.evaluatePipelineStr(failing + " |> " + step)
			if len(got) != 1 || !strings.Contains(got[0], "[ERROR]") || !strings.Contains(got[0], "division by zero") {
				t.Fatalf("%s over a failed stream = %q; want exactly the [ERROR] with its cause", step, got)
			}
		})
	}

	// Control: the same step over the clean stream is a value, not an error, so the test
	// above can only pass by the error being delivered.
	for _, step := range []string{"build_bst", "fold sum", "to_set"} {
		in := NewInterpreter()
		got := in.evaluatePipelineStr(`spawn S: x |> emit 1, 3, 2 |> drop 1 |> map (\x -> 6 // x) |> ` + step)
		if len(got) != 1 || strings.Contains(got[0], "[ERROR]") {
			t.Errorf("%s over a clean stream = %q; want one value", step, got)
		}
	}
}

// TestPivotFiltersFailLoud: filter_less / filter_greater with an undefined pivot variable
// used to vanish from the pipeline (the step was silently not added), so the whole stream
// came through unfiltered.
func TestPivotFiltersFailLoud(t *testing.T) {
	for _, step := range []string{"filter_less", "filter_greater"} {
		in := NewInterpreter()
		got := in.evaluatePipelineStr(`spawn S: x |> emit 1, 5, 9 |> drop 1 |> ` + step + ` nosuchpivot`)
		if len(got) == 0 || !strings.Contains(strings.Join(got, "\n"), "nosuchpivot") {
			t.Errorf("%s with an undefined pivot = %q; want an [ERROR] naming it", step, got)
		}
	}
}
