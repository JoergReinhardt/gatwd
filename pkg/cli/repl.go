package cli

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

// REPL reads statements from in until `exit`, `quit` or end of input and evaluates each one
// as a top-level statement, printing to i.Out. Besides DSL statements it understands
// `:load FILE`, `:macros` and `:vars`.
func (i *Interpreter) REPL(in io.Reader) {
	out := i.out()
	fmt.Fprintln(out, "Running Interactive Mode. Type ':load <file>' to load a script, or 'exit' to quit.")
	scanner := bufio.NewScanner(in)
	for {
		fmt.Fprint(out, "gatwd ⊳ ")
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
		switch {
		case strings.HasPrefix(line, ":load "):
			i.LoadScript(strings.TrimSpace(line[6:]))
		case line == ":macros":
			fmt.Fprintln(out, "Registered Macros:")
			for _, k := range sortedKeys(i.Funcs) {
				fmt.Fprintf(out, "  %s = %s\n", k, i.Funcs[k])
			}
		case line == ":vars":
			fmt.Fprintln(out, "Registered Bindings (let):")
			for _, k := range sortedKeys(i.Variables) {
				fmt.Fprintf(out, "  %s = %s\n", k, i.Variables[k])
			}
		default:
			i.EvaluateLine(line)
		}
	}
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
