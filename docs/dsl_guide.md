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
* `build_bst`: Binary Search Tree of the stream's *values*, ordered by comparison like `sort` (`3, 1, 10, 9` is in order `1 3 9 10`, not the text order); two elements that cannot be ordered (a String and an Int) are an error naming both. The same `*BST[any]` as `bst_from`.
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

**Literal tokens are typed.** The unquoted value of a `const`, of `-D NAME=VALUE` and of each `emit` token follows one rule: a
*canonical* literal is that value, anything else (`alice`, `007`, `1.50`, `kind:Pod`) is a `String`, and a quoted value
(`const ZIP = "10115"`) is always a `String`. The canonical literals are
numbers (`10`, `-5`, `1.5` — text that prints back unchanged), **Durations** (`5s`, `1h30m`, `30d`, `-2w`, `1.5h`),
**Times** (`@2026-10-13T10:00:00Z`, `@2026-10-13T11:30:00+02:00`) and **Dates** (`@2026-10-13`). A Time or Date is canonical
when it prints back unchanged (`@2026-10-13T10:00:00+00:00` and `.250Z` are not: they print `Z` and `.25Z`). A Duration has
many spellings of one value (`90m`, `1h30m`, `5400s`) and prints as its own chain (`30d` shows as `720h`), so for it
"canonical" is about the *numbers*, not the unit: each segment is written without a `+` sign, a leading zero (`007s`) or a
trailing zero in a fraction (`1.50h`), so `007s`, `1.50h`, `+5s`, `-0s`, `0.0s` stay `String`s, exactly like `007` and `1.50`.
So `const N = 5` and `const MAX_AGE = 30d` can be used inside a lambda (`map (\x -> x + N)`, `filter (\age -> age > MAX_AGE)`),
`emit 5s, 6s |> fold (\a b -> a + b) 0s` adds durations, and `-D X=5s` binds a `Duration`, while the text of a log line or
a config value never silently becomes a value; convert it explicitly with `str_to_int` / `str_to_float` (`Maybe`) or
`dur_parse` / `time_parse` / `date_parse`. A `const` whose value contains an operator character is an expression and is
evaluated (`const P = +5` is `5`, `const D = 1h + 30m` a `Duration`, `const D = @2026-10-13` a `Date`); a *quoted*
`const` value stays the text you wrote, whatever it would evaluate to (`const Q = "2+3"` is the String `2+3`, not `5`). Operator sections read the same literals
(`|> + 5s`, `|> - 7d`), and a time constant is spliced into a pipeline head or step as its literal, so the line re-lexes to the
same value. **Quotes are the one way to say "text"** for a spelling that is also a literal, in every position: `const Q = "2+3"`, `emit "5", 5` (the String `5`, then the Int), a spawn head `spawn S: Q`, an operator operand (`|> ++ "!"`; `|> + "5"` is an error, a String does not do arithmetic), and `-D`. On the command line the shell removes quotes before gatwd sees them, so `-D Q="2+3"` binds the text `2+3` anyway (it is not canonical) and `-D N=5` binds the Int `5`; to pass a *String that is spelled like a literal* keep the quotes in the value, `-D 'ZIP="10115"'` is the String `10115`, while `-D ZIP=10115` is the Int. Only a value that is *exactly one* quoted literal is unquoted: the interior is taken literally (no escapes), and `const J = "a" ++ "b"` is an expression, as before. An unquoted expression keeps being evaluated (`const N = 2+3` is `5`). A `spawn` seed is not a literal token: for a type with no registered parser it stays text (`spawn T: 5s` seeds the String `5s`; use `emit 5s` or `let d = 5s`). A String
constant used inside a lambda is spliced quoted, so `const Q = "30d"` is a `String` there (it used to read as a Duration). A spawn of an **undeclared** type keeps its value as text, so declare
the type you compute with (`type IntVal = Int`) — `x + 1` on text is an error that names both types.

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

**A key is a value.** A HAMT key may be any value, and `==` decides "the same key": equal keys are one entry and different keys never are.

| Key written | Is the same entry as | Is a different entry from |
| :-- | :-- | :-- |
| `[1, 2]` | `(pair 1 2)`, a Chunk `[1, 2]`, `[1.0, 2]` | `"[1 2]"`, `[2, 1]`, `[[1, 2]]`, `(to_set [1, 2])` |
| `3` | `3.0`, `3r1` | `"3"`, `true` |
| `0.5` | `1r2` | `"0.5"`; and `0.1` is not `1r10` (`==` says they differ) |
| `"a"` | `'a'` (D1) | `["a"]` |
| `(to_set [1, 2])` | `(to_set [2, 1])` (a Set is unordered) | `[1, 2]` |

`hamt_keys` and `hamt_pairs` return the keys as the values they were (`frequencies` over coordinate pairs gives pairs of `[x, y]` lists, so `fst p == [2, 1]` is `true` and the Game of Life compares real coordinates). A `String` key is still itself, so records from `json_parse`, `yaml_parse` and manifests are unchanged; `json_encode`/`yaml_encode` of a HAMT whose keys are not Strings is an error (object keys are Strings). A value that cannot be a key (a function, a Graph, `NaN`) is an error naming it, never a guess. The same rule is applied by the function forms (`hamt_set`, `hamt_get`, `hamt_has`, `hamt_del`, `hamt_get_or`, `hamt_from`), the step forms (`hamt_get K`, `hamt_set K V`, `hamt_del K`, a bare canonical number token is a number, a quoted one a String), `to_hamt` and `frequencies`.

*Design (for contributors).* `evaluator.EncodeKey` writes a String as itself and anything else as a NUL byte plus a tagged, length-prefixed canonical text (numbers through `functions.CanonNumber`, the representative of their `==` class; Sets and HAMT entries sorted); `evaluator.DecodeKey` is the inverse. Order is kept for lists, `1` and `1.0` meet, and two different collections can never share a key because every element is tagged and length-prefixed. **Graph ids and colors** are different: they are Strings by construction, so `graph_add_node "c" 7 g` names the node `"7"` (a number is its text) but a list or pair is an error — use `show` to name a node by one. The Set and `distinct` use the same key design (section 11).

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
* `set_elements`: Unfolds the Set into a flat stream of its elements (the values, never their text).

