package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/joergreinhardt/gatwd/functions"
)

// This file is the command line of every Edition and Appliance: a cobra tree that maps flags
// to Interpreter.Run and nothing more. What a script sees (ARGS, ARG0, tasks, `main`) is
// defined in run.go; the tree owns only the syntax.
//
//	NAME [flags] [script] [-- arg ...]    script runner / REPL (backward compatible)
//	NAME run SCRIPT [-- arg ...]          explicit form
//	NAME eval -e 'statement' [-e …]       one-liners
//	NAME check FILE...                    structural errors + advisory typing findings
//	NAME task TASK [-s SCRIPT] [-- arg …] run a def from the prelude or a script
//	NAME repl                             interactive
//	NAME prelude list | cat NAME          the embedded preludes
//	NAME builtins                         function forms, defs and constants
//	NAME version · NAME completion SHELL
//
// A script literally named like a subcommand is run as `NAME ./check`.

// Prelude names one embedded prelude for `prelude list` and `prelude cat`.
type Prelude struct {
	Name   string
	Source func() (string, error)
}

// CommandSpec describes the command line of one binary.
type CommandSpec struct {
	// Name is the program name shown in help and error messages.
	Name string
	// Build returns a fresh interpreter. noPrelude asks it to skip the embedded preludes;
	// a Build with no preludes ignores it.
	Build func(noPrelude bool) *Interpreter
	// Preludes are the embedded preludes, listed by `prelude`.
	Preludes []Prelude
	// Commands are domain subcommands added to the root (Edition.WithCommand).
	Commands []*cobra.Command
	// Flags declare extra persistent flags on the root (Edition.WithFlags); each is called
	// with the root's flag set every time the tree is built.
	Flags []func(*pflag.FlagSet)
	// NormalizeArgs, when set, rewrites the arguments before parsing; it is the hook for
	// accepting a retired spelling (for example a single-dash long flag) for a release.
	NormalizeArgs func([]string) []string
}

// CommandTree is a CommandSpec turned into a runnable command line.
type CommandTree struct {
	spec CommandSpec
	// signals and exit are hooks for tests; NewCommandTree installs the real ones.
	signals func() (<-chan os.Signal, func())
	exit    func(int)
}

// NewCommandTree builds the command line for spec.
func NewCommandTree(spec CommandSpec) *CommandTree {
	if spec.Name == "" {
		spec.Name = "gatwd"
	}
	return &CommandTree{spec: spec, signals: notifyStopSignals, exit: os.Exit}
}

// The exit statuses of every command line in this package (Edition trees and Appliance trees).
// They are documented in docs/dsl_guide.md §23 and pinned by tests; callers (CI, a scheduler)
// can rely on them to tell "the check found problems" from "the check could not run".
const (
	// ExitOK means the command ran and, where a check was asked to fail on findings, found none.
	ExitOK = 0
	// ExitFailed means the command broke: a script or check that errored or yielded nothing,
	// rules that do not load, a static check that found an error.
	ExitFailed = 1
	// ExitUsage means a malformed command line (unknown flag, missing argument, no source).
	ExitUsage = 2
	// ExitViolations means the check ran correctly and found violations, and the caller asked
	// for that to fail the command (--fail-on-violation). It is distinct from ExitFailed so a
	// CI job can tell a policy finding from a broken run (the diff/grep convention).
	ExitViolations = 3
)

// ViolationsError reports that a check ran correctly and found violations. A command returns
// it (or wraps it) when the caller asked findings to fail the command; the tree reports the
// message on stderr and exits with ExitViolations instead of ExitFailed.
type ViolationsError struct {
	// Count is the number of violations found.
	Count int
}

func (v ViolationsError) Error() string { return fmt.Sprintf("%d violation(s) found", v.Count) }

// usageError marks a malformed command line (exit status 2).
type usageError struct{ err error }

func (u usageError) Error() string { return u.err.Error() }
func (u usageError) Unwrap() error { return u.err }

// silentExit ends the process with a status after the command already reported why.
type silentExit struct{ code int }

func (s silentExit) Error() string { return fmt.Sprintf("exit status %d", s.code) }

func usagef(format string, a ...any) error { return usageError{fmt.Errorf(format, a...)} }

// UsageError marks err as a malformed command line: the tree reports it with a pointer to
// `--help` and exits with status 2. A domain command (Edition.WithCommand, an Appliance's
// setup) uses it for "you invoked me wrongly", as opposed to "I ran and failed" (status 1).
func UsageError(err error) error { return usageError{err} }

