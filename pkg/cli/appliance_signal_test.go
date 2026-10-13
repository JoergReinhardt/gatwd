package cli

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// Tests for the stop requests (SIGINT, SIGTERM) of `check` and `run`. The signal source is
// injected (ApplianceTree.signals), so no test sends a real signal or waits for luck: the
// request is raised from inside the run — by a pipeline step while a check is in flight, by
// the writer that receives the report while the loop is about to wait — and the code under test
// is what happens next. The real-binary counterpart is cmd/gatwd-k8s-operator/signal_test.go.

// signalFixture is a toy Appliance command line whose stop requests the test raises itself.
type signalFixture struct {
	tree *ApplianceTree
	sigs chan os.Signal
	// trigger is the step `trigger`: it runs for every node of a check and calls onNode.
	onNode func()
	exits  chan int
}

func newSignalFixture(t *testing.T) *signalFixture {
	t.Helper()
	f := &signalFixture{sigs: make(chan os.Signal, 4), exits: make(chan int, 1)}
	tree := NewApplianceTree(ApplianceSpec{
		Name: "toy",
		Bind: func(cmd *cobra.Command) ApplianceSetup {
			src := cmd.PersistentFlags().String("source", "", "the pipeline to evaluate")
			return func(s ApplianceSettings) (*Appliance, string, error) {
				app, err := NewAppliance(append(s.Options(),
					WithEmbeddedScript(toyRules), WithRequiredDefs("entry"),
					WithStep(SinkStep("trigger", func(any) {
						if f.onNode != nil {
							f.onNode()
						}
					})))...)
				return app, *src, err
			}
		},
	})
	tree.signals = func() (<-chan os.Signal, func()) { return f.sigs, func() {} }
	tree.exit = func(code int) { f.exits <- code }
	f.tree = tree
	return f
}

// run executes argv with stdout as given (a plain buffer when nil) and returns the exit status
// and the two output streams. It fails the test if the command does not end by itself.
func (f *signalFixture) run(t *testing.T, stdout *bytes.Buffer, w func(*bytes.Buffer) writerFunc, argv ...string) (int, string, string) {
	t.Helper()
	var errb bytes.Buffer
	if stdout == nil {
		stdout = new(bytes.Buffer)
	}
	done := make(chan int, 1)
	go func() {
		out := w
		if out == nil {
			done <- f.tree.Execute(argv, strings.NewReader(""), stdout, &errb)
			return
		}
		done <- f.tree.Execute(argv, strings.NewReader(""), out(stdout), &errb)
	}()
	select {
	case code := <-done:
		return code, stdout.String(), errb.String()
	case <-time.After(30 * time.Second):
		t.Fatal("the command did not stop; a stop request must end it promptly")
		return 0, "", ""
	}
}

// writerFunc adapts a function to io.Writer.
type writerFunc func([]byte) (int, error)

func (w writerFunc) Write(p []byte) (int, error) { return w(p) }

// onFirstWrite wraps the buffer so that the first report written raises sig: the loop has just
// finished a check and is about to wait for the next one.
func onFirstWrite(sig chan<- os.Signal, s os.Signal) func(*bytes.Buffer) writerFunc {
	var once sync.Once
	return func(b *bytes.Buffer) writerFunc {
		return func(p []byte) (int, error) {
			n, err := b.Write(p)
			once.Do(func() { sig <- s })
			return n, err
		}
	}
}

const endlessCheck = "spawn StrVal: s |> mock_stream 400000000 |> trigger |> fold sum"

func TestApplianceRunStopsOnSIGTERMDuringACheck(t *testing.T) {
	f := newSignalFixture(t)
	var once sync.Once
	f.onNode = func() { once.Do(func() { f.sigs <- syscall.SIGTERM }) }
	start := time.Now()
	code, stdout, stderr := f.run(t, nil, nil, "run", "--source", endlessCheck, "--interval", "1h")
	if code != 0 {
		t.Fatalf("a stop request is a normal end of a daemon: exit %d, want 0\nstderr: %s", code, stderr)
	}
	if stdout != "" {
		t.Errorf("the cancelled check must print nothing, got %q", stdout)
	}
	if want := "toy: received terminated, stopping, the check in progress was cancelled\n"; !strings.HasSuffix(stderr, want) {
		t.Errorf("stderr = %q, want it to end with %q", stderr, want)
	}
	if elapsed := time.Since(start); elapsed > 20*time.Second {
		t.Errorf("stopping took %v: the check was not cancelled", elapsed)
	}
}

