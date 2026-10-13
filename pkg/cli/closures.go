package cli

import (
	"errors"
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
	return i.applyStep(stepStr, head, fn, fromDef, true)
}

// builtinStep is the function-form-as-step fallback (F18): when no existing step claims
// the head and it names a registry builtin, `x |> f a b` applies the curried function
// data-last, exactly as `f a b x` — so every function form is also a step. It is
// consulted after the whole step chain and before the OS-exec fallback.
func (i *Interpreter) builtinStep(stepStr string) (evaluator.PipelineStep[functions.Either[error, any]], bool) {
	fields := strings.Fields(stepStr)
	if len(fields) == 0 || !isIdent(fields[0]) {
		return nil, false
	}
	fn, ok := dslBuiltinLookup(fields[0])
	if !ok || !functions.IsCallable(fn) {
		return nil, false
	}
	return i.applyStep(stepStr, fields[0], fn, false, false)
}

// applyStep compiles the step `head a b …` for a resolved callable fn (see closureStep).
// structural selects the let-bound-closure reading of a bare arity-1 `f` (map it over the
// incoming value); a builtin is applied to the whole value instead (`xs |> length`).
func (i *Interpreter) applyStep(stepStr, head string, fn any, fromDef, structural bool) (evaluator.PipelineStep[functions.Either[error, any]], bool) {
	fields := strings.Fields(stepStr)
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
	} else if c, isClosure := fn.(*functions.Closure); structural && isClosure && c.Arity() == 1 {
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
					if !strings.HasPrefix(evalErr.Error(), head+":") {
						evalErr = fmt.Errorf("%s: %w", head, evalErr)
					}
					out = evaluator.InjectError[any](evalErr, node.DSLType)
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
			upstreamFailed := false
			for node := range stream {
				if node.Value.IsLeft() {
					upstreamFailed = true
					if !yield(node) {
						return
					}
					continue
				}
				v, _ := node.Value.UnwrapRight()
				items = append(items, v)
				lastFlag = node.DSLType
			}
			if upstreamFailed {
				return
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

// isLambdaText reports whether arg is written as a lambda, optionally parenthesised.
func isLambdaText(arg string) bool {
	arg = strings.TrimSpace(arg)
	return strings.HasPrefix(arg, "\\") || strings.HasPrefix(arg, "(\\")
}

// stepClosure resolves the function argument of a higher-order step (`map`, `filter`,
// `reject`, `flatmap`) — an inline lambda, a closure name or a partial application —
// to one *functions.Closure of the given remaining arity. This is the single
// lambda implementation: a lambda is parsed once by parser.LambdaExpr and applied
// through Closure.Apply, so errors inside the body reach the caller as values.
//
// ok=false means arg is not function-shaped (a keyword, a sub-pipeline, a
// pipeline-bodied lambda) and the caller keeps its other readings. A non-nil err means
// it is a lambda but cannot be used here (wrong number of parameters).
func (i *Interpreter) stepClosure(arg string, arity int) (c *functions.Closure, ok bool, err error) {
	expr, perr := parser.ParseInfixExpr(arg)
	if perr != nil {
		return nil, false, nil
	}
	hv, eerr := expr.Eval(i.graphLookup(nil))
	if eerr != nil {
		return nil, false, nil
	}
	cl, isClosure := hv.(*functions.Closure)
	if !isClosure {
		return nil, false, nil
	}
	if cl.Arity() != arity {
		if isLambdaText(arg) {
			return nil, false, fmt.Errorf("function takes %d argument(s), want %d", cl.Arity(), arity)
		}
		return nil, false, nil
	}
	return cl, true, nil
}

// failStep is a pipeline step that emits one error node, used when a step cannot be
// compiled (a bad argument) so the failure is reported where the step runs instead of
// the step silently vanishing from the pipeline.
func (i *Interpreter) failStep(err error) evaluator.PipelineStep[functions.Either[error, any]] {
	return func(iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			yield(evaluator.InjectError[any](err, i.Parser.GetFlag("Variable")))
		}
	}
}

// closureMapper adapts a Closure to the element function mapStepOver expects.
func closureMapper(c *functions.Closure) func(any, dsltypes.BitFlag) (any, error) {
	return func(item any, _ dsltypes.BitFlag) (any, error) { return c.Apply([]any{item}) }
}

// closurePredicate adapts a Closure to the predicate filterStepOver expects,
// requiring a boolean result.
func closurePredicate(c *functions.Closure, negate bool) func(any, dsltypes.BitFlag) (bool, error) {
	return func(item any, _ dsltypes.BitFlag) (bool, error) {
		b, err := callPred(c, item)
		return b != negate, err
	}
}

// pipelineLambda compiles a lambda whose body is a pipeline rather than an infix
// expression (`\x -> x |> step …`) into an element function: the body runs as a
// sub-pipeline over the element, in a child interpreter that sees the current
// bindings. A body that yields no value is an error, not the unchanged element.
func (i *Interpreter) pipelineLambda(varName, body string) func(any, dsltypes.BitFlag) (any, error) {
	return func(item any, dslType dsltypes.BitFlag) (any, error) {
		child := &Interpreter{
			Engine:         i.Engine,
			EnableBuiltins: i.EnableBuiltins,
			Funcs:          i.Funcs,
			RuntimeNodes:   make(map[string]dsltypes.ASTNode[functions.Either[error, any]]),
			Variables:      make(map[string]string),
			Constants:      i.Constants,
			CustomSteps:    i.CustomSteps,
		}
		for k, v := range i.Variables {
			child.Variables[k] = v
		}
		for k, v := range i.RuntimeNodes {
			child.RuntimeNodes[k] = v
		}
		child.RuntimeNodes[varName] = dsltypes.NewASTNode(functions.Right[error, any](item), dslType)
		child.Variables[varName] = fmt.Sprintf("%v", item)

		subParts, _ := splitPipelineAware(body)
		subPipeline := child.parsePipelineSteps(subParts)
		singleton := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			yield(dsltypes.NewASTNode(functions.Right[error, any](item), dslType))
		}
		for r := range child.Engine.EvaluatePipeline(singleton, subPipeline) {
			if r.Value.IsRight() {
				v, _ := r.Value.UnwrapRight()
				return v, nil
			}
			err, _ := r.Value.UnwrapLeft()
			return nil, err
		}
		return nil, fmt.Errorf("lambda body %q yielded no value", body)
	}
}

