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

## 7. ADT Pattern Matching & List Destructuring

Pattern matching supports full list deconstruction (`[pivot, ...rest]`), enabling functional recursion like Quicksort:

```haskell
def quicksort = match {
    [] -> empty
    | [pivot, ...rest] -> seq { rest |> filter_less pivot |> collect_list |> quicksort | pivot | rest |> filter_greater pivot |> collect_list |> quicksort }
}

spawn IntVal: 0 |> emit 5, 2, 8, 1, 9, 3 |> drop 1 |> collect_list |> quicksort |> collect_list |> assert [1 2 3 5 8 9]
```

## 8. Persistent HAMT Dictionaries (`to_hamt`, `hamt_get`)

Fold key-value streams into a 32-way branching persistent Hash Array Mapped Trie:

```haskell
spawn StrVal: Start |> emit user:alice, role:admin |> to_hamt |> hamt_get role |> assert admin
```

## 9. Performance Chunks & Continuations (`chunk`, `checkpoint`, `telemetry`)

* `chunk <size>`: Batches lazy streams into eager, bounded contiguous arrays for CPU cache efficiency.
* `checkpoint <id> <interval>`: Watermarks stream offsets, allowing computations to pause and resume.
* `telemetry`: Zero-overhead element counter and duration profiler.

```haskell
spawn IntVal: 0 |> emit 1, 2, 3, 4 |> drop 1 |> chunk 2 |> head |> assert [1 2]
spawn IntVal: 0 |> emit 1, 2, 3 |> drop 1 |> checkpoint job_1 2 |> math_add 10 |> head |> assert 11
```

## 10. IO Streaming & Network Adapters (`scan_file`, `dial_tcp`)

* `scan_file <path>`: Streams lines from a file using the generalized scanner.
* `dial_tcp <address>`: Transmits payloads across a network socket and streams back response lines.

```haskell
spawn StrVal: Start |> scan_file examples/tour/data/data1.txt |> head |> assert M
spawn StrVal: "ping" |> dial_tcp "localhost:9090" |> assert "ECHO:ping"
```
