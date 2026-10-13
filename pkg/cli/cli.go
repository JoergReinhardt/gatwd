package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"os"
	pathpkg "path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/joergreinhardt/gatwd/collections"

	"github.com/joergreinhardt/gatwd/adapters"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/interactive"
	"github.com/joergreinhardt/gatwd/parser"
	"github.com/joergreinhardt/gatwd/pkg/plugins"
	"github.com/joergreinhardt/gatwd/visualizer"
)

// StepHandler is an extension function that parses a DSL pipeline step string.
// If the handler recognizes the step, it returns the compiled PipelineStep and true.
type StepHandler func(stepStr string, interp *Interpreter) (evaluator.PipelineStep[functions.Either[error, any]], bool)

type Interpreter struct {
	Engine       *evaluator.Engine[functions.Either[error, any]]
	Parser       *parser.Parser
	Funcs        map[string]string // Maps function name to pipeline string
	RuntimeNodes map[string]dsltypes.ASTNode[functions.Either[error, any]]
	Variables    map[string]string // Maps local 'let' variables
	Constants    map[string]any    // Registered and script-defined constants
	// compiledFuncs caches each def's parsed pipeline steps, keyed by name.
	// Only populated/served when the active scope carries no local Variables or
	// RuntimeNodes (so no value-dependent substitution can occur), which makes
	// the cached parse identical to a fresh one. Shared across forks by
	// reference (like Funcs) and cleared when a def or constant is (re)defined.
	compiledFuncs map[string][]evaluator.PipelineStep[functions.Either[error, any]]
	// lambdaCache caches the parsed lambda expression of each lambda-bodied def,
	// keyed by name; shared across forks and cleared with compiledFuncs.
	lambdaCache    *sync.Map     // map[string]parser.Expr; concurrent-safe because ||> evaluates in parallel
	CustomSteps    []StepHandler // Pluggable extension step handlers
	EnableBuiltins bool          // Whether Go std builtins are active
	// Out receives the printed result of every top-level statement; nil means os.Stdout.
	Out io.Writer
}

// out is the writer top-level statements print to.
func (i *Interpreter) out() io.Writer {
	if i.Out == nil {
		return os.Stdout
	}
	return i.Out
}

// NewBareInterpreter creates an interpreter with bare-bones execution capabilities:
// variable bindings ('let'), macro definitions ('def'), constants ('const'), pattern
// matching ('match'), and core stream primitives ('emit', 'take', 'drop', 'head', 'tail', 'assert').
// Builtin standard library adapters and transducers are disabled unless registered.
func NewBareInterpreter() *Interpreter {
	i := &Interpreter{
		Engine:         evaluator.NewEngine[functions.Either[error, any]](),
		Parser:         parser.NewParser(),
		Funcs:          make(map[string]string),
		RuntimeNodes:   make(map[string]dsltypes.ASTNode[functions.Either[error, any]]),
		Variables:      make(map[string]string),
		Constants:      make(map[string]any),
		compiledFuncs:  make(map[string][]evaluator.PipelineStep[functions.Either[error, any]]),
		lambdaCache:    &sync.Map{},
		EnableBuiltins: false,
	}
	i.setArgs("", nil)
	i.Engine.OnMissingAdapter = func(flag dsltypes.BitFlag) evaluator.Adapter[functions.Either[error, any]] {
		// Suppress REPL pause in batch mode; just return the exec adapter
		return evaluator.EitherT(interactive.GenericExecAdapter("", flag))
	}
	return i
}

// NewInterpreter creates the standard full interpreter with all Go standard library
// builtins (HAMT, Set, Chunk, Net, Scanner, Checkpoint, Testing, Visualizer, Math) enabled.
func NewInterpreter() *Interpreter {
	i := NewBareInterpreter()
	i.EnableBuiltins = true
	return i
}

// RegisterStep registers a custom step handler for third-party libraries or domain-specific language extensions.
func (i *Interpreter) RegisterStep(handler StepHandler) {
	i.CustomSteps = append(i.CustomSteps, handler)
}

func (i *Interpreter) fork() *Interpreter {
	localInterp := NewBareInterpreter()
	localInterp.EnableBuiltins = i.EnableBuiltins
	localInterp.Out = i.Out
	localInterp.CustomSteps = i.CustomSteps
	localInterp.Variables = make(map[string]string)
	for k, v := range i.Variables {
		localInterp.Variables[k] = v
	}
	localInterp.Funcs = i.Funcs
	localInterp.Constants = i.Constants
	localInterp.compiledFuncs = i.compiledFuncs
	localInterp.lambdaCache = i.lambdaCache
	localInterp.RuntimeNodes = make(map[string]dsltypes.ASTNode[functions.Either[error, any]])
	for k, v := range i.RuntimeNodes {
		localInterp.RuntimeNodes[k] = v
	}
	return localInterp
}

// invalidateCompiledCache clears every cached def parse. Called whenever a def
// or constant is (re)defined, since a cached parse may have baked in a prior
// constant value or referenced a since-changed body. It deletes keys in place
// to preserve the map identity shared with any live forks.
func (i *Interpreter) invalidateCompiledCache() {
	for k := range i.compiledFuncs {
		delete(i.compiledFuncs, k)
	}
	if i.lambdaCache != nil {
		i.lambdaCache.Range(func(k, _ any) bool {
			i.lambdaCache.Delete(k)
			return true
		})
	}
}

func toListElements(val any) ([]any, bool) {
	if list, ok := val.(*collections.List[any]); ok {
		var elems []any
		for item := range list.Seq() {
			elems = append(elems, item)
		}
		return elems, true
	}
	if list, ok := val.(*collections.List[int]); ok {
		var elems []any
		for item := range list.Seq() {
			elems = append(elems, item)
		}
		return elems, true
	}
	if list, ok := val.(*collections.List[string]); ok {
		var elems []any
		for item := range list.Seq() {
			elems = append(elems, item)
		}
		return elems, true
	}
	if sl, ok := val.([]any); ok {
		return sl, true
	}
	if sl, ok := val.([]int); ok {
		var elems []any
		for _, x := range sl {
			elems = append(elems, x)
		}
		return elems, true
	}
	if p, ok := val.(functions.Pair[any, any]); ok {
		return []any{p.Fst, p.Snd}, true
	}
	if p, ok := val.(interface{ PairElements() (any, any) }); ok {
		f, s := p.PairElements()
		return []any{f, s}, true
	}
	return nil, false
}

func isIdent(s string) bool {
	if s == "" {
		return false
	}
	runes := []rune(s)
	if !unicode.IsLetter(runes[0]) && runes[0] != '_' {
		return false
	}
	for _, ch := range runes {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			return false
		}
	}
	return true
}

func (i *Interpreter) pipelineToAdapter(subPipelineStr string) evaluator.Adapter[functions.Either[error, any]] {
	subPipelineStr = strings.TrimSpace(subPipelineStr)
	var subPipeline []evaluator.PipelineStep[functions.Either[error, any]]
	if strings.HasPrefix(subPipelineStr, "let ") || (strings.HasPrefix(subPipelineStr, "(") && strings.Contains(subPipelineStr, "let ")) {
		subPipeline = i.parsePipelineSteps([]string{subPipelineStr})
	} else {
		subParts, _ := splitPipelineAware(subPipelineStr)
		subPipeline = i.parsePipelineSteps(subParts)
	}

	return func(node dsltypes.ASTNode[functions.Either[error, any]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		singleton := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) { yield(node) }
		return i.Engine.EvaluatePipeline(singleton, subPipeline)
	}
}

