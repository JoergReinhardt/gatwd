# Design & Implementation Plan: Manifest ⊆ Live Parity Check (read-only "is it up to date?" verb)

> Status: **IMPLEMENTED — Phases 1, 1b, 2, 3 all done (2026-10-14).** §6 #1–#3 were
> decided as recommended (non-injective list matching kept, verb placed in
> `pkg/k8sgraph`, `graph_payload_at` added to core). One thing this plan's sketches
> got wrong, found only by running them: **the interpreter's task-closure path
> (`defClosure`/`ParseInfixExpr`, `pkg/cli/closures.go`) does not reliably resolve a
> lambda whose body chains more than one `|>`, or that calls another `|>`-bodied def
> by name** — `k8s_parity_task` as originally sketched (§5.1, a two-positional-arg
> lambda task) never got past "unknown step or command" / "undefined identifier"
> errors, for reasons unrelated to this feature's own logic. The working fix: no
> `k8s_parity_task` def at all; `pkg/k8sgraph/parity_command.go`'s `ParityCommand`
> builds a top-level `let m = (... |> manifests |> k8s_graph) in let l = (... |>
> k8s_resources |> k8s_graph) in k8s_parity_report m l` string from the command's own
> flags and runs it via `RunOptions.Eval`, not `RunOptions.Task`/`cli.TaskCommand` —
> a top-level `let` with two independent `|>` chains IS reliably resolved (it is the
> exact shape `pkg/k8sgraph/parity_test.go` uses throughout). A second, smaller
> surprise: a bare `def NAME = [a, b, c]` (no lambda) is a pipeline fragment in this
> language, not a reusable constant — `const K8S_RUNTIME_METADATA_FIELDS = [...]` is
> what actually works. Neither of these is a defect in the design itself; both are
> pre-existing interpreter behavior this plan's own "verify before implementing"
> principle (§9's note on Appendix C) was written to catch, and did. Grounded against
> the actual codebase on 2026-10-14. Follows the format and conventions of
> [`docs/k8s_graph_design.md`](k8s_graph_design.md) (that document's Phases 1–8 are
> DONE and are the substrate this plan builds on).
> Scope: (1) a read-only CLI verb that compares a subset of Kubernetes resources
> described by YAML/kustomize manifests against the same resources crawled live from
> a cluster, answering "is everything in this manifest subset present and
> up to date live?" — never writing to the cluster. (2) is noted as an **explicitly
> out-of-scope, lower-priority, separate design** (live→YAML rendering and a
> kustomize-overlay-with-generator-template layout from a live scrape) — see §1.2.

---

## 0. Guiding principles (the 8 mandates applied)

- **Mandate 5 (Clean Core)** is the spine, same as the base design. The generic
  containment primitive — "does every field I declared hold, possibly with more
  present, at every level" — is domain-blind math (a lattice/preorder on partially
  known records) and belongs in `evaluator/` (Layer 3) and `pkg/cli` (Layer 5). No
  Kubernetes identifier or knowledge may appear there. The *k8s-specific* knowledge
  — which fields are server-managed noise (`status`, `metadata.resourceVersion`,
  …) — belongs in the `pkg/k8sgraph` shim's `.gatwd` prelude, not core.
- **Mandate 2 (Operator-Type Matrix)**: `subset_of` already has a full cell across
  Record/HAMT/Map and Set/collection carriers (step form, function form, DSL
  exhaustive test, structural Go test). This plan does not add a new column; it
  makes the **existing** cell mathematically correct (recursive, not flat) — the
  "definition of done" in §3 is the same list, applied to a bug fix rather than a
  new operator.
- **Mandate 1 & 7 (Anti-Mockup / Pure Assembly)**: the comparison logic (what
  "stable view" strips, how a report is phrased) is pure `.gatwd`; Go changes are
  confined to making `evaluator.IsSubsetValue` recursive (a correctness fix to an
  existing generic primitive) and to the zero-logic `cli.TaskCommand` wiring.
  Nothing here talks to the cluster in a new way — it reuses the live/offline
  sources `k8s_graph`/`k8s_resources`/`manifests` that Phases 5–8 of the base
  design already built and proved at parity (`TestLiveOfflineParity`).
- **Mandate 6 (Fail Loud)**: a resource declared in the manifest subset but absent
  live is a **reported finding**, not a silently-empty result; the negative
  fixture (§5, Phase 3) is a manifest that is deliberately stale against its live
  fixture, so "no drift" and "comparison didn't run" cannot be confused.
- **Mandate 8 (Phase-Gated, Decision-Recorded Delivery)**: three small phases
  (§5), each independently gated by `go test ./... && go vet ./... &&
  ./run_dsl_tests.sh`; §6 records the decision points this plan deliberately does
  **not** resolve by assumption (list-element matching semantics, verb placement).

---

## 1. Problem statement & use cases

### 1.1 In scope: read-only manifest ⊆ live parity (this plan)

A user has a YAML/kustomize manifest subset (could be one overlay, could be the
whole repo) and a live cluster reachable via their local kubeconfig. They want to
know, **without writing anything**:

1. For every resource the manifest subset declares: does a same-identity resource
   (`Kind/namespace/name`, per `pkg/k8sgraph/lib/k8s/ids.gatwd`) exist live?
2. If it exists, is every field the manifest declared actually present, with the
   declared value, in the live object? (Not: is the live object *identical* to the
   manifest — the live object is always allowed to have *more*: server defaults,
   injected sidecars, controller-added labels, a `kubectl.kubernetes.io/
   last-applied-configuration` annotation, …)

| Use case | What answers it |
| :-- | :-- |
| "Did my last `kubectl apply`/GitOps sync actually land?" | per-resource present? + up-to-date? |
| "Is this overlay's `replicas: 3` still what's running, or did someone `kubectl scale`?" | up-to-date? (a scalar mismatch surfaces by name) |
| "Which of these manifests were never applied at all?" | present? == false |
| "Drift audit before a GitOps reconcile" | the whole report, read-only, safe to run anytime |

