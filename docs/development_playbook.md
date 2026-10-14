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
  packages import only the standard library and this module (plus an explicit allowlist: cobra/pflag and the YAML library in
  `pkg/cli` only), and no non-test identifier names a
  domain. Comments may still use a domain as an *example*.
- **Documentation links** (`TestDocLinksResolve`): every relative markdown link in the root docs,
  `docs/`, `.agents/` and the skills must resolve, and none may hard-code a machine-specific path.

`go test -run Example ./...` is redundant with `go test ./...` but is the fast loop for godoc.

Command-line changes live in `pkg/cli/command*.go` and are tested three ways: a **backward-compatibility
table** (`TestCommandTreeBackwardCompatible` — every invocation the old binary accepted or rejected, with exit
code and output), a subcommand table (`cmd.Execute` against `bytes.Buffer`s, no process), and a **real-binary
test** per `cmd/` (`cmd/gatwd/cli_test.go`, `cmd/gatwd-k8s/cli_test.go`). The tree holds no semantics: a new
behaviour goes into `Interpreter.Run` first. Capture statement output with `Interpreter.Out`, never by swapping
`os.Stdout`.

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
  dependency, unless the user decides otherwise. (The user did decide, later: D7 puts the kustomize
  library into the k8s shim, `pkg/k8sgraph.InProcessRenderer`, behind that very seam — plan §8.1.)
- **Source-agnostic downstream.** Live and offline sources yield the *same* payload shape (nested
  HAMT), so the same prelude runs on both. Prove it with a parity test over a fake client
  (`pkg/k8sgraph/k8sgraphtest` + `TestLiveOfflineParity`), never by inspection.
- **Name nodes `Kind/scope/name`.** A node is identified by one string, so every adapter's ids are
  qualified by kind and by scope (`Pod/default/web`, `Database/pg-prod/orders`, `DBMS/_/pg-prod`;
  `_` = no scope): equal local names under another kind or scope are then other nodes. Ingest
  with `graph_insert_node`, which turns a repeated id into an `[ERROR]` naming both declarations
  instead of a silent overwrite (`graph_add_node` is the upsert, for updates). A file-reading adapter
  attaches a `collections.Source` (file, document position) to each payload with `HAMT.WithSource`: a
  side channel that rules, equality and JSON never see (so live/offline parity is untouched), shown only
  in the duplicate error. Never put provenance into the payload as a key. The id is built in
  the *prelude* from the payload (`pkg/k8sgraph/lib/k8s/ids.gatwd`), never in the adapter, and a
  reference resolves in the referrer's scope (`k8s_ref`). Fixtures need an explicit scope on both
  the live and the offline side, or the two sources yield different ids and parity fails.
- **Merge sources generically.** `(a <+> b <+> c)` (`arrows.Plus` → `evaluator.StepPlus`) is the
  domain-free way to feed one pipeline from many producers; the domain only names the list
  (`lib/k8s/sources.gatwd`).
- **Tests use fakes and fixtures:** fake dynamic client, manifest files under `testdata/`, tests
  that need an external binary are skipped when it is absent. Never depend on a real cluster.

Minting a new Edition/Appliance: follow `.agents/skills/edition-appliance/SKILL.md`.

---

## 5a. Child repos: when a project moves outside this tree

Every Edition/Appliance built so far (`cmd/gatwd-k8s`, `cmd/gatwd-k8s-operator`, and formerly
`userland/*`) lived inside the gatwd tree. That stops working the moment a project needs its
own release cadence, its own hosting, or — the case that prompted this section — confidential
content that must never share gatwd's public commit history. The answer is a **child repo**: a
separate, independently-hosted repository that depends on gatwd as an ordinary Go module
(`go get github.com/joergreinhardt/gatwd`) rather than as a subtree of this one.

**What stays the same.** A child repo follows the same mandates as any Edition built in this
tree: pure assembly in `main.go` (§5), domain logic in `.gatwd` not Go, the Edition/Appliance/
shim pattern, exhaustive DSL tests, fakes-and-fixtures over real I/O. The generic/domain split
still applies *within* the child repo too — if a child repo needs a capability a second domain
could also use, that capability belongs in gatwd core (or a core-adjacent library), not
reinvented privately; see Mandate 5's "generic first, domain second."

