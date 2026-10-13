# Follow-up plan: the larger items left after the hardening run

> Status (2026-10-14): written after `docs/hardening_plan.md` closed. The hardening plan is
> finished; every item below was **deferred there with a reason** and none blocks anything.
> This file turns the deferrals into phases that can be picked up one at a time. Decisions
> already taken are in `hardening_plan.md` §1; open ones are marked **Decision** and carry a
> recommendation. The small silent-failure fixes (guard-only comprehension, collectors
> dropping error nodes, `zip`) are not here: they were done as their own tracks first (merged 2026-10-14).

## 0. How to work through it

- **Same delivery rules as before** (`AGENTS.md` Mandates 6 and 8, `.agents/skills/phase-delivery/SKILL.md`): one commit per sub-step, every gate green, negative-control each guard, write deferrals down, stage by name, commit and push only when asked.
- **Parallel where the files allow** (user instruction, 2026-10-13). The dependency table in §9 says which phases touch disjoint files.
- **Resource limits** (user, 2026-10-13, after a laptop crash): at most **two** heavy agents at once; every command under `nice -n 10`, `GOMAXPROCS=4`, `go test -p 2`; `-race` on touched packages per commit and the full set once per merge; **never** run `benchmarks/*.gatwd` in a routine gate. The one exception is phase F2, which changes the evaluator's hot path: it runs the two benchmarks **once before and once after**, deliberately, and says so in its report.
- **Gates** are the usual set (`gofmt`, `go vet`, `go test -p 2 ./...`, `./run_dsl_tests.sh`, `-race`, silent `gatwd --check` over every `*.gatwd`, script regression diff for interpreter, parser or evaluator changes).

## 1. Phase F1: error positions (`parser.FormatError`)  — size M, highest user value

**Today.** `parser/errors.go` (60 lines) has a `SyntaxErrorTracker`, an `Expect` combinator and `FormatError`, an Elm/Rust-style pointer at the failing character. Nothing outside the file calls it. A bad statement reaches the user as `cannot evaluate "<line>": <reason>` (`pkg/cli/cli.go`, near line 2219): it names *what* failed but not *where*. The infix expression parser (`parser/infix.go`) and the line parsers do not thread a tracker, which is why W0/W1b deferred this.

**Goal.** A syntax error names the line and column in the script, quotes the line and points at the column, in the REPL, in `-e`, in script files and in `--check`.

**Design.**
1. Give the lexer a position on every token (offset into the statement) and the statement joiner a map from joined-statement offsets back to file line and column. This is the real work: the joiner (`logicalLines`) currently discards positions.
2. Make the infix parser report the furthest failing token position through the existing tracker instead of a bare error.
3. Render with `FormatError` in `cli.go` where `cannot evaluate` is built; keep the old sentence as the first line so existing substring tests still match, add the pointer lines below it.
4. `gatwd --check` prints `file:line:col: error: …` (the shape editors parse).

**Decision D-F1.** Keep the old message as the first line (recommended, no test churn) or replace it. Recommended: keep and append.

**Tests.** Table of malformed statements with the expected line, column and pointer; a multi-line statement; a string literal containing delimiters (the earlier W2 joiner bugs); negative control: break the position map and watch the table fail. Docs: `docs/error_handling.md`, `docs/dsl_guide.md`.

## 2. Phase F2: interruptible evaluation of a closure that never returns  — size L, touches the hot path

**Today.** Cancellation (`--timeout`, SIGINT/SIGTERM, the output limit) is cooperative and checked at pipeline-node boundaries (`execState.cancelled()`, `pkg/cli/policy.go`; called at `cli.go` around line 2450 and between statements in `run.go`). `Closure.Apply` (`functions/closure.go`) runs the body to completion with no check, and `functions/` is Layer 0 so it cannot import `pkg/cli`. A pure infinite recursion (`def spin = \n -> spin (n + 1)`) is therefore not stopped by a timeout or a signal.

**Goal.** A timeout or stop request ends a runaway pure computation within a bounded time, as an `[ERROR] … timed out` value, never a hang and never a leaked goroutine spinning forever.

