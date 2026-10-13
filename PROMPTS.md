# AI and Developer Refactoring Instructions

When tasked with refactoring, extending, or maintaining this codebase, you must adhere strictly to the architectural paradigm and style guidelines of `gatwd`. 

Read these instructions carefully before proposing or committing code changes.

## 1. Architectural Alignment
`gatwd` began as an educational functional Go library (pre-Go 1.18); that migration to a generic, iterator-based engine is **complete**. Today's work is extending a modern **Infrastructure Graph Query Engine and functional language** — see [`AGENTS.md`](AGENTS.md) for the layers and the eight mandates, and [`docs/development_playbook.md`](docs/development_playbook.md) for how work is done here.
- **Host Engine vs Guest DSL:** The internal engine must rely completely on Go Generics (`[T any]`) and Go Iterators (`iter.Seq`). The `BitFlag` system is strictly reserved for evaluating the dynamically parsed DSL at runtime.
- **Isolate Side Effects (The IO Monad):** Never execute raw side effects (SSH, file I/O) directly in the evaluation loops. They must be isolated in Adapters that return a lazy stream (`iter.Seq[ASTNode]`).
- **Parallelism (Functional Concurrency):** When mapping over unbounded sequences, prefer parallel abstractions (like `ParMap`) powered by goroutines, safely leveraging the immutable nature of the AST.

## 2. Coding Style & Naming
- Read and follow `STYLEGUIDE.md`. 
- You must use Haskell/Category Theory vocabulary (`Monad`, `Bind`, `Applicative`, `Functor`, `SumType`) for core abstractions.
- You must write idiomatic, clean Go 1.23+ code (no panic-driven control flow, proper error handling, standard go formatting).

## 3. The Documentation Mandate
The most critical requirement for any generated code is the **educational quality of its documentation**.
- Treat every `godoc` comment as an opportunity to teach a junior developer Functional Programming theory.
- Never assume the reader knows what an "Algebraic Data Type" or "Applicative Functor" is. Break it down using the exact pedagogical approach seen in Go's `regexp` documentation.
- **Required Structure for Core Abstraction Docs:**
  1. What is the mathematical concept?
  2. Why do we need it here (practical problem it solves)?
  3. How is it implemented idiomatically in Go?
  4. A brief usage example.

## 4. Refactoring Process
The pre-generics migration (reflection wrappers, `Expression`/`Native`/`Group` interfaces, `for head, tail = tail.Continue()` loops) is finished and lives in `legacy/` for git history only. For any refactor today:
1. Preserve the `BitFlag` registry — it is the guest DSL's runtime type system and lives in `dsltypes/`. Generics are for the host engine; `BitFlag` is for the dynamically parsed DSL.
2. Prefer generic type parameters over `any` in host code, and `for v := range seq` over manual iteration.
3. Write the pinning test **first** (the behavior you must not change), then refactor, then run the full gates (`AGENTS.md` §5) and — for interpreter/parser/evaluator changes — the script regression diff (`docs/development_playbook.md` §3).
4. Dogfood: after adding an abstraction, revisit older code that can use it (`.agents/skills/recursive-dogfood`). Recent examples: the three copy-pasted lambda-name lookups in `map`/`filter`/`flatmap` collapsed into one `graphLookup`; `map`/`filter` step bodies extracted into `mapStepOver`/`filterStepOver` so closures share them.

## 5. Mathematical Rigor and Structural Testing
When implementing new features or migrating existing ones, you must adopt a **Test-Driven Design (TDD) approach**.
- **No Dummy Code or Mocks:** If you implement a functional concept (like `FanOut` or `Split`), do not implement a naive string-splitting parser or OS-fallback hack. You must implement the true mathematical structure in the base layers (`arrows/`, `comonads/`, `evaluator/`).
- **Exhaustive Structural Assertions:** You must write tests that bypass the string CLI and directly construct the generic structures in Go, asserting exactly on the memory types yielded by the `iter.Seq` (e.g. `TestKleisli`, `TestFanOut`, `TestStepMap`).
- **DSL Native Testing:** Every DSL-facing feature must be added to `tests/dsl_exhaustive_test.gatwd` using the `assert` keyword. Make sure the DSL itself can mathematically verify its own execution.
- **Fail CI on Regression:** Design your tests so that if a future agent attempts to lazily replace a mathematical combinator with a bash-script execution, the tests will physically fail to compile or hard-crash.

