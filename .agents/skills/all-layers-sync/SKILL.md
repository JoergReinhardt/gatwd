---
name: all-layers-sync
description: "Audit lower-level Go abstractions (recent first), ensure they are wired all the way through to the DSL as builtins, and enforce exhaustive test coverage in tests/dsl_exhaustive_test.gatwd."
---

# All-Layers-Through Synchronization Skill

## Principle

In Gatwd, lower-layer functional abstractions (Layers 0 through 4) must never remain stranded in Go API space without script-layer accessibility. Whenever support for a Go standard library capability, persistent data structure, or streaming adapter is built in lower layers, it MUST be wired all the way through to the scripting language (Layer 5 `pkg/cli/`) as a **builtin**.

Third-party Go packages must be integrated by creating **Custom Editions** (or **Distributions**) that import `pkg/cli` and register custom types and `StepHandler` transducers.

## When to Activate

- Whenever a new collection type, transducer, or I/O adapter is introduced or refactored in lower layers.
- During periodic architecture sweeps to audit feature completeness across layers.
- When requested by the user to verify layer-to-layer feature completeness or build custom extensions.

## Procedure

### Step 1: Chronological Audit of Lower Layers (Recent First)

Run git log across lower layers to identify recently introduced types, methods, or transducers:

```bash
git log -n 10 --oneline --stat collections/ adapters/ evaluator/
```

Identify any functions or types that lack corresponding DSL pipeline keywords. Common categories:
- **Collections:** Operations on `HAMT`, `Set`, `Chunk`, `List`, `Tree`.
- **I/O & Adapters:** Unfolding from `bufio.Scanner`, `net.Listener`, `net.Conn`, `fs.FS`.
- **Testing & Telemetry:** Benchmarking, mock streams, ordered/unordered assertions.

### Step 2: Implement Missing Evaluator Transducers (if needed)

If the lower layer provides a data structure or adapter but lacks a stream transducer (`evaluator.PipelineStep`), implement it in Layer 3 (`evaluator/` or `adapters/`):
- Transducers must follow signature: `func(...) evaluator.PipelineStep[functions.Either[error, any]]`.
- Properly handle error short-circuiting: pass Left nodes through untouched unless explicitly acting as error handlers.

### Step 3: Wire into DSL as a Builtin

In `pkg/cli/cli.go`:
1. Check `i.EnableBuiltins` guarding standard library builtins.
2. In `parsePipelineSteps`, add keyword detection for the new step.
3. If the step requires arguments, parse them cleanly (handling quotes, whitespace, and numbers).
4. Append the compiled step to `pipeline`.

### Step 4: Add Self-Asserting DSL Test Cases

Add one or more test cases to `tests/dsl_exhaustive_test.gatwd`:
- Use `spawn <Type>: <Init> |> <new_step> |> assert <Expected>` syntax.
- Ensure the test asserts the true output of the operation.

### Step 5: Verify End-to-End

```bash
# Verify both native Go unit tests and DSL execution
go test ./...
./run_dsl_tests.sh
```

Ensure all tests pass with zero errors.

### Step 6: Synchronize Documentation

1. Update `docs/dsl_guide.md` with the new keyword syntax, parameter descriptions, and DSL examples.
2. Update `docs/components.md` to reflect the newly exposed capability.
