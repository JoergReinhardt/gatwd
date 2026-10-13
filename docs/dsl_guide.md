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

### Sum-Type Constructor Matching
When a `match` arm names a constructor of a declared sum type, it discriminates on the value's branch — each constructor routes to its own arm. Arms whose name is *not* a registered constructor are treated as fresh variable bindings that match any value (as in `conway_step`'s `live -> …`). A `_` arm or a bare variable arm is a catch-all.

```haskell
type OS = Linux | Windows | Darwin

def route_os = match { Linux -> echo PENGUIN | Windows -> echo WINDOW | Darwin -> echo APPLE }

spawn OS: Linux   |> route_os   -- PENGUIN
spawn OS: Windows |> route_os   -- WINDOW
```

> **Tip:** `gatwd --check` warns when a `match` over a closed sum type is non-exhaustive (see §19), e.g. *"non-exhaustive match on OS: missing Darwin"*.

### List & Tuple Patterns
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

### `Char` literals and functions (hardening W2, D1)

`'a'`, `'é'`, `'\n'`, `'\x41'`, `'\u00e9'` (escapes `\n \t \r \0 \\ \' \" \xHH \uHHHH`) are values of type **`Char`**: one Unicode
code point. A `Char` is neither an `Int` nor a `String` — `'a' + 1` is an error — but they convert **on demand**: wherever a
`Char` is expected a one-character `String` is accepted (`char_code "é"` is `233`, `'a' == "a"` is `true`), and ordering is
by code point (`'a' < 'b'`). The function forms are data-last and curried (they also work inside lambdas and comprehensions):

| Function | Meaning |
| :-- | :-- |
| `char_code c` · `char_from_code n` | `Char` ↔ `Int` code point (`char_from_code` is an error for a surrogate or `> 0x10FFFF`) |
| `char_to_str c` · `char_of_str s` | `Char` → `String` · `String` → `Maybe Char` (`Nothing` unless `s` is exactly one character) |
| `char_is_digit` `char_is_alpha` `char_is_space` `char_is_upper` `char_is_lower` | `Char` → `Bool` |
| `char_upper c` · `char_lower c` | `Char` → `Char` |

The `Char` annotation is understood by `gatwd --check` (`def isa : Char -> Bool = \c -> c == 'a'`). 

**`String` is a collection row of `Char`s.** The expression function forms see a `String` as its characters:
`length`, `head`/`nth` (a `Char`), `member`, `fold`, `to_list`/`to_set`, and `map`/`flatmap` (a `List` of whatever the
function returns — `map char_upper "abc"` is a list of `Char`s). The *closed* operations answer with a `String`:
`filter`, `reject`, `reverse`, `tail`, `cons 'x' s`. A collection of `Char`s used where a `String` is demanded converts:
`"ab" ++ ['c', 'd']` is `"abcd"` (a mixed collection is an error naming the element; the empty collection spells `""`).
**In a pipeline a `String` node is one element of the stream** (a line, a word), so the stream steps `map`, `filter`,
`fold`, `head` … do not descend into it; inside a lambda or a function-form call use the function forms
(`["abc", "de"] |> map (\s -> reverse s)`).

### The `str_*` library (hardening W2)

Data-last, curried, usable as steps (`"a,b" |> str_split ","`), positions count **characters**, and a collection of `Char`s is
accepted wherever a `String` is demanded. Malformed text is a `Maybe`; an impossible request (empty separator, a slice out of
range, a `%d` given a string) is an `[ERROR]`.

| Group | Functions |
| :-- | :-- |
| split / join | `str_split sep s` · `str_join sep xs` · `str_lines s` · `str_words s` · `str_chars s` · `str_from_chars cs` |
| transform | `str_trim` · `str_upper` · `str_lower` · `str_repeat n s` · `str_replace old new s` · `str_slice from to s` (half-open) · `str_pad_left n ch s` · `str_pad_right n ch s` |
| query | `str_contains sub s` · `str_starts_with p s` · `str_ends_with p s` · `str_index_of sub s` (`Maybe Int`, in characters) |
| parse | `str_to_int s` · `str_to_float s` — `Maybe`; decimal syntax only (no surrounding space, no `nan`/`inf`/hex) |
| display | `str_from x` (the printed form) · `show x` (`"a"`, `'a'`, `['a', "b", 1]`) · `str_format fmt args` |

