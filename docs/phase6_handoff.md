# Phase 6 Handoff — the pure-DSL Kubernetes relation prelude

> Handoff for a **fresh context** to implement Phase 6 of the K8s-graph project.
> Written 2026-10-13 after Phases P1–P3, 1–5 (+kustomize) landed and were pushed
> to `origin/master` (HEAD `3cce4c3`). All gates green. Read this, then the design
> doc §4 + §7 + Appendix C, then start.

---

## 1. Where we are (done, pushed, green)

The **generic graph** and the **K8s I/O layer** are complete. What remains is the
*substance*: the relations, authored in **pure gatwd DSL**, plus the assembly that
ships them.

Done & pushed (`origin/master`):

- **Core graph DSL row COMPLETE** — construct / project / walk / Functor / Foldable
  / Filterable / Monoid / Lens, ~41 `graph_*` builtins, asserted in
  `tests/dsl_exhaustive_test.gatwd` §§65–72. Core has **zero** K8s.
- **`hamt_path "a.b.c"`** deep-field accessor (`evaluator.StepHAMTPath`) +
  `optics.HAMTPathLens` composed optic — the "functional struct field accessor"
  over open-schema payloads. The relation predicates use this.
- **Live K8s I/O (P2/P3b)** — `adapters.NativeKubernetesResourceAdapter(client,
  gvr)` yields each resource as a nested-HAMT payload; `adapters.EnumerateResources`
  (discovery, incl. CRDs). Shim `pkg/k8sgraph`: `CrawlStep(keyword, client, gvr)`,
  `DynamicClientFromKubeconfig`/`DiscoveryClientFromKubeconfig`/`RestConfigFromKubeconfig`
  (in-cluster → `$KUBECONFIG` → `~/.kube/config`), convenience GVRs.
- **Offline I/O (Phase 5)** — `adapters.ManifestAdapter(opts…)` reads YAML/JSON
  files/dirs **and renders kustomize libraries** via an external `kustomize build`
  (pluggable `WithRenderer`/`WithKustomizeCommand`; adapter stays kustomize-agnostic,
  no kustomize Go dep). Shim steps `ManifestStep`/`KustomizeStep`.

**All three sources converge on the identical nested-HAMT payload**, so everything
Phase 6 builds is source-agnostic (live vs. manifests vs. kustomize). The DSL
surface is already symmetric:

```
spawn StrVal: default           |> k8s_pods  |> …   -- live
spawn StrVal: ./overlays/prod   |> kustomize |> …   -- kustomize
spawn StrVal: ./manifests/x.yaml|> manifests |> …   -- raw
```

---

## 2. Phase 6 goal

Deliver, per design doc §4.2/§4.3:

1. **`pkg/k8sgraph/lib/k8s/*.gatwd`** — the relation library, authored **entirely in
   DSL** on top of the graph builtins + `hamt_path`:
   - `types.gatwd` — node/edge *colors* as DSL sum types
     (`type NodeColor = Pod | Service | …`, `type EdgeColor = Owns | Selects | …`).
   - `ingest.gatwd` — lift the crawled resource HAMT stream into typed graph nodes.
   - `relations.gatwd` — the **relation rules** as DSL predicates producing colored,
     weighted edges. **This is the heart of the tool.**
   - `queries.gatwd` — reusable walks for the use cases (dependency resolve, request
     trace, compliance folds).
   - `prelude.gatwd` — `import`s the above (single embed entry point).
2. **`pkg/k8sgraph` assembly growth** — `ApplyEdition(e *cli.Edition, client) *cli.Edition`
   that registers the crawl/kustomize steps + `WithPreludeFS(embeddedFS,
   "lib/k8s/prelude.gatwd")`, and a `//go:embed lib/k8s/*.gatwd`. (`Edition.WithPreludeFS`
   already exists — Phase P1.)
3. **Tests** — DSL tests over **fixture manifests** (no live cluster in CI) proving
   `Owns` / `Selects` / `RoutesTo` end-to-end. Prove those **three first**, then
   expand the relation matrix.

**The discipline (Mandates 1 & 8, non-negotiable):** relations are DSL predicates,
**never Go loops faking a graph**. Go only ingests raw resources (already done).

---

## 3. ⚠️ FIRST BLOCKER — graph ingest from a resource stream (resolve before relations)

