package cli_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

// ExampleNewEdition demonstrates building a custom Gatwd Edition (distribution)
// using ergonomic step and type constructors without touching iter.Seq or Either closures.
func ExampleNewEdition() {
	edition := cli.NewEdition("gatwd-demo").
		WithType("Word", func(s string) any { return s }).
		WithStep(cli.MapStep("shout", func(val any) any {
			return fmt.Sprintf("%v!", val)
		})).
		WithStep(cli.ArgStep("prefix", func(arg string, val any) (any, error) {
			return fmt.Sprintf("%s:%v", arg, val), nil
		}))

	interp := edition.BuildInterpreter()
	results := interp.EvaluateToValues("spawn Word: hello |> shout |> prefix GATWD |> assert GATWD:hello!")
	for _, res := range results {
		fmt.Println(res)
	}

	// Output:
	// GATWD:hello!
}

// ExampleMapStep demonstrates how MapStep easily transforms items in a pipeline.
func ExampleMapStep() {
	interp := cli.NewInterpreter()
	interp.EvaluateLine("type Num = Int")

	interp.RegisterStep(cli.MapStep("double_val", func(val any) any {
		return fmt.Sprintf("%v%v", val, val)
	}))

	results := interp.EvaluateToValues("spawn Num: 42 |> double_val |> assert 4242")
	for _, res := range results {
		fmt.Println(res)
	}

	// Output:
	// 4242
}

// ExampleFilterStep demonstrates how FilterStep filters stream elements.
func ExampleFilterStep() {
	interp := cli.NewInterpreter()
	interp.EvaluateLine("type Word = String")

	interp.RegisterStep(cli.FilterStep("keep_admin", func(val any) bool {
		return val == "admin"
	}))

	results := interp.EvaluateToValues("spawn Word: user |> emit admin |> keep_admin |> assert admin")
	for _, res := range results {
		fmt.Println(res)
	}

	// Output:
	// admin
}

// ExampleNewAppliance demonstrates creating a hot-reloadable appliance with embedded rules.
func ExampleNewAppliance() {
	embeddedDSL := `
type Status = OK | DEGRADED
const THRESHOLD = 90
def check = assert OK
`
	app, err := cli.NewAppliance(
		cli.WithEmbeddedScript(embeddedDSL),
		cli.WithStep(cli.MapStep("tag", func(val any) any {
			return fmt.Sprintf("checked_%v", val)
		})),
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	thresh, _ := app.GetConstant("THRESHOLD")
	fmt.Println("Threshold:", thresh)

	res := app.Execute("spawn Status: OK |> check |> tag |> assert checked_OK")
	for _, r := range res {
		fmt.Println(r)
	}

	// Output:
	// Threshold: 90
	// checked_OK
}
