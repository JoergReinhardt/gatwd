> **COMPLETED — historical record.** Every goal below was delivered; see `docs/k8s_graph_design.md` §5 (Phase 8 entry) for what actually landed and the remaining ideas. The goal list and the "Gotchas" are kept because the gotchas are still true — they are now consolidated, with their prevention rules, in `docs/development_playbook.md` §6 and `AGENTS.md` §5a. To write the next handoff use `.agents/prompts/phase-handoff.md`. Where the text below says "multi-arg lambda `def` mis-evaluates" or "`EvaluateToValues(\"\")` panics": both are fixed (closures; empty input returns no values).

# Phase 8 Handoff — live operator source, relation matrix, perf

> Handoff for a **fresh context**. Phases 1–7 of the k8s-graph project are **COMPLETE and
> committed**. Read this, then `docs/k8s_graph_design.md` §5 (Phase 7 + 8 entries), §6
> #8/#9, §7, then start. Supersedes `docs/phase7_handoff.md` (done) and
> `docs/phase6_handoff.md` (stale).

---

## 1. Where we are (done, committed)

- **Core graph DSL row** (~41 `graph_*` steps) + expression builtins in
  `pkg/cli/dslfuncs.go`: `graph_empty`, `graph_add_node/edge/edges/inferred_edge`,
  `graph_payloads(_of)`, `graph_in_of`, `graph_out_of`, `hamt_get/path/has_path`, `subset_of`.
- **K8s I/O:** live `CrawlStep` (`k8s_pods`, `k8s_services`, ...), offline
  `ManifestStep`/`KustomizeStep` (`manifests`, `kustomize`).
- **Prelude** `pkg/k8sgraph/lib/k8s/*.gatwd`: `k8s_ingest`, relations `Owns`/`Selects`/
  `RoutesTo`, `k8s_graph = k8s_ingest |> k8s_relations`, `k8s_apply_order`, `k8s_cycles`.
  `k8sgraph.ApplyEdition(e, client)` assembles it (nil client = offline).
- **Phase 7 (this session):**
  - `cmd/gatwd-k8s/main.go` — Edition, falls back to offline when kubeconfig fails.
  - `cmd/gatwd-k8s-operator/` — Appliance: `rules/compliance.gatwd` (`compliance` =
    `UnselectedPod` + `IdleDeployment`), flags `-manifests -rules -interval`,
    `newAppliance`/`checkPipeline` helpers, tests incl. ConfigMap hot-reload.
  - E2E tests over `pkg/k8sgraph/testdata/stack.yaml`: request trace
    (`graph_shortest_path web nginx-pod`) and dependency resolve (`graph_reachable`,
    `k8s_apply_order`).
  - DSL asserts through §77.

## 2. Phase 8 goals

