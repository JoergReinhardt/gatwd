# gatwd DSL Guide

`gatwd` is a purely functional, Turing-complete Scripting Language for declarative infrastructure crawling, evaluation, and mutation. It heavily borrows syntax and philosophy from Haskell, Elm, and F#.

## 1. Modules, Types, and Go Primitives

In `gatwd`, you define **Types** and **Functions** (`def`) in library modules, and bring them into your runtime using `import`.

**types.gatwd**
```haskell
-- Sum Types (Dynamic Enums)
type OS = Linux | Windows | Darwin

-- Record & Collection Types
type ServerConfig = { IP: String, Port: Int }
type GiantList = List(Int)
type BinaryTree = RBTree(String)

-- Native Go Primitive Bindings! 
-- (Assuming your main.go registered these via Interpreter.RegisterTypeFactory)
type Balance = BigInt 
type Timestamp = GoTime 
```

**main.gatwd**
```haskell
import "types.gatwd"

-- Functions (Aliases) encapsulate pipeline logic
def audit_linux = uname -a |> expect "Linux" "CRITICAL: Not Linux"
def extract_ip = parse_record IP Port |> focus IP

spawn OS: Linux |> audit_linux |> extract_ip
```

## 2. Advanced Tree Structures

`gatwd` provides immutable, natively implemented Trees designed to be graphed dynamically.

* `build_tree "Root"`: Standard Rose Tree.
* `build_bst`: Standard Binary Search Tree.
* `build_btree`: N-ary chunked B-Tree.
* `build_rbtree`: Okasaki purely functional Red-Black Tree.

```haskell
-- Feed 10 strings into a B-Tree and beautifully graph it.
spawn GiantList: Init |> cat data.txt |> build_btree |> visualize --color=green --curved
```

## 3. The Visualizer ("Blackjack and Hookers")

The `visualize` keyword invokes an extremely advanced UTF-8 / ASCII formatting engine that runs natively over Go structural reflection.

* **Conditional Coloring**: `--color-type=TypeName:cyan`, `--color-val=Root:green`
* **Outer Borders**: `--border`, `--auto-border=5` (only draw if height > 5 lines)
* **Nested Sub-Boxes**: `--box-all`, `--box=Record,RBNodeGraph` (Draws elegant flow-graph boxes precisely aligned with the tree layout).
* **Array Slicing**: `--head=N`, `--tail=M` (Safely skips central indices and outputs `... (X items omitted) ...`).
* **Styles**: `--curved` (`╭─╮`), `--heavy` (`┏━┓`), `--ascii` (`+--`), `--dot` (Graphviz DOT compilation).
* **Go fmt.Stringer**: The visualizer defaults to invoking `String() string` interfaces for native Go objects (like `net.IP` or `time.Time`), while intelligently preserving structural ANSI coloring!

```haskell
spawn GiantList: Init |> cat big.txt |> build_array |> visualize --head=2 --tail=2 --color-type=List:cyan
```

## 4. Local State (`let ... in`)

While pipelines (`|>`) flow data sequentially, you often need to store intermediate state. Use `let` bindings to capture terminal nodes and substitute them into subsequent branches.

```haskell
let cluster_ip = spawn IPAddress: 192.168.1.100 |> math_add 1 in
spawn Server: Worker |> ping cluster_ip
```

## 5. Lambdas, Kleisli Arrows, & Optics

`gatwd` relies entirely on Category Theory to safely mutate values without side-effects.
* `(A &&& B)` (Kleisli FanOut): Routes a node into TWO independent functions concurrently, yielding a Tuple `(Res1, Res2)`.
* `[A *** B]` (Applicative Split): Takes a Tuple, and routes the Left side to Function A, and Right side to Function B.
* `update` / `focus`: First-class Lenses for manipulating inner structs immutably.
* `into_zipper` / `up` / `down`: Comonadic navigation over Trees.

```haskell
def get_time = date
def get_user = whoami

spawn OS: Linux |> (get_time &&& get_user) |> [echo Time_%s *** echo User_%s] |> format_tuple "(%s, %s)"
```

## 6. Top-Level Script Constants (`const`)

Define compile-time immutable constants that can be used across spawn values and pipeline arguments:

```haskell
const BASE_PORT = 8080
const MAX_RETRIES = 5

spawn IntVal: BASE_PORT |> math_add 1 |> assert 8081
```

