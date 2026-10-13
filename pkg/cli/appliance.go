package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
)

// OverrideMode says how an Appliance combines its embedded rules with an override file
// (typically a mounted ConfigMap).
//
// Think of systemd drop-ins, Helm values files or Kustomize overlays: the base is loaded
// first and the override is applied on top, so the override names only what it changes.
// That is OverrideLayer, the default. OverrideReplace is the explicit alternative, "the
// override file is the whole rule set", for a deployment that wants to own every rule.
type OverrideMode int

const (
	// OverrideLayer loads the embedded rules first and the override on top of them in the
	// same interpreter. A `def` in the override redefines the embedded one of the same name
	// and may call any other embedded def, so an override can reuse the embedded rule_*
	// closures and redefine only what it needs. This is the default.
	OverrideLayer OverrideMode = iota
	// OverrideReplace loads the override alone; the embedded rules are not loaded while an
	// override is present. With no override (file absent or empty) the embedded rules run.
	OverrideReplace
)

// String returns the flag spelling of the mode: "layer" or "replace".
func (m OverrideMode) String() string {
	if m == OverrideReplace {
		return "replace"
	}
	return "layer"
}

// ParseOverrideMode reads the flag spelling of a mode ("layer" or "replace"); anything else
// is an error naming the accepted values.
func ParseOverrideMode(s string) (OverrideMode, error) {
	switch s {
	case "layer":
		return OverrideLayer, nil
	case "replace":
		return OverrideReplace, nil
	}
	return OverrideLayer, fmt.Errorf("unknown rules mode %q: want layer or replace", s)
}

// LoadReport describes one successful (re)load of an Appliance: which override, if any,
// was applied and what it changed relative to the embedded rules. It is what an operator
// prints so that nobody has to diff the ConfigMap against the binary by hand.
type LoadReport struct {
	// Mode is the OverrideMode in force.
	Mode OverrideMode
	// Override is the path of the override that was applied; empty when the embedded rules
	// ran alone (no override configured, file absent, or file empty).
	Override string
	// Overridden lists the defs (and, as "const NAME", the constants) that the override
	// redefined with a body different from the embedded one. Always empty in
	// OverrideReplace mode, where there is nothing underneath to compare with.
	Overridden []string
	// Added lists the defs and constants the override introduced. In OverrideReplace mode
	// it is empty as well; the whole override is the rule set.
	Added []string
}

// Lines renders the report as the human-readable lines an operator logs. The first line
// names the override (or says the embedded rules run alone); a layered override then gets
// one line for the rules it overrides and one for those it adds.
func (r LoadReport) Lines() []string {
	if r.Override == "" {
		return []string{"rules: embedded rules, no override"}
	}
	if r.Mode == OverrideReplace {
		return []string{fmt.Sprintf("rules: override %s REPLACES the embedded rules (mode replace)", r.Override)}
	}
	lines := []string{fmt.Sprintf("rules: override %s layered on the embedded rules (mode layer)", r.Override)}
	if len(r.Overridden) > 0 {
		lines = append(lines, "rules: overridden: "+strings.Join(r.Overridden, ", "))
	}
	if len(r.Added) > 0 {
		lines = append(lines, "rules: added: "+strings.Join(r.Added, ", "))
	}
	return lines
}

// String is the report's lines joined with newlines.
func (r LoadReport) String() string { return strings.Join(r.Lines(), "\n") }

// LoadError is returned when the rules of an Appliance cannot be loaded: an unreadable
// override file, a structural error in it, a failing statement or import, or a required
// def left undefined. The Appliance keeps serving the previous rules when a hot reload
// fails; at construction the error is what NewAppliance returns.
type LoadError struct {
	// Source names what failed to load: an override path, "embedded rules" or "appliance".
	Source string
	// Problems are the individual findings, one line each, with line numbers where known.
	Problems []string
}

// Error lists every problem so a broken ConfigMap is diagnosed from one log line.
func (e *LoadError) Error() string {
	return fmt.Sprintf("%s: %d problem(s): %s", e.Source, len(e.Problems), strings.Join(e.Problems, "; "))
}

// ApplianceOption configures a Gatwd Appliance instance.
type ApplianceOption func(*Appliance)

// applianceState is one fully loaded rule set; Reload swaps whole states atomically, so
// Execute never sees a half-loaded interpreter and a failed reload leaves the last good
// state in place.
type applianceState struct {
	interp *Interpreter
	report LoadReport
}