This is **strictly a read**: both sides of the comparison (`manifests`,
`k8s_resources`/`k8s_graph`) are existing read-only sources (§2.3); nothing in this
plan touches `k8s_apply`, `k8s_rollout_*` or `k8s_scale`, and no new capability is
needed — reads in this codebase already require none (`pkg/cli/policy.go:85`, "reading
the rollout status is a read and needs no capability"; the write family is the only
one gated by `CapClusterApply`/`CapClusterRollout`).

### 1.2 Out of scope (explicitly, lower priority, separate design)

The user also wants, eventually: (a) rendering a live object back out as YAML, (b)
generating a kustomize-overlay-with-generator-template (`configMapGenerator`/
`secretGenerator`-style) layout from a live scrape, and (c) applying graph-derived
objects (from YAML transforms and/or a live scrape) back to a cluster.

Most of the building blocks for (c) **already exist and are proven**:
`k8s_graph |> k8s_apply_objects |> k8s_apply` (topologically-ordered, server-side
apply, gated by `CapClusterApply`, `--out DIR` for an offline dry-run render —
`pkg/k8sgraph/doc.go:56-69`, `pkg/k8sgraph/write.go`), and `pkg/cli/dslfuncs_yaml.go`
already registers `yaml_parse`/`yaml_parse_all`/`yaml_encode`/`yaml_encode_all` as
generic, domain-free builtins (a nested-HAMT payload round-trips to YAML today).
What is genuinely missing, and genuinely **not** designed here, is the **policy**
piece of (b): the directory/file layout, which fields count as "cluster noise" to
strip before emitting YAML, and whether a ConfigMap/Secret becomes a
`configMapGenerator`/`secretGenerator` stanza or a static resource file. That is a
real design question with no obviously-right answer (different teams want
different overlay conventions) and is **out of scope for this plan's phases**. It
is flagged here so a future design doc for it starts from this plan's §2 and §4,
not from zero.

---

## 2. Grounding: what the engine already has (verified 2026-10-14)

### 2.1 The generic containment primitive — `subset_of` / `IsSubsetValue`

`evaluator/pipeline.go:648-700`:

```go
// IsSubsetValue reports whether sub ⊆ super, generically and domain-free, across the
// two families where containment is well-defined:
//   - Record/map-like (HAMT, Map, map[string]any): sub ⊆ super iff every (key,value)
//     of sub is present in super with a structurally-equal value.
//   - Set/collection-like (Set, List, Chunk, any ToSliceAny-able value): sub ⊆ super
//     iff every element of sub is a member of super (by structural equality).
func IsSubsetValue(sub, super any) (bool, error)
```

Read in full (lines 661-700), the **current** behavior is:

- Map-like branch (`rangeMapLike`, line 626, dispatches over `*collections.HAMT`,
  `*collections.Map`, `map[string]any`): for every `(k, v)` in `sub`, look up `v`'s
  counterpart in `super` via `HAMTGetKey` (line 1653) and compare with
  `functions.StructuralEqual` (`functions/operators.go:1071`). **If the matched
  value is itself map-like on both sides, this still calls `StructuralEqual`, not
  `IsSubsetValue`** — i.e. a nested record must match **exactly**, not merely
  contain the declared fields. This is the bug for the K8s use case: `spec` (and
  even `metadata`) are flat-equal-compared at every nesting level once you are
  inside the top-level map.
- Collection-like branch (`functions.ToSliceAny`, `functions/operators.go:880`):
  for every element `e` of `sub`, is there some `s` in `super` with
  `functions.StructuralEqual(e, s)`? **Same issue inside list elements** — a Pod's
  `containers` list, `ports` list, etc. are structurally-equal-compared per
  element, so a server-added field anywhere inside a list element (e.g.
  `imagePullPolicy: Always` added to a container) breaks containment.
- Empty `sub` is vacuously `⊆` anything (both branches); a type mismatch (map-like
  vs not) is a `fmt.Errorf`, not a silent `false` (Mandate 6 already honored here).

Registered as the step `StepSubsetOf` (`evaluator/pipeline.go:706-730`, `subject |>
subset_of other`) and as the expression/function form in `pkg/cli/dslfuncs.go:340-345`:

```go
// subset_of sub super → bool: is sub ⊆ super? ...
m["subset_of"] = functions.DSLFunc(func(args []any) (any, error) {
    if len(args) != 2 { ... }
    return evaluator.IsSubsetValue(args[0], args[1])
})
```

