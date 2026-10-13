> **SUPERSEDED — historical record.** Phase 7 is complete (design doc §5, Phase 7). Superseded by `docs/phase8_handoff.md`, itself now complete.

# Phase 7 Handoff — K8s graph Edition + Appliance

> Handoff for a **fresh context**. Phase 6 (the pure-DSL relation prelude +
> `ApplyEdition`) is **COMPLETE and committed**. Read this, then design doc
> `docs/k8s_graph_design.md` §4.4 / §5 (Phase 6 + 7 entries) / §7, then start.
> Supersedes `docs/phase6_handoff.md`, which is stale (it still describes the ingest
> work as uncommitted) and should be deleted.

---

## 1. Where we are (done, committed)

- **Core graph DSL row** (~41 `graph_*` steps), **K8s I/O** (live `CrawlStep`, offline
  `ManifestStep`/`KustomizeStep`), **ingest primitive** (infix function application +
  `fold` with a graph accumulator) — all earlier phases.
- **Phase 6 (this session):** `pkg/k8sgraph/lib/k8s/{types,ingest,relations,queries,prelude}.gatwd`,
  `pkg/k8sgraph/edition.go` (`ApplyEdition(e, client)`, `//go:embed lib/k8s/*.gatwd`,
  `PreludeFS()`, `PreludePath`), fixture `pkg/k8sgraph/testdata/stack.yaml`, tests in
  `pkg/k8sgraph/edition_test.go`. `Owns`, `Selects`, `RoutesTo` are proven end-to-end:

  ```
  spawn StrVal: <path> |> manifests |> k8s_graph |> graph_edges_of Owns |> collect_list
  ```

  `ApplyEdition` with a **nil client** builds an offline edition (`manifests` +
  `kustomize` + prelude); with a client it also registers `k8s_pods`, `k8s_services`,
  `k8s_configmaps`, `k8s_secrets`, `k8s_endpoints`, `k8s_deployments`,
  `k8s_replicasets`, `k8s_ingresses`.

## 2. The authoring shape (copy it for new relations)

- **Ingest** = zero-parameter `def` fragment wrapping a `fold` with a graph accumulator:
  `def k8s_ingest = fold (\g r -> graph_add_node (hamt_get "kind" r) (hamt_path "metadata.name" r) r g) graph_empty`
- **Relation** = graph-to-graph pass. A graph stream element reaches the reducer as the
  second argument, so the whole graph is readable:
  `def rel_x = fold (\_ g -> graph_add_edges [ [color, w, from, to] | a <- (graph_payloads_of "A" g), hamt_has_path "…" a, … ] g) graph_empty`
  The reducer returns `g` grown by the edges, so passes chain with `|>`.
- **Edge specs** are 4-element lists `["Color", weight, "from", "to"]` (no tuple literal).
  `Connect` does not create endpoint nodes; ingest nodes first (the relation passes run
  after `k8s_ingest`, so endpoints exist).
- **Guards:** comprehension filters are positional, so `hamt_has_path "k" x` placed
  before a dependent generator protects it. Use it for every optional field.
- Register a new relation in `relations.gatwd` and append it to `k8s_relations`.

## 3. Gotchas (don't re-derive)

- A multi-argument lambda `def` used as `fold add_node graph_empty` still mis-evaluates
  (re-parsed-text macro; deferred Workstream B). Zero-parameter fragment defs work.
- `map (\g -> …)` on a graph element is the graph Functor (payload map), not
  whole-graph application. Use `fold (\_ g -> …) seed`.
- `EvaluateToValues` treats its argument as one pipeline: `def` lines in a multi-line
  string are not registered. Use `EvaluateScript` for setup, `EvaluateToValues` for the
  query (or build the interpreter via `ApplyEdition(...).BuildInterpreter()`).
- `EvaluateToValues("")` panics on empty input (`pkg/cli/cli.go` ~2276); minor, worth a guard.
- Result lists come back as `*collections.List[any]`; use `functions.ToSliceAny`.
- A top-level `let r = <hamt> in hamt_get "kind" r` still fails (stringified substitution).
- Tooling: Bash permission checks were flaky in the previous session on long commands and
  heredocs; prefer Read/Write/Edit for file work and keep shell commands short.

