# AI and Developer Refactoring Instructions

When tasked with refactoring, extending, or maintaining this codebase, you must adhere strictly to the architectural paradigm and style guidelines of `gatwd`. 

Read these instructions carefully before proposing or committing code changes.

## 1. Architectural Alignment
You are migrating an educational functional Go library (pre-Go 1.18) into a modern **Infrastructure Graph Query Engine**.
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
When refactoring old files (like `functions/composition.go` or `data/conversion_table.go`):
1. Do not indiscriminately delete the `BitFlag` logic; it must be preserved and moved to a dedicated `dsltypes` package.
2. Rip out the heavy `Expression`, `Native`, and `Group` interface wrappers for host operations and replace them with generic type parameters.
3. Replace custom `for head, tail = tail.Continue()` loops with `for value := range seq` using `iter.Seq`.

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

## 8. The Anti-Mockup / Hard Work First Mandate
- **No Bare-Minimum Passing:** Never implement the bare minimum just to make a unit test pass or dodge a difficult design challenge.
- **Pure DSL-First Solutions:** When asked to implement algorithms or domain programs in Gatwd (e.g. Conway's Game of Life, sorting, tree traversal, simulations), NEVER write Go-level mockups, hardcoded sample sequences, or external shortcuts and claim it runs in Gatwd.
- **Implement Missing Primitives First:** If essential language or engine features are missing to implement the algorithm cleanly in pure DSL, *immediately identify the missing primitives, implement the true functional abstractions in the engine and DSL first*, and then write the program purely in the DSL.
- **Industrial-Strength Goal:** The target is a production-ready, mathematically rigorous functional programming language and interpreter. Every feature must be genuinely implemented, general, and robust.

## 9. Exhaustive Operator-Type Matrix Completeness
- **Orthogonal Parity Across All Types:** "Wiring all things up" must not be selective. Whenever an operator, method, or combinator is added or expanded (`++`, `map`, `fold`, `filter`, `compare`, `<`, `<=`, `>`, `>=`, `==`, `!=`, set operations), it MUST be implemented across the entire Cartesian product of data types where that operation is mathematically and conceptually valid:
  - Collections: `List`, `Chunk`, `Set`, `HAMT`, `Tree`, sequences (`iter.Seq`).
  - Algebraic Data Types (ADTs): `Pair`, `Either`, `Maybe`, `Variant`.
  - Primitives: numbers, strings, runes, booleans.
- **Never Leave Second-Class Citizens:** Never implement an operation for `List` without ensuring `Chunk`, `Set`, and `HAMT` have their corresponding algebraic implementations.

## 10. In-Repository Agent Skills (`.agents/skills/`)
All agent skills and operational instructions are maintained directly within the repository under `.agents/skills/` so that any AI assistant or human collaborator can discover and execute them:
- `.agents/skills/recursive-dogfood/SKILL.md`: Audit older code for replacement with newer abstractions, enforce anti-mockup pure DSL development, and verify operator-type matrix completeness.
- `.agents/skills/all-layers-sync/SKILL.md`: Audit lower-level Go abstractions and wire them all the way through to the DSL as builtins with exhaustive DSL tests.
- `.agents/skills/project-doc-sync/SKILL.md`: Maintain package `doc.go` overviews, godoc on all exported identifiers, and testable `Example*` functions.

## 11. Clean Core Prevention Rule (Generic Core vs Domain Isolation)
- **Zero Domain Pollution in Core:** Never add use-case-specific or domain-specific types, keywords, or pipeline steps directly into core packages (`evaluator/`, `functions/`, `collections/`, `parser/`, `pkg/cli/`). For example, never introduce Conway-specific steps (`StepConway`, `LifeGrid`), simulation hacks, or database/Kubernetes commands into core. Core MUST remain strictly mathematical, category-theoretic, and universally generic.
- **The Go Standard Library Rule (Generic Upward Wiring):** When a specific application reveals that a capability from the Go Standard Library (e.g. `io`, `bufio`, `net`, `math/big`, `crypto`, `time`, `sync`) is missing in Gatwd, implement it **generically** across all layers (Layers 0 through 5) so it becomes a first-class, general-purpose Gatwd builtin available to all programs.
- **Third-Party Domain Dependencies (Adapters/Plugins for Editions & Appliances):** Third-party Go dependencies (Redis, Postgres, Kafka, Kubernetes, AWS, etc.) must NEVER be imported into core `gatwd` packages. Implement them as **Adapters / Plugins** inside the repository (e.g. under `adapters/` or `pkg/plugins/`), registrable in `pkg/cli` via `RegisterStep` and `RegisterType`. Use them to mint **Gatwd Editions** (`cmd/gatwd-<name>`) or **Gatwd Appliances** (single-purpose distributions with embedded `.gatwd` code) without bloating core.




