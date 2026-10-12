package cli

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"iter"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
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
	Engine         *evaluator.Engine[functions.Either[error, any]]
	Parser         *parser.Parser
	Funcs          map[string]string // Maps function name to pipeline string
	RuntimeNodes   map[string]dsltypes.ASTNode[functions.Either[error, any]]
	Variables      map[string]string // Maps local 'let' variables
	Constants      map[string]any    // Registered and script-defined constants
	CustomSteps    []StepHandler     // Pluggable extension step handlers
	EnableBuiltins bool              // Whether Go std builtins are active
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
		EnableBuiltins: false,
	}
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
	localInterp.CustomSteps = i.CustomSteps
	localInterp.Variables = make(map[string]string)
	for k, v := range i.Variables {
		localInterp.Variables[k] = v
	}
	localInterp.Funcs = i.Funcs
	localInterp.Constants = i.Constants
	localInterp.RuntimeNodes = make(map[string]dsltypes.ASTNode[functions.Either[error, any]])
	for k, v := range i.RuntimeNodes {
		localInterp.RuntimeNodes[k] = v
	}
	return localInterp
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
					// Compile the function body lazily at runtime
					funcPipeline := i.Funcs[funcName]
					funcParts, _ := splitPipelineAware(funcPipeline)
					compiledSteps := i.parsePipelineSteps(funcParts)

					// Evaluate the compiled pipeline over the incoming stream
					eng := evaluator.NewEngine[functions.Either[error, any]]()
					outStream := eng.EvaluatePipeline(stream, compiledSteps)
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
								bindAdapter := localInterp.pipelineToAdapter(bindingPipelineStr)
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
						return nil, false
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
				return nil, false
			}
			if strings.ContainsAny(stepStr, "+-*/%&|<>= ") || (len(stepStr) > 0 && unicode.IsDigit(rune(stepStr[0]))) {
				if _, err := parser.EvalInfixString(stepStr, lookup); err == nil {
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
		} else if strings.HasPrefix(stepStr, "filter") {
			rawArg := strings.TrimSpace(stepStr[6:])
			arg := rawArg
			if strings.HasPrefix(arg, "(") && strings.HasSuffix(arg, ")") {
				arg = strings.TrimSuffix(strings.TrimPrefix(arg, "("), ")")
			}
			if strings.Contains(arg, "->") {
				parts2 := strings.SplitN(arg, "->", 2)
				varName := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(parts2[0]), "\\"))
				exprBody := strings.TrimSpace(parts2[1])

				evalPred := func(item any, dslType dsltypes.BitFlag) bool {
					lookup := func(name string) (any, bool) {
						if name == varName {
							return item, true
						}
						if rt, ok := i.RuntimeNodes[name]; ok {
							v, _ := rt.Value.UnwrapRight()
							return v, true
						}
						if c, ok := i.Constants[name]; ok {
							return c, true
						}
						if v, ok := i.Variables[name]; ok {
							return v, true
						}
						return nil, false
					}
					if parsedExpr, err := parser.ParseInfixExpr(exprBody); err == nil {
						if resVal, errEval := parsedExpr.Eval(lookup); errEval == nil {
							if b, ok := resVal.(bool); ok {
								return b
							}
						}
					}
					return false
				}

				filterStep := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
					return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
						for node := range stream {
							if node.Value.IsLeft() {
								if !yield(node) {
									return
								}
								continue
							}
							val, _ := node.Value.UnwrapRight()

							// If val is a List:
							if l, ok := val.(*collections.List[any]); ok {
								var kept []any
								for it := range l.Seq() {
									if evalPred(it, node.DSLType) {
										kept = append(kept, it)
									}
								}
								resList := collections.Nil[any]()
								for idx := len(kept) - 1; idx >= 0; idx-- {
									resList = collections.Cons(kept[idx], resList)
								}
								if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(resList)), node.DSLType)) {
									return
								}
								continue
							}

							// If val is a Chunk:
							if c, ok := val.(*collections.Chunk[any]); ok {
								resChunk := collections.NewChunk[any](c.Capacity())
								for _, it := range c.Items() {
									if evalPred(it, node.DSLType) {
										resChunk.Append(it)
									}
								}
								if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(resChunk)), node.DSLType)) {
									return
								}
								continue
							}

							// If val is a Set:
							if _, isSet := val.(interface{ IsSet() bool }); isSet {
								if seq, okSeq := functions.ToSliceAny(val); okSeq {
									resSet := collections.NewSet[any]()
									for _, it := range seq {
										if evalPred(it, node.DSLType) {
											resSet = resSet.Add(it)
										}
									}
									if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(resSet)), node.DSLType)) {
										return
									}
									continue
								}
							}

							// If val is a slice:
							if sl, okSl := val.([]any); okSl {
								var newSl []any
								for _, it := range sl {
									if evalPred(it, node.DSLType) {
										newSl = append(newSl, it)
									}
								}
								if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(newSl)), node.DSLType)) {
									return
								}
								continue
							}

							// Stream node
							if evalPred(val, node.DSLType) {
								if !yield(node) {
									return
								}
							}
						}
					}
				}
				pipeline = append(pipeline, filterStep)
			} else {
				keyword := strings.Trim(rawArg, "\"")
				predicate := func(val any) bool {
					if strVal, ok := val.(string); ok {
						return strings.Contains(strVal, keyword)
					}
					return false
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
			pipeline = append(pipeline, evaluator.StepFold(op))
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
			var trimmed []string
			for _, p := range parts {
				trimmed = append(trimmed, strings.TrimSpace(p))
			}
			pipeline = append(pipeline, evaluator.StepEmit(trimmed...))
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
					return nil, false
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
				return nil, false
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
				return nil, false
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
				return nil, false
			}
			if evaled, err := parser.EvalInfixString(operandStr, lookup); err == nil {
				pipeline = append(pipeline, evaluator.StepHAMTMerge(evaled))
			} else if rt, ok := i.RuntimeNodes[operandStr]; ok {
				v, _ := rt.Value.UnwrapRight()
				pipeline = append(pipeline, evaluator.StepHAMTMerge(v))
			} else {
				pipeline = append(pipeline, evaluator.StepHAMTMerge(operandStr))
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
		} else if strings.HasPrefix(stepStr, "math_mod") {
			amount := strings.TrimSpace(stepStr[9:])
			pipeline = append(pipeline, evaluator.StepMathMod(amount))
		} else if strings.HasPrefix(stepStr, "+ ") {
			pipeline = append(pipeline, evaluator.StepBinaryOp(functions.OpAdd, strings.TrimSpace(stepStr[2:])))
		} else if strings.HasPrefix(stepStr, "- ") {
			pipeline = append(pipeline, evaluator.StepBinaryOp(functions.OpSub, strings.TrimSpace(stepStr[2:])))
		} else if strings.HasPrefix(stepStr, "* ") {
			pipeline = append(pipeline, evaluator.StepBinaryOp(functions.OpMul, strings.TrimSpace(stepStr[2:])))
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
		} else if strings.HasPrefix(stepStr, "map") {
			lambda := strings.TrimSpace(stepStr[3:])
			lambda = strings.TrimPrefix(lambda, "(")
			lambda = strings.TrimSuffix(lambda, ")")
			parts2 := strings.SplitN(lambda, "->", 2)
			if len(parts2) == 2 {
				varName := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(parts2[0]), "\\"))
				exprBody := strings.TrimSpace(parts2[1])

				evalElem := func(item any, dslType dsltypes.BitFlag) (any, error) {
					lookup := func(name string) (any, bool) {
						if name == varName {
							return item, true
						}
						if rt, ok := i.RuntimeNodes[name]; ok {
							v, _ := rt.Value.UnwrapRight()
							return v, true
						}
						if c, ok := i.Constants[name]; ok {
							return c, true
						}
						if v, ok := i.Variables[name]; ok {
							return v, true
						}
						return nil, false
					}
					// 1. Try Infix parser
					if parsedExpr, err := parser.ParseInfixExpr(exprBody); err == nil {
						if resVal, errEval := parsedExpr.Eval(lookup); errEval == nil {
							return resVal, nil
						}
					}
					// 2. Try sub-pipeline evaluation
					childInterp := &Interpreter{
						Engine:         i.Engine,
						EnableBuiltins: i.EnableBuiltins,
						Funcs:          i.Funcs,
						RuntimeNodes:   make(map[string]dsltypes.ASTNode[functions.Either[error, any]]),
						Variables:      make(map[string]string),
						Constants:      i.Constants,
						CustomSteps:    i.CustomSteps,
					}
					for k, v := range i.Variables {
						childInterp.Variables[k] = v
					}
					for k, v := range i.RuntimeNodes {
						childInterp.RuntimeNodes[k] = v
					}
					childInterp.RuntimeNodes[varName] = dsltypes.NewASTNode(functions.Right[error, any](item), dslType)
					childInterp.Variables[varName] = fmt.Sprintf("%v", item)

					subParts, _ := splitPipelineAware(exprBody)
					subPipeline := childInterp.parsePipelineSteps(subParts)
					singleton := func(subYield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
						subYield(dsltypes.NewASTNode(functions.Right[error, any](item), dslType))
					}
					subRes := childInterp.Engine.EvaluatePipeline(singleton, subPipeline)
					for r := range subRes {
						if r.Value.IsRight() {
							v, _ := r.Value.UnwrapRight()
							return v, nil
						} else {
							err, _ := r.Value.UnwrapLeft()
							return nil, err
						}
					}
					return item, nil
				}

				mapStep := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
					return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
						for node := range stream {
							if node.Value.IsLeft() {
								if !yield(node) {
									return
								}
								continue
							}
							val, _ := node.Value.UnwrapRight()

							// If val is a List collection:
							if l, ok := val.(*collections.List[any]); ok {
								var newItems []any
								for it := range l.Seq() {
									mVal, _ := evalElem(it, node.DSLType)
									newItems = append(newItems, mVal)
								}
								resList := collections.Nil[any]()
								for idx := len(newItems) - 1; idx >= 0; idx-- {
									resList = collections.Cons(newItems[idx], resList)
								}
								if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(resList)), node.DSLType)) {
									return
								}
								continue
							}
							if _, isList := val.(interface{ IsList() bool }); isList {
								if seq, okSeq := functions.ToSliceAny(val); okSeq {
									var newItems []any
									for _, it := range seq {
										mVal, _ := evalElem(it, node.DSLType)
										newItems = append(newItems, mVal)
									}
									resList := collections.Nil[any]()
									for idx := len(newItems) - 1; idx >= 0; idx-- {
										resList = collections.Cons(newItems[idx], resList)
									}
									if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(resList)), node.DSLType)) {
										return
									}
									continue
								}
							}

							// If val is a Chunk:
							if c, ok := val.(*collections.Chunk[any]); ok {
								resChunk := collections.NewChunk[any](c.Capacity())
								for _, it := range c.Items() {
									mVal, _ := evalElem(it, node.DSLType)
									resChunk.Append(mVal)
								}
								if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(resChunk)), node.DSLType)) {
									return
								}
								continue
							}
							if _, isChunk := val.(interface{ IsChunk() bool }); isChunk {
								if seq, okSeq := functions.ToSliceAny(val); okSeq {
									resChunk := collections.NewChunk[any](len(seq))
									for _, it := range seq {
										mVal, _ := evalElem(it, node.DSLType)
										resChunk.Append(mVal)
									}
									if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(resChunk)), node.DSLType)) {
										return
									}
									continue
								}
							}

							// If val is a Set:
							if _, isSet := val.(interface{ IsSet() bool }); isSet {
								if seq, okSeq := functions.ToSliceAny(val); okSeq {
									resSet := collections.NewSet[any]()
									for _, it := range seq {
										mVal, _ := evalElem(it, node.DSLType)
										resSet = resSet.Add(mVal)
									}
									if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(resSet)), node.DSLType)) {
										return
									}
									continue
								}
							}

							// If val is a HAMT:
							if h, ok := val.(*collections.HAMT[string, any]); ok {
								resHamt := collections.NewHAMT[string, any]()
								for k, v := range h.Seq() {
									mVal, _ := evalElem(v, node.DSLType)
									resHamt = resHamt.Set(k, mVal)
								}
								if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(resHamt)), node.DSLType)) {
									return
								}
								continue
							}

							// If val is a Tree:
							if tr, ok := val.(*collections.Tree[any]); ok {
								var mapTree func(t *collections.Tree[any]) *collections.Tree[any]
								mapTree = func(t *collections.Tree[any]) *collections.Tree[any] {
									if t == nil {
										return nil
									}
									mV, _ := evalElem(t.Value(), node.DSLType)
									var newChList *collections.List[*collections.Tree[any]] = collections.Nil[*collections.Tree[any]]()
									if t.Children() != nil {
										var chs []*collections.Tree[any]
										for c := range t.Children().Seq() {
											chs = append(chs, mapTree(c))
										}
										for i := len(chs) - 1; i >= 0; i-- {
											newChList = collections.Cons(chs[i], newChList)
										}
									}
									return collections.NewTree(mV, newChList)
								}
								if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(mapTree(tr))), node.DSLType)) {
									return
								}
								continue
							}

							// If val is a Pair:
							if p, ok := val.(interface{ PairElements() (any, any) }); ok {
								f, s := p.PairElements()
								mS, _ := evalElem(s, node.DSLType)
								newPair := functions.Pair[any, any]{Fst: f, Snd: mS}
								if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(newPair)), node.DSLType)) {
									return
								}
								continue
							}

							// If val is a Maybe:
							if mb, ok := val.(interface{ MaybeValue() (any, bool) }); ok {
								v, isJ := mb.MaybeValue()
								if isJ {
									mV, _ := evalElem(v, node.DSLType)
									newMb := functions.Just(mV)
									if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(newMb)), node.DSLType)) {
										return
									}
								} else {
									if !yield(dsltypes.NewASTNode(functions.Right[error, any](val), node.DSLType)) {
										return
									}
								}
								continue
							}

							// If val is a standard Go slice:
							if sl, okSl := val.([]any); okSl {
								var newSl []any
								for _, it := range sl {
									mVal, _ := evalElem(it, node.DSLType)
									newSl = append(newSl, mVal)
								}
								if !yield(dsltypes.NewASTNode(functions.Right[error, any](any(newSl)), node.DSLType)) {
									return
								}
								continue
							}

							// Scalar / individual stream node
							resVal, errEval := evalElem(val, node.DSLType)
							if errEval != nil {
								if !yield(evaluator.InjectError[any](errEval, node.DSLType)) {
									return
								}
							} else {
								if !yield(dsltypes.NewASTNode(functions.Right[error, any](resVal), node.DSLType)) {
									return
								}
							}
						}
					}
				}
				pipeline = append(pipeline, mapStep)
			} else {
				targetStr := strings.TrimSpace(lambda)
				subParts, _ := splitPipelineAware(targetStr)
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
	for i, r := range line {
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

func splitCasesAware(line string) []string {
	var parts []string
	var current strings.Builder
	depthParen := 0
	depthBrace := 0
	depthBracket := 0

	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

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

func splitPipelineAware(line string) ([]string, bool) {
	var parts []string
	var current strings.Builder
	depthParen := 0
	depthBrace := 0
	depthBracket := 0
	isParallel := false

	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

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
				current.Reset()
				i++ // skip '>'
				continue
			} else if r == '|' && i+2 < len(runes) && runes[i+1] == '|' && runes[i+2] == '>' {
				parts = append(parts, strings.TrimSpace(current.String()))
				current.Reset()
				i += 2 // skip '|>'
				isParallel = true
				continue
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}

	return parts, isParallel
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
		return nil, false
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

	if strings.HasPrefix(firstPart, "(") || (strings.ContainsAny(firstPart, "+-*/%&|<>= ") && !strings.HasPrefix(firstPart, "emit ")) {
		if val, err := parser.EvalInfixString(firstPart, lookup); err == nil {
			flag := i.Parser.GetFlag("Variable")
			return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
				yield(dsltypes.NewASTNode(functions.Right[error, any](val), flag))
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
					return nil, false
				}

				cleanPipeline := pipelineStr
				if strings.HasPrefix(cleanPipeline, "(") && strings.HasSuffix(cleanPipeline, ")") {
					cleanPipeline = strings.TrimSpace(cleanPipeline[1 : len(cleanPipeline)-1])
				}
				if strings.Contains(cleanPipeline, "|>") || strings.HasPrefix(cleanPipeline, "spawn ") {
					res := i.EvaluateToValues(cleanPipeline)
					if len(res) > 0 {
						val := res[0]
						i.Variables[varName] = fmt.Sprintf("%v", val)
						i.RuntimeNodes[varName] = dsltypes.NewASTNode(functions.Right[error, any](val), i.Parser.GetFlag("Variable"))
					}
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

	isParallel := strings.Contains(line, "||>")
	line = strings.ReplaceAll(line, "||>", "|>")
	parts, _ := splitPipelineAware(line)

	firstPart := parts[0]
	if _, isVar := i.Variables[firstPart]; !isVar {
		if _, isRt := i.RuntimeNodes[firstPart]; !isRt {
			for k, v := range i.Variables {
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
		words := strings.Split(firstPart, " ")
		for idx, w := range words {
			if w == k {
				words[idx] = fmt.Sprintf("%v", v)
			}
		}
		firstPart = strings.Join(words, " ")
	}

	initialStream := i.buildInitialStream(firstPart)

	pipeline := i.parsePipelineSteps(parts[1:])
	terminalNodes := i.Engine.EvaluatePipeline(initialStream, pipeline)

	if isParallel {
		terminalNodes = i.Engine.ParEvaluate(terminalNodes, 10)
	}

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
	parts, isParallel := splitPipelineAware(line)

	firstPart := parts[0]
	if _, isVar := i.Variables[firstPart]; !isVar {
		if _, isRt := i.RuntimeNodes[firstPart]; !isRt {
			for k, v := range i.Variables {
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
					words[idx] = fmt.Sprintf("%v", v)
				}
			}
			firstPart = strings.Join(words, " ")
		}
	}

	initialStream := i.buildInitialStream(firstPart)

	pipeline := i.parsePipelineSteps(parts[1:])
	terminalNodes := i.Engine.EvaluatePipeline(initialStream, pipeline)

	if isParallel {
		terminalNodes = i.Engine.ParEvaluate(terminalNodes, 10)
	}

	var results []string
	for node := range terminalNodes {
		if node.Value.IsLeft() {
			err, _ := node.Value.UnwrapLeft()
			results = append(results, fmt.Sprintf("[ERROR] %v", err))
		} else {
			val, _ := node.Value.UnwrapRight()
			if b, ok := val.([]byte); ok {
				results = append(results, string(b))
			} else {
				results = append(results, fmt.Sprintf("%+v", val))
			}
		}
	}
	return results
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
					return nil, false
				}

				// Evaluate the left side completely
				cleanPipeline := pipelineStr
				if strings.HasPrefix(cleanPipeline, "(") && strings.HasSuffix(cleanPipeline, ")") {
					cleanPipeline = strings.TrimSpace(cleanPipeline[1 : len(cleanPipeline)-1])
				}
				if strings.Contains(cleanPipeline, "|>") || strings.HasPrefix(cleanPipeline, "spawn ") {
					res := i.EvaluateToValues(cleanPipeline)
					if len(res) > 0 {
						val := res[0]
						i.Variables[varName] = fmt.Sprintf("%v", val)
						i.RuntimeNodes[varName] = dsltypes.NewASTNode(functions.Right[error, any](val), i.Parser.GetFlag("Variable"))
					}
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
					return nil, false
				}

				cleanPipeline := pipelineStr
				if strings.HasPrefix(cleanPipeline, "(") && strings.HasSuffix(cleanPipeline, ")") {
					cleanPipeline = strings.TrimSpace(cleanPipeline[1 : len(cleanPipeline)-1])
				}
				if strings.Contains(cleanPipeline, "|>") || strings.HasPrefix(cleanPipeline, "spawn ") {
					res := i.EvaluateToValues(cleanPipeline)
					if len(res) > 0 {
						val := res[0]
						i.Variables[varName] = fmt.Sprintf("%v", val)
						i.RuntimeNodes[varName] = dsltypes.NewASTNode(functions.Right[error, any](val), i.Parser.GetFlag("Variable"))
					}
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
			valStr = strings.Trim(valStr, "\"")
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
					i.Constants[constName] = fmt.Sprintf("%v", evaluated)
				} else {
					i.Constants[constName] = valStr
				}
			} else {
				i.Constants[constName] = valStr
			}
		}
		return
	}

	// Function Definitions: def func = ...
	if strings.HasPrefix(line, "def ") {
		parts := strings.SplitN(line[4:], "=", 2)
		if len(parts) == 2 {
			funcName := strings.TrimSpace(parts[0])
			pipelineStr := strings.TrimSpace(parts[1])
			i.Funcs[funcName] = pipelineStr
		}
		return
	}

	results := i.evaluatePipelineStr(line)
	for _, res := range results {
		fmt.Printf("∷ %s\n", res)
	}
	if len(results) == 0 && !strings.HasPrefix(line, "type ") && !strings.HasPrefix(line, "import ") && !strings.HasPrefix(line, "const ") {
		fmt.Println("(No output or node unhandled)")
	}
}

// EvaluateScript evaluates an entire multi-line script string, resolving comments,
// multi-line expressions, macros, and pipeline definitions sequentially.
func (i *Interpreter) EvaluateScript(script string) {
	i.LoadScriptContent(script, ".")
}

// LoadScriptContent parses and evaluates a script string with a base directory for resolving relative imports.
func (i *Interpreter) LoadScriptContent(content string, baseDir string) {
	for _, stmt := range logicalLines(content) {
		buffer := stmt.text
		// Module Imports
		if strings.HasPrefix(buffer, "import ") {
			importFile := strings.Trim(strings.TrimSpace(buffer[7:]), "\"")
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
			if strings.Count(buffer, "\"")%2 != 0 {
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
			strings.Count(buffer, "{") > strings.Count(buffer, "}") ||
			strings.Count(buffer, "(") > strings.Count(buffer, ")") ||
			strings.Count(buffer, "[") > strings.Count(buffer, "]") ||
			strings.Count(buffer, "\"")%2 != 0 {
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
		fmt.Printf("Error reading file '%s': %v\n", filename, err)
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
	i.LoadScriptContent(string(bytes), filepath.Dir(filename))
	return nil
}

// RunWithInterpreter executes batch script execution or interactive REPL using the supplied Interpreter.
func RunWithInterpreter(interpreter *Interpreter) {
	fmt.Println("=== gatwd Interpreter (V2 Modernized) ===")
	fmt.Println("UTF-8 Mathematical Aliases Enabled (e.g. ⊳, →, ∷)")

	if len(os.Args) > 1 && (os.Args[1] == "--check" || os.Args[1] == "-c") {
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: gatwd --check <file.gatwd>")
			os.Exit(1)
		}
		targetFile := os.Args[2]
		content, err := os.ReadFile(targetFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s:1:1: error: cannot read file: %v\n", targetFile, err)
			os.Exit(1)
		}
		// Validate statically over logical statements (physical lines joined the
		// same way the evaluator joins them). Only structural, zero-false-positive
		// checks are reported — not type errors, which gatwd resolves at runtime.
		stmts := logicalLines(string(content))

		// Pre-register the script's sum types so exhaustiveness checks know each
		// closed type's full set of constructors.
		checkParser := parser.NewParser()
		for _, stmt := range stmts {
			if strings.HasPrefix(stmt.text, "type ") {
				for range checkParser.Parse(stmt.text) {
				}
			}
		}

		hasError := false
		for _, stmt := range stmts {
			line := stmt.text
			report := func(col int, msg string) {
				fmt.Fprintf(os.Stderr, "%s:%d:%d: error: %s\n", targetFile, stmt.num, col, msg)
				hasError = true
			}
			warn := func(msg string) {
				fmt.Fprintf(os.Stderr, "%s:%d:1: warning: %s\n", targetFile, stmt.num, msg)
			}
			checkExhaustiveMatch(checkParser, line, warn)
			// 1. Legacy assignment operator.
			if strings.Contains(line, ":=") && !strings.Contains(line, "ebnf_parse") {
				report(strings.Index(line, ":=")+1, "invalid assignment ':=' (use '=')")
			}
			// 2. Unbalanced (), [], or {} (string literals are ignored).
			if col, ok := unbalancedDelimiter(line); !ok {
				report(col+1, "unbalanced delimiter")
			}
			// 3. A declaration keyword missing its '='.
			for _, kw := range []string{"def ", "const ", "type "} {
				if strings.HasPrefix(line, kw) && !strings.Contains(line, "=") {
					report(1, "'"+strings.TrimSpace(kw)+"' declaration is missing '='")
				}
			}
			// 4. A trailing pipe with no following stage (unterminated at EOF).
			if strings.HasSuffix(line, "|>") {
				report(len(line)-1, "pipeline ends with '|>' but has no following stage")
			}
		}
		if hasError {
			os.Exit(1)
		}
		// Clean check
		os.Exit(0)
	}

	interactiveMode := len(os.Args) == 1
	var scriptToLoad string
	if len(os.Args) > 1 {
		if os.Args[1] == "-i" {
			interactiveMode = true
			if len(os.Args) > 2 {
				scriptToLoad = os.Args[2]
			}
		} else {
			interactiveMode = false
			scriptToLoad = os.Args[1]
		}
	}

	if scriptToLoad != "" {
		if !interactiveMode {
			fmt.Printf("Running Batch Mode: %s\n", scriptToLoad)
		}
		interpreter.LoadScript(scriptToLoad)
		if !interactiveMode {
			return
		}
	}

	fmt.Println("Running Interactive Mode. Type ':load <file>' to load a script, or 'exit' to quit.")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("gatwd ⊳ ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "exit" || line == "quit" {
			break
		}
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		if strings.HasPrefix(line, ":load ") {
			interpreter.LoadScript(strings.TrimSpace(line[6:]))
			continue
		}

		// Handle REPL introspection commands
		if line == ":macros" {
			fmt.Println("Registered Macros:")
			for k, v := range interpreter.Funcs {
				fmt.Printf("  %s = %s\n", k, v)
			}
			continue
		}
		if line == ":vars" {
			fmt.Println("Registered Bindings (let):")
			for k, v := range interpreter.Variables {
				fmt.Printf("  %s = %s\n", k, v)
			}
			continue
		}

		interpreter.EvaluateLine(line)
	}
}

// Run executes the default CLI using the full NewInterpreter() with all builtins.
func Run() {
	RunWithInterpreter(NewInterpreter())
}
