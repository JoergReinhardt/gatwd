package cli_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

// embeddedRuleSet is a miniature of the operator's rules: two rules and an entry point.
const embeddedRuleSet = `
def rule_a = \x -> x + 1
def rule_b = \x -> x + 10
def entry = \x -> [(rule_a x), (rule_b x)]
const LIMIT = 5
`

func writeRules(t *testing.T, dir, name, text string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func newRules(t *testing.T, override string, opts ...cli.ApplianceOption) (*cli.Appliance, error) {
	t.Helper()
	all := append([]cli.ApplianceOption{
		cli.WithEmbeddedScript(embeddedRuleSet),
		cli.WithConfigMapOverride(override),
		cli.WithRequiredDefs("entry"),
	}, opts...)
	return cli.NewAppliance(all...)
}

func execOne(t *testing.T, app *cli.Appliance, line string) string {
	t.Helper()
	vals := app.Execute(line)
	if len(vals) != 1 {
		t.Fatalf("%s: want one value, got %v", line, vals)
	}
	return fmt.Sprint(vals[0])
}

// An override redefines rule_b, adds rule_c and a new entry point that REUSES the embedded
// rule_a: that is only possible when the override is layered on top of the embedded rules.
const layeringOverride = `
def rule_b = \x -> x + 100
def rule_c = \x -> x + 1000
def entry = \x -> [(rule_a x), (rule_b x), (rule_c x)]
const LIMIT = 6
const EXTRA = 1
`

func TestApplianceLayersOverrideByDefault(t *testing.T) {
	dir := t.TempDir()
	app, err := newRules(t, writeRules(t, dir, "rules.gatwd", layeringOverride))
	if err != nil {
		t.Fatal(err)
	}
	if got := execOne(t, app, "entry 1"); got != "[2 101 1001]" {
		t.Fatalf("layered entry = %s, want [2 101 1001] (embedded rule_a reused, rule_b redefined, rule_c added)", got)
	}
	rep := app.Report()
	if rep.Mode != cli.OverrideLayer || rep.Override == "" {
		t.Fatalf("report = %+v", rep)
	}
	// Defs are listed before constants; rule_a is untouched so it does not appear.
	if want := []string{"entry", "rule_b", "const LIMIT"}; !reflect.DeepEqual(rep.Overridden, want) {
		t.Errorf("Overridden = %v, want %v", rep.Overridden, want)
	}
	if want := []string{"rule_c", "const EXTRA"}; !reflect.DeepEqual(rep.Added, want) {
		t.Errorf("Added = %v, want %v", rep.Added, want)
	}
	text := rep.String()
	for _, want := range []string{"layered", "overridden: entry, rule_b, const LIMIT", "added: rule_c, const EXTRA"} {
		if !strings.Contains(text, want) {
			t.Errorf("report text lacks %q:\n%s", want, text)
		}
	}
}

func TestApplianceReplaceModeDropsEmbeddedRules(t *testing.T) {
	dir := t.TempDir()
	override := writeRules(t, dir, "rules.gatwd", layeringOverride)
	app, err := newRules(t, override, cli.WithOverrideMode(cli.OverrideReplace))
	if err != nil {
		t.Fatal(err)
	}
	// rule_a lived only in the embedded rules; in replace mode the override cannot reach it.
	got := execOne(t, app, "entry 1")
	if !strings.Contains(got, "rule_a") {
		t.Fatalf("replace mode: want an error naming the missing embedded rule_a, got %s", got)
	}
	rep := app.Report()
	if rep.Mode != cli.OverrideReplace || len(rep.Overridden) != 0 || len(rep.Added) != 0 {
		t.Errorf("replace report = %+v, want mode replace and no diff", rep)
	}
	if !strings.Contains(rep.String(), "REPLACES") {
		t.Errorf("replace report text = %q", rep.String())
	}

	// Replace mode without an override file still runs the embedded rules.
	app2, err := newRules(t, filepath.Join(dir, "absent.gatwd"), cli.WithOverrideMode(cli.OverrideReplace))
	if err != nil {
		t.Fatal(err)
	}
	if got := execOne(t, app2, "entry 1"); got != "[2 11]" {
		t.Fatalf("replace mode, no override: entry = %s, want the embedded [2 11]", got)
	}
	if app2.Report().Override != "" || app2.Report().String() != "rules: embedded rules, no override" {
		t.Errorf("no-override report = %+v", app2.Report())
	}
}

// A copy of the embedded rules is not an override: only a def whose body changed is reported.
func TestApplianceReportIgnoresIdenticalRedefinition(t *testing.T) {
	dir := t.TempDir()
	app, err := newRules(t, writeRules(t, dir, "rules.gatwd", "def rule_a = \\x -> x + 1\ndef rule_b = \\x -> x + 11\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := app.Report().Overridden, []string{"rule_b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Overridden = %v, want %v (rule_a is an identical copy)", got, want)
	}
	if len(app.Report().Added) != 0 {
		t.Errorf("Added = %v, want none", app.Report().Added)
	}
}

// Every way an override can be broken is a LoadError at construction — never a silent
// fall-back to the embedded rules.
func TestApplianceBrokenOverrideFailsLoudlyAtLoad(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name, text, want string
		mode             cli.OverrideMode
	}{
		{"unbalanced delimiter", "def rule_a = \\x -> (x + 1\n", "unbalanced delimiter", cli.OverrideLayer},
		{"failing statement", "1 / 0\n", "[ERROR]", cli.OverrideLayer},
		{"unresolvable import", "import \"no_such_file.gatwd\"\n", "no_such_file.gatwd", cli.OverrideLayer},
		{"entry point missing in replace mode", "def rule_a = \\x -> x + 1\n", `required def "entry" is not defined`, cli.OverrideReplace},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := writeRules(t, dir, "broken.gatwd", tc.text)
			app, err := newRules(t, p, cli.WithOverrideMode(tc.mode))
			var le *cli.LoadError
			if !errors.As(err, &le) || app != nil {
				t.Fatalf("want a *LoadError and no appliance, got app=%v err=%v", app, err)
			}
			if le.Source != p || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want source %s and mention of %q", err, p, tc.want)
			}
		})
	}

	// A ConfigMap mounted where a file was expected is an error, not "no override".
	if _, err := newRules(t, dir); err == nil {
		t.Error("an override path that is a directory must be a load error")
	}
}

