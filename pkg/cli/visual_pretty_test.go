package cli_test

import (
	"fmt"
	"testing"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func TestVisualPretty_EditionPipeline(t *testing.T) {
	type ServiceTelemetry struct {
		Service string
		Uptime  string
		Healthy bool
	}

	edition := cli.NewEdition("microservices-telemetry").
		WithType("Node", func(s string) any { return s }).
		WithStep(cli.MapStep("probe", func(val any) any {
			return ServiceTelemetry{
				Service: fmt.Sprintf("svc-%v", val),
				Uptime:  "99.98%",
				Healthy: true,
			}
		}))

	interp := edition.BuildInterpreter()
	results := interp.EvaluateToValues("spawn Node: auth |> probe")

	opts := visualizer.ParseOptions("--curved --border --color-type=ServiceTelemetry:cyan")
	t.Log("=== GATWD CUSTOM EDITION PIPELINE OUTPUT ===")
	for _, res := range results {
		card := visualizer.Format(res, opts)
		t.Logf("\n%s", card)
	}
}

func TestVisualPretty_ApplianceDashboard(t *testing.T) {
	embeddedRules := `
type ServerMetric = Int
const MAX_CONCURRENT = 5000
const BUFFER_SIZE = 1024
`
	app, err := cli.NewAppliance(cli.WithEmbeddedScript(embeddedRules))
	if err != nil {
		t.Fatalf("failed to create appliance: %v", err)
	}

	maxConc, _ := app.GetConstant("MAX_CONCURRENT")
	bufSize, _ := app.GetConstant("BUFFER_SIZE")

	type ApplianceDashboard struct {
		ApplianceID   string
		MaxConcurrent any
		BufferSize    any
		EmbeddedRules int
		State         string
	}

	dashboard := ApplianceDashboard{
		ApplianceID:   "gatwd-appliance-edge-01",
		MaxConcurrent: maxConc,
		BufferSize:    bufSize,
		EmbeddedRules: 2,
		State:         "RUNNING_HOT_RELOADABLE",
	}

	opts := visualizer.ParseOptions("--heavy --border")
	card := visualizer.Format(dashboard, opts)
	t.Logf("=== GATWD APPLIANCE RUNTIME DASHBOARD ===\n%s", card)
}