## 7. ADT Pattern Matching, List Destructuring & Non-Primitive Recursion

Pattern matching supports full list deconstruction (`[pivot, ...rest]`) as well as fixed-length tuple/list patterns (`[0, n]`, `[m, 0]`, `[m, n]`). This enables pure functional recursion like Quicksort using list comprehensions, as well as non-primitive recursive functions like Ackermann:

```haskell
-- Turing-complete Quicksort with List Comprehensions
def quicksort = match {
    [] -> empty
    | [pivot, ...rest] -> seq { [x | x <- rest, x < pivot] |> quicksort | pivot | [x | x <- rest, x >= pivot] |> quicksort }
}

spawn IntVal: 0 |> emit 5, 2, 8, 1, 9, 3 |> drop 1 |> collect_list |> quicksort |> collect_list |> assert [1 2 3 5 8 9]

-- Computability Theory: Non-Primitive Recursive Ackermann Function
def ackermann = match {
    [0, n] -> n + 1
    | [m, 0] -> [m - 1, 1] |> ackermann
    | [m, n] -> let res = [m, n - 1] |> ackermann in [m - 1, res] |> ackermann
}

[2, 2] |> ackermann |> assert 7
[3, 2] |> ackermann |> assert 29
```

## 8. Persistent HAMT Dictionaries (`to_hamt`, `hamt_get`, `hamt_set`, `hamt_del`, `hamt_keys`, `hamt_values`)

Fold key-value streams into a 32-way branching persistent Hash Array Mapped Trie, and perform pure structural sharing updates:

* `to_hamt`: Aggregates the stream of key-value pairs into an immutable HAMT dictionary.
* `hamt_get <key>`: Looks up a value by key.
* `hamt_set <key> <val>`: Yields a new persistent HAMT with the key-value pair set.
* `hamt_del <key>`: Yields a new persistent HAMT with the key removed.
* `hamt_keys`: Unfolds a HAMT into a stream of its keys.
* `hamt_values`: Unfolds a HAMT into a stream of its values.

```haskell
-- Aggregate and update HAMT
spawn StrVal: Start |> emit user:alice |> to_hamt |> hamt_set env prod |> hamt_get env |> assert prod

-- Delete key from HAMT
spawn StrVal: Start |> emit a:1, b:2 |> to_hamt |> hamt_del a |> hamt_get b |> assert 2
```

## 9. Performance Chunks & Continuations (`chunk`, `unchunk`, `checkpoint`, `telemetry`)

* `chunk <size>`: Batches lazy streams into eager, bounded contiguous arrays for CPU cache efficiency.
* `unchunk` (or `flatten_chunk`): Unpacks chunk nodes back into a flat sequential stream of elements.
* `checkpoint <id> <interval>`: Watermarks stream offsets, allowing computations to pause and resume.
* `telemetry`: Zero-overhead element counter and duration profiler.

```haskell
spawn IntVal: 0 |> emit 1, 2, 3, 4 |> drop 1 |> chunk 2 |> head |> assert [1 2]
spawn IntVal: 0 |> emit 10, 20, 30 |> drop 1 |> chunk 2 |> unchunk |> collect_list |> assert [10 20 30]
spawn IntVal: 0 |> emit 1, 2, 3 |> drop 1 |> checkpoint job_1 2 |> math_add 10 |> head |> assert 11
```

## 10. IO Streaming & Network Adapters (`scan_file`, `scan_words`, `dial_tcp`, `listen_tcp`, `conn_lines`)

* `scan_file <path>`: Streams lines from a file using the generalized scanner.
* `scan_words <path>`: Streams whitespace-delimited words using Go's `bufio.ScanWords`.
* `dial_tcp <address>`: Transmits payloads across a TCP socket and streams back response lines.
* `listen_tcp <address> [count]`: Opens a TCP listener and unfolds accepted connections into a stream.
* `conn_lines`: Unfolds an accepted network connection into a line-by-line stream.

```haskell
spawn StrVal: Start |> scan_file examples/tour/data/data1.txt |> head |> assert M
spawn StrVal: Start |> scan_words examples/tour/data/data1.txt |> head |> assert M
spawn StrVal: "ping" |> dial_tcp "localhost:9090" |> assert "ECHO:ping"
```

## 11. Persistent Hash Sets (`to_set`, `set_add`, `set_contains`, `set_remove`, `set_elements`)