func (i *Interpreter) parsePipelineSteps(parts []string) []evaluator.PipelineStep[functions.Either[error, any]] {
	var pipeline []evaluator.PipelineStep[functions.Either[error, any]]

	for _, stepStr := range parts {
		stepStr = strings.TrimSpace(stepStr)
		if stepStr == "" {
			continue
		}

		// Closure application: `f`, `f a b` where f is a let-bound closure or a
		// lambda-bodied def (see closures.go). Must precede runtime-node yielding.
		if closureStep, ok := i.closureStep(stepStr); ok {
			pipeline = append(pipeline, closureStep)
			continue
		}

		// Runtime Variable Yielding
		if rtNode, exists := i.RuntimeNodes[stepStr]; exists {
			pipeline = append(pipeline, evaluator.StepYieldRuntimeNode(rtNode))
			continue
		}

		// Variable Substitution
		for k, v := range i.Variables {
			if _, isRt := i.RuntimeNodes[k]; isRt {
				continue
			}
			// Naive word replacement for variables
			words := strings.Split(stepStr, " ")
			for idx, w := range words {
				if w == k {
					words[idx] = v
				}
			}
			stepStr = strings.Join(words, " ")
		}

		// Constant Substitution
		for k, v := range i.Constants {
			if !scalarValue(v) {
				continue
			}
			words := strings.Split(stepStr, " ")
			for idx, w := range words {
				if w == k {
					words[idx] = fmt.Sprintf("%v", v)
				}
			}
			stepStr = strings.Join(words, " ")
		}

		// Bare lambda in pipeline position: `v |> \x -> expr` applies the lambda
		// to each incoming value (identical to mapping it). This also makes
		// lambda-bodied functions first-class and callable — e.g.
		// `def inc = \x -> x + 1` then `5 |> inc` — because the stored body
		// re-enters here as a bare lambda step and is applied element-wise.
		if (strings.HasPrefix(stepStr, "\\") || strings.HasPrefix(stepStr, "(\\")) && strings.Contains(stepStr, "->") {
			stepStr = "map " + stepStr
		}

		// Function Substitution (Lazy Call)
		if _, exists := i.Funcs[stepStr]; exists {
			funcName := stepStr
			lazyStep := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
				return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
					// Compile the function body lazily at runtime. When the active
					// scope carries no local Variables/RuntimeNodes, no value-
					// dependent substitution can occur, so the parse is purely
					// structural and can be cached by function name — the hot path
					// for pure recursive defs. Scoped calls skip the cache and
					// re-parse exactly as before, preserving semantics.
					cacheable := len(i.Variables) == 0 && len(i.RuntimeNodes) == 0
					var compiledSteps []evaluator.PipelineStep[functions.Either[error, any]]
					if cacheable {
						compiledSteps = i.compiledFuncs[funcName]
					}
					if compiledSteps == nil {
						funcPipeline := i.Funcs[funcName]
						funcParts, _ := splitPipelineAware(funcPipeline)
						compiledSteps = i.parsePipelineSteps(funcParts)
						if cacheable {
							i.compiledFuncs[funcName] = compiledSteps
						}
					}

					// Evaluate over the incoming stream, reusing the interpreter's
					// engine instead of allocating a fresh one per call.
					outStream := i.Engine.EvaluatePipeline(stream, compiledSteps)
					for node := range outStream {
						if !yield(node) {
							return
						}
					}
				}
			}
			pipeline = append(pipeline, lazyStep)
			continue
		}

		// Pluggable Third-Party Extension Step Check
		customHandled := false
		for _, handler := range i.CustomSteps {
			if customStep, ok := handler(stepStr, i); ok {
				pipeline = append(pipeline, customStep)
				customHandled = true
				break
			}
		}
		if customHandled {
			continue
		}

		// Let Expression Binding as Pipeline Step: let x = ... in ... or (let x = ... in ...)
		stepTrimmed := stepStr
		if strings.HasPrefix(stepTrimmed, "(") && strings.HasSuffix(stepTrimmed, ")") {
			innerP := strings.TrimSpace(stepTrimmed[1 : len(stepTrimmed)-1])
			if strings.HasPrefix(innerP, "let ") {
				stepTrimmed = innerP
			}
		}
		if strings.HasPrefix(stepTrimmed, "let ") {
			parts := strings.SplitN(stepTrimmed[4:], " in ", 2)
			if len(parts) == 2 {
				defParts := strings.SplitN(parts[0], "=", 2)
				if len(defParts) == 2 {
					varName := strings.TrimSpace(defParts[0])
					bindingPipelineStr := strings.TrimSpace(defParts[1])
					bodyPipelineStr := strings.TrimSpace(parts[1])

					letStep := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
						return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
							for node := range stream {
								localInterp := i.fork()
								bindAdapter := localInterp.bindingAdapter(bindingPipelineStr)
								var boundVal any
								found := false
								for bNode := range bindAdapter(node) {
									if bNode.Value.IsRight() {
										val, _ := bNode.Value.UnwrapRight()
										boundVal = val
										found = true
										break
									}
								}
								if found {
									localInterp.RuntimeNodes[varName] = dsltypes.NewASTNode(functions.Right[error, any](boundVal), node.DSLType)
								}
								bodyAdapter := localInterp.pipelineToAdapter(bodyPipelineStr)
								for res := range bodyAdapter(node) {
									if !yield(res) {
										return
									}
								}
							}
						}
					}
					pipeline = append(pipeline, letStep)
					continue
				}
			}
		}

		// List Comprehension or List Literal Yielding
		if strings.HasPrefix(stepStr, "[") && strings.HasSuffix(stepStr, "]") && !strings.Contains(stepStr, " *** ") {
			rawExpr := stepStr
			listStep := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
				return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
					lookup := func(k string) (any, bool) {
						if rt, ok := i.RuntimeNodes[k]; ok {
							if rt.Value.IsRight() {
								val, _ := rt.Value.UnwrapRight()
								return val, true
							}
						}
						if v, ok := i.Variables[k]; ok {
							return v, true
						}
						if v, ok := i.Constants[k]; ok {
							return v, true
						}
						return i.exprFallback(k)
					}
					if val, err := parser.EvalInfixString(rawExpr, lookup); err == nil {
						wrapped := functions.Right[error, any](val)
						flag := i.Parser.GetFlag("List")
						if !yield(dsltypes.NewASTNode(wrapped, flag)) {
							return
						}
					}
				}
			}
			pipeline = append(pipeline, listStep)
			continue
		}

		// Infix Arithmetic / Logic Expression Yielding
		if !strings.HasPrefix(stepStr, "echo ") && !strings.HasPrefix(stepStr, "assert ") && !strings.HasPrefix(stepStr, "expect ") && !strings.HasPrefix(stepStr, "uname") && !strings.HasPrefix(stepStr, "ls") && !strings.HasPrefix(stepStr, "date") && !strings.HasPrefix(stepStr, "whoami") {
			lookup := func(k string) (any, bool) {
				if rt, ok := i.RuntimeNodes[k]; ok {
					if rt.Value.IsRight() {
						val, _ := rt.Value.UnwrapRight()
						return val, true
					}
				}
				if v, ok := i.Variables[k]; ok {
					return v, true
				}
				if v, ok := i.Constants[k]; ok {
					return v, true
				}
				return i.exprFallback(k)
			}
			// `- 2` is the operator section "subtract 2", not the expression "minus two" (which
			// parses since unary minus exists): a step that opens with an operator and a space is
			// always a section, handled below.
			if !strings.HasPrefix(stepStr, "- ") && (strings.ContainsAny(stepStr, "+-*/%&|<>= ") || (len(stepStr) > 0 && unicode.IsDigit(rune(stepStr[0])))) {
				// A step that merely evaluates to a function (a partial application such
				// as `map (\x -> …)`) is a pipeline step, not an expression value.
				if probe, err := parser.EvalInfixString(stepStr, lookup); err == nil && !functions.IsCallable(probe) {
					rawStep := stepStr
					infixStep := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
						return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
							v, err := parser.EvalInfixString(rawStep, lookup)
							if err == nil {
								wrapped := functions.Right[error, any](v)
								flag := i.Parser.GetFlag("Int")
								yield(dsltypes.NewASTNode(wrapped, flag))
							}
						}
					}
					pipeline = append(pipeline, infixStep)
					continue
				}
			}
		}

		if strings.HasPrefix(stepStr, "take") {
			for res := range parser.ParseTake()(stepStr) {
				pipeline = append(pipeline, evaluator.StepTake[functions.Either[error, any]](res.Fst))
				break
			}
		} else if strings.HasPrefix(stepStr, "drop") {
			for res := range parser.ParseDrop()(stepStr) {
				pipeline = append(pipeline, evaluator.StepDrop[functions.Either[error, any]](res.Fst))
				break
			}
		} else if strings.HasPrefix(stepStr, "filter_less ") {
			varName := strings.TrimSpace(stepStr[12:])
			if pivotNode, ok := i.RuntimeNodes[varName]; ok {
				pipeline = append(pipeline, evaluator.StepFilterLess(pivotNode))
			}
		} else if strings.HasPrefix(stepStr, "filter_greater ") {
			varName := strings.TrimSpace(stepStr[15:])
			if pivotNode, ok := i.RuntimeNodes[varName]; ok {
				pipeline = append(pipeline, evaluator.StepFilterGreaterEqual(pivotNode))
			}
		} else if strings.HasPrefix(stepStr, "filter") || strings.HasPrefix(stepStr, "reject") {
			// `reject p` keeps the complement of `filter p`; both share this branch
			// ("filter" and "reject" are both 6 characters) with a negate flag.
			negate := strings.HasPrefix(stepStr, "reject")
			rawArg := strings.TrimSpace(stepStr[6:])
			if c, ok, err := i.stepClosure(rawArg, 1); err != nil {
				pipeline = append(pipeline, i.failStep(fmt.Errorf("%s: %w", stepStr[:6], err)))
			} else if ok {
				pipeline = append(pipeline, filterStepOver(closurePredicate(c, negate)))
			} else if isLambdaText(rawArg) {
				pipeline = append(pipeline, i.failStep(fmt.Errorf("%s: cannot use %s as a function: %v", stepStr[:6], rawArg, lambdaParseCause(rawArg))))
			} else {
				keyword := strings.Trim(rawArg, "\"")
				predicate := func(val any) bool {
					match := false
					if strVal, ok := val.(string); ok {
						match = strings.Contains(strVal, keyword)
					}
					if negate {
						return !match
					}
					return match
				}
				pipeline = append(pipeline, evaluator.StepFilter(predicate))
			}
		} else if strings.HasPrefix(stepStr, "expect") {
			for res := range parser.ParseExpect()(stepStr) {
				keyword := res.Fst.Fst
				errMsg := res.Fst.Snd
				predicate := func(val any) bool {
					if strVal, ok := val.(string); ok {
						return strings.Contains(strVal, keyword)
					}
					return false
				}
				pipeline = append(pipeline, evaluator.StepExpect(predicate, fmt.Errorf("%s", errMsg)))
				break
			}
		} else if stepStr == "empty" {
			pipeline = append(pipeline, evaluator.StepEmpty())
		} else if stepStr == "distinct" {
			pipeline = append(pipeline, evaluator.StepDistinct())
		} else if strings.HasPrefix(stepStr, "fold") {
			op := strings.TrimSpace(stepStr[4:])
			if foldStep, ok := i.foldClosureStep(op); ok {
				pipeline = append(pipeline, foldStep)
			} else if isLambdaText(op) {
				pipeline = append(pipeline, i.failStep(fmt.Errorf("fold: expected a two-argument function and an optional seed (fold (\\acc x -> …) seed), got: %s", op)))
			} else {
				pipeline = append(pipeline, evaluator.StepFold(op))
			}
		} else if strings.HasPrefix(stepStr, "memoize") {
			cmd := strings.TrimSpace(stepStr[7:])
			adapter := evaluator.EitherT(interactive.GenericExecAdapter(cmd, 0))
			pipeline = append(pipeline, evaluator.StepMemoize(adapter))
		} else if strings.HasPrefix(stepStr, "zip") {
			for res := range parser.ParseZip()(stepStr) {
				cmd := res.Fst
				secAdapter := evaluator.EitherT(interactive.GenericExecAdapter(cmd, 0))
				pipeline = append(pipeline, evaluator.StepZip(secAdapter))
				break
			}
		} else if strings.HasPrefix(stepStr, "format_tuple") {
			formatStr := strings.TrimSpace(stepStr[12:])
			pipeline = append(pipeline, evaluator.StepFormatTuple(strings.Trim(formatStr, "\"")))
		} else if branches, ok := plusBranches(stepStr); ok {
			// (a <+> b <+> …): merge several sub-pipelines into one stream (ArrowPlus).
			adapters := make([]evaluator.Adapter[functions.Either[error, any]], len(branches))
			for idx, b := range branches {
				adapters[idx] = i.pipelineToAdapter(b)
			}
			pipeline = append(pipeline, evaluator.StepPlus(adapters...))
		} else if strings.HasPrefix(stepStr, "(") && strings.Contains(stepStr, "&&&") {
			for res := range parser.ParseFanOut()(stepStr) {
				leftAdapter := i.pipelineToAdapter(res.Fst.Fst)
				rightAdapter := i.pipelineToAdapter(res.Fst.Snd)
				pipeline = append(pipeline, evaluator.StepFanOut(leftAdapter, rightAdapter))
				break
			}
		} else if strings.HasPrefix(stepStr, "[") && strings.Contains(stepStr, "***") {
			for res := range parser.ParseSplit()(stepStr) {
				leftAdapter := i.pipelineToAdapter(res.Fst.Fst)
				rightAdapter := i.pipelineToAdapter(res.Fst.Snd)
				pipeline = append(pipeline, evaluator.StepSplit(leftAdapter, rightAdapter))
				break
			}
		} else if strings.HasPrefix(stepStr, "parse_record") {
			fieldsStr := strings.TrimSpace(stepStr[12:])
			fields := strings.Fields(fieldsStr)
			adapter := evaluator.EitherT(interactive.ParseColumnsAdapter(fields, 0))
			pipeline = append(pipeline, evaluator.StepAdapter(adapter))
		} else if strings.HasPrefix(stepStr, "focus") {
			field := strings.TrimSpace(stepStr[5:])
			pipeline = append(pipeline, evaluator.StepFocus(field))
		} else if strings.HasPrefix(stepStr, "update") {
			args := strings.Fields(strings.TrimSpace(stepStr[6:]))
			if len(args) >= 2 {
				pipeline = append(pipeline, evaluator.StepUpdate(args[0], strings.Join(args[1:], " ")))
			}
		} else if strings.HasPrefix(stepStr, "emit ") {
			parts := strings.Split(strings.TrimSpace(stepStr[5:]), ",")
			// The tokens are literals: a canonical number (`10`, `-5`, `1.5`) is that number,
			// anything else (`alice`, `007`, `kind:Pod`) stays text.
			var values []any
			for _, p := range parts {
				values = append(values, functions.LiteralToken(strings.TrimSpace(p)))
			}
			pipeline = append(pipeline, evaluator.StepEmitValues(values...))
		} else if strings.HasPrefix(stepStr, "walk_fs ") {
			pathStr := strings.TrimSpace(stepStr[8:])
			adapter := adapters.WalkFSAdapter(pathStr)
			pipeline = append(pipeline, evaluator.StepAdapter(adapter))
		} else if strings.HasPrefix(stepStr, "assert_unordered ") {
			expected := strings.TrimSpace(stepStr[17:])
			expected = strings.ReplaceAll(expected, ",", "\n")
			pipeline = append(pipeline, func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
				return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
					var actualLines []string
					var lastFlag dsltypes.BitFlag = 1
					for node := range stream {
						lastFlag = node.DSLType
						if node.Value.IsRight() {
							val, _ := node.Value.UnwrapRight()
							actualLines = append(actualLines, fmt.Sprintf("%v", val))
						} else if node.Value.IsLeft() {
							err, _ := node.Value.UnwrapLeft()
							actualLines = append(actualLines, err.Error())
						}
					}
					pass, diff := evaluator.CompareExampleOutput(expected, strings.Join(actualLines, "\n"), true)
					if pass {
						yield(dsltypes.NewASTNode(functions.Right[error, any](any("PASS")), lastFlag))
					} else {
						yield(evaluator.InjectError[any](fmt.Errorf("assert_unordered failure: %s", diff), lastFlag))
					}
				}
			})
		} else if strings.HasPrefix(stepStr, "example_test ") {
			expected := strings.TrimSpace(stepStr[13:])
			expected = strings.ReplaceAll(expected, ",", "\n")
			pipeline = append(pipeline, func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
				return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
					var actualLines []string
					var lastFlag dsltypes.BitFlag = 1
					for node := range stream {
						lastFlag = node.DSLType
						if node.Value.IsRight() {
							val, _ := node.Value.UnwrapRight()
							actualLines = append(actualLines, fmt.Sprintf("%v", val))
						} else if node.Value.IsLeft() {
							err, _ := node.Value.UnwrapLeft()
							actualLines = append(actualLines, err.Error())
						}
					}
					pass, diff := evaluator.CompareExampleOutput(expected, strings.Join(actualLines, "\n"), false)
					if pass {
						yield(dsltypes.NewASTNode(functions.Right[error, any](any("PASS")), lastFlag))
					} else {
						yield(evaluator.InjectError[any](fmt.Errorf("example_test failure: %s", diff), lastFlag))
					}
				}
			})
		} else if strings.HasPrefix(stepStr, "mock_stream ") {
			countStr := strings.TrimSpace(stepStr[12:])
			if count, err := strconv.Atoi(countStr); err == nil && count >= 0 {
				pipeline = append(pipeline, func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
					return evaluator.GenerateStream(func(i int) any {
						return i
					}, count, dsltypes.BitFlag(1))
				})
			}
		} else if strings.HasPrefix(stepStr, "mock_users ") {
			countStr := strings.TrimSpace(stepStr[11:])
			if count, err := strconv.Atoi(countStr); err == nil && count >= 0 {
				pipeline = append(pipeline, func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
					return evaluator.GenerateMockUsers(count, dsltypes.BitFlag(1))
				})
			}
		} else if strings.HasPrefix(stepStr, "assert ") {
			expected := strings.TrimSpace(stepStr[7:])
			pipeline = append(pipeline, evaluator.StepAssert(expected))
		} else if strings.HasPrefix(stepStr, "debug_print ") {
			prefix := strings.TrimSpace(stepStr[12:])
			pipeline = append(pipeline, evaluator.StepDebugPrint(prefix))
		} else if strings.HasPrefix(stepStr, "visualize") {
			pipeline = append(pipeline, visualizer.StepVisualize(strings.TrimSpace(stepStr[9:])))
		} else if strings.HasPrefix(stepStr, "sort") {
			pipeline = append(pipeline, evaluator.StepSort(evaluator.DefaultLess))
		} else if strings.HasPrefix(stepStr, "to_set") {
			pipeline = append(pipeline, evaluator.StepToSet())
		} else if strings.HasPrefix(stepStr, "to_list") {
			pipeline = append(pipeline, evaluator.StepToList())
		} else if strings.HasPrefix(stepStr, "to_chunk") {
			pipeline = append(pipeline, evaluator.StepToChunk())
		} else if stepStr == "pair_swap" {
			pairSwapStep := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
				return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
					for node := range stream {
						if node.Value.IsLeft() {
							if !yield(node) {
								return
							}
							continue
						}
						val, _ := node.Value.UnwrapRight()
						if p, ok := val.(interface{ PairElements() (any, any) }); ok {
							f, s := p.PairElements()
							newNode := dsltypes.NewASTNode(functions.Right[error, any](any(functions.Pair[any, any]{Fst: s, Snd: f})), node.DSLType)
							if !yield(newNode) {
								return
							}
							continue
						}
						if l, ok := val.(*collections.List[any]); ok && l.Len() == 2 {
							f, _ := l.Head()
							t, _ := l.Tail()
							s, _ := t.Head()
							swapped := collections.Cons(s, collections.Cons(f, collections.Nil[any]()))
							newNode := dsltypes.NewASTNode(functions.Right[error, any](any(swapped)), node.DSLType)
							if !yield(newNode) {
								return
							}
							continue
						}
						if sl, ok := functions.ToSliceAny(val); ok && len(sl) == 2 {
							swapped := []any{sl[1], sl[0]}
							newNode := dsltypes.NewASTNode(functions.Right[error, any](any(swapped)), node.DSLType)
							if !yield(newNode) {
								return
							}
							continue
						}
						yield(node)
					}
				}
			}
			pipeline = append(pipeline, pairSwapStep)
		} else if stepStr == "tree_size" {
			treeSizeStep := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
				return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
					for node := range stream {
						if node.Value.IsLeft() {
							if !yield(node) {
								return
							}
							continue
						}
						val, _ := node.Value.UnwrapRight()
						if t, ok := val.(interface{ Size() int }); ok {
							newNode := dsltypes.NewASTNode(functions.Right[error, any](any(t.Size())), node.DSLType)
							if !yield(newNode) {
								return
							}
						} else {
							yield(node)
						}
					}
				}
			}
			pipeline = append(pipeline, treeSizeStep)
		} else if stepStr == "tree_depth" {
			treeDepthStep := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
				return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
					for node := range stream {
						if node.Value.IsLeft() {
							if !yield(node) {
								return
							}
							continue
						}
						val, _ := node.Value.UnwrapRight()
						if t, ok := val.(interface{ Depth() int }); ok {
							newNode := dsltypes.NewASTNode(functions.Right[error, any](any(t.Depth())), node.DSLType)
							if !yield(newNode) {
								return
							}
						} else {
							yield(node)
						}
					}
				}
			}
			pipeline = append(pipeline, treeDepthStep)
		} else if strings.HasPrefix(stepStr, "set_contains ") {
			item := strings.Trim(strings.TrimSpace(stepStr[13:]), "\"")
			pipeline = append(pipeline, evaluator.StepSetContains(item))
		} else if strings.HasPrefix(stepStr, "set_add ") {
			item := strings.Trim(strings.TrimSpace(stepStr[8:]), "\"")
			pipeline = append(pipeline, evaluator.StepSetAdd(item))
		} else if strings.HasPrefix(stepStr, "set_remove ") {
			item := strings.Trim(strings.TrimSpace(stepStr[11:]), "\"")
			pipeline = append(pipeline, evaluator.StepSetRemove(item))
		} else if strings.HasPrefix(stepStr, "set_union") {
			operandStr := strings.TrimSpace(stepStr[9:])
			if operandStr == "" {
				pipeline = append(pipeline, evaluator.StepSetUnion())
			} else {
				lookup := func(k string) (any, bool) {
					if rt, ok := i.RuntimeNodes[k]; ok {
						v, _ := rt.Value.UnwrapRight()
						return v, true
					}
					if v, ok := i.Constants[k]; ok {
						return v, true
					}
					if v, ok := i.Variables[k]; ok {
						return v, true
					}
					return i.exprFallback(k)
				}
				if evaled, err := parser.EvalInfixString(operandStr, lookup); err == nil {
					pipeline = append(pipeline, evaluator.StepSetUnion(evaled))
				} else if rt, ok := i.RuntimeNodes[operandStr]; ok {
					v, _ := rt.Value.UnwrapRight()
					pipeline = append(pipeline, evaluator.StepSetUnion(v))
				} else {
					pipeline = append(pipeline, evaluator.StepSetUnion(operandStr))
				}
			}
		} else if strings.HasPrefix(stepStr, "subset_of ") {
			// subject |> subset_of <operand>  →  bool (subject ⊆ operand); operand is
			// a let-bound collection/Record resolved through the standard env lookup.
			operandStr := strings.TrimSpace(stepStr[len("subset_of "):])
			pipeline = append(pipeline, evaluator.StepSubsetOf(i.resolveGraphOperand(operandStr)))
		} else if strings.HasPrefix(stepStr, "set_intersect ") {
			operandStr := strings.TrimSpace(stepStr[14:])
			lookup := func(k string) (any, bool) {
				if rt, ok := i.RuntimeNodes[k]; ok {
					v, _ := rt.Value.UnwrapRight()
					return v, true
				}
				if v, ok := i.Constants[k]; ok {
					return v, true
				}
				if v, ok := i.Variables[k]; ok {
					return v, true
				}
				return i.exprFallback(k)
			}
			if evaled, err := parser.EvalInfixString(operandStr, lookup); err == nil {
				pipeline = append(pipeline, evaluator.StepSetIntersect(evaled))
			} else if rt, ok := i.RuntimeNodes[operandStr]; ok {
				v, _ := rt.Value.UnwrapRight()
				pipeline = append(pipeline, evaluator.StepSetIntersect(v))
			} else {
				pipeline = append(pipeline, evaluator.StepSetIntersect(operandStr))
			}
		} else if strings.HasPrefix(stepStr, "set_diff ") {
			operandStr := strings.TrimSpace(stepStr[9:])
			lookup := func(k string) (any, bool) {
				if rt, ok := i.RuntimeNodes[k]; ok {
					v, _ := rt.Value.UnwrapRight()
					return v, true
				}
				if v, ok := i.Constants[k]; ok {
					return v, true
				}
				if v, ok := i.Variables[k]; ok {
					return v, true
				}
				return i.exprFallback(k)
			}
			if evaled, err := parser.EvalInfixString(operandStr, lookup); err == nil {
				pipeline = append(pipeline, evaluator.StepSetDiff(evaled))
			} else if rt, ok := i.RuntimeNodes[operandStr]; ok {
				v, _ := rt.Value.UnwrapRight()
				pipeline = append(pipeline, evaluator.StepSetDiff(v))
			} else {
				pipeline = append(pipeline, evaluator.StepSetDiff(operandStr))
			}
		} else if strings.HasPrefix(stepStr, "hamt_merge ") {
			operandStr := strings.TrimSpace(stepStr[11:])
			lookup := func(k string) (any, bool) {
				if rt, ok := i.RuntimeNodes[k]; ok {
					v, _ := rt.Value.UnwrapRight()
					return v, true
				}
				if v, ok := i.Constants[k]; ok {
					return v, true
				}
				if v, ok := i.Variables[k]; ok {
					return v, true
				}
				return i.exprFallback(k)
			}
			if evaled, err := parser.EvalInfixString(operandStr, lookup); err == nil {
				pipeline = append(pipeline, evaluator.StepHAMTMerge(evaled))
			} else if rt, ok := i.RuntimeNodes[operandStr]; ok {
				v, _ := rt.Value.UnwrapRight()
				pipeline = append(pipeline, evaluator.StepHAMTMerge(v))
			} else {
				pipeline = append(pipeline, evaluator.StepHAMTMerge(operandStr))
			}
		} else if strings.HasPrefix(stepStr, "hamt_intersect ") {
			operandStr := strings.TrimSpace(stepStr[15:])
			lookup := func(k string) (any, bool) {
				if rt, ok := i.RuntimeNodes[k]; ok {
					v, _ := rt.Value.UnwrapRight()
					return v, true
				}
				if v, ok := i.Constants[k]; ok {
					return v, true
				}
				if v, ok := i.Variables[k]; ok {
					return v, true
				}
				return i.exprFallback(k)
			}
			// StepSetIntersect dispatches by reflection to the value's Intersect
			// method, so it applies to HAMT (keyed intersection) as well as Set.
			if evaled, err := parser.EvalInfixString(operandStr, lookup); err == nil {
				pipeline = append(pipeline, evaluator.StepSetIntersect(evaled))
			} else if rt, ok := i.RuntimeNodes[operandStr]; ok {
				v, _ := rt.Value.UnwrapRight()
				pipeline = append(pipeline, evaluator.StepSetIntersect(v))
			} else {
				pipeline = append(pipeline, evaluator.StepSetIntersect(operandStr))
			}
		} else if strings.HasPrefix(stepStr, "hamt_diff ") {
			operandStr := strings.TrimSpace(stepStr[10:])
			lookup := func(k string) (any, bool) {
				if rt, ok := i.RuntimeNodes[k]; ok {
					v, _ := rt.Value.UnwrapRight()
					return v, true
				}
				if v, ok := i.Constants[k]; ok {
					return v, true
				}
				if v, ok := i.Variables[k]; ok {
					return v, true
				}
				return i.exprFallback(k)
			}
			// StepSetDiff dispatches by reflection to the value's Diff method,
			// so it applies to HAMT (keyed difference) as well as Set.
			if evaled, err := parser.EvalInfixString(operandStr, lookup); err == nil {
				pipeline = append(pipeline, evaluator.StepSetDiff(evaled))
			} else if rt, ok := i.RuntimeNodes[operandStr]; ok {
				v, _ := rt.Value.UnwrapRight()
				pipeline = append(pipeline, evaluator.StepSetDiff(v))
			} else {
				pipeline = append(pipeline, evaluator.StepSetDiff(operandStr))
			}
		} else if stepStr == "set_elements" {
			pipeline = append(pipeline, evaluator.StepSetElements())
		} else if strings.HasPrefix(stepStr, "write_file") {
			filename := strings.TrimSpace(stepStr[10:])
			adapter := adapters.WriteFile(filename, false)
			pipeline = append(pipeline, evaluator.StepAdapter(adapter))
		} else if strings.HasPrefix(stepStr, "append_file") {
			filename := strings.TrimSpace(stepStr[11:])
			adapter := adapters.WriteFile(filename, true)
			pipeline = append(pipeline, evaluator.StepAdapter(adapter))
		} else if stepStr == "route" {
			pipeline = append(pipeline, evaluator.StepRoute(i.Engine))
		} else if stepStr == "build_bst" {
			pipeline = append(pipeline, evaluator.StepBuildBST())
		} else if stepStr == "build_btree" {
			pipeline = append(pipeline, evaluator.StepBuildBTree())
		} else if stepStr == "build_rbtree" {
			pipeline = append(pipeline, evaluator.StepBuildRBTree())
		} else if strings.HasPrefix(stepStr, "math_add") {
			amount := strings.TrimSpace(stepStr[9:])
			pipeline = append(pipeline, evaluator.StepMathAdd(amount))
		} else if strings.HasPrefix(stepStr, "math_sub") {
			amount := strings.TrimSpace(stepStr[9:])
			pipeline = append(pipeline, evaluator.StepMathSub(amount))
		} else if strings.HasPrefix(stepStr, "math_mul") {
			amount := strings.TrimSpace(stepStr[9:])
			pipeline = append(pipeline, evaluator.StepMathMul(amount))
		} else if strings.HasPrefix(stepStr, "math_div") {
			amount := strings.TrimSpace(stepStr[9:])
			pipeline = append(pipeline, evaluator.StepMathDiv(amount))
		} else if strings.HasPrefix(stepStr, "math_idiv") {
			amount := strings.TrimSpace(stepStr[10:])
			pipeline = append(pipeline, evaluator.StepMathIDiv(amount))
		} else if strings.HasPrefix(stepStr, "math_pow") {
			amount := strings.TrimSpace(stepStr[9:])
			pipeline = append(pipeline, evaluator.StepMathPow(amount))
		} else if strings.HasPrefix(stepStr, "math_mod") {
			amount := strings.TrimSpace(stepStr[9:])
			pipeline = append(pipeline, evaluator.StepMathMod(amount))
		} else if strings.HasPrefix(stepStr, "+ ") {
			pipeline = append(pipeline, evaluator.StepBinaryOp(functions.OpAdd, strings.TrimSpace(stepStr[2:])))
		} else if strings.HasPrefix(stepStr, "- ") {
			pipeline = append(pipeline, evaluator.StepBinaryOp(functions.OpSub, strings.TrimSpace(stepStr[2:])))
		} else if strings.HasPrefix(stepStr, "* ") {
			pipeline = append(pipeline, evaluator.StepBinaryOp(functions.OpMul, strings.TrimSpace(stepStr[2:])))
		} else if strings.HasPrefix(stepStr, "** ") {
			pipeline = append(pipeline, evaluator.StepBinaryOp(functions.OpPow, strings.TrimSpace(stepStr[3:])))
		} else if strings.HasPrefix(stepStr, "// ") {
			pipeline = append(pipeline, evaluator.StepBinaryOp(functions.OpIDiv, strings.TrimSpace(stepStr[3:])))
		} else if strings.HasPrefix(stepStr, "/ ") {
			pipeline = append(pipeline, evaluator.StepBinaryOp(functions.OpDiv, strings.TrimSpace(stepStr[2:])))
		} else if strings.HasPrefix(stepStr, "% ") {
			pipeline = append(pipeline, evaluator.StepBinaryOp(functions.OpMod, strings.TrimSpace(stepStr[2:])))
		} else if strings.HasPrefix(stepStr, "++ ") {
			pipeline = append(pipeline, evaluator.StepBinaryOp(functions.OpConcat, strings.TrimSpace(stepStr[3:])))
		} else if stepStr == "build_array" {
			pipeline = append(pipeline, evaluator.StepBuildArray())
		} else if stepStr == "collect_list" {
			pipeline = append(pipeline, evaluator.StepCollectList())
		} else if strings.HasPrefix(stepStr, "build_tree") {
			rootName := strings.Trim(strings.TrimSpace(stepStr[10:]), "\"")
			pipeline = append(pipeline, evaluator.StepBuildTree(rootName))
		} else if stepStr == "into_zipper" || stepStr == "down" || stepStr == "up" || stepStr == "extract" || stepStr == "left" || stepStr == "right" {
			pipeline = append(pipeline, evaluator.StepZipper(stepStr))
		} else if stepStr == "to_hamt" {
			pipeline = append(pipeline, evaluator.StepToHAMT())
		} else if strings.HasPrefix(stepStr, "hamt_get ") {
			key := strings.Trim(strings.TrimSpace(stepStr[9:]), "\"")
			pipeline = append(pipeline, evaluator.StepHAMTGet(key))
		} else if strings.HasPrefix(stepStr, "hamt_path ") {
			path := strings.Trim(strings.TrimSpace(stepStr[10:]), "\"")
			pipeline = append(pipeline, evaluator.StepHAMTPath(path))
		} else if strings.HasPrefix(stepStr, "hamt_set ") {
			parts := strings.Fields(strings.TrimSpace(stepStr[9:]))
			if len(parts) >= 2 {
				k := strings.Trim(parts[0], "\"")
				v := strings.Trim(strings.Join(parts[1:], " "), "\"")
				pipeline = append(pipeline, evaluator.StepHAMTSet(k, v))
			}
		} else if strings.HasPrefix(stepStr, "hamt_del ") {
			key := strings.Trim(strings.TrimSpace(stepStr[9:]), "\"")
			pipeline = append(pipeline, evaluator.StepHAMTDel(key))
		} else if stepStr == "hamt_keys" {
			pipeline = append(pipeline, evaluator.StepHAMTKeys())
		} else if stepStr == "hamt_values" {
			pipeline = append(pipeline, evaluator.StepHAMTValues())
		} else if stepStr == "hamt_pairs" {
			pipeline = append(pipeline, evaluator.StepHAMTPairs())
		} else if stepStr == "frequencies" {
			pipeline = append(pipeline, evaluator.StepFrequencies())
		} else if step, ok := i.buildGraphLambdaStep(stepStr); ok {
			pipeline = append(pipeline, step)
		} else if step, ok := buildGraphStep(stepStr); ok {
			pipeline = append(pipeline, step)
		} else if strings.HasPrefix(stepStr, "chunk ") {
			sizeStr := strings.TrimSpace(stepStr[6:])
			if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 {
				pipeline = append(pipeline, evaluator.StepChunk(size))
			}
		} else if stepStr == "unchunk" || stepStr == "flatten_chunk" {
			pipeline = append(pipeline, evaluator.StepUnchunk())
		} else if strings.HasPrefix(stepStr, "checkpoint ") {
			args := strings.Fields(strings.TrimSpace(stepStr[11:]))
			id := "default"
			var interval uint64 = 10
			if len(args) > 0 {
				id = strings.Trim(args[0], "\"")
			}
			if len(args) > 1 {
				if iv, err := strconv.ParseUint(args[1], 10, 64); err == nil {
					interval = iv
				}
			}
			pipeline = append(pipeline, evaluator.StepCheckpoint(context.Background(), evaluator.NewMemoryCheckpointStore(), id, interval))
		} else if stepStr == "telemetry" {
			var count uint64
			pipeline = append(pipeline, evaluator.StepTelemetry(&count))
		} else if strings.HasPrefix(stepStr, "ebnf_parse ") {
			// ebnf_parse <startRule> <grammarText>
			parts := strings.SplitN(strings.TrimSpace(stepStr[11:]), " ", 2)
			if len(parts) == 2 {
				startRule := strings.Trim(parts[0], "\"")
				grammarText := strings.Trim(parts[1], "\"")
				pipeline = append(pipeline, evaluator.StepEBNFParse(startRule, grammarText))
			}
		} else if strings.HasPrefix(stepStr, "dial_tcp ") {
			addr := strings.Trim(strings.TrimSpace(stepStr[9:]), "\"")
			pipeline = append(pipeline, adapters.StepDial("tcp", addr, 5*time.Second))
		} else if strings.HasPrefix(stepStr, "listen_tcp ") {
			args := strings.Fields(strings.TrimSpace(stepStr[11:]))
			addr := "127.0.0.1:0"
			maxConns := 1
			if len(args) > 0 {
				addr = strings.Trim(args[0], "\"")
			}
			if len(args) > 1 {
				if mc, err := strconv.Atoi(args[1]); err == nil {
					maxConns = mc
				}
			}
			pipeline = append(pipeline, adapters.StepListenTCP("tcp", addr, maxConns))
		} else if stepStr == "conn_lines" {
			pipeline = append(pipeline, adapters.StepConnLines())
		} else if strings.HasPrefix(stepStr, "scan_words ") {
			filePath := strings.Trim(strings.TrimSpace(stepStr[11:]), "\"")
			pipeline = append(pipeline, adapters.StepScanFile(filePath, bufio.ScanWords))
		} else if strings.HasPrefix(stepStr, "scan_file ") {
			filePath := strings.Trim(strings.TrimSpace(stepStr[10:]), "\"")
			pipeline = append(pipeline, adapters.StepScanFile(filePath, nil))
		} else if stepStr == "head" {
			pipeline = append(pipeline, evaluator.StepTake[functions.Either[error, any]](1))
		} else if stepStr == "tail" {
			pipeline = append(pipeline, evaluator.StepDrop[functions.Either[error, any]](1))
		} else if strings.HasPrefix(stepStr, "seq {") && strings.HasSuffix(stepStr, "}") {
			inner := strings.TrimSpace(stepStr[5 : len(stepStr)-1])
			casesStr := splitCasesAware(inner) // this splits by | which is fine, we can use |
			var adapters []evaluator.Adapter[functions.Either[error, any]]
			for _, c := range casesStr {
				adapters = append(adapters, i.pipelineToAdapter(strings.TrimSpace(c)))
			}
			pipeline = append(pipeline, evaluator.StepSeq(adapters))
		} else if strings.HasPrefix(stepStr, "match {") && strings.HasSuffix(stepStr, "}") {
			inner := strings.TrimSpace(stepStr[7 : len(stepStr)-1])
			casesStr := splitCasesAware(inner)
			type matchCase struct {
				pattern string
				cmd     string
			}
			var cases []matchCase
			matchCaseParser := parser.ParseMatchCase()
			for _, c := range casesStr {
				c = strings.TrimSpace(c)
				if c == "" {
					continue
				}
				for res := range matchCaseParser(c) {
					cases = append(cases, matchCase{pattern: strings.TrimSpace(res.Fst.Fst), cmd: strings.TrimSpace(res.Fst.Snd)})
					break
				}
			}

			dynamicMatch := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
				return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
					for node := range stream {
						if node.Value.IsLeft() {
							if !yield(node) {
								return
							}
							continue
						}
						val, _ := node.Value.UnwrapRight()
						strVal := fmt.Sprintf("%v", val)

						for _, mc := range cases {
							pattern := mc.pattern
							cmd := mc.cmd

							// 1. Head/Tail list pattern: [head, ...tail]
							if strings.HasPrefix(pattern, "[") && strings.Contains(pattern, ", ...") && strings.HasSuffix(pattern, "]") {
								if list, isList := val.(*collections.List[any]); isList {
									if list.IsEmpty() {
										continue
									}
									parts := strings.Split(pattern[1:len(pattern)-1], ", ...")
									headVar := strings.TrimSpace(parts[0])
									tailVar := strings.TrimSpace(parts[1])

									localInterp := i.fork()
									headVal, _ := list.Head()
									tailVal, _ := list.Tail()
									localInterp.RuntimeNodes[headVar] = dsltypes.NewASTNode(functions.Right[error, any](headVal), node.DSLType)
									localInterp.RuntimeNodes[tailVar] = dsltypes.NewASTNode(functions.Right[error, any](any(tailVal)), node.DSLType)

									adapter := localInterp.pipelineToAdapter(cmd)
									for res := range adapter(node) {
										if !yield(res) {
											return
										}
									}
									break
								}
							}

							// 2. Empty list pattern: []
							if pattern == "[]" {
								if list, isList := val.(*collections.List[any]); isList && list.IsEmpty() {
									adapter := i.pipelineToAdapter(cmd)
									for res := range adapter(node) {
										if !yield(res) {
											return
										}
									}
									break
								}
							}

							// 3. Fixed-length list / tuple pattern: [p1, p2, ...] or (p1, p2, ...)
							if (strings.HasPrefix(pattern, "[") && strings.HasSuffix(pattern, "]") && !strings.Contains(pattern, ", ...")) ||
								(strings.HasPrefix(pattern, "(") && strings.HasSuffix(pattern, ")")) {
								innerPat := pattern[1 : len(pattern)-1]
								patParts := strings.Split(innerPat, ",")
								var elemPats []string
								for _, p := range patParts {
									elemPats = append(elemPats, strings.TrimSpace(p))
								}
								if elems, ok := toListElements(val); ok && len(elems) == len(elemPats) {
									matchedElems := true
									localInterp := i.fork()
									for idx, p := range elemPats {
										elem := elems[idx]
										if p == "_" {
											continue
										}
										if isIdent(p) {
											localInterp.RuntimeNodes[p] = dsltypes.NewASTNode(functions.Right[error, any](elem), node.DSLType)
										} else {
											if fmt.Sprintf("%v", elem) != p {
												matchedElems = false
												break
											}
										}
									}
									if matchedElems {
										adapter := localInterp.pipelineToAdapter(cmd)
										for res := range adapter(node) {
											if !yield(res) {
												return
											}
										}
										break
									}
								}
							}

							// 4. Wildcard pattern: _
							if pattern == "_" {
								adapter := i.pipelineToAdapter(cmd)
								for res := range adapter(node) {
									if !yield(res) {
										return
									}
								}
								break
							}

							// 5. Constructor pattern: an identifier naming a registered sum-type
							// branch (e.g. `Linux`, `Red`, `Nothing`). It discriminates on the
							// incoming value's variant tag rather than binding a variable. A
							// registered constructor whose tag does not match this value falls
							// through to the NEXT case (continue), never to the variable arm —
							// so `match { Red -> a | Green -> b }` routes correctly instead of
							// letting the first arm win.
							if isIdent(pattern) && i.Parser.IsConstructor(pattern) {
								matched := false
								if variant, ok := val.(dsltypes.Variant); ok {
									// Value was boxed as a sum-type Variant (parser path).
									matched = variant.Is(pattern)
								} else {
									// Value flows as its bare tag string (interpreter spawn
									// path): discriminate by exact tag equality, not substring.
									matched = strVal == pattern
								}
								if matched {
									adapter := i.pipelineToAdapter(cmd)
									for res := range adapter(node) {
										if !yield(res) {
											return
										}
									}
									break
								}
								// Constructor pattern that did not match: try the next case.
								continue
							}

							// 6. Variable identifier pattern: x -> cmd (binds incoming value to variable x)
							if isIdent(pattern) && pattern != "true" && pattern != "false" {
								localInterp := i.fork()
								localInterp.RuntimeNodes[pattern] = dsltypes.NewASTNode(functions.Right[error, any](val), node.DSLType)
								adapter := localInterp.pipelineToAdapter(cmd)
								for res := range adapter(node) {
									if !yield(res) {
										return
									}
								}
								break
							}

							// 7. String/Value substring matching
							if !strings.HasPrefix(pattern, "[") && !strings.HasPrefix(pattern, "(") && strings.Contains(strVal, pattern) {
								adapter := i.pipelineToAdapter(cmd)
								for res := range adapter(node) {
									if !yield(res) {
										return
									}
								}
								break
							}
						}
					}
				}
			}
			pipeline = append(pipeline, dynamicMatch)
		} else if strings.HasPrefix(stepStr, "bimap") {
			// bimap (\x -> e1) (\y -> e2): the Bifunctor map. Applies e1 to the
			// first/Left component and e2 to the second/Right component of a Pair,
			// an Either, or a 2-element list.
			rest := strings.TrimSpace(stepStr[5:])
			extract := func(s string) (string, string) {
				s = strings.TrimSpace(s)
				if strings.HasPrefix(s, "(") {
					depth := 0
					for idx, r := range s {
						if r == '(' {
							depth++
						} else if r == ')' {
							depth--
							if depth == 0 {
								return strings.TrimSpace(s[1:idx]), strings.TrimSpace(s[idx+1:])
							}
						}
					}
				}
				return s, ""
			}
			parseLam := func(l string) (string, string, bool) {
				l = strings.TrimPrefix(strings.TrimSpace(l), "\\")
				pp := strings.SplitN(l, "->", 2)
				if len(pp) != 2 {
					return "", "", false
				}
				return strings.TrimSpace(pp[0]), strings.TrimSpace(pp[1]), true
			}
			lam1, rem := extract(rest)
			lam2, _ := extract(rem)
			v1, b1, ok1 := parseLam(lam1)
			v2, b2, ok2 := parseLam(lam2)
			if ok1 && ok2 {
				evalOne := func(vn, body string, item any) any {
					lookup := func(name string) (any, bool) {
						if name == vn {
							return item, true
						}
						if rt, ok := i.RuntimeNodes[name]; ok {
							vv, _ := rt.Value.UnwrapRight()
							return vv, true
						}
						if c, ok := i.Constants[name]; ok {
							return c, true
						}
						if vv, ok := i.Variables[name]; ok {
							return vv, true
						}
						return nil, false
					}
					if pe, err := parser.ParseInfixExpr(body); err == nil {
						if rv, e := pe.Eval(lookup); e == nil {
							return rv
						}
					}
					return item
				}
				bimapStep := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
					return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
						for node := range stream {
							if node.Value.IsLeft() {
								if !yield(node) {
									return
								}
								continue
							}
							val, _ := node.Value.UnwrapRight()
							if p, ok := val.(interface{ PairElements() (any, any) }); ok {
								a, b := p.PairElements()
								np := functions.Pair[any, any]{Fst: evalOne(v1, b1, a), Snd: evalOne(v2, b2, b)}
								if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(np)), node.DSLType)) {
									return
								}
								continue
							}
							if e, ok := val.(interface {
								EitherValue() (any, any, bool)
							}); ok {
								lv, rv, isLeft := e.EitherValue()
								var res any
								if isLeft {
									res = functions.Left[any, any](evalOne(v1, b1, lv))
								} else {
									res = functions.Right[any, any](evalOne(v2, b2, rv))
								}
								if !yield(dsltypes.NewASTNode(functions.Right[error, any](res), node.DSLType)) {
									return
								}
								continue
							}
							if elems, ok := toListElements(val); ok && len(elems) == 2 {
								np := []any{evalOne(v1, b1, elems[0]), evalOne(v2, b2, elems[1])}
								if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(np)), node.DSLType)) {
									return
								}
								continue
							}
							if !yield(node) {
								return
							}
						}
					}
				}
				pipeline = append(pipeline, bimapStep)
			}
		} else if strings.HasPrefix(stepStr, "flatmap") {
			// flatmap f: the Monad bind (bindValue) at the DSL level; a scalar stream
			// element's result is flattened and emitted in turn.
			arg := strings.TrimSpace(stepStr[7:])
			if c, ok, err := i.stepClosure(arg, 1); err != nil {
				pipeline = append(pipeline, i.failStep(fmt.Errorf("flatmap: %w", err)))
			} else if ok {
				pipeline = append(pipeline, flatmapStepOver(closureMapper(c)))
			} else {
				pipeline = append(pipeline, i.failStep(fmt.Errorf("flatmap: cannot use %q as a one-argument function", arg)))
			}
		} else if strings.HasPrefix(stepStr, "map") {
			arg := strings.TrimSpace(stepStr[3:])
			if c, ok, err := i.stepClosure(arg, 1); err != nil {
				pipeline = append(pipeline, i.failStep(fmt.Errorf("map: %w", err)))
			} else if ok {
				elem := closureMapper(c)
				if vars, body, isLam := parseLambdaInner(strings.TrimSuffix(strings.TrimPrefix(arg, "("), ")")); isLam && isLambdaText(arg) && len(vars) == 1 {
					elem = withPipelineFallback(elem, i.pipelineLambda(vars[0], body), body)
				}
				pipeline = append(pipeline, mapStepOver(elem))
			} else if vars, body, isLam := parseLambdaInner(strings.TrimSuffix(strings.TrimPrefix(arg, "("), ")")); isLam && isLambdaText(arg) && len(vars) == 1 {
				pipeline = append(pipeline, mapStepOver(i.pipelineLambda(vars[0], body)))
			} else if isLambdaText(arg) {
				pipeline = append(pipeline, i.failStep(fmt.Errorf("map: expected a one-argument function, got: %s", arg)))
			} else {
				subParts, _ := splitPipelineAware(arg)
				subPipeline := i.parsePipelineSteps(subParts)
				pipeline = append(pipeline, evaluator.StepMap(subPipeline, i.Engine))
			}
		} else {
			parts := strings.Fields(stepStr)
			if len(parts) > 0 {
				if factory, exists := plugins.Registry[parts[0]]; exists {
					pipeline = append(pipeline, evaluator.StepAdapter(factory()))
					continue
				}
			}
			if bs, ok := i.builtinStep(stepStr); ok {
				pipeline = append(pipeline, bs)
				continue
			}
			adapter := evaluator.EitherT(interactive.GenericExecAdapter(stepStr, 0))
			pipeline = append(pipeline, evaluator.StepAdapter(adapter))
		}
	}
	return pipeline
}

