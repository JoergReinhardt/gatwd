package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/interactive"
	"github.com/joergreinhardt/gatwd/parser"
)

func main() {
	fmt.Println("=== gatwd Interpreter (V2 Modernized) ===")
	fmt.Println("UTF-8 Mathematical Aliases Enabled (e.g. ⊳, →, ∷)")

	p := parser.NewParser()
	engine := evaluator.NewEngine[functions.Either[error, any]]()

	engine.OnMissingAdapter = func(flag dsltypes.BitFlag) evaluator.Adapter[functions.Either[error, any]] {
		fmt.Printf("\n[REPL] Encountered unknown node type %d. Enter CLI command: ", flag)
		reader := bufio.NewReader(os.Stdin)
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		if cmd == "" { return nil }
		return evaluator.EitherT(interactive.GenericExecAdapter(cmd, 0))
	}

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

	evaluateLine := func(line string) {
		isParallel := strings.Contains(line, "||>")
		line = strings.ReplaceAll(line, "||>", "|>")
		parts := strings.Split(line, "|>")
		
		stringStream := p.Parse(parts[0])
		initialStream := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stringStream {
				wrappedValue := functions.Right[error, any](any(node.Value))
				yield(dsltypes.NewASTNode(wrappedValue, node.DSLType))
			}
		}

		var pipeline []evaluator.PipelineStep[functions.Either[error, any]]

		for _, stepStr := range parts[1:] {
			stepStr = strings.TrimSpace(stepStr)
			
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
			} else if strings.HasPrefix(stepStr, "zip") {
				for res := range parser.ParseZip()(stepStr) {
					cmd := res.Fst
					secAdapter := evaluator.EitherT(interactive.GenericExecAdapter(cmd, 0))
					pipeline = append(pipeline, evaluator.StepZip(secAdapter))
					break
				}
			} else if strings.HasPrefix(stepStr, "(") {
				for res := range parser.ParseFanOut()(stepStr) {
					leftAdapter := evaluator.EitherT(interactive.GenericExecAdapter(res.Fst.Fst, 0))
					rightAdapter := evaluator.EitherT(interactive.GenericExecAdapter(res.Fst.Snd, 0))
					pipeline = append(pipeline, evaluator.StepFanOut(leftAdapter, rightAdapter))
					break
				}
			} else if stepStr == "distinct" {
				pipeline = append(pipeline, evaluator.StepDistinct())
			} else if strings.HasPrefix(stepStr, "format_tuple") {
				formatStr := strings.TrimSpace(stepStr[12:])
				pipeline = append(pipeline, evaluator.StepFormatTuple(strings.Trim(formatStr, "\"")))
			} else if strings.HasPrefix(stepStr, "fold") {
				op := strings.TrimSpace(stepStr[4:])
				pipeline = append(pipeline, evaluator.StepFold(op))
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
			} else if strings.HasPrefix(stepStr, "memoize") {
				cmd := strings.TrimSpace(stepStr[7:])
				adapter := evaluator.EitherT(interactive.GenericExecAdapter(cmd, 0))
				pipeline = append(pipeline, evaluator.StepMemoize(adapter))
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
			} else {
				adapter := evaluator.EitherT(interactive.GenericExecAdapter(stepStr, 0))
				pipeline = append(pipeline, evaluator.StepAdapter(adapter))
			}
		}

		terminalNodes := engine.EvaluatePipeline(initialStream, pipeline)
		
		if isParallel {
			terminalNodes = engine.ParEvaluate(terminalNodes, 10)
		}

		count := 0
		for node := range terminalNodes {
			count++
			if node.Value.IsLeft() {
				err, _ := node.Value.UnwrapLeft()
				fmt.Printf("∷ [ERROR] %v\n", err)
			} else {
				val, _ := node.Value.UnwrapRight()
				fmt.Printf("∷ %+v\n", val)
			}
		}
		if count == 0 && !strings.HasPrefix(line, "type ") {
			fmt.Println("(No output or node unhandled)")
		}
	}

	loadScript := func(filename string) {
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Printf("Error reading file '%s': %v\n", filename, err)
			return
		}
		lines := strings.Split(string(bytes), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "--") {
				evaluateLine(line)
			}
		}
	}

	if scriptToLoad != "" {
		if !interactiveMode { fmt.Printf("Running Batch Mode: %s\n", scriptToLoad) }
		loadScript(scriptToLoad)
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
			loadScript(strings.TrimSpace(line[6:]))
			continue
		}
		evaluateLine(line)
	}
}
