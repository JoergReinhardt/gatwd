package cli

// Edition represents a custom Gatwd binary distribution (e.g. "gatwd-cloud", "gatwd-k8s")
// configured with domain types, constants, and custom step transducers.
type Edition struct {
	Name         string
	bare         bool
	types        map[string]func(string) any
	constants    map[string]any
	stepHandlers []StepHandler
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

// BuildInterpreter instantiates and configures a new Interpreter from this Edition blueprint.
func (e *Edition) BuildInterpreter() *Interpreter {
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
	return interp
}

// Run executes the CLI using this configured Edition.
func (e *Edition) Run() {
	RunWithInterpreter(e.BuildInterpreter())
}