// unbalancedDelimiter reports whether a line's (), [], and {} delimiters are
// balanced and correctly nested, ignoring anything inside double-quoted string
// literals. On imbalance it returns the rune index of the offending delimiter
// and false; when balanced it returns (0, true). Used by `gatwd --check`.
func unbalancedDelimiter(line string) (int, bool) {
	var stack []struct {
		open rune
		pos  int
	}
	closer := map[rune]rune{')': '(', ']': '[', '}': '{'}
	inString := false
	escaped := false
	for i, r := range line {
		if inString && escaped {
			escaped = false
			continue
		}
		if inString && r == '\\' {
			escaped = true
			continue
		}
		switch {
		case r == '"':
			inString = !inString
		case inString:
			// skip delimiters inside string literals
		case r == '(' || r == '[' || r == '{':
			stack = append(stack, struct {
				open rune
				pos  int
			}{r, i})
		case r == ')' || r == ']' || r == '}':
			if len(stack) == 0 || stack[len(stack)-1].open != closer[r] {
				return i, false
			}
			stack = stack[:len(stack)-1]
		}
	}
	if len(stack) > 0 {
		return stack[len(stack)-1].pos, false
	}
	return 0, true
}

// checkExhaustiveMatch scans a statement for `match { ... }` blocks whose arms are
// all constructors of a single sum type, and reports — via warn — any branches of
// that type left uncovered when there is no catch-all (`_` or a bare variable) arm.
// This is a local, registry-driven check: no type inference is required because the
// constructor arms themselves identify the sum type. It is advisory (a warning),
// since gatwd permits partial matches at runtime.
func checkExhaustiveMatch(p *parser.Parser, stmt string, warn func(msg string)) {
	const marker = "match {"
	for idx := 0; ; {
		mi := strings.Index(stmt[idx:], marker)
		if mi < 0 {
			return
		}
		start := idx + mi + len(marker)
		depth, end := 1, start
		for end < len(stmt) && depth > 0 {
			switch stmt[end] {
			case '{':
				depth++
			case '}':
				depth--
			}
			if depth == 0 {
				break
			}
			end++
		}
		if depth != 0 {
			return // unbalanced; already reported by the delimiter check
		}

		var ctors []string
		typeName := ""
		catchAll, mixed := false, false
		for _, arm := range splitCasesAware(stmt[start:end]) {
			pp := strings.SplitN(strings.TrimSpace(arm), "->", 2)
			pat := strings.TrimSpace(pp[0])
			switch {
			case pat == "" || pat == "_":
				if pat == "_" {
					catchAll = true
				}
			case isIdent(pat) && p.IsConstructor(pat):
				tn, _, _ := p.SumTypeOf(pat)
				if typeName == "" {
					typeName = tn
				} else if typeName != tn {
					mixed = true
				}
				ctors = append(ctors, pat)
			case isIdent(pat):
				catchAll = true // a bare variable pattern binds anything
			}
			// non-identifier patterns (lists, literals) don't constrain a sum type
		}

		if typeName != "" && !catchAll && !mixed {
			_, all, _ := p.SumTypeOf(ctors[0])
			covered := make(map[string]bool, len(ctors))
			for _, c := range ctors {
				covered[c] = true
			}
			var missing []string
			for _, b := range all {
				if !covered[b] {
					missing = append(missing, b)
				}
			}
			if len(missing) > 0 {
				sort.Strings(missing)
				warn(fmt.Sprintf("non-exhaustive match on %s: missing %s", typeName, strings.Join(missing, ", ")))
			}
		}
		idx = end + 1
	}
}

