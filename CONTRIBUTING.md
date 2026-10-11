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
