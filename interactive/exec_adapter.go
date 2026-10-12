package interactive

import (
	"bufio"
	"fmt"
	"iter"
	"os/exec"
	"strings"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
)

func GenericExecAdapter(commandTemplate string, yieldedFlag dsltypes.BitFlag) evaluator.Adapter[any] {
	return func(node dsltypes.ASTNode[any]) iter.Seq[dsltypes.ASTNode[any]] {
		return func(yield func(dsltypes.ASTNode[any]) bool) {
			strValue := "unknown"
			if node.Value != nil {
				if s, ok := node.Value.(string); ok {
					strValue = s
				} else {
					// Non-string payloads (e.g. a sum-type Variant) render via their
					// Stringer / %v form rather than falling back to "unknown".
					strValue = fmt.Sprintf("%v", node.Value)
				}
			}
			cmdStr := commandTemplate
			if strings.Contains(commandTemplate, "%s") {
				cmdStr = fmt.Sprintf(commandTemplate, strValue)
			}
			
			parts := strings.Fields(cmdStr)
			if len(parts) == 0 {
				return
			}

			// Loud failure: a step that is neither a recognized builtin nor an
			// available executable surfaces an error instead of silently yielding
			// nothing. This turns a typo'd pipeline step (e.g. `frobnicate`) from a
			// silent no-op into a visible `[ERROR] unknown step or command` node.
			if _, lookErr := exec.LookPath(parts[0]); lookErr != nil {
				yield(dsltypes.NewASTNode[any](fmt.Errorf("unknown step or command: %q", parts[0]), yieldedFlag))
				return
			}

			cmd := exec.Command(parts[0], parts[1:]...)
			stdout, err := cmd.StdoutPipe()
			if err != nil { return }

			if err := cmd.Start(); err != nil { return }

			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				text := strings.TrimSpace(scanner.Text())
				if text != "" {
					childNode := dsltypes.NewASTNode[any](text, yieldedFlag)
					if !yield(childNode) {
						cmd.Process.Kill()
						return
					}
				}
			}
			cmd.Wait()
		}
	}
}

// ParseColumnsAdapter takes a string node (like a line from 'ls -l') 
// and structurally transforms it into a Record (map[string]any).
// It splits the string by whitespace and assigns them to the provided column names.
func ParseColumnsAdapter(columns []string, flag dsltypes.BitFlag) evaluator.Adapter[any] {
	return func(node dsltypes.ASTNode[any]) iter.Seq[dsltypes.ASTNode[any]] {
		return func(yield func(dsltypes.ASTNode[any]) bool) {
			strVal, ok := node.Value.(string)
			if !ok {
				// Bypass if it's not a string
				yield(node)
				return
			}
			
			parts := strings.Fields(strVal)
			record := make(map[string]any)
			
			for i, colName := range columns {
				if i < len(parts) {
					record[colName] = parts[i]
				} else {
					record[colName] = ""
				}
			}
			
			// If there are leftover parts, stuff them in the last column
			if len(parts) > len(columns) && len(columns) > 0 {
				record[columns[len(columns)-1]] = strings.Join(parts[len(columns)-1:], " ")
			}
			
			yield(dsltypes.NewASTNode[any](record, flag))
		}
	}
}
