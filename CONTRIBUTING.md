# Contributing to gatwd

Thank you for your interest in contributing to `gatwd`!

Before opening a pull request or asking an AI to generate code for this repository, you **must** read the following documents:

1. [AGENTS.md](./AGENTS.md) - The architecture manual, the layered hierarchy, the **eight mandatory development invariants** (Anti-Mockup, Operator-Type Matrix, All Layers Through, Recursive Dogfooding, Clean Core, Fail Loud, Pure Assembly, Phase-Gated Delivery), the verification gates and the prevention rules.
2. [docs/development_playbook.md](./docs/development_playbook.md) - *How* work is done here and why: the delivery protocol, verification gates, the Edition/Appliance/shim pattern, DSL traps, performance and handoff protocols, and the Kubernetes-graph case study.
3. [STYLEGUIDE.md](./STYLEGUIDE.md) - Naming conventions (Category Theory), the documentation mandate, test standards, the operator-type matrix (§7) and how to write DSL code (§8). All code must contain exceptionally educational `godoc` comments explaining the functional concepts used.
4. [PROMPTS.md](./PROMPTS.md) and [.agents/](./.agents/README.md) - Refactoring constraints, reusable task prompts and the in-repository skills.
5. [tests/dsl_exhaustive_test.gatwd](./tests/dsl_exhaustive_test.gatwd) - The native DSL test suite. All new syntax and features must be self-asserted here.
6. [docs/refactoring_plan.md](./docs/refactoring_plan.md), [docs/language_roadmap.md](./docs/language_roadmap.md) and [docs/k8s_graph_design.md](./docs/k8s_graph_design.md) - The roadmap and the worked design example.

We welcome contributions that push us closer to a fully lazy, dynamically-extensible, mathematically pure graph exploration DSL!

## The Contribution Workflow

1. **Pick the right prompt/skill** (`.agents/prompts/`, `.agents/skills/`) for your kind of change.
2. **Ground it in the code.** For anything beyond a bug fix, write or extend a design doc (`docs/<topic>_design.md`) with real signatures and file anchors. Split *generic* work (core) from *domain* work (shim/prelude) and land the generic part first.
3. **Ask at decision points.** If several designs are plausible and the choice is hard to reverse (a new dependency, a data representation, a merge strategy), present options with a recommendation and record the decision.
4. **Test first**, then implement: structural Go test on the raw memory type; a parser test for syntax; a self-asserting DSL case; a *negative* case; for rule sets a fixture that violates every rule.
5. **Implement through every layer** — Go method → operator → evaluator step → DSL step form → DSL function form → DSL assertion — and fill the whole operator-type matrix row/column.
6. **Run the gates** (below) after every sub-step.
7. **One logical commit per sub-step**: `feat(scope): summary (Phase N)` with a body listing what was added, decisions taken and the tests that prove it. Do not amend published commits, do not skip hooks, stage files by name.
8. **Update the docs in the same change**, and close the phase in the design doc (DONE, tests, remaining ideas, deferrals with reasons). Leave a handoff (`.agents/prompts/phase-handoff.md`) if work continues in another context.

## Verification Gates

```bash
gofmt -l <changed dirs>          # nothing may be listed for files you touched
go vet ./...                     # compile, vet, no import cycles
go test ./...                    # unit, structural, Example*, and internal/archtest
./run_dsl_tests.sh               # DSL suite: [ERROR], silent statements, --check errors
go test -race ./pkg/... ./evaluator ./parser   # when touching interpreter/parallel code
```

- `internal/archtest` fails the build if a library layer imports upward, if core imports a third-party package (the sole exception is cobra/pflag in `pkg/cli`), if core code names a domain (`kubernetes`, `k8s`, `conway`, …), or if a documentation link is broken or hard-codes a machine-specific path.
- Interpreter, parser or evaluator changes: also diff the output of every `*.gatwd` script against a baseline binary (build it from a `git worktree` of the previous commit) and make sure `gatwd --check` is silent over all of them. A few scripts differ between runs on `master` alone (Go map iteration order) — run the baseline more than once.
- **A test that cannot fail is worthless.** `assert` passes on an empty stream, so inspect every output line of a new DSL assertion; pair "no violations" tests with a violating fixture; break each new guard once to watch it fail.

## Recursive Dogfooding

After implementing any significant new abstraction, contributors (human or AI) **must** perform a "recursive dogfood" pass:

1. **Use `git log` to build a chronological timeline** of when each abstraction was introduced.
2. **Scan older code** (starting from the oldest files) asking: "Can this be simplified or strengthened using something we built later?"
3. **Verify no cyclic dependencies** are created (`go vet ./...`). Layer N may only import from Layers 0 through N-1.
4. **Write a test first** before replacing old code with newer abstractions.
5. **Run the full test suite** to confirm no regressions.