**Options.**
- **A. A cooperative budget in the closure call path** (recommended). `functions` defines a tiny hook, for example `type Interrupter interface{ Interrupted() error }`, carried by the evaluation environment; `Closure.Apply` checks it every N applications (a counter, N around 1024) and returns the error. `pkg/cli` implements it with `execState.cancelled`. Cost: one counter increment per application; measure it.
- **B. Run the evaluation in a goroutine and abandon it on timeout.** Simple, but the goroutine keeps burning CPU and memory after the "timeout"; rejected.
- **C. A recursion-depth or step limit as a language rule** (like a fuel parameter). Changes semantics for legitimate deep programs; offered only as an additional, opt-in `--max-steps`.

**Decision D-F2.** Option A, plus optional C later. Recommended: A.

**Tests.** A spinning closure at top level, inside `map`, inside a comprehension and inside a `fold` is stopped within a deadline (use short timeouts, no sleeps); a legitimately deep non-spinning recursion still finishes; the interrupt error is a value on the Left channel and the operator reports it with the existing `timeout` diagnostic. **Performance protocol:** run `benchmarks/ackermann.gatwd` and `benchmarks/conway_gun.gatwd` once on the baseline and once after; accept at most a few percent; record the numbers in the phase entry. Negative control: remove the check, watch the spin test hang-fail within its deadline.

**Risk.** The check must not run while holding the interpreter's locks; test with `-race`.

## 3. Phase F3: signal handling in `gatwd` itself  — size S

**Today.** Only the operator (`ApplianceTree.serve`) handles SIGINT/SIGTERM. `gatwd script.gatwd` and the REPL die on the default signal handler; a second Ctrl-C is the only way out of a long script, and child OS commands started by exec steps are killed only if the process dies.

**Goal.** SIGINT/SIGTERM cancel the running script's context (the machinery from W7 and Track K already exists: `RunOptions.Context`), stop child commands, print one line, exit non-zero (130 for SIGINT, 143 for SIGTERM is the Unix convention; a decision below). The REPL: Ctrl-C cancels the current statement and returns to the prompt; Ctrl-D exits.

**Decision D-F3.** Exit statuses 130/143 (recommended, what shells expect) or the existing 1. A second signal exits immediately.

**Depends on** F2 for the "stops a spinning closure" part; works without it for everything else. Tests: a real-binary signal test like the operator's (skip where unsupported).

## 4. Phase F4: rollout completion  — size M

**Today** (`adapters/kubernetes_rollout.go`, Track L). `Applier.Status` is one reading; `Undo`, `Pause`, `Resume`, status work for Deployments only. StatefulSets and DaemonSets keep their history in `ControllerRevision` objects, not ReplicaSets.

**F4a `rollout status --watch`.** A step that waits until the rollout is complete or a deadline passes, using the cluster's watch API with the existing `--timeout` context. Output: a stream of status records ending in a final `complete` or an error. **Decision D-F4a:** a separate step `k8s_rollout_wait` (recommended: `k8s_rollout_status` stays a pure single read, so the family keeps its "status is read-only and ungated" rule; the watch is still read-only and ungated) versus a flag on `status`. Fake-client tests drive a fake watcher; never a real cluster.

**F4b StatefulSet and DaemonSet.** `Undo` selects the revision through `ControllerRevision` (owner reference + `revision` field), `Status` uses `updatedReplicas`/`currentRevision` for StatefulSets and `numberReady`/`updatedNumberScheduled` for DaemonSets, `Pause`/`Resume` are not defined for them by Kubernetes (loud, named error, not a silent no-op). Same capability gate (`cluster-rollout`), same dry-run default.

**Tests.** Fake dynamic client with emulated writes (`k8sgraphtest.EmulateWrites`); negative: undo to a revision that does not exist is an error naming the available revisions. Docs: how-to §3.5, `doc_test.go` rows.

## 5. Phase F5: richer rule results (a message per violation)  — size M, **needs a decision**

**Today.** A rule result is `[Rule, resource]` (text, JSON `{"rule","resource"}`), and a cycle is a list resource. `str_format` exists but a rule cannot carry a human message, so the report names *what* and *where* but not *why*.