// Execute runs the command line argv (without the program name) and returns the exit
// status: ExitOK (0) on success, ExitFailed (1) when a command failed, ExitUsage (2) when the
// command line is malformed. The Edition tree never returns ExitViolations; a domain command
// that returns ViolationsError does.
func (t *CommandTree) Execute(argv []string, stdin io.Reader, stdout, stderr io.Writer) int {
	args := append([]string{}, argv...)
	if t.spec.NormalizeArgs != nil {
		args = t.spec.NormalizeArgs(args)
		if args == nil {
			args = []string{}
		}
	}
	return executeRoot(t.spec.Name, t.Root(), args, stdin, stdout, stderr)
}

// executeRoot runs a cobra root with the streams and maps its error to the exit status shared
// by every command line in this package: ExitOK, ExitFailed (a command failed), ExitUsage (a
// malformed command line) and ExitViolations (a ViolationsError).
func executeRoot(name string, root *cobra.Command, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(args)
	err := root.Execute()
	var silent silentExit
	var usage usageError
	var violations ViolationsError
	switch {
	case err == nil:
		return ExitOK
	case errors.As(err, &silent):
		return silent.code
	case errors.As(err, &usage):
		fmt.Fprintf(stderr, "%s: %v\nRun '%s --help' for usage.\n", name, err, name)
		return ExitUsage
	case errors.As(err, &violations):
		fmt.Fprintf(stderr, "%s: %v\n", name, err)
		return ExitViolations
	default:
		fmt.Fprintf(stderr, "%s: %v\n", name, err)
		return ExitFailed
	}
}

// globalFlags are the persistent flags every command shares.
type globalFlags struct {
	preludes  []string
	noPrelude bool
	defines   []string
}

// runFlags are the flags of the script-running commands.
type runFlags struct {
	interactive bool
	task        string
	eval        []string
	check       bool
}