### Layer Dependency Order (never import upward):
```
Layer 0: functions/, dsltypes/                       (Either, Pair, Maybe, Variant, Closure, Node/Edge; type flags)
Layer 1: collections/                                (List, Chunk, Set, HAMT, Tree, BST, Graph)
Layer 2: arrows/, comonads/, optics/, transducers/, transformers/
Front end: parser/                                   (imports only Layers 0–1; used by evaluator and pkg/cli)
Layer 3: evaluator/                                  (Engine, PipelineStep, Adapter)
Layer 4: adapters/, interactive/, visualizer/, pkg/plugins/
Layer 5: pkg/cli/, pkg/<domain>/, cmd/               (Interpreter, builtins, Editions, Appliances, domain shims)
```
Verify with `go list -f '{{.Imports}}' ./<pkg>`; `internal/archtest` guards the library layers.

### Anti-patterns:
- Don't dogfood just because a newer tool exists — only when it genuinely improves correctness, clarity, or performance.
- Don't break import layering to force dogfooding.
- Every dogfood change must pass all tests. No exceptions.

## Documentation Standards (`godoc`)
All code must integrate seamlessly with Go's built-in documentation system:
1. **Full godoc Coverage:** Every exported type, method, function, and const must have descriptive doc comments starting with the identifier name.
2. **Explain Parameters & Return Values:** Describe generic type parameters (`[T any]`, `[K comparable]`), inputs, and output values.
3. **Package `doc.go`:** Each package must have a `doc.go` providing architectural overview and usage patterns.
4. **Current State:** Documentation must only describe active modern code, never structures in `legacy/`.

## Testing Standards
1. **Generic Type Diversity:** Tests must verify operations across primitives (`int`, `float64`, `string`), custom structs, and nested collections.
2. **Integration Chaining:** Build end-to-end integration tests chaining generators, filters, transforms, and sinks.
3. **Performance & Mocking:** Use mocked data generators dogfooding our own engine to test high-throughput pipelines.
4. **Tests Must Be Able To Fail:** assert on non-empty output, add negative fixtures next to "no violations" tests, negative-control every new guard, and never use `t.Logf` as a substitute for an assertion.
5. **Offline, Fake, Fixture:** I/O-backed features are tested against fakes and checked-in fixtures (e.g. `pkg/k8sgraph/k8sgraphtest`), prove live/offline **parity** by running the same prelude over both, and skip when an optional external binary is absent. Never require a real cluster or network.
6. **Testable Examples:** Write `Example*` functions with `// Output:` comments in `*_example_test.go` to provide simultaneously executable tests and documentation.
7. **DSL Exhaustive Test Suite (`tests/dsl_exhaustive_test.gatwd`):** Every new keyword, transducer, or language feature exposed to the DSL must have an explicit, self-asserting test (`|> assert <expected>`) in `tests/dsl_exhaustive_test.gatwd`. Run via `./run_dsl_tests.sh`.
8. **Visual & Pretty Output Testing:** Incorporate visualizer testing (`visualize`) to ensure output cards, ANSI color borders, and pretty formatting render correctly and deterministically. `evaluator.StepAssert` strips ANSI escape codes so formatted cards can be asserted cleanly without brittle color-code matching.

## The Multi-Layer Implementation Checklist ("All Layers Through")
When introducing support for Go standard library features or new transducers:
1. **Layer 0–4 Implementation:** Implement the pure Go abstraction with thorough unit tests and godoc.
2. **Layer 5 Builtin Wiring:** Wire the abstraction into `pkg/cli/cli.go` as a builtin step or adapter **and** register its curried, data-last, type-preserving **function form** in `pkg/cli/dslfuncs*.go` (`defn`) so lambdas, folds and comprehensions can call it with computed values.
3. **DSL Exhaustive Test:** Add assertion cases to `tests/dsl_exhaustive_test.gatwd` (step form, function form inside a lambda/comprehension, a negative case) and verify `./run_dsl_tests.sh` passes 100% — after reading every output line of your new section. Core primitives need a **non-domain** assertion.
4. **Visual Verification:** Include pretty visual printing or card formatting tests where applicable.
5. **Documentation Sync:** Document the new keyword in `docs/dsl_guide.md`, `docs/grammar.md` (if syntax), `docs/components.md`, and `README.md`; delete any "not yet supported" sentence the change made false.

## The Operator-Type Matrix Completeness Checklist (Orthogonal Parity)

Where the "All Layers Through" checklist guarantees *vertical* completeness (an abstraction wired from Layer 0 up to the DSL), this one guarantees *horizontal* completeness. **Whenever you add or extend an operator, method, or combinator, you must implement it across the entire Cartesian product of types where it is mathematically valid** — picture a matrix whose rows are types and whose columns are operations, and never ship a half-filled row or column. This is **Mandate 2** in [`AGENTS.md`](AGENTS.md) and [`STYLEGUIDE.md §7`](STYLEGUIDE.md); it is operationalized by the [`recursive-dogfood`](.agents/skills/recursive-dogfood/SKILL.md) skill.

