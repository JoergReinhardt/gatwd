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
	
	// Upgrade Engine to use EitherT for compliance auditing!
	engine := evaluator.NewEngine[functions.Either[error, string]]()

	engine.OnMissingAdapter = func(flag dsltypes.BitFlag) evaluator.Adapter[functions.Either[error, string]] {
		fmt.Printf("\n[REPL] Encountered unknown node type %d. Enter CLI command (or enter to skip): ", flag)
		reader := bufio.NewReader(os.Stdin)
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			return nil
		}
		
		// Wrap the generic string adapter in EitherT so it handles Errors safely
		genericAdapter := interactive.GenericExecAdapter(cmd, 0)
		
		return evaluator.EitherT(genericAdapter)
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
		parts := strings.Split(line, "|>")
		
		// Parse the Root Node (e.g., spawn Server)
		// p.Parse returns iter.Seq[ASTNode[string]], we need to map it to Either
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
					
					predicate := func(val string) bool {
						return strings.Contains(val, keyword)
					}
					pipeline = append(pipeline, evaluator.StepExpect(predicate, fmt.Errorf("%s", errMsg)))
					break
				}
			} else if strings.HasPrefix(stepStr, "(") {
				// FanOut &&&
				for res := range parser.ParseFanOut()(stepStr) {
					// We'd look up the adapters dynamically, but for CLI let's use REPL
					leftCmd := res.Fst.Fst
					rightCmd := res.Fst.Snd
					
					leftAdapter := evaluator.EitherT(interactive.GenericExecAdapter(leftCmd, 0))
					rightAdapter := evaluator.EitherT(interactive.GenericExecAdapter(rightCmd, 0))
					
					pipeline = append(pipeline, evaluator.StepFanOut(leftAdapter, rightAdapter))
					break
				}
			} else {
				// Normal Adapter via REPL fallback for demo
				adapter := evaluator.EitherT(interactive.GenericExecAdapter(stepStr, 0))
				pipeline = append(pipeline, evaluator.StepAdapter(adapter))
			}
		}

		terminalNodes := engine.EvaluatePipeline(initialStream, pipeline)
		count := 0
		for node := range terminalNodes {
			count++
			if node.Value.IsLeft() {
				err, _ := node.Value.UnwrapLeft()
				fmt.Printf("∷ [ERROR] %v\n", err)
			} else {
				val, _ := node.Value.UnwrapRight()
				fmt.Printf("∷ %s (Type: %d)\n", val, node.DSLType)
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
		if !interactiveMode {
			fmt.Printf("Running Batch Mode: %s\n", scriptToLoad)
		}
		loadScript(scriptToLoad)
		if !interactiveMode {
			return
		}
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
