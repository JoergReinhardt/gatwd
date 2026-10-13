# Plan: Base Types & Stdlib, Core CLI, Node Identity, In-Process Kustomize, Fail-Loud Sweep

> **Status (2026-10-13): W0, W1a and W1b (the cobra command tree) are implemented; W2–W7 are untouched.** Progress table: §0.
> **All decisions D1–D10 are answered** (§1, 2026-10-13). Two answers carry a point to confirm when their phase starts: D1 (the
> exact `Char` conversion rule, W2) and D7 (auto-render, W5).
> Facts marked **(verified)** were reproduced against the current tree while writing this plan;
> facts marked **(researched)** come from the sources listed at the end.
> **To start implementing, read [`hardening_handoff.md`](hardening_handoff.md).**
> Companion to [`k8s_howto.md`](k8s_howto.md) §4 (known limitations) and
> [`development_playbook.md`](development_playbook.md) (how each phase is run).

---

## 0. Summary

| # | Workstream | Why | Size | Breaking? |
| :-- | :-- | :-- | :-- | :-- |
| W0 | **Fail-loud sweep** | many failures are silent (verified list in §2.4) — the project's own Mandate 6 | M | mild (turns silence into errors) |
| W1 | **CLI**: a core `Run` API (script + arguments, `-e`, optional/additive preludes) **and a cobra command tree built into `pkg/cli`** with subcommands, completion, `--help` | needed by every future project; Editions/Appliances inherit it | M–L | operator flags become `--long` (migration shim) |
| W2 | **Base types & stdlib**: `Char`, string library, `not`, `Ordering`, conversions, `Bytes` + encodings, formatting | the DSL has a String *value* but no string *library*, no `Char`, no boolean `not` (§2) | L | `"x" + 1` becomes an error |
| W3 | **Numeric tower, specified**: logical `Int` (unbounded), `Float`, `Ratio?`, `Complex`; `Num`/`Integral`/`Fractional` as *constraint names*; `**`, rounding, conversions | promotion exists but is undocumented and untested as a matrix | M | display fix (`+123`) |
| W4 | **Node identity**: collision-proof ids | ids are `metadata.name` only (verified) | M–L | yes for ids in queries/tests |
| W5 | **In-process kustomize** in the k8s edition | no external binary; `manifests` can render | S–M | no |
| W6 | **Operator/Appliance**: layered overrides, exit codes, formats | override replaces rules (verified); exit 0 always | S | override mode default (D5) |
| W7 | **Execution policy & script sources** (later) | unknown steps run OS commands (verified) — unsafe for ConfigMap/API scripts | M | opt-in |

Order: **W0 → W1 → W2/W3 (parallel) → W4 → W5 → W6 → W7.** Rationale in §10.

**Progress**

| Phase | State |
| :-- | :-- |
| W0 | **done** (committed). One deviation: item 3 is half done — unparsable statements report *why* (`cannot evaluate "…": <reason>`) but not *where*; `parser/errors.go` `FormatError` is not wired into the infix/line parsers — **decided at W1b: deferred to W7**, see §4.5. F2's missing operators (`Char`, `not`, `**`, unary minus) belong to W2/W3. |
| W1a (`Run` API) | **done**: `RunOptions{Script, Args, Task, Eval, Define, Preludes}`, `Interpreter.Run`/`RunTask`, `ARGS`/`ARG0`/`main`/`--task` (D4), `Interpreter.Out` (statement output as a writer), `ParseDefines`. Deferred with reasons in §4.5. |
| W1b (cobra tree) | **done** (2026-10-13): `pkg/cli/command.go` (`CommandTree`, `CommandSpec`), `repl.go`; `Edition.Command/WithCommand/WithArgNormalizer`; the hand-rolled `parseRunArgs` and `--check` branch are deleted; backward-compatibility table `TestCommandTreeBackwardCompatible`, subcommand table, two real-binary tests (`cmd/gatwd`, `cmd/gatwd-k8s`). Design calls and deferrals: §4.5. |
| W2–W7 | not started |

---

## 1. Decisions

**Decided (2026-10-13):**

| ID | Decision | Notes |
| :-- | :-- | :-- |
| **D2** | **Remove the implicit string coercion.** `+` is numeric-only (so `x + 1` with a numeric `x` works, of course); `++` concatenates; `show`/`str_from` converts. | A *proposed corollary*, open to veto: `"a" + "b"` also becomes `[ERROR] operator '+' is not defined for String and String — use '++'` (what Elm and Haskell do). Verified: no script in the repo relies on string `+`. |
| **D5** | **Layered overrides by default**: the embedded rules load first, the override on top; `replace` stays as an explicit mode (`-rules-mode replace`). | Chosen as the idiomatic model: systemd drop-ins, Helm values, Kustomize overlays and CSS all layer; "replace the whole file" is the exception, not the default. |
| **D6** | **Qualified node ids** `Kind/namespace/name` + loud duplicate detection (A+C in §7). **All tests and examples are migrated in the same phase.** | |
| **D9** | **`assert` on an empty stream is an error.** | Scripts that relied on a vacuous assert get fixed. |
| **D4** | **Argument model, as recommended:** `ARGS` (List of String, after `--`), `ARG0`, `main` as the optional default task, `--task NAME` to run any `def`. | The flag syntax itself is cobra's (§4); D4 is what a script *sees*. **Extension made while implementing:** a pipeline-bodied task is a *fragment* whose input stream is the arguments (one String node each), a lambda-bodied task is called with the `ARGS` list. **Implemented** in `pkg/cli/run.go` (`RunOptions{Script, Args, Task}`, `Interpreter.Run`/`RunTask`) and, since W1b, on the cobra tree. Spec: `docs/dsl_guide.md` §23. |
| **D10** | **Use a real CLI library: cobra (+ pflag), integrated directly into core (`pkg/cli`).** The Clean Core rule gets an explicit, narrow exemption: cobra/pflag may be imported by `pkg/cli` only; the lower layers stay standard-library only. | Cobra is what kubectl, Helm and kind use; `pflag` is already in our module graph via client-go. A CLI framework is plumbing, not domain. **Implemented in `internal/archtest` (allowlist) and the rule docs on 2026-10-13.** |

