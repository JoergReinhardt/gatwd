package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync/atomic"

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
	// Eval statements run, in order, after the script loads. When Eval is given and Task is
	// empty, `main` does not run: the statements replace the default task. A statement that
	// prints an `[ERROR]` or nothing at all makes Run fail.
	Eval []string
	// Define binds string constants (`-D NAME=VALUE`) before anything else runs.
	Define map[string]string
	// Preludes are extra scripts loaded, in order, before Script.
	Preludes []string
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

// defineName is a constant name `-D` may bind.
var defineName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ParseDefines reads `NAME=VALUE` strings into a map; the later of two equal names wins.
// A missing `=`, an invalid name, or one of the reserved names ARGS and ARG0 is an error.
func ParseDefines(defs []string) (map[string]string, error) {
	out := make(map[string]string, len(defs))
	for _, d := range defs {
		name, value, ok := strings.Cut(d, "=")
		if !ok {
			return nil, fmt.Errorf("invalid definition %q: want NAME=VALUE", d)
		}
		if err := checkDefineName(name); err != nil {
			return nil, err
		}
		out[name] = value
	}
	return out, nil
}

func checkDefineName(name string) error {
	if !defineName.MatchString(name) {
		return fmt.Errorf("invalid definition name %q: want an identifier (NAME=VALUE)", name)
	}
	if name == "ARGS" || name == "ARG0" {
		return fmt.Errorf("cannot define %s: it is bound to the script arguments", name)
	}
	return nil
}

// setDefines binds the `-D` string constants.
func (i *Interpreter) setDefines(defs map[string]string) error {
	for name, value := range defs {
		if err := checkDefineName(name); err != nil {
			return err
		}
		i.Constants[name] = value
	}
	if len(defs) > 0 {
		i.invalidateCompiledCache()
	}
	return nil
}

// errorCounter forwards writes and counts the lines that report a failure: an `[ERROR]`
// result, or a statement that printed nothing (the DSL gate rejects both).
type errorCounter struct {
	w io.Writer
	n atomic.Int64
}

func (e *errorCounter) Write(p []byte) (int, error) {
	if bytes.HasPrefix(p, []byte("∷ [ERROR]")) || bytes.HasPrefix(p, []byte("(No output or node unhandled)")) {
		e.n.Add(1)
	}
	return e.w.Write(p)
}

// runEval evaluates statements as if typed at the prompt and fails if any printed an [ERROR].
func (i *Interpreter) runEval(stmts []string) error {
	if len(stmts) == 0 {
		return nil
	}
	counter := &errorCounter{w: i.out()}
	saved := i.Out
	i.Out = counter
	defer func() { i.Out = saved }()
	for _, s := range stmts {
		i.EvaluateLine(s)
	}
	if n := counter.n.Load(); n > 0 {
		return fmt.Errorf("%d evaluated statement(s) failed or printed nothing", n)
	}
	return nil
}

// Run binds ARGS, ARG0 and the -D constants, loads the extra preludes and then o.Script (its
// top-level statements run and print as usual), runs o.Eval, and finally runs the selected
// task: o.Task, or `main` if the script defines it. An unreadable script or prelude, an
// unknown task, or a task that reports an `[ERROR]` is an error; the task's lines are
// returned either way so the caller can print them first.
func (i *Interpreter) Run(o RunOptions) (RunResult, error) {
	i.setArgs(o.Script, o.Args)
	if err := i.setDefines(o.Define); err != nil {
		return RunResult{}, err
	}
	for _, p := range o.Preludes {
		content, err := os.ReadFile(p)
		if err != nil {
			return RunResult{}, fmt.Errorf("cannot read prelude %q: %w", p, err)
		}
		i.LoadScriptContent(string(content), filepath.Dir(p))
	}
	if o.Script != "" {
		content, err := os.ReadFile(o.Script)
		if err != nil {
			return RunResult{}, fmt.Errorf("cannot read script %q: %w", o.Script, err)
		}
		i.LoadScriptContent(string(content), filepath.Dir(o.Script))
	}
	if err := i.runEval(o.Eval); err != nil {
		return RunResult{}, err
	}
	name := o.Task
	if name == "" {
		if len(o.Eval) > 0 {
			return RunResult{}, nil
		}
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
