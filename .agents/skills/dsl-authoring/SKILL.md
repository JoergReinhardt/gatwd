---
name: dsl-authoring
description: "Write pure gatwd DSL code — programs, graph relations, compliance rules, ingest passes — in the project's idioms (closures, folds, comprehensions, step vs function forms) and avoid the known traps (list-literal application, nodes before edges, positional guards, vacuous asserts, text-macro defs)."
---

# DSL Authoring Skill

Use this when you write or review `.gatwd` code: tests in `tests/dsl_exhaustive_test.gatwd`,
tour/examples, an Edition prelude (`pkg/<domain>/lib/**`), or Appliance rules
(`cmd/**/rules/*.gatwd`). The style reference is `pkg/k8sgraph/lib/k8s/relations.gatwd` and
`cmd/gatwd-k8s-operator/rules/compliance.gatwd`. Rules and rationale: `AGENTS.md` §5a and
`docs/development_playbook.md` §4 and §6.

## 1. Pick the right kind of definition

| You want | Write | Because |
| :-- | :-- | :-- |
| a function of values | `def add = \x y -> x + y` | a **closure**: curried, lexically scoped, `add 1 2`, `map (add 10)`, `fold add 0`, `x \|> add 3` (= `add 3 x`, data-last) |
| a reusable stream transformer | `def k8s_ingest = fold (\g r -> …) graph_empty` (no parameters) | a pipeline-bodied `def` is a re-parsed **text macro** |
| a whole-graph pass | `fold (\_ g -> graph_add_edges […] g) graph_empty` | a graph stream element reaches the reducer as `g`; `map` on it would be the graph Functor over payloads |
| a rule | `def rule_x = \g -> [ ["Violation", res] \| … ]`, combined with `++` | rules are closures from the graph to `[violation, resource]` pairs |
| a function by cases | `def fact 0 = 1` then `def fact n = n * (fact (n - 1))`; guards `def f x if x > 0 = …` | consecutive same-name clause lines form one curried closure; arms take `pat if cond -> body` |
| a lookup/helper | an ordinary closure (`def k8s_ref = \kind ns name g -> […]`) | closures are values; call them inside comprehensions |
| an index | `fold (\idx x -> hamt_set k (cons x (hamt_get_or [] k idx)) idx) hamt_empty xs` | avoids O(n·m) joins, stays pure DSL |

## 2. Step form vs function form

The step form takes **literal tokens** (`graph_add_node Pod nginx`); the function form takes
**computed values** (`graph_add_node (hamt_get "kind" r) (hamt_get "name" r) r g`). Use the step
form at the top of a pipe, the function form inside lambdas, folds and comprehensions. Function
forms are **data-last** and **curried** (`map (hamt_get "a") records`).

## 3. The traps (memorise these)

1. **Calls inside plain list literals need parentheses**: `[(f x), (g y)]`. `[f x, g y]` is four
   elements — lists are serialised space-separated and re-parsed. Comprehensions `[e | …]` and
   `( … )` are ordinary expression contexts.
2. **Nodes before edges.** `Connect`/`graph_add_edges` do not create endpoints; edges without
   nodes do not show in listings. Guard relations with `graph_has_node`.
3. **Guards are positional.** In `[… | x <- xs, hamt_has_path "k" x, y <- (hamt_path "k" x)]` the
   `hamt_has_path` runs as soon as `x` is bound — put it *before* the dependent generator.
   A comprehension needs **a generator**: `[ e | guard ]` alone is an `[ERROR] … no generator`
   (also flagged by `gatwd --check`); use `[ e | _ <- [0], guard ]` to yield `e` once when the guard holds.
4. **`hamt_get`/`hamt_path` error on a missing key** (naming the segment); use `hamt_has_path` /
   `hamt_get_or` when absence is normal.
5. **Edge specs are 4-element lists** `["Color", weight, "from", "to"]`; the DSL has no tuple
   literal. Colours are plain strings; a declared constructor (`Owns`) resolves to its tag string.
6. **`assert` passes on an empty stream.** After writing a test, run just that section and read
   every output line. `./run_dsl_tests.sh` fails on silent statements.
7. **No unary minus on identifiers** — write `0 - x`.
8. **A `def` used as a step with arguments**: `x |> f a` applies `f a x` for closures; a bare
   one-argument closure name maps element-wise.
