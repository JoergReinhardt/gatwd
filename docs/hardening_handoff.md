# Handoff — hardening plan: W0–W6 and YAML are done and merged; next are `Duration`/`Time`/`Date` (§5.6), W7, then the closing checklist

> Handoff for a **fresh context**. It **supersedes the earlier version of this file** (the "W2 is done; run W3–W6 in parallel" brief) and
> nothing else; the k8s-graph handoffs `phase6/7/8_handoff.md` are historical and carry banners.
>
> **First, do §0** (put the open questions to the user). **Then read in this order:** `AGENTS.md` (mandates 1–8, §5 gates, §5a prevention rules) → `docs/hardening_plan.md`
> (§0 progress table, §1 decisions, §5.6 time types, §9 W7, §10.1 closing checklist, and the closure notes §6.2/§7.1/§8.1/§9) → `docs/development_playbook.md` (§3 gates and the
> regression check, §4 step vs function form, §6 DSL traps) → this file. Then start at "First steps" (§8).

---

## 0. START HERE — put the open questions to the user first

Everything below was **found or left open by the four parallel tracks** (2026-10-13). None blocks the next phase except where marked; present each with its default
so the user can answer in one line, record the answers (with the date) in `docs/hardening_plan.md` §1 and in §2 here.

| # | Open question | Recommended default | What it gates |
| :-- | :-- | :-- | :-- |
| Q1 | **Push** the local commits (`git log --oneline origin/master..HEAD`; about 25, from the W3–W6 merges and YAML)? | push, once the user has read the list | publication only |
| Q2 | **Remote kustomize bases** (`https://…`, `git@…`) are refused loudly (the library cannot be told not to fetch). Allow them behind an explicit opt-in flag? | stay refused; revisit with W7's capability policy | W5 follow-up |
| Q3 | `go.mod` carries `replace k8s.io/kube-openapi => 2026…` because kustomize v0.21.3 needs a newer one than `apimachinery` v0.37.1 allows. Adopt `client-go`/`apimachinery` v0.38 alpha to drop it? | keep the replace until v0.38 is stable | none |
| Q4 | **Exit codes (W6):** should `check` fail on violations by default (better for CI, breaks scripts that rely on exit 0), and should violations get their own code (1 = broken run, 3 = violations)? | opt-in `--fail-on-violation` stays; add code 3 for violations | W6 follow-up |
| Q5 | The duplicate-id error (W4) names both payloads but not the **source file**: no adapter attaches one. Add provenance to manifest documents (touches `adapters/manifest_adapter.go`)? | yes, a `_source` field set by the manifest adapter, shown only in errors | W4 follow-up |
| Q6 | An `ownerReferences` entry without `kind` is now an error (it used to match by name). Keep it strict? | keep | none |
| Q7 | **String vs collection equality (W3):** `"[2 1]" == [2, 1]` is still `true` because the pure-DSL Game of Life depends on it through HAMT keys. Fix HAMT keys so the equality can go? | yes, as a separate small phase after §5.6 | D2 completeness |
| Q8 | `2 ** -3` is a `Float` (0.125) today; or the exact `1r8`? | keep `Float` (a `Ratio` result needs an explicit `ratio`) | §6.2 |
| Q9 | `n -1` with `n` a number is an application error with a hint (a negative literal is an argument: `f -3`). Keep, or make it a subtraction again? | keep, the hint names the fix | §6.2 |

If the user is away or says "defaults", use the recommended column, **say so in the plan entry**, and continue.

---

## 1. Where we are

**State (2026-10-13):** `master` at the merge of Track B (`a09a901`) plus the doc commits after it; the remote is at `9373f01` — **everything since is local** (Q1). `output.txt` in the repo
root is untracked and **not ours** — do not add it. `.claude/worktrees/` is empty again (the four track worktrees and branches are deleted). Stage by name; commit and push only when asked.

