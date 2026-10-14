# Gatwd Agent Guidelines & Architecture Manual

This document defines the architectural standards, development rules, and repository-specific skills for AI coding agents and human contributors working on `gatwd`.

---

## 1. Project Mission & Identity

`gatwd` is an industrial-strength, mathematically rigorous, pure functional programming language, execution engine, and dynamic infrastructure graph processor written in Go 1.23+.

- **Theoretical Foundation**: Category Theory (Monads, Comonads, Arrows, Bifunctors, Optics/Lenses, Catamorphisms) and Persistent Data Structures (HAMT, Cons Lists, Chunks, Sets, Trees).
- **Host Engine vs Guest DSL**: The internal engine relies on Go Generics (`[T any]`) and Go Iterators (`iter.Seq`). The dynamic guest DSL uses runtime type tokens and parser combinators.
- **Pure Functional Integrity**: Side effects are strictly isolated into stream adapters. State and transitions are modeled immutably.
- **Language surface today**: first-class curried closures with lexical scope (`\x y -> …`, partial application, data-last pipeline application), an expression-level function form for every cell of the operator-type matrix, optional `def f : A -> B` annotations with an advisory gradual type checker (`gatwd --check`), constructor pattern matching with exhaustiveness warnings, a generic typed/colored/weighted `Graph` collection with its full DSL row, the `<+>` (ArrowPlus) stream-merge combinator, and Editions/Appliances with embedded DSL preludes. The Kubernetes graph tool (`pkg/k8sgraph`, `cmd/gatwd-k8s*`) is the reference for how a domain is built on top without touching core.

---

## 2. Core Architectural Layers

All code adheres strictly to a unidirectional layered hierarchy. Layer $N$ may only import from Layers $0 \le k < N$:

| Layer | Package(s) | Role & Abstractions |
| :--- | :--- | :--- |
| **Layer 0** | `functions/`, `dsltypes/` | ADTs (`Either`, `Maybe`, `Pair`, `Variant`), `Closure`/`DSLFunc` callables, graph `Node`/`Edge`, stream operators, type flags |
| **Layer 1** | `collections/` | Persistent collections (`List`, `Chunk`, `Set`, `HAMT`, `Tree`, `BST`, `Graph`) |
| **Layer 2** | `arrows/`, `comonads/`, `optics/`, `transducers/`, `transformers/` | Category-theoretic combinators (Kleisli `Compose`/`FanOut`/`Split`/`Plus`), zippers, lenses, prisms |
| **Front end** | `parser/` | Monadic parser combinators and the infix expression language. *Physically* it imports only Layers 0–1 and is consumed by `evaluator/` and `pkg/cli/`, so it sits between Layer 2 and Layer 3 even though it is the language front end |
| **Layer 3** | `evaluator/` | Execution engine, stream pipelines (`PipelineStep`), step handlers |
| **Layer 4** | `adapters/`, `interactive/`, `visualizer/`, `pkg/plugins/` | Isolated I/O adapters (`fs.FS`, `bufio.Scanner`, `net`, Kubernetes dynamic client, manifests), terminal visualizer, plugin registry |
| **Layer 5** | `pkg/cli/`, `pkg/<domain>/`, `cmd/` | Builtins, interpreter, REPL, Edition/Appliance builders; thin domain shims (`pkg/k8sgraph`); Edition and Appliance binaries |

The *direction* of imports is what matters, and it is machine-checked: `go vet ./...` rejects
cycles, and `internal/archtest` enforces the library-layer boundaries and the Clean Core rule
(Mandate 5). Nothing in Layers 0–2 may import `parser/`, `evaluator/` or anything above.

---

## 3. Mandatory Development Invariants

