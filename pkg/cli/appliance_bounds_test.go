package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// Tests for the per-check bounds of an Appliance (WithTimeout, WithMaxOutput). The slow
// sources are real: an OS command that sleeps (killed when the deadline passes) and a pure
// stream that would take about a minute (stopped at its next node). Timeouts are short
// (50-150ms) against work that cannot finish, so the tests wait for the bound, never for luck.

const (
	slowCommand = "spawn StrVal: seed |> sleep 20"
	longStream  = "spawn StrVal: s |> mock_stream 400000000 |> fold sum"
	// the seed and four 10-byte results: 2 + 4*11 = 46 bytes in the text report
	linesPipeline = "spawn StrVal: s |> lines"
	endless       = "spawn StrVal: s |> mock_stream 400000000"
)

func boundedAppliance(t *testing.T, opts ...ApplianceOption) *Appliance {
	t.Helper()
	app, err := NewAppliance(append([]ApplianceOption{
		WithEmbeddedScript(toyRules), WithPolicy(Policy{AllowExec: true}),
	}, opts...)...)
	if err != nil {
		t.Fatal(err)
	}
	return app
}

func TestApplianceTimeoutStopsASlowCommand(t *testing.T) {
	app := boundedAppliance(t, WithTimeout(100*time.Millisecond))
	start := time.Now()
	got := app.Execute(slowCommand)
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("the check took %v: the child process was not killed", elapsed)
	}
	if len(got) != 1 {
		t.Fatalf("a timed-out check must report exactly one error, got %v", got)
	}
	var te *TimeoutError
	err, _ := got[0].(error)
	if !errors.As(err, &te) || te.Limit != 100*time.Millisecond {
		t.Fatalf("got %T %v, want *TimeoutError{100ms}", got[0], got[0])
	}
	if err.Error() != "check timed out after 100ms" {
		t.Errorf("message = %q", err.Error())
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Error("a TimeoutError must match context.DeadlineExceeded")
	}
	if f := FailureOf(err); f.Kind != FailTimeout || f.Limit != "100ms" {
		t.Errorf("FailureOf = %+v", f)
	}
}

func TestApplianceTimeoutStopsAPureStreamAndDropsPartialResults(t *testing.T) {
	app := boundedAppliance(t, WithTimeout(150*time.Millisecond))
	// the endless stream emits nodes until the deadline; none of them may reach the caller
	got := app.Execute(endless)
	if len(got) != 1 {
		t.Fatalf("partial results leaked: %d values, first %v", len(got), got[0])
	}
	if _, ok := got[0].(*TimeoutError); !ok {
		t.Fatalf("got %T %v, want *TimeoutError", got[0], got[0])
	}
	got = app.Execute(longStream)
	if _, ok := got[0].(*TimeoutError); !ok || len(got) != 1 {
		t.Fatalf("got %v, want one *TimeoutError", got)
	}
}

func TestApplianceTimeoutIsFreshForEveryCheck(t *testing.T) {
	// each check takes about half the timeout; a deadline shared between checks would pass
	// during the third or fourth one
	app := boundedAppliance(t, WithTimeout(300*time.Millisecond))
	for k := 0; k < 6; k++ {
		got := app.Execute("spawn StrVal: seed |> sleep 0.15")
		for _, v := range got {
			if err, ok := v.(error); ok {
				t.Fatalf("check %d failed: %v", k+1, err)
			}
		}
	}
	// ... and the check after a timed-out one is unaffected
	if got := app.Execute(slowCommand); len(got) != 1 {
		t.Fatalf("got %v", got)
	}
	if got := app.Execute("entry 1"); len(got) != 1 || !strings.Contains(fmt.Sprint(got[0]), "2 11") {
		t.Fatalf("a check after a timeout returned %v", got)
	}
	if app.Interpreter().exec.context().Err() != nil {
		t.Fatal("the interpreter was left under a finished context")
	}
}

func TestApplianceWithoutBoundsIsUnchanged(t *testing.T) {
	app := boundedAppliance(t)
	if app.Timeout() != 0 || app.MaxOutput() != 0 {
		t.Fatalf("defaults: timeout %v, max output %d", app.Timeout(), app.MaxOutput())
	}
	got := app.Execute("entry 1")
	if len(got) != 1 || fmt.Sprint(got[0]) != "[2 11]" {
		t.Fatalf("got %v", got)
	}
	if app.Interpreter().exec.outputBudget() != nil {
		t.Fatal("an unbounded appliance installed a budget")
	}
}

