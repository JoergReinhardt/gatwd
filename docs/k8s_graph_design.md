# Design & Implementation Plan: Generic Typed Graph Core + Kubernetes Graph Edition

> Status: **IMPLEMENTED — Phases 1–8 complete** (2026-10-13). Originally a living design
> document grounded against the actual codebase (grammar, engine types, edition/appliance
> machinery); now retained as the **design and decision record**. §5 records what each phase
> actually delivered, §6 the decisions with their rationale, §8 the verified mandate checklist.
> The way this project was run is distilled in [`development_playbook.md`](development_playbook.md); how to *use* and
> extend the result is in [`k8s_howto.md`](k8s_howto.md).
> Scope: (1) a generic cyclic / directed / typed ("colored") / weighted **graph
> abstraction** in gatwd core, wired through every layer to the DSL; and (2) a
> **Kubernetes resource-graph** built *on top of that DSL* as adapters + plugins +
> a preloaded `.gatwd` library, delivered as Editions and Appliances.

---

## 0. Guiding principles (the 5 mandates applied)

- **Mandate 5 (Clean Core)** is the spine. The *graph* is a pure,
  category-theoretic, domain-free abstraction and belongs in core. No **import** of a
  third-party package (except cobra/pflag in `pkg/cli`, allowlisted 2026-10-13) and no **identifier** naming `kubernetes`/`k8s` may appear in
  `functions/`, `dsltypes/`, `collections/`, `arrows/`, `comonads/`, `optics/`,
  `transducers/`, `transformers/`, `evaluator/`, `parser/`, or `pkg/cli/` (comments may
  still use Kubernetes as an *example*). This is enforced by `internal/archtest`
  (`TestCleanCoreImportsNoThirdParty`, `TestCleanCoreNoDomainIdentifiers`).
- **Mandate 3 (All-Layers-Through)**: every graph capability is exposed from
  Layer 0/1 up to a DSL builtin and asserted in `tests/dsl_exhaustive_test.gatwd`.
- **Mandate 2 (Operator-Type Matrix)**: the graph joins
  `List/Chunk/Set/HAMT/Tree` as a new *collection row* and must fill the
  `map / fold / filter / ++ / == / walk` columns where mathematically valid.
- **Mandate 1 & 8 (Anti-Mockup / Hard-Work-First)**: the Kubernetes node/edge
  *types and relation rules* are authored in **pure gatwd DSL** (a preloaded
  library), NOT Go `for`-loops faking a graph. Go is used only for real I/O and
  SDK access (client-go), isolated in adapters/plugins.

### The Edition boundary rule (user clarification, 2026-10-13)

A gatwd Edition/Appliance **is** a Go binary, so it **must** have a `main.go`.
That is expected. The rule constrains *what that `main.go` contains*:

> The edition `main.go` is **pure assembly**. Everything it registers
> (`WithType`/`WithStep`/`WithConstant`, adapter factories, the `//go:embed`'d
> prelude) must be *supplied by* **core**, **`adapters/`**, **`pkg/plugins/`**, or
> the **preloaded `.gatwd` DSL**. No Kubernetes-specific algorithm, loop, relation
> rule, or type definition may be authored inline in the edition `main.go`.

---

## 1. Problem statement & use cases

Build a tool that:

1. Reads Kubernetes resources (built-in **and** CRDs) from a live cluster API,
   and *also* offline (`kustomize build` YAML, raw manifests).
2. Lifts each resource into a typed graph **node**, and discovers typed,
   directional, weighted **edges**: `Service—(selects)→Pod`,
   `Ingress—(routes-to)→Service`, `Deployment—(owns)→ReplicaSet—(owns)→Pod`,
   `Pod—(mounts)→Secret/ConfigMap/PVC`, `ServiceAccount—(binds)→Role`, …
   Some edges are **inferred**, not read from the API — the `Configures` chain is
   the canonical example: a `ConfigMap—(configures)→Pod` (the Pod mounts/envs the
   CM), the Pod runs an image, and a CM *value* equals a Service name/port. The
   API never states "this app calls that service," so we **assume** a
   `Pod—(consumes)→Service` edge: the value would not be in the ConfigMap
   otherwise. Inferred edges carry a lower weight / an `inferred` marker so walks
   can trust-weight or exclude them.
3. Produces a **cyclic**, multi-typed, directed, weighted graph.
4. Lets the user define node/edge types + relation rules **on the fly in the DSL**.
5. Walks/transforms/accumulates purely in the DSL (BFS/DFS, reachability,
   weighted paths, folds, projections, subgraph extraction, rewrites).

| Use case | Graph operation |
| :-- | :-- |
| Dependency resolver | reachability / topological walk over `depends-on` edges |
| Request tracking (ingress → nodePort → backend → aux svc) | weighted directed path-find |
| Compliance management | predicate folds + subgraph match for violations |
| Orchestration | ordered traversal + staged accumulation |
| Config-as-code refactoring | optic-based focused mutation, re-emit YAML |
| Kustomize offline analysis | same graph, file source instead of API |
| **Live → kustomize transform** (DSL-definable) | read live resources → graph → DSL rewrite → emit kustomize base+overlays |

Delivery: **Editions** (CLI tools) and **Appliances** (operators, API servers)
minted from a shared **Kubernetes Graph base library**.

---

## 2. Grounding: what the engine already is (and already has)

The scan of the repo established the real substrate the design must build on:

### 2.1 The DSL is already a node-expansion graph-traversal trampoline

From `docs/grammar.md`, a pipeline *is* a traversal:

```ebnf
Pipeline  ::= SpawnNode { ("|>" | "||>") PipelineStep }
SpawnNode ::= "spawn " Identifier ":" Identifier
```

- `spawn IntVal: 10 |> math_add 5 |> assert 15` — `spawn T: V` mints a root
  `ASTNode`, `|>` threads it through steps, `||>` is the parallel variant.
- An **Adapter** expands one node into a lazy stream of child nodes:
  ```go
  // evaluator/engine.go:14
  type Adapter[T any] func(dsltypes.ASTNode[T]) iter.Seq[dsltypes.ASTNode[T]]
  // evaluator/pipeline.go:29
  type PipelineStep[T any] func(iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]]
  ```
- The runtime value type carried through the DSL is `functions.Either[error, any]`
  (steps are `StepHandler = func(stepStr string, interp *Interpreter)
  (evaluator.PipelineStep[functions.Either[error,any]], bool)`).

So `adapters/kubernetes_adapter.go`'s `NativeKubernetesAdapter` is *already* a
"graph crawler" — but a **generative/streaming** one (unfold), not a **stored**
graph. The user wants a *persistent cyclic graph value* you can fold, find cycles
in, weight-path over, and rewrite. **That stored `Graph` collection is the core's
missing piece.** The crawl trampoline *feeds* it; relation rules *build its
edges*; graph ops *consume* it.

### 2.2 The DSL already defines types — including collection types

```ebnf
TypeDef        ::= "type " Identifier "=" (SumType | RecordType | TupleType | CollectionType | NativeType)
SumType        ::= Identifier { "|" Identifier }          -- e.g. type Color = Red | Green | Blue
RecordType     ::= "{" FieldDef { "," FieldDef } "}"
CollectionType ::= ("Set(" | "List(" | "Tree(" | "BST(" | "BTree(" | "RBTree(") Identifier ")"
NativeType     ::= Identifier   -- maps to a Go factory via WithType/RegisterType
```

Consequences for us:
- **Node/edge "colors" are DSL `SumType`s** — `type EdgeColor = Selects | RoutesTo
  | Configures | Owns | Mounts | Binds`. Workstream A (constructor matching, DONE)
  makes `match { Selects -> … | Owns -> … }` route correctly.
- Adding **`Graph(T)`** to `CollectionType` in the grammar is the natural,
  consistent way to surface the new collection — siblings `Set/Tree/BST/...`
  already exist.
- `RecordType` gives us resource shapes in-DSL; field access already works via the
  HAMT builtins (`to_hamt`, `hamt_get`, `hamt_set`, `hamt_del` — seen in
  `tests/dsl_exhaustive_test.gatwd`).

### 2.3 The edition / appliance / preload machinery (exact APIs)

- **Edition** (`pkg/cli/edition.go`): fluent builder
  `cli.NewEdition(name).Bare()?.WithType(name, func(string) any).
  WithConstant(name, any).WithStep(StepHandler).Run()`; `BuildInterpreter()`
  applies everything onto `NewInterpreter()` / `NewBareInterpreter()`.
- **Appliance** (`pkg/cli/appliance.go`): `cli.NewAppliance(opts...)` with
  `WithEmbeddedFS(fsys, path)` / `WithEmbeddedScript(str)` /
  `WithConfigMapOverride(path)` / `WithStep` / `WithType` / `WithEdition`;
  `Reload()` reads ConfigMap-or-embedded, `EvaluateScript`s it, and atomically
  swaps the interpreter pointer (zero-downtime hot reload). This is the operator
  substrate.
