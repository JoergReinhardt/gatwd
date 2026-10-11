---
name: project-doc-sync
description: "Audit, synchronize, and enforce Go built-in documentation (godoc), package doc.go overviews, testable Example* tests, and eliminate outdated references to legacy code."
---

# Project Documentation Synchronization & Integrity (godoc)

## Principle

Documentation in `gatwd` is not an afterthought; it is an active, executable part of the codebase.
This skill guides contributors and AI agents to:
1. Enforce Go's native documentation system (`godoc`) across all packages.
2. Eliminate all drift where documentation references deprecated code that was moved to `legacy/`.
3. Provide package-level architectural overviews via `doc.go`.
4. Ensure every core abstraction is accompanied by testable `Example*` functions with `// Output:` comments.

## When to Activate

- Whenever introducing a new package, type, transducer, or adapter.
- After refactoring or moving deprecated prototypes into `legacy/`.
- During periodic documentation sweeps or release preparation.
- When requested by the user to improve or audit project documentation.

## Procedure

### Step 1: Scan for Legacy References in Documentation

Run ripgrep / grep across the `docs/` folder, `README.md`, and non-legacy `.go` files:

```bash
# Check if docs reference legacy packages or obsolete types
grep -rn "legacy/" docs/ README.md
grep -rn "github.com/joergreinhardt/gatwd/legacy" --exclude-dir=legacy/ .
```

If found, update the documentation to refer to the modern generic abstractions (`collections.List`, `collections.HAMT`, `evaluator.Engine`, `evaluator.PipelineStep`, `adapters.ScannerAdapter`, etc.).

### Step 2: Enforce godoc on Exported Identifiers

Check for exported symbols without comments:

```bash
# Verify all exported identifiers are documented
go doc ./...
```

For every exported type, method, function, and constant:
- Comment must start with the exact name of the identifier.
- Explain generic type parameters (`[T any]`, `[K comparable]`).
- Explain parameter meanings and return values.
- Explain the category-theoretic or mathematical concept in clear, educational terms.

### Step 3: Maintain Package `doc.go` Files

Every package (`collections/`, `evaluator/`, `adapters/`, `arrows/`, `comonads/`, `dsltypes/`, `pkg/cli/`) must have a `doc.go` file containing:
- High-level package comment (`// Package <name> provides...`).
- Architectural role in the dataflow pipeline.
- Theoretical foundations (Category Theory, persistent data structures, streaming iterators).
- Minimal quickstart snippet.

### Step 4: Write Testable Go Examples (`Example*`)

Create or update `*_example_test.go` files:
- Use standard Go testable example syntax (`func ExampleHAMT_Set()`, `func ExamplePipeline()`).
- Include the exact expected output in `// Output:` comments.
- Verify examples pass: `go test -run Example ./...`.

### Step 5: Verify with `go test` and `go vet`

```bash
go vet ./...
go test -run Example ./...
go test ./...
```
