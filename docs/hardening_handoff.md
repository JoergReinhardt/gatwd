# Handoff — hardening plan: W0–W7, YAML, the time types and the closing checklist are done and merged; what is left are small follow-ups

> Handoff for a **fresh context**. It **supersedes the earlier version of this file** (the "tracks E–G in flight" brief) and
> nothing else; the k8s-graph handoffs `phase6/7/8_handoff.md` are historical and carry banners.
>
> **First, do §0** (put the open questions to the user). **Then read in this order:** `AGENTS.md` (mandates 1–8, §5 gates, §5a prevention rules) → `docs/hardening_plan.md`
> (§0 progress table, §1 decisions, the closure sections §5.6/§6.2/§6.3/§7.1/§8.1/§9, §10.1 closing table) → `docs/development_playbook.md` (§3 gates and the
> regression check, §4 step vs function form, §6 DSL traps) → this file.

---

## 0. START HERE — put the open questions to the user first

All were raised by the tracks and the closing checklist (2026-10-13); none blocks anything. Present each with its default so the user can answer in one line, record the answers
(with the date) in `docs/hardening_plan.md` §1 and in §2 here. If the user says "defaults", use the recommended column and say so in the plan.

**The Q1–Q11 of the previous handoff were answered on 2026-10-13** (plan §1, "Answered after the closing checklist") and implemented: pushed; typed time tokens; value-typed sets; operator `--timeout`/`--max-output`;
`Duration` stays literal text in JSON/YAML; override imports stay behind `--allow-fs`; no `env` builtin yet; age rules stay an example override; Secret `data` is documented, not redacted. **Open now:**