- **Steps** (`pkg/cli/steps.go`): ergonomic `MapStep(name, fn)`,
  `FilterStep(name, pred)`, `ArgStep(prefix, fn)`, `SinkStep(name, fn)` — all
  return a `StepHandler` that pattern-matches `stepStr` and operates on the
  `Either[error,any]` stream.
- **Registration**: `interp.RegisterStep(StepHandler)`,
  `interp.RegisterType(definition string, factory func(string) any)`,
  `interp.RegisterConstant(name, any)`.
- **Third-party adapter registry** (`pkg/plugins/registry.go`):
  `plugins.Register(keyword string, factory AdapterFactory)` where
  `AdapterFactory = func() evaluator.Adapter[functions.Either[error,any]]`.
- **Preload / embed**:
  - DSL statement `import "path.gatwd"` → `LoadScriptContent(content, baseDir)` →
    resolves relative imports → `LoadScript(path)`.
  - `LoadScriptFS(fsys, filename)` evaluates a script from an `embed.FS` — this is
    how a prelude ships *inside* a binary.

### 2.4 How an Adapter reaches the DSL today (the crawl/step glue)

Verified against the engine + examples:

- `evaluator.StepAdapter[T](Adapter[T]) PipelineStep[T]` turns any adapter into a
  pipeline step (`functions.Bind(stream, adapter)`).
- `evaluator.EitherT[T](Adapter[T]) Adapter[Either[error,T]]` lifts a pure adapter
  into the DSL's error-aware payload type.
- A **plugin keyword** is the DSL-facing path: `plugins.Register("kw", factory)`
  in an `init()`, blank-import the plugin package, and the token `kw` in a `|>`
  pipeline is looked up in `plugins.Registry` and wrapped via
  `StepAdapter(factory())` (`pkg/cli/cli.go` pipeline-step fallthrough).
- A **custom `StepHandler`** (registered via `WithStep`) is the alternative path
  and the one that can **close over configuration** (a dynamic client, a GVR).
- The lower-level **flag-keyed crawl** (`engine.Register(sourceFlag, adapter)` +
  `engine.Evaluate(seed)`, with `OnMissingAdapter` fallback) is how
  `examples/kubernetes_repl` drives `NativeKubernetesAdapter` today — in raw Go,
  **not** surfaced as a DSL keyword. For graph-building we prefer *explicit* DSL
  relation steps (`spawn … |> k8s_pods |> …`) over implicit type-keyed recursion.

### 2.5 Gaps surfaced (prerequisite work)

1. **`Edition` has no embedded-prelude option.** Only `Appliance` can embed a
   `.gatwd` (`WithEmbeddedFS`/`WithEmbeddedScript`); `game_of_life_edition`
   confirms Editions take their `.gatwd` as a runtime CLI arg.
   **DECISION (approved 2026-10-13): add generic prelude support to the `Edition`
   core** — this is expected to be used by *many* future editions, so it is built
   as a first-class, general-purpose feature, not a workaround. Scope:
   `Edition.WithPreludeFS(fsys fs.FS, path string)` **and** `Edition.WithPrelude(
   script string)`, both accumulating (an edition may stack several preludes, in
   registration order), evaluated by `BuildInterpreter()` via
   `LoadScriptFS`/`EvaluateScript` *after* type/constant/step registration so the
   prelude can use registered builtins/types. **Generic, core-level, domain-free**
   → `pkg/cli/edition.go`. Tracked as **Phase P1** below.
2. **`NativeKubernetesAdapter` yields only resource *names* (`string`), and is
   typed `Adapter[string]`.** Two coupled problems: (a) graph-building needs the
   **full resource** (labels, selectors, ownerRefs, spec), so widen it to emit the
   `unstructured.Unstructured` content as a **nested `HAMT`** payload (DSL reads
   `… |> hamt_get spec |> hamt_get selector` via existing builtins); (b) the DSL
   engine runs over `Adapter[Either[error, any]]`, so the widened adapter must be
   `Adapter[any]` lifted through `evaluator.EitherT` — it cannot be dropped into
   the engine as `Adapter[string]`.
3. **`pkg/plugins.AdapterFactory` takes no arguments** (`func()
   evaluator.Adapter[Either[error,any]]`), so a *configured* client/GVR can't be
   injected through the keyword-plugin path (the registry's own comment flags this
   as future work). Two options: widen the factory to accept config
   (`func(cfg any) Adapter[...]`), **or** register the K8s source as a custom
   `StepHandler` that closes over the client at edition-assembly time
   (recommended near-term — no registry change, config via closure).
4. **No `Graph` collection** exists (there is `Tree`, acyclic single-root). This
   is the main core build.

---

## 3. The generic graph abstraction (CORE)

### 3.1 Layer 0 — ADTs (`functions/graph_node.go`)

- `Node[K comparable, V any]` — identity `K`, payload `V`, **color** = a `Variant`
  tag (reuse `dsltypes.Variant` / `SumTypeDefinition.Branches`; do **not** invent
  a parallel enum). A node's "type" is just its variant tag.
- `Edge[K comparable, W any]` — `From K`, `To K`, **color** (`Variant`),
  **weight** `W`, optional payload, plus two small fields that fall out of the
  cardinality decision (§6 #8): an optional **`group`** id (ties the binary edges
  that together form one logical 1:n / n:n relation) and an **`inferred`** flag
  (edges we *assume* rather than read from the API, e.g. `consumes`). Directional
  by construction; undirected = two directed edges (documented). Default weight
  `1` (int) so unweighted walks work unchanged.
- Operator wiring in `functions/operators.go`: structural `==`/`!=` on
  nodes/edges; edges order lexicographically by `(from, to, color, weight)`.

### 3.2 Layer 1 — the collection (`collections/graph.go`)

`Graph[K comparable, N any, E any]` — a **persistent** directed multigraph,
dogfooding existing structures (Mandate 4), **no raw Go `map`**:

- out-edges `HAMT[K, List[Edge[K,E]]]`, reverse in-edges `HAMT[K,
  List[Edge[K,E]]]` (cheap predecessor queries), node payloads `HAMT[K, N]`.