* `to_set`: Aggregates the stream into a deduplicated, persistent HAMT-backed Set.
* `set_add <item>`: Yields a new Set with `<item>` added.
* `set_contains <item>`: Checks membership, yielding `true` or `false`.
* `set_remove <item>`: Yields a new Set with `<item>` removed.
* `set_elements`: Unfolds the Set into a flat stream of its elements.

```haskell
spawn StrVal: Start |> emit ruby |> to_set |> set_add golang |> set_contains golang |> assert true
spawn StrVal: Start |> emit a, b |> drop 1 |> to_set |> set_remove a |> set_contains a |> assert false
spawn StrVal: Start |> emit x |> drop 1 |> to_set |> set_elements |> assert x
```

## 12. Stream Transducers (`filter`, `distinct`, `fold`)

* `filter <substring>`: Filters nodes whose string representation contains the target substring.
* `distinct`: Deduplicates consecutive and recurring elements across the stream, maintaining order.
* `fold <op>`: Catamorphism reducing the stream to a single summary value (e.g. `fold count` counts elements).

```haskell
-- Filter stream elements matching a substring
spawn StrVal: Start |> emit alpha, beta, gamma |> drop 1 |> filter bet |> assert beta

-- Deduplicate elements preserving stream order
spawn IntVal: 0 |> emit 5, 10, 5, 10 |> drop 1 |> distinct |> collect_list |> assert [5 10]

-- Fold catamorphism counting items
spawn StrVal: Start |> emit x, y, z |> drop 1 |> fold count |> assert 3
```

## 13. First-Class Optics Lenses (`focus`, `update`)

* `focus <field>`: Projects into a record or map, selecting the named field.
* `update <field> <new_val>`: Immutably updates the named field on a record or map, yielding a new structure via structural sharing.

```haskell
spawn RecordVal: {Owner: guest, Role: viewer} |> update Owner admin |> focus Owner |> assert admin
```

## 14. Pipeline Visualization (`visualize`)

* `visualize [flags...]`: Directly pipe any AST node, collection, tree, or record through the visual formatting engine. Formats the data structure into an ANSI card or box. Supported flags include `--curved`, `--heavy`, `--ascii`, `--border`, `--border-color=<color>`, `--bg-color=<color>`, `--color=<color>`, `--head=N`, `--tail=M`.

```haskell
spawn RecordVal: {Title: PipelineCard, Status: Active} |> visualize --curved --border --border-color=yellow
```

## 15. Infix Expressions & Operator Overloading (`+`, `-`, `*`, `/`, `//`, `%`, `++`, `==`, `<`, `&&`, `||`)

Gatwd supports full infix mathematical, relational, and boolean expressions with standard operator precedence and parentheses everywhere: in `spawn <Type>: <expr>`, in `let <var> = <expr>`, and as pipeline operator sections (`|> * 5`, `|> / 2`).

### Supported Infix Operators
- **Arithmetic**:
  - `+` (Addition, string concatenation)
  - `-` (Subtraction)
  - `*` (Multiplication)
  - `/` (True division: **automatically promotes integer division to `Float` / `float64`**, e.g. `7 / 2 == 3.5`)
  - `//` (Euclidean integer division, e.g. `7 // 2 == 3`)
  - `%` (Modulo / remainder)
- **Monoidal Concatenation**:
  - `++` (String and collection concatenation, e.g. `"Hello " ++ "Gatwd"`)
- **Relational & Equality**:
  - `==`, `!=`, `<`, `<=`, `>`, `>=` (Structural equality and scalar comparisons)
- **Boolean Logic**:
  - `&&`, `||`

### Automatic Overflow Promotion to `BigInt` & `BigFloat`
Gatwd protects numerical pipelines against arithmetic overflow across all numeric types:
- **Signed Integers (`Int`, `int64`)**: Operations exceeding `math.MaxInt64` / `math.MinInt64` automatically promote to arbitrary-precision `*big.Int`.
- **Unsigned Integers (`Uint`, `uint64`)**: Addition or multiplication exceeding `math.MaxUint64` promotes to `*big.Int`; unsigned subtraction underflow promotes to negative signed `int64` / `*big.Int`.
- **Floating Point (`Float`, `float64`)**: Operations that would overflow IEEE-754 limits into `+Inf`/`-Inf` automatically promote to arbitrary-precision `*big.Float`.
- **Complex Numbers (`Complex`, `complex128`)**: Overloaded arithmetic over complex numbers (e.g. `1+2i + 3+4i == (4+6i)`), promoting component floats to `*big.Float` if infinite overflow occurs.