**The gap:** every graph-construction builtin today takes **literal string args** from
the step string:

```go
// pkg/cli/cli.go  buildGraphStep
case "graph_add_node":                       // graph_add_node Pod nginx
    return evaluator.StepGraphAddNode(args[0], args[1]), true   // args are literals
```

So there is **no way yet** to turn a *stream/list of crawled resource HAMTs* into
graph nodes/edges — you cannot derive a node's `id`/`color`/`payload` from a HAMT
payload with the current builtins. Ingest (`… |> k8s_pods |> <build a graph>`) is
therefore the first thing to build, and per the anti-mockup mandate it must be a
**real engine primitive**, wired L3→L5, before any `.gatwd` relation is written.

**Design options** (pick one, implement it generically in `evaluator/graph_steps.go`
+ `pkg/cli/cli.go`, then use it from the prelude):

- **(A) Stream-fold ingest** — a step that consumes the upstream resource stream and
  folds it into one graph, deriving each node via lambdas:
  `… |> k8s_pods |> graph_build_nodes (\r -> hamt_path kind r) (\r -> hamt_path metadata.name r)`
  (color-fn, id-fn; payload = the resource HAMT). This mirrors `StepToHAMT`/`StepToList`
  (stream → single aggregate value) and reuses the lambda-compilation infra in
  `(*Interpreter).buildGraphLambdaStep` (`compileUnary`/`compilePred`/`splitParen`).
- **(B) List-fold ingest** — `… |> collect_list` then a graph fold
  `graph_from_list (\g r -> …)` where the lambda adds a node per element. Needs a
  way to call node-adding *inside* a fold lambda (today `graph_add_node` is a step,
  not a function) — i.e. expose graph construction as lambda-callable combinators.
- **(C) Reducer builtin** — generalize the existing `fold` so its accumulator can be
  a graph and the reducer a DSL lambda returning the updated graph.

**(A) is the recommended shape** — smallest, matches the existing stream-aggregate
pattern, and keeps edges as a second pass (`relations.gatwd` folds the same resource
list into edges via comprehensions). Whatever you pick: **generic, domain-free**,
lives in core, asserted in `tests/dsl_exhaustive_test.gatwd` with a non-K8s example,
then reused by the prelude.

Edges have the same need: `rel_owns` must, per child, read `ownerReferences` (a
`*collections.List[any]` of HAMTs) and emit an `Owns` edge per owner. That is a
comprehension/fold producing edges folded into the graph — same ingest primitive,
edge variant (e.g. `graph_build_edges (\r -> [ (color, w, from, to) | … ])`).

---

## 4. DSL capabilities you already have (cheat sheet — don't re-derive)

Verified present in `pkg/cli/cli.go` + `tests/dsl_exhaustive_test.gatwd`:

- **Graph builtins** (~41): `graph_empty`, `graph_add_node <color> <id>`,
  `graph_add_edge <color> <w> <from> <to>`, `graph_add_edge_inferred`,
  `graph_add_fanout <color> <w> <from> <to…>`, `graph_add_fanin`, `graph_remove_*`;
  projection `graph_nodes`/`graph_edges`/`graph_out`/`graph_in`/`graph_neighbors`/
  `graph_nodes_of <color>`/`graph_edges_of <color>`/`graph_edges_of_group`/
  `graph_edges_inferred`/`graph_reachable`; walks `graph_bfs`/`graph_dfs`/
  `graph_walk_from`/`graph_shortest_path`/`graph_topo_sort`/`graph_scc`; higher-order
  `graph_map_nodes`/`graph_map_edges`/`graph_bimap`/`graph_filter_nodes`/
  `graph_filter_edges`/`graph_fold_nodes`/`graph_fold_edges`/`graph_fold_walk`;
  set-algebra `graph_union`/`graph_diff`/`graph_intersect`; Lens `graph_focus <id>`/
  `graph_set <id> <v>`/`graph_over <id> (\p -> …)`.
  Edges project as the string `from-Color->to`.
  **Binding convention:** node-lambda steps bind the node *payload*; edge-lambda
  steps bind the edge *weight*.
  **DSL color is a plain string tag** (`graph_*` pins `Graph[string,string,any,any]`),
  matching sum-type constructor names — NOT `dsltypes.Variant`.