// copyLiteral copies the String literal ("…", with backslash escapes) or the Char literal
// ('x', '\n') that starts at runes[i] into out and returns the index of its last rune, so a
// splitter can treat the whole literal as text. It reports false when runes[i] starts none (an
// unterminated string is copied to the end of the line).
func copyLiteral(runes []rune, i int, out *strings.Builder) (int, bool) {
	switch runes[i] {
	case '"':
		out.WriteRune('"')
		for j := i + 1; j < len(runes); j++ {
			out.WriteRune(runes[j])
			if runes[j] == '\\' && j+1 < len(runes) {
				j++
				out.WriteRune(runes[j])
			} else if runes[j] == '"' {
				return j, true
			}
		}
		return len(runes) - 1, true
	case '\'':
		if i+2 < len(runes) && runes[i+1] != '\\' && runes[i+2] == '\'' {
			out.WriteString(string(runes[i : i+3]))
			return i + 2, true
		}
		if i+3 < len(runes) && runes[i+1] == '\\' && runes[i+3] == '\'' {
			out.WriteString(string(runes[i : i+4]))
			return i + 3, true
		}
	}
	return 0, false
}

func splitCasesAware(line string) []string {
	var parts []string
	var current strings.Builder
	depthParen := 0
	depthBrace := 0
	depthBracket := 0

	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Delimiters and `|>` inside a String or Char literal are text, not structure.
		if last, ok := copyLiteral(runes, i, &current); ok {
			i = last
			continue
		}

		if r == '(' {
			depthParen++
		}
		if r == ')' {
			depthParen--
		}
		if r == '{' {
			depthBrace++
		}
		if r == '}' {
			depthBrace--
		}
		if r == '[' {
			depthBracket++
		}
		if r == ']' {
			depthBracket--
		}

		if depthParen == 0 && depthBrace == 0 && depthBracket == 0 {
			if r == '|' {
				// Make sure it's not |> or ||>
				if i+1 < len(runes) && runes[i+1] == '>' {
					// It's |>, don't split, just add it
					current.WriteRune(r)
					continue
				}
				if i+2 < len(runes) && runes[i+1] == '|' && runes[i+2] == '>' {
					// It's ||>, don't split, just add it
					current.WriteRune(r)
					continue
				}

				// It's a genuine case separator!
				parts = append(parts, strings.TrimSpace(current.String()))
				current.Reset()
				continue
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}

	return parts
}

func splitPipelineOps(line string) ([]string, []bool) {
	var parts []string
	var current strings.Builder
	depthParen := 0
	depthBrace := 0
	depthBracket := 0
	// parallel[k] is true when the operator in front of parts[k] was `||>`.
	var parallel []bool
	nextParallel := false

	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if last, ok := copyLiteral(runes, i, &current); ok {
			i = last
			continue
		}

		if r == '(' {
			depthParen++
		}
		if r == ')' {
			depthParen--
		}
		if r == '{' {
			depthBrace++
		}
		if r == '}' {
			depthBrace--
		}
		if r == '[' {
			depthBracket++
		}
		if r == ']' {
			depthBracket--
		}

		if depthParen == 0 && depthBrace == 0 && depthBracket == 0 {
			if r == '|' && i+1 < len(runes) && runes[i+1] == '>' {
				parts = append(parts, strings.TrimSpace(current.String()))
				parallel = append(parallel, nextParallel)
				nextParallel = false
				current.Reset()
				i++ // skip '>'
				continue
			} else if r == '|' && i+2 < len(runes) && runes[i+1] == '|' && runes[i+2] == '>' {
				parts = append(parts, strings.TrimSpace(current.String()))
				parallel = append(parallel, nextParallel)
				nextParallel = true
				current.Reset()
				i += 2 // skip '|>'
				continue
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
		parallel = append(parallel, nextParallel)
	}

	return parts, parallel
}

// splitPipelineAware splits a pipeline line at its top-level `|>` / `||>` operators and
// reports whether any of them was the parallel `||>`.
func splitPipelineAware(line string) ([]string, bool) {
	parts, parallel := splitPipelineOps(line)
	for _, p := range parallel {
		if p {
			return parts, true
		}
	}
	return parts, false
}

// parallelWorkers is the worker-pool size of a `||>` step.
const parallelWorkers = 10

// pipelineStepsFor parses parts[1:] into steps. A part reached through `||>` becomes a
// StepPar over the steps it parses to, so it runs across the stream's nodes on a
// worker pool; the sequential case is parsed exactly as before.
func (i *Interpreter) pipelineStepsFor(parts []string, parallel []bool) []evaluator.PipelineStep[functions.Either[error, any]] {
	anyParallel := false
	for k := 1; k < len(parallel); k++ {
		anyParallel = anyParallel || parallel[k]
	}
	if !anyParallel {
		return i.parsePipelineSteps(parts[1:])
	}
	var out []evaluator.PipelineStep[functions.Either[error, any]]
	for k := 1; k < len(parts); k++ {
		steps := i.parsePipelineSteps(parts[k : k+1])
		if parallel[k] && len(steps) > 0 {
			out = append(out, evaluator.StepPar(evaluator.ComposeSteps(steps...), parallelWorkers))
		} else {
			out = append(out, steps...)
		}
	}
	return out
}

// letResultError turns a failed `let` right-hand side into an error: the pipeline
// yielded no value at all, or its first result is an error value. Binding either would
// leave later statements to fail far from the cause.
func letResultError(name string, res []any) error {
	if len(res) == 0 {
		return fmt.Errorf("let %s: right-hand side yielded no value", name)
	}
	if e, ok := res[0].(error); ok {
		return fmt.Errorf("let %s: %w", name, e)
	}
	return nil
}

// startsWithLiteral reports whether a pipeline head begins like a number or a string
// literal, so a bare `5 |> f` or `"x" |> f` is evaluated as an expression instead of being
// handed to Parser.Parse, which only knows `spawn` and `type` lines and would yield nothing.
func startsWithLiteral(s string) bool {
	if s == "" {
		return false
	}
	if s == "true" || s == "false" {
		return true
	}
	r := rune(s[0])
	return r == '"' || r == '\'' || (r >= '0' && r <= '9')
}

// scalarValue reports whether v prints and re-parses losslessly as text: booleans,
// numbers and strings (and nil). Collections, closures and other structures do not, so
// they must reach the evaluator as values and never be spliced into pipeline text.
func scalarValue(v any) bool {
	if functions.IsChar(v) {
		return false // its text ("a") would re-parse as a String
	}
	if functions.IsTemporal(v) {
		return false // a Duration is an int64 underneath, but its text is a typed literal, not a number
	}
	switch reflect.ValueOf(v).Kind() {
	case reflect.Invalid, reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// textSubstitutable reports whether the variable k may be replaced by its text in a
// pipeline head. A name bound to a real, non-scalar value (a HAMT, list, graph, closure…)
// must reach the expression evaluator as that value: its `%v` text is lossy, and
// substituting it would turn `hamt_get "kind" r` into a call on a string.
func (i *Interpreter) textSubstitutable(k string) bool {
	rt, ok := i.RuntimeNodes[k]
	if !ok || !rt.Value.IsRight() {
		return true
	}
	v, _ := rt.Value.UnwrapRight()
	return scalarValue(v)
}

// constSplice is the text a constant's value is replaced by when a constant name appears as a
// word of a pipeline head: a String as its quoted literal (a bare word would be read as an
// undefined name, a space would split the line), anything else as its printed form.
func constSplice(v any) string {
	if s, ok := v.(string); ok {
		return quoteString(s)
	}
	return fmt.Sprintf("%v", v)
}

// spliceText is the text a let-bound variable k (printed form v) is replaced by when it is
// substituted into a pipeline head. A String is spliced as its quoted literal so the line
// re-parses to the same String (a bare word would be read as an undefined name, and a
// space, quote or newline would split or corrupt the line); everything else as its text.
func (i *Interpreter) spliceText(k, v string) string {
	if rt, ok := i.RuntimeNodes[k]; ok && rt.Value.IsRight() {
		rv, _ := rt.Value.UnwrapRight()
		if sv, isString := rv.(string); isString {
			return quoteString(sv)
		}
	}
	return v
}

func (i *Interpreter) buildInitialStream(firstPart string) func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
	lookup := func(k string) (any, bool) {
		if rt, ok := i.RuntimeNodes[k]; ok {
			if rt.Value.IsRight() {
				val, _ := rt.Value.UnwrapRight()
				return val, true
			}
		}
		if v, ok := i.Variables[k]; ok {
			return v, true
		}
		if v, ok := i.Constants[k]; ok {
			return v, true
		}
		return i.exprFallback(k)
	}

	// 1. Direct variable reference: e.g. "nbrs |> unchunk"
	if rtNode, exists := i.RuntimeNodes[firstPart]; exists {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			yield(rtNode)
		}
	}
	if v, exists := i.Variables[firstPart]; exists {
		var parsedVal any = v
		if evaluated, err := parser.EvalInfixString(v, lookup); err == nil {
			parsedVal = evaluated
		}
		flag := i.Parser.GetFlag("Variable")
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			yield(dsltypes.NewASTNode(functions.Right[error, any](parsedVal), flag))
		}
	}

	// 2. Bracket expression (list literals or list comprehensions)
	if strings.HasPrefix(firstPart, "[") && strings.HasSuffix(firstPart, "]") && !strings.Contains(firstPart, " *** ") {
		if val, err := parser.EvalInfixString(firstPart, lookup); err == nil {
			flag := i.Parser.GetFlag("List")
			return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
				yield(dsltypes.NewASTNode(functions.Right[error, any](val), flag))
			}
		}
	}

	if strings.HasPrefix(firstPart, "spawn ") {
		colonIdx := strings.Index(firstPart, ":")
		if colonIdx != -1 {
			typeName := strings.TrimSpace(firstPart[6:colonIdx])
			valStr := strings.TrimSpace(firstPart[colonIdx+1:])
			var parsedValue any = valStr
			if strings.ContainsAny(valStr, "+-*/%&|<>= ") {
				if evaluated, err := parser.EvalInfixString(valStr, nil); err == nil {
					parsedValue = evaluated
				} else if factory, ok := i.Parser.GetFactory(typeName); ok {
					parsedValue = factory(valStr)
				}
			} else if factory, ok := i.Parser.GetFactory(typeName); ok {
				parsedValue = factory(valStr)
			} else if strings.HasPrefix(valStr, "\"") && strings.HasSuffix(valStr, "\"") && len(valStr) >= 2 {
				parsedValue = valStr[1 : len(valStr)-1]
			}
			flag := i.Parser.GetFlag(typeName)
			return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
				yield(dsltypes.NewASTNode(functions.Right[error, any](parsedValue), flag))
			}
		}
	}

	// If firstPart is a bound runtime variable (e.g. nbrs in nbrs |> ...), yield its bound value
	if val, ok := lookup(firstPart); ok {
		flag := i.Parser.GetFlag("Variable")
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			yield(dsltypes.NewASTNode(functions.Right[error, any](val), flag))
		}
	}

	// A leading `!` is the prefix operator (`!true`, `!ok`), so the head is an expression.
	if strings.HasPrefix(firstPart, "(") || strings.HasPrefix(firstPart, "!") || startsWithLiteral(firstPart) || (strings.ContainsAny(firstPart, "+-*/%&|<>= ") && !strings.HasPrefix(firstPart, "emit ")) {
		val, err := parser.EvalInfixString(firstPart, lookup)
		if err == nil {
			flag := i.Parser.GetFlag("Variable")
			return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
				yield(dsltypes.NewASTNode(functions.Right[error, any](val), flag))
			}
		}
		// A call that was actually made and failed is reported, not swallowed; a mere
		// name-resolution miss still falls through to the other parses.
		var applyErr *parser.ApplyError
		if errors.As(err, &applyErr) {
			flag := i.Parser.GetFlag("Variable")
			return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
				yield(evaluator.InjectError[any](err, flag))
			}
		}
	}

	stringStream := i.Parser.Parse(firstPart)
	return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		for node := range stringStream {
			wrappedValue := functions.Right[error, any](any(node.Value))
			yield(dsltypes.NewASTNode(wrappedValue, node.DSLType))
		}
	}
}