**What's different: the agent-instructions file is a pointer, not a copy.** A child repo's own
`AGENTS.md`/`CLAUDE.md` does not paste in gatwd's mandates — they would rot the moment they're
copied. Instead it:

1. Names gatwd's `AGENTS.md`, `PROMPTS.md`, `STYLEGUIDE.md` and `.agents/skills/` as its base
   ruleset, by local path while gatwd is a sibling checkout (`../gatwd/AGENTS.md`), or by URL +
   pinned commit/tag once the child repo has moved away from a filesystem-adjacent gatwd.
2. Lists only its own deltas: its domain vocabulary, its confidentiality/hosting rule (below),
   any mandate it deliberately relaxes and why.
3. May mirror a short excerpt of the parts that matter operationally (the 8-mandate summary,
   the verification gate commands) inline for convenience when a path/URL reference isn't
   practical to re-read every session — clearly marked "mirrored from gatwd, not authoritative."

**The hosting/confidentiality rule, stated generically (gatwd names no domain or employer,
Mandate 5).** A child repo may be confidential — owned by an employer, a client, or otherwise
non-public. **A confidential child repo is never hosted on, or has any commit pushed to, a
public forge.** It lives on whatever private forge its owner designates, and is pushed only by
a human acting deliberately — never by an agent inferring that a generic "push" or "set up the
repo" instruction license a public remote. gatwd itself stays public and domain-free: no child
repo's name, employer, or domain vocabulary is ever written into gatwd's own tracked files.
*Which* repos are confidential, and where each may be hosted, is recorded by the operator
outside this tree (their own agent memory, the child repo's own privately-hosted docs) — never
here, which is exactly what keeps this file safe to keep public.

**Dependency style before the child repo has a real remote.** Use a placeholder module path and
a `replace` directive to the local gatwd checkout:

```
// child repo's go.mod
module example.invalid/placeholder-name   // swap for the real path once hosted

require github.com/joergreinhardt/gatwd v0.0.0
replace github.com/joergreinhardt/gatwd => ../gatwd
```

Switch to a real `require` with a pinned version once the child repo has a remote and gatwd has
a tagged (or otherwise pinned) commit the child repo is compatible with.

**Extracting an existing subtree into a child repo.** `git subtree split --prefix=<path> -b
<branch>` preserves the subtree's history as a standalone branch; build the new repo from that
branch, give it its own `go.mod` (above), add the pointer `AGENTS.md`/`CLAUDE.md`, verify it
builds and tests standalone, then remove the subtree from gatwd. If the subtree had *already*
been pushed to gatwd's public remote before the confidentiality need was caught, purging it
from gatwd's history (an interactive rebase or `git filter-repo`, followed by a force-push) is a
separate, explicitly-confirmed, high-blast-radius step on its own — never fold it silently into
the extraction, and never force-push shared history without the operator asking for it at the
time.

---

## 5b. Where a CLI verb belongs (core vs. domain shim vs. child repo)

A recurring question once a domain tool grows past its first few pipelines: a new CLI
"verb" (a named operation — filter, rank, diff, show-dependents, …) feels domain-specific
because it is used *on* domain objects, but is it actually generic? The test is not "does
it touch Kubernetes (or any domain) objects" — almost everything does, if the tool is for
that domain. The test is **does it need domain-shaped *data knowledge*, or only the
domain's already-generic container.**

**Why this works: Functor/Filterable/Foldable genericity.** Once a container has `map`,
`filter` and `fold` (collections/*, the Graph instance included), *every* query that is
really "pick some elements, maybe transform them, maybe reduce them" is already expressible
with those three combinators plus a predicate or accessor over the payload — no new
algorithm needed. "Find Services missing a label," "rank nodes by replica count," "show
everything with a hairpin dependency" are not new graph operations; they are
`graph_filter_nodes`/`graph_fold_nodes`/`graph_nodes_of` plus a `hamt_path`-shaped
predicate. The *topology* operations (reachability, degree, shortest path, SCC, unrolling a
cycle into a tree) are the genuinely new mathematical content, and they are domain-blind by
construction — a dependency graph does not care whether its nodes are Kubernetes resources,
Go packages, or people on an org chart.

That gives a three-way placement test, cheapest check first:

1. **Is it topology-only (data-blind)?** → **gatwd core** (`pkg/cli`/`collections`/
   `evaluator`), as a new matrix cell if the algorithm is genuinely missing (Mandate 2's
   full definition of done), or free if it already composes from existing generic ops
   (`sort`/`take` over a list of `(degree, id)` pairs needs no new "rank" verb — see
   `graph_degree`/`graph_out_degree`/`graph_in_degree`, added for exactly this reason: the
   *count* was missing, the *ranking* was not).
2. **Does it need domain-shaped data knowledge but nothing repo/employer-specific?** → the
   domain shim's own prelude (`pkg/k8sgraph/lib/k8s/*.gatwd`), as a **`.gatwd` def naming a
   predicate or accessor over the existing generic ops** — new Go code only when the
   knowledge is genuinely I/O (a new adapter, a new client call), never to reimplement a
   combinator the shim could call instead. Live-vs-offline source selection is the
   canonical example: it reads like it needs new plumbing, but `k8sgraph.LazyClient` (defers
   connecting until a live step actually runs) composing with `k8sgraph.ApplyEdition(e,
   client)` (client-or-nil) already means passing `LazyClient(DynamicClientFromKubeconfig)`
   instead of `nil` makes a *single* binary serve both `manifests` (offline) and
   `k8s_pods`/… (live) — no new flag, no new Go, because the two generic pieces already
   composed into the answer.
3. **Is it specific to one repo's conventions (an employer's DNS scheme, a directory
   layout, a particular overlay naming)?** → the child repo only (`AGENTS.md` §7), never
   upstreamed into the shim just because it is convenient to.

**The CLI-surface half of this is separately generic.** Whether a verb lands at tier 1, 2 or
3, *exposing* it as both a DSL task and a cobra subcommand is the same zero-logic wiring
every time: `cli.TaskCommand(edition, name, short, taskName)` promotes a `def` already in an
Edition's prelude to its own subcommand (`NAME name -- args`), beside the generic `NAME task
name -- args` it already had. Tier 1/2 verbs call it from `pkg/cli`/`pkg/k8sgraph` directly;
a child repo calls it the same way for its own tier-3 verbs. This is what "a generalized
cobra-subcommand and pipeline layer available to all CLIs" means in practice: one Go
function, reused at every tier, never a bespoke RunE per verb.

---

## 5c. Two sharpenings that came from using gatwd as a dependency, not just building it

These generalize from real usage of gatwd *as a library* underneath other tools — the
clearest test of whether a principle actually holds, since nobody is grading their own
homework. Both are domain-free by construction; the domains that prompted them are not
named here (Mandate 5) — only the shape.

### When "a second domain could use it" is enough to generalize now

Mandate 5's "generic first, domain second" needs a sharper test than "I can imagine
someone else wanting this," because imagination is cheap and a wrong abstraction is
expensive to unwind. Two conditions, and generalizing now is justified when **either**
holds:

1. **A second domain already exists and needs it.** The uncontroversial case — this is
   what "generic first" always meant.
2. **A second consumer is *committed*, not merely imaginable**: a named project on an
   actual backlog, not a "someone, someday might." A long-lived platform that expects to
   keep adding tools against the same handful of external systems (a version-control
   host, a datastore, an inventory system, a config-management target — whatever they are
   for the project in question) can reasonably treat "we will need an adapter for this
   again" as settled fact rather than speculation, *once it already has one working,
   concrete instance to generalize from*. Generalizing on that basis — before the second
   consumer is literally built — is correct engineering, not premature abstraction.

What is **not** justified either way: generalizing from a *single* instance with **no**
second consumer, committed or otherwise, and no backlog naming one. That is designing an
interface for an audience of one imagined reader — the classic premature-abstraction
trap, just dressed up in Mandate 5's language. The discipline is the same one phase
delivery already asks for (§2): ground the generalization in something real (an existing
second user, or a dated, named line in an actual backlog), not in "this feels reusable."

### Write adapters follow the same gated-capability shape, whatever the domain

`pkg/cli/policy.go`'s `Capability` system (`CapClusterApply`, `CapClusterRollout` are the
first two instances) is not a Kubernetes-specific mechanism that happens to live next to
one — it is the **general shape every write-capable adapter should follow**, regardless
of what external system it mutates:

1. A new named `Capability` constant, **denied by `TrustedPolicy` by default** — reading
   is cheap and safe by default; mutating something outside the process never is.
2. An explicit, per-Edition opt-in flag (`Edition.WithCapabilityFlag`) — the capability
   exists in the binary, but a script only gets it if the person running the command line
   asked for it that run, not because the Edition's author assumed it was fine.
3. **A dry-run default where the underlying system has the concept** (a server-side dry
   run, a diff instead of a write, a draft instead of a publish) — a second explicit flag
   turns the dry run off, mirroring the capability flag's own "opt in twice for something
   consequential" shape.

A future write adapter for some other external system — anything that can create,
modify, or delete state outside the current process — should reuse this exact shape
(new `Capability`, denied-by-default, `WithCapabilityFlag`, dry-run-first) rather than
inventing a bespoke gate per domain. The point of naming it as a pattern here, instead of
leaving it implicit in `pkg/k8sgraph/write.go`'s comments, is so the *next* write adapter
copies the shape on purpose instead of re-deriving it (or, worse, skipping the gate
because nothing in the docs said it was a general rule and not a Kubernetes quirk).

---

## 6. Writing DSL: conventions and traps

The project's own relations and rules are the style reference
(`pkg/k8sgraph/lib/k8s/*.gatwd`, `cmd/gatwd-k8s-operator/rules/compliance.gatwd`).

**Shapes that work**

- *Ingest* = a zero-parameter `def` wrapping a `fold` with a graph accumulator.
- *Relation* = a graph→graph pass: `fold (\_ g -> graph_add_edges [spec | …] g) graph_empty`
  (a graph stream element reaches the reducer as `g`). Chain passes with `|>`.
- *Rule* = a closure `\g -> [ [violation, resource] | … ]`; combine rules with `++`.
- *Helpers* = ordinary closures (`def k8s_ref = \kind ns name g -> […]`); they are values, so call them
  inside comprehensions.
- *Reconciliation* = compare two sources describing the same identity space (a designated
  state vs. a live one, or any two independently-derived views of the same resources) by
  recursive `subset_of` over a "stable view" of each (a domain-shaped projection that
  strips the fields each source alone is allowed to vary — e.g. the first instance of
  this, `k8s_stable_view`/`k8s_up_to_date`/`k8s_parity_report` in
  `pkg/k8sgraph/lib/k8s/parity.gatwd`, drops server-managed Kubernetes fields before
  comparing). The report is "present and matching" / "present but drifted" / "missing" —
  never a bare boolean, so a caller can tell "nothing to reconcile" from "didn't run."
  This shape is not specific to its first use: any two sources over the same identity
  space fit it, and a write-back correction from the reconciliation (if the system has
  one) is itself a write adapter and follows §5c's gated-capability shape, not a bespoke
  one.
- *Optimise with an index, in DSL.* The label-indexed `Selects` pass builds a HAMT of pod lists
  with `fold` + `hamt_set` + `hamt_get_or` — no Go needed.
- *Reusable view/query, source-agnostic.* A task script defines what to compute, never which
  source to compute it over: `def main = manifests |> gitops_graph |> gp_trees |> visualize`
  names no path. Run with `gatwd-<edition> task main -s script.gatwd -- /path/to/source`; swap
  sources by rerunning with a different argument, never by editing the script. This works because
  a pipeline-bodied def run as a task is a **fragment**: its input stream is the `--` arguments,
  one String node each — exactly what `manifests` (and `walk_fs` used with no inline path) reads
  its path from. See §5a's prevention-rule row and `.agents/skills/dsl-authoring`.

  A step that insists on a *literal* argument (the old `walk_fs <dir>`) cannot be parameterized
  this way — `-D` constants are substituted into expressions, never into a step's literal text, so
  `walk_fs REPO` would try to walk a directory literally named `REPO`. The fix is never to bake
  the one real path into the script; it is to give the step a stream-driven form (`walk_fs` with
  no inline argument reads the root from the upstream value) so the *existing* fragment mechanism
  parameterizes it for free. Mandate 1 applies here as much as anywhere: implement the missing
  primitive, don't hack the script around its absence.

**Traps (each cost real time; prevention rule in bold)**

| Symptom | Cause | Rule |
| :-- | :-- | :-- |
| `[f x, g y]` gives four elements | gatwd serialises lists space-separated; inside a *plain* `[...]` adjacent atoms are separate elements | **Parenthesise calls in list literals: `[(f x), (g y)]`.** Comprehensions and `( … )` re-enable application |
| Edges missing from listings | `Connect` does not create endpoint nodes | **Add nodes first, then edges**; guard with `graph_has_node` |
| A comprehension errors on a missing field | filters are positional (run when their variables are bound) | **Put `hamt_has_path` guards *before* the dependent generator** |
| `map (\g -> …)` changes payloads | `map` on a graph element is the graph Functor | **Whole-graph passes use `fold (\_ g -> …) seed`** |
| A `def` mis-evaluates its arguments | a `def` whose body is a *pipeline* is a re-parsed text macro | **Value functions are lambda-bodied (`def f = \x y -> …`); pipeline fragments are zero-parameter defs** |
| A "reusable" script only works for the one path it was written against | the source got typed into the `def` instead of read from the task's `--` arguments | **Never put a path/URL in a task script; make the step read it from the stream (`manifests`, `walk_fs` with no inline arg) and pass it with `-- <source>`** |
| `assert` reports "empty stream" | `assert` now fails when no node arrives (D9) | **Fix the pipeline that produced nothing; do not weaken the assert** |
| `n -1` is "cannot apply non-function value" | since W3 `-` is unary minus, and `f -3` (space before, none after) is a negative *argument* | **Write a subtraction with spaces: `n - 1`** (`-x`, `- 5`, `abs -3` all parse now) |
| `-7 // 2` / `-7 % 3` differ from Go | `//` floors and `%` takes the sign of the divisor (W3); `quot`/`rem` are the truncated forms | **Use `quot`/`rem` for Go's `/` and `%`; `a == b*(a//b) + a%b` always holds** |
| `"5" == 5` is `false`, `nth "1" xs` errors | strings are not numbers (D2 finished in W3) | **Parse with `str_to_int`/`str_to_float` (`Maybe`)**; a String still equals a *collection's* printed form (HAMT keys) |
| A partial application is run as a step | a step whose text evaluates to a *function* is not an expression value | **The infix-step detector ignores callables; keep it that way** |
| `to_set` set ≠ function-form set (**fixed**) | the step used to build `Set[string]` (printed form) and the function forms `Set[any]` (Go identity) | **Both build a `*collections.Setoid[any]` keyed by `==` (`evaluator.NewValueSet`); `==` across the two forms is fine. Never key a new set-like by `fmt.Sprintf("%v")`** |
| Test helper wrote nonsense | `echo '…\x…'` interprets escapes in zsh | **Write scratch scripts with `cat > f <<'EOF'`** |
| Results look odd in Go tests | `EvaluateToValues` treats its argument as one pipeline; `def` lines in a multi-line string are not registered | **Use `EvaluateScript` for setup, `EvaluateToValues` for the query** (empty input returns no values) |
| Result type surprises | lists come back as `*collections.List[any]` | **Use `functions.ToSliceAny`** |
| `[ x \| guard ]` (no `<-`) is an `[ERROR] … no generator` | a comprehension with **no generator** loops over nothing: it used to yield `[]` even when the guard held (`[ 1 \| true ]`), and `head ([ x \| guard ] ++ [d])` always gave `d`. It is now rejected at parse time, in every statement shape (`let`, `def`, lambda, `emit`), and `gatwd --check` reports it | **Bind the value with a generator first: `[ x \| q <- [r], guard q ]`** (`k8s_scope` in `lib/k8s/ids.gatwd`) **or a dummy one: `[ x \| _ <- [0], guard ]`**. A parity test that passed vacuously exposed it |
| `||>` output order differs run to run | `||>` fans the next step out on a worker pool (`evaluator.StepPar`) | **Sort or assert order-insensitively; use `|>` when order matters** |
| A `let`-bound HAMT/list/graph misbehaves in a head position | head-word text substitution used the lossy `%v` text | **Fixed: non-scalar runtime values are never substituted; do not re-add text substitution for them** |
| A step that needs quotes (`sh -c "…"`) lost them | exec steps used `strings.Fields` | **Fixed: `interactive.SplitCommandLine` tokenizes like a shell (no expansion); an unterminated quote is an error** |
| A top-level reference to a `def` holding a list/record resolves to nothing, or "undefined identifier" when passed as an argument | a non-lambda `def` (no `\`) is a **pipeline fragment** (a zero-parameter text macro expecting an input stream), not a reusable value, even when its body is a plain literal like `["a", "b"]` | **Use `const NAME = …` for a genuine constant value** (`tests/dsl_exhaustive_test.gatwd`'s `TAX_BASE`/`MAX_AGE`); reserve non-lambda `def` for pipeline fragments meant to be chained with `|>` |
| A `task`-run lambda (`cli.TaskCommand`/`RunOptions.Task` on a `\args -> …` def) fails with a misleading "unknown step or command" or "undefined identifier", naming something that plainly *is* defined | confirmed root cause (`docs/task_closure_pipe_fix_design.md`): `parser.ParseInfixExpr` — what `defClosure` (`pkg/cli/closures.go`) uses to detect a task-closure — cannot parse `\|>` **at all**, not even once (`ParseInfixExpr("\x -> x \|> f")` itself errors "unexpected trailing tokens"). `defClosure` swallows that error and returns `false`, and `RunTask` then reinterprets the lambda's literal source text as a pipeline-step name, which is never right for text starting with `\` — hence the misleading error. The same failure hits transitively: calling a *named* `\|>`-bodied def from inside an otherwise pipe-free closure body also fails, because resolving that name internally re-runs the same broken `defClosure` check | **A task-closure's body must contain zero `\|>` anywhere, including in any def it calls by name.** When a task genuinely needs one or more pipelines (reading a graph via `manifests`, say), skip `RunOptions.Task`/`cli.TaskCommand` and build a **top-level `let x = (… \|> …) in let y = (… \|> …) in f x y`** string instead, run via `RunOptions.Eval` — that shape *is* reliably resolved (every case in `pkg/k8sgraph/parity_test.go` uses it). See the design doc for the planned real fix (reusing `pipelineLambda`'s existing pipe-bodied-lambda machinery) |

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

- **Spec and editor pipeline.** `spec/spec.go` is the source for keywords, steps, types, operators
  and snippets; `go run ./cmd/gatwd-gen .` generates the editor plugins (Neovim, VS Code,
  tree-sitter) and `parser/tokens_gen.go` from it plus `cli.FunctionForms()`. In the same change
  as any syntax/keyword/step/builtin/snippet edit: update the spec, update `docs/grammar.md`,
  regenerate, review and commit the generated diff. `cmd/gatwd-gen/main_test.go` guards spec
  coverage and grammar-doc agreement. Procedure: `.agents/prompts/maintain-language-spec-and-editors.md`.

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
| 5 | offline manifests, kustomize via injectable renderer | user decision then: external binary, no heavy dep (superseded by D7: in-process library, same seam) |
| ingest | `fold` with a graph accumulator + **juxtaposition application** | a missing language primitive implemented properly (Mandate 1), not faked in Go |
| 6 | embedded pure-DSL prelude + `ApplyEdition` | relations in `.gatwd` |
| 7 | `cmd/gatwd-k8s`, `cmd/gatwd-k8s-operator` | pure-assembly main; hot-reloadable rules |
| 8 | `<+>` live source, relation matrix, six rules, perf | parity test, negative fixture, profile-then-fix |
| roadmap B–E | `Closure`, currying, gradual typing, function-form matrix | the language grew *because* the application demanded it |

Lesson that repeats: **the application exposes the missing primitive; implement the primitive
generically, wire it through every layer, then write the application in DSL.** Every time the
shortcut (a Go loop, a string hack) was avoided, the primitive turned out to be reusable (the
function-form matrix, `subset_of`, `<+>`, `hamt_get_or`, `graph_scc` in comprehensions).
