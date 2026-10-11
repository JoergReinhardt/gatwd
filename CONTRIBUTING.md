# Contributing to gatwd

Thank you for your interest in contributing to `gatwd`!

Before opening a pull request or asking an AI to generate code for this repository, you **must** read the following documents:

1. [STYLEGUIDE.md](./STYLEGUIDE.md) - This dictates our naming conventions (Category Theory) and our documentation mandate. All code must contain exceptionally educational `godoc` comments explaining the functional concepts used.
2. [PROMPTS.md](./PROMPTS.md) - Instructions and constraints for refactoring the architecture towards modern Go Generics and Iterators while preserving the BitFlag runtime registry.
3. [tests/dsl_exhaustive_test.gatwd](./tests/dsl_exhaustive_test.gatwd) - The native DSL test suite. All new syntax and features must be self-asserted here.
4. [docs/refactoring_plan.md](./docs/refactoring_plan.md) - The architectural roadmap for the project.

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

## The Multi-Layer Implementation Checklist ("All Layers Through")
When introducing support for Go standard library features or new transducers:
1. **Layer 0–4 Implementation:** Implement the pure Go abstraction with thorough unit tests and godoc.
2. **Layer 5 Builtin Wiring:** Wire the abstraction into `pkg/cli/cli.go` as a builtin step or adapter.
3. **DSL Exhaustive Test:** Add an assertion test case to `tests/dsl_exhaustive_test.gatwd`.
4. **Documentation Sync:** Document the new keyword in `docs/dsl_guide.md` and `docs/components.md`.

## Building Custom Editions, Servers & Kubernetes Operators
Do not import third-party domain packages (e.g. AWS, Kubernetes, specific databases) into the core `gatwd` module. Instead, author a **Custom Edition** binary or release:
1. **Choose Your Archetype**: Build a specialized CLI (`cmd/gatwd-<name>/main.go`), an HTTP/gRPC server daemon, or a Kubernetes controller/operator loop.
2. **Import Engine & CLI Packages**: Import `github.com/joergreinhardt/gatwd/pkg/cli` and lower-layer packages (`evaluator`, `collections`, `adapters`).
3. **Choose Runtime**: Instantiate `cli.NewInterpreter()` (standard builtins) or `cli.NewBareInterpreter()` (lean runtime).
4. **Register Domain Types & Steps**: Expose third-party types and transducers via `interp.RegisterType(...)`, `interp.RegisterStep(...)`, and `interp.RegisterConstant(...)`.
5. **Embed Gatwd Scripts (`//go:embed`)**: Use `//go:embed` with `interp.LoadScriptFS(fsys, path)` or `interp.EvaluateScript(content)` to ship self-contained logic without external file dependencies.
6. **Package Releases**: Compile static binaries (`CGO_ENABLED=0 go build`) or distroless container images, supporting live script overrides via arguments or Kubernetes ConfigMaps. See [`examples/custom_cli/main.go`](examples/custom_cli/main.go) and [`examples/kubernetes_repl/main.go`](examples/kubernetes_repl/main.go).



