package main

import (
	_ "embed"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

//go:embed default_rules.gatwd
var embeddedRules string

// Appliance represents a long-running service, Kubernetes Operator, or CLI daemon
// that evaluates Gatwd DSL pipelines with atomic hot-reloading capabilities.
type Appliance struct {
	activeInterp atomic.Pointer[cli.Interpreter]
	overridePath string
}

// NewAppliance initializes an appliance with the specified ConfigMap override path.
// It loads the live override if present, or falls back to the embedded rules.
func NewAppliance(overridePath string) (*Appliance, error) {
	app := &Appliance{
		overridePath: overridePath,
	}
	if err := app.Reload(); err != nil {
		return nil, err
	}
	return app, nil
}

// Reload reads either the mounted ConfigMap override file or falls back to the
// embedded default rules, compiles a fresh Interpreter, and atomically swaps the pointer.
func (a *Appliance) Reload() error {
	var scriptContent string

	// 1. Attempt reading live ConfigMap mounted file
	if a.overridePath != "" {
		if data, err := os.ReadFile(a.overridePath); err == nil && len(data) > 0 {
			scriptContent = string(data)
			fmt.Printf("[+] Reloaded live rules from ConfigMap: %s\n", a.overridePath)
		}
	}

	// 2. Fall back to embedded default rules if no override is present
	if scriptContent == "" {
		scriptContent = embeddedRules
		fmt.Println("[*] Loaded embedded fallback rules")
	}

	// 3. Compile a fresh interpreter with domain types and new rules
	newInterp := cli.NewInterpreter()
	newInterp.EvaluateScript(scriptContent)

	// 4. Atomically swap active pointer (Zero downtime!)
	a.activeInterp.Store(newInterp)
	return nil
}

// Execute evaluates a DSL pipeline line against the currently active interpreter.
func (a *Appliance) Execute(line string) []any {
	interp := a.activeInterp.Load()
	return interp.EvaluateToValues(line)
}

// GetConstant retrieves an evaluated constant from the currently active interpreter.
func (a *Appliance) GetConstant(name string) (any, bool) {
	interp := a.activeInterp.Load()
	val, ok := interp.Constants[name]
	return val, ok
}

func main() {
	overridePath := "/etc/gatwd/rules.gatwd"
	if len(os.Args) > 1 {
		overridePath = os.Args[1]
	}

	app, err := NewAppliance(overridePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize appliance: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Gatwd Hot-Reloading Appliance Running ===")
	res := app.Execute("spawn HealthStatus: OK |> eval_status")
	fmt.Printf("Health Check Result: %v\n", res)
}
