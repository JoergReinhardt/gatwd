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

1. **Layer 0 (`functions/`, `dsltypes/`)**: Foundational algebraic types (`Either[error, any]`, `Pair`, `Maybe`, `Variant`), the callable values `Closure` (curried, lexically scoped) and `DSLFunc`, graph `Node`/`Edge`, sequence combinators, and the DSL's `BitFlag` type tokens.
2. **Layer 1 (`collections/`)**: Persistent, immutable data structures with structural sharing (`HAMT`, `List`, `Tree`, `BST`, `Set`, `Map`, `Chunk`, `Graph`).
3. **Layer 2 (`arrows/`, `comonads/`, `optics/`, `transducers/`, `transformers/`)**: Category theory combinators (Kleisli arrows `&&&` / `***` / `<+>`, Rose Tree and graph navigation, lenses/traversals, `EitherT`/`StateT`).
   *Front end (`parser/`)* sits here physically: it imports only Layers 0–1 and is consumed by `evaluator/` and `pkg/cli/`.
4. **Layer 3 (`evaluator/`)**: Stream execution engine, transducers (`PipelineStep`), checkpointing, performance benchmarking (`testing.B`), and diff assertions.
5. **Layer 4 (`adapters/`)**: IO bridges mapping external systems (generalized `Scanner`, network `net.Listener` / `StepDial`, Go `fs.FS` trees) into lazy `iter.Seq` streams.
   Also at this level: `interactive/`, `visualizer/` (terminal visualization) and `pkg/plugins/`.
6. **Layer 5 (`pkg/cli/`, `pkg/<domain>/`, `cmd/`)**: The interpreter, builtins, Edition/Appliance builders, thin domain shims and the distribution binaries.

The direction is machine-checked: `go vet ./...` rejects cycles and `internal/archtest` enforces the library-layer boundaries and the Clean Core rule (core imports only the standard library and this module and names no domain).

### 2.1 Planned: Layers 0–2 as a Standalone Functional Library

Layers 0–2 depend only on each other (and, for now, `dsltypes/` in `optics/` and `transformers/`), use
real Go generics, and have no third-party dependencies, so they can in principle be used on their own to
bring functional tools (persistent collections, `Either`/`Maybe`, `State`/`Reader`/`Writer`, lenses,
lazy `iter.Seq` pipelines) to procedural Go code. Today that is a property of the layout, not an enforced
or polished one. Known gaps: `any`-typed folds, `Map` limited to `func(T) T`, DSL-specific files inside
the library layers, and per-type duplication of the operator-type matrix.

Planned direction (proposal, subject to change; see [`library_extraction_plan.md`](library_extraction_plan.md)
and Workstream F in [`language_roadmap.md`](language_roadmap.md)):

- **Dependency inversion:** the DSL depends on the typed core, never the reverse; dynamic `any`/BitFlag
  adapters move into a facade above Layer 2.
- **`iter.Seq[T]` as interchange format:** collections expose `All()`/`FromSeq`; `Map`, `Filter`, `Fold`
  and friends are written once as generic functions, shrinking the matrix cost from N×M to N+M.
- **Enforced boundaries:** a `go list -deps` guard test, a dependency-free `collections/`, and optional
  `fp` / `persistent` / `graph` module split if an external consumer emerges.

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

---

## 5. Runtimes, Builtins, Editions & Appliances

Gatwd's extension mechanism and standalone release model are formalized around four core concepts:

### 5.1 Core Nomenclature
1. **Bare Runtime (`cli.NewBareInterpreter()`)**: A minimalist execution kernel providing only fundamental language primitives (`let`, `def`, `const`, `match`, `seq`, `emit`, `take`, `drop`, `head`, `tail`, `assert`). Standard library adapters and transducers are excluded for ultra-lean, sandboxed environments.
2. **Gatwd Builtins (`cli.NewInterpreter()`)**: The standard interpreter distribution exposing Go standard library capabilities directly to the DSL (`HAMT`, `Set`, `Chunk`, `List`, `Tree`, `Scanner`, `Net`, `Checkpoint`, `Testing`, `Visualizer`, `Math`).
3. **Gatwd Editions (or Distributions)**: Tailored Go binary distributions (e.g., `cmd/gatwd-cloud`, `cmd/gatwd-k8s`) importing `pkg/cli`, bundling third-party Go packages, custom domain types (`interp.RegisterType`), and domain step transducers (`interp.RegisterStep`).
4. **Gatwd Appliances**: Self-contained, single-purpose release packages (such as Kubernetes Operators, stream processing microservices, or compliance auditing CLIs) that bundle pre-compiled, embedded `.gatwd` DSL code (`//go:embed`) via `LoadScriptFS` or `EvaluateScript`. Appliances run out of the box without external script files, while allowing live overrides via flags or Kubernetes ConfigMaps.

