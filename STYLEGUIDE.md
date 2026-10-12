# Style Guide: Gophers All The Way Down

This project walks the line between pure functional programming theory (Category Theory) and pragmatic, modern Go (1.18+ Generics and 1.23+ Iterators). 

When contributing to this project, adhere to the following rules:

## 1. Go Idiomatic First
While the concepts are functional, the syntax must remain strictly idiomatic Go. 
- Use standard Go formatting (`gofmt`).
- Error handling should use multiple return values or `Result[T]` types, avoiding panics.
- Use native Go language features (`iter.Seq`, Generics `[T any]`) rather than deep interface boxing or manual closures whenever possible.
- If there is a conflict between "how Haskell does it" and "how Go does it" (e.g., naming conventions, package layouts), **Go idiomatic conventions win**.

## 2. Naming: The Category Theory Theme
We intentionally use the vocabulary of functional programming and Category Theory to name our abstractions. This is the domain language of the project.
- **Do use:** `Functor`, `Applicative`, `Monad`, `Bind`, `Map`, `Fold`, `FlatMap`, `Maybe`, `Either`, `Sum`, `Product`.
- **Do not** invent Go-specific names for mathematically established concepts (e.g., do not name `Bind` as `ChainExecutions`). 

## 3. Mitigating "Smarty-Pants" Syndrome (The Golden Rule of Docs)
Category Theory vocabulary can be intimidating. We mitigate this through **exceptional, educational documentation**. 

Our standard is the Go standard library's `regexp` package, which beautifully teaches the reader what a regular language actually is. **Do not take functional concepts for granted.** 

### Documentation Requirements:
1. **Explain the Concept:** If you write a type called `Monad`, the `godoc` must explain what a Monad is conceptually in simple terms (e.g., "A context that wraps a value and allows sequencing dependent operations safely").
2. **Explain the Mechanics:** Explain *how* the implementation works (e.g., "This uses an `iter.Seq` to lazily evaluate the operations without blowing up the stack").
3. **Provide Examples:** Every core interface and function must include a practical, readable Go example showing why it is useful.

### Example of Good Documentation:
```go
// Bind (often called FlatMap) chains two operations together. 
//
// In functional programming, this is the defining operation of a "Monad". 
// Imagine you have a lazy sequence of Hosts, and you want to look up the 
// Network Interfaces for each Host. Because the lookup function itself returns 
// a sequence of Interfaces, a simple Map would result in a sequence of sequences.
//
// Bind solves this by applying the function and immediately flattening the 
// resulting sequences into a single continuous stream. 
// 
// Example:
//    // hosts is an iter.Seq[Host]
//    // getInterfaces is func(Host) iter.Seq[Interface]
//    interfaces := Bind(hosts, getInterfaces)
func Bind[T, U any](seq iter.Seq[T], f func(T) iter.Seq[U]) iter.Seq[U] {
    // ...
}
```

## 4. Test-Driven Mathematical Rigor
This project relies on mathematically proven Category Theory architectures (Monads, Comonadic Zippers, Kleisli arrows). We do not accept "dummy implementations" or mocked side-effects masquerading as structural features. 

**Testing Intention and Style:**
- **Aspire to Production-Readiness:** All tests must assume the goal is a fully production-ready system. Tests should exhaustively assert against edge cases inferred from the documentation and common-sense functional design.
- **Structural Go Tests over String Matching:** Do not test the core engine merely by parsing CLI output strings. You must write structural integration tests (like in `evaluator/pipeline_test.go`) that instantiate the raw Go memory arrays (`[]PipelineStep`) and hard-assert against the resulting memory types (`int`, `functions.Pair`, `*collections.Tree`).
- **DSL Native Assertions:** All DSL features must be tested natively using the Gatwd DSL `assert` keyword. If you add a new feature to the DSL, you must add it to `tests/dsl_exhaustive_test.gatwd` to ensure the mathematical engine can evaluate it natively without failing.
- **TDD (Test-Driven Design) Philosophy:** Always write tests that enforce the intended architectural purity of the code. If a pipeline expects a Tuple, assert that a native Go struct Tuple is yielded—not a formatted string. 
- **Variety of Test Data & True Genericity:** Test across multiple generic types—primitives (`int`, `float64`, `string`), custom structs, nested collections (`List`, `HAMT`), and error branches. Never assume generic code works with only integers.

## 5. Built-in Go Documentation System (godoc Mandate)
We strictly leverage Go's native documentation system (`go doc`, `pkgsite`, testable `Example*` functions).

1. **Every Exported Identifier Must Be Documented**:
   - Comments MUST begin with the exact name of the exported identifier (e.g. `// HAMT represents an immutable...`).
   - Clearly document all generic type parameters (e.g. `// K must satisfy comparable because...`).
   - Clearly document all input arguments and returned values (including tuples and `Either[error, any]`).
2. **Package Overviews (`doc.go`)**:
   - Every major package MUST have a dedicated `doc.go` explaining its role in the Gatwd architecture, its mathematical foundations, and a quick-start code snippet.
3. **Testable Examples (`Example*`)**:
   - Provide standard Go `Example*` functions in `*_example_test.go` with `// Output:` comments. These are validated by `go test` and rendered directly into generated godoc!
4. **Current State Integrity**:
   - Documentation must ALWAYS reflect the active codebase. Never refer to outdated structures that were migrated into `legacy/`.

## 6. The "All Layers Through" Mandate & Release Architecture

### The "All Layers Through" Principle
Lower-layer abstractions (Layers 0–4) must never remain stranded in Go API space without script-layer accessibility. Whenever a Go standard library capability, collection, transducer, or I/O adapter is built, it MUST be wired all the way through to the scripting language (Layer 5 `pkg/cli/`) as a **builtin**.

