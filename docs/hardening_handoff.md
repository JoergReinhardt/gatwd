# Handoff — continue the hardening plan at W3 (the numeric tower and `Ratio`); W2 is done

> Handoff for a **fresh context**. It **supersedes the earlier version of this file** (the W2/W3 brief written after W1b) and
> nothing else; the k8s-graph handoffs `phase6/7/8_handoff.md` are historical and carry banners.
>
> **Read in this order:** `AGENTS.md` (mandates 1–8, §5 gates, §5a prevention rules) → `docs/hardening_plan.md`
> (§0 progress table, §1 decisions, §5 W2, §6 W3, §10 sequencing) → `docs/development_playbook.md` (§3 gates and the
> regression check, §4 step vs function form, §6 DSL traps) → this file. Then start at "First steps" (§8).

---

## 1. Where we are

**State (2026-10-13):** everything up to `master` at `19a4b03` is pushed; **W2 is committed locally and not pushed** (about a dozen commits from `4cc0c8a` to
`e5e38ff`; `git log --oneline 19a4b03..HEAD`). `output.txt` in the repo root is untracked and **not ours** — do not add it. Stage by name; commit and push only when asked.

**W2 is done** (plan §0 has one row per step): `Char`; `String` and `Bytes` as collection rows; the `str_*` library, `show`, `str_format`; D18 string escapes; Bool functions and the
`Ordering` sum type (`compare` returns it); **D2** (`+` numeric-only, and `- * / // %` and the ordering operators reject a `String` beside a number; two Strings order as text); typed literal tokens for
`emit`/`const`/`-D`; F18 (every builtin is also a step); `Bytes` with base64/hex/url/utf8 and `sha256`/`md5`; JSON; RE2 regex. Several pre-existing silent defects were fixed on the way
(let-bound and constant Strings spliced as bare words, delimiters inside string literals confusing the statement joiner and the splitters). **Deferred, with reasons:** YAML (needs a
third-party library; Clean Core — an Edition registers it), `Unit`, prefix `!`. **Still planned for later:** `Duration`/`Time`/`Date` (plan §5.6, D14–D17 proposed, **ask first**), scheduled after W3.

What W1b changed (for orientation):

| Group | Files |
| :-- | :-- |
| The tree | `pkg/cli/command.go` (`CommandSpec`, `CommandTree`, `Execute` → exit status), `pkg/cli/repl.go` (`Interpreter.REPL`) |
| Run API | `pkg/cli/run.go` (`RunOptions` + `Eval`, `Define`, `Preludes`; `ParseDefines`; `parseRunArgs` is **deleted**), `pkg/cli/cli.go` (`RunWithInterpreter` is now three lines; `Interpreter.Out`) |
| Edition | `pkg/cli/edition.go` (`Command()`, `WithCommand`, `WithArgNormalizer`, `buildInterpreter(skipPreludes)`) |
| Tests | `pkg/cli/command_test.go` (backward-compat table, subcommand table, normaliser, `--no-prelude`, `Out`), `cmd/gatwd/cli_test.go` and `cmd/gatwd-k8s/cli_test.go` (build the real binary), `pkg/cli/cheatsheet_test.go` (now uses `Interpreter.Out`), `pkg/cli/example_test.go` (`ExampleCommandTree_Execute`) |
| Other | `go.mod`/`go.sum` (cobra v1.10.2, mousetrap), `.gitignore` (`gatwd` → `/gatwd`: the bare name also ignored the `cmd/gatwd/` directory), docs: `docs/{hardening_plan,dsl_guide,k8s_howto,development_playbook}.md`, `pkg/cli/doc.go` |

Gates at the end of W1b: `gofmt`, `go vet ./...`, `go test ./...`, `./run_dsl_tests.sh`, `-race` on the touched packages, a silent `--check`
over every `*.gatwd`, and an **empty script regression diff** against a baseline built from the previous `HEAD` — all green. Negative
controls done: the exit-code mapping and the `--no-prelude` plumbing were each broken once and the tests went red.