// A hot reload that finds a broken override reports it and keeps serving the last good rules.
func TestApplianceReloadKeepsLastGoodRules(t *testing.T) {
	dir := t.TempDir()
	p := writeRules(t, dir, "rules.gatwd", "def rule_b = \\x -> x + 100\n")
	app, err := newRules(t, p)
	if err != nil {
		t.Fatal(err)
	}
	if got := execOne(t, app, "entry 1"); got != "[2 101]" {
		t.Fatalf("v1 = %s", got)
	}
	v1 := app.Report()

	writeRules(t, dir, "rules.gatwd", "def rule_b = \\x -> (x + 100\n")
	err = app.Reload()
	var le *cli.LoadError
	if !errors.As(err, &le) {
		t.Fatalf("Reload of a broken override: err = %v, want a *LoadError", err)
	}
	if got := execOne(t, app, "entry 1"); got != "[2 101]" {
		t.Fatalf("after a failed reload the last good rules must keep serving, got %s", got)
	}
	if !reflect.DeepEqual(app.Report(), v1) {
		t.Errorf("a failed reload must not change the report: %+v vs %+v", app.Report(), v1)
	}

	writeRules(t, dir, "rules.gatwd", "def rule_b = \\x -> x + 200\n")
	if err := app.Reload(); err != nil {
		t.Fatalf("fixed override: %v", err)
	}
	if got := execOne(t, app, "entry 1"); got != "[2 201]" {
		t.Fatalf("v2 = %s", got)
	}
}

// F16: an import in an override resolves against the override file's directory, wherever
// the process runs.
func TestApplianceImportsResolveAgainstOverrideDirectory(t *testing.T) {
	dir := t.TempDir()
	writeRules(t, dir, "extra.gatwd", "def rule_extra = \\x -> x + 7\n")
	p := writeRules(t, dir, "rules.gatwd", "import \"extra.gatwd\"\ndef entry = \\x -> [(rule_a x), (rule_extra x)]\n")
	// an import reads a file: the override needs the fs capability (denied by default, W7)
	if _, err := newRules(t, p); err == nil || !strings.Contains(err.Error(), "requires the fs capability") {
		t.Fatalf("an import in an override must be denied by default: err = %v", err)
	}
	app, err := newRules(t, p, cli.WithPolicy(cli.Policy{AllowFS: true}))
	if err != nil {
		t.Fatalf("import beside the override: %v", err)
	}
	if got := execOne(t, app, "entry 1"); got != "[2 8]" {
		t.Fatalf("entry = %s, want [2 8]", got)
	}
	if !reflect.DeepEqual(app.Report().Added, []string{"rule_extra"}) || !reflect.DeepEqual(app.Report().Overridden, []string{"entry"}) {
		t.Errorf("report = %+v, want rule_extra added and entry overridden", app.Report())
	}
}

// Embedded rules read from an fs.FS import their siblings inside it, not from the disk.
func TestApplianceEmbeddedFSImportsStayInsideTheFS(t *testing.T) {
	fsys := fstest.MapFS{
		"rules/main.gatwd":   {Data: []byte("import \"part.gatwd\"\ndef entry = \\x -> rule_part x\n")},
		"rules/part.gatwd":   {Data: []byte("def rule_part = \\x -> x + 3\n")},
		"rules/broken.gatwd": {Data: []byte("import \"absent.gatwd\"\n")},
	}
	app, err := cli.NewAppliance(cli.WithEmbeddedFS(fsys, "rules/main.gatwd"))
	if err != nil {
		t.Fatal(err)
	}
	if got := execOne(t, app, "entry 1"); got != "4" {
		t.Fatalf("entry = %s, want 4", got)
	}
	// A broken embedded rule set is a load error as well (a programmer error, surfaced).
	if _, err := cli.NewAppliance(cli.WithEmbeddedFS(fsys, "rules/broken.gatwd")); err == nil || !strings.Contains(err.Error(), "absent.gatwd") {
		t.Fatalf("broken embedded import: err = %v", err)
	}
}

func TestParseOverrideMode(t *testing.T) {
	for _, m := range []cli.OverrideMode{cli.OverrideLayer, cli.OverrideReplace} {
		got, err := cli.ParseOverrideMode(m.String())
		if err != nil || got != m {
			t.Errorf("round trip of %v: %v, %v", m, got, err)
		}
	}
	if _, err := cli.ParseOverrideMode("merge"); err == nil || !strings.Contains(err.Error(), "layer or replace") {
		t.Errorf("unknown mode: err = %v", err)
	}
	if cli.OverrideMode(0) != cli.OverrideLayer {
		t.Error("the zero OverrideMode must be layer (decision D5)")
	}
}