- Cyclic by design; walks are cycle-safe via a `Set[K]` visited set.
- Operations (all immutable, return a new graph):
  - construct: `Empty`, `AddNode`, `AddEdge`, `RemoveNode`, `RemoveEdge`;
    cardinality sugar `AddFanOut(color, w, from, tos []K)` /
    `AddFanIn(color, w, froms []K, to)` (expand to grouped binary edges, §6 #8)
  - project: `Nodes`, `Edges`, `OutEdges`, `InEdges`, `Neighbors`,
    `NodesByColor`, `EdgesByColor`, `EdgesByGroup(g)` (fetch a whole n-ary relation)
  - **Functor**: `MapNodes`, `MapEdges`, `Bimap`
  - **Foldable**: `FoldNodes`, `FoldEdges`, `FoldWalk` (fold along traversal)
  - **Filterable**: `FilterNodes`, `FilterEdges` → subgraph, `Reject`
  - **Monoid**: `Union`/`++`, `Diff`, `Intersect`
  - **Walks**: `BFS`, `DFS`, `WalkFrom(k, strategy) iter.Seq[Node]`,
    `Reachable`, `ShortestPath`/`WeightedPath` (Dijkstra when `E` is
    `Ord`+additive), `TopoSort → Either[Cycle, []K]`, `SCC`.
- Structural Go tests + `Example*` with `// Output:` (STYLEGUIDE §4/§5).

### 3.3 Layer 2 — traversal structure

- **Comonad zipper** (`comonads/graph_cursor.go`): focus a node with its
  neighborhood as context. `Extract` = focused node; `Extend`/`Duplicate` =
  recompute a value at every node from its local neighborhood — this is how
  compliance/propagation rules are expressed comonadically. (The DSL already has
  `into_zipper |> down |> extract` for trees — mirror that surface for graphs.)
- **Optics** (`optics/graph_optics.go`): lenses onto node payload / edge weight;
  prisms onto "nodes/edges of color C"; a traversal onto "all out-edges" — gives
  the DSL focused immutable rewrites (`focus`/`update` already exist in grammar).
- **Transducers** (`transducers/graph_walk.go`): walk-as-transducer so traversal
  composes with `map/filter/take/fold` and streams lazily.

### 3.4 Layer 3 — evaluator steps (`evaluator/graph_steps.go`)

Generic graph walk/fold/map steps as `PipelineStep`s emitting `ASTNode`s into the
trampoline, composable with the existing stream pipeline. No domain logic.

### 3.5 Layer 5 — grammar + DSL builtins

- **Grammar**: add `Graph(` to `CollectionType` in `parser/` so `type K8sGraph =
  Graph(Resource)` parses.
- **Builtins** in `pkg/cli/cli.go` (follow the `to_hamt`/`hamt_get` naming
  precedent; prefer `MapStep`/`FilterStep`/`ArgStep`, raw `StepHandler` only for
  the walk trampoline):
  ```
  graph_empty  graph_add_node  graph_add_edge  graph_remove_node  graph_remove_edge
  graph_nodes  graph_edges  graph_out  graph_in  graph_neighbors
  graph_nodes_of  graph_edges_of           -- by color/variant
  graph_map_nodes  graph_map_edges  graph_bimap
  graph_fold_nodes  graph_fold_edges  graph_fold_walk
  graph_filter_nodes  graph_filter_edges
  graph_union (++)  graph_diff  graph_intersect
  graph_add_fanout  graph_add_fanin  graph_edges_of_group   -- n-ary relation sugar (§6 #8)
  graph_add_edge_inferred  graph_edges_inferred             -- inferred/assumed edges (§6 #9)
  graph_bfs  graph_dfs  graph_walk_from  graph_reachable
  graph_shortest_path  graph_topo_sort  graph_scc
  graph_focus / graph_over / graph_set     -- optic rewrites
  ```
- Each gets a self-asserting case in `tests/dsl_exhaustive_test.gatwd`.

---

## 4. The Kubernetes layer (NON-core)

Split exactly as the user requires.

### 4.1 "Any K8s edition could use it" — generic K8s I/O (`adapters/`, `pkg/plugins/`)

- `adapters/kubernetes_adapter.go` — **exists** (dynamic client + GVR, CRD-capable).
  **Fix (Phase 5):** emit full `unstructured` content as a nested-`HAMT` payload,
  not a bare name; add discovery-client GVR/CRD enumeration so "all resources incl.
  CRDs" holds. Keep client-go confined to this file.
- `adapters/kustomize_adapter.go` — **new**: runs `kustomize build` / reads
  manifests, yields the *same* nested-`HAMT` payloads → identical graph-building
  DSL for live vs. offline. `sigs.k8s.io/kustomize` confined here.
- Register both via `pkg/plugins.Register("k8s", …)` / `Register("kustomize", …)`
  so editions enable them by keyword. (Third-party deps **only** here — Mandate 5.)

### 4.2 "Only this thing needs it" — the Kubernetes Graph base library

**(a) Thin Go shim `pkg/k8sgraph/` (registration + I/O plumbing only, no logic).**
Its jobs: build the dynamic client (I/O), register the adapter-backed crawl steps
(`CrawlStep`, see §4.3) and the kustomize step, register any `Resource` type
token, and provide the `//go:embed`'d prelude via an `ApplyEdition` assembly
helper. It contains **no relation logic and no graph algorithms** — those are in
core (graph) or in the DSL prelude (rules). Full shape in §4.3.

**(b) The pure-DSL prelude `pkg/k8sgraph/lib/k8s/*.gatwd` (the real substance).**
Authored entirely in DSL on top of §3 graph builtins + §4.1 adapters:

- `types.gatwd` — node & edge colors as DSL sum types:
  ```
  type NodeColor = Pod | Service | Ingress | ConfigMap | Secret | Deployment | ReplicaSet | PVC | ServiceAccount | Role
  type EdgeColor = Selects | RoutesTo | Configures | Owns | Mounts | Binds | Calls
  ```
- `ingest.gatwd` — map raw `unstructured` (nested HAMT) payloads → typed graph
  nodes.
- `relations.gatwd` — the **relation rules** as pure-DSL predicates producing
  colored, weighted edges (see §7 for concrete sketches). This is the heart of
  the tool and is DSL, not Go.
- `queries.gatwd` — reusable walks for the use cases.
- `prelude.gatwd` — `import`s the above (the single embed entry point).

Because colors/types are DSL sum types and rules are DSL predicates, **users
extend the type/edge system on the fly** by editing `.gatwd` — zero Go.

### 4.3 The client→DSL-step bridge (`pkg/k8sgraph`)

The lifting + step-exposure machinery is **generic** and lives in core as
`cli.AdapterStep` (Phase P3a). The only domain-specific Go here is *constructing*
the configured K8s adapter and naming it — then handing it to the generic helper:

```go
// pkg/cli/steps.go  (CORE, Phase P3a — domain-free)
// AdapterStep exposes any pure Adapter[any] as a named |> pipeline step,
// lifting it into the Either[error,any] engine. Reused by every data-gathering edition.
func AdapterStep(name string, adapter evaluator.Adapter[any]) StepHandler {
    lifted := evaluator.EitherT(adapter)
    return func(stepStr string, interp *Interpreter) (evaluator.PipelineStep[functions.Either[error,any]], bool) {
        if stepStr != name { return nil, false }
        return evaluator.StepAdapter(lifted), true
    }
}

// pkg/k8sgraph/steps.go  (DOMAIN, Phase P3b — client-go wiring only)
func CrawlStep(keyword string, client dynamic.Interface, gvr schema.GroupVersionResource) cli.StepHandler {
    adapter := adapters.NativeKubernetesResourceAdapter(client, gvr) // Adapter[any], §4.1 widened
    return cli.AdapterStep(keyword, adapter)                         // all generic glue delegated to core
}
```

`pkg/k8sgraph` then exposes an assembly helper:
```go
//go:embed lib/k8s/*.gatwd
var preludeFS embed.FS

func ApplyEdition(e *cli.Edition, client dynamic.Interface) *cli.Edition {
    return e.WithStep(CrawlStep("k8s_pods",     client, podsGVR)).
             WithStep(CrawlStep("k8s_services", client, svcGVR)).
             WithStep(KustomizeStep("kustomize")).            // offline source (adapters/)
             WithPreludeFS(preludeFS, "lib/k8s/prelude.gatwd") // needs §2.5 gap #1
}
```

### 4.4 Editions & Appliances (`cmd/`)

- `cmd/gatwd-k8s/main.go` — **Edition**, pure assembly (templated on
  `examples/custom_cli` + `examples/game_of_life_edition`):
  ```go
  func main() {
      client := k8sgraph.DynamicClientFromKubeconfig() // helper in pkg/k8sgraph (I/O only)
      k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s"), client).Run()
  }
  ```
  If `WithPreludeFS` (gap #1) is not yet added, the assembly-only fallback is:
  ```go
  func main() {
      client := k8sgraph.DynamicClientFromKubeconfig()
      e := k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s"), client) // without the prelude line
      interp := e.BuildInterpreter()
      interp.LoadScriptFS(k8sgraph.PreludeFS(), "lib/k8s/prelude.gatwd")
      cli.RunWithInterpreter(interp)
  }
  ```
- `cmd/gatwd-k8s-operator/main.go` — **Appliance** (templated on
  `examples/hot_reload_appliance`):
  ```go
  //go:embed rules/compliance.gatwd
  var complianceRules string
  func main() {
      client := k8sgraph.DynamicClientFromKubeconfig()
      base := k8sgraph.ApplyEdition(cli.NewEdition("k8s-operator"), client)
      app, err := cli.NewAppliance(
          cli.WithEdition(base),
          cli.WithEmbeddedScript(complianceRules),          // baseline rules
          cli.WithConfigMapOverride("/etc/gatwd/rules.gatwd"), // live hot-reload
      )
      // ... watch ConfigMap -> app.Reload(); serve app.Execute(...) ...
  }
  ```
  (The embedded prelude of node/edge *types* + base relations comes in via
  `ApplyEdition`'s `WithPreludeFS`; the appliance's own embed carries the
  *use-case* rules, e.g. compliance checks, which hot-reload.)
- Further appliances (API server, request-tracer, dependency-resolver) = same
  base, different embedded DSL — proving "mint many from one base."

Templates confirmed: `examples/custom_cli` + `examples/game_of_life_edition`
(Edition builder), `examples/hot_reload_appliance` (Appliance + `//go:embed`),
`examples/14_plugins` (plugin `Register`→DSL-keyword path),
`examples/kubernetes_repl` (raw dynamic-client crawl wiring). **Not** a template:
`cmd/gatwd-gen` (it is a grammar/codegen tool, not an edition).

---

## 5. Implementation plan (phased; gate each with `go test ./... && go vet ./... && ./run_dsl_tests.sh`)

### Prerequisite phases (user-approved 2026-10-13, committed)

These three are independent of the graph build and may land first / in parallel;
the numbered phases depend on them where noted.

**Phase P1 — generic `Edition` prelude support (core, `pkg/cli/edition.go`). ✅ DONE (2026-10-13).**
Added `Edition.WithPreludeFS(fsys fs.FS, path string)` and
`Edition.WithPrelude(script string)`, both **accumulating** via a single ordered
`[]preludeEntry` (interleaved registration order preserved across both kinds).
`BuildInterpreter()` evaluates all preludes *after* type/constant/step
registration (so preludes can use registered builtins), via
`interp.LoadScriptFS`/`interp.EvaluateScript`; a failed embedded load panics
(loud, programmer-error). Tests in `pkg/cli/edition_prelude_test.go`:
`TestEditionWithPrelude`, `TestEditionWithPreludeStacking` (second prelude uses
first), `TestEditionWithPreludeFS` (via `fstest.MapFS`). All green; full
`go test ./... && go vet ./... && ./run_dsl_tests.sh` pass. (No K8s.)

**Phase P2 — structured K8s resource payload (`adapters/`, L4). ✅ DONE (2026-10-13).**
Added `adapters.NativeKubernetesResourceAdapter(client, gvr) evaluator.Adapter[any]`
yielding each resource's full `unstructured.Unstructured` content as a **nested
`HAMT`** payload. A recursive `unstructuredToValue` lifts the decoded Go shape into
gatwd persistent collections — JSON object `map[string]any` → `*collections.HAMT[
string, any]` (navigable via `hamt_get`/`hamt_values`), JSON array `[]any` →
order-preserving `*collections.List[any]` (iterable: ownerReferences, containers,
…), scalars carried through. A `List` failure is yielded as an `error`-valued node
so `evaluator.EitherT` lifts it to `Left` (loud, not silent-halt). Added
`EnumerateResources(discovery.DiscoveryInterface) ([]GVR, error)` → discovery-client
GVR/CRD enumeration (skips subresources `name` with `/` and non-`list`-verb kinds;
CRD groups surface like core kinds; tolerates partial-discovery errors). The legacy
`Adapter[string]` name-only adapter is kept for back-compat. Tests
(`adapters/kubernetes_adapter_test.go`): `TestNativeKubernetesResourceAdapter`
(asserts the nested-HAMT/List shape — metadata→labels→app, ownerReferences list,
spec — not strings), `TestNativeKubernetesResourceAdapterThroughEitherT` (clean
`Right` lift into the engine payload), `TestEnumerateResources` (stub discovery:
CRD included, subresource + non-listable excluded). client-go confined to this
file. All gates green (`go build/vet/test ./...`, `./run_dsl_tests.sh`).

**Phase P3 — adapter→DSL-step bridge, split generic vs. domain (DECISION).**
The user asked whether the crawl step (§4.3) belongs in core, since *every* future
data-gathering edition will "gather via an adapter → lift into core types →
crawl." **Decision: yes, split it.** The reusable machinery is domain-free; only
the adapter *construction* (client-go) is K8s-specific.

- **P3a — generic (core, `pkg/cli/steps.go`): `cli.AdapterStep(name string,
  adapter evaluator.Adapter[any]) StepHandler`. ✅ DONE (2026-10-13).** A new
  sibling to `MapStep`/`FilterStep`/`ArgStep`/`SinkStep` that takes *any* pure
  `Adapter[any]`, lifts it with `evaluator.EitherT`, and exposes it as a named `|>`
  step via `evaluator.StepAdapter`. This is the generalization: K8s/AWS/DB/file
  editions all reuse it. (It also subsumes the keyword-plugin path's ergonomics
  while allowing config via closure, which `pkg/plugins.AdapterFactory` can't.)
  Test `TestAdapterStep` in `pkg/cli/edition_prelude_test.go` (in-memory fan-out
  adapter, no K8s). Green.
- **P3b — domain (`pkg/k8sgraph/steps.go`): the construction only. ✅ DONE (2026-10-13).**
  `k8sgraph.CrawlStep(keyword, client, gvr)` builds the configured K8s
  `Adapter[any]` (from P2) and hands it to `cli.AdapterStep(keyword, adapter)` —
  no lifting/step logic, just client-go wiring. Shim also provides the only genuine
  I/O: `RestConfigFromKubeconfig` (in-cluster → $KUBECONFIG → ~/.kube/config),
  `DynamicClientFromKubeconfig`, `DiscoveryClientFromKubeconfig`, plus convenience
  core GVRs (`PodsGVR`/`ServicesGVR`/…). `pkg/k8sgraph/doc.go` documents the "pure
  assembly, zero graph/relation logic" boundary. Test `TestCrawlStep`
  (`pkg/k8sgraph/steps_test.go`): registers a fake-client CrawlStep on a real
  `cli.NewInterpreter()` and drives it from the DSL —
  `spawn StrVal: default |> k8s_pods |> hamt_path spec.selector.app` → `nginx`,
  `… |> hamt_get kind` → `Pod` (proves P2→P3a→P3b wired through to Layer 5). All
  gates green.
- *Optional stretch (generic `pkg/plugins`):* widen `AdapterFactory` to
  `func(cfg any) evaluator.Adapter[...]` so the keyword-plugin registry can also
  carry config (its comment already anticipates this). Independent of P3a/P3b.

Tests: `spawn KubeNS: default |> k8s_pods` over a fake client yields resource
HAMTs (P3b); `cli.AdapterStep` surfaces an arbitrary adapter as a step (P3a).

### Numbered phases

**Phase 0 — spikes.** Confirm node-payload rep (nested `HAMT`, recommended);
confirm the `Graph(` grammar insertion point in `parser/`. (Mostly done via this
doc's scan.)

**Phase 1 — Graph ADTs + collection (CORE L0–L1). ✅ DONE (2026-10-13).**
`functions/graph_node.go` — `Node[K,C,N]` + `Edge[K,C,E]` ADTs (color is a
**generic `C comparable` param**, not hard-coded `dsltypes.Variant`, to keep
Layer 0/1 pure generics per the host/guest split; the DSL instantiates `C` in
Phase 4); `Edge` carries `Group` + `Inferred`; `Equal`, lexicographic `Compare`,
marker/stringer methods. `collections/graph.go` — persistent directed multigraph
`Graph[K,C,N,E]` over three HAMT adjacency indexes (nodes, out, in):
construct (`AddNode`/`AddEdge`/`Connect`/`AddFanOut`/`AddFanIn`/`AddInferredEdge`/
`RemoveNode`/`RemoveEdge`), project (`Order`/`Size`/`Nodes`/`Edges`/`OutEdges`/
`InEdges`/`Neighbors`/`Predecessors`/`NodesByColor`/`EdgesByColor`/`EdgesByGroup`/
`InferredEdges`), Functor (`MapNodes`/`MapEdges`/`Bimap`), Foldable
(`FoldNodes`/`FoldEdges`), Filterable (`FilterNodes`/`FilterEdges`/`Reject*`),
Monoid/set-algebra (`Union`/`Diff`/`Intersect`), cycle-safe walks
(`WalkFrom`/`BFS`/`DFS`/`Reachable`), order-independent `Equal`, `IsGraph`/
`GraphEqual`/`Graph()` visualizer hook. Operators wired in
`functions/operators.go`: `++`→`Union` (EvalBinaryOp) and `==`→`GraphEqual`
(StructuralEqual). Structural tests `collections/graph_test.go` (assert memory
types, incl. operator dispatch through `functions.EvalBinaryOp`) + `Example*`
with `// Output:` in `collections/graph_example_test.go`. All gates green
(`go build/vet/test ./...`, `./run_dsl_tests.sh`). *Deferred to Phase 2:*
`FoldWalk`, weighted `ShortestPath`/`TopoSort`/`SCC`, zipper, optics. *Deferred to
Phase 4:* `Graph(` grammar + DSL builtins.

**Phase 2 — traversal structure (CORE L1–L2). ✅ DONE (2026-10-13).** The
traversal *algorithms* landed beside the collection (L1, `collections/graph_walk.go`):
`FoldWalk` (catamorphism along a cycle-safe walk), `ShortestPath` (Dijkstra, numeric
weight via `functions.ToFloat64`, O(V²), deterministic tie-break, `from==to`→cost 0),
`TopoSort` → `functions.Either[[]K cyclic-remainder, []K order]` (Kahn, lexicographic),
`SCC` → `[][]K` (Tarjan, components sorted; partitions the declared vertex set). The
L2 traversal *structures*: **comonad** `comonads/graph_cursor.go` —
`GraphCursor` (`Extract`/`To`/`Successors`/`Predecessors`/`OutEdges`/`InEdges`),
`Extend` (N→N, iterable to a fixpoint) + free `ExtendGraph` (N→R cobind) +
`DuplicateGraph` (W a→W (W a)); **optics** `optics/graph_optics.go` —
`NodePayloadLens` (Maybe payload; Set Nothing removes), `ColorNodePrism`,
`AllEdgeWeights`/`OutEdgeWeights` traversals, plus a new generic
`optics/traversal.go` `Traversal[S,A]` optic (`GetAll`/`Modify`/`Set`/`ComposeTraversals`);
**transducer** `transducers/graph_walk.go` — `WalkNodes` + `TransduceWalk` (Sequence
over the cycle-safe walk, fuses Map/Filter). Structural Go tests + `Example*` with
`// Output:` in each package; `doc.go` overviews updated. All gates green
(`go build/vet/test ./...`, `go test -run Example ./...`, `./run_dsl_tests.sh`).
*Deferred to Phase 4:* DSL builtins (`graph_fold_walk`/`graph_shortest_path`/
`graph_topo_sort`/`graph_scc`/`graph_focus`/`graph_over`/…) + exhaustive asserts.

**Phase 3 — evaluator steps (CORE L3). ✅ DONE (2026-10-13).** Landed together with
Phase 4 wiring as `evaluator/graph_steps.go`: the DSL pins the generic
`Graph[K,C,N,E]` to `dslGraph = *collections.Graph[string,string,any,any]` (string
id/color, `any` payload/weight — `ShortestPath` reads weight via `ToFloat64`). Two
reusable step shapes — `graphMap` (graph→graph, Left/non-graph pass-through) and
`graphUnfold` (graph→stream of derived values) — back every `StepGraph*`. Godoc +
testable `Example*` (`graph_steps_example_test.go`: `ExampleStepGraphNodes`,
`ExampleStepGraphShortestPath`, both with `// Output:`).

**Phase 4 — grammar + DSL builtins (CORE L5). ✅ DONE (2026-10-13).** `Graph(` added
to `CollectionType` in `parser/parser.go` (so `type K8sGraph = Graph(String)`
parses) and `docs/grammar.md`. `pkg/cli/cli.go` gains `buildGraphStep(stepStr)`, a
`graph_*` dispatcher wired into the pipeline-step chain, exposing **26 builtins**:
- *construction:* `graph_empty`, `graph_add_node`, `graph_add_edge`,
  `graph_add_edge_inferred`, `graph_add_fanout`, `graph_add_fanin`,
  `graph_remove_node`, `graph_remove_edge`
- *projection:* `graph_order`, `graph_size`, `graph_nodes`, `graph_edges`,
  `graph_out`, `graph_in`, `graph_neighbors`, `graph_nodes_of`, `graph_edges_of`,
  `graph_edges_of_group`, `graph_edges_inferred`, `graph_reachable`
- *walks/algos:* `graph_bfs`, `graph_dfs`, `graph_walk_from`,
  `graph_shortest_path`, `graph_topo_sort`, `graph_scc`

Exhaustive self-asserting cases in `tests/dsl_exhaustive_test.gatwd` §§65–68 (edges
project as `from-Color->to`; all green via `./run_dsl_tests.sh`). (Prelude support is
**Phase P1**, done independently.) *Gate met:* a graph is constructible, walkable,
and queryable from pure DSL; core still has zero K8s.

**Phase 4b — DSL higher-order / binary-operand cells (CORE L3+L5). ✅ DONE
(2026-10-13).** The Functor/Foldable/Filterable/Monoid matrix cells are now DSL
builtins. L3 (`evaluator/graph_steps.go`): `StepGraphMapNodes`/`MapEdges`/`Bimap`,
`FilterNodes`/`FilterEdges`, `FoldNodes`/`FoldEdges`/`FoldWalk` — each takes the
compiled lambda as a plain Go closure, so Layer 3 stays parser-free. L5
(`pkg/cli/cli.go`): `(*Interpreter).buildGraphLambdaStep` (a *method*, so it can
resolve names against the interpreter's `Constants`/`Variables`/`RuntimeNodes`),
with lambda-compilation helpers `compileUnary`/`compilePred`/`compileBinary`
(built on `parser.EvalInfixString`), `splitParen` (nesting-aware lambda/seed
split), and `resolveGraphOperand`/`evalSeed`. Set-algebra reuses the existing
reflection-dispatched `StepSetUnion`/`StepSetIntersect`/`StepSetDiff` (Graph already
has `Union`/`Intersect`/`Diff` + `++`→`Union` via `EvalBinaryOp`), named
`graph_union`/`graph_diff`/`graph_intersect` and taking a let-bound graph operand.
**Binding convention:** node steps bind the node *payload* (its id by default); edge
steps bind the edge *weight*. New builtins (12):
`graph_map_nodes`/`graph_map_edges`/`graph_bimap`,
`graph_filter_nodes`/`graph_filter_edges`,
`graph_fold_nodes`/`graph_fold_edges`/`graph_fold_walk`,
`graph_union`/`graph_diff`/`graph_intersect`. Exhaustive asserts in
`tests/dsl_exhaustive_test.gatwd` §§69–71 (all green).

**Phase 4c — optics rewrites (CORE L3+L5). ✅ DONE (2026-10-13).** The single-node
payload **Lens** trio closes the last matrix column. L3 (`evaluator/graph_steps.go`):
`StepGraphFocus` (read/preview the payload — missing node → injected error, not
silent), `StepGraphSet` (Lens Set with Just), `StepGraphOver` (Lens "over": Get →
f → Set), all backed by `optics.NodePayloadLens` so editing one node copies a path,
not the whole graph. L5: `graph_focus <id>` + `graph_set <id> <value>` in the free
`buildGraphStep`; `graph_over <id> (\p -> …)` in `buildGraphLambdaStep` (reuses
`compileUnary`). These are the *focused* counterpart to the bulk Functor
`graph_map_nodes` — one node, siblings/edges untouched. Asserts §72 +
`ExampleStepGraphOver`. **The graph collection row of the operator-type matrix is
now complete through the DSL** (construct / project / walk / Functor / Foldable /
Filterable / Monoid / Lens).

**Phase 5 — generic offline I/O (L4). ✅ DONE (2026-10-13).** Depends on **P2**
(structured adapter). Added `adapters/manifest_adapter.go`:
`adapters.ManifestAdapter() evaluator.Adapter[any]` reads YAML/JSON manifests from a
file *or* a directory tree (every `*.yaml`/`*.yml`/`*.json`, lexical `WalkDir`
order), decodes multi-document files via `k8s.io/apimachinery/pkg/util/yaml`
(`NewYAMLOrJSONDecoder`, empty docs skipped), and yields **the same nested-HAMT
payload as P2** through the shared `unstructuredToValue` — so every downstream
graph/relation rule is source-agnostic (live vs. offline), giving the §6 #10
live→offline parity. Path comes from the spawned node value; read/parse errors
surface as error nodes → `Left` via `EitherT`. Edition-facing twin of `CrawlStep`
in the shim: `k8sgraph.ManifestStep(keyword) cli.StepHandler`
(`= cli.AdapterStep(keyword, adapters.ManifestAdapter())`). Tests:
`adapters.TestManifestAdapterMultiDoc` / `…DirectoryWalk` (ignores non-manifests,
asserts lexical order) / `…MissingPath`; `k8sgraph.TestManifestStep` drives it from
the DSL — `spawn StrVal: <file> |> manifests |> hamt_path spec.selector.app` →
`nginx` (confirms a filesystem path survives as a `spawn` value). All gates green.

**Kustomize rendering — EXTERNAL renderer, adapter stays kustomize-agnostic
(✅ 2026-10-13, user-directed).** Kustomize libraries are a daily occurrence, so
`ManifestAdapter` is parameterized (functional options `ManifestOption`) to *deal
with* them: when pointed at a **kustomization root** (`isKustomizationRoot`: dir
directly holds `kustomization.yaml`/`.yml`/`Kustomization`) it first renders the
library to a YAML stream, then decodes it via the shared `decodeManifests` — i.e.
render, then treat as any other YAML. Crucially **no kustomize specifics live in the
adapter**: it only knows "render a dir → YAML bytes via a pluggable function". The
default renderer `execRenderer` shells out to the external `kustomize build` (NO
`sigs.k8s.io/kustomize` Go dep — just `os/exec`). Options:
- `WithKustomize()` — enable rendering (default cmd `kustomize build`);
- `WithKustomizeCommand(name, args…)` — e.g. `WithKustomizeCommand("kubectl",
  "kustomize")` → `kubectl kustomize <dir>` (dir appended last);
- `WithRenderer(func(dir)([]byte,error))` — inject a custom renderer; this is also
  the seam through which an edition could later plug the in-process
  `sigs.k8s.io/kustomize` API **without the adapter importing it**.
Plain files/dirs still read directly (mixed raw+overlay trees "just work"); a failed
render (binary missing, bad kustomization) → error node → `Left`. Shim step
`k8sgraph.KustomizeStep(keyword)` = `AdapterStep(kw, ManifestAdapter(WithKustomize()))`.
Tests: injected-renderer (render-root rendered + decoded; plain dir falls back to
walk, renderer NOT called) + **real `kubectl kustomize` e2e** (gated on the binary;
`namePrefix` transformer proves genuine rendering). All gates green.

**Phase 6 — K8s Graph base library.** Depends on **P1, P2, P3** and Phases 1–4.
`pkg/k8sgraph/` shim (`ApplyEdition`, `DynamicClientFromKubeconfig`, embedded
prelude) + `lib/k8s/*.gatwd`. Prove 3 relations end-to-end first (`Owns` via
ownerRefs, `Selects` via selector ⊆ labels, `RoutesTo` via Ingress backend), then
expand the relation matrix. DSL tests over fixture manifests.

**Phase 6 — ingest blocker RESOLVED (2026-10-13), option C.** The first blocker
(handoff §3: no way to turn a crawled resource stream into graph nodes/edges, since
every graph builtin took literal step args) is cleared by the user-chosen **option
C** — generalize the existing custom `fold` so its accumulator can be a graph and the
reducer a DSL lambda returning the updated graph. The enabling, generic, domain-free
primitive is **function application (juxtaposition) in the infix expression
language**: Layer 0 `functions.DSLFunc` (type-erased callable); `parser/infix.go`
`ApplyExpr` + juxtaposition parsing (with an `appDisabled` guard so space-separated
list literals `[2 1]`/`[nginx web]` still parse as separate elements — parens &
comprehensions re-enable application); shared `evaluator.HAMTGetKey`/
`NavigateHAMTPath` helpers (extracted from the hamt_get/hamt_path steps); and a
registry of expression-level builtins in `pkg/cli/dslfuncs.go` wired into the fold &
graph-lambda lookups — `graph_empty` (seed value), `graph_add_node`,
`graph_add_edge`, `graph_add_inferred_edge`, `graph_add_edges`, `hamt_get`,
`hamt_path` (each the function form of the like-named step). Ingest is now pure DSL:
`resources |> fold (\g r -> graph_add_node (hamt_get "kind" r) (hamt_get "name" r) r g) graph_empty`,
edges via a second fold feeding `graph_add_edges`. Asserts: `tests/dsl_exhaustive_test.gatwd`
§73 (non-K8s: id-list, record ingest via hamt_get/hamt_path, edge ingest);
`parser/apply_test.go`; `pkg/cli/dslfuncs_test.go` (structural, asserts the raw
`*collections.Graph`) + Examples. Grammar documented in `docs/grammar.md`. All gates
green.

**Phase 6 — prelude + `ApplyEdition` LANDED (2026-10-13).** `pkg/k8sgraph/lib/k8s/`
`{types,ingest,relations,queries,prelude}.gatwd` + `pkg/k8sgraph/edition.go`
(`ApplyEdition`, `//go:embed lib/k8s/*.gatwd`, `PreludeFS`/`PreludePath`; `client` may
be nil for an offline edition). Shape: `k8s_ingest` is a zero-parameter `def` fragment
(`fold (\g r -> graph_add_node …) graph_empty`); each relation is a **graph-to-graph
pass** `fold (\_ g -> graph_add_edges [ spec | … <- graph_payloads_of … ] g) graph_empty`
(a graph stream element reaches the reducer as `g`, so the whole graph is readable and
the rule can join across nodes); `k8s_graph = k8s_ingest |> k8s_relations`. `Owns`,
`Selects` (via `subset_of`), `RoutesTo` proven over `testdata/stack.yaml` in
`pkg/k8sgraph/edition_test.go`. Findings that shaped it: (1) multi-arg lambda `def`s
used as `fold add_node …` still mis-evaluate — use zero-param fragments, which work;
(2) `map` on a graph element is the graph Functor (payload map), not whole-graph
application — use the `fold (\_ g -> …)` form. **Generic core additions (Mandate 3/5):**
`graph_payloads` / `graph_payloads_of color g` and `hamt_has_path` expression builtins;
**positional comprehension guards** (`parser.ListComprehensionExpr.FilterAfter` — a
filter now runs as soon as its variables are bound, Haskell-style, so a guard can
protect a later dependent generator); and **FS-aware `import`** (a prelude loaded from
an `embed.FS` now imports its siblings from that FS). Asserts: DSL §76,
`parser.TestComprehensionPositionalGuard`, `cli.TestLoadScriptFSResolvesImportsInFS`.
**NEXT:** Phase 7 (`cmd/gatwd-k8s` Edition + `cmd/gatwd-k8s-operator` Appliance); then
expand the relation matrix (`Configures`, `Mounts`, `Binds`, `Calls`, inferred edges).

**Phase 7 — Editions & Appliances. DONE.** `cmd/gatwd-k8s/` Edition (pure assembly;
falls back to an offline edition when there is no kubeconfig). `cmd/gatwd-k8s-operator/`
Appliance: embedded `rules/compliance.gatwd` (UnselectedPod, IdleDeployment) over
`k8s_graph`, ConfigMap override + `Reload`, `-manifests`/`-rules`/`-interval` flags.
New generic core builtins `graph_in_of` / `graph_out_of color id g` (DSL §77) back the
rules. E2E tests over `pkg/k8sgraph/testdata/stack.yaml`: request trace
(`graph_shortest_path Ingress/default/web Pod/default/nginx-pod`; the ids were the bare `web`/`nginx-pod` until hardening W4 qualified them) and dependency resolve (`graph_reachable`,
`k8s_apply_order`) in `pkg/k8sgraph/edition_test.go`; operator tests in
`cmd/gatwd-k8s-operator/main_test.go`. Gap: the operator checks offline manifests; a
live source needs the multi-GVR crawl streams merged into one `k8s_graph` input.

**Phase 8 — live source, relation matrix, rules, perf. DONE (2026-10-13).**
- *Live source (design choice (a), generic).* `arrows.Plus` (ArrowPlus / MonadPlus
  choice, Layer 2) → `evaluator.StepPlus` (Layer 3) → the DSL step `(a <+> b <+> …)`
  merges several sub-pipelines into one stream. `lib/k8s/sources.gatwd` defines
  `k8s_resources` as the `<+>` of all twelve crawl steps (`k8sgraph.CoreCrawls`,
  now incl. PVCs, ServiceAccounts, Roles, RoleBindings), so
  `spawn StrVal: default |> k8s_resources |> k8s_graph` feeds the SAME prelude as the
  offline path. `TestLiveOfflineParity` (fake dynamic client, helper package
  `pkg/k8sgraph/k8sgraphtest`) proves §6 #10: identical nodes, declared edges and
  inferred edges for both fixtures. The operator gained a live mode (`-namespace`).
- *Relation matrix.* `Configures` (ConfigMap → Pod/Deployment-template via volume,
  `configMapKeyRef`, `envFrom`), `Mounts` (PVC → workload), `Binds` (RoleBinding →
  ServiceAccount subjects and → its Role) and the **inferred** `Calls` (weight 0.5:
  a workload env value / ConfigMap data value equal to a Service name). Authored in
  pure DSL with first-class closures (`k8s_workloads`, `k8s_names` — replaced by `k8s_ref` in hardening W4 —, `k8s_config_refs`,
  `k8s_claim_refs`); edges only join existing nodes. Fixture `testdata/stack_full.yaml`;
  `TestPodSpecRelations`. Node color `PVC` renamed to the real kind
  `PersistentVolumeClaim`; `RoleBinding` added.
- *Generic core additions it needed* (all domain-free, DSL §§79–82): the full
  operator-type matrix as curried expression functions (Workstream E), the
  graph read-side function forms (`graph_scc`, `graph_topo_sort`, `graph_reachable`,
  `graph_shortest_path`, `graph_edge_pairs`, `graph_has_node`, `graph_nodes`,
  `graph_order`, `graph_size`, `graph_add_inferred_edges`), `hamt_get_or`.
- *Compliance rules* (`rules/compliance.gatwd`, each a closure graph → violations):
  UnselectedPod, IdleDeployment, **UnusedService, DanglingIngress, MissingConfigMap,
  DependencyCycle** (graph_scc inside a comprehension). `testdata/broken.yaml`
  violates every rule; `TestOperatorAllRules`, `TestOperatorLiveSource`.
- *Perf.* `BenchmarkK8sGraph` + `benchmarks/run_k8s_graph.sh`. Profiling showed the
  O(services × pods) `Selects` pass was about three quarters of the run; replaced by a label-indexed
  pass (pure DSL, HAMT of pod lists): 2040 ms → 294 ms at 100 × 1000, scaling now
  ~linear. Semantics pinned by `TestSelectsIndexedPassSemantics`.
- *Dogfood/docs.* Relations rewritten on closures; `Example_graphInOfOutOf`,
  `Example_graphAlgorithmsInComprehension`; `EvaluateToValues("")` no longer panics;
  `doc.go` for the prelude. **Remaining ideas:** ~~edges keyed by `metadata.name`
  only (a Pod and Service with the same name collide)~~ — **done in the hardening plan's
  W4: ids are `Kind/namespace/name` and a duplicate id is an error** (`docs/hardening_plan.md`
  §7), a leak/soak harness for a long-running operator `-interval` loop,
  CRD-driven GVR discovery in the edition (`EnumerateResources` exists), live
  `kustomize`-vs-cluster drift diff (`graph_diff`).

---

## 6. Key decisions & open questions

1. **Node payload** = nested `HAMT` (dogfood + reuse `hamt_get`), *recommended*,
   vs. a dedicated `Resource` type token. Leaning HAMT.
2. **Edge weights** = generic `E` with `Ord`+monoid for weighted paths; default `1`.
3. **Colors/types = `Variant` sum types** (reuse Workstream A). Confirm the DSL
   can `match` the variants we need at the scale required; if a gap appears,
   implement the language primitive (Anti-Mockup), don't fake it in Go.
4. **Relations = DSL**, firmly (Mandate 1/8). Go only ingests raw resources.
5. **`Graph` vs `Tree`** — keep `Tree` (acyclic single-root) as-is; `Graph` is the
   cyclic generalization. Later dogfood opportunity: re-express tree ops via graph.
6. **Incremental updates** — persistent HAMT adjacency gives structural sharing so
   an operator's watch events add/remove a few nodes/edges without copying the
   graph. Validate with a bench.
7. **`||>` parallel crawl** — the parallel pipeline operator already exists; large
   clusters should crawl resources in parallel then assemble the graph.
8. **Edge cardinality — DECISION: binary multigraph + grouping + reification; no
   true hyperedges in core.** The question was whether edges can be 1:n / n:1 /
   n:n (e.g. a Service exposes one IP:port but fans out to *n* backend Pods). True
   hyperedges (an edge over *sets* of nodes) would force us to re-derive every
   algorithm — BFS/DFS/Dijkstra/SCC/`TopoSort` are all defined on *binary* digraphs
   — for little gain. Instead the core graph stays an ordinary **directed
   multigraph with binary (1:1) edges**, and higher cardinality is expressed two
   ways, both of which keep every algorithm standard:
   - **Grouped binary edges** — one logical 1:n relation = *n* binary edges that
     share a `group` id + `color`. `AddFanOut`/`AddFanIn` create them;
     `EdgesByGroup(g)` fetches the whole relation back as a unit. This covers
     "Service `selects` 3 Pods" directly (3 `Selects` edges, one group).
   - **Reification** — when the relation *itself* has identity/attributes, promote
     it to a **node**. This is the clean category-theoretic n:n, and Kubernetes
     usually *already gives you that node*: `Service —(exposes)→ EndpointSlice
     —(targets)→ Pod×n` models the one-IP:port / n-backends case exactly, because
     `EndpointSlice` is the API's own reification of that relation. No core change
     needed — it's just nodes + binary edges.

   So the answer to "is it possible?": yes, and it's modeled *without* leaving the
   binary-digraph core — via grouping for simple fan-out/in and reification for
   attributed n:n. A genuine hypergraph variant stays an *open option* for a future
   use case that truly needs set-valued edges, but is explicitly out of scope now.
9. **Inferred edges** — relations not stated by the API (e.g. the `consumes`
   assumption from a ConfigMap value equal to a Service name, §1) are marked
   `inferred` and given a configurable (lower) weight, so walks/compliance can
   trust-weight, down-rank, or exclude them explicitly.
10. **Live → kustomize transform** — because ingest from the live adapter and from
    the kustomize adapter produce the *same* graph node payloads (nested HAMT), a
    pure-DSL rewrite over the graph can emit kustomize base+overlays; the transform
    rules are themselves DSL (user-definable), reusing the graph optics
    (`graph_over`/`graph_set`) for focused mutation before re-serialization.

---

## 7. Appendix A — concrete DSL sketches (syntax matched to the real grammar)

> Illustrative; exact builtin names finalize in Phase 4. Uses confirmed grammar:
> `type`, `spawn`, `|>`/`||>`, `match`, `\x -> …`, comprehensions `[e | p <- xs,
> guard]`, `emit`, `collect_list`, `to_hamt`/`hamt_get`, `focus`/`update`.

**types.gatwd**
```
type NodeColor = Pod | Service | Ingress | ConfigMap | Deployment | ReplicaSet
type EdgeColor = Selects | RoutesTo | Configures | Owns | Mounts
type Resource  = { kind: String, name: String, ns: String, labels: HAMT, spec: HAMT }
```

**relations.gatwd — `Owns` (ownerReferences), `Selects` (selector ⊆ labels)**
```
-- an edge-builder takes the full node set and emits colored edges.
-- Owns: child.metadata.ownerReferences[].name -> that owner node.
def rel_owns = \child ->
    child |> hamt_get ownerRefs
          |> map (\owner -> graph_add_edge Owns 1 owner child)

-- Selects: Service whose spec.selector is a subset of a Pod's labels.
-- 1:n fan-out (one Service -> its matching Pods) as ONE grouped relation (§6 #8):
def selector_matches = \svc -> \pod ->
    svc |> hamt_get spec |> hamt_get selector
        |> subset_of (pod |> hamt_get labels)

def rel_selects = \svc -> \pods ->
    graph_add_fanout Selects 1 svc [ pod | pod <- pods, selector_matches svc pod ]
```

**relations.gatwd — `Configures` (the inferred chain, §1 corrected)**
```
-- Direct edge: a ConfigMap configures the Pods that mount it or env-ref it.
def rel_configures = \cm -> \pods ->
    graph_add_fanout Configures 1 cm [ pod | pod <- pods, pod_uses_cm pod cm ]

-- INFERRED edge: a CM value equals a Service name/port, so we ASSUME the Pod
-- configured by that CM consumes the Service (the value would not be there
-- otherwise). Marked inferred + lower weight so walks can trust-weight it (§6 #9).
def rel_consumes = \cm -> \pods -> \services ->
    [ graph_add_edge_inferred Consumes 0.5 pod svc
    | svc <- services
    , val <- (cm |> hamt_values)
    , val == (svc |> hamt_get name)
    , pod <- pods, pod_uses_cm pod cm ]
```

**queries.gatwd — dependency resolve & request trace**
```
-- all resources a Service transitively depends on (reverse of Owns/Selects):
def deps_of = \svc -> graph_reachable svc |> collect_list

-- request path from an Ingress to backends along RoutesTo |> Selects:
def trace_request = \ingress ->
    ingress |> graph_walk_from (RoutesTo, Selects) |> collect_list
```

**A user adding a new edge type on the fly (no Go):**
```
-- append to types.gatwd:  type EdgeColor = … | Calls
-- append to relations.gatwd:
def rel_calls = \pod -> \services ->
    [ graph_add_edge Calls 1 pod svc
    | svc <- services
    , (pod |> hamt_get spec |> hamt_get env) contains (svc |> hamt_get name) ]
```

## 8. Appendix B — mandate compliance checklist

- [x] Core contains no `kubernetes`/`k8s` identifier or third-party import (Mandate 5) — enforced by `internal/archtest`. ✅
- [x] Graph fills `map/fold/filter/++/==` as a new collection row at the Go
      level (Mandate 2) — Phase 1. ✅ (DSL builtins for the row land in Phase 4.)
- [x] Graph construction/projection/walk ops → DSL builtins + `tests/dsl_exhaustive_test.gatwd`
      asserts §§65–68 (M3) — Phases 3–4. ✅
- [x] Graph Functor/Foldable/Filterable/Monoid → DSL builtins + asserts §§69–71 (M3) —
      Phase 4b. ✅
- [x] Graph optics (single-node payload Lens: `graph_focus`/`graph_set`/`graph_over`) →
      DSL builtins + asserts §72 (M3) — Phase 4c. ✅ **Matrix row complete through DSL.**
- [x] K8s relations are pure DSL; no Go loops faking a graph (M1/8) — `pkg/k8sgraph/lib/k8s/relations.gatwd` (Owns, Selects, RoutesTo, Configures, Mounts, Binds, inferred Calls). ✅
- [x] Third-party K8s deps only in `adapters/` + the `pkg/k8sgraph` shim (M5); client-go is confined to adapter/shim files. ✅
- [x] Edition `main.go` is assembly-only; registered pieces come from
      core/adapters/plugins/preload DSL (user rule, §0; Mandate 7). ✅
- [x] Generic `Edition.WithPreludeFS`/`WithPrelude` added to `pkg/cli` (Phase P1). ✅
- [x] Generic `cli.AdapterStep` bridges any `Adapter[any]`→DSL step (Phase P3a). ✅
- [x] K8s adapter emits structured nested-HAMT payloads, not bare names (Phase P2). ✅
- [x] Domain `k8sgraph.CrawlStep` constructs the K8s adapter → `cli.AdapterStep` (Phase P3b). ✅
- [x] godoc on every new exported identifier (verified by an AST scan of the touched files), `doc.go` for new packages, `Example*` for the main new exports (project-doc-sync). ✅
- [x] Live/offline parity proven by test (`TestLiveOfflineParity`) and every compliance rule proven by a violating fixture (`testdata/broken.yaml`) — Mandates 6/7. ✅
- [x] Generic-first: every core primitive the domain needed (`<+>`, `subset_of`, `hamt_path`, `graph_scc` as a function, `Edition.WithPreludeFS`, `cli.AdapterStep`, …) has a non-domain DSL assertion (M3). ✅

---

## 9. Appendix C — Implementation reference (signatures & file anchors)

Verified against the codebase on 2026-10-13. Lets a fresh implementation session
skip re-exploration.

**Engine core types**
- `evaluator/engine.go:14` — `type Adapter[T any] func(dsltypes.ASTNode[T]) iter.Seq[dsltypes.ASTNode[T]]`
- `evaluator/pipeline.go:29` — `type PipelineStep[T any] func(iter.Seq[dsltypes.ASTNode[T]]) iter.Seq[dsltypes.ASTNode[T]]`
- `evaluator/pipeline.go:50` — `func StepAdapter[T any](adapter Adapter[T]) PipelineStep[T]` (= `functions.Bind(stream, adapter)`)
- `evaluator/eithert.go:17` — `func EitherT[T any](adapter Adapter[T]) Adapter[functions.Either[error, T]]` (bypasses `Left`, lifts `Right`)
- `evaluator/engine.go` — `NewEngine[T any]()`, `(*Engine[T]).Register(flag dsltypes.BitFlag, adapter Adapter[T])`, `(*Engine[T]).Evaluate(stream)`, field `OnMissingAdapter func(BitFlag) Adapter[T]`
- `dsltypes/types.go:40` — `type ASTNode[T any] struct { Value T; DSLType BitFlag; ... }`; `:56` — `func NewASTNode[T any](val T, dslType BitFlag) ASTNode[T]`

**Interpreter / registration (`pkg/cli/cli.go`)**
- `:31` — `type StepHandler func(stepStr string, interp *Interpreter) (evaluator.PipelineStep[functions.Either[error, any]], bool)`
- `:46` — `Interpreter.CustomSteps []StepHandler`; `:278` custom-step dispatch loop in `parsePipelineSteps`
- `:54` `NewBareInterpreter()`, `:74` `NewInterpreter()`, `:81` `RegisterStep`, `:2373` `RegisterConstant`, `:2386` `RegisterType(definition string, factory func(string) any)`
- `:2567` `EvaluateScript(script string)`, `:2572` `LoadScriptContent(content, baseDir)`, `:2647` `LoadScript(filename)`, `:2659` `LoadScriptFS(fsys fs.FS, filename string) error`
- `EvaluateToValues(line string) []any` — runs one pipeline line, returns unwrapped Right values (what tests assert on)
- `RunWithInterpreter(*Interpreter)` / `Run()` — REPL/script driver (also handles `--check`)

**Edition / Appliance / Steps (`pkg/cli/`)**
- `edition.go` — `NewEdition(name).Bare()?.WithType(name, func(string) any).WithConstant(name, any).WithStep(StepHandler).WithPrelude(script).WithPreludeFS(fsys, path).BuildInterpreter()/.Run()` (preludes run last, in order)
- `steps.go` — `MapStep(name, func(any) any)`, `FilterStep(name, func(any) bool)`, `ArgStep(prefix, func(arg string, val any)(any,error))`, `SinkStep(name, func(any))`, `AdapterStep(name, evaluator.Adapter[any])`
- `appliance.go` — `NewAppliance(WithEmbeddedFS(fsys,path)|WithEmbeddedScript(str)|WithConfigMapOverride(path)|WithEdition(e)|WithStep|WithType|WithConstant)`, `(*Appliance).Reload()` (atomic swap), `.Execute(line) []any`

**Plugins (`pkg/plugins/registry.go`)**
- `type AdapterFactory func() evaluator.Adapter[functions.Either[error, any]]`; `Register(keyword string, factory AdapterFactory)`; `var Registry map[string]AdapterFactory` — consumed in `pkg/cli/cli.go` pipeline-step fallthrough via `StepAdapter(factory())`. (Nullary → no config injection; P3a's `AdapterStep` is the config-carrying alternative.)

**Kubernetes adapter (`adapters/kubernetes_adapter.go`)**
- `func NativeKubernetesAdapter(client dynamic.Interface, gvr schema.GroupVersionResource, targetFlag dsltypes.BitFlag) evaluator.Adapter[string]` — legacy (yields resource *names*, `Adapter[string]`; kept for back-compat).
- `func NativeKubernetesResourceAdapter(client dynamic.Interface, gvr schema.GroupVersionResource) evaluator.Adapter[any]` — **P2 ✅** (nested-HAMT payloads; `unstructuredToValue` object→HAMT / array→List / scalar; List error→node→`Left` via EitherT).
- `func EnumerateResources(disco discovery.DiscoveryInterface) ([]schema.GroupVersionResource, error)` — **P2 ✅** (GVR/CRD discovery; skips subresources + non-listable).
- Raw crawl wiring template: `examples/kubernetes_repl/main.go`; fake-client test: `adapters/kubernetes_adapter_test.go` (`fake.NewSimpleDynamicClient`).

**Graph core (Phases 1–2, verified 2026-10-13)**
- `functions/graph_node.go` — `Node[K,C,N]`, `Edge[K,C,E]` (color generic `C comparable`; `Group`/`Inferred`; `Equal`/`Compare`).
- `collections/graph.go` — `Graph[K,C,N,E]` (3 HAMT indexes); construct/project/Functor/Foldable/Filterable/Monoid; `WalkFrom`/`BFS`/`DFS`/`Reachable`; `WalkStrategy{WalkBFS,WalkDFS}`.
- `collections/graph_walk.go` — `FoldWalk`, `ShortestPath(from,to)([]K,float64,bool)`, `TopoSort() functions.Either[[]K,[]K]`, `SCC() [][]K`.
- `comonads/graph_cursor.go` — `GraphCursor[K,C,N,E]`, `NewGraphCursor`, `Extract`/`To`/`Successors`/`Predecessors`/`OutEdges`/`InEdges`/`Extend`; free `ExtendGraph[K,C,N,R,E]`, `DuplicateGraph[K,C,N,E]`.
- `optics/traversal.go` — generic `Traversal[S,A]{GetAll,Modify}`, `ComposeTraversals`, `(Traversal).Set`.
- `optics/graph_optics.go` — `NodePayloadLens[K,C,N,E](id)`, `ColorNodePrism[K,C,N,E](id,color)`, `AllEdgeWeights[K,C,N,E]()`, `OutEdgeWeights[K,C,N,E](id)`.
- `transducers/graph_walk.go` — `WalkNodes[K,C,N,E](g,start,strategy)`, `TransduceWalk[K,C,N,E,B](g,start,strategy,xf)`.

**DSL grammar (`docs/grammar.md`, parser in `parser/`)**
- `TypeDef ::= "type" Ident "=" (SumType | RecordType | TupleType | CollectionType | NativeType)`;
  `CollectionType ::= ("Set("|"List("|"Tree("|"BST("|"BTree("|"RBTree(") Ident ")"` — **add `Graph(`** here for Phase 4.
- `Pipeline ::= SpawnNode {("|>"|"||>") PipelineStep}`; `SpawnNode ::= "spawn" Ident ":" Ident`
- Lambdas `\x -> …`, `match { Ctor -> … | _ -> … }` (constructor matching = Workstream A, done), comprehensions `[e | p <- xs, guard]`.
- HAMT builtins (payload navigation): `to_hamt`, `hamt_get`, `hamt_set`, `hamt_del`, `hamt_values`, `hamt_path` (**new, Phase-6 prereq**).

**Deep field access — the "functional struct" accessor (Phase-6 prereq, ✅ 2026-10-13).**
Resources are open-schema, so payloads are nested HAMTs (a product over a *dynamic*
key set), not fixed-schema records/Go structs. The functional equivalent of
`pod.Spec.Selector.App` over that open payload is a **composed optic**, implemented
two ways:
- `optics.HAMTPathLens(keys ...string) Lens[*HAMT[string,any], Maybe[any]]`
  (`optics/lens.go`) — composes one `HAMTLens` per segment; Get is an Optional
  "preview" (`Maybe`), Set rebuilds only the spine along the path (structural
  sharing, so a focused `spec.selector.app` rewrite copies a path, not the whole
  resource). This is the reusable primitive for the live→kustomize rewrite (§6 #10).
- DSL `hamt_path "spec.selector.app"` → `evaluator.StepHAMTPath` — the dynamic,
  multi-carrier (HAMT / Map / `map[string]any` Record) preview: walks all levels in
  one step, erroring (naming the failing segment + path) on a missing/non-navigable
  segment instead of silently yielding nil. Tests: `evaluator.TestStepHAMTPath`
  (nested/missing/scalar/record/empty), `ExampleStepHAMTPath`,
  `optics.ExampleHAMTPathLens`, DSL §13b (single-level over HAMT + Record; deep
  nesting asserted in Go since the DSL can't yet build a nested HAMT). The
  relation predicates (§7, Phase 6) use `hamt_path` instead of chained `hamt_get`.

**Edition template files** — `examples/custom_cli/main.go`, `examples/game_of_life_edition/main.go` (Edition); `examples/hot_reload_appliance/main.go` (Appliance + `//go:embed`); `examples/14_plugins/` (plugin→keyword). NOT a template: `cmd/gatwd-gen` (codegen tool).
