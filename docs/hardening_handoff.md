# Handoff — continue the hardening plan at W1b (the cobra command tree)

> Handoff for a **fresh context**. It **supersedes the earlier version of this file** (the W0 brief, now done) and
> nothing else; the k8s-graph handoffs `phase6/7/8_handoff.md` are historical and carry banners.
>
> **Read in this order:** `AGENTS.md` (mandates 1–8, §5 gates, §5a prevention rules) → `docs/hardening_plan.md`
> (§0 progress table, §1 decisions, §4 W1) → `docs/development_playbook.md` (§3 gates and the regression check, §6
> DSL traps) → `docs/dsl_guide.md` §23 (the argument model you are about to put a command line on) → this file.
> Then start at "First steps" (§9).

---

## 1. Where we are

**Committed:** the language roadmap, the Kubernetes graph tool, the agent rules/skills/guards, the how-to, the hardening
*plan*, and now **W0 and the D4 slice of W1a**, as two commits — the code, tests and examples together (`cli.go` carries both
the fail-loud changes and the argument model, so they could not be split without a broken intermediate tree), then the docs.
**Pushed:** none of the latest commits — `git status -sb` shows the branch ahead of `origin`. Ask before pushing.
**Uncommitted:** nothing of ours; `output.txt` in the repo root is untracked and **not ours** — do not add it.

What those commits contain (by area, for orientation):

| Group | Files |
| :-- | :-- |
| Operator and type errors | `parser/infix.go` (`ApplyError` now also wraps operator failures; new `UndefinedError`), `parser/parser.go` (godoc) |
| `assert` on an empty stream | `evaluator/pipeline.go`, `evaluator/assert_test.go` |
| One lambda implementation | `pkg/cli/closures.go` (`stepClosure`, `failStep`, `pipelineLambda`, `withPipelineFallback`, `flatmapStepOver`, graph `latchedStep`), `pkg/cli/valueops.go` (`notCarrierError`), `pkg/cli/cli.go` (`mapStepOver`/`filterStepOver` now sit on the shared `mapValue`/`filterValue`; the old text parsers `compileUnary`/`compilePred`/`compileBinary` and the inline fold parser are gone) |
| `let`, heads, constants | `pkg/cli/cli.go` (`textSubstitutable`, `scalarValue`, `letResultError`, `startsWithLiteral`, `explainSilentStatement`, bare-constant heads) |
| `\|\|>` and exec steps | `evaluator/pipeline.go` (`StepPar`, `ComposeSteps`), `pkg/cli/cli.go` (`splitPipelineOps`, `pipelineStepsFor`), `interactive/exec_adapter.go` + `exec_adapter_test.go` (`SplitCommandLine`, loud failures) |
| D4 arguments and tasks | `pkg/cli/run.go`, `pkg/cli/run_test.go`, `examples/k8s_howto/tasks.gatwd`, `cmd/gatwd-k8s-operator/howto_test.go`, `examples/scripts/cheatsheet.gatwd`, `pkg/cli/cheatsheet_test.go` |
| Tests and gate | `pkg/cli/failloud_test.go`, `tests/dsl_exhaustive_test.gatwd` (the last sections), `run_dsl_tests.sh` (comment) |
| Docs | `AGENTS.md`, `CLAUDE.md`, `STYLEGUIDE.md`, `.agents/skills/dsl-authoring/SKILL.md`, `docs/{development_playbook,dsl_guide,k8s_howto,hardening_plan}.md`, this file |

**Capabilities now true, and the tests that prove them**

- No evaluation path returns nothing, `nil`, `0` or its input on failure: `TestFailLoudSweep` (a table, one row per
  defect F1–F8/F10), `TestLetBoundValueIsNotTextSubstituted`, `TestLiteralHeadIsAnExpression`, `TestBareConstantStatement`,
  `TestStepAssertEmptyStreamIsAnError`, and the DSL sections "Fail-loud sweep", "ARGS / ARG0" and "bare literal" in
  `tests/dsl_exhaustive_test.gatwd`.
- `map`/`filter`/`reject`/`flatmap`/`fold` and the graph lambda steps go through one closure path that reports errors
  naming the element: the `TestFailLoudSweep` graph rows and the `evaluator` structural tests.
- `||>` fans the following per-element step out on a worker pool; exec steps tokenize quotes like a shell and report
  failed starts and non-zero exits: `TestParallelPipeYieldsResults`, `TestExecStepHonoursQuotes`, `TestSplitCommandLine`,
  `TestExecAdapterNonZeroExitIsAnError`.