1. **Live source for the operator.** Today the operator only evaluates
   `spawn StrVal: <manifests> |> manifests |> compliance`. A live run needs the crawl
   streams of several GVRs (`k8s_pods`, `k8s_services`, `k8s_deployments`,
   `k8s_replicasets`, `k8s_ingresses`, ...) merged into **one** resource stream feeding
   `k8s_graph`. Open design question: there is no obvious generic stream-concatenation
   step in the pipeline dispatch (`grep concat pkg/cli/cli.go` finds only per-element
   flatmap). Options, in order of preference:
   (a) a generic multi-source step (core, domain-free: concatenate N spawned streams);
   (b) a `k8s_resources` step in `pkg/k8sgraph` that composes the per-GVR
   `adapters.NativeKubernetesResourceAdapter`s into one `evaluator.Adapter[any]`
   (still no relation logic in Go);
   (c) fold ingest across several pipelines into one graph with `graph_union` / `++`
   (the graph Monoid exists; relation passes then run once on the union).
   Pick one with the user before building. Then add a fake-dynamic-client test (see
   `pkg/k8sgraph/steps_test.go` `TestCrawlStep`) proving the live path derives the same
   edges as the manifest fixture (live/offline parity, design §6 #10).
2. **Relation matrix expansion** in `lib/k8s/relations.gatwd` (append to
   `k8s_relations`): `Configures` (ConfigMap → Pods that mount/env-reference it),
   `Mounts` (PVC/volumes), `Binds` (RoleBinding → ServiceAccount), `Calls`, and
   **inferred edges** (`graph_add_inferred_edge`, weight 0.5, e.g. a ConfigMap value
   equal to a Service name). `NodeColor`/`EdgeColor` in `types.gatwd` already list them.
   Extend `testdata/stack.yaml` (or add a second fixture) and assert each in
   `pkg/k8sgraph/edition_test.go`.
3. **More compliance rules** in `cmd/gatwd-k8s-operator/rules/compliance.gatwd`
   (e.g. Service with no selected Pods, Ingress routing to a missing Service,
   dependency cycles via `graph_scc`).
4. **Perf bench** on a large synthetic cluster graph (the `Selects` pass is
   O(services × pods)); mirror `benchmarks/` style (`run_ackermann.sh`, `README.md`).
5. **Recursive dogfood** pass (Mandate 4) over the K8s shim and prelude; godoc/`doc.go`
   sync for `cmd/gatwd-k8s*` and the new builtins (`graph_in_of`/`graph_out_of` have
   DSL asserts but no Go `Example*` yet — add to `pkg/cli/dslfuncs_example_test.go`).
6. Update design doc §5 (Phase 8 entry) and memory; commit per sub-step.

## 3. Authoring shape (copy it for new relations/rules)

- **Ingest** = zero-parameter `def` fragment wrapping a `fold` with a graph accumulator.
- **Relation** = graph-to-graph pass:
  `def rel_x = fold (\_ g -> graph_add_edges [ [color, w, from, to] | a <- (graph_payloads_of "A" g), hamt_has_path "…" a, … ] g) graph_empty`
- **Rule** = comprehension over the graph yielding `[violation, resource]` pairs; use
  `graph_in_of` / `graph_out_of Color id g` `== []` as the "nothing points here" guard;
  concatenate rules with `++`.
- Edge specs are 4-element lists `["Color", weight, "from", "to"]`. `Connect` does not
  create endpoint nodes; ingest first.
- Comprehension filters are positional: guard optional fields with `hamt_has_path`
  *before* the dependent generator.

## 4. Gotchas (don't re-derive)

- Inside a list literal a function application needs parens:
  `["X", (hamt_path "a" p)]` — otherwise the args split into separate elements.
- A multi-argument lambda `def` used as `fold add_node graph_empty` mis-evaluates
  (re-parsed-text macro; deferred Workstream B). Zero-parameter fragment defs work.
- `map (\g -> …)` on a graph element is the graph Functor (payload map); use
  `fold (\_ g -> …) seed` for whole-graph passes.
- `EvaluateToValues` treats its argument as one pipeline: `def` lines in a multi-line
  string are not registered. Use `EvaluateScript` for setup.
- `EvaluateToValues("")` panics on empty input (`pkg/cli/cli.go` ~2276); worth a guard.
- A top-level `let r = <hamt> in hamt_get "kind" r` still fails (stringified substitution).
- Result lists come back as `*collections.List[any]`; use `functions.ToSliceAny`.
- Kubeconfig with an `oidc` auth provider fails `DynamicClientFromKubeconfig` in this
  environment; editions fall back to offline. Live tests must use the fake dynamic client.
- Prefer Read/Write/Edit for file work and keep shell commands short.

## 5. Constraints (hard rules)

- **Mandate 5:** no `kubernetes`/`k8s` under `functions/ dsltypes/ collections/ arrows/
  comonads/ optics/ transducers/ transformers/ evaluator/ parser/ pkg/cli/`. K8s lives in
  `adapters/`, `pkg/k8sgraph/`, `cmd/`, `.gatwd`.
- **Edition/Appliance `main.go` = pure assembly.** Relations and rules live in `.gatwd`.
- New core primitives need a DSL builtin + a non-K8s assert in
  `tests/dsl_exhaustive_test.gatwd` (currently through §77) and godoc/Example coverage.
- Prefer `go doc` / `gopls` over grep; `gofmt -w` changed files.

## 6. Verification (gate every step)

```bash
go test ./...
go vet ./...
go test -run Example ./...
./run_dsl_tests.sh
```

## 7. Pointers

- Design doc `docs/k8s_graph_design.md`: §4.4, §5 Phases 7–8, §6 #8/#9/#10, §7, Appendix C.
- Shim: `pkg/k8sgraph/` (`steps.go`, `edition.go`, `lib/k8s/*.gatwd`).
- Operator: `cmd/gatwd-k8s-operator/{main.go,main_test.go,rules/,testdata/}`.
- Appliance API: `pkg/cli/appliance.go` (`NewAppliance`, `Reload`, `Execute`).
- Core internals: `pkg/cli/dslfuncs.go`, `parser/infix.go`
  (`ListComprehensionExpr.FilterAfter`), `pkg/cli/cli.go` (`loadScriptContent`, FS imports).
- Memory: `[[k8s-graph-project]]`, `[[language-roadmap-progress]]`.
