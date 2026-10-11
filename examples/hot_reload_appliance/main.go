package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

//go:embed default_rules.gatwd
var embeddedRules string

func main() {
	overridePath := "/etc/gatwd/rules.gatwd"
	if len(os.Args) > 1 {
		overridePath = os.Args[1]
	}

	// First-class Appliance constructor handles embedded fallback, ConfigMap loading,
	// and zero-downtime atomic hot-reloading with zero boilerplate.
	app, err := cli.NewAppliance(
		cli.WithEmbeddedScript(embeddedRules),
		cli.WithConfigMapOverride(overridePath),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize appliance: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Gatwd Hot-Reloading Appliance Running ===")
	res := app.Execute("spawn HealthStatus: OK |> eval_status")
	fmt.Printf("Health Check Result: %v\n", res)
}
