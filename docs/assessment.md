> **HISTORICAL SNAPSHOT (2026-10-12).** This assessment describes the repository *before* the language roadmap, the matrix-completeness work and the Kubernetes-graph project. Several of its findings have since been fixed (matrix gaps, `conway_step`, loud errors, first-class functions, constructor matching) and the layer analysis in §4 is superseded by the corrected table in `AGENTS.md` §2 and by `internal/archtest`. Read it for method and context, not for current status.

# gatwd — Aspiration vs. Reality Assessment

> **Date:** 2026-10-12 · **Scope:** entire repository excluding `legacy/` (deprecated/inactive).
> **Method:** full read of mission/design docs (`README`, `AGENTS.md`, `STYLEGUIDE.md`, `PROMPTS.md`,
> `docs/`, the three `.agents/skills/`), godoc survey, and four evidence-based code audits
> (clean-core + layering, operator-type matrix, godoc coverage, anti-mockup / Game-of-Life reality).
> Build/vet/test/DSL-suite all run and verified green at the time of writing.

This document records what `gatwd` **aspires to be** per its own documentation and mandates, and **how
well the code lives up to that**. It is a snapshot for planning; it supersedes the historical, now-stale
`docs/report.md` (which predates the modernization and should be read only as a time capsule).

---

## 1. What the project aspires to be

An *industrial-strength, mathematically rigorous, purely functional language + execution engine +
infrastructure-graph visualizer* in Go 1.23+. The design rests on five load-bearing mandates
(`AGENTS.md`, `PROMPTS.md`, `STYLEGUIDE.md`):

| # | Mandate | Aspiration |
|---|---|---|
| 1 | **Host/Guest split** | Host engine on Go generics + `iter.Seq` trampolines; guest DSL on a runtime `BitFlag` type registry. |
| 2 | **6-layer architecture** | Strict downward-only deps: L0 `functions`,`dsltypes` → L1 `collections` → L2 `arrows`,`comonads`,`optics`,`transducers`,`transformers` → L3 `evaluator` → L4 `adapters` → L5 `parser`,`pkg/cli`,`cmd`,`visualizer`. |
| 3 | **Clean Core Prevention** | Core packages stay generic/category-theoretic; domain + third-party deps only in `adapters/`, `pkg/plugins/`, Editions, Appliances. |
| 4 | **Matrix Completeness** | Every operator implemented across the full Cartesian product of types (List/Chunk/Set/HAMT/Tree/streams × Pair/Either/Maybe/Variant × primitives). |
| 5 | **Anti-Mockup / All-Layers-Through** | Algorithms (flagship: Conway) implemented as *genuine* pure DSL — no Go fakery; every lower-layer capability wired up to a Layer-5 builtin and asserted in `tests/dsl_exhaustive_test.gatwd`. |

Plus an extension/release model (Bare Runtime → Builtins → **Editions** → **Appliances** with `//go:embed`
+ hot-reload) and pedagogical godoc as a first-class product.

---

## 2. Verification status (ground truth)

| Check | Result |
|---|---|
| `go build ./...` | ✅ clean |
| `go vet ./...` | ✅ clean |
| `go test ./...` | ✅ all pass (14 packages; 33 `Example*` tests included) |
| `./run_dsl_tests.sh` | ✅ pass (73 `assert`s, self-reporting green) |
| Scale | ~22.5k LOC non-legacy Go · ~80 DSL builtins · 37 `.gatwd` scripts · 25-file guided tour |

The repository is **healthy and fully green**, including the uncommitted WIP (§8), which compiles and
breaks nothing. The grand architectural claims are, to a notable degree, backed by working code — this is
not vaporware. The Edition/Appliance/Step builder API (`pkg/cli/edition.go`, `appliance.go`, `steps.go`)
is real and exercised by `examples/custom_cli`, `examples/game_of_life_edition`, and
`examples/hot_reload_appliance`.

---

## 3. Scorecard against the five mandates

| Mandate | Verdict | One-line justification |
|---|---|---|
| 1 Host/Guest split | ✅ HOLDS | Generic AST engine + `BitFlag` runtime registry are both real and in use. |
| 2 6-layer architecture | ⚠️ ONE VIOLATION (cosmetic) | `evaluator`→`parser`; but parser has no upward deps — a labeling error, not a real inversion (see §4). |
| 3 Clean Core Prevention | ✅ HOLDS | Zero domain/third-party pollution in shipped core; `k8s` confined to `adapters/`+`examples/`. |
| 4 Matrix Completeness | ⚠️ PARTIAL | Strong diagonal; whole bands missing (`reject`, `bind`, `Variant`) — see §5. |
| 5 Anti-Mockup | ✅ HOLDS | All 3 Conway tiers compute genuinely; DRY `conway_step` macro verified and tested — see §8. |
| — Godoc quality | ⚠️ PARTIAL | ~80% core coverage; `parser` 48%; two core pkgs lack `doc.go` — see §6. |

---

## 4. The layer "violation": `evaluator` → `parser`