**What the command line is now** (the full list is in `docs/dsl_guide.md` §23): `gatwd [script] [-- args]`, `-i`, `--task NAME`, `-e`, `-D NAME=VALUE`,
`--prelude FILE`, `--no-prelude`, `--check`; subcommands `run eval check task repl prelude builtins version completion`. Exit `0` ok, `1` a command
failed, `2` a malformed command line. **Design calls made in W1b are listed in plan §4.5** (output via `Interpreter.Out`; `-e` replaces `main`
and fails on `[ERROR]` or silence; `steps` is `builtins`; `task` takes `-s SCRIPT`; deferrals with reasons).

## 2. Decisions — **all answered (2026-10-13)**

| ID | Decision |
| :-- | :-- |
| D1 | `Char` added; `String` is a collection row of `Char`s; **conversion works both ways, on demand** (user, 2026-10-13: *"if a serialized collection of chars gets treated as a string, convert it"*): a `Char` collection used as a `String` becomes the `String`, a one-character `String` becomes a `Char`, a `String` used as a `Char` collection becomes one. **Details accepted 2026-10-13:** `List`/`Chunk` of `Char` count as a string, a mixed collection is an error naming the element, an empty collection converts only where a `String` is demanded. Keep the conversion table narrow and in `valueops.go` (D2 removed implicit `+` coercion for the opposite reason). |
| D2 | `+` is numeric-only; `++` concatenates. **Corollary confirmed (2026-10-13):** `"a" + "b"` is also an error with a hint to use `++`. Check how `const`s (stored as text) behave first. |
| D11, D12, D13 | **Decided on the recommendation (2026-10-13)**, alternatives written down in plan §1/§6.1: `/` on `Int`/`Int` unchanged; `Ratio` vs `Float` promotes in arithmetic and compares exactly; `%` stays modulo (`%` as fraction operator is the recorded alternative). |
| D3 | (a): keep the runtime types, specify the tower, `Num`/`Integral`/`Fractional` are annotation-level constraint names. |
| D4, D5, D6, D9, D10 | as in plan §1 (D4 and D10 are implemented). |
| D7 | Bump the repo to **Go 1.27**, kustomize v0.21.3 (W5). No system Go update needed: `GOTOOLCHAIN=auto` fetches `go1.27.x`; only CI/Docker images need ≥ 1.27. Auto-render of kustomization roots was **not** explicitly answered — confirm at W5. |
| D8 | Yes: unknown step ⇒ OS command stays for trusted CLI use, denied by default for overrides and non-command-line scripts (W7). |

## 3. Goals for the next phase (W3, one commit per sub-step)

Start from `.agents/prompts/add-language-feature.md` / `fill-matrix-cell.md` and the `phase-delivery` skill. Ask before the choices that are hard to reverse. D11–D13 (the `Ratio` design) are decided.

