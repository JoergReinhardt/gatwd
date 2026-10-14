package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// This file is the generic half of "a named pipeline becomes a cobra subcommand
// too": it is domain-free and knows nothing about graphs, Kubernetes, or any
// other domain — it only promotes a DSL def that already exists in one of an
// Edition's own preludes to a first-class subcommand, so a user who never wants
// to type `task NAME -- ...` or learn `|>` syntax can run `NAME verb -- ...`
// instead. The pipeline itself stays pure DSL (Mandate 7): this is wiring, not
// logic, and Edition.WithCommand is the existing, generic extension point it
// plugs into.

// TaskCommand returns a cobra.Command named cmdName that runs the DSL def
// taskName — already defined in one of edition's own preludes — as its own
// subcommand: everything after `--` becomes the task's input stream, exactly
// like `NAME task taskName -- arg...` (see command.go's `task` subcommand), but
// promoted to `NAME cmdName -- arg...` and given its own `--help` text. Pass the
// result to Edition.WithCommand.
//
// A command built this way carries its own `-D`/`--define NAME=VALUE` flag
// (ParseDefines), independent of the root's, so a task that reads a named
// constant (gp_roots' `-D OVERLAY`/`-D PATTERN` convention) still works when
// invoked as a verb. It does not expose `--prelude`/`--no-prelude`: those pick
// which script loads, which is meaningless once the script (the Edition's own
// prelude) is already fixed. Positional args need no completion wiring of
// their own: with no ValidArgsFunction registered, cobra already tells the
// shell to fall back to its native default (filename completion) — the right
// behavior for `--` args, which are almost always paths, for free.
//
// knownDefines, if given, names the `-D`/`--define` constants taskName actually
// reads (e.g. "OVERLAY", "PATTERN" for a task following the gp_roots
// convention), so `--define <TAB>` completes "OVERLAY=", "PATTERN=" instead of
// nothing. TaskCommand has no way to discover these on its own — a task is a
// plain DSL def, not a declared signature — so the caller, which already names
// taskName and therefore already knows what it reads, supplies the list; this
// keeps the completion MECHANISM generic here while the domain knowledge of
// which names matter for a given task stays with whichever layer defines that
// task (`docs/development_playbook.md` §5b). Omit it for a task that takes no
// `-D` constants.
//
// Example:
//
//	edition := cli.NewEdition("gatwd-x").WithPreludeFS(lib, "lib/tasks.gatwd")
//	edition = edition.WithCommand(cli.TaskCommand(edition, "filter", "filter the graph", "filter_task"))
//	edition = edition.WithCommand(cli.TaskCommand(edition, "wide", "whole-repo view", "wide", "OVERLAY", "PATTERN"))
//	// Now both work:
//	//   gatwd-x task filter_task -- ...
//	//   gatwd-x filter -- ...
func TaskCommand(edition *Edition, cmdName, short, taskName string, knownDefines ...string) *cobra.Command {
	var defines []string
	cmd := &cobra.Command{
		Use:           cmdName + " [-- arg ...]",
		Short:         short,
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			d := cmd.ArgsLenAtDash()
			var argv []string
			if d >= 0 {
				argv = args[d:]
			}
			defs, err := ParseDefines(defines)
			if err != nil {
				return UsageError(err)
			}
			interp := edition.BuildInterpreter()
			interp.Out = cmd.OutOrStdout()
			res, runErr := interp.Run(RunOptions{Task: taskName, Args: argv, Define: defs})
			out := cmd.OutOrStdout()
			for _, line := range res.Lines {
				fmt.Fprintf(out, "∷ %s\n", line)
			}
			if runErr == nil && len(res.Lines) == 0 {
				fmt.Fprintln(out, "(No output or node unhandled)")
			}
			return runErr
		},
	}
	cmd.Flags().StringArrayVarP(&defines, "define", "D", nil, "bind the constant `NAME=VALUE` (repeatable)")
	if len(knownDefines) > 0 {
		names := append([]string{}, knownDefines...)
		_ = cmd.RegisterFlagCompletionFunc("define", func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			var out []string
			for _, n := range names {
				if candidate := n + "="; strings.HasPrefix(candidate, toComplete) {
					out = append(out, candidate)
				}
			}
			return out, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
		})
	}
	return cmd
}