- **D4 (decided):** `ARGS` (List of String after `--`), `ARG0`, `main` as the default task, `--task NAME`. A lambda-bodied
  task is called with the `ARGS` list; a pipeline-bodied task is a *fragment* whose input stream is the arguments, one
  String node each. `TestParseRunArgs`, `TestRun*`, `TestArgsAreNotSplicedAsText`, `TestHowToTasks`,
  `TestHowToTasksCommandLine` (builds the real binary), `TestCheatsheet*`.
- `examples/scripts/cheatsheet.gatwd` is a runnable, deterministic, self-asserting script and a test (`TestCheatsheet*`).

## 2. Decisions

**Decided by the user (2026-10-13)** — implement, do not re-ask:

| ID | Decision |
| :-- | :-- |
| D2 | `+` is numeric-only; `++` concatenates. *Proposed corollary, not yet confirmed:* `"a" + "b"` is also an error with a hint to use `++` — **mention it when you reach W2 and confirm.** |
| D4 | `ARGS`, `ARG0`, `main`, `--task` — **implemented** (see §1). Cobra owns the flag *syntax*; D4 is what a script *sees*. |
| D5 | Appliance override layers on the embedded rules by default; `replace` is an explicit mode (W6). |
| D6 | Qualified node ids `Kind/namespace/name` + loud duplicate detection; migrate all tests and examples in the same phase (W4). |
| D9 | `assert` on an empty stream is an error — **implemented**. |
| D10 | Cobra (+pflag) directly in `pkg/cli`; the Clean Core allowlist is already in place. No cobra code exists yet. |

**Still open — ask before the phase that needs it** (recommendations are in plan §1):
D1 (`Char`; String as a collection row) → W2 · D3 (numeric tower spec; `Num`/`Integral`/`Fractional` constraint names) → W3 ·
D7 (kustomize pin v0.21.0 + auto-render) → W5 · D8 (exec capability policy) → W7.

**W1b needs none of them.** One design question *inside* W1 is yours to raise with the user first (§3, item 2).

## 3. Goals for the next phase (W1b, then the rest of W1a)

1. **The cobra command tree in `pkg/cli`** (plan §4.3). Replace the hand-rolled `parseRunArgs` and the `--check` branch in
   `RunWithInterpreter` with the tree: root = backward-compatible script runner / REPL (`Args: cobra.ArbitraryArgs`, script
   arguments after `--` via `cmd.ArgsLenAtDash()`), plus `run`, `eval -e`, `check`, `task`, `repl`, `prelude list|cat`,
   `steps`, `version`, `completion`, and `Edition.WithCommand(*cobra.Command)`. Keep the tree in its own files
   (`command*.go`). It is a thin mapping from flags to `Interpreter.Run`; **semantics stay in `Run`, not in cobra code.**
   Backward compatibility is a tested table: `script`, `-i script`, `--check file`, `script -- args`, `--task`.
   Known cobra edge: pflag rejects single-dash long flags (`-manifests`) — the operator migration to `--manifests` etc. with a
   one-release shim is W6, but design the normaliser hook now.
2. **Open design question — where does output go?** `Run` returns only the *task's* lines; top-level statements still print
   through `fmt.Printf` inside the evaluator. The W6/W7 API-server and ConfigMap hook need output as values or writers.
   Options, in order of preference: **(a)** an `Out io.Writer` on the interpreter/`RunOptions` (default `os.Stdout`) used by
   `EvaluateLine`/`LoadScript*` instead of `fmt.Printf` — small and testable (it would replace the `captureStdout` hack in
   `pkg/cli/cheatsheet_test.go`); (b) a returned `[]Result` for every statement (a bigger API change); (c) leave printing and defer to W7.
   Pick with the user before building.
3. **Finish `RunOptions`:** `Eval []string` (`-e`), `Define` (`-D NAME=VALUE`), `Preludes`, `NoPrelude`, `Check`, `Interactive`,
   `Stdin/Stdout/Stderr`, `context.Context`. A generic `env "NAME"` builtin returning `Maybe String` is planned but **subject to the
   W7 policy** — add it only together with its policy hook, or leave it for W7.
