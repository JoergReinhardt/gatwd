# Gophers All The Way Down (gatwd) - Architecture

`gatwd` is a purely functional, category-theoretic streaming engine and DSL written in modern Go (Go 1.18+ Generics and Go 1.23+ Iterators). It models data processing and infrastructure traversal as composable, trampolined streams over `iter.Seq`.

---

## 1. The Core Paradigm: Host Engine vs. Guest DSL

`gatwd` strictly separates the **Host Language Mechanics** from the **Guest DSL Runtime**:

```
+--------------------------------------------------------------+
|                    Guest DSL (Scripts)                       |
|   type SumType = A | B | let x = ... | def func = ...        |
+--------------------------------------------------------------+
                               |
                               v
+--------------------------------------------------------------+
|                 Layer 5: Interpreter & Parser                |
|   BitFlag Dynamic Type Registry | Monadic Parser Combinators |
+--------------------------------------------------------------+
                               |
                               v
+--------------------------------------------------------------+
|                 Layer 3 & 4: Engine & Adapters               |
|   iter.Seq Trampoline Pipelines | Network & FS Stream Adapters|
+--------------------------------------------------------------+
                               |
                               v
+--------------------------------------------------------------+
|                 Layer 1 & 2: Collections & Arrows            |
|   32-way HAMT | Cons List | Chunks | Kleisli Arrows | Zippers |
+--------------------------------------------------------------+
                               |
                               v
+--------------------------------------------------------------+
|                 Layer 0: Mathematical Primitives              |
|   Either[L, R] | Pair[A, B] | Option[T] | Lazy Iterators     |
+--------------------------------------------------------------+
```

### The Host Engine (Modern Go)
The internal engine—memory management, AST node pipelines, lazy evaluation loops, and functional transducers (`Map`, `Fold`, `Filter`, `Take`, `Chunk`)—is powered strictly by native Go Generics (`[T any]`) and Iterators (`iter.Seq`). Using `iter.Seq` as a trampoline prevents stack overflow on recursive and infinite streams.

### The Guest DSL Runtime (BitFlag Registry)
Because Go Generics are compile-time only and cannot evaluate user-defined dynamic types parsed at runtime, the interpreter maintains a **BitFlag runtime type registry**:
- Every AST node carries a `BitFlag uint64` quasi-type.
- Sum types (`type OS = Linux | Darwin`) are represented by bitwise OR (`|`).
- Pattern matching (`match { ... }`) evaluates via fast bitmask operations (`FlagMatch`).
- List destructuring (`[pivot, ...rest]`) dynamically binds AST nodes into local branch scopes.

---

## 2. Layered Architecture & Dependency Order

The project strictly follows a **one-way dependency hierarchy** (Layer $N$ imports only from Layers $0 \dots N-1$):

1. **Layer 0 (`functions/`)**: Foundational algebraic types (`Either[error, any]`, `Pair`, `Option`, sequence combinators).
2. **Layer 1 (`collections/`)**: Persistent, immutable data structures with structural sharing (`HAMT`, `List`, `Tree`, `Set`, `Map`, `Chunk`).
3. **Layer 2 (`arrows/`, `comonads/`)**: Category theory combinators (Kleisli arrows `&&&` / `***`, Rose Tree spatial navigation).
4. **Layer 3 (`evaluator/`)**: Stream execution engine, transducers (`PipelineStep`), checkpointing, performance benchmarking (`testing.B`), and diff assertions.
5. **Layer 4 (`adapters/`)**: IO bridges mapping external systems (generalized `Scanner`, network `net.Listener` / `StepDial`, Go `fs.FS` trees) into lazy `iter.Seq` streams.
6. **Layer 5 (`dsltypes/`, `parser/`, `visualizer/`, `pkg/cli/`)**: Language parsing, constants registry, AST nodes, and terminal visualization.

*Note on `legacy/`: Early pre-generics prototypes using reflection and non-generic interfaces were quarantined into the `legacy/` directory to preserve git history while ensuring the active project runs on pure modern Go.*

---

## 3. Quick Start & Setup

### Prerequisites
- Go 1.23 or later (Go 1.24+ recommended for native `maphash.Comparable`).

### Building the CLI
```bash
# Clone the repository
git clone git@github.com:JoergReinhardt/gatwd.git
cd gatwd

# Build the interpreter binary
go build -o gatwd ./cmd/gatwd/...
```

### Running Tests
```bash
# Run all unit tests across all packages
go test ./...

# Run the native self-asserting DSL test suite
./run_dsl_tests.sh
```

### Executing a Gatwd Script
```bash
# Run a script in batch mode
./gatwd examples/tour/01_basics.gatwd

# Launch the interactive REPL
./gatwd
```

---

## 4. Key Capabilities & How To Use Them

### Persistent HAMT Maps & Sets
```go
import "github.com/joergreinhardt/gatwd/collections"

// Create immutable HAMT map
m := collections.NewHAMT[string, int]()
m2 := m.Set("alpha", 100).Set("beta", 200)

val, ok := m2.Get("alpha") // 100, true
// m remains completely unmodified (len 0)
```

### Stream Pipelines with Checkpointing
```go
import (
    "context"
    "github.com/joergreinhardt/gatwd/evaluator"
)

// Pipeline that checkpoints every 50 items
store := evaluator.NewMemoryCheckpointStore()
step := evaluator.StepCheckpoint(context.Background(), store, "job_1", 50)
```

### Network Streaming
```go
import "github.com/joergreinhardt/gatwd/adapters"

// Unfold TCP listener into a stream of connection nodes
connStream := adapters.ListenerToSeq(listener, myFlag)
```

### Go Testable Examples & Synthetic Streams
```go
import "github.com/joergreinhardt/gatwd/evaluator"

// Run programmatic testable example with stdout interception and visual diffs
ex := evaluator.Example{
    Name: "Demo",
    F: func() { fmt.Println("Result") },
    Output: "Result",
    Unordered: false,
}
passed, captured, diff := evaluator.RunExample(ex)

// Run whole suite with subtests under testing.TB
evaluator.RunExamples(t, []evaluator.Example{ex})

// Generate lazy streams of mock domain data for testing generic pipelines
mockStream := evaluator.GenerateMockUsers(1000, flag)
```