`str_format` is a checked printf subset: `%s %v %d %f %g %e %x %q %c %%` with flags, width and precision
(`str_format "%5d|%-5s|%.2f" [42, "ab", 3.14159]`); each verb names the type it needs, and a mismatch, a missing or a surplus
argument is an error. String literals have **no escapes yet** (`"\n"` is a backslash and an `n`): build a newline with
`char_to_str '\n'`.

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

## 18. Functional Operations Completeness & Typeclass Parity

Gatwd achieves full operation completeness across persistent types (`List`, `Chunk`, `Set`, `HAMT`, `Tree`, `Pair`, `Either`, `Maybe`), bringing parity with Haskell (`Foldable`, `Functor`, `Semigroup`/`Monoid`, `Ord`/`Eq`), Clojure, and Scala:

### Extended Folds (Foldable)
Collapse lazy streams or persistent collections into scalar reductions:
* `fold sum`: Numerically adds elements (supports integers, floats, and big numbers).
* `fold product` (or `fold prod`): Computes element product (starting at 1).
* `fold min`: Finds minimum element using total lexicographical ordering.
* `fold max`: Finds maximum element using total lexicographical ordering.
* `fold concat`: Monoidal reduction over elements with `++`.
* `fold any` (or `fold or`): Boolean disjunction (short-circuits on first truthy item).
* `fold all` (or `fold and`): Boolean conjunction.

```haskell
spawn IntVal: 0 |> emit 10, 20, 30 |> drop 1 |> fold sum |> assert 60
spawn IntVal: 0 |> emit 2, 3, 4 |> drop 1 |> fold product |> assert 24
spawn IntVal: 0 |> emit 50, 15, 80 |> drop 1 |> fold min |> assert 15
```

#### Custom-Accumulator Fold (`fold (\acc x -> expr) seed`)
Beyond the fixed reducers, a two-argument lambda with an explicit seed expresses any left fold, threading `acc` across the elements:

```haskell
let xs = [1, 2, 3, 4] in xs |> fold (\acc x -> acc + x) 0 |> assert 10
spawn IntVal: 0 |> emit 2, 3, 4 |> drop 1 |> fold (\acc x -> acc * x) 1 |> assert 24
```

### Bifunctor Mapping (`bimap`)
`bimap (\x -> e1) (\y -> e2)` applies `e1` to the first/`Left` component and `e2` to the second/`Right` component of a `Pair` (e.g. from `&&&`), an `Either`, or a 2-element list:

```haskell
let p = [3, 4] in p |> bimap (\x -> x + 10) (\y -> y * 2) |> assert [13 8]
```

### Functor Mapping (`map`)
Applies higher-order functions or sub-pipelines across both stream nodes and persistent collections (`List`, `Chunk`, `Set`, `Slice`):
* Lambda mapping: `map (\x -> expr)` evaluates `expr` with `x` bound to each item.
* Pointfree step mapping: `map <step>` or `map (<subpipeline>)`.

```haskell
-- Map across stream nodes
spawn IntVal: 0 |> emit 1, 2, 3 |> drop 1 |> map (\x -> x * 10) |> collect_list |> assert [10 20 30]

-- Map across persistent lists directly
let xs = [1, 2, 3] in xs |> map (\x -> x + 5) |> assert [6 7 8]
```

### Monadic FlatMap (`flatmap`)
`flatmap (\x -> expr)` is the Monad bind: the lambda yields a collection per element, and the results are flattened into one. Over a `List` the per-element lists are concatenated; over a stream each element's result is flattened and re-emitted. (At the Go API level `Bind` is also available on `List`, `Chunk`, `Set`, `Either`, and `Variant`.)

```haskell
let xs = [1, 2, 3] in xs |> flatmap (\x -> [x, x]) |> assert [1 1 2 2 3 3]
spawn IntVal: 0 |> emit 1, 2 |> drop 1 |> flatmap (\x -> [x, x]) |> collect_list |> assert [1 1 2 2]
```