Both forms are exercised in `tests/dsl_exhaustive_test.gatwd` §75 (lines 386-397:
Set membership, HAMT containment, and the function form used inside a `fold`) and
structurally in `evaluator/subset_test.go` (`TestIsSubsetValue`, 8 cases: HAMT
subset/value-mismatch/missing-key/empty, Set subset/not-subset/empty, type
mismatch). **None of the existing test cases nest a map inside a map, or a map
inside a list element** — every case is one level deep. This matters for risk
assessment (§4, §6 #1): making the recursion real does not contradict any pinned
test, but the matrix of "value is itself map-like/collection-like on one or both
sides" has no existing coverage to protect against regressing a case no one wrote
down yet, so Phase 1's own tests become the first exhaustive pin for this shape.

Already a production relation predicate, not just a test fixture:
`pkg/k8sgraph/lib/k8s/relations.gatwd:46` —

```
, subset_of (hamt_path "spec.selector" s) (hamt_path "metadata.labels" p) ]
```

`Selects` compares a Service's `spec.selector` (a flat string-to-string map) against
a Pod's `metadata.labels` (also flat) — this is exactly the one-level-deep case the
current flat implementation already handles correctly, which is why `Selects` has
never surfaced the recursion bug. The bug only shows up once you compare `spec`
(or `metadata`) *wholesale*, which no existing relation does — they all compare
single sub-fields (`hamt_path`-selected leaves), never the resource record itself
against its live counterpart.

### 2.2 Why flat equality is wrong for manifest-vs-live (restated precisely)

- **`status`** is a clean, well-known, server-only subresource. Safe to exclude
  wholesale — no manifest the user wrote should declare it, and recursion would
  not help here: dropping the key before comparing is the right move, not making
  the comparison smarter about it.
- **`metadata`** mixes user-declared fields (`name`, `namespace`, `labels`,
  `annotations`) with server-managed ones (`resourceVersion`, `uid`,
  `generation`, `managedFields`, `creationTimestamp`, `deletionTimestamp`,
  `selfLink`) as **plain siblings at the same level** — `kubernetes`'s own
  `ObjectMeta` has no structural separation between "you set this" and "the API
  server owns this." A manifest will generally not declare the server-managed
  keys at all, so the current flat-equal-on-nested-maps bug does not actually
  bite `metadata` *today* (an absent key in `sub` is never compared) — **except**
  for the one real risk: if the manifest's `metadata` is compared as the whole
  value of the `metadata` key (not merged/recursed), `StructuralEqual` between the
  two `metadata` HAMTs would require them to have the exact same key set, which
  they never will. This is k8s domain knowledge either way (which fields are
  server-managed), so it is handled by an explicit, small, named exclude-list in
  the Phase 2 shim, not solved generically.
- **`spec`** is the real problem: it is pervasively, recursively, additively
  defaulted by the API server and admission webhooks **at every nesting level,
  including inside list elements** — a Pod's `containers[i]` gains
  `imagePullPolicy`, `terminationMessagePath`, `terminationMessagePolicy`, and
  each `ports[j]` gains `protocol: TCP`, purely as examples. No static
  field-exclude list is tractable (the defaulting surface changes across API
  versions, admission webhooks, and third-party mutating webhooks) or
  future-proof. **Recursive subset containment is the only generically correct
  mechanism**: "every field I declared is present and equal, live may have more,
  at every level, including inside list elements." It also auto-tolerates
  `kubectl.kubernetes.io/last-applied-configuration`, Helm/kustomize/controller
  labels, and any other addition, for free — those keys simply never appear in
  `sub`, so they are never inspected.
- **Residual limitation, recorded as deferred, not solved here**: the API server
  sometimes **canonicalizes/rewrites** a value the manifest *did* set — e.g. a
  resource `Quantity` string normalizes ("1" the manifest wrote may read back as
  "1" or "1000m" depending on the field and version), and certain fields get
  reformatted (durations, IP ranges). Recursive subset containment compares by
  `functions.StructuralEqual`, which is exact value equality — it cannot paper
  over a value the server legitimately rewrote. This is an explicit open question
  in §6, not addressed by this plan; a false "drift" report on a canonicalized
  quantity is a known, accepted limitation to document in the verb's `--help`/docs.

### 2.3 Live and offline sources — both already exist, already proven source-agnostic

- Offline: `spawn StrVal: <path> |> manifests |> k8s_graph` (`manifests` =
  `ManifestStep`, auto-renders a kustomization root in-process — `pkg/k8sgraph/
  doc.go:41-52`).
- Live: `spawn StrVal: <namespace> |> k8s_resources |> k8s_graph` (`k8s_resources`
  = the `<+>`-merge of all twelve `CoreCrawls` crawl steps, `pkg/k8sgraph/lib/k8s/
  sources.gatwd`, `pkg/k8sgraph/steps.go:43-60`), using the client built by
  `k8sgraph.LazyDynamicClient()` (`pkg/k8sgraph/lazy.go:14-26` — reads
  `$KUBECONFIG`/`~/.kube/config` lazily, on first crawl, never for `version`/
  `completion`/offline work).
- Both paths decode to the **same nested-HAMT payload shape** (`unstructuredToValue`
  in `adapters/kubernetes_adapter.go`, shared by `NativeKubernetesResourceAdapter`
  and `ManifestAdapter`), node ids are the same `Kind/namespace/name` scheme
  (`pkg/k8sgraph/lib/k8s/ids.gatwd:1-37`), and `TestLiveOfflineParity`
  (`pkg/k8sgraph/live_parity_test.go:17-40`) already proves the *same* prelude
  (`k8s_graph`, `graph_nodes`, `graph_edges`, `graph_edges_inferred`) returns
  identical results over a fixture read both ways via
  `pkg/k8sgraph/k8sgraphtest.FakeClusterFromFile` (a fake dynamic client built
  from the same YAML fixture `manifests` reads). **This plan's Phase 3 test reuses
  exactly this fixture/fake-client pattern** — no new test infrastructure needed.

### 2.4 The read side's existing projection/query vocabulary

- `graph_has_node id g` — does the graph have this node? (used throughout
  `relations.gatwd` to keep edges joining only existing nodes).
- `graph_payloads g` / `graph_payloads_of color g` — the node payload HAMTs, by
  color (`pkg/cli/dslfuncs.go:238-258`), already the comprehension idiom every
  relation in `relations.gatwd` uses.
- `k8s_id_of r`, `k8s_scope r`, `k8s_ref kind ns name g` (`pkg/k8sgraph/lib/k8s/
  ids.gatwd`) — identity and reference resolution, reused as-is (the comparison
  needs no new identity scheme: "same resource" already means "same
  `Kind/namespace/name` id").
- `hamt_del k m`, `hamt_set k v m`, `hamt_get_or default k m`
  (`pkg/cli/dslfuncs_matrix.go:454-490`, each data-last/curried) — the existing
  projection primitives Phase 2's `k8s_strip_runtime` composes from; no new HAMT
  primitive is needed.
- `cli.TaskCommand(edition, cmdName, short, taskName, knownDefines...)`
  (`pkg/cli/verb_command.go:55-100`) — the existing, generic, zero-logic
  "promote an Edition's own prelude `def` to a cobra subcommand" mechanism; this
  is the entire Phase 3 CLI-surface wiring, reused exactly as `docs/
  development_playbook.md` §5b describes it ("one Go function, reused at every
  tier, never a bespoke RunE per verb").

---

## 3. Phase 1 — recursive `IsSubsetValue` (CORE, `evaluator/`, domain-blind)

### 3.1 The change

`evaluator/pipeline.go:661-700`. Two call sites change from `functions.
StructuralEqual` to a recursive call, with a scalar/type-mismatch fallback to the
existing exact-equality behavior:

- **Map-like branch** (currently line 668: `!functions.StructuralEqual(v, sv)`):
  when the sub-value `v` *and* its counterpart `sv` are **both** map-like
  (`rangeMapLike` succeeds on both) or **both** collection-like
  (`functions.ToSliceAny` succeeds on both), recurse via `IsSubsetValue(v, sv)`
  instead of flat equality. When they are neither (both scalars, or a type
  mismatch between the two), keep `functions.StructuralEqual` — this is the
  correct behavior for a genuine leaf (`"nginx" == "nginx"`) and for a value that
  does not recurse (comparing a map to a scalar is a mismatch, which
  `StructuralEqual` already reports as unequal, not an error — consistent with
  the existing "no violations found" vs "type mismatch" distinction: a
  leaf-shape mismatch is a normal "not a subset," only a *top-level* map-vs-set
  mismatch is the `fmt.Errorf` case).
- **Collection-like branch** (currently line 687: `functions.StructuralEqual(e,
  s)`): when both `e` and `s` are map-like or both collection-like, the
  element-level test becomes `IsSubsetValue(e, s)` (ignoring its error — a type
  mismatch between two list elements is "not a match for this candidate," not a
  propagated error, matching the existing semantics where only the *top-level*
  operand mismatch errors); otherwise fall back to `StructuralEqual`.

The membership semantics for the collection-like branch are **unchanged**: for
each `e` in `subEls`, *some* `s` in `superEls` must satisfy the (now recursive)
predicate — this is **not** an injective (1:1) matching, the same as today. See
§6 #1 for why this is the recommended, lower-risk choice and what it does and does
not guarantee.

### 3.2 Definition of done (Mandate 2's full checklist, applied to this fix)

| Step | Artifact |
| :-- | :-- |
| Go method / function | `evaluator.IsSubsetValue` (no new exported name — fixing the existing one is correct per Mandate 2 since it is "the same cell," not a new one; a rename would break `pkg/k8sgraph/lib/k8s/relations.gatwd:46`'s call site and the DSL `subset_of` name) |
| Operator overload | n/a — `subset_of` is not a `functions/operators.go` infix operator |
| Evaluator step | `evaluator.StepSubsetOf` already calls `IsSubsetValue` (line 717) — no change needed, it inherits the fix automatically |
| DSL step form | `subject \|> subset_of other` — unchanged registration, inherits the fix |
| DSL function form | `pkg/cli/dslfuncs.go:340` `subset_of` — unchanged registration, inherits the fix |
| DSL exhaustive test | New cases in `tests/dsl_exhaustive_test.gatwd` §122 (next free number — current last is §121, line 1562): nested-HAMT-in-HAMT subset (holds when live has *more* keys at the nested level), nested-List-in-HAMT subset (a sub list element with fewer fields than its live counterpart), and a negative case (a nested field whose *value* differs) — each in the function form inside a `fold`/comprehension per the existing §75/§122 convention, non-domain (plain records, no Kubernetes) |
| Structural Go test | `evaluator/subset_test.go`: new cases nesting one level deeper — `HAMT{spec: HAMT{...}} ⊆ HAMT{spec: HAMT{...extra...}}` (new: holds), `HAMT{spec: HAMT{a: 1}} ⊆ HAMT{spec: HAMT{a: 2}}` (new: fails — nested value mismatch still caught), `List[HAMT{name: X}] ⊆ List[HAMT{name: X, extra: Y}]` (new: holds) |
| `Example*` | `evaluator/subset_example_test.go` (new file, or added to an existing `*_example_test.go` if one exists for this package) — one `Example` showing a two-level-nested HAMT subset with `// Output:` |
| Negative control | Run each new nested test with the *old* flat-`StructuralEqual` code first (temporarily), confirm it fails (proves the test can fail), then apply the fix and confirm it passes — standard TDD per `PROMPTS.md` §4 ("write the pinning test first... then refactor") |

### 3.3 Risk check performed (not just claimed)

`grep -rn "subset_of\|IsSubsetValue"` across `evaluator/*_test.go`,
`tests/dsl_exhaustive_test.gatwd`, and `pkg/k8sgraph/lib/k8s/*.gatwd` (done as part
of writing this doc, §2.1) found:

- `evaluator/subset_test.go` — 8 cases, all one level deep (HAMT-of-scalars,
  Set-of-scalars). None nest a map/list value. **No existing test asserts the old
  flat-nested behavior**, so there is nothing to break by making it recursive.
- `tests/dsl_exhaustive_test.gatwd` §75 (lines 390-397) — Set-of-ints and
  HAMT-of-strings, both one level deep. Same conclusion.
- `pkg/k8sgraph/lib/k8s/relations.gatwd:46` — `subset_of (hamt_path "spec.selector"
  s) (hamt_path "metadata.labels" p)`. Both operands are already **pre-selected**
  flat string-to-string maps (via `hamt_path`, not the resource record itself), so
  they are one level deep regardless of the recursion fix — `Selects`'s behavior
  is provably unchanged (its operands never contain a nested map/list value to
  recurse into).

Conclusion: **low risk, verified** — not merely low risk by assumption. No
existing caller compares genuinely nested structures with `subset_of` today, which
is exactly why the bug has been latent rather than caught.

---

## 4. Phase 2 — `k8s_strip_runtime` / `k8s_up_to_date` (pkg/k8sgraph, domain shim)

### 4.1 Placement

Per `docs/development_playbook.md` §5b's three-way test: this needs **domain-shaped
data knowledge** (which fields are server-managed noise) but **nothing
repo/employer-specific** (no DNS scheme, no directory convention) — tier 2, the
shim's own prelude (`pkg/k8sgraph/lib/k8s/*.gatwd`), as plain `.gatwd` defs
composed from Phase 1's now-recursive `subset_of` plus the existing `hamt_del`/
`hamt_get_or`. No new Go code at all in this phase — it is pure DSL, consistent
with every other relation in this library.

### 4.2 New file: `pkg/k8sgraph/lib/k8s/parity.gatwd`

Following the naming and comment style of `relations.gatwd`/`ids.gatwd` (a
`--` comment stating the relation/invariant, helper closures named `k8s_*`, the
`fold`/comprehension idioms):

```
-- Manifest <-> live parity (read-only). A manifest-declared resource is "up to
-- date" when every field it declares is present and equal live — the live object
-- is always allowed to have MORE (server defaults, injected sidecars, controller
-- labels): see docs/manifest_live_parity_design.md §2.2 for why this must be a
-- RECURSIVE subset test, not flat equality, and what it does not catch (a value
-- the API server canonicalizes, e.g. a Quantity string, still reads as drift).

-- status is a clean, well-known, server-only subresource: never user-declared.
-- These metadata fields are server-managed siblings of the user-declared ones
-- (name, namespace, labels, annotations) with no structural separation in the
-- API, so they are named explicitly rather than inferred.
def k8s_runtime_metadata_fields = ["resourceVersion", "uid", "generation", "managedFields", "creationTimestamp", "deletionTimestamp", "selfLink"]

def k8s_strip_metadata = \meta -> fold (\m f -> hamt_del f m) meta k8s_runtime_metadata_fields

-- The "stable view" of a resource: what a manifest could plausibly have declared,
-- with the server-only/server-managed fields removed. Comparing two stable views
-- with subset_of answers "is everything I declared still true live," not "are
-- these two objects identical" (spec's additive, recursive defaulting makes
-- exact equality the wrong question — see design doc §2.2).
def k8s_stable_view = \r -> hamt_set "metadata" (k8s_strip_metadata (hamt_get "metadata" r)) (hamt_del "status" r)

-- Is a manifest-declared resource present, and up to date, in a live graph?
-- Returns False for "not present" as well as "present but drifted" — see
-- k8s_parity_report for which one happened.
def k8s_up_to_date = \sub live_g -> (graph_has_node (k8s_id_of sub) live_g) && (subset_of (k8s_stable_view sub) (k8s_stable_view (hamt_get (k8s_id_of sub) (graph_payloads_index live_g))))
```

(`graph_payloads_index` — see §4.3: a small helper to look a payload up by id,
needed because `graph_payloads`/`graph_payloads_of` yield a *list*, not a
by-id map; built from existing primitives, not a new core builtin — see the
decision in §6 #3.)

### 4.3 The "look up a live payload by id" gap

Scanning `pkg/cli/dslfuncs.go` (graph read-side functions, lines ~230-370) and
`pkg/cli/cli.go`'s `graph_*` dispatcher (`buildGraphStep`, `graph_focus`/
`graph_set`/`graph_over`, around line 2900+): the existing single-node accessor is
the **optic** trio `graph_focus`/`graph_set`/`graph_over` (`StepGraphFocus` in
`evaluator/graph_steps.go`, backed by `optics.NodePayloadLens`) — these are
**step** forms taking a literal id argument, not expression-level functions usable
inside a lambda/comprehension the way `graph_payloads_of` is. Phase 2 needs a
*function-form* "get the payload at this id, or report absent" usable inside
`k8s_up_to_date`'s lambda body. Two ways to close this, decided in §6 #3:

- **(a)** add `graph_payload_at id g` as a new DSL function form (a thin wrapper:
  `graph_payloads_of (k8s_id_kind id) g` filtered by `k8s_id_of p == id`, or more
  directly a Go function form backed by the existing `Graph.Nodes()`/HAMT payload
  index already inside `collections.Graph`) — **generic, core, Mandate 2/3-correct**
  (any graph, any domain, "get the node payload by id" is a missing matrix cell
  for the function-form row, since today only the step-form optics do this).
- **(b)** build it in DSL from what already exists: `head [p | p <- graph_payloads
  live_g, (k8s_id_of p) == (k8s_id_of sub)]` — no new core builtin, but O(n) per
  lookup (a linear scan of the whole payload list per manifest resource compared,
  i.e. O(manifest-subset-size × live-graph-size) overall).

Recommendation: **(a)**, because this is precisely a case the base design's own
recursive-dogfood principle (Mandate 4) flags — "get the current payload of node
X" is generically useful (any consumer of `graph_focus`'s *step* form will
eventually want its *function* form inside a comprehension, the same gap that
motivated `graph_payloads`/`graph_payloads_of`/`graph_in_of`/`graph_out_of`
originally), and the O(n) DSL-only version does not scale past toy fixtures for a
real cluster-sized comparison. This is a small, self-contained Phase 1b (see §5) —
one Go function form, one DSL function form, matching the exact "Go method →
function form → exhaustive test → structural test" checklist of Mandate 2, by
analogy to the already-landed `graph_in_of`/`graph_out_of` (added for the same
"needed inside a comprehension" reason — `docs/k8s_graph_design.md` Phase 7).

### 4.4 `k8s_parity_report`

```
-- One report line per manifest resource: [id, status] where status is one of
-- "ok" (present and up to date), "missing" (not present live) or "drift"
-- (present, but stable views differ). Read-only: never calls k8s_apply or any
-- write step.
def k8s_parity_status = \sub live_g ->
    head ([ "missing" | (graph_has_node (k8s_id_of sub) live_g) == false ]
       ++ [ "ok"      | (graph_has_node (k8s_id_of sub) live_g), (k8s_up_to_date sub live_g) ]
       ++ [ "drift"   | (graph_has_node (k8s_id_of sub) live_g), (k8s_up_to_date sub live_g) == false ])

def k8s_parity_report = \manifest_g live_g ->
    [ [(k8s_id_of sub), (k8s_parity_status sub live_g)] | sub <- (graph_payloads manifest_g) ]
```

### 4.5 Definition of done

- `.gatwd` defs in `pkg/k8sgraph/lib/k8s/parity.gatwd`, `import`ed from
  `prelude.gatwd` alongside the other library files.
- DSL test over fixtures (not the exhaustive-matrix file — this is domain, lives
  beside `pkg/k8sgraph/edition_test.go`'s style): a "clean" fixture pair
  (identical resources, live has extra server fields) → every status `"ok"`; a
  **negative fixture** (one resource scaled differently live, one resource never
  applied live) → `"drift"` and `"missing"` respectively, pairing with the
  existing `testdata/broken.yaml`-style convention (Mandate 6).
- godoc-equivalent: a `--` header comment per def (already modeled above),
  `pkg/k8sgraph/doc.go` gains a one-line mention under "the prelude" bullet list
  alongside `queries.gatwd`.

---

## 5. Phase 3 — the CLI verb (placement decided in §6 #2; default: `pkg/k8sgraph`)

### 5.1 Wiring (zero logic, per Mandate 7)

`cli.TaskCommand(edition, cmdName, short, taskName, knownDefines...)`
(`pkg/cli/verb_command.go:55`) already does exactly this: promote a `def` already
in the Edition's prelude to its own subcommand. `k8s_parity_report` takes two
graphs, not a stream, so the task wrapper is a thin pipeline `def` (the existing
"zero-parameter fragment reads its path from `--` arguments" idiom,
`docs/development_playbook.md` §6):

```
-- Added to pkg/k8sgraph/lib/k8s/parity.gatwd. Two `--` arguments: the manifest
-- path/root, then the namespace to crawl live (empty string = all namespaces,
-- matching k8s_resources's existing convention, e.g. TestLiveOfflineParity's
-- fx.ns `""`). Read-only.
def k8s_parity_task = \manifest_path ns ->
    k8s_parity_report (manifest_path |> manifests |> k8s_graph) (ns |> k8s_resources |> k8s_graph)
```

`cmd/gatwd-k8s/main.go` (the existing pure-assembly Edition main) gains one line:

```go
edition = edition.WithCommand(cli.TaskCommand(edition, "parity",
    "check whether a manifest subset is present and up to date live (read-only)",
    "k8s_parity_task"))
```

This is the entire Go surface of Phase 3 — no new flags, no new client wiring (it
reuses the Edition's existing `LazyDynamicClient()`), no new capability (§1.1).

### 5.2 Definition of done

- `cmd/gatwd-k8s/main.go` one-line `WithCommand` addition.
- `cmd/gatwd-k8s/cli_test.go` (existing file, same pattern as its other subcommand
  tests): run `gatwd-k8s parity -- <manifest> <ns>` against
  `k8sgraphtest.FakeClusterFromFile` (reusing the exact helper
  `TestLiveOfflineParity` already uses, §2.3) over a fixture with one resource
  deliberately scaled differently and one resource never applied — assert the
  report names both by id and status, and that a clean fixture reports all `"ok"`
  (the negative-fixture pairing of Mandate 6, applied at the CLI-integration
  level since `k8s_parity_report` already has its own DSL-level negative fixture
  in Phase 2).
- No change to `pkg/cli` policy/capability code — confirmed by re-reading
  `pkg/cli/policy.go` (§2.1): reads need no capability grant.

---

## 6. Key decisions & open questions

1. **List-element matching inside recursive `IsSubsetValue`: non-injective
   "exists a satisfying element" (current semantics, extended recursively) vs.
   positional pairing vs. full injective (bipartite) matching.**
   **Recommendation: keep the current non-injective semantics** — for each `e` in
   `sub`'s list, *some* element of `super`'s list must satisfy (recursively) `e`;
   the same `super` element may satisfy more than one `sub` element (no 1:1
   assignment is required or attempted). This is the **lowest-risk** choice
   because it is exactly what the existing flat implementation already does
   (§3.1), so Phase 1 is purely "make the per-element test recursive," not "also
   change the matching strategy" — two independent changes bundled into one would
   make the risk assessment (§3.3) harder to trust. **Known limitation, recorded
   not solved**: this can produce a false "subset holds" for two structurally
   similar-but-distinct `sub` list elements that both happen to be satisfied by
   the *same* `super` element (e.g. two containers in the manifest that declare
   only a common subset of fields, both matched against one richer live
   container, while a different live container that should have matched the
   second sub-container goes unmatched and unnoticed). A positional variant
   (pair `sub[i]` with `super[i]`, trusting Kubernetes to preserve list order and
   only *augment* each element — true for `containers`/`ports`/`volumes` in
   practice) would be more precise for this exact use case, but it is K8s domain
   knowledge about *which* lists are order-stable, not a generic claim about
   lists in general (an arbitrary DSL program's lists are not guaranteed
   order-stable under "augmentation"). **Decision: defer the stricter variant**;
   if real manifests show false "ok" reports from this gap, add a k8s-specific
   positional comparator in `pkg/k8sgraph`'s `parity.gatwd` (Phase 2's own file,
   not core) rather than changing core's generic semantics. Not resolved by
   assumption — flagged here precisely because it is a real algorithmic choice.

2. **Where does the parity verb's CLI surface live: `pkg/k8sgraph` (shared,
   any K8s tool) or a child/employer repo?**
   **Argued for `pkg/k8sgraph` (recommended) and against, per
   `docs/development_playbook.md` §5b's three-way test:**
   - *For `pkg/k8sgraph`*: the verb needs domain-shaped data knowledge (which
     fields are server-managed noise — §2.2) but **nothing** repo/employer-specific:
     no DNS scheme, no particular overlay directory convention, no assumption
     about which namespaces or clusters an organization uses. Every piece it
     composes (`manifests`, `k8s_resources`, `k8s_graph`, `subset_of`,
     `k8s_id_of`) is already in the shared shim or core. "Is this manifest subset
     present and up to date live" is a question *any* Kubernetes tool built on
     this engine would want, the same way `k8s_apply_order`/`k8s_cycles` already
     are generic-enough-for-any-k8s-consumer queries living in `queries.gatwd`
     rather than a child repo.
   - *For a child repo*: one could argue the **reporting shape** (is it a table?
     JSON? which exit code means "drift found"?) and **which clusters/namespaces
     get compared against which overlay** are operational conventions that differ
     per organization, and a specific team's wrapper script (e.g. "compare
     `overlays/prod` against the `prod-eu` context") is legitimately tier-3.
   - **Recommendation**: put the **primitive** (`k8s_parity_report`,
     `k8s_up_to_date`, the `parity` subcommand reading two positional `--`
     arguments) in `pkg/k8sgraph`, exactly as `k8s_apply_order` sits there rather
     than in a child repo, because it reads its source from `--` arguments rather
     than hard-coding any path/namespace (the existing "never put a path in a
     task script" discipline, `docs/development_playbook.md` §6). Any
     organization-specific wrapper (which overlay maps to which cluster context,
     custom report formatting, Slack/ticket integration) stays firmly in a child
     repo that calls `gatwd-k8s parity -- <path> <ns>` or imports
     `k8sgraph.ApplyEdition` directly. This mirrors exactly how `k8s_apply`
     already separates "the generic write primitive, gated" (shim) from "which
     cluster to point it at" (caller's responsibility, never baked in).

3. **The "payload lookup by id" gap (§4.3): new core function form
   `graph_payload_at` vs. DSL-only O(n) lookup.**
   **Recommendation: add `graph_payload_at id g`** as a genuine new matrix cell
   (Mandate 2/3 — Go method on `*collections.Graph` if one is not already exposed
   as a function form, wired to `pkg/cli/dslfuncs.go`, exhaustive test, structural
   test) rather than building it from `graph_payloads` + a linear filter in DSL.
   This is a small, separable **Phase 1b** (core, before Phase 2 needs it) — not
   bundled into Phase 1's `IsSubsetValue` fix because it is a different kind of
   change (a new function-form cell, not a bugfix to an existing one) and should
   get its own commit/gate per Mandate 8. Deferred to the implementer to confirm
   at build time whether `collections.Graph` already exposes a by-id node-payload
   accessor Go-side (the base design's §3.2 lists `Nodes` as a projection, but its
   signature was not re-verified for this doc — Appendix C's "verify before
   implementing" note applies) and wire only the missing layers.

4. **The residual canonicalization gap (§2.2, e.g. Quantity normalization) is
   explicitly not solved by this plan.** No generic or domain mechanism is
   proposed to canonicalize values before comparison; a future phase could add a
   k8s-specific value-normalizer (e.g. parse-and-compare `Quantity` strings
   numerically) as a `.gatwd` def composed into `k8s_stable_view`, but that is
   new domain logic beyond "strip known-noisy fields," and is left as an
   explicitly deferred idea, not a phase of this plan.

5. **Capability (2) from the task background — kustomize-generator-layout
   rendering from a live scrape — is out of scope** (§1.2), confirmed by rereading
   `pkg/k8sgraph/doc.go` and `pkg/cli/dslfuncs_yaml.go`: the apply/render and
   YAML-marshal primitives already exist; only the layout *policy* is missing,
   and it has no single right answer across teams. Not designed here.

---

## 7. Appendix A — concrete DSL sketches

See §4.2, §4.4, §5.1 above for the full `parity.gatwd` sketch (reproduced here as
one block for an implementer to start from verbatim, pending Phase 1/1b's
`graph_payload_at` landing):

```
-- pkg/k8sgraph/lib/k8s/parity.gatwd (NEW FILE — sketch, not yet implemented)

def k8s_runtime_metadata_fields = ["resourceVersion", "uid", "generation", "managedFields", "creationTimestamp", "deletionTimestamp", "selfLink"]

def k8s_strip_metadata = \meta -> fold (\m f -> hamt_del f m) meta k8s_runtime_metadata_fields

def k8s_stable_view = \r -> hamt_set "metadata" (k8s_strip_metadata (hamt_get "metadata" r)) (hamt_del "status" r)

def k8s_up_to_date = \sub live_g ->
    (graph_has_node (k8s_id_of sub) live_g)
    && (subset_of (k8s_stable_view sub) (k8s_stable_view (graph_payload_at (k8s_id_of sub) live_g)))

def k8s_parity_status = \sub live_g ->
    head ([ "missing" | (graph_has_node (k8s_id_of sub) live_g) == false ]
       ++ [ "ok"      | (graph_has_node (k8s_id_of sub) live_g), (k8s_up_to_date sub live_g) ]
       ++ [ "drift"   | (graph_has_node (k8s_id_of sub) live_g), (k8s_up_to_date sub live_g) == false ])

def k8s_parity_report = \manifest_g live_g ->
    [ [(k8s_id_of sub), (k8s_parity_status sub live_g)] | sub <- (graph_payloads manifest_g) ]

def k8s_parity_task = \manifest_path ns ->
    k8s_parity_report (manifest_path |> manifests |> k8s_graph) (ns |> k8s_resources |> k8s_graph)
```

```go
// cmd/gatwd-k8s/main.go — one added line, pure assembly (Mandate 7)
edition = edition.WithCommand(cli.TaskCommand(edition, "parity",
    "check whether a manifest subset is present and up to date live (read-only)",
    "k8s_parity_task"))
```

---

## 8. Appendix B — mandate compliance checklist

- [ ] Core change (`evaluator.IsSubsetValue` made recursive) names no domain
      identifier and imports nothing new (Mandate 5) — verify with
      `internal/archtest` after the change.
- [ ] Matrix cell `subset_of` stays the **same** cell across Record/HAMT and
      Set/collection carriers, step form and function form, now with correct
      recursive semantics (Mandate 2) — no new cell, no second-class citizen
      introduced.
- [ ] All-layers-through unaffected: `StepSubsetOf` and the `subset_of` DSL
      function both call `IsSubsetValue` directly, so the fix is inherited
      without touching either registration (Mandate 3).
- [ ] `k8s_strip_runtime`/`k8s_stable_view`/`k8s_up_to_date`/`k8s_parity_report`
      are pure `.gatwd`, composed from existing generic + existing domain
      primitives; no Go loop fakes the comparison (Mandate 1).
- [ ] The parity verb is read-only: no new capability, no touch to
      `k8s_apply`/`k8s_rollout_*`/`k8s_scale`, confirmed by rereading
      `pkg/cli/policy.go` (Mandate 6 — read failures, e.g. "no cluster access,"
      already surface loudly via `LazyDynamicClient`, not silently).
- [ ] Phase 2's "no violations"/"all ok" case is paired with a negative fixture
      (a deliberately-drifted and a never-applied resource) at both the DSL
      level (Phase 2) and the CLI-integration level (Phase 3) (Mandate 6).
- [ ] `cmd/gatwd-k8s/main.go`'s only change is one `WithCommand` line; the
      verb's logic lives entirely in `pkg/k8sgraph/lib/k8s/parity.gatwd`
      (Mandate 7).
- [ ] Each phase gated by `go test ./... && go vet ./... && ./run_dsl_tests.sh`
      before the next starts; decisions in §6 recorded with rationale before
      implementation proceeds past them (Mandate 8).
- [ ] `graph_payload_at` (if added, §6 #3) gets its own godoc, `Example*`, and
      DSL exhaustive-test entry, non-domain, like every other core graph
      builtin (project-doc-sync discipline).

---

## 9. Appendix C — Implementation reference (signatures & file anchors)

Verified against the codebase on 2026-10-14. **Re-verify line numbers before
implementing** — this tree is shared and may have moved since.

**The containment primitive**
- `evaluator/pipeline.go:626` — `func rangeMapLike(v any, fn func(k string, val any)) bool` (dispatches `*collections.HAMT`, `*collections.Map[string,any]`, `map[string]any`)
- `evaluator/pipeline.go:648-700` — `func IsSubsetValue(sub, super any) (bool, error)` (the function to make recursive; see §3.1 for the exact two call sites, lines 668 and 687 in the as-read version)
- `evaluator/pipeline.go:706-730` — `func StepSubsetOf(other any) PipelineStep[functions.Either[error, any]]` (pipeline-step twin, calls `IsSubsetValue` at line 717 — inherits the fix unmodified)
- `evaluator/pipeline.go:1653` — `func HAMTGetKey(val any, key string) (value any, found bool, navigable bool)`
- `evaluator/pipeline.go:1676` — `func NavigateHAMTPath(cur any, path string) (any, error)`
- `functions/operators.go:880` — `func ToSliceAny(val any) ([]any, bool)`
- `functions/operators.go:1071` — `func StructuralEqual(a, b any) bool`
- `pkg/cli/dslfuncs.go:340-345` — `m["subset_of"] = functions.DSLFunc(...)` (calls `evaluator.IsSubsetValue` directly)
- `evaluator/subset_test.go` — `TestIsSubsetValue` (8 structural cases, all one level deep — see §3.3)
- `tests/dsl_exhaustive_test.gatwd:386-397` — §75, the DSL-level `subset_of` assertions (step form, function form inside `fold`)

**Graph read-side (existing)**
- `pkg/cli/dslfuncs.go:238-258` — `graph_payloads` / `graph_payloads_of` function forms (`m["graph_payloads"]`, `m["graph_payloads_of"]`)
- `pkg/cli/cli.go:2848` (comment) / `:2906` — `graph_filter_nodes` dispatcher case (Filterable row, for reference on the existing step-form conventions around the `graph_*` dispatcher)
- `pkg/cli/dslfuncs_matrix.go:454-490` — `hamt_set`, `hamt_del`, `hamt_get_or` function forms (curried via `defn(m, name, "args...", fn)`)
- Graph collection (base design, Phase 1, `docs/k8s_graph_design.md` §9) — `collections/graph.go`, `Graph[K,C,N,E]`; **re-verify at implementation time** whether a by-id node-payload accessor already exists Go-side before adding `graph_payload_at` (§6 #3) — this doc does not re-confirm that signature.

**Kubernetes shim (`pkg/k8sgraph/`)**
- `pkg/k8sgraph/lib/k8s/ids.gatwd:1-37` — `k8s_scope`, `k8s_id`, `k8s_id_of`, `k8s_id_kind`, `k8s_id_scope`, `k8s_short`, `k8s_ref` (identity scheme: `Kind/namespace/name`)
- `pkg/k8sgraph/lib/k8s/relations.gatwd:46` — the one existing `subset_of` call site in the k8s library (`Selects`), confirmed one-level-deep (§3.3)
- `pkg/k8sgraph/lib/k8s/queries.gatwd:1-18` — `k8s_apply_order`, `k8s_cycles`, `k8s_apply_objects` (the style/placement precedent for `parity.gatwd`)
- `pkg/k8sgraph/lib/k8s/sources.gatwd` — `k8s_resources` (the `<+>`-merge of `CoreCrawls`)
- `pkg/k8sgraph/steps.go:43-60` — `var CoreCrawls = []struct{ Keyword string; GVR schema.GroupVersionResource }{...}` (12 entries)
- `pkg/k8sgraph/lazy.go:14-49` — `LazyDynamicClient()`, `LazyClient(build func() (dynamic.Interface, error)) dynamic.Interface` (defers kubeconfig load to first crawl; a build failure surfaces as `"no cluster access (...)"` from every subsequent call, not silently)
- `pkg/k8sgraph/edition.go:71-108` — `func ApplyEdition(e *cli.Edition, client dynamic.Interface, opts ...Option) *cli.Edition` (registers `manifests`/`kustomize`, the write steps gated by capability, and every `CoreCrawls` entry when `client != nil`, then `e.WithPreludeFS(preludeFS, PreludePath)`)
- `pkg/k8sgraph/doc.go:1-79` — package overview; the prelude file list (line 24-33) is where `parity.gatwd` gets a new bullet
- `pkg/k8sgraph/k8sgraphtest/apply.go` (package `k8sgraphtest`, doc comment lines 1-6) — `FakeClusterFromYAML`/`FakeClusterFromFile` (test-only, builds a fake dynamic client from the same fixture `manifests` reads — reused by Phase 3's test)
- `pkg/k8sgraph/live_parity_test.go:17-40` — `TestLiveOfflineParity` (the exact fixture/fake-client/comparison pattern Phase 3's test follows)

**CLI verb wiring (generic, `pkg/cli/`)**
- `pkg/cli/verb_command.go:55-100` — `func TaskCommand(edition *Edition, cmdName, short, taskName string, knownDefines ...string) *cobra.Command`
- `pkg/cli/policy.go:66-119` — `type Capability string`; `CapExec`, `CapFS`, `CapNet`, `CapEnv`, `CapClusterApply`, `CapClusterRollout`; `type Policy struct{...}`; `TrustedPolicy()`, `DenyAll()` — confirms reads need no capability (line 85's comment: "reading the rollout status is a read and needs no capability")
- `pkg/cli/dslfuncs_yaml.go:1-60` — `yaml_parse`/`yaml_parse_all`/`yaml_encode`/`yaml_encode_all` (already registered, relevant to the out-of-scope §1.2 capability, not to this plan's phases)
- `pkg/k8sgraph/write.go` — `k8s_apply` / `k8s_apply_objects` (`pkg/k8sgraph/lib/k8s/queries.gatwd:18`), the existing write path this plan deliberately never touches

**Templates / precedent for phase style**
- `docs/k8s_graph_design.md` — the base design this plan is the direct sequel to; its Phases 5-8 and §6 #10 (live→kustomize parity) are the grounding for §2.3 and the `TestLiveOfflineParity` pattern reused in Phase 3.
- `docs/development_playbook.md` §5b — the three-way CLI-verb placement test applied in §6 #2.
