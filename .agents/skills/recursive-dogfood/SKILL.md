---
name: recursive-dogfood
description: "Eat Your Own Dogfood Recursively: systematically identify older code that can be improved by newer abstractions, enforce pure DSL-first implementations over bare-minimum Go mockups, and ensure full operator-type orthogonal matrix completeness across all data types."
---

# Recursive Dogfooding & Matrix Completeness Skill

## Principle

**"Eat your own dogfood recursively"** means: every time new abstractions, utilities, or patterns are added to the codebase, systematically revisit older code to determine if it can be simplified, strengthened, or replaced by the newer constructs — unless doing so would create cyclic dependencies.

In `gatwd`, this principle is augmented by two foundational mandates:
1. **The "No Mockups / Hard Work First" Mandate**: Never implement a bare-minimum Go mockup or external workaround when an application reveals missing language or engine features. Implement the true functional primitives first, then build the application purely in the DSL.
2. **Exhaustive Operator-Type Matrix Completeness**: Every operator and combinator must apply across the entire Cartesian product of data types where it is conceptually and algebraically valid. Never leave any data type as a second-class citizen.

This is a **structural integrity and language completeness pass** ensuring the codebase converges toward a production-ready, mathematically rigorous functional programming language and engine.

---

## The Core Mandates

### 1. The Anti-Mockup / Hard Work First Mandate
- **No Bare-Minimum Passing**: Do not write the bare minimum code just to make a unit test pass or dodge a complex design problem.
- **Pure DSL-First Implementation**: When asked to implement an algorithm or application in Gatwd (e.g. Conway's Game of Life, Quicksort, graph algorithms, stream processors), NEVER write a Go-side loop or fake Go wrapper and pretend it is a Gatwd solution.
- **Implement Missing Primitives First**: If the language or runtime lacks the necessary functional constructs (e.g., set comprehensions, frequency counting, lambda transforms, recursive let-bindings, higher-order transducers, continuations), do NOT hack around it with Go mockups. Immediately identify what is missing in the language, implement the true functional abstractions in the engine (`evaluator/`, `functions/`, `collections/`, `parser/`, `pkg/cli/`), wire them all the way through, and then write the true, pure Gatwd program.

### 2. Exhaustive Operator-Type Matrix Completeness
Gatwd maintains an explicit taxonomy of types and operators. Whenever a type or operator is touched, verify the full matrix:

#### Data Types (Rows)
- **Persistent Collections**: `List[T]`, `Chunk[T]`, `Set[T]`, `HAMT[K, V]`, `Tree[T]`, lazy streams (`iter.Seq[T]`)
- **Algebraic Data Types (ADTs)**: `Pair[A, B]`, `Either[L, R]`, `Maybe[T]`, `Variant`
- **Scalar Primitives**: `int`, `float64`, `big.Int`, `complex128`, `string`, `bool`, `rune`

#### Operations & Combinators (Columns)
| Operation / Operator | Mathematical Concept | Types Supported |
| :--- | :--- | :--- |
| `++` / `concat` | Monoid operation | `List`, `Chunk`, `Set` (union), `HAMT` (merge), `string`, `iter.Seq` |
| `map` / `bimap` / `mapValues` | Covariant Functor / Bifunctor | `List`, `Chunk`, `Set`, `HAMT`, `Tree`, `Pair`, `Either`, `Maybe`, streams |
| `fold` / `foldLeft` / `foldRight` | Catamorphism / Foldable | `List`, `Chunk`, `Set`, `HAMT`, `Tree`, `Pair`, `Either`, `Maybe`, streams |
| `filter` / `reject` | Predicate filtering | `List`, `Chunk`, `Set`, `HAMT`, streams |
| `==`, `!=` | Structural Equivalence | All collections (order-independent for `Set`/`HAMT`), ADTs, primitives |
| `<`, `<=`, `>`, `>=` | Total / Partial Order | Primitives, `List`, `Chunk`, `Pair`, streams (lexicographic) |
| `union`, `intersect`, `diff` | Set-Theoretic Algebra | `Set[T]`, `HAMT[K, V]` (keyed set), sequences |
| `flatMap` / `bind` | Monad bind | `List`, `Maybe`, `Either`, streams |

**Completeness Rule**: When introducing a new operator or type, you MUST implement all cells in the matrix where the operation is mathematically well-defined. Never implement `++` or `fold` for `List` without ensuring `Chunk`, `Set`, and `HAMT` also support their algebraic equivalents.

### 3. Clean Core Prevention Rule (Generic Core vs Domain Isolation)
- **Zero Domain Pollution in Core**: Never introduce use-case specific or domain-specific types, keywords, or pipeline steps (e.g. `StepConway`, `LifeGrid`) into core packages (`evaluator/`, `functions/`, `collections/`, `pkg/cli/`). Core must remain strictly generic, mathematical, and category-theoretic.
- **Go Standard Library Upward Wiring**: If a specific use case reveals that a capability from the Go Standard Library is missing, implement it **generically** across all layers (Layers 0 through 5) so it becomes a general-purpose Gatwd builtin available to all programs.
- **Third-Party Domain Dependencies**: Third-party Go dependencies (Redis, Postgres, Kafka, Kubernetes, AWS, etc.) must NEVER be imported into core. Implement them as **Adapters / Plugins** within the repository (e.g. in `adapters/` or `pkg/plugins/`), registrable in `pkg/cli` to build **Gatwd Editions** (`cmd/gatwd-<name>`) and **Gatwd Appliances**.

---

## When to Activate

- After implementing any new abstraction (collection type, transducer, operator, adapter).
- When asked to build an application, algorithm, or simulation in the DSL.
- During architectural review sweeps or before minor/major releases.
- When requested by the user to dogfood new features or eliminate gaps.

---

## Procedure

### Step 1: Build the Chronological Dependency Map
Use git history to understand what was built and when:

```bash
# List all Go files by first commit date (oldest first)
git log --diff-filter=A --name-only --pretty=format:"%H %ai" -- '*.go' | \
  awk '/^[0-9a-f]/ {date=$2; next} NF {print date, $0}' | sort | uniq

# Get a summary of what recent commits introduced
git log -n 15 --oneline --reverse -- '*.go'
```

### Step 2: Audit the Operator-Type Completeness Matrix
Check every row (type) and column (operation):
1. Does the Go type in `collections/` or `functions/` provide the method?
2. Does `functions/operators.go` overload the operator for this type?
3. Does `evaluator/pipeline.go` provide a stream transducer for this operation on this type?
4. Does `pkg/cli/cli.go` parse and execute the operation in expressions and pipeline steps?
5. Does `tests/dsl_exhaustive_test.gatwd` assert its correct execution?

If any cell in the matrix is missing, implement it immediately across all layers before proceeding.

### Step 3: Scan Older Code for Dogfooding Opportunities
For each file in the codebase, starting from the **oldest**, ask:
1. **Does this file use raw Go primitives where a newer collection type is more appropriate?**
   - Raw `map[string]any` -> immutable `collections.HAMT`
   - Raw `[]T` slices needing bounds safety -> `collections.Chunk[T]`
   - Ad-hoc tuples -> `functions.Pair[A, B]`
   - Untyped error/value returns -> `functions.Either[error, T]`
2. **Does this file manually implement loops that newer combinators provide?**
   - Manual iteration loops -> `iter.Seq`, `FoldLeft`, `Map`, `Filter`
   - Manual error checking -> `Either` monadic binds or `StepAssert`
3. **Does this file duplicate pattern matching or parsing?**
   - Ad-hoc `strings.Split`/`strings.HasPrefix` -> monadic Parser Combinators in `parser/`
4. **Can tests be strengthened using newer assertion tools?**
   - String assertions -> `functions.StructuralEqual` and typed structural checks

### Step 4: Verify Architectural Layering & Dependency Safety
Ensure dependency direction is strictly downward:

```
Layer 0: functions/ (ADTs, operators, streams)
Layer 1: collections/ (HAMT, Set, List, Chunk, Tree)
Layer 2: arrows/, comonads/, optics/, transducers/, transformers/
Layer 3: evaluator/ (Engine, Pipeline, Steps)
Layer 4: adapters/ (IO, Scanner, FS, Net)
Layer 5: pkg/cli/ (Parser, Builtins, Interpreter, CLI)
```

Run `go vet ./...` to guarantee no circular package imports.

### Step 5: Test-Driven Verification
1. Add Go unit tests in `<package>_test.go` asserting generic and memory-level properties.
2. Add DSL-level tests in `tests/dsl_exhaustive_test.gatwd` verifying script execution.
3. Run the full verification suite:
   ```bash
   go test ./...
   ./run_dsl_tests.sh
   ```

### Step 6: Synchronize Documentation
1. Update `docs/dsl_guide.md` with new operator semantics and keywords.
2. Update `docs/components.md` with new type methods and algebraic properties.

---

## Anti-Patterns to Avoid

- **The Bare-Minimum Mockup**: Writing Go helper code or hardcoded values to bypass missing DSL capabilities.
- **Piecemeal Operator Support**: Implementing an operator for only one type while ignoring the rest of the type matrix.
- **Stranded Lower-Layer Abstractions**: Implementing a Go feature in `collections/` without wiring it to `evaluator/` and `pkg/cli/`.
- **Cyclic Imports**: Importing a higher layer into a lower layer just to dogfood.
- **Forced Dogfooding**: Replacing clear, simple code with over-abstracted combinators where no clarity or correctness benefit exists.
