package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
// --format text|json, --fail-on-violation, and the execution policy --allow-exec, --allow-fs,
// --allow-net, --allow-env (all off: the rules, an override above all, may only compute), and the
// per-check limits --timeout DUR and --max-output BYTES (both off by default; a check that goes
// over one fails, and each iteration of `run` starts with fresh limits). A domain adds its own
// (its data source) in Bind.
//
// Exit status, the same mapping as the Edition command tree plus one: 0 the check ran and,
// with --fail-on-violation, found nothing; 1 the command failed — rules that do not load, a
// check that errored, yielded nothing or went over its --timeout/--max-output; 2 a malformed command line; 3 (ExitViolations) the
// check ran correctly and, under --fail-on-violation, found violations. An error always wins
// over violations: a check that both errored and found violations exits 1. `run` is a daemon: a failing check is reported and the loop continues, and
// only a start-up failure or a stop request ends it.
//
// Stopping. SIGINT and SIGTERM (what Kubernetes sends a pod, to PID 1, before the grace period
// ends) are a request to stop, handled for `check` and `run` alike: the context of the check in
// progress is cancelled (every pipeline step stops at its next node and an OS command a step
// started is killed), so the process ends promptly rather than after the interval or the check.
// For `run` that is a normal shutdown of a daemon: one line on stderr ("NAME: received
// terminated, stopping") and exit status 0; the interrupted check prints nothing, and a check
// that had already failed before the signal was reported when it failed. For `check` the check
// did not finish, so it exits 1 ("the check was interrupted"). A second signal while the
// process is still stopping exits at once with status 1: cancellation is cooperative, and one
// evaluation that never reaches a node boundary cannot be interrupted any other way.
//
// Loading. --timeout also bounds LOADING the rules (the top-level statements of the embedded
// rules and of the override are evaluated while loading): one budget per load, at start-up and
// again at every reload of `run`. A start-up load that runs over fails the command with status
// 1 ("PATH: 1 problem(s): loading the rules timed out after 30s"); a hot reload that runs over
// is reported on stderr like any broken override and the last good rules keep serving. A stop
// request cancels a load in progress the same way it cancels a check.

// ApplianceSettings is what the generic flags resolved to; a domain's ApplianceSetup turns it
// into an Appliance.
type ApplianceSettings struct {
	// Rules is the override file path (--rules).
	Rules string
	// Mode is how the override combines with the embedded rules (--rules-mode).
	Mode OverrideMode
	// Policy is what the rules may do beyond computing (--allow-exec, --allow-fs, --allow-net,
	// --allow-env); every capability is denied unless its flag is given.
	Policy Policy
	// Timeout is how long one check may take (--timeout); zero means no limit.
	Timeout time.Duration
	// MaxOutput is how many bytes one check may produce (--max-output); zero means no limit.
	MaxOutput int64
	// Context bounds loading the rules (the stop request of the command, see below); nil means
	// no bound beyond Timeout. The command fills it in; a setup passes it on with BoundOptions.
	Context context.Context
}

// Options returns the ApplianceOptions that apply the settings: WithConfigMapOverride,
// WithOverrideMode, WithPolicy, WithTimeout and WithMaxOutput. A domain setup passes them to
// NewAppliance next to its own edition and embedded rules.
func (s ApplianceSettings) Options() []ApplianceOption {
	return []ApplianceOption{WithConfigMapOverride(s.Rules), WithOverrideMode(s.Mode), WithPolicy(s.Policy), s.BoundOptions()}
}

// BoundOptions returns the options for the limits alone (WithTimeout, which also bounds
// loading and reloading the rules, WithMaxOutput and WithLoadContext), for a domain setup that
// assembles the other options itself.
func (s ApplianceSettings) BoundOptions() ApplianceOption {
	return func(a *Appliance) {
		WithTimeout(s.Timeout)(a)
		WithMaxOutput(s.MaxOutput)(a)
		WithLoadContext(s.Context)(a)
	}
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
	spec       ApplianceSpec
	watchEvery time.Duration            // how often wait polls the override file; zero means one second
	sleep      func(time.Duration) bool // waits between checks; false stops the loop; nil: wait on a timer or a stop request (replaced in tests)
	// signals starts listening for stop requests and returns their channel and a function that
	// stops listening; tests inject a channel instead of sending the process a real signal.
	signals func() (<-chan os.Signal, func())
	exit    func(code int) // ends the process on a second stop request (replaced in tests)
}

