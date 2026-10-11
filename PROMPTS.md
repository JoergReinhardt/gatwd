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