func TestApplianceMaxOutputIsALoudError(t *testing.T) {
	const script = "def lines = emit aaaaaaaaaa, bbbbbbbbbb, cccccccccc, dddddddddd\n"
	rows := []struct {
		name  string
		limit int64
		over  bool
	}{
		{"well under", 1000, false},
		{"exactly at the limit (the seed and 4 results of 10 bytes, each plus a newline)", 46, false},
		{"one byte over", 45, true},
		{"far over", 5, true},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			app, err := NewAppliance(WithEmbeddedScript(script), WithMaxOutput(r.limit))
			if err != nil {
				t.Fatal(err)
			}
			got := app.Execute(linesPipeline)
			if !r.over {
				if len(got) != 5 {
					t.Fatalf("got %v, want the 5 results", got)
				}
				return
			}
			if len(got) != 1 {
				t.Fatalf("an over-limit check must report one error and none of its results, got %v", got)
			}
			var oe *OutputLimitError
			err, _ = got[0].(error)
			if !errors.As(err, &oe) || oe.Limit != r.limit {
				t.Fatalf("got %T %v, want *OutputLimitError{%d}", got[0], got[0], r.limit)
			}
			if !errors.Is(err, ErrOutputLimit) {
				t.Error("an OutputLimitError must match ErrOutputLimit")
			}
			if f := FailureOf(err); f.Kind != FailOutputLimit || f.Limit == "" {
				t.Errorf("FailureOf = %+v", f)
			}
		})
	}
}

func TestApplianceMaxOutputStopsAnEndlessStream(t *testing.T) {
	// no timeout: only the budget can end this check
	app := boundedAppliance(t, WithMaxOutput(200))
	done := make(chan []any, 1)
	go func() { done <- app.Execute(endless) }()
	select {
	case got := <-done:
		if _, ok := got[0].(*OutputLimitError); !ok || len(got) != 1 {
			t.Fatalf("got %v, want one *OutputLimitError", got)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("the endless stream was not stopped by the output limit")
	}
}

func TestApplianceBoundsAreFreshForEveryCheck(t *testing.T) {
	// 46 bytes per check; a budget that accumulated across checks would run out
	app, err := NewAppliance(WithEmbeddedScript("def lines = emit aaaaaaaaaa, bbbbbbbbbb, cccccccccc, dddddddddd\n"), WithMaxOutput(46))
	if err != nil {
		t.Fatal(err)
	}
	for k := 0; k < 5; k++ {
		if got := app.Execute(linesPipeline); len(got) != 5 {
			t.Fatalf("check %d: %v", k+1, got)
		}
	}
}

func TestExecuteContextHonoursTheCallersContext(t *testing.T) {
	app := boundedAppliance(t)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	got := app.ExecuteContext(ctx, slowCommand)
	var ce *CancelError
	if len(got) != 1 {
		t.Fatalf("got %v", got)
	}
	if err, _ := got[0].(error); !errors.As(err, &ce) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %T %v, want a *CancelError caused by the deadline", got[0], got[0])
	}
	if f := FailureOf(got[0].(error)); f.Kind != FailStopped {
		t.Errorf("kind = %q, want %q", f.Kind, FailStopped)
	}
}

func TestNegativeBoundsAreRejected(t *testing.T) {
	if _, err := NewAppliance(WithEmbeddedScript(toyRules), WithTimeout(-time.Second)); err == nil || !strings.Contains(err.Error(), "negative timeout") {
		t.Errorf("negative timeout: err = %v", err)
	}
	if _, err := NewAppliance(WithEmbeddedScript(toyRules), WithMaxOutput(-1)); err == nil || !strings.Contains(err.Error(), "negative output limit") {
		t.Errorf("negative output limit: err = %v", err)
	}
}

func TestBoundErrorsReportAsDiagnostics(t *testing.T) {
	var out strings.Builder
	rep := classify([]any{&TimeoutError{Limit: 30 * time.Second}})
	rep.writeText(&out)
	if out.String() != "[ERROR] check timed out after 30s\n" {
		t.Errorf("text = %q", out.String())
	}
	out.Reset()
	if n := rep.writeJSON(&out); n != 1 {
		t.Errorf("errors reported = %d", n)
	}
	want := `{"status":"error","count":0,"violations":[],"errors":["check timed out after 30s"],"diagnostics":[{"kind":"timeout","message":"check timed out after 30s","limit":"30s"}]}` + "\n"
	if out.String() != want {
		t.Errorf("json = %q\nwant   %q", out.String(), want)
	}
	out.Reset()
	classify([]any{&OutputLimitError{Limit: 1024}}).writeJSON(&out)
	want = `{"status":"error","count":0,"violations":[],"errors":["check output exceeded the limit of 1024 bytes"],"diagnostics":[{"kind":"output_limit","message":"check output exceeded the limit of 1024 bytes","limit":"1024"}]}` + "\n"
	if out.String() != want {
		t.Errorf("json = %q\nwant   %q", out.String(), want)
	}
}