// NewApplianceTree builds the command line for spec.
func NewApplianceTree(spec ApplianceSpec) *ApplianceTree {
	if spec.Name == "" {
		spec.Name = "appliance"
	}
	return &ApplianceTree{spec: spec, signals: notifyStopSignals, exit: os.Exit}
}

// notifyStopSignals listens for SIGINT and SIGTERM. Installing a handler is also what lets a
// process that runs as PID 1 receive SIGTERM at all: the kernel drops signals sent to PID 1
// that have no handler.
func notifyStopSignals() (<-chan os.Signal, func()) {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	return ch, func() { signal.Stop(ch) }
}

// stopRequest is the cause of a context cancelled by SIGINT or SIGTERM.
type stopRequest struct{ sig os.Signal }

func (s stopRequest) Error() string { return "received " + s.sig.String() }

// stopContext returns a context that a stop request cancels (with a stopRequest cause) and the
// function that releases it. After the first request the goroutine keeps listening: a second
// one means the graceful stop is stuck, and ends the process at once.
func (t *ApplianceTree) stopContext(parent context.Context, stderr io.Writer) (context.Context, func()) {
	return stopOnSignal(t.spec.Name, t.signals, t.exit, parent, stderr)
}

// stopOnSignal is the shared stop-request machinery of every command line: signals supplies
// the stop signals, exit ends the process when a second one arrives.
func stopOnSignal(name string, signals func() (<-chan os.Signal, func()), exit func(int), parent context.Context, stderr io.Writer) (context.Context, func()) {
	ch, stopListening := signals()
	ctx, cancel := context.WithCancelCause(parent)
	done := make(chan struct{})
	go func() {
		select {
		case sig := <-ch:
			cancel(stopRequest{sig})
		case <-done:
			return
		}
		select {
		case sig := <-ch:
			fmt.Fprintf(stderr, "%s: received %v again, exiting without waiting for the check\n", name, sig)
			exit(ExitFailed)
		case <-done:
		}
	}()
	return ctx, func() {
		close(done)
		stopListening()
		cancel(nil)
	}
}

// exitForSignal is the Unix convention for a process ended by a signal: 128 plus its number.
func exitForSignal(sig os.Signal) int {
	if s, ok := sig.(syscall.Signal); ok {
		return 128 + int(s)
	}
	return ExitFailed
}