**Done and merged** (the plan's §0 progress table has one row each): W0, W1a/b (cobra tree), W2 (`Char`, `String`/`Bytes` rows, `str_*`, `Ordering`, D2, F18, JSON, regex), **YAML**
(`yaml_parse/_all`, `yaml_encode/_all`, `go.yaml.in/yaml/v3` on the Clean Core allowlist for `pkg/cli`), **W3** (numeric tower, `Ratio`, unary minus, `**`, prefix `!`), **W4** (qualified ids),
**W5** (kustomize, Go 1.27), **W6** (layered overrides, exit codes, formats, lazy client). Gates on the merged tree: `gofmt`, `go vet`, `go test ./...`, `./run_dsl_tests.sh`, `-race`, silent `--check`,
script regression diff against `2526cc0` — all green, every difference explained.

**What changed that the next phases build on**
- **Numbers (W3):** `Int` is logical (int64 fast path, `big.Int`, demoted when it fits); `//` floors, `%` takes the divisor's sign; `Ratio` is `*functions.Ratio` (`3r4`); `Int` vs `Float` compares exactly; `"5" == 5` is `false`; `-` in operand position negates (`f -3` is a negative argument).
  `Duration / Duration` therefore has `Ratio` to land on (§5.6).
- **Ids (W4):** `Kind/namespace/name` (`_` for no namespace), built in `lib/k8s/ids.gatwd`; `graph_insert_node` refuses duplicates; `k8s_names` is gone; `collect_list` no longer drops error nodes.
- **Kustomize (W5):** `pkg/k8sgraph/kustomize.go`; `manifests` renders a kustomization root, `--no-kustomize` (a persistent `gatwd-k8s` flag, `Edition.WithFlags`); the operator renders too and has the option `k8sgraph.WithoutKustomize()` but **no flag yet**.
- **Operator (W6):** `cli.NewApplianceTree` (`check`, `run --interval`, `rules cat`, `version`, `completion`, `--rules-mode`), `--fail-on-violation`, `--format text|json`, `*cli.LoadError`, `LoadReport`; single-dash flags still work with a notice. **Layering is only safe once W7's deny-by-default policy exists** (an override can run OS commands).

## 2. Decisions — all answered (2026-10-13)

The user's answers to the first round of open questions are recorded in `docs/hardening_plan.md` §1 ("Answered at the start of the parallel run"): push (done), D14–D17 as proposed, YAML into core (done), auto-render with
`--no-kustomize` (done), prefix `!` as a synonym of `not` (done), parallel tracks (done). D1–D13 and D18 are in plan §1. **The questions of §0 above are new and open.**

## 3. Goals for the next phases (one commit per sub-step; sequential from here)

Start from `.agents/prompts/add-language-feature.md` / `fill-matrix-cell.md` and the `phase-delivery` skill.

1. **§5.6 `Duration` / `Time` / `Date`** (plan §5.6; D14–D17 accepted as proposed): `@…` and unit-suffix literals (`5s`, `@2026-10-13`), `now` as an injectable-clock adapter step (so tests are deterministic and core stays pure), `Time` carries a display zone and there is no `DateTime`,
   `Date ± whole-day Duration`, `Duration / Duration` is a `Ratio`. Every cell of Mandate 2 (including `==`/ordering/`+`/`-` and the function forms, parse/format) with negative cases; DST edge cases are named in the plan. Motivation: `age > 30d`, `expires - now < 14d` in `.gatwd` rules.
2. **W7** (plan §9): execution policy (D8: unknown steps run OS commands only for trusted command-line scripts; denied by default for overrides and any script not given on the command line), `context`, structured diagnostics (`cli.checkReport` is the one place that renders text and JSON), and the `parser/errors.go` `FormatError` wiring deferred from W0/W1b. Hard-to-reverse design choices: ask first.
3. **The closing checklist** for the k8s how-to and k8s code (plan §10.1; user request): every row the tracks appended; re-run every command of the how-to against the final binaries.
4. **Smaller follow-ups from §0** once answered (Q2–Q9), plus: an operator flag for `--no-kustomize`, `GATWD_<FLAG>` environment defaults and ConfigMap watching (deferred in W6), `json_encode` of a `Ratio` (an error today), exponent float literals (`1e-3`), the `--check` advisory beyond literals.

**Definition of done per cell** is in `AGENTS.md` Mandate 2 (Go method → operator → evaluator step → step form → function form → DSL assertion *inside a lambda* → structural Go test → `Example`). Each partial function needs a negative case; "step form ≡ function form" is a test, not a hope.

**Merging lessons from the parallel run:** the tracks conflicted only in additive places (the plan, the how-to, `tests/dsl_exhaustive_test.gatwd`, one test file) — keep both sides, renumber the DSL test sections, and migrate expectations that a later track changed (the operator tests needed the W4 ids). Run the *whole* gate set after every merge; `git show ":1:path"` needs braces in zsh (`":${v}:path"`).

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
- Exec steps split quotes like a shell and a non-zero exit is an error. An unknown pipeline step still **runs an OS command** (F17) until W7: never feed untrusted scripts to the interpreter.
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

0. **Put the §0 questions to the user and wait.** Nothing else happens before that (they do not block §5.6, so "defaults" lets work start immediately).
1. `git status -sb` (clean apart from `output.txt`), then the §7 gates once for a known-green start. Handle Q1 (push) as answered.
2. Build a baseline binary (`go build -o /tmp/gatwd_base ./cmd/gatwd`), write the pinning tests for §5.6 first, then the generic Go in `functions/` (types, literals in `parser/`), then `pkg/cli/dslfuncs_*.go`, then DSL assertions (read every output line; **a statement ending in `=` continues on the next line — do not end an `assert` argument with `=`**), then docs. One commit per sub-step, all gates after each.
3. Then W7 (ask first at its decision points), then the §10.1 closing checklist.
4. The four-track parallel mode is spent; a *future* parallel run should again use one worktree per track, self-contained briefs, "never push / never edit the progress table" and a merge order chosen by file overlap (the id migration last).

## 9. Pointers

- Plan and decisions: `docs/hardening_plan.md` (§0 progress, §1 decisions, §2.4 defect table F1–F18, §4.5 what W1b decided, §5 W2, §6 W3, §10.1 closing checklist).
- Command line: `pkg/cli/command.go`, `docs/dsl_guide.md` §23; the argument model: `pkg/cli/run.go`.
- Guards: `internal/archtest/{layers,clean_core,doclinks}_test.go`, `run_dsl_tests.sh`.
- Prompts and skills: `.agents/prompts/{add-language-feature,fill-matrix-cell,review-against-mandates,phase-handoff}.md`, `.agents/skills/{all-layers-sync,dsl-authoring,phase-delivery,recursive-dogfood}`.
- Memory notes for the assistant: `project-hardening-plan-decisions`, `project-agent-docs-and-guards`.
