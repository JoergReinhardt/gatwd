package cli

import (
	"fmt"
	"iter"
	"strings"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/parser"
)

// This file connects the first-class Closure value (functions.Closure, built by the
// parser's LambdaExpr) to the interpreter's three function surfaces:
//
//   - expression position — a `def f = \x y -> …` resolves, by name, to a Closure
//     through exprFallback, so `f 1 2`, `map (f 1)` and `fold f seed` all see a real
//     curried function value rather than re-parsed macro text;
//   - pipeline position — `x |> f a b` applies f to a b x (data-LAST, the Haskell
//     convention), and a bare one-argument closure name maps over collections exactly
//     like the inline `map` lambda it replaces;
//   - higher-order steps — `fold` and `filter`/`reject` accept a closure (a name or a
//     partial application) wherever they accept an inline lambda.
//
// Lexical scoping: a lambda captures the name-lookup of the interpreter that
// evaluated it (let-bound values included), so a closure returned from or stored in a
// let keeps its definition-site bindings.

// defClosure resolves a `def name = \params -> body` to a *functions.Closure whose
// free variables resolve through the interpreter's expression environment. It
// returns false when name is not a def, or its body is not a single lambda
// expression (such defs remain pipeline macros).
func (i *Interpreter) defClosure(name string) (*functions.Closure, bool) {
	src, ok := i.Funcs[name]
	if !ok {
		return nil, false
	}
	src = strings.TrimSpace(src)
	if !strings.HasPrefix(src, "\\") && !strings.HasPrefix(src, "(\\") {
		return nil, false
	}
	var expr parser.Expr
	if i.lambdaCache != nil {
		if cached, ok := i.lambdaCache.Load(name); ok {
			expr = cached.(parser.Expr)
		}
	}
	if expr == nil {
		parsed, err := parser.ParseInfixExpr(src)
		if err != nil {
			return nil, false
		}
		expr = parsed
		if i.lambdaCache != nil {
			i.lambdaCache.Store(name, expr)
		}
	}
	lam, isLambda := expr.(parser.LambdaExpr)
	if !isLambda {
		return nil, false
	}
	v, err := lam.Eval(i.graphLookup(nil))
	if err != nil {
		return nil, false
	}
	c, ok := v.(*functions.Closure)
	return c, ok
}

// callableNamed resolves head to a callable value: a let-bound closure first
// (RuntimeNodes), then a lambda-bodied def. fromDef reports the latter.
func (i *Interpreter) callableNamed(head string) (fn any, fromDef, ok bool) {
	if rt, found := i.RuntimeNodes[head]; found && rt.Value.IsRight() {
		if v, _ := rt.Value.UnwrapRight(); functions.IsCallable(v) {
			return v, false, true
		}
	}
	if c, found := i.defClosure(head); found {
		return c, true, true
	}
	return nil, false, false
}

// closureStep compiles a pipeline step whose head names a closure.
//
//   - `f` (no arguments), arity 1, let-bound: map the closure over the incoming value
//     structurally (identical to `map (\x -> …)`); a lambda-bodied def with arity 1
//     keeps its existing macro path, which has the same element-wise meaning.
//   - `f a b`: apply f to the evaluated a, b and the incoming value LAST.
//   - `f` (no arguments), arity ≥ 2: apply to the incoming value, yielding the
//     partially-applied closure.
func (i *Interpreter) closureStep(stepStr string) (evaluator.PipelineStep[functions.Either[error, any]], bool) {
	fields := strings.Fields(stepStr)
	if len(fields) == 0 || !isIdent(fields[0]) {
		return nil, false
	}
	head := fields[0]
	fn, fromDef, ok := i.callableNamed(head)
	if !ok {
		return nil, false
	}
	var argExprs []parser.Expr
	if len(fields) > 1 {
		expr, err := parser.ParseInfixExpr(stepStr)
		app, isApp := expr.(parser.ApplyExpr)
		if err != nil || !isApp {
			return nil, false
		}
		if v, isVar := app.Head.(parser.VarExpr); !isVar || v.Name != head {
			return nil, false
		}
		argExprs = app.Args
	} else if c, isClosure := fn.(*functions.Closure); isClosure && c.Arity() == 1 {
		if fromDef {
			return nil, false
		}
		return mapStepOver(func(item any, _ dsltypes.BitFlag) (any, error) {
			return c.Apply([]any{item})
		}), true
	}

	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) {
						return
					}
					continue
				}
				val, _ := node.Value.UnwrapRight()
				lookup := i.graphLookup(nil)
				args := make([]any, 0, len(argExprs)+1)
				var evalErr error
				for _, ae := range argExprs {
					av, err := ae.Eval(lookup)
					if err != nil {
						evalErr = err
						break
					}
					args = append(args, av)
				}
				var res any
				if evalErr == nil {
					res, evalErr = functions.Apply(fn, append(args, val))
				}
				out := dsltypes.NewASTNode(functions.Right[error, any](res), node.DSLType)
				if evalErr != nil {
					out = evaluator.InjectError[any](fmt.Errorf("%s: %w", head, evalErr), node.DSLType)
				}
				if !yield(out) {
					return
				}
			}
		}
	}, true
}

