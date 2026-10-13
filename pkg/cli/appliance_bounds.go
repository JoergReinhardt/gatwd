package cli

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

// This file bounds ONE check of an Appliance: how long it may take and how much it may
// produce. An operator that evaluates untrusted or merely slow rules on an interval needs
// both, or one runaway check (a command that never returns, a stream that never ends) stalls
// the loop or fills the log.
//
// The mechanism is the one Interpreter.Run already has (hardening W7): a context that every
// pipeline step observes between nodes and that kills the OS commands a step started, and a
// byte budget. An Appliance applies them per Execute, so each check — and each iteration of
// `run --interval` — gets a FRESH deadline and a fresh budget; one slow check does not eat
// the next one's time. A check that runs over a limit reports ONE error naming the limit and
// nothing else: the results it had produced up to then are discarded, never mixed in
// (Mandate 6: a partial answer that looks complete is worse than an error).

// TimeoutError is the failure of a check that ran longer than its timeout. It is an error
// VALUE in the result, so it prints as `[ERROR] check timed out after 30s`, and
// errors.Is(err, context.DeadlineExceeded) holds. The same budget bounds loading and
// reloading the rules (see WithTimeout); there the error is the cause of a *LoadError and
// says "loading the rules timed out after 30s".
type TimeoutError struct {
	// Limit is the timeout that was exceeded.
	Limit time.Duration
	// What names what ran over; empty means "check".
	What string
}

// Error names what ran over and the limit.
func (e *TimeoutError) Error() string {
	what := e.What
	if what == "" {
		what = "check"
	}
	return what + " timed out after " + e.Limit.String()
}

// Is makes a TimeoutError match context.DeadlineExceeded.
func (e *TimeoutError) Is(target error) bool { return target == context.DeadlineExceeded }

// OutputLimitError is the failure of a check that produced more than its output limit. It
// prints as `[ERROR] check output exceeded the limit of 1024 bytes` and matches ErrOutputLimit.
type OutputLimitError struct {
	// Limit is the number of bytes the check was allowed to produce.
	Limit int64
}

// Error names the limit.
func (e *OutputLimitError) Error() string {
	return fmt.Sprintf("check output exceeded the limit of %d bytes", e.Limit)
}

// Is makes an OutputLimitError match ErrOutputLimit.
func (e *OutputLimitError) Is(target error) bool { return target == ErrOutputLimit }

// outputBudget counts the bytes one check collects and cancels the check, once, when they
// pass the limit. It is read from the goroutine that collects the results, and `||>` steps
// finish before the collector sees their nodes, so one counter suffices; the atomics keep it
// safe regardless.
type outputBudget struct {
	max    int64
	n      atomic.Int64
	over   atomic.Bool
	cancel context.CancelCauseFunc
}

// take accounts for n bytes and reports whether they are within the limit. The first byte
// past it cancels the run with an *OutputLimitError cause.
func (b *outputBudget) take(n int) bool {
	if b.n.Add(int64(n)) <= b.max {
		return true
	}
	if b.over.CompareAndSwap(false, true) {
		b.cancel(&OutputLimitError{Limit: b.max})
	}
	return false
}

// outputBudget returns the budget in force, nil when the output is not bounded.
func (s *execState) outputBudget() *outputBudget {
	if s == nil {
		return nil
	}
	return s.budget.Load()
}

// resultSize is the number of bytes a result takes in the text report: its printed form and
// the newline (see checkReport.writeText).
func resultSize(v any) int {
	if err, ok := v.(error); ok {
		return len("[ERROR] ") + len(err.Error()) + 1
	}
	return len(fmt.Sprint(v)) + 1
}

