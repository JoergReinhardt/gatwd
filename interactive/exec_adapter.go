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

// GenericExecAdapter dynamically generates an Adapter that runs a CLI command.
//
// This is the embodiment of the IO Monad for Tool Agnosticity (Phase 5).
// It wraps arbitrary side-effects (`os/exec`) into a pure `iter.Seq` stream.
// 
// commandTemplate: e.g. "echo discovering %s" where %s is replaced by the ASTNode's value.
// yieldedFlag: The BitFlag assigned to the newly discovered nodes (lines of stdout).
func GenericExecAdapter(commandTemplate string, yieldedFlag dsltypes.BitFlag) evaluator.Adapter[string] {
	return func(node dsltypes.ASTNode[string]) iter.Seq[dsltypes.ASTNode[string]] {
		return func(yield func(dsltypes.ASTNode[string]) bool) {
			// Construct command
			cmdStr := fmt.Sprintf(commandTemplate, node.Value)
			parts := strings.Fields(cmdStr)
			if len(parts) == 0 {
				return
			}

			cmd := exec.Command(parts[0], parts[1:]...)
			
			// Pipe stdout
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				// In a pure Monad, this would yield an Error ASTNode. 
				// For simplicity, we silently stop the sequence.
				return
			}

			if err := cmd.Start(); err != nil {
				return
			}

			// Read line by line lazily, yielding immediately to the trampoline
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				text := strings.TrimSpace(scanner.Text())
				if text != "" {
					childNode := dsltypes.NewASTNode(text, yieldedFlag)
					if !yield(childNode) {
						// Consumer requested stop; kill the process and exit
						cmd.Process.Kill()
						return
					}
				}
			}

			cmd.Wait()
		}
	}
}
