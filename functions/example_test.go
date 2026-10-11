package functions_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/visualizer"
)

type EitherStatus struct {
	Status  string
	IsRight bool
	IsLeft  bool
}

func ExampleEither() {
	opts := visualizer.ParseOptions("--curved --border")

	val := functions.Right[error, string]("healthy-service")
	rightVal, _ := val.UnwrapRight()

	status := EitherStatus{
		Status:  rightVal,
		IsRight: val.IsRight(),
		IsLeft:  val.IsLeft(),
	}

	fmt.Print(visualizer.Format(status, opts))

	// Output:
	// ╭────────────────────────────────╮
	// │ ╰─ EitherStatus                │
	// │     ├─ Status: healthy-service │
	// │     ├─ IsRight: true           │
	// │     ╰─ IsLeft: false           │
	// ╰────────────────────────────────╯
}

type CurryResult struct {
	Function   string
	Result     int
	Expression string
}

func ExampleCurry2() {
	opts := visualizer.ParseOptions("--curved --border")

	add := func(a, b int) int { return a + b }
	curriedAdd := functions.Curry2(add)
	addTen := curriedAdd(10)

	result := addTen(5)

	res := CurryResult{
		Function:   "Curry2(add)(10)(5)",
		Result:     result,
		Expression: "10 + 5",
	}

	fmt.Print(visualizer.Format(res, opts))

	// Output:
	// ╭─────────────────────────────────────╮
	// │ ╰─ CurryResult                      │
	// │     ├─ Function: Curry2(add)(10)(5) │
	// │     ├─ Result: 15                   │
	// │     ╰─ Expression: 10 + 5           │
	// ╰─────────────────────────────────────╯
}
