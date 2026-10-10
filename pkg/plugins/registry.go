package plugins

import (
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// AdapterFactory is a function that generates an adapter. 
// We use a factory so it can optionally take configuration in the future.
type AdapterFactory func() evaluator.Adapter[functions.Either[error, any]]

// Registry holds all globally registered third-party Go adapters.
var Registry = make(map[string]AdapterFactory)

// Register allows third-party packages to seamlessly register native adapters
// via their init() functions.
func Register(keyword string, factory AdapterFactory) {
	Registry[keyword] = factory
}