**Options.**
- **A. Optional third element:** `[Rule, resource, message]`, fully backward compatible (two-element results unchanged, JSON gains `"message"` only when present). Recommended.
- **B. A record type** (`Violation{rule, resource, message}` as a HAMT or Variant): cleaner, but breaks every existing rule and override.
- **C. Message by convention in the rule name:** no change, no structure; rejected.

**Decision D-F5.** Option A. Touches `pkg/cli/appliance_report.go` (the one place text and JSON are rendered), the embedded rules (optional: give the six rules messages), the how-to and `doc_test.go`. Negative fixture per rule with an expected message. Mandate 7: the messages are `.gatwd`, never Go.

## 6. Phase F6: small language and library items  — each S, independent, pick when a consumer appears

| Item | Today | Plan |
| :-- | :-- | :-- |
| Exponent float literals | `1e-3` is not lexed (`parser`/`functions.ParseNumberLiteral`) | Lex `1e3`, `1.5e-3`, `2E10` as `Float` (an `e` exponent makes it a Float, never an Int); `1e400` is a range error, not `+Inf`; pinning tests for `1e3x` (still a lex error), interaction with the `3r4` and duration suffixes (`5e` is not a unit), grammar and guide text. |
| `Unit` | deferred in W2 (no consumer) | Add `()` as a `Unit` value only when a function needs it (for example effectful steps returning nothing); until then keep the deferral. |
| `json_encode` of a `Ratio` | error | **Decision D-F6a:** encode as the exact string `"3r4"` (recommended, reads back through the DSL, never lossy) or as a Float (lossy, rejected). `yaml_encode` follows. |
| `strptime`-style time parsing | only RFC 3339/ISO literals and `time_parse` | A layout function `time_parse_with layout text` using the documented strftime subset already used for formatting; `Maybe` for malformed text. Only when a rule needs a non-ISO timestamp. |
| `env` builtin | none; `AllowEnv` gates nothing | `env_get name` returning `Maybe String`, gated by `CapEnv` through `RequireCapability` and the same structural guard test. Only when something needs it. |
| `--check` advisory beyond literals | only literals are judged (the 2^53 Int-vs-Float case) | Extend the checker's value-range reasoning to annotated bindings; keep the zero-false-positive rule on every committed script as the acceptance test. |

## 7. Phase F7: operator configuration  — size S–M, low priority

- **`GATWD_<FLAG>` environment defaults** for the operator and edition flags (design in `hardening_plan.md` §4.3): flag > environment > default, one table in `pkg/cli/command.go`, documented per flag. Useful for container images (`env:` in `deploy.yaml`) instead of long `args:`. **Decision D-F7a:** do it for the operator only first (recommended).
- **Watch the ConfigMap** instead of interval-driven reload: a `fsnotify`-style watch is a new third-party dependency in core, which the Clean Core rule forbids; the in-process alternative is polling the file's modification time on a short ticker (recommended, stdlib only, no allowlist change). The reload path already keeps the last good rules on failure.

## 7a. Phase F8: multi-line statements and expression-group diagnostics  — size S–M, found 2026-10-14

Raised by the user's question "why are all gatwd defs so long, is there a multiline syntax?". Three findings, each reproduced with the built binary.

**F8a. Statement continuation is too narrow.** `logicalLines` (`pkg/cli/cli.go`, shared by script evaluation and `gatwd --check`) joins physical lines only when delimiters are unbalanced, a string is unterminated, or the buffer ends in `=`, `|>`, `||>`, ` let`, ` in` or `\`. So a bracketed body already spans lines (a comprehension over five lines works), but
```
def area = \w h ->
  w * h