### Set Algebra Builtins
Full mathematical set algebra with persistent structural sharing:
* `set_union <other>`: Set-theoretic union ($A \cup B$).
* `set_intersect <other>`: Set-theoretic intersection ($A \cap B$).
* `set_diff <other>`: Set-theoretic relative complement ($A \setminus B$).
* `hamt_merge <other>`: Right-biased map merge / keyed union ($H_1 \cup H_2$).
* `hamt_intersect <other>`: Keyed map intersection, values from the left ($H_1 \cap H_2$).
* `hamt_diff <other>`: Keyed map difference ($H_1 \setminus H_2$).

### Monoidal Concatenation (`++`) & Lexicographical Ordering (`<`, `<=`, `>`, `>=`)
Binary infix operators work uniformly across all types:
* `String ++ String`: String concatenation.
* `List ++ List` / `Chunk ++ Chunk`: Persistent sequence concatenation.
* `Set ++ Set`: Set Union.
* `HAMT ++ HAMT`: Map Merge.
* `Tree ++ Tree`: Rose Tree node attachment.
* `Pair ++ Pair`: Tuple element-wise concatenation.
* `[1, 2] < [1, 3]`: Lexicographical comparison over sequences, pairs, and ADTs (`Maybe`, `Either`).

### Stream & Collection Conversions
First-class transducers to aggregate lazy streams into persistent structures:
* `to_list`: Aggregates the stream into an immutable persistent Cons List (`List`).
* `to_chunk`: Aggregates the stream into a contiguous bounded array buffer (`Chunk`).
* `to_set`: Aggregates and deduplicates the stream into a persistent HAMT-backed `Set`.
* `to_hamt`: Aggregates key-value pairs or "k:v" strings into an immutable `HAMT` map.

### Lambda Predicate Filtering (`filter`, `reject`)
Filters streams or persistent collections (`List`, `Chunk`, `Set`, `Tree`, `Slice`) using lambdas:
* `filter (\x -> expr)`: Retains elements where `expr` evaluates to `true`.
* `reject (\x -> expr)`: Retains the complement — elements where `expr` is `false`.
* `filter "keyword"` / `reject "keyword"`: Substring matching (and its complement) over string nodes.

Over a `Tree`, `filter`/`reject` prune child subtrees by the predicate (the root is always kept).

```haskell
-- Filter over persistent List:
let xs = [1, 2, 3, 4, 5] in xs |> filter (\x -> x > 3) |> assert [4 5]

-- Filter over stream:
spawn IntVal: 0 |> emit 2, 5, 8, 11 |> drop 1 |> filter (\x -> x > 6) |> to_list |> assert [8 11]
```

### Algebraic Tuple & Tree Transducers
* `pair_swap`: Swaps components of a pair or 2-element collection `[a, b]` $\to$ `[b, a]`.
* `tree_size`: Evaluates the total node count of a Rose Tree.
* `tree_depth`: Evaluates the maximum depth/height of a Rose Tree.

## 19. First-Class Functions: Closures, Currying & Lexical Scope

A lambda `\x y -> expr` is an **expression** that evaluates to a first-class *closure*: a function value carrying its parameter list and the environment it was defined in (lexical scope). Closures are **curried** — applying fewer arguments than the arity yields a function awaiting the rest, and applying more applies the result to the surplus. A lambda-bodied `def` and a `let`-bound lambda are the same kind of value.

```haskell
def add = \x y -> x + y
def mk_adder = \k -> \x -> x + k       -- a function that returns a function

add 1 2 |> assert 3                      -- application by juxtaposition
(add 10) 5 |> assert 15                  -- partial application
mk_adder 5 10 |> assert 15               -- over-application through the returned closure

-- In a pipeline the incoming value is the LAST argument (data-last):
spawn IntVal: 5 |> add 3 |> assert 8     -- add 3 5

-- Pass closures to higher-order steps by name or as partial applications:
[1, 2, 3] |> map (add 10)  |> assert [11 12 13]
[1, 2, 3] |> fold add 0    |> assert 6
[1, -2, 3] |> filter pos   |> assert [1 3]       -- with def pos = \x -> x > 0

-- Bind a partial application, or a lambda, with let:
let add7 = mk_adder 7 in [1, 2] |> map add7 |> assert [8 9]
let f = \x y -> x * y in spawn IntVal: 6 |> f 7 |> assert 42
```

Rules worth knowing:

