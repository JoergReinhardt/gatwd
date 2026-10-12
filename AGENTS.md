# Gatwd Agent Guidelines & Architecture Manual

This document defines the architectural standards, development rules, and repository-specific skills for AI coding agents and human contributors working on `gatwd`.

---

## 1. Project Mission & Identity

`gatwd` is an industrial-strength, mathematically rigorous, pure functional programming language, execution engine, and dynamic infrastructure graph processor written in Go 1.23+.

- **Theoretical Foundation**: Category Theory (Monads, Comonads, Arrows, Bifunctors, Optics/Lenses, Catamorphisms) and Persistent Data Structures (HAMT, Cons Lists, Chunks, Sets, Trees).
- **Host Engine vs Guest DSL**: The internal engine relies on Go Generics (`[T any]`) and Go Iterators (`iter.Seq`). The dynamic guest DSL uses runtime type tokens and parser combinators.
- **Pure Functional Integrity**: Side effects are strictly isolated into stream adapters. State and transitions are modeled immutably.

---

## 2. Core Architectural Layers

All code adheres strictly to a unidirectional layered hierarchy. Layer $N$ may only import from Layers $0 \le k < N$:

| Layer | Package(s) | Role & Abstractions |
| :--- | :--- | :--- |
| **Layer 0** | `functions/`, `dsltypes/` | ADTs (`Either`, `Maybe`, `Pair`, `Variant`), stream operators, type flags |
| **Layer 1** | `collections/` | Persistent collections (`List`, `Chunk`, `Set`, `HAMT`, `Tree`) |
| **Layer 2** | `arrows/`, `comonads/`, `optics/`, `transducers/`, `transformers/` | Category-theoretic combinators, Kleisli arrows, zippers, lenses, prisms |
| **Layer 3** | `evaluator/` | Execution engine, stream pipelines (`PipelineStep`), step handlers |
| **Layer 4** | `adapters/` | Isolated I/O adapters (`fs.FS`, `bufio.Scanner`, `net.Listener`, `net.Conn`) |
| **Layer 5** | `parser/`, `pkg/cli/`, `cmd/` | Monadic parser combinators, builtins, interpreter, REPL, CLI |

---

## 3. Mandatory Development Invariants

### Mandate 1: The Anti-Mockup / Hard Work First Rule
- **No Bare-Minimum Passing**: Never implement superficial code or dummy stubs just to pass a test or appease a prompt.
- **Pure DSL-First Solutions**: When asked to implement an algorithm, application, or simulation (e.g. Conway's Game of Life, sorting, graph traversal), **NEVER** write a Go-side loop or fake Go wrapper and pretend it is a Gatwd solution.
- **Implement Missing Primitives First**: If the language or runtime lacks the necessary functional constructs (e.g., set comprehensions, frequency counting, lambda transforms, recursive let-bindings, higher-order transducers, continuations), do NOT hack around it with Go mockups. Immediately identify what is missing in the language, implement the true functional abstractions in the engine (`evaluator/`, `functions/`, `collections/`, `parser/`, `pkg/cli/`), wire them all the way through, and then write the true, pure Gatwd program.

### Mandate 2: Exhaustive Operator-Type Matrix Completeness
- **Orthogonal Parity Across All Types**: "Wiring all things up" must not be selective. Whenever an operator, method, or combinator is added or expanded (`++`, `map`, `fold`, `filter`, `compare`, `<`, `<=`, `>`, `>=`, `==`, `!=`, set operations), it MUST be implemented across the entire Cartesian product of data types where that operation is mathematically and conceptually valid:
  - Collections: `List`, `Chunk`, `Set`, `HAMT`, `Tree`, sequences (`iter.Seq`).
  - Algebraic Data Types (ADTs): `Pair`, `Either`, `Maybe`, `Variant`.
  - Primitives: numbers, strings, runes, booleans.
- **Never Leave Second-Class Citizens**: Never implement an operation for `List` without ensuring `Chunk`, `Set`, and `HAMT` have their corresponding algebraic implementations.

### Mandate 3: All-Layers-Through Synchronization
- Whenever a lower-level capability or persistent data structure is added or extended, it must be wired all the way through to Layer 5 (`pkg/cli/cli.go`) as a **Gatwd Builtin** and tested in `tests/dsl_exhaustive_test.gatwd`.
- Third-party domain dependencies (Kubernetes, AWS, databases) belong in **Gatwd Editions** (`cmd/gatwd-<name>`) or **Appliances**, never in core.

### Mandate 4: Recursive Dogfooding
- Every time new abstractions or utilities are added, systematically revisit older code to determine if it can be simplified, strengthened, or replaced by newer constructs (e.g., raw maps $\to$ `HAMT`, raw slices $\to$ `Chunk`, loops $\to$ `iter.Seq`/`FoldLeft`).

---

## 4. In-Repository Agent Skills

Repository-specific skills are located under `.agents/skills/`. Agents must read and execute these skills when performing their respective tasks:

1. **[`recursive-dogfood`](.agents/skills/recursive-dogfood/SKILL.md)**:
   - Systematically audit older code in chronological order to dogfood newer abstractions.
   - Enforce anti-mockup pure DSL development over bare-minimum Go wrappers.
   - Enforce full Cartesian operator-type matrix completeness.

2. **[`all-layers-sync`](.agents/skills/all-layers-sync/SKILL.md)**:
   - Audit lower-level Go abstractions (recent first).
   - Wire them all the way through to the DSL as builtins and grammar constructs.
   - Enforce exhaustive test coverage across all supported types in `tests/dsl_exhaustive_test.gatwd`.

3. **[`project-doc-sync`](.agents/skills/project-doc-sync/SKILL.md)**:
   - Enforce Go's native documentation system (`godoc`) across all exported symbols.
   - Maintain package-level architectural overviews via `doc.go`.
   - Write testable Go `Example*` functions with `// Output:` comments.
   - Eliminate outdated references to legacy code.

---

## 5. Verification Commands

Always run the full verification suite before submitting or committing changes:

```bash
# 1. Run all native Go unit tests
go test ./...

# 2. Check for compile and vet errors
go vet ./...

# 3. Run the exhaustive DSL integration test suite
./run_dsl_tests.sh
```
