# Contributing to gatwd

Thank you for your interest in contributing to `gatwd`!

Before opening a pull request or asking an AI to generate code for this repository, you **must** read the following documents:

1. [STYLEGUIDE.md](./STYLEGUIDE.md) - This dictates our naming conventions (Category Theory) and our documentation mandate. All code must contain exceptionally educational `godoc` comments explaining the functional concepts used.
2. [PROMPTS.md](./PROMPTS.md) - Instructions and constraints for refactoring the architecture towards modern Go Generics and Iterators while preserving the BitFlag runtime registry.
3. [tests/dsl_exhaustive_test.gatwd](./tests/dsl_exhaustive_test.gatwd) - The native DSL test suite. All new syntax and features must be self-asserted here.
4. [docs/refactoring_plan.md](./docs/refactoring_plan.md) - The architectural roadmap for the project.

We welcome contributions that push us closer to a fully lazy, dynamically-extensible, mathematically pure graph exploration DSL!