### Mandate 1: The Anti-Mockup / Hard Work First Rule
- **No Bare-Minimum Passing**: Never implement superficial code or dummy stubs just to pass a test or appease a prompt.
- **Pure DSL-First Solutions**: When asked to implement an algorithm, application, or simulation (e.g. Conway's Game of Life, sorting, graph traversal), **NEVER** write a Go-side loop or fake Go wrapper and pretend it is a Gatwd solution.
- **Implement Missing Primitives First**: If the language or runtime lacks the necessary functional constructs (e.g., set comprehensions, frequency counting, lambda transforms, recursive let-bindings, higher-order transducers, continuations), do NOT hack around it with Go mockups. Immediately identify what is missing in the language, implement the true functional abstractions in the engine (`evaluator/`, `functions/`, `collections/`, `parser/`, `pkg/cli/`), wire them all the way through, and then write the true, pure Gatwd program.

### Mandate 2: Exhaustive Operator-Type Matrix Completeness
- **Orthogonal Parity Across All Types**: "Wiring all things up" must not be selective. Whenever an operator, method, or combinator is added or expanded (`++`, `map`, `fold`, `filter`, `compare`, `<`, `<=`, `>`, `>=`, `==`, `!=`, set operations), it MUST be implemented across the entire Cartesian product of data types where that operation is mathematically and conceptually valid:
  - Collections: `List`, `Chunk`, `Set`, `HAMT`, `Tree`, `BST`, `Graph`, sequences (`iter.Seq`).
  - Algebraic Data Types (ADTs): `Pair`, `Either`, `Maybe`, `Variant`.
  - Primitives: numbers, strings, runes, booleans.
- **Never Leave Second-Class Citizens**: Never implement an operation for `List` without ensuring `Chunk`, `Set`, and `HAMT` have their corresponding algebraic implementations.
- **Definition of done for one cell** (a type × operation): Go method on the type → operator overload in `functions/operators.go` (if infix) → evaluator step in `evaluator/` (if a stream op) → DSL **step form** in `pkg/cli/cli.go` → DSL **function form** in `pkg/cli/dslfuncs*.go` → a self-asserting case in `tests/dsl_exhaustive_test.gatwd` that exercises the function form *inside a lambda or comprehension* → structural Go test asserting the raw memory type → `Example*`. Skip a cell only for a mathematical reason, and record the reason.
- **The function form** is data-last, curried (register with `defn`, giving a `functions.Closure`), type-preserving (never stringify elements) and returns errors as values. Step form and function form must behave identically; if they cannot, give the function a distinct name. Carrier dispatch lives once, in `pkg/cli/valueops.go` — do not re-implement per-carrier loops in a new builtin.

### Mandate 3: All-Layers-Through Synchronization
- Whenever a lower-level capability or persistent data structure is added or extended, it must be wired all the way through to Layer 5 (`pkg/cli/cli.go` for steps, `pkg/cli/dslfuncs*.go` for expression functions) as a **Gatwd Builtin** and tested in `tests/dsl_exhaustive_test.gatwd`.
- Every new *core* primitive needs a **non-domain** DSL assertion. If the only test that exercises it is a Kubernetes (or any domain) test, it is not generic enough yet.
- Third-party domain dependencies (Kubernetes, AWS, databases) belong in **Gatwd Editions** (`cmd/gatwd-<name>`) or **Appliances**, never in core.

### Mandate 4: Recursive Dogfooding
- Every time new abstractions or utilities are added, systematically revisit older code to determine if it can be simplified, strengthened, or replaced by newer constructs (e.g., raw maps $\to$ `HAMT`, raw slices $\to$ `Chunk`, loops $\to$ `iter.Seq`/`FoldLeft`).

### Mandate 5: Clean Core Prevention Rule (Generic Core vs Domain Isolation)
- **Zero Domain Pollution in Core**: Never add use-case-specific or domain-specific types, keywords, or pipeline steps directly into core packages (`functions/`, `dsltypes/`, `collections/`, `arrows/`, `comonads/`, `optics/`, `transducers/`, `transformers/`, `evaluator/`, `parser/`, `pkg/cli/`). For example, never introduce Conway-specific steps (`StepConway`, `LifeGrid`), simulation hacks, or database/Kubernetes commands into core. Core MUST remain strictly mathematical, category-theoretic, and universally generic.
- **What exactly is forbidden, and how it is enforced**: in core, (a) no import outside the standard library and this module — **except** the explicit allowlist in `internal/archtest/clean_core_test.go`, currently **cobra and pflag (a command-line framework is plumbing, not domain; kubectl, Helm and kind are built on it) and `go.yaml.in/yaml/v3` (YAML is a data format the standard library lacks; decision 2026-10-13), and only in `pkg/cli`**; the lower layers stay standard-library only — and (b) no non-test identifier that names a domain (`kubernetes`, `k8s`, `kube`, `conway`, …). A comment or doc example *mentioning* a domain is fine; code that *knows* one is not. `internal/archtest` (`TestCleanCoreImportsNoThirdParty`, `TestCleanCoreNoDomainIdentifiers`) fails the build on violation — extend its word list when a new domain appears, never add an exemption to make a violation pass.
- **Generic first, domain second**: when a domain needs something a second domain could also use (a stream merge, a prelude hook, a subset test, a path accessor), implement it *generically* in core and wire it through every layer, then let the domain shim use it. This repeatedly paid off (`<+>`, `subset_of`, `hamt_path`, `Edition.WithPreludeFS`, `cli.AdapterStep`). **"A second domain could use it" is a real test, not a license to generalize from one data point**: it is satisfied either by a second domain that already exists, or by a *committed* (named, on an actual backlog) future consumer — not by an imagined one. Generalizing on a committed-backlog basis, before the second consumer is literally built, is correct; generalizing on a merely-imaginable basis, with zero concrete second consumer, risks guessing the wrong abstraction from a sample of one. See `docs/development_playbook.md` §5c for the longer version and the write-capability pattern this same test applies to.
- **The Go Standard Library Rule (Generic Upward Wiring)**: When a specific application reveals that a capability from the Go Standard Library (e.g. `io`, `bufio`, `net`, `math/big`, `crypto`, `time`, `sync`) is missing in Gatwd, implement it **generically** across all layers (Layers 0 through 5) so it becomes a first-class, general-purpose Gatwd builtin available to all programs.
- **Third-Party Domain Dependencies (Adapters/Plugins for Editions & Appliances) — "Batteries Included"**: Third-party Go dependencies (Redis, Postgres, Kafka, Kubernetes, AWS, etc.) must NEVER be imported into core `gatwd` packages. Implement them as **Adapters / Plugins** inside the repository, **each domain in its own Go package** (`adapters/<domain>/`, never sharing a package with another domain or with the generic `adapters` package core imports — see `docs/development_playbook.md` §5's table for the precise split and `docs/adapter_roadmap.md` for why this is a named, tracked principle, not just a convention), registrable in `pkg/cli` via `RegisterStep` and `RegisterType`. Use them to mint **Gatwd Editions** (`cmd/gatwd-<name>`) or **Gatwd Appliances** (single-purpose distributions with embedded `.gatwd` code). The name is deliberate: gatwd grows a wide, ever-expanding shelf of ready-made domain adapters in this one repository, the same way a "batteries included" language ships a large standard library — but **no Edition pays, in DSL keyword space or compiled size, for a domain it didn't pick up**. That second half is enforced by the per-domain package boundary, not assumed from it.

### Mandate 6: Fail Loud — Nothing Silent, and Every Test Must Be Able to Fail
- **No silent failures.** A step, builtin or call that cannot do its job emits an error node / returns an error value that reaches the Left channel (`[ERROR] …`). It never yields nothing, `nil`, `0` or the unchanged input. A failure message names *what* failed (the missing key's path segment, the unknown step, the undeclared constructor).
- **A pipeline that yields nothing is not a passing pipeline.** `assert` on a stream that delivers no node at all is an error (`assertion on an empty stream`), and `./run_dsl_tests.sh` also fails on `(No output or node unhandled)` as well as on `[ERROR]`. A rule that legitimately returns `[]` is still indistinguishable from a broken pipeline, hence the negative fixtures below. When adding DSL tests, run your new section on its own and read every output line.
- **Pair "no violations" with a negative fixture.** A compliant system returns `[]`, which looks exactly like a broken pipeline. Any rule set needs a fixture that violates *every* rule and a test expecting exactly those violations.
- **Negative-control every new guard** (a test, a lint, a gate): introduce a deliberate violation, watch it fail, remove it. Never use `t.Logf` where an assertion belongs.
- **Advisory is not silent.** Static checks (`gatwd --check` typing findings) are warnings, but they must be deterministic and must stay at zero false positives on every script in the repo; `Any` (unknown) is consistent with everything.

### Mandate 7: Pure Assembly — Logic Lives in the DSL, Mains Only Wire
- **Editions, Appliances and domain shims assemble; they do not compute.** A `cmd/gatwd-*` `main.go` registers only what core, `adapters/`, `pkg/plugins/` or an embedded `.gatwd` prelude supplies (`WithStep`, `WithType`, `WithPreludeFS`, `//go:embed`). No domain algorithm, loop, relation rule or type definition inline. Tiny testable helpers for flags and pipeline strings are fine.
- **Domain logic is `.gatwd`.** Node/edge vocabularies, ingest, relations and compliance rules are authored in the embedded prelude/rules, never as Go loops faking a graph. Go does only real I/O (client libraries, files, network) behind an adapter.
- **The shim has zero logic** (`pkg/k8sgraph`: construct an adapter, hand it to generic `cli.AdapterStep`; `ApplyEdition` wires steps + prelude; a `nil` client builds an offline-only edition).
- **Parity is proven, not assumed.** If two sources (live cluster, offline manifests) must yield the same graph, add a test that runs the same prelude over both through fakes (`pkg/k8sgraph/k8sgraphtest`). Tests never need a real cluster, network or external binary (gate those on availability).
- See `.agents/skills/edition-appliance/SKILL.md` for the procedure.

### Mandate 8: Phase-Gated, Decision-Recorded Delivery
- **Work larger than a bug fix is delivered in small gated phases** with a design doc grounded in the real code, generic prerequisites before domain phases, one commit per sub-step, and the full verification gates (§5) green at every step.
- **Stop at real decision points.** When several designs are plausible and the choice is hard to reverse, present the options with a recommendation, ask, and record the decision with its date in the design doc. Already-delegated work ("pick up from here") may proceed on the recommended option — say which and why.
- **Deferrals are written down** in the phase entry with the reason. Never leave a matrix cell, relation or feature silently missing.
- **Hand off in writing** when a context ends mid-project (`.agents/prompts/phase-handoff.md`); mark superseded handoffs; close each phase by updating the design doc entry and checklist.
- **Shared working tree.** Other people or sessions may be editing the same tree. Re-read before editing, make additive edits, never revert changes you did not make, stage files by name, and commit or push only when asked.
- See `.agents/skills/phase-delivery/SKILL.md` and `docs/development_playbook.md`.

---

## 4. In-Repository Agent Skills

Repository-specific skills are located under `.agents/skills/`. Agents must read and execute these skills when performing their respective tasks:

1. **[`recursive-dogfood`](.agents/skills/recursive-dogfood/SKILL.md)**:
   - Systematically audit older code in chronological order to dogfood newer abstractions.
   - Enforce anti-mockup pure DSL development over bare-minimum Go wrappers.
   - Enforce full Cartesian operator-type matrix completeness.

2. **[`all-layers-sync`](.agents/skills/all-layers-sync/SKILL.md)**:
   - Audit lower-level Go abstractions (recent first).
   - Wire them all the way through to the DSL as builtins and grammar constructs.
   - Enforce exhaustive test coverage across all supported types in `tests/dsl_exhaustive_test.gatwd`.

3. **[`project-doc-sync`](.agents/skills/project-doc-sync/SKILL.md)**:
   - Enforce Go's native documentation system (`godoc`) across all exported symbols.
   - Maintain package-level architectural overviews via `doc.go`.
   - Write testable Go `Example*` functions with `// Output:` comments.
   - Eliminate outdated references to legacy code, hard-coded counts and stale handoffs.

4. **[`dsl-authoring`](.agents/skills/dsl-authoring/SKILL.md)**:
   - Write pure-DSL programs, relations and rules in the project's idioms (closures, folds, comprehensions, function forms).
   - Avoid the known traps (list-literal application, node-before-edge, positional guards, vacuous asserts).

5. **[`edition-appliance`](.agents/skills/edition-appliance/SKILL.md)**:
   - Mint an Edition/Appliance as `adapter → thin shim → embedded prelude → pure-assembly main`, with live/offline parity and negative-fixture tests.

6. **[`phase-delivery`](.agents/skills/phase-delivery/SKILL.md)**:
   - Run larger work as gated phases: design doc, generic prerequisites, decision points, deferrals, handoffs, perf protocol.

Reusable task prompts (add a language feature, fill a matrix cell, add a relation or rule, investigate performance, write a handoff, sync docs) live in [`.agents/prompts/`](.agents/prompts/README.md). The long-form "how we work here, and why" is [`docs/development_playbook.md`](docs/development_playbook.md).

---

## 5. Verification Commands

Always run the full verification suite before submitting or committing changes:

```bash
# 0. Format what you touched (nothing may be listed)
gofmt -l <changed dirs>

# 1. Compile + vet + no import cycles
go vet ./...

# 2. Native Go tests: unit, structural, Example*, and internal/archtest
#    (layer boundaries, Clean Core, doc links). `go test ./...` already runs the Examples.
go test ./...

# 3. The exhaustive DSL suite. Fails on [ERROR], on any statement that produced
#    no output (vacuous assert), and on `gatwd --check` structural errors.
./run_dsl_tests.sh

# 4. When you touched pipeline/parallel/interpreter code
go test -race ./pkg/... ./evaluator ./parser
```

Interpreter, parser or evaluator changes additionally need the **script regression diff** and a clean `gatwd --check` over every `*.gatwd` (see `docs/development_playbook.md` §3).

---

## 4a. Language Spec & Editor-Support Pipeline

`spec/spec.go` is the single hand-edited source for the language's surface vocabulary (keywords, steps, types, operators, snippets). `go run ./cmd/gatwd-gen .` turns it, plus the registered function forms (`cli.FunctionForms()`), into the Neovim, VS Code and tree-sitter files under `editors/` and `parser/tokens_gen.go`. Generated files carry a `DO NOT EDIT` header; never edit them by hand.

- **Whenever you add or change syntax, a keyword, a step, a builtin or a snippet**: update `spec/spec.go` (and `docs/grammar.md` for syntax), run the generator, review the diff, and commit generated files together with the spec change. Follow `.agents/prompts/maintain-language-spec-and-editors.md`.
- `cmd/gatwd-gen/main_test.go` fails when the spec misses a registered function form or pipeline step, or when `docs/grammar.md` and the spec disagree. Negative-control any new coverage guard.
- Snippets are documentation: keep them valid, current and in the most readable form (clause defs, guards); check them with `gatwd --check`.
- The generator is domain-free like core: domain vocabulary (e.g. `k8s_*`) comes from an Edition's own prelude, not from the spec.

---

## 5a. Prevention Rules (traps that already cost time)

Each row is a mistake that happened; the rule is what prevents the next one. The long table with causes is in `docs/development_playbook.md` §6.

| Do | Don't |
| :-- | :-- |
| Parenthesise calls inside plain list literals: `[(f x), (g y)]` | Write `[f x, g y]` (gatwd re-parses lists space-separated; you get four elements) |
| Add graph **nodes first**, then edges; guard edges with `graph_has_node` | Assume `Connect`/`graph_add_edges` creates endpoints |
| Put `hamt_has_path` guards *before* the dependent generator | Rely on filters running after all generators |
| Use `fold (\_ g -> … g) graph_empty` for whole-graph passes | Use `map` on a graph element (that is the graph Functor over payloads) |
| Write value functions as lambda-bodied `def f = \x y -> …` | Expect a pipeline-bodied `def` to take arguments (it is a re-parsed text macro) |
| Use `EvaluateScript` for setup and `EvaluateToValues` for one query | Put `def` lines inside an `EvaluateToValues` string |
| Use `||>` only before a per-element step (map, filter, a command); expect any output order | Put a whole-stream step (`fold`, `sort`) after `||>` — it sees single nodes |
| Check the output of every new DSL assertion | Trust `assert` on a stream that may be empty |
| Pair a rule set with a violating fixture | Test only the compliant case (`[]` is also what a broken pipeline returns) |
| Name a function differently when it cannot match its step form | Overload a name so step and function disagree |
| Re-read a file immediately before editing it | Overwrite a file another session is changing |
| Write scratch scripts with `cat > f <<'EOF'` | `echo '…\x…' > f` (zsh interprets escapes) |
| Give core a generic primitive and put the domain in a shim/prelude | Put a domain word, import or step in core |
| Change `spec/spec.go`, then run `go run ./cmd/gatwd-gen .` and commit the regenerated files | Hand-edit anything under `editors/` or `parser/tokens_gen.go`, or add a keyword/step/builtin without the spec entry |
| Make a reusable task script read its source from the `--` arguments (`manifests`, `walk_fs` with no inline arg); swap sources by rerunning with a different argument | Put a path/URL/directory inside a task's `def` — it then only ever works against the one source it was written for |
| Use `const NAME = …` for a genuine constant value (a list, a record) | Use a non-lambda `def` for a value — it's a pipeline fragment (a text macro), not a constant, and silently yields nothing or "undefined identifier" |
| Keep a task-closure's (`cli.TaskCommand`/`RunOptions.Task` on a `\args -> …` def) body free of `\|>` entirely, including in any def it calls by name | Put even one `\|>` anywhere in a task-closure's body (or call a `\|>`-bodied def by name from it) — `parser.ParseInfixExpr` cannot parse `\|>` at all (confirmed: `docs/task_closure_pipe_fix_design.md`), so `defClosure` silently falls back to misinterpreting the lambda's own source as a pipeline-step name; combine multiple pipelines via a top-level `let x = (… \|> …) in let y = (… \|> …) in f x y` run with `RunOptions.Eval` instead |

---

## 6. Code Navigation & Refactoring Tooling (use the type-aware Go tools)

This is a large, heavily-generic codebase (e.g. `pkg/cli/cli.go` is ~3k lines; the
same method name — `Union`, `Map`, `Fold`, `Filter` — exists on `List`, `Chunk`,
`Set`, `HAMT`, `Tree`, and `Graph`). Prefer Go's **type-aware** tooling over
`grep`/`sed` for navigation and refactoring: text search cannot tell which type a
method belongs to, misses multi-line signatures, and makes renames unsafe. The
toolchain, `gofmt`, `goimports`, and `gopls` are all available in this environment.

**Navigation — prefer these to reading whole files or grepping:**

```bash
go doc ./collections Graph.AddNode     # authoritative signature + doc, no file read
go doc ./pkg/cli                       # package overview + exported symbol index
gopls definition      path/file.go:LINE:COL   # jump to a symbol's declaration
gopls references      path/file.go:LINE:COL   # EVERY use, module-wide, type-aware
gopls implementation  path/file.go:LINE:COL   # concrete impls of an interface
gopls call_hierarchy  path/file.go:LINE:COL   # callers/callees of a function
gopls symbols         path/file.go             # outline of one file
gopls workspace_symbol Name                    # find a symbol across the module
```

Use `gopls references` to find all call sites before changing a signature, and to
check a proposed new identifier does not collide — it is authoritative where `grep`
guesses.

**Refactoring:**

```bash
gopls rename -w path/file.go:LINE:COL NewName   # type-aware, module-wide safe rename
```

Prefer `gopls rename` over an editor find-replace / `Edit replace_all` for renaming
any exported or widely-used identifier: it renames only the right binding across
every package, never a same-named method on a different type or a substring match.

**After every code change, before the §5 verification suite:**

```bash
gofmt -w <changed files>       # or: go fmt ./<changed pkg>
goimports -w <changed files>   # fix/order imports
```

---

## 7. Child Repos

A **child repo** is a separate, independently-hosted repository that depends on gatwd as a
library (`go get github.com/joergreinhardt/gatwd`) instead of living under this tree. It is
the right home for a project that needs its own release cadence or hosting constraints — in
particular, one that is confidential and must never share gatwd's public commit history.

A child repo still follows gatwd's mandates, but does not copy this file: its own
`AGENTS.md`/`CLAUDE.md` is a short pointer naming gatwd's `AGENTS.md`, `PROMPTS.md`,
`STYLEGUIDE.md` and `.agents/skills/` as its base ruleset, plus only its own deltas (domain
vocabulary, confidentiality/hosting rule, any mandate it deliberately relaxes and why).

**The hosting rule, stated generically (gatwd names no domain, Mandate 5):** a confidential
child repo is never hosted on, or has a commit pushed to, a public forge. It lives on
whatever private forge its owner designates, pushed only by a human. Which repos are
confidential, and where they may go, is recorded by the operator outside this tree — never
inside gatwd's tracked files.

Full detail (extraction procedure, dependency style before a remote exists, pointer-file
shape): `docs/development_playbook.md` §5a.

**Deciding where a new CLI verb belongs** (core vs. a domain shim like `pkg/k8sgraph` vs. a
child repo) is a related, recurring question with its own test — `docs/development_playbook.md`
§5b.

These complement — do not replace — the standard file-editing tools; reach for them
whenever a task involves finding call sites, confirming a signature, checking for
name collisions, or renaming.
