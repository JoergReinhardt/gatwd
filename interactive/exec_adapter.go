package interactive

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"iter"
	"os/exec"
	"strings"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
)

// UnknownStepError is the failure of a step that is neither a recognised step nor an
// executable on the PATH: usually a typo. Name is the word that was looked up. Its message
// is `unknown step or command: "name"`.
type UnknownStepError struct{ Name string }

// Error is the one-line message.
func (e *UnknownStepError) Error() string {
	return fmt.Sprintf("unknown step or command: %q", e.Name)
}

// GenericExecAdapter runs commandTemplate (a `%s` is replaced by the incoming node) as an
// operating-system command and yields each non-empty line of its standard output as a node.
// A command that cannot start or exits non-zero yields an error node carrying its standard
// error. It runs without a context; see GenericExecAdapterContext.
func GenericExecAdapter(commandTemplate string, yieldedFlag dsltypes.BitFlag) evaluator.Adapter[any] {
	return GenericExecAdapterContext(nil, commandTemplate, yieldedFlag)
}

// GenericExecAdapterContext is GenericExecAdapter bound to a context supplied by ctx each
// time the adapter runs (a function, so an adapter built once and cached still sees the
// context of the run it executes in; nil means no context). When the context is done the
// child process is killed and the error node carries the context's cause
// (errors.Is(err, context.DeadlineExceeded) holds).
func GenericExecAdapterContext(ctx func() context.Context, commandTemplate string, yieldedFlag dsltypes.BitFlag) evaluator.Adapter[any] {
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
				yield(dsltypes.NewASTNode[any](&UnknownStepError{Name: parts[0]}, yieldedFlag))
				return
			}

			runCtx := context.Background()
			if ctx != nil {
				if c := ctx(); c != nil {
					runCtx = c
				}
			}
			fail := func(err error) {
				if runCtx.Err() != nil {
					// the process was killed by the context: say why, not "signal: killed"
					err = context.Cause(runCtx)
				}
				yield(dsltypes.NewASTNode[any](fmt.Errorf("%s: %w", parts[0], err), yieldedFlag))
			}

			cmd := exec.CommandContext(runCtx, parts[0], parts[1:]...)
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
