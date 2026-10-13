package cli

import (
	"fmt"
	"io/fs"

	"github.com/spf13/cobra"
)

// preludeEntry is one DSL prelude queued on an Edition. Exactly one of its two
// shapes is populated: an inline script (fsys == nil) or an embedded file read
// from fsys at path. Entries are evaluated in append order when the interpreter
// is built, so a later prelude may build on definitions from an earlier one.
type preludeEntry struct {
	script string // inline prelude source (used when fsys == nil)
	fsys   fs.FS  // embedded filesystem (e.g. a //go:embed embed.FS)
	path   string // path within fsys
}

// Edition represents a custom Gatwd binary distribution (e.g. "gatwd-cloud", "gatwd-k8s")
// configured with domain types, constants, custom step transducers, and
// preloaded DSL preludes.
type Edition struct {
	Name         string
	bare         bool
	types        map[string]func(string) any
	constants    map[string]any
	stepHandlers []StepHandler
	preludes     []preludeEntry
	commands     []*cobra.Command
	normalize    func([]string) []string
}

// NewEdition creates a new fluent builder for authoring a custom Gatwd distribution.
//
// Example:
//
//	func main() {
//	    cli.NewEdition("gatwd-cloud").
//	        WithType("Email", parseEmail).
//	        WithStep(cli.MapStep("mask", maskFn)).
//	        Run()
//	}
func NewEdition(name string) *Edition {
	return &Edition{
		Name:      name,
		types:     make(map[string]func(string) any),
		constants: make(map[string]any),
	}
}

// Bare configures the edition to use the lean Bare Runtime instead of full builtins.
func (e *Edition) Bare() *Edition {
	e.bare = true
	return e
}

// WithType registers a domain type constructor for the DSL parser.
func (e *Edition) WithType(name string, factory func(string) any) *Edition {
	e.types[name] = factory
	return e
}

// WithConstant registers an immutable host constant into the interpreter environment.
func (e *Edition) WithConstant(name string, val any) *Edition {
	e.constants[name] = val
	return e
}

// WithStep registers a custom step handler (such as MapStep, FilterStep, ArgStep).
func (e *Edition) WithStep(handler StepHandler) *Edition {
	e.stepHandlers = append(e.stepHandlers, handler)
	return e
}

// WithPrelude registers an inline .gatwd prelude script to preload into the
// edition. The script is evaluated once, when the interpreter is built, *after*
// all types, constants, and steps are registered — so a prelude may freely use
// the edition's registered builtins and types. Preludes accumulate and run in
// registration order, so a later prelude can build on definitions introduced by
// an earlier one.
//
// Use this to ship DSL-defined types, functions, and domain rules built into the
// edition (e.g. a graph standard library, a set of relation rules) instead of
// requiring the user to pass them as a script at runtime.
func (e *Edition) WithPrelude(script string) *Edition {
	e.preludes = append(e.preludes, preludeEntry{script: script})
	return e
}

// WithPreludeFS registers an embedded .gatwd prelude, read from fsys at path,
// to preload into the edition. It is the canonical way to bake a DSL standard
// library into an edition binary via a //go:embed embed.FS. Evaluation timing
// and ordering are identical to WithPrelude (after registration, in registration
// order, interleaved across both WithPrelude and WithPreludeFS calls).
//
// Example:
//
//	//go:embed lib/*.gatwd
//	var libFS embed.FS
//
//	cli.NewEdition("gatwd-k8s").
//	    WithStep(crawlStep).
//	    WithPreludeFS(libFS, "lib/prelude.gatwd").
//	    Run()
func (e *Edition) WithPreludeFS(fsys fs.FS, path string) *Edition {
	e.preludes = append(e.preludes, preludeEntry{fsys: fsys, path: path})
	return e
}

// WithCommand adds a domain subcommand to the edition's command line. The command maps
// flags to a DSL task or pipeline defined in the prelude — it holds no logic itself — and
// obtains its interpreter from BuildInterpreter.
func (e *Edition) WithCommand(cmd *cobra.Command) *Edition {
	e.commands = append(e.commands, cmd)
	return e
}

// WithArgNormalizer registers a rewrite applied to the command-line arguments before they
// are parsed, so an edition can keep accepting a retired spelling (a single-dash long flag,
// a renamed flag) for a release while it prints its own notice.
func (e *Edition) WithArgNormalizer(fn func([]string) []string) *Edition {
	e.normalize = fn
	return e
}

// Command returns the edition's command line: the script runner, `eval`, `check`, `task`,
// `repl`, `prelude`, `builtins`, `version`, shell completion, and the commands added with
// WithCommand.
func (e *Edition) Command() *CommandTree {
	preludes := make([]Prelude, len(e.preludes))
	for k, p := range e.preludes {
		p := p
		name := p.path
		if p.fsys == nil {
			name = fmt.Sprintf("inline-%d", k+1)
		}
		preludes[k] = Prelude{Name: name, Source: func() (string, error) {
			if p.fsys == nil {
				return p.script, nil
			}
			b, err := fs.ReadFile(p.fsys, p.path)
			return string(b), err
		}}
	}
	return NewCommandTree(CommandSpec{
		Name:          e.Name,
		Build:         e.buildInterpreter,
		Preludes:      preludes,
		Commands:      e.commands,
		NormalizeArgs: e.normalize,
	})
}

// BuildInterpreter instantiates and configures a new Interpreter from this Edition blueprint.
func (e *Edition) BuildInterpreter() *Interpreter {
	return e.buildInterpreter(false)
}

// buildInterpreter is BuildInterpreter with the option to leave the embedded preludes out
// (types, constants and steps stay registered).
func (e *Edition) buildInterpreter(skipPreludes bool) *Interpreter {
	var interp *Interpreter
	if e.bare {
		interp = NewBareInterpreter()
	} else {
		interp = NewInterpreter()
	}

	for name, factory := range e.types {
		interp.RegisterType(name, factory)
	}
	for name, val := range e.constants {
		interp.RegisterConstant(name, val)
	}
	for _, handler := range e.stepHandlers {
		interp.RegisterStep(handler)
	}
	// Preludes run last, in registration order, so they can use every registered
	// type, constant, and step. An embedded prelude that fails to load (e.g. a bad
	// //go:embed path) is a programmer error; surface it loudly rather than
	// shipping a silently half-configured edition.
	for _, p := range e.preludes {
		if skipPreludes {
			break
		}
		if p.fsys != nil {
			if err := interp.LoadScriptFS(p.fsys, p.path); err != nil {
				panic("gatwd edition " + e.Name + ": failed to load embedded prelude '" + p.path + "': " + err.Error())
			}
		} else {
			interp.EvaluateScript(p.script)
		}
	}
	return interp
}

// Run executes the CLI using this configured Edition.
func (e *Edition) Run() {
	e.Command().exitWith()
}