```
is two statements and fails loudly (`undefined identifier: w`). That is why the embedded rules (`cmd/gatwd-k8s-operator/rules/compliance.gatwd`, `lib/k8s/*.gatwd`) are single very long lines: there was no natural place to break them.

*Design* (the Go/Kotlin rule: a line that cannot end a statement continues):
1. Trailing `->`, `,` and binary operators (`+ - * / // % ** ++ == != < <= > >= && || <+>`) continue the statement.
2. A line that *starts* with `|>`, `||>`, `++`, `<+>` or `|` continues the previous one (a statement can never start with them, so this is safe).
3. One implementation in `logicalLines`, so `--check` and evaluation agree; keep the position map F1 needs (F1 and F8a share this function: same agent, F1 first or together).
4. Reformat the embedded rules and the k8s prelude into readable multi-line form once it works (comments per clause, one generator or guard per line), pinned by the existing parity and rule tests so behaviour cannot drift.

*Risk.* An exec step whose last argument is an operator-like token (`echo a,`, `echo -`) would now continue onto the next line. **Decision D-F8a:** apply rules 1 and 2 to every statement (recommended: simple, one rule to document; the regression diff over every committed script and the k8s fixtures is the acceptance test and shows any victim) or only to statements that begin with `def`, `let` or a lambda (safer, but two rules). Recommended: every statement, and add an escape only if the diff finds a real victim (then record it here).

*Tests.* A table of (physical lines, expected logical statements) including every operator, a trailing `,` inside a list literal, comments between continued lines, a string literal ending in an operator character (must not continue), `--check` line numbers pointing at the *first* physical line of a joined statement; negative control: remove rule 1 and watch the area example fail. Docs: the joiner's rules are not documented anywhere today; add them to `docs/dsl_guide.md` and `docs/grammar.md`.

**F8b. A pipeline inside an expression group gives bad or no guidance.** `(3 |> dbl) + 1` and `1 + (3 |> dbl)` at top level fail with `expected closing parenthesis ')'`, which does not say what is wrong. Inside a `match` arm the same text surfaces as `unknown step or command: "((n"`. The working idioms are function application (`dbl 3 + 1`) and `let r = 3 |> dbl in r + 1`.

*Design.* **Decision D-F8b:** (1) always: a precise error at the point of failure, `a pipeline (|>) cannot appear inside an expression; apply the function directly (f x) or bind it: let r = x |> f in …`, from the infix parser and inside `match` arms (same message, with the position once F1 lands); (2) optionally, support `(x |> f)` as an expression when `f` resolves to a function *value* (a closure or builtin curried function, where `x |> f` is just `f x`), and keep it an error when `f` is a stream step (it yields a stream, not a value). Mandate 1 argues for (2); the ambiguity argues for doing (1) first and (2) as its own sub-step with the decision recorded. Recommended: (1) now, (2) after F1.

**F8c. Possible silent failure inside recursive `match` defs: investigate.** A naive recursive Fibonacci written with the expression-group form,
```
def fib = match { 0 -> 0 | 1 -> 1 | n -> ((n - 1) |> fib) + ((n - 2) |> fib) }
10 |> fib |> assert 55
```
produced `[ERROR] Assertion failed: expected '55', got '0'`: the assert caught it, but the evaluation itself neither failed loudly nor returned the right value. In a smaller case (`n -> ((n - 1) |> dbl) + 1`) the error is the confusing `unknown step or command: "((n"`, so two different arms of the same code path behave differently. *Work:* reproduce with a Go test, find which path turns the arm body into a value or an empty result, make it a loud error (Mandate 6) or correct, add rows to `failLoudRows`. If F8b(2) is implemented, the same test must then produce `55`.

**Order inside F8:** F8a (with F1) → F8b(1) → F8c (it may disappear once F8b(1) rejects the form loudly; then keep the test as the negative control) → F8b(2) as an optional last step.

## 7b. Phase F9: pattern-matching function definitions  — size M–L, found 2026-10-14

Raised by the user's question about Haskell-style definitions ("first line defines the function name and input values, following lines each start with `|` with the pattern the arguments must match, then the body, one variant per line"). What exists today, verified with the built binary:

| Haskell | gatwd today |
| :-- | :-- |
| `fact 0 = 1` / `fact n = n * fact (n-1)` (one clause per line) | not supported: `def fact 0 = 1` and `def fact n = …` are **silently accepted and define nothing**; a later `3 \|> fact` is `unknown step or command: "fact"` |
| `f x \| x > 0 = …` (guards) | not supported: `\| n if n > 0 -> 1` is not a recognised pattern, the arm **never matches** and the next arm wins |
| several arguments `add 0 y = y` | one incoming value only; several arguments are matched as a tuple/list pattern `[0, y]` |
| patterns in lambda parameters | `\x y ->` binds variables only |
| what exists | `def f = match { pat -> body \| pat -> body … }`, one arm per line (multi-line works, brackets are open), applied with `x \|> f`; patterns: literal, variable, `_`, constructor, `[]`, `[a, b]`, `(a, b)`, `[h, ...t]` |

**Already fixed as Track R (merged 2026-10-14, `hardening_plan.md` §6.4, `pkg/cli/matchpattern.go`):** literal patterns matched by *substring* (`0 -> …` matched 10 and 20; a recursive `fib` was wrong from 10 on), element patterns compared printed text, and a value matching no arm vanished silently. They are value equality, a loud error for an unrecognised pattern and a loud `no arm matches` error now. Everything below builds on that.

**F9a. Guards.** `| pattern if condition -> body`, evaluated after the pattern binds its variables; a failing guard falls through to the next arm (as in Haskell); the condition is an ordinary expression (`n > 0`, `x == y`, `str_starts_with "a" s`) and must be a `Bool` (anything else is a loud error naming the arm). `--check`: variables bound by the pattern are in scope in the guard (zero false positives on committed scripts); exhaustiveness treats a guarded arm as not covering its pattern (so `n if n > 0 -> …` alone still warns). **Decision D-F9a:** spelling `pat if cond ->` (recommended: reads like the existing arms and like Rust/OCaml `when`/Scala `if`) versus Haskell's `| cond` after the pattern. Recommended: `if`.

**F9b. Clause-style definitions.** Consecutive `def NAME <patterns> = body` lines with the same name form one function, in order, exactly Haskell's layout:
```
def fact 0 = 1
def fact n = n * (fact (n - 1))
def add 0 y = y
def add x 0 = x
def add x y = x + y
```
*Design.* A desugaring in the statement front end (not the evaluator): the joined clauses become `def fact = \a1 -> match a1 { … }`-equivalent; with N parameters the clauses match the **tuple** of the arguments, so `def add = \a b -> [a, b] |> match { [0, y] -> y | [x, 0] -> x | [x, y] -> x + y }` (the existing tuple patterns, no new matcher). The result is a normal curried closure: `add 3 4`, `map (add 1)`, partial application and the function-form call all work, which the one-input `match`-def (a pipeline macro) cannot do. Clauses of different arity, a gap (another statement between clauses of one name), or a redefinition later are loud errors naming both lines. Optional guards on a clause: `def f x if x > 0 = …` (after F9a). `--check`: the clauses are one function for typing and exhaustiveness (a final catch-all clause is required to silence "non-exhaustive").
*Why not just a Haskell `|` layout:* it combines with F8a (a line starting with `|` continues the previous statement), so both spellings work: the repeated-`def` form above and the one-`def` block `def fact = match { 0 -> 1 \| n -> … }`.
**Decision D-F9b:** repeated `def` clauses (recommended: unambiguous with the current line-oriented statements, no layout rule needed) versus a Haskell-style `fact 0 = 1` without `def` (rejected: a bare `name pattern = expr` is today a malformed statement and would collide with application and equality).

**F9c. Patterns in lambda parameters** (`\0 y -> y`, `\(a, b) -> a + b`): the same tuple desugaring applied to `\` parameters. Optional; do it when a consumer exists (the clause form of F9b covers the use case).

**F9e. A crash found by Track R (fix first, it is a panic):** `[0, 1] |> map (\n -> n |> one)`, where `one` is a `match`-bodied `def`, panics with a nil `Parser` in the lambda's child interpreter (`Interpreter.fork` for a lambda body does not carry the parser that `match` needs for constructor lookup). It reproduces on the pre-Track-R binary too, so it is old. *Work:* a Go regression test (RED), carry the parser into the forked interpreter (check every other `fork`/child construction in `pkg/cli/closures.go` for the same omission), then make any remaining nil-parser path an error value, never a panic. Calling a `match` def from inside a lambda is exactly what clause-style definitions (F9b) will do, so this is a prerequisite for F9b.

**F9d. The silent acceptance of `def fact 0 = 1` is a defect of its own** until F9b lands: make the current form a loud error (`def NAME` must be followed by `=`; a pattern before `=` is not supported yet, see clause-style definitions) as the first, 10-line sub-step, so nothing can be silently defined-as-nothing in the meantime.

*Tests.* A table of clause definitions (factorial, fib, `add` with tuple patterns, a guard, partial application, a catch-all missing → `--check` warning, arity mismatch error, a gap error); step form vs closure parity; negative controls: break the clause grouping and the guard fall-through and watch the tests fail; the committed pure-DSL examples (`examples/game_of_life_pure_dsl`, `quicksort`/`ackermann` in `tests/dsl_exhaustive_test.gatwd`) get a clause-style twin as dogfooding (Mandate 4), asserting the same results. Docs: `docs/dsl_guide.md` (a Haskell-comparison table like the one above), `docs/grammar.md`, the `dsl-authoring` skill.

**Order inside F9:** F9e (the panic) → F9d (loud error) → F9a guards → F9b clause definitions (needs F8a if the clauses use trailing-operator continuation) → F9c if wanted.

## 8. Standing decisions that stay as they are

Remote kustomize bases stay refused; Secret `data` is documented, not redacted; age rules stay an example override; deleting or pruning cluster objects stays out of scope; `--out` does not need the apply capability; the operator never applies offline manifests to a live client; one `--timeout` bounds loading and checks. Reopen any of these only on the user's word.

## 9. Sequencing and parallelism

| Phase | Files it owns | Runs in parallel with | Must wait for |
| :-- | :-- | :-- | :-- |
| F1 error positions | `parser/`, `pkg/cli/cli.go` (message), docs | F4, F5, F6, F7 | nothing |
| F2 interruptible evaluation | `functions/closure.go`, `pkg/cli/policy.go`, evaluator hot path | F4, F5 | nothing (benchmarks once, deliberately) |
| F3 `gatwd` signals | `cmd/gatwd/`, `pkg/cli/repl.go`, `run.go` | F4, F5, F6 | F2 for the runaway-closure case only |
| F4 rollout | `adapters/kubernetes_rollout.go`, `pkg/k8sgraph/write.go` | F1, F2, F3, F5, F6 | nothing |
| F5 rule messages | `pkg/cli/appliance_report.go`, `lib/k8s` rules, how-to | F1, F2, F3, F4, F6 | decision D-F5 |
| F6 small items | `functions/`, `parser` lexer, `dslfuncs_*`, `typecheck.go` | F4, F5, F7 | F1 if it renumbers the lexer positions (exponent literals) |
| F8 multi-line statements, expression-group diagnostics | `pkg/cli/cli.go` (`logicalLines`), `parser/infix.go`, `pkg/cli/typecheck.go` (`--check` joiner use), embedded rules and prelude (reformat), docs | F2, F3, F4, F5, F7 | **F1** (same joiner; do F8a with or right after F1) |
| F9 pattern-matching definitions (guards, clause-style `def`) | `pkg/cli/cli.go` (match step, `def` front end), `parser/` (`def` and arm grammar), `pkg/cli/typecheck.go`, docs | F2, F3, F4, F5, F7 | Track R (merged first), F8a for line-continuation inside clauses |
| F7 operator config | `pkg/cli/appliance_command.go`, `cmd/gatwd-k8s-operator` | F1, F2, F4, F6 | nothing |

**Suggested order** (at most two heavy agents at a time): **F1 (+ F8a, one agent) and F4** first (disjoint, highest value), then **F2 and F5**, then **F3 and F7**, with F8b/F8c, **F9 (it follows Track R and F8a; F9d first)** and the F6 items interleaved on demand. F1 before the exponent-literal item of F6 because both touch the lexer.

## 10. Definition of done for the whole plan

Each phase: a design entry here updated with the decision and date; commits per sub-step; the verification gates green; deferrals written down; a handoff entry in `docs/hardening_handoff.md` (or a new handoff file if that one is retired); and, for anything user-visible in the Kubernetes tool, a row in `hardening_plan.md` §10.1 and a pinned how-to example in `cmd/gatwd-k8s-operator/doc_test.go`.