* **Lexical capture.** Two closures made by one factory keep their own captured values (`mk_adder 5` vs `mk_adder 9`).
* **Bare one-argument closure in a pipeline maps element-wise** over collections (`[1,2,3] |> inc`), exactly like `map inc`; with arguments (`x |> f a b`) it is applied to the whole incoming value as the last argument.
* **A call inside a plain list literal needs parentheses**: `[(f x), (g y)]`. Inside `[ … | … ]` comprehensions and `( … )` application works directly.
* **A `def` whose body is a pipeline stays a text macro** (a stream transformer such as `def k8s_ingest = fold (\g r -> …) graph_empty`); it takes no arguments. Use a lambda-bodied `def` for value functions.
* **Errors are values.** Applying a closure to an incompatible value yields an `[ERROR]` result naming the function (`sub: unsupported binary operator "-" for types string and string`), never a silent empty result.
* Unary minus applies only to numeric literals; write `0 - x` for a negated variable.

## 20. Static Checking, Gradual Typing & Loud Errors (`gatwd --check`)

`gatwd` fails loudly rather than silently. A pipeline step that is neither a known builtin/function nor an available executable, and a function call that fails, surface an error node instead of producing nothing:

```
spawn IntVal: 5 |> frobnicate
∷ [ERROR] unknown step or command: "frobnicate"
```

`gatwd --check <file.gatwd>` statically validates a script without running it, emitting `file:line:col:` diagnostics:

* **Errors** (exit 1): unbalanced `()`/`[]`/`{}`, a `def`/`const`/`type` declaration missing `=`, a pipeline ending in a dangling `|>`, and the legacy `:=` operator.
* **Warnings** (exit 0, advisory):
  * a `match` over a closed sum type that is **non-exhaustive** (`warning: non-exhaustive match on OS: missing Darwin`);
  * **gradual-typing findings**: wrong argument or result type, too many arguments (to a `def` or a builtin), an operator the runtime rejects for the operand types (`"a" - 1`, `[1] * 2`), an unbound name in an expression or lambda body, and a `spawn IntVal: hello` seed that contradicts a numeric/Bool `type` alias.

### Optional type annotations

`def name : Type = body` carries an optional signature. It never changes evaluation; it gives the checker something to check against and documents intent.

```haskell
type IntVal = Int
def inc   : Int -> Int        = \x -> x + 1
def add   : Int -> Int -> Int = \x y -> x + y
def total : List(Int) -> Int  = \xs -> fold (\a x -> a + x) 0 xs

inc "x"                       -- warning: argument 1 of 'inc' has type String, expected Int
spawn IntVal: 5 |> add "a"    -- warning: argument 1 of 'add' has type String, expected Int
```

`Type ::= Atom { "->" Type }` (right-associative) with `Atom ::= Name [ "(" Type {"," Type} ")" ] | "(" Type ")"`; names declared by `type Alias = Int` resolve to their right-hand side. Types flow locally from literals, operators, annotated parameters, `let` bindings and comprehension generators (`x <- [1, 2]` types `x`). Operator checks ask the **runtime's own operator matrix** (`functions.EvalBinaryOp`) on representative values, so the checker cannot disagree with the evaluator about what is defined.

The checker is *gradual*: anything it cannot determine — a runtime value, a pipeline step, a builtin without a signature — is `Any`, which is consistent with every type. That is why committed scripts check clean and why unannotated code is never second-class.

```bash
gatwd --check examples/tour/09_pattern_matching.gatwd
# examples/tour/09_pattern_matching.gatwd:6:1: warning: non-exhaustive match on OS: missing Darwin
```

## 21. The Operator-Type Matrix as Functions

Every operation exists in two forms. The **step form** takes literal tokens and sits in a pipeline (`xs |> hamt_set k 1`); the **function form** takes *computed values* and composes inside lambdas, folds and comprehensions (`hamt_set (hamt_get "k" r) n m`). Function forms are **curried**, **data-last** (`x |> f a` ≡ `f a x`), **type-preserving** (elements are never stringified) and report errors as values. **Every function form is also a step**: a pipeline step whose head names a builtin that no other step claims applies it to the whole incoming value (`[1, 2, 3] |> length` is `3`, `'a' |> char_upper` is `A`, `xs |> member 2`); the older stream steps (`head`, `tail`, `map`, `fold`, …) keep their meaning, and inside a lambda the function form is the one to use.

