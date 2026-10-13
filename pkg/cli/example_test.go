package cli_test

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/visualizer"
)

// Example demonstrates basic evaluation of a Gatwd DSL pipeline using NewInterpreter with visualizer output.
func Example() {
	interp := cli.NewInterpreter()
	interp.EvaluateLine("type Count = Int")
	res := interp.EvaluateToValues("spawn Count: 10 |> math_add 5 |> assert 15 |> visualize --curved --border")
	for _, r := range res {
		fmt.Print(r)
	}

	// Output:
	// ╭──────────────╮
	// │ ╰─ Count: 15 │
	// ╰──────────────╯
}

// ExampleNewEdition demonstrates building a custom Gatwd Edition (distribution)
// using ergonomic step and type constructors, visualized via DSL pipeline.
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
	results := interp.EvaluateToValues("spawn Word: hello |> shout |> prefix GATWD |> assert GATWD:hello! |> visualize --curved --border")
	for _, res := range results {
		fmt.Print(res)
	}

	// Output:
	// ╭───────────────────────╮
	// │ ╰─ Word: GATWD:hello! │
	// ╰───────────────────────╯
}

// ExampleNewEdition_visualized demonstrates formatting custom Edition outputs with visualizer cards.
func ExampleNewEdition_visualized() {
	type Profile struct {
		User string
		Role string
	}

	edition := cli.NewEdition("gatwd-visual").
		WithType("User", func(s string) any { return s }).
		WithStep(cli.MapStep("make_profile", func(val any) any {
			return Profile{User: fmt.Sprintf("%v", val), Role: "admin"}
		}))

	interp := edition.BuildInterpreter()
	results := interp.EvaluateToValues("spawn User: alice |> make_profile")

	opts := visualizer.ParseOptions("--curved --border")
	for _, res := range results {
		fmt.Print(visualizer.Format(res, opts))
	}

	// Output:
	// ╭────────────────────╮
	// │ ╰─ Profile         │
	// │     ├─ User: alice │
	// │     ╰─ Role: admin │
	// ╰────────────────────╯
}

// ExampleMapStep demonstrates how MapStep easily transforms items in a pipeline.
func ExampleMapStep() {
	interp := cli.NewInterpreter()
	interp.EvaluateLine("type Num = Int")

	interp.RegisterStep(cli.MapStep("double_val", func(val any) any {
		return fmt.Sprintf("%v%v", val, val)
	}))

	results := interp.EvaluateToValues("spawn Num: 42 |> double_val |> assert 4242 |> visualize --curved --border")
	for _, res := range results {
		fmt.Print(res)
	}

	// Output:
	// ╭──────────────╮
	// │ ╰─ Num: 4242 │
	// ╰──────────────╯
}

// ExampleFilterStep demonstrates how FilterStep filters stream elements.
func ExampleFilterStep() {
	interp := cli.NewInterpreter()
	interp.EvaluateLine("type Word = String")

	interp.RegisterStep(cli.FilterStep("keep_admin", func(val any) bool {
		return val == "admin"
	}))

	results := interp.EvaluateToValues("spawn Word: admin |> keep_admin |> assert admin |> visualize --curved --border")
	for _, res := range results {
		fmt.Print(res)
	}

	// Output:
	// ╭────────────────╮
	// │ ╰─ Word: admin │
	// ╰────────────────╯
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
	type ApplianceReport struct {
		Threshold any
		Execution string
	}
	rep := ApplianceReport{Threshold: thresh, Execution: "checked_OK"}
	fmt.Print(visualizer.Format(rep, visualizer.ParseOptions("--curved --border")))

	// Output:
	// ╭──────────────────────────────╮
	// │ ╰─ ApplianceReport           │
	// │     ├─ Threshold: 90         │
	// │     ╰─ Execution: checked_OK │
	// ╰──────────────────────────────╯
}

// ExampleArgStep demonstrates using ArgStep to create steps that accept string arguments.
func ExampleArgStep() {
	interp := cli.NewInterpreter()
	interp.EvaluateLine("type Text = String")

	interp.RegisterStep(cli.ArgStep("greet", func(arg string, val any) (any, error) {
		return fmt.Sprintf("%s, %v!", arg, val), nil
	}))

	results := interp.EvaluateToValues("spawn Text: Alice |> greet Hello |> assert Hello, Alice! |> visualize --curved --border")
	for _, res := range results {
		fmt.Print(res)
	}

	// Output:
	// ╭────────────────────────╮
	// │ ╰─ Text: Hello, Alice! │
	// ╰────────────────────────╯
}

// ExampleAppliance_Reload demonstrates dynamically reloading an appliance with modified rules.
func ExampleAppliance_Reload() {
	initialRules := `
type Metric = Int
const VAL = 100
`
	app, _ := cli.NewAppliance(cli.WithEmbeddedScript(initialRules))
	v1, _ := app.GetConstant("VAL")

	// Update embedded rules and reload
	updatedRules := `
type Metric = Int
const VAL = 200
`
	_ = app.Reload()
	app.Interpreter().EvaluateScript(updatedRules)
	v2, _ := app.GetConstant("VAL")

	type ReloadSummary struct {
		Initial  any
		Reloaded any
	}
	summary := ReloadSummary{Initial: v1, Reloaded: v2}
	fmt.Print(visualizer.Format(summary, visualizer.ParseOptions("--curved --border")))

	// Output:
	// ╭──────────────────────╮
	// │ ╰─ ReloadSummary     │
	// │     ├─ Initial: 100  │
	// │     ╰─ Reloaded: 200 │
	// ╰──────────────────────╯
}

// ExampleCommandTree_Execute runs a command line without a process: the tree returns the
// exit status (ExitOK, ExitFailed when a command failed, ExitUsage for a malformed command line) and writes to the
// given streams.
func ExampleCommandTree_Execute() {
	tree := cli.NewCommandTree(cli.CommandSpec{
		Name:  "gatwd",
		Build: func(bool) *cli.Interpreter { return cli.NewInterpreter() },
	})
	code := tree.Execute([]string{"eval", "-e", "1 + 2"}, strings.NewReader(""), os.Stdout, os.Stderr)
	fmt.Println("exit", code)
	code = tree.Execute([]string{"--no-such-flag"}, strings.NewReader(""), io.Discard, io.Discard)
	fmt.Println("exit", code)
	// Output:
	// ∷ 3
	// exit 0
	// exit 2
}
