# Design & Fix Plan: task-closures cannot contain `|>` (`parser.ParseInfixExpr`)

> **Handoff for a fresh context.** Nothing in this plan is implemented yet — this
> session only investigated and documented. Read this whole file before writing any
> code; it supersedes nothing (first handoff on this topic). Related, already-shipped
> work that hit this bug as a symptom: `docs/manifest_live_parity_design.md`
> (IMPLEMENTED; its `ParityCommand` is the worked-around caller, not a fix). The
> general trap-table entries this plan grounds: `AGENTS.md` §5a (Prevention Rules) and
> `docs/development_playbook.md` §6 (both already updated with the corrected,
> probe-confirmed description of the bug — read those two rows first, they are the
> one-paragraph version of this whole document).

---

## 1. The bug, precisely (confirmed by direct probe, not inferred)

A **task-closure** is a `def name = \params -> body` run via `RunOptions.Task`
(`cli.TaskCommand` builds exactly this). `Interpreter.RunTask`
(`pkg/cli/run.go:294`) special-cases a closure-shaped def: it calls
`defClosure(name)` (`pkg/cli/closures.go:35`), and if that succeeds, applies the
resulting `*functions.Closure` to `ARGS` directly (`c.Apply([]any{ARGS})`) instead
of reinterpreting the def's name as a pipeline-step string.

`defClosure` detects "closure-shaped" by parsing the def's raw source text with
**`parser.ParseInfixExpr`** (`parser/infix.go:1491`) and checking the result is a
`parser.LambdaExpr`. **`ParseInfixExpr` cannot parse `|>` at all — not once, let
alone more than once.** Confirmed directly (not from task-runner symptoms, which
are a red herring — see §2) with a throwaway `go run` against the `parser` package:

```go
parser.ParseInfixExpr(`\args -> (nth 0 args) |> manifests |> k8s_graph`)
// → err: "unexpected trailing tokens: |"
```

Every `|>`-bodied lambda, with zero exceptions, fails this parse. `defClosure`
treats a parse failure exactly like "this def is not closure-shaped at all" —
`return nil, false` (`closures.go:52`) — so `RunTask` falls through to its OTHER
branch (`run.go:312-321`): it reinterprets the **literal source text of the
lambda**, `\`-prefix included, as a pipeline-step string via
`i.parsePipelineSteps([]string{name})`, and runs `ARGS` through it as the input
stream. That reinterpretation is never meaningful for text starting with `\` — it
just produces whatever garbage falls out of treating `\args -> f (g x) (h y)` as a
sequence of step names, which is why the surfaced errors ("unknown step or command:
k8s_parity_report", "undefined identifier: build_from" — see §2) name things that
are, in fact, perfectly well defined. **The error message is actively misleading
about where the real problem is**, which is the second, independent defect worth
fixing (§4, Fix C) regardless of whether the first one (§4, Fix B) ever lands.

### Why this is worse than "the parser is missing one feature"

`defClosure` is also what resolves a **named reference to another lambda-bodied
def from inside a closure's body** (any identifier lookup that could be a closure
goes through `callableNamed` → `defClosure`, `closures.go:73-82`). So even a
task-closure whose **own** text contains zero `|>` fails the same way the moment it
calls, by name, some other def whose body contains one — that other def's own
`defClosure` call hits the identical parse failure. This is why `mytask = \args ->
k8s_parity_report (build_from (nth 0 args)) (...)` (zero `|>` in `mytask`'s own
text) still failed with `"undefined identifier: build_from"`: resolving
`build_from` — whose body is `\path -> path |> manifests` — recursed into the same
broken check.

---

## 2. Repro scripts (preserved so the next session does not re-derive them)

All run against a `gatwd-k8s`-equivalent binary (any Edition with `manifests`
registered). `KUBECONFIG=/nonexistent` is irrelevant to these — they fail before
any cluster/file I/O.

```gatwd
-- (1) the baseline: a closure with NO pipe anywhere resolves and runs for real.
def inc = \x -> x + 1
def callit = \args -> inc (nth 0 args)
-- `task callit -- 5` → runs, fails INSIDE inc on "String + Int" (real execution,
-- proves defClosure succeeded and Apply actually ran).

