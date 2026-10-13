# Core Components Reference

This document breaks down the major packages and abstractions that form the active `gatwd` architecture.

---

## 1. Layer 0: `functions/` (Foundational Functional Primitives)
The foundational mathematical primitives:
- **`Either[L, R]`**: Standard sum-type container for functional error handling. Left represents errors/failures, Right represents valid values. Provides `Left()`, `Right()`, `IsLeft()`, `IsRight()`, `UnwrapLeft()`, `UnwrapRight()`, `Map()`, `MapLeft()`, `Bimap()`, `Fold()`, `Bind()` (railway-oriented short-circuiting), and `Equal()`.
- **`Pair[A, B]`**: Immutable 2-element product type (2-tuple) with `Fst`, `Snd`, `Map()`, `MapFst()`, `Bimap()`, `Fold()`, and `Equal()`.
- **`Option[T]` / `Maybe[T]`**: Functional optional container (`Just` / `Nothing`) with `Map()`, `Bind()`, `Fold()`, `Filter()`, and `Equal()`.
- **Operators & Ordering (`Compare`, `StructuralEqual`, `EvalBinaryOp`)**: Total ordering comparison across scalars, strings, pairs, and sequences; order-independent mathematical set equality, HAMT key-value equality, and overloaded monoidal concatenation (`++`).
- **`Closure` & `DSLFunc` (callables)**: `Closure` is the DSL's first-class function value — a parameter list plus a body closed over its defining environment, **curried** (`Apply` returns a closure when under-applied and re-applies the result when over-applied). `DSLFunc` is the type-erased builtin callable; `IsCallable`/`Apply` treat both uniformly.
- **`Node[K, C, N]` / `Edge[K, C, E]`**: the graph's ADTs — an identity, a generic comparable *color*, and a payload/weight; edges also carry a group id and an `Inferred` flag. Colors are a generic `C comparable` (the DSL instantiates it with `string`), keeping Layers 0–1 pure generics.
- **Sequences & Iterators**: High-order sequence utilities (`Take`, `Drop`, `Filter`, `Map`, `Fold`, `Zip`, `ZipWith`) built on Go 1.23+ `iter.Seq`.

---

