package main

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// Real-binary tests for SIGINT and SIGTERM: Kubernetes stops a pod with SIGTERM, and a
// container runs the operator as PID 1, so the signal must end the process promptly. The
// structural counterpart, with an injectable signal channel, is pkg/cli/appliance_signal_test.go.

// syncBuffer is a bytes.Buffer that the test may read while the process still writes to it.
type syncBuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func skipWithoutSignals(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		t.Skip("no POSIX signals on " + runtime.GOOS)
	}
}

// startOperator starts the binary and returns the process, its stdout lines and a way to wait
// for its end. Stderr is collected into the returned buffer once the process has ended.
func startOperator(t *testing.T, bin string, args ...string) (*exec.Cmd, <-chan string, *syncBuffer, <-chan error) {
	t.Helper()
	c := exec.Command(bin, args...)
	c.Env = append(os.Environ(), "KUBECONFIG="+filepath.Join(t.TempDir(), "absent"), "HOME="+t.TempDir(), "KUBERNETES_SERVICE_HOST=")
	stdout, err := c.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	errb := new(syncBuffer)
	c.Stderr = errb
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Process.Kill() })
	lines := make(chan string, 16)
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()
	done := make(chan error, 1)
	go func() { done <- c.Wait() }()
	return c, lines, errb, done
}

func exitCodeOf(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode()
	}
	t.Fatal(err)
	return -1
}

func waitEnd(t *testing.T, done <-chan error, what string) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(30 * time.Second):
		t.Fatalf("%s: the process did not stop after the signal", what)
		return nil
	}
}

// waitLine waits for the first stdout line, which proves the process is past its start-up and
// inside the loop or the check, so the signal is not racing the handler installation.
func waitLine(t *testing.T, lines <-chan string) string {
	t.Helper()
	select {
	case l, ok := <-lines:
		if !ok {
			t.Fatal("stdout closed before the first line")
		}
		return l
	case <-time.After(30 * time.Second):
		t.Fatal("no output from the process")
		return ""
	}
}

// `run` between two checks: a stop request ends it with status 0 and one line on stderr.
func TestOperatorRunStopsOnSignalWhileWaiting(t *testing.T) {
	skipWithoutSignals(t)
	bin := buildOperator(t)
	for _, sig := range []syscall.Signal{syscall.SIGTERM, syscall.SIGINT} {
		t.Run(sig.String(), func(t *testing.T) {
			// a long interval: only the signal can end the process within the test's patience
			c, lines, stderr, done := startOperator(t, bin, "run", "--manifests", "testdata/idle.yaml",
				"--rules", filepath.Join(t.TempDir(), "absent"), "--interval", "1h")
			if l := waitLine(t, lines); l != strings.TrimSpace(idleWant) {
				t.Fatalf("first check: %q", l)
			}
			if err := c.Process.Signal(sig); err != nil {
				t.Fatal(err)
			}
			if code := exitCodeOf(t, waitEnd(t, done, "run")); code != 0 {
				t.Fatalf("exit %d, want 0 (a requested stop is a normal shutdown)\nstderr: %s", code, stderr)
			}
			if got := stderr.String(); !strings.Contains(got, "gatwd-k8s-operator: received ") || !strings.Contains(got, ", stopping\n") {
				t.Errorf("stderr lacks the one-line stop notice:\n%s", got)
			}
		})
	}
}

// A check in flight (a real `sleep` run by an override) is cancelled and its child killed.
func TestOperatorStopsACheckInFlight(t *testing.T) {
	skipWithoutSignals(t)
	bin := buildOperator(t)
	slow := filepath.Join(t.TempDir(), "slow.gatwd")
	if err := os.WriteFile(slow, []byte("def compliance = sleep 300\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	base := []string{"--manifests", "testdata/idle.yaml", "--rules", slow, "--allow-exec"}
	rows := []struct {
		name   string
		args   []string
		code   int
		stderr string
	}{
		{"run: exit 0 and a notice", append([]string{"run", "--interval", "1h"}, base...), 0, "stopping, the check in progress was cancelled"},
		{"check: exit 1 and a message", append([]string{"check"}, base...), 1, "the check was interrupted (received terminated)"},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			c, _, stderr, done := startOperator(t, bin, r.args...)
			// the check has no output to wait for, and the handler is installed before the
			// rules load: poll until the process has started the sleep child (or the report of
			// the loaded rules appeared on stderr).
			waitStderr(t, stderr, "rules: ")
			if err := c.Process.Signal(syscall.SIGTERM); err != nil {
				t.Fatal(err)
			}
			start := time.Now()
			code := exitCodeOf(t, waitEnd(t, done, r.name))
			if code != r.code {
				t.Fatalf("exit %d, want %d\nstderr: %s", code, r.code, stderr)
			}
			if !strings.Contains(stderr.String(), r.stderr) {
				t.Errorf("stderr lacks %q:\n%s", r.stderr, stderr)
			}
			if time.Since(start) > 20*time.Second {
				t.Errorf("stopping took %v", time.Since(start))
			}
		})
	}
}

// waitStderr polls the process's stderr for text. The poll is bounded by a deadline, not by a
// fixed sleep: it returns as soon as the text is there.
func waitStderr(t *testing.T, b *syncBuffer, text string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for !strings.Contains(b.String(), text) {
		if time.Now().After(deadline) {
			t.Fatalf("stderr never contained %q", text)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
