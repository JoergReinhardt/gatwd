package evaluator

import (
	"math"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

func list(items ...any) *collections.List[any] {
	l := collections.Nil[any]()
	for i := len(items) - 1; i >= 0; i-- {
		l = collections.Cons(items[i], l)
	}
	return l
}

func mustValueSet(t *testing.T, items ...any) *collections.Setoid[any] {
	t.Helper()
	s, err := NewValueSet(items...)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// Element identity is the language's `==`, never the printed form (user decision 2026-10-13,
// plan 6.3). Each row is a pair that is one member or two.
func TestValueSetMembershipIsEquality(t *testing.T) {
	utc, _ := functions.ParseTime("2026-10-13T09:30:00Z")
	berlin, _ := functions.ParseTime("2026-10-13T11:30:00+02:00")
	for _, c := range []struct {
		name string
		a, b any
		same bool
	}{
		{"Int and its String are two", 1, "1", false},
		{"Int and the equal Float are one", 1, 1.0, true},
		{"narrow and wide Int are one", int8(3), int64(3), true},
		{"equal lists are one (two pointers, one value)", list(1, 2), list(1, 2), true},
		{"the same elements in another order are two", list(1, 2), list(2, 1), false},
		{"a list and its printed form are two", list(1, 2), "[1 2]", false},
		{"a list of Ints and a list of Strings are two", list(1), list("1"), false},
		{"a Pair and the equal list are one", functions.Pair[any, any]{Fst: 1, Snd: 2}, list(1, 2), true},
		{"equal Maybes are one", functions.Just[any](1), functions.Just[any](1), true},
		{"a Maybe and its payload are two", functions.Just[any](1), 1, false},
		{"one instant in two zones is one", utc, berlin, true},
		{"a Set of Sets: element order does not matter", mustValueSet(t, 1, 2), mustValueSet(t, 2, 1), true},
		{"true and 1 are two", true, 1, false},
	} {
		s := mustValueSet(t, c.a, c.b)
		if got := s.Len() == 1; got != c.same {
			t.Errorf("%s: Len = %d, want one member = %v", c.name, s.Len(), c.same)
		}
	}
}

func TestValueSetStoresTheValueNotText(t *testing.T) {
	berlin, _ := functions.ParseTime("2026-10-13T11:30:00+02:00")
	s := mustValueSet(t, list(1, 2), 1.5, berlin)
	var sawList, sawFloat, sawUTC bool
	for _, e := range s.Elements() {
		switch v := e.(type) {
		case *collections.List[any]:
			sawList = true
		case float64:
			sawFloat = v == 1.5
		case functions.Time:
			sawUTC = v.String() == "2026-10-13T09:30:00Z"
		default:
			t.Errorf("member %#v was turned into %T", e, e)
		}
	}
	if !sawList || !sawFloat || !sawUTC {
		t.Errorf("members lost their type: list=%v float=%v utc-time=%v", sawList, sawFloat, sawUTC)
	}
}

func TestValueSetRefusesWhatHasNoEquality(t *testing.T) {
	fn := functions.NewClosure([]string{"x"}, func(a []any) (any, error) { return a[0], nil })
	if _, err := NewValueSet(1, fn); err == nil || !strings.Contains(err.Error(), "a Function cannot be a Set member") {
		t.Errorf("a function as a member: err = %v", err)
	}
	if _, err := NewValueSet(math.NaN()); err == nil {
		t.Error("NaN is equal to nothing, so it is no member")
	}
	if _, err := AsValueSet(42); err == nil || !strings.Contains(err.Error(), "want a Set, got Int") {
		t.Errorf("AsValueSet(42): err = %v", err)
	}
	// the HAMT's wording is unchanged
	if _, err := EncodeKey(fn); err == nil || !strings.Contains(err.Error(), "cannot be a HAMT key") {
		t.Errorf("EncodeKey(function): err = %v", err)
	}
}

// A Go caller's own carriers are read as DSL Sets, under the language's equality.
func TestAsValueSetReadsForeignCarriers(t *testing.T) {
	for name, carrier := range map[string]any{
		"Set[string]": collections.NewSet("a", "b"),
		"Set[any]":    collections.NewSet[any]("a", "b"),
	} {
		s, err := AsValueSet(carrier)
		if err != nil || s.Len() != 2 {
			t.Errorf("%s: %v, %v", name, s, err)
			continue
		}
		if ok, _ := s.Contains("a"); !ok {
			t.Errorf("%s: lost a member", name)
		}
	}
}

// runPipeline runs steps over an empty stream and returns every terminal node, errors included.
func runPipeline(steps []PipelineStep[functions.Either[error, any]]) []dsltypes.ASTNode[functions.Either[error, any]] {
	engine := NewEngine[functions.Either[error, any]]()
	initial := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {}
	var out []dsltypes.ASTNode[functions.Either[error, any]]
	for n := range engine.EvaluatePipeline(initial, steps) {
		out = append(out, n)
	}
	return out
}

func stepValues(steps ...PipelineStep[functions.Either[error, any]]) []any {
	return runPipelineToValues(steps)
}

func emitted(vals ...any) PipelineStep[functions.Either[error, any]] { return StepEmitValues(vals...) }

func TestStepDistinctIsValueTyped(t *testing.T) {
	got := stepValues(emitted(1, 1.0, "1", list(1, 2), list(1, 2), list(2, 1), int8(1)), StepDistinct())
	if len(got) != 4 { // 1 (== 1.0 == int8(1)), "1", [1 2], [2 1]
		t.Fatalf("distinct kept %d elements %v, want 4", len(got), got)
	}
	if got[0] != 1 || got[1] != "1" {
		t.Errorf("the first of each class stays, in stream order: %#v", got[:2])
	}
	// no node survives for a value with no equality: an error node, the rest still flows
	fn := functions.NewClosure([]string{"x"}, func(a []any) (any, error) { return a[0], nil })
	nodes := runPipeline([]PipelineStep[functions.Either[error, any]]{emitted(1, fn, 1, 2), StepDistinct()})
	var errs, vals int
	for _, n := range nodes {
		if n.Value.IsLeft() {
			errs++
			e, _ := n.Value.UnwrapLeft()
			if !strings.Contains(e.Error(), "distinct: a Function cannot be a Set member") {
				t.Errorf("error = %v", e)
			}
		} else {
			vals++
		}
	}
	if errs != 1 || vals != 2 {
		t.Errorf("errors = %d, values = %d, want 1 and 2", errs, vals)
	}
}

func TestStepToSetYieldsOneSetoid(t *testing.T) {
	got := stepValues(emitted(1, 1.0, "1", list(1, 2), list(1, 2)), StepToSet())
	if len(got) != 1 {
		t.Fatalf("to_set yielded %d nodes", len(got))
	}
	s, ok := got[0].(*collections.Setoid[any])
	if !ok || s.Len() != 3 {
		t.Fatalf("to_set = %#v, want a *Setoid[any] of 3 members", got[0])
	}
	// an unkeyable element: errors, and NO set (a Set silently missing a member is worse)
	fn := functions.NewClosure([]string{"x"}, func(a []any) (any, error) { return a[0], nil })
	for _, n := range runPipeline([]PipelineStep[functions.Either[error, any]]{emitted(1, fn), StepToSet()}) {
		if !n.Value.IsLeft() {
			t.Errorf("a Set was yielded despite an unkeyable element: %#v", n.Value)
		}
	}
}

func TestSetStepsOverTheDSLSet(t *testing.T) {
	base := mustValueSet(t, 1, list(1, 2))
	one := func(step PipelineStep[functions.Either[error, any]]) any {
		t.Helper()
		out := stepValues(emitted(any(base)), step)
		if len(out) != 1 {
			t.Fatalf("want one value, got %v", out)
		}
		return out[0]
	}
	for _, c := range []struct {
		name string
		step PipelineStep[functions.Either[error, any]]
		want any
	}{
		{"contains 1.0 (== 1)", StepSetContains(1.0), true},
		{"contains \"1\"", StepSetContains("1"), false},
		{"contains an equal list", StepSetContains(list(1, 2)), true},
		{"contains the list's printed form", StepSetContains("[1 2]"), false},
	} {
		if got := one(c.step); got != c.want {
			t.Errorf("%s = %v, want %v", c.name, got, c.want)
		}
	}
	added := one(StepSetAdd(list(1, 2))).(*collections.Setoid[any])
	if added.Len() != 2 {
		t.Errorf("adding an equal list changed the size to %d", added.Len())
	}
	added = one(StepSetAdd("1")).(*collections.Setoid[any])
	if added.Len() != 3 {
		t.Errorf("adding the String \"1\" gave %d members, want 3", added.Len())
	}
	removed := one(StepSetRemove(1.0)).(*collections.Setoid[any])
	if removed.Len() != 1 {
		t.Errorf("removing 1.0 left %d members, want 1", removed.Len())
	}
	// set_elements gives the members back as the values they are
	elems := stepValues(emitted(any(base)), StepSetElements())
	if len(elems) != 2 {
		t.Fatalf("set_elements = %v", elems)
	}
	var sawInt, sawList bool
	for _, e := range elems {
		switch e.(type) {
		case int, int8, int64:
			sawInt = true
		case *collections.List[any]:
			sawList = true
		}
	}
	if !sawInt || !sawList {
		t.Errorf("set_elements lost types: %#v", elems)
	}
	// the algebra, across carriers: another DSL Set, and a plain List operand
	other := mustValueSet(t, 1.0, "x")
	for name, c := range map[string]struct {
		step PipelineStep[functions.Either[error, any]]
		want int
	}{
		"union":                 {StepSetUnion(any(other)), 3},
		"intersect":             {StepSetIntersect(any(other)), 1},
		"diff":                  {StepSetDiff(any(other)), 1},
		"intersect with a List": {StepSetIntersect(any(list(list(1, 2), 7))), 1},
		"diff with a List":      {StepSetDiff(any(list(1, 7))), 1},
	} {
		got := one(c.step).(*collections.Setoid[any])
		if got.Len() != c.want {
			t.Errorf("%s: %d members, want %d", name, got.Len(), c.want)
		}
	}
}

// A step that needs a Set and gets something else says so, and names the step.
func TestSetStepsRefuseNonSets(t *testing.T) {
	for name, step := range map[string]PipelineStep[functions.Either[error, any]]{
		"set_contains": StepSetContains(1),
		"set_add":      StepSetAdd(1),
		"set_remove":   StepSetRemove(1),
		"set_elements": StepSetElements(),
		"set_diff":     StepSetDiff(any(mustValueSet(t, 1))),
	} {
		nodes := runPipeline([]PipelineStep[functions.Either[error, any]]{emitted(42), step})
		if len(nodes) != 1 || !nodes[0].Value.IsLeft() {
			t.Errorf("%s over an Int: %v, want one error node", name, nodes)
			continue
		}
		e, _ := nodes[0].Value.UnwrapLeft()
		if !strings.HasPrefix(e.Error(), name+": ") {
			t.Errorf("%s: error %q does not name the step", name, e)
		}
	}
}
