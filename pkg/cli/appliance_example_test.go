package cli_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

// ExampleWithOverrideMode shows the default, layered override: the override file names only
// what it changes, calls the embedded rule it keeps, and the appliance reports the difference.
func ExampleWithOverrideMode() {
	dir, _ := os.MkdirTemp("", "rules")
	defer os.RemoveAll(dir)
	override := filepath.Join(dir, "rules.gatwd")
	_ = os.WriteFile(override, []byte("def threshold = \\x -> x + 100\n"), 0o644)

	app, err := cli.NewAppliance(
		cli.WithEmbeddedScript("def threshold = \\x -> x + 1\ndef check = \\x -> [(threshold x)]\n"),
		cli.WithConfigMapOverride(override),
		cli.WithOverrideMode(cli.OverrideLayer), // the default; cli.OverrideReplace is the explicit alternative
		cli.WithRequiredDefs("check"),
	)
	if err != nil {
		fmt.Println("load failed:", err)
		return
	}
	fmt.Println(app.Execute("check 1")[0])
	fmt.Println(app.Report().Overridden)
	// Output:
	// [101]
	// [threshold]
}

// ExampleWithMaxOutput bounds one check: a check that produces more than its limit reports a
// single error naming the limit instead of its results (WithTimeout does the same for time).
func ExampleWithMaxOutput() {
	rules := "def entry = \\x -> [(x + 1), (x + 10)]\n"
	app, _ := cli.NewAppliance(cli.WithEmbeddedScript(rules), cli.WithMaxOutput(8))
	fmt.Println(app.Execute("entry 1")[0])

	tight, _ := cli.NewAppliance(cli.WithEmbeddedScript(rules), cli.WithMaxOutput(5))
	got := tight.Execute("entry 1")
	fmt.Println(len(got), got[0])
	// Output:
	// [2 11]
	// 1 check output exceeded the limit of 5 bytes
}
