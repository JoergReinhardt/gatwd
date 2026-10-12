# Contributing to gatwd

Thank you for your interest in contributing to `gatwd`!

Before opening a pull request or asking an AI to generate code for this repository, you **must** read the following documents:

1. [AGENTS.md](./AGENTS.md) - The architecture manual and the five mandatory development invariants, including **Mandate 2: Exhaustive Operator-Type Matrix Completeness** (orthogonal operator parity across every type) and the "All Layers Through" and Clean Core rules.
2. [STYLEGUIDE.md](./STYLEGUIDE.md) - This dictates our naming conventions (Category Theory), our documentation mandate, and the operator-type matrix completeness rule (§7). All code must contain exceptionally educational `godoc` comments explaining the functional concepts used.
3. [PROMPTS.md](./PROMPTS.md) - Instructions and constraints for refactoring the architecture towards modern Go Generics and Iterators while preserving the BitFlag runtime registry.
4. [tests/dsl_exhaustive_test.gatwd](./tests/dsl_exhaustive_test.gatwd) - The native DSL test suite. All new syntax and features must be self-asserted here.
5. [docs/refactoring_plan.md](./docs/refactoring_plan.md) - The architectural roadmap for the project.

We welcome contributions that push us closer to a fully lazy, dynamically-extensible, mathematically pure graph exploration DSL!

## Recursive Dogfooding

After implementing any significant new abstraction, contributors (human or AI) **must** perform a "recursive dogfood" pass:

1. **Use `git log` to build a chronological timeline** of when each abstraction was introduced.
2. **Scan older code** (starting from the oldest files) asking: "Can this be simplified or strengthened using something we built later?"
3. **Verify no cyclic dependencies** are created (`go vet ./...`). Layer N may only import from Layers 0 through N-1.
4. **Write a test first** before replacing old code with newer abstractions.
5. **Run the full test suite** to confirm no regressions.

### Layer Dependency Order (never import upward):
```
Layer 0: functions/       (Either, Pair, Option)
Layer 1: collections/     (List, Tree, Set, Map, HAMT, Chunk)
Layer 2: arrows/, monads/, comonads/  (Category theory combinators)
Layer 3: evaluator/       (Pipeline, Steps, Adapters)
Layer 4: adapters/        (FS, Scanner, Writer)
Layer 5: pkg/cli/         (Parser, Interpreter, DSL)
```

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
4. **Testable Examples:** Write `Example*` functions with `// Output:` comments in `*_example_test.go` to provide simultaneously executable tests and documentation.
5. **DSL Exhaustive Test Suite (`tests/dsl_exhaustive_test.gatwd`):** Every new keyword, transducer, or language feature exposed to the DSL must have an explicit, self-asserting test (`|> assert <expected>`) in `tests/dsl_exhaustive_test.gatwd`. Run via `./run_dsl_tests.sh`.
6. **Visual & Pretty Output Testing:** Incorporate visualizer testing (`visualize`) to ensure output cards, ANSI color borders, and pretty formatting render correctly and deterministically. `evaluator.StepAssert` strips ANSI escape codes so formatted cards can be asserted cleanly without brittle color-code matching.

## The Multi-Layer Implementation Checklist ("All Layers Through")
When introducing support for Go standard library features or new transducers:
1. **Layer 0–4 Implementation:** Implement the pure Go abstraction with thorough unit tests and godoc.
2. **Layer 5 Builtin Wiring:** Wire the abstraction into `pkg/cli/cli.go` as a builtin step or adapter.
3. **DSL Exhaustive Test:** Add an assertion test case to `tests/dsl_exhaustive_test.gatwd` and verify `./run_dsl_tests.sh` passes 100%.
4. **Visual Verification:** Include pretty visual printing or card formatting tests where applicable.
5. **Documentation Sync:** Document the new keyword in `docs/dsl_guide.md`, `docs/components.md`, and `README.md`.

## The Operator-Type Matrix Completeness Checklist (Orthogonal Parity)