**Type rows:** persistent collections (`List`, `Chunk`, `Set`, `HAMT`, `Tree`, `BST`, `Graph`, `iter.Seq` streams), ADTs (`Pair`, `Either`, `Maybe`, `Variant`), and scalar primitives (`int`, `float64`, `big.Int`, `complex128`, `string`, `bool`, `rune`).

**Operation columns:** `++`/`concat` (Monoid), `map`/`bimap`/`mapValues` (Functor/Bifunctor), `fold`/`foldLeft`/`foldRight` (Foldable), `filter`/`reject` (predicate), `flatMap`/`bind` (Monad), `==`/`!=` (Eq), `<`/`<=`/`>`/`>=` (Ord), `union`/`intersect`/`diff` (Set algebra).

When you touch any cell, complete it everywhere it is well-defined:

1. **Fill the whole row/column:** never implement `fold` (or `++`, `filter`, `<`, `bind`, …) for `List` without giving `Chunk`, `Set`, `HAMT`, `Tree`, and the eligible ADTs their corresponding algebraic implementation. `Variant` is a full ADT, not a second-class citizen.
2. **Skip only on mathematical grounds:** omit a cell *only* when the operation is genuinely meaningless for that type (e.g. ordering on an unordered `Set`) — never because it was tedious.
3. **Carry each cell through every layer:** Go method on the type → operator overload in `functions/operators.go` (if infix) → evaluator transducer in `evaluator/pipeline.go` (if a stream op) → DSL step form in `pkg/cli/cli.go` → DSL function form in `pkg/cli/dslfuncs*.go` → self-asserting cases in `tests/dsl_exhaustive_test.gatwd`. A cell implemented in Go but not reachable from the DSL is only half-filled.
4. **Record asymmetries:** if a type supports fewer operations than its siblings, the reason must be the math, and it should be obvious why.

## Building Gatwd Editions & Appliances
Do not import third-party domain packages (e.g. AWS, Kubernetes, specific databases) into the core `gatwd` packages — `internal/archtest` will fail the build. Instead author a **Gatwd Edition** (custom binary distribution) or a **Gatwd Appliance** (self-contained release). The Kubernetes graph tool is the reference implementation: `adapters/kubernetes_adapter.go` + `adapters/manifest_adapter.go` → `pkg/k8sgraph` (zero-logic shim, embedded prelude) → `cmd/gatwd-k8s` (Edition) → `cmd/gatwd-k8s-operator` (Appliance).
1. **Choose Your Role**: an **Edition** extends the language for arbitrary user scripts; an **Appliance** is a single-purpose CLI, daemon or Kubernetes Operator with embedded `.gatwd` rules.
2. **Domain logic is `.gatwd`**: vocabulary, ingest, relations and rules live in an embedded prelude (`Edition.WithPreludeFS`), never as Go loops. Go does only real I/O behind an adapter, lifted to a step with `cli.AdapterStep`.
3. **`main.go` is pure assembly** (Mandate 7): it registers only what core, `adapters/`, `pkg/plugins/` or the prelude supplies. Small testable helpers for flags/pipeline strings are fine.
4. **Use ergonomic constructors and fluent builders**: `cli.MapStep`, `cli.FilterStep`, `cli.ArgStep`, `cli.SinkStep`, `cli.NewEdition(name)`, `cli.NewAppliance(cli.WithEdition(...), cli.WithEmbeddedFS(...), cli.WithConfigMapOverride(...))`. Merge several producers into one stream with `(a <+> b <+> c)`.
5. **Prove it**: fixtures + a fake client, a live/offline **parity** test, a **negative fixture** violating every rule, a hot-reload test, and a benchmark on a synthetic workload (`.agents/skills/edition-appliance`).
6. **Package Releases**: compile static binaries (`CGO_ENABLED=0 go build`) or distroless images with live script overrides via arguments or ConfigMaps. See [`examples/custom_cli/main.go`](examples/custom_cli/main.go), [`examples/hot_reload_appliance/main.go`](examples/hot_reload_appliance/main.go) and [`cmd/gatwd-k8s-operator`](cmd/gatwd-k8s-operator/main.go).

## Maintaining the Language Spec & Editor Support
Syntax, keywords, steps, types and snippets are declared once in `spec/spec.go`. `go run ./cmd/gatwd-gen .` generates the Neovim, VS Code and tree-sitter support and `parser/tokens_gen.go` from it (registered function forms are picked up automatically). When you add or change any of these: edit the spec, update `docs/grammar.md`, run the generator, review the diff and commit the regenerated files with your change. Never hand-edit generated files (`DO NOT EDIT` headers). `go test ./cmd/gatwd-gen` fails if the spec misses a function form or step, or if `docs/grammar.md` and the spec disagree. Snippets must be valid, readable gatwd. Details: [`.agents/prompts/maintain-language-spec-and-editors.md`](.agents/prompts/maintain-language-spec-and-editors.md).

## Working in a Shared Tree
Another contributor or an AI session may be editing the same checkout. Re-read a file right before editing it, prefer additive edits, never revert changes you did not make, stage files by explicit path (never `git add -A`), and commit or push only when asked.
