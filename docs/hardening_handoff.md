# Handoff — start implementing the hardening plan (W0 first)

> Handoff for a **fresh context**. Nothing in `docs/hardening_plan.md` is implemented yet; this document says what is
> decided, what is still open, and exactly how to begin. It supersedes no earlier handoff (the k8s-graph handoffs
> `phase6/7/8_handoff.md` are historical and carry banners).
>
> **Read in this order:** `AGENTS.md` (mandates 1–8, §5 gates, §5a prevention rules) → `docs/hardening_plan.md`
> (the plan; §1 decisions, §2.4 the verified defect table F1–F18, §3 W0) → `docs/development_playbook.md` (§3 gates,
> §6 DSL traps) → this file. Then start at "First steps".

---

## 1. Where we are

- **Done and committed:** the language roadmap (closures, currying, gradual typing, function-form matrix, `<+>`),
  the Kubernetes graph tool through Phase 8, the agent rules/skills/prompts, machine-enforced guards, the how-to
  (`docs/k8s_howto.md` + `examples/k8s_howto/`, executed by `cmd/gatwd-k8s-operator/howto_test.go`), and the
  hardening **plan**. The Clean Core test already allows **cobra/pflag in `pkg/cli` only**
  (`coreImportAllowlist` in `internal/archtest/clean_core_test.go`); no cobra code exists yet.
- **Not pushed.** The latest commits are local only; ask before pushing (`git status -sb`).
- **Shared tree:** another session may touch `docs/architecture.md`, `docs/language_roadmap.md`,
  `internal/archtest/layers_test.go` (library-extraction work). Re-read before editing, keep edits additive,
  stage by path.

## 2. Decisions

**Decided by the user (2026-10-13)** — implement, do not re-ask:

| ID | Decision |
| :-- | :-- |
| D2 | `+` is numeric-only. `"x" + 1` is an error; `++` concatenates; numeric `x + 1` is unaffected. *Proposed corollary, not yet confirmed:* `"a" + "b"` also an error with a hint to use `++` — **mention it when you reach W2 and confirm.** |
| D5 | Appliance override **layers on** the embedded rules by default; `replace` is an explicit mode. |
| D6 | Qualified node ids `Kind/namespace/name` + loud duplicate detection; **migrate all tests and examples in the same phase.** |
| D9 | `assert` on an empty stream is an error. |
| D10 | Cobra (+pflag) integrated directly into `pkg/cli`; allowlist already in place. |

**Still open — ask the user before the phase that needs it** (recommendations are in plan §1):
D1 (Char; String as a collection row), D3 (numeric tower spec + `Num`/`Integral`/`Fractional` constraint names),
D4 (`ARGS` / `main` / `--task` shape), D7 (kustomize pin v0.21.0 + auto-render), D8 (exec capability policy).

**W0 needs none of the open decisions.** Start there.

## 3. First steps — W0, the fail-loud sweep (plan §3)

Do these as separate commits, each with the gates (§6) green. Write the **pinning test first** for every row.

1. **Make a regression table of the silent failures.** A Go test (new file in `pkg/cli`, e.g. `failloud_test.go`) with
   one row per defect: statement → expected `[ERROR]` substring. Start red. Rows (all reproduced; each printed
   nothing, `[<nil> <nil>]`, `[]` or the fold seed before):
   - `1 / 0`, `1 // 0`, `1 % 0` → error text contains `by zero`
   - `true + 1` → `unsupported binary operator`
   - `[1, 2] |> map (\x -> nosuchfn x)` → names `nosuchfn` (today `[<nil> <nil>]`)
   - `[1, 2] |> map (\x -> x + [1])` (today `[<nil> <nil>]`)
   - `[1, 2] |> filter (\x -> nosuchfn x)` (today `[]`)
   - `[1, 2] |> fold (\a x -> a + nosuch) 0` (today `0`)
   - `let r = <hamt> in hamt_get "kind" r` as a top-level statement (today nothing; F6 — time-boxed spike, see below)
   - `spawn IntVal: 1 ||> math_add 1` (F7)
   - `assert` on an empty stream (F8 / D9)
   - `'a'`, `**`, `!true`, `not true`, `- 5`: must report *why* (a parse/evaluation error), not nothing. (Adding `Char`,
     `not`, `**` themselves is W2/W3.)
2. **Wire operator errors through `buildInitialStream`.** The operator layer already returns the errors
   (`functions.EvalBinaryOp` → `division by zero`, `unsupported binary operator …`; verified). The silence is in
   `Interpreter.buildInitialStream`: when `parser.EvalInfixString` fails it falls through to `Parser.Parse` and yields
   nothing, except for `*parser.ApplyError`. Add an analogous typed error for *evaluation* failures (an operator
   failing, as opposed to `undefined identifier`, which must keep falling through because a first part may be a pipeline
   word) and surface it as an error node. Keep "(No output or node unhandled)" for statements that legitimately yield
   no value (`type`, `def`, `const`).
3. **One lambda implementation (the big one).** The inline `map`, `filter`/`reject`, `flatmap` and custom `fold` steps in
   `parsePipelineSteps` (`pkg/cli/cli.go`) and the graph lambdas (`compileUnary`/`compilePred`/`compileBinary`, which
   return the input on error; `parseLambdaInner`) each hand-parse `\x -> …` text and swallow errors. Replace them with:
   parse once via `parser.LambdaExpr` → `*functions.Closure`, then run through `mapStepOver`/`filterStepOver` and the
   closure `foldClosureStep`/`closureMapper`/`closurePredicate` in `pkg/cli/closures.go`, which already propagate errors.
   Make `mapStepOver` yield an error node for a collection when an element fails (it currently ignores per-element
   errors). Delete the dead text parsers. This is Mandate 4 dogfooding and closes F3–F5 at the root.
