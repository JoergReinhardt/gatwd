package cli

import (
	"io/fs"
	"os"
	"sync/atomic"
)

// ApplianceOption configures a Gatwd Appliance instance.
type ApplianceOption func(*Appliance)

// Appliance represents a production-grade, self-contained service, Kubernetes Operator,
// or CLI daemon that evaluates Gatwd DSL pipelines with atomic zero-downtime hot-reloading.
type Appliance struct {
	edition      *Edition
	activeInterp atomic.Pointer[Interpreter]
	overridePath string
	embeddedFS   fs.FS
	embeddedPath string
	embeddedStr  string
}

// NewAppliance creates and initializes a hot-reloadable Gatwd Appliance.
// It loads the live ConfigMap override if available on disk, or falls back to the embedded rules.
//
// Example:
//
//	app, err := cli.NewAppliance(
//	    cli.WithEmbeddedFS(rulesFS, "rules.gatwd"),
//	    cli.WithConfigMapOverride("/etc/gatwd/rules.gatwd"),
//	    cli.WithStep(cli.MapStep("mask", maskFn)),
//	)
func NewAppliance(opts ...ApplianceOption) (*Appliance, error) {
	app := &Appliance{
		edition: NewEdition("appliance"),
	}
	for _, opt := range opts {
		opt(app)
	}
	if err := app.Reload(); err != nil {
		return nil, err
	}
	return app, nil
}

// WithEmbeddedScript configures raw embedded script text as the baseline fallback.
func WithEmbeddedScript(script string) ApplianceOption {
	return func(a *Appliance) { a.embeddedStr = script }
}

// WithEmbeddedFS configures an embedded filesystem (such as embed.FS) as the baseline fallback.
func WithEmbeddedFS(fsys fs.FS, path string) ApplianceOption {
	return func(a *Appliance) {
		a.embeddedFS = fsys
		a.embeddedPath = path
	}
}

// WithConfigMapOverride specifies the filesystem path where a live ConfigMap is mounted.
func WithConfigMapOverride(path string) ApplianceOption {
	return func(a *Appliance) { a.overridePath = path }
}

// WithStep registers a step handler on the appliance's underlying edition.
func WithStep(handler StepHandler) ApplianceOption {
	return func(a *Appliance) { a.edition.WithStep(handler) }
}

// WithType registers a domain type factory on the appliance's underlying edition.
func WithType(name string, factory func(string) any) ApplianceOption {
	return func(a *Appliance) { a.edition.WithType(name, factory) }
}

// WithConstant registers an immutable host constant on the appliance's underlying edition.
func WithConstant(name string, val any) ApplianceOption {
	return func(a *Appliance) { a.edition.WithConstant(name, val) }
}

// WithBare specifies whether the appliance should run with the lean Bare Runtime.
func WithBare(bare bool) ApplianceOption {
	return func(a *Appliance) {
		if bare {
			a.edition.Bare()
		}
	}
}

// WithEdition attaches a pre-configured Edition blueprint to the appliance.
func WithEdition(edition *Edition) ApplianceOption {
	return func(a *Appliance) { a.edition = edition }
}

// Reload re-reads either the mounted ConfigMap or embedded fallback rules,
// compiles a fresh Interpreter, and atomically swaps the active pointer without downtime.
func (a *Appliance) Reload() error {
	var scriptContent string

	// 1. Attempt reading live override file from disk (e.g. ConfigMap)
	if a.overridePath != "" {
		if data, err := os.ReadFile(a.overridePath); err == nil && len(data) > 0 {
			scriptContent = string(data)
		}
	}

	// 2. Fall back to embedded FS if present
	if scriptContent == "" && a.embeddedFS != nil && a.embeddedPath != "" {
		if data, err := fs.ReadFile(a.embeddedFS, a.embeddedPath); err == nil {
			scriptContent = string(data)
		}
	}

	// 3. Fall back to embedded string if present
	if scriptContent == "" && a.embeddedStr != "" {
		scriptContent = a.embeddedStr
	}

	// 4. Build fresh interpreter from edition blueprint
	newInterp := a.edition.BuildInterpreter()
	if scriptContent != "" {
		newInterp.EvaluateScript(scriptContent)
	}

	// 5. Atomically swap pointer
	a.activeInterp.Store(newInterp)
	return nil
}

// Execute evaluates a DSL pipeline line against the currently active interpreter.
func (a *Appliance) Execute(line string) []any {
	interp := a.activeInterp.Load()
	if interp == nil {
		return nil
	}
	return interp.EvaluateToValues(line)
}

// GetConstant retrieves an evaluated constant from the currently active interpreter.
func (a *Appliance) GetConstant(name string) (any, bool) {
	interp := a.activeInterp.Load()
	if interp == nil {
		return nil, false
	}
	val, ok := interp.Constants[name]
	return val, ok
}

// Interpreter returns the currently active Interpreter instance.
func (a *Appliance) Interpreter() *Interpreter {
	return a.activeInterp.Load()
}
