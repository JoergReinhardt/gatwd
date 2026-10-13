# Plan: Layers 0–2 as a Reusable Functional Go Library

> **Status: proposal, deliberately loose.** The repo is under heavy development; file names, package
> names and ordering below are directional and will change at implementation time. Treat each step as
> "the intent", not a spec. Nothing here is started.

## 1. Idea

Layers 0–2 (`functions/`, `collections/`, `arrows/`, `comonads/`, `optics/`, `transducers/`,
`transformers/`) already depend only on each other (plus `dsltypes/` in two places) and use real Go
generics with no third-party dependencies. They could serve as a standalone functional toolkit that
lets *procedural* Go code borrow functional tools: persistent collections, `Either`/`Maybe`, state and
reader monads, lenses, lazy `iter.Seq` pipelines.

The goal is to make that an enforced, documented property of the repo rather than an accident, without
hurting the interpreter that is the project's main product.

## 2. Observed gaps (as of 2026-10)

1. **`any` leaks into the typed API.** Folds are `Fold(initial any, f func(acc any, item T) any) any`;
   `Elements()` returns `[]any`. Callers must type-assert. Go methods cannot introduce type
   parameters, so a generic `Fold[B]` cannot be a method.
2. **`Map` is `func(T) T`.** A method cannot change the element type (`List[int]` → `List[string]`).
3. **DSL concerns sit inside the "library" layers.** `optics/` and `transformers/` import `dsltypes/`;
   `functions/` holds `dslfunc.go`, `graph_node.go` and the DSL-oriented `closure.go`.
4. **Per-type matrix cost.** Mandate 2 asks for every operation on every collection (N types × M
   operations), implemented per type.
5. **`collections/` imports `functions/`** (a couple of files; to be verified which types), so the
   persistent collections are not a zero-dependency unit.
6. **Go interop is thin or unverified.** `iter.Seq` entry/exit points, `slices`/`maps` conversions,
   `(T, error)` helpers.
7. **Docs and examples target the DSL**, not Go developers.

## 3. Principles

- **The DSL depends on the core, never the reverse.** Dynamic typing (`any`, BitFlag tokens) lives in a
  facade above the typed core.
- **Rigor is kept where it is cheap, relaxed where it is expensive.** Law tests and matrix parity stay
  as goals; per-type duplication is what we eliminate, not the guarantees.
- **Additive first.** New generic APIs land beside existing methods; removals come only after the
  interpreter has migrated.
- **Guard rails before moves.** Boundary tests go in early so decoupling cannot silently regress.

## 4. Workstreams (order is a suggestion)

### L1. Boundary guard test (small, do first) — DONE
Implemented as `internal/archtest/layers_test.go` (parses imports; ranks `functions` < `collections` < the Layer 2 packages; shrink-only allowlist `knownViolations`, which also fails on stale entries). Current allowlist: `optics`/`transformers` → `dsltypes` (L3) and library tests → `visualizer`/`dsltypes` (L1/L3). Note: `collections` → `functions` is *allowed* by the rank rule; L4 (dependency-free `collections/`) is a stricter goal and is not yet enforced.

Original intent:
A Go test (or script) that runs `go list -deps` for Layers 0–2 and fails on imports of anything above
them, or of `dsltypes/`/`visualizer/` from non-test code. Initially it will list the known violations
as an allowlist that shrinks as later steps land. Also move `visualizer` imports out of library tests.

### L2. `iter.Seq` as the interchange format (highest value)
- Every collection gets `All() iter.Seq[T]` (and keyed variants for `HAMT`/`Graph`) plus `FromSeq`
  constructors.
- `Map`, `Filter`, `Fold`, `FlatMap`, `Zip`, etc. are written once as package-level generic functions
  over `iter.Seq`, with type-changing signatures (`Map[T, U]`, `Fold[T, B]`).