-- (2) one pipe, directly in the task-closure's own body: broken.
def direct_task = \args -> (nth 0 args) |> manifests |> k8s_graph |> graph_order
-- `task direct_task -- somepath` → "ManifestAdapter: expected a file or directory
-- path, got functions.Char" — NOT real success: RunTask's pipeline-reparse
-- fallback happened to still invoke `manifests` on a byte of garbage text; this
-- looked like progress during the original investigation but is not evidence
-- defClosure/ParseInfixExpr handled the pipe. Direct probe (§1) is the ground truth.

-- (3) two independent pipe chains, combined with a plain operator: broken.
def two_pipes_task = \args -> ((nth 0 args) |> manifests |> graph_order) + ((nth 0 args) |> manifests |> graph_order)
-- `task two_pipes_task -- somepath` → "map: unknown step or command: ((nth"

-- (4) calling a separately-named pipe-bodied helper: broken (the transitive case, §1).
def build_from = \path -> path |> manifests |> k8s_graph
def mytask = \args -> k8s_parity_report (build_from (nth 0 args)) (build_from (nth 0 args))
-- `task mytask -- somepath` → "undefined identifier: build_from"

-- (5) the WORKING alternative used in production (pkg/k8sgraph/parity_command.go):
-- a top-level `let` (not inside any lambda) with two independent `|>` chains,
-- run via RunOptions.Eval, not RunOptions.Task. This goes through the interpreter's
-- general statement evaluator (EvaluateToValues, see §4 Fix B), not defClosure.
let m = ("somepath" |> manifests |> k8s_graph) in let l = ("default" |> k8s_resources |> k8s_graph) in k8s_parity_report m l
-- works correctly — this is the existence proof that the INTERPRETER overall can
-- do this; only the task-closure DETECTION/EXECUTION path (defClosure) cannot.
```

A fresh session can paste any of these into a scratch `.gatwd` file and reproduce
in under a minute; no fixture files are needed for (1)-(4).

---

## 3. Why this matters beyond the one feature that found it

`pkg/k8sgraph/parity_command.go` worked around this by not using
`cli.TaskCommand`/`RunOptions.Task` at all — it builds a `let ... in ...` string in
Go and runs it via `RunOptions.Eval`. That is a legitimate, permanent pattern (now
documented as such), **not** just a stopgap for this one verb. But it means:

- `cli.TaskCommand` — the generic, domain-free "promote a DSL def to a cobra
  subcommand" bridge (`docs/development_playbook.md` §5b) — silently cannot serve
  any task whose logic needs more than trivial (zero-pipe) combination of values.
  Every *future* verb that needs to build more than one thing via `manifests`/
  `k8s_resources`/any other pipeline-only source and combine them will hit this
  exact wall, rediscover it the same slow way, or (worse) ship a silently-wrong
  result if the misleading fallback error happens to look like something else.
- The failure mode is a **Mandate 6 (Fail Loud) violation** independent of whether
  the missing capability is ever added: a lambda that plainly starts with `\` and
  fails to parse should never be silently reinterpreted as a pipeline-step name.
  That reinterpretation cannot ever be right for `\`-prefixed text, pipe or no pipe.

---

## 4. Candidate fixes, in recommended order

### Fix C — make the failure loud (do this first; cheap, low-risk, independently valuable)

In `defClosure` (`closures.go:35`), distinguish "this def's source doesn't look
like a lambda at all" (current `return nil, false` at line 42, correct as is — a
genuine pipeline fragment) from "this def's source starts with `\`/`(\` but failed
to parse or evaluate as one" (lines 52 and 65-66 today: still `return nil, false`,
silently). For the second case, surface a real error instead: the simplest
mechanism is likely a sentinel error type (`var errLambdaParseFailed = ...` or a
small struct implementing `error`) that `defClosure` returns via a changed
signature, or a second return value `(c *functions.Closure, ok bool, parseErr
error)`; `RunTask` (`run.go:294`) then returns `parseErr` directly instead of
falling into the `i.parsePipelineSteps` branch when it's non-nil. Message should
name the def, say it looks like a lambda, show the underlying parse/eval error, and
(until Fix B lands) say plainly that a task-closure body cannot contain `|>` yet,
pointing at this doc. **Verify**: every existing test that exercises `RunTask` on a
genuine pipeline-fragment task (`main`/`dot`/`wide`-shaped, `pkg/cli/command_test.go`
`TestCommandTreeFragmentTask`, easyk8s's own task tests) must be unaffected — those
defs never start with `\`, so `defClosure`'s FIRST check (line 41-43) already
returns false before any parse is attempted; confirm this with the existing test
suite, don't just assume it from reading the code.

### Fix B — make it actually work (the real capability fix; needs investigation before design, not just before coding)

**Recommended direction, not yet a committed design:** `defClosure` should stop
relying solely on the narrow `parser.ParseInfixExpr` grammar and instead bind the
lambda's parameter to the applied argument (exactly as `pipelineLambda`,
`closures.go` — search `func (i *Interpreter) pipelineLambda` — already does for
map/filter lambdas with pipeline bodies: a child `*Interpreter` with the parameter
name added to `RuntimeNodes`/`Variables`), then evaluate the **body** through the
interpreter's general-purpose statement evaluator — `Interpreter.EvaluateToValues`
(`cli.go:1953`) is the leading candidate, since it is demonstrably already capable
of everything needed: `let`, multiple independent `|>` chains, calling named defs
that themselves contain `|>` — repro (5) in §2 is proof, because that is exactly
the code path a top-level script statement goes through.

**Open questions the next session must resolve before writing code (do not guess):**

1. **Does `pipelineLambda` already generalize to this, or is a new helper needed?**
   `pipelineLambda(varName, body string)` (`closures.go`, search for its
   definition) splits `body` with `splitPipelineAware` and feeds
   `child.parsePipelineSteps(subParts)` a stream seeded with the bound value —
   read it closely enough to determine whether its **first** split segment must be
   exactly the bound variable name (resolved through the existing runtime-node/
   variable step-substitution machinery), or whether an arbitrary expression there
   (`(nth 0 args) |> manifests |> ...`, not just `args |> manifests |> ...`) already
   works. If it requires a bare variable, task-closures need something closer to
   "evaluate the whole body via `EvaluateToValues` with the parameter bound," not a
   direct reuse of `pipelineLambda` as-is.
2. **Error-value convention.** `functions.Closure.Apply` returns `(any, error)` — a
   real Go `error`. `EvaluateToValues` returns `[]any` and, for most failure modes
   in this codebase, failures surface as a rendered `"[ERROR] ..."` **value**
   inside that slice (via `functions.Either`'s Left channel, rendered), not
   necessarily as a Go `error` on the return path itself (`cli.go:1953`'s own
   early-return cases, like a `rewriteLine` failure, DO return a bare Go `error`
   boxed in `[]any{err}` — but a failure *inside* a pipeline the function evaluates
   may come back as a string-rendered `[ERROR]` value instead). Whatever bridges
   `EvaluateToValues`'s output back into a Closure's `(any, error)` contract must
   correctly turn an embedded `"[ERROR] ..."` result into a real `error` (so a
   failing task-closure actually fails the task, per Mandate 6), not pass it
   through as a successful string value. Read `RunTask`'s own existing handling of
   `[ERROR]`-prefixed lines (`run.go:322-330`) for the established convention to
   reuse, rather than inventing a new one.
3. **Multi-value results.** A pipeline can legitimately emit more than one node.
   `pipelineLambda` takes "the first result, or error" (`closures.go`, the `for r
   := range ... { return v, nil }` / fallthrough-to-error shape). Decide whether a
   task-closure's pipe-containing body should have the same "first value" semantics
   or something else (e.g. an error if it emits anything other than exactly one
   value, to fail loud on an ambiguous case rather than silently dropping extra
   output) — this is a real design choice, not an implementation detail; record
   the decision and why.
4. **Scope of the change.** Should this new capability apply to *every* task-closure
   unconditionally (replace `ParseInfixExpr`'s role in `defClosure` entirely with
   the `EvaluateToValues`-based approach), or should `ParseInfixExpr` still be tried
   FIRST (cheaper, already fast) with the new path as a fallback specifically when
   parsing fails because of a pipe? The fallback-only approach is lower-risk
   (unaffected behavior for every existing zero-pipe task-closure) and is the
   recommended starting point; only promote it to "the only path" if the fallback
   version is clearly superior in every case once built.

### Fix A — teach `ParseInfixExpr`/`InfixParser` to understand `|>` as a first-class operator (NOT recommended as the first move)

Extending the core expression grammar so `|>` is a real `BinaryExpr`-producing
operator at the `InfixParser` level would close the gap at its root and might look
more "proper." **Recommended against, for now**: the right-hand side of `|>` is
usually a **step name** (`manifests`, `k8s_graph`), not an ordinary callable value —
evaluating it as a plain expression would need to thread through the whole
`evaluator.PipelineStep`/stream/`Either` machinery that `Interpreter.parsePipelineSteps`
already owns, not just extend `Expr.Eval`'s scalar-value evaluation. That is a much
larger, riskier surface (the main interpreter/parser, touched by every `.gatwd`
script in the repo) for the same outcome Fix B gets more cheaply by reusing
existing pipeline-execution machinery instead of reimplementing it inside the
expression evaluator. Revisit only if Fix B turns out to be structurally impossible.

---

## 5. Verification plan (once Fix B has a concrete design)

Per `AGENTS.md` §5 and `docs/development_playbook.md` §3 — this touches
`pkg/cli/closures.go` and `pkg/cli/run.go`, interpreter-adjacent code, so the full
bar applies, not just `go test ./pkg/cli/...`:

1. **Pinning tests first** (`PROMPTS.md` §4 TDD discipline, already modeled in this
   session's own Phase 1 — `git stash` the fix, confirm the new tests fail, restore,
   confirm they pass): add the five repros from §2 as table-driven cases in
   `pkg/cli/closures_test.go` (already exists; follow its style) — (1) and (5) must
   keep passing throughout (they already work); (2)-(4) are the ones that currently
   fail and must start passing once Fix B lands; add one more case for Fix C alone
   (a lambda with a syntax error unrelated to pipes, e.g. unbalanced parens, must
   surface a parse-error message naming the def, not fall through to a confusing
   pipeline reinterpretation).
2. **Negative control for Fix C specifically**: confirm a genuine zero-parameter
   pipeline-fragment task (`main`/`dot`-shaped) still resolves via the OTHER
   `RunTask` branch and is completely unaffected — add this as its own explicit
   test case, don't rely on the existing suite happening to cover it.
3. `go test ./... && go vet ./...` and `./run_dsl_tests.sh` (`AGENTS.md` §5).
4. **The script regression diff** (`docs/development_playbook.md` §3): build a
   baseline binary from a `git worktree` of the commit before this fix, run every
   `*.gatwd` script in the repo through both binaries, diff the output. Parser/
   interpreter changes specifically require this — don't skip it because the
   change is "only" in `defClosure`.
5. `gatwd --check` silent over every `.gatwd` file in the repo (no new structural
   errors introduced).
6. `go test -race ./pkg/cli/... ./evaluator/... ./parser/...` (`AGENTS.md` §5,
   interpreter/parser changes).
7. Once Fix B lands, go back to `pkg/k8sgraph/parity_command.go` and
   `cli.TaskCommand`: decide whether `ParityCommand` should be simplified to use
   `cli.TaskCommand` now that the underlying limitation is gone (functionally
   equivalent either way; simplifying removes the now-unnecessary custom cobra
   command and its `let...in` string-building, converging back onto the generic
   bridge) — this is a nice-to-have cleanup, not required, and should not block
   closing this fix's own phase.

---

## 6. Mandates and constraints that apply

- **Mandate 5 (Clean Core)**: `pkg/cli` is core (domain-free); this fix stays
  entirely within it and `evaluator`/`parser` if Fix B needs anything there — no
  Kubernetes or other domain identifier belongs anywhere near this change.
- **Mandate 6 (Fail Loud)**: the whole point of Fix C; don't regress it while
  building Fix B (a task-closure that fails inside its new pipe-execution path must
  still surface a real error, not silently succeed with a wrong/partial value —
  see open question 2 above).
- **Mandate 8 (Phase-Gated, Decision-Recorded)**: do Fix C and Fix B as two
  separately-gated commits/phases (Fix C is small and independently valuable even
  if Fix B is deferred further); record the four open-question decisions in §4
  Fix B in this file's own revision once made, with rationale, before implementing
  past them — they are real either-way choices, not implementation details to
  improvise silently.
- **Shared working tree**: re-read `closures.go`/`run.go` immediately before
  editing (this is core, high-traffic code); this file may itself have been
  updated by someone else by the time you pick it up — check `git log -- docs/
  task_closure_pipe_fix_design.md` first.

---

## 7. First steps for the next session

1. `git log --oneline -5 -- pkg/cli/closures.go pkg/cli/run.go` — confirm nothing
   has changed here since this plan was written (2026-10-14).
2. Re-run the §2 repros yourself (they take under a minute) to confirm the
   starting state before touching anything.
3. Read `pipelineLambda`'s full body and `splitPipelineAware`'s splitting rule
   closely enough to answer open question 1 in §4 Fix B — this determines whether
   Fix B is "call pipelineLambda" or "build something new alongside it."
4. Read `EvaluateToValues`'s full body (`cli.go:1953` onward) to answer open
   question 2 (the `[ERROR]`-value vs. Go-`error` convention) with certainty, not
   inference.
5. Implement Fix C alone first, gate it fully (§5), commit it on its own.
6. Design Fix B concretely (answer all four open questions, write them into this
   doc), implement, gate fully, commit separately.
7. Optionally revisit `pkg/k8sgraph/parity_command.go` per §5 point 7.

---

## 8. Pointers

- `pkg/cli/closures.go` — `defClosure`, `callableNamed`, `pipelineLambda`,
  `bindingAdapter`, `splitPipelineAware`'s caller.
- `pkg/cli/run.go` — `RunTask` (`:294`), the two branches this plan's Fix C/B change
  the boundary between.
- `pkg/cli/cli.go` — `EvaluateToValues` (`:1953`), `parsePipelineSteps` (`:240`),
  `splitPipelineAware` (`:1693`).
- `parser/infix.go` — `ParseInfixExpr` (`:1491`), `InfixParser.parsePrecedence`/
  `parseApplication`/`parsePrimary` (the restricted grammar Fix A would have to
  extend, if ever pursued).
- `pkg/cli/closures_test.go` — existing closure test file; add the §5 cases here.
- `pkg/k8sgraph/parity_command.go`, `pkg/k8sgraph/parity_test.go` — the working
  `let ... in` / `RunOptions.Eval` pattern, proof the interpreter itself handles
  this correctly outside the task-closure path.
- `AGENTS.md` §5a, `docs/development_playbook.md` §6 — the short trap-table
  versions of this document, already corrected to match the probe-confirmed root
  cause; keep them in sync if this document's understanding changes.
- Memory: `project-manifest-live-parity` (this session's notes on where the bug
  was found) — read-only context, not authoritative over this document.