func (i *Interpreter) EvaluateToValues(line string) []any {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	// Let Bindings: let x = ... in ...
	if strings.HasPrefix(line, "let ") {
		parts := strings.SplitN(line[4:], " in ", 2)
		if len(parts) == 2 {
			defParts := strings.SplitN(parts[0], "=", 2)
			if len(defParts) == 2 {
				varName := strings.TrimSpace(defParts[0])
				pipelineStr := strings.TrimSpace(defParts[1])

				prevVal, hadPrev := i.Variables[varName]
				prevNode, hadPrevNode := i.RuntimeNodes[varName]

				lookup := func(k string) (any, bool) {
					if rt, ok := i.RuntimeNodes[k]; ok {
						if rt.Value.IsRight() {
							val, _ := rt.Value.UnwrapRight()
							return val, true
						}
					}
					if v, ok := i.Variables[k]; ok {
						return v, true
					}
					if v, ok := i.Constants[k]; ok {
						return v, true
					}
					return i.exprFallback(k)
				}

				cleanPipeline := pipelineStr
				if strings.HasPrefix(cleanPipeline, "(") && strings.HasSuffix(cleanPipeline, ")") {
					cleanPipeline = strings.TrimSpace(cleanPipeline[1 : len(cleanPipeline)-1])
				}
				if strings.Contains(cleanPipeline, "|>") || strings.HasPrefix(cleanPipeline, "spawn ") {
					res := i.EvaluateToValues(cleanPipeline)
					if err := letResultError(varName, res); err != nil {
						return []any{err}
					}
					val := res[0]
					i.Variables[varName] = fmt.Sprintf("%v", val)
					i.RuntimeNodes[varName] = dsltypes.NewASTNode(functions.Right[error, any](val), i.Parser.GetFlag("Variable"))
				} else {
					if val, err := parser.EvalInfixString(pipelineStr, lookup); err == nil {
						i.Variables[varName] = fmt.Sprintf("%v", val)
						i.RuntimeNodes[varName] = dsltypes.NewASTNode(functions.Right[error, any](val), i.Parser.GetFlag("Variable"))
					} else {
						i.Variables[varName] = pipelineStr
					}
				}

				defer func() {
					if hadPrev {
						i.Variables[varName] = prevVal
					} else {
						delete(i.Variables, varName)
					}
					if hadPrevNode {
						i.RuntimeNodes[varName] = prevNode
					} else {
						delete(i.RuntimeNodes, varName)
					}
				}()

				return i.EvaluateToValues(parts[1])
			}
		}
	}

	parts, parallel := splitPipelineOps(line)

	firstPart := parts[0]
	if _, isVar := i.Variables[firstPart]; !isVar {
		if _, isRt := i.RuntimeNodes[firstPart]; !isRt {
			for k, v := range i.Variables {
				if !i.textSubstitutable(k) {
					continue
				}
				v = i.spliceText(k, v)
				words := strings.Split(firstPart, " ")
				for idx, w := range words {
					if w == k {
						words[idx] = v
					}
				}
				firstPart = strings.Join(words, " ")
			}
		}
	}
	for k, v := range i.Constants {
		// A head that is exactly a constant's name resolves to the value itself (lookup),
		// not to its text, which no parser would accept as a statement.
		if !scalarValue(v) || strings.TrimSpace(firstPart) == k {
			continue
		}
		words := strings.Split(firstPart, " ")
		for idx, w := range words {
			if w == k {
				words[idx] = constSplice(v)
			}
		}
		firstPart = strings.Join(words, " ")
	}

	initialStream := i.buildInitialStream(firstPart)

	pipeline := i.pipelineStepsFor(parts, parallel)
	terminalNodes := i.Engine.EvaluatePipeline(initialStream, pipeline)

	var results []any
	for result := range terminalNodes {
		if result.Value.IsRight() {
			val, _ := result.Value.UnwrapRight()
			results = append(results, val)
		} else if result.Value.IsLeft() {
			err, _ := result.Value.UnwrapLeft()
			results = append(results, err)
		}
	}
	return results
}