- **Payload navigation:** `to_hamt`, `hamt_get <k>`, `hamt_path <a.b.c>` (deep, multi-
  carrier, errors on missing), `hamt_set`, `hamt_del`, `hamt_keys`, `hamt_values`,
  `hamt_pairs`, `frequencies`. Records (`map[string]any`) via `parse_record` +
  `focus`/`update`.
- **Lists & FP:** comprehensions **with multiple generators + guards + nesting**
  (`[e | p <- xs, q <- ys, guard]` — real, used in `cgol_step`/`quicksort`),
  `map`/`filter`/`reject`/`fold`/`collect_list`/`collect_set`, `++`.
- **Sets:** `to_set`, `set_contains <elem>`, `set_union`/`set_intersect`/`set_diff`.
  (Selector⊆labels can be expressed with `set_contains` over a comprehension, or add
  a `subset_of` primitive if cleaner — see §6.)
- **Binding/flow:** `let … in …`, `def name = …`, `match { Ctor -> … | [h, ...t] ->
  … | _ -> … }` (constructor matching = Workstream A, done), lambdas `\x -> …`,
  `|>`/`||>`.
- **Grammar:** `type K8sGraph = Graph(String)` parses (`Graph(` in `CollectionType`);
  sum/record/tuple/collection type defs all work.

---

## 5. Relation sketches (match to the real grammar; finalize names in impl)

From design §7, corrected to the primitives above. Each notes what it needs.

**types.gatwd**
```
type NodeColor = Pod | Service | Ingress | ConfigMap | Secret | Deployment | ReplicaSet | PVC | ServiceAccount | Role
type EdgeColor = Selects | RoutesTo | Configures | Owns | Mounts | Binds | Consumes
```

**Owns** — `child.metadata.ownerReferences[].name → child`. Needs: iterate a
`List[HAMT]` (ownerReferences) → edge per owner. (ingest §3 edge variant + comprehension.)
```
-- for each owner ref on a child, add Owns(owner -> child)
def rel_owns = \child -> \g ->
    [ (Owns, 1, owner_name, child_name)
    | owner <- (child |> hamt_path metadata.ownerReferences)
    , owner_name <- [owner |> hamt_get name]
    , child_name <- [child |> hamt_path metadata.name] ]
    -- folded into g by the edge-ingest primitive
```

**Selects** — Service whose `spec.selector ⊆ Pod.metadata.labels`; 1:n fan-out as one
grouped relation. Needs: subset test over label maps (`set_contains`/comprehension or
a `subset_of` primitive).
```
def selector_matches = \svc -> \pod ->
    (svc |> hamt_path spec.selector) `subset_of` (pod |> hamt_path metadata.labels)

def rel_selects = \svc -> \pods ->
    graph_add_fanout Selects 1 (svc_id svc) [ pod_id pod | pod <- pods, selector_matches svc pod ]
```

**RoutesTo** — Ingress `spec.rules[].http.paths[].backend.service.name → Service`.
Needs: nested `List`/`HAMT` navigation via `hamt_path` + comprehension.

