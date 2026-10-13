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
