package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/visualizer"
)

//go:embed default_rules.gatwd
var embeddedRules string

func main() {
	overridePath := "/etc/gatwd/rules.gatwd"
	if len(os.Args) > 1 {
		overridePath = os.Args[1]
	}

	opts := visualizer.Options{Border: true, Style: visualizer.CurvedStyle}

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

	res := app.Execute("spawn HealthStatus: OK |> eval_status")

	fmt.Println(visualizer.Format(map[string]any{
		"Appliance":    "Gatwd Hot-Reloading Appliance",
		"OverridePath": overridePath,
		"Status":       "Running",
		"HealthCheck":  res,
	}, opts))
}
