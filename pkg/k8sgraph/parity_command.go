package k8sgraph

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

// ParityCommand returns the read-only `parity` subcommand: does every resource a
// manifest subset declares exist, and match, live? See lib/k8s/parity.gatwd for
// the comparison itself (k8s_parity_report); this is only the cobra/argv wiring
// (Mandate 7 — the logic is the DSL def, this just calls it).
//
// It does not use cli.TaskCommand. k8s_parity_report needs two independently
// sourced graphs (manifests and k8s_resources), and this interpreter's
// task-closure path does not reliably resolve a lambda whose body chains more
// than one `|>`, or that calls another `|>`-bodied def by name — found while
// building this feature (docs/manifest_live_parity_design.md). A top-level
// `let ... in` expression with two independent `|>` chains IS reliable (it is
// the exact shape pkg/k8sgraph/parity_test.go's parityRows helper uses), so
// this command builds and runs one directly via RunOptions.Eval instead of
// RunOptions.Task.
//
// Usage: `gatwd-k8s parity -D NS=<namespace> -- <manifest-path>`. NS is
// required with no silent default ("" means every namespace, matching
// k8s_resources's own convention) — the same "no silent default" discipline
// as the easyk8s child repo's `-D OVERLAY`/`-D PATTERN` convention.
func ParityCommand(edition *cli.Edition) *cobra.Command {
	var defines []string
	cmd := &cobra.Command{
		Use:           "parity -D NS=<namespace> [-- manifest-path]",
		Short:         "check whether a manifest subset is present and up to date live (read-only)",
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			d := cmd.ArgsLenAtDash()
			var argv []string
			if d >= 0 {
				argv = args[d:]
			}
			if len(argv) != 1 {
				return cli.UsageError(fmt.Errorf("parity needs exactly one manifest path after --, got %d", len(argv)))
			}
			defs, err := cli.ParseDefines(defines)
			if err != nil {
				return cli.UsageError(err)
			}
			ns, ok := defs["NS"]
			if !ok {
				return cli.UsageError(fmt.Errorf(`parity needs -D NS=<namespace> ("" means every namespace)`))
			}
			interp := edition.BuildInterpreter()
			interp.Out = cmd.OutOrStdout()
			stmt := fmt.Sprintf(
				`let m = (%q |> manifests |> k8s_graph) in let l = (%q |> k8s_resources |> k8s_graph) in k8s_parity_report m l`,
				argv[0], ns)
			res, runErr := interp.Run(cli.RunOptions{Eval: []string{stmt}})
			out := cmd.OutOrStdout()
			for _, line := range res.Lines {
				fmt.Fprintf(out, "∷ %s\n", line)
			}
			return runErr
		},
	}
	cmd.Flags().StringArrayVarP(&defines, "define", "D", nil, "bind the constant `NAME=VALUE` (repeatable); NS is required")
	return cmd
}
