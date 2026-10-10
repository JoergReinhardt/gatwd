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
	engine := evaluator.NewEngine[functions.Either[error, string]]()

	engine.OnMissingAdapter = func(flag dsltypes.BitFlag) evaluator.Adapter[functions.Either[error, string]] {
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
		initialStream := func(yield func(dsltypes.ASTNode[functions.Either[error, string]]) bool) {
			for node := range stringStream {
				wrappedValue := functions.Right[error, string](node.Value)
				yield(dsltypes.NewASTNode(wrappedValue, node.DSLType))
			}
		}

		var pipeline []evaluator.PipelineStep[functions.Either[error, string]]

		for _, stepStr := range parts[1:] {
			stepStr = strings.TrimSpace(stepStr)
			
			if strings.HasPrefix(stepStr, "take") {
				for res := range parser.ParseTake()(stepStr) {
					pipeline = append(pipeline, evaluator.StepTake[functions.Either[error, string]](res.Fst))
					break
				}
			} else if strings.HasPrefix(stepStr, "drop") {
				for res := range parser.ParseDrop()(stepStr) {
					pipeline = append(pipeline, evaluator.StepDrop[functions.Either[error, string]](res.Fst))
					break
				}
			} else if strings.HasPrefix(stepStr, "expect") {
				for res := range parser.ParseExpect()(stepStr) {
					keyword := res.Fst.Fst
					errMsg := res.Fst.Snd
					predicate := func(val string) bool { return strings.Contains(val, keyword) }
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
			} else {
				adapter := evaluator.EitherT(interactive.GenericExecAdapter(stepStr, 0))
				pipeline = append(pipeline, evaluator.StepAdapter(adapter))
			}
		}

		terminalNodes := engine.EvaluatePipeline(initialStream, pipeline)
		
		if isParallel {
			// Convert Pipeline sequence to ParEvaluate pool (10 workers)
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
				fmt.Printf("∷ %s\n", val)
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
