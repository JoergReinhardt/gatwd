package cli_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

// Example_executionPolicy shows the capability policy: the same statement runs a program for
// a trusted script and is an error value, naming the capability, once the policy denies it.
func Example_executionPolicy() {
	interp := cli.NewInterpreter() // a script from the command line is trusted
	fmt.Println(interp.EvaluateToValues("spawn StrVal: x |> echo hello")[0])

	interp.SetPolicy(cli.DenyAll()) // what a mounted override gets
	denied := interp.EvaluateToValues("spawn StrVal: x |> echo hello")[0].(error)
	fmt.Println(denied)

	f := cli.FailureOf(denied)
	fmt.Println(f.Kind, f.Step, f.Capability)
	// Output:
	// hello
	// step 'echo' requires the exec capability (policy denies it)
	// capability_denied echo exec
}

// Example_runBounds shows RunOptions bounding a script that did not come from the command
// line: it has no capabilities of its own and is stopped when its deadline passes.
func Example_runBounds() {
	interp := cli.NewInterpreter()
	interp.Out = io.Discard
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := interp.Run(cli.RunOptions{
		Context: ctx,
		Eval:    []string{"spawn StrVal: s |> mock_stream 2000000000 |> fold sum"},
	})
	fmt.Println(errors.Is(err, context.DeadlineExceeded))
	// Output: true
}
