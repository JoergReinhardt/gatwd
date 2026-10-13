package cli

import (
	"iter"
	"strings"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// MapStep creates an ergonomic StepHandler that transforms every successful (Right) value
// in the stream using fn. It automatically handles iter.Seq streams, AST nodes, and Either wrapping.
//
// Example:
//
//	interp.RegisterStep(cli.MapStep("mask_secret", func(val any) any {
//	    return fmt.Sprintf("***[%v]***", val)
//	}))
func MapStep(name string, fn func(val any) any) StepHandler {
	return func(stepStr string, interp *Interpreter) (evaluator.PipelineStep[functions.Either[error, any]], bool) {
		if stepStr != name {
			return nil, false
		}
		step := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
			return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
				for node := range stream {
					if node.Value.IsRight() {
						val, _ := node.Value.UnwrapRight()
						res := fn(val)
						newNode := dsltypes.NewASTNode(functions.Right[error, any](res), node.DSLType)
						if !yield(newNode) {
							return
						}
					} else {
						if !yield(node) {
							return
						}
					}
				}
			}
		}
		return step, true
	}
}

// FilterStep creates an ergonomic StepHandler that keeps only items for which predicate returns true.
//
// Example:
//
//	interp.RegisterStep(cli.FilterStep("only_numbers", func(val any) bool {
//	    _, ok := val.(int)
//	    return ok
//	}))
func FilterStep(name string, predicate func(val any) bool) StepHandler {
	return func(stepStr string, interp *Interpreter) (evaluator.PipelineStep[functions.Either[error, any]], bool) {
		if stepStr != name {
			return nil, false
		}
		step := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
			return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
				for node := range stream {
					if node.Value.IsRight() {
						val, _ := node.Value.UnwrapRight()
						if predicate(val) {
							if !yield(node) {
								return
							}
						}
					} else {
						if !yield(node) {
							return
						}
					}
				}
			}
		}
		return step, true
	}
}

// ArgStep creates an ergonomic StepHandler for parameterized steps with arguments (e.g. "prefix_with <arg>").
// If the handler function returns an error, it is wrapped in an Either Left node and safely handled by the pipeline.
//
// Example:
//
//	interp.RegisterStep(cli.ArgStep("prefix", func(arg string, val any) (any, error) {
//	    return fmt.Sprintf("%s_%v", arg, val), nil
//	}))
func ArgStep(prefix string, fn func(arg string, val any) (any, error)) StepHandler {
	return func(stepStr string, interp *Interpreter) (evaluator.PipelineStep[functions.Either[error, any]], bool) {
		trimmedPrefix := strings.TrimSpace(prefix)
		if !strings.HasPrefix(stepStr, trimmedPrefix+" ") && stepStr != trimmedPrefix {
			return nil, false
		}
		arg := strings.TrimSpace(strings.TrimPrefix(stepStr, trimmedPrefix))
		step := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
			return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
				for node := range stream {
					if node.Value.IsRight() {
						val, _ := node.Value.UnwrapRight()
						res, err := fn(arg, val)
						if err != nil {
							errNode := dsltypes.NewASTNode(functions.Left[error, any](err), node.DSLType)
							if !yield(errNode) {
								return
							}
						} else {
							resNode := dsltypes.NewASTNode(functions.Right[error, any](res), node.DSLType)
							if !yield(resNode) {
								return
							}
						}
					} else {
						if !yield(node) {
							return
						}
					}
				}
			}
		}
		return step, true
	}
}

// AdapterStep exposes an arbitrary pure Adapter[any] as a named DSL pipeline step.
//
// This is the generic bridge every data-gathering edition needs: hand it any
// evaluator.Adapter[any] — a Kubernetes crawler, an AWS lister, a database query,
// a filesystem walker — together with a keyword, and the adapter becomes callable
// as "|> keyword" in the DSL. Because an Adapter maps one node to a *sequence* of
// nodes (a Kleisli fan-out), AdapterStep is how an edition lifts external data
// sources into the engine's node-expansion trampoline so their results can be
// walked, folded, and composed like any other stream.
//
// The adapter is lifted into the engine's error-aware payload type via the EitherT
// monad transformer (evaluator.EitherT), so a failing adapter surfaces a
// Left(error) node downstream instead of crashing the pipeline, and then bound
// over the stream via evaluator.StepAdapter.
//
// Unlike the pkg/plugins keyword registry — whose AdapterFactory is a nullary
// func() and therefore cannot carry configuration — AdapterStep accepts a
// fully-constructed adapter, so edition authors inject configuration (clients,
// endpoints, credentials, GVRs) by closure at assembly time.
//
// Example:
//
//	adapter := mycloud.ListInstances(client, region) // evaluator.Adapter[any]
//	interp.RegisterStep(cli.AdapterStep("list_instances", adapter))
//	// DSL:  spawn Region: eu-central-1 |> list_instances |> ...
func AdapterStep(name string, adapter evaluator.Adapter[any]) StepHandler {
	step := evaluator.StepAdapter(evaluator.EitherT(adapter))
	return func(stepStr string, interp *Interpreter) (evaluator.PipelineStep[functions.Either[error, any]], bool) {
		if stepStr != name {
			return nil, false
		}
		return step, true
	}
}

// SinkStep creates an ergonomic StepHandler for side effects (logging, metric recording, printing).
// The stream items pass through unchanged to downstream steps.
//
// Example:
//
//	interp.RegisterStep(cli.SinkStep("log_trace", func(val any) {
//	    log.Printf("[TRACE] %v", val)
//	}))
func SinkStep(name string, fn func(val any)) StepHandler {
	return func(stepStr string, interp *Interpreter) (evaluator.PipelineStep[functions.Either[error, any]], bool) {
		if stepStr != name {
			return nil, false
		}
		step := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
			return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
				for node := range stream {
					if node.Value.IsRight() {
						val, _ := node.Value.UnwrapRight()
						fn(val)
					}
					if !yield(node) {
						return
					}
				}
			}
		}
		return step, true
	}
}