| Family | Functions |
| :--- | :--- |
| Functor / Filterable / Foldable / Monad | `map` `filter` `reject` `fold` `flatmap` `bimap` — dispatch on the carrier: List, Chunk, Set, HAMT (values), Tree, Pair (second component), Maybe, Either (right), Variant (payload) |
| Sequences | `length` `is_empty` `head` `tail` `nth` `reverse` `cons` `concat` `range` `member` `to_list` `to_set` |
| Set | `set_empty` `set_add` `set_remove` `set_contains` `set_union` `set_intersect` `set_diff` `set_elements` |
| HAMT | `hamt_empty` `hamt_from` `hamt_set` `hamt_del` `hamt_has` `hamt_get` `hamt_get_or` `hamt_path` `hamt_has_path` `hamt_keys` `hamt_values` `hamt_pairs` `hamt_merge` `hamt_diff` `hamt_intersect` |
| Tree / BST | `tree_leaf` `tree_node` `tree_add_child` `tree_value` `tree_children` `tree_size` `tree_depth` `tree_flatten` `tree_member` · `bst_empty` `bst_from` `bst_insert` `bst_member` (a BST folds in order) |
| Pair / Maybe / Either / Variant | `pair` `fst` `snd` `pair_swap` · `just` `nothing` `is_just` `is_nothing` `maybe_default` `maybe_fold` · `either_left` `either_right` `is_left` `is_right` `either_default` `either_fold` · `variant` `variant_tag` `variant_payload` `is_variant` |
| Primitives | `min` `max` `abs` `negate` `compare` |
| Containment | `subset_of sub super` — Record/HAMT key-value containment or Set membership |

```haskell
map (\x -> x * 2) [1, 2, 3] |> assert [2 4 6]
hamt_get "ab" (fold (\h w -> hamt_set w (length w) h) hamt_empty ["ab", "c"]) |> assert 2
map (set_contains 2) [(to_set [1, 2]), (to_set [3])] |> assert [true false]
maybe_default 0 (map (\x -> x + 1) (just 1)) |> assert 2
type Shape = Circle | Square
variant_payload (map (\r -> r * 2) (variant "Circle" 3)) |> assert 6
```

If a function form cannot mean exactly what its step form means, it has a different name (`graph_edges_of` the step yields formatted edge strings; `graph_edge_pairs` the function yields `[from to]` pairs).

## 22. Graphs, Inferred Edges & Stream Merge (`<+>`)

`Graph(T)` is a persistent directed **multigraph** with colored nodes and edges. Colors are plain strings (a declared constructor such as `Owns` resolves to its tag); edge weights are any value; an edge can be **inferred** (assumed, lower-trust). Higher edge cardinality is modeled with grouped binary edges or by promoting a relation to a node (reification).

```haskell
-- Ingest: fold a record stream into a graph. Node id/color are computed from the record.
def ingest = fold (\g r -> graph_add_node (hamt_get "kind" r) (hamt_path "metadata.name" r) r g) graph_empty

-- A relation is a graph→graph pass. Edge specs are 4-element lists: [color, weight, from, to].
-- Add NODES FIRST; guard edges with graph_has_node.
def rel_owns = fold (\_ g -> graph_add_edges [ ["Owns", 1, (hamt_get "name" o), (hamt_path "metadata.name" x)] | x <- (graph_payloads g), hamt_has_path "metadata.ownerReferences" x, o <- (hamt_path "metadata.ownerReferences" x) ] g) graph_empty

-- Inferred evidence is recorded separately from declared relations:
--   graph_add_inferred_edges [["Calls", 0.5, "a", "b"]] g

-- Whole-graph algorithms are functions, so a rule can run them inside a comprehension:
--   [ c | c <- (graph_scc g), (length c) > 1 ]        dependency cycles
--   graph_shortest_path "web" "pod" g                  a request trace
--   graph_topo_sort g                                  an apply order ([] when cyclic)
```

Read side: `graph_payloads` `graph_payloads_of color` `graph_in_of color id` `graph_out_of color id` `graph_has_node` `graph_nodes` `graph_order` `graph_size` `graph_edge_pairs color` `graph_reachable` `graph_shortest_path` `graph_topo_sort` `graph_scc`. The pipeline steps (`graph_nodes_of`, `graph_edges_of`, `graph_walk_from`, `graph_map_nodes`, `graph_focus`/`graph_set`/`graph_over`, `graph_union`/`graph_diff`/`graph_intersect`, …) are exercised in `tests/dsl_exhaustive_test.gatwd`.