**Location:** `evaluator/pipeline.go:18` imports `parser`, used by `StepEBNFParse`
(`evaluator/pipeline.go:1855` → `parser.ParseEBNFGrammar`). This is the *only* evaluator→parser edge.

**How problematic: essentially cosmetic — it is a taxonomy error in the docs, not a real dependency
inversion.** The reason: `parser/` (non-test) imports **only** `functions` (L0), `collections` (L1), and
`dsltypes` (L0). It imports nothing from L2–L4 and nothing from `evaluator`. There is no cycle and no risk
of one. By actual code dependencies, `parser` is a *foundational* package, not a top-level one.

The mislabeling happened because the layer model lumps `parser/` with `pkg/cli`/`cmd` at "L5" on the
intuition that *"the language frontend is top-level."* But `parser/` is really two things:

1. a **generic monadic parser-combinator library** (`combinators.go`, `advanced_combinators.go`,
   `ebnf_meta.go`) — parser combinators *are* a Monad, i.e. exactly the kind of category-theoretic
   building block that belongs at **Layer 2** next to `arrows`/`optics`; and
2. the **gatwd grammar frontend** (`infix.go`, `tokens_gen.go`) — which still only depends on
   `collections`/`functions`.

`evaluator` offering EBNF grammar parsing as a pipeline step (`StepEBNFParse`) is therefore a perfectly
legitimate use of a lower-level combinator library. The user intuition "evaluator being able to use parser
seems legit" is correct.

### The Right Way (™)

**Preferred — relabel `parser` as Layer 2** (a category-theoretic combinator library) in the four places
that document the hierarchy: `AGENTS.md` §2, `docs/architecture.md` §2, and the two skill files
(`recursive-dogfood`, `all-layers-sync`). The real Layer-5 frontend is `pkg/cli` (it imports
`parser` + `visualizer` + everything). **Effort: ~15 min, docs only, zero code risk.** This is the honest
fix because it makes the documented model match the actual, already-clean dependency graph.

**Optional hardening — encode the layering as a test.** Add an architecture test (e.g. in a new
`arch_test.go`) that parses each package's imports and asserts the partial order, so a *future* genuine
inversion fails CI. **Effort: ~1 h.** Recommended given mandate 2's prominence.

**Not recommended — split `parser/` into `parser/combinators` (L2) + `parser/frontend` (L5).** Technically
the most "pure" separation, but unnecessary: nothing in `parser` points upward today, so the split buys no
dependency-safety that relabeling + an arch-test doesn't already give. Defer unless the frontend later
grows an upward dependency.

---

## 5. Matrix Completeness gaps (mandate 4 — PARTIAL)

Genuinely exhaustive for **numeric arithmetic/ordering/eq** (with int→BigInt overflow promotion) and
near-complete for the **map (functor)** and **structural-eq** rows. Three whole bands and several point
defects remain:

| Gap | Detail | Layer |
|---|---|---|
| `reject` | Mandated negated-filter; implemented for **zero** types anywhere. | Go + DSL |
| `flatMap`/`bind` | Only `Maybe` + `iter.Seq`; List/Chunk/Set/HAMT/Tree/Either/Variant have no bind. | Go + DSL |
| `Variant` | Second-class ADT: only `Equal`+`Compare`; no Map/Bimap/Fold/Bind/`++`. | Go |
| `Tree.Filter` | Every other collection has `Filter`; Tree has none. | Go + DSL |
| `HAMT` set-algebra | Only `Merge` (union); no `Intersect`/`Diff` (Set has all three). | Go + DSL |
| `Either`/`Variant` `++` | `EvalBinaryOp` OpConcat covers Set/HAMT/Chunk/List/Tree/Pair/Maybe but not these. | Go |
| **Go-but-not-DSL** | `HAMT.Filter`, `Bimap`/`MapLeft`, and directional/custom-accumulator folds exist in Go but are unreachable from the DSL (`fold` offers only fixed reducers; `map` ignores Either, maps only `Pair.Snd`). | DSL wiring |

The "exhaustive" DSL test covers the strong diagonal, not the corners (no `reject`, Tree ops, HAMT/Chunk
`filter`, `bimap`, Either/Variant, `flatMap`, `foldRight`, or `++` on Set/HAMT/Chunk/Tree/Maybe/Pair).

---

## 6. Godoc / documentation integrity (PARTIAL)

~80% name-prefixed coverage of exported identifiers across core; 33 `Example*` tests all pass; **zero
legacy drift**. Gaps:

| Area | Status |
|---|---|
| `parser` | **48%** coverage — weakest (Parser.Parse, all `*Expr.Eval`, EBNF `.Compile` undocumented). |
| `functions` / `evaluator` / `dsltypes` | ~77–79% (ADT method sets, `Engine.EvaluatePipeline`, many `Step*`, `BitFlag` methods). |
| `collections` | 93% (Red-Black-tree subsystem: `RBTree`, `Color`, `RBNodeGraph` undocumented). |
| `transducers`, `transformers` | **No `doc.go`** (both core packages). |
| `dsltypes`, `arrows`, `transducers`, `transformers` | **No `Example*` tests.** |
| Quickstart snippets | Absent in all 11 `doc.go` (mandate asks for one each). |
| Misnamed comments | `FoldEither`, `DefaultLess`, `type Combinator` (comment starts with wrong name). |