---

### 5.2 Two Execution Tiers: API Level vs. Embedded DSL
Each Gatwd Edition or Appliance can operate across two execution tiers:
- **API-Level Execution (Pure Go Engine)**: Directly builds and evaluates pipelines using `evaluator.Engine`, `collections.HAMT`, `collections.Chunk`, transducers (`evaluator.PipelineStep`), and adapters (`adapters.NativeKubernetesAdapter`, `adapters.ScannerToSeq`). Delivers maximum performance (~7.3M items/sec) with zero DSL parsing overhead and full Go type safety.
- **Embedded Gatwd DSL (`//go:embed`)**: Compiles `.gatwd` scripts directly into static binary releases using `embed.FS` and executes them via `interp.LoadScriptFS(...)` or `interp.EvaluateScript(...)`. Users can override embedded scripts at runtime via CLI arguments or Kubernetes ConfigMap volume mounts.

---

### 5.3 Architectural Archetypes & Reference Patterns

#### Archetype 1: Specialized CLI Edition
```go
package main

import (
    "github.com/joergreinhardt/gatwd/pkg/cli"
)

func main() {
    cli.NewEdition("gatwd-audit").
        WithType("AuditLog", parseLog).
        WithStep(cli.MapStep("mask", maskSecret)).
        WithStep(cli.FilterStep("critical_only", isCritical)).
        WithStep(cli.ArgStep("tag", tagWithEnv)).
        Run() // Handles flags, -i REPL, script loading automatically
}
```

#### Archetype 2: High-Throughput HTTP Streaming Server
```go
package main

import (
    "fmt"
    "net/http"
    "github.com/joergreinhardt/gatwd/evaluator"
    "github.com/joergreinhardt/gatwd/dsltypes"
)

func streamHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain")
    engine := evaluator.NewEngine[string]()
    
    // Inbound stream
    source := func(yield func(dsltypes.ASTNode[string]) bool) {
        for i := 0; i < 10000; i++ {
            if !yield(dsltypes.NewASTNode(fmt.Sprintf("event_%d", i), 1)) {
                return
            }
        }
    }
    
    // Pure Go pipeline with batching chunk transducer
    chunkStep := evaluator.StepChunk[string](250)
    for chunkNode := range engine.EvaluatePipeline(source, []evaluator.PipelineStep[string]{chunkStep}) {
        fmt.Fprintf(w, "Batch: %v\n", chunkNode.Value)
    }
}
```

#### Archetype 3: Kubernetes Operator Appliance with Hot-Reloading
```go
package main

import (
    "embed"
    "github.com/joergreinhardt/gatwd/pkg/cli"
)

//go:embed reconcile_rules.gatwd
var rulesFS embed.FS

func main() {
    // Zero-boilerplate appliance with embedded baseline and live ConfigMap overrides
    app, _ := cli.NewAppliance(
        cli.WithEmbeddedFS(rulesFS, "reconcile_rules.gatwd"),
        cli.WithConfigMapOverride("/etc/gatwd/rules.gatwd"),
        cli.WithStep(cli.MapStep("enrich_node", enrichClusterNode)),
    )
    
    // In operator reconcile loop:
    res := app.Execute("spawn Cluster: default |> audit_and_reconcile")
    _ = res
}
```

---

#### Archetype 4: A Domain Graph Tool (the Kubernetes reference implementation)
The most complete example is `gatwd-k8s` (Edition) and `gatwd-k8s-operator` (Appliance). The architecture it demonstrates is the template for any domain:
```
adapters/kubernetes_adapter.go    live crawl  → nested-HAMT payload per resource   (client-go confined here)
adapters/manifest_adapter.go      offline     → the SAME payload shape (YAML / kustomize build)
pkg/k8sgraph/                     thin shim: CrawlStep/ManifestStep → generic cli.AdapterStep; ApplyEdition(e, client)
pkg/k8sgraph/lib/k8s/*.gatwd      the logic, embedded: types, ingest, relations, queries, sources, prelude
cmd/gatwd-k8s/                    Edition main = pure assembly
cmd/gatwd-k8s-operator/           Appliance: edition + embedded rules/compliance.gatwd + ConfigMap hot-reload
```
```haskell
-- offline and live feed the same prelude; (a <+> b <+> …) merges the crawl streams:
spawn StrVal: ./manifests |> manifests       |> k8s_graph |> graph_shortest_path web nginx-pod
spawn StrVal: default     |> k8s_resources   |> compliance
```
Relations (`Owns`, `Selects`, `RoutesTo`, `Configures`, `Mounts`, `Binds`, inferred `Calls`) and compliance rules are DSL, not Go; Go does only real I/O. Live/offline equivalence is proven by a test over a fake dynamic client rather than assumed. See [`k8s_graph_design.md`](k8s_graph_design.md) and the [Development Playbook](development_playbook.md) §5.