func (i *Interpreter) evaluatePipelineStr(line string) []string {
	parts, parallel := splitPipelineOps(line)

	firstPart := parts[0]
	if _, isVar := i.Variables[firstPart]; !isVar {
		if _, isRt := i.RuntimeNodes[firstPart]; !isRt {
			for k, v := range i.Variables {
				if !i.textSubstitutable(k) {
					continue
				}
				v = i.spliceText(k, v)
				words := strings.Split(firstPart, " ")
				for idx, w := range words {
					if w == k {
						words[idx] = v
					}
				}
				firstPart = strings.Join(words, " ")
			}
		}
	}
	for k, v := range i.Constants {
		// A head that is exactly a constant's name resolves to the value itself (lookup),
		// not to its text, which no parser would accept as a statement.
		if !scalarValue(v) || strings.TrimSpace(firstPart) == k {
			continue
		}
		// Replace bare identifier in firstPart or after "spawn Type: IDENT"
		colonParts := strings.SplitN(firstPart, ":", 2)
		if len(colonParts) == 2 {
			rhs := strings.TrimSpace(colonParts[1])
			if rhs == k {
				firstPart = colonParts[0] + ": " + fmt.Sprintf("%v", v)
			}
		} else {
			words := strings.Split(firstPart, " ")
			for idx, w := range words {
				if w == k {
					words[idx] = constSplice(v)
				}
			}
			firstPart = strings.Join(words, " ")
		}
	}

	initialStream := i.buildInitialStream(firstPart)

	pipeline := i.pipelineStepsFor(parts, parallel)
	terminalNodes := i.Engine.EvaluatePipeline(initialStream, pipeline)

	results := renderNodes(terminalNodes)
	if len(results) == 0 {
		if err := i.explainSilentStatement(line, parts); err != nil {
			results = append(results, fmt.Sprintf("[ERROR] %v", err))
		}
	}
	return results
}

// renderNodes drains a stream into the printed form of its nodes: `[ERROR] …` for a Left,
// the raw text of a byte slice, plain digits for a big integer, and `%+v` for everything else
// (displayText).
func renderNodes(nodes iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) []string {
	var out []string
	for node := range nodes {
		out = append(out, renderValue(node.Value))
	}
	return out
}

func renderValue(v functions.Either[error, any]) string {
	if v.IsLeft() {
		err, _ := v.UnwrapLeft()
		return fmt.Sprintf("[ERROR] %v", err)
	}
	val, _ := v.UnwrapRight()
	return displayText(val)
}

// explainSilentStatement reports why a single-part statement that produced nothing was
// not understood: it is neither a declaration (`type`, `import`, `const`, `def`, `let`)
// nor a spawn, and it does not evaluate as an infix expression. It returns nil when the
// silence is legitimate or has no better explanation, leaving the caller's generic
// "no output" message in place.
func (i *Interpreter) explainSilentStatement(line string, parts []string) error {
	line = strings.TrimSpace(line)
	if len(parts) != 1 || line == "" || strings.HasPrefix(line, "--") {
		return nil
	}
	for _, kw := range []string{"type ", "import ", "const ", "def ", "let ", "spawn ", "emit "} {
		if strings.HasPrefix(line, kw) {
			return nil
		}
	}
	if _, err := parser.EvalInfixString(line, i.graphLookup(nil)); err != nil {
		return fmt.Errorf("cannot evaluate %q: %w", line, err)
	}
	return nil
}

// RegisterConstant binds a native Go constant into the DSL environment.
func (i *Interpreter) RegisterConstant(name string, val any) {
	i.Constants[name] = val
}

// RegisterConstants binds multiple Go constants into the DSL environment.
func (i *Interpreter) RegisterConstants(constants map[string]any) {
	for k, v := range constants {
		i.Constants[k] = v
	}
}

// RegisterType allows third-party Go code to register native constructors for DSL types.
// Accepts either a simple type name like "IntVal" or a full definition like "type IntVal = Int".
func (i *Interpreter) RegisterType(definition string, factory func(string) any) {
	typeName := strings.TrimSpace(definition)
	if strings.HasPrefix(typeName, "type ") {
		typeName = strings.TrimSpace(typeName[5:])
	}
	if strings.Contains(typeName, "=") {
		parts := strings.SplitN(typeName, "=", 2)
		typeName = strings.TrimSpace(parts[0])
	}
	i.Parser.RegisterTypeFactory(typeName, factory)
	i.Parser.GetFlag(typeName)
}