// Root returns a fresh cobra root command for the spec. Flag values live in the returned
// tree, so each call starts from the defaults.
func (t *CommandTree) Root() *cobra.Command {
	var g globalFlags
	var rf runFlags
	name := t.spec.Name

	root := &cobra.Command{
		Use:   name + " [flags] [script] [-- arg ...]",
		Short: "run a gatwd script, evaluate statements, or start the REPL",
		Long: name + " runs a gatwd script, a task from it, one-liners (-e), or an interactive REPL.\n\n" +
			"Arguments after -- are passed to the script as ARGS (a List of String); ARG0 is the script path.\n" +
			"With neither a script, a task nor -e, the REPL starts.",
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			before, after := splitAtDash(cmd, args)
			if rf.check {
				return checkFiles(cmd, args)
			}
			if len(before) > 1 {
				return usagef("unexpected argument %q (script arguments go after --)", before[1])
			}
			script := ""
			if len(before) == 1 {
				script = before[0]
			}
			return t.runScript(cmd, &g, &rf, script, after)
		},
	}
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return usageError{err} })
	pf := root.PersistentFlags()
	pf.StringArrayVar(&g.preludes, "prelude", nil, "load `FILE` before the script (repeatable)")
	pf.BoolVar(&g.noPrelude, "no-prelude", false, "skip the embedded preludes")
	pf.StringArrayVarP(&g.defines, "define", "D", nil, "bind the constant `NAME=VALUE`: a number, Duration (5s, 30d), Time or Date (@2026-10-13) literal is typed, anything else a String (repeatable)")
	for _, declare := range t.spec.Flags {
		declare(pf)
	}

	addRunFlags(root, &rf)
	root.Flags().BoolVarP(&rf.check, "check", "c", false, "only check the given files (see `check`)")

	run := &cobra.Command{
		Use:   "run SCRIPT [-- arg ...]",
		Short: "run a script",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			before, after := splitAtDash(cmd, args)
			switch {
			case len(before) == 0:
				return usagef("run needs a script: %s run SCRIPT [-- arg ...]", name)
			case len(before) > 1:
				return usagef("unexpected argument %q (script arguments go after --)", before[1])
			}
			return t.runScript(cmd, &g, &rf, before[0], after)
		},
	}
	addRunFlags(run, &rf)

	var evalStmts []string
	eval := &cobra.Command{
		Use:   "eval -e STATEMENT [-e STATEMENT ...]",
		Short: "evaluate statements and print their results",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if len(evalStmts) == 0 {
				return usagef("eval needs at least one -e STATEMENT")
			}
			return t.run(cmd, &g, RunOptions{Eval: evalStmts}, false)
		},
	}
	eval.Flags().StringArrayVarP(&evalStmts, "eval", "e", nil, "evaluate `STATEMENT` (repeatable)")

	var taskScript string
	task := &cobra.Command{
		Use:   "task TASK [-s SCRIPT] [-- arg ...]",
		Short: "run a def from the prelude or a script, with the arguments after --",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			before, after := splitAtDash(cmd, args)
			switch {
			case len(before) == 0:
				return usagef("task needs a name: %s task TASK [-s SCRIPT] [-- arg ...]", name)
			case len(before) > 1:
				return usagef("unexpected argument %q (task arguments go after --)", before[1])
			}
			return t.run(cmd, &g, RunOptions{Script: taskScript, Task: before[0], Args: after}, false)
		},
	}
	task.Flags().StringVarP(&taskScript, "script", "s", "", "load `SCRIPT` before running the task")

	check := &cobra.Command{
		Use:   "check FILE...",
		Short: "check scripts statically (exit 1 on structural errors; typing findings are advisory)",
		Args:  cobra.ArbitraryArgs,
		RunE:  func(cmd *cobra.Command, args []string) error { return checkFiles(cmd, args) },
	}

	repl := &cobra.Command{
		Use:   "repl",
		Short: "start the interactive REPL",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return t.runScript(cmd, &g, &runFlags{interactive: true}, "", nil)
		},
	}

	preludeCmd := &cobra.Command{Use: "prelude", Short: "list or print the embedded preludes"}
	preludeCmd.AddCommand(
		&cobra.Command{
			Use: "list", Short: "list the embedded preludes", Args: noArgs,
			Run: func(cmd *cobra.Command, _ []string) {
				for _, p := range t.spec.Preludes {
					fmt.Fprintln(cmd.OutOrStdout(), p.Name)
				}
			},
		},
		&cobra.Command{
			Use: "cat NAME", Short: "print an embedded prelude", Args: exactArgs(1, "cat needs the name of a prelude (see `prelude list`)"),
			RunE: func(cmd *cobra.Command, args []string) error {
				for _, p := range t.spec.Preludes {
					if p.Name == args[0] {
						src, err := p.Source()
						if err != nil {
							return fmt.Errorf("prelude %q: %w", p.Name, err)
						}
						_, err = io.WriteString(cmd.OutOrStdout(), src)
						return err
					}
				}
				return fmt.Errorf("no prelude %q (see `prelude list`)", args[0])
			},
		},
	)

	builtins := &cobra.Command{
		Use:   "builtins",
		Short: "list the function forms, defs and constants of this edition",
		Long: "Lists every curried function form with its parameters, then the defs and constants the\n" +
			"edition's preludes define. Pipeline step names are matched by prefix and cannot be listed.",
		Args: noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			printBuiltins(cmd.OutOrStdout(), t.interp(cmd, &g))
			return nil
		},
	}

	version := versionCommand(name)

	root.AddCommand(run, eval, check, task, repl, preludeCmd, builtins, version)
	root.AddCommand(t.spec.Commands...)
	return root
}

// versionCommand is the `version` subcommand every command line has: the program name and the
// module version the binary was built from.
func versionCommand(name string) *cobra.Command {
	return &cobra.Command{
		Use: "version", Short: "print the version", Args: noArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			v := "(unknown)"
			if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" {
				v = bi.Main.Version
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", name, v)
		},
	}
}

func addRunFlags(cmd *cobra.Command, rf *runFlags) {
	f := cmd.Flags()
	f.BoolVarP(&rf.interactive, "interactive", "i", false, "enter the REPL after the script ran")
	f.StringVar(&rf.task, "task", "", "run the def `NAME` instead of main")
	f.StringArrayVarP(&rf.eval, "eval", "e", nil, "evaluate `STATEMENT` after the script (repeatable; replaces main)")
}

// splitAtDash separates the positional arguments before `--` from those after it.
func splitAtDash(cmd *cobra.Command, args []string) (before, after []string) {
	d := cmd.ArgsLenAtDash()
	if d < 0 {
		return args, nil
	}
	return args[:d], args[d:]
}

func noArgs(_ *cobra.Command, args []string) error {
	if len(args) > 0 {
		return usagef("unexpected argument %q", args[0])
	}
	return nil
}

func exactArgs(n int, msg string) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if len(args) != n {
			return usagef("%s", msg)
		}
		return nil
	}
}

// interp builds an interpreter honouring the global flags, printing to the command's output.
func (t *CommandTree) interp(cmd *cobra.Command, g *globalFlags) *Interpreter {
	in := t.spec.Build(g.noPrelude)
	in.Out = cmd.OutOrStdout()
	return in
}

