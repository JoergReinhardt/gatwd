# Language Roadmap — Staying Interpreted: First-Class Functions & Gradual Typing

> **Status:** in progress (2026-10-12). Companion to [`docs/assessment.md`](assessment.md), which
> diagnosed the two main gaps vs. Haskell/Elm. This doc scopes *how much of those gaps is reachable
> while keeping `gatwd` an interpreted, dynamically/gradually-typed language* — and sketches the
> concrete approach against the code that exists today.
>
> **Progress:**
> - ✅ **Workstream A — ADT constructor matching: DONE.** `Parser.IsConstructor`/`MakeVariant` added;
>   a constructor arm (matching a `Variant` tag or exact tag string) now runs before the variable-binding
>   arm in the `match` evaluator, so `match { Red -> … | Green -> … }` routes correctly while fresh
>   variables (e.g. `conway_step`'s `live ->`) still bind. Covered by `tests/dsl_exhaustive_test.gatwd`
>   (#56) and `parser.TestParser_IsConstructorAndMakeVariant`.
> - ◑ **Workstream B — first-class functions: pipeline-position slice DONE.** A bare/paren lambda in
>   pipeline position now applies to the incoming value, making lambda-bodied `def`s first-class and
>   composable: `def inc = \x -> x + 1` then `5 |> inc`, `5 |> (\x -> x*10)`, `inc |> dbl`,
>   `let f = \x -> … in 5 |> f`, and `map inc` all work (test #57). **Deferred (higher-risk, infix-parser
>   surgery):** expression-position application `f x`, multi-argument lambdas, and currying — see §2.
> - ☐ **Workstream C, D:** not started.

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

## 5. Sequencing, effort, and the scorecard

| Order | Workstream | Effort | Depends on | Interpreter-safe? |
|---|---|---|---|---|
| 1 | **A** — fix constructor matching | ~0.5–1 day | — | yes (match-arm change) |
| 2 | **B** — first-class functions | ~several days (core evaluator) | — | yes (new value kind) |
| 3 | **C** — loud errors + wire diagnostics | ~1–2 days | — | yes (runtime + parser) |
| 4 | **D1/D2** — annotations + checker/inference | ~3–5 days | C (`--check` entry) | yes (pre-pass) |
| 4 | **D3** — exhaustiveness | ~1 day | A | yes (pre-pass) |

**How much of each hurdle this recovers:**

- **Hurdle B (first-class functions): ~fully.** Nothing about it needs static types or AOT; Workstream B
  closes it.
- **Hurdle A (static types): the gradual/dynamic-done-well tier.** Workstreams C + D deliver loud errors,
  optional annotations, local inference, and exhaustiveness — most of the *day-to-day* safety of Elm/Haskell,
  minus the global soundness guarantee that scope rules out.

Net: of the two hurdles from `docs/assessment.md`, B is essentially recoverable and A is recoverable to its
gradual ceiling — all while `gatwd` stays a lazy, interpreted, `iter.Seq`-trampolined language.
</content>
