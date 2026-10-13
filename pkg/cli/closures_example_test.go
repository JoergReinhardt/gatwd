package cli_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

// Example_closures shows lambdas as first-class values: a two-argument def is
// curried, so supplying one argument yields a function that map can apply to every
// element, and a closure returned by another closure keeps its captured argument.
func Example_closures() {
	interp := cli.NewInterpreter()
	interp.EvaluateScript("def add = \\x y -> x + y\ndef mk = \\k -> \\x -> x + k")
	fmt.Println(interp.EvaluateToValues(`[1, 2, 3] |> map (add 10)`)[0])
	fmt.Println(interp.EvaluateToValues(`mk 5 10`)[0])
	fmt.Println(interp.EvaluateToValues(`[1, 2, 3] |> fold add 0`)[0])
	// Output:
	// [11 12 13]
	// 15
	// 6
}

// Example_matrixFunctionForms shows the expression-level function forms of the
// operator-type matrix: the same operation is available over every carrier, takes
// computed values (not string tokens), and is curried so a partial application is a
// reusable function. Here a HAMT is built with hamt_set inside a fold, then read back.
func Example_matrixFunctionForms() {
	interp := cli.NewInterpreter()
	fmt.Println(interp.EvaluateToValues(`hamt_get "ab" (fold (\h w -> hamt_set w (length w) h) hamt_empty ["ab", "c"])`)[0])
	fmt.Println(interp.EvaluateToValues(`map (set_contains 2) [(to_set [1, 2]), (to_set [3])]`)[0])
	fmt.Println(interp.EvaluateToValues(`maybe_default 0 (map (\x -> x + 1) (just 1))`)[0])
	// Output:
	// 2
	// [true false]
	// 2
}