func (i *Interpreter) EvaluateLine(line string) {
	// Let Bindings: let x = ... in ... or top-level let x = ...
	if strings.HasPrefix(line, "let ") {
		parts := strings.SplitN(line[4:], " in ", 2)
		if len(parts) == 2 {
			defParts := strings.SplitN(parts[0], "=", 2)
			if len(defParts) == 2 {
				varName := strings.TrimSpace(defParts[0])
				pipelineStr := strings.TrimSpace(defParts[1])

				prevVal, hadPrev := i.Variables[varName]
				prevNode, hadPrevNode := i.RuntimeNodes[varName]

				lookup := func(k string) (any, bool) {
					if rt, ok := i.RuntimeNodes[k]; ok {
						if rt.Value.IsRight() {
							val, _ := rt.Value.UnwrapRight()
							return val, true
						}
					}
					if v, ok := i.Variables[k]; ok {
						return v, true
					}
					if v, ok := i.Constants[k]; ok {
						return v, true
					}
					return i.exprFallback(k)
				}

				// Evaluate the left side completely
				cleanPipeline := pipelineStr
				if strings.HasPrefix(cleanPipeline, "(") && strings.HasSuffix(cleanPipeline, ")") {
					cleanPipeline = strings.TrimSpace(cleanPipeline[1 : len(cleanPipeline)-1])
				}
				if strings.Contains(cleanPipeline, "|>") || strings.HasPrefix(cleanPipeline, "spawn ") {
					res := i.EvaluateToValues(cleanPipeline)
					if err := letResultError(varName, res); err != nil {
						fmt.Fprintf(i.out(), "∷ [ERROR] %v\n", err)
						return
					}
					val := res[0]
					i.Variables[varName] = fmt.Sprintf("%v", val)
					i.RuntimeNodes[varName] = dsltypes.NewASTNode(functions.Right[error, any](val), i.Parser.GetFlag("Variable"))
				} else {
					// Direct infix expression evaluation
					if val, err := parser.EvalInfixString(pipelineStr, lookup); err == nil {
						i.Variables[varName] = fmt.Sprintf("%v", val)
						i.RuntimeNodes[varName] = dsltypes.NewASTNode(functions.Right[error, any](val), i.Parser.GetFlag("Variable"))
					} else {
						i.Variables[varName] = pipelineStr
					}
				}

				// Evaluate the right side
				i.EvaluateLine(parts[1])

				// Restore previous variable value after right side completes
				if hadPrev {
					i.Variables[varName] = prevVal
				} else {
					delete(i.Variables, varName)
				}
				if hadPrevNode {
					i.RuntimeNodes[varName] = prevNode
				} else {
					delete(i.RuntimeNodes, varName)
				}
			}
		} else {
			// Top-level variable definition: let x = ...
			defParts := strings.SplitN(line[4:], "=", 2)
			if len(defParts) == 2 {
				varName := strings.TrimSpace(defParts[0])
				pipelineStr := strings.TrimSpace(defParts[1])

				lookup := func(k string) (any, bool) {
					if rt, ok := i.RuntimeNodes[k]; ok {
						if rt.Value.IsRight() {
							val, _ := rt.Value.UnwrapRight()
							return val, true
						}
					}
					if v, ok := i.Variables[k]; ok {
						return v, true
					}
					if v, ok := i.Constants[k]; ok {
						return v, true
					}
					return i.exprFallback(k)
				}

				cleanPipeline := pipelineStr
				if strings.HasPrefix(cleanPipeline, "(") && strings.HasSuffix(cleanPipeline, ")") {
					cleanPipeline = strings.TrimSpace(cleanPipeline[1 : len(cleanPipeline)-1])
				}
				if strings.Contains(cleanPipeline, "|>") || strings.HasPrefix(cleanPipeline, "spawn ") {
					res := i.EvaluateToValues(cleanPipeline)
					if err := letResultError(varName, res); err != nil {
						fmt.Fprintf(i.out(), "∷ [ERROR] %v\n", err)
						return
					}
					val := res[0]
					i.Variables[varName] = fmt.Sprintf("%v", val)
					i.RuntimeNodes[varName] = dsltypes.NewASTNode(functions.Right[error, any](val), i.Parser.GetFlag("Variable"))
				} else {
					if val, err := parser.EvalInfixString(pipelineStr, lookup); err == nil {
						i.Variables[varName] = fmt.Sprintf("%v", val)
						i.RuntimeNodes[varName] = dsltypes.NewASTNode(functions.Right[error, any](val), i.Parser.GetFlag("Variable"))
					} else {
						i.Variables[varName] = pipelineStr
					}
				}
			}
		}
		return
	}

	// Constant Definitions: const NAME = ...
	if strings.HasPrefix(line, "const ") {
		parts := strings.SplitN(line[6:], "=", 2)
		if len(parts) == 2 {
			constName := strings.TrimSpace(parts[0])
			valStr := strings.TrimSpace(parts[1])
			quoted := strings.HasPrefix(valStr, "\"")
			valStr = strings.Trim(valStr, "\"")
			i.invalidateCompiledCache()
			if strings.Contains(valStr, "\n") {
				i.Constants[constName] = valStr
				return
			}
			lookup := func(k string) (any, bool) {
				if v, ok := i.Constants[k]; ok {
					return v, true
				}
				return nil, false
			}
			// If it's a compound expression, evaluate it
			if strings.ContainsAny(valStr, "+-*/%&|<>= ") {
				if evaluated, err := parser.EvalInfixString(valStr, lookup); err == nil {
					if functions.IsNumeric(evaluated) {
						i.Constants[constName] = evaluated
					} else {
						i.Constants[constName] = fmt.Sprintf("%v", evaluated)
					}
				} else {
					i.Constants[constName] = valStr
				}
			} else if quoted {
				i.Constants[constName] = valStr
			} else {
				// A bare canonical number is that number (`const N = 5`), anything else text.
				i.Constants[constName] = functions.LiteralToken(valStr)
			}
		}
		return
	}

	// Function Definitions: def func = ...
	if strings.HasPrefix(line, "def ") {
		parts := strings.SplitN(line[4:], "=", 2)
		if len(parts) == 2 {
			funcName, _ := splitDefHead(parts[0]) // optional `: type` annotation is advisory only
			pipelineStr := strings.TrimSpace(parts[1])
			i.Funcs[funcName] = pipelineStr
			i.invalidateCompiledCache()
		}
		return
	}

	results := i.evaluatePipelineStr(line)
	for _, res := range results {
		fmt.Fprintf(i.out(), "∷ %s\n", res)
	}
	if len(results) == 0 && !strings.HasPrefix(line, "type ") && !strings.HasPrefix(line, "import ") && !strings.HasPrefix(line, "const ") {
		fmt.Fprintln(i.out(), "(No output or node unhandled)")
	}
}

// EvaluateScript evaluates an entire multi-line script string, resolving comments,
// multi-line expressions, macros, and pipeline definitions sequentially.
func (i *Interpreter) EvaluateScript(script string) {
	i.LoadScriptContent(script, ".")
}

// LoadScriptContent parses and evaluates a script string with a base directory for resolving relative imports.
func (i *Interpreter) LoadScriptContent(content string, baseDir string) {
	i.loadScriptContent(content, baseDir, nil)
}

// loadScriptContent is LoadScriptContent with an optional virtual filesystem: when
// fsys is non-nil, `import` statements are resolved inside it (slash-separated,
// relative to baseDir) instead of the OS filesystem, so a multi-file prelude
// embedded with //go:embed can import its sibling modules.
func (i *Interpreter) loadScriptContent(content string, baseDir string, fsys fs.FS) {
	for _, stmt := range logicalLines(content) {
		buffer := stmt.text
		// Module Imports
		if strings.HasPrefix(buffer, "import ") {
			importFile := strings.Trim(strings.TrimSpace(buffer[7:]), "\"")
			if fsys != nil {
				importFile = pathpkg.Join(baseDir, importFile)
				bytes, err := fs.ReadFile(fsys, importFile)
				if err != nil {
					fmt.Fprintf(i.out(), "Error reading embedded import '%s': %v\n", importFile, err)
					continue
				}
				i.loadScriptContent(string(bytes), pathpkg.Dir(importFile), fsys)
				continue
			}
			if !filepath.IsAbs(importFile) {
				importFile = filepath.Join(baseDir, importFile)
			}
			i.LoadScript(importFile)
			continue
		}
		i.EvaluateLine(buffer)
	}
}

// logicalLine is a complete gatwd statement assembled from one or more physical
// lines, tagged with the 1-based line number where the statement began.
type logicalLine struct {
	num  int
	text string
}

// lineState is what a scan of statement text leaves open at its end: whether a double-quoted
// string is unterminated and how many `{`, `(` and `[` are still unclosed. Delimiters inside
// string literals and Char literals (`'('`, `'\"'`) do not count, and a backslash escapes the
// next character inside a string, so `"a\"b"` is one string.
type lineState struct {
	inString                 bool
	braces, parens, brackets int
}

// open reports whether the statement must continue on the next physical line.
func (st lineState) open() bool {
	return st.inString || st.braces > 0 || st.parens > 0 || st.brackets > 0
}

func scanLine(s string) lineState {
	var st lineState
	dec := func(n *int) {
		if *n > 0 {
			*n--
		}
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case st.inString && c == '\\':
			i++
		case st.inString && c == '"':
			st.inString = false
		case st.inString:
		case c == '"':
			st.inString = true
		case c == '\'' && i+2 < len(s) && s[i+1] != '\\' && s[i+2] == '\'':
			i += 2 // 'x'
		case c == '\'' && i+3 < len(s) && s[i+1] == '\\' && s[i+3] == '\'':
			i += 3 // '\n'
		case c == '{':
			st.braces++
		case c == '}':
			dec(&st.braces)
		case c == '(':
			st.parens++
		case c == ')':
			dec(&st.parens)
		case c == '[':
			st.brackets++
		case c == ']':
			dec(&st.brackets)
		}
	}
	return st
}

// unterminatedString reports whether s ends inside a double-quoted string literal.
func unterminatedString(s string) bool { return scanLine(s).inString }

// logicalLines splits a script into complete statements, joining physical lines
// that continue a statement (unbalanced delimiters, an unterminated string, or a
// trailing continuation operator like `=`, `|>`, or `in`). Comment and blank
// lines are skipped. This is the single source of truth for statement assembly,
// shared by script evaluation and `gatwd --check`.
func logicalLines(content string) []logicalLine {
	lines := strings.Split(content, "\n")
	var out []logicalLine
	var buffer string
	startNum := 0

	for idx, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		if buffer != "" {
			if unterminatedString(buffer) {
				buffer += "\n" + rawLine
			} else {
				buffer += " " + line
			}
		} else {
			buffer = line
			startNum = idx + 1
		}

		// Check for multi-line continuation
		if strings.HasSuffix(buffer, "=") ||
			strings.HasSuffix(buffer, "|>") ||
			strings.HasSuffix(buffer, "||>") ||
			strings.HasSuffix(buffer, " let") ||
			strings.HasSuffix(buffer, " in") ||
			strings.HasSuffix(buffer, "\\") ||
			scanLine(buffer).open() {
			continue // Wait for next line
		}

		out = append(out, logicalLine{num: startNum, text: buffer})
		buffer = ""
	}
	if buffer != "" {
		out = append(out, logicalLine{num: startNum, text: buffer})
	}
	return out
}

// LoadScript reads a script from the filesystem and evaluates it.
func (i *Interpreter) LoadScript(filename string) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(i.out(), "Error reading file '%s': %v\n", filename, err)
		return
	}
	i.LoadScriptContent(string(bytes), filepath.Dir(filename))
}

// LoadScriptFS reads a script from a virtual filesystem (such as embed.FS) and evaluates it.
// This enables embedding Gatwd scripts directly into Go binaries for standalone releases,
// CLI tools, servers, and Kubernetes operators.
func (i *Interpreter) LoadScriptFS(fsys fs.FS, filename string) error {
	bytes, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return fmt.Errorf("failed to read embedded script '%s': %w", filename, err)
	}
	i.loadScriptContent(string(bytes), pathpkg.Dir(filename), fsys)
	return nil
}

// RunWithInterpreter runs the command line of the process (see command.go) against
// interpreter and exits with a non-zero status on failure. The interpreter has no
// embedded preludes to skip, so --no-prelude changes nothing.
func RunWithInterpreter(interpreter *Interpreter) {
	NewCommandTree(CommandSpec{
		Name:  "gatwd",
		Build: func(bool) *Interpreter { return interpreter },
	}).exitWith()
}

// Run executes the default CLI using the full NewInterpreter() with all builtins.
func Run() {
	RunWithInterpreter(NewInterpreter())
}

// buildGraphStep is the Layer-5 string dispatch for the generic Graph collection
// builtins. It parses a `graph_*` pipeline token into its evaluator.StepGraph*
// counterpart (evaluator/graph_steps.go), returning ok=false for any token it does
// not recognise so the surrounding pipeline dispatch can keep trying other steps.
//
// The builtins mirror the collections.Graph API as a new DSL collection row:
//
//	Construction : graph_empty, graph_add_node, graph_insert_node, graph_add_edge,
//	               graph_add_edge_inferred, graph_add_fanout, graph_add_fanin,
//	               graph_remove_node, graph_remove_edge
//	Projection   : graph_order, graph_size, graph_nodes, graph_edges,
//	               graph_out, graph_in, graph_neighbors, graph_nodes_of,
//	               graph_edges_of, graph_edges_of_group, graph_edges_inferred,
//	               graph_reachable
//	Walks/algos  : graph_bfs, graph_dfs, graph_walk_from, graph_shortest_path,
//	               graph_topo_sort, graph_scc
//
// Arguments are whitespace-separated; node identities and colors are bare tokens.
func buildGraphStep(stepStr string) (evaluator.PipelineStep[functions.Either[error, any]], bool) {
	fields := strings.Fields(stepStr)
	if len(fields) == 0 || !strings.HasPrefix(fields[0], "graph_") {
		return nil, false
	}
	name, args := fields[0], fields[1:]
	switch name {
	// --- construction ---
	case "graph_empty":
		return evaluator.StepGraphEmpty(), true
	case "graph_add_node":
		if len(args) >= 2 {
			return evaluator.StepGraphAddNode(args[0], args[1]), true
		}
	case "graph_insert_node":
		if len(args) >= 2 {
			return evaluator.StepGraphInsertNode(args[0], args[1]), true
		}
	case "graph_add_edge":
		if len(args) >= 4 {
			return evaluator.StepGraphAddEdge(args[0], parseGraphWeight(args[1]), args[2], args[3]), true
		}
	case "graph_add_edge_inferred":
		if len(args) >= 4 {
			return evaluator.StepGraphAddInferredEdge(args[0], parseGraphWeight(args[1]), args[2], args[3]), true
		}
	case "graph_add_fanout":
		if len(args) >= 3 {
			return evaluator.StepGraphAddFanOut(args[0], parseGraphWeight(args[1]), args[2], args[3:]), true
		}
	case "graph_add_fanin":
		if len(args) >= 3 {
			return evaluator.StepGraphAddFanIn(args[0], parseGraphWeight(args[1]), args[2], args[3:]), true
		}
	case "graph_remove_node":
		if len(args) >= 1 {
			return evaluator.StepGraphRemoveNode(args[0]), true
		}
	case "graph_remove_edge":
		if len(args) >= 3 {
			return evaluator.StepGraphRemoveEdge(args[0], args[1], args[2]), true
		}
	// --- projection ---
	case "graph_order":
		return evaluator.StepGraphOrder(), true
	case "graph_size":
		return evaluator.StepGraphSize(), true
	case "graph_nodes":
		return evaluator.StepGraphNodes(), true
	case "graph_edges":
		return evaluator.StepGraphEdges(), true
	case "graph_out":
		if len(args) >= 1 {
			return evaluator.StepGraphOut(args[0]), true
		}
	case "graph_in":
		if len(args) >= 1 {
			return evaluator.StepGraphIn(args[0]), true
		}
	case "graph_neighbors":
		if len(args) >= 1 {
			return evaluator.StepGraphNeighbors(args[0]), true
		}
	case "graph_nodes_of":
		if len(args) >= 1 {
			return evaluator.StepGraphNodesOf(args[0]), true
		}
	case "graph_edges_of":
		if len(args) >= 1 {
			return evaluator.StepGraphEdgesOf(args[0]), true
		}
	case "graph_edges_of_group":
		if len(args) >= 1 {
			return evaluator.StepGraphEdgesOfGroup(args[0]), true
		}
	case "graph_edges_inferred":
		return evaluator.StepGraphEdgesInferred(), true
	case "graph_reachable":
		if len(args) >= 1 {
			return evaluator.StepGraphReachable(args[0]), true
		}
	// --- walks & algorithms ---
	case "graph_bfs":
		if len(args) >= 1 {
			return evaluator.StepGraphBFS(args[0]), true
		}
	case "graph_dfs":
		if len(args) >= 1 {
			return evaluator.StepGraphDFS(args[0]), true
		}
	case "graph_walk_from":
		if len(args) >= 2 {
			return evaluator.StepGraphWalkFrom(args[0], args[1]), true
		}
	case "graph_shortest_path":
		if len(args) >= 2 {
			return evaluator.StepGraphShortestPath(args[0], args[1]), true
		}
	case "graph_topo_sort":
		return evaluator.StepGraphTopoSort(), true
	case "graph_scc":
		return evaluator.StepGraphSCC(), true
	// --- optics: single-node payload lens (read/set) ---
	case "graph_focus":
		if len(args) >= 1 {
			return evaluator.StepGraphFocus(args[0]), true
		}
	case "graph_set":
		if len(args) >= 2 {
			return evaluator.StepGraphSet(args[0], parseGraphWeight(strings.Join(args[1:], " "))), true
		}
	}
	return nil, false
}