## 6. Project Documentation Integrity & No Legacy Drift
- **Zero Legacy Drift:** Documentation in `docs/` and exported code comments must NEVER reference structures that were moved into `legacy/`. All architectural overviews and examples must run against the modern generic and iterator-based engine.
- **Package `doc.go` Overviews:** Every package must maintain a top-level `doc.go` documenting its role in the dataflow pipeline.
- **Go Testable Examples:** Combine testing and documentation by writing `Example*` functions with `// Output:` comments in `*_example_test.go` files.

## 7. The "All Layers Through" Mandate & Custom Editions
- **Full Stack Wiring:** Never stop at Layer 1–4 when implementing Go standard library integrations or transducers. You MUST expose them all the way through to Layer 5 (`pkg/cli/cli.go`) as **builtins**.
- **DSL Verification:** Add every newly exposed builtin to `tests/dsl_exhaustive_test.gatwd` and verify with `./run_dsl_tests.sh`.
- **Custom Editions for Third-Party Libraries:** Never add third-party domain dependencies (Kubernetes, AWS, Postgres, etc.) directly into `gatwd` core. Guide users and agents to create Custom Editions (`cmd/gatwd-<name>`) importing `pkg/cli` and registering types/steps via `RegisterStep` and `RegisterType`.

## 8. The Mandates (single source of truth: `AGENTS.md`)
The rules that earlier versions of this file duplicated now live only in [`AGENTS.md`](AGENTS.md) §3, so they cannot drift:
- **Mandate 1** Anti-Mockup / Hard Work First — implement the missing primitive, then write the program in the DSL.
- **Mandate 2** Exhaustive Operator-Type Matrix — every type × operation cell, including the curried function form.
- **Mandate 3** All Layers Through — wired to `pkg/cli` as a builtin, asserted in `tests/dsl_exhaustive_test.gatwd`, with a non-domain assertion for every core primitive.
- **Mandate 4** Recursive Dogfooding.
- **Mandate 5** Clean Core — enforced by `internal/archtest`.
- **Mandate 6** Fail Loud — no silent failures; tests must be able to fail.
- **Mandate 7** Pure Assembly — Editions/Appliances/shims wire; logic lives in the DSL.
- **Mandate 8** Phase-Gated, Decision-Recorded Delivery.

## 9. Test Intent You Must Preserve
- A test that cannot fail proves nothing. Assert on **non-empty** output; add a **negative fixture** next to every "no violations" test; **negative-control** every new guard by breaking it once.
- Structural Go tests assert raw memory types (`*collections.Graph`, `functions.Pair`), not formatted strings; DSL tests use `assert` and run through `./run_dsl_tests.sh`.
- Tests use fakes and fixtures (fake dynamic client, manifest files); they never need a real cluster, network or external binary.

## 10. Task Prompts
Ready-made, self-contained prompts for recurring tasks are in [`.agents/prompts/`](.agents/prompts/README.md): add a language feature, fill a matrix cell, add a DSL relation or rule, mint an Edition/Appliance, investigate performance, write a phase handoff, sync documentation. Start from the matching prompt, fill in the blanks, and keep its "definition of done".

## 11. Agent Skills (`.agents/skills/`)
`recursive-dogfood`, `all-layers-sync`, `project-doc-sync`, `dsl-authoring`, `edition-appliance`, `phase-delivery`. Read the relevant `SKILL.md` before starting the matching kind of task.
