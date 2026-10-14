package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Real-binary tests for the bound on loading rules: --timeout also stops a start-up load or a
// hot reload whose top-level statements never finish. The hanging statement is a pure pipeline
// over a stream that would take seconds, against a timeout of a fraction of that.

const (
	hangingStatement = "spawn StrVal: s |> mock_stream 40000000 |> fold sum\n"
	ruleIdle         = "def compliance = k8s_graph |> fold (\\_ g -> rule_idle_deployment g) 0\n"
	rulePod          = "def compliance = k8s_graph |> fold (\\_ g -> rule_unselected_pod g) 0\n"
	idleLine         = "[[IdleDeployment Deployment/default/lonely this Deployment owns no ReplicaSet, so it manages no Pods]]"
	podLine          = "[[UnselectedPod Pod/default/stray no Service selects this Pod, so it receives no traffic]]"
)

// writeAtomic replaces path in one step, so the loop never reads half an override.
func writeAtomic(t *testing.T, path, text string) {
	t.Helper()
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmp, path); err != nil {
		t.Fatal(err)
	}
}

// At start-up a load that runs over --timeout is a failed start: exit 1, the message names the
// override and the budget, and nothing is checked. Without the flag nothing is bounded.
func TestOperatorStartupLoadTimeout(t *testing.T) {
	bin := buildOperator(t)
	hang := filepath.Join(t.TempDir(), "hang.gatwd")
	if err := os.WriteFile(hang, []byte(hangingStatement+ruleIdle), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"check", "--manifests", "testdata/idle.yaml", "--rules", hang, "--timeout", "300ms"},
		{"run", "--manifests", "testdata/idle.yaml", "--rules", hang, "--timeout", "300ms", "--interval", "1h"},
		{"check", "--manifests", "testdata/idle.yaml", "--rules", hang, "--timeout", "300ms", "--format", "json"},
	} {
		t.Run(args[0]+" "+args[len(args)-1], func(t *testing.T) {
			start := time.Now()
			code, stdout, stderr := operator(t, bin, args...)
			if time.Since(start) > 20*time.Second {
				t.Fatalf("took %v: the load was not stopped", time.Since(start))
			}
			if code != 1 || stdout != "" {
				t.Fatalf("exit %d, stdout %q, want 1 and nothing\nstderr: %s", code, stdout, stderr)
			}
			for _, want := range []string{"gatwd-k8s-operator: " + hang, "loading the rules timed out after 300ms"} {
				if !strings.Contains(stderr, want) {
					t.Errorf("stderr lacks %q:\n%s", want, stderr)
				}
			}
		})
	}
	// negative control: a load that fits its budget is untouched by the flag
	fast := filepath.Join(t.TempDir(), "fast.gatwd")
	if err := os.WriteFile(fast, []byte("spawn StrVal: s |> mock_stream 100 |> fold sum\n"+ruleIdle), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := operator(t, bin, "check", "--manifests", "testdata/idle.yaml", "--rules", fast, "--timeout", "1m")
	if code != 0 || stdout != idleLine+"\n" {
		t.Errorf("a load inside its budget: exit %d, stdout %q\nstderr: %s", code, stdout, stderr)
	}
}

// The running daemon: a hot reload that hangs is stopped at --timeout, reported, and the last
// good rules keep serving every check; fixing the override recovers; SIGTERM ends it with 0.
func TestOperatorHotReloadTimeoutKeepsServing(t *testing.T) {
	skipWithoutSignals(t)
	bin := buildOperator(t)
	override := filepath.Join(t.TempDir(), "rules.gatwd")
	writeAtomic(t, override, ruleIdle)
	c, lines, stderr, done := startOperator(t, bin, "run", "--manifests", "testdata/idle.yaml",
		"--rules", override, "--timeout", "500ms", "--interval", "100ms")
	if l := waitLine(t, lines); l != idleLine {
		t.Fatalf("first check: %q", l)
	}

	writeAtomic(t, override, hangingStatement+rulePod)
	waitStderr(t, stderr, "reload failed, still using the previous rules")
	if got := stderr.String(); !strings.Contains(got, "loading the rules timed out after 500ms") {
		t.Errorf("the reload failure must name the timeout:\n%s", got)
	}
	// still the previous rules, check after check
	for k := 0; k < 3; k++ {
		if l := waitLine(t, lines); l != idleLine {
			t.Fatalf("check %d after the failed reload: %q, want the last good rules' %q", k+1, l, idleLine)
		}
	}

	writeAtomic(t, override, rulePod)
	deadline := time.Now().Add(30 * time.Second)
	for {
		l := waitLine(t, lines)
		if l == podLine {
			break
		}
		if l != idleLine {
			t.Fatalf("unexpected line %q while waiting for the fixed rules", l)
		}
		if time.Now().After(deadline) {
			t.Fatal("the fixed override never took effect")
		}
	}

	if err := c.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if code := exitCodeOf(t, waitEnd(t, done, "run")); code != 0 {
		t.Fatalf("exit %d, want 0\nstderr: %s", code, stderr)
	}
}
