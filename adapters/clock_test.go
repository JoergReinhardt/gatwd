package adapters_test

import (
	"testing"
	"time"

	"github.com/joergreinhardt/gatwd/adapters"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

func readClock(clock func() time.Time) []dsltypes.ASTNode[any] {
	var out []dsltypes.ASTNode[any]
	for n := range adapters.ClockAdapter(clock)(dsltypes.NewASTNode[any]("seed", dsltypes.BitFlag(4))) {
		out = append(out, n)
	}
	return out
}

// A fixed clock yields exactly one functions.Time in UTC, with the incoming node's flag.
func TestClockAdapterFixed(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Skip("no tzdata:", err)
	}
	fixed := func() time.Time { return time.Date(2026, 10, 13, 11, 30, 0, 123, berlin) }
	out := readClock(fixed)
	if len(out) != 1 {
		t.Fatalf("want one node, got %d", len(out))
	}
	got, ok := out[0].Value.(functions.Time)
	if !ok {
		t.Fatalf("value is %T, want functions.Time", out[0].Value)
	}
	if got.String() != "2026-10-13T09:30:00.000000123Z" {
		t.Errorf("clock reading = %s, want the instant shown in UTC", got)
	}
	if out[0].DSLType != dsltypes.BitFlag(4) {
		t.Errorf("the node flag %v is not the incoming one", out[0].DSLType)
	}
}

// A scripted clock is read once per node, so two reads see two readings.
func TestClockAdapterIsReadEachTime(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	n := 0
	clock := func() time.Time { n++; return base.Add(time.Duration(n) * time.Hour) }
	a := readClock(clock)[0].Value.(functions.Time)
	b := readClock(clock)[0].Value.(functions.Time)
	if a.String() != "2026-01-01T01:00:00Z" || b.String() != "2026-01-01T02:00:00Z" {
		t.Errorf("readings %s, %s", a, b)
	}
}

// nil means the wall clock, and the monotonic reading never leaks.
func TestClockAdapterDefaultsToWallClock(t *testing.T) {
	before := time.Now().Add(-time.Second)
	out := readClock(nil)
	got := out[0].Value.(functions.Time)
	if got.Std().Before(before) || got.Std().After(time.Now().Add(time.Second)) {
		t.Errorf("wall clock reading %s is not now", got)
	}
	if got.Std() != got.Std().Round(0) {
		t.Error("the monotonic reading must be stripped")
	}
	if got.Std().Location() != time.UTC {
		t.Errorf("the reading is shown in %s, never Local", got.Std().Location())
	}
}

// A reading outside the supported years is an error node, never a wrong Time.
func TestClockAdapterOutOfRangeIsAnError(t *testing.T) {
	out := readClock(func() time.Time { return time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) })
	if len(out) != 1 {
		t.Fatalf("want one node, got %d", len(out))
	}
	if err, ok := out[0].Value.(error); !ok || err == nil {
		t.Errorf("want an error value, got %#v", out[0].Value)
	}
}
