package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// Tests for the bound on LOADING rules (NewAppliance, Reload): the top-level statements of the
// embedded rules and of an override are evaluated while loading, so a slow or endless one must
// not stall the start-up or the hot reload. The slow statements are real: a pure pipeline over a
// long stream (stopped at its next node) and, under the exec policy, an OS command that sleeps
// (killed). They are sized to take seconds if nothing stops them, against timeouts of a
// fraction of a second, so a guard that does not work makes the test slow AND fail.

const (
	// about four seconds of pure work, if it is never stopped
	slowLoadStatement  = "spawn StrVal: s |> mock_stream 40000000 |> fold sum\n"
	sleepLoadStatement = "spawn StrVal: s |> sleep 3\n"
	fastLoadStatement  = "spawn StrVal: s |> mock_stream 1000 |> fold sum\n"
)

func loadWith(t *testing.T, override string, opts ...ApplianceOption) (*Appliance, error) {
	t.Helper()
	path := writeFile(t, "rules.gatwd", override)
	return NewAppliance(append([]ApplianceOption{
		WithEmbeddedScript(toyRules), WithRequiredDefs("entry"), WithConfigMapOverride(path),
	}, opts...)...)
}

func requireLoadTimeout(t *testing.T, err error, limit time.Duration) *LoadError {
	t.Helper()
	var le *LoadError
	if !errors.As(err, &le) {
		t.Fatalf("got %T %v, want a *LoadError", err, err)
	}
	var te *TimeoutError
	if !errors.As(err, &te) || te.Limit != limit {
		t.Fatalf("the LoadError must wrap a *TimeoutError{%v}, got %v", limit, le.Err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Error("a load timeout must match context.DeadlineExceeded")
	}
	want := "loading the rules timed out after " + limit.String()
	if len(le.Problems) != 1 || le.Problems[0] != want {
		t.Errorf("problems = %q, want exactly [%q]", le.Problems, want)
	}
	if f := FailureOf(err); f.Kind != FailTimeout {
		t.Errorf("FailureOf kind = %q, want %q", f.Kind, FailTimeout)
	}
	return le
}

func TestApplianceLoadTimeoutStopsAPureStatement(t *testing.T) {
	start := time.Now()
	app, err := loadWith(t, slowLoadStatement+"def extra = \\x -> x\n", WithTimeout(150*time.Millisecond))
	if elapsed := time.Since(start); elapsed > 1500*time.Millisecond {
		t.Fatalf("loading took %v: the statement was not stopped", elapsed)
	}
	if app != nil {
		t.Error("a failed load must not return an Appliance")
	}
	le := requireLoadTimeout(t, err, 150*time.Millisecond)
	if !strings.HasSuffix(le.Source, "rules.gatwd") {
		t.Errorf("the source must name the override, got %q", le.Source)
	}
}

func TestApplianceLoadTimeoutKillsACommand(t *testing.T) {
	start := time.Now()
	_, err := loadWith(t, sleepLoadStatement, WithTimeout(150*time.Millisecond), WithPolicy(Policy{AllowExec: true}))
	if elapsed := time.Since(start); elapsed > 1500*time.Millisecond {
		t.Fatalf("loading took %v: the command was not killed", elapsed)
	}
	requireLoadTimeout(t, err, 150*time.Millisecond)
}

// Without the exec policy the same override fails the way it always did, not by timing out.
func TestApplianceLoadTimeoutDoesNotHideAPolicyDenial(t *testing.T) {
	_, err := loadWith(t, sleepLoadStatement, WithTimeout(time.Minute))
	var le *LoadError
	if !errors.As(err, &le) || le.Err != nil || !strings.Contains(err.Error(), "exec") {
		t.Fatalf("got %v, want the capability denial as a plain LoadError", err)
	}
}

func TestApplianceLoadTimeoutBoundsTheEmbeddedRulesToo(t *testing.T) {
	start := time.Now()
	_, err := NewAppliance(WithEmbeddedScript(toyRules+slowLoadStatement), WithRequiredDefs("entry"), WithTimeout(150*time.Millisecond))
	if time.Since(start) > 1500*time.Millisecond {
		t.Fatal("the embedded rules were not stopped")
	}
	le := requireLoadTimeout(t, err, 150*time.Millisecond)
	if le.Source != "embedded rules" {
		t.Errorf("source = %q, want embedded rules", le.Source)
	}
}

// The budget is per load, and a load inside it is untouched.
func TestApplianceLoadWithinTheTimeoutAndWithoutOne(t *testing.T) {
	for name, opts := range map[string][]ApplianceOption{
		"a generous timeout": {WithTimeout(time.Minute)},
		"no timeout":         nil,
	} {
		t.Run(name, func(t *testing.T) {
			app, err := loadWith(t, fastLoadStatement+"def rule_b = \\x -> x + 100\n", opts...)
			if err != nil {
				t.Fatal(err)
			}
			if got := app.Execute("entry 1"); len(got) != 1 || sprint(got[0]) != "[2 101]" {
				t.Errorf("got %v", got)
			}
		})
	}
}

// The context a load runs under must not outlive it: a later Execute (which has no bound of its
// own here) would find it done and stop at once.
func TestApplianceLoadContextDoesNotOutliveTheLoad(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	app, err := loadWith(t, fastLoadStatement, WithLoadContext(ctx)) // no timeout: Execute takes the unbounded path
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	for k := 0; k < 2; k++ {
		got := app.Execute("spawn StrVal: s |> mock_stream 5 |> fold sum")
		if len(got) != 1 {
			t.Fatalf("check %d: %v", k, got)
		}
		// a pipeline (its steps observe the context); 0+1+2+3+4
		if _, isErr := got[0].(error); isErr || sprint(got[0]) != "10" {
			t.Fatalf("check %d gave %v: the load context leaked into the check", k, got[0])
		}
	}
}

func TestApplianceLoadStoppedByItsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var once sync.Once
	// a step of the override cancels the context while the statement is running
	trigger := WithStep(SinkStep("trigger", func(any) { once.Do(cancel) }))
	start := time.Now()
	_, err := loadWith(t, "spawn StrVal: s |> mock_stream 40000000 |> trigger |> fold sum\n", WithLoadContext(ctx), trigger)
	if time.Since(start) > 1500*time.Millisecond {
		t.Fatal("the load was not stopped")
	}
	var le *LoadError
	var ce *CancelError
	if !errors.As(err, &le) || !errors.As(err, &ce) || !errors.Is(err, context.Canceled) {
		t.Fatalf("got %T %v, want a LoadError wrapping a CancelError", err, err)
	}
	if len(le.Problems) != 1 || !strings.Contains(le.Problems[0], "context canceled") {
		t.Errorf("problems = %q", le.Problems)
	}
}