// wait pauses for d between two checks, or until the override file at path changes (polled
// once per watchEvery), and reports whether the loop goes on: false when the context is done first
// (a stop request) or, in a test, when the sleep hook says so.
func (t *ApplianceTree) wait(ctx context.Context, d time.Duration, path string) bool {
	if t.sleep != nil {
		return t.sleep(d)
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	poll := t.watchEvery
	if poll <= 0 {
		poll = time.Second
	}
	tick := time.NewTicker(poll)
	defer tick.Stop()
	before := fileStamp(path)
	for {
		select {
		case <-timer.C:
			return true
		case <-tick.C:
			// a mounted ConfigMap swaps its file in place: re-check at once rather than
			// after the rest of the interval
			if path != "" && fileStamp(path) != before {
				return true
			}
		case <-ctx.Done():
			return false
		}
	}
}

// fileStamp identifies the current content of path by modification time and size (a missing
// file has its own stamp), so a change is noticed by polling with the standard library only.
func fileStamp(path string) string {
	if path == "" {
		return ""
	}
	fi, err := os.Stat(path)
	if err != nil {
		return "missing"
	}
	return fmt.Sprintf("%d/%d", fi.ModTime().UnixNano(), fi.Size())
}

// stopped ends the command after a stop request. `run` is a daemon and a stop request is its
// normal end (nil, status 0), announced on one line; `check` did not finish what it was asked
// to do, so it fails.
func (t *ApplianceTree) stopped(ctx context.Context, stderr io.Writer, loop, inFlight bool) error {
	cause := cancelCause(ctx)
	if !loop {
		return fmt.Errorf("the check was interrupted (%v)", cause)
	}
	note := "stopping"
	if inFlight {
		note = "stopping, the check in progress was cancelled"
	}
	fmt.Fprintf(stderr, "%s: %v, %s\n", t.spec.Name, cause, note)
	return nil
}

// applianceFlags are the generic flags shared by the commands.
type applianceFlags struct {
	rules           string
	rulesMode       string
	interval        time.Duration
	format          string
	failOnViolation bool
	allow           Policy
	timeout         time.Duration
	maxOutput       int64
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
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error { return applyEnvDefaults(cmd) }
	pf := root.PersistentFlags()
	pf.StringVar(&f.rules, "rules", t.spec.RulesPath, "override rules `FILE` (a mounted ConfigMap); a missing or empty file means none")
	pf.StringVar(&f.rulesMode, "rules-mode", OverrideLayer.String(), "how the override combines with the embedded rules: `layer` (on top of them) or `replace` (instead of them)")
	pf.StringVar(&f.format, "format", FormatText, "output format: `text` (one result per line) or `json` (one JSON object per check)")
	pf.BoolVar(&f.failOnViolation, "fail-on-violation", false, "exit 3 when the check finds violations (check only; run never exits on them); a failed check still exits 1")
	pf.BoolVar(&f.allow.AllowExec, "allow-exec", false, "let the rules run OS commands (an unknown step is a program); denied by default")
	pf.BoolVar(&f.allow.AllowFS, "allow-fs", false, "let the rules read and write files and `import` other files; denied by default")
	pf.BoolVar(&f.allow.AllowNet, "allow-net", false, "let the rules open network connections; denied by default")
	pf.BoolVar(&f.allow.AllowEnv, "allow-env", false, "let the rules read the process environment; denied by default")
	pf.BoolVar(&f.allow.AllowClusterApply, "allow-cluster-apply", false, "let the rules create and update objects in a cluster (a write step such as k8s_apply); denied by default")
	pf.BoolVar(&f.allow.AllowClusterRollout, "allow-cluster-rollout", false, "let the rules restart, roll back, pause, resume and scale workloads in a cluster; denied by default")
	pf.DurationVar(&f.timeout, "timeout", 0, "stop a check that takes longer than `DUR` (a failed check, exit 1; each check of `run` gets its own); 0 means no limit")
	pf.Int64Var(&f.maxOutput, "max-output", 0, "fail a check that produces more than `BYTES` of output (never truncated silently; each check of `run` gets its own); 0 means no limit")
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

// EnvPrefix is the prefix of the environment variables that default an Appliance's flags:
// --allow-exec is GATWD_ALLOW_EXEC, --rules-mode is GATWD_RULES_MODE.
const EnvPrefix = "GATWD_"

// envName is the environment variable that defaults the flag name.
func envName(name string) string {
	return EnvPrefix + strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
}

// applyEnvDefaults gives every flag of cmd that was not set on the command line the value of
// its environment variable, if there is one: flag > environment > default. A value that does
// not parse is a usage error naming the variable; an empty variable counts as unset.
func applyEnvDefaults(cmd *cobra.Command) error {
	var err error
	cmd.Flags().VisitAll(func(fl *pflag.Flag) {
		if err != nil || fl.Changed || fl.Name == "help" {
			return
		}
		v, ok := os.LookupEnv(envName(fl.Name))
		if !ok || v == "" {
			return
		}
		if serr := fl.Value.Set(v); serr != nil {
			err = usagef("%s=%q: %v", envName(fl.Name), v, serr)
		}
	})
	return err
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
	if f.timeout < 0 {
		return usagef("--timeout must not be negative")
	}
	if f.maxOutput < 0 {
		return usagef("--max-output must not be negative")
	}
	if f.format != FormatText && f.format != FormatJSON {
		return usagef("unknown --format %q: want text or json", f.format)
	}
	if loop && f.failOnViolation {
		return usagef("--fail-on-violation applies to a single check; run keeps checking and never exits on violations")
	}
	stdout, stderr := cmd.OutOrStdout(), cmd.ErrOrStderr()
	parent := cmd.Context()
	if parent == nil {
		parent = context.Background()
	}
	ctx, release := t.stopContext(parent, stderr)
	defer release()

	app, pipeline, err := setup(ApplianceSettings{Rules: f.rules, Mode: mode, Policy: f.allow, Timeout: f.timeout, MaxOutput: f.maxOutput, Context: ctx})
	if err != nil {
		if ctx.Err() != nil { // a stop request during start-up: not a failure of the setup
			return t.stopped(ctx, stderr, loop, false)
		}
		return err
	}
	last := app.Report().String()
	t.logLines(stderr, app.Report().Lines())

	for {
		values := app.ExecuteContext(ctx, pipeline)
		if ctx.Err() != nil {
			// a stop request: the values are the cancelled check's, not a result
			return t.stopped(ctx, stderr, loop, true)
		}
		rep := classify(values)
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
		if !t.wait(ctx, f.interval, f.rules) {
			if ctx.Err() != nil {
				return t.stopped(ctx, stderr, loop, false)
			}
			return nil
		}
		if err := app.ReloadContext(ctx); err != nil {
			if ctx.Err() != nil {
				return t.stopped(ctx, stderr, loop, false)
			}
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