**A member is a value.** "The same member" is the language's `==`, exactly as for a HAMT key (section 8), never the printed form: `1` and `"1"` are two members, `1` and `1.0` one, `[1, 2]` and `[1, 2]` one (two lists, one value), `[1, 2]` and `[2, 1]` two, `[1, 2]` and the text `"[1 2]"` two, and one instant written in two zones is one member. The step operand of `set_contains`/`set_add`/`set_remove` is read as a value too: `3` is the number, `"3"` a String, `[1 2]` a list, a bare word (`golang`) its text, a let-bound name its value. `distinct` uses the same relation, so `[1, 1.0, "1"] |> unchunk |> distinct` keeps `1` and `"1"`. An element with no equality (a function, a Graph, `NaN`) cannot be a member: `to_set`, `distinct` and the `set_*` forms report an error that names it, and a `set_*` step handed something that is not a Set says so (it used to pass the value through unchanged). Step form and function form build the same Set (`to_set`, `set_add`, `set_remove`, `set_contains`, `set_union`, `set_intersect`, `set_diff`, `set_elements`, and `map`/`filter`/`flatmap`/`member`/`==`/`++` over a Set), so they can be compared with `==` across the two forms.

*Design (for contributors).* The Set is a `*collections.Setoid[any]` (`collections/setoid.go`): the elements themselves, stored in a `HAMT[string, any]` under `evaluator.ValueSetKey` (= `EncodeKey`, with a Time reduced to its UTC instant). Nothing is decoded from text, so reading a member gives back the stored value; the first of several equal members stays; a Set read from a Go carrier (`*collections.Set[string]`, `*collections.Set[any]`) is rebuilt by `evaluator.AsValueSet`. `evaluator.NewValueSet`, `ValueSetAdd` and `EmptyValueSet` are the constructors; a Go caller's own `*collections.Set[T]` still works under Go's `==` and still unites with a Set of the same type by its own method.

```haskell
spawn StrVal: Start |> emit ruby |> to_set |> set_add golang |> set_contains golang |> assert true
spawn StrVal: Start |> emit a, b |> drop 1 |> to_set |> set_remove a |> set_contains a |> assert false
spawn StrVal: Start |> emit x |> drop 1 |> to_set |> set_elements |> assert x
```

## 12. Stream Transducers (`filter`, `distinct`, `fold`)

* `filter <substring>`: Filters nodes whose string representation contains the target substring.
* `distinct`: Deduplicates consecutive and recurring elements across the stream, maintaining order. "The same element" is `==` (as for a Set member, section 11): `1` and `"1"` are two, `1` and `1.0` one.
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
- **Arithmetic** (the numeric tower, see "The numeric tower" below):
  - `+` (**numeric addition only**: `"a" + "b"` and `"1" + 2` are errors — use `++` to concatenate and `str_to_int` to parse; D2). The same holds for `-`, `*`, `/`, `//`, `%`, `**` and `<`, `<=`, `>`, `>=`: a `String` is never read as a number (`"5" - 1`, `"5" < 10` are errors), and two `String`s **order as text** (`"10" < "9"` is `true`)
  - `-` (Subtraction; **unary minus** in operand position: `- 5`, `-x`, `-(a + b)`)
  - `*` (Multiplication)
  - `/` (True division: **exact when it divides, otherwise a `Float`**, e.g. `8 / 2 == 4`, `7 / 2 == 3.5`; exact with a `Ratio` operand: `1r3 / 2 == 1r6`)
  - `//` (**Floor** division: `7 // 2 == 3`, `-7 // 2 == -4`; the result is an `Int`)
  - `%` (**Modulo**, taking the sign of the divisor: `-7 % 3 == 2`, `7 % -3 == -2`; always `a == b * (a // b) + a % b`)
  - `**` (Exponentiation, right-associative, tighter than unary minus: `2 ** 3 ** 2 == 512`, `-2 ** 2 == -4`)
- **Monoidal Concatenation**:
  - `++` (String and collection concatenation, e.g. `"Hello " ++ "Gatwd"`)
- **Relational & Equality**:
  - `==`, `!=`, `<`, `<=`, `>`, `>=` (Structural equality and scalar comparisons; numbers compare **exactly**, never through a double, and `"5" == 5` is `false`)
- **Boolean Logic**:
  - `&&`, `||`, and the prefix **`!`** (a synonym of `not`: `!true`, `!(a == b)`; `!` is a prefix operator only when the next character is not `=`, so `a != b` is unchanged and `! =` is an error)

Precedence, loosest to tightest: `||` · `&&` · `==` `!=` · `<` `<=` `>` `>=` · `+` `-` `++` · `*` `/` `//` `%` · prefix `-` `!` · `**`; function application binds tighter than all of them.

**Minus, sign and application.** A `-` in operand position (after an operator, `(`, `[`, `,`, `->`, or at the start) is the prefix operator, and it is folded into a numeric literal, so `-9223372036854775808` is the smallest int64. After an operand, `a - 3`, `a-3` and `a- 3` are subtractions, but **`f -3` (a space before the `-`, none after) is a negative *argument***: `abs -3` is `3`, `min -1 3` is `-1`, `divmod -7 2` is `[-4, 1]`. The price is that `n -1` with `n` bound to a number is an application error (with a hint to write `n - 1`). In a plain list `[3 -1]` is two elements, so a printed list of negatives reads back as written. A literal cannot be applied, so `5 -3` is still `2`.

### The numeric tower (hardening W3, D3, D11–D13)

The language has four number kinds: **`Int`** (unbounded, exact), **`Ratio`** (an exact fraction), **`Float`** (an IEEE double) and **`Complex`**, ordered `Int ⊂ Ratio ⊂ Float(approx) ⊂ Complex`. An operation on two numbers is carried out in the *join* of their kinds: `Int + Ratio` is a `Ratio`, `Ratio + Float` a `Float`, anything with a `Complex` a `Complex`. `BigInt` and `Uint` remain accepted type names for `Int` (so `type Balance = BigInt` keeps working).

| | `Int` | `Ratio` | `Float` | `Complex` |
| :-- | :-- | :-- | :-- | :-- |
| `+` `-` `*` | `Int` | `Ratio` (an `Int` when whole) | `Float` | `Complex` |
| `/` | `Int` if it divides, else `Float` (**D11**) | `Ratio` (exact; an `Int` when whole) | `Float` | `Complex` |
| `//` | `Int` | `Int` (floor) | `Int` (floor of the quotient) | error |
| `%` | `Int` | `Ratio` (`a - b*(a//b)`) | `Float` | error |
| `**` | `Int` for a non-negative `Int` exponent, `Float` for a negative one | exact for an `Int` exponent, `Float` otherwise | `Float` | `Complex` |
| `<` `<=` `>` `>=` `==` | exact | exact | exact against `Int`/`Ratio` (**D12**) | `==` only |

