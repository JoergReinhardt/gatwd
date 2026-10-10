package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/interactive"
	"github.com/joergreinhardt/gatwd/parser"
)

type Interpreter struct {
	Engine    *evaluator.Engine[functions.Either[error, any]]
	Parser    *parser.Parser
	Funcs     map[string]string // Maps function name to pipeline string
	Variables map[string]string // Maps local 'let' variables
}

func NewInterpreter() *Interpreter {
	i := &Interpreter{
		Engine:    evaluator.NewEngine[functions.Either[error, any]](),
		Parser:    parser.NewParser(),
		Funcs:     make(map[string]string),
		Variables: make(map[string]string),
	}
	i.Engine.OnMissingAdapter = func(flag dsltypes.BitFlag) evaluator.Adapter[functions.Either[error, any]] {
		// Suppress REPL pause in batch mode; just return the exec adapter
		return evaluator.EitherT(interactive.GenericExecAdapter("", flag))
	}
	return i
}

func (i *Interpreter) parsePipelineSteps(parts []string) []evaluator.PipelineStep[functions.Either[error, any]] {
	var pipeline []evaluator.PipelineStep[functions.Either[error, any]]

	for _, stepStr := range parts {
		stepStr = strings.TrimSpace(stepStr)
		if stepStr == "" { continue }

		// Variable Substitution
		for k, v := range i.Variables {
			// Naive word replacement for variables
			words := strings.Split(stepStr, " ")
			for idx, w := range words {
				if w == k { words[idx] = v }
			}
			stepStr = strings.Join(words, " ")
		}

		// Function Substitution (Macros)
		if funcPipeline, exists := i.Funcs[stepStr]; exists {
			funcParts := strings.Split(funcPipeline, "|>")
			pipeline = append(pipeline, i.parsePipelineSteps(funcParts)...)
			continue
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
		} else if strings.HasPrefix(stepStr, "filter") {
			keyword := strings.Trim(strings.TrimSpace(stepStr[6:]), "\"")
			predicate := func(val any) bool {
				if strVal, ok := val.(string); ok { return strings.Contains(strVal, keyword) }
				return false
			}
			pipeline = append(pipeline, evaluator.StepFilter(predicate))
		} else if strings.HasPrefix(stepStr, "expect") {
			for res := range parser.ParseExpect()(stepStr) {
				keyword := res.Fst.Fst
				errMsg := res.Fst.Snd
				predicate := func(val any) bool {
					if strVal, ok := val.(string); ok { return strings.Contains(strVal, keyword) }
					return false
				}
				pipeline = append(pipeline, evaluator.StepExpect(predicate, fmt.Errorf("%s", errMsg)))
				break
			}
		} else if stepStr == "distinct" {
			pipeline = append(pipeline, evaluator.StepDistinct())
		} else if strings.HasPrefix(stepStr, "fold") {
			op := strings.TrimSpace(stepStr[4:])
			pipeline = append(pipeline, evaluator.StepFold(op))
		} else if strings.HasPrefix(stepStr, "memoize") {
			cmd := strings.TrimSpace(stepStr[7:])
			adapter := evaluator.EitherT(interactive.GenericExecAdapter(cmd, 0))
			pipeline = append(pipeline, evaluator.StepMemoize(adapter))
		} else if strings.HasPrefix(stepStr, "match") {
			cases := make(map[string]evaluator.Adapter[functions.Either[error, any]])
			block := strings.TrimSpace(stepStr[5:])
			block = strings.TrimPrefix(block, "{")
			block = strings.TrimSuffix(block, "}")
			rules := strings.Split(block, "|")
			for _, rule := range rules {
				parts := strings.Split(rule, "->")
				if len(parts) == 2 {
					pattern := strings.TrimSpace(parts[0])
					cmd := strings.TrimSpace(parts[1])
					cases[pattern] = evaluator.EitherT(interactive.GenericExecAdapter(cmd, 0))
				}
			}
			pipeline = append(pipeline, evaluator.StepMatch(cases, nil))
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
				leftAdapter := evaluator.EitherT(interactive.GenericExecAdapter(res.Fst.Fst, 0))
				rightAdapter := evaluator.EitherT(interactive.GenericExecAdapter(res.Fst.Snd, 0))
				pipeline = append(pipeline, evaluator.StepFanOut(leftAdapter, rightAdapter))
				break
			}
		} else if strings.HasPrefix(stepStr, "[") && strings.Contains(stepStr, "***") {
			stepStr = strings.TrimSpace(stepStr)
			stepStr = strings.TrimPrefix(stepStr, "[")
			stepStr = strings.TrimSuffix(stepStr, "]")
			parts2 := strings.Split(stepStr, "***")
			if len(parts2) == 2 {
				leftAdapter := evaluator.EitherT(interactive.GenericExecAdapter(strings.TrimSpace(parts2[0]), 0))
				rightAdapter := evaluator.EitherT(interactive.GenericExecAdapter(strings.TrimSpace(parts2[1]), 0))
				pipeline = append(pipeline, evaluator.StepSplit(leftAdapter, rightAdapter))
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
		} else if stepStr == "route" {
				pipeline = append(pipeline, evaluator.StepRoute(i.Engine))
			} else if stepStr == "collect_list" {
			pipeline = append(pipeline, evaluator.StepCollectList())
		} else if strings.HasPrefix(stepStr, "build_tree") {
			rootName := strings.Trim(strings.TrimSpace(stepStr[10:]), "\"")
			pipeline = append(pipeline, evaluator.StepBuildTree(rootName))
		} else if stepStr == "into_zipper" || stepStr == "down" || stepStr == "up" || stepStr == "extract" {
			pipeline = append(pipeline, evaluator.StepZipper(stepStr))
		} else if strings.HasPrefix(stepStr, "map") {
			// Lambda: map (\x -> x + 2)
			// For OS evaluation, we just replace \x -> echo x with generic exec adapter
			lambda := strings.TrimSpace(stepStr[3:])
			lambda = strings.TrimPrefix(lambda, "(")
			lambda = strings.TrimSuffix(lambda, ")")
			parts2 := strings.Split(lambda, "->")
			if len(parts2) == 2 {
				// We don't mathematically evaluate, we just run OS command replacing var with %s
				cmd := strings.TrimSpace(parts2[1])
				varName := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(parts2[0]), "\\"))
				cmd = strings.ReplaceAll(cmd, varName, "%s")
				adapter := evaluator.EitherT(interactive.GenericExecAdapter(cmd, 0))
				pipeline = append(pipeline, evaluator.StepAdapter(adapter))
			}
		} else {
			adapter := evaluator.EitherT(interactive.GenericExecAdapter(stepStr, 0))
			pipeline = append(pipeline, evaluator.StepAdapter(adapter))
		}
	}
	return pipeline
}

