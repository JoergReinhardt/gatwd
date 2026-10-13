package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

const toyRules = `
def rule_a = \x -> x + 1
def rule_b = \x -> x + 10
def entry = \x -> [(rule_a x), (rule_b x)]
`

// toyTree is an Appliance command line over a miniature domain: --source is its data source.
func toyTree(embedded string) *ApplianceTree {
	return NewApplianceTree(ApplianceSpec{
		Name:      "toy",
		RulesPath: "/nonexistent/rules.gatwd",
		Bind: func(cmd *cobra.Command) ApplianceSetup {
			src := cmd.PersistentFlags().String("source", "", "the number to check")
			return func(s ApplianceSettings) (*Appliance, string, error) {
				if *src == "" {
					return nil, "", UsageError(errors.New("no --source given"))
				}
				app, err := NewAppliance(append(s.Options(),
					WithEmbeddedScript(embedded), WithRequiredDefs("entry"))...)
				return app, "entry " + *src, err
			}
		},
		Embedded:    func() (string, error) { return embedded, nil },
		LegacyFlags: []string{"source", "rules", "rules-mode", "interval"},
	})
}

func toyRun(t *testing.T, tree *ApplianceTree, argv ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errb bytes.Buffer
	code = tree.Execute(argv, strings.NewReader(""), &out, &errb)
	return code, out.String(), errb.String()
}

func TestApplianceTreeCheckAndExitCodes(t *testing.T) {
	override := writeFile(t, "rules.gatwd", "def rule_b = \\x -> x + 100\n")
	reuse := writeFile(t, "reuse.gatwd", "def entry = \\x -> [(rule_a x)]\n")
	broken := writeFile(t, "broken.gatwd", "def rule_b = \\x -> (x + 100\n")
	rows := []struct {
		name       string
		argv       []string
		code       int
		stdout     string // exact
		stderrHas  []string
		stderrNone []string
	}{
		{"check, embedded rules", []string{"check", "--source", "1"}, 0, "[2 11]\n", []string{"toy: rules: embedded rules, no override"}, nil},
		{"the legacy form without a subcommand checks once", []string{"--source", "1"}, 0, "[2 11]\n", nil, nil},
		{"a layered override is reported", []string{"check", "--source", "1", "--rules", override}, 0, "[2 101]\n",
			[]string{"layered on the embedded rules", "overridden: rule_b"}, nil},
		{"layer mode: the same override reuses the embedded rule_a", []string{"check", "--source", "1", "--rules", reuse}, 0, "[2]\n", []string{"overridden: entry"}, nil},
		{"replace mode: the embedded rule_a is gone", []string{"check", "--source", "1", "--rules", reuse, "--rules-mode", "replace"}, 0, "",
			[]string{"REPLACES"}, nil},
		{"a broken override fails the command", []string{"check", "--source", "1", "--rules", broken}, 1, "",
			[]string{"toy: " + broken, "unbalanced delimiter"}, nil},
		{"an unreadable override (a directory) fails the command", []string{"check", "--source", "1", "--rules", filepath.Dir(broken)}, 1, "", []string{"toy:"}, nil},
		{"no data source is a malformed command line", []string{"check"}, 2, "", []string{"no --source given", "--help"}, nil},
		{"unknown rules mode", []string{"check", "--source", "1", "--rules-mode", "merge"}, 2, "", []string{"layer or replace"}, nil},
		{"unknown flag", []string{"check", "--source", "1", "--bogus"}, 2, "", []string{"unknown flag"}, nil},
		{"unexpected argument", []string{"chek"}, 2, "", []string{"unexpected argument"}, nil},
		{"--interval is not a check flag", []string{"check", "--source", "1", "--interval", "1s"}, 2, "", []string{"unknown flag"}, nil},
		{"run needs an interval", []string{"run", "--source", "1"}, 2, "", []string{"--interval"}, nil},
		{"negative interval", []string{"--source", "1", "--interval", "-1s"}, 2, "", []string{"negative"}, nil},
		{"single-dash legacy flags still work, with a notice", []string{"-source", "1", "-rules", override}, 0, "[2 101]\n",
			[]string{"-source is deprecated; use --source", "-rules is deprecated; use --rules"}, nil},
		{"version", []string{"version"}, 0, "toy ", nil, nil},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			code, stdout, stderr := toyRun(t, toyTree(toyRules), r.argv...)
			if code != r.code {
				t.Fatalf("exit %d, want %d\nstdout: %s\nstderr: %s", code, r.code, stdout, stderr)
			}
			if r.name == "version" {
				if !strings.HasPrefix(stdout, r.stdout) {
					t.Errorf("stdout = %q", stdout)
				}
			} else if r.name == "replace mode: the embedded rule_a is gone" {
				if !strings.Contains(stdout, "rule_a") {
					t.Errorf("stdout = %q, want an error naming rule_a", stdout)
				}
			} else if stdout != r.stdout {
				t.Errorf("stdout = %q, want %q", stdout, r.stdout)
			}
			for _, want := range r.stderrHas {
				if !strings.Contains(stderr, want) {
					t.Errorf("stderr lacks %q:\n%s", want, stderr)
				}
			}
		})
	}
}