| # | Open question | Recommended default |
| :-- | :-- | :-- |
| Q1 | **Apply to the cluster** (the tool is read-only today; the shipped RBAC grants `get`/`list`): a write adapter in `adapters/` behind `cli.AdapterStep`, a new `AllowCluster` capability in `cli.Policy` (denied by default), `--dry-run` as the default, server-side apply with a field manager, an offline twin that writes rendered YAML to a directory, a separate opt-in RBAC role; first version apply only (no delete/prune), in `gatwd-k8s` before the operator. | yes, as proposed |
| Q2 | Push the commits since `66a2abe` (the merges of Tracks I, J, K and the doc commits)? | push |
| Q3 | Signal handling in the operator `run` loop, bounding rule load/hot reload (Track K's open items); the quoted `const Q = "2+3"` evaluation (Track I); `BST[string]` still stringifies (Track J). | queue as small tracks |

## 1. Where we are

**State (2026-10-13):** `master` contains everything; the remote is at `66a2abe`. `output.txt` in the repo root is untracked and **not ours**. `.claude/worktrees/` is empty and all `hardening/*` branches are deleted.
Stage by name; commit and push only when asked.

**Done and merged** (plan §0 has a row each): W0, W1a/b, W2 (+ YAML, `go.yaml.in/yaml/v3` on the Clean Core allowlist for `pkg/cli`), W3 (numeric tower, `Ratio`, prefix `!`), W4 (qualified ids), W5 (kustomize, Go 1.27),
W6 (layered overrides, exit codes, formats, lazy client), the follow-ups (exit status **3** for violations, operator `--no-kustomize`, source provenance), client-go/apimachinery v0.38 alpha (the `replace` is gone),
canonical HAMT keys and `String != printed collection`, §5.6 `Duration`/`Time`/`Date`, W7 (`cli.Policy`, capability gates, `--allow-*`, `diagnostics`), and the §10.1 closing checklist (the how-to is pinned by
`cmd/gatwd-k8s-operator/doc_test.go`). Gates on the final tree: `gofmt`, `go vet`, `go test ./...`, `./run_dsl_tests.sh`, `-race`, silent `--check`; script regression diff against `2526cc0` explained.

**Deferred with reasons (all in the plan):** `Unit`; `parser.FormatError` wiring (the infix and line parsers do not use the tracker); operator timeout flags (Q8); a guard-only comprehension (`[ e | guard ]` with no generator silently yields `[]`,
documented as a trap, a parser fix is open); `GATWD_<FLAG>` environment defaults; ConfigMap watching; remote kustomize bases (stay refused, user decision); `zip` is suspect (`echo Primary: %s |> zip date`); `json_encode` of a `Ratio`;
exponent float literals (`1e-3`); the `--check` advisory beyond literals.

## 2. Decisions — all answered (2026-10-13)

Two rounds of answers are in `docs/hardening_plan.md` §1 (D1–D18, "Answered at the start of the parallel run", "Answered after the parallel run", the W7 design points). **§0 above is new and open.**

## 3. Working mode

**Implement in parallel wherever possible (user instruction, 2026-10-13).** Decide by file overlap: independent phases run as concurrent subagents, one `isolation: "worktree"` branch each, started in one message, self-contained briefs,
"never push, never edit the progress table or this handoff"; sub-steps inside a phase stay sequential; the main session merges one branch at a time (the track touching the most files last), re-runs every gate and the
script regression diff after each merge, resolves the additive conflicts (keep both sides, renumber `tests/dsl_exhaustive_test.gatwd` sections, migrate expectations a later track changed) and records the closure.
Lessons of this run: a branch cut before another track merged conflicts in the shared command-line files — brief tracks with the latest `master` hash; a resumed agent (SendMessage) keeps its context after a spend-limit stop;
`git show ":${n}:path"` needs braces in zsh. Start from `.agents/prompts/*` and the `phase-delivery` skill; Mandate 2's definition of done applies to every cell.

## 4. Authoring shapes (copy these)

*A fail-loud row* — add to `failLoudRows` in `pkg/cli/failloud_test.go`: `{"id", statement, "substring of the [ERROR]"}`. It must be seen **red** before the fix.

*A test of `Run`* — `pkg/cli/run_test.go`: `writeScript(t, "def main = \\args -> …")`, then `NewInterpreter().Run(RunOptions{Script: p, Args: …, Task: …})`; assert `res.Task` and `res.Lines`.
To see what a statement prints, set `in.Out = &buf` (see `quietInterpreter` in `pkg/cli/cheatsheet_test.go`) — never swap `os.Stdout`.

*A command-line test* — `treeRun(t, stdin, argv...)` in `pkg/cli/command_test.go` returns the exit code, stdout and stderr; add a row to the backward-compat or subcommand table.
A change to a binary's command line also gets a row in that binary's real-binary test.

*An Edition command* — `ed.WithCommand(&cobra.Command{…})`: a thin mapping from flags to a prelude `def`; it gets its interpreter from `ed.BuildInterpreter()`.

*A negative control* (every new guard): copy the file to `/tmp`, break the guard, run the test, watch it fail, restore the file.

## 5. Gotchas (don't re-derive)

**Language and runtime**
- A bare literal head (`5 |> f`) and a bare constant (`N`) work; a *non-scalar* constant or let-bound value is never spliced as text — `ARGS` is a list value, use it in expressions.
- `||>` applies to the following per-element step, order is nondeterministic, a whole-stream step (`fold`, `sort`) after it sees singletons. Every builtin is also a step now (F18, applied to the whole value); inside a pipeline-bodied lambda use the function form.
- A bare name used as an expression looks up a *value*: `eval -e "k8s_graph"` is `undefined identifier` because `k8s_graph` is a pipeline-bodied def (a fragment). Use it in a pipeline: `spawn StrVal: DIR |> manifests |> k8s_graph |> graph_order`.
- `main` and `--task` run after the script's top-level statements, which always run — keep invocations out of the top level. With `-e`, `main` does not run.
- Exec steps split quotes like a shell and a non-zero exit is an error. An unknown pipeline step **runs an OS command** (F17) only under a policy that allows exec: the command-line script and `-e` are trusted, while Appliance overrides and `RunOptions.Source` scripts default to `DenyAll` (W7; `--allow-*` on the operator).
- `+` is numeric-only (D2, done): `"a" + 1` is an error naming both types and `++`. Text reaching arithmetic must be converted (`str_to_int`); the tokens of `emit`/`const`/`-D` are typed literals; spawn an **undeclared** type and the value is text (declare `type IntVal = Int`). a numeric `String` is no longer a number anywhere (W3: arithmetic, comparison, `==` and the number helpers reject it).
- Known, **not fixed**: `zip` is suspect (`echo Primary: %s |> zip date` yields only the date). F11 (a leading `+` on big integers) is fixed.
- `-D X=5` binds the **number** 5 (canonical numbers only; `-D X=007` or `bob` are Strings) — this amends D4, see plan §0 step 7.

**CLI and tooling**
- `cobra` with `Args: ArbitraryArgs` on the root: a first argument that is not a subcommand is the script, so a script literally named `run`, `check`, `task`, `eval`, `repl`, `prelude`, `builtins`, `version`, `completion` or `help` must be passed as `./check`. A mistyped subcommand (`chek`) is read as a script path and fails with `cannot read script`.
- `ArgsLenAtDash()` is how script arguments are split off; `root.SetArgs(nil)` makes cobra read `os.Args` — `CommandTree.Execute` always passes a non-nil slice.
- `CommandTree.Root()` builds fresh flag state on every call (cobra flags otherwise accumulate across `Execute` calls). Edition commands are shared instances.
- The k8s edition's `main` builds a cluster client before the tree runs, so `version`/`completion` print the "no cluster access" notice on stderr (W6: make it lazy).
- `.gitignore` used to ignore `cmd/gatwd/` (fixed to `/gatwd`); if a new test file in an existing directory does not show in `git status`, run `git check-ignore -v <file>`.
- `gofmt -l` lists files that were **already unformatted** and are not part of this work (`parser/*`, `evaluator/graph_steps.go`, `pkg/plugins/registry.go`, `cmd/gatwd-gen/main.go`). Format only what you touch.
- Scratch scripts: `cat > f <<'EOF'` (zsh mangles `\x` in `echo`). `run_dsl_tests.sh` builds `./gatwd` in the repo root.
- Run the baseline binary **more than once** before blaming a diff: several tour scripts differ run to run on old code (map-iteration order). At the end of W1b the diff was empty on all scripts.
- Results of `EvaluateToValues` are `*collections.List[any]` (use `functions.ToSliceAny`); `let … in …` goes through `EvaluateLine`/`EvaluateToValues`.

## 6. Constraints (hard rules)

Mandates 1–8 (`AGENTS.md`). **Mandate 5:** core imports only stdlib + this module, except cobra/pflag in `pkg/cli` only; new base types and libraries are *generic* and live in
`functions/` / `collections/` with registration in `pkg/cli`. **Mandate 6:** nothing silent, `Maybe` (not an error) for malformed text in `str_to_int`/decoders, negative-control every guard.
**Mandate 8:** ask at hard-to-reverse choices and record them with a date in the plan, write deferrals down, shared tree — re-read before editing, stage by path, commit/push only when asked.

## 7. Verification gates

```bash
gofmt -l <dirs you touched>         # nothing listed for files you touched
go vet ./...
go test ./...                       # includes internal/archtest (layers, Clean Core, doc links)
./run_dsl_tests.sh                  # fails on [ERROR], silent statements, --check errors
go test -race ./pkg/... ./evaluator ./parser ./interactive ./cmd/gatwd-k8s-operator
```

Interpreter, parser and evaluator changes also need the **script regression diff** and a silent `--check` (recreate the helper; it is ephemeral):

```bash
git worktree add /tmp/base <previous-commit> && (cd /tmp/base && go build -o /tmp/gatwd_base ./cmd/gatwd)   # or: build from HEAD before editing
go build -o /tmp/gatwd_new ./cmd/gatwd
for f in $(find . -name '*.gatwd' -not -path './tests/*' -not -path './legacy/*' -not -path './.git/*' | sort); do
  timeout 20 <binary> $f </dev/null > <outdir>/$(echo $f | tr / _).txt 2>&1; done      # for base (twice) and new
diff -rq <base-out> <new-out>
for f in $(find . -name '*.gatwd' -not -path './tests/*' -not -path './legacy/*'); do /tmp/gatwd_new --check $f; done   # must print nothing
```

D2 will change the output of any script that concatenated with `+` (none found); every difference must be explained in the commit.

## 8. First steps

0. **Put the §0 questions to the user and wait** (they block nothing; "defaults" lets work start at once).
1. `git status -sb` (clean apart from `output.txt`), then the §7 gates once for a known-green start. Handle Q1 (push) as answered.
2. Pick the follow-ups the answers enable; independent ones run as parallel tracks (§3). Pinning tests first, then generic Go in `functions/`/`collections/`, then `pkg/cli/dslfuncs_*.go`, then DSL assertions (read every output line;
   **a statement ending in `=` continues on the next line — do not end an `assert` argument with `=`**), then docs. One commit per sub-step, all gates after each.

## 9. Pointers

- Plan and decisions: `docs/hardening_plan.md` (§0 progress, §1 decisions, §2.4 defect table F1–F18, §4.5 what W1b decided, §5 W2, §6 W3, §10.1 closing checklist).
- Command line: `pkg/cli/command.go`, `docs/dsl_guide.md` §23; the argument model: `pkg/cli/run.go`.
- Guards: `internal/archtest/{layers,clean_core,doclinks}_test.go`, `run_dsl_tests.sh`.
- Prompts and skills: `.agents/prompts/{add-language-feature,fill-matrix-cell,review-against-mandates,phase-handoff}.md`, `.agents/skills/{all-layers-sync,dsl-authoring,phase-delivery,recursive-dogfood}`.
- Memory notes for the assistant: `project-hardening-plan-decisions`, `project-agent-docs-and-guards`.
