package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Tests for what RunOptions bounds: the policy, the context and the output limit (W7).

func TestRunSourceDefaultsToDenyAll(t *testing.T) {
	in, buf := quietInterpreter()
	res, err := in.Run(RunOptions{Source: "def main = echo hello-from-source\n", Args: []string{"a"}})
	if err == nil {
		t.Fatalf("a Source task that needs the exec capability must fail; lines %q", res.Lines)
	}
	if len(res.Lines) == 0 || !strings.Contains(res.Lines[0], "step 'echo' requires the exec capability (policy denies it)") {
		t.Fatalf("lines = %q, want the capability error", res.Lines)
	}
	if strings.Contains(strings.Join(res.Lines, "\n")+buf.String(), "hello-from-source") {
		t.Fatal("the denied command ran")
	}
	if in.Policy() != TrustedPolicy() {
		t.Fatalf("Run must restore the interpreter's policy; it is now %v", in.Policy())
	}
}

func TestRunSourcePolicyCanBeGrantedByTheCaller(t *testing.T) {
	in, _ := quietInterpreter()
	p := Policy{AllowExec: true}
	res, err := in.Run(RunOptions{Source: "def main = echo hello-from-source\n", Args: []string{"a"}, Policy: &p})
	if err != nil || len(res.Lines) == 0 || !strings.Contains(res.Lines[0], "hello-from-source") {
		t.Fatalf("with AllowExec the Source task must run: lines %q, err %v", res.Lines, err)
	}
	// ... but only that capability: a file step is still denied
	p2 := Policy{AllowExec: true}
	_, err = in.Run(RunOptions{Source: "def main = write_file /tmp/never-written-by-run-test\n", Args: []string{"a"}, Policy: &p2})
	if err == nil {
		t.Fatal("AllowExec must not grant the fs capability")
	}
	if _, statErr := os.Stat("/tmp/never-written-by-run-test"); statErr == nil {
		os.Remove("/tmp/never-written-by-run-test")
		t.Fatal("write_file ran without the fs capability")
	}
}

func TestRunSourceAndScriptAreExclusive(t *testing.T) {
	in, _ := quietInterpreter()
	if _, err := in.Run(RunOptions{Script: "x.gatwd", Source: "const A = 1"}); err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("err = %v", err)
	}
}

func TestRunScriptFromThePathStaysTrusted(t *testing.T) {
	// the command-line script keeps the interpreter's policy (trusted for NewInterpreter)
	path := writeScript(t, "def main = echo hello-from-file\n")
	in, _ := quietInterpreter()
	res, err := in.Run(RunOptions{Script: path, Args: []string{"a"}})
	if err != nil || len(res.Lines) == 0 || !strings.Contains(res.Lines[0], "hello-from-file") {
		t.Fatalf("lines %q, err %v", res.Lines, err)
	}
}

func TestRunPolicyOverridesAScriptPath(t *testing.T) {
	path := writeScript(t, "def main = echo hello-from-file\n")
	in, _ := quietInterpreter()
	deny := DenyAll()
	_, err := in.Run(RunOptions{Script: path, Args: []string{"a"}, Policy: &deny})
	if err == nil {
		t.Fatal("an explicit DenyAll must apply to a script given by path")
	}
}

// ---- context: timeout and cancel ----

func TestRunTimeoutKillsASlowCommand(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	in, _ := quietInterpreter()
	start := time.Now()
	_, err := in.Run(RunOptions{Context: ctx, Eval: []string{"spawn StrVal: seed |> sleep 20"}})
	elapsed := time.Since(start)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want the deadline", err)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("the run took %v: the child process was not killed", elapsed)
	}
}

func TestRunTimeoutStopsAPureStream(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	in, _ := quietInterpreter()
	start := time.Now()
	// roughly a minute of work if it were allowed to finish
	_, err := in.Run(RunOptions{Context: ctx, Eval: []string{"spawn StrVal: s |> mock_stream 400000000 |> fold sum"}})
	elapsed := time.Since(start)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want the deadline", err)
	}
	if elapsed > 10*time.Second {
		t.Fatalf("the run took %v: the pipeline did not observe the context", elapsed)
	}
}

