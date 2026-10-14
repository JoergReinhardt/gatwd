package spec

// KeywordDef describes a core language keyword.
type KeywordDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// TypeDef describes a native or algebraic data type constructor.
type TypeDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// BuiltinDef describes a built-in transducer, optic, adapter, or pipeline step.
type BuiltinDef struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Signature   string `json:"signature"`
	Description string `json:"description"`
}

// OperatorDef describes an operator, pipe, or arrow.
type OperatorDef struct {
	Symbol        string `json:"symbol"`
	Unicode       string `json:"unicode,omitempty"`
	Category      string `json:"category"`
	Precedence    int    `json:"precedence,omitempty"`
	Associativity string `json:"associativity,omitempty"` // "left" or "right"
	Description   string `json:"description"`
}

// SnippetDef describes an editor snippet template.
type SnippetDef struct {
	Prefix      string `json:"prefix"`
	Body        string `json:"body"`
	Description string `json:"description"`
}

// GrammarSpec is the single source of truth for the Gatwd DSL syntax, tokens, and metadata.
type GrammarSpec struct {
	LanguageName  string        `json:"language_name"`
	FileExtension string        `json:"file_extension"`
	CommentPrefix string        `json:"comment_prefix"`
	Keywords      []KeywordDef  `json:"keywords"`
	Types         []TypeDef     `json:"types"`
	Builtins      []BuiltinDef  `json:"builtins"`
	Operators     []OperatorDef `json:"operators"`
	Snippets      []SnippetDef  `json:"snippets"`
}