4. **`assert` on empty input** → error (`evaluator.StepAssert`, `evaluator/pipeline.go`). Run the DSL suite; fix scripts
   that relied on vacuous asserts (the gate already flags silent statements, so few should).
5. **Top-level `let` binding a real value (F6).** The `let` path in `EvaluateLine`/`EvaluateToValues` stores a stringified
   copy in `Variables` beside the real node in `RuntimeNodes`, and later substitution clobbers the real value. **Spike
   first, time-boxed (~half a day):** bind only the node, fall back to text only for non-representable values, and
   measure with the script regression diff. If it spreads, stop and report before continuing.
6. **`||>` (F7) and quoted args to OS-exec steps (F10).** Isolate the cause before deciding fix-or-error. (One probe,
   `sh -c "echo …$((6*7))"`, printed nothing and the `$(( ))` may be the real cause.)
7. Fix the stale `Parser.Parse` godoc claim that an unregistered `spawn` type yields no node (F15; the interpreter
   registers types on demand).
8. **Docs:** update `docs/k8s_howto.md` §2.1/§4 (the inline-`fold` limitation disappears), the playbook trap table, and
   AGENTS §5a if a row becomes obsolete; mark the F-rows done in `docs/hardening_plan.md`.

**Definition of done for W0:** F1–F8, F10, F15 closed with regression tests; script regression diff clean except the
intended new errors; `gatwd --check` silent over all scripts; race-clean.

## 4. After W0 (summary — details in the plan)

W1 (`Run` API, then the cobra tree in `pkg/cli`; operator flags become `--long` with a one-release shim) →
W2/W3 (base types, numeric tower; needs D1/D3 answers) → W4 (qualified ids) → W5 (in-process kustomize; D7) →
W6 (operator layering/exit codes/formats) → W7 (exec policy; D8). Each phase starts from the matching prompt in
`.agents/prompts/` and the `phase-delivery` skill.

## 5. Gotchas (still true)

- `assert` passes on an empty stream and the old inline lambdas swallow errors — so **a green test proves nothing until it
  has been seen to fail.** Negative-control every new guard once.
- Calls inside plain list literals need parentheses: `[(f x), (g y)]`.
- Add graph nodes before edges; guard with `graph_has_node`; `hamt_has_path` guards go before the dependent generator.
- `EvaluateToValues` treats its argument as one pipeline; use `EvaluateScript` for `def`/`type` setup. Results are
  `*collections.List[any]` (use `functions.ToSliceAny`).
- Scratch scripts: `cat > f <<'EOF'`, never `echo` (zsh mangles `\x`).
- `||>` printing nothing at top level is pre-existing; `examples/tour/15`, `20`, `23` differ run to run on `master` alone
  (map-iteration order) — run the baseline more than once before blaming a change.
- An unknown pipeline step **runs an OS command** (`|> printf …` executes). Do not feed untrusted scripts to the
  interpreter until W7.
- A `-rules` operator override currently **replaces** the embedded rules (W6 changes the default to layering).

## 6. Hard constraints and gates

- Mandates 1–8 (`AGENTS.md`). In particular **Mandate 5**: core imports only stdlib + this module, except cobra/pflag in
  `pkg/cli`; **Mandate 6**: nothing silent; **Mandate 8**: stop and ask at open decisions, one commit per sub-step.
- Run before every commit:

```bash
gofmt -l <changed dirs>          # nothing listed for files you touched
go vet ./...
go test ./...                    # includes internal/archtest (layers, Clean Core, doc links)
./run_dsl_tests.sh               # fails on [ERROR], silent statements, --check errors
go test -race ./pkg/... ./evaluator ./parser
```

- Interpreter/parser/evaluator changes also need the **script regression diff**: build a baseline binary from a
  `git worktree` of the previous commit, run every `*.gatwd` outside `tests/` and `legacy/` with both, and diff
  (`docs/development_playbook.md` §3). Expect differences only where a silent failure became an error.
- Do not commit or push unless asked; stage by explicit path.

## 7. Pointers

- Plan and decisions: `docs/hardening_plan.md`; defect table F1–F18 in §2.4.
- Code: `pkg/cli/cli.go` (`parsePipelineSteps`, `buildInitialStream`, `EvaluateLine`, `EvaluateToValues`,
  `mapStepOver`, `filterStepOver`, `compileUnary`/`compilePred`/`compileBinary`), `pkg/cli/closures.go`,
  `parser/infix.go` (`LambdaExpr`, `ApplyError`), `functions/operators.go` (`EvalBinaryOp`),
  `evaluator/pipeline.go` (`StepAssert`).
- Guards: `internal/archtest/{layers,clean_core,doclinks}_test.go`; `run_dsl_tests.sh`.
- Prompts/skills: `.agents/prompts/add-language-feature.md`, `.agents/prompts/review-against-mandates.md`,
  `.agents/skills/{all-layers-sync,dsl-authoring,phase-delivery}`.
- Memory notes for the assistant: `project-hardening-plan-decisions`, `project-agent-docs-and-guards`.