func TestApplianceRunStopsOnSIGINTWhileWaiting(t *testing.T) {
	f := newSignalFixture(t)
	code, stdout, stderr := f.run(t, nil, onFirstWrite(f.sigs, syscall.SIGINT), "run", "--source", "entry 1", "--interval", "1h")
	if code != 0 {
		t.Fatalf("exit %d, want 0\nstderr: %s", code, stderr)
	}
	if stdout != "[2 11]\n" {
		t.Errorf("stdout = %q: the check before the request is reported once and nothing after", stdout)
	}
	if want := "toy: received interrupt, stopping\n"; !strings.HasSuffix(stderr, want) {
		t.Errorf("stderr = %q, want it to end with %q", stderr, want)
	}
	if strings.Contains(stderr, "in progress") {
		t.Errorf("no check was in progress: %q", stderr)
	}
}

// A check that failed before the request was reported when it failed; the request after it
// does not turn the shutdown into a failure.
func TestApplianceRunStopAfterAFailedCheckStillReportsIt(t *testing.T) {
	f := newSignalFixture(t)
	code, stdout, stderr := f.run(t, nil, onFirstWrite(f.sigs, syscall.SIGTERM),
		"run", "--source", "spawn StrVal: s |> no_such_step_m", "--interval", "1h")
	if code != 0 {
		t.Fatalf("exit %d, want 0\nstderr: %s", code, stderr)
	}
	if !strings.HasPrefix(stdout, "[ERROR] ") || !strings.Contains(stdout, "no_such_step_m") {
		t.Errorf("stdout = %q, want the failed check reported", stdout)
	}
	if !strings.Contains(stderr, "received terminated, stopping") {
		t.Errorf("stderr = %q", stderr)
	}
}

func TestApplianceCheckInterruptedExitsOne(t *testing.T) {
	for _, sig := range []os.Signal{syscall.SIGINT, syscall.SIGTERM} {
		t.Run(sig.String(), func(t *testing.T) {
			f := newSignalFixture(t)
			var once sync.Once
			f.onNode = func() { once.Do(func() { f.sigs <- sig }) }
			code, stdout, stderr := f.run(t, nil, nil, "check", "--source", endlessCheck)
			if code != 1 {
				t.Fatalf("exit %d, want 1: an interrupted check did not finish\nstderr: %s", code, stderr)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want nothing from a cancelled check", stdout)
			}
			if want := "toy: the check was interrupted (received " + sig.String() + ")\n"; stderr != want && !strings.HasSuffix(stderr, want) {
				t.Errorf("stderr = %q, want it to end with %q", stderr, want)
			}
		})
	}
}

// Negative control for the tests above: with no request the very same command runs to its end.
func TestApplianceCheckWithoutARequestIsUnaffected(t *testing.T) {
	f := newSignalFixture(t)
	code, stdout, stderr := f.run(t, nil, nil, "check", "--source", "entry 1")
	if code != 0 || stdout != "[2 11]\n" || strings.Contains(stderr, "interrupted") || strings.Contains(stderr, "stopping") {
		t.Errorf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

// A second request while the first is still being honoured ends the process at once: here the
// step that never reaches a node boundary again is a blocked one.
func TestApplianceSecondStopRequestExitsAtOnce(t *testing.T) {
	f := newSignalFixture(t)
	release := make(chan struct{})
	var once sync.Once
	f.onNode = func() {
		once.Do(func() {
			f.sigs <- syscall.SIGINT
			f.sigs <- syscall.SIGINT
			<-release // stuck: only the second request can end the process now
		})
	}
	done := make(chan struct{})
	var code int
	var stderr string
	go func() {
		code, _, stderr = f.run(t, nil, nil, "run", "--source", endlessCheck, "--interval", "1h")
		close(done)
	}()
	select {
	case got := <-f.exits:
		if got != 1 {
			t.Errorf("forced exit status %d, want 1", got)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("the second request did not end the process")
	}
	close(release)
	<-done
	if code != 0 || !strings.Contains(stderr, "received interrupt again, exiting without waiting for the check") {
		t.Errorf("exit %d, stderr %q", code, stderr)
	}
}