**W3 (plan §6):** the numeric tower as a tested specification.
1. Logical `Int` (int64 fast path + `big.Int`, **demote** results that fit; fix the leading `+` of big integers, F11), the promotion lattice, `/ // %` and zero-division errors.
2. `**`/`pow`/`sqrt`/`floor`/`ceil`/`round`/`trunc`/`sign`/`gcd`/`lcm`/`to_int`/`to_float`/`is_nan`/`is_inf`, and **unary minus** (`- 5`, `abs -3`, `str_slice -1 …` currently parse as subtraction; the F2 rows `- 5` and `2 ** 3` in `pkg/cli/failloud_test.go` expect `cannot evaluate`: flip them).
3. The constraint names `Num`/`Integral`/`Fractional`/`Comparable`/`Appendable` in `ParseTypeAnnotation` and the checker (`pkg/cli/typecheck.go`), **zero false positives** on every committed script.
4. **Strings are not numbers — the rest of D2 (plan §6 item 7):** equality (`"5" == 5` must be `false`; `StructuralEqual`'s `%v` fallback and numeric-string branch), and the helpers `ToFloat64`/`toInt64`/`ToBigInt`/`ToComplex128` stop parsing `string` (they are why `nth "1" xs`, `range "0" "2"`, `min "5" 3` work).
5. **`Ratio`, exact fractions (plan §6.1):** `3r4` and `0.75r` literals, `ratio n d`, `*big.Rat` with demotion to `Int`, `to_ratio` (shortest decimal) / `to_ratio_exact` / `to_ratio_approx x max_den`, correctly rounded `to_float`; D11 (`/` on two Ints unchanged), D12 (Ratio vs Float: arithmetic promotes, comparison exact), D13 (`%` stays modulo) are decided.
6. An operator × type-pair matrix test, including the `String × number` rows.

**Definition of done per cell** is in `AGENTS.md` Mandate 2 (Go method → operator → evaluator step → step form → function form → DSL assertion *inside a lambda* → structural Go test → `Example`).
Each partial function needs a negative case; the parity check "step form ≡ function form" is a test, not a hope.

**After W3:** `Duration`/`Time`/`Date` (plan §5.6; **D14–D17 are proposed and must be asked before implementing**), then **W4** (qualified node ids, D6 — migrate all tests/examples in the same phase) → **W5** (Go 1.27 bump first) → **W6** (operator flags `--long` with the normaliser hook, layered rules, exit codes) → **W7** (policy, `context`, structured diagnostics).

**When everything is done:** work through the **closing checklist for the k8s how-to and k8s code** (plan §10.1; user request); add a row there whenever a phase changes behaviour the k8s edition or its how-to touches.

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
- `+` is numeric-only (D2, done): `"a" + 1` is an error naming both types and `++`. Text reaching arithmetic must be converted (`str_to_int`); the tokens of `emit`/`const`/`-D` are typed literals; spawn an **undeclared** type and the value is text (declare `type IntVal = Int`). `- * / // %` and comparisons still coerce numeric strings — a W3 item.
- Known, **not fixed**: `zip` is suspect (`echo Primary: %s |> zip date` yields only the date); big integers print with a leading `+` (F11, W3).
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

1. `git status -sb` (clean apart from `output.txt`), then the §7 gates once for a known-green start. Build the baseline binary from `HEAD` **before** editing: `go build -o /tmp/gatwd_base ./cmd/gatwd`.
2. Ask the user whether to **push** the W2 commits (never without asking).
3. W3 sub-step order I recommend: unary minus and `**` (they unblock the F2 rows and `abs -3`) → the logical `Int` demotion rule and the F11 display → the String-is-not-a-number leftovers (item 4) → the constraint names in the checker → `Ratio` → the matrix test last, so it documents the final behaviour.
4. For each: pinning tests first, then the generic Go in `functions/`, then `pkg/cli/dslfuncs_*.go`, then DSL assertions (read every output line; a statement ending in `=` continues on the next line — don't end an `assert` argument with `=`), then docs (`dsl_guide.md`, `grammar.md`). Update the plan's progress table and rewrite this handoff at the end of the phase; mark the old one superseded.

## 9. Pointers

- Plan and decisions: `docs/hardening_plan.md` (§0 progress, §1 decisions, §2.4 defect table F1–F18, §4.5 what W1b decided, §5 W2, §6 W3, §10.1 closing checklist).
- Command line: `pkg/cli/command.go`, `docs/dsl_guide.md` §23; the argument model: `pkg/cli/run.go`.
- Guards: `internal/archtest/{layers,clean_core,doclinks}_test.go`, `run_dsl_tests.sh`.
- Prompts and skills: `.agents/prompts/{add-language-feature,fill-matrix-cell,review-against-mandates,phase-handoff}.md`, `.agents/skills/{all-layers-sync,dsl-authoring,phase-delivery,recursive-dogfood}`.
- Memory notes for the assistant: `project-hardening-plan-decisions`, `project-agent-docs-and-guards`.
