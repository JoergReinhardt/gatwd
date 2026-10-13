package cli_test

import (
	"fmt"
	"time"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

// The clock is injected: `now` reads Interpreter.Clock, so a rule such as "expires within
// 14 days" is a pure function of a fixed reading and the example never depends on the wall clock.
func Example_nowWithAnInjectedClock() {
	interp := cli.NewInterpreter()
	interp.Clock = func() time.Time { return time.Date(2026, 10, 13, 9, 30, 0, 0, time.UTC) }
	interp.EvaluateScript(`def expiring = \now exp -> (exp - now) < 14d`)

	res := interp.EvaluateToValues(`now |> map (\t -> expiring t @2026-10-20T00:00:00Z)`)
	fmt.Println(res[0])
	res = interp.EvaluateToValues(`now |> map (\t -> expiring t @2026-12-01T00:00:00Z)`)
	fmt.Println(res[0])
	// Output:
	// true
	// false
}
