# Development Playbook — How Work Gets Done in gatwd

> Audience: human contributors **and** AI agents. This is the long-form companion to the
> rules in [`AGENTS.md`](../AGENTS.md): the rules say *what must hold*; this document says
> *how work is actually done here* and *why*, using the project's own history as evidence.
> The reference case study is the Kubernetes-graph arc (`docs/k8s_graph_design.md`), which
> took a generic graph from nothing to a live operator across eight phases while the
> language grew first-class functions, gradual typing and an expression-level function
> matrix underneath it.
>
> Facts here were verified against the tree and `git log`. Avoid copying counts
> ("N assertions", "§77", commit hashes of HEAD) into other docs — they rot. Link to a
> section or a test name instead.

---

## 1. The shape of a change

Every feature is a **point in three dimensions**. A change is not done until all three are
filled for the point it touches.

| Dimension | Question | Mandate |
| :-- | :-- | :-- |
| **Vertical** (layers) | Is it reachable from the DSL, not stranded in Go? | 3 — All Layers Through |
| **Horizontal** (types) | Does every type where it is well-defined have it? | 2 — Operator-Type Matrix |
| **Form** (step vs value) | Is there both a pipeline *step* form and a curried *function* form? | 2/3 (see §4) |

Plus two cross-cutting duties: it must **fail loudly** (Mandate 6) and it must **stay in the
right place** (Mandate 5 Clean Core, Mandate 7 Pure Assembly).

Worked example — adding a graph read operation (`graph_scc`):

1. `collections.Graph.SCC()` — Go method with structural tests (Layer 1).
2. `evaluator.StepGraphSCC` — stream step (Layer 3), via the shared `graphUnfold` shape.
3. `pkg/cli` — the `graph_scc` step is dispatched by `buildGraphStep`; the **function form**
   `graph_scc g` is registered in `pkg/cli/dslfuncs_graph.go` so a comprehension can call it.
4. `tests/dsl_exhaustive_test.gatwd` — a self-asserting case for the step form *and* one that
   calls the function form inside a lambda/comprehension.
5. `Example*` with `// Output:` plus godoc; `docs/dsl_guide.md` / `docs/grammar.md` if syntax changed.

---

## 2. Delivery protocol (phase-gated)

The k8s-graph arc was delivered as a sequence of small, gated phases. This pattern worked
and is the default for anything larger than a bug fix.

1. **Design first.** Write or extend a design doc (`docs/<topic>_design.md`) *grounded in the
   real code* (grammar, engine types, edition/appliance machinery): exact signatures and file
   anchors in an appendix so a fresh session does not re-derive them.
2. **Split generic from domain.** Anything a *second* domain could use lands first, generically
   (Edition preludes `cli.Edition.WithPreludeFS`, `cli.AdapterStep`, `hamt_path`, `subset_of`,
   the `Graph` collection). The domain layer is then a thin shim over it (`pkg/k8sgraph`).
   Record the split as an explicit decision in the design doc.
3. **Prerequisite phases before domain phases.** Gaps found while designing become numbered
   prerequisite phases (P1, P2, P3a/b), approved and committed before the phase that needs them.
4. **One commit per sub-step**, subject `feat(scope): summary (project Phase N)`, body = what
   was added, the decision taken, the tests that prove it, and "all gates green".
5. **Gate every step** (§3). Never stack unverified work.
6. **Stop at decision points.** When several designs are plausible (stream-merge strategy,
   in-process vs external kustomize, how to ingest), present the options with a
   recommendation, **ask**, and record the decision with its date in the design doc. This
   happened at P1, the ingest "option C", the kustomize dependency, and the live-source merge.
7. **Record deferrals, never silently skip.** A matrix cell, relation or feature left out goes
   into the phase entry with the reason. "Deferred, not silent."
8. **Hand off in writing** (§9) whenever a context ends mid-project.
9. **Close the loop in docs:** design-doc phase entry → DONE with what changed, Appendix B
   checklist ticked, remaining ideas listed.

