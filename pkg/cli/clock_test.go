package cli

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/joergreinhardt/gatwd/functions"
)

func fixedClock(text string) func() time.Time {
	t, err := time.Parse(time.RFC3339, text)
	if err != nil {
		panic(err)
	}
	return func() time.Time { return t }
}

// `now` is a stream-source step that reads the injected Interpreter.Clock.
func TestNowStepReadsTheInjectedClock(t *testing.T) {
	in := NewInterpreter()
	in.Clock = fixedClock("2026-10-13T11:30:00+02:00")
	got, ok := evalOne(t, in, `spawn S: x |> now`).(functions.Time)
	if !ok || got.String() != "2026-10-13T09:30:00Z" {
		t.Fatalf("spawn S: x |> now = %#v", got)
	}
	// the bare head is the same reading
	head, ok := evalOne(t, in, `now`).(functions.Time)
	if !ok || head.Compare(got) != 0 {
		t.Errorf("bare now = %#v", head)
	}
	// moving the clock moves the reading (nothing is cached)
	in.Clock = fixedClock("2027-01-01T00:00:00Z")
	if later := evalOne(t, in, `now`).(functions.Time); later.String() != "2027-01-01T00:00:00Z" {
		t.Errorf("after moving the clock: %s", later)
	}
}

// The motivating rule shapes from the plan, with a fixed clock: no wall clock, no sleep.
func TestMotivatingRulesWithAnInjectedClock(t *testing.T) {
	in := NewInterpreter()
	in.Clock = fixedClock("2026-10-13T09:30:00Z")
	in.EvaluateScript(`def stale = \now created -> (now - created) > 30d
def expiring = \now expires -> (expires - now) < 14d`)
	for _, tc := range []struct {
		src  string
		want any
	}{
		{`now |> map (\n -> stale n @2026-08-01T00:00:00Z)`, true},
		{`now |> map (\n -> stale n @2026-10-01T00:00:00Z)`, false},
		{`now |> map (\n -> expiring n @2026-10-20T00:00:00Z)`, true},
		{`now |> map (\n -> expiring n @2026-12-01T00:00:00Z)`, false},
		{`now |> map (\n -> expiring n @2026-10-01T00:00:00Z)`, true}, // already expired: negative remaining
		{`spawn S: x |> now |> map (\n -> ((@2026-10-20T09:30:00Z) - n) < 14d)`, true},
	} {
		vals := in.EvaluateToValues(tc.src)
		items, _ := functions.ToSliceAny(vals[0])
		if len(items) != 1 || items[0] != tc.want {
			if len(items) == 0 && vals[0] == tc.want {
				continue
			}
			t.Errorf("%s = %v (%v), want %v", tc.src, vals, items, tc.want)
		}
	}
}

// A user binding named now shadows the clock; the clock is for builtins only.
func TestNowIsShadowedByABinding(t *testing.T) {
	in := NewInterpreter()
	in.Clock = fixedClock("2026-10-13T09:30:00Z")
	in.EvaluateLine(`let now = 5`)
	if got := evalOne(t, in, `now`); !functions.StructuralEqual(got, 5) {
		t.Errorf("a let-bound now = %#v", got)
	}
	if bare := NewBareInterpreter(); len(bare.evaluatePipelineStr(`now`)) == 1 && !strings.Contains(bare.evaluatePipelineStr(`now`)[0], "ERROR") && strings.Contains(bare.evaluatePipelineStr(`now`)[0], "2026") {
		t.Error("the bare runtime has no builtins, so no clock")
	}
}

// The clock reaches forks (parallel stages) and pipeline-bodied lambdas.
func TestClockPropagatesToForksAndLambdas(t *testing.T) {
	in := NewInterpreter()
	in.Clock = fixedClock("2031-10-13T09:30:00Z") // a year the wall clock is not in
	vals := in.EvaluateToValues(`[1, 2, 3] |> map (\x -> x |> now |> time_year)`)
	if len(vals) != 1 {
		t.Fatalf("got %v", vals)
	}
	items, _ := functions.ToSliceAny(vals[0])
	if len(items) != 3 {
		t.Fatalf("items = %v", items)
	}
	for _, it := range items {
		if !functions.StructuralEqual(it, 2031) {
			t.Errorf("a lambda saw %v, want the injected 2031", it)
		}
	}
	// goroutines sharing one interpreter read the same fixed clock safely
	var wg sync.WaitGroup
	for k := 0; k < 8; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f := in.fork()
			if got := f.EvaluateToValues(`now`); len(got) != 1 || got[0].(functions.Time).String() != "2031-10-13T09:30:00Z" {
				t.Errorf("fork read %v", got)
			}
		}()
	}
	wg.Wait()
}

// Without an injected clock the step reads the real one: near time.Now, in UTC.
func TestNowDefaultsToTheWallClock(t *testing.T) {
	in := NewInterpreter()
	before := time.Now().Add(-time.Second)
	got := evalOne(t, in, `now`).(functions.Time)
	if got.Std().Before(before) || got.Std().After(time.Now().Add(time.Second)) {
		t.Errorf("now = %s", got)
	}
	if got.Std().Location() != time.UTC {
		t.Errorf("now is shown in %s, never Local", got.Std().Location())
	}
}

// An out-of-range clock is an error, not a wrong value.
func TestNowOutOfRangeClockFailsLoud(t *testing.T) {
	in := NewInterpreter()
	in.Clock = func() time.Time { return time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) }
	for _, src := range []string{`now`, `spawn S: x |> now`} {
		res := in.evaluatePipelineStr(src)
		if len(res) != 1 || !strings.HasPrefix(res[0], "[ERROR]") || !strings.Contains(res[0], "out of range") {
			t.Errorf("%s = %v, want an out-of-range [ERROR]", src, res)
		}
	}
}
