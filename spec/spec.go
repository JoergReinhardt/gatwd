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
	LanguageName string        `json:"language_name"`
	FileExtension string       `json:"file_extension"`
	CommentPrefix string       `json:"comment_prefix"`
	Keywords     []KeywordDef  `json:"keywords"`
	Types        []TypeDef     `json:"types"`
	Builtins     []BuiltinDef  `json:"builtins"`
	Operators    []OperatorDef `json:"operators"`
	Snippets     []SnippetDef  `json:"snippets"`
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
			{Name: "ebnf_parse", Category: "parser", Signature: "ebnf_parse <rule> <grammar>", Description: "Compile dynamic EBNF grammar and parse string stream into AST"},
		},
		Operators: []OperatorDef{
			// Pipes
			{Symbol: "|>", Unicode: "⊳", Category: "pipe", Description: "Sequential pipeline stream forward"},
			{Symbol: "||>", Category: "pipe", Description: "Parallel pipeline stream execution"},

			// Arrows
			{Symbol: "->", Unicode: "→", Category: "arrow", Description: "Pattern match branch arrow and lambda return"},
			{Symbol: "<-", Category: "arrow", Description: "List comprehension generator source binding"},

			// Applicatives & Kleisli
			{Symbol: "&&&", Category: "arrows", Description: "Kleisli FanOut: route single input to two parallel arrows"},
			{Symbol: "***", Category: "arrows", Description: "Applicative Split: apply independent arrows to Pair halves"},

			// Infix Math & Logic
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
				Body:        "def ackermann = match {\n    [0, n] -> n + 1\n    | [m, 0] -> [m - 1, 1] |> ackermann\n    | [m, n] -> let res = [m, n - 1] |> ackermann in [m - 1, res] |> ackermann\n}",
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
