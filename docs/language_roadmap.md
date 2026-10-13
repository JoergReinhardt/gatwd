# Language Roadmap — Staying Interpreted: First-Class Functions & Gradual Typing

> **Status:** COMPLETE (2026-10-13) — every workstream is implemented, tested (Go structural tests,
> `Example*`, self-asserting `tests/dsl_exhaustive_test.gatwd` §§56–80) and documented
> ([`docs/grammar.md`](grammar.md)). Companion to [`docs/assessment.md`](assessment.md), which diagnosed
> the two main gaps vs. Haskell/Elm. This doc scoped *how much of those gaps is reachable while keeping
> `gatwd` an interpreted, dynamically/gradually-typed language*.
>
> **Progress:**
> - ✅ **Workstream A — ADT constructor matching.** `Parser.IsConstructor`/`MakeVariant`; a constructor
>   arm runs before the variable-binding arm in `match`. Tests: dsl #56, `TestParser_IsConstructorAndMakeVariant`.
> - ✅ **Workstream B — first-class functions.** `functions.Closure` (`functions/closure.go`) is a real,
>   curried, lexically-scoped function value: `\x y -> body` is an *expression* (`parser.LambdaExpr`, new
>   `\`/`->` tokens) that captures its defining environment; `Apply` curries on under-application and
>   re-applies on over-application. A lambda-bodied `def`, a `let`-bound lambda and a partial application
>   are the same kind of value, usable in expression position (`add 1 2`, `(add 10) 5`, `mk 5 10`), in
>   pipeline position data-last (`x |> add 3` = `add 3 x`), and by `map`/`filter`/`reject`/`fold`
>   (`map (add 10)`, `fold add 0`, `filter is_pos`). Interpreter glue: `pkg/cli/closures.go`;
>   `mapStepOver`/`filterStepOver` were extracted so inline lambdas and closures share one structural
>   implementation. This also fixed the long-standing "multi-arg lambda `def` as `fold` reducer
>   mis-evaluates" gotcha. Tests: `functions/closure_test.go`, `parser/lambda_test.go`,
>   `pkg/cli/closures_test.go`, dsl #78. **Not done (by design):** a `def` whose body is a pipeline stays a
>   re-parsed macro — it is a stream transformer, not a value function.
> - ✅ **Workstream C — loud errors + diagnostics.** Unknown pipeline steps surface `[ERROR] unknown step
>   or command`; `gatwd --check` is a real structural validator over logical lines (`CheckScript`). Now also:
>   a first-part expression whose *call* fails returns the error (`parser.ApplyError`) instead of silently
>   falling through to "(No output)". **Deferred:** wiring `parser/errors.go` `FormatError` into the
>   line-based top-level parser (combinator-level; low ROI).
> - ✅ **Workstream D — gradual typing.** D3 exhaustiveness (non-exhaustive `match` warnings) plus
>   **D1** optional annotations (`def inc : Int -> Int = …`, `ParseTypeAnnotation`, advisory only) and
>   **D2** local inference (`pkg/cli/typecheck.go`): types flow from literals, operators, annotated
>   parameters, let bindings and comprehension generators; reports wrong argument/result types, arity,
>   operator misuse (decided by the *runtime operator matrix* on sample values, so checker and evaluator
>   cannot disagree), unbound names, and bad `spawn` seeds — all as non-blocking warnings, `Any` consistent
>   with everything, zero warnings across every repo script. Tests: `pkg/cli/typecheck_test.go`.
> - ✅ **Workstream E — expression-level function forms across the operator-type matrix.** 74 new curried,
>   data-last, type-preserving function forms in `pkg/cli/dslfuncs_matrix.go` (+ carrier-polymorphic
>   `mapValue`/`filterValue`/`foldValue`/`bindValue` in `pkg/cli/valueops.go`): generic
>   `map filter reject fold flatmap bimap`; sequences; Set; HAMT; Tree; **BST** (new `Contains`/`Size`/`Seq`
>   at Layer 1); Pair; Maybe/Either; Variant (`variant "Circle" 3`); primitives. Pre-existing fixed-arity
>   builtins (`hamt_get`, `graph_in_of`, …) became curried too. Tests: dsl #79 (91 assertions),
>   `pkg/cli/dslfuncs_matrix_test.go`, `collections/bst_test.go`, `Example_matrixFunctionForms`.

## 0. Scope & non-goals

**In scope** (everything here keeps the interpreter model intact — changes are either new *runtime value
kinds* or an optional *pre-evaluation pass*; the `iter.Seq` trampoline evaluation loop is untouched):

- First-class functions (closures as values, application, currying) — a dynamically-typed interpreter
  feature, no type system required.
- A gradual-typing slice: loud runtime type errors, optional annotations, a best-effort pre-eval checker,
  local inference, and exhaustiveness checking.
- Fixing the broken ADT constructor pattern matching (a prerequisite bug).

**Explicit non-goals** (ruled out by project scope):

- Full Hindley–Milner global type inference / unification.
- Static ahead-of-time compilation.
- Whole-program soundness, type classes, inference-driven parametric polymorphism, and Elm's *guarantee*
  of no runtime type errors. These are intrinsically global-type-system properties and are out of reach
  without the two non-goals above — and that is an accepted trade.

The target end-state is roughly **"a well-typed-at-the-edges, dynamically-typed functional scripting
language"** — Clojure/Elixir/TypeScript-lite in spirit, not Haskell.

---

## 1. Workstream A — Fix ADT constructor pattern matching (prerequisite bug)

### The defect
`match` is an ordered scan of 6 pattern arms over each value's string form
(`pkg/cli/cli.go:1045-1160`). **Arm 5** (`:1138`) fires for *any* identifier and binds it as a catch-all
variable:

```go
// 5. Variable identifier pattern: x -> cmd (binds incoming value to variable x)
if isIdent(pattern) && pattern != "true" && pattern != "false" { ... bind pattern = val ... break }
```

So when the arms are constructor names, the **first arm always wins**. The project's own tour shows it:
`spawn OS: Windows |> route_os` prints the `Linux` branch's output. The engine never consults the Variant
machinery that already exists — `dsltypes.Variant.Is(tag)` and `Variant.MatchType(flag)`
(`dsltypes/variant.go:23,28`), nor the declared branches in `dsltypes.SumTypeDefinition.Branches`
(`dsltypes/variant.go`).

### The subtlety (why this is a *designed* fix, not a one-liner)
Arm 5's catch-all binding is **load-bearing** — it is exactly what makes the newly-landed
`def conway_step = match { live -> ... }` work (commit `92c9253`): `live` is a fresh variable bound to the
incoming grid. If we naively make bare identifiers match constructors, we break `conway_step`. The fix must
**disambiguate "identifier that names a registered constructor" from "identifier that is a fresh
variable."**

### The design
Insert a new **constructor arm before arm 5**:

```
4.5 Constructor pattern:  if pattern names a registered sum-type branch
      (parser type registry / SumTypeDefinition.Branches has it)
        → match iff the incoming value is a Variant whose Tag == pattern
          (val.(dsltypes.Variant).Is(pattern)), optionally binding payload;
        → if it is a registered constructor but the value's tag differs, CONTINUE (try next arm);
    else fall through to arm 5 (fresh-variable binding, unchanged).
```

Because `live`, `x`, `rest`, … are not registered constructors, they still hit arm 5 and bind as before —
`conway_step` keeps working. `Red`/`Green`/`Blue` now discriminate correctly. This also requires `spawn T:
Red` to actually box the value as a `dsltypes.Variant{Tag:"Red", Flag:…}` (today it carries only the sum
`BitFlag`), so the constructor arm has a tag to test.

### Follow-ons unlocked
- **Payload binding** (`Just x -> …`) once variants carry a payload slot (the struct already has `Value`).
- **Exhaustiveness checking** (Workstream D) — the set of required arms is literally
  `SumTypeDefinition.Branches`.

### Guardrails
- Add tour/exhaustive cases that currently mis-route (`OS: Windows → WINDOWS_…`) as regression asserts.
- Keep a test proving `conway_step`'s `live ->` binding still works (it must stay green).

---

## 2. Workstream B — First-class functions on the existing Kleisli arrow

This is the highest-leverage change and is **independent of typing** — every dynamic interpreter has it.

### Why the architecture is already friendly
`gatwd` *already* represents functions-as-values at the Go layer. `arrows/kleisli.go:13`:

```go
// Kleisli ... mathematically formalizes our Evaluator `Adapter`.
type Kleisli[A, B any] func(A) iter.Seq[B]
```

and its doc comment says so explicitly. An evaluator `Adapter[T] = func(ASTNode[T]) iter.Seq[ASTNode[T]]`
*is* a `Kleisli[ASTNode, ASTNode]`, and `pipelineToAdapter` already turns any pipeline text into one.
**First-class functions = exposing that arrow as a nameable, storable, applicable DSL value** instead of
the current re-parsed-text macro.

### Today's limitation
`def name = body` stores the **raw source string** (`i.Funcs[name] = pipelineStr`, `pkg/cli/cli.go:1999`)
and re-parses it on every `|> name` call. A lambda (`\x -> …`) is not a value — it only works as an inline
argument inside `map`/`filter`, and is single-arg (`SplitN(lambda,"->",2)`, `:1169`).

### The design
1. **A `Function` value kind.** Introduce a runtime value
   `type Closure struct { Params []string; Body ASTNode; Env *Scope }` carried with a dedicated `Function`
   `BitFlag`. A `Closure` evaluates to an `arrows.Kleisli[ASTNode, ASTNode]` (or a plain value→value fn for
   scalar use) — reusing the existing arrow, not inventing a new abstraction.
2. **Keep the body AST.** The parser already builds the body; stop re-stringifying it. `def`/lambda both
   construct a `Closure`. `i.Funcs` (text) migrates to binding `Closure` values in the normal environment,
   so functions live alongside other values and can be passed, returned, and stored in collections.
3. **Application in both positions.**
   - *Pipeline position* `x |> f`: `buildInitialStream`/`pipelineToAdapter` look up `f`; if it is a
     `Closure`, apply it instead of re-parsing macro text.
   - *Expression position* `f x` (new): add application to the infix/expression evaluator so functions can
     be called outside a pipe.
4. **Lexical closures.** `fork()` + `RuntimeNodes` already give a child scope; promote it to a captured
   scope chain stored *in the Closure at definition site* (lexical), replacing today's dynamic
   name-lookup-at-call capture.
5. **Multi-arg + currying.** Parse `\x y -> …` to `Params:["x","y"]`; applying fewer args than arity
   yields a new `Closure` with a partial `Env` — currying falls out for free.

### Payoff
Pass/return/store functions; higher-order user functions (not just builtin `map`/`filter`); real closures;
currying/partial application — all without a type system. It also makes the **operator-type matrix**
(STYLEGUIDE §7) expressible *in the language itself* (user-defined `fold`/`map` over any type), instead of
only as Go builtins.

### Risk
The one real design decision is unifying *expression* evaluation (value→value) with *pipeline* evaluation
(stream→stream). Resolve by making the canonical `Closure` form the Kleisli arrow and bridging scalar calls
through a single-element sequence — consistent with how the engine already treats everything as `iter.Seq`.

---

## 3. Workstream C — Loud dynamic typing + wire the dead diagnostics

Cheap, high-value, no new type theory.

- **Fail loud, not silent.** Today `3 + +` yields `0` and an unknown step yields nothing
  (`parser.go` silently skips unrecognized lines). Make operator application and step lookup raise real
  runtime errors through the existing `EitherT` channel (`Left` carries the message), e.g.
  *"cannot `++` an Int with a Set"*, *"unknown step `frobnicate`"*.
- **Surface the diagnostics that already exist.** `parser/errors.go` has `FormatError` /
  `SyntaxErrorTracker` producing Elm/Rust-style `🔥 Syntax Error … ^-- Expected … here` output — but it has
  **zero call sites** at the top level. Wire it into the line parser (`parser.go:28-86`) and the step
  dispatcher so malformed scripts report *where* and *why* instead of `(No output or node unhandled)`.
- **Make `--check` real.** It currently only greps for the literal `:=` (`pkg/cli/cli.go:2105`). Turn it
  into the entry point for Workstream D's pre-eval pass.

---

## 4. Workstream D — Gradual typing (the reachable slice of static safety)

A single **optional pre-evaluation pass over the AST**, ordered by power. Each tier is independent; stop
wherever appetite runs out. Interpretation is unchanged — unannotated code still runs fully dynamically.

1. **Optional annotations + best-effort checker (TypeScript-lite).** Allow `def inc : Int -> Int = …` and a
   non-blocking pass that flags the obvious: arity mismatches, unbound names, known-type operator misuse,
   wrong `spawn` seed. Catches the thing Elm is loved for — "you passed the wrong thing *here*" — with no
   inference.
2. **Local / bidirectional inference (not global HM).** Literals, operators, and `spawn T:` seeds give
   types locally; propagate them to catch more without annotations. Still a pre-pass, still interpreted.
   This is the ceiling of type safety while interpreted — an 80/20 of HM without global unification.
3. **Exhaustiveness checking.** Keyed off `dsltypes.SumTypeDefinition.Branches`: a `match` on a value of a
   known closed sum type warns on missing constructors. Local, registry-driven, no inference — and one of
   Elm's headline safety features. **Depends on Workstream A** (constructors must match first).

**Out of reach here (accepted):** whole-program soundness, type-class resolution, inference-driven
generics, and the no-runtime-type-error guarantee.

---

## 4b. Workstream E — Expression-level function forms across the operator-type matrix (RECOMMENDED)

> **Status:** DONE (2026-10-13). **Independent of the Kubernetes-graph plan**
> ([`docs/k8s_graph_design.md`](k8s_graph_design.md)) — that plan only *motivated* and *seeded* it. This is a
> general language change that stands on its own and benefits every program, so it is tracked here, not in
> the k8s plan.

### The observation

The k8s-graph ingest work ("option C": `fold` with a graph accumulator) needed a reducer body to *construct*
a value, not just combine scalars with operators. Delivering that required **function application in the
expression language** (Workstream B's expression slice, now landed) plus a few **expression-level builtins**
that take and return *real typed values* — `graph_add_node color id payload g`, `hamt_get key record`,
`hamt_path "a.b.c" record` — registered as `functions.DSLFunc`s in `pkg/cli/dslfuncs.go` and resolved via the
reducer/lambda name lookups.

The insight the user named: this is **"replacing strings with real types."** Historically a DSL operation was
reachable only as a *pipeline step* whose arguments are literal string tokens parsed out of the step text
(`graph_add_node Pod nginx` — `Pod`/`nginx` are raw strings). The new *expression-function* form lets the
arguments be **computed, real runtime values** (`graph_add_node (hamt_get "kind" r) (hamt_get "name" r) r g`).
The step form is point-free and convenient at the top of a pipe; the function form is applicative and
composes inside lambdas, folds, and comprehensions. They are complementary, and the function form is strictly
more expressive.

### The proposal

Fill the **operator-type matrix** (STYLEGUIDE §7) with expression-level `DSLFunc` forms for every collection
and ADT where the operation is mathematically valid — the horizontal complement to the pipeline-step forms
that already exist. The enabling machinery is done; the remaining work is *registration* plus asserts.

Illustrative cells (each the function form of an existing step, returning the real typed value):

- **List / Chunk** — `cons x xs`, `append xs ys`, `head xs`, `tail xs`, `nth i xs`, `length xs`
- **Set** — `set_insert x s`, `set_member x s` → `bool`, `set_union a b`, `set_intersect a b`, `set_diff a b`
- **HAMT** — `hamt_set k v m`, `hamt_del k m`, `hamt_keys m`, `hamt_values m` (`hamt_get`/`hamt_path` ✅ done)
- **Tree / BST** — `tree_insert x t`, `tree_member x t`
- **Pair** — `pair a b`, `fst p`, `snd p`
- **Either / Maybe** — `right x`, `left e`, `just x`, `nothing`, `either_map f e`, `maybe_default d m`
- **Variant** — constructor application yielding a `dsltypes.Variant`
- **Primitives** — `min a b`, `max a b`, `abs n` (arithmetic itself already works via operators)

### Design constraints (carry over from the landed slice)

1. **Preserve element types; do not stringify.** The graph pins `K`/`C` to `string`, so its constructors
   coerce via `dslStr` — correct *there only*. A `Set` of ints, a `Pair[int,int]`, or a `Maybe[float64]` must
   thread values through untouched. Each registered `DSLFunc` owns its own coercion policy.
2. **Honor the `appDisabled` list-literal guard.** gatwd serializes lists space-separated and re-parses them,
   so inside a bare `[...]` adjacent atoms stay separate elements; a function call used there must be
   parenthesized (`[(cons 1 xs), ys]`) — the same rule as Haskell. The parser already enforces this.
3. **One wiring generalization.** `dslBuiltinLookup` is currently consulted only by the `fold` reducer and the
   graph-lambda lookups. To make these functions callable in *every* expression context (a bare
   `let x = cons 1 xs in …`, any comprehension), wire the registry into the general evaluator name-resolution
   path. Small, clean, and the natural home for Workstream B's eventual `Closure`-backed user functions too.
4. **Namespaces don't collide.** A name may be *both* a pipeline step and an expression function (`graph_empty`,
   `hamt_get` already are): step dispatch keys off the pipeline-step text, function application off the
   expression lookup. Keep the two forms behaviorally identical.

### Definition of done (per cell, matching the matrix mandate)

The Go method on the type → the pipeline-step form (if infix/stream) → the expression-function `DSLFunc` in
`pkg/cli/dslfuncs.go` → a self-asserting case in `tests/dsl_exhaustive_test.gatwd` exercising the *function*
form inside a lambda/comprehension (not only the step form). Structural Go tests assert the raw memory type
returned, per STYLEGUIDE §4.

### Payoff & relation to Workstream B

This makes the operator-type matrix usable *from inside the language* — user-written `fold`/`map`/relation
predicates over any type — without waiting for the full `Closure` subsystem (B.1–B.5). When B's `Closure`
value kind lands, these registered `DSLFunc`s become the builtin tier that user closures compose with; the two
are consistent (both are callables resolved through the same lookup). Low risk: additive registration on top of
infra already covered by `parser/apply_test.go` and `pkg/cli/dslfuncs_test.go`.

---

## 5. Sequencing, effort, and the scorecard

| Order | Workstream | Effort | Depends on | Interpreter-safe? |
|---|---|---|---|---|
| 1 | **A** — fix constructor matching | ~0.5–1 day | — | yes (match-arm change) |
| 2 | **B** — first-class functions | ~several days (core evaluator) | — | yes (new value kind) |
| 3 | **C** — loud errors + wire diagnostics | ~1–2 days | — | yes (runtime + parser) |
| 4 | **D1/D2** — annotations + checker/inference | ~3–5 days | C (`--check` entry) | yes (pre-pass) |
| 4 | **D3** — exhaustiveness | ~1 day | A | yes (pre-pass) |
| — | **E** — expression-level function forms across the operator-type matrix (recommended, independent) | ~1–3 days, incremental per type | B expression slice (✅ landed) | yes (additive registration) |
| — | **F** — Layers 0–2 as a reusable Go library (proposal, see §6) | multi-week, incremental | coordinate with B/E (same files) | yes (additive first; DSL moves behind a facade) |
| — | **G** — program-level IR + pluggable backends (proposal, see §7) | IR: multi-week; backends: incremental | benefits from F (typed core); IR itself is independent | yes (interpreter becomes the first IR consumer) |

**How much of each hurdle this recovers:**

- **Hurdle B (first-class functions): ~fully.** Nothing about it needs static types or AOT; Workstream B
  closes it. Its expression-application slice is already landed, and **Workstream E** then turns that into
  broad day-to-day leverage by exposing the operator-type matrix as composable expression-level functions.
- **Hurdle A (static types): the gradual/dynamic-done-well tier.** Workstreams C + D deliver loud errors,
  optional annotations, local inference, and exhaustiveness — most of the *day-to-day* safety of Elm/Haskell,
  minus the global soundness guarantee that scope rules out.

Net: of the two hurdles from `docs/assessment.md`, B is essentially recoverable and A is recoverable to its
gradual ceiling — all while `gatwd` stays a lazy, interpreted, `iter.Seq`-trampolined language.
</content>

---

## 6. Workstream F — Layers 0–2 as a reusable procedural-Go library (proposal)

Not a language feature, but it shares files with Workstreams B and E, so it is tracked here.

**Observation.** Layers 0–2 depend only on each other (plus `dsltypes/` in `optics/` and `transformers/`)
and already use real generics. With modest decoupling they could serve as a standalone toolkit that adds
functional tools (persistent collections, `Either`/`Maybe`, state/reader monads, lenses, lazy `iter.Seq`
pipelines) to ordinary procedural Go code.

**Main gaps.** `any`-typed folds and `Elements()`; `Map` limited to `func(T) T`; DSL-specific files and
`dsltypes` imports inside the library layers; per-type duplication of the operator-type matrix.

**Direction (details in [`library_extraction_plan.md`](library_extraction_plan.md)):**

1. Boundary guard test for Layers 0–2 (`go list -deps`).
2. `iter.Seq[T]` as the interchange format, with generic `Map[T,U]`/`Fold[T,B]` written once. This also
   collapses the matrix cost from N×M to N+M while keeping parity (Mandate 2).
3. Move DSL-bound code behind a dynamic facade above Layer 2, so the DSL depends on the typed core and
   never the reverse.
4. Make `collections/` dependency-free; optionally split into `fp` / `persistent` / `graph` modules if a
   real external consumer appears.

**Trade-off.** Relaxing strict per-type rigor for the library surface buys a smaller, more idiomatic API
and faster iteration, at the cost of weaker per-type parity guarantees. The plan keeps parity via the
iterator interchange and keeps correctness via optional law tests. Status: proposal; step 1 (boundary guard
test, `internal/archtest/layers_test.go`) is done, the rest is unscheduled.

---

## 7. Workstream G — Program-level IR and pluggable backends (proposal)

**Observation.** The parser produces an AST only for *expressions* (`parser.Expr`: `BinaryExpr`,
`ApplyExpr`, `LambdaExpr`, …). Statements and pipelines (`def`, `let`, `match`, `|>` steps) are parsed
and executed together, line by line, largely by string handling in `pkg/cli/cli.go`; `match` works on the
string form of values. There is therefore no program representation that a different execution strategy
could consume, so neither the backend nor the middle layer can be swapped today. (`evaluator/` is already
a generic, DSL-independent engine; Go callers can use Layers 0–3 directly.)

**Direction (intent only; details will change):**

1. **G1 — Program-level IR.** Extend the expression AST to cover statements and pipelines, produced by the
   parser *before* anything runs. The existing interpreter becomes the first consumer. Side benefits: real
   error positions, a single place for the gradual checker (Workstream D) and exhaustiveness checks to
   run, and removal of the re-parsed-macro `def` special case.
2. **G2 — Closure-compiling backend (optional).** Compile the IR to Go closures instead of re-evaluating
   text. Conventional procedural execution, no new typing requirements, expected large speedup.
3. **G3 — Typed dialect + Go codegen (optional, only with a real use case).** Parse → IR → mandatory type
   check → emit Go source that instantiates the generic core (`List[int]`, `Either[E, T]`, …) → `go build`.
   Go generics cannot be instantiated at runtime, so this must be a transpiler, not a typed interpreter.

**Constraints and risks for G3.**

- The checker must become mandatory and complete. `pkg/cli/typecheck.go` is advisory and local today, and
  global inference is an explicit non-goal in §0, so G3 means a statically typed *dialect or subset*, not
  the current language.
- Dynamic features need typed equivalents or exclusion: string-form `match`, `spawn`, macro-style `def`
  bodies, `Any`. `Variant` would become generated Go types.
- Depends on Workstream F: codegen should target the typed core (generic `Map[T,U]`/`Fold[T,B]`, no `any`
  leakage), while the interpreter keeps using the `any` facade.

**Sequencing.** G1 first (useful on its own); G2 and G3 are independent follow-ons and are not committed
to. Coordinate G1 with Workstream F step 3, since both move code out of `pkg/cli/cli.go` and the
DSL-bound files. Status: proposal, unscheduled.
