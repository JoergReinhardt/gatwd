package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"io/ioutil"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/interactive"
	"github.com/joergreinhardt/gatwd/parser"
)

func main() {
	fmt.Println("=== gatwd Interpreter (V2 Modernized) ===")
	fmt.Println("UTF-8 Mathematical Aliases Enabled (e.g. ⊳, →, ∷)")

	p := parser.NewParser()
	engine := evaluator.NewEngine[string]()

	engine.OnMissingAdapter = func(flag dsltypes.BitFlag) evaluator.Adapter[string] {
		fmt.Printf("\n[REPL] Encountered unknown node type %d. Enter CLI command (or enter to skip): ", flag)
		reader := bufio.NewReader(os.Stdin)
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			return nil
		}
		return interactive.GenericExecAdapter(cmd, 0)
	}

	// Simple flag parsing
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

	// Function to load and execute a script
	loadScript := func(filename string) {
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Printf("Error reading file '%s': %v\n", filename, err)
			return
		}
		astStream := p.Parse(string(bytes))
		terminalNodes := engine.Evaluate(astStream)
		count := 0
		for node := range terminalNodes {
			fmt.Printf("-> %s\n", node.Value)
			count++
		}
		if count > 0 {
			fmt.Printf("Loaded and executed %d terminal nodes from %s\n", count, filename)
		} else {
			fmt.Printf("Loaded definitions from %s\n", filename)
		}
	}

	if scriptToLoad != "" {
		if !interactiveMode {
			fmt.Printf("Running Batch Mode: %s\n", scriptToLoad)
		} else {
			fmt.Printf("Pre-loading script into REPL: %s\n", scriptToLoad)
		}
		loadScript(scriptToLoad)
		if !interactiveMode {
			return
		}
	}

	fmt.Println("Running Interactive Mode. Type ':load <file>' to load a script, or 'exit' to quit.")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("gatwd ⊳ ") // Using mathematical right-triangle for the prompt!
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

		if strings.HasPrefix(line, ":load ") {
			filename := strings.TrimSpace(line[6:])
			loadScript(filename)
			continue
		}

		astStream := p.Parse(line)
		terminalNodes := engine.Evaluate(astStream)
		
		count := 0
		for node := range terminalNodes {
			count++
			fmt.Printf("∷ %s (Type: %d)\n", node.Value, node.DSLType)
		}
		
		if count == 0 && !strings.HasPrefix(line, "type ") {
			fmt.Println("(No output or node unhandled)")
		}
	}
}
