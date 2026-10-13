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

// Example_graphInOfOutOf shows the colour-filtered adjacency function forms. A rule
// asks "does anything of this colour point at x?" — an empty list means no — so a
// comprehension can find the roots of a relation (here, the nodes no Owns edge enters).
func Example_graphInOfOutOf() {
	interp := cli.NewInterpreter()
	interp.EvaluateScript(`def g_chain = fold (\g n -> graph_add_node "N" n g) graph_empty |> fold (\_ g -> graph_add_edges [["Owns", 1, "a", "b"], ["Owns", 1, "b", "c"]] g) graph_empty`)
	in := interp.EvaluateToValues(`["a", "b", "c"] |> g_chain |> fold (\_ g -> graph_in_of "Owns" "b" g) 0`)
	out := interp.EvaluateToValues(`["a", "b", "c"] |> g_chain |> fold (\_ g -> graph_out_of "Owns" "b" g) 0`)
	roots := interp.EvaluateToValues(`["a", "b", "c"] |> g_chain |> fold (\_ g -> [ n | n <- ["a", "b", "c"], (graph_in_of "Owns" n g) == [] ]) 0`)
	fmt.Println(in[0], out[0], roots[0])
	// Output: [a] [c] [a]
}

// Example_graphAlgorithmsInComprehension runs whole-graph algorithms (strongly
// connected components, topological order, shortest path) from inside a rule: here a
// dependency cycle is found with graph_scc, and a safe apply order is empty exactly
// when a cycle exists.
func Example_graphAlgorithmsInComprehension() {
	interp := cli.NewInterpreter()
	interp.EvaluateScript(`def g_cycle = fold (\g n -> graph_add_node "N" n g) graph_empty |> fold (\_ g -> graph_add_edges [["Owns", 1, "x", "y"], ["Owns", 1, "y", "x"], ["Owns", 1, "y", "z"]] g) graph_empty`)
	cycles := interp.EvaluateToValues(`["x", "y", "z"] |> g_cycle |> fold (\_ g -> [ c | c <- (graph_scc g), (length c) > 1 ]) 0`)
	order := interp.EvaluateToValues(`["x", "y", "z"] |> g_cycle |> fold (\_ g -> graph_topo_sort g) 0`)
	path := interp.EvaluateToValues(`["x", "y", "z"] |> g_cycle |> fold (\_ g -> graph_shortest_path "x" "z" g) 0`)
	fmt.Println(cycles[0], order[0], path[0])
	// Output: [[x y]] [] [x y z]
}

// Example_charFuncs shows the Char type: a literal, the function forms inside a lambda, and
// the on-demand conversion of a one-character String to a Char.
func Example_charFuncs() {
	interp := cli.NewInterpreter()
	up := interp.EvaluateToValues(`['a', 'b'] |> map (\c -> char_upper c)`)
	code := interp.EvaluateToValues(`char_code "é"`)
	fmt.Println(up[0], code[0])
	// Output: [A B] 233
}

// Example_stringRow shows String as a collection of Chars: filter and reverse answer with a
// String, map with a list of whatever the function returns.
func Example_stringRow() {
	interp := cli.NewInterpreter()
	kept := interp.EvaluateToValues(`filter (\c -> c != 'b') "abcb"`)
	back := interp.EvaluateToValues(`reverse "abc"`)
	codes := interp.EvaluateToValues(`map (\c -> char_code c) "ab"`)
	fmt.Println(kept[0], back[0], codes[0])
	// Output: ac cba [97 98]
}