### List Comprehensions (`[expr | var <- source, filter...]`)
Gatwd provides native Haskell/Python-style list comprehensions over collections and ranges:

```haskell
-- List comprehension with generator and filter predicate:
spawn ListVal: [x * 2 | x <- [1, 2, 3, 4], x > 2] |> assert [6 8]

-- True division promotes integers to float:
spawn FloatVal: 7 / 2 |> assert 3.5

-- Automatic integer and unsigned overflow promotion:
spawn BigIntVal: 9223372036854775807 + 1 |> assert 9223372036854775808
spawn BigIntVal: 18446744073709551615 + 1 |> assert 18446744073709551616

-- Complex number arithmetic:
spawn ComplexVal: 1+2i + 3+4i |> assert (4+6i)

-- Pipeline Operator Sections:
spawn IntVal: 10 |> * 5 |> / 2 |> assert 25
```

## 16. Synthetic Mock Streams & Go Example Assertions

Gatwd brings Go standard library testing conventions (`testing.InternalExample`) directly into script pipelines:

* `mock_stream <count>`: Generates lazy streams of integers `0, 1, ..., N-1` for mock benchmarking.
* `mock_users <count>`: Generates lazy streams of structured domain structs (`MockUser{ID, Name, Role, Score, Active}`).
* `assert_unordered <values...>`: Verifies that pipeline elements match the expected set regardless of permutation order (mirroring Go's `// Unordered output:`).
* `example_test <values...>`: Validates exact sequential stream output (mirroring Go's `// Output:`).

```haskell
-- Generate synthetic integers and chain with math
spawn IntVal: 0 |> mock_stream 3 |> math_add 10 |> collect_list |> assert [10 11 12]

-- Stream domain structs through filtering
spawn StrVal: Start |> mock_users 2 |> head |> assert {0 admin_A admin 0 true}

-- Unordered permutation-invariant matching
spawn IntVal: 0 |> emit 20, 10 |> drop 1 |> assert_unordered 10, 20 |> assert PASS

-- Ordered example output validation
spawn IntVal: 0 |> emit 10, 20 |> drop 1 |> example_test 10, 20 |> assert PASS
```

## 17. Runtimes, Builtins, Editions & Appliances

Gatwd's runtime model is categorized into four distinct concepts:

* **Bare Runtime (`cli.NewBareInterpreter()`)**: A minimalist execution kernel providing only fundamental language primitives (`let`, `def`, `const`, `match`, `seq`, `emit`, `take`, `drop`, `head`, `tail`, `assert`). It excludes heavy standard library adapters and transducers, ideal for sandboxed or footprint-constrained environments.
* **Gatwd Builtins (`cli.NewInterpreter()`)**: The standard full distribution exposing Go's standard library capabilities directly to the DSL as built-in transducers (persistent collections `HAMT`, `Set`, `Chunk`, `List`, `Tree`; stream adapters `scan_file`, `dial_tcp`, `walk_fs`; testing & telemetry `example_test`, `assert_unordered`, `checkpoint`).
* **Gatwd Editions (or Distributions)**: Tailored Go binary distributions (such as `cmd/gatwd-cloud` or `cmd/gatwd-k8s`) that import `github.com/joergreinhardt/gatwd/pkg/cli`, bundling third-party Go packages, custom domain types, and specialized transducers using the fluent builder:
  ```go
  cli.NewEdition("gatwd-cloud").
      WithType("Email", parseEmail).
      WithStep(cli.MapStep("mask", maskSecret)).
      Run()
  ```
* **Gatwd Appliances**: Self-contained, single-purpose releases (such as Kubernetes Operators, stream processing microservices, or compliance auditing CLIs) that bundle pre-compiled, embedded `.gatwd` DSL code (`//go:embed`) via `cli.NewAppliance`. Appliances run out of the box without external script files, while allowing live overrides via command-line flags or Kubernetes ConfigMaps with zero-downtime hot-reloading:
  ```go
  app, _ := cli.NewAppliance(
      cli.WithEmbeddedScript(embeddedRules),
      cli.WithConfigMapOverride("/etc/gatwd/rules.gatwd"),
  )
  res := app.Execute("spawn Status: OK |> check_rules")
  ```



