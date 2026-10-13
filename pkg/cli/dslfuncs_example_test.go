package cli_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

// Example_graphIngestFold demonstrates the pure-DSL graph ingest primitive: a fold
// whose accumulator is a graph and whose reducer applies the graph_add_node function
// form (callable now that the infix expression language supports juxtaposition
// application). A list of records is folded into one persistent graph — the generic,
// domain-free foundation the Kubernetes relation prelude builds on.
func Example_graphIngestFold() {
	interp := cli.NewInterpreter()
	// Fold a list of ids into graph nodes, then read back the node count.
	res := interp.EvaluateToValues(`let ids = ["a", "b", "c"] in ids |> fold (\g r -> graph_add_node "N" r g) graph_empty |> graph_order`)
	fmt.Println(res[0])
	// Output: 3
}

// Example_graphIngestFromRecords shows color and identity DERIVED from a record
// payload via the hamt_get function form during the ingest fold, with the record
// kept as the node payload — the shape a crawled resource stream takes.
func Example_graphIngestFromRecords() {
	interp := cli.NewInterpreter()
	res := interp.EvaluateToValues(`let r1 = spawn S: x |> emit kind:Pod, name:nginx |> drop 1 |> to_hamt in let r2 = spawn S: x |> emit kind:Service, name:web |> drop 1 |> to_hamt in [r1, r2] |> fold (\g x -> graph_add_node (hamt_get "kind" x) (hamt_get "name" x) x g) graph_empty |> graph_nodes_of Service |> head`)
	fmt.Println(res[0])
	// Output: web
}