**Consumes (inferred, §6 #9)** — CM value equals a Service name ⇒ assume the Pod
configured by that CM consumes the Service. Use `graph_add_edge_inferred` + lower
weight (`0.5`).

**queries.gatwd**
```
def deps_of      = \svc -> svc |> graph_reachable |> collect_list
def trace_request = \ing -> ing |> graph_walk_from |> collect_list
```

---

## 6. Anticipated missing primitives — implement first (anti-mockup), then author DSL

Expect to hit these. Implement each **generically, L3→L5, with a non-K8s exhaustive
assert**, before the relation that needs it:

1. **Graph ingest from a stream/list** (§3) — the big one; blocks *all* of ingest.
2. **`subset_of` / map-subset test** — selector⊆labels. May be composable from
   `set_contains` + comprehension; add a builtin only if the DSL expression is ugly.
   (Decide, don't fake.)
3. **Iterating `List[HAMT]` payloads in comprehensions** — confirm a `*collections.List[any]`
   from `hamt_path metadata.ownerReferences` is a valid comprehension source
   (`x <- thatList`). If not, that's a primitive to fix.
4. **Edge/tuple construction in DSL** — the ingest edge variant needs the lambda to
   yield `(color, weight, from, to)`; confirm tuple/list literals carry through, or
   add an edge-spec builtin.

For anything genuinely missing: implement the true functional abstraction in the
engine (`evaluator/`, `pkg/cli/`), wire to the DSL, assert it — **do not** hack
around it in Go or hardcode. This is the mandate the whole project is graded on.

---

## 7. Constraints (hard rules)

- **Mandate 5 — Clean Core.** The strings `kubernetes`/`k8s` must NOT appear under
  `functions/`, `dsltypes/`, `collections/`, `arrows/`, `comonads/`, `optics/`,
  `transducers/`, `transformers/`, `evaluator/`, `parser/`, `pkg/cli/`. Any new
  ingest/subset primitive is generic and domain-free. K8s lives only in `adapters/`,
  `pkg/k8sgraph/`, and the `.gatwd` prelude.
- **Edition `main.go` = pure assembly** (user rule, design §0). Everything it
  registers comes from core/adapters/plugins/preload DSL. No K8s algorithm, loop,
  relation rule, or type authored inline in Go. Relations live in `.gatwd`.
- **All-Layers-Through (M3)** + **exhaustive matrix (M2)** — new core primitives get
  a DSL builtin + an assert in `tests/dsl_exhaustive_test.gatwd`.
- **Docs (project-doc-sync)** — godoc + `doc.go` + `Example*` on all new exports.

---

## 8. Verification (gate every step)

```bash
go test ./...          # native
go vet ./...           # vet + no import cycles
go test -run Example ./...
./run_dsl_tests.sh     # exhaustive DSL suite
```

Phase-6 tests additionally run the prelude over **fixture manifests** (commit small
YAML fixtures; drive via `ManifestStep`/`KustomizeStep` through a `cli.Interpreter`,
as `pkg/k8sgraph/steps_test.go` already does) and assert the produced edges, e.g.
`… |> manifests |> <ingest> |> <rel_owns> |> graph_edges_of Owns |> collect_list`.

---

## 9. Pointers

- **Design doc** `docs/k8s_graph_design.md`: §4.2 (prelude lib layout), §4.3
  (`ApplyEdition`/`CrawlStep` shim), §6 #8 (cardinality: grouped binary edges +
  reification, *no* hyperedges), §6 #9 (inferred edges), §7 (relation DSL sketches),
  Appendix C (verified signatures & file anchors — read before coding).
- **Shim to grow:** `pkg/k8sgraph/steps.go` (+ new `edition.go` for `ApplyEdition`,
  `embed.go` for `//go:embed lib/k8s/*.gatwd`). Templates:
  `examples/custom_cli`, `examples/game_of_life_edition` (Edition builder),
  `examples/hot_reload_appliance` (Appliance + `//go:embed`).
- **Lambda-compilation infra** to extend for ingest:
  `(*Interpreter).buildGraphLambdaStep` + `compileUnary`/`compilePred`/`compileBinary`/
  `splitParen`/`resolveGraphOperand` in `pkg/cli/cli.go`; L3 step shapes
  `graphMap`/`graphUnfold` in `evaluator/graph_steps.go`.
- **Memory:** `[[k8s-graph-project]]` (full phase-by-phase log + commit hashes).

---

## 10. First concrete steps (ordered)

1. Re-read design §4.2/§4.3/§7 + Appendix C.
2. **Design & implement the ingest primitive** (§3, option A recommended): L3
   `evaluator/graph_steps.go` + L5 `pkg/cli/cli.go` + non-K8s assert in
   `tests/dsl_exhaustive_test.gatwd`. Gate green.
3. Confirm/implement `subset_of` + `List[HAMT]` comprehension iteration (§6.2/6.3)
   with non-K8s asserts, only as needed by the relations.
4. Create `pkg/k8sgraph/lib/k8s/{types,ingest,relations,queries,prelude}.gatwd`;
   author `Owns` first, prove it over a fixture manifest.
5. Add `ApplyEdition` + `//go:embed`; wire `WithPreludeFS`. Prove `Owns`/`Selects`/
   `RoutesTo` end-to-end over fixtures.
6. Update design doc (Phase 6 ✅ + checklist), memory, commit per sub-step, push.
7. Phase 7 (`cmd/gatwd-k8s` Edition + `cmd/gatwd-k8s-operator` Appliance) comes next.