// parseGraphWeight interprets a DSL weight token as int, then float64, then string,
// matching the evaluator's own parseWeight so graph_add_* and ShortestPath agree.
func parseGraphWeight(tok string) any {
	if i, err := strconv.Atoi(tok); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(tok, 64); err == nil {
		return f
	}
	return tok
}

// buildGraphLambdaStep is the Phase-4b dispatch for the graph builtins that need
// the interpreter's binding environment: the Functor/Foldable/Filterable steps
// (which take a DSL lambda) and the set-algebra steps (which take a second graph
// operand named by a let-bound variable). It is a method, not the free
// buildGraphStep, precisely because it must resolve names against i's Constants,
// Variables, and RuntimeNodes.
//
//	Functor     : graph_map_nodes (\p -> …)   graph_map_edges (\w -> …)
//	              graph_bimap (\p -> …) (\w -> …)
//	Filterable  : graph_filter_nodes (\p -> …)   graph_filter_edges (\w -> …)
//	Foldable    : graph_fold_nodes (\acc p -> …) seed
//	              graph_fold_edges (\acc w -> …) seed
//	              graph_fold_walk  (\acc p -> …) seed start [bfs|dfs]
//	Set-algebra : graph_union g2   graph_diff g2   graph_intersect g2
//
// Binding convention: node steps bind the node *payload* (its id by default); edge
// steps bind the edge *weight*. Returns ok=false for any unrecognised/ill-formed
// token so the surrounding dispatch keeps trying other steps.
func (i *Interpreter) buildGraphLambdaStep(stepStr string) (evaluator.PipelineStep[functions.Either[error, any]], bool) {
	fields := strings.Fields(stepStr)
	if len(fields) == 0 || !strings.HasPrefix(fields[0], "graph_") {
		return nil, false
	}
	name := fields[0]
	rest := strings.TrimSpace(stepStr[len(name):])
	switch name {
	case "graph_union", "graph_diff", "graph_intersect":
		if rest == "" {
			return nil, false
		}
		operand := i.resolveGraphOperand(rest)
		switch name {
		case "graph_union":
			return evaluator.StepSetUnion(operand), true
		case "graph_diff":
			return evaluator.StepSetDiff(operand), true
		case "graph_intersect":
			return evaluator.StepSetIntersect(operand), true
		}
	case "graph_map_nodes":
		if inside, _, ok := splitParen(rest); ok {
			if c, ok := i.graphLambda(inside, 1); ok {
				return latchedStep(func(fail errLatch) graphStep {
					return evaluator.StepGraphMapNodes(unaryFn(c, fail))
				}), true
			}
		}
	case "graph_map_edges":
		if inside, _, ok := splitParen(rest); ok {
			if c, ok := i.graphLambda(inside, 1); ok {
				return latchedStep(func(fail errLatch) graphStep {
					return evaluator.StepGraphMapEdges(unaryFn(c, fail))
				}), true
			}
		}
	case "graph_bimap":
		if in1, rest1, ok := splitParen(rest); ok {
			if in2, _, ok := splitParen(rest1); ok {
				cn, ok1 := i.graphLambda(in1, 1)
				ce, ok2 := i.graphLambda(in2, 1)
				if ok1 && ok2 {
					return latchedStep(func(fail errLatch) graphStep {
						return evaluator.StepGraphBimap(unaryFn(cn, fail), unaryFn(ce, fail))
					}), true
				}
			}
		}
	case "graph_filter_nodes":
		if inside, _, ok := splitParen(rest); ok {
			if c, ok := i.graphLambda(inside, 1); ok {
				return latchedStep(func(fail errLatch) graphStep {
					p := predFn(c, fail)
					return evaluator.StepGraphFilterNodes(func(_ string, payload any) bool { return p(payload) })
				}), true
			}
		}
	case "graph_filter_edges":
		if inside, _, ok := splitParen(rest); ok {
			if c, ok := i.graphLambda(inside, 1); ok {
				return latchedStep(func(fail errLatch) graphStep {
					p := predFn(c, fail)
					return evaluator.StepGraphFilterEdges(func(weight any, _, _, _ string) bool { return p(weight) })
				}), true
			}
		}
	case "graph_fold_nodes":
		if inside, seedStr, ok := splitParen(rest); ok {
			if c, ok := i.graphLambda(inside, 2); ok {
				return latchedStep(func(fail errLatch) graphStep {
					return evaluator.StepGraphFoldNodes(i.evalSeed(seedStr), binaryFn(c, fail))
				}), true
			}
		}
	case "graph_fold_edges":
		if inside, seedStr, ok := splitParen(rest); ok {
			if c, ok := i.graphLambda(inside, 2); ok {
				return latchedStep(func(fail errLatch) graphStep {
					return evaluator.StepGraphFoldEdges(i.evalSeed(seedStr), binaryFn(c, fail))
				}), true
			}
		}
	case "graph_fold_walk":
		if inside, tail, ok := splitParen(rest); ok {
			if c, ok := i.graphLambda(inside, 2); ok {
				tf := strings.Fields(tail)
				if len(tf) >= 2 {
					strategy := "bfs"
					if len(tf) >= 3 {
						strategy = tf[2]
					}
					return latchedStep(func(fail errLatch) graphStep {
						return evaluator.StepGraphFoldWalk(tf[1], strategy, i.evalSeed(tf[0]), binaryFn(c, fail))
					}), true
				}
			}
		}
	case "graph_over":
		// graph_over <id> (\p -> expr) — focused single-node payload rewrite
		if pidx := strings.Index(rest, "("); pidx > 0 {
			id := strings.TrimSpace(rest[:pidx])
			if inside, _, ok := splitParen(rest[pidx:]); ok && id != "" {
				if c, ok := i.graphLambda(inside, 1); ok {
					return latchedStep(func(fail errLatch) graphStep {
						return evaluator.StepGraphOver(id, unaryFn(c, fail))
					}), true
				}
			}
		}
	}
	return nil, false
}

// graphLookup builds a name resolver over the interpreter's binding environment,
// overlaid with a lambda's locally-bound variables (binds, highest precedence).
func (i *Interpreter) graphLookup(binds map[string]any) func(string) (any, bool) {
	return func(name string) (any, bool) {
		if v, ok := binds[name]; ok {
			return v, true
		}
		if rt, ok := i.RuntimeNodes[name]; ok {
			v, _ := rt.Value.UnwrapRight()
			return v, true
		}
		if v, ok := i.Constants[name]; ok {
			return v, true
		}
		if v, ok := i.Variables[name]; ok {
			return v, true
		}
		return i.exprFallback(name)
	}
}

// exprFallback resolves an identifier that is NOT a local or user binding — the
// LAST step of every expression-eval name lookup, so user bindings always win. It
// supplies two things to any expression context (fold/map/filter reducer bodies,
// comprehension sources/yields, bare top-level expressions):
//
//   - the generic expression-level builtin FUNCTIONS (graph_*/hamt_get/hamt_path, …)
//     registered in dslBuiltinFuncs, so they are callable by juxtaposition `f a b`;
//   - a bare nullary sum-type CONSTRUCTOR (e.g. the edge color `Owns` declared by
//     `type EdgeColor = Owns | …`) resolved to its tag STRING, so relation rules can
//     write `graph_add_edge Owns …` without quoting the color. This is a low-priority
//     fallback (after RuntimeNodes/Constants/Variables), and constructor resolution
//     in `match` arms is unaffected — that path is handled by the match evaluator, not
//     by this expression lookup.
func (i *Interpreter) exprFallback(name string) (any, bool) {
	if c, ok := i.defClosure(name); ok {
		return c, true
	}
	if v, ok := dslBuiltinLookup(name); ok {
		return v, true
	}
	if name == "variant" && i.Parser != nil {
		return i.variantConstructor(), true
	}
	if v, ok := parser.OrderingVariant(name); ok {
		return v, true
	}
	if i.Parser != nil && i.Parser.IsConstructor(name) {
		return name, true
	}
	return nil, false
}

// resolveGraphOperand resolves a set-algebra operand token to its value — normally
// a let-bound graph in RuntimeNodes — falling back to the raw token so the step can
// still type-check (and no-op) rather than panic.
func (i *Interpreter) resolveGraphOperand(s string) any {
	if v, err := parser.EvalInfixString(s, i.graphLookup(nil)); err == nil {
		return v
	}
	if rt, ok := i.RuntimeNodes[s]; ok {
		v, _ := rt.Value.UnwrapRight()
		return v
	}
	return s
}

// evalSeed evaluates a fold's seed token (a literal or a bound name) to a value,
// defaulting to 0 for an empty seed.
func (i *Interpreter) evalSeed(s string) any {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if v, err := parser.EvalInfixString(s, i.graphLookup(nil)); err == nil {
		return v
	}
	return parseGraphWeight(s)
}

// parseLambdaInner splits an un-parenthesised lambda body `\v1 v2 -> expr` into its
// bound variable names and expression body.
func parseLambdaInner(s string) ([]string, string, bool) {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "->") {
		return nil, "", false
	}
	parts := strings.SplitN(s, "->", 2)
	left := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(parts[0]), "\\"))
	vars := strings.Fields(left)
	if len(vars) == 0 {
		return nil, "", false
	}
	return vars, strings.TrimSpace(parts[1]), true
}

// splitParen returns the contents of the first balanced parenthesised group in s
// (nesting-aware) and the remainder after it, e.g. "(\x -> x) 0" → ("\x -> x", "0").
func splitParen(s string) (inside, rest string, ok bool) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "(") {
		return "", "", false
	}
	depth := 0
	for idx, r := range s {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return strings.TrimSpace(s[1:idx]), strings.TrimSpace(s[idx+1:]), true
			}
		}
	}
	return "", "", false
}

// mapStepOver lifts an element function into the polymorphic `map` pipeline step:
// the Functor action (mapValue) over every carrier, and — for a scalar — the function
// applied to the value itself. It is shared by the inline-lambda `map` step and by
// closure application in pipeline position, and it is the stream-level twin of the
// `map` function form. A failing element yields one error node naming that element.
func mapStepOver(evalElem func(item any, dslType dsltypes.BitFlag) (any, error)) evaluator.PipelineStep[functions.Either[error, any]] {
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
				res, err := mapValue(val, func(it any) (any, error) {
					r, e := evalElem(it, node.DSLType)
					if e != nil {
						return nil, fmt.Errorf("map: element %v: %w", it, e)
					}
					return r, nil
				})
				if _, scalar := err.(*notCarrierError); scalar {
					res, err = evalElem(val, node.DSLType)
					if err != nil {
						err = fmt.Errorf("map: %w", err)
					}
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

// filterStepOver lifts a predicate into the polymorphic `filter` pipeline step: the
// Filterable action (filterValue) over every carrier, and — for a scalar stream
// node — pass or drop the node. It is shared by inline lambdas and closure predicates.
// A predicate that fails (or is not boolean) yields an error node, never "false".
func filterStepOver(keep func(item any, dslType dsltypes.BitFlag) (bool, error)) evaluator.PipelineStep[functions.Either[error, any]] {
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
				res, err := filterValue(val, func(it any) (bool, error) {
					k, e := keep(it, node.DSLType)
					if e != nil {
						return false, fmt.Errorf("filter: element %v: %w", it, e)
					}
					return k, nil
				})
				if _, scalar := err.(*notCarrierError); scalar {
					k, e := keep(val, node.DSLType)
					if e != nil {
						if !yield(evaluator.InjectError[any](fmt.Errorf("filter: %w", e), node.DSLType)) {
							return
						}
					} else if k && !yield(node) {
						return
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

// plusBranches recognises a parenthesised `(branch <+> branch <+> …)` step and returns
// its branch pipelines. The split happens only at top-level `<+>` operators, so a
// branch may itself contain parentheses, brackets, strings and `|>` pipelines.
func plusBranches(stepStr string) ([]string, bool) {
	s := strings.TrimSpace(stepStr)
	if !strings.HasPrefix(s, "(") || !strings.HasSuffix(s, ")") || !strings.Contains(s, "<+>") {
		return nil, false
	}
	inner := s[1 : len(s)-1]
	var branches []string
	depth, start := 0, 0
	for idx := 0; idx < len(inner); idx++ {
		switch inner[idx] {
		case '"':
			for idx++; idx < len(inner) && inner[idx] != '"'; idx++ {
			}
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
			if depth < 0 {
				return nil, false // the leading '(' closed before the trailing one: not one group
			}
		case '<':
			if depth == 0 && strings.HasPrefix(inner[idx:], "<+>") {
				branches = append(branches, strings.TrimSpace(inner[start:idx]))
				idx += 2
				start = idx + 1
			}
		}
	}
	if depth != 0 || len(branches) == 0 {
		return nil, false
	}
	branches = append(branches, strings.TrimSpace(inner[start:]))
	for _, b := range branches {
		if b == "" {
			return nil, false
		}
	}
	return branches, true
}
