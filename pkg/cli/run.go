package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// This file is the argument model of a script (hardening decision D4). The command-line
// syntax (flags, subcommands, `--`) belongs to the command tree; what a script SEES is
// defined here:
//
//   - ARGS  — a List of String: everything the caller passed after `--`;
//   - ARG0  — the script path ("" when there is none);
//   - tasks — any `def`, run with ARGS. A lambda-bodied def is a function called with the
//     ARGS list; a pipeline-bodied def is a fragment whose input stream is the arguments,
//     one String node each. `main` is the default task, run after the script loads
//     when no other task is selected.

// RunOptions describes one invocation of a script.
type RunOptions struct {
	// Script is the path of the script to load; empty means none (REPL or edition prelude only).
	Script string
	// Args is exposed to the DSL as ARGS.
	Args []string
	// Task names the def to run instead of `main`.
	Task string
}

// RunResult is what the selected task produced.
type RunResult struct {
	// Task is the name of the task that ran; empty when the script defines no `main` and
	// none was selected.
	Task string
	// Lines is the printed form of the task's results; an `[ERROR] …` line is a failure.
	Lines []string
}

// setArgs binds the constants ARGS and ARG0.
func (i *Interpreter) setArgs(script string, args []string) {
	items := make([]any, len(args))
	for k, a := range args {
		items[k] = a
	}
	i.Constants["ARGS"] = listOf(items)
	i.Constants["ARG0"] = script
	i.invalidateCompiledCache()
}

// Run binds ARGS and ARG0, loads o.Script (its top-level statements run and print as
// usual) and then runs the selected task: o.Task, or `main` if the script defines it. An
// unreadable script, an unknown task, or a task that reports an `[ERROR]` is an error;
// the task's lines are returned either way so the caller can print them first.
func (i *Interpreter) Run(o RunOptions) (RunResult, error) {
	i.setArgs(o.Script, o.Args)
	if o.Script != "" {
		content, err := os.ReadFile(o.Script)
		if err != nil {
			return RunResult{}, fmt.Errorf("cannot read script %q: %w", o.Script, err)
		}
		i.LoadScriptContent(string(content), filepath.Dir(o.Script))
	}
	name := o.Task
	if name == "" {
		if _, ok := i.Funcs["main"]; !ok {
			return RunResult{}, nil
		}
		name = "main"
	}
	return i.RunTask(name)
}

// RunTask runs the def called name with the current ARGS (see the file comment for the
// two kinds of task).
func (i *Interpreter) RunTask(name string) (RunResult, error) {
	res := RunResult{Task: name}
	if _, ok := i.Funcs[name]; !ok {
		return res, fmt.Errorf("unknown task %q (defined: %s)", name, strings.Join(i.taskNames(), ", "))
	}

	if c, ok := i.defClosure(name); ok {
		if c.Arity() != 1 {
			return res, fmt.Errorf("task %q takes %d arguments; a task function takes one, the ARGS list", name, c.Arity())
		}
		out, err := c.Apply([]any{i.Constants["ARGS"]})
		if err != nil {
			return res, fmt.Errorf("task %q: %w", name, err)
		}
		res.Lines = []string{renderValue(functions.Right[error, any](out))}
		return res, nil
	}

	flag := i.Parser.GetFlag("StrVal")
	args, _ := functions.ToSliceAny(i.Constants["ARGS"])
	input := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		for _, a := range args {
			if !yield(dsltypes.NewASTNode(functions.Right[error, any](a), flag)) {
				return
			}
		}
	}
	res.Lines = renderNodes(i.Engine.EvaluatePipeline(input, i.parsePipelineSteps([]string{name})))
	failed := 0
	for _, l := range res.Lines {
		if strings.HasPrefix(l, "[ERROR]") {
			failed++
		}
	}
	if failed > 0 {
		return res, fmt.Errorf("task %q reported %d error(s)", name, failed)
	}
	return res, nil
}

// taskNames lists the defined defs, sorted, for error messages.
func (i *Interpreter) taskNames() []string {
	names := make([]string, 0, len(i.Funcs))
	for n := range i.Funcs {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// runArgs is the parsed command line of RunWithInterpreter.
type runArgs struct {
	script      string
	task        string
	interactive bool
	args        []string
}

// parseRunArgs reads `[-i] [--task NAME] [script] [-- arg ...]`. It rejects what it does
// not understand — an unknown flag, a second positional argument — instead of treating
// it as a file name.
func parseRunArgs(argv []string) (runArgs, error) {
	var ra runArgs
	for idx := 0; idx < len(argv); idx++ {
		a := argv[idx]
		switch {
		case a == "--":
			ra.args = append([]string(nil), argv[idx+1:]...)
			return ra, nil
		case a == "-i":
			ra.interactive = true
		case a == "--task":
			if idx+1 >= len(argv) {
				return ra, fmt.Errorf("--task needs a task name")
			}
			idx++
			ra.task = argv[idx]
		case strings.HasPrefix(a, "--task="):
			ra.task = strings.TrimPrefix(a, "--task=")
		case strings.HasPrefix(a, "-") && len(a) > 1:
			return ra, fmt.Errorf("unknown flag %q (script arguments go after --)", a)
		case ra.script == "":
			ra.script = a
		default:
			return ra, fmt.Errorf("unexpected argument %q (script arguments go after --)", a)
		}
	}
	return ra, nil
}