## 4. Phase 7 goal (design §4.4)

1. **`cmd/gatwd-k8s/main.go` — Edition, pure assembly:**
   ```go
   client, _ := k8sgraph.DynamicClientFromKubeconfig()
   k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s"), client).Run()
   ```
   Handle a missing kubeconfig gracefully (fall back to `nil` client = offline edition).
   No K8s algorithm, loop or relation inline in Go (user rule, design §0).
2. **`cmd/gatwd-k8s-operator/main.go` — Appliance** (template:
   `examples/hot_reload_appliance`): `cli.NewAppliance(cli.WithEdition(base),
   cli.WithEmbeddedScript(rules) or WithEmbeddedFS, cli.WithConfigMapOverride("/etc/gatwd/rules.gatwd"))`
   with embedded **use-case rules** under `cmd/gatwd-k8s-operator/rules/*.gatwd`
   (e.g. a compliance check: Pods not selected by any Service, Deployments with no
   owned Pods). Check `Appliance.Reload` / `Execute` in `pkg/cli/appliance.go` for how
   the baseline script and the hot-reload watch are driven.
3. **E2E demos over fixtures** for the two use cases: dependency resolve
   (`graph_reachable`) and request trace (Ingress → Service → Pod via
   `RoutesTo`/`Selects`). Add them as tests (fixture YAML, no live cluster).
4. Update design doc §5 (Phase 7 entry), memory, and commit per sub-step.

## 5. After Phase 7 — relation matrix expansion (Phase 8 candidates)

`Configures` (ConfigMap → Pods that mount/env-reference it), `Mounts` (PVC/volumes),
`Binds` (Role bindings → ServiceAccounts), `Calls`, and **inferred edges**
(`graph_add_inferred_edge`, weight 0.5, e.g. a ConfigMap value equal to a Service
name). `NodeColor`/`EdgeColor` in `types.gatwd` already list them. Also: perf bench on a
large synthetic cluster graph (the Selects pass is O(services × pods)).

## 6. Constraints (hard rules)

- **Mandate 5:** no `kubernetes`/`k8s` under `functions/ dsltypes/ collections/ arrows/
  comonads/ optics/ transducers/ transformers/ evaluator/ parser/ pkg/cli/`. K8s lives in
  `adapters/`, `pkg/k8sgraph/`, `cmd/`, `.gatwd`.
- **Edition/Appliance `main.go` = pure assembly.** Relations live in `.gatwd`.
- New core primitives need a DSL builtin + a non-K8s assert in
  `tests/dsl_exhaustive_test.gatwd` (currently through §76) and godoc/Example coverage.
- Prefer `go doc` / `gopls` over grep; `gofmt -w` changed files.

## 7. Verification (gate every step)

```bash
go test ./...
go vet ./...
go test -run Example ./...
./run_dsl_tests.sh
```

## 8. Pointers

- Design doc `docs/k8s_graph_design.md`: §4.4 (editions/appliances), §5 Phase 6 + 7,
  §6 #8/#9 (cardinality, inferred edges), §7 (relation sketches), Appendix C (signatures).
- Shim: `pkg/k8sgraph/` (`steps.go`, `edition.go`, `lib/k8s/*.gatwd`).
- Templates: `examples/custom_cli`, `examples/game_of_life_edition`,
  `examples/hot_reload_appliance`, `examples/kubernetes_repl`.
- Core internals added for the prelude: `pkg/cli/dslfuncs.go` (`graph_payloads`,
  `graph_payloads_of`, `hamt_has_path`, …), `parser/infix.go`
  (`ListComprehensionExpr.FilterAfter`), `pkg/cli/cli.go` (`loadScriptContent`, FS-aware
  imports).
- Memory: `[[k8s-graph-project]]`, `[[language-roadmap-progress]]`.