// foldClosureStep compiles `fold f [seed]` where f resolves to a 2-ary closure (a
// named def, a let-bound closure, or a partial application of a wider one). With a
// seed it is a left fold from the seed; without one the first element seeds the
// accumulator (foldl1), and an empty input is an error rather than a silent zero.
// It reports false when f is not a closure, leaving the fixed reducers
// (sum/product/min/max/…) untouched.
func (i *Interpreter) foldClosureStep(op string) (evaluator.PipelineStep[functions.Either[error, any]], bool) {
	expr, err := parser.ParseInfixExpr(op)
	if err != nil {
		return nil, false
	}
	var headExpr, seedExpr parser.Expr
	switch e := expr.(type) {
	case parser.ApplyExpr:
		headExpr = e.Head
		if len(e.Args) > 1 {
			return nil, false
		}
		if len(e.Args) == 1 {
			seedExpr = e.Args[0]
		}
	case parser.VarExpr:
		headExpr = e
	default:
		return nil, false
	}
	hv, err := headExpr.Eval(i.graphLookup(nil))
	if err != nil {
		return nil, false
	}
	c, ok := hv.(*functions.Closure)
	if !ok || c.Arity() != 2 {
		return nil, false
	}
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			var items []any
			var lastFlag dsltypes.BitFlag
			for node := range stream {
				if node.Value.IsLeft() {
					continue
				}
				v, _ := node.Value.UnwrapRight()
				items = append(items, v)
				lastFlag = node.DSLType
			}
			if len(items) == 1 {
				if col, ok := functions.ToSliceAny(items[0]); ok {
					items = col
				}
			}
			fail := func(err error) {
				yield(evaluator.InjectError[any](fmt.Errorf("fold: %w", err), lastFlag))
			}
			var acc any
			if seedExpr != nil {
				sv, err := seedExpr.Eval(i.graphLookup(nil))
				if err != nil {
					fail(err)
					return
				}
				acc = sv
			} else {
				if len(items) == 0 {
					fail(fmt.Errorf("empty input and no seed"))
					return
				}
				acc, items = items[0], items[1:]
			}
			for _, it := range items {
				next, err := c.Apply([]any{acc, it})
				if err != nil {
					fail(err)
					return
				}
				acc = next
			}
			yield(dsltypes.NewASTNode(functions.Right[error, any](acc), lastFlag))
		}
	}, true
}

// closurePredicate resolves a filter/reject argument that is NOT an inline lambda
// (a name or partial application such as `is_even` or `gt 3`) to a boolean
// predicate over elements. It reports false when the expression does not evaluate
// to a callable closure of remaining arity 1, so the keyword form is unaffected.
func (i *Interpreter) closurePredicate(arg string) (func(any, dsltypes.BitFlag) bool, bool) {
	expr, err := parser.ParseInfixExpr(arg)
	if err != nil {
		return nil, false
	}
	hv, err := expr.Eval(i.graphLookup(nil))
	if err != nil {
		return nil, false
	}
	c, ok := hv.(*functions.Closure)
	if !ok || c.Arity() != 1 {
		return nil, false
	}
	return func(item any, _ dsltypes.BitFlag) bool {
		res, err := c.Apply([]any{item})
		if err != nil {
			return false
		}
		b, _ := res.(bool)
		return b
	}, true
}

// bindingAdapter compiles the right-hand side of an in-pipeline `let x = <rhs> in …`.
// A lambda right-hand side is a VALUE (a Closure to bind), not a step to run on the
// incoming node, so it is evaluated directly in the current environment; anything
// else is an ordinary pipeline.
func (i *Interpreter) bindingAdapter(rhs string) evaluator.Adapter[functions.Either[error, any]] {
	if strings.HasPrefix(rhs, "\\") || strings.HasPrefix(rhs, "(\\") {
		if v, err := parser.EvalInfixString(rhs, i.graphLookup(nil)); err == nil {
			return func(node dsltypes.ASTNode[functions.Either[error, any]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
				return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
					yield(dsltypes.NewASTNode(functions.Right[error, any](v), node.DSLType))
				}
			}
		}
	}
	return i.pipelineToAdapter(rhs)
}

// closureMapper resolves a `map` argument that is NOT an inline lambda — a closure
// name or partial application such as `inc` or `add 10` — to an element function.
// It reports false unless the expression evaluates to a closure of remaining arity 1
// (e.g. `map (add 10)`), so sub-pipeline arguments are unaffected.
func (i *Interpreter) closureMapper(arg string) (func(any, dsltypes.BitFlag) (any, error), bool) {
	expr, err := parser.ParseInfixExpr(arg)
	if err != nil {
		return nil, false
	}
	hv, err := expr.Eval(i.graphLookup(nil))
	if err != nil {
		return nil, false
	}
	c, ok := hv.(*functions.Closure)
	if !ok || c.Arity() != 1 {
		return nil, false
	}
	return func(item any, _ dsltypes.BitFlag) (any, error) { return c.Apply([]any{item}) }, true
}
