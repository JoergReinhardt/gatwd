package main

import (
	"fmt"
	"iter"
	"strings"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/pkg/cli"
)

// This example demonstrates how to build a custom Gatwd Edition (e.g., "Gatwd Cloud Edition",
// "Gatwd Math Edition", or "Gatwd DataScience Edition") by embedding the core interpreter,
// importing third-party Go libraries, and registering custom pipeline steps and types.
func main() {
	// 1. Choose either NewBareInterpreter() (minimal) or NewInterpreter() (full Go std builtins)
	interp := cli.NewInterpreter()

	// 2. Register custom types from third-party Go libraries
	interp.RegisterType("Email", func(val string) any {
		return val
	})

	// 3. Register custom third-party step handlers (transducers)
	// Example: A step that masks sensitive data using a custom algorithm
	interp.RegisterStep(func(stepStr string, i *cli.Interpreter) (evaluator.PipelineStep[functions.Either[error, any]], bool) {
		if strings.HasPrefix(stepStr, "mask_secret") {
			return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
				return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
					for node := range stream {
						if node.Value.IsRight() {
							val, _ := node.Value.UnwrapRight()
							masked := fmt.Sprintf("***[%v]***", val)
							maskedNode := dsltypes.NewASTNode(functions.Right[error, any](any(masked)), node.DSLType)
							if !yield(maskedNode) {
								return
							}
						} else {
							if !yield(node) {
								return
							}
						}
					}
				}
			}, true
		}
		return nil, false
	})

	// 4. Run the customized CLI binary (interactive REPL or batch script runner)
	cli.RunWithInterpreter(interp)
}