(a row names the result when both operands are of the column's kind or lower; mixed pairs follow the join). Division, floor division and modulo by zero are errors in every kind; an `Int` too large for a double never silently becomes `+Inf` — mixing it with a `Float` is an error.

**`Int` is logical.** Behind the one type the runtime keeps an int64 fast path and a big integer; every integer result is *canonical* (an int when it fits, a big integer only beyond), so `2 ** 100 - 2 ** 100` is the plain `0`, `(2 ** 64) / (2 ** 64)` is `1`, and big integers print as plain digits (no `+`). `**` refuses a result over four million bits (`2 ** 99999999` is an error), and an exact power with an exponent beyond int64 is an error.

**Exact comparison.** `==`, `<`, … never round an `Int` or `Ratio` to a double: `9007199254740993 > 9007199254740992.0` is `true`, `1r2 == 0.5` is `true`, `1r10 == 0.1` is **`false`** (the double nearest to 0.1 is not one tenth; ask `to_ratio 0.1 == 1r10`). A `NaN` is unordered: every ordering with it is `false`, `NaN == NaN` is `false`.

**`Ratio` literals and functions.** `3r4` (numerator `r` denominator) and the exact decimal `0.75r` are `Ratio` values, always in lowest terms and with a positive denominator; a whole fraction **is an `Int`** (`4r2` is `2`, `1r2 + 1r2` is `1`). `1r0` and a bare `5r` are lex errors. A `Ratio` prints as the literal (`3r4`), so printed values read back. `%` stays *modulo* (**D13**) and `/` on two `Int`s keeps its meaning (**D11**): exactness is opt-in through `3r4` or `ratio n d`.

| function (data-last, curried, also a step) | meaning |
| :-- | :-- |
| `ratio n d` | the exact quotient of two `Int`/`Ratio` (`ratio 1 3` is `1r3`; `d = 0` is an error; a `Float` argument is an error) |
| `to_ratio x` | the fraction of the **shortest decimal** that round-trips: `to_ratio 0.1` is `1r10`, `to_ratio 0.75` is `3r4` |
| `to_ratio_exact x` | the exact binary value of the double: `to_ratio_exact 0.1` is `3602879701896397r36028797018963968` |
| `to_ratio_approx x max_den` | the best fraction with denominator ≤ `max_den` (continued fractions): `to_ratio_approx 3.14159265 100` is `311r99` |
| `str_to_ratio s` | `Maybe Ratio` from `"3/4"`, `"3r4"`, `"-0.75"`, `"1.5e-3"`; malformed text is `Nothing` |
| `numer r`, `denom r`, `ratio_parts r` | the parts (`denom` of an `Int` is `1`); `ratio_mixed 7r2` is `[3, 1r2]` |
| `ratio_cf r`, `ratio_from_cf xs`, `ratio_mediant a b` | continued-fraction coefficients, their inverse, the Stern–Brocot mediant |

`NaN` and `±Inf` have no `Ratio` (an error). `to_float r` is the correctly rounded double (a magnitude beyond the double range is an error; a fraction below the smallest double is `0.0`).

**Numeric functions** (curried, data-last, usable as steps; every partial one has an error case): `pow base exp` (the function form of `**`) · `sqrt x` (a `Float`; negative is an error — use a `Complex`) · `isqrt n` (exact integer square root) · `floor` `ceil` `trunc` `round` (half away from zero) `round_even` (a number to an `Int`; `NaN`/`Inf` are errors) · `sign x` · `abs x` · `negate x` · `gcd a b` · `lcm a b` · `divmod a b` (floor, a pair) · `quot a b` / `rem a b` (truncated) · `to_int x` (**exact or an error**: `to_int 3.5` names `floor`/`round`) · `to_float x` · `is_nan x` · `is_inf x`.

**Strings are not numbers** (the end of D2). `"5" == 5` is `false`; `ToFloat64`-style helpers no longer parse text, so `nth "1" xs`, `range "0" "2"`, `min "5" 3` and `compare "5" 3` are errors that point to `str_to_int`. **A String is not the printed form of a collection either:** `"[2 1]" == [2, 1]` is `false` (it was `true` until HAMT keys carried values, below). A `String` equals only a `String`, and a one-character `String` a `Char` (D1).

**Constraint names.** In annotations `Num` (any number), `Integral` (`Int`), `Fractional` (`Ratio`, `Float`, `Complex`), `Comparable` (everything but `Complex` and functions) and `Appendable` (`String`, `Bytes`, `List`, `Chunk`, `Set`, `HAMT`, `Tree`, `Graph`, `Pair`, `Maybe`) state intent without a runtime wrapper: `def sum2 : Num -> Num -> Num = \a b -> a + b`. `gatwd --check` reports only definite contradictions (`sum2 "a" 1`, `x - "a"` under `Num`); because the tower converts at runtime a concrete number is consistent with every *numeric* constraint, so `Integral` and `Fractional` contradict only non-numbers and each other.

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

### `Duration`, `Time` and `Date` (hardening §5.6, D14–D17)

Three distinct types answer three questions: **how long** (`Duration`), **which instant** (`Time`), **which calendar day**
(`Date`). They are separate because mixing them silently is the classic bug (`+ 24h` is not "tomorrow" across a
daylight-saving change), so no operator converts between them; the conversions are named functions that take the zone.
The rule behind the operators is the one for **points and vectors**: a `Duration` is a vector, `Time` and `Date` are
points. Point − point is a vector, point ± vector is a point, vector ± vector is a vector, and point + point is an error.

**Literals.**

| Form | Value |
| :-- | :-- |
| `5s` `250ms` `1h30m` `2d` `1w` `1.5h` `-5s` | `Duration`. Units `ns us µs ms s m h d w`; a day is **exactly 24 h** and a week 7 days (fixed lengths, not calendar days). A number *immediately* followed by a unit is one token: `5s` is a literal, `5 s` is still an application. Segments chain (`1h30m15s`, any order, each unit once); only the first segment may be fractional. |
| `@2026-10-13T09:30:00Z` `@2026-10-13T11:30:00+02:00` `@2026-10-13T09:30:00.25Z` | `Time` (RFC 3339). **A missing zone is an error**: local time is never implicit. |
| `@2026-10-13` | `Date` (ISO 8601, years 1 to 9999) |

A malformed literal is a lex error that names the problem (`5q` unknown unit, `1h1h` unit used twice, `1h30` no unit,
`0.1ns` not whole nanoseconds, `@2026-02-30` day out of range, `@2026-10-13T09:30:00` no zone). They print as the
canonical text, which lexes back: a `Duration` as `1h30m`, `1s500ms`, `-5s`, `0s` (the largest unit printed is the hour, so
`30d` prints `720h`), a `Time` as RFC 3339 (`2026-10-13T09:30:00Z`), a `Date` as `2026-10-13`.

**Operators** (anything not listed is an error naming both types; overflow is an error, never a wrap):

| Operation | Result |
| :-- | :-- |
| `Duration ± Duration` · unary `-` · `abs` | `Duration` |
| `Duration * n` · `n * Duration` · `Duration / n` (`n` an `Int`, `Ratio` or `Float`) | `Duration`, truncated toward zero to whole nanoseconds; `/ 0` is an error |
| `Duration / Duration` | the **exact `Ratio`**, an `Int` when whole (`90m / 1h` is `3r2`, `2h / 1h` is `2`) |
| `Duration // Duration` · `Duration % Duration` | floor quotient (`Int`) · remainder (`Duration`, the divisor's sign) |
| `Time + Duration` · `Time - Duration` | `Time` (same display zone) |
| `Time - Time` | `Duration` (an error beyond about ±292 years) |
| `Date + Duration` · `Date - Duration` | `Date`, **only for a whole number of days** (`2d`, `-1w`, `0s`); `Date + 90m` is an error that says so (**D17**) |
| `Date - Date` | `Duration` (a multiple of `1d`) |
| `== != < <= > >=`, `min` `max` `compare` `sort` | within one type only; **`Date` against `Time` is an error** (no implicit zone) |

`5s + 1`, `5s * 5s`, `Time + Time`, `Date + Date` and `Time * 2` are all errors. A `Duration` is never an `Int`
(`5s == 5` is `false`). Calendar months and years are not fixed lengths and go through functions, not operators.

**`Time` is an instant that remembers a display zone** (**D16**). `time_in "Europe/Berlin" t` is the *same instant* shown
in another zone; `==` and the orderings compare instants, the calendar fields (`time_hour`, `date_of_time`) read the zone,
and a parsed offset sets it. The default zone is UTC; the host's zone is never used (`time_in "Local" t` is an error). There
is **no `DateTime` type**: a date and time of day without a zone is not a point on the timeline (deferred until a real
need appears). The zone database is embedded in the binary, so zone names never depend on the host.

| Function (data-last, curried; usable as steps) | Meaning |
| :-- | :-- |
| `dur_ns d` · `dur_ms d` · `to_seconds d` | `Int` nanoseconds · `Int` milliseconds (toward zero) · `Float` seconds |
| `dur_from_seconds x` · `dur_parse s` | `Duration` from an `Int`/`Ratio`/`Float` of seconds (toward zero) · `Maybe Duration` (`dur_parse "1h30m"`) |
| `dur_truncate unit d` · `dur_round unit d` | to a multiple of `unit` (a positive `Duration`, `1m`): toward zero · halves away from zero |
| `time_parse s` · `time_format layout t` | `Maybe Time` (RFC 3339, zone required) · text, see the layout table below |
| `time_unix t` · `time_from_unix n` | seconds since 1970 (`Int`) · `Time` in UTC |
| `time_year` `time_month` `time_day` `time_hour` `time_minute` `time_second` `time_nano` `time_weekday` | `Int` fields **in the display zone**; the weekday is ISO, `1` Monday … `7` Sunday |
| `time_in zone t` | the same instant in an IANA zone; an unknown or empty zone is an error |
| `time_truncate d t` · `time_round d t` | to a multiple of `d` counted from `0001-01-01T00:00:00Z`, so `1d` is **UTC** midnight whatever the display zone |
| `date_of y m d` · `date_parse s` · `date_format layout d` | `Date` (an invalid day such as 2026-02-30 is an error) · `Maybe Date` · text |
| `date_year` `date_month` `date_day` `date_weekday` `date_days_in_month` · `is_leap_year y` | `Int` / `Bool` |
| `date_add_days n d` · `date_add_months n d` · `date_add_years n d` | `Date`; months and years **clamp to the end of the month** (31 Jan + 1 month is 28/29 Feb) and never skip a month |
| `date_of_time zone t` · `time_of_date zone d` | the calendar day of an instant in a zone · the **first instant of a day** in a zone (midnight; a midnight swallowed by a daylight-saving gap is the first valid instant after it, a repeated midnight is the earlier) |

The parsers answer `Nothing` for malformed text (never an error); impossible requests are errors. **Layouts** use a
strftime subset, never Go's reference-time layouts: `%Y %m %d %H %M %S` (zero padded), `%z` (`+0200`), `%Z` (`CEST`, or
`+02:00` for a bare offset), `%%`, and the names `rfc3339`, `date` (`%Y-%m-%d`), `clock` (`%H:%M:%S`). An unknown
directive is an error; `date_format` accepts only `%Y %m %d %%` (a `Date` has no time of day or zone).

```
-- Europe/Berlin, the clocks go forward on 2026-03-29: that day lasts 23 hours, so "+ 24h" is not "tomorrow"
(time_of_date "Europe/Berlin" @2026-03-30) - (time_of_date "Europe/Berlin" @2026-03-29)   -- 23h
@2026-03-29 |> time_of_date "Europe/Berlin" |> map (\t -> time_format "%Y-%m-%d %H:%M" (t + 24h))   -- 2026-03-30 01:00
date_add_days 1 @2026-03-29                                                              -- 2026-03-30
```

**`now`** reads the clock, which is a side effect, so it is a stream-source **step** backed by an adapter
(`adapters.ClockAdapter`), exactly like `scan_file` reads a file: `spawn S: x |> now` emits one `Time` (UTC), and a bare
`now` at the head of a pipeline is the same reading (a `let now = …` binding shadows it). It reads `Interpreter.Clock`
(`func() time.Time`, `nil` means `time.Now`); **tests and Editions inject a fixed clock** (`interp.Clock = func() …`), so
nothing sleeps and nothing depends on the wall clock. The clock is read once at the edge and passed on as a value, which
keeps rules pure: `def expiring = \now exp -> (exp - now) < 14d`, then `now |> map (\t -> expiring t cert_expiry)`; resource
age is `(now - created) > 30d`. `now` is not a function form (inside a lambda, take the reading as a parameter). A
`sleep` step is deliberately not provided (blocking steps need the execution policy of W7 first).

**In collections.** Structural equality reaches inside lists, pairs and `Maybe`s. A `Time` is its instant, so one instant
shown in two zones is **one** `Set` member and **one** `HAMT` key (a `Set` stores the canonical UTC form; use
`time_in` to show it again). `sort`, `min`, `max`, `compare` order by value. **Encoding:** `json_encode`/`yaml_encode` write
a `Time` as its RFC 3339 `String`, a `Date` as its ISO `String`, a `Duration` as its literal text (`"1h30m"`); decoding leaves
text as text and `time_parse`/`date_parse`/`dur_parse` revive it on request (the same stance as YAML dates).
**Typed tokens:** `const X = 30d`, `emit 5s` and `-D X=5s` bind a real `Duration` (and `@2026-10-13…` a `Time` or `Date`) by the
same rule as numbers (§6). The checker knows all three types
(`Comparable`, never `Num`; `Duration / Duration` is typed `Ratio`; `now` is a `Time`).

### Booleans and `Ordering` (hardening W2)

`not b`, `xor a b`, `and a b`, `or a b` are curried function forms (also steps: `true |> not`) for folds and lambdas —
`[true, false] |> fold and true`. They are **strict** (both arguments are evaluated) and accept only `Bool`: `not 1` is an
error, a number is never "truthy". `&&` and `||` remain the short-circuiting infix operators. There is no prefix `!`
(it would collide with `!=`); write `not`.

`compare a b` answers with the **`Ordering`** sum type `LT | EQ | GT` (declared in every parser, so it is an ordinary sum
type): `compare 1 2 == LT`, `compare x y |> match { LT -> … | EQ -> … | GT -> … }` (`gatwd --check` reports a missing branch).
`ordering_to_int o` gives `-1`/`0`/`1`, `ordering_from_int n` the Ordering of a sign, `ordering_reverse o` flips it, and
`ordering_then a b` is `a` unless it is `EQ`, then `b` — the lexicographic combination
(`ordering_then (compare x1 y1) (compare x2 y2)`). `min`, `max` and the infix comparisons are unchanged. (`compare` used to
return `-1`/`0`/`1`; wrap it in `ordering_to_int` for that.)

### Bytes, encodings and hashes (hardening W2, tier 2)

`Bytes` is immutable binary data (`functions.Bytes`); it prints as `0x`-hex (`0x6869`) and is a collection row of `Int`s
0..255 — `length`, `head`, `nth`, `member`, `fold`, `to_list`, `map`/`flatmap` (a list), and the closed operations
`filter`, `reject`, `reverse`, `tail`, `cons` (Bytes again). `++` concatenates two `Bytes`, `==` compares them. A `String` is
**not** silently `Bytes`, but wherever a function wants `Bytes` (the encoders and hashes) a `String` is accepted as its UTF-8
bytes; the way back is explicit and partial.

| Group | Functions |
| :-- | :-- |
| text ↔ bytes | `utf8_encode s` → `Bytes` · `utf8_decode b` → `Maybe String` (`Nothing` for invalid UTF-8) |
| build / take apart | `bytes_from_list xs` (Ints 0..255, else an error naming the element) · `bytes_to_list b` · `bytes_length b` · `bytes_slice from to b` (half-open) |
| encodings | `base64_encode`/`base64_decode` (standard, padded) · `base64url_encode`/`base64url_decode` · `hex_encode`/`hex_decode` · `url_encode s`/`url_decode s` (query escaping) — encoders return a `String`, **decoders return a `Maybe`** |
| hashes | `sha256 b` · `md5 b` — lowercase hex `String` (md5 is a fingerprint, not a security measure) |

### Regular expressions (hardening W2, tier 2)

Go **RE2** syntax: the engine runs in time linear in the input, so a pattern from an untrusted source cannot cause
catastrophic backtracking — and so there are no backreferences or lookaround (such a pattern is a compile error). Matching
is unanchored; write `\d` as `"\\d"` in a String literal. An invalid pattern is an error naming it; *no match* is a `Maybe` or an
empty list, never an error.

| Function | Result |
| :-- | :-- |
| `re_match pat s` | `Bool` |
| `re_find pat s` · `re_find_all pat s` | `Maybe String` (leftmost) · `List` of `String` |
| `re_groups pat s` | `Maybe` of the whole match followed by each group (a group that took no part is `""`) |
| `re_replace pat repl s` | `String`; `repl` expands `$1` / `${name}`, `$$` is a literal `$` |
| `re_split pat s` · `re_quote s` | `List` of `String` · the pattern that matches `s` literally |

### JSON (hardening W2, tier 2)

`json_parse s` reads JSON text into the **same nested shape a resource has**: an object is a `HAMT` (so `hamt_path "a.b" d`
navigates it), an array a `List`, a number an `Int` / `BigInt` (an integer) or a `Float`, a string a `String`, `true`/`false`
a `Bool`, and `null` the null value (`is_null x`). The result is a **`Maybe`**: malformed text, or anything after the one value,
is `Nothing`, never an error (`json_parse "null"` is `Just` null). `json_encode v` is the reverse and deterministic (object keys
sorted, no whitespace, no HTML escaping); `json_pretty v` indents by two spaces. A value JSON cannot hold — a `Pair`, a closure,
a `Set` (no defined order), `Bytes` (use `base64_encode`), `NaN`/`Inf` — is an error that names the path to it (`$.a[1]`).

### YAML (hardening W2, tier 2)

`yaml_parse s` reads YAML into the same shape as `json_parse` (mapping `HAMT`, sequence `List`, `Int`/`BigInt`, `Float`,
`String`, `Bool`, null) inside a **`Maybe`**; `yaml_parse_all s` reads a `---` stream into a `Maybe` of a `List` of documents
(an empty stream is `[]`). `yaml_encode v` writes block-style YAML with sorted keys and no final newline, quoting a `String`
only where it would otherwise read back as something else (`"5"`, `"yes"`); `yaml_encode_all docs` joins documents with `---`.
It fails on the same values `json_encode` does. **Nothing is altered silently:** an unquoted timestamp (`2026-10-13`) and a
binary scalar stay the text they were written as (parse them later with the time and `base64_decode` functions), anchors and
aliases are expanded and `<<` merge keys applied (an explicit key wins), a non-string scalar key (`1: a`) becomes the String
`"1"`. **Malformed or unsupported text is `Nothing`:** a syntax error, a duplicate key, a complex (non-scalar) key, a tag
outside the core schema (`!Ref x`), a multi-document stream given to `yaml_parse`, and a document that expands past a million
nodes (an alias bomb). The library (`go.yaml.in/yaml/v3`) is on the Clean Core allowlist for `pkg/cli` only (decision
2026-10-13; the visualizer is expected to use it later).

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
argument is an error.

**String literals have escapes** (D18): `\n \t \r \0 \\ \' \" \xHH \uHHHH`, the same table as `Char` literals; an unknown escape is an
error. So `str_lines "a\nb"`, `str_split "\n" s` and `"say \"hi\""` work, a literal backslash is `"\\"`, and `show` emits text that
lexes back to the same `String`. (This changed `"\n"` from two characters to one; no committed script was affected.)
A let-bound `String` or a `String` constant (`const`, `-D`) is spliced into a pipeline head as its quoted literal, so `let s = "a b"` then `length s` is `3`, and `-D NAME=bob` then `str_upper NAME` is `BOB` (both used to print nothing). Statement assembly ignores `(`, `[`, `{` inside string and `Char` literals, so a pattern such as `"("` does not swallow the next line.
`char_to_str '\n'`.

### Automatic Overflow Promotion to `BigInt` & `BigFloat`
Gatwd protects numerical pipelines against arithmetic overflow across all numeric types (the integer rules are the logical `Int` of "The numeric tower" above):
- **Integers (`Int`, `Uint`, `int64`, `uint64`)**: Operations exceeding `math.MaxInt64` / `math.MinInt64` promote to arbitrary-precision `*big.Int`, and a result that fits int64 again **demotes**; unsigned values take part as ordinary integers, so unsigned subtraction underflow is simply a negative `Int`.
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
  The override is **layered on top of** the embedded rules by default (`cli.WithOverrideMode(cli.OverrideLayer)`): it names only what it changes and may call the embedded defs; `cli.OverrideReplace` makes it the whole rule set. A broken override is a `*cli.LoadError` (from `NewAppliance`, or from `Reload`, which then keeps the last good rules); `app.Report()` lists the defs it overrode or added. See §23 for the Appliance command line.

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

Besides concrete types, annotations accept the **constraint names** `Num`, `Integral`, `Fractional`, `Comparable` and `Appendable` (see §15, "The numeric tower"): `def sum2 : Num -> Num -> Num = \a b -> a + b`; a constraint takes no type arguments.

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
| HAMT | `hamt_empty` `hamt_from` `hamt_set` `hamt_del` `hamt_has` `hamt_get` `hamt_get_or` `hamt_path` `hamt_has_path` `hamt_keys` `hamt_values` `hamt_pairs` `hamt_merge` `hamt_diff` `hamt_intersect` `with_source` `source_of` |
| Tree / BST | `tree_leaf` `tree_node` `tree_add_child` `tree_value` `tree_children` `tree_size` `tree_depth` `tree_flatten` `tree_member` · `bst_empty` `bst_from` `bst_insert` `bst_member` (a BST holds values ordered by comparison, so `10` follows `9`, and folds in order; `1` and `1.0` are one member; no `map`, which could break the order) |
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
def ingest = fold (\g r -> graph_insert_node (hamt_get "kind" r) (hamt_get "kind" r ++ "/" ++ hamt_path "metadata.name" r) r g) graph_empty

-- A relation is a graph→graph pass. Edge specs are 4-element lists: [color, weight, from, to].
-- Add NODES FIRST; guard edges with graph_has_node.
def rel_owns = fold (\_ g -> graph_add_edges [ ["Owns", 1, (hamt_get "kind" o ++ "/" ++ hamt_get "name" o), (hamt_get "kind" x ++ "/" ++ hamt_path "metadata.name" x)] | x <- (graph_payloads g), hamt_has_path "metadata.ownerReferences" x, o <- (hamt_path "metadata.ownerReferences" x) ] g) graph_empty

-- Inferred evidence is recorded separately from declared relations:
--   graph_add_inferred_edges [["Calls", 0.5, "a", "b"]] g

-- Whole-graph algorithms are functions, so a rule can run them inside a comprehension:
--   [ c | c <- (graph_scc g), (length c) > 1 ]        dependency cycles
--   graph_shortest_path "Ingress/default/web" "Pod/default/web-1" g   a request trace
--   graph_topo_sort g                                  an apply order ([] when cyclic)
```

Read side: `graph_payloads` `graph_payloads_of color` `graph_in_of color id` `graph_out_of color id` `graph_has_node` `graph_nodes` `graph_order` `graph_size` `graph_edge_pairs color` `graph_reachable` `graph_shortest_path` `graph_topo_sort` `graph_scc`. The pipeline steps (`graph_nodes_of`, `graph_edges_of`, `graph_walk_from`, `graph_map_nodes`, `graph_focus`/`graph_set`/`graph_over`, `graph_union`/`graph_diff`/`graph_intersect`, …) are exercised in `tests/dsl_exhaustive_test.gatwd`.

**Node identity.** A node is identified by its id string alone, so two different things given the same id would become one node. The convention is therefore that **an id is qualified**: `Kind/scope/name` (the Kubernetes edition: `Pod/default/web`, `Service/default/web`, `Role/_/admin`, with `_` for "no namespace"), so equal local names under different kinds or scopes are different nodes. `graph_add_node` keeps its upsert meaning (a later declaration replaces the earlier one); **`graph_insert_node color id [payload] g`** (function form; step form `graph_insert_node COLOR ID`) is its strict twin and returns an `[ERROR] duplicate node id "…": declared as … and again as …` — naming the id and both colors and payloads (structured payloads print as JSON) — instead of overwriting. When a payload remembers where it was read (**`with_source origin document h`** attaches an origin to a HAMT *beside* its contents, like Clojure's `with-meta`; **`source_of v`** reads it back as a `Maybe String`), the error says so too: `declared as Pod in a.yaml (…) and again as Pod in b.yaml, document 2 (…)`. The source is invisible to `==`, `hamt_*`, `json_encode` and printing, survives `hamt_set`/`hamt_del`, and an adapter that has no file (a live crawl) simply attaches none; the manifest reader attaches the file and, for multi-document files, the document's position. Use it for ingest. Two related rules: an edge whose far endpoint is not a declared node is stored (`graph_size`, `graph_edges` count it) but **not followed** (`graph_in_of`, `graph_out_of`, `graph_out`/`graph_in`, walks, `graph_shortest_path`), so adjacency agrees with `graph_nodes`; and `graph_add_edges` / `graph_add_inferred_edges` add an identical spec (same color, weight, endpoints) **once**, because a derivation pass can find one fact twice (the single-edge step `graph_add_edge` stays a multigraph insert). `collect_list` no longer drops an `[ERROR]` node from the stream it collects: it yields the errors instead of a list.

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
gatwd prelude list | cat NAME           # the embedded preludes: each entry and the .gatwd files beside it (its imports)
gatwd builtins                          # function forms (with parameters), defs, constants
gatwd version · gatwd completion bash|zsh|fish|powershell · gatwd --help
```

Global flags: `-D NAME=VALUE` (a constant: a canonical number such as `10` or `1.5`, Duration such as `5s` or `30d`, Time or Date such as `@2026-10-13` is that value, anything else — `bob`, `007`, `007s` — the String; `ARGS`/`ARG0` are reserved), `--prelude FILE` (repeatable, loaded
before the script), `--no-prelude` (skip the embedded preludes; their steps stay registered). Exit status: `0` ok,
`1` a command failed, `2` a malformed command line (an Appliance adds `3` for violations, see below). A script literally named like a subcommand is run as `gatwd ./check`.
Go callers can capture output by setting `Interpreter.Out`; an Edition adds a command with `Edition.WithCommand`.

### Execution policy, timeouts and output limits

A script can start OS commands (a word that is not a step runs as a program), touch files, `import` files and open
sockets. That is what a person at a terminal wants and what a mounted ConfigMap must not get, so the interpreter
checks a **policy** — a set of capabilities — before such a step runs (decision D8):

| Capability | Steps it gates |
| :-- | :-- |
| `exec` | an unknown step word, `memoize CMD`, `zip CMD` |
| `fs` | `scan_file`, `scan_words`, `walk_fs`, `write_file`, `append_file`, `import "file"` (an embedded import is not file access) |
| `net` | `dial_tcp`, `listen_tcp` |
| `env` | nothing yet: no builtin reads the environment; an Edition step takes part through `Interpreter.RequireCapability` |
| `cluster-apply` | a **remote write** an Edition wires behind `cli.GatedStep(cli.CapClusterApply, …)`: the Kubernetes edition's `k8s_apply` (core has no such step) |
| `cluster-rollout` | the same for operating on running workloads: `k8s_rollout_restart/undo/pause/resume`, `k8s_scale` (reading a rollout's status needs no capability) |

A denied step yields `[ERROR] step 'printf' requires the exec capability (policy denies it)` — an error value on the
same channel as every other failure, never an empty result. Steps an Edition wires itself (`RegisterStep`,
`AdapterStep`: the cluster crawl, `manifests`) are the author's decision and stay allowed.

**Two capabilities even the trusted script does not get.** `cluster-apply` and `cluster-rollout` name *writes to a
remote system of record*: they can hurt a shared production system, often by a typo, and a script is run by habit. So
`cli.TrustedPolicy()` grants `exec`, `fs`, `net` and `env` but **not** these two, `cli.DenyAll()` grants none, and they
need an explicit opt-in even for the user's own script: an Edition declares the flag with
`Edition.WithCapabilityFlag("allow-cluster-apply", cli.CapClusterApply, usage)` (`gatwd-k8s --allow-cluster-apply`), an
Appliance has `--allow-cluster-apply` and `--allow-cluster-rollout` among its generic flags, and Go code uses
`Policy.Grant(c)`. An Edition gates a step with `cli.GatedStep(cap, handler)`; the denial is the same
`[ERROR] step 'k8s_apply' requires the cluster-apply capability (policy denies it)` error value and the same structured
`capability_denied` diagnostic (`"capability":"cluster-apply"`). The capability decides *whether* a script may write; the
step is still a server dry run until its own switch says otherwise (the Kubernetes how-to, §3.5).

**Who is trusted.** The script you name on the command line, `-e`, `--prelude` files and the REPL run under
`cli.TrustedPolicy()` (everything allowed; `NewInterpreter` starts there). Anything that did not come from the
command line starts from `cli.DenyAll()`: an Appliance's rules and override (`cli.WithPolicy`, the `--allow-*` flags),
and a script given as text to `Run`. Nothing inside a script can widen the policy — it is set in Go and no statement
names it.

**`Run` and its bounds.** `RunOptions` carries them (they are installed for the call and restored after it):

| Field | Meaning |
| :-- | :-- |
| `Source` | a script as text (a ConfigMap value, an API body) instead of `Script`'s path; with no `Policy` it runs under `DenyAll` |
| `Policy` | a `*cli.Policy` for this run (nil: the interpreter's own) |
| `Context` | a deadline or cancel: every pipeline step stops at the next node, the loop over top-level statements stops, an OS command a step started is killed; `Run` returns an error wrapping the cause (`errors.Is(err, context.DeadlineExceeded)`). Cooperative: one non-terminating evaluation of a single element cannot be interrupted |
| `MaxOutput` | bytes the run may print and return; past it the output is dropped, the run is cancelled and `Run` returns an error wrapping `cli.ErrOutputLimit` |

```go
p := cli.Policy{}                                  // may only compute
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
res, err := interp.Run(cli.RunOptions{Source: body, Policy: &p, Context: ctx, MaxOutput: 1 << 20})
```

**Failures have a shape.** `cli.FailureOf(err)` classifies an error value — `capability_denied` (with `Step` and
`Capability`), `unknown_step`, `stopped`, `timeout` and `output_limit` (with `Limit`), `error` — and the Appliance's `--format json` report carries the same
objects in a `diagnostics` array (below). The free-text `[ERROR]` line is unchanged.

### The Appliance command line

An Appliance binary (for example `gatwd-k8s-operator`) gets its own tree from `cli.NewApplianceTree`, because it is
single-purpose: it evaluates one pipeline against its rules, once or on a schedule, and does not run scripts.

```bash
NAME check [flags]                 # evaluate the rules once
NAME run --interval 5m [flags]     # evaluate every 5m, re-reading the override before each check
NAME rules cat                     # print the embedded rules (copy, edit, mount as the override)
NAME version · NAME completion bash|zsh|fish|powershell
NAME [flags]                       # legacy form: loops when --interval is given, else checks once
```

| Flag | Meaning |
| :-- | :-- |
| `--rules FILE` | override rules (a mounted ConfigMap); missing or empty means none |
| `--rules-mode layer\|replace` | `layer` (default, decision D5): the override is loaded on top of the embedded rules and may call them; `replace`: the override is the whole rule set |
| `--interval DUR` | `run`/legacy form: wait between checks; the override is re-read each time |
| `--format text\|json` | `text` (default): one result per line; `json`: one JSON object per check (JSON Lines) |
| `--fail-on-violation` | `check`: exit `3` when the result is non-empty (opt-in; off by default; rejected for `run`, which never exits on violations) |
| `--allow-exec` `--allow-fs` `--allow-net` `--allow-env` `--allow-cluster-apply` `--allow-cluster-rollout` | grant the rules (the override above all) a capability; all denied by default, see *Execution policy* below |
| `--timeout DUR` | stop a check that runs longer than `DUR`: it fails with `[ERROR] check timed out after DUR` (exit `1`); every `run` round gets a fresh limit; it also bounds each load and reload of the rules (below); `0` (default) is no limit, a negative value is a usage error |
| `--max-output BYTES` | fail a check that produces more than `BYTES` of text report (`[ERROR] check output exceeded the limit of N bytes`, exit `1`; never silently truncated); `0` (default) is no limit |

In Go the limits are `cli.WithTimeout(d)` and `cli.WithMaxOutput(n)` (or `ApplianceSettings.Options()`); each `Appliance.Execute`
starts with a fresh deadline and budget, a check over a limit yields one `*cli.TimeoutError` / `*cli.OutputLimitError` and none of its
results, and `Appliance.ExecuteContext(ctx, line)` adds the caller's own context.

The same timeout bounds **loading**: `NewAppliance`, `Reload` and `ReloadContext(ctx)` evaluate the top-level statements of the
embedded rules and the override, and one that runs over fails with a `*cli.LoadError` whose `Err` is a `*cli.TimeoutError`
(`loading the rules timed out after 30s`, `errors.Is(err, context.DeadlineExceeded)`): `NewAppliance` returns it, and a failed
`Reload` leaves the last good rules in place. `cli.WithLoadContext(ctx)` (or `ApplianceSettings.Context`) bounds a load by a
context of the caller's, with a `*cli.CancelError` cause.

**Stop requests.** SIGINT and SIGTERM stop `check` and `run`: the context of the check (or of a load) in progress is cancelled, so
every pipeline step stops at its next node and a started OS command is killed. `run` prints one line on stderr (`NAME: received
terminated, stopping`) and exits `0` — a requested stop is a daemon's normal end; `check` exits `1` (`the check was interrupted`).
A second signal exits `1` at once. In Go the loop is `ApplianceTree`; a Kubernetes pod's SIGTERM, delivered to PID 1, is handled
(a handler is what lets PID 1 receive it at all).

The Appliance's own data-source flags (the operator's `--manifests`, `--namespace`) are added by the domain; older
single-dash spellings (`-manifests`) are rewritten with a deprecation notice.

**Loading is loud.** An unreadable override, a structural error in it, a statement that prints `[ERROR]`, an `import`
that cannot be read (imports resolve against the override's directory) or a required def left undefined fails the
start with exit `1` and names the problems; on a hot reload the error is printed and the **last good rules keep
running**. At start and whenever it changes, the Appliance prints which defs the override overrode and which it added.

**Exit status** — the same mapping as everywhere, plus one code for findings: `0` the check ran (and, with
`--fail-on-violation`, found nothing); `1` the command failed — rules that do not load, a check that errored (`[ERROR]`
result, a `--timeout` or `--max-output` overrun included), a pipeline that yielded no result at all; `2` a malformed command line (unknown flag or `--format`, no data
source); `3` the check ran correctly and, under `--fail-on-violation`, found violations. A CI job can therefore tell
"policy findings" (`3`) from "the check could not run" (`1`). A check that both errored and found violations exits `1`:
a broken run is never reported as a mere finding. In Go the codes are `cli.ExitOK`, `ExitFailed`, `ExitUsage` and
`ExitViolations`; a domain command returns `cli.ViolationsError` to exit `3`.

**Formats.** A rule set answers with a list of `[rule, resource]` findings; `[]` is compliant. `text` prints the
value as the pipeline produced it. `json` prints one line per check, every key always present:

```json
{"status":"violations","count":2,"violations":[{"rule":"UnselectedPod","resource":"Pod/default/stray"},{"rule":"IdleDeployment","resource":"Deployment/default/lonely"}],"errors":[],"diagnostics":[]}
{"status":"ok","count":0,"violations":[],"errors":[],"diagnostics":[]}
{"status":"error","count":0,"violations":[],"errors":["undefined identifier: no_such_rule"],"diagnostics":[{"kind":"error","message":"undefined identifier: no_such_rule"}]}
```

`status` is `ok`, `violations` or `error`; a finding that is not a `[String, value]` pair becomes `{"value": …}`; a
value JSON cannot hold (a Set, a function) is reported in `errors`, never dropped.

`errors` lists the failure messages as plain strings, as before. `diagnostics` describes the same failures, one object
each and **in the same order**, so a consumer need not parse prose: `kind` is one of `capability_denied`, `unknown_step`,
`stopped`, `timeout`, `output_limit`, `no_result`, `unencodable_value` or `error` (treat an unknown kind like `error`), `message` is the text of the
`[ERROR]` line, and `step`, `capability` and (for `timeout` and `output_limit`) `limit` appear where the failure names them:

```json
{"status":"error","count":0,"violations":[],"errors":["step 'echo' requires the exec capability (policy denies it)"],"diagnostics":[{"kind":"capability_denied","message":"step 'echo' requires the exec capability (policy denies it)","step":"echo","capability":"exec"}]}
```