// A hot reload that runs over keeps serving the last good rules, reports the failure, and the
// next reload — with a fixed override — recovers. The statement that hangs is the override's,
// so every reload gets its own budget.
func TestApplianceReloadTimeoutKeepsTheLastGoodRules(t *testing.T) {
	path := writeFile(t, "rules.gatwd", "def rule_b = \\x -> x + 100\n")
	app, err := NewAppliance(WithEmbeddedScript(toyRules), WithRequiredDefs("entry"),
		WithConfigMapOverride(path), WithTimeout(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	before := app.Report()
	const check = "entry 1"
	if got := app.Execute(check); len(got) != 1 || !strings.Contains(sprint(got[0]), "101") {
		t.Fatalf("first rules: %v", got)
	}

	// the same Appliance, now with a tight budget, reloading a hanging override
	app.timeout = 150 * time.Millisecond
	if err := os.WriteFile(path, []byte(slowLoadStatement+"def rule_b = \\x -> x + 999\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	err = app.Reload()
	if time.Since(start) > 1500*time.Millisecond {
		t.Fatalf("the reload took %v", time.Since(start))
	}
	requireLoadTimeout(t, err, 150*time.Millisecond)
	if got := app.Execute(check); len(got) != 1 || !strings.Contains(sprint(got[0]), "101") {
		t.Errorf("after the failed reload the last good rules must serve, got %v", got)
	}
	if app.Report().String() != before.String() {
		t.Errorf("the report changed on a failed reload: %q", app.Report())
	}

	if err := os.WriteFile(path, []byte("def rule_b = \\x -> x + 999\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := app.Reload(); err != nil {
		t.Fatalf("a fixed override must reload: %v", err)
	}
	if got := app.Execute(check); len(got) != 1 || !strings.Contains(sprint(got[0]), "1000") {
		t.Errorf("after the recovery: %v", got)
	}
}

func TestApplianceReloadContextCancelledKeepsTheLastGoodRules(t *testing.T) {
	path := writeFile(t, "rules.gatwd", "def rule_b = \\x -> x + 100\n")
	app, err := NewAppliance(WithEmbeddedScript(toyRules), WithRequiredDefs("entry"), WithConfigMapOverride(path))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = app.ReloadContext(ctx)
	var ce *CancelError
	if !errors.As(err, &ce) {
		t.Fatalf("got %v, want a LoadError wrapping a CancelError", err)
	}
	if got := app.Execute("entry 1"); len(got) != 1 || !strings.Contains(sprint(got[0]), "101") {
		t.Errorf("the last good rules must keep serving: %v", got)
	}
}

func sprint(v any) string { return fmt.Sprint(v) }

// --- the command line ---

func TestApplianceTreeLoadTimeout(t *testing.T) {
	slow := writeFile(t, "slow.gatwd", slowLoadStatement)
	rows := []struct {
		name string
		argv []string
	}{
		{"check", []string{"check", "--source", "entry 1", "--rules", slow, "--timeout", "150ms"}},
		{"run", []string{"run", "--source", "entry 1", "--rules", slow, "--timeout", "150ms", "--interval", "1h"}},
		{"json", []string{"check", "--source", "entry 1", "--rules", slow, "--timeout", "150ms", "--format", "json"}},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			start := time.Now()
			code, stdout, stderr := toyRun(t, toyTree(toyRules), r.argv...)
			if time.Since(start) > 1500*time.Millisecond {
				t.Fatal("the start-up load was not bounded")
			}
			if code != 1 || stdout != "" {
				t.Fatalf("exit %d, stdout %q, want 1 and nothing (the rules never loaded)\nstderr: %s", code, stdout, stderr)
			}
			for _, want := range []string{"toy: ", slow, "loading the rules timed out after 150ms"} {
				if !strings.Contains(stderr, want) {
					t.Errorf("stderr lacks %q:\n%s", want, stderr)
				}
			}
		})
	}
}

// Negative control: the same override, no timeout, loads (it only takes a few seconds, so the
// control uses the fast statement).
func TestApplianceTreeLoadWithoutTimeoutIsUnbounded(t *testing.T) {
	fast := writeFile(t, "fast.gatwd", fastLoadStatement)
	code, stdout, stderr := toyRun(t, toyTree(toyRules), "check", "--source", "entry 1", "--rules", fast)
	if code != 0 || stdout != "[2 11]\n" {
		t.Fatalf("exit %d, stdout %q\nstderr: %s", code, stdout, stderr)
	}
}

// In the interval loop the hang is in the reload: the loop reports it, keeps the previous
// rules, checks again, and recovers when the override is fixed.
func TestApplianceTreeRunReloadTimeoutKeepsTheLoopAlive(t *testing.T) {
	override := writeFile(t, "rules.gatwd", "def rule_b = \\x -> x + 100\n")
	edits := []string{
		slowLoadStatement + "def rule_b = \\x -> x + 999\n", // after check 1: hangs
		"def rule_b = \\x -> x + 999\n",                     // after check 2: fixed
	}
	tree := toyTree(toyRules)
	waits := 0
	tree.sleep = func(time.Duration) bool {
		if waits < len(edits) {
			if err := os.WriteFile(override, []byte(edits[waits]), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		waits++
		return waits <= len(edits)
	}
	start := time.Now()
	code, stdout, stderr := toyRun(t, tree, "run", "--source", "entry 1", "--rules", override, "--interval", "1s", "--timeout", "150ms")
	if time.Since(start) > 10*time.Second {
		t.Fatalf("the loop took %v", time.Since(start))
	}
	if code != 0 {
		t.Fatalf("exit %d\nstderr: %s", code, stderr)
	}
	// check 1: old rules; check 2: reload failed, STILL the old rules; check 3: the fix
	if want := "[2 101]\n[2 101]\n[2 1000]\n"; stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	if !strings.Contains(stderr, "reload failed, still using the previous rules") || !strings.Contains(stderr, "loading the rules timed out after 150ms") {
		t.Errorf("stderr must report the reload timeout:\n%s", stderr)
	}
}

// A stop request during the start-up load (the first load, before any check) ends the command
// like any other stop request, and the load is cancelled rather than left to finish.
func TestApplianceStopRequestDuringTheStartupLoad(t *testing.T) {
	override := writeFile(t, "rules.gatwd", "spawn StrVal: s |> mock_stream 40000000 |> trigger |> fold sum\n")
	for _, c := range []struct {
		cmd  string
		code int
		want string
	}{
		{"run", 0, "toy: received terminated, stopping\n"},
		{"check", 1, "toy: the check was interrupted (received terminated)\n"},
	} {
		t.Run(c.cmd, func(t *testing.T) {
			f := newSignalFixture(t)
			var once sync.Once
			f.onNode = func() { once.Do(func() { f.sigs <- syscall.SIGTERM }) }
			argv := []string{c.cmd, "--source", "entry 1", "--rules", override}
			if c.cmd == "run" {
				argv = append(argv, "--interval", "1h")
			}
			start := time.Now()
			code, stdout, stderr := f.run(t, nil, nil, argv...)
			if time.Since(start) > 1500*time.Millisecond {
				t.Fatal("the load was not cancelled")
			}
			if code != c.code || stdout != "" || !strings.HasSuffix(stderr, c.want) {
				t.Errorf("exit %d (want %d), stdout %q, stderr %q (want suffix %q)", code, c.code, stdout, stderr, c.want)
			}
		})
	}
}

// And a stop request while a reload runs.
func TestApplianceStopRequestDuringAReload(t *testing.T) {
	override := writeFile(t, "rules.gatwd", "def rule_b = \\x -> x + 100\n")
	f := newSignalFixture(t)
	var once sync.Once
	f.onNode = func() { once.Do(func() { f.sigs <- syscall.SIGINT }) }
	f.tree.sleep = func(time.Duration) bool {
		// between the checks: the override now hangs, on a step that raises the request
		_ = os.WriteFile(override, []byte("spawn StrVal: s |> mock_stream 40000000 |> trigger |> fold sum\n"), 0o644)
		return true
	}
	start := time.Now()
	code, stdout, stderr := f.run(t, nil, nil, "run", "--source", "entry 1", "--rules", override, "--interval", "1h")
	if time.Since(start) > 1500*time.Millisecond {
		t.Fatal("the reload was not cancelled")
	}
	if code != 0 || stdout != "[2 101]\n" || !strings.HasSuffix(stderr, "toy: received interrupt, stopping\n") || strings.Contains(stderr, "reload failed") {
		t.Errorf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}