**`(a <+> b <+> c)`** merges several sub-pipelines into one stream: each incoming node runs through every branch and the outputs are concatenated in branch order (the ArrowPlus / MonadPlus choice combinator; the zero arrow is the identity and it is associative). It is how many producers feed one downstream pipeline:

```haskell
spawn IntVal: 5 |> (math_add 1 <+> math_add 10 <+> math_add 100) |> collect_list |> assert [6 15 105]
-- the Kubernetes edition merges every crawl stream the same way:
--   spawn StrVal: default |> k8s_resources |> k8s_graph
```

A complete domain built on all of this — Kubernetes resources lifted into a typed graph with relations and compliance rules written in the DSL, over a live cluster or offline manifests — is documented in [`k8s_graph_design.md`](k8s_graph_design.md); its entry points are `gatwd-k8s` (Edition) and `gatwd-k8s-operator` (Appliance).


## 23. Arguments, Tasks and `main`

What a script *sees* of its command line is defined by the interpreter, independent of the flag syntax:

| Name | Meaning |
| :-- | :-- |
| `ARGS` | a `List` of `String`: everything after `--` (always defined; empty when there are none) |
| `ARG0` | the script path (`""` when there is no script) |
| task | any `def`, run with `ARGS` |
| `main` | the default task, run after the script has loaded when no `--task` is given |

```bash
gatwd script.gatwd -- prod --strict          # loads script.gatwd, then runs `main`
gatwd script.gatwd --task audit -- prod      # loads it, then runs `audit` instead of `main`
```

A task is one of two kinds, chosen by how the `def` is written:

- **Lambda-bodied** (`def audit = \args -> …`) — a function called with the `ARGS` list; its value is printed.
- **Pipeline-bodied** (`def audit = manifests |> k8s_graph |> …`) — a fragment. Its input stream is the
  arguments, one `String` node each, so a fragment started with `-- some/path` receives that path as
  its first node. With no arguments the stream is empty (a generator such as `emit` still produces output).

```haskell
def shout = \args -> map (\a -> a ++ "!") args
def main  = \args -> [(length args), (shout args)]
-- gatwd script.gatwd -- a "b c"      prints   ∷ [2 [a! b c!]]
```

Put invocations in `main` (or in tasks), not at the top level: top-level statements always run, which defeats
`--task`. `ARGS` is a list *value* — use it inside expressions (`length ARGS`, `nth 0 ARGS`) rather than
splicing it into step text. A task name that does not exist, a script that cannot be read, or a task that prints
an `[ERROR]` makes the process exit non-zero (`1`; a malformed command line exits `2`).

### The command line

Every Edition and Appliance gets the same command line (a cobra tree in `pkg/cli`; it only maps flags to
`Interpreter.Run`). The old forms keep working:

```bash
gatwd [script] [-- arg ...]            # run a script (then its `main`); no script → the REPL
gatwd -i script                         # run it, then a prompt with its defs loaded
gatwd --task NAME script -- arg ...     # run one def instead of `main`
gatwd --check file                      # same as `gatwd check file`
```

and these are new:

```bash
gatwd run script -- arg ...             # explicit form
gatwd eval -e '1 + 2' -e '"a" ++ "b"'   # one-liners; exit 1 if a statement errors or prints nothing
gatwd script -e 'main_helper 3'         # -e after a script: replaces `main`
gatwd check a.gatwd b.gatwd             # structural errors exit 1; typing findings are advisory
gatwd task NAME [-s script] -- arg ...  # a def from the prelude, or from `script`
gatwd prelude list | cat NAME           # the embedded prelude files
gatwd builtins                          # function forms (with parameters), defs, constants
gatwd version · gatwd completion bash|zsh|fish|powershell · gatwd --help
```

Global flags: `-D NAME=VALUE` (a string constant; `ARGS`/`ARG0` are reserved), `--prelude FILE` (repeatable, loaded
before the script), `--no-prelude` (skip the embedded preludes; their steps stay registered). Exit status: `0` ok,
`1` a command failed, `2` a malformed command line. A script literally named like a subcommand is run as `gatwd ./check`.
Go callers can capture output by setting `Interpreter.Out`; an Edition adds a command with `Edition.WithCommand`.
