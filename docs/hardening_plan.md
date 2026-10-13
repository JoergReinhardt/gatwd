# Plan: Base Types & Stdlib, Core CLI, Node Identity, In-Process Kustomize, Fail-Loud Sweep

> **Status (2026-10-13): W0–W7, YAML, the time types and the follow-ups are implemented and merged (see §0); everything in the plan is done; what remains are the optional follow-ups of the handoff §0.** Progress table: §0.
> **All decisions D1–D10 are answered** (§1, 2026-10-13); **D11–D13 (the `Ratio` type, §6.1) are decided on the recommendation**. Two answers carry a point to settle when their phase starts: D1 (the
> details of the two-way `Char`/`String` conversion, W2) and D7 (auto-render, W5).
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
| W3 | **Numeric tower, specified**: logical `Int` (unbounded), `Float`, **`Ratio` (exact fractions, §6.1)**, `Complex`; `Num`/`Integral`/`Fractional` as *constraint names*; `**`, rounding, conversions | promotion exists but is undocumented and untested as a matrix | M | display fix (`+123`) |
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
| W2 tier 1, step 1 (`Char`) | **done** (2026-10-13, commit `4cc0c8a`): `functions.Char` (`type Char rune`, so it never matches the numeric `int32` arms), `'a'` literals with escapes (`parser.lexCharLiteral`, `TokChar`), `==`/ordering incl. the D1 one-character-String rule (`Compare`, `StructuralEqual`), `char_*` function forms (`dslfuncs_text.go`), the `asChar` entry of the conversion table (`valueops.go`), checker support (`Char`; `Char`/`String` are consistent), docs (`dsl_guide.md` §15, `grammar.md`). Not yet: String as a collection row, `str_*`, `not`/`xor`/`and`/`or`, `Ordering`, D2, F18 (function-form-as-step — `'a' |> char_upper` is `unknown step` until it lands), `Bytes`. Notes: `const K = 'k'` stores text like every constant (a pre-existing limitation, revisit with D2); `StructuralEqual`'s string/`%v` fallback already equated `'a'` and `"a"`, the explicit `Char` case documents D1 and survives removing that fallback. |
| W2 tier 1, step 2 (F18) | **done** (2026-10-13): function-form-as-step fallback, see §5.4. |
| W2 tier 1, step 3 (String row) | **done** (2026-10-13): the function forms `length head tail nth reverse cons member fold map filter reject flatmap to_list to_set` and `++` see a `String` as its `Char`s (`valueops.go`: `asString`, `charList`, `seqOf`, `textResult`; `functions.CharsAsString`). Closed operations answer with a `String`, `map`/`flatmap` with a `List`. **Boundary, deliberate:** the stream *steps* keep treating a `String` node as one element (a stream of lines), because changing that would rewrite every `emit`/`echo` script — pinned by `TestStreamStepsKeepStringNodesScalar`; the step form therefore differs from the function form for `map`/`filter`/`flatmap`/`fold`/`head`/`tail` on a `String` (use `str_chars` once it lands, or the function form). `==`/`<` stay strict (a `String` never equals a list of `Char`s): comparison does not *demand* a `String`. Not yet: the `str_*` library (next), `asString` is its entry point. |
| W2 tier 1, step 4 (`str_*`) | **done** (2026-10-13): `str_split/join/lines/words/chars/from_chars/trim/upper/lower/repeat/contains/starts_with/ends_with/index_of/replace/slice/pad_left/pad_right/to_int/to_float/from`, `show`, `str_format` in `pkg/cli/dslfuncs_text.go` (`asString` at every String argument; `strictInt` so a numeric String is not an Int; `maxStrLen` bounds repeat/pad). Deviations from §5.1: `str_to_int`/`str_to_float` accept decimal syntax only (no whitespace, `nan`, `inf`, hex) and return `Maybe`; `str_split ""` and `str_replace ""` are errors (they would silently split per character / insert everywhere) — use `str_chars`; a `%s` given a non-String is an error (use `%v`/`str_from`) instead of printf's `%!s(int=3)`. **Known gaps recorded:** (a) string literals had no escapes — closed by D18, step 5; (b) `str_slice -1 …` parses as a subtraction until unary minus lands in W3, so write `(0 - 1)`; (c) `show` of a `String` containing `"` did not re-lex (closed by step 5); (d) `str_pad_*` takes a `Char`, not a padding string. Not yet: regex (tier 2), `Bytes`/encodings, `not`/`xor`/`and`/`or`, `Ordering`, D2. |
| W2 tier 1, step 5 (D18 string escapes) | **done** (2026-10-13): `parser.lexEscape` shared by `Char` and `String` literals; every scanner that looks for the end of a string agrees — `unterminatedString` (multi-line statement joining, which used to count raw `"`), `unbalancedDelimiter`, `lambdaGroups`; `show` and `quoteString` emit text that lexes back to the same `String`. **Found and fixed on the way (pre-existing, silent):** a let-bound `String` was spliced into a pipeline head as a *bare word*, so `let s = "hello"` then `str_upper s` or `length s` printed nothing (`(No output …)`; `hello` was read as an undefined name) and a value with a space, quote or newline would have split the line. `Interpreter.spliceText` now splices the quoted literal (both substitution loops use it). No committed script changed meaning (regression diff empty, grep of `.gatwd` files). Known and kept: the `ebnf_meta` grammar strings and the exec-adapter's own quote splitting have their own escape handling and were not touched. |
| W2 tier 1, step 6 (Bool, `Ordering`) | **done** (2026-10-13): `not`/`xor`/`and`/`or` (strict, Bool-only, `pkg/cli/dslfuncs_bool.go`); **`Ordering = LT \| EQ \| GT` is declared first in `parser.NewParser`**, so the flags are identical in every parser and `parser.OrderingVariant` builds the very value `compare` returns; `compare` now answers with it (**breaking:** it returned `-1`/`0`/`1` — one DSL test and two Go tests migrated; `ordering_to_int` gives the old value); `ordering_from_int/reverse/then`; the constants `LT`/`EQ`/`GT` in expressions are the real Variants (`exprFallback`); `match` and the `--check` exhaustiveness warning work on it. `true`/`false` are now valid pipeline heads (`true \|> not`; they used to yield nothing). **Decided:** no prefix `!` (it collides with `!=`; the F2 row for `!true` stays an error), `Unit` is deferred (no consumer yet). **Risk recorded:** a user sum type that also declares `LT`, `EQ` or `GT` shares those constructors and `EQ`/`LT`/`GT` in an expression is the Ordering Variant (a `match` on the user's type still routes by tag). The three flags cost three of the 64 BitFlag bits. |
| W2 tier 1, step 7 (**D2**) | **done** (2026-10-13): `functions.EvalBinaryOp` rejects a `String` or `Char` operand of `+` with `operator '+' is not defined for String and Int; use ++ to concatenate` (the `"a" + "b"` corollary adds `or str_to_int / str_to_float to parse`); `functions.KindName` names types in messages; `functions.ParseNumberLiteral` (the lexer's number reader, now shared — dogfooded in `parsePrimary`), `ParseCanonicalNumber` (text that prints back unchanged), `LiteralToken`, `OperandValue`. **Consequences handled, all found by the regression diff and the tests:** (1) pipeline operator sections and `math_*` steps took their operand as *text* and relied on the coercion — the operand is now parsed as the number it spells (`evaluator.StepBinaryOp/StepMathAdd/StepMathSub`); (2) `emit 10, 20 \|> fold sum` summed text — **the unquoted tokens of `emit`, `const` and `-D` are now typed by one rule** (`LiteralToken`: a *canonical* number is a number, anything else — `alice`, `007`, `1.50` — a `String`; a quoted `const` value is always a `String`); (3) a spawn of an **undeclared** type keeps its value as text, so `examples/new_features/showcase.gatwd` now declares `type IntVal = Int` (the only committed script that needed a change); (4) three Go tests read `Appliance.GetConstant` as a string and now read the number. **Amends D4:** `-D NAME=VALUE` binds a *number* when the value is a canonical number (it said "the string"); the reason is that `-D LIMIT=10` then `x + LIMIT` inside a lambda would otherwise be an error whose only fix is `str_to_int LIMIT`, which the text-splicing of heads turns into `str_to_int 10`. **Extended the same day (user: do the operators now, leave equality for W3):** `- * / // %` and `< <= > >=` reject a `String` next to a number (`"5" - 1`, `"5" < 10` are errors naming `str_to_int`), two `String`s never do arithmetic, and two `String`s **order as text** (`Compare`: `"10" < "9"` is `true`; it was `false`, as numbers); a `Char` beside a `String` stays the D1 conversion. **Left for W3 (§6):** equality (`"5" == 5` is `true` through `StructuralEqual`'s `%v` fallback, `"10" == "10.0"` is `true` through its numeric-string branch) and the function forms that parse numeric strings (`asInt`/`ToFloat64`: `nth "1" xs`, `range "0" "2"`, `min "5" 3`, `compare "5" 3`). |
| W2 tier 2, step 1 (`Bytes`, encodings, hashes) | **done** (2026-10-13): `functions.Bytes` (`type Bytes []byte`, prints `0x…`, `AsBytes`/`ByteOf`), `Bytes ++ Bytes`/`==`, `Bytes` as a collection row of Ints in the function forms (`rowResult` generalises the String row: closed operations answer with `Bytes`), `utf8_*`, `bytes_*`, `base64`/`base64url`/`hex`/`url` encode/decode (decoders `Maybe`), `sha256`, `md5` in `pkg/cli/dslfuncs_encoding.go`. The conversion table gains a row: *Bytes wanted, String given → its UTF-8 bytes*; the reverse is never implicit. RFC 4648 and published-digest test vectors. **Deferred with reason (Mandate 5):** `yaml_parse`/`yaml_encode` — Go's standard library has no YAML and core may import only the standard library; a YAML library may not be added to core (the cobra allowlist was an explicit decision for plumbing, not for data formats). It belongs in an adapter/edition that registers it (the k8s edition already links `sigs.k8s.io/yaml` for manifests). *Alternative for you to decide:* add the YAML library to the Clean Core allowlist for `pkg/cli` (smaller, but widens core). |
| W2 tier 2, step 2 (JSON) | **done** (2026-10-13): `json_parse` (`Maybe`; trailing data is malformed), `json_encode` (sorted keys, no HTML escaping, errors name the path), `json_pretty`, `is_null` in `pkg/cli/dslfuncs_json.go`; the standard `encoding/json` only. `functions.KindName` now names the DSL types (`Pair`, `Set`, `Function`, …). **Known:** duplicate object keys — the last wins (Go's behaviour); `null` is the Go nil, so `json_parse "null"` is a `Just` of nil. |
| W2 tier 2, step 3 (regex) | **done** (2026-10-13): `re_match`, `re_find`, `re_find_all`, `re_groups`, `re_replace`, `re_split`, `re_quote` (Go RE2, `pkg/cli/dslfuncs_regex.go`); compiled patterns are cached (bounded at 256) and a pattern is limited to 4096 bytes; a test pins linear time on `^(a+)+$` over 200 000 characters. **Two pre-existing defects found and fixed on the way:** (a) a **String constant** (`const`, `-D`) was spliced into a pipeline head as a bare word, so `-D NAME=bob` then `str_upper NAME` printed nothing — now `constSplice` quotes it (same fix as let-bound Strings, `spliceText`); (b) the multi-line statement joiner, the pipeline splitter (`splitPipelineOps`) and the `match`-arm splitter (`splitCasesAware`) counted `(`, `[`, `{` and `|>` **inside string literals**, so a pattern like `"("` or `"[(]"` swallowed the following lines or hid the `|>` — `scanLine` and `copyLiteral` now treat String and `Char` literals as text (`unterminatedString` is a view of `scanLine`). Left: the step-argument `Constants` splice (`parsePipelineSteps`, line-level text for exec steps) still uses bare text. |
| W2 | **done** except what is deferred: `Duration`/`Time`/`Date` (designed in §5.6, D14–D17 **accepted as proposed 2026-10-13**, scheduled after W3), YAML (**decided 2026-10-13: a YAML library joins the Clean Core allowlist for `pkg/cli`; done 2026-10-13**: `yaml_parse`/`yaml_parse_all`/`yaml_encode`/`yaml_encode_all`, `go.yaml.in/yaml/v3`, `pkg/cli/dslfuncs_yaml.go`, semantics in `docs/dsl_guide.md`; `Maybe` for malformed text, dates/binary kept as text, merge keys and aliases expanded with a node budget, YAML 1.1 boolean words quoted on output; guards negative-controlled, allowlist scope proven by importing it from `functions/`), `Unit`, prefix `!` (**decided 2026-10-13: add it as a synonym for `not`, in Track A**). Answers: §1 "Answered at the start of the parallel run". |
| W3 (numeric tower, `Ratio`, prefix `!`) | **done and merged** (2026-10-13, Track A, `d20dc21`): unary minus and `**`, the logical `Int` (demotes; F11 fixed), exact `/ // %` (`//` floors, `%` takes the divisor's sign), the numeric functions, strings are not numbers (`"5" == 5` is `false`), constraint names in the checker, `Ratio` (`3r4`, `ratio`, `to_ratio*`; D11–D13), prefix `!` (a synonym of `not`; no collision with `!=`), the operator × type matrix test. Notes and open questions: §6.2. |
| W4 (qualified node ids, D6) | **done and merged** (2026-10-13, Track B, `a09a901`): `Kind/namespace/name` ids, loud duplicates (`Graph.InsertNode`, `graph_insert_node`), F13/F14 fixed, `collect_list` passes errors on, every test/fixture/example migrated. Notes: §7.1. |
| W5 (in-process kustomize, Go 1.27) | **done and merged** (2026-10-13, Track C, `0f6c6db`): `go 1.27.0`, kustomize v0.21.3 in `pkg/k8sgraph/kustomize.go`, `manifests` auto-renders, `--no-kustomize`. Notes: §8.1. |
| W6 (operator: layering, exit codes, formats, lazy client) | **done and merged** (2026-10-13, Track D, `9e8fde6`): layered overrides by default (D5), `*LoadError`, `NewApplianceTree`, `--fail-on-violation`, `--format json`, `LazyDynamicClient`. Notes: §9. |
| YAML | **done** (`fca8917`), see the W2 row. |
| Operator follow-ups (Track E) | **done and merged** (2026-10-13): exit status 3 for violations under the opt-in `--fail-on-violation`, the operator's `--no-kustomize`, source provenance (`collections.Source` side channel, `with_source`/`source_of`) so duplicate-id errors name both files. Notes: §7.1, §9. |
| Dependencies | **done** (`3e4927b`): `client-go`/`apimachinery`/`api` v0.38.0-alpha.2, the kube-openapi `replace` removed. |
| String vs collection equality (Track G, `0d30e85`) | **done and merged**: canonical, collision-free HAMT keys (`evaluator/hamt_key.go`), the printed-form fallback removed from `StructuralEqual`. Notes: §6.3. Open: `to_set`/`distinct` steps still deduplicate by printed form (recommended: make them value-typed). |
| §5.6 `Duration`/`Time`/`Date` (Track F, `51830ae`) | **done and merged**: three core types, unit/`@` literals, arithmetic, function forms, `now` through an injectable clock, checker support, JSON/YAML as strings; time keys encoded by UTC instant in `EncodeKey`. Notes: §5.6; open T1-T3 there. |
| W7 execution policy (Track H, `84a5c85`) | **done and merged**: `cli.Policy`, capability gates (exec, fs, net; env has no builtin yet), `RunOptions.Source/Policy/Context/MaxOutput`, operator `--allow-*` flags, deny-by-default overrides, `diagnostics` in the JSON report. Notes: §9. Deferred: `parser.FormatError` wiring, operator timeout/output-limit flags. |
| Closing checklist (§10.1) | **done and merged** (2026-10-13): every row marked done with a commit or deferred with a reason; the how-to is pinned by `cmd/gatwd-k8s-operator/doc_test.go`. |
| Follow-ups after the checklist (Tracks I, J, K) | **done and merged** (2026-10-13): typed `Duration`/`Time`/`Date` tokens for `emit`/`const`/`-D` (§5.6.2); value-typed `to_set`/`distinct` on the new `collections.Setoid` (§6.3a); operator `--timeout`/`--max-output` with failure kinds `timeout`/`output_limit` (§9). |
| Merged-tree gates (2026-10-13) | `go vet`, `go test ./...`, `./run_dsl_tests.sh`, `-race`, silent `--check` all green; script regression diff against `2526cc0`: only explained differences (ids, `collect_list` now loud, the big-integer `+`, the cheatsheet's new step; the visualizer tours and `tour_23` differ between runs of the same binary). |

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

**Confirmed by the user (2026-10-13, after W2 step 1):**

| ID | Decision | Notes |
| :-- | :-- | :-- |
| **D2** (corollary) | **`"a" + "b"` is an error** with a hint to use `++` (the operator's message names `++`). | Applies when D2 is implemented (W2, after `str_*`). Check first how `const`s (stored as text) behave under a numeric-only `+`. |
| **D1** (details) | **Accepted as proposed:** a `List`/`Chunk` of `Char`s counts as a serialized string; a mixed collection (`['a', 1]`) is an error naming the element; an empty collection converts to `""` only where a `String` is demanded (annotation or string function). | The table is in `pkg/cli/valueops.go` (`asChar` is its first row). |

**Decided on the recommendation (2026-10-13, the user: "do it like you recommend and propose in both cases"; §6.1):**

| ID | Decision | Alternative, kept as a written proposal |
| :-- | :-- | :-- |
| **D11** | `/` on two `Int`s keeps today's meaning (exact when divisible, else `Float`); exactness is opt-in through the `Ratio` literal `3r4` or `ratio n d`. | Scheme/Clojure-style exact `/` for `Int`/`Int` (breaking: `7 / 2 == 3.5` is pinned in docs and tests). |
| **D12** | `Ratio` ⊕ `Float` arithmetic promotes to `Float`; `==`/`<` are **exact** (the `Float` is compared by its exact value), so `1r10 == 0.1` is `false` and `to_ratio 0.1 == 1r10` is `true`. | (a) compare as doubles (`1r10 == 0.1` is `true`, like `Int == Float` today; `==` stops being transitive); (b) make `Ratio` vs `Float` comparison an error that demands `to_float`/`to_ratio`. |
| **D13** | `%` stays **modulo**. | Haskell-style `%` as the fraction operator: `1 % 3` is `1r3` for `Int`/`Ratio` operands; modulo moves to a function `mod a b` (also `rem`). Cost: three script lines use `%` as an operator (`examples/scripts/cheatsheet.gatwd:44,105`, `examples/game_of_life_pure_dsl/game_of_life.gatwd:24`), plus docs and tests. Cheap to add later *alongside* `3r4`; ask again if wanted. |

**Decided on the recommendation (2026-10-13, "follow recommendation for D18"):** **D18** — escapes in `"…"` literals (`\n \t \r \0 \\ \' \" \xHH \uHHHH`, one table shared with `Char`; an unknown escape is an error). Implemented (W2 step 5, below). The recorded alternative was a raw-string prefix only.

**Decided later the same day (2026-10-13):**

| ID | Decision | Notes |
| :-- | :-- | :-- |
| **D1** | **`Char` is added; `String` is a collection row of `Char`s; conversion works both ways, on demand.** The user's words: *"convert to char on demand if possible, otherwise yes"*, clarified the same day: *"should work both ways: i.e. if a serialized collection of chars gets treated as a string, convert it"*. | **Rule:** wherever a `String` is expected, a collection of `Char`s (a `List`/`Chunk` of `Char`, e.g. the result of `map`/`filter` over a string) converts to the `String`; wherever a `Char` is expected, a one-character `String` converts; wherever a collection of `Char`s is expected, a `String` converts to one. Otherwise a real distinct `Char`. One narrow, tested conversion table (D2 removed implicit `+` coercion for the opposite reason). **To settle in the W2 design note:** which carriers count as "serialized" (`List`, `Chunk`, `iter.Seq` of `Char`), what a mixed collection (`['a', 1]`) does (an error naming the element), and the empty collection (`[]` is not a `String` on its own — convert only when the context demands a `String`, as an annotation or a string function does). |
| **D3** | **(a)** keep the runtime representations; specify the tower; `Num`/`Integral`/`Fractional` are annotation-level constraint names (no boxed `Num`). | W3 as written in §6. |
| **D7** | **Bump the repo to Go 1.27** and use `sigs.k8s.io/kustomize/api` **v0.21.3** (not the v0.21.0 pin). | The user also asked whether system Go must be updated: **no** — `GOTOOLCHAIN=auto` (the default; this machine has system Go 1.26.0 and the toolchain cache) downloads `go1.27.x` (1.27.2 is on proxy.golang.org) when `go.mod` says `go 1.27.0`. Only CI images, Dockerfiles and other contributors need Go ≥ 1.27 or auto-toolchain with network access (the repo has no Dockerfile or CI file that pins a Go version). **Not answered explicitly:** auto-render of kustomization roots with `--no-kustomize` — proceed on the recommendation at W5 and confirm. |
| **D8** | **Yes** — unknown pipeline step ⇒ OS command stays for trusted CLI use; **denied by default** for Appliance overrides and any script not given on the command line (capability policy, §9). | W7. |

**Answered at the start of the parallel run (2026-10-13; the user's answers to the handoff's open questions Q1–Q6):**

| ID | Decision | Notes |
| :-- | :-- | :-- |
| **Q1** | **Push now.** The W2 commits were pushed (`19a4b03..9373f01`). | Later pushes only when asked. |
| **D14–D17** | **As proposed** (§5.6): `@…` and unit-suffix literals; `now` as an injectable-clock adapter step; `Time` carries a display zone, no `DateTime`; `Date ± whole-day Duration`. | Runs after Track A (W3) lands. |
| **YAML** | **Add a YAML library to the Clean Core allowlist for `pkg/cli`** (user: "add to core, will help visualizer later"). `yaml_parse`/`yaml_encode` become core builtins instead of an Edition function. | Amends D10's allowlist (cobra/pflag only) in `internal/archtest/clean_core_test.go` and the Mandate 5 text. Done **after** the Go 1.27 bump of Track C has merged (both touch `go.mod`). Library to be chosen then (prefer one already in the module graph via client-go). The visualizer is a Layer 4 package; it would still import only what its layer rule allows — the allowlist change names `pkg/cli` only until a second consumer needs it. |
| **D7** (auto-render) | **Auto-render** kustomization roots in `manifests`; `--no-kustomize` turns it off. | W5. |
| **Prefix `!`** | **Add it as a synonym for `not`**, *if it does not break anything* (user: "decide"). Decision: do it in Track A next to unary minus (both are prefix operators in the expression parser). `!=` stays the inequality token; `!` is a prefix operator only when the next character is not `=`. | Track A's last parser sub-step. Needs rows `!true`, `!(a == b)`, `a != b` (unchanged), `! =` (still an error). If a pinning test shows a real collision, defer again and record why. |
| **Q5** (rest) | **Keep:** `Unit` deferred; `LT`/`EQ`/`GT` reserved by `Ordering`; `compare` returns `Ordering`; typed `-D` numbers. | |
| **Q6** | **Run the four tracks (W3, W4, W5, W6) in parallel** in their own worktrees; merge one at a time in the order C, D, A, B. | Handoff §3a. |

**Answered after the parallel run (2026-10-13; the user's answers to the second round of questions, defaults where the user said "keep default"):**

| ID | Decision | Notes |
| :-- | :-- | :-- |
| **Push** | Pushed (`9373f01..24ef7c8`). | |
| **Remote kustomize bases** | **Stay refused** ("keep as is"). | Revisit with W7's capability policy. |
| **kube-openapi `replace`** | **Update `client-go`/`apimachinery` to v0.38 (alpha) and drop the `replace`.** | **Done** on `master` (`client-go`/`apimachinery`/`api` v0.38.0-alpha.2, `replace` removed); see §8.1 deviation 1. |
| **Exit codes** | **Keep the default:** `--fail-on-violation` stays opt-in; violations get their own exit code **3** (1 = a broken run, 2 = a malformed command line). | Operator follow-up, §9. |
| **Source provenance** | **Yes:** the manifest adapter attaches the source file to a document and a duplicate-id error names it. | W4 follow-up, §7.1. |
| **`ownerReferences` without `kind`** | **Keep strict** (an error). | |
| **String vs collection equality** | **Yes, fix:** HAMT keys must stop depending on the printed form of a collection, then `"[2 1]" == [2, 1]` becomes `false`. | A small phase after §5.6 (D2 completeness). |
| **`2 ** -3`** | **Keep the default:** a `Float`; an exact result needs an explicit `ratio`. | §6.2. |
| **`n -1`** | **Keep:** an application error with a hint. | §6.2. |
| **W7 design points** | **Taken on the recommendation** (the user said "continue" after the three points were put to them, 2026-10-13): (1) the policy gates only capabilities reachable from script text — the unknown-step OS-command fallback and file/network/env builtins; steps an Edition wires itself (`manifests`, the cluster crawl) stay allowed; (2) the operator opts in with `--allow-exec/--allow-fs/--allow-net/--allow-env`, all off by default, and a ConfigMap can never widen them; (3) diagnostics keep the free-text `[ERROR] …` and add structured fields (kind, step, capability) to the JSON `errors` entries. | Reversible while W7 is unmerged; veto at any time. |

**Answered after the closing checklist (2026-10-13; the user's answers to handoff §0 Q1–Q11, the rest on the recommendation):**

| ID | Decision | Notes |
| :-- | :-- | :-- |
| **Push** | Pushed (`24ef7c8..66a2abe`). | |
| **T1 typed tokens** | **Yes:** `const X = 30d`, `emit 5s`, `-D X=5s` bind a real `Duration` (and `@…` a `Time`/`Date`), by the same canonical-literal rule as numbers. | Track I. |
| **T2 `Duration` in JSON/YAML** | **Keep the literal text** (`"1h30m"`). | No change. |
| **T3 `strptime`-style parsing** | Deferred until a rule needs it. | |
| **`to_set` / `distinct`** | **Value-typed** (no printed-form dedup). | Track J. |
| **Override `import` without `--allow-fs`** | **No**, stays strict. | No change. |
| **`env` builtin** | Only when something needs it. | No change. |
| **Operator timeout / output limit** | **Yes:** `--timeout`, `--max-output`, exit 1 on timeout. | Track K. |
| **Age/expiry rule in the embedded rules** | **Stays an example override.** | No change. |
| **Secret `data` redaction** | **Document, do not redact** ("kubeconfig/RBAC on the user side will restrict it when necessary"). | how-to §2.4 Security. |
| **Third message element in rule results** | Deferred. | |

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
| F18 | `length` has no step form (function forms in general are not usable as steps) | probe | **fixed** (W2 step 2): generic builtin-as-step fallback (§5.4) |

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
- **Two-way conversion (D1):** a collection of `Char`s used as a `String` becomes that `String` (`str_upper (map char_upper "abc")` works), and a `String` used as a collection of `Char`s becomes one. Table lives in `valueops.go`, one place, listed in the doc.
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

**Done (W2 step 2, 2026-10-13):** `Interpreter.builtinStep` in `pkg/cli/closures.go`, hooked in after the plugin registry
check and before `GenericExecAdapter`. The semantics differ from the let-bound-closure step in one deliberate way: a
builtin is applied to the **whole incoming value** (`[1,2,3] |> length` is `3`), never mapped over its elements; a
bare arity-1 *let-bound* closure keeps its map-over-the-value meaning. Tests: `builtinstep_test.go` (step ≡ function
table, whole-value, existing steps and unknowns unchanged, **every registry name resolves as a step**). Known and kept:
(a) the pre-existing stream steps `head` and `tail` operate on stream *nodes*, so `[3,1,2] |> head` is not `head [3,1,2]`
(the step form and function form differ for a reason; rename or document at the matrix review); (b) inside a
pipeline-bodied lambda the input's elements stream, so write `\xs -> length xs`; (c) an OS command named like a builtin
is now shadowed by the builtin — today only `compare` (ImageMagick) is affected; (d) `abs -3` parses as `abs - 3` until
unary minus lands in W3.

### 5.5 Later tier

`Duration`, `Time` and `Date` (Go `time`; **designed in §5.6**), bit operations, n-ary tuples. (`Ratio`, exact fractions, is in W3: §6.1.)

**DoD (W2):** each type/library row has structural Go tests, DSL assertions (output verified non-empty),
`Example*`, docs in `dsl_guide.md`/`grammar.md` (new literal), checker support for `Char`/`Bytes`/`Ordering`, and a
negative case per partial function.


### 5.6 `Duration`, `Time` and `Date` (proposed; scheduled after W3, because `Duration / Duration` uses `Ratio`)

Three distinct scalar types over Go's `time` (standard library only, Mandate 5), in `functions/`, each with its own
arithmetic, ordering and function forms. They are separate because they answer different questions: *how long*
(`Duration`), *which instant* (`Time`), *which calendar day* (`Date`). Mixing them silently is the classic bug (adding
24 h is not "tomorrow" across a daylight-saving change), so the language keeps them apart and makes the conversions
named, explicit functions.

| Type | Representation | Meaning | Prints as (and re-lexes from) |
| :-- | :-- | :-- | :-- |
| `Duration` | `functions.Duration` = `type Duration time.Duration` (int64 nanoseconds; a named type, so it never passes for an `Int`) | a signed length of elapsed time, about ±292 years | `1h30m`, `250ms`, `-5s`, `0s` |
| `Time` | `time.Time` (an instant; Go also keeps a location for display) | a point on the timeline | RFC 3339, `2026-10-13T09:30:00Z` |
| `Date` | `functions.Date{Year, Month, Day}` (proleptic Gregorian, no zone, comparable) | a calendar day | ISO 8601, `2026-10-13` |

**Literals (D14, proposed).**

| Form | Value |
| :-- | :-- |
| `5s`, `250ms`, `1h30m`, `2d`, `1w` | `Duration`. Units: `ns us µs ms s m h` as in Go's `time.ParseDuration`, plus `d` (= 24 h) and `w` (= 7 d), **fixed lengths, not calendar days**. A number *immediately* followed by a unit (no space) and then a non-identifier character lexes as one `Duration` token; `5 s` is still an application. Segments chain (`1h30m15s`); a fractional first segment is allowed (`1.5h`); a unit used twice, an unknown unit or an overflow is a lex error. |
| `@2026-10-13T09:30:00Z`, `@2026-10-13T11:30:00+02:00` | `Time` (RFC 3339; a missing zone is an error, never "local") |
| `@2026-10-13` | `Date` |
| `time "…"`, `date "…"`, `duration "…"` | the same as function forms returning `Maybe` for computed text (`time_parse`, `date_parse`, `dur_parse`); malformed text is `Nothing`, never an error |

The `@` prefix marks a time-like literal and cannot start anything else today (`@` is unused in the lexer).
**Alternative recorded:** functions only (no `@`, no unit suffix) — smaller lexer change, but `now - @…` and
`timeout > 30s` are the readable forms this type family exists for.

**Arithmetic and ordering (the matrix; everything else is an error naming both types).**

| Operation | Result |
| :-- | :-- |
| `Duration ± Duration` | `Duration` |
| `Duration * Int`, `Int * Duration`, `Duration * Float`/`Ratio`, `Duration / Int`/`Float`/`Ratio` | `Duration`, **truncated toward zero** to whole nanoseconds (as Go's `time.Duration` does); `/ 0` is an error |
| `Duration / Duration` | **exact `Ratio`, demoted to `Int` when whole** (`90m / 1h` is `3r2`, `2h / 1h` is `2`); `to_float` for the double. Before W3 lands: `Float`. |
| `Duration // Duration`, `Duration % Duration` | floor quotient (`Int`) and remainder (`Duration`) |
| unary `-`, `abs` | `Duration` |
| `Time + Duration`, `Time - Duration` | `Time` |
| `Time - Time` | `Duration` (an error if it overflows int64 ns) |
| `Date ± Duration` | `Date` **only when the duration is a whole number of days** (`2d`, `-1w`); `Date + 90m` is an error saying so. `Date - Date` is a `Duration` (a multiple of `1d`). |
| `==`, `!=`, `<`, `<=`, `>`, `>=`, `min`, `max`, `compare` | within a type only; `Time` compares by instant (`Equal`), not by location; **`Date` vs `Time` is an error** (no implicit zone) |
| `Time + Time`, `Date + Date`, `Time * n`, … | errors (not meaningful) |

Overflow is an error (Mandate 6), never wrap-around.

**Functions** (data-last, curried, `Maybe` for malformed text, error values for impossible requests):
- *Duration:* `dur_ns d`, `dur_ms d`, `to_seconds d` (`Float`), `dur_from_seconds x`, `dur_parse s`, `dur_truncate unit d`, `dur_round unit d`.
- *Time:* `time_parse s`, `time_format layout t` (a documented strftime subset `%Y-%m-%d %H:%M:%S %z %Z`, plus the names `rfc3339`, `date`, `clock`; Go's reference-time layouts are not exposed), `time_unix t`/`time_from_unix n` (seconds), `time_year/month/day/hour/minute/second/nano t`, `time_weekday t` (ISO `1` = Monday … `7`), `time_in zone t` (IANA name; an unknown zone is an error), `time_truncate d t`, `time_round d t`.
- *Date:* `date t`-style constructors `date_of y m d` (an invalid day such as 2026-02-30 is an error), `date_parse s`, `date_format layout d`, `date_year/month/day`, `date_weekday`, `date_add_days n d`, **`date_add_months n d`** and `date_add_years n d` (the day clamps to the end of the month: 31 Jan + 1 month is 28/29 Feb, as in Java's `plusMonths`), `date_days_in_month d`, `is_leap_year y`, `date_of_time zone t`, `time_of_date zone d` (midnight in `zone`; a midnight skipped by a daylight-saving gap resolves to the first valid instant after it, documented).
- *Common:* all three work with `map`/`fold`/`filter`, sort with `compare`, and print through `show`/`str_from`.

**`now` is not a pure function (D15, proposed).** The clock is a side effect, so it lives where side effects live
(Mandate: effects isolated in adapters): `now` is a **stream-source step** reading an injectable
`Interpreter.Clock func() time.Time` (default `time.Now`), exactly like `scan_file` reads a file. Everything above is pure;
tests set a fixed clock, so no test sleeps or depends on the wall clock. A `sleep d` adapter step is deliberately
**not** part of this tier (deferral: blocking steps need the W7 capability policy first).

**Zone model (D16, proposed).** A `Time` is an instant that *remembers a display location* (Go's model): `time_in
"Europe/Berlin" t` is the same instant rendered in another zone, `==` ignores the location, calendar fields
(`time_hour`, `date_of_time`) use it, and a parsed RFC 3339 offset sets it. There is **no separate `DateTime` type**
(a zone-less wall-clock time); that is the usual source of bugs and is a recorded deferral (add it only when a real
need appears). The default location is UTC; local time is never implicit.

**Date arithmetic with Durations (D17, proposed):** `Date ± whole-day Duration` is allowed (above) because it makes
`today + 7d` and `deadline - today` readable and exact; **alternative:** `Date` only through `date_add_*` functions and
no operators. Calendar months and years always go through functions, since they are not fixed lengths.

**Where it plugs in (Mandate 2 cells):** the types and helpers in `functions/time.go` → `EvalBinaryOp`/`Compare`/
`StructuralEqual` arms (before the numeric arms, so a `Duration` is never read as an `int64`) → the `@`/unit lexing
in `parser/infix.go` (`TokDuration`, `TokTime`, `TokDate`) → function forms in `pkg/cli/dslfuncs_time.go` → the `now`
adapter step and `Interpreter.Clock` → `tyDuration`/`tyTime`/`tyDate` in the checker (operator results come from the
runtime matrix, so `Duration * Int` is typed for free; `Comparable ⊇` all three) → DSL assertions inside lambdas and
comprehensions → structural Go tests on the raw types → `Example`. The text-substitution guard `scalarValue` treats all
three as non-scalar. A `Duration` also joins the `Num`-like constraint list only as `Comparable`, never as `Num`.

**Negative cases (each must be seen red first):** `Date + 90m`, `@2026-02-30`, `@2026-10-13T09:30:00` (no zone),
`Time + Time`, `Date < Time`, `1h / 0s`, `time_in "Nowhere/Land" t`, `1h1h`, `5q`, `9223372036854775807ns + 1ns` (overflow),
`date_of 2026 13 1`. **Property tests:** `(t + d) - t == d`, `date_add_months` clamps and never skips a month,
`date_parse (date_format "%Y-%m-%d" d) == just d`, `time_parse (time_format rfc3339 t) == just t`, DST boundaries
(`Europe/Berlin` 2026-03-29 and 2026-10-25) for `time_in`, `time_of_date` and `+ 24h` versus `date_add_days 1`.

**Motivating use (generic, not domain code):** resource ages and expiry — `now - (time_parse created)` is a `Duration`, a
rule such as `age > 30d` or a certificate `expires - now < 14d` reads as written; the Kubernetes prelude may use
it from `.gatwd`, core stays domain-free. **Size:** about two days inside the later tier (the three lexer tokens and the
`strftime` translation are the long poles).

**Proposed decisions, unanswered:** **D14** the `@` and unit-suffix literals (vs functions only); **D15** `now` as an
injectable-clock adapter step; **D16** `Time` carries a display location, no `DateTime` type; **D17** `Date ± whole-day
Duration` operators. Each recommendation is the first option named; ask before implementing this section.

#### 5.6.1 Closure notes (Track F, 2026-10-13; D14–D17 accepted as proposed)

**Done, one commit per sub-step, all gates green after each** (gofmt, `go vet`, `go test ./...`, `./run_dsl_tests.sh`,
`-race` on `./pkg/... ./evaluator ./parser ./interactive`, silent `--check`, script regression diff with the base
binary run twice: empty apart from the visualizer tours 15/20/23, which differ between runs of one binary):

| Step | Commit | What |
| :-- | :-- | :-- |
| 1 `Duration` | `452a99f` | `functions.Duration` (`type Duration time.Duration`), `ParseDuration`, the `TokDuration` lexer token (`5s 1h30m 1.5h 2d 1w`), canonical printing, the operator matrix in `functions/timeops.go` (`± * / // %`, `Duration / Duration` = exact `Ratio`), unary `-`/`abs`, overflow errors; `sort`'s `DefaultLess` now orders numbers and time values by value |
| 2 `Time` | `75d5ae0` | `functions.Time` (struct around `time.Time`: RFC 3339 printing, years 1–9999, monotonic part stripped), `@…T…Z`/`@…+02:00` literals (`TokTime`), `Time ± Duration`, `Time - Time`, comparison by instant, embedded zone database (`time/tzdata`), DST pinned for `Europe/Berlin` 2026-03-29 and 2026-10-25 |
| 3 `Date` | `a5b7249` | `functions.Date` (comparable struct), `@2026-10-13` literal (`TokDate`), D17 (`Date ± whole-day Duration`, `Date - Date`), `AddDays/Months/Years` with end-of-month clamping, `Date` vs `Time` is an error |
| 4 function forms | `21eebd9` | `dur_*`, `time_*`, `date_*`, `is_leap_year`, `date_of_time`, `time_of_date` (`pkg/cli/dslfuncs_time.go`, helpers in `functions/timefmt.go`); strftime-subset layouts; `Maybe` parsers |
| 5 `now` | `aa860ef` | `adapters.ClockAdapter`, the `now` step and a bare `now` head, `Interpreter.Clock` (inherited by forks and pipeline lambdas) |
| 6 checker, matrix, encoding | `f895387`, `982631d` | `Duration`/`Time`/`Date` known to `--check`; `CanonicalKey`/`KeyText` (a `Time` is its instant as a Set member and HAMT key, in function and step forms); `json_encode`/`yaml_encode` |
| 7 docs | this commit | `dsl_guide.md` ("`Duration`, `Time` and `Date`"), `grammar.md`, godoc (`functions`, `adapters`, `pkg/cli`), `Example*` |

**Decisions taken on the recommendation (reversible, say if you want the alternative):**
1. **Encoding.** `json_encode`/`yaml_encode` write a `Time` as RFC 3339, a `Date` as ISO 8601 and a `Duration` as its literal
   text (`"1h30m"`), all Strings; before this they failed with "has no JSON form" (never silent). Alternatives for `Duration`:
   ISO 8601 (`PT1H30M`) or integer nanoseconds. Decoding leaves text as text (the YAML-date stance); `time_parse`/`date_parse`/`dur_parse` revive it.
2. **`dur_truncate`/`dur_round` take a `Duration` as the unit** (`dur_round 1m d`), the same argument shape as `time_truncate d t`;
   the plan wrote `unit` without saying whether a name string or a `Duration`.
3. **`time_truncate`/`time_round` count from `0001-01-01T00:00:00Z`** (Go's rule), so `1d` is UTC midnight whatever the display zone; use `time_of_date zone (date_of_time zone t)` for the zone's midnight.
4. **A `Time` Set member / HAMT key is its canonical UTC form** (`==` and membership agree); the display zone is not kept in a `Set`.
5. **`time_in ""` and `time_in "Local"` are errors** (local time is never implicit); `time_of_date` resolves a midnight lost to a daylight-saving gap to the first valid instant after it and a repeated midnight to the earlier one (tested with `Asia/Beirut` 2026-03-29 and `America/Havana` 2026-11-01).
6. **Range:** `Time`/`Date` years 1–9999 (what a four-digit RFC 3339 year reads back); `Duration` about ±292 years; anything beyond is an error.
7. **`now` is not a function form**: it is a step (and a bare head); inside a lambda take the reading as a parameter (`\now exp -> exp - now < 14d`). A user binding named `now` shadows it.

**Deferred, with reason:** (a) **typed tokens** — `const X = 30d`, `emit 5s`, `-D X=5s` keep the *text*, because `LiteralToken` types only canonical numbers and a Duration's canonical form is not how people write it (`30d` prints `720h`); use `let limit = 30d` or `dur_parse "30d"` (documented in the guide); open question below. (b) `DateTime` (D16) and a `sleep` step (needs W7's policy). (c) ISO 8601 durations (`PT1H`) and `Time`/`Date` parsing of other layouts: `strptime`-style parsing is not provided, only RFC 3339 and `YYYY-MM-DD`. (d) The k8s prelude does not use the new types yet (§10.1 row added). (e) Weekday names, week numbers, `time_format` locale names.

**Behaviour changes to know:** a number immediately followed by letters is now one token, so `5s` is a `Duration`, and `5q`/`5min`/`3x` are lex errors naming the unit (they were a number applied to an identifier, `cannot apply non-function value`, an error before as well; the regression diff over every committed script is empty); `5 s` is unchanged. `sort` (`evaluator.DefaultLess`) now orders any two numbers (`int8` vs `int64`, big integers, `Ratio`) and time values by value instead of by printed text. The binary grows by about 0.5 MB (embedded zone database).

**Open questions:** **T1** make `const`/`emit`/`-D` tokens typed for `Duration`/`Time`/`Date` literals (`const MAX_AGE = 30d`)? Recommended: yes for `Time` and `Date` (`@…` is unambiguous) and for a Duration whose *value* is written (not only canonical text), at the price of changing how a printed value looks (`30d` shows `720h`). **T2** keep the `Duration` JSON form as the literal String or switch to ISO 8601? **T3** add `str`-style parsing with a layout (`time_parse_with "%d.%m.%Y" s`) when a real need appears?

#### 5.6.2 T1 closure notes (Track I, 2026-10-13; user: yes, by the same canonical-literal rule as numbers)

`const X = 30d`, `emit 5s`, `-D X=5s` and `@2026-10-13[T…]` now bind a real `Duration`/`Time`/`Date`. One rule, one function:
`functions.LiteralToken` (a canonical number, then `ParseCanonicalDuration`, then `ParseCanonicalTemporal`, else the text).
Commits: `4289c3e` (the rule and its Go tests), `07c2ef8` (wiring, splicing, tests), the docs commit after them.

**The Duration decision (the open point in the brief).** A number is canonical when it *prints back unchanged*. A Duration cannot
meet that test as stated (`90m`, `1h30m` and `5400s` are one value; `30d` prints `720h`), and requiring it would leave exactly
the tokens people write (`30d`) as text. The rule chosen never rewrites the *spelling of the number*, and lets the *unit* be
free: a token is a typed Duration when `ParseDuration` accepts it **and** every numeric segment is written canonically (no
`+` sign, no leading zero, no trailing zero in a fraction, no `-0s`). So `5s`, `90m`, `1h30m`, `30d`, `-2w`, `1.5h`, `0s` are
Durations; `007s`, `1.50h`, `+5s`, `0.0s`, `-0s`, `5q`, `5` (no unit) stay Strings. This is also the only choice that never
contradicts the expression lexer: every typed token lexes as the same Duration inside an expression (the lexer is more
liberal, `007s` is 7 s there); the token rule only declines to guess for spellings it would normalise. `Time` and `Date`
keep the literal test: typed only when the value prints back as written (`@2026-10-13T10:00:00+00:00` prints `…Z`, `.250Z`
prints `.25Z`, so they stay Strings; a bad calendar day such as `@2026-02-30`, a missing zone or a missing `@` stays text,
never an error). **Cost:** a typed Duration prints in its canonical chain (`const X = 30d` shows `720h`).

**What changed besides the rule.**
1. `const`: a bare time token is typed through the existing expression path (`@2026-10-13` holds a `-`, so it was already
   evaluated, then turned into text; now the value is kept); `const D = 1h + 30m` is a `Duration`; a *quoted* value that would
   evaluate to a time value stays the text written. A const value with an operator character is an expression (`const P = +5`
   is `5`, so `const P = +5s` is a `Duration`; unchanged for numbers).
2. Splicing: a time constant is spliced into a pipeline head, a step and `spawn T: C` as its **literal** (`functions.TemporalLiteral`:
   a `Time`/`Date` prints without its `@`, which would re-lex as a subtraction); `scalarValue` (used for `let` variables) is
   unchanged, `spliceableConst` adds the three types. The regression diff shows no script using them.
3. Operator sections take the same literals as operands (`|> + 5s`, `|> - 7d`, `|> + AGE`): `OperandValue` reads a Duration/
   Time/Date token as the lexer does (it already accepted non-canonical numbers such as `007`). Before, the operand stayed text
   and `emit 5s |> + 5s` failed with `Duration and String`.
4. **Found on the way (pre-existing, fixed):** a **String** constant was spliced *unquoted* into a lambda step, so
   `const Q = "5"` then `\n -> Q == "5"` was `false` (the lambda saw the Int `5`) and, with the Duration lexer, `const Q = "30d"` read
   as a Duration there. Inside a lambda step (`isLambdaStep`: a `\` and a `->`) a String constant is now spliced quoted. Raw
   step arguments (`walk_fs DIR`, `take N`) are unchanged: they are not expression text.
5. `hamt_get`/`hamt_set`/`hamt_del` step forms read a bare time token as the value it is (`hamt_get 5s`), via `hamtStepKey`'s
   `LiteralToken`; `hamt_get @2026-10-13` was `key not found` before.
6. Checker: no code change; it already typed a bare `const` through the expression reader (a Duration is a `Duration`, a quoted
   `"30d"` a `String`), now pinned by `TestCheckerTypesTemporalConstants`. `--check` stays silent on every committed script.
7. Docs: `docs/dsl_guide.md` (§6 literal tokens, the time section's "Not yet", the `-D` flag), `docs/grammar.md` (`LiteralToken`),
   `docs/k8s_howto.md` (the two places that said `-D`/`const` stay text), the `-D` flag help, godoc.

**Tests.** `functions/temporal_literal_test.go` (the rule, canonical and non-canonical spellings, round trip through
`TemporalLiteral`), `pkg/cli/typed_tokens_test.go` (const kinds, quoted/non-canonical stay String, lambda use, emit kinds,
`-D`, sections, statement-path lambda splice, the checker), DSL section 115 in `tests/dsl_exhaustive_test.gatwd` (everything
inside lambdas or per element). Negative controls, each failing before being restored: dropping the zero check or the
print-back check in the canonical readers, ignoring the quoted case, `%v` instead of the literal at the step splice, `%v`
instead of `TemporalLiteral` in `constWord`, disabling the lambda quoting, and disabling the operand reader.

**Regression diff** over every committed script (base binary twice, new once): no difference apart from the nondeterministic
`examples/scripts/03_kleisli_fanout`, `examples/tour/05`, `15`, `20`, `23` (they differ between two runs of the *base* binary).

**Deferred / open (decided on the recommendation, say if you want the alternative):**
- **T1a** `spawn T: 5s` for a type with no registered parser still seeds the String `5s` (a spawn seed is not a literal token;
  `W2` step 7 decided undeclared spawns keep text). Typing it would change `spawn S: 5` from `"5"` to `5` too. Use `emit`/`let`.
- **T1b** a *quoted* `const` value that evaluates to a number is still evaluated (`const Q = "2+3"` is `5`), against the
  documented "a quoted value is always a String"; only the time-valued case was fixed here. Fixing it means skipping the
  expression path for every quoted value.

#### 5.6.2a T1b closure notes (Track N, 2026-10-13): a quoted value is always a String

Closes T1b. Branch `hardening/n`, commit `0fadf61`.

- **Root cause.** `const` stripped the quotes with `strings.Trim` and then decided with a flag; the expression path ran for any text
  holding an operator character, so `"2+3"` was evaluated (Track I patched only the time-valued result). Quotes were not part of the
  token rule at all: `emit "5"` kept the quotes as text, an operator operand `++ "!"` kept them too (`a"!"`), and a spawn head
  `spawn S: Q` spliced the constant unquoted.
- **The rule, in one place.** `functions.UnquoteLiteral` (a token that is *exactly one* double-quoted literal: no unescaped inner quote,
  the interior taken literally, no escape processing, so a path or pattern holds what was typed) is the first test of
  `functions.LiteralToken` (so `const`, `emit` and `-D` follow it) and of `functions.OperandValue` (operator sections). `const` returns
  before the expression path for a quoted value; the spawn head splices a String constant with `constSplice` (quoted). Unquoted
  tokens are unchanged: `const N = 2+3` is the Int `5` (pinned), `const B = 007` the String `007`, `const J = "a" ++ "b"` is not one
  literal and stays an expression (`ab`).
- **`-D` (decided, recommended option).** The shell removes quotes before the program sees them, so `-D Q="2+3"` is the value `2+3`
  (a String since it is not canonical) and `-D N=5` an Int; keep the quotes in the value (`-D 'ZIP="10115"'`) to get a String that is
  spelled like a literal. No flag syntax was added. Documented in the guide.
- **Behaviour changes to know:** `emit "5"` is now the String `5` (it was the text `"5"` with quotes, a String either way);
  `|> ++ "!"` appends `!` (it appended `"!"` with quotes); `|> + "5"` is an error (a String and an Int). No committed script
  exercises these (the regression diff differs only in the five nondeterministic scripts; `--check` silent).
- **Tests.** `functions.TestLiteralTokenQuotedIsString`, `cli.TestQuotedConstantIsAlwaysString` (stored value by raw type `string`,
  every position, `-D`), DSL section 117 (inside lambdas, heads, operands, `emit`, the unquoted pins). Negative control: making
  `UnquoteLiteral` always false fails 28 assertions and the DSL section fails on the base binary (13 errors).



---

## 6. W3 — The numeric tower, specified (D3)

**Decision D3(a):** keep the runtime representations; make the tower an explicit, tested specification.

1. **Logical types:** `Int` (unbounded; implemented as `int64` fast path + `big.Int`, canonical after each operation —
   **demote** `big.Int` results that fit, fixing F11's "never demoted"), `Float` (IEEE 754 double), `Complex`;
   **`Ratio`** (exact fractions, §6.1). `BigInt`/`Uint` remain accepted *names* (aliases of `Int`) so existing
   `type X = BigInt` lines keep parsing.
2. **Promotion lattice (documented):** `Int → Float` only when the other operand is `Float` or the operation is
   inexact; `Int/BigInt` overflow stays exact; `Int → Ratio` when the other operand is a `Ratio` (always exact);
   `Ratio → Float` when the other operand is a `Float`; anything `→ Complex` when an imaginary operand appears.
   Mathematically `Int ⊂ Ratio ⊂ Float(approx) ⊂ Complex`.
   `Int → Float` can lose precision beyond 2⁵³ — **raise a *warning* from `--check`** when the operand is a literal
   or a known `Int` that large, and document `toFloat` as the explicit form.
3. **Division:** keep `/` (exact when divisible, else `Float`) and `//` (floor division); add `divmod`, `quot/rem`
   if wanted; **`/ 0`, `// 0`, `% 0` are errors** (W0).
4. **New functions:** `**`/`pow`, `sqrt`, `floor`, `ceil`, `round`, `trunc`, `sign`, `gcd`, `lcm`, `to_int`,
   `to_float`, `is_nan`, `is_inf`; unary minus on any expression; `min/max/abs/compare` already exist.
5. **Constraint names for annotations** (the generalized `Num`): in `ParseTypeAnnotation` and the checker,
   - `Num` ≡ any of Int, Ratio, Float, Complex (what `number` is in Elm),
   - `Integral` ≡ Int (operands of `//`, `%`), `Fractional` ≡ Ratio/Float/Complex (operands of `/`),
   - `Comparable` ≡ Num ∪ Char ∪ String ∪ lists/tuples of Comparable, `Appendable` ≡ String ∪ List,
   so `def sum2 : Num -> Num -> Num = \a b -> a + b` documents and checks intent without a runtime wrapper.
6. **Tests:** an operator × type-pair matrix for the tower (Mandate 2 for numerics), promotion/demotion cases,
   overflow boundaries, and the display of big integers without `+`.

7. **Strings are not numbers — the part of D2 deliberately left for W3 (user, 2026-10-13):** (a) **equality** — `==`/`!=`
   between a `String` and a number must be `false` (not `true` through `StructuralEqual`'s `fmt.Sprintf("%v")` fallback), and
   two `String`s compare as text (`"10" == "10.0"` is `false`; today the numeric-string branch says `true`); the fallback is
   the one place that also makes `'a' == "a"` true, so it is replaced by explicit D1 rules, not just deleted;
   (b) **the helpers** `ToFloat64`/`toInt64`/`ToBigInt`/`ToComplex128` stop parsing `string` (they are what lets `nth "1" xs`,
   `range "0" "2"`, `min "5" 3` and `compare "5" 3` accept text); each caller is checked, since a few feed `Compare`/sort over
   text nodes; (c) the matrix test (item 6) gains the rows `String × {Int, Float, Complex, BigInt, Ratio}` for every operator.

### 6.1 `Ratio` — exact fractions

**Representation.** A `Ratio` is a Go `*math/big.Rat`, always normalised (lowest terms, positive denominator) and
never mutated. Core stays standard-library only (Mandate 5); the type lives in `functions/ratio.go` next to `Char`.
**Canonical form:** a result whose denominator is 1 **demotes to `Int`** (so `1r2 + 1r2` is `1`, and `ℤ ⊂ ℚ` holds
without a `4r1` that is "almost an Int"); this is the same rule as the `big.Int` → `int64` demotion in item 1.

**Syntax (D11–D13 decided).**

| Form | Meaning |
| :-- | :-- |
| `3r4` | literal: numerator `r` denominator, both unsigned integers (`-3r4` is unary minus on it; `6r8` normalises to `3r4`; `1r0` is a lex error: *denominator is zero*) |
| `0.75r`, `12.5r` | **exact decimal literal**: the decimal text is parsed as written, so `0.1r` is exactly `1r10` and never the nearest double |
| `ratio n d` | function form (data-last, curried): `ratio 1 3`, `3 \|> ratio 1`; `d = 0` is an error; arguments are `Int`/`Ratio` (so `ratio 1r2 3` is `1r6`) |
| `a + b`, `-`, `*`, `/`, `**` with a `Ratio` operand | exact; `Ratio / Int` and `Int / Ratio` stay exact (`1r3 / 2` is `1r6`); `x ** n` is exact for an `Int` exponent (negative exponent: reciprocal, `0` base is an error), a non-integer exponent promotes to `Float` |
| `a // b`, `a % b` with a `Ratio` | floor quotient (`Int`) and the matching remainder (`a - b * (a // b)`, a `Ratio`); `b = 0` is an error |
| `==`, `<`, … | exact across `Int`, `BigInt`, `Ratio` (D12: also exact against `Float`, so `1r10 == 0.1` is `false`) |
| `3r4` printed | `3r4` (round-trips through the lexer); `show` and `str_from` agree |

Why a suffix and not `1/3`: `/` on two `Int`s already means "exact when divisible, otherwise `Float`" (`7 / 2 == 3.5`,
pinned in `dsl_guide.md` and the tests) and **D11 keeps it** — silently turning every division into a
fraction would change the output of existing scripts, and `1:3` would collide with the `key:value` syntax of `emit`
and `spawn`. `%` is modulo (Haskell's `%` is not available), `1/3` as a *literal* would make `/` context-dependent.
The `r` suffix reads like Clojure's `1/3`, Scheme's `1/3` and Python's `Fraction`, keeps the lexer change to one
character in the number scanner (`TokNumber` already scans digits, `.` and the complex `i`), and `ratio` covers computed
operands. The recorded alternative to D11 is *exact `/` for `Int`/`Int`* (Scheme/Clojure/Julia-`//`-style); the
costs are in the D11 note below and the choice is cheap to reverse because `/` goes through one function.

**Converting from a `Float` — three named, explicit functions** (a float never converts to a `Ratio` implicitly):

| Function | Result | Use |
| :-- | :-- | :-- |
| `to_ratio x` | **the shortest decimal that round-trips** (`strconv.FormatFloat(x, 'g', -1, 64)`), parsed exactly: `to_ratio 0.1` is `1r10`, `to_ratio 0.75` is `3r4`, `to_ratio 1e-3` is `1r1000` | the default and the "smart" one: it returns the number the user *meant*, not the binary artefact |
| `to_ratio_exact x` | the exact value of the double (`big.Rat.SetFloat64`): `0.1` is `3602879701896397r36028797018963968` | when exactness of the *stored* value matters (reproducing a computation, hashing) |
| `to_ratio_approx x max_den` | the **best rational approximation with denominator ≤ `max_den`** by continued-fraction convergents (Python's `limit_denominator`): `to_ratio_approx 3.14159265 100` is `311r99`, `to_ratio_approx 0.3333333 100` is `1r3` | measured data, recovering a simple fraction from a noisy float |

`NaN` and `±Inf` have no `Ratio`: all three return an error value (Mandate 6). `to_ratio` on an `Int` is the
`Int` itself; on a `Ratio`, itself; on a `String`, see `str_to_ratio` below.

**Converting to a `Float` and `Int`:** `to_float r` is `big.Rat.Float64` — the **correctly rounded** nearest double,
never the lossy `numerator/denominator` division of two doubles; a magnitude beyond the double range is an error, an
underflow is `0.0` (documented). `to_int r` is an error unless `r` is integral (so a lossy conversion is never
silent); the rounding family is explicit and returns `Int`: `floor`, `ceil`, `trunc`, `round` (half away from zero,
Go's `math.Round`; `round_even` for banker's rounding). `numer r` and `denom r` give the parts (`Int`; `denom` of an `Int`
is `1`), `ratio_parts r` the `Pair`.

**Text:** `str_to_ratio s` is `Maybe Ratio` and accepts `"3/4"`, `"3r4"`, `"-0.75"` and `"1.5e-3"` (malformed text is
`Nothing`, never an error — same rule as `str_to_int`); `show`/`str_from` print the literal form. `ratio_mixed r`
returns the `Pair` (whole part, proper fraction) for display, `7r2 → (3, 1r2)`.

**Smart extras (cheap, decided with the rest of the tier):** `ratio_cf r` the continued-fraction coefficients
(`List Int`) and `ratio_from_cf xs`, which is also the building block of `to_ratio_approx`; `ratio_mediant a b`
(Stern–Brocot mediant). Each is one function over `big.Rat`; skipped if time-boxed, with the reason recorded.

**Where it plugs in (Mandate 2 cells):** `functions.Ratio` helpers and `EvalBinaryOp`/`Compare`/`StructuralEqual`
arms (the promotion lattice of item 2 decides the arm order: `Ratio` before `Float`, `Float` before `Complex`) →
number scanner `r`-suffix in `parser/infix.go` and `parsePrimary` → `functions.TryNarrowInteger`-style demotion
(`fitRatio`) → function forms in `pkg/cli/dslfuncs_numeric.go` → `tyRatio` in the checker (`Ratio` is numeric, so
the existing "all numerics are consistent" rule applies, plus `Fractional ⊇ Ratio`) → `sampleOf` → DSL assertions inside
lambdas/comprehensions (`fold (\a x -> a + x) 0r1 …` summing `1r2 … 1r10` exactly) → a structural Go test on the raw
`*big.Rat` → `Example`. Negative cases: `1r0`, `ratio 1 0`, `to_ratio nan`, `to_int 1r2`, `1r2 // 0`, `0r1 ** -1`.
`Ratio` is a scalar, so no carrier row is needed; as a collection element it already travels through `map`/`fold`.
`scalarValue` (the text-substitution guard in `cli.go`) must treat `*big.Rat` as non-scalar for the same reason as `Char`.

**Tests specific to the type:** the field axioms by property loop (associativity, distributivity, `a - a == 0`,
`a * (1 / a) == 1` for `a ≠ 0`) over a table of mixed `Int`/`Ratio` values; `(1r10 + 2r10) == 3r10` as the headline case
that `0.1 + 0.2 == 0.3` fails for `Float`; round-trips `to_ratio (to_float r)` for decimal-representable `r`;
`to_ratio_approx` on the convergents of π; every `Ratio × {Int, BigInt, Float, Complex}` operator pair in the W3 matrix.

**Decision notes (all decided 2026-10-13, alternatives recorded in §1).**
- **D11:** keep `/` on `Int`/`Int`; exactness is opt-in. No existing script changes and the literal makes the intent visible at the call site. Against: users coming from Scheme/Clojure expect `(/ 1 3)` to be exact.
- **D12:** arithmetic with a `Float` promotes to `Float`; `==`/`<` are exact — what Python's `fractions.Fraction` does. Price: `1r10 == 0.1` is `false` (the answer is `to_ratio 0.1 == 1r10`). It is stricter than today's `Int == Float` (compared as doubles); the difference only shows beyond 2⁵³.
- **D13:** `%` stays modulo; the `%` fraction operator is the recorded alternative.
- **Size:** about a day of work inside W3, after the logical-`Int` demotion rule exists (the `Ratio` demotion reuses it).

### 6.2 Track A closure notes (W3 implemented on branch `hardening/w3`, 2026-10-13)

One commit per sub-step, every gate green after each (`gofmt` on touched files, `go vet`, `go test ./...`, `./run_dsl_tests.sh`,
`go test -race ./pkg/... ./evaluator ./parser ./interactive ./cmd/gatwd-k8s-operator`, script regression diff against a baseline built at the
branch point, `--check` silent over every `*.gatwd`). The regression diff differs from the baseline only in `examples/tour/18_big_math.gatwd`
(the F11 `+` is gone) plus five tour scripts that are nondeterministic on the baseline itself (dates, map order).

| Step | What it did | Commit |
| :-- | :-- | :-- |
| 1 | unary minus, `**`; the F2 rows `- 5` and `2 ** 3` flipped; `|> ** n`, `math_pow` | `7f7603b` |
| 2 | the logical `Int` (canonical, demotes), the lattice `Int < Float < Complex`, exact comparison, zero-division errors, `pow sqrt isqrt floor ceil trunc round round_even sign abs negate gcd lcm divmod quot rem to_int to_float is_nan is_inf`, F11 | `a348057` |
| 3 | strings are not numbers (`"5" == 5` is `false`; `ToFloat64`/`ToBigInt`/`ToComplex128`/`toInt64`/`toUint64` no longer parse text) | `c8eaf0c` |
| 4 | `Num` `Integral` `Fractional` `Comparable` `Appendable` in the checker | `6c4acb9` |
| 5 | `Ratio` (D11–D13) | `be3633a` |
| 6 | prefix `!` | `39d68e5` |
| 7 | the operator x type-pair matrix test (it found: a big integer next to a `Complex` failed) | `99b6d18` |

**Behaviour changes (breaking, all intended; none changed a committed script's output).**
- `//` now **floors** and `%` takes the **sign of the divisor** (`-7 // 2 == -4`, `-7 % 3 == 2`); it truncated before. `quot`/`rem` are the truncated forms. The plan said "floor division" and defined the `Ratio` remainder as `a - b*(a//b)`, so the Int rule follows it.
- A `-` directly before an operand after a space (`f -3`) is a negative **argument**; `n -1` with `n` a number is an application error with a hint. `5 -3` (a literal cannot be applied) and `a - 3`, `a-3` are unchanged. In a plain list `[3 -1]` is two elements (a printed list of negatives now reads back; it used to read as `[2]`).
- `==`/`<` between an `Int` and a `Float` are **exact** (they compared as doubles; they differ only beyond 2^53). A `NaN` is unordered. Comparing a `String` with a number in `Compare` (`min "5" 3`) is an error naming `str_to_int`.
- Integer results are canonical: unsigned operands take part as ordinary integers (a `uint64` operation no longer answers with a `uint64`).
- Big integers print without a leading `+` (F11).

**Deviations from the text above (each a call made while implementing; veto any).**
- **`Ratio` is `*functions.Ratio`, a defined type over `big.Rat`, not a bare `*big.Rat`.** A bare `*big.Rat` prints `3/4` (it has only `String()`), which does not lex back and is what lists print their elements with; the named type prints `3r4`. Structural tests assert `*functions.Ratio`.
- **`Int ** negative Int` is a `Float`** (`2 ** -1 == 0.5`), like `/` on integers that do not divide; a `Ratio` base with a negative exponent is exact. The plan only described the `Ratio` case.
- **`//` and `%` are defined for `Float` too** (`//` gives the `Int` floor of the quotient, `%` a `Float`); `Complex` is an error (no floor on a plane).
- **`Integral` and `Fractional` are consistent with every *numeric* type** in the checker (the tower converts at runtime), so they contradict only non-numbers and each other; the plan's `Integral ≡ Int` would otherwise warn on every `Float` argument. `Appendable` is `String, Bytes, List, Chunk, Set, HAMT, Tree, Graph, Pair, Maybe` (the plan said `String ∪ List`; `++` is defined on all of them), `Comparable` is everything but `Complex` and functions.
- **(Superseded by §6.3: the fallback is gone.) The `%v` fallback of `StructuralEqual` was kept for String-versus-*collection*** (`"[2 1]" == [2, 1]` was `true`): HAMT keys are Strings, and the pure-DSL Game of Life (`examples/game_of_life_pure_dsl`, `TestStructuralPipelines/Pure_DSL_Conway_Step`) compares keys that went through `frequencies |> hamt_pairs` with coordinate pairs. The real fix is HAMT keys that carry values (a Layer-1 change). A `String` never equals a number, `Bool` or `Char` (the D1 one-character rule is explicit).
- `to_ratio 1e-3` in §6.1 cannot be written: the lexer has no exponent literals. `to_ratio 0.001` is `1r1000`. `str_to_ratio "1.5e-3"` does parse exponents (bounded at ±1000).
- The `--check` advisory for an `Int` beyond 2^53 mixed with a `Float` (§6 item 2) judges **literals** only (the checker does not track values), so "a known `Int` that large" is deferred.
- A fraction below the smallest double converts to `0.0` in `to_float` (documented in §6.1); beyond the range it is an error.

**Deferred, with reasons.** `quot`/`rem`/`divmod` are in; `Duration / Duration` (§5.6) is not started (needs Q2). `ratio_cf`, `ratio_from_cf`, `ratio_mediant` are in (the "smart extras", not time-boxed out). JSON has no form for a `Ratio` (`json_encode 1r2` is an error naming it; use `to_float` or `show`). A negative index in `str_slice` is an error (not "from the end"); no decision exists for that. `Float` literals with exponents (`1e-3`) are not in the lexer.

**Open questions for the user.** (1) Keep the String-versus-collection equality boundary, or schedule the HAMT-keys-carry-values change so the fallback can go? (2) Is `Int ** negative Int` as a `Float` right, or should it be an exact `Ratio` (`2 ** -3 == 1r8`)? It is one function (`intPow`). (3) `n -1` as an application error (Elm's rule): keep, or require `f (-1)` and leave `n -1` a subtraction?

---

### 6.3 Track G closure notes: HAMT keys carry values; a String is not a printed collection (2026-10-13)

Closes handoff Q7 / §6.2 "the `%v` fallback is kept for String-versus-collection". Branch `hardening/g`, four commits (key design + pinning
tests; every user migrated; the fallback removed; docs). Gates after each: `gofmt` on touched files, `go vet`, `go test ./...`,
`./run_dsl_tests.sh`, `-race ./pkg/... ./evaluator ./parser ./interactive`, `--check` silent (two pre-existing warnings in `tests/`), script
regression diff (differences only in the five nondeterministic scripts, `tour_23` and `tour_08`, which counts files matching `gatwd` in the
working directory; the Game of Life prints `3` and `5` as before).

- **Key design (decided, no open choice remained).** *Options:* (A) a canonical string encoding of the value, kept in the existing
  `*HAMT[string, any]`; (B) re-key the HAMT by a structural hash with equality buckets (`HAMT[any, V]`); (C) keep printing and leave `==` as it was.
  *Chosen: A* — B changes the type every adapter, optic, JSON/YAML reader and the Kubernetes records rely on (a Layer-1 rewrite, hard to reverse) for no
  gain; C is the lossy status quo. A keeps String keys byte-for-byte themselves (records unchanged) and encodes any other value as NUL + a tagged,
  length-prefixed canonical text: `evaluator.EncodeKey` / `DecodeKey` (`evaluator/hamt_key.go`; it lives in Layer 3 because Layer 1 may not import `dsltypes`
  for `Variant`), numbers through `functions.CanonNumber` (`functions/numkey.go`: whole numbers are the canonical Int, a fraction a double holds exactly is that
  float, any other `Ratio` stays one — exactly the `==` classes, so `1r10` and `0.1` are different keys because `1r10 == 0.1` is false). Deterministic,
  list order kept, Sets/HAMT entries sorted, injective (a property test over a pool and 70 random nested values checks `keys equal <=> StructuralEqual`
  on every pair), invertible (`DecodeKey` gives a value `==` the original, a real `*List`/`*Set`/`*HAMT`).
- **Migrated:** `hamt_from/set/get/has/del/get_or`, the `hamt_get/set/del` steps (`hamtStepKey`), `to_hamt` (a Pair key is any value),
  `frequencies`, `hamt_keys/hamt_pairs` (steps and functions decode). `json_encode`/`yaml_encode` of a HAMT with non-String keys is an error.
  **Graph ids and colors** were already Strings; `dslStr` printed any value, so a list id was its printed form. They now use `idStr`: a String, a
  `Char`, or a number/Bool as its text, and a collection is an error naming `show`. No committed script or the k8s prelude passed anything else.
- **Removed:** the printed-form block and `isCollectionValue` in `functions.StructuralEqual`; `"[2 1]" == [2, 1]` is `false`, `String == String` and the D1
  `Char` rule are unchanged, `Compare(String, collection)` is an error. Tests: carrier x `String` matrix (`pkg/cli/hamt_keys_test.go`, negative-controlled by
  restoring the fallback), DSL sections 106 and 107, `TestHAMT*`, `TestFrequencies*`, `TestGameOfLife*`.
- **Deviations / deferred.** (1) The `to_set` and `distinct` *steps* still deduplicate by printed form (`StepToSet` builds a `Set[string]`, `StepDistinct`
  hashes `%v`), so `emit [1,2] |> to_set` and `"[1 2]"` collapse — the same lossy coercion on a different type (the Set row, not equality); the function
  form `to_set` is already value-typed. Not changed here because a Set of values changes the `set_*` step semantics (DSL test 55 asserts `set_contains [1 2]` on such a Set);
  **open question for the user:** make the step value-typed too (recommended, small follow-up)? (2) A Complex number is `==` through doubles, a key is
  exact, so two big integers differing only below double precision next to a Complex could be `==` but different keys; documented in `hamt_key.go`.
  (3) `hamt_keys` of a HAMT built by the optics/JSON readers returns the same Strings as before.
- **k8s.** `pkg/k8sgraph` tests (including live/offline parity) and `lib/k8s/*.gatwd` pass unchanged; ids are Strings (`Kind/ns/name`).

#### 6.3a Track J closure notes: the Set and `distinct` are value-typed (2026-10-13)

Closes the §6.3 deferral (1) on the user's decision "make the step value-typed too". Branch `hardening/j`.

- **The premise was half wrong.** §6.3 said the function form `to_set` was already value-typed. It was not: it built a `*Set[any]` keyed by Go's `==` on the
  interface value, so `length (to_set [1, 1.0, "1", 2])` was 4, `to_set [[1,2],[1,2],[2,1]]` had three members (two list pointers) and an
  unhashable element was an "unhashable" error. Both forms needed the change, and they now share one type.
- **Representation (decided, recommended option taken).** *Options:* (A) a `Set[string]` of `EncodeKey` strings with decode on read; (B) a `Set[any]` made value-equal;
  (C) a new Layer-1 set whose element identity is a caller-supplied canonical key, elements stored as themselves. *Chosen: C* — `collections.Setoid[T]`
  (`collections/setoid.go`, a `HAMT[string, T]` from key to element) with `evaluator.ValueSetKey` (= `EncodeKey`, a Time by its UTC instant) as the key function. A would
  have made every duck-typed reader (`ElementsAny`, `functions.ToSliceAny`, JSON, the visualizer) decode, and left a `Set[string]` of NUL-tagged text visible to Go callers;
  B is impossible (Go cannot hash a slice, and a list is a pointer). With C no member is ever decoded from text (a member read back is the value stored), the key design is the
  HAMT's own (so a Set and a HAMT key agree on `==`), and it is hard to reverse only in the sense that every producer now builds it, which is the point. `*collections.Set[T]`
  stays as the Go-native set (Go's `==`); `evaluator.AsValueSet` reads any set carrier as the DSL Set, and two sets of one carrier still combine by the carrier's own method.
- **Behaviour.** `[1, 1.0, "1"] |> unchunk |> distinct` keeps `1` and `"1"` and merges `1` and `1.0`; `[[1,2],[1,2],[2,1]]` has two members; a Time in two zones is one
  member; `set_contains "[1 2]"` is false on a Set of lists (DSL test 55, which asserted `set_contains [1 2]` on such a Set, still asserts it, now on real lists, and gained
  the negatives `[1 3]`, `"[1 2]"` and the size 3). The first of several equal members stays. Step operands of `set_contains`/`set_add`/`set_remove` are values (`i.setItem`:
  canonical number, quoted String, expression such as `[1 2]` or a let-bound name, bare text as the fallback).
- **Fail loud (Mandate 6), two changes beyond the brief.** (1) A `set_contains`/`set_add`/`set_remove`/`set_elements` step handed a non-Set used to *pass the value through unchanged*
  (`emit 1 |> set_contains 1` printed `1`); it is now an error naming the step. (2) An element with no equality (a function, a Graph, `NaN`, an unknown Go type) is an error in
  `to_set` (no Set is yielded then), `distinct` (an error node, the rest flows) and the `set_*` function forms; `distinct` of a stream of Graphs, which used to work by printed text, is now
  that error. No committed script was affected (the script regression diff is clean).
- **Mandate 2 matrix.** `to_set` over List, Set, HAMT values, Tree and a stream; `set_add/remove/contains/union/intersect/diff/elements`, `member`, `==`, `++`, and `map`/`filter`/`flatmap`
  over a Set, each in step and function form where both exist: `pkg/cli/value_set_test.go`, `evaluator/valueset_test.go`, `collections/setoid_test.go`, DSL section 115 (function forms
  inside lambdas, then the step forms). Negative-controlled by restoring the printed-form key (21 DSL errors, 5 evaluator and 15 cli test failures). The k8s graph tests, live/offline parity included, pass unchanged.
- **Deviations / deferred.** (1) `BST[string]` (`bst_insert`/`bst_member` over `dslStr`) still stringifies its elements; a BST is ordered by `Compare`, not a set, and `BST[any]` is the
  default, so it is recorded here, not changed. (2) `functions.KeyText` stays for `dslStr` (path segments, a Variant tag, a graph id) and is documented as display text, not identity.
  (3) `set_*` step operands accept a let-bound name now (a bare word that names a variable is its value, not its text), the same resolution `set_union`/`subset_of` always had.
  (4) `subset_of` over two non-Set collections is still a pairwise `StructuralEqual` scan (O(n·m), value-typed, unchanged); the `set_intersect`/`set_diff` fallbacks over a List now build DSL Sets.
- **Open question.** `emit 1.0` is a String (`LiteralToken` keeps a non-canonical number as text), so a Float needs a function form or a `let` list in step-form tests; that is Track I's area.

**Track N follow-up (2026-10-13), closes deferral (1): `BST` holds values.** Branch `hardening/n`, commit `b627fca`.
- **Finding.** Only the `build_bst` step produced a `BST[string]` (`%v` text, `strings.Compare`); the function forms already built a
  `*BST[any]` ordered by `functions.Compare`, but accepted a `BST[string]` and stringified into it (`bst_insert`/`bst_member` over
  `dslStr`), so a step-built tree behaved differently from `bst_from` (`[10, 9]` was `[10 9]`).
- **Representation (decided, recommended option).** One type, `*collections.BST[any]`, and one implementation:
  `collections.InsertOrdered` / `ContainsOrdered` (`collections/bst_ordered.go`, ordered by `functions.Compare`; the tree returned with an error is the
  unchanged receiver). `StepBuildBST`, `bst_from`, `bst_insert` and `bst_member` all call them; every `BST[string]` branch (`length`, element
  access, the function forms) is deleted. `collections.BST[T]` itself is unchanged (`BST[int]` etc. stay available to Go callers).
- **Behaviour.** `[3, 1, 2, 10, 9]` is in order the Ints `1 2 3 9 10`; Strings order as text; an Int and the equal Float are one member
  (`1`, `1.0`); Lists order lexicographically. Elements that cannot be ordered are an error naming both
  (`bst_from: cannot order "a" and 1 in a BST: incomparable types: String and Int …`); `build_bst` yields that one error node and no tree.
  `bst_member "10"` on a tree of Ints is the same error, not a quiet `false` (it used to stringify to match).
- **Tests.** `collections` (`TestInsertOrdered*`, `ExampleInsertOrdered`), `evaluator` (`TestStepBuildBST*`, raw type), `pkg/cli`
  (`TestBSTKeepsValuesStepAndFunctionForm`: step and function forms build `StructuralEqual` trees; `TestBSTIncomparable…`), DSL section 118
  (function forms inside lambdas and comprehensions, then the step form). Negative controls: comparing printed text instead (12 Go failures),
  and the base binary on section 118 (5 errors). `tour_16` (`build_bst |> visualize`) is byte-identical.
- **Deferred / open.** (1) `==` on two trees is shape-sensitive (generic fallback): the same elements inserted in a different order are not `==`;
  compare `to_list`. *Open question:* give a BST content equality (in-order sequence), which is what a Set-like ordered collection means. (2) `map` over a
  BST is still an error (a Functor could break the ordering invariant; skipped for a mathematical reason). (3) `build_bst` still skips upstream error
  nodes silently (pre-existing, not changed here). (4) There is no `to_bst`; `bst_from` is the constructor.

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

### 7.1 What W4 built, decided and left open (2026-10-13, branch `hardening/w4`)

Four commits, each with every gate green: (1) the generic graph pieces, (2) the fake cluster keeps namespaces,
(3) the Kubernetes prelude + every test/fixture/example migrated, (4) docs and this closure note.
The design above is implemented as written; these are the calls made while doing it.

- **Generic first.** `collections.Graph.InsertNode` and `*DuplicateNodeError` (id, both colors, both payloads, an
  optional `Render` so the DSL can print a nested-map payload as JSON); the step `graph_insert_node COLOR ID`
  (`evaluator.StepGraphInsertNode`) and the function form `graph_insert_node color id [payload] g` (3 or 4
  arguments, like `graph_add_node`, whose raw-`DSLFunc` exception it shares). No Kubernetes word is in core; the id
  scheme itself is *prelude* code (`pkg/k8sgraph/lib/k8s/ids.gatwd`: `k8s_scope`, `k8s_id`, `k8s_id_of`,
  `k8s_ref`, `k8s_short`, `k8s_id_kind`, `k8s_id_scope`) built from `str_*` and the payload; adapters and the shim
  needed **no change** (the payload already carries `kind`, `metadata.namespace`, `metadata.name`).
- **F13 (one adjacency rule).** `Graph.Neighbors`, `Predecessors`, `ShortestPath`, and so the walks/`reachable`/`graph_out`/
  `graph_in`, and the function forms `graph_in_of`/`graph_out_of`, no longer follow an edge whose far endpoint is not
  a declared node. The edge stays stored (`Edges`, `Size`, `graph_edges` still count it) and the raw
  `OutEdges`/`InEdges` accessors are unchanged: `Connect` still does not create endpoints. `TopoSort` and `SCC` already
  ignored undeclared targets. One DSL test (section 77) relied on undeclared endpoints and now declares its nodes.
- **F14.** `Graph.AddEdgesOnce`; `graph_add_edges` and `graph_add_inferred_edges` add an *identical* spec (color,
  weight, endpoints, group, inferred flag) once, also against edges already in the graph. Edges that differ in any
  field stay; the single-edge step `graph_add_edge` stays a multigraph insert. (The plan said "identical
  `(from, to, color)` specs"; a different weight is a different fact, so the whole spec is compared.)
- **A fail-loud defect this exposed and fixed:** `collect_list` dropped error nodes from its stream, so a duplicate
  id would have printed as a plausible list. It now yields the errors instead of a list (`evaluator`, pinned by
  `TestStepCollectListKeepsErrors`); the regression diff shows only `examples/k8s_howto/infra.gatwd` run with the
  plain binary changing from `[]` to `unknown step "manifests"`, which is the point.
- **Relations.** Every reference resolves with the referrer's scope through `k8s_ref` and every edge is guarded by
  `graph_has_node` (the `Owns` and `RoutesTo` passes used to add edges to missing targets). `Selects` keeps its label
  index, now keyed by namespace, and an empty selector selects the Service's namespace only. `RoleBinding` subjects use
  their own `namespace`, a `roleRef` its `kind`; `Calls` resolves env values and ConfigMap data in the referrer's
  namespace. `k8s_names` is deleted (a `graph_has_node` on a `k8s_ref` replaces it) and the compliance rules print
  qualified ids; **an override file that used `k8s_names` or `metadata.name` ids must be migrated** (the how-to's
  `rules.gatwd` is).
- **Fixtures and parity.** Every fixture now names `namespace: default` (a live crawl always reports one, so a
  namespace-less offline fixture would yield `Pod/_/x` against the fake cluster's `Pod/default/x`); the fake cluster
  keeps the namespaces a manifest names. New: `collisions.yaml` (same names across kinds and namespaces, **with**
  live/offline parity over all namespaces), `scopes.yaml` (the `_` fallback), `duplicates.yaml` (the negative fixture;
  also through the operator). All were seen red first (guards negative-controlled).
- **Cost.** About 1.6x slower on the synthetic cluster, still linear (`benchmarks/README.md`). Not optimised.
- **Left open, with the reason:**
  1. *"Naming both sources"*: the duplicate error names the id, both colors and both payloads as JSON, but **not the
     file** each document came from, because no adapter attaches a source to a payload. Adding a provenance field to
     the payload would touch `adapters/manifest_adapter.go` (Track C's file) and the live/offline parity shape; decide
     the shape (a reserved `_source` key? a side channel?) before doing it. **Open question for the user.**
  2. *`ownerReferences` without `kind`* (invalid in Kubernetes, but easy to write by hand) is now an error from
     `hamt_get` (`key 'kind' not found`) instead of matching the owner by name. Loud and cheap to change: a lenient
     fallback (match any kind by name) would reintroduce the ambiguity D6 removes. **Open question: keep strict?**
  3. *Not W4's files, found on the way:* a comprehension with **no generator** (`[ e | guard ]`) yields `[]` even when
     the guard holds (worked around in `k8s_scope`; the playbook trap table has it) — a silent defect for the parser
     owner; `gatwd-k8s prelude list|cat` shows only the entry file `lib/k8s/prelude.gatwd`, not the files it imports
     (`relations.gatwd` cannot be read, so the "copy, edit, shadow" workflow of the how-to is half a workflow) — W6's
     `command.go`.


**Track E closure — source provenance (item 1 above; 2026-10-13, branch `hardening/e`).** The duplicate-id error now names
the file each declaration was read from (`declared as Pod in a.yaml (…) and again as Pod in more/b.yaml, document 2 (…)`).
*Design, and why not the suggested `_source` key:* the source is a **side channel on the payload value**, not a key in
it — `collections.Source{Origin, Document}` attached with `HAMT.WithSource` (Clojure's `with-meta`), read with
`collections.SourceOf`. It is invisible to `Get`/`Seq`/`Keys`/`Len`/`Equal`/`StructuralEqual`/JSON and the DSL, so no rule,
no graph and no export can see it and live/offline parity is untouched by construction; a `_source` key would have
changed `hamt_keys`, `json_encode` of every payload (and any live→manifest export) and forced every parity comparison to
strip it. A test proves the difference: with the source stored as a key, `TestSourceProvenanceIsInvisible` fails.
`Set`/`Delete` (and so `Merge`/`Diff` from the receiver) keep the source, `Filter`/`MapValues`/`Intersect` do not. Core
stays domain-free (Origin is an opaque string; `collections` only). `adapters/manifest_adapter.go` attaches the file, and
the 1-based document position only when the file holds several documents (empty documents count, so it matches the
file); a rendered kustomization root is `kustomize render of DIR, document N`; live crawls attach nothing. `DuplicateNodeError`
gained `ExistingSource`/`NewSource` and its text is unchanged when neither payload has one. DSL: `with_source origin
document h` and `source_of v` (a `Maybe String`), registered from `pkg/cli/source_funcs.go` through `registerGraphFuncs`
(not `dslfuncs.go`, to stay out of Track F's registry lines). Tests: the two-file fixture `testdata/duplicates-files/`,
`duplicates.yaml` (document 1 and 2), payload parity live vs offline by JSON, the operator real-binary row; negative
controls: source not attached, source stored as a key (both seen red). **Deferred:** a source on *live* payloads (a
namespace/cluster label would be possible with the same mechanism; nothing needs it); `with_source` only carries a HAMT
(the record shape) — other carriers are a recorded gap, not a mathematical one.
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

### 8.1 What W5 built (Track C, 2026-10-13; branch `hardening/w5`)

Four commits, each with all gates green: the Go bump · `Edition.WithFlags` · the in-process renderer · auto-render plus `--no-kustomize`; this section and the how-to are the fifth.

- **Go bump (D7):** `go.mod` says `go 1.27.0`; the system `go` is 1.26.0 and `GOTOOLCHAIN=auto` selected the cached 1.27.0 toolchain, no manual step. Every gate was green on the bump alone (the only script-diff entries were the three tour scripts that differ run to run on the *base* binary too: map iteration order). `go mod tidy` also made cobra and pflag direct dependencies, which they already were in the code. No CI file or Dockerfile exists; the Go requirement is documented in `docs/architecture.md` and `docs/k8s_howto.md` §3.1.
- **`Edition.WithFlags(func(*pflag.FlagSet))`** (`pkg/cli/edition.go`, `CommandSpec.Flags` in `command.go`, three additive lines): the generic hook that lets an edition declare a persistent flag. It was missing: `WithCommand` adds subcommands only, and `--no-kustomize` must be a flag of the root. Generic and domain-free (Mandate 5); test `pkg/cli/edition_flags_test.go` (flag before/after the subcommand, `--help`, fresh state per tree, an undeclared flag stays exit 2), negative-controlled.
- **`pkg/k8sgraph/kustomize.go`:** `InProcessRenderer()` is the `adapters.WithRenderer` function; the dependency (`sigs.k8s.io/kustomize/api` and `kyaml` v0.21.3) is imported there and nowhere else (`internal/archtest` green). `KustomizeStep` uses it by default; the external binary stays selectable through `adapters.WithKustomizeCommand`. Output order is kustomize's legacy order (what `kustomize build` prints); a per-call `Kustomizer`, renders serialised by a mutex, a panic becomes an error.
- **Safety:** default load restriction kept; plugins/KRM/Helm disabled (`DisabledPluginConfig`); **remote bases and files are refused before kustomize can fetch them** — the library hard-wires `git clone` and HTTP GET with no switch (`internal/loader`, `internal/git` are unexported), so a `guardedFS` checks every kustomization file it reads (and every generator/transformer/validator config it names) against a copy of kustomize's own remote-URL rules (`ssh://`, `https://`, `http://`, `file://`, `git::`, `user@host`, `github.com/`); an unparsable kustomization is refused rather than waved through. Enabling remote bases would be a deliberate, separate option (none yet; recorded below).
- **`manifests` auto-renders (user decision 2026-10-13):** `ApplyEdition(e, client, opts ...Option)` registers `manifests` with the renderer and `adapters.WithKustomizeSwitch`; the persistent flag **`--no-kustomize`** (and the option `k8sgraph.WithoutKustomize()`) turns it off; the explicit `kustomize` step ignores the switch. Only the *path itself* is tested for being a root (a tree containing roots is read file by file, as before). The shim stays wiring: no render logic, no algorithm.
- **Tests:** fixtures `pkg/k8sgraph/testdata/kustomize/` (a base, an overlay with `namePrefix`, `namespace`, `labels`, a strategic-merge patch and an added Ingress; `expected-prod/` is the hand-written plain equivalent; negative fixtures `broken`, `remote`, `remote-config`, `exec`, `escape`). The rendered overlay is compared resource by resource with the hand-written expectation and with `kubectl kustomize` (skipped when absent — it ran and passed here); the overlay and its plain twin build the same graph (non-empty: 4 nodes, 3 edges); plain files and directories give the baseline step's graph with the switch on and off; the switch in every spelling; the failure cases are `[ERROR]`s naming the root; a real-binary test runs with an empty `PATH`; 16 concurrent renders under `-race`. Negative controls (each seen red, then restored): remote guard off, load restriction off, switch ignored in the adapter, auto-render removed, flag ignored, every directory treated as a root, binary switch forced. Ids in the tests are asserted by substring (`prod-web`), so they survive W4's qualified ids.
- **Size:** the stripped static binaries (`CGO_ENABLED=0 -trimpath -ldflags="-s -w"`) grew from 28.1 to 31.3 MB (`gatwd-k8s`) and from 27.8 to 31.0 MB (`gatwd-k8s-operator`), about +3 MB (+11%). Unstripped, `gatwd-k8s` went from 39.8 to 71.0 MB (debug information of the library tree). The core `gatwd` binary does not link kustomize and is unchanged (35.5 to 36.0 MB unstripped, Go 1.27 and the flag hook).
- **Deviation 1 — a `replace` in `go.mod`.** `kustomize/api v0.21.3` requires `k8s.io/kube-openapi v0.0.0-20261001230523`, which moved to `structured-merge-diff/v7`; `k8s.io/apimachinery v0.37.1` (the newest non-alpha) still builds against v6, so with the newer kube-openapi nothing compiles. `go.mod` therefore pins `k8s.io/kube-openapi` back to `v0.0.0-20260721132016-d427ff9ee9ad` (the version we had) with a `replace`; everything builds and every test passes, including kustomize's OpenAPI schema handling. **Remove the `replace` when `apimachinery`/`client-go` v0.38 (currently only alpha) is adopted.** The alternative — moving the Kubernetes libraries to `v0.38.0-alpha.2` — was not guessed. The module graph also moved `go.yaml.in/yaml/v3` v3.0.4 to v3.0.5, `google/gnostic-models` to v0.7.1, `golang.org/x/sys` to v0.48.0 and gained `blang/semver`, `go-errors`, `treeprint`, `go-gitignore`. **Resolved 2026-10-13 (user decision):** `client-go`, `apimachinery` and `api` are now `v0.38.0-alpha.2`, which uses `structured-merge-diff/v7`, so the `replace` is deleted; the one source change was a new method on `dynamic.ResourceInterface` (`DeleteWithResult`) that the lazy client stub in `pkg/k8sgraph/lazy.go` had to implement. All gates green; the stripped static `gatwd-k8s` is 32.2 MB (it was 33.1 MB). The alpha is a pre-release: revisit when v0.38.0 is tagged.
- **Deviation 2 — the mutex is defence in depth.** Upstream's race report (#4824) no longer reproduces: kyaml v0.21.3 locks its global schema (`schemaLock`). Removing `renderMu` did not make the concurrency test fail in 30 runs of 16 goroutines under `-race`; the test pins the promised behavior but cannot prove the lock necessary. Kept, as the plan says.
- **Found; **fixed later by W4 (`collect_list` passes error nodes on, commit 0840c97)**; was "not fixed" (pre-existing, outside this track):** `collect_list` drops error nodes, so any failing source piped into it prints `[]` with exit 0 (`spawn StrVal: /nonexistent |> manifests |> k8s_graph |> graph_edges |> collect_list`; same on the base binary). A Mandate 6 hole in the evaluator/step layer that W0 missed. The tests and the how-to work around it (no `collect_list` where a failure must be seen); it needs its own fix and a row in `pkg/cli/failloud_test.go`.
- **Deferred, with reasons:** (1) remote bases/files behind an explicit opt-in — a security decision for the user, not guessed; (2) a `--no-kustomize` flag for the operator Appliance — `cmd/gatwd-k8s-operator` belongs to W6; the option `k8sgraph.WithoutKustomize()` is ready for it; (3) in-memory rendering of bases shipped in a ConfigMap or an `embed.FS` (§8 "Filesystem") — nothing consumes it yet; `filesys.MakeFsInMemory` is the route; (4) skipping the `Kustomization` document when a root is read plain (`--no-kustomize` on a root fails in ingest with `key 'kind' not found`) — an ingest/W4 matter in the prelude; (5) the how-to's container section still shows a `scratch` image built from a host-built binary; a `golang:1.27` multi-stage build is not added because no Dockerfile is committed.

---

## 9. W6 / W7 — Operator, policy and script sources

**W6 (small):**
- **Override mode (D5, decided: layer by default):** `Appliance` gets `WithOverrideMode(Layer | Replace)`; default **Layer** (embedded first,
  override second, so an override can reuse `rule_*` and redefine only what it needs); flag
  `--rules-mode layer|replace`. Imports in an override resolve against **the override file's directory**, not the cwd (F16).
- **Exit codes and formats:** `--fail-on-violation` (exit 1 when the result is non-empty), `--format text|json`
  (JSON for log pipelines), still exit 2 for "no source".
- Operator `--help`/flags come from W1.

**W6 closure (2026-10-13, track D; veto any call below).** Built as specified, in five commits (layering, lazy client, command tree, exit codes and formats, this closure):
- **Layering (D5)** — `cli.OverrideMode` (`OverrideLayer`, the zero value and default; `OverrideReplace`), `WithOverrideMode`, `Appliance.Report()` (`LoadReport`: the override applied, the defs/constants it **overrode** — a different body — and **added**; an identical copy is not reported), imports in an override resolve against the override file's directory and embedded-FS imports stay inside the FS (F16 closed). Types are not diffed in the report.
- **Loud loading** — `*cli.LoadError`: an unreadable override (a directory is not "absent"), a structural `--check` error, a top-level statement that prints `[ERROR]` or nothing, an unreadable import, or a `WithRequiredDefs` name left undefined fails `NewAppliance`; `Reload` returns it and **keeps the last good rules**. Top-level output while loading is discarded (it would pollute a machine-readable stdout). Not detectable at load, by design: a rule that is well-formed but fails when it runs — that is reported by the check (`[ERROR]`, exit 1). A dry-run probe of the pipeline at load was considered and left out (it needs the data source).
- **Command line** — `cli.NewApplianceTree(ApplianceSpec{Name, RulesPath, Bind, Embedded, LegacyFlags})`: `check`, `run --interval`, `rules cat`, `version`, `completion`; the no-subcommand legacy form loops with `--interval`, else checks once; single-dash flags (`-manifests` …) are rewritten with a stderr notice. The operator `main` only supplies `Bind` (its source flags) and a setup that builds the client **only on the live path**. The Edition tree's `built-in check/run` names are not reused (an Appliance does not run scripts), hence a second tree sharing `executeRoot` and `versionCommand`.
- **Exit codes** — `0` ok; `1` the command failed (rules that do not load, a check that **errored**, a pipeline that **yielded nothing** — Mandate 6 —, or violations under `--fail-on-violation`); `2` a malformed line (unknown flag/`--format`/`--rules-mode`, no source). **Decision made on the plan's wording:** violations exit non-zero only with `--fail-on-violation` (default off), so a scheduled `run` is not restarted for finding something; `run` never exits on a failing check and rejects `--fail-on-violation`. Open question for the user: should `check` fail on violations by default (CI-friendly, but breaks the old "exit 0" scripts)? Also: violations and "the command broke" both exit 1 (`diff`-style 1/2/3 would let CI tell them apart; not in the plan).
- **Formats** — `--format text` (as before; an error value now prints `[ERROR] …`) and `json`: one line per check `{"status":"ok|violations|error","count":N,"violations":[{"rule","resource"}],"errors":[…]}`; a finding that is not a `[String, value]` pair is `{"value":…}`; a value JSON cannot hold is an entry in `errors`, never dropped. Reports and diagnostics go to stderr. W7's structured diagnostics will replace the free-text `errors` strings.
- **Lazy client** — `k8sgraph.LazyDynamicClient()` / `LazyClient(build)`; `gatwd-k8s` registers the live steps always and builds the client on the first crawl, so `version`/`completion` are silent; with no cluster a crawl is `[ERROR] no cluster access (…)` (it used to be `unknown step`, i.e. a step that did not exist). Handoff gotcha closed.
- **Found; **fixed later by W4 (`collect_list` passes error nodes on, commit 0840c97)**; was "not fixed" (outside W6):** `collect_list` drops error nodes (`… |> manifests |> collect_list` over a missing path prints `∷ []`, exit 0) — a silent failure for the next fail-loud pass (`evaluator/`, track B's files). `-e`-less script top-level statements still keep their old exit status (§4.5).
- **Security, unchanged until W7 (closed by the W7 closure below):** a mounted override is still *trusted code* — its defs can use any step, including the unknown-step OS-command fallback (F17); layering does not widen that (replace mode ran the same text), but the §10 "replace-only until W7" mitigation no longer means anything once the default is `layer`: W7's deny-by-default policy for overrides is what closes it.
- **Deferred:** `GATWD_<FLAG>` environment defaults (§4.3, nothing needs them); a `--rules-mode` that merges per-def (`layer` is whole-def redefinition); watching the ConfigMap (reload stays interval-driven); `context`/signal handling in the loop (W7).

**Track E closure — exit code 3, operator `--no-kustomize` (2026-10-13, branch `hardening/e`; user decisions).**
- **Exit codes, now:** `0` ok · `1` the command broke (error nodes — a duplicate id included —, a pipeline that yielded nothing,
  rules that do not load, a failed `check`) · `2` a malformed command line · **`3` the check ran and found violations, only
  with `--fail-on-violation`** (the flag stays opt-in; `run` still rejects it). An error always wins: a check that both errored
  and found violations exits 1. Built as `cli.ExitOK/ExitFailed/ExitUsage/ExitViolations` and `cli.ViolationsError{Count}`
  (`command.go`, mapped in `executeRoot`, which both command trees share), not a magic number; a domain command can return
  `ViolationsError` too. Pinned by `TestExitStatusMapping`, the toy Appliance rows (including "an error wins over violations")
  and the real-binary operator rows (violations → 3, JSON → 3, a duplicate id → 1); the mapping was negative-controlled
  (violations mapped to 1, and the error/violation order swapped, both seen red). Docs: `dsl_guide.md` §23, `k8s_howto.md`
  §3.2 (with a CI `case $?` snippet), the `edition-appliance` skill. The W6 open question "should `check` fail on violations
  by default" is **answered: no**.
- **Operator `--no-kustomize`:** a persistent flag of the operator tree declared in its `Bind` and mapped to
  `k8sgraph.WithoutKustomize()` in `setup` (the main stays wiring). **Call made:** with the live source (no `--manifests`) the
  flag would silently do nothing, so it is a usage error (exit 2) — loud instead of ignored; veto if you want it accepted.
  Test: `TestOperatorNoKustomize` (a checked-in root `testdata/kustomize-root`: rendered gives `Pod/default/x-web`, with the flag
  the plain files `Kustomization/_/k` and `Pod/default/web`), negative-controlled twice (option not applied; live guard off).
  Closes the §8.1 deferral (2).

**W7 (later; design now, build when needed):** running scripts that did **not** come from the command line —
ConfigMaps, API requests — is only safe with a **capability policy**, because an unknown step executes an OS
command (verified, F17) and the DSL can read files and sockets.
- `cli.Policy{AllowExec, AllowFS, AllowNet, AllowEnv}`; the CLI script argument is trusted (all allowed);
  Appliance overrides and API-submitted scripts default to **deny exec/FS/net/env**.
- Deny-by-default means the exec fallback becomes an `[ERROR] step 'printf' requires the exec capability`.
- The `Run(ctx, Request) Result` API from W1 carries the policy, a timeout (`ctx`), and an output size limit.
- Out of scope here: authentication/authorisation of API callers (belongs to the server Appliance).

**W7 decision record (2026-10-13; the user: "continue", on the recommendation).**

1. The policy gates only the capabilities **reachable from script text**. Steps an Edition wires itself (`manifests`, the cluster crawl, any `AdapterStep`/`RegisterStep`) are the author's decision, made in Go, and stay allowed.
2. `cli.Policy{AllowExec, AllowFS, AllowNet, AllowEnv}`. The script named on the command line, `-e`, preludes and the REPL are **trusted** (`cli.TrustedPolicy()`; `NewInterpreter` starts there). An Appliance override and any script that did not come from the command line (`RunOptions.Source`) default to **`cli.DenyAll()`**. The operator gets `--allow-exec/--allow-fs/--allow-net/--allow-env`, all off; a ConfigMap-sourced override can never widen the policy. `RunOptions` carries the policy, a `context.Context` and an output limit (D8).
3. A denied capability is an **error value on the Left channel** (`[ERROR] step 'printf' requires the exec capability (policy denies it)`), never an empty result (Mandate 6). Structured diagnostics are added to the JSON report and the free-text line is kept.

**W7 closure (2026-10-13, track H, branch `hardening/h`; veto any call below).** Built in five commits (policy, Run bounds, operator, failures, docs).

*Inventory of script-reachable side effects* (every Go call that gives script text an effect, found by reading `pkg/cli`, `adapters/`, `interactive/`, `evaluator/` for `os`, `os/exec`, `net`, `net/http`; the guard test `TestEveryScriptReachableSideEffectIsGated` fails the build when a new one is wired in without `i.gated`):

| Capability | Script text | Go |
| :-- | :-- | :-- |
| `exec` | the unknown-step fallback (`printf …`, `sh -c …`), `memoize CMD`, `zip CMD` | `interactive.GenericExecAdapter` |
| `fs` | `scan_file`, `scan_words`, `walk_fs`, `write_file`, `append_file`, `import "file"` | `adapters.StepScanFile/WalkFSAdapter/WriteFile`, `loadScriptContent` |
| `net` | `dial_tcp`, `listen_tcp` (`conn_lines` only reads connections those steps produced) | `adapters.StepDial/StepListenTCP` |
| `env` | **none exists** — the generic `env "NAME"` deferred in W1b was never added | — |
| not gated, by design | `manifests`, the cluster crawl, Edition steps; the `OnMissingAdapter` fallback (its command template is empty, so it starts nothing: `TestEmptyExecTemplateRunsNothing`); embedded `import`s inside an `fs.FS` (the author's code, not file access); `--prelude FILE`, the REPL's `:load`, the script path (command line, trusted); the appliance's own read of its override file (Go code) | — |

*What was built.* `pkg/cli/policy.go`: `Capability`, `Policy` (the zero value grants nothing), `TrustedPolicy`, `DenyAll`, `CapabilityError`, `Interpreter.SetPolicy/Policy/RequireCapability/SetContext`. The policy and the run context live in **one `execState` shared by pointer with every fork and child interpreter**, and a gated step reads it **when its stream runs**, so a def compiled (and cached) before the policy changed still obeys it and a fork can never hold a more permissive copy (an interpreter assembled without `NewBareInterpreter` fails closed). `Interpreter.Run` takes `RunOptions.Source/Policy/Context/MaxOutput`, installs them for the call and restores them; a `Source` run with no `Policy` is `DenyAll`. Cancellation is cooperative: every pipeline step checks the context per node (and before it starts), the statement loop checks between statements, `GenericExecAdapterContext` kills the child process; a run that ends with a done context returns an error wrapping the cause (`context.DeadlineExceeded`, `context.Canceled`, `cli.ErrOutputLimit`). `Appliance.WithPolicy` (default `DenyAll`) applies to the embedded rules, the override and every later `Execute`; `ApplianceSettings.Policy` and the four flags on `NewApplianceTree` carry the process's grant to the operator `main`, which only passes it on (Mandate 7). Failures: `cli.Failure`/`FailureOf` (`capability_denied`, `unknown_step`, `stopped`, `no_result`, `unencodable_value`, `error`) and a `diagnostics` array in the operator's JSON report; `interactive.UnknownStepError` gives the unknown-step failure a type (same message). Tests: the `Policy` type, one row per gated step (denied exactly, allowed under `TrustedPolicy`), forks, lambda bodies, cached defs, Edition steps, `RequireCapability`, an override that tries to widen the policy, the real operator binary (denied by default, `--allow-exec` grants it, the other flags do not, top-level statement stops the start, replace mode is no way round it), timeout with a slow command and with a long pure stream, output limits, JSON diagnostics. Each guard was negative-controlled.

*Deviations and calls.*
- **The Run API keeps its one-argument signature.** §4.2 sketched `Run(ctx, o)`; the context is a `RunOptions` field instead, so callers (and the parallel tracks' tests) did not change. Documented as against the usual Go advice.
- **`RunOptions.Source`** (a script as text) was added: it is what makes "a script not from the command line defaults to deny-all" expressible in `Run`.
- **JSON diagnostics are a new `"diagnostics"` array next to the unchanged `"errors"` strings**, not objects inside `errors` (that would have changed the type of an existing key). It is always present, so healthy output changed by one key; every repo test and doc sample was updated.
- **An override's `import` needs `--allow-fs`** (the request listed "`import` of files if applicable"). This makes the W6 feature "imports resolve against the override's directory" opt-in. *Open question:* allow imports confined to the override's own directory without the flag?
- **`AllowEnv` gates nothing today** (see the inventory); `Interpreter.RequireCapability(CapEnv, …)` is the hook a future `env` builtin or Edition step must call. *Open question:* add the `env "NAME"` builtin (`Maybe String`) now? Left out because it is a new capability with an API to name, not something the policy needed.
- Error nodes yielded by the policy and cancellation wrappers carry the zero type flag (`Parser.GetFlag` registers on first use and is not safe from `||>` goroutines).
- The embedded rules and the Edition's preludes run under the Appliance's policy too (one policy per interpreter); the Edition's own preludes run trusted inside `BuildInterpreter`, before the policy is applied.

*Deferred, with reasons.*
- **`parser/errors.go` `FormatError` stays unwired** (decision of W1b/W7 was "only if small and safe"): it needs a `SyntaxErrorTracker` threaded through `Expect`-wrapped combinators, and the infix and line parsers are hand-written scanners that do not use the tracker; wiring it means rewriting their error paths, which is parser work, not diagnostics plumbing. What the report needs (a stable kind plus the step and capability) does not depend on it. Line and column for `LoadError.Problems` (the §10.1 follow-up) wait for the same reason: `EvaluateLine` does not know the line.
- **Timeout and output-limit flags on the operator**, and `Appliance.ExecuteContext`: the mechanism is in `Run`; the flag names and what a timed-out check should exit with are a product call (nothing in the plan or the answers settles them). `context`/signal handling of the `run --interval` loop stays open for the same reason.
- **A single non-terminating evaluation of one element** (a closure that never returns) cannot be interrupted: cancellation is checked between nodes and statements. Interrupting it needs a check inside `functions.Closure.Apply` (Layer 0) or a goroutine that is abandoned; both are bigger decisions than W7.
- **The output limit bounds what is printed and returned**, not the memory a pipeline uses while it computes (a task's lines are collected before they are returned); the context is the bound for that.
- CLI flags `--timeout`/`--max-output` for `gatwd` itself: the command line is trusted and nothing asked for them.

**Track K closure — operator `--timeout` and `--max-output` (2026-10-13, branch `hardening/k`; user decision of the same day: flags `--timeout DUR` and `--max-output BYTES`, a timed-out check exits 1; veto any call below).** Closes the W7 deferral "timeout and output-limit flags on the operator". Three code commits and the docs:
- **Appliance** (`pkg/cli/appliance_bounds.go`): `cli.WithTimeout(d)` and `cli.WithMaxOutput(n)` (zero, the default, is off; a negative value makes `NewAppliance` fail), `Appliance.Timeout()/MaxOutput()`, and `Appliance.ExecuteContext(ctx, line)`; `Execute` is `ExecuteContext(context.Background(), …)`, so **every check — and every round of `run --interval` — gets a fresh deadline and a fresh byte budget**. It reuses W7's machinery (the run context every step observes, children killed), it does not duplicate it. With no bound and a context that cannot end, `Execute` touches nothing on the interpreter (pinned).
- **A stopped check reports exactly one error value and none of its partial results**: `*cli.TimeoutError` (`check timed out after 30s`, matches `context.DeadlineExceeded`) or `*cli.OutputLimitError` (`check output exceeded the limit of N bytes`, matches `cli.ErrOutputLimit`); when the caller's own context ends it is a `*cli.CancelError`. Output over the limit is **loud, never truncated**; the budget is enforced in `EvaluateToValues`'s collecting loop, so an endless stream is cancelled as soon as it passes the limit (a test runs one with no timeout).
- **Diagnostics, additively**: failure kinds `timeout` and `output_limit` (`cli.FailTimeout`, `cli.FailOutputLimit`) and a new optional `Failure.Limit` / JSON `"limit"` key (`"30s"`, `"40"`); the key is omitted for every other kind, so existing output is unchanged.
- **Command line** (`ApplianceTree`): persistent flags `--timeout DUR` and `--max-output BYTES` (also in the legacy no-subcommand form), carried by `ApplianceSettings.Timeout/MaxOutput`; `Options()` applies them and `BoundOptions()` applies just them. A negative or unparsable value is a usage error (exit 2, before the source is read). A timed-out or over-limit check is a failed check: **exit 1, not 3**, also under `--fail-on-violation`; in `run` it is printed like any failing check and the loop continues, each round isolated (a test rewrites the override between rounds so round 1 times out and round 2 succeeds; the real binary test sees three timeouts in a row). The operator `main` only adds `s.BoundOptions()` next to `WithPolicy` on both paths.
- **Tests**: structural (`appliance_bounds_test.go`: slow command killed, endless/long pure streams stopped, partial results dropped, freshness per check, exact byte boundary, caller's context, negative values, diagnostics JSON), command tree (`appliance_command_bounds_test.go`), `ExampleWithMaxOutput`, real binary (`cli_test.go`: slow override with `--allow-exec` running `sleep` stopped at `--timeout 200ms`, output limit at the exact byte boundary, JSON, usage errors, `run` surviving a timeout, `--help`), and the how-to doc run (`doc_test.go`: three new rows, two stale-claim phrases). Negative controls (each seen red, then restored): output budget disabled; the stopped-check discard removed; timeout/limit dropped from the settings; negative `--max-output` not rejected; a deadline shared between checks; `BoundOptions` removed from the operator main.
- **Docs**: `k8s_howto.md` §3.2 (flags table, exit codes, a *Limits per check* section with sensible values, JSON diagnostics), §3.4 and `deploy.yaml` (commented `args:` line, defaults unchanged), §3.5 (two troubleshooting rows), §4 (the old "no timeout flags" limitation rewritten); `dsl_guide.md` §23 (flags, exit status, diagnostics kinds, the Go options).
- **Calls made, veto any**: (1) the output limit counts bytes of the **text** report (each result's printed form plus a newline) whatever `--format` is, so it does not depend on the format; the budget also counts the intermediate results of a `let … = pipeline in …` evaluated on the way (a slight over-count). (2) A context that ends just after the last result still reports the check as stopped (the same rule as `Run`). (3) Loading and hot-reloading the rules is not a check and is not bounded. (4) `--max-output` takes plain bytes, no `1MiB` suffixes.
- **Open questions**: (a) `ExecuteContext` installs the bounds on the active interpreter for the call, so two checks must not overlap on one Appliance (documented; an HTTP-server Appliance would need a per-call fork). (b) Signal handling (SIGTERM ending `run --interval` cleanly) is still open; a root context passed to the loop would now be easy to add. (c) A single non-terminating closure is still not interruptible (W7).

**Track L closure — writing to the Kubernetes cluster API (2026-10-13, branch `hardening/l`; user decisions of the same day: two capability families, both denied by default in the CLI *and* the operator, first version apply-only with no delete or prune).** Six commits:
- **Capabilities** (`pkg/cli/policy.go`, generic, no domain word): `CapClusterApply` and `CapClusterRollout` with `Policy.AllowClusterApply/AllowClusterRollout`; **`TrustedPolicy()` deliberately does not grant them** (`DenyAll()` neither), so even the user's own command-line script needs an opt-in. `Policy.Grant(c)`, `cli.GatedStep(cap, handler)` (wraps every step a handler builds in `Interpreter.gated`, checked when the stream runs) and `Edition.WithCapabilityFlag(name, cap, usage)` (a persistent flag that adds exactly that capability when the interpreter is built). `ApplianceTree` gets `--allow-cluster-apply` and `--allow-cluster-rollout` next to the four existing flags. The denial is the existing `CapabilityError` / `capability_denied` diagnostic (`"capability":"cluster-apply"`); `cli.Failure` needed no change.
- **Adapter** (`adapters/kubernetes_write.go`, `kubernetes_rollout.go`, the only place client-go is touched): `Applier` with server-side apply (`Apply`, field manager `gatwd`, `--force-conflicts` off), `RenderTo` (the offline twin), `Restart/Undo/Pause/Resume/Scale/Status`; a kind resolver (built-in table, then lazy discovery) and `ObjectFromPayload` (the inverse of the read adapter's HAMT lifting, validating apiVersion/kind/name). **It works through a narrow `writableResource` interface (Get, List, Apply, Patch), so a delete is not expressible**, and a test scans the adapter and shim sources for one. **Dry run is the default** (`DryRun: [All]` on every request) and `Persist` (`--dry-run=false`) is the only way out.
- **Shim** (`pkg/k8sgraph/write.go`, zero logic): `ApplyStep`, `RolloutSteps`, `WriteFlags` (`--dry-run`, `--force-conflicts`, `--field-manager`, `--out`, `--apply-namespace`), every write step built by `recordStep` inside `cli.GatedStep(...)`; `ApplyEdition` registers them always (so the offline-only edition has `--out`) and adds the two capability flags unless the host brings its own `WithWriteFlags` (the operator). Results are the DSL list `[verb, Kind/ns/name, outcome]` per object; a list input is a batch applied in order, and the prelude gains `k8s_apply_objects` (the payloads in `k8s_apply_order`).
- **Tests** (fake dynamic client + `k8sgraphtest.EmulateWrites`, which adds create-on-apply and honoring of `DryRun` to the stock fake): denied for the trusted script with zero requests reaching the cluster; dry run persists nothing and every request carries `DryRun=[All]`; create, idempotent second apply (`unchanged`), changed field (`configured`); each rollout step sends exactly one expected patch (merge / JSON patch) and only get/list/patch ever reach the server; `--out` files, round trip through `manifests`, parity with the live apply, fs capability; real binaries (`gatwd-k8s`: opt-in rows; operator: an override cannot widen, the other flags do not grant, `--out` needs fs, help text). **Negative controls seen red, then restored:** TrustedPolicy granting apply; `GatedStep` not gating; a write step wired ungated; default dry run flipped; dry-run option dropped from apply and from patches; a delete call added; status gated; fs check removed; `--out` ignored; `ownedBy` accepting any ReplicaSet; an Appliance flag always granted.
- **RBAC and docs:** `examples/k8s_howto/deploy-write-role.yaml` (separate, commented, opt-in; get/create/patch, never delete/update/wildcard; the shipped `deploy.yaml` is still get/list and a test pins both); `k8s_howto.md` §3.5 "Writing to the cluster" (the three layers of defence, the how-to's samples run by `cmd/gatwd-k8s-operator/howto_write_test.go`), `dsl_guide.md` §23.
- **Calls made, veto any:** (1) `rollout status` needs only `get`, so it is **not** gated; (2) `scale` is in the rollout family behind `cluster-rollout`; (3) dry-run spelling is `--dry-run` (default true) and real writes are `--dry-run=false`; (4) `--out` writes files whatever `--dry-run` says (naming the directory is the opt-in) but needs the **fs** capability on top of `cluster-apply`; (5) the apply Role grants `get`, `create`, `patch` and **not** `update` (nothing calls it; the brief listed it); (6) scale patches `spec.replicas` of the workload, so no `…/scale` RBAC; (7) a namespaced object with no namespace is an error unless `--apply-namespace` is given; (8) `k8s_apply` takes no per-step argument; (9) the operator's offline source (`--manifests`) has no client, so a write there is a loud error unless `--out`.
- **Deferred, with reasons:** delete/prune (user decision); `rollout status --watch` (a wait loop is a product call); undo/status/pause for StatefulSet and DaemonSet (their history is `ControllerRevision`); apply of a Secret value from the DSL is allowed like any object (no redaction decision); a client-side field-owner conflict report beyond the API's own error; a `kubectl diff`-style output for dry runs (the outcome is `would-*`, not a diff). **Open questions:** (a) should `--out` be allowed without `cluster-apply` as a plain `write_file` of objects (it is today refused)? (b) should the operator get `k8s_apply` against `--manifests` with an explicit `--target-cluster` flag, so an offline source can be applied to a live client?

**Track M closure — signals and bounded loading (2026-10-13, branch `hardening/m`; user: "queue them"; veto any call below).** Closes the two open items Tracks H and K left (the signal handling of `run --interval`, the unbounded load/reload of rules). Two code commits and the docs:

- **Signals** (`ApplianceTree.serve`): SIGINT and SIGTERM cancel one context (cause `stopRequest`, "received terminated"); the in-flight check runs under it through `Appliance.ExecuteContext`, so the existing machinery stops the steps and kills the OS command. **Exit statuses (decided, recorded):** `run` exits **0** with one stderr line (`NAME: received terminated, stopping`, plus `, the check in progress was cancelled` when one was running) — a requested stop is a daemon's normal shutdown, the cancelled check prints nothing and a check that failed before the signal was reported when it failed; `check` exits **1** (`the check was interrupted (received terminated)`) because it did not finish; a **second** signal exits 1 at once (the cooperative cancel cannot interrupt one non-terminating evaluation, so there is a way out). The signal source (`ApplianceTree.signals`) and the wait between rounds (`sleep`, now nil by default = a timer raced against the context) are injectable, so the structural tests raise the request from inside the run — no sleeps, no real signals. Real binary (`signal_test.go`): SIGTERM and SIGINT while `run` waits, SIGTERM during a check that is a real `sleep 300`.
- **Bounded loading**: `--timeout` (`WithTimeout`) is one budget per load — embedded rules and override together — at `NewAppliance` and at every `Reload`. A load that runs over is a `*cli.LoadError` whose new `Err`/`Unwrap` is a `*cli.TimeoutError{What: "loading the rules"}` (`loading the rules timed out after 30s`; `FailureOf` kind `timeout`). At start: exit 1; on a hot reload the last good rules keep serving and the failure is printed (`reload failed, still using the previous rules: …`), a fixed override recovers. The context is installed on the load interpreter and **removed again** before it is returned (a leaked, finished context would stop every later check). New: `WithLoadContext`, `Appliance.ReloadContext`, `ApplianceSettings.Context` (set by the command to the stop context, passed on by `BoundOptions`), so a stop request cancels a load too. The edition's own preludes are trusted and unbounded. Output limit does not apply to loading (its output is discarded already).
- **Tests**: `pkg/cli/appliance_signal_test.go`, `appliance_load_bounds_test.go` (pure statement, `--allow-exec` `sleep`, embedded rules, generous/no timeout, leak of the load context, cancelled context, reload keeps the last good rules and recovers, the toy command tree for check/run/json, a hot reload hanging inside the loop, a stop request during start-up load and during a reload), real binary (`signal_test.go`, `load_bounds_test.go`: start-up `check`/`run`/`json`, hot reload that hangs then is fixed, SIGTERM ends it with 0), `doc_test.go` rows. Hanging statements are sized to take ~4 s if nothing stops them against timeouts of 150–500 ms and assertions at 1.5 s, so a dead guard is red, not merely slow. **Negative controls (each seen red, then restored):** the context not installed on the load interpreter (eight tests), the context left on it (the leak test), `Context` dropped from the settings (start-up stop request), `ExecuteContext` replaced by `Execute` in the loop (the check tests; it exhausted memory on the endless stream). One control was vacuous and was fixed: `ReloadContext` in the loop, because `Reload()` already uses the stop context through `WithLoadContext`.
- **Existing test changed**: `TestApplianceFlagsMapToTheLimits` built an Appliance with `Timeout: 3` (3 ns); that now also bounds the load and made it fail, so it uses 3 minutes. A timeout too small to load the rules is, by design, a failed start.
- **Supersedes** Track K's call (3) ("loading is not bounded") and open question (b) (signal handling).
- **Deferred / open**: (a) a single pure closure that never returns (`def spin = \n -> spin (n+1)`) is still not interruptible — a load or check on it is not stopped by the timeout or a signal (a process that runs it dies of the Go stack limit, as before); a per-application context check belongs to the closure layer (Layer 0) and was not attempted. (b) `Appliance.ExecuteContext` still must not overlap with itself on one Appliance (unchanged). (c) `gatwd` itself (the script runner and REPL) has no signal handling; only the Appliance command line does. (d) No separate budget for loading vs checking: one `--timeout` serves both; split it (`--load-timeout`) only if a deployment needs a different value.

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
- **§5.6 (`Duration`/`Time`/`Date`) after W3**: `Duration / Duration` is an exact `Ratio`, so it needs §6.1; it is not on W4's path.
- **W2 before W4**: `k8s_id` needs string building (already possible with `++`, so W4 can start earlier if needed).
- **W5/W6 are independent** of W2–W4 and can be done any time after W1.

Each phase follows the playbook: design note → generic prerequisite → pinning tests first → one commit per
sub-step → full gates (`gofmt`, `go vet`, `go test ./...`, `./run_dsl_tests.sh`, `-race`) → **script regression
diff against a baseline binary** (the breaking changes in W0/W2/W4 will show up there on purpose) →
docs/playbook/skills updated in the same change.

### 10.1 Closing checklist: the Kubernetes how-to and the k8s code (user request, 2026-10-13)

When **all** phases are done, `docs/k8s_howto.md` and the Kubernetes edition (`pkg/k8sgraph`, `cmd/gatwd-k8s`,
`cmd/gatwd-k8s-operator`, the embedded `lib/k8s/*.gatwd`, their tests and `testdata/`) are reviewed and updated
against what changed. Run the `project-doc-sync` skill, `gatwd --check` over every `.gatwd` and every code block of
the how-to, and the script regression diff. Collect here as phases land; **each row is a task, not a decision.**

| Source of the change | Update in `docs/k8s_howto.md` | Update in the k8s code |
| :-- | :-- | :-- |
| **D2, D18, typed literal tokens** (done) | re-run every sample (`+` on text is an error, `-D`/`const`/`emit` numbers are numbers, strings have escapes); fixed already: the "no string builtins" claims in §2.4 and §4 | check the prelude and operator rules for text reaching arithmetic or ordering (a scan found none today); `cmd/gatwd-k8s-operator/main.go` builds `spawn StrVal: …` strings — keep them as they are, they are text <br>**Closing: done (c3ddf2d)** every how-to command re-run by `cmd/gatwd-k8s-operator/doc_test.go`; the prelude/rules scan (`+`, `<`, `//`, `%`, `== "…"`) is clean. |
| **`str_*`, `show`, `str_format`, `Char`, `Ordering`, Bool functions** (done) | document them where §2.3/§2.4 build ids or labels; show `str_format` in a rule's violation message | **adopt in the prelude** instead of ad-hoc text: the id builder (W4), violation messages, label matching <br>**Closing: done (c3ddf2d) except one deferral.** The id builder is `str_*` already (W4); the how-to's id and data sections use it. **Deferred:** `str_format` in a violation message: the rule result is the contract `[Rule, resource]` that the text and JSON reports and every test render, a message would be a third element; needs a report-format decision. |
| **W2 tier 2:** base64/hex/url, JSON/YAML, regex (RE2), hashes | rewrite §2.4's "why an annotation" paragraph: decoding is now possible, **keep the security call** (credentials must not enter graph payloads — decide, then document) | `json_parse`/`yaml_parse` for ConfigMap data and annotations; `re_match` for image tags and label patterns; `sha256` fingerprints of Secret data (never the data) <br>**Closing: done (c3ddf2d):** §2.4 rewritten with the security call (a rule may inspect a decoded value, may report a `sha256` fingerprint, never the value), new §2.5 and `examples/k8s_howto/data.gatwd` (YAML and JSON from a ConfigMap, base64 and fingerprints from Secrets) with `TestHowToDataExample`. **Deferred:** `re_match` for image tags / label patterns as an *embedded* rule — the six embedded rules are the product contract (tests pin the exact violation lists), a new rule is a product decision; `re_match` is shown in the §2.5 table. **Open question:** ingest-time redaction of Secret `data` (the crawl still holds it in memory). |
| **§5.6 `Duration`/`Time`/`Date`** (proposed) | an "ages and expiry" example (`now - (time_parse created) > 30d`, certificate `expires - now < 14d`) | `creationTimestamp`-based rules in `lib/k8s/*.gatwd`; the injectable `Interpreter.Clock` in the edition's tests (no wall clock) <br>**Closing: done (c3ddf2d):** §2.6, `age_rules.gatwd` (layered override) and `age.gatwd` (script), `TestHowToAgeRule` (injected `Interpreter.Clock`, three instants, negative control) and `TestHowToAgeOverrideInTheOperator` (real binary). **Deferred:** a `creationTimestamp` rule in `lib/k8s/*.gatwd` / the embedded rules (same reason as above: product decision; the example override is the supported route); certificate expiry is described, not shipped. **Found:** an age rule cannot call `now` inside a lambda (by design), so the reading is a top-level `let clock = (now |> head)`, which the operator re-evaluates before each check (verified with `run --interval`). |
| **W3 numeric tower, `Ratio`** | resource quantities (`500m`, `2Gi`) as an example of exact arithmetic; the numeric-constraint annotations in rule signatures | quantity parsing helpers in the prelude, if wanted (`Ratio` gives exact CPU/memory sums); `--check` stays at zero findings <br>**Closing: done (c3ddf2d):** exact quantities in `data.gatwd` (`500m + 250m` is `3r4`) and the §2.5 table; `def cores : String -> Num` and `def older_than : Time -> Duration -> Any -> Bool` show the annotations and `--check` stays silent on every `.gatwd` (54 files). No quantity helpers were added to the prelude (nothing consumes them). |
| **W4 qualified node ids (D6)** | rewrite §2.3 ("Ids are `metadata.name` only") and §4 (the collision limitation); every query and example output | `ingest.gatwd`, `relations.gatwd`, `queries.gatwd`, `types.gatwd`, `steps.go`, **all** tests, `testdata/` expectations and the live/offline parity test (`live_parity_test.go`) migrate in the same phase <br>**Closing: done (W4 commits; re-verified c3ddf2d):** every id in the how-to is printed by the doc test. |
| **W5 in-process kustomize (D7, Go 1.27)** | §3.1 build (Go version), §3.3 container image (golang tag), manifests/kustomization notes, remove "external binary" caveats | `adapters/` manifest ingestion, `pkg/k8sgraph/edition.go` (`ApplyEdition`), `go.mod`; auto-render of kustomization roots was **not** explicitly answered (D7) — confirm before documenting it <br>**Closing: done (c3ddf2d):** §3.1 (Go 1.27, client-go `v0.38.0-alpha.2` pre-release note, 32.7 MB / 31 MiB measured), §3.3 re-run with Docker (`FROM scratch` image: offline check and a kustomize overlay both ran), no external-binary caveats left (`TestHowToHasNoStaleClaims`). |
| **W6 operator: `--long` flags, layered rules (D5), exit codes, lazy client** | §2.2 (override rules now *layer*; `replace` is a mode), §3.2 flag table, §3.4 deploy manifests (args, exit codes), §4 (drop the "replace, not layer" and "exit 0 always" limitations) | `cmd/gatwd-k8s/main.go` (lazy client: `version`/`completion` must not print the "no cluster access" notice), `cmd/gatwd-k8s-operator/main.go`, `pkg/cli/appliance.go` reload semantics, the operator's embedded default rules <br>**Closing: done (c3ddf2d):** §2.2/§3.2/§3.4 against the final operator; `version`/`completion` quiet verified. |
| **W7 policy (D8), `context`, structured diagnostics** | a security section: which scripts may run OS commands; how a ConfigMap-sourced override is sandboxed | the Appliance's default policy for overrides; diagnostics instead of free-text `[ERROR]` in the operator output <br>**Closing: done (c3ddf2d):** §2.2 security section re-run with the real binary (`printf` denied, `--allow-exec`, `import` needs `--allow-fs`, JSON `diagnostics`) and pinned by `TestHowToDocSamples`; §3.4 says the Deployment's `args:` are the only place a capability is granted. |
| **Anything that renames or removes a builtin** (`compare` now returns `Ordering`, `-D` typing, …) | grep the how-to and `examples/k8s_howto/` for the old behaviour | grep `lib/k8s/*.gatwd`, the operator rules and `testdata` <br>**Closing: done (c3ddf2d):** grep of the how-to, `examples/k8s_howto`, `lib/k8s`, the operator rules and `testdata` for `compare`, `-D`, numeric strings: none. |
| **W5 closure (Track C): auto-render, `--no-kustomize`, in-process kustomize** (done in `hardening/w5`) | §3.2 now has a *Kustomize* paragraph and §4 two bullets (`collect_list`, operator flag); §3.1/§3.3 state Go 1.27 and the ~31 MB static size; re-run the two commands of the Kustomize paragraph after the other tracks merge (they print `∷ 4` and `∷ 3`); when W6 renames the operator flags, update the "no flag for this yet" sentence | **W6:** offer `--no-kustomize` on the operator through `k8sgraph.WithoutKustomize()` (the Appliance has its own flag set). **After W4:** re-run `pkg/k8sgraph/kustomize_edition_test.go` (it compares ids by substring and graphs by equality, so it should pass unchanged) and add a qualified-id expectation for the overlay. **Fix `collect_list` swallowing `[ERROR]`** (a failing source prints `[]`), then drop the workaround notes in the tests and the how-to. Remove the `replace k8s.io/kube-openapi` in `go.mod` once `client-go`/`apimachinery` v0.38 is adopted <br>**Closing: done (c3ddf2d):** both Kustomize commands print `∷ 4` / `∷ 3` and are pinned; the operator flag exists and the stale sentence is gone; `collect_list` keeps errors (W4); the `replace` is gone from `go.mod`. |
| **W6 as built** (layered overrides, `LoadError`, `NewApplianceTree`, exit codes, `--format json`, lazy client) | §2.2/§3.2/§3.4 already describe layering, the flags, exit codes and formats — re-run every operator command of the how-to against the final binary after W4/W5 merge (ids in the sample output); add the `--format json` example to any CI section | after **W4** (qualified ids): `idleWant`/`idleJSON`/`brokenJSON` in `cmd/gatwd-k8s-operator/cli_test.go`, `main_test.go`, `howto_test.go` and `testdata/compliant.yaml`'s expectations carry ids; after **W5**: `ApplyEdition` must still pass `TestLazyClientBuildsOnFirstUseOnly` (`pkg/k8sgraph/lazy_test.go`); the operator rules (`rules/compliance.gatwd`) stay unchanged by W6 <br>**Closing: done (c3ddf2d):** all operator commands re-run; the CI section now has the exit-code `case` and `--format json`. |
| **W7 structured diagnostics** (follow-up to W6) | the JSON `errors` are plain strings today; document the structured form when it lands | `cli.checkReport` (`pkg/cli/appliance_report.go`) is the single place that renders text and JSON — change it there; the Appliance load problems (`LoadError.Problems`) become diagnostics with line/column <br>**Closing: done (c3ddf2d):** the JSON `diagnostics` shown and tested (`doc_test.go`); line/column stay deferred (§9). |
| **W3 numeric display and operators** (done, §6.2): `//`/`%` floor, `==` exact and String-never-a-number, big integers print without `+`, `-x`/`!x`/`**`, `Ratio` | re-run every how-to sample that prints numbers or negative values; `n -1` in a sample is now an application error (write `n - 1`); mention `Ratio` for resource quantities (`500m` is `1r2`) | grep `lib/k8s/*.gatwd`, the operator rules and `testdata` for `//`, `%`, a negative literal after an identifier (`x -1`), `== "<number text>"` and numeric strings given to `nth`/`range`/`min`/`max`/`compare`; the scan done in W3 found none in the committed scripts <br>**Closing: done (c3ddf2d):** no `n -1`, `//`, `%` or numeric-string comparison in the how-to, examples, prelude or rules. |
| **HAMT keys carry values; a String is not a printed collection** (done, §6.3): graph ids/colors accept a String or a number's text and refuse a list/pair; `"[2 1]" == [2, 1]` is `false` | none of the how-to's samples compare a String with a collection; mention that an id is a String (`Kind/ns/name`) | `pkg/k8sgraph` parity tests and `lib/k8s/*.gatwd` pass unchanged; grep the how-to and `examples/k8s_howto/*.gatwd` for `graph_add_node` with a non-String id and for `== "[`; none found <br>**Closing: done (c3ddf2d):** nothing to change; the how-to says an id is a String. |
| **`compare` and ordering of mixed values** (W3): `compare`, `min`, `max`, `<` on a String and a number are errors; two Strings order as text | anything that sorts or compares labels/ports as text vs numbers: convert with `str_to_int` first | rules that compare a label or annotation value with a number <br>**Closing: done (c3ddf2d):** no rule compares a label or annotation with a number. |
| **W7 as built** (execution policy, `Run` bounds, structured failures; branch `hardening/h`) | §2.2 has a "What an override may do (security)" section (the capability table, `--allow-*`, imports need `--allow-fs`), §3.2 lists the flags, the JSON samples carry `diagnostics` (§3.2 and the DSL guide §23); **after the other tracks merge:** re-run every override example of the how-to with the real binary — an override that runs a program or imports a file now needs `--allow-exec`/`--allow-fs` (the committed `examples/k8s_howto/rules.gatwd` needs neither) — and say in the how-to's §4 that `run`'s loop has no timeout/signal handling yet | `cmd/gatwd-k8s-operator/main.go` passes `cli.WithPolicy(s.Policy)` (its `newAppliance` takes extra options); expectations that spell the JSON report (`idleJSON`, `brokenJSON`, the compliant row) gained `"diagnostics":[]`; **when Track E's exit code 3 / `--no-kustomize` merge**, keep both the new flags and the policy flags in `appliance_command.go` and the operator tests; a deployment whose override needs exec/fs/net adds the flag to the Deployment's `args:` (the sample `deploy.yaml` needs none) <br>**Closing: done (c3ddf2d):** override examples re-run (the committed `rules.gatwd` needs no flag; an `import` or OS command does); §4 says the operator has no timeout, output-limit or signal handling flags. |
| **W4 follow-ups** (the W4 row above is done on branch `hardening/w4`: how-to §2.3/§2.4/§3.5/§4, all ids and tests migrated) | after W5 (kustomize): say that `kubectl kustomize` output carries namespaces and that two overlays defining one object now fail with the duplicate-id error; if provenance (W4 §7.1 item 1) is decided, show the file in the error text | a source path in the duplicate error (adapter + payload shape, after Track C); decide `ownerReferences` without `kind` (§7.1 item 2); an override that still uses `k8s_names` or `metadata.name` ids fails loudly: mention it in the W6 layered-rules text <br>**Closing: done (c3ddf2d), one open question.** The duplicate error with files is documented and pinned (`rendered/a.yaml` / `rendered/more/b.yaml, document 2`); an override using `k8s_names` fails with `undefined identifier` (documented in §2.2). **Open question (unchanged):** `ownerReferences` without `kind` stay a loud error; the how-to documents the strict behavior. |
| **W4 finds for other tracks** | `prelude cat` reads only the entry file: either teach it the imported files (W6, `command.go`) or drop the claim "copy, edit, load with `--prelude`" for `relations.gatwd` | the operator prints a duplicate-id error and **exits 0** like a violation list: W6's exit codes (`--fail-on-violation`, `2` for no source) must also fail on an error result; the guard-only comprehension defect (§7.1 item 3) belongs to the parser <br>**Closing: done (f7b78e1, c3ddf2d):** `prelude list`/`cat` now show the imported files (generic, `Edition.Command`), the how-to uses them; a duplicate id exits 1 (Track E). **Still open (parser):** the comprehension-without-generator defect is documented as a trap in §2.1, not fixed. |
| **Track E: exit code 3, operator `--no-kustomize`, source provenance** (done in `hardening/e`) | §3.2 documents exit `3`, the CI `case $?` snippet, the flag and the file in the duplicate error (done); re-run the three commands (`check --fail-on-violation` → 3, `--no-kustomize`, a duplicate) after the other tracks merge | the operator `main.go` and `cli_test.go` carry the flag and the rows; **remove the stale "operator has no flag" sentences if another track re-adds them**; a source on live payloads is deferred (§7.1). When Track F's time types land, check that `with_source`/`source_of` (`pkg/cli/source_funcs.go`) did not collide with a registry change in `dslfuncs.go` <br>**Closing: done (c3ddf2d):** the three commands re-run and pinned (`3`, the flag, the duplicate with files); no stale operator-flag sentence remains. |
| **§5.6 as built (Track F)**: `Duration`/`Time`/`Date`, `now`, `Interpreter.Clock`, JSON/YAML as Strings | the "ages and expiry" example in `docs/k8s_howto.md`, run against the final binary with an injected clock (`now |> map (\t -> (t - created) > 30d)`); mention `time_parse` for `metadata.creationTimestamp` and that decoded YAML/JSON dates are text until `time_parse`/`date_parse` | adopt in `lib/k8s/*.gatwd` (`creationTimestamp` age rules, certificate expiry `expires - now < 14d`, `Secret` rotation); the edition's tests set `Interpreter.Clock` (`Edition.BuildInterpreter()` then `interp.Clock = …`), never the wall clock; **typed tokens** (T1) decide whether `-D MAX_AGE=30d` works for the operator; Editions that need a different clock set it on the interpreter they build <br>**Closing: done (c3ddf2d):** the ages example runs against the final binary with an injected clock; `time_parse` for `creationTimestamp` and text dates from YAML/JSON are in §2.5/§2.6; `-D MAX_AGE=30d` stays text (documented, §4). |
| **Track L: writing to the cluster** (`hardening/l`) | `docs/k8s_howto.md` §3.5 (three layers of defence, apply, `--out`, rollout, operator flags), §2.2 capability table, §3.6/§4 rows (done); re-run the §3.5 commands after the other tracks merge (the operator `--timeout`/`--max-output` rows share the flag table) | `pkg/k8sgraph/write.go`, `adapters/kubernetes_write.go`, `kubernetes_rollout.go`, `deploy-write-role.yaml`; if W6's operator `main.go` or `ApplianceTree` flags are edited by another track, keep `WithWriteFlags` and the two `--allow-cluster-*` flags; **adding any step that mutates a remote system must go through `cli.GatedStep`** (the guard `TestEveryWriteStepIsGated` and `TestWriteAdapterIsReachedOnlyThroughTheGatedSteps` fail the build otherwise) <br>**Closing: done (this branch).** |

**Closing run (2026-10-13, branch `hardening/closing`; commits `f7b78e1`, `c3ddf2d`).** Every row above is marked. What the run changed: the how-to is pinned to the real binaries by `doc_test.go` (needles in both the doc and the output; excerpts quote their files; a stale-claim list), a generic `prelude list|cat` that shows imported files, the data and age examples with tests, and a recipe test for "no Secrets in the crawl". **Open questions for the user:** (1) should an age / image-tag / expiry rule join the *embedded* operator rules (a product change: the six-rule contract and every exact-violation test move), or stay an example override as now? (2) Redact Secret `data` at ingest? (3) Should a rule result be allowed a message (`[Rule, resource, message]`), which would let `str_format` build it? (4) An operator flag for a per-check timeout and output limit (W7 deferral).

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