// DefaultSpec returns the canonical, comprehensive Gatwd Grammar Specification.
func DefaultSpec() GrammarSpec {
	return GrammarSpec{
		LanguageName:  "Gatwd",
		FileExtension: "gatwd",
		CommentPrefix: "--",
		Keywords: []KeywordDef{
			{Name: "def", Description: "Define a reusable function pipeline"},
			{Name: "spawn", Description: "Spawn an initial typed stream AST node"},
			{Name: "type", Description: "Declare an algebraic data type, record, or collection alias"},
			{Name: "const", Description: "Declare a compile-time or script constant"},
			{Name: "let", Description: "Introduce a locally scoped variable binding"},
			{Name: "in", Description: "Delimiter for the evaluation body of a let expression"},
			{Name: "match", Description: "Pattern matching catamorphism with ADT and fixed-list destructuring"},
			{Name: "seq", Description: "Applicative stream concatenation of pipeline branches"},
			{Name: "import", Description: "Import declarations and functions from another Gatwd module"},
			{Name: "empty", Description: "Terminate a pipeline stream with an empty output"},
		},
		Types: []TypeDef{
			{Name: "Int", Description: "Arbitrary or signed integer"},
			{Name: "String", Description: "UTF-8 string value"},
			{Name: "Float", Description: "Floating point real number"},
			{Name: "BigInt", Description: "Arbitrary-precision integer (*big.Int)"},
			{Name: "BigFloat", Description: "Arbitrary-precision floating point number (*big.Float)"},
			{Name: "Complex", Description: "Complex number (complex128)"},
			{Name: "Uint", Description: "Unsigned 64-bit integer"},
			{Name: "Bool", Description: "Boolean value (true / false)"},
			{Name: "List", Description: "Immutable persistent cons-cell linked list"},
			{Name: "Set", Description: "Persistent HAMT-backed immutable set"},
			{Name: "Tree", Description: "Immutable Rose Tree multi-way branch structure"},
			{Name: "BST", Description: "Binary search tree"},
			{Name: "BTree", Description: "Balanced B-tree"},
			{Name: "RBTree", Description: "Self-balancing red-black tree"},
			{Name: "Pair", Description: "Heterogeneous 2-element tuple product type"},
			{Name: "Tuple", Description: "N-ary product type"},
			{Name: "Grid", Description: "Immutable 2D spatial plane or comonadic torus grid"},
			{Name: "Char", Description: "A single Unicode code point ('a', '\\n')"},
			{Name: "Bytes", Description: "Immutable byte string"},
			{Name: "Ratio", Description: "Exact rational number (3r4)"},
			{Name: "Duration", Description: "Fixed-length time span (5s, 1h30m, 2d)"},
			{Name: "Time", Description: "RFC 3339 instant with zone (@2026-10-13T12:00:00Z)"},
			{Name: "Date", Description: "ISO 8601 calendar day (@2026-10-13)"},
			{Name: "Chunk", Description: "Contiguous, cache-friendly immutable batch of values"},
			{Name: "HAMT", Description: "Persistent hash array mapped trie (immutable dictionary)"},
			{Name: "Graph", Description: "Typed, coloured, weighted persistent directed graph"},
			{Name: "Maybe", Description: "Optional value: just x or nothing"},
			{Name: "Either", Description: "Value that is either a Left (error) or a Right (success)"},
			{Name: "Variant", Description: "Boxed constructor of a declared sum type with its payload"},
		},
		Builtins: []BuiltinDef{
			// Transducers & Core
			{Name: "take", Category: "transducer", Signature: "take <n>", Description: "Take the first n nodes from stream"},
			{Name: "drop", Category: "transducer", Signature: "drop <n>", Description: "Drop the first n nodes from stream"},
			{Name: "head", Category: "transducer", Signature: "head", Description: "Take exactly 1 node from stream"},
			{Name: "tail", Category: "transducer", Signature: "tail", Description: "Drop the first 1 node from stream"},
			{Name: "distinct", Category: "transducer", Signature: "distinct", Description: "Filter out duplicate elements in stream"},
			{Name: "fold", Category: "transducer", Signature: "fold (count|sum)", Description: "Aggregate stream into a single scalar value"},
			{Name: "memoize", Category: "transducer", Signature: "memoize <step>", Description: "Cache execution results of pure steps"},
			{Name: "map", Category: "transducer", Signature: "map (\\x -> ...)", Description: "Monadic bind / flatMap over stream elements"},
			{Name: "filter", Category: "transducer", Signature: "filter <substring>", Description: "Keep nodes matching text filter"},
			{Name: "filter_less", Category: "transducer", Signature: "filter_less <var>", Description: "Filter elements strictly less than pivot"},
			{Name: "filter_greater", Category: "transducer", Signature: "filter_greater <var>", Description: "Filter elements greater than or equal to pivot"},
			{Name: "sort", Category: "transducer", Signature: "sort", Description: "Strict sorting boundary over stream"},
			{Name: "emit", Category: "transducer", Signature: "emit <items...>", Description: "Emit additional values into the stream"},
			{Name: "chunk", Category: "transducer", Signature: "chunk <size>", Description: "Batch stream into contiguous CPU-cache Chunks"},
			{Name: "unchunk", Category: "transducer", Signature: "unchunk", Description: "Flatten batched Chunks back into individual stream nodes"},
			{Name: "collect_list", Category: "transducer", Signature: "collect_list", Description: "Collect entire stream into an immutable List"},

			// Persistent Collections (HAMT & Sets)
			{Name: "to_hamt", Category: "collections", Signature: "to_hamt", Description: "Fold key:value stream into persistent HAMT dictionary"},
			{Name: "hamt_get", Category: "collections", Signature: "hamt_get <key>", Description: "Lookup key in persistent HAMT"},
			{Name: "hamt_set", Category: "collections", Signature: "hamt_set <key> <val>", Description: "Insert or update key-value in HAMT with structural sharing"},
			{Name: "hamt_del", Category: "collections", Signature: "hamt_del <key>", Description: "Remove key from HAMT"},
			{Name: "hamt_keys", Category: "collections", Signature: "hamt_keys", Description: "Stream all keys from HAMT"},
			{Name: "hamt_values", Category: "collections", Signature: "hamt_values", Description: "Stream all values from HAMT"},
			{Name: "hamt_pairs", Category: "collections", Signature: "hamt_pairs", Description: "Stream all key-value pairs from HAMT"},
			{Name: "frequencies", Category: "collections", Signature: "frequencies", Description: "Count occurrences of stream elements into persistent HAMT"},
			{Name: "to_set", Category: "collections", Signature: "to_set", Description: "Fold stream into persistent immutable set"},
			{Name: "set_add", Category: "collections", Signature: "set_add <item>", Description: "Add element to immutable set"},
			{Name: "set_contains", Category: "collections", Signature: "set_contains <item>", Description: "Test set membership"},
			{Name: "set_remove", Category: "collections", Signature: "set_remove <item>", Description: "Remove element from immutable set"},
			{Name: "set_elements", Category: "collections", Signature: "set_elements", Description: "Unfold set into stream"},

			// Optics (Lenses & Records)
			{Name: "parse_record", Category: "optics", Signature: "parse_record <fields...>", Description: "Parse whitespace columns into named Record fields"},
			{Name: "focus", Category: "optics", Signature: "focus <field>", Description: "Lens view: extract focused field from Record"},
			{Name: "update", Category: "optics", Signature: "update <field> <val>", Description: "Lens set: pure functional field update"},

			// Comonads & Zippers
			{Name: "build_tree", Category: "comonad", Signature: "build_tree <root>", Description: "Construct immutable Rose Tree"},
			{Name: "into_zipper", Category: "comonad", Signature: "into_zipper", Description: "Focus Rose Tree into Comonadic Zipper cursor"},
			{Name: "down", Category: "comonad", Signature: "down", Description: "Move Zipper cursor to first child"},
			{Name: "up", Category: "comonad", Signature: "up", Description: "Move Zipper cursor to parent node"},
			{Name: "left", Category: "comonad", Signature: "left", Description: "Move Zipper cursor to left sibling"},
			{Name: "right", Category: "comonad", Signature: "right", Description: "Move Zipper cursor to right sibling"},
			{Name: "extract", Category: "comonad", Signature: "extract", Description: "Comonadic counit: extract focused node value"},

			// Arrow Combinators & Tuples
			{Name: "zip", Category: "arrows", Signature: "zip <adapter>", Description: "Pairwise zip incoming stream with secondary stream"},
			{Name: "format_tuple", Category: "arrows", Signature: "format_tuple <fmt>", Description: "Format product pair into formatted string"},

			// Verification & Observability
			{Name: "assert", Category: "testing", Signature: "assert <expected>", Description: "Assert exact node value equality"},
			{Name: "expect", Category: "testing", Signature: "expect <pred> <errMsg>", Description: "Assert predicate holds or abort with error"},
			{Name: "debug_print", Category: "observability", Signature: "debug_print <prefix>", Description: "Print inspect nodes in stream"},
			{Name: "visualize", Category: "observability", Signature: "visualize [flags]", Description: "Render stream as pretty UTF-8 visual cards"},
			{Name: "telemetry", Category: "observability", Signature: "telemetry", Description: "Track and emit stream throughput telemetry"},
			{Name: "checkpoint", Category: "reliability", Signature: "checkpoint <id> <interval>", Description: "Periodically persist pipeline state for recovery"},
			{Name: "to_list", Category: "transducer", Signature: "to_list", Description: "Collect the stream into an immutable List"},
			{Name: "to_chunk", Category: "transducer", Signature: "to_chunk", Description: "Collect the stream into a Chunk"},
			{Name: "flatten_chunk", Category: "transducer", Signature: "flatten_chunk", Description: "Flatten Chunks back into individual stream nodes"},
			{Name: "reject", Category: "transducer", Signature: "reject <pred>", Description: "Drop nodes matching a predicate"},
			{Name: "flatmap", Category: "transducer", Signature: "flatmap (\\x -> ...)", Description: "Map then flatten (monadic bind)"},
			{Name: "bimap", Category: "transducer", Signature: "bimap <f> <g>", Description: "Map both sides of a Pair or Either"},
			{Name: "subset_of", Category: "collections", Signature: "subset_of <super>", Description: "Key-value containment (Record/HAMT) or Set membership test"},
			{Name: "hamt_path", Category: "collections", Signature: "hamt_path <k1.k2...>", Description: "Look up a nested path in a HAMT"},
			{Name: "hamt_merge", Category: "collections", Signature: "hamt_merge <other>", Description: "Merge two HAMTs"},
			{Name: "hamt_diff", Category: "collections", Signature: "hamt_diff <other>", Description: "Keys of this HAMT absent from the other"},
			{Name: "hamt_intersect", Category: "collections", Signature: "hamt_intersect <other>", Description: "Keys present in both HAMTs"},
			{Name: "set_union", Category: "collections", Signature: "set_union <other>", Description: "Set union"},
			{Name: "set_intersect", Category: "collections", Signature: "set_intersect <other>", Description: "Set intersection"},
			{Name: "set_diff", Category: "collections", Signature: "set_diff <other>", Description: "Set difference"},
			{Name: "build_bst", Category: "collections", Signature: "build_bst", Description: "Build a binary search tree from the stream"},
			{Name: "build_btree", Category: "collections", Signature: "build_btree", Description: "Build a B-tree from the stream"},
			{Name: "build_rbtree", Category: "collections", Signature: "build_rbtree", Description: "Build a red-black tree from the stream"},
			{Name: "build_array", Category: "collections", Signature: "build_array", Description: "Build an array from the stream"},
			{Name: "math_add", Category: "math", Signature: "math_add <n>", Description: "Add n to each value (also the step form `+ n`)"},
			{Name: "math_sub", Category: "math", Signature: "math_sub <n>", Description: "Subtract n (also `- n`)"},
			{Name: "math_mul", Category: "math", Signature: "math_mul <n>", Description: "Multiply by n (also `* n`)"},
			{Name: "math_div", Category: "math", Signature: "math_div <n>", Description: "True division by n (also `/ n`)"},
			{Name: "math_idiv", Category: "math", Signature: "math_idiv <n>", Description: "Integer floor division by n (also `// n`)"},
			{Name: "math_mod", Category: "math", Signature: "math_mod <n>", Description: "Remainder modulo n (also `% n`)"},
			{Name: "math_pow", Category: "math", Signature: "math_pow <n>", Description: "Raise to the power n (also `** n`)"},
			{Name: "now", Category: "time", Signature: "now", Description: "Emit the current Time (reads the injectable clock)"},
			{Name: "walk_fs", Category: "io", Signature: "walk_fs <dir>", Description: "Stream the files below a directory"},
			{Name: "write_file", Category: "io", Signature: "write_file <path>", Description: "Write the stream to a file"},
			{Name: "env_get", Category: "io", Signature: "env_get <NAME>", Description: "Look up an environment variable: Just value or Nothing (needs the env capability)"},
			{Name: "append_file", Category: "io", Signature: "append_file <path>", Description: "Append the stream to a file"},
			{Name: "scan_file", Category: "io", Signature: "scan_file <path>", Description: "Stream a file line by line"},
			{Name: "scan_words", Category: "io", Signature: "scan_words <path>", Description: "Stream a file word by word"},
			{Name: "dial_tcp", Category: "io", Signature: "dial_tcp <addr>", Description: "Connect to a TCP address"},
			{Name: "listen_tcp", Category: "io", Signature: "listen_tcp <addr>", Description: "Listen on a TCP address"},
			{Name: "conn_lines", Category: "io", Signature: "conn_lines", Description: "Stream the lines of a connection"},
			{Name: "route", Category: "io", Signature: "route", Description: "Route connection traffic through the pipeline"},
			{Name: "assert_unordered", Category: "testing", Signature: "assert_unordered <items...>", Description: "Assert the stream equals the items in any order"},
			{Name: "example_test", Category: "testing", Signature: "example_test <name> ...", Description: "Run an inline example test"},
			{Name: "mock_stream", Category: "testing", Signature: "mock_stream <n>", Description: "Deterministic mock stream of n nodes"},
			{Name: "mock_users", Category: "testing", Signature: "mock_users <n>", Description: "Deterministic mock user records"},
			{Name: "graph_add_edge_inferred", Category: "graph", Signature: "graph_add_edge_inferred <color> <from> <to>", Description: "Add an inferred (derived) edge"},
			{Name: "graph_add_fanout", Category: "graph", Signature: "graph_add_fanout <color> <from> <to...>", Description: "Add one edge from a node to many"},
			{Name: "graph_add_fanin", Category: "graph", Signature: "graph_add_fanin <color> <from...> <to>", Description: "Add edges from many nodes to one"},
			{Name: "graph_remove_node", Category: "graph", Signature: "graph_remove_node <id>", Description: "Remove a node and its edges"},
			{Name: "graph_remove_edge", Category: "graph", Signature: "graph_remove_edge <color> <from> <to>", Description: "Remove an edge"},
			{Name: "graph_edges", Category: "graph", Signature: "graph_edges", Description: "Stream all edges"},
			{Name: "graph_edges_inferred", Category: "graph", Signature: "graph_edges_inferred", Description: "Stream the inferred edges"},
			{Name: "graph_edges_of", Category: "graph", Signature: "graph_edges_of <id>", Description: "Stream the edges touching a node"},
			{Name: "graph_edges_of_group", Category: "graph", Signature: "graph_edges_of_group <color>", Description: "Stream the edges of one colour"},
			{Name: "graph_nodes_of", Category: "graph", Signature: "graph_nodes_of <color>", Description: "Stream the nodes of one colour"},
			{Name: "graph_in", Category: "graph", Signature: "graph_in <id>", Description: "Incoming neighbours of a node"},
			{Name: "graph_out", Category: "graph", Signature: "graph_out <id>", Description: "Outgoing neighbours of a node"},
			{Name: "graph_neighbors", Category: "graph", Signature: "graph_neighbors <id>", Description: "All neighbours of a node"},
			{Name: "graph_bfs", Category: "graph", Signature: "graph_bfs <id>", Description: "Breadth-first walk from a node"},
			{Name: "graph_dfs", Category: "graph", Signature: "graph_dfs <id>", Description: "Depth-first walk from a node"},
			{Name: "graph_walk_from", Category: "graph", Signature: "graph_walk_from <id>", Description: "Walk the graph from a node"},
			{Name: "graph_focus", Category: "graph", Signature: "graph_focus <id>", Description: "Focus the graph on a node's payload (lens)"},
			{Name: "graph_set", Category: "graph", Signature: "graph_set <id> <val>", Description: "Replace a node's payload"},
			{Name: "graph_over", Category: "graph", Signature: "graph_over <id> <f>", Description: "Modify a node's payload with a function"},
			{Name: "graph_union", Category: "graph", Signature: "graph_union <other>", Description: "Union of two graphs"},
			{Name: "graph_map_nodes", Category: "graph", Signature: "graph_map_nodes <f>", Description: "Map node payloads"},
			{Name: "graph_map_edges", Category: "graph", Signature: "graph_map_edges <f>", Description: "Map edge payloads"},
			{Name: "graph_bimap", Category: "graph", Signature: "graph_bimap <f> <g>", Description: "Map node and edge payloads"},
			{Name: "graph_filter_nodes", Category: "graph", Signature: "graph_filter_nodes <pred>", Description: "Keep nodes matching a predicate"},
			{Name: "graph_filter_edges", Category: "graph", Signature: "graph_filter_edges <pred>", Description: "Keep edges matching a predicate"},
			{Name: "graph_fold_nodes", Category: "graph", Signature: "graph_fold_nodes <f> <seed>", Description: "Fold over nodes"},
			{Name: "graph_fold_edges", Category: "graph", Signature: "graph_fold_edges <f> <seed>", Description: "Fold over edges"},
			{Name: "graph_fold_walk", Category: "graph", Signature: "graph_fold_walk <id> <f> <seed>", Description: "Fold over a walk from a node"},
			{Name: "ebnf_parse", Category: "parser", Signature: "ebnf_parse <rule> <grammar>", Description: "Compile dynamic EBNF grammar and parse string stream into AST"},
		},
		Operators: []OperatorDef{
			// Pipes
			{Symbol: "|>", Unicode: "⊳", Category: "pipe", Description: "Sequential pipeline stream forward"},
			{Symbol: "||>", Category: "pipe", Description: "Parallel pipeline stream execution"},

			// Arrows
			{Symbol: "->", Unicode: "→", Category: "arrow", Description: "Pattern match branch arrow and lambda return"},
			{Symbol: "\\", Category: "arrow", Description: "Lambda introducer: \\x y -> body"},
			{Symbol: "<-", Category: "arrow", Description: "List comprehension generator source binding"},

			// Applicatives & Kleisli
			{Symbol: "<+>", Category: "arrows", Description: "ArrowPlus: merge several producers into one stream, in order"},
			{Symbol: "&&&", Category: "arrows", Description: "Kleisli FanOut: route single input to two parallel arrows"},
			{Symbol: "***", Category: "arrows", Description: "Applicative Split: apply independent arrows to Pair halves"},

			// Infix Math & Logic
			{Symbol: "!", Category: "logic", Description: "Prefix logical NOT (when not followed by =)"},
			{Symbol: "||", Category: "logic", Precedence: 1, Associativity: "left", Description: "Logical OR"},
			{Symbol: "&&", Category: "logic", Precedence: 2, Associativity: "left", Description: "Logical AND"},
			{Symbol: "==", Category: "comparison", Precedence: 3, Associativity: "left", Description: "Structural equality test"},
			{Symbol: "!=", Category: "comparison", Precedence: 3, Associativity: "left", Description: "Structural inequality test"},
			{Symbol: "<", Category: "comparison", Precedence: 4, Associativity: "left", Description: "Less than comparison"},
			{Symbol: "<=", Category: "comparison", Precedence: 4, Associativity: "left", Description: "Less than or equal comparison"},
			{Symbol: ">", Category: "comparison", Precedence: 4, Associativity: "left", Description: "Greater than comparison"},
			{Symbol: ">=", Category: "comparison", Precedence: 4, Associativity: "left", Description: "Greater than or equal comparison"},
			{Symbol: "+", Category: "math", Precedence: 5, Associativity: "left", Description: "Infix addition with automatic overflow promotion"},
			{Symbol: "-", Category: "math", Precedence: 5, Associativity: "left", Description: "Infix subtraction with unsigned underflow protection"},
			{Symbol: "++", Category: "math", Precedence: 5, Associativity: "left", Description: "Monoidal concatenation for strings and lists"},
			{Symbol: "*", Category: "math", Precedence: 6, Associativity: "left", Description: "Infix multiplication with automatic overflow promotion"},
			{Symbol: "/", Category: "math", Precedence: 6, Associativity: "left", Description: "Infix true division (promotes integers to float)"},
			{Symbol: "//", Category: "math", Precedence: 6, Associativity: "left", Description: "Infix Euclidean integer floor division"},
			{Symbol: "%", Category: "math", Precedence: 6, Associativity: "left", Description: "Infix modulo / remainder"},
			{Symbol: "**", Category: "math", Precedence: 7, Associativity: "right", Description: "Exponentiation (exact for Int ** non-negative Int)"},
		},
		Snippets: []SnippetDef{
			{
				Prefix:      "def",
				Body:        "def ${1:function_name} = ${2:pipeline_body}",
				Description: "Define a reusable Gatwd pipeline function",
			},
			{
				Prefix:      "spawn",
				Body:        "spawn ${1:Type}: ${2:initial_value} |> ${3:step}",
				Description: "Spawn a typed AST stream pipeline",
			},
			{
				Prefix:      "match",
				Body:        "match {\n    [] -> empty\n    | [${1:pivot}, ...${2:rest}] -> ${3:step}\n}",
				Description: "Pattern matching with ADT head/tail list destructuring",
			},
			{
				Prefix:      "match_tuple",
				Body:        "match {\n    [0, ${1:n}] -> ${1:n} + 1\n    | [${2:m}, 0] -> [${2:m} - 1, 1] |> ${3:func}\n    | [${2:m}, ${1:n}] -> let res = [${2:m}, ${1:n} - 1] |> ${3:func} in [${2:m} - 1, res] |> ${3:func}\n}",
				Description: "Pattern matching with 2-element tuple/list destructuring (Ackermann style)",
			},
			{
				Prefix:      "let",
				Body:        "let ${1:variable} = ${2:expression} in ${3:body}",
				Description: "Scoped let-binding expression",
			},
			{
				Prefix:      "comp",
				Body:        "[${1:x} | ${1:x} <- ${2:source}, ${3:condition}]",
				Description: "Declarative list comprehension",
			},
			{
				Prefix:      "quicksort",
				Body:        "def quicksort = match {\n    [] -> empty\n    | [pivot, ...rest] -> seq { [x | x <- rest, x < pivot] |> quicksort | pivot | [x | x <- rest, x >= pivot] |> quicksort }\n}",
				Description: "Pure Turing-complete Quicksort using list comprehensions",
			},
			{
				Prefix:      "ackermann",
				Body:        "def ackermann 0 n = n + 1\ndef ackermann m 0 = ackermann (m - 1) 1\ndef ackermann m n = ackermann (m - 1) (ackermann m (n - 1))",
				Description: "Canonical non-primitive recursive Ackermann function",
			},
			{
				Prefix:      "type",
				Body:        "type ${1:TypeName} = ${2:Definition}",
				Description: "Declare an algebraic data type or record",
			},
			{
				Prefix:      "const",
				Body:        "const ${1:NAME} = ${2:value}",
				Description: "Declare a script constant",
			},
			{
				Prefix:      "hamt",
				Body:        "to_hamt |> hamt_set \"${1:key}\" \"${2:value}\" |> hamt_get \"${1:key}\"",
				Description: "Persistent HAMT key-value pipeline",
			},
		},
	}
}

// WithFunctionForms returns s extended with one builtin per expression-level function
// form. names maps a function name to its parameter names (data-last). A name already
// listed in s.Builtins (a pipeline step with the same name) keeps its richer entry, so
// the merge only ever adds. cmd/gatwd-gen feeds it from the live interpreter registry
// (cli.FunctionForms), which is what keeps the generated editor support from drifting.
func (s GrammarSpec) WithFunctionForms(names []string, params map[string][]string) GrammarSpec {
	seen := map[string]bool{}
	for _, b := range s.Builtins {
		seen[b.Name] = true
	}
	out := s
	out.Builtins = append([]BuiltinDef(nil), s.Builtins...)
	for _, n := range names {
		if seen[n] {
			continue
		}
		sig := n
		for _, p := range params[n] {
			sig += " " + p
		}
		out.Builtins = append(out.Builtins, BuiltinDef{Name: n, Category: "function", Signature: sig, Description: "Expression-level function form (curried, data-last)"})
	}
	return out
}
