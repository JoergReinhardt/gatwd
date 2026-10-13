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

// Example_strLibrary shows the str_* library: data-last functions that double as pipeline
// steps, with Maybe-valued parsing instead of errors for malformed text.
func Example_strLibrary() {
	interp := cli.NewInterpreter()
	parts := interp.EvaluateToValues(`"a,b,c" |> str_split ","`)
	n := interp.EvaluateToValues(`maybe_default 0 (str_to_int "4x")`)
	msg := interp.EvaluateToValues(`str_format "%s=%d" ["x", 3]`)
	fmt.Println(parts[0], n[0], msg[0])
	// Output: [a b c] 0 x=3
}

// Example_stringEscapes shows escapes in String literals: a newline splits lines, and a
// quote is written \".
func Example_stringEscapes() {
	interp := cli.NewInterpreter()
	lines := interp.EvaluateToValues(`str_lines "a\nb\n"`)
	quoted := interp.EvaluateToValues(`"say \"hi\""`)
	fmt.Println(lines[0], quoted[0])
	// Output: [a b] say "hi"
}

// Example_ordering shows compare answering with the Ordering sum type, and the strict Bool
// functions used as a fold.
func Example_ordering() {
	interp := cli.NewInterpreter()
	o := interp.EvaluateToValues(`compare 1 2`)
	all := interp.EvaluateToValues(`[true, true, false] |> fold and true`)
	fmt.Println(o[0], all[0])
	// Output: LT false
}

// Example_plusIsNumeric shows D2: `+` adds numbers, `++` concatenates, and text is
// converted explicitly.
func Example_plusIsNumeric() {
	interp := cli.NewInterpreter()
	sum := interp.EvaluateToValues(`1 + 2`)
	text := interp.EvaluateToValues(`"1" ++ "2"`)
	parsed := interp.EvaluateToValues(`maybe_default 0 (str_to_int "41") + 1`)
	bad := interp.EvaluateToValues(`"a" + 1`)
	fmt.Println(sum[0], text[0], parsed[0])
	fmt.Println(bad[0])
	// Output:
	// 3 12 42
	// operator '+' is not defined for String and Int; use ++ to concatenate
}

// Example_encodings shows Bytes: encoders take a String or Bytes, decoders are Maybe, and a
// hash is a hex String.
func Example_encodings() {
	interp := cli.NewInterpreter()
	enc := interp.EvaluateToValues(`base64_encode "hello"`)
	back := interp.EvaluateToValues(`maybe_default "?" (utf8_decode (maybe_default (utf8_encode "") (base64_decode "aGVsbG8=")))`)
	bad := interp.EvaluateToValues(`is_just (base64_decode "!!!")`)
	sum := interp.EvaluateToValues(`sha256 "abc"`)
	fmt.Println(enc[0], back[0], bad[0])
	fmt.Println(sum[0])
	// Output:
	// aGVsbG8= hello false
	// ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
}

// Example_json shows JSON: parse into the resource shape (a Maybe), navigate with hamt_path,
// and encode back deterministically.
func Example_json() {
	interp := cli.NewInterpreter()
	name := interp.EvaluateToValues(`maybe_default "?" (map (\d -> hamt_path "metadata.name" d) (json_parse "{\"metadata\": {\"name\": \"web\"}}"))`)
	bad := interp.EvaluateToValues(`is_just (json_parse "[1,]")`)
	out := interp.EvaluateToValues(`json_encode (hamt_from [["b", 1], ["a", [1, 2]]])`)
	fmt.Println(name[0], bad[0])
	fmt.Println(out[0])
	// Output:
	// web false
	// {"a":[1,2],"b":1}
}

// Example_regex shows RE2 regular expressions: matching, extraction with groups, and
// replacement with $1 expansion. A pattern that does not match is Nothing, not an error.
func Example_regex() {
	interp := cli.NewInterpreter()
	ok := interp.EvaluateToValues(`re_match "^v[0-9]+$" "v12"`)
	g := interp.EvaluateToValues(`maybe_default [] (re_groups "(\\w+)@(\\w+)" "bob@example")`)
	r := interp.EvaluateToValues(`re_replace "(\\w+)@(\\w+)" "$2 at $1" "bob@example"`)
	none := interp.EvaluateToValues(`is_just (re_find "z" "abc")`)
	fmt.Println(ok[0], g[0], r[0], none[0])
	// Output: true [bob@example bob example] example at bob false
}
