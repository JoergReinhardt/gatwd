package cli

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"iter"
	"os"
	"path/filepath"
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
			keyword := strings.Trim(strings.TrimSpace(stepStr[6:]), "\"")
			predicate := func(val any) bool {
				if strVal, ok := val.(string); ok {
					return strings.Contains(strVal, keyword)
				}
				return false
			}
			pipeline = append(pipeline, evaluator.StepFilter(predicate))
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
		} else if strings.HasPrefix(stepStr, "set_contains ") {
			item := strings.Trim(strings.TrimSpace(stepStr[13:]), "\"")
			pipeline = append(pipeline, evaluator.StepSetContains(item))
		} else if strings.HasPrefix(stepStr, "set_add ") {
			item := strings.Trim(strings.TrimSpace(stepStr[8:]), "\"")
			pipeline = append(pipeline, evaluator.StepSetAdd(item))
		} else if strings.HasPrefix(stepStr, "set_remove ") {
			item := strings.Trim(strings.TrimSpace(stepStr[11:]), "\"")
			pipeline = append(pipeline, evaluator.StepSetRemove(item))
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
		} else if stepStr == "conway_step" {
			pipeline = append(pipeline, evaluator.StepConway())
		} else if strings.HasPrefix(stepStr, "conway_render") {
			parts := strings.Fields(stepStr)
			mode := "braille"
			color := "cyan"
			if len(parts) > 1 {
				mode = strings.Trim(parts[1], "\"")
			}
			if len(parts) > 2 {
				color = strings.Trim(parts[2], "\"")
			}
			pipeline = append(pipeline, visualizer.StepConwayRender(mode, color))
		} else if strings.HasPrefix(stepStr, "parse_life") {
			parts := strings.Fields(stepStr)
			w, h := 80, 40
			if len(parts) > 1 {
				if val, err := strconv.Atoi(parts[1]); err == nil {
					w = val
				}
			}
			if len(parts) > 2 {
				if val, err := strconv.Atoi(parts[2]); err == nil {
					h = val
				}
			}
			pipeline = append(pipeline, evaluator.StepParseLife(w, h))
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

							// 5. String/Value substring matching
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
				innerStr := strings.TrimSpace(parts2[1])
				varName := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(parts2[0]), "\\"))
				if varName != "" {
					innerStr = strings.ReplaceAll(innerStr, varName, "%s")
				}
				subParts, _ := splitPipelineAware(innerStr)
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
	if strings.HasPrefix(firstPart, "[") && strings.HasSuffix(firstPart, "]") && !strings.Contains(firstPart, " *** ") {
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

	stringStream := i.Parser.Parse(firstPart)
	return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		for node := range stringStream {
			wrappedValue := functions.Right[error, any](any(node.Value))
			yield(dsltypes.NewASTNode(wrappedValue, node.DSLType))
		}
	}
}

func (i *Interpreter) EvaluateToValues(line string) []any {
	isParallel := strings.Contains(line, "||>")
	line = strings.ReplaceAll(line, "||>", "|>")
	parts, _ := splitPipelineAware(line)

	firstPart := parts[0]
	for k, v := range i.Variables {
		words := strings.Split(firstPart, " ")
		for idx, w := range words {
			if w == k {
				words[idx] = v
			}
		}
		firstPart = strings.Join(words, " ")
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
	for k, v := range i.Variables {
		words := strings.Split(firstPart, " ")
		for idx, w := range words {
			if w == k {
				words[idx] = v
			}
		}
		firstPart = strings.Join(words, " ")
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
	// Let Bindings: let x = ... in ...
	if strings.HasPrefix(line, "let ") {
		parts := strings.SplitN(line[4:], " in ", 2)
		if len(parts) == 2 {
			defParts := strings.SplitN(parts[0], "=", 2)
			if len(defParts) == 2 {
				varName := strings.TrimSpace(defParts[0])
				pipelineStr := strings.TrimSpace(defParts[1])

				prevVal, hadPrev := i.Variables[varName]

				// Evaluate the left side completely
				if strings.Contains(pipelineStr, "|>") || strings.HasPrefix(pipelineStr, "spawn ") {
					res := i.evaluatePipelineStr(pipelineStr)
					if len(res) > 0 {
						i.Variables[varName] = res[0] // Bind to first result
					}
				} else {
					// Direct infix expression evaluation
					lookup := func(k string) (any, bool) {
						if v, ok := i.Variables[k]; ok {
							return v, true
						}
						if v, ok := i.Constants[k]; ok {
							return v, true
						}
						return nil, false
					}
					if val, err := parser.EvalInfixString(pipelineStr, lookup); err == nil {
						i.Variables[varName] = fmt.Sprintf("%v", val)
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
	lines := strings.Split(content, "\n")
	var buffer string

	for _, rawLine := range lines {
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

		// Module Imports
		if strings.HasPrefix(buffer, "import ") {
			importFile := strings.Trim(strings.TrimSpace(buffer[7:]), "\"")
			if !filepath.IsAbs(importFile) {
				importFile = filepath.Join(baseDir, importFile)
			}
			i.LoadScript(importFile)
			buffer = ""
			continue
		}

		i.EvaluateLine(buffer)
		buffer = ""
	}
	if buffer != "" {
		i.EvaluateLine(buffer)
	}
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
		// Validate lines
		hasError := false
		lines := strings.Split(string(content), "\n")
		for idx, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "--") {
				continue
			}
			// Infix syntax validation check
			if strings.Contains(line, ":=") && !strings.Contains(line, "ebnf_parse") {
				fmt.Fprintf(os.Stderr, "%s:%d:1: error: invalid syntax ':='\n", targetFile, idx+1)
				hasError = true
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
