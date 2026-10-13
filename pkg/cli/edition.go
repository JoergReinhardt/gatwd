package cli

import (
	"fmt"
	"io/fs"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	flags        []func(*pflag.FlagSet)
	grants       []*capabilityFlag
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

// WithFlags lets the edition declare persistent flags on its root command (they apply to
// every subcommand and show up in --help and shell completion). fn is called each time the
// command line is built, with a fresh pflag.FlagSet, so it should bind the flags to variables
// the edition's steps read when they run, never at registration time: the flags are parsed
// after the steps are registered. It is a generic hook, like WithCommand: the edition names
// a flag, the steps (or a prelude constant) give it meaning.
//
// Example:
//
//	var verbose bool
//	cli.NewEdition("gatwd-x").
//	    WithFlags(func(f *pflag.FlagSet) { f.BoolVar(&verbose, "verbose", false, "chatty steps") }).
//	    WithStep(stepThatReads(&verbose)).
//	    Run()
func (e *Edition) WithFlags(fn func(*pflag.FlagSet)) *Edition {
	e.flags = append(e.flags, fn)
	return e
}

// capabilityFlag is one opt-in switch: a persistent flag that grants a capability to the
// interpreter the Edition builds.
type capabilityFlag struct {
	name string
	cap  Capability
	on   bool
}

// WithCapabilityFlag declares a persistent boolean flag (--NAME, default false) that grants
// the capability c to the interpreter this Edition builds. It is how a capability that
// TrustedPolicy withholds (CapClusterApply, CapClusterRollout) is opted in to, even for the
// user's own command-line script: no flag, no capability. The flag is read when the
// interpreter is built, after the command line is parsed, and only ever widens the policy by
// that one capability; nothing inside a script can do so. usage is the help text.
//
// Example:
//
//	cli.NewEdition("gatwd-k8s").
//	    WithStep(cli.GatedStep(cli.CapClusterApply, applyStep)).
//	    WithCapabilityFlag("allow-cluster-apply", cli.CapClusterApply, "let scripts create and update objects in the cluster").
//	    Run()
func (e *Edition) WithCapabilityFlag(name string, c Capability, usage string) *Edition {
	g := &capabilityFlag{name: name, cap: c}
	e.grants = append(e.grants, g)
	return e.WithFlags(func(f *pflag.FlagSet) {
		f.BoolVar(&g.on, name, false, usage+"; denied by default, even for your own script")
	})
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
	preludes = withSiblingPreludes(preludes, e.preludes)
	return NewCommandTree(CommandSpec{
		Name:          e.Name,
		Build:         e.buildInterpreter,
		Preludes:      preludes,
		Commands:      e.commands,
		Flags:         e.flags,
		NormalizeArgs: e.normalize,
	})
}

// withSiblingPreludes appends, after the entry preludes, every other `.gatwd` file that sits
// next to an embedded entry file. A prelude is usually one entry that `import`s its
// siblings (imports resolve relative to the importing file), so `prelude cat` of the entry
// alone shows only the import lines; listing the siblings makes the whole library readable
// (and copyable) from the binary. Names are the file's path inside its filesystem, an
// entry already listed is not repeated, and an unreadable directory simply adds nothing.
func withSiblingPreludes(listed []Prelude, entries []preludeEntry) []Prelude {
	seen := map[string]bool{}
	for _, p := range listed {
		seen[p.Name] = true
	}
	for _, e := range entries {
		if e.fsys == nil {
			continue
		}
		dir := path.Dir(e.path)
		files, err := fs.ReadDir(e.fsys, dir)
		if err != nil {
			continue
		}
		for _, f := range files {
			name := path.Join(dir, f.Name())
			if f.IsDir() || path.Ext(name) != ".gatwd" || seen[name] {
				continue
			}
			seen[name] = true
			fsys := e.fsys
			listed = append(listed, Prelude{Name: name, Source: func() (string, error) {
				b, err := fs.ReadFile(fsys, name)
				return string(b), err
			}})
		}
	}
	return listed
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

	for _, g := range e.grants {
		if g.on {
			interp.SetPolicy(interp.Policy().Grant(g.cap))
		}
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
