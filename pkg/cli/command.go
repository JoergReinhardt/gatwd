package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/spf13/cobra"

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
	// NormalizeArgs, when set, rewrites the arguments before parsing; it is the hook for
	// accepting a retired spelling (for example a single-dash long flag) for a release.
	NormalizeArgs func([]string) []string
}

// CommandTree is a CommandSpec turned into a runnable command line.
type CommandTree struct {
	spec CommandSpec
}

// NewCommandTree builds the command line for spec.
func NewCommandTree(spec CommandSpec) *CommandTree {
	if spec.Name == "" {
		spec.Name = "gatwd"
	}
	return &CommandTree{spec: spec}
}

// usageError marks a malformed command line (exit status 2).
type usageError struct{ err error }

func (u usageError) Error() string { return u.err.Error() }
func (u usageError) Unwrap() error { return u.err }

// silentExit ends the process with a status after the command already reported why.
type silentExit struct{ code int }

func (s silentExit) Error() string { return fmt.Sprintf("exit status %d", s.code) }

func usagef(format string, a ...any) error { return usageError{fmt.Errorf(format, a...)} }

// Execute runs the command line argv (without the program name) and returns the exit
// status: 0 on success, 1 when a command failed, 2 when the command line is malformed.
func (t *CommandTree) Execute(argv []string, stdin io.Reader, stdout, stderr io.Writer) int {
	root := t.Root()
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)
	args := append([]string{}, argv...)
	if t.spec.NormalizeArgs != nil {
		args = t.spec.NormalizeArgs(args)
		if args == nil {
			args = []string{}
		}
	}
	root.SetArgs(args)
	err := root.Execute()
	var silent silentExit
	var usage usageError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &silent):
		return silent.code
	case errors.As(err, &usage):
		fmt.Fprintf(stderr, "%s: %v\nRun '%s --help' for usage.\n", t.spec.Name, err, t.spec.Name)
		return 2
	default:
		fmt.Fprintf(stderr, "%s: %v\n", t.spec.Name, err)
		return 1
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
	pf.StringArrayVarP(&g.defines, "define", "D", nil, "bind the string constant `NAME=VALUE` (repeatable)")

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

	version := &cobra.Command{
		Use: "version", Short: "print the version", Args: noArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			v := "(unknown)"
			if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" {
				v = bi.Main.Version
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", name, v)
		},
	}

	root.AddCommand(run, eval, check, task, repl, preludeCmd, builtins, version)
	root.AddCommand(t.spec.Commands...)
	return root
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
	res, runErr := in.Run(opts)
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
		return silentExit{1}
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
