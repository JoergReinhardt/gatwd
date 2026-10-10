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

func GenericExecAdapter(commandTemplate string, yieldedFlag dsltypes.BitFlag) evaluator.Adapter[string] {
	return func(node dsltypes.ASTNode[string]) iter.Seq[dsltypes.ASTNode[string]] {
		return func(yield func(dsltypes.ASTNode[string]) bool) {
			cmdStr := commandTemplate
			if strings.Contains(commandTemplate, "%s") {
				cmdStr = fmt.Sprintf(commandTemplate, node.Value)
			}
			
			parts := strings.Fields(cmdStr)
			if len(parts) == 0 {
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
					childNode := dsltypes.NewASTNode(text, yieldedFlag)
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
