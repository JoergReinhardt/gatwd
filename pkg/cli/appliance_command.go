package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// This file is the command line of a rule-checking Appliance: one binary that evaluates a
// pipeline against embedded rules (plus an optional mounted override) once, or on an
// interval with hot reload. Like the Edition command tree it maps flags to library calls and
// holds no logic; what the rules are lives in .gatwd, and what the *source* of the checked
// data is belongs to the domain, which supplies it through ApplianceSpec.Bind.
//
//	NAME check [flags]            evaluate the rules once
//	NAME run --interval DUR       evaluate every DUR, re-reading the override before each check
//	NAME rules cat                print the embedded rules (copy, edit, mount as the override)
//	NAME version · NAME completion SHELL
//	NAME [flags]                  the legacy form: `run` when --interval is given, else `check`
//
// Flags every Appliance has: --rules FILE, --rules-mode layer|replace, --interval DUR,
// --format text|json, --fail-on-violation. A domain adds its own (its data source) in Bind.
//
// Exit status, the same mapping as the Edition command tree plus one: 0 the check ran and,
// with --fail-on-violation, found nothing; 1 the command failed — rules that do not load, a
// check that errored or yielded nothing; 2 a malformed command line; 3 (ExitViolations) the
// check ran correctly and, under --fail-on-violation, found violations. An error always wins
// over violations: a check that both errored and found violations exits 1. `run` is a daemon: a failing check is reported and the loop continues, and
// only a start-up failure ends it.

// ApplianceSettings is what the generic flags resolved to; a domain's ApplianceSetup turns it
// into an Appliance.
type ApplianceSettings struct {
	// Rules is the override file path (--rules).
	Rules string
	// Mode is how the override combines with the embedded rules (--rules-mode).
	Mode OverrideMode
}

// Options returns the ApplianceOptions that apply the settings: WithConfigMapOverride and
// WithOverrideMode. A domain setup passes them to NewAppliance next to its own edition and
// embedded rules.
func (s ApplianceSettings) Options() []ApplianceOption {
	return []ApplianceOption{WithConfigMapOverride(s.Rules), WithOverrideMode(s.Mode)}
}

// ApplianceSetup builds the Appliance and the pipeline to evaluate, once the command line has
// been parsed. This is where a domain decides which data source its flags selected — and
// where it builds anything expensive (a cluster client) only if that source needs it. An
// error wrapped with UsageError ends the command with status 2 ("no source given"); any
// other error with status 1.
type ApplianceSetup func(s ApplianceSettings) (app *Appliance, pipeline string, err error)

// ApplianceSpec describes the command line of one Appliance binary.
type ApplianceSpec struct {
	// Name is the program name shown in help and messages.
	Name string
	// RulesPath is the default for --rules (a mounted ConfigMap path).
	RulesPath string
	// Bind registers the domain's flags on cmd.PersistentFlags() and returns the setup that
	// reads them. It is called once per command-line build, so flag values are fresh.
	Bind func(cmd *cobra.Command) ApplianceSetup
	// Embedded returns the embedded rules for `rules cat`; nil omits the command.
	Embedded func() (string, error)
	// LegacyFlags are the long flag names (without dashes) that older releases accepted with
	// a single dash (`-manifests`); they are still accepted, with a notice on stderr.
	LegacyFlags []string
}

// ApplianceTree is an ApplianceSpec turned into a runnable command line.
type ApplianceTree struct {
	spec  ApplianceSpec
	sleep func(time.Duration) bool // waits between checks; false stops the loop (replaced in tests)
}

// NewApplianceTree builds the command line for spec.
func NewApplianceTree(spec ApplianceSpec) *ApplianceTree {
	if spec.Name == "" {
		spec.Name = "appliance"
	}
	return &ApplianceTree{spec: spec, sleep: func(d time.Duration) bool { time.Sleep(d); return true }}
}

// applianceFlags are the generic flags shared by the commands.
type applianceFlags struct {
	rules           string
	rulesMode       string
	interval        time.Duration
	format          string
	failOnViolation bool
}