- Existing methods become thin wrappers or are deprecated once the interpreter no longer needs them.
- Effect on Mandate 2: the matrix shrinks from N×M implementations to N (iterators) + M (operations),
  keeping parity across types for free. Cells that need structure-preserving results (e.g. `Set.Map`
  re-deduplicating) stay as methods.

### L3. DSL facade extraction
- Move `dslfunc.go`, `closure.go`, `graph_node.go` (if DSL-bound) and the `dsltypes` use in
  `optics/`/`transformers/` into a facade package above Layer 2 (exact home TBD: `evaluator/` or a new
  dynamic layer).
- The facade wraps the typed core with `any`-typed adapters so the interpreter keeps its uniform shapes.
- Likely the first place the roadmap's Workstream B/E work (first-class functions, function forms
  across the matrix) needs coordinating, since it also touches these files.

### L4. Make `collections/` dependency-free
Replace the few `functions/` imports (verify which: probably `Pair`/`Maybe`) with stdlib types or small
local types, or invert the direction. Target: persistent `List`, `Set`, `HAMT`, `Chunk`, `Tree`, `Graph`
importable with no other gatwd package.

### L5. Split by audience (module boundaries)
Candidate grouping, to be revisited once L2–L4 settle:

| Group | Contents |
|---|---|
| `fp` | `Either`, `Maybe`, `Pair`, `State`/`Reader`/`Writer`/`Cont`, currying, parallel helpers |
| `persistent` | `List`, `Chunk`, `Set`, `HAMT`, `Map`, `Tree`, `BST`/`RBTree` |
| `graph` | `Graph[K,C,N,E]` and traversals |
| optional | `optics`, `comonads`, `arrows`, `transducers`, `transformers` |

Mechanics (separate Go modules vs. `internal/` boundaries, `go.work`, versioning) are open; a module
split is only worth doing if there is an actual external consumer.

### L6. Consumer-side interfaces
Prefer tiny interfaces defined where used (`Foldable[T]` = "has `All()`") over cross-package concrete
types, to keep L4/L5 boundaries from re-forming.

### L7. Go-facing docs and examples
A package overview and `Example*` functions written for procedural Go: slice → pipeline → slice,
`Either` error chains, `Set`/`HAMT` as drop-in persistent maps, snapshots/undo with structural sharing.
Keep the educational category-theory docs; add a "no theory required" quick start.

### L8. Law tests (optional, rigor-preserving)
Property-style tests for functor/monad/monoid laws on the core types, so relaxing per-type matrix
parity does not mean relaxing correctness.

## 5. Risks and open questions

- **Interpreter churn.** `pkg/cli/` and `evaluator/` rely on the `any` folds and `func(T) T` maps; L2/L3
  must migrate callers or keep wrappers during the transition.
- **Parallel work in flight.** Uncommitted language-roadmap work (closures, matrix function forms,
  `typecheck.go`) touches the same files as L3; sequence after it lands or coordinate.
- **Performance.** Generic `iter.Seq` pipelines can be slower than hand-rolled loops for small
  collections; benchmark before deprecating existing methods.
- **Is there a consumer?** If nobody outside this repo uses the library, L5 (module split) is
  speculative; L1–L4 are still worthwhile as hygiene.
- **Which `any` uses are essential?** Some (graph payloads, DSL values) are intrinsic; only the
  accidental ones should be removed.

## 6. Related

Workstream G in [`language_roadmap.md`](language_roadmap.md) (program-level IR, optional closure compiler,
optional typed-dialect Go codegen) is the consumer that benefits most from L2/L3: a code generator would
target the typed core directly while the interpreter keeps the `any` facade.

## 7. Done-when (rough)

- Boundary test passes with an empty allowlist.
- No exported Layer 0–2 API requires `any` where a type parameter would do.
- `collections/` builds with no other gatwd package imported.
- A Go developer can use `persistent` + `fp` from a README without reading DSL docs.
- `go test ./...`, `go vet ./...` and `./run_dsl_tests.sh` stay green throughout.