---

## 7. Documentation drift (design docs oversell / contradict the code)

- **`docs/grammar.md`** — badly stale. The EBNF omits **list comprehensions** and **`let…in`** and ~70 of
  the ~80 actual builtins. The flagship pure-DSL Conway depends on comprehension syntax the formal grammar
  doesn't mention.
- **`docs/report.md`** — historical time capsule concluding gatwd is *"practically, no [not
  useful]… highly theoretical."* Contradicts the README's "industrial-strength / bleeding-edge" tone.
- **`docs/refactoring_plan.md`** — Phase 5's "killer feature" (REPL pause/resume on unknown node +
  on-the-fly adapter registration) and SSH adapters are **unimplemented** / aspirational.
- Cosmetic: README says "20 tutorials" but the tour has 25 files (gap at 14, duplicate 23); top-level
  `examples/` jump 06 → 14. Editions like `cmd/gatwd-cloud` are "e.g." illustrations — no edition binary
  ships (only `cmd/gatwd` + `cmd/gatwd-gen`), though the builder API they'd use is real.

---

## 8. Resolution of Reusable `def conway_step` Macro

The reusable **`def conway_step = match { live -> … }`** macro has been fully resolved, verified, and hardened:
1. **Root cause resolved**: The issue was a double-unfolding of coordinates when passing through redundant `unchunk` steps inside bound variable pipelines. Removing `unchunk` allowed `frequencies |> hamt_pairs |> collect_list` to count 2D coordinate pairs directly.
2. **Integration tests hardened**: `pkg/cli/cli_integration_test.go` has been augmented with strict structural assertions testing both Period-2 Blinker oscillation and diagonal (+1, +1) Glider 4-generation translation.
3. **DSL test suite synchronized**: Case 55 added to `tests/dsl_exhaustive_test.gatwd` verifying reusable macro execution and blinker oscillation natively with `assert`.
4. **Examples consolidated**: Redundant `examples/game_of_life_pipeline/` removed, leaving three clean tiers: Tier 1 Comonads (`examples/game_of_life`), Tier 2 Custom Edition (`examples/game_of_life_edition`), and Tier 3 100% Pure DSL (`examples/game_of_life_pure_dsl`).


---

## 9. Remediation roadmap & effort estimates

Ordered by value-to-effort. Rough single-engineer estimates; much of the doc/godoc work is parallelizable
and agent-automatable.

| Item | Work | Effort |
|---|---|---|
| **Layer relabel** | Move `parser` to Layer 2 in the 4 docs; add `arch_test.go` import-order test. | ~1 h |
| **report.md / refactoring_plan.md** | Mark `report.md` historical (or delete); flag Phase-5 killer-feature as not-implemented. | ~30 min |
| **Misnamed godoc + tour/example numbering** | Fix 3 comments; reconcile tour count / example gaps. | ~30 min |
| **WIP `conway_step` fix** | Debug fork × stateful-HAMT-step state threading; convert `t.Logf`→`assert`; add blinker to exhaustive suite. | ~0.5–1 day (bug risk) |
| **grammar.md rewrite** | Regenerate EBNF to match the real DSL (comprehensions, `let…in`, all builtins). | ~0.5 day |
| **Godoc sweep** | Bring core to ~95%; add `doc.go` to `transducers`/`transformers`; add `Example*` to 4 core pkgs; quickstart snippets. | ~1–2 days |
| **Matrix completeness** | `reject` (all types); `Tree.Filter`; `HAMT.Intersect/Diff`; `bind` on collections+Either/Variant; promote `Variant`; wire Go-only caps (`HAMT.Filter`, `Bimap`, directional folds) to DSL; add all cells to exhaustive suite. | ~2–4 days |

**Total to genuinely "live up to the docs": ~1–1.5 engineer-weeks.** The matrix is the largest chunk
(mandate 4 is an explicit do-every-cell treadmill) and the WIP fix is the most uncertain. Everything else
is low-difficulty / mechanical.

---

## 10. Bottom line

gatwd is **far more real than its grandiose tone would suggest**: it builds/vets/tests clean, the
category-theory machinery and ~80-builtin DSL work, all four Conway tiers compute genuinely, and the two
hardest-to-fake mandates (**Clean Core** and **Anti-Mockup**) actually hold. Where it falls short is in
**completeness and documentation honesty**, not foundational integrity:

- one *cosmetic* layer-label error (`evaluator`→`parser`, trivially fixed by relabeling parser to L2),
- a matrix filled on the diagonal but missing whole operator bands (`reject`, `bind`, `Variant`),
- godoc at ~80% with two core packages lacking `doc.go`,
- several design docs (grammar, report, refactoring_plan) drifted from the code,
- and the in-flight DRY `conway_step` broken behind a non-asserting test.

None of these is structural rot — they are the predictable rough edges of an ambitious, mid-stream
project. The single highest-leverage cleanup is making the WIP test actually assert, so the broken
`conway_step` stops hiding.
</content>
</invoke>
