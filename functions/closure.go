package functions

import (
	"fmt"
	"strings"
)

// Closure is a first-class, curried, lexically-scoped function value of the guest
// DSL — the runtime representation of `\x y -> x + y`.
//
// # The mathematical concept
//
// A closure pairs a function with the environment it was defined in. In the lambda
// calculus, `\x -> body` is an abstraction; evaluating it in an environment ρ yields
// the pair (λx.body, ρ) so that free variables in the body keep the meaning they had
// at the definition site (lexical scope), no matter where the function is later
// called. Currying is the isomorphism (A×B → C) ≅ (A → (B → C)): a function of n
// arguments is a chain of n one-argument functions, so applying fewer arguments than
// the arity simply yields another function awaiting the rest.
//
// # Why we need it here
//
// Before Closure, a `def` stored raw source text that was re-parsed on every call and
// a lambda only existed inline inside map/filter/fold. Closures make functions values:
// they can be bound with let, passed to user-written higher-order functions, returned
// from functions, stored in collections and partially applied (`map (add 1)`).
//
// # How it is implemented
//
// Params lists the names still awaiting arguments and Body receives exactly
// len(Params) evaluated arguments. The defining environment is captured inside Body
// by whoever builds the Closure (the parser's LambdaExpr captures the lookup function
// at its Eval site). Apply implements currying by wrapping Body when under-applied
// and re-applying the result when over-applied. Closure is immutable, so it is safe to
// share across goroutines and fold iterations.
//
// # Example
//
//	add := NewClosure([]string{"x", "y"}, func(a []any) (any, error) {
//	    return a[0].(int) + a[1].(int), nil
//	})
//	inc, _ := add.Apply([]any{1})   // *Closure awaiting y
//	v, _ := Apply(inc, []any{41})   // 42
type Closure struct {
	// Params are the names of the parameters that have not been supplied yet.
	Params []string
	// Body evaluates the function once all len(Params) arguments are supplied.
	Body func(args []any) (any, error)
}

// NewClosure builds a Closure over params whose Body is called with exactly
// len(params) arguments once fully applied.
func NewClosure(params []string, body func(args []any) (any, error)) *Closure {
	return &Closure{Params: params, Body: body}
}

// Arity returns the number of arguments the closure still awaits.
func (c *Closure) Arity() int { return len(c.Params) }

// Apply supplies args to the closure with currying semantics: fewer than Arity
// arguments return a new Closure awaiting the remainder, exactly Arity runs the body,
// and more than Arity applies the body's result (which must itself be callable) to
// the surplus. Applying zero arguments returns the closure unchanged.
func (c *Closure) Apply(args []any) (any, error) {
	n, arity := len(args), len(c.Params)
	switch {
	case n == 0:
		return c, nil
	case n < arity:
		bound := append([]any(nil), args...)
		body := c.Body
		return &Closure{
			Params: c.Params[n:],
			Body: func(rest []any) (any, error) {
				all := make([]any, 0, len(bound)+len(rest))
				all = append(all, bound...)
				all = append(all, rest...)
				return body(all)
			},
		}, nil
	case n == arity:
		return c.Body(args)
	default:
		res, err := c.Body(args[:arity])
		if err != nil {
			return nil, err
		}
		if !IsCallable(res) {
			return nil, fmt.Errorf("applied %d arguments to a function of arity %d (result %T is not callable)", n, arity, res)
		}
		return Apply(res, args[arity:])
	}
}

// String renders the closure as `<fn x y>`, listing its remaining parameters.
func (c *Closure) String() string {
	return "<fn " + strings.Join(c.Params, " ") + ">"
}

// IsCallable reports whether v can be applied: a *Closure or a DSLFunc.
func IsCallable(v any) bool {
	switch v.(type) {
	case *Closure, DSLFunc:
		return true
	}
	return false
}

// Apply applies the callable fn (a *Closure or DSLFunc) to args, returning an error
// instead of panicking when fn is not callable.
func Apply(fn any, args []any) (any, error) {
	switch f := fn.(type) {
	case *Closure:
		return f.Apply(args)
	case DSLFunc:
		return f(args)
	default:
		return nil, fmt.Errorf("cannot apply non-function value of type %T", fn)
	}
}
