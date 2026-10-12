---
name: all-layers-sync
description: "Audit lower-level Go abstractions (recent first), ensure they are wired all the way through to the DSL as builtins, and enforce exhaustive test coverage in tests/dsl_exhaustive_test.gatwd."
---

# All-Layers-Through Synchronization Skill

## Principle

In Gatwd, lower-layer functional abstractions (Layers 0 through 4) must never remain stranded in Go API space without script-layer accessibility. Whenever support for a Go standard library capability, persistent data structure, or streaming adapter is built in lower layers, it MUST be wired all the way through to the scripting language (Layer 5 `pkg/cli/`) as a **Gatwd Builtin**.

Crucially, **"wiring all things up" must be orthogonal and complete across the type-operator matrix**:
1. **No Selective Wiring**: Introducing or expanding an operation (e.g., `++`, `map`, `fold`, `filter`, `<`, `==`) requires wiring it across ALL data types where it is algebraically sound (`List`, `Chunk`, `Set`, `HAMT`, `Tree`, `Pair`, `Either`, `Maybe`, streams).
2. **No Mockups / Shortcuts**: Never implement bare-minimum Go wrappers or mockups to satisfy a test or demo. Provide full, genuine language and engine primitives so the DSL can solve problems natively.
3. **Clean Core Prevention Rule**: Never introduce use-case specific or domain-specific types, keywords, or pipeline steps (e.g. `StepConway`, `LifeGrid`) into core packages (`evaluator/`, `functions/`, `collections/`, `pkg/cli/`). Core must remain strictly generic, mathematical, and category-theoretic.
4. **Go Standard Library Upward Wiring**: If a specific use case reveals that a capability from the Go Standard Library is missing, implement it generically across all layers (Layers 0 through 5) so it becomes a general-purpose Gatwd builtin.
5. **Third-Party Domain Dependencies**: Third-party Go dependencies (Redis, Postgres, Kafka, Kubernetes, AWS, etc.) must NEVER be imported into core. Implement them as **Adapters / Plugins** within the repository (e.g. in `adapters/` or `pkg/plugins/`), registrable in `pkg/cli` to build **Gatwd Editions** (`cmd/gatwd-<name>`) and **Gatwd Appliances**.

---

## When to Activate

- Whenever a new collection type, transducer, or I/O adapter is introduced or refactored in lower layers.
- During periodic architecture sweeps to audit feature completeness across layers.
- When requested by the user to verify layer-to-layer feature completeness, eliminate gaps, or build custom extensions.

---

## Procedure

### Step 1: Chronological Audit of Lower Layers (Recent First)

Run git log across lower layers to identify recently introduced types, methods, or transducers:

```bash
git log -n 15 --oneline --stat collections/ functions/ adapters/ evaluator/
```

Identify any functions, types, or operators that lack corresponding DSL pipeline keywords or expression operators:
- **Collections**: Operations on `HAMT`, `Set`, `Chunk`, `List`, `Tree`.
- **Algebraic Data Types**: `Pair`, `Either`, `Maybe`, `Variant`.
- **Operators**: `++` (concat/merge/union), `map`, `fold`, `filter`, comparison operators `<`, `<=`, `>`, `>=`, `==`, `!=`.
- **I/O & Adapters**: Unfolding from `bufio.Scanner`, `net.Listener`, `net.Conn`, `fs.FS`.
- **Testing & Telemetry**: Benchmarking, mock streams, ordered/unordered assertions.

### Step 2: Implement Missing Evaluator Transducers & Operators

If the lower layer provides a data structure or adapter but lacks stream transducers (`evaluator.PipelineStep`) or expression operators:
- **Transducers**: Follow signature `func(...) evaluator.PipelineStep[functions.Either[error, any]]`.
- **Operator Matrix Completeness**: Ensure the operator handles all eligible types in `functions/operators.go` and `evaluator/pipeline.go` (e.g. `StepFold` supporting sum, product, min, max, concat, all, any over all foldables; `OpConcat` supporting List, Chunk, Set union, HAMT merge, sequences).
- **Error Propagation**: Pass Left nodes through untouched unless explicitly acting as error handlers.

### Step 3: Wire into DSL as Builtins & Grammar Constructs

In `pkg/cli/cli.go`:
1. Check `i.EnableBuiltins` guarding standard library builtins.
2. In `parsePipelineSteps`, add keyword detection for the new step.
3. Support lambda arguments, literal operands, and named options cleanly.
4. If it is an infix operator, ensure `parser/infix.go` and `evaluator` evaluate it correctly across all types.
5. Append the compiled step to `pipeline`.

### Step 4: Add Self-Asserting DSL Test Cases

Add test cases to `tests/dsl_exhaustive_test.gatwd`:
- Cover the entire Cartesian product of types supported by the feature.
- Use `spawn <Type>: <Init> |> <new_step> |> assert <Expected>` syntax.
- Ensure the test asserts the true structural output of the operation, avoiding hardcoded dummies.

### Step 5: Verify End-to-End

```bash
# Verify both native Go unit tests and DSL execution
go test ./...
./run_dsl_tests.sh
```

Ensure all tests pass with zero errors.

### Step 6: Synchronize Documentation

1. Update `docs/dsl_guide.md` with the new keyword syntax, parameter descriptions, and DSL examples.
2. Update `docs/components.md` to reflect the newly exposed capability and supported type matrix.
