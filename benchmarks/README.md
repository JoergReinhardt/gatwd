# gatwd benchmarks — Ackermann

A performance, memory, and deep-recursion probe for the gatwd DSL interpreter,
built around the Ackermann function `A(3, n)`. Reuses the tour definition
(`examples/tour/23_ackermann.gatwd`) as a deliberately allocation-heavy,
call-bound workload.

> **Why Ackermann `A(3, n)`.** It stays on the native `int64` fast path
> (`A(3, n) = 2^(n+3) - 3`), so it measures *interpreter overhead*, not `big.Int`
> arithmetic. It is also the worst case for any interpreter — pure function-call
> overhead with almost no work per call — which makes the per-call cost visible.
> Input `n` is a single calibration knob: wall-clock grows ~4×/step, recursion
> depth grows ~2×/step.

## Files

| File | What it is |
|---|---|
| `ackermann.gatwd` | Self-asserting DSL workload, ramped to `A(3,6)` (~5s). |
| `run_ackermann.sh` | Builds + runs the DSL script under `/usr/bin/time -v`; reports wall + peak RSS. |
| `native/ackermann.go` | Native-Go baseline. `go run ./benchmarks/native <n>` (default 6). |
| `native/crashprobe/main.go` | Stack-overflow probe. `go run ./benchmarks/native/crashprobe <n> <maxStackBytes>`. |

## Running

```sh
./benchmarks/run_ackermann.sh          # DSL: wall-clock + peak RSS
go run ./benchmarks/native 7           # native Go baseline for A(3,7)
go run ./benchmarks/native/crashprobe 11 1048576   # force a native stack overflow
```

---

## 1. Throughput & scaling

Wall-clock for a single `A(3, n)` call (measured on the dev machine; native Go
times are small and carry startup/GC noise):

| n | result | gatwd DSL | native Go |
|---|---|---|---|
| 3 | 61    | 0.04 s  | — |
| 4 | 125   | 0.29 s  | — |
| 5 | 253   | 1.18 s  | — |
| 6 | 509   | 4.49 s  | 0.82 ms |
| 7 | 1021  | 20.6 s  | 2.39 ms |
| 8 | 2045  | ~80 s *(est.)* | 4.17 ms |
| 10| 8189  | —       | 141.6 ms |
| 12| 32765 | —       | 0.975 s |

Both runtimes scale **~4×/step** (call count ~`4^n`). Recursion *depth* grows only
~`2^n`, so **cost ≈ depth²** — the key structural fact below.

**Interpretation tax:** gatwd is **~6,000–9,000× slower than native Go**,
roughly constant across depth (it is per-call overhead, not depth-dependent). At
`A(3,7)`: 20.6 s vs 2.39 ms ≈ **~8,600×**.

## 2. Memory / leak behavior

Peak RSS (`/usr/bin/time -v`), gatwd:

| input | recursion depth | wall | peak RSS |
|---|---|---|---|
| A(3,6) | ~509  | 5.3 s  | 18.8 MB |
| A(3,7) | ~1021 | 20.6 s | 21.7 MB |

Depth doubled and total work 4×'d, yet peak RSS rose only ~15%. **The workload is
high-churn, not leaking** — the interpreter allocates furiously per call (reparse
+ engine alloc + scope fork) but GC reclaims it; the live set stays bounded.
`/usr/bin/time -v` gives a clean perf + peak-memory number but will **not** surface
a true leak on its own — that needs an in-process repeat-identical-work loop
asserting `runtime.HeapInuse` plateaus (`inuse_space` flat vs `alloc_space` huge).
*(Not yet built — see TODO.)*

## 3. Deep recursion: gatwd vs native Go

The payoff. gatwd uses an `iter.Seq` trampoline, so recursion depth does **not**
consume the Go call stack — depth is bounded by time/memory, with no stack wall.
Native Go recurses on the goroutine stack and has a hard ceiling.

Measured with `crashprobe`: each native Ackermann frame is **exactly 128 bytes**,
and `peakDepth(A(3,n)) = 2^(n+3) - 1 = A(3,n) + 2`.

| Go stack limit | Overflow depth | Smallest crashing input |
|---|---|---|
| 1 MiB (forced) | 8,192 frames | **A(3,11)** — depth 16,381 (A(3,10) survives at 8,191) |
| 1 GiB (Go default) | ~8.39M frames | **A(3,21)** — depth 16.78M (A(3,20) survives at 8.39M) *(extrapolated)* |

**But the default-stack crash is unreachable in practice.** Since cost ≈ depth²,
reaching `A(3,21)`'s depth takes ~`4^9 × 0.975 s ≈ ~3 days` of native CPU. Native
Go effectively never stack-crashes on Ackermann — it runs out of *time* first. The
only way to observe the crash is to cap the stack (what `crashprobe` does).

**Contrast at the 1-MiB crash point (A(3,11), depth 16,381):** native Go dies in
milliseconds; gatwd finishes in **~1.5–2 hours** *(estimated: `20.6 s × 4^4`)* at
~20–25 MB flat RSS, no stack limit in sight. The trampoline trades a hard crash
for a soft ~6,000× time penalty — that is the design tradeoff, quantified.

