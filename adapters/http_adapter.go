package adapters

import (
	"bufio"
	"iter"
	"net/http"
	"strings"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
)

// NativeHTTPAdapter is a reference implementation of a Native Go Adapter.
//
// By using the "Closure Pattern", we can inject persistent state (like an
// *http.Client with connection pooling and timeouts) into the adapter ONCE.
// The engine will reuse this optimized client across potentially millions
// of graph node evaluations.
func NativeHTTPAdapter(client *http.Client, targetFlag dsltypes.BitFlag) evaluator.Adapter[string] {
	
	// Return the pure monadic function expected by the Evaluator
	return func(node dsltypes.ASTNode[string]) iter.Seq[dsltypes.ASTNode[string]] {
		return func(yield func(dsltypes.ASTNode[string]) bool) {
			
			// node.Value is assumed to be a URL for this specific adapter
			resp, err := client.Get(node.Value)
			if err != nil {
				// In a pure Monad, we might yield an error node.
				// Here we silently halt this branch of the graph.
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return
			}

			// Lazily scan the response body line by line
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					childNode := dsltypes.NewASTNode(line, targetFlag)
					
					// Yield into the Trampoline
					if !yield(childNode) {
						// Safe short-circuit! If the evaluator stops reading
						// (e.g. ParMap cancels context), we immediately close 
						// the response body via defer and exit.
						return
					}
				}
			}
		}
	}
}