func (i *Interpreter) evaluatePipelineStr(line string) []string {
	isParallel := strings.Contains(line, "||>")
	line = strings.ReplaceAll(line, "||>", "|>")
	parts := strings.Split(line, "|>")
	
	stringStream := i.Parser.Parse(parts[0])
	initialStream := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		for node := range stringStream {
			wrappedValue := functions.Right[error, any](any(node.Value))
			yield(dsltypes.NewASTNode(wrappedValue, node.DSLType))
		}
	}

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
			results = append(results, fmt.Sprintf("%+v", val))
		}
	}
	return results
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
				
				// Evaluate the left side completely
				res := i.evaluatePipelineStr(pipelineStr)
				if len(res) > 0 {
					i.Variables[varName] = res[0] // Bind to first result
				}
				
				// Evaluate the right side
				i.EvaluateLine(parts[1])
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
	if len(results) == 0 && !strings.HasPrefix(line, "type ") && !strings.HasPrefix(line, "import ") {
		fmt.Println("(No output or node unhandled)")
	}
}

func (i *Interpreter) LoadScript(filename string) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file '%s': %v\n", filename, err)
		return
	}
	
	dir := filepath.Dir(filename)
	
	lines := strings.Split(string(bytes), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") { continue }
		
		// Module Imports
		if strings.HasPrefix(line, "import ") {
			importFile := strings.Trim(strings.TrimSpace(line[7:]), "\"")
			if !filepath.IsAbs(importFile) {
				importFile = filepath.Join(dir, importFile)
			}
			i.LoadScript(importFile)
			continue
		}

		i.EvaluateLine(line)
	}
}

func main() {
	fmt.Println("=== gatwd Interpreter (V2 Modernized) ===")
	fmt.Println("UTF-8 Mathematical Aliases Enabled (e.g. ⊳, →, ∷)")

	interpreter := NewInterpreter()

	interactiveMode := len(os.Args) == 1
	var scriptToLoad string
	if len(os.Args) > 1 {
		if os.Args[1] == "-i" && len(os.Args) > 2 {
			interactiveMode = true
			scriptToLoad = os.Args[2]
		} else {
			interactiveMode = false
			scriptToLoad = os.Args[1]
		}
	}

	if scriptToLoad != "" {
		if !interactiveMode { fmt.Printf("Running Batch Mode: %s\n", scriptToLoad) }
		interpreter.LoadScript(scriptToLoad)
		if !interactiveMode { return }
	}

	fmt.Println("Running Interactive Mode. Type ':load <file>' to load a script, or 'exit' to quit.")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("gatwd ⊳ ")
		if !scanner.Scan() { break }
		line := strings.TrimSpace(scanner.Text())
		if line == "exit" || line == "quit" { break }
		if line == "" || strings.HasPrefix(line, "--") { continue }

		if strings.HasPrefix(line, ":load ") {
			interpreter.LoadScript(strings.TrimSpace(line[6:]))
			continue
		}
		
		interpreter.EvaluateLine(line)
	}
}