## 4. Cross-language penalty

Same input `A(3,7) = 1021`, same machine. "Penalty" = slowdown on this call-bound
microbenchmark (the worst case for interpreters; real workloads compress these):

| Runtime | Time | vs native Go | vs CPython |
|---|---|---|---|
| Native Go | 2.39 ms | 1× | 0.047× |
| CPython 3.12.3 | 51.2 ms | ~21× | 1× |
| **gatwd** | **20.6 s** | **~8,600×** | **~400×** |

For reference, typical slowdown-vs-native on call-bound recursion: JIT runtimes
(LuaJIT/V8/PyPy/JVM) 1–5×; Lua ~10–30×; CPython/Ruby ~50–100×; naive AST
tree-walker ~500–2,000×; POSIX shell ~1,000–10,000×. gatwd's ~400×-behind-CPython
places it below the naive-tree-walker tier, around shell-script territory.

## 5. Why gatwd is ~400× behind CPython — and the fix

The gap is **not** evaluator quality (CPython and gatwd do similar per-call
interpretive work). It is almost entirely **per-call re-parsing**. On every single
recursive call gatwd:

1. **re-lexes/re-parses the `def` body string** (CPython compiles each function
   once, to bytecode, forever);
2. allocates a fresh `evaluator.NewEngine(...)`;
3. `fork()`s a full map-copy of scope for the `let ... in ...` arm.

CPython pays ~one dispatch + stack push per call; gatwd pays a *compile* per call.

The obvious-looking fix — cache each `def`'s parsed `[]PipelineStep` by name —
does **not** work. See §6 for why (we tried it and measured it).

## 6. Optimization attempt: name-keyed step cache (negative result)

Hypothesis: the ~400×-vs-CPython gap is per-call re-parsing, so caching each
`def`'s parsed `[]PipelineStep` by name should collapse it by ~an order of
magnitude. We prototyped it on branch `perf/cache-pipeline-steps`:

- **Tier 0** — reuse `i.Engine` instead of `evaluator.NewEngine()` per call.
- **Tier 1** — cache `funcName → []PipelineStep`, served only when the active
  scope has empty `Variables`/`RuntimeNodes` (so no value-dependent substitution
  can occur — making the cached parse provably identical to a fresh one),
  invalidated when any `def`/`const` is (re)defined.

**Measured result: A(3,7) 21.09s → 20.83s (~1%, noise).** The cache never
engages on recursion:

```
cache: hit=0  miss=1  skip(scoped)=42437   (A(3,5), 42,438 calls)
```

**Why it can't work.** Every recursive `ackermann` call fires under a *forked*
scope that carries the match-bound `m`, `n`, `res` in `Variables`/`RuntimeNodes`,
so the guard correctly refuses to cache (42,437 of 42,438 skipped). The re-parse
is not redundant recompilation of invariant structure — it is **how arguments are
bound**: the evaluator substitutes `m`/`n`/`res` values into the body string and
re-parses the match cases *per call*. Parsing and argument-binding are the same
operation; a name-keyed cache cannot separate them without changing behavior.

**CPU profile (A(3,5), `go tool pprof`)** confirms the cost is real but
inseparable from binding:

| Cost center | cum % |
|---|---|
| `parsePipelineSteps` (re-parse body every call) | ~66% |
| ↳ `ParseMatchCase` + parser-combinator `Bind` | ~25% |
| ↳ `splitPipelineAware` / `splitCasesAware` | ~12% |
| ↳ `EvalInfixString` / `TokenizeInfix` | ~6% |
| `mallocgc` + GC (driven by the above) | ~50% |

**The real lever is Tier 2, not caching:** parse each `def`'s match structure
*once* into a reusable template, and bind arguments as **data resolved at eval
time** rather than by substitute-and-re-parse. That attacks the ~66% parse cost
*and* the allocation churn it drives. It is a ~1–3 day architectural change (high
regression risk — touches the string-substitution model `let`, constants, and
lambda-inlining all share), now justified by a measured number instead of a guess.

Regression status of the Tier 0+1 branch: full Go suite (18 pkgs) green; 29/32
batch scripts byte-identical vs master (the 3 others are nondeterministic Go
map-iteration order on master alone). Behavior-preserving, but only a ~1% win —
kept mainly as a documented stepping-stone and for the argument-free top-level
call paths where the cache *does* engage.

## TODO

- In-process leak harness: loop a fixed `A(3,n)` N times, assert `HeapInuse`
  plateaus; two variants (fresh `NewInterpreter()` vs reused interpreter) to catch
  registry leaks vs `Funcs`/`Variables`/`RuntimeNodes` map accumulation.
- Optional `make bench` / CI wiring (currently manual-run).

---

*All single-number timings are from individual runs on one dev machine and carry
normal variance; treat them as order-of-magnitude. Rows marked estimated/extrapolated
were computed, not measured.*