// Appliance represents a production-grade, self-contained service, Kubernetes Operator,
// or CLI daemon that evaluates Gatwd DSL pipelines with atomic zero-downtime hot-reloading.
//
// Its rules come from two layers: the embedded rules (a //go:embed file or a string) and
// an optional override file such as a mounted ConfigMap. By default the override is layered
// on top of the embedded rules (OverrideLayer); WithOverrideMode(OverrideReplace) makes it
// replace them. Loading fails loudly: a broken override is an error (from NewAppliance, or
// from Reload while the previous rules keep running), never a silent fall-back.
type Appliance struct {
	edition      *Edition
	state        atomic.Pointer[applianceState]
	overridePath string
	embeddedFS   fs.FS
	embeddedPath string
	embeddedStr  string
	mode         OverrideMode
	required     []string
	policy       Policy // zero value: DenyAll
}

// NewAppliance creates and initializes a hot-reloadable Gatwd Appliance.
// It loads the embedded rules and, if a live ConfigMap override is present on disk,
// layers it on top (or, in OverrideReplace mode, loads it instead). A broken override is
// an error.
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
// A missing or empty file means "no override"; an unreadable or broken one is a LoadError.
// `import` statements inside the override resolve against the override file's directory,
// not the process's working directory.
func WithConfigMapOverride(path string) ApplianceOption {
	return func(a *Appliance) { a.overridePath = path }
}

// WithOverrideMode selects how the override combines with the embedded rules: OverrideLayer
// (the default) or OverrideReplace.
func WithOverrideMode(mode OverrideMode) ApplianceOption {
	return func(a *Appliance) { a.mode = mode }
}

// WithPolicy sets the execution policy the Appliance's rules run under: what the embedded
// rules and the override may do beyond computing (see Policy). The default is DenyAll — an
// override is a script that did not come from the command line (decision D8), so it cannot
// run OS commands, touch files, open sockets or read the environment unless the operator of
// the process grants it here (the command line has --allow-exec, --allow-fs, --allow-net and
// --allow-env). Steps the Edition wired itself (a cluster crawl, `manifests`) are not
// affected. The policy is set by Go code from the process's flags; no rule text can widen it.
func WithPolicy(p Policy) ApplianceOption {
	return func(a *Appliance) { a.policy = p }
}

// Policy returns the execution policy the Appliance's rules run under.
func (a *Appliance) Policy() Policy { return a.policy }

// WithRequiredDefs names defs (or constants) that must be defined after every load. An
// override that leaves one undefined — in OverrideReplace mode, typically by forgetting
// the entry point — is a LoadError instead of a pipeline that fails later at run time.
func WithRequiredDefs(names ...string) ApplianceOption {
	return func(a *Appliance) { a.required = append(a.required, names...) }
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

// loadProblems is an io.Writer for Interpreter.Out while rules load. Loading prints a line
// per failing statement (`∷ [ERROR] …`, `(No output or node unhandled)`, an unreadable
// import); this collects those as problems and drops everything else, so a rule file that
// happens to print a result cannot pollute the appliance's own output.
type loadProblems struct {
	lines []string
}

func (p *loadProblems) Write(b []byte) (int, error) {
	for _, line := range strings.Split(strings.TrimRight(string(b), "\n"), "\n") {
		switch {
		case strings.HasPrefix(line, "∷ [ERROR]"):
			p.lines = append(p.lines, strings.TrimSpace(strings.TrimPrefix(line, "∷")))
		case strings.HasPrefix(line, "(No output or node unhandled)"):
			p.lines = append(p.lines, "a statement produced no output")
		case strings.HasPrefix(line, "Error reading"):
			p.lines = append(p.lines, line)
		}
	}
	return len(b), nil
}

var _ io.Writer = (*loadProblems)(nil)

// readOverride returns the override text. A missing or blank file is "no override"
// (ok == false); any other read failure is an error — a ConfigMap mounted as a directory
// must not be mistaken for an absent one.
func (a *Appliance) readOverride() (text string, ok bool, err error) {
	if a.overridePath == "" {
		return "", false, nil
	}
	data, err := os.ReadFile(a.overridePath)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "", false, nil
	case err != nil:
		return "", false, &LoadError{Source: a.overridePath, Problems: []string{err.Error()}}
	case len(bytes.TrimSpace(data)) == 0:
		return "", false, nil
	}
	return string(data), true, nil
}

// loadBase evaluates the embedded rules into interp: the embedded filesystem when one is
// configured (its imports resolve inside it), else the embedded string.
func (a *Appliance) loadBase(interp *Interpreter) error {
	switch {
	case a.embeddedFS != nil && a.embeddedPath != "":
		return interp.LoadScriptFS(a.embeddedFS, a.embeddedPath)
	case a.embeddedStr != "":
		interp.LoadScriptContent(a.embeddedStr, ".")
	}
	return nil
}

// snapshot copies what a rule file can define: def bodies and constants (as text).
func snapshot(in *Interpreter) (defs, consts map[string]string) {
	defs = make(map[string]string, len(in.Funcs))
	for k, v := range in.Funcs {
		defs[k] = v
	}
	consts = make(map[string]string, len(in.Constants))
	for k, v := range in.Constants {
		consts[k] = fmt.Sprintf("%v", v)
	}
	return defs, consts
}