// flatmapStepOver lifts an element function into the `flatmap` step: the Monad bind
// (bindValue) over every carrier; for a scalar stream node the result is flattened
// and each element emitted in turn.
func flatmapStepOver(evalElem func(item any, dslType dsltypes.BitFlag) (any, error)) evaluator.PipelineStep[functions.Either[error, any]] {
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
				res, err := bindValue(val, func(it any) (any, error) {
					r, e := evalElem(it, node.DSLType)
					if e != nil {
						return nil, fmt.Errorf("flatmap: element %v: %w", it, e)
					}
					return r, nil
				})
				if _, scalar := err.(*notCarrierError); scalar {
					r, e := evalElem(val, node.DSLType)
					if e != nil {
						if !yield(evaluator.InjectError[any](fmt.Errorf("flatmap: %w", e), node.DSLType)) {
							return
						}
						continue
					}
					elems, ok := functions.ToSliceAny(r)
					if !ok {
						elems = []any{r}
					}
					for _, el := range elems {
						if !yield(dsltypes.NewASTNode(functions.Right[error, any](el), node.DSLType)) {
							return
						}
					}
					continue
				}
				out := dsltypes.NewASTNode(functions.Right[error, any](res), node.DSLType)
				if err != nil {
					out = evaluator.InjectError[any](err, node.DSLType)
				}
				if !yield(out) {
					return
				}
			}
		}
	}
}

// withPipelineFallback lets a lambda whose body is really a pipeline (`\x -> echo
// Processing %s` parses as an infix expression, but `echo` is a step, not a value) run
// as one: when the closure fails because the body's head word is an undefined name,
// the sub-pipeline reading takes over, and its own failures are reported. Every other
// failure of the closure is returned as is.
func withPipelineFallback(primary, fallback func(any, dsltypes.BitFlag) (any, error), body string) func(any, dsltypes.BitFlag) (any, error) {
	head := ""
	if f := strings.Fields(body); len(f) > 0 {
		head = f[0]
	}
	return func(item any, flag dsltypes.BitFlag) (any, error) {
		res, err := primary(item, flag)
		var undefined *parser.UndefinedError
		if err != nil && errors.As(err, &undefined) && undefined.Name == head {
			return fallback(item, flag)
		}
		return res, err
	}
}

// lambdaParseCause explains why a lambda argument is not an expression.
func lambdaParseCause(arg string) error {
	_, err := parser.ParseInfixExpr(arg)
	return err
}

// graphStep is the pipeline-step type of the graph builtins.
type graphStep = evaluator.PipelineStep[functions.Either[error, any]]

// errLatch records the first error raised by a lambda inside a graph step. The graph
// steps in evaluator/ take error-free Go callbacks, so a failing lambda reports here
// instead of returning its input unchanged.
type errLatch func(error)

// graphLambda resolves the text of one graph-step lambda (`\p -> expr`) to a Closure
// of the given arity, parsed once. It reports false when the text is not such a lambda,
// so the surrounding dispatch keeps trying other steps (and ends in "unknown step").
func (i *Interpreter) graphLambda(inside string, arity int) (*functions.Closure, bool) {
	c, ok, err := i.stepClosure(inside, arity)
	return c, ok && err == nil
}

func unaryFn(c *functions.Closure, fail errLatch) func(any) any {
	return func(x any) any {
		r, err := c.Apply([]any{x})
		if err != nil {
			fail(err)
			return x
		}
		return r
	}
}

func predFn(c *functions.Closure, fail errLatch) func(any) bool {
	return func(x any) bool {
		b, err := callPred(c, x)
		if err != nil {
			fail(err)
		}
		return b
	}
}

func binaryFn(c *functions.Closure, fail errLatch) func(acc, x any) any {
	return func(acc, x any) any {
		r, err := c.Apply([]any{acc, x})
		if err != nil {
			fail(err)
			return acc
		}
		return r
	}
}

// latchedStep runs the graph step built by build once per incoming node with a fresh
// error latch. If any lambda fails while processing a node, that node's output is
// replaced by a single error node carrying the first failure.
func latchedStep(build func(fail errLatch) graphStep) graphStep {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				var first error
				inner := build(func(err error) {
					if first == nil {
						first = err
					}
				})
				single := func(y func(dsltypes.ASTNode[functions.Either[error, any]]) bool) { y(node) }
				var out []dsltypes.ASTNode[functions.Either[error, any]]
				for o := range inner(single) {
					out = append(out, o)
				}
				if first != nil {
					out = []dsltypes.ASTNode[functions.Either[error, any]]{evaluator.InjectError[any](first, node.DSLType)}
				}
				for _, o := range out {
					if !yield(o) {
						return
					}
				}
			}
		}
	}
}