When the user has already delegated ("pick up from here"), still apply 6 for choices that are
hard to reverse or that the handoff explicitly flags; otherwise pick the recommended option and
say which you picked and why.

---

## 3. Verification gates

Run **all** of these before calling anything done, and again before a commit:

```bash
gofmt -l <changed dirs>          # must print nothing for files you touched
go vet ./...                      # compile, vet, and no import cycles
go test ./...                     # unit + structural + Example tests (Examples run here too)
./run_dsl_tests.sh                # DSL suite: fails on [ERROR], on silent output, on --check errors
go test -race <touched pkgs>      # parallel steps (||>) share interpreter state
```

`go test ./...` includes `internal/archtest`, which enforces:

- **Layer boundaries** of the library layers (`TestLibraryLayersBoundaries`; its
  `knownViolations` list may only shrink).
- **Clean Core** (`TestCleanCoreImportsNoThirdParty`, `TestCleanCoreNoDomainIdentifiers`): core
  packages import only the standard library and this module (plus an explicit allowlist: cobra/pflag in
  `pkg/cli` only), and no non-test identifier names a
  domain. Comments may still use a domain as an *example*.
- **Documentation links** (`TestDocLinksResolve`): every relative markdown link in the root docs,
  `docs/`, `.agents/` and the skills must resolve, and none may hard-code a machine-specific path.

`go test -run Example ./...` is redundant with `go test ./...` but is the fast loop for godoc.

### Regression check for interpreter changes

Changes to `pkg/cli`, `parser` or `evaluator` can alter script behavior in ways unit tests miss.
Build the pre-change binary (`git worktree add /tmp/base HEAD && go build -o /tmp/gatwd_base
./cmd/gatwd`), run every `*.gatwd` outside `tests/` and `legacy/` with both binaries and diff
the output. A few scripts are nondeterministic on `master` alone (Go map iteration order:
`examples/tour/15`, `20`, `23`); run the baseline several times before blaming your change.
Run `gatwd --check` over all scripts too: it must stay silent (zero false positives is the bar).

### Tests must be able to fail

A passing test proves nothing unless it could have failed. Three rules, each from a real
incident:

1. **Assert on non-empty output.** `assert` on a stream that delivers no node is an error
   (`assertion on an empty stream`), and `run_dsl_tests.sh` also fails on
   `(No output or node unhandled)`. When you add DSL tests, run just your new section and look
   at every output line.
2. **Pair "no violations" with a negative fixture.** A compliant cluster yields `[]`, which is
   indistinguishable from a broken pipeline. Every rule set needs a fixture that violates every
   rule (`cmd/gatwd-k8s-operator/testdata/broken.yaml`) and a test that expects exactly those
   violations.
3. **Negative-control new guards.** After adding a guard test, add a deliberate violation, see
   it fail, remove it. (The Clean Core tests and the DSL gate were verified this way.)
   Never use `t.Logf` as a test — the `conway_step` bug hid behind one for a long time.

---

## 4. Step form, function form, and "replacing strings with real types"

A DSL operation can exist in two forms, and **both must exist and behave identically**:

| | Step form | Function form |
| :-- | :-- | :-- |
| Syntax | `xs \|> hamt_set k 1` | `hamt_set k 1 m` (juxtaposition) |
| Arguments | literal string tokens parsed out of step text | computed runtime values (`hamt_set (hamt_get "k" r) n m`) |
| Position | point-free, top of a pipe | inside lambdas, folds, comprehensions |
| Registered in | `pkg/cli/cli.go` step dispatch (+ `evaluator/` step) | `pkg/cli/dslfuncs*.go` (`defn`) |