// WithTimeout limits how long ONE check (one Execute) may take; zero, the default, is no
// limit. A check that runs over reports a single *TimeoutError ("check timed out after 30s"):
// every pipeline step stops at its next node and an OS command a step started is killed. The
// deadline is per Execute, so each iteration of a loop starts a fresh one. A negative timeout
// makes NewAppliance fail.
//
// The same duration also bounds LOADING the rules — NewAppliance, Reload and ReloadContext —
// as one budget for the embedded rules plus the override together. Loading evaluates the
// top-level statements of the rules, so a slow or endless statement (a pipeline over a stream
// that never ends, an OS command under --allow-exec) would otherwise stall the start-up or the
// hot reload for good. A load that runs over fails with a *LoadError whose Err is a
// *TimeoutError ("loading the rules timed out after 30s", errors.Is(err,
// context.DeadlineExceeded)); NewAppliance returns it, and Reload leaves the last good rules in
// place. Like a check, loading is stopped cooperatively: at the next statement, at the next
// node of a pipeline, by killing an OS command — a single pure evaluation of one element that
// never returns cannot be interrupted. The Edition's own preludes are part of the binary and are
// not bounded.
func WithTimeout(d time.Duration) ApplianceOption {
	return func(a *Appliance) { a.timeout = d }
}

// WithMaxOutput limits how many bytes ONE check may produce, counted as the text report
// prints its results (each result's printed form and a newline); zero, the default, is no
// limit. A check that produces more reports a single *OutputLimitError and none of the
// results it had produced; the run is cancelled as soon as the limit is passed, so a stream
// that never ends is stopped too. It bounds what the check emits, not the memory it uses
// while it computes — use WithTimeout for that. A negative limit makes NewAppliance fail.
func WithMaxOutput(bytes int64) ApplianceOption {
	return func(a *Appliance) { a.maxOutput = bytes }
}

// Timeout returns the per-check timeout; zero means none.
func (a *Appliance) Timeout() time.Duration { return a.timeout }

// MaxOutput returns the per-check output limit in bytes; zero means none.
func (a *Appliance) MaxOutput() int64 { return a.maxOutput }

// checkBounds validates the options that bound a check.
func (a *Appliance) checkBounds() error {
	switch {
	case a.timeout < 0:
		return fmt.Errorf("appliance: negative timeout %v", a.timeout)
	case a.maxOutput < 0:
		return fmt.Errorf("appliance: negative output limit %d", a.maxOutput)
	}
	return nil
}

// ExecuteContext is Execute under ctx: the check stops when ctx is done as well as when the
// Appliance's own timeout or output limit is reached. When the check is stopped the result is
// exactly one error value saying why — a *TimeoutError, an *OutputLimitError, or a
// *CancelError when ctx itself was cancelled — and no partial results.
//
// An Appliance with no timeout and no limit given a context that cannot end (Execute) changes
// nothing about the interpreter. Otherwise the bounds are installed on the active
// interpreter for the call and restored afterwards, so checks that use them must not overlap
// on one Appliance.
func (a *Appliance) ExecuteContext(ctx context.Context, line string) []any {
	interp := a.Interpreter()
	if interp == nil {
		return nil
	}
	if a.timeout == 0 && a.maxOutput == 0 && ctx.Done() == nil {
		return interp.EvaluateToValues(line)
	}

	var cancelTimeout context.CancelFunc = func() {}
	if a.timeout > 0 {
		ctx, cancelTimeout = context.WithTimeoutCause(ctx, a.timeout, &TimeoutError{Limit: a.timeout})
	}
	defer cancelTimeout()
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	defer interp.SetContext(interp.SetContext(ctx))
	if a.maxOutput > 0 {
		budget := &outputBudget{max: a.maxOutput, cancel: cancel}
		interp.exec.budget.Store(budget)
		defer interp.exec.budget.Store(nil)
	}

	values := interp.EvaluateToValues(line)
	if ctx.Err() == nil {
		return values
	}
	// stopped: the values are what a half-finished run happened to emit; say why instead
	cause := cancelCause(ctx)
	var timeout *TimeoutError
	var limit *OutputLimitError
	switch {
	case errors.As(cause, &timeout), errors.As(cause, &limit):
		return []any{cause}
	}
	return []any{&CancelError{Cause: cause}}
}
