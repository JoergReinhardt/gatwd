package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/interactive"
	"github.com/joergreinhardt/gatwd/parser"
)

func main() {
	fmt.Println("=== gatwd Interpretter (V2 Modernized) ===")

	p := parser.NewParser()
	engine := evaluator.NewEngine[string]()

	// Setup Interactive REPL Hook
	engine.OnMissingAdapter = func(flag dsltypes.BitFlag) evaluator.Adapter[string] {
		fmt.Printf("\n[REPL] Encountered unknown node type %d. Enter CLI command to evaluate (or press enter to skip): ", flag)
		reader := bufio.NewReader(os.Stdin)
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			return nil
		}
		return interactive.GenericExecAdapter(cmd, 0) // Yields generic output nodes
	}

	// Batch Mode
	if len(os.Args) > 1 {
		filename := os.Args[1]
		fmt.Printf("Running Batch Mode: %s\n", filename)
		bytes, err := os.ReadFile(filename)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
		
		astStream := p.Parse(string(bytes)) // Assuming p.Parse is the Phase 4 parser
		
		terminalNodes := engine.Evaluate(astStream)
		for node := range terminalNodes {
			fmt.Printf("-> %s\n", node.Value)
		}
		return
	}

	// Interactive Mode
	fmt.Println("Running Interactive Mode. Type 'exit' to quit.")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("gatwd> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "exit" || line == "quit" {
			break
		}
		if line == "" {
			continue
		}

		astStream := p.Parse(line)
		terminalNodes := engine.Evaluate(astStream)
		
		count := 0
		for node := range terminalNodes {
			count++
			fmt.Printf("-> %s (Type: %d)\n", node.Value, node.DSLType)
		}
		
		if count == 0 && !strings.HasPrefix(line, "type ") {
			fmt.Println("(No output or node unhandled)")
		}
	}
}