Where the "All Layers Through" checklist guarantees *vertical* completeness (an abstraction wired from Layer 0 up to the DSL), this one guarantees *horizontal* completeness. **Whenever you add or extend an operator, method, or combinator, you must implement it across the entire Cartesian product of types where it is mathematically valid** — picture a matrix whose rows are types and whose columns are operations, and never ship a half-filled row or column. This is **Mandate 2** in [`AGENTS.md`](AGENTS.md) and [`STYLEGUIDE.md §7`](STYLEGUIDE.md); it is operationalized by the [`recursive-dogfood`](.agents/skills/recursive-dogfood/SKILL.md) skill.

**Type rows:** persistent collections (`List`, `Chunk`, `Set`, `HAMT`, `Tree`, `iter.Seq` streams), ADTs (`Pair`, `Either`, `Maybe`, `Variant`), and scalar primitives (`int`, `float64`, `big.Int`, `complex128`, `string`, `bool`, `rune`).

**Operation columns:** `++`/`concat` (Monoid), `map`/`bimap`/`mapValues` (Functor/Bifunctor), `fold`/`foldLeft`/`foldRight` (Foldable), `filter`/`reject` (predicate), `flatMap`/`bind` (Monad), `==`/`!=` (Eq), `<`/`<=`/`>`/`>=` (Ord), `union`/`intersect`/`diff` (Set algebra).

When you touch any cell, complete it everywhere it is well-defined:

1. **Fill the whole row/column:** never implement `fold` (or `++`, `filter`, `<`, `bind`, …) for `List` without giving `Chunk`, `Set`, `HAMT`, `Tree`, and the eligible ADTs their corresponding algebraic implementation. `Variant` is a full ADT, not a second-class citizen.
2. **Skip only on mathematical grounds:** omit a cell *only* when the operation is genuinely meaningless for that type (e.g. ordering on an unordered `Set`) — never because it was tedious.
3. **Carry each cell through every layer:** Go method on the type → operator overload in `functions/operators.go` (if infix) → evaluator transducer in `evaluator/pipeline.go` (if a stream op) → DSL builtin in `pkg/cli/cli.go` → self-asserting case in `tests/dsl_exhaustive_test.gatwd`. A cell implemented in Go but not reachable from the DSL is only half-filled.
4. **Record asymmetries:** if a type supports fewer operations than its siblings, the reason must be the math, and it should be obvious why.

## Building Gatwd Editions & Appliances
Do not import third-party domain packages (e.g. AWS, Kubernetes, specific databases) into the core `gatwd` repository. Instead, author a **Gatwd Edition** (custom binary distribution) or a **Gatwd Appliance** (self-contained release):
1. **Choose Your Role**: Build a **Gatwd Edition** (extending the language with third-party libraries for arbitrary user scripts) or a **Gatwd Appliance** (single-purpose CLI, server daemon, or Kubernetes Operator with embedded `.gatwd` rules).
2. **Import Engine & CLI Packages**: Import `github.com/joergreinhardt/gatwd/pkg/cli` and lower-layer packages (`evaluator`, `collections`, `adapters`).
3. **Use Ergonomic Step Constructors**: Use `cli.MapStep(name, fn)`, `cli.FilterStep(name, pred)`, and `cli.ArgStep(prefix, fn)` instead of writing nested `iter.Seq` closures.
4. **Use Fluent Builders**: Build distributions via `cli.NewEdition(name)` and hot-reloadable appliances via `cli.NewAppliance(cli.WithEmbeddedFS(...), cli.WithConfigMapOverride(...))`.
5. **Package Releases**: Compile static binaries (`CGO_ENABLED=0 go build`) or distroless container images, supporting live script overrides via arguments or Kubernetes ConfigMaps. See [`examples/custom_cli/main.go`](examples/custom_cli/main.go) and [`examples/hot_reload_appliance/main.go`](examples/hot_reload_appliance/main.go).