## 2. Layer 1: `collections/` (Persistent Data Structures)
Purely functional, persistent data structures with structural sharing (all implement `Graphable` for the visualizer):
- **`HAMT[K, V]`**: 32-way branching Hash Array Mapped Trie using bitmap index compression (`math/bits.OnesCount32`) and standard `hash/maphash.Comparable`. Provides $O(\log_{32} N) \approx O(1)$ persistent key-value mapping, `Merge()` (union), `Intersect()`, `Diff()` (keyed set algebra), `Equal()`, `MapValues()`, `Fold()`, `Filter()`, and `Reject()`. Implements `json.Marshaler`, `driver.Valuer`, `sql.Scanner`, and `Graphable`.
- **`List[T]`**: Persistent singly-linked Cons list (`Nil()`, `Cons()`, `Head()`, `Tail()`, `Seq()`, `Len()`, `Reverse()`, `Concat()`, `Map()`, `Bind()` (flatMap), `FoldLeft()`, `FoldRight()`, `Filter()`, `Reject()`, `Equal()`). Implements `json.Marshaler`, `driver.Valuer`, `sql.Scanner`, and `Graphable`.
- **`Tree[T]`**: Multi-way Rose Tree (`NewTree()`, `Leaf()`, `Children()`, `AddChild()`, `Size()`, `Depth()`, `Flatten()`, `FlattenChunk()`, `Map()`, `Fold()` bottom-up catamorphism, `Filter()`/`Reject()` subtree pruning, `Equal()`). Implements `json.Marshaler`, `json.Unmarshaler`, `driver.Valuer`, `sql.Scanner`, and `Graphable`.
- **`Set[T]`**: Persistent hash set backed by `HAMT[T, struct{}]`. Provides $O(\log_{32} N)$ structural sharing, `Contains()`, `Add()`, `Remove()`, `Union()`, `Intersect()`, `Diff()`, `IsSubset()`, `Equal()`, `Map()`, `Bind()` (idempotent flatMap), `Fold()`, `Filter()`, and `Reject()`.
- **`BST[T]`**: Persistent binary search tree (`LeafBST`, `NodeBST`, `Insert(v, cmp)`, `Contains(v, cmp)`, `Size()`, in-order `Seq()` — a search tree's catamorphism is its sorted sequence).
- **`Graph[K, C, N, E]`**: Persistent directed **multigraph** over three HAMT indexes (nodes, out-edges, in-edges): `AddNode`, `Connect`/`AddEdge`, `AddInferredEdge`, `AddFanOut`/`AddFanIn` (grouped binary edges), projections (`Nodes`, `Edges`, `NodesByColor`, `EdgesByColor`, `InEdges`/`OutEdges`), Functor/Foldable/Filterable (`MapNodes`, `Bimap`, `FoldNodes`, `FilterEdges`, …), Monoid (`Union`/`Intersect`/`Diff`, `++`), order-independent `Equal`, and cycle-safe algorithms: `BFS`/`DFS`/`Reachable`, `ShortestPath` (Dijkstra), `TopoSort` (Kahn; returns the unordered remainder when cyclic), `SCC` (Tarjan). Higher edge cardinality is modeled with grouped binary edges or by reifying a relation as a node — there are no hyperedges.
- **`Map[K, V]`**: Persistent dictionary backed by `HAMT[K, V]`, implementing `Graphable`.
- **`Chunk[T]`**: Eager, contiguous, bounds-checked slice wrapper for CPU spatial locality and cache efficiency, providing `Concat()`, `Reverse()`, `Slice()`, `Map()`, `Bind()` (flatMap), `Fold()`, `Filter()`, `Reject()`, and `Equal()`. Implements `json.Marshaler`, `json.Unmarshaler`, `driver.Valuer`, and `sql.Scanner`.

---

## 3. Layer 2: Category Theory Combinators
- **`arrows/`**: Kleisli arrows for monadic composition. Provides identity arrows, `Compose`, FanOut (`&&&`), Split (`***`), curry composition, and **`Plus` (`<+>`)** — the ArrowPlus/MonadPlus choice combinator that feeds one input to many arrows and concatenates their outputs (the monoid of arrows under sequence concatenation; zero arrow = empty).
- **`comonads/`**:
  - **Rose Tree Zipper**: Spatial navigation via zippers (`IntoZipper`, `Left`, `Right`, `Down`, `Up`, `Extract`).
  - **Graph cursor**: `GraphCursor` with `Extend`/`ExtendGraph` (cobind) and `DuplicateGraph` — navigation by node identity, since a graph has neighbor *sets* rather than one up/down.
  - **2D Toroidal Grid Comonad**: Discrete 2-dimensional toroidal surfaces (`Grid[T]`, `GridZipper[T]`, `Extract`, `Duplicate`, `Extend`/cobind, `MooreNeighbors`, `VonNeumannNeighbors`, `Count`, `CoordinatesWhere`).


---

## 4. Layer 3: `evaluator/` (Streaming Execution Engine & Testing Harness)
- **`Engine[T]`**: Trampolined stream processing engine that evaluates pipelines over `iter.Seq`.
- **`PipelineStep[T]`**: First-class stream transducers (`func(iter.Seq) iter.Seq`). Includes `StepTake`, `StepDrop`, `StepFilter` (supporting lambda predicates `\x -> ...`), `StepMap` (supporting lambda functions across HAMTs, Trees, Pairs, and Maybes), `StepSort`, `StepToSet`, `StepToList`, `StepToChunk`, `StepSetUnion`, `StepSetIntersect`, `StepSetDiff`, `StepToHAMT`, `StepHAMTGet`, `StepHAMTMerge`, `StepFold` (expanded for `sum`, `product`, `min`, `max`, `count`, `concat`, `any`, `all`), `StepChunk`, `StepAssert`, `StepDebugPrint`.

- **Graph steps** (`graph_steps.go`): construct/project/walk/Functor/Foldable/Filterable/Monoid/Lens over the DSL's string-keyed `Graph`, built from two shapes (`graphMap`: graph→graph, `graphUnfold`: graph→stream); lambdas arrive as plain Go closures so Layer 3 stays parser-free.
- **`StepPlus`**: the pipeline form of `arrows.Plus` — each incoming node runs through every branch adapter and the outputs concatenate (DSL: `(a <+> b <+> c)`). Left (error) nodes pass through.
- **Value helpers**: `HAMTGetKey`/`NavigateHAMTPath` (dotted deep access with errors naming the failing segment), `IsSubsetValue`/`StepSubsetOf` (Record/HAMT key-value containment or Set membership).
- **`StepCheckpoint` & `CheckpointStore`**: Non-intrusive stream watermarking and continuation state tracking without steady-state performance overhead.
- **`BenchmarkPipeline` & `StepTelemetry`**: Performance measurement harness integrating with `testing.B` and live stream telemetry.
- **`Example` & Go Testing Integration**: Full alignment with Go's `testing.InternalExample` interface. Provides `RunExample`, `RunExamples` (with subtests), `RunInternalExamples` (Go standard library bridge), and thread-safe `CaptureStdout` supporting ordered (`// Output:`) and permutation-invariant (`// Unordered output:`) line verification with formatted visual diff cards.
- **`GenerateStream` & `GenerateMockUsers`**: Lazy synthetic stream generators for unit testing generic diversity, chunking, and high-throughput pipelines.

---

## 5. Layer 4: `adapters/` (IO Bridges & Anamorphisms)
- **`ScannerAdapter` & `ScannerToSeq`**: Anamorphic unfold from any `bufio.Scanner` or generalized `Scanner` into lazy ASTNode streams.
- **`ListenerToSeq`, `ConnToLines`, `StepDial`**: Streaming network socket ingress and egress for TCP and Unix domain sockets.
- **`WalkFSAdapter`**: Traverses Go `fs.FS` directory trees, converting directory hierarchies into Comonadic Zippers.
- **`WriteFile`**: Streaming sink to files using Go `io.Writer`.
- **`NativeKubernetesResourceAdapter` / `EnumerateResources`** (`kubernetes_adapter.go`): crawl one GVR through a client-go dynamic client, yielding each resource as a **nested-HAMT payload** (JSON objects → `HAMT`, arrays → ordered `List`, scalars as-is) with list errors lifted to `Left`; discovery enumerates all GVRs including CRDs. client-go is confined to adapter/shim files.
- **`ManifestAdapter`** (`manifest_adapter.go`): the offline twin — YAML/JSON files or directory trees (multi-document), optionally rendered through an injectable renderer (`WithRenderer`; `WithKustomize`/`WithKustomizeCommand` for an external `kustomize build`; `WithKustomizeSwitch` makes rendering switchable at read time), yielding the *same* nested-HAMT payload so downstream DSL is source-agnostic.

---

## 6. Language Front End, Interpreter & Tooling
- **`dsltypes/`** (Layer 0): `BitFlag` dynamic quasi-type system, `ASTNode[T]` host-guest bridge (implementing `fmt.Stringer`), and `Variant` tagged-union values with full Functor (`Map`), catamorphism (`Fold`), and Monad (`Bind`) instances over the payload, plus `SumTypeDefinition` branch tracking.
- **`parser/`** (front end — physically imports only Layers 0–1 and is consumed by `evaluator/` and `pkg/cli/`): monadic parser combinators and the infix expression language. `LambdaExpr` (`\x y -> body`, evaluating to a `Closure` that captures its defining lookup), `ApplyExpr` (juxtaposition application of any callable; failures wrapped in `ApplyError` so the interpreter can tell "a call failed" from "a name did not resolve"), list comprehensions with positional guards, and `ConcatValues` (the `++` monoid shared by the operator and the `concat` function). Sum-type awareness is exposed via `IsConstructor(name)`, `MakeVariant(type, value)`, and `SumTypeOf(branch)` — used by `match` and by `--check` exhaustiveness.
- **`pkg/cli/`** (Layer 5): the `Interpreter` combining the parser, engine, runtime scopes, macro expansions (`def`), local bindings (`let`), script constants (`const`), and host Go constant registrations (`RegisterConstant`).
  - Runtimes and builders: `NewBareInterpreter()`, `NewInterpreter()`, `NewEdition()` (with `WithStep`/`WithType`/`WithPrelude`/`WithPreludeFS`), `NewAppliance()` (hot-reloadable via `atomic.Pointer`), `AdapterStep` (lifts any `evaluator.Adapter` to a step), and the ergonomic step helpers (`MapStep`, `FilterStep`, `ArgStep`, `SinkStep`).
  - **Closures** (`closures.go`): lambda-bodied `def`s resolve to `Closure` values; pipeline-position application is data-last; `fold`/`filter`/`reject`/`map` accept closures; `mapStepOver`/`filterStepOver` give inline lambdas and closures one implementation.
  - **Function forms** (`dslfuncs.go`, `dslfuncs_matrix.go`, `dslfuncs_graph.go`) and the carrier-polymorphic `valueops.go` (`mapValue`, `filterValue`, `foldValue`, `bindValue`): the operator-type matrix as curried, data-last expression functions, resolved in every expression context through `exprFallback`.
  - **`CheckScript` / `gatwd --check`** (`typecheck.go`): structural errors plus advisory findings — non-exhaustive `match`, and a gradual type checker over optional `def f : A -> B` annotations with local inference (`ParseTypeAnnotation`, operator checks decided by the runtime's own `EvalBinaryOp`).

## 7. Domain Shims, Editions & Appliances
- **`pkg/k8sgraph/`**: the thin Kubernetes shim — `CrawlStep`/`ManifestStep`/`KustomizeStep` (adapter → generic `cli.AdapterStep`), `CoreCrawls` (keyword → GVR table), client construction from kubeconfig/in-cluster, `InProcessRenderer` (the sigs.k8s.io/kustomize library, imported only here — `manifests` auto-renders a kustomization root, `--no-kustomize` turns it off), and `ApplyEdition(e, client, opts...)` (steps + embedded prelude; a `nil` client builds an offline-only edition). **No graph or relation logic in Go.** `k8sgraphtest` builds a fake dynamic client from manifest fixtures for tests.
- **`pkg/k8sgraph/lib/k8s/*.gatwd`** (embedded): `types` (node/edge vocabularies), `ingest`, `relations` (`Owns`, `Selects` — label-indexed —, `RoutesTo`, `Configures`, `Mounts`, `Binds`, inferred `Calls`), `queries`, `sources` (`k8s_resources` = every crawl stream merged with `<+>`), and `prelude` (entry point; defines `k8s_graph`).
- **`cmd/gatwd-k8s`** (Edition, pure assembly) and **`cmd/gatwd-k8s-operator`** (Appliance: edition + embedded `rules/compliance.gatwd` — `UnselectedPod`, `IdleDeployment`, `UnusedService`, `DanglingIngress`, `MissingConfigMap`, `DependencyCycle` — with ConfigMap hot-reload, offline `-manifests` or live `-namespace`).
- Usage and extension guide: [`k8s_howto.md`](k8s_howto.md). Design and decision record: [`k8s_graph_design.md`](k8s_graph_design.md). How to mint your own: `.agents/skills/edition-appliance/SKILL.md`.

## 8. Architecture Tests
`internal/archtest` turns architectural rules into tests: the library layers (`functions` → `collections` → `arrows`/`comonads`/`optics`/`transducers`/`transformers`) may import only lower library layers (its `knownViolations` list may only shrink), and core packages import only the standard library and this module (the explicit exception: cobra/pflag in `pkg/cli`) and contain no identifier naming a domain (Clean Core, Mandate 5).

> Note: Early pre-generics prototypes and non-generic interface wrappers have been cleanly separated into the `legacy/` directory and are not part of the active pipeline.