// runScript is the script runner behind the root command, `run` and `repl`.
func (t *CommandTree) runScript(cmd *cobra.Command, g *globalFlags, rf *runFlags, script string, args []string) error {
	out := cmd.OutOrStdout()
	interactive := rf.interactive || (script == "" && rf.task == "" && len(rf.eval) == 0)
	fmt.Fprintln(out, "=== gatwd Interpreter (V2 Modernized) ===")
	fmt.Fprintln(out, "UTF-8 Mathematical Aliases Enabled (e.g. ⊳, →, ∷)")
	opts := RunOptions{Script: script, Args: args, Task: rf.task, Eval: rf.eval}
	if script != "" && !interactive {
		fmt.Fprintf(out, "Running Batch Mode: %s\n", script)
	}
	return t.run(cmd, g, opts, interactive)
}

func (t *CommandTree) run(cmd *cobra.Command, g *globalFlags, opts RunOptions, interactive bool) error {
	defs, err := ParseDefines(g.defines)
	if err != nil {
		return usageError{err}
	}
	opts.Define = defs
	opts.Preludes = g.preludes
	in := t.interp(cmd, g)
	out := cmd.OutOrStdout()
	if !interactive && t.signals != nil {
		// SIGINT/SIGTERM cancel the script; a second signal ends the process at once.
		ctx, release := stopOnSignal(t.spec.Name, t.signals, t.exit, context.Background(), cmd.ErrOrStderr())
		defer release()
		opts.Context = ctx
	}
	res, runErr := in.Run(opts)
	var stop stopRequest
	if errors.As(runErr, &stop) {
		for _, line := range res.Lines {
			fmt.Fprintf(out, "∷ %s\n", line)
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "%s: %v, stopped\n", t.spec.Name, stop)
		return silentExit{exitForSignal(stop.sig)}
	}
	for _, line := range res.Lines {
		fmt.Fprintf(out, "∷ %s\n", line)
	}
	if runErr == nil && res.Task != "" && len(res.Lines) == 0 {
		fmt.Fprintln(out, "(No output or node unhandled)")
	}
	if runErr != nil {
		if !interactive {
			return runErr
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "%s: %v\n", t.spec.Name, runErr)
	}
	if interactive {
		in.REPL(cmd.InOrStdin())
	}
	return nil
}

// checkFiles runs the static checker over each file. Diagnostics go to stderr as
// `file:line:col: severity: message`; any error-severity diagnostic or unreadable file makes
// the command fail with status 1, warnings alone do not.
func checkFiles(cmd *cobra.Command, files []string) error {
	if len(files) == 0 {
		return usagef("check needs at least one file")
	}
	stderr := cmd.ErrOrStderr()
	failed := false
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(stderr, "%s:1:1: error: cannot read file: %v\n", file, err)
			failed = true
			continue
		}
		for _, d := range CheckScript(string(content)) {
			fmt.Fprintf(stderr, "%s:%d:%d: %s: %s\n", file, d.Line, d.Col, d.Severity, d.Msg)
			if d.Severity == "error" {
				failed = true
			}
		}
	}
	if failed {
		return silentExit{ExitFailed}
	}
	return nil
}

func printBuiltins(w io.Writer, in *Interpreter) {
	names := make([]string, 0, len(dslBuiltinFuncs))
	for n := range dslBuiltinFuncs {
		names = append(names, n)
	}
	sort.Strings(names)
	fmt.Fprintln(w, "function forms:")
	for _, n := range names {
		params := ""
		if c, ok := dslBuiltinFuncs[n].(*functions.Closure); ok {
			params = " " + strings.Join(c.Params, " ")
		}
		fmt.Fprintf(w, "  %s%s\n", n, strings.TrimRight(params, " "))
	}
	fmt.Fprintln(w, "defs:")
	for _, n := range sortedKeys(in.Funcs) {
		fmt.Fprintf(w, "  %s\n", n)
	}
	consts := make([]string, 0, len(in.Constants))
	for n := range in.Constants {
		consts = append(consts, n)
	}
	sort.Strings(consts)
	fmt.Fprintln(w, "constants:")
	for _, n := range consts {
		fmt.Fprintf(w, "  %s\n", n)
	}
}

// exitWith runs the tree against the process streams and exits on a non-zero status.
func (t *CommandTree) exitWith() {
	if code := t.Execute(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); code != 0 {
		os.Exit(code)
	}
}