### Official Architectural Nomenclature
1. **Bare Runtime (`cli.NewBareInterpreter()`)**: A minimalist evaluation engine providing only host language primitives (`let`, `def`, `const`, `match`, `seq`, `emit`, `take`, `drop`, `head`, `tail`, `assert`). Builtins are disabled unless explicitly registered.
2. **Gatwd Builtins (`cli.NewInterpreter()`)**: Features derived from Go's standard library and Gatwd core abstractions (`math`, `collections`, `adapters`, `testing`, `visualizer`, `checkpoints`, `scanner`, `net`). Enabled by default in the full interpreter.
3. **Gatwd Editions (or Distributions)**: Standalone binary distributions (e.g., `cmd/gatwd-cloud`) that import `github.com/joergreinhardt/gatwd/pkg/cli`, authored via the fluent `cli.NewEdition()` builder.
4. **Gatwd Appliances**: Single-purpose, self-contained release artifacts (Kubernetes Operators, HTTP/gRPC streaming servers, compliance daemons, diagnostic CLIs) authored via `cli.NewAppliance()`, compiling embedded `.gatwd` DSL rules (`//go:embed`) and supporting live ConfigMap hot-reloading.

### Ergonomic Extension Guidelines
To prevent intimidating contributors with nested Category Theory closures (`iter.Seq`, `Either[error, any]`, `ASTNode`), extension code should prefer the high-level constructors:
- Use **`cli.MapStep(name, fn)`** for mapping values directly.
- Use **`cli.FilterStep(name, predicate)`** for stream predicate filtering.
- Use **`cli.ArgStep(prefix, fn)`** for parameterized pipeline steps.
- Use **`cli.SinkStep(name, fn)`** for non-mutating side-effects (logging, telemetry).
- Use **`cli.NewEdition(name)`** for building custom CLI distributions.
- Use **`cli.NewAppliance(...)`** for building production-grade services and operators.
Raw `StepHandler` closures should only be authored when custom trampoline short-circuiting or multiplexing is explicitly required.

## 7. Exhaustive Operator-Type Matrix Completeness

Where [§6](#6-the-all-layers-through-mandate--release-architecture) demands *vertical* completeness (every abstraction wired from Layer 0 up to the DSL), this rule demands *horizontal* completeness: **whenever you add or extend an operator, method, or combinator, implement it across the entire Cartesian product of types where it is mathematically and algebraically valid.** Think of the type system as a matrix — rows are types, columns are operations — and never ship a half-filled row or column.

### The Matrix

| Operation (column) | Concept | Valid across (rows) |
| :--- | :--- | :--- |
| `++` / `concat` | Monoid | `List`, `Chunk`, `Set` (union), `HAMT` (merge), `Tree`, `Pair`, `string`, `iter.Seq` |
| `map` / `bimap` / `mapValues` | Functor / Bifunctor | `List`, `Chunk`, `Set`, `HAMT`, `Tree`, `Pair`, `Either`, `Maybe`, `Variant`, streams |
| `fold` / `foldLeft` / `foldRight` | Foldable / Catamorphism | `List`, `Chunk`, `Set`, `HAMT`, `Tree`, `Pair`, `Either`, `Maybe`, streams |
| `filter` / `reject` | Predicate filtering | `List`, `Chunk`, `Set`, `HAMT`, `Tree`, streams |
| `flatMap` / `bind` | Monad | `List`, `Chunk`, `Set`, `Maybe`, `Either`, streams |
| `==`, `!=` | Structural Eq | all collections (order-independent for `Set`/`HAMT`), ADTs, primitives |
| `<`, `<=`, `>`, `>=` | Total / Partial Ord | primitives, `List`, `Chunk`, `Pair`, ADTs, streams (lexicographic) |
| `union`, `intersect`, `diff` | Set algebra | `Set`, `HAMT` (keyed set) |

Type rows span **persistent collections** (`List`, `Chunk`, `Set`, `HAMT`, `Tree`, `iter.Seq`), **ADTs** (`Pair`, `Either`, `Maybe`, `Variant`), and **scalar primitives** (`int`, `float64`, `big.Int`, `complex128`, `string`, `bool`, `rune`).

### The Rules
1. **No second-class citizens.** Never implement an operation for `List` without giving `Chunk`, `Set`, `HAMT`, `Tree`, and the eligible ADTs their corresponding algebraic implementation. `Variant` is a full ADT, not an afterthought.
2. **Only where it is mathematically well-defined.** Skip a cell *only* when the operation is genuinely meaningless for that type (e.g. ordering on an unordered `Set`). When you skip, it must be because of the math, not because it was tedious.
3. **A new cell is not "done" until it is complete through every layer.** The definition of done for one matrix cell is: the Go method on the type → the operator overload in `functions/operators.go` (if infix) → the evaluator transducer in `evaluator/pipeline.go` (if a stream op) → the DSL builtin in `pkg/cli/cli.go` → the self-asserting case in `tests/dsl_exhaustive_test.gatwd`.

### Anti-Patterns
- **Piecemeal operator support:** adding `fold` for `List` only, leaving `Chunk`/`Set`/`HAMT`/`Tree` folds unimplemented.
- **Go-but-not-DSL cells:** implementing a capability in Go (e.g. `HAMT.Filter`, `Bimap`) but never wiring it through to the DSL — the cell is only half-filled.
- **Silent asymmetry:** a type that quietly supports fewer operations than its siblings with no mathematical reason recorded.

This rule is enforced as **Mandate 2** in [`AGENTS.md`](AGENTS.md) and operationalized by the [`recursive-dogfood`](.agents/skills/recursive-dogfood/SKILL.md) skill.