4. **Carry-over from W0 (decide, don't drop):** unparsable statements say *why* but not *where*; wire `parser/errors.go`
   `FormatError` into the infix/line parsers, or strike the item from plan §3 with a written reason.
5. After W1: **W2 and W3** (parallel; need D1 and D3 answered) → **W4** → **W5** (D7) → **W6** → **W7** (D8); details in the plan.
   Each phase starts from the matching prompt in `.agents/prompts/` and the `phase-delivery` skill; one commit per sub-step.

## 4. Authoring shape (copy these)

*A fail-loud row* — add to `failLoudRows` in `pkg/cli/failloud_test.go`: `{"id", statement, "substring of the [ERROR]"}`.
It must be seen **red** before the fix.

*A test of `Run`* — `pkg/cli/run_test.go`: `writeScript(t, "def main = \\args -> …")`, then
`NewInterpreter().Run(RunOptions{Script: p, Args: …, Task: …})`; assert `res.Task` and `res.Lines`.
For a printed top-level statement use `captureStdout` (`pkg/cli/cheatsheet_test.go`).

*A task script* (`examples/k8s_howto/tasks.gatwd`): fragments for pipelines (`def trace_to_pod = graph |> … |> collect_list`),
lambda-bodied defs for values, **no invocations at top level**, `def main = (a <+> b <+> c)`.
Run it: `gatwd-k8s tasks.gatwd -- <manifest-dir>` (main) or `… --task trace_to_pod -- <manifest-dir>`.

*A command-tree test* (for W1b): build the root command, `cmd.SetArgs([]string{…})`, `cmd.SetOut(&buf)`, table-driven; plus one
end-to-end test that builds the binary (pattern: `TestHowToTasksCommandLine`).

*A negative control* (every new guard): copy the file to `/tmp`, break the guard (for example `sed` the condition off), run the test,
watch it fail, restore the file.

## 5. Gotchas (don't re-derive)

**Language and runtime**
- A bare literal head (`5 |> f`, `"x" |> f`) and a bare constant (`N`) used to print nothing; both work now. A *non-scalar*
  constant or let-bound value is never spliced as text (`scalarValue`, `textSubstitutable`) — `ARGS` is a list value, use it in
  expressions (`length ARGS`), not inside step text.
- `ARGS`/`ARG0` live in `Constants`, **not** `RuntimeNodes`: a non-empty `RuntimeNodes` disables the def-parse cache for every pipeline.
- The spawn node is the first element of the stream: `spawn T: 0 |> emit 10, 20 |> drop 1` leaves `10, 20`.
- `||>` applies to the *following per-element* step, order is nondeterministic, and a whole-stream step (`fold`, `sort`) after it sees
  singletons. `length` has no step form (F18); use the function form `length x`.
- A fragment run as a task with **no arguments** gets an empty stream (a generator like `emit` still runs; `manifests` yields `[]`).
- `main` and `--task` run *after* the script's top-level statements, which always run — keep invocations out of the top level.
- Exec steps now split quotes like a shell (no expansion) and a **non-zero exit is an error** (so `grep` with no match is an
  `[ERROR]`). An unknown pipeline step still **runs an OS command** (F17) until W7: never feed untrusted scripts to the interpreter.
- A lambda whose body is really a pipeline (`\x -> echo Processing %s`) works only through `withPipelineFallback`, which triggers
  when the body's *head word* is an undefined name; any other failure is returned.
- `"a" + [1]` and `"a" + 1` still concatenate until D2 lands (W2); the graph fail-loud rows use an undefined name for that reason.
- Known, **not fixed**: `zip` is suspect (`echo Primary: %s |> zip date` yields only the date; `zip echo …` prints nothing); big
  integers print with a leading `+` (F11, W3). The cheatsheet deliberately avoids both.
- Exit codes today: a `Run` error exits `1`, a malformed command line exits `2`; the operator still exits `0` on violations (W6).

**Tests and tooling**
- The F2 rows in `failloud_test.go` (`'a'`, `**`, `!true`, `not true`, `- 5`) expect `cannot evaluate`; **W2/W3 must flip them to
  values** when `Char`, `not`, `**` and unary minus are real.
- `EvaluateToValues` treats its argument as one pipeline; `let … in …` goes through `EvaluateLine`/`EvaluateToValues`, not
  `evaluatePipelineStr`. Results are `*collections.List[any]` (use `functions.ToSliceAny`).
- `gofmt -l` lists files that were **already unformatted** and are not part of this work (`parser/*`, `evaluator/graph_steps.go`,
  `pkg/plugins/registry.go`, `cmd/gatwd-gen/main.go`). Format only what you touch; do not mass-format.
- Scratch scripts: `cat > f <<'EOF'` (zsh mangles `\x` in `echo`). `run_dsl_tests.sh` builds `./gatwd` in the repo root.
- Run the baseline binary **more than once** before blaming a diff: `examples/scripts/03_kleisli_fanout`, `examples/tour/05`, `15`, `20`
  and `23` differ run to run on the old code (map-iteration order).
- Intended script-diff differences vs the pre-hardening binary: `examples/k8s_howto/tasks.gatwd` (now needs arguments and the k8s
  edition), `examples/new_features/showcase.gatwd` (`echo "…"` no longer prints the quotes), and the rewritten cheatsheet.

## 6. Constraints (hard rules)

- Mandates 1–8 (`AGENTS.md`). In particular **Mandate 5**: core imports only stdlib + this module, except cobra/pflag in `pkg/cli`
  only — nothing else may import cobra, and the lower layers stay stdlib-only; **Mandate 6**: nothing silent, negative-control every
  guard; **Mandate 7**: Editions add commands (`WithCommand`), not logic; **Mandate 8**: ask at hard-to-reverse choices and record
  them with a date in the plan, write deferrals down, shared tree — re-read before editing, stage by path, commit/push only when asked.
- The command tree must not hold semantics: it maps flags to `Interpreter.Run`.

## 7. Verification gates

```bash
gofmt -l <dirs you touched>         # nothing listed for files you touched
go vet ./...
go test ./...                       # includes internal/archtest (layers, Clean Core, doc links)
./run_dsl_tests.sh                  # fails on [ERROR], silent statements, --check errors
go test -race ./pkg/... ./evaluator ./parser ./interactive ./cmd/gatwd-k8s-operator
```

Interpreter, parser and evaluator changes also need the **script regression diff** and a silent `--check`. The helper script is
ephemeral, so recreate it:

```bash
git worktree add /tmp/base <previous-commit> && (cd /tmp/base && go build -o /tmp/gatwd_base ./cmd/gatwd)
go build -o /tmp/gatwd_new ./cmd/gatwd
# for each of base, new:
for f in $(find . -name '*.gatwd' -not -path './tests/*' -not -path './legacy/*' -not -path './.git/*' | sort); do
  timeout 20 <binary> $f </dev/null > <outdir>/$(echo $f | tr / _).txt 2>&1; done
diff -rq <base-out> <new-out>       # run the baseline twice first to learn which scripts are nondeterministic
for f in $(find . -name '*.gatwd' -not -path './tests/*' -not -path './legacy/*'); do /tmp/gatwd_new --check $f; done   # must print nothing
```

For W1b the "previous commit" is the current `HEAD`: build the baseline from it **before** editing.

## 8. Pointers

- Plan and decisions: `docs/hardening_plan.md` (§0 progress, §1 decisions, §2.4 defect table F1–F18, §4 W1, §10 sequencing).
- The argument model: `docs/dsl_guide.md` §23 and `pkg/cli/run.go`; the entry point is `RunWithInterpreter` in `pkg/cli/cli.go`.
- Guards: `internal/archtest/{layers,clean_core,doclinks}_test.go` (the cobra allowlist is `coreImportAllowlist`), `run_dsl_tests.sh`.
- Prompts and skills: `.agents/prompts/add-language-feature.md`, `.agents/prompts/review-against-mandates.md`,
  `.agents/prompts/phase-handoff.md`, `.agents/skills/{all-layers-sync,dsl-authoring,phase-delivery,edition-appliance}`.
- Memory notes for the assistant: `project-hardening-plan-decisions`, `project-agent-docs-and-guards`.

## 9. First steps

1. `git status -sb` and `git log --oneline -5`: confirm the two hardening commits are the tip and the tree is clean apart from
   the untracked `output.txt`.
2. Run the §7 gates once so you start from a known-green tree; build the baseline binary from the current `HEAD`.
3. Ask the user the output question (§3, item 2) and record the answer with its date in plan §4.
4. Add the cobra dependency (`go get github.com/spf13/cobra`; pflag comes with it), confirm `internal/archtest` still passes, and
   write the **backward-compatibility table test first** (every invocation `parseRunArgs` accepts today, expected to behave
   identically on the new tree) — red against an empty tree, green when the root command exists.
5. Build the tree in `pkg/cli/command*.go` on top of `Interpreter.Run`; delete `parseRunArgs` and the hand-rolled `--check` branch in
   the same commit that makes the table green. Add `eval -e`, `check`, `task`, `repl`, `steps`, `version`, `completion`, then
   `prelude list|cat` and `Edition.WithCommand`.
6. Update `docs/dsl_guide.md` §23, `docs/k8s_howto.md` (flags, the `-i` route) and the playbook; mark W1b in the plan's progress
   table; rewrite this handoff for W2/W3 at the end of the phase and mark it superseded.
