package interactive

import (
	"bufio"
	"bytes"
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

			parts, splitErr := SplitCommandLine(cmdStr)
			if splitErr != nil {
				yield(dsltypes.NewASTNode[any](fmt.Errorf("%s: %w", cmdStr, splitErr), yieldedFlag))
				return
			}
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

			fail := func(err error) {
				yield(dsltypes.NewASTNode[any](fmt.Errorf("%s: %w", parts[0], err), yieldedFlag))
			}

			cmd := exec.Command(parts[0], parts[1:]...)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				fail(err)
				return
			}

			if err := cmd.Start(); err != nil {
				fail(err)
				return
			}

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
			if err := cmd.Wait(); err != nil {
				if msg := strings.TrimSpace(stderr.String()); msg != "" {
					fail(fmt.Errorf("%w: %s", err, msg))
				} else {
					fail(err)
				}
			}
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

// SplitCommandLine splits a command line into arguments the way a shell tokenizes it,
// without any expansion: whitespace separates arguments, single quotes take their
// content literally, double quotes group (with \" and \\ as the only escapes), and an
// unterminated quote is an error rather than a silently mangled argument.
func SplitCommandLine(s string) ([]string, error) {
	var args []string
	var cur strings.Builder
	inArg := false
	var quote rune
	runes := []rune(s)
	for idx := 0; idx < len(runes); idx++ {
		r := runes[idx]
		switch {
		case quote == '\'':
			if r == '\'' {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case quote == '"':
			switch {
			case r == '"':
				quote = 0
			case r == '\\' && idx+1 < len(runes) && (runes[idx+1] == '"' || runes[idx+1] == '\\'):
				idx++
				cur.WriteRune(runes[idx])
			default:
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
			inArg = true
		case r == ' ' || r == '\t' || r == '\n':
			if inArg {
				args = append(args, cur.String())
				cur.Reset()
				inArg = false
			}
		default:
			cur.WriteRune(r)
			inArg = true
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated %c quote in command", quote)
	}
	if inArg {
		args = append(args, cur.String())
	}
	return args, nil
}