Function-form conventions (all enforced by review; see `dslfuncs_matrix.go`'s header):

- **Data-last**: the thing operated on is the last argument, so `x |> f a` ≡ `f a x` and partial
  application (`map (hamt_get "a")`) works.
- **Curried**: register through `defn` so the value is a `functions.Closure`; builtins and user
  lambdas are then indistinguishable to callers.
- **Preserve element types.** Never stringify. Only the HAMT *key* (a string by construction) and
  the graph's id/color (pinned to `string`) are coerced, each function owning its own policy.
- **Errors are values.** Wrong carrier, non-callable argument, out-of-range index, unhashable Set
  element → an error that reaches the Left channel, never a panic or a silent no-op.
- **Carrier polymorphism lives in one place** (`pkg/cli/valueops.go`: `mapValue`, `filterValue`,
  `foldValue`, `bindValue`). Do not re-implement per-carrier loops in a new builtin.

A name can be both a step and a function (`graph_scc`, `hamt_get`). If the two would diverge in
meaning (the step `graph_edges_of` yields formatted edge strings), give the function a distinct
name (`graph_edge_pairs`) instead of overloading.

---

## 5. Placing code: the Edition / Appliance / shim pattern

Reference implementation: `pkg/k8sgraph`, `cmd/gatwd-k8s`, `cmd/gatwd-k8s-operator`.

```
core (generic, domain-free)  ──►  adapters/ (I/O, client-go confined here)
                                      │
                         pkg/<domain>/  (thin shim: clients + step registration)
                                      │            embeds lib/<domain>/*.gatwd (the logic)
                       cmd/gatwd-<domain>  (Edition main = pure assembly)
                       cmd/gatwd-<domain>-<role>  (Appliance: edition + embedded rules)
```

Rules (Mandates 5 and 7):

- **Domain logic is `.gatwd`, not Go.** Node/edge vocabularies (`type NodeColor = …`), ingest,
  relations and compliance rules live in the embedded prelude/rules. Go does only real I/O.
- **`main.go` is pure assembly.** Everything it registers is supplied by core, `adapters/`,
  `pkg/plugins/` or the embedded DSL. No domain algorithm, loop, relation or type inline. Small
  testable helpers for flags and pipeline strings (`newAppliance`, `checkPipeline`,
  `livePipeline`) are fine; logic is not.
- **The shim has zero logic.** `k8sgraph.CrawlStep` constructs an adapter and hands it to the
  generic `cli.AdapterStep`. `k8sgraph.ApplyEdition(e, client)` assembles steps + prelude; a nil
  client builds an offline-only edition (CI, tests).
- **Third-party SDKs are confined** to specific adapter/shim files. Prefer an external binary plus
  an injectable seam (`ManifestAdapter` + `WithRenderer`, shelling out to `kustomize`) over a heavy
  dependency, unless the user decides otherwise.
- **Source-agnostic downstream.** Live and offline sources yield the *same* payload shape (nested
  HAMT), so the same prelude runs on both. Prove it with a parity test over a fake client
  (`pkg/k8sgraph/k8sgraphtest` + `TestLiveOfflineParity`), never by inspection.
- **Merge sources generically.** `(a <+> b <+> c)` (`arrows.Plus` → `evaluator.StepPlus`) is the
  domain-free way to feed one pipeline from many producers; the domain only names the list
  (`lib/k8s/sources.gatwd`).
- **Tests use fakes and fixtures:** fake dynamic client, manifest files under `testdata/`, tests
  that need an external binary are skipped when it is absent. Never depend on a real cluster.

Minting a new Edition/Appliance: follow `.agents/skills/edition-appliance/SKILL.md`.

---

## 6. Writing DSL: conventions and traps

The project's own relations and rules are the style reference
(`pkg/k8sgraph/lib/k8s/*.gatwd`, `cmd/gatwd-k8s-operator/rules/compliance.gatwd`).

**Shapes that work**

- *Ingest* = a zero-parameter `def` wrapping a `fold` with a graph accumulator.
- *Relation* = a graph→graph pass: `fold (\_ g -> graph_add_edges [spec | …] g) graph_empty`
  (a graph stream element reaches the reducer as `g`). Chain passes with `|>`.
- *Rule* = a closure `\g -> [ [violation, resource] | … ]`; combine rules with `++`.
- *Helpers* = ordinary closures (`def k8s_names = \kind g -> […]`); they are values, so call them
  inside comprehensions.
- *Optimise with an index, in DSL.* The label-indexed `Selects` pass builds a HAMT of pod lists
  with `fold` + `hamt_set` + `hamt_get_or` — no Go needed.

**Traps (each cost real time; prevention rule in bold)**

| Symptom | Cause | Rule |
| :-- | :-- | :-- |
| `[f x, g y]` gives four elements | gatwd serialises lists space-separated; inside a *plain* `[...]` adjacent atoms are separate elements | **Parenthesise calls in list literals: `[(f x), (g y)]`.** Comprehensions and `( … )` re-enable application |
| Edges missing from listings | `Connect` does not create endpoint nodes | **Add nodes first, then edges**; guard with `graph_has_node` |
| A comprehension errors on a missing field | filters are positional (run when their variables are bound) | **Put `hamt_has_path` guards *before* the dependent generator** |
| `map (\g -> …)` changes payloads | `map` on a graph element is the graph Functor | **Whole-graph passes use `fold (\_ g -> …) seed`** |
| A `def` mis-evaluates its arguments | a `def` whose body is a *pipeline* is a re-parsed text macro | **Value functions are lambda-bodied (`def f = \x y -> …`); pipeline fragments are zero-parameter defs** |
| `assert` reports "empty stream" | `assert` now fails when no node arrives (D9) | **Fix the pipeline that produced nothing; do not weaken the assert** |
| `-x` fails to parse | unary minus applies to numeric literals only | **Write `0 - x`** |
| A partial application is run as a step | a step whose text evaluates to a *function* is not an expression value | **The infix-step detector ignores callables; keep it that way** |
| `to_set` set ≠ function-form set | the step builds `Set[string]`, function forms normalise to `Set[any]` | **Compare via `member`/`length`, not `==`, across the two forms** |
| Test helper wrote nonsense | `echo '…\x…'` interprets escapes in zsh | **Write scratch scripts with `cat > f <<'EOF'`** |
| Results look odd in Go tests | `EvaluateToValues` treats its argument as one pipeline; `def` lines in a multi-line string are not registered | **Use `EvaluateScript` for setup, `EvaluateToValues` for the query** (empty input returns no values) |
| Result type surprises | lists come back as `*collections.List[any]` | **Use `functions.ToSliceAny`** |
| `||>` output order differs run to run | `||>` fans the next step out on a worker pool (`evaluator.StepPar`) | **Sort or assert order-insensitively; use `|>` when order matters** |
| A `let`-bound HAMT/list/graph misbehaves in a head position | head-word text substitution used the lossy `%v` text | **Fixed: non-scalar runtime values are never substituted; do not re-add text substitution for them** |
| A step that needs quotes (`sh -c "…"`) lost them | exec steps used `strings.Fields` | **Fixed: `interactive.SplitCommandLine` tokenizes like a shell (no expansion); an unterminated quote is an error** |

---

## 7. Performance protocol

1. **Measure first**, on a synthetic workload that scales (`BenchmarkK8sGraph`,
   `benchmarks/run_k8s_graph.sh`).
2. **Profile per stage** before touching anything (time each relation pass separately). The
   O(services × pods) `Selects` pass was about three quarters of the run; guessing would have
   optimised the wrong pass.
3. **Fix in the layer that owns the problem** — here a pure-DSL index, not a Go special case.
4. **Pin semantics first**: write the test that defines correct behavior (multi-pair selector,
   unlabeled Pod, empty selector) so the optimisation cannot change results.
5. **Record honestly**, including negative results (`benchmarks/README.md`: the step cache was
   ~1%). Numbers are single runs on one machine; say so.

---

## 8. Documentation protocol

- **godoc on every exported identifier**, starting with the identifier's name; package `doc.go`
  for every package; `Example*` with `// Output:` (they are tests). The educational standard in
  `STYLEGUIDE.md` §3 applies: explain the concept, why it is needed here, how it is implemented,
  and give a usage example.
- **Docs describe the present.** Nothing in `docs/` or comments may point into `legacy/`. Dated
  snapshots (`docs/assessment.md`, `docs/report.md`) carry a "historical" banner.
- **Update in the same change** (not later): `docs/grammar.md` for syntax; `docs/dsl_guide.md`
  for behavior and builtins; `docs/components.md` for types and methods; `README.md` features;
  the design doc phase entry; the roadmap status block.
- **Do not hard-code counts or hashes** that change every commit; link to the test/section.
- **Docs-as-tests:** an `Example` found a real bug (`Chunk.Bind` truncating past capacity); a
  self-asserting tour exposed first-arm-wins `match`. Prefer self-asserting docs.

---

## 9. Handoffs, memory and git hygiene

A **handoff** (`docs/phaseN_handoff.md`) lets a fresh context continue without the chat. Required
sections: where we are (done, committed) · goals · authoring shape (copy-pasteable patterns) ·
gotchas ("don't re-derive") · hard constraints · verification gates · pointers (files, design-doc
sections) · ordered first steps. Rules:

- It states which earlier handoff it **supersedes**; mark superseded ones with a banner (or delete
  them in the commit that completes the phase) so nobody follows a stale guide.
- Refer to sections and test names, not line numbers or `HEAD`.
- Update the design doc's phase entry and Appendix B checklist when the phase closes.

Template: `.agents/prompts/phase-handoff.md`.

**Git:** one logical commit per sub-step; never amend published commits; never skip hooks; stage
files by name (this tree often holds other people's untracked files — `.claude/`, scratch docs);
commit and push only when asked. **Work may be happening in the same tree concurrently** (library
extraction, docs): re-read a file before editing it, make additive edits, never revert changes you
did not make.

---

## 10. Case study: the k8s-graph arc, as a timeline

| Phase | What | Why it is instructive |
| :-- | :-- | :-- |
| P1 + P3a | `Edition.WithPrelude(FS)`, `cli.AdapterStep` | generic enablers landed *before* the domain |
| 1 | `collections.Graph` + `functions.Node/Edge` | color is a generic `C comparable`, not a DSL type: L0–1 stay pure generics |
| 2 | walks, SCC, Dijkstra, comonad cursor, optics | structure first, DSL later |
| 3–4c | `Graph(T)` grammar, ~40 `graph_*` steps, Functor/Fold/Filter/Monoid/Lens | whole matrix row through the DSL, deferrals recorded |
| P2 | structured nested-HAMT resources (not names) | "replace strings with real types" |
| 6-prereq | `hamt_path` + `HAMTPathLens` | an absent-key path is an *Optional*, not a Lens |
| P3b | `k8sgraph.CrawlStep` | shim with zero logic |
| 5 | offline manifests, kustomize via injectable renderer | user decision: external binary, no heavy dep |
| ingest | `fold` with a graph accumulator + **juxtaposition application** | a missing language primitive implemented properly (Mandate 1), not faked in Go |
| 6 | embedded pure-DSL prelude + `ApplyEdition` | relations in `.gatwd` |
| 7 | `cmd/gatwd-k8s`, `cmd/gatwd-k8s-operator` | pure-assembly main; hot-reloadable rules |
| 8 | `<+>` live source, relation matrix, six rules, perf | parity test, negative fixture, profile-then-fix |
| roadmap B–E | `Closure`, currying, gradual typing, function-form matrix | the language grew *because* the application demanded it |

Lesson that repeats: **the application exposes the missing primitive; implement the primitive
generically, wire it through every layer, then write the application in DSL.** Every time the
shortcut (a Go loop, a string hack) was avoided, the primitive turned out to be reusable (the
function-form matrix, `subset_of`, `<+>`, `hamt_get_or`, `graph_scc` in comprehensions).
