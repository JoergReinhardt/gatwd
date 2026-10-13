package cli

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// Command-line tests for --timeout and --max-output (the per-check limits of an Appliance),
// through the toy domain of appliance_command_test.go. The slow source is a real `sleep`
// (so --allow-exec); the timeouts are short against it, the tests never wait for luck.

func TestApplianceTreeCheckLimits(t *testing.T) {
	const slow = "spawn StrVal: s |> sleep 20"
	const timedOut = "[ERROR] check timed out after 100ms\n"
	rows := []struct {
		name      string
		argv      []string
		code      int
		stdout    string
		stderrHas string
	}{
		{"a slow check times out and exits 1", []string{"check", "--source", slow, "--allow-exec", "--timeout", "100ms"}, 1, timedOut, ""},
		{"the legacy form takes the flag too", []string{"--source", slow, "--allow-exec", "--timeout", "100ms"}, 1, timedOut, ""},
		{"json: the error, a structured timeout diagnostic, exit 1", []string{"check", "--source", slow, "--allow-exec", "--timeout", "100ms", "--format", "json"}, 1,
			`{"status":"error","count":0,"violations":[],"errors":["check timed out after 100ms"],"diagnostics":[{"kind":"timeout","message":"check timed out after 100ms","limit":"100ms"}]}` + "\n", ""},
		{"a timeout is a failed check, not a violation: 1, never 3", []string{"check", "--source", slow, "--allow-exec", "--timeout", "100ms", "--fail-on-violation"}, 1, timedOut, "the check failed"},
		{"a fast check under a generous timeout is untouched", []string{"check", "--source", "entry 1", "--timeout", "1m"}, 0, "[2 11]\n", ""},
		{"no flag, no limit", []string{"check", "--source", "entry 1"}, 0, "[2 11]\n", ""},
		{"--timeout 0 is no limit", []string{"check", "--source", "entry 1", "--timeout", "0"}, 0, "[2 11]\n", ""},
		{"output exactly at the limit passes", []string{"check", "--source", "entry 1", "--max-output", "7"}, 0, "[2 11]\n", ""},
		{"output one byte over fails loudly, with none of the results", []string{"check", "--source", "entry 1", "--max-output", "6"}, 1,
			"[ERROR] check output exceeded the limit of 6 bytes\n", "the check failed"},
		{"json: output_limit diagnostic", []string{"check", "--source", "entry 1", "--max-output", "6", "--format", "json"}, 1,
			`{"status":"error","count":0,"violations":[],"errors":["check output exceeded the limit of 6 bytes"],"diagnostics":[{"kind":"output_limit","message":"check output exceeded the limit of 6 bytes","limit":"6"}]}` + "\n", ""},
		{"--max-output 0 is no limit", []string{"check", "--source", "entry 1", "--max-output", "0"}, 0, "[2 11]\n", ""},
		{"a negative timeout is a usage error", []string{"check", "--source", "entry 1", "--timeout", "-1s"}, 2, "", "--timeout must not be negative"},
		{"a negative output limit is a usage error", []string{"check", "--source", "entry 1", "--max-output", "-1"}, 2, "", "--max-output must not be negative"},
		{"an unparsable timeout is a usage error", []string{"check", "--source", "entry 1", "--timeout", "soon"}, 2, "", "invalid argument"},
		{"an unparsable output limit is a usage error", []string{"check", "--source", "entry 1", "--max-output", "lots"}, 2, "", "invalid argument"},
		{"the limits are validated before the source is read", []string{"check", "--timeout", "-1s"}, 2, "", "--timeout must not be negative"},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			code, stdout, stderr := toyRun(t, toyTree(toyRules), r.argv...)
			if code != r.code {
				t.Fatalf("exit %d, want %d\nstdout: %s\nstderr: %s", code, r.code, stdout, stderr)
			}
			if stdout != r.stdout {
				t.Errorf("stdout = %q\nwant     %q", stdout, r.stdout)
			}
			if !strings.Contains(stderr, r.stderrHas) {
				t.Errorf("stderr lacks %q:\n%s", r.stderrHas, stderr)
			}
		})
	}
}

func TestApplianceFlagsMapToTheLimits(t *testing.T) {
	var got ApplianceSettings
	tree := NewApplianceTree(ApplianceSpec{Name: "toy", Bind: func(cmd *cobra.Command) ApplianceSetup {
		return func(s ApplianceSettings) (*Appliance, string, error) {
			got = s
			return nil, "", UsageError(os.ErrInvalid)
		}
	}})
	toyRun(t, tree, "check", "--timeout", "30s", "--max-output", "1048576")
	if got.Timeout.String() != "30s" || got.MaxOutput != 1048576 {
		t.Errorf("settings = %+v", got)
	}
	got = ApplianceSettings{Timeout: 1, MaxOutput: 1}
	toyRun(t, tree, "check")
	if got.Timeout != 0 || got.MaxOutput != 0 {
		t.Errorf("the defaults must be no limit: %+v", got)
	}
	// the settings reach the Appliance
	app, err := NewAppliance(append(ApplianceSettings{Timeout: 3, MaxOutput: 9}.Options(), WithEmbeddedScript(toyRules))...)
	if err != nil || app.Timeout() != 3 || app.MaxOutput() != 9 {
		t.Errorf("Options() lost the limits: %v %v %v", err, app.Timeout(), app.MaxOutput())
	}
}

// A timed-out check must not end `run`: the loop reports it and checks again, and the output of
// the next check is its own. The override is rewritten between the iterations (the sleep hook
// plays the part of the interval), so iteration 1 is slow and iteration 2 is fast.
func TestApplianceRunLoopContinuesAfterATimeout(t *testing.T) {
	override := writeFile(t, "rules.gatwd", "def probe = sleep 20\n")
	tree := toyTree(toyRules)
	waits := 0
	tree.sleep = func(time.Duration) bool {
		waits++
		switch waits {
		case 1: // after the slow check: fix the rule, as an operator editing the ConfigMap would
			if err := os.WriteFile(override, []byte("def probe = echo recovered\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		case 3:
			return false
		}
		return true
	}
	code, stdout, stderr := toyRun(t, tree, "run", "--source", "spawn StrVal: s |> probe", "--rules", override,
		"--interval", "1s", "--timeout", "100ms", "--allow-exec")
	if code != 0 {
		t.Fatalf("exit %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) < 3 || lines[0] != "[ERROR] check timed out after 100ms" {
		t.Fatalf("the first check must report the timeout alone: %q", stdout)
	}
	for _, l := range lines[1:] {
		if strings.Contains(l, "[ERROR]") || strings.Contains(l, "timed out") {
			t.Errorf("a later, healthy check reports %q: the timeout leaked into it\n%s", l, stdout)
		}
	}
	if !strings.Contains(stdout, "recovered") {
		t.Errorf("the loop did not check again after the timeout:\n%s", stdout)
	}
}
