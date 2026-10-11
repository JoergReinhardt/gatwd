package cli_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

func TestErgonomicSteps(t *testing.T) {
	interp := cli.NewInterpreter()

	// 1. Test MapStep
	interp.RegisterStep(cli.MapStep("shout", func(val any) any {
		return strings.ToUpper(fmt.Sprint(val))
	}))

	// 2. Test FilterStep
	interp.RegisterStep(cli.FilterStep("non_empty", func(val any) bool {
		return strings.TrimSpace(fmt.Sprint(val)) != ""
	}))

	// 3. Test ArgStep
	interp.RegisterStep(cli.ArgStep("prefix", func(arg string, val any) (any, error) {
		return fmt.Sprintf("%s:%v", arg, val), nil
	}))

	// 4. Test SinkStep
	var captured []any
	interp.RegisterStep(cli.SinkStep("record", func(val any) {
		captured = append(captured, val)
	}))

	interp.EvaluateLine("type StrVal = String")
	res := interp.EvaluateToValues("spawn StrVal: hello |> shout |> prefix GATWD |> record |> assert GATWD:HELLO")
	if len(res) == 0 || res[0] != "GATWD:HELLO" {
		t.Fatalf("expected pipeline to yield 'GATWD:HELLO', got %v", res)
	}

	if len(captured) == 0 || captured[0] != "GATWD:HELLO" {
		t.Fatalf("expected SinkStep to capture 'GATWD:HELLO', got %v", captured)
	}
}

func TestEditionBuilder(t *testing.T) {
	edition := cli.NewEdition("gatwd-test").
		WithType("Word", func(s string) any { return s }).
		WithConstant("APP_ENV", "staging").
		WithStep(cli.MapStep("double_str", func(val any) any {
			return fmt.Sprintf("%v%v", val, val)
		}))

	interp := edition.BuildInterpreter()

	env, ok := interp.Constants["APP_ENV"]
	if !ok || env != "staging" {
		t.Fatalf("expected constant APP_ENV to be 'staging', got %v", env)
	}

	res := interp.EvaluateToValues("spawn Word: ab |> double_str |> assert abab")
	if len(res) == 0 || res[0] != "abab" {
		t.Fatalf("expected double_str to yield 'abab', got %v", res)
	}
}

func TestFirstClassAppliance(t *testing.T) {
	tmpDir := t.TempDir()
	overrideFile := filepath.Join(tmpDir, "override.gatwd")

	defaultScript := `
type Status = OK | FAIL
const RETRY_LIMIT = 3
def check = assert OK
`
	app, err := cli.NewAppliance(
		cli.WithEmbeddedScript(defaultScript),
		cli.WithConfigMapOverride(overrideFile),
		cli.WithStep(cli.MapStep("tag_prod", func(val any) any {
			return fmt.Sprintf("prod_%v", val)
		})),
	)
	if err != nil {
		t.Fatalf("failed to create appliance: %v", err)
	}

	limit, ok := app.GetConstant("RETRY_LIMIT")
	if !ok || limit != "3" {
		t.Fatalf("expected embedded default RETRY_LIMIT '3', got %v", limit)
	}

	res := app.Execute("spawn Status: OK |> check |> tag_prod |> assert prod_OK")
	if len(res) == 0 || res[0] != "prod_OK" {
		t.Fatalf("expected result 'prod_OK', got %v", res)
	}

	// Write override to simulate live ConfigMap update
	updatedScript := `
type Status = OK | FAIL
const RETRY_LIMIT = 99
def check = assert FAIL
`
	if err := os.WriteFile(overrideFile, []byte(updatedScript), 0644); err != nil {
		t.Fatalf("failed to write override file: %v", err)
	}

	if err := app.Reload(); err != nil {
		t.Fatalf("failed to reload appliance: %v", err)
	}

	limitUpdated, ok := app.GetConstant("RETRY_LIMIT")
	if !ok || limitUpdated != "99" {
		t.Fatalf("expected updated RETRY_LIMIT '99', got %v", limitUpdated)
	}
}
