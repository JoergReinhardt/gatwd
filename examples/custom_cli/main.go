package main

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

// This example demonstrates how to build a custom Gatwd Edition (e.g., "Gatwd Cloud Edition",
// "Gatwd Math Edition", or "Gatwd DataScience Edition") using the high-level fluent Edition builder.
//
// Contributors and authors do NOT need to write complex Category Theory closures or iter.Seq trampolines;
// simply use NewEdition(), WithType(), and ergonomic step helpers like MapStep, FilterStep, and ArgStep.
func main() {
	cli.NewEdition("gatwd-cloud").
		// 1. Register domain types
		WithType("Email", func(val string) any {
			return val
		}).
		// 2. Register custom transducers using ergonomic helpers
		WithStep(cli.MapStep("mask_secret", func(val any) any {
			return fmt.Sprintf("***[%v]***", val)
		})).
		WithStep(cli.ArgStep("tag", func(arg string, val any) (any, error) {
			return fmt.Sprintf("[%s] %v", arg, val), nil
		})).
		// 3. Launch the customized CLI runtime (interactive REPL or script execution)
		Run()
}