// Root returns a fresh cobra root for the spec.
func (t *ApplianceTree) Root() *cobra.Command {
	var f applianceFlags
	name := t.spec.Name
	var setup ApplianceSetup

	root := &cobra.Command{
		Use:   name + " [check|run] [flags]",
		Short: "check the rules once, or on an interval with hot reload",
		Long: name + " evaluates its embedded rules, with an optional override file layered on top (--rules,\n" +
			"--rules-mode). `check` evaluates once; `run --interval DUR` evaluates repeatedly and re-reads the\n" +
			"override before each check. With no subcommand it behaves like `run` when --interval is given and like\n" +
			"`check` otherwise.",
		Args:          noArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return t.serve(cmd, &f, setup, f.interval > 0)
		},
	}
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return usageError{err} })
	pf := root.PersistentFlags()
	pf.StringVar(&f.rules, "rules", t.spec.RulesPath, "override rules `FILE` (a mounted ConfigMap); a missing or empty file means none")
	pf.StringVar(&f.rulesMode, "rules-mode", OverrideLayer.String(), "how the override combines with the embedded rules: `layer` (on top of them) or `replace` (instead of them)")
	pf.StringVar(&f.format, "format", FormatText, "output format: `text` (one result per line) or `json` (one JSON object per check)")
	pf.BoolVar(&f.failOnViolation, "fail-on-violation", false, "exit 3 when the check finds violations (check only; run never exits on them); a failed check still exits 1")
	root.Flags().DurationVar(&f.interval, "interval", 0, "re-check (and hot-reload the rules) every `DUR`; 0 checks once")
	if t.spec.Bind != nil {
		setup = t.spec.Bind(root)
	}

	check := &cobra.Command{
		Use:   "check",
		Short: "evaluate the rules once",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return t.serve(cmd, &f, setup, false)
		},
	}
	run := &cobra.Command{
		Use:   "run --interval DUR",
		Short: "evaluate the rules every DUR, re-reading the override before each check",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if f.interval <= 0 {
				return usagef("run needs --interval DUR greater than zero (use `check` to evaluate once)")
			}
			return t.serve(cmd, &f, setup, true)
		},
	}
	run.Flags().DurationVar(&f.interval, "interval", 0, "wait `DUR` between checks")

	root.AddCommand(check, run, versionCommand(name))
	if t.spec.Embedded != nil {
		rules := &cobra.Command{
			Use:   "rules",
			Short: "print the embedded rules",
			Long:  "`rules cat` prints the embedded rules: the starting point for an override file (layer mode needs only\nyour changes; replace mode needs the whole file).",
		}
		rules.AddCommand(&cobra.Command{
			Use: "cat", Short: "print the embedded rules", Args: noArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				src, err := t.spec.Embedded()
				if err != nil {
					return err
				}
				_, err = io.WriteString(cmd.OutOrStdout(), src)
				return err
			},
		})
		root.AddCommand(rules)
	}
	return root
}

// Execute runs the command line argv (without the program name) and returns the exit status.
// Single-dash spellings of the spec's LegacyFlags are rewritten to their `--long` form, with
// one notice per flag on stderr.
func (t *ApplianceTree) Execute(argv []string, stdin io.Reader, stdout, stderr io.Writer) int {
	args, rewritten := rewriteLegacyFlags(argv, t.spec.LegacyFlags)
	for _, name := range rewritten {
		fmt.Fprintf(stderr, "%s: -%s is deprecated; use --%s\n", t.spec.Name, name, name)
	}
	return executeRoot(t.spec.Name, t.Root(), args, stdin, stdout, stderr)
}

// Exit runs the command line against the process streams and exits on a non-zero status.
func (t *ApplianceTree) Exit() {
	if code := t.Execute(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); code != 0 {
		os.Exit(code)
	}
}

// rewriteLegacyFlags turns `-name` and `-name=value` into `--name` and `--name=value` for the
// listed names, up to a `--` terminator, and returns the names it rewrote (once each, in order
// of appearance). Every other argument — including real short flags — passes through.
func rewriteLegacyFlags(argv []string, names []string) (out []string, rewritten []string) {
	known := make(map[string]bool, len(names))
	for _, n := range names {
		known[n] = true
	}
	seen := map[string]bool{}
	out = make([]string, 0, len(argv))
	for k, a := range argv {
		if a == "--" {
			return append(out, argv[k:]...), rewritten
		}
		if strings.HasPrefix(a, "-") && !strings.HasPrefix(a, "--") && len(a) > 1 {
			name, _, _ := strings.Cut(a[1:], "=")
			if known[name] {
				if !seen[name] {
					seen[name] = true
					rewritten = append(rewritten, name)
				}
				a = "-" + a
			}
		}
		out = append(out, a)
	}
	return out, rewritten
}

// serve loads the Appliance and evaluates its pipeline once, or forever when loop is set.
func (t *ApplianceTree) serve(cmd *cobra.Command, f *applianceFlags, setup ApplianceSetup, loop bool) error {
	if setup == nil {
		return fmt.Errorf("%s has no data source configured", t.spec.Name)
	}
	mode, err := ParseOverrideMode(f.rulesMode)
	if err != nil {
		return UsageError(err)
	}
	if f.interval < 0 {
		return usagef("--interval must not be negative")
	}
	if f.format != FormatText && f.format != FormatJSON {
		return usagef("unknown --format %q: want text or json", f.format)
	}
	if loop && f.failOnViolation {
		return usagef("--fail-on-violation applies to a single check; run keeps checking and never exits on violations")
	}
	stdout, stderr := cmd.OutOrStdout(), cmd.ErrOrStderr()

	app, pipeline, err := setup(ApplianceSettings{Rules: f.rules, Mode: mode})
	if err != nil {
		return err
	}
	last := app.Report().String()
	t.logLines(stderr, app.Report().Lines())

	for {
		rep := classify(app.Execute(pipeline))
		failures := len(rep.errs)
		if f.format == FormatJSON {
			failures = rep.writeJSON(stdout)
		} else {
			rep.writeText(stdout)
		}
		if !loop {
			switch {
			case failures > 0:
				return fmt.Errorf("the check failed (%d error(s), see the output above)", failures)
			case f.failOnViolation && len(rep.findings) > 0:
				return ViolationsError{Count: len(rep.findings)}
			}
			return nil
		}
		if !t.sleep(f.interval) {
			return nil
		}
		if err := app.Reload(); err != nil {
			fmt.Fprintf(stderr, "%s: reload failed, still using the previous rules: %v\n", t.spec.Name, err)
		} else if now := app.Report().String(); now != last {
			last = now
			t.logLines(stderr, app.Report().Lines())
		}
	}
}

func (t *ApplianceTree) logLines(w io.Writer, lines []string) {
	for _, l := range lines {
		fmt.Fprintf(w, "%s: %s\n", t.spec.Name, l)
	}
}
