package cli

import (
	"sort"

	"github.com/joergreinhardt/gatwd/functions"
)

// FunctionForm describes one expression-level builtin: its name and, for curried
// closures, the parameter names in application order (data-last: the operated-on
// value is the final parameter).
type FunctionForm struct {
	Name   string
	Params []string
}

// FunctionForms lists every registered expression-level builtin (the function forms of
// the operator-type matrix, graph, text, time, JSON, ... rows) sorted by name. It reads
// the live registry, so tooling such as cmd/gatwd-gen never carries a hand-written copy
// that can drift from what the interpreter actually resolves.
func FunctionForms() []FunctionForm {
	out := make([]FunctionForm, 0, len(dslBuiltinFuncs))
	for name, v := range dslBuiltinFuncs {
		f := FunctionForm{Name: name}
		if c, ok := v.(*functions.Closure); ok {
			f.Params = append(f.Params, c.Params...)
		}
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