// diffNames splits the names an override defines into those it redefined with a different
// value and those that are new.
func diffNames(prefix string, before, after map[string]string) (overridden, added []string) {
	for name, v := range after {
		old, existed := before[name]
		switch {
		case !existed:
			added = append(added, prefix+name)
		case old != v:
			overridden = append(overridden, prefix+name)
		}
	}
	sort.Strings(overridden)
	sort.Strings(added)
	return overridden, added
}

// load builds a complete rule set without touching the active one.
func (a *Appliance) load() (*applianceState, error) {
	overrideText, hasOverride, err := a.readOverride()
	if err != nil {
		return nil, err
	}

	interp := a.edition.BuildInterpreter()
	// The Edition's own preludes ran trusted inside BuildInterpreter (they are part of the
	// binary); everything loaded from here on — the embedded rules and above all the override —
	// runs under the Appliance's policy, and so does every later Execute.
	interp.SetPolicy(a.policy)
	problems := &loadProblems{}
	interp.Out = problems
	defer func() { interp.Out = nil }()

	report := LoadReport{Mode: a.mode}
	if !(hasOverride && a.mode == OverrideReplace) {
		if err := a.loadBase(interp); err != nil {
			return nil, &LoadError{Source: "embedded rules", Problems: []string{err.Error()}}
		}
		if len(problems.lines) > 0 {
			return nil, &LoadError{Source: "embedded rules", Problems: problems.lines}
		}
	}

	if hasOverride {
		report.Override = a.overridePath
		if diags := structuralErrors(overrideText); len(diags) > 0 {
			return nil, &LoadError{Source: a.overridePath, Problems: diags}
		}
		beforeDefs, beforeConsts := snapshot(interp)
		interp.LoadScriptContent(overrideText, filepath.Dir(a.overridePath))
		if len(problems.lines) > 0 {
			return nil, &LoadError{Source: a.overridePath, Problems: problems.lines}
		}
		if a.mode == OverrideLayer {
			afterDefs, afterConsts := snapshot(interp)
			od, ad := diffNames("", beforeDefs, afterDefs)
			oc, ac := diffNames("const ", beforeConsts, afterConsts)
			report.Overridden = append(od, oc...)
			report.Added = append(ad, ac...)
		}
	}

	var missing []string
	for _, name := range a.required {
		_, isDef := interp.Funcs[name]
		_, isConst := interp.Constants[name]
		if !isDef && !isConst {
			missing = append(missing, fmt.Sprintf("required def %q is not defined", name))
		}
	}
	if len(missing) > 0 {
		source := "embedded rules"
		if hasOverride {
			source = a.overridePath
		}
		return nil, &LoadError{Source: source, Problems: missing}
	}
	return &applianceState{interp: interp, report: report}, nil
}

// structuralErrors runs the static checker over rule text and returns its error-severity
// findings as `line N: message`. Typing findings are advisory and are not load failures.
func structuralErrors(text string) []string {
	var out []string
	for _, d := range CheckScript(text) {
		if d.Severity == "error" {
			out = append(out, fmt.Sprintf("line %d: %s", d.Line, d.Msg))
		}
	}
	return out
}

// Reload re-reads the embedded rules and the mounted ConfigMap override, compiles a fresh
// Interpreter, and atomically swaps it in without downtime.
//
// If the new rules cannot be loaded (a *LoadError) the swap does not happen: the previous,
// last good rules keep serving Execute, and the error says what is wrong so it can be
// logged and fixed. Reload therefore never degrades a running Appliance.
func (a *Appliance) Reload() error {
	st, err := a.load()
	if err != nil {
		return err
	}
	a.state.Store(st)
	return nil
}

// Report describes the rules currently in force: the override applied, if any, and which
// defs it overrode or added. The zero LoadReport is returned before the first load.
func (a *Appliance) Report() LoadReport {
	st := a.state.Load()
	if st == nil {
		return LoadReport{}
	}
	return st.report
}

// Execute evaluates a DSL pipeline line against the currently active interpreter.
func (a *Appliance) Execute(line string) []any {
	interp := a.Interpreter()
	if interp == nil {
		return nil
	}
	return interp.EvaluateToValues(line)
}

// GetConstant retrieves an evaluated constant from the currently active interpreter.
func (a *Appliance) GetConstant(name string) (any, bool) {
	interp := a.Interpreter()
	if interp == nil {
		return nil, false
	}
	val, ok := interp.Constants[name]
	return val, ok
}

// Interpreter returns the currently active Interpreter instance.
func (a *Appliance) Interpreter() *Interpreter {
	st := a.state.Load()
	if st == nil {
		return nil
	}
	return st.interp
}
