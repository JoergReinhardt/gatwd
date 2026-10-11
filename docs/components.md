# Core Components Reference

This document breaks down the major packages and abstractions that form the active `gatwd` architecture.

---

## 1. Layer 0: `functions/` (Foundational Functional Primitives)
The foundational mathematical primitives:
- **`Either[L, R]`**: Standard sum-type container for functional error handling. Left represents errors/failures, Right represents valid values. Provides `Left()`, `Right()`, `IsLeft()`, `IsRight()`, `UnwrapLeft()`, and `UnwrapRight()`.
- **`Pair[A, B]`**: Immutable 2-element product type (2-tuple) with `Fst` and `Snd`.
- **`Option[T]`**: Functional optional container (`Some` / `None`).
- **Sequences & Iterators**: High-order sequence utilities (`Take`, `Drop`, `Filter`, `Map`, `Fold`) built on Go 1.23+ `iter.Seq`.

---

## 2. Layer 1: `collections/` (Persistent Data Structures)
Purely functional, persistent data structures with structural sharing:
- **`HAMT[K, V]`**: 32-way branching Hash Array Mapped Trie using bitmap index compression (`math/bits.OnesCount32`) and standard `hash/maphash.Comparable`. Provides $O(\log_{32} N) \approx O(1)$ persistent key-value mapping. Implements `json.Marshaler`, `driver.Valuer`, and `sql.Scanner`.
- **`List[T]`**: Persistent singly-linked Cons list (`Nil()`, `Cons()`, `Head()`, `Tail()`, `Seq()`). Implements `json.Marshaler`, `driver.Valuer`, and `sql.Scanner`.
- **`Tree[T]`**: Multi-way Rose Tree (`NewTree()`, `AddChild()`, `Children()`).
- **`Set[T]`**: Persistent hash set backed by `HAMT[T, struct{}]`. Provides $O(\log_{32} N)$ structural sharing, `Contains()`, `Add()`, and `Remove()`.
- **`Map[K, V]`**: Persistent dictionary backed by `HAMT[K, V]`.
- **`Chunk[T]`**: Eager, contiguous, bounds-checked slice wrapper for CPU spatial locality.

---

## 3. Layer 2: Category Theory Combinators
- **`arrows/`**: Kleisli arrows for monadic composition. Provides identity arrows, FanOut (`&&&`), Split (`***`), and curry composition.
- **`comonads/`**: Spatial navigation via Rose Tree Zippers (`IntoZipper`, `Left`, `Right`, `Down`, `Up`, `Extract`).

---

## 4. Layer 3: `evaluator/` (Streaming Execution Engine)
- **`Engine[T]`**: Trampolined stream processing engine that evaluates pipelines over `iter.Seq`.
- **`PipelineStep[T]`**: First-class stream transducers (`func(iter.Seq) iter.Seq`). Includes `StepTake`, `StepDrop`, `StepFilter`, `StepMap`, `StepSort`, `StepToSet`, `StepToHAMT`, `StepHAMTGet`, `StepChunk`, `StepAssert`, `StepDebugPrint`.
- **`StepCheckpoint` & `CheckpointStore`**: Non-intrusive stream watermarking and continuation state tracking without steady-state performance overhead.
- **`BenchmarkPipeline` & `StepTelemetry`**: Performance measurement harness integrating with `testing.B` and live stream telemetry.
- **`AssertVisualDiff` & `VerifyExampleOutput`**: Visual assertion diffing and testable example verification.

---

## 5. Layer 4: `adapters/` (IO Bridges & Anamorphisms)
- **`ScannerAdapter` & `ScannerToSeq`**: Anamorphic unfold from any `bufio.Scanner` or generalized `Scanner` into lazy ASTNode streams.
- **`ListenerToSeq`, `ConnToLines`, `StepDial`**: Streaming network socket ingress and egress for TCP and Unix domain sockets.
- **`WalkFSAdapter`**: Traverses Go `fs.FS` directory trees, converting directory hierarchies into Comonadic Zippers.
- **`WriteFile`**: Streaming sink to files using Go `io.Writer`.

---

## 6. Layer 5: Language & Tooling
- **`dsltypes/`**: `BitFlag` dynamic quasi-type system and `ASTNode[T]` host-guest bridge (implementing `fmt.Stringer`).
- **`parser/`**: Monadic parser combinators and syntax analyzers.
- **`visualizer/`**: UTF-8 and ASCII graph rendering with borders, box alignment, DOT compilation, and first-class error card formatting.
- **`pkg/cli/`**: The `Interpreter` combining the parser, engine, runtime scopes, macro expansions (`def`), local bindings (`let`), script constants (`const`), and host Go constant registrations (`RegisterConstant`).

> Note: Early pre-generics prototypes and non-generic interface wrappers have been cleanly separated into the `legacy/` directory and are not part of the active pipeline.
