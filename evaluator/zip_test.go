package evaluator

import (
	"errors"
	"iter"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// secondaryOf builds a one-to-many adapter from a function of the incoming value: it
// yields one Right node per returned value, or the Left of a failure.
func secondaryOf(f func(in any) []either) Adapter[either] {
	return func(node dsltypes.ASTNode[either]) iter.Seq[dsltypes.ASTNode[either]] {
		return func(yield func(dsltypes.ASTNode[either]) bool) {
			in, _ := node.Value.UnwrapRight()
			for _, v := range f(in) {
				if !yield(dsltypes.NewASTNode(v, node.DSLType)) {
					return
				}
			}
		}
	}
}

func pairOf(t *testing.T, e either) functions.Pair[any, any] {
	t.Helper()
	v, isRight := e.UnwrapRight()
	if !isRight {
		t.Fatalf("want a Right Pair, got %v", e)
	}
	p, ok := v.(functions.Pair[any, any])
	if !ok {
		t.Fatalf("want a functions.Pair[any, any], got %T (%v)", v, v)
	}
	return p
}

// TestStepZipPairsEachNodeWithItsPartner: zip yields the Pair (primary, secondary) - the
// primary is no longer dropped in favour of the secondary alone (the bug: `echo A: %s |>
// zip date` produced only the date).
func TestStepZipPairsEachNodeWithItsPartner(t *testing.T) {
	ok := func(v any) either { return functions.Right[error, any](v) }
	upper := secondaryOf(func(in any) []either { return []either{ok(strings.ToUpper(in.(string)))} })

	out := drain(StepZip(upper)(streamOf(ok("a"), ok("b"))))
	if len(out) != 2 {
		t.Fatalf("want 2 pairs, got %v", out)
	}
	for idx, want := range [][2]string{{"a", "A"}, {"b", "B"}} {
		p := pairOf(t, out[idx])
		if p.Fst != want[0] || p.Snd != want[1] {
			t.Errorf("pair %d = (%v, %v), want (%s, %s)", idx, p.Fst, p.Snd, want[0], want[1])
		}
	}

	// One-to-many secondary: one pair per secondary result, primary repeated.
	two := secondaryOf(func(in any) []either { return []either{ok(1), ok(2)} })
	many := drain(StepZip(two)(streamOf(ok("x"))))
	if len(many) != 2 || pairOf(t, many[0]).Fst != "x" || pairOf(t, many[1]).Snd != 2 {
		t.Errorf("one-to-many zip = %v", many)
	}
}

// TestStepZipFailsLoud: an error from the secondary, an error already in the stream and a
// primary with no partner each reach the Left channel; none becomes a half-empty Pair.
func TestStepZipFailsLoud(t *testing.T) {
	ok := func(v any) either { return functions.Right[error, any](v) }
	boom := errors.New("boom: secondary failed")

	failing := secondaryOf(func(in any) []either { return []either{functions.Left[error, any](boom)} })
	out := drain(StepZip(failing)(streamOf(ok("a"))))
	if len(out) != 1 || !out[0].IsLeft() {
		t.Fatalf("a failing secondary must yield its error, got %v", out)
	}

	good := secondaryOf(func(in any) []either { return []either{ok("p")} })
	out = drain(StepZip(good)(streamOf(functions.Left[error, any](errors.New("upstream")), ok("a"))))
	if len(out) != 2 || !out[0].IsLeft() || out[1].IsLeft() {
		t.Fatalf("an upstream error must pass through ahead of the next pair, got %v", out)
	}

	none := secondaryOf(func(in any) []either { return nil })
	out = drain(StepZip(none)(streamOf(ok("lonely"))))
	if len(out) != 1 || !out[0].IsLeft() {
		t.Fatalf("a primary with no partner must be an error, not silence: %v", out)
	}
	if e, _ := out[0].UnwrapLeft(); !strings.Contains(e.Error(), "lonely") {
		t.Errorf("the error should name the unpaired value: %v", e)
	}
}

// TestStepFanOutDoesNotPairAnError (Mandate 6): a branch that failed used to be unwrapped
// with its error discarded, giving the Pair (<nil>, value): `(cat /missing &&& date)`.
func TestStepFanOutDoesNotPairAnError(t *testing.T) {
	ok := func(v any) either { return functions.Right[error, any](v) }
	boom := errors.New("boom: branch failed")
	fine := secondaryOf(func(in any) []either { return []either{ok("fine")} })
	bad := secondaryOf(func(in any) []either { return []either{functions.Left[error, any](boom)} })

	for name, step := range map[string]PipelineStep[either]{
		"left branch fails":  StepFanOut(bad, fine),
		"right branch fails": StepFanOut(fine, bad),
	} {
		out := drain(step(streamOf(ok("in"))))
		if len(out) != 1 || !out[0].IsLeft() {
			t.Errorf("%s: want the branch error alone, got %v", name, out)
		}
	}

	out := drain(StepFanOut(fine, fine)(streamOf(ok("in"))))
	if len(out) != 1 || pairOf(t, out[0]).Snd != "fine" {
		t.Errorf("control: both branches fine should pair, got %v", out)
	}
}