---

### 5.4 Standalone Release & Container Packaging Strategy

- **Static Binary Releases**: Built with `CGO_ENABLED=0 go build -ldflags="-s -w"`, packaging all Go code and embedded `.gatwd` scripts into a single, zero-dependency executable.
- **Distroless Container Images**: Deployed using `FROM gcr.io/distroless/static:nonroot` for secure, minimal attack-surface Kubernetes pods and server containers.

---

### 5.5 Zero-Downtime Hot-Reloading & ConfigMap Mutation Architecture

While compiled bytes within `embed.FS` reside in immutable binary memory (`.rodata`), the **active rule execution state** of a Gatwd Appliance can be dynamically mutated at runtime without restarting the container or dropping network connections.

#### Architectural Mechanics
1. **Baked-in Baseline (`embed.FS`)**: The appliance bundles standard `.gatwd` scripts as an unalterable baseline.
2. **Mounted ConfigMap Override**: In production, Kubernetes mounts a ConfigMap volume (e.g., at `/etc/gatwd/rules.gatwd`).
3. **Atomic Pointer Swapping (`atomic.Pointer[cli.Interpreter]`)**: When a configuration change occurs:
   - The appliance reads the updated script from disk (or falls back to `embed.FS` if the override file is absent or deleted).
   - A fresh `cli.NewInterpreter()` instance is instantiated and populated with the new rules.
   - The appliance atomically swaps its active interpreter pointer (`activeInterp.Store(newInterp)`).
   - In-flight stream evaluations complete safely on the previous interpreter, while incoming requests immediately execute the updated logic.

#### Three Hot-Reload Trigger Modes
- **Kubernetes API Informer (Sub-Second)**: Operators watch the `ConfigMap` resource directly via `client-go` Informers, reacting to `kubectl apply` in milliseconds without waiting for kubelet's volume sync period.
- **Volume Symlink Watcher (`fsnotify`)**: Daemons watch the mounted `/etc/gatwd` directory. Kubelet updates mounted ConfigMaps by rotating the `..data` symlink atomically, which triggers the file watcher.
- **Admin Endpoint & Signals**: Microservices expose a `POST /-/reload` webhook or handle `syscall.SIGHUP`.

See [`examples/hot_reload_appliance/main.go`](../examples/hot_reload_appliance/main.go) for an executable reference implementation and [`examples/hot_reload_appliance/main_test.go`](../examples/hot_reload_appliance/main_test.go) for automated fallback tests.

---

## 6. Architectural Case Study: Conway's Game of Life

For a complete pedagogical walkthrough of how `gatwd` balances **mathematical rigor**, **execution performance**, **developer intimidation**, and the **Clean Core Prevention Rule**, see the comprehensive architectural case study:

👉 **[`docs/GAME_OF_LIFE.md`](GAME_OF_LIFE.md)**

This case study compares three distinct tiers of cellular automata implementations across the repository:
1. **Tier 1: Category Theory Comonads** ([`examples/game_of_life/`](../examples/game_of_life)) — Spatial zippers, counit (`extract`), and cobind (`extend`) over 2D toroidal grids in pure Go.
2. **Tier 2: Minted Custom Editions** ([`examples/game_of_life_edition/`](../examples/game_of_life_edition)) — Builder pattern (`cli.NewEdition`) isolating domain types while exposing declarative `.gatwd` pipelines.
3. **Tier 3: 100% Pure Gatwd DSL** ([`examples/game_of_life_pure_dsl/`](../examples/game_of_life_pure_dsl)) — Sparse coordinate sets, multi-generator list comprehensions, HAMT frequencies, and reusable `def conway_step` macro running on vanilla, unextended `gatwd`.