9. **`to_set` (step) builds `Set[string]`; function forms use `Set[any]`.** Compare with
    `member`/`length`.
10. **Scratch scripts**: `cat > f.gatwd <<'EOF'`, never `echo` (zsh mangles `\x`).
11. **Go tests**: `EvaluateScript` for `def`/`type` setup, `EvaluateToValues` for one query
    (returns `*collections.List[any]` — use `functions.ToSliceAny`).

## 4. Annotate and check

`def inc : Int -> Int = \x -> x + 1` documents intent; it never changes evaluation. Run
`gatwd --check file.gatwd`: structural problems are errors (exit 1); arity, argument/result types,
operator misuse (`"a" - 1`), unbound names and bad numeric `spawn` seeds are advisory warnings.
Committed scripts must be warning-free (`Any` is consistent with everything, so valid dynamic code
is never flagged). Inside application heads/arguments unbound words are not reported because
lambda bodies can be pipeline-style (`echo hello %s`).

## 5. Procedure for a new test section

1. Add `-- NN. Title (what it proves)` and the cases to `tests/dsl_exhaustive_test.gatwd`; number
   after the last section (`grep -c '^-- [0-9]'`; do not copy counts into docs).
2. Extract your section into a temp file (prepend the `type` lines it needs), run `./gatwd` on it
   and verify **every** line printed a value and the value is what you meant (a wrong expectation
   shows `[ERROR] Assertion failed`; an assertion that never saw a value shows
   `[ERROR] assertion on an empty stream`; any other silence is a bug).
3. Include a **negative** case where behavior should differ (`assert false`, `assert []`).
4. Run `./run_dsl_tests.sh`, then `gatwd --check tests/dsl_exhaustive_test.gatwd`.

## 6. Reusable task scripts: parameterize by argument, never by editing the script

When a script defines a *view*, *transformation* or *query* meant to be applied to more than one
source (a different overlay, repo, file, cluster — "apply this to a variety of sources"), the
source is always a **runtime argument**, never text written into the script:

```gatwd
-- Good: the script names no path. Run as:
--   gatwd-<edition> task main -s script.gatwd -- /path/to/source-a
--   gatwd-<edition> task main -s script.gatwd -- /path/to/source-b
def main = manifests |> gitops_graph |> gp_trees |> visualize
```

```gatwd
-- Bad: this view now only works for one source, ever — it has to be *edited* to point anywhere
-- else, which defeats the entire point of a reusable script.
def main = spawn StrVal: /path/to/source-a |> manifests |> gitops_graph |> gp_trees |> visualize
```

This works because of the **fragment** convention (`pkg/cli/run.go`): a pipeline-bodied `def` run
as a task (`gatwd-<edition> task NAME -s script.gatwd -- arg ...`) gets the `--` arguments as its
*input stream*, one String node each — that is what `manifests` (and `walk_fs` used with no inline
argument) read their path from. `task` (the subcommand, not the bare `gatwd-<edition> script.gatwd`
form) also skips the startup banner, so a task's output pipes straight into `dot`/`jq`/etc.

A step that only accepts a **literal** argument (`walk_fs <dir>`, fixed at parse time) cannot be
parameterized this way, and a `-D NAME=VALUE` constant does not help either — constants substitute
into *expressions*, never into a step's literal text, so `walk_fs REPO` tries to walk a directory
literally named `REPO`. Per Mandate 1, the fix is never to hardcode the one path the script
happens to need right now; it is to give the step a stream-driven form (no inline argument, reads
its subject from the upstream value, exactly like `manifests` already does) so the existing
fragment mechanism parameterizes it for free. If the step you need doesn't have that form yet,
add it (`adapters/`, wired through `pkg/cli`) before reaching for a workaround in the script.

A view that genuinely consumes its argument (`manifests`, `walk_fs`) fixes itself this way: give it
a source and it runs, give it none and it correctly produces nothing — there is no source to view.
A fragment that has *no* per-run use for its argument at all (a fixed computation, not a view over
a source) still needs at least one `--` argument to fire, because the task's input stream has
exactly as many elements as arguments given and every step runs per upstream element: zero
arguments is an empty stream is no output, not a sensible default.