func TestApplianceTreeRulesCat(t *testing.T) {
	code, stdout, _ := toyRun(t, toyTree(toyRules), "rules", "cat")
	if code != 0 || stdout != toyRules {
		t.Fatalf("rules cat: exit %d, stdout %q", code, stdout)
	}
	// An appliance with no embedded-rules accessor has no `rules` command.
	tree := NewApplianceTree(ApplianceSpec{Name: "bare"})
	if code, _, _ := toyRun(t, tree, "rules", "cat"); code != 2 {
		t.Errorf("rules cat without Embedded: exit %d, want 2", code)
	}
}

// The interval loop re-reads the override before each check: a good edit takes effect, a
// broken edit is reported and the last good rules keep checking, a fixed edit recovers.
func TestApplianceTreeRunReloadsAndKeepsLastGood(t *testing.T) {
	dir := t.TempDir()
	rules := filepath.Join(dir, "rules.gatwd")
	write := func(text string) {
		if err := os.WriteFile(rules, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("def rule_b = \\x -> x + 100\n")

	tree := toyTree(toyRules)
	edits := []string{
		"def rule_b = \\x -> (x + 100\n", // after check 1: broken
		"def rule_b = \\x -> x + 200\n",  // after check 2: fixed
	}
	var waited []time.Duration
	tree.sleep = func(d time.Duration) bool {
		waited = append(waited, d)
		if len(waited) > len(edits) {
			return false // three checks done
		}
		write(edits[len(waited)-1])
		return true
	}
	code, stdout, stderr := toyRun(t, tree, "run", "--source", "1", "--rules", rules, "--interval", "5m")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, stderr)
	}
	if want := "[2 101]\n[2 101]\n[2 201]\n"; stdout != want {
		t.Fatalf("stdout = %q, want %q (the broken edit must leave the last good rules checking)", stdout, want)
	}
	if !reflect.DeepEqual(waited, []time.Duration{5 * time.Minute, 5 * time.Minute, 5 * time.Minute}) {
		t.Errorf("waited %v", waited)
	}
	for _, want := range []string{"reload failed, still using the previous rules", "unbalanced delimiter", "overridden: rule_b"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr lacks %q:\n%s", want, stderr)
		}
	}
	// The report is printed when it changes, not on every reload: the identical report of
	// check 3 versus check 1 (same override, same names) is not printed a second time.
	if n := strings.Count(stderr, "layered on the embedded rules"); n != 1 {
		t.Errorf("the unchanged layer report was printed %d times, want once:\n%s", n, stderr)
	}

	// A broken override at start is not recoverable by looping: the command fails.
	write("def rule_b = \\x -> (x + 100\n")
	tree2 := toyTree(toyRules)
	tree2.sleep = func(time.Duration) bool { t.Fatal("must not loop"); return false }
	if code, _, _ := toyRun(t, tree2, "run", "--source", "1", "--rules", rules, "--interval", "1s"); code != 1 {
		t.Errorf("broken override at start: exit %d, want 1", code)
	}
}

func TestRewriteLegacyFlags(t *testing.T) {
	names := []string{"manifests", "rules"}
	rows := []struct {
		in, out, rewritten []string
	}{
		{[]string{"-manifests", "d"}, []string{"--manifests", "d"}, []string{"manifests"}},
		{[]string{"-manifests=d", "-rules", "f", "-manifests", "e"}, []string{"--manifests=d", "--rules", "f", "--manifests", "e"}, []string{"manifests", "rules"}},
		{[]string{"--manifests", "d"}, []string{"--manifests", "d"}, nil},
		{[]string{"-h", "-x"}, []string{"-h", "-x"}, nil},
		{[]string{"-", "check"}, []string{"-", "check"}, nil},
		{[]string{"check", "--", "-manifests"}, []string{"check", "--", "-manifests"}, nil},
		{nil, []string{}, nil},
	}
	for _, r := range rows {
		out, rw := rewriteLegacyFlags(r.in, names)
		if !reflect.DeepEqual(out, r.out) || !reflect.DeepEqual(rw, r.rewritten) {
			t.Errorf("rewrite(%v) = %v, %v; want %v, %v", r.in, out, rw, r.out, r.rewritten)
		}
	}
}