func TestRunWithACancelledContextRunsNothing(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "ran.txt")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	in, _ := quietInterpreter()
	_, err := in.Run(RunOptions{Context: ctx, Source: "", Eval: []string{"spawn StrVal: seed |> write_file " + marker}, Policy: &Policy{AllowFS: true}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if _, statErr := os.Stat(marker); statErr == nil {
		t.Fatal("a statement ran under an already cancelled context")
	}
}

func TestRunStopsBetweenTopLevelStatements(t *testing.T) {
	// the first statement takes long enough for the deadline to pass; the second must not run
	marker := filepath.Join(t.TempDir(), "second.txt")
	script := "spawn StrVal: s |> mock_stream 400000000 |> fold sum\n" +
		"spawn StrVal: seed |> write_file " + marker + "\n"
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	in, _ := quietInterpreter()
	_, err := in.Run(RunOptions{Context: ctx, Source: script, Policy: &Policy{AllowFS: true}})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want the deadline", err)
	}
	if _, statErr := os.Stat(marker); statErr == nil {
		t.Fatal("the statement after the stopped one still ran")
	}
}

func TestRunRestoresTheContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in, _ := quietInterpreter()
	if _, err := in.Run(RunOptions{Context: ctx, Eval: []string{"spawn StrVal: seed |> emit 1"}}); err != nil {
		t.Fatal(err)
	}
	cancel()
	// a later run without a context is not affected by the cancelled one
	if _, err := in.Run(RunOptions{Eval: []string{"spawn StrVal: seed |> emit 1"}}); err != nil {
		t.Fatalf("the cancelled context of an earlier run leaked into this one: %v", err)
	}
}

// ---- output limit ----

func TestRunOutputLimitStopsAndReports(t *testing.T) {
	in, buf := quietInterpreter()
	_, err := in.Run(RunOptions{MaxOutput: 40, Eval: []string{
		"spawn StrVal: seed |> emit aaaaaaaaaa, bbbbbbbbbb, cccccccccc, dddddddddd, eeeeeeeeee, ffffffffff",
	}})
	if !errors.Is(err, ErrOutputLimit) {
		t.Fatalf("err = %v, want ErrOutputLimit", err)
	}
	if !strings.Contains(err.Error(), "40 bytes") {
		t.Errorf("the error should name the limit: %v", err)
	}
	if buf.Len() > 40 {
		t.Fatalf("%d bytes were written past the limit of 40: %q", buf.Len(), buf.String())
	}
}

func TestRunOutputLimitAppliesToTaskLines(t *testing.T) {
	in, _ := quietInterpreter()
	res, err := in.Run(RunOptions{MaxOutput: 25, Source: "def main = emit aaaaaaaaaa, bbbbbbbbbb, cccccccccc, dddddddddd\n", Policy: &Policy{}})
	if !errors.Is(err, ErrOutputLimit) {
		t.Fatalf("err = %v, want ErrOutputLimit", err)
	}
	total := 0
	for _, l := range res.Lines {
		total += len(l) + 1
	}
	if total > 25 || len(res.Lines) == 0 {
		t.Fatalf("lines %q total %d bytes; want some lines, at most 25 bytes", res.Lines, total)
	}
}

func TestRunWithinTheLimitsSucceeds(t *testing.T) {
	// the control: bounds that are not exceeded change nothing
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	in, buf := quietInterpreter()
	res, err := in.Run(RunOptions{Context: ctx, MaxOutput: 1 << 20, Source: "def main = emit 1, 2, 3\n", Policy: &Policy{}})
	if err != nil || len(res.Lines) != 3 {
		t.Fatalf("lines %q, err %v, out %q", res.Lines, err, buf.String())
	}
}

func TestCancelledContextStopsAStepBeforeItsEffect(t *testing.T) {
	// no Run, no statement loop: the pipeline step itself must refuse to start
	marker := filepath.Join(t.TempDir(), "step.txt")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	in, _ := quietInterpreter()
	in.SetContext(ctx)
	got := evalStrings(in, "spawn StrVal: seed |> write_file "+marker)
	if !anyContains(got, "the run was stopped: context canceled") {
		t.Fatalf("got %q, want the stop reported as an error value", got)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("write_file ran under a cancelled context")
	}
}