**Decided later the same day (2026-10-13):**

| ID | Decision | Notes |
| :-- | :-- | :-- |
| **D1** | **`Char` is added; `String` becomes a collection row — "convert to char on demand if possible, otherwise yes"** (the user's words). | **Reading to confirm when W2 starts:** `Char` is a real type with `'a'` literals; where a one-character `String` meets a `Char` position (`char_code "a"`, `'a' == "a"`, a `Char` annotation) it converts on demand instead of erroring; where on-demand conversion is not possible (a multi-character string, a different type) the answer is "yes, a real distinct `Char` and a real collection row". Keep the coercion rule narrow and listed in one table, because D2 removed implicit `+` coercion for the opposite reason. |
| **D3** | **(a)** keep the runtime representations; specify the tower; `Num`/`Integral`/`Fractional` are annotation-level constraint names (no boxed `Num`). | W3 as written in §6. |
| **D7** | **Bump the repo to Go 1.27** and use `sigs.k8s.io/kustomize/api` **v0.21.3** (not the v0.21.0 pin). | The user also asked whether system Go must be updated: **no** — `GOTOOLCHAIN=auto` (the default; this machine has system Go 1.26.0 and the toolchain cache) downloads `go1.27.x` (1.27.2 is on proxy.golang.org) when `go.mod` says `go 1.27.0`. Only CI images, Dockerfiles and other contributors need Go ≥ 1.27 or auto-toolchain with network access (the repo has no Dockerfile or CI file that pins a Go version). **Not answered explicitly:** auto-render of kustomization roots with `--no-kustomize` — proceed on the recommendation at W5 and confirm. |
| **D8** | **Yes** — unknown pipeline step ⇒ OS command stays for trusted CLI use; **denied by default** for Appliance overrides and any script not given on the command line (capability policy, §9). | W7. |

---

## 2. Findings

### 2.1 What the DSL has today (verified)

| Area | Present | Missing |
| :-- | :-- | :-- |
| String *value* | `"abc"`; `++` concat; `==`, `<`; `length` (function form; counts characters) | **no string library** (split, join, trim, replace, contains, case, slice, parse, format) — and **no step form** of `length` at all |
| Char | — | no `'a'` literal (it parses to nothing and prints nothing) |
| Bool | `true`, `false`, `&&`, `\|\|`, comparisons | **no `not`/`!`** (both silently produce nothing); no `xor` |
| Integers | arbitrary size by auto-promotion to `big.Int`; `//`, `%`; `abs`, `min`, `max`, `negate`, `compare` | no `**`/`pow`; no unary minus on a spaced literal (`- 5`); no `floor/ceil/round/trunc`; no `toInt/toFloat`; no `gcd`, `divmod` |
| Floats | IEEE `Float`; `/` | silent precision loss on `Int→Float` (`9007199254740993 + 0.0` → `9.007199254740992e+15`) |
| Complex | `1 + 2i`, products | — |
| Unit / Ordering | `compare` returns `-1/0/1` | no `Unit`; no `Ordering` ADT (`LT/EQ/GT`) |
| Bytes | — | no `Bytes`; no base64/hex/utf8 conversions |
| Collections | List, Chunk, Set, HAMT, Tree, BST, Graph, Pair, Maybe, Either, Variant | n-ary tuples (only `Pair`); `Result` is `Either` |
| Time, JSON/YAML, regex | — | none (Go stdlib is available; per Mandate 5 it must be wired *generically*) |

### 2.2 The numeric tower as it behaves today (verified)

`4 / 2 = 2` (exact), `10 / 4 = 2.5` (so `/` is *exact-when-divisible, else float*); `7 // 2 = 3`; `3 == 3.0` is true;
`1 < 2.5` works across kinds; `Int + Int` overflow promotes to `big.Int`
(`9223372036854775807 + 1 = 9223372036854775808`) and results are **not demoted** back; small integers are
stored narrowed (int8…int64) as an implementation detail; `Uint` and `BigInt` exist as separate named types
but behave as part of the same tower; `+9223372036854775808` prints with a stray `+` (a `big.Int` formatting
bug). So the *behavior* is already a numeric tower (Scheme-style: exact integers → bigints → floats →
complex). What is missing is its **specification, tests, and a name** for it.

### 2.3 What comparable languages treat as essential (researched)

| | Haskell (Prelude/base) | Elm | Gleam | Dhall | Jsonnet | Clojure |
| :-- | :-- | :-- | :-- | :-- | :-- | :-- |
| Bool | ✓ `not`, `&&`, `||` | ✓ `not`, `xor` | ✓ `bool` module | ✓ | ✓ | ✓ |
| Integers | `Int` (bounded) + `Integer` (unbounded) | `Int` | `Int` (unbounded on BEAM) | `Natural`, `Integer` | one `number` | `long`, `BigInt` |
| Floating | `Float`, `Double` | `Float` | `Float` | `Double` | `number` | `double`, `BigDecimal` |
| Exact fractions | `Rational` | — | — | — | — | `Ratio` |
| Char | ✓ | ✓ (`'a'`) | — | — | — | ✓ |
| String | `String = [Char]`; `Text` library | `String` + module | `String` + `gleam/string` | `Text` + `Text/replace`, `Text/show` | `std.split`, `std.format`, `std.base64`, … | `String` + `clojure.string` |
| Bytes | `ByteString` | — | `BitArray` | — | via `std.base64` of arrays | byte arrays |
| Unit / Ordering | `()`, `Ordering` | `()`, `Order` | `Nil`, `Order` | — | — | `nil` |
| Maybe / Result | ✓ / `Either` | ✓ / `Result` | `Option` / `Result` | `Optional` | — | `nil` punning |
| Numeric generics | class hierarchy `Num ⊃ Real ⊃ Integral`, `Fractional ⊃ Floating`, `RealFrac` | **constrained type variables** `number`, `comparable`, `appendable` | explicit (`int.add` vs `float.add`) | explicit | one numeric type | tower + **opt-in** auto-promoting ops (`+'`) |
| Overflow | wraps (`Int`) / unbounded (`Integer`) | JS number | unbounded (BEAM) | n/a | float | **throws** unless `+'` |

Takeaways for gatwd:

- **We are not missing `String` as a *type* — we are missing its library, `Char`, and `not`.** Every language above
  that has `String` ships a string module; infra/config languages (Jsonnet, Dhall) ship `split/format/base64/replace`
  and JSON parsing as their core standard library.
- **A generalized `Num` is a *constraint*, not a value type.** Haskell expresses it as a class, Elm as the
  constrained variable `number` (Int or Float). gatwd's checker already treats all numeric types as mutually
  consistent; naming that relation (`Num`, plus `Integral` for `//`/`%` and `Fractional` for `/`) completes the idea
  with no runtime cost (D3). Explicit conversion (`toFloat`, `toInt`) is what Elm and Gleam require; our automatic
  `Int→Float` promotion is the one place we diverge, and Clojure's design shows the cost (silent precision loss).
- **Essential and cheap:** `Char`, string library, `not`, `Ordering`, `Maybe`-returning parsers (`str_to_int`),
  `Bytes` + base64/hex/utf8. **Infra-specific but generic:** JSON/YAML encode/decode, regex (Go RE2), hashes,
  `Time`/`Duration`. These follow the Go-standard-library rule: wired generically through every layer.

### 2.4 Silent failures and other defects (verified)

| # | Defect | Evidence | Plan |
| :-- | :-- | :-- | :-- |
| F1 | `1 / 0`, `1 // 0`, `1 % 0` print nothing | probe | **done (W0)** — W0: arithmetic errors → `[ERROR]` |
| F2 | `true + 1`, `'a'`, `**`, `!true`, `not true`, `- 5` print nothing | probe | **W0 done:** every one now reports `[ERROR] cannot evaluate …` with the reason. **Open (W2/W3):** add `Char`, `not`, `**`, unary minus themselves — their rows in `pkg/cli/failloud_test.go` must then change from "error" to "value" |
| F3 | inline `map (\x -> x + [1])` → `[<nil> <nil>]`; `map (\x -> nosuchfn x)` → `[<nil> <nil>]` | probe | **done (W0)** — W0: unify lambda evaluation (§3) |
| F4 | inline `filter (\x -> nosuchfn x)` → `[]` (error read as "false") | probe | **done (W0)** — W0 |
| F5 | inline `fold (\a x -> a + nosuch) 0` → `0` (the seed) | probe | **done (W0)** — W0 |
| F6 | top-level `let r = <hamt> in hamt_get "kind" r` prints nothing (string-substituted value) | probe | **done (W0)** — W0 spike: bind values, not text |
| F7 | `||>` at top level prints nothing | probe | **done (W0)** — W0 |
| F8 | `assert` on an empty stream passes | design | **done (W0)** — W0 (D9) |
| F9 | `gatwd --help` is read as a file name | probe | W1 |
| F10 | operator exits 0 with violations; one OS-step probe with quoted arguments (`sh -c "echo …$((6*7))"`) printed nothing (cause not yet isolated — may be the `$(( ))`) | probe | **W0 done:** the quoted-argument probe was `strings.Fields` ignoring quotes (now `interactive.SplitCommandLine`), and failed starts / non-zero exits were dropped (now error nodes with stderr). **Open (W6):** the operator still exits 0 with violations |
| F11 | `big.Int` results print with a leading `+` and are never demoted | probe | W3 |
| F12 | `to_set` builds `Set[string]`, function forms build `Set[any]` | code | W2: one `Set[any]` |
| F13 | edges to missing nodes are invisible in listings but visible in `graph_in_of` | code | W4: one consistent rule |
| F14 | `Configures` derived twice (volume + env) → duplicate edge | fixture | W4: dedupe identical specs |
| F15 | `Parser.Parse` godoc claims an unregistered `spawn` type yields no node (the interpreter registers on demand) | probe | **done (W0)** — doc fix |
| F16 | override replaces embedded rules; imports in an override resolve against the process cwd | probe | W6 |
| F17 | **unknown pipeline step runs an OS command** (`|> printf …` executes) | probe | W7 |
| F18 | `length` has no step form (function forms in general are not usable as steps) | probe | W2: generic builtin-as-step fallback (§5.4) |

---

## 3. W0 — Fail-loud sweep  (**done — see the progress table in §0**)

Goal: **no evaluation path returns nothing, `nil`, `0`, or its input unchanged on failure.**

1. **One lambda implementation.** The inline `map`/`filter`/`flatmap`/`fold` steps and the graph lambdas
   (`compileUnary` returns its input on error) hand-parse `\x -> …` text. Replace them with: parse once to
   `parser.LambdaExpr` → `*functions.Closure`, then run through `mapStepOver`/`filterStepOver` and the closure
   `fold`, all of which propagate errors. This deletes three hand-written parsers (Mandate 4) and fixes F3–F5 at the root.
   *Mechanism for errors in collection maps:* `mapStepOver` currently ignores `evalElem` errors for collections —
   change to yield an error node for the whole collection naming the failing element.
2. **Arithmetic and type errors** (F1, F2): `EvalBinaryOp` returns errors for division/modulo by zero and
   unsupported operand pairs; make sure every statement-level path prints `[ERROR] <reason>` instead of
   "(No output or node unhandled)". That message itself should be reserved for statements that legitimately
   produce no value (`type`, `def`, `const`).
3. **Surface parser errors** (F2): wire `parser/errors.go` `FormatError` into the infix and line parsers so a
   statement that cannot be parsed reports *where* and *why* (the roadmap's deferred C item).
4. **`assert` on an empty stream is an error** (F8, D9); update the few tests that rely on it.
5. **`let` binds values** (F6): the top-level `let` stores a stringified copy next to the real value and later
   substitution can clobber it. Spike first (a day): bind only the real node, and fall back to text only for
   non-representable values; measure with the script regression diff.
6. **`||>`** (F7) and quoted arguments to OS-exec steps (F10): isolate the cause first, then fix or make it a loud error.
7. **Tests:** each F-row gets a regression test that asserts an `[ERROR]` with a message naming the cause;
   `run_dsl_tests.sh` already fails on silence.

**DoD:** the F-table rows F1–F8, F10, F15 closed; script regression diff clean except the intended
new errors; `--check` silent; race-clean.

---

## 4. W1 — CLI: the `Run` API and a cobra command tree in `pkg/cli` (D4, D5, D10)

Today `cli.RunWithInterpreter` accepts exactly one script path (plus `--check` and `-i`); everything else is read as
a file name (F9), and the operator uses the standard `flag` package. W1 gives every Edition and Appliance a real
command line. **Cobra is integrated directly into `pkg/cli`** (D10): the Clean Core test now carries an explicit
allowlist — cobra and pflag, `pkg/cli` only — so the lower layers (`functions`, `collections`, `evaluator`, `parser`, …)
remain standard-library only and any other third-party import in core still fails the build.

### 4.1 Which library — cobra

| Library | Verdict |
| :-- | :-- |
| **spf13/cobra** (+ pflag) | **chosen** — the Kubernetes standard (kubectl, Helm, kind all use it); POSIX flags, nested subcommands, generated help and shell completion (bash/zsh/fish/powershell), `--` handling, doc generation. `pflag` is **already an indirect dependency** (client-go). Version checked: v1.10.2 builds on our Go 1.26 line. |
| urfave/cli | simpler, but no Kubernetes alignment and weaker completion/nesting |
| alecthomas/kong | struct-tag driven and neat, but small ecosystem; unfamiliar to Kubernetes contributors |

Cobra's known rough edge for us — a bare script path being rejected as an "unknown command" when subcommands exist —
was **tested in a scratch module**: `Args: cobra.ArbitraryArgs` on the root keeps `gatwd script.gatwd -- a b`
working next to `gatwd run …` / `gatwd check …`, and `cmd.ArgsLenAtDash()` separates script arguments. The one
confirmed incompatibility is that pflag rejects **single-dash long flags** (`-manifests` → `unknown shorthand flag`).

### 4.2 The `Run` API (semantics, independent of the command tree)

```go
// package cli
type RunOptions struct {
    Script      string            // path, or "-" for stdin; or Source for an in-memory script
    Eval        []string          // statements/pipelines to evaluate (-e)
    Args        []string          // exposed to the DSL as ARGS (List of String); ARG0 = script path
    Define      map[string]string // -D NAME=VALUE string constants
    Task        string            // run `def NAME` (fragment) or call closure NAME with ARGS
    Preludes    []string          // extra preludes, additive
    NoPrelude   bool              // skip the edition's embedded preludes (steps stay registered)
    Check       bool              // validate only
    Interactive bool
    Stdin       io.Reader; Stdout, Stderr io.Writer
}
func (i *Interpreter) Run(ctx context.Context, o RunOptions) (Result, error)
```

The command tree is a thin mapping from flags to this call, so the same semantics serve the **API-server / ConfigMap
hook** (W7): the call returns values and errors instead of printing, with a `context` for timeouts and (later) a
capability `Policy`. Keeping the semantics out of the cobra code also keeps tests simple and leaves the command tree
easy to split out again if the library-extraction work (Workstream F) wants a cobra-free core.

- **Arguments in the DSL:** `ARGS`, `ARG0`, and a generic `env "NAME"` builtin returning `Maybe String` (subject to the W7 policy).
- **Entry point convention:** if the script defines `def main = \args -> …` it is called with `ARGS` after loading and its
  result printed; otherwise top-level statements run as today. `--task` selects any def.
- **Preludes:** `Edition.WithPrelude*` stays additive in registration order; `NoPrelude` skips them; `Preludes`
  appends user preludes (`GATWD_PRELUDE_PATH` is the optional environment form).

### 4.3 The command tree (cobra, in `pkg/cli`)

`Edition.Run()` and `RunWithInterpreter` become the cobra tree below (the old hand-rolled argument parsing is deleted):

```
gatwd-k8s [flags] [script] [-- arg ...]       # root: backward-compatible script runner / REPL
gatwd-k8s run   SCRIPT [-- arg ...]           # explicit form
gatwd-k8s eval  -e 'pipeline' [-e …]          # one-liners
gatwd-k8s check FILE...                       # structural + advisory typing; exit 1 on errors
gatwd-k8s task  NAME [-- arg ...]             # run a def from the loaded script/prelude
gatwd-k8s repl                                # interactive
gatwd-k8s prelude list | cat NAME             # list / print the EMBEDDED prelude files (copy-edit-shadow workflow)
gatwd-k8s steps                               # list registered steps and builtin functions
gatwd-k8s version
gatwd-k8s completion bash|zsh|fish|powershell
persistent flags: --prelude FILE (repeatable) · --no-prelude · -D NAME=VALUE · -e/--eval · --check
```

- **Editions add commands, not logic** (Mandate 7): `Edition.WithCommand(cmd *cobra.Command)`. A domain subcommand is a
  thin mapping from flags to a DSL task defined in the prelude (e.g. the operator's `check` and `run` subcommands call
  `compliance`; a future `gatwd-k8s trace FROM TO` calls a prelude `def`).
- **Operator:** `gatwd-k8s-operator check` (once), `run` (interval loop), `version`; flags `--manifests`, `--namespace`,
  `--rules`, `--rules-mode layer|replace`, `--interval`, `--fail-on-violation`, `--format text|json` (W6).
- **Migration of single-dash flags:** the operator's documented `-manifests/-namespace/-rules/-interval` and the sample
  `deploy.yaml` become `--long` flags. A small argument normaliser rewrites a leading single dash on *known long flag
  names* to `--` and prints a one-line deprecation notice on stderr for one release; docs, `deploy.yaml`, tests and the
  how-to are updated in the same commit.
- **Environment:** an optional tiny helper maps `GATWD_<FLAG>` to flag defaults (12-factor, handy in containers);
  **no Viper** (heavy, unneeded).
- **Script named like a subcommand:** `gatwd check` runs the subcommand; run a script literally called `check` as
  `gatwd ./check`. Documented and tested.
- **Optional kubectl plugin:** an executable named `kubectl-gatwd` on `PATH` is a kubectl plugin; shipping the k8s edition
  under that name costs nothing.
- **Cost, accepted:** anything that imports `pkg/cli` now links cobra. Put the command tree in its own files
  (`command*.go`) so it stays separable.

### 4.4 Tests and definition of done

- **Guard (done):** `internal/archtest`'s `coreImportAllowlist` permits cobra/pflag in `pkg/cli` only; negative-controlled
  (cobra in `functions`/`parser` and client-go in `pkg/cli` are rejected).
- Table-driven tests with `cmd.SetArgs(...)` for every command and flag; a **backward-compatibility table** (every
  documented old invocation maps to the same behavior, including `script`, `-i script`, `--check file`, `script -- args`).
- End-to-end tests with a built binary: script + args + `main`; `--no-prelude` leaves `k8s_graph` undefined but `manifests`
  defined; `check` exit codes; completion scripts generate.
- `docs/dsl_guide.md`, `docs/k8s_howto.md` (§3.2 flags table, deploy manifest) and the playbook updated.
- **DoD:** `--help` everywhere; old invocations unchanged (or shimmed with a notice); only `pkg/cli` imports cobra.

### 4.5 What W1b built, and what it decided (2026-10-13)

The command tree is exactly §4.3 with these differences. Each was a call made while implementing; veto any of them.

- **Where output goes:** option (a) — `Interpreter.Out io.Writer` (default `os.Stdout`, inherited by forks). Top-level statements,
  the REPL and the tree all print through it; `captureStdout` is gone from the tests. `StepDebugPrint` (the pipeline tap) and
  `evaluator/testing.go` still use the process stdout by design.
- **`RunOptions`** gained `Eval`, `Define`, `Preludes`. `NoPrelude` is **construction-time**, so it lives in the tree
  (`CommandSpec.Build(noPrelude)`), not in `Run`. `-e` statements replace the default `main` task; a statement that prints an
  `[ERROR]` **or nothing at all** makes the run fail (Mandate 6; script top-level statements keep today's exit code until W6).
- **`-D NAME=VALUE`** binds a string constant; an invalid name, a missing `=`, or `ARGS`/`ARG0` is a usage error (exit 2).
  `--prelude FILE` (repeatable) loads before the script; an unreadable prelude is an error, not a silent skip.
- **`steps` is `builtins`:** step-form names are matched by prefix and cannot be enumerated, so the command lists the curried
  function forms (with parameters), the prelude defs and the constants; the help says so.
- **`task NAME [-s SCRIPT] [-- args]`:** the script comes from `-s/--script`, so `--` stays unambiguous.
- **Exit codes:** 0 ok · 1 a command failed · 2 a malformed command line (unknown flag, second positional, missing value, bad `-D`).
  One intended change from the old binary: `--check` with no file now exits **2**, not 1. The banner is printed only by the
  script runner and the REPL, not by `check`, `eval`, `task`, `builtins`, `version` or `completion`.
- **The argument normaliser hook** exists (`CommandSpec.NormalizeArgs`, `Edition.WithArgNormalizer`) and is tested with a toy
  rewrite; the single-dash operator migration itself is W6.
- **Deferred, with reasons:** `context.Context` in `Run` (the evaluator has no cancellation; adding the parameter would be a no-op
  that looks like a feature — W7 brings the policy and the timeout together); `Stdin`/`Stderr` fields (the tree owns them);
  `RunOptions.Check`/`Interactive` (command-line concerns, in the tree); the `GATWD_<FLAG>` environment defaults (optional in §4.3,
  nothing needs them yet); a generic `env "NAME"` builtin (only together with the W7 policy hook); **"where" for unparsable
  statements** — `FormatError` belongs to the combinator tracker, which the infix/line parser does not use, and the script loader
  knows the line while `EvaluateLine` does not; structured diagnostics (line, column, code) are what the W7 API server needs
  anyway, so it is done once there. `gatwd --check` already reports `file:line:col` for structural errors.
- **Known, not fixed here:** the k8s edition's `main` builds a cluster client *before* the tree runs, so even `version` and
  `completion` print the "no cluster access" notice on stderr; make the client lazy in W6.

---

## 5. W2 — Base types and standard library

All additions are **generic** (Mandate 5), live as pure Go in `functions/` (Layer 0, structurally tested), are
registered as curried data-last function forms in `pkg/cli/dslfuncs_*.go`, and get DSL assertions
(Mandate 3) and `Example*`s. Naming follows the existing prefix style (`set_`, `hamt_`, `tree_`).

### 5.1 Char and String (D1, D2 decided)

- **`Char`**: literal `'a'`, escapes `'\n'`, `'é'`; a Unicode code point (Go `rune`); `==`, `<`, `compare`;
  `char_code`, `char_from_code`, `char_is_digit/alpha/space/upper/lower`, `char_upper/lower`.
- **String is a collection row of Chars** (Mandate 2): `map`, `filter`, `fold`, `reverse`, `length`, `++`,
  `==`, `<`, `member` added to `valueops.go` — Elm has `String.map/filter/foldl` for exactly this.
- **`str_*` library** (data-last, curried): `str_split sep s`, `str_join sep xs`, `str_trim`, `str_upper`, `str_lower`,
  `str_contains/starts_with/ends_with`, `str_replace old new s`, `str_slice from to s` (by characters),
  `str_index_of` (`Maybe Int`), `str_repeat`, `str_pad_left/right`, `str_chars`/`str_from_chars`,
  `str_lines`/`str_words`, `str_to_int`/`str_to_float` (**`Maybe`**, never an error for malformed text),
  `str_from` / `show` (any value → text), `str_format fmt args` (a printf subset).
- **Regex** via Go RE2 (linear-time, safe for untrusted patterns): `re_match`, `re_find`, `re_find_all`, `re_replace`.
- **`+` is numeric only** (D2, decided): `"x" + 1` becomes `[ERROR] operator '+' is not defined for String and Int`,
  and — the proposed corollary — `"a" + "b"` becomes an error with a hint to use `++`; numeric `x + 1` is unaffected.
  `++` is concatenation; `show` converts. The checker already flags this class.
- **One `Set[any]`** (F12): `to_set` stops building `Set[string]`.

### 5.2 Booleans, Ordering, Unit

`not`, `xor`, `and`/`or` as functions (for folds: `fold and true`), `==`/`!=` on Bool. `compare` returns an
`Ordering` ADT (`LT | EQ | GT`) with `ordering_to_int`; `Unit` (`()`) as the value of side-effect-only steps.

### 5.3 Bytes and encodings (the Secret question)

- **`Bytes`** (immutable Go `[]byte`): `bytes_length/slice/++`, `==`, `utf8_encode`, `utf8_decode` (**`Maybe String`**).
- **`base64_encode/decode`**, **`hex_encode/decode`**, **`url_encode/decode`** (decoders return `Maybe`).
- **`json_parse`/`json_encode`** (JSON → nested HAMT/List/scalars; the same shape resources already have) and
  **`yaml_parse`/`yaml_encode`**.
- **Hashes:** `sha256`, `md5` (as hex `String`) for fingerprints.
- With these, "Secret `data` → decode → host" becomes expressible in the DSL; **whether the graph should do that is
  a separate security call** (credentials should not be copied into graph payloads — see the security note in the how-to).

### 5.4 Every function form is also a step (F18, Mandate 2 parity)

Instead of hand-writing a step per function, add a **generic fallback in step dispatch**: if a step's head names a
registry builtin (a curried closure) and no existing step claims it, apply it data-last — `"abc" |> length`,
`xs |> str_join ","`. The fallback sits *after* the existing step chain (so `head`, `map`, … keep their current
meaning) and *before* the OS-exec fallback. This closes the step/function parity gap for the whole matrix at once.

### 5.5 Later tier

`Time`/`Duration` (Go `time`), `Rational` (exact fractions; see W3), bit operations, n-ary tuples.

**DoD (W2):** each type/library row has structural Go tests, DSL assertions (output verified non-empty),
`Example*`, docs in `dsl_guide.md`/`grammar.md` (new literal), checker support for `Char`/`Bytes`/`Ordering`, and a
negative case per partial function.

---

## 6. W3 — The numeric tower, specified (D3)

**Decision D3(a):** keep the runtime representations; make the tower an explicit, tested specification.

1. **Logical types:** `Int` (unbounded; implemented as `int64` fast path + `big.Int`, canonical after each operation —
   **demote** `big.Int` results that fit, fixing F11's "never demoted"), `Float` (IEEE 754 double), `Complex`;
   optional `Ratio` (exact fractions) as a later addition. `BigInt`/`Uint` remain accepted *names* (aliases of
   `Int`) so existing `type X = BigInt` lines keep parsing.
2. **Promotion lattice (documented):** `Int → Float` only when the other operand is `Float` or the operation is
   inexact; `Int/BigInt` overflow stays exact; anything `→ Complex` when an imaginary operand appears.
   `Int → Float` can lose precision beyond 2⁵³ — **raise a *warning* from `--check`** when the operand is a literal
   or a known `Int` that large, and document `toFloat` as the explicit form.
3. **Division:** keep `/` (exact when divisible, else `Float`) and `//` (floor division); add `divmod`, `quot/rem`
   if wanted; **`/ 0`, `// 0`, `% 0` are errors** (W0).
4. **New functions:** `**`/`pow`, `sqrt`, `floor`, `ceil`, `round`, `trunc`, `sign`, `gcd`, `lcm`, `to_int`,
   `to_float`, `is_nan`, `is_inf`; unary minus on any expression; `min/max/abs/compare` already exist.
5. **Constraint names for annotations** (the generalized `Num`): in `ParseTypeAnnotation` and the checker,
   - `Num` ≡ any of Int, Float, Complex (what `number` is in Elm),
   - `Integral` ≡ Int (operands of `//`, `%`), `Fractional` ≡ Float/Complex-ish (operands of `/`),
   - `Comparable` ≡ Num ∪ Char ∪ String ∪ lists/tuples of Comparable, `Appendable` ≡ String ∪ List,
   so `def sum2 : Num -> Num -> Num = \a b -> a + b` documents and checks intent without a runtime wrapper.
6. **Tests:** an operator × type-pair matrix for the tower (Mandate 2 for numerics), promotion/demotion cases,
   overflow boundaries, and the display of big integers without `+`.

---

## 7. W4 — Node identity (D6, decided)

**Problem (verified):** a node's id is `metadata.name`. Two resources with the same name collide — across kinds
(a `Service` and a `Pod` named `web`) and across namespaces (`default/web` vs `prod/web`) — and a later
`graph_add_node` silently replaces the earlier one. Extension kinds (Databases, DBMSs) make it worse.

**Options:**

| | Approach | Verdict |
| :-- | :-- | :-- |
| A | **Convention**: id = `Kind/namespace/name` (cluster-scoped: `Kind/_/name`), built in the prelude | **recommended** — pure prelude change; no core change |
| B | Structured (tuple) ids in the core `Graph` | rejected — rewrites Layers 0–2 for no gain |
| C | Keep names, **fail loudly on duplicates** | adopt *with* A: detect collisions at ingest |

**Design (A + C):**
- Ingest uses `k8s_id kind ns name` (`kind ++ "/" ++ ns ++ "/" ++ name`; unspecified namespace → the literal `_`).
- **References resolve with the referrer's scope**, because Kubernetes references are namespace-local:
  ownerReferences (same namespace as the dependent), Service selectors, Ingress backends, PVC/ConfigMap refs, and
  RoleBinding subjects (which carry their own namespace). A helper `k8s_ref kind ns name g` returns the existing id
  (`Kind/ns/name`, falling back to `Kind/_/name` for cluster-scoped or namespace-less offline manifests).
- **Duplicate detection:** a generic `graph_insert_node` (errors on an existing id) alongside the upserting
  `graph_add_node`; the prelude's ingest uses it, so a collision is an `[ERROR]`, not a silent overwrite.
- **Display:** a helper `k8s_short id` and rule outputs that print `Kind/ns/name`; the `[violation, resource]`
  pairs stay readable.
- **Consistency of edges (F13) and duplicates (F14):** adjacency queries ignore edges whose endpoint node does not
  exist, matching the listings; identical `(from, to, color)` specs from one pass are de-duplicated.
- **Generic guideline** (playbook + `edition-appliance` skill): *every adapter names nodes `Kind/scope/name`*;
  the SQL example becomes `Database/pg-prod/orders`, `DBMS/_/pg-prod`.
- **Migration cost (be honest) — decided to pay it in the same phase, tests and examples included:** every relation, rule, fixture expectation and how-to example changes ids;
  queries become `graph_shortest_path Ingress/default/web Pod/default/nginx-pod`. Mechanical but wide; one commit
  for the prelude + tests, one for docs/examples.

---

## 8. W5 — In-process kustomize (D7: bump to Go 1.27)

**Findings (researched, and a scratch module built and ran):** `sigs.k8s.io/kustomize/api/krusty` is the supported
embedding API (`krusty.MakeKustomizer(krusty.MakeDefaultOptions()).Run(fs, path)` → `ResMap.AsYaml()`). In a scratch
module it **coexisted with our `k8s.io/client-go v0.37.1`** and rendered a real overlay (`namePrefix` applied). The scratch binary
built to ~33 MB unstripped (measure the real delta against the stripped operator when implementing). Constraints:

- **Go version (D7, decided):** the newest `api v0.21.3` requires **Go ≥ 1.27**; `v0.20.1` and `v0.21.0` stay on Go 1.26.
  → **bump the repo to Go 1.27** (`go 1.27.0` in `go.mod`; the go command fetches the toolchain itself, no system update
  needed) and use v0.21.3. Re-run every gate after the bump, before any kustomize code.
- **Concurrency:** upstream reports a data race when one `Kustomizer` is used from several goroutines → create a
  `Kustomizer` per call and serialise renders with a mutex (our `||>` parallel steps could otherwise hit it).
- **Filesystem:** `filesys.MakeFsOnDisk()` for paths, `MakeFsInMemory()` to render bases shipped in a ConfigMap or
  embed; `embed.FS` itself is not supported (read-only) — copy into the in-memory FS.
- **Safety defaults:** keep the default load restrictions (no files outside the kustomization root), leave exec
  plugins / Helm / KRM functions **disabled**, and disable remote (git/http) bases unless explicitly enabled.

**Design:** `pkg/k8sgraph/kustomize.go` provides `InProcessRenderer()` implementing the adapter's existing
`WithRenderer` seam; `KustomizeStep` uses it by default (the external-binary renderer stays selectable via
`WithKustomizeCommand`). The dependency is imported **only** by the k8s shim, never by core
(`internal/archtest` stays green). `manifests` **auto-renders a kustomization root** (D7, to confirm) with `--no-kustomize` /
an option to opt out. Tests: render the overlay fixture and compare with `kubectl kustomize` when present (skip
otherwise); a `-race` test with concurrent renders; a test that an exec plugin is refused.

---

## 9. W6 / W7 — Operator, policy and script sources

**W6 (small):**
- **Override mode (D5, decided: layer by default):** `Appliance` gets `WithOverrideMode(Layer | Replace)`; default **Layer** (embedded first,
  override second, so an override can reuse `rule_*` and redefine only what it needs); flag
  `--rules-mode layer|replace`. Imports in an override resolve against **the override file's directory**, not the cwd (F16).
- **Exit codes and formats:** `--fail-on-violation` (exit 1 when the result is non-empty), `--format text|json`
  (JSON for log pipelines), still exit 2 for "no source".
- Operator `--help`/flags come from W1.

**W7 (later; design now, build when needed):** running scripts that did **not** come from the command line —
ConfigMaps, API requests — is only safe with a **capability policy**, because an unknown step executes an OS
command (verified, F17) and the DSL can read files and sockets.
- `cli.Policy{AllowExec, AllowFS, AllowNet, AllowEnv}`; the CLI script argument is trusted (all allowed);
  Appliance overrides and API-submitted scripts default to **deny exec/FS/net/env**.
- Deny-by-default means the exec fallback becomes an `[ERROR] step 'printf' requires the exec capability`.
- The `Run(ctx, Request) Result` API from W1 carries the policy, a timeout (`ctx`), and an output size limit.
- Out of scope here: authentication/authorisation of API callers (belongs to the server Appliance).

---

## 10. Sequencing, gates, risk

```
W0 fail-loud ──► W1 CLI ──┬─► W2 base types/stdlib ──┐
                          └─► W3 numeric tower ──────┼─► W4 node identity ─► W5 kustomize ─► W6 operator ─► W7 policy
                                                     └─(W2 string ops needed by W4's k8s_id)
```

- **W0 first**: everything after relies on errors being visible; it also removes the hand-parsed lambda code the
  later work would otherwise have to touch.
- **W1 before W2/W3**: the new library is exercised through scripts that take arguments. W1 itself is two steps: the `Run`
  API first (W1a), then the cobra command tree in `pkg/cli` (W1b), which **W6 (operator subcommands/flags) depends on**.
- **W2 before W4**: `k8s_id` needs string building (already possible with `++`, so W4 can start earlier if needed).
- **W5/W6 are independent** of W2–W4 and can be done any time after W1.

Each phase follows the playbook: design note → generic prerequisite → pinning tests first → one commit per
sub-step → full gates (`gofmt`, `go vet`, `go test ./...`, `./run_dsl_tests.sh`, `-race`) → **script regression
diff against a baseline binary** (the breaking changes in W0/W2/W4 will show up there on purpose) →
docs/playbook/skills updated in the same change.

| Risk | Mitigation |
| :-- | :-- |
| Breaking changes (D2 string `+`, D9 assert, ids) | decide first (§1); the regression diff lists every affected script; a one-line migration note per change in the doc |
| Wide id migration (W4) | mechanical; do prelude + tests in one commit; keep `k8s_short` for readable output |
| Every importer of `pkg/cli` links cobra | accepted (D10); command tree isolated in `command*.go`; allowlist is `pkg/cli`-only and negative-controlled |
| Operator flag migration (`-x` → `--x`) | normaliser + deprecation notice for one release; docs/`deploy.yaml`/tests updated in the same commit |
| `let` rework (F6) | time-boxed spike before committing to the design |
| kustomize version/Go skew | D7: bump the repo to Go 1.27 (kustomize v0.21.3) as the first W5 step and re-run every gate; the toolchain is fetched automatically, CI images must carry Go ≥ 1.27 |
| Scope creep in the stdlib | tiers (§5): Tier 1 = Char/String/`not`/`Ordering`/parsers/Bytes+encodings; JSON/YAML/regex/hashes second; time later |
| Exec policy added late | keep the Appliance on **replace-only embedded rules** until W7, and never run API-submitted scripts before it |

## 11. Sources

- Elm — basic types and constrained type variables (`number`, `comparable`, `appendable`):
  [Reading Types · An Introduction to Elm](https://guide.elm-lang.org/types/reading_types.html),
  [Core Language Features](https://deepwiki.com/elm/core/2-core-language-features)
- Haskell — numeric classes (`Num`, `Integral`, `Fractional`, `Floating`, `RealFrac`), `Int`/`Integer`/`Rational`:
  [Prelude (base)](https://hackage-content.haskell.org/package/base-4.22.0.0/docs/Prelude.html),
  [A Gentle Introduction: Numbers](https://www.haskell.org/tutorial/numbers.html)
- Clojure — numeric tower, `Ratio`, throwing overflow vs auto-promoting `+'`:
  [Clojure data structures](https://clojure.org/reference/data_structures),
  [1.3 Numerics](https://dev.clojure.org/pages/viewpage.action?pageId=2097186)
- Dhall — builtin types and text functions: [Built-in types, functions, and operators](https://docs.dhall-lang.org/references/Built-in-types.html)
- Jsonnet — standard library (`std.split`, `std.format`, `std.base64`, `std.parseJson`, `std.md5`): [Jsonnet standard library](https://jsonnet.org/ref/stdlib.html)
- Gleam — prelude types, `BitArray`, stdlib modules: [Gleam language tour](https://tour.gleam.run/everything/),
  [Core data types](https://deepwiki.com/gleam-lang/stdlib/2-core-data-types)
- CLI libraries — [Go CLI library comparison](https://github.com/gschauer/go-cli-comparison),
  [How To Develop Kubernetes CLIs Like a Pro](https://iximiuz.com/en/posts/kubernetes-api-go-cli/),
  [kubectl reference docs generation](https://kubernetes.io/docs/contribute/generate-ref-docs/kubectl/),
  [spf13/cobra](https://github.com/spf13/cobra)
- Kustomize embedding — [krusty package](https://pkg.go.dev/sigs.k8s.io/kustomize/api/krusty),
  [data race with concurrent `Run`](https://github.com/kubernetes-sigs/kustomize/issues/4824),
  [no `embed.FS` support](https://github.com/kubernetes-sigs/kustomize/issues/4515)
