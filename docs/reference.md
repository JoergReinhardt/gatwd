# gatwd Language Reference

A friendly, example-driven reference for the `gatwd` scripting language. It assumes you can
read a little code in *some* language, but **not** that you know functional programming.
Every concept is explained in plain words first, then shown.

> **These examples are tested.** Every `gatwd` code block in this file is executed by
> `pkg/cli/reference_test.go` on each `go test`; a block that errors, prints nothing, or fails
> `gatwd --check` fails the build. Blocks marked `text` are illustrations only.
> For the formal grammar see [grammar.md](grammar.md); for a one-page tour see
> [`examples/scripts/cheatsheet.gatwd`](../examples/scripts/cheatsheet.gatwd);
> for the editor/completion tables see [Keeping tooling in sync](#14-keeping-editor-tooling-in-sync).

Run a script with `gatwd file.gatwd`, check it without running with `gatwd --check file.gatwd`,
and list every built-in with `gatwd builtins`.

---

## 1. The big idea in 60 seconds

In gatwd you do not change things; you **transform values** and pass the result on.

* A **value** is data: `5`, `"hello"`, `[1, 2, 3]`.
* A **function** turns a value into another value.
* A **pipeline** (`|>`) feeds the output of one step into the next, like a conveyor belt.
* Values are **immutable**: "adding" to a list gives you a *new* list; the old one is untouched.
  This makes programs predictable and safe to run in parallel.

```gatwd
-- "take 5, add 1, multiply by 2, then check the answer is 12"
5 |> + 1 |> * 2 |> assert 12
```

`assert` is your friend: it checks a value and turns into an `[ERROR]` if it is wrong.
Examples in this document use it so they prove themselves.

Comments start with `--` and run to the end of the line.

---

## 2. Values and literals

| Kind | Examples | Notes |
| :-- | :-- | :-- |
| Int | `42`, `-7` | Overflow promotes to a big integer; it never wraps. |
| Float | `3.14`, `2.5` | |
| Ratio | `3r4`, `0.75r` | Exact fractions, no rounding error. |
| Complex | `2i`, `1.5i` | |
| Bool | `true`, `false` | |
| String | `"hello"` | Escapes: `\n \t \" \\ \x41 \u00e9`. |
| Char | `'a'`, `'\n'` | One Unicode character. |
| Duration | `5s`, `1h30m`, `2d` | `d` = 24 h, `w` = 7 d. |
| Date / Time | `@2026-10-13`, `@2026-10-13T12:00:00Z` | |
| List | `[1, 2, 3]` | Ordered, immutable. |

```gatwd
7 / 2 |> assert 3.5
7 // 2 |> assert 3
7 % 2 |> assert 1
2 ** 10 |> assert 1024
"ab" ++ "cd" |> assert abcd
[1, 2] ++ [3] |> assert [1 2 3]
```

`/` is true division, `//` whole-number division, `%` remainder, `**` power, and `++` joins
(strings, lists, sets, maps, graphs).

### Operators, loosest to tightest

| Operators | Meaning |
| :-- | :-- |
| `\|>` `\|\|>` | pipe; parallel pipe (section 4) |
| `\|\|` | or |
| `&&` | and |
| `==` `!=` | equal, not equal (compares structure, not identity) |
| `<` `<=` `>` `>=` | ordering |
| `+` `-` `++` | add, subtract, concatenate |
| `*` `/` `//` `%` | multiply, divide, whole-divide, remainder |
| `**` | power (right-associative) |
| `-x` `!x` | negate, logical not |

Mixing types that make no sense (`"a" - 1`) is an error, not a silent guess.

---

## 3. Names: `let`, `const`, `def`

```gatwd
let x = 10 in x * 2 |> assert 20
```

* `let name = value in expression` — a name that only exists inside that expression.
* `const NAME = value` — a name for the whole script.
* `def name = ...` — defines a reusable function (next section).

```gatwd
const LIMIT = 3
spawn Int: 0 |> emit 1, 2, 3, 4, 5 |> drop 1 |> take LIMIT |> collect_list |> assert [1 2 3]
```

---

## 4. Pipelines

`a |> step` means "give `a` to `step`". Steps chain left to right.

```gatwd
spawn Int: 0 |> emit 5, 3, 8, 1 |> drop 1 |> sort |> take 2 |> collect_list |> assert [1 3]
[1, 2, 3, 4] |> filter (\x -> x > 2) |> map (\x -> x * 10) |> assert [30 40]
```

There are two kinds of step. *List functions* (`map`, `filter`, `fold`, `flatmap`, `reject`)
work on a list value. *Stream steps* (`take`, `drop`, `sort`, `distinct`, `chunk`, `head`) work on
a stream of separate values, which you create with `emit a, b, c` and gather again with
`collect_list`. Because the `spawn` node is the first element of the stream, `drop 1` removes it.

Common steps: `take n`, `drop n`, `map f`, `filter p`, `reject p`, `sort`, `distinct`,
`fold f seed`, `flatmap f`, `head`, `collect_list`, `chunk n`.

`||>` is the **parallel pipe**: the next per-element step runs on many workers at once.
The order of results is not guaranteed, so use it for independent work.

```gatwd
[1, 2, 3] ||> map (\x -> x * x) |> sort |> assert [1 4 9]
```

A pipeline can also start from `spawn Type: value`, which creates a typed first node; the
first element of the stream *is* that node:

```gatwd
type IntVal = Int
spawn IntVal: 10 |> + 5 |> * 2 |> assert 30
```

---

## 5. Functions, lambdas and currying

A **lambda** is a function without a name: `\x -> x + 1` ("a function of `x` that returns
`x + 1`"). Several parameters are separated by spaces.

```gatwd
def add = \x y -> x + y
add 1 2 |> assert 3
```

**Currying** (the one scary word) simply means: *if you give a function fewer arguments than
it wants, you get back a smaller function that waits for the rest.*

```gatwd
def add = \x y -> x + y
let inc = add 1 in inc 5 |> assert 6
```

`add 1` is "the function that adds one". That is exactly what you want inside a pipeline,
because the piped-in value is always passed as the **last** argument ("data-last"):

```gatwd
def add = \x y -> x + y
5 |> add 1 |> assert 6
[1, 2, 3] |> map (add 10) |> assert [11 12 13]
[1, 2, 3] |> fold add 0 |> assert 6
```

Functions can take and return functions, and close over names around them:

```gatwd
def compose = \f g x -> f (g x)
def add = \x y -> x + y
(compose (add 1) (add 2)) 3 |> assert 6
```

**Parentheses** group a call: `f (g x)`. Inside a plain list literal, parenthesise calls too:
`[(f x), (g y)]`, otherwise the list is split on spaces.

> `def name = ...` with a *lambda* is a real function value. `def name = step |> step` with a
> *pipeline* is a reusable pipeline fragment (a text macro), so keep those without parameters.

### Type annotations (optional)

```gatwd
def inc : Int -> Int = \x -> x + 1
inc 41 |> assert 42
```

Annotations document intent. `gatwd --check` reads them and warns about mismatches; they never
change how the program runs.

---

## 6. Lists and comprehensions

```gatwd
[1, 2, 3] |> map (\x -> x * 2) |> assert [2 4 6]
length [10, 20, 30] |> assert 3
reverse [1, 2, 3] |> assert [3 2 1]
```

A **comprehension** builds a list from other lists, reading like set notation
"the squares of the x in 1..4 where x > 1":

```gatwd
[x * x | x <- [1, 2, 3, 4], x > 1] |> assert [4 9 16]
[[a, b] | a <- [1, 2], b <- [10, 20]] |> assert [[1 10] [1 20] [2 10] [2 20]]
```

`x <- list` is a *generator* (take each element); a bare condition is a *guard* (keep only if
true). Use `_ <- [0]` when you only want guards.

`flatmap` maps and then flattens one level:

```gatwd
[1, 2, 3] |> flatmap (\x -> [x, x * 10]) |> assert [1 10 2 20 3 30]
```

---

## 7. Types and pattern matching

Declare your own types:

```gatwd
type OS = Linux | Windows | Darwin
type Config = { IP: String, Port: Int }
type Ints = List(Int)
```

`match` chooses a branch by the *shape* of a value. Arms are separated by `|`:

```gatwd
type OS = Linux | Windows | Darwin
def describe = match { Linux -> echo PENGUIN | Windows -> echo WINDOW | Darwin -> echo APPLE }
spawn OS: Linux |> describe |> assert PENGUIN
```

Lists can be taken apart too: `[]` is empty, `[head, ...rest]` splits first from the others, and
`[a, b]` matches exactly two elements. Together with recursion this replaces loops:

```gatwd
type IntList = List(Int)
def quicksort = match {
    [] -> empty
    | [pivot, ...rest] -> seq { [x | x <- rest, x < pivot] |> quicksort | pivot | [x | x <- rest, x >= pivot] |> quicksort }
}
spawn IntList: [5, 2, 8, 1, 9, 3] |> quicksort |> collect_list |> assert [1 2 3 5 8 9]
```

`gatwd --check` warns when the arms of a `match` over a declared type do not cover all cases.

---

## 8. Collections

All collections are **persistent**: operations return a new collection and share structure with
the old one, so they are cheap.

| Collection | Think of it as | Build | Some functions |
| :-- | :-- | :-- | :-- |
| List | ordered sequence | `[1, 2]` | `head tail nth length map filter fold reverse` |
| Set | unique items | `to_set` | `set_add set_contains set_union set_intersect set_diff set_elements` |
| HAMT | immutable dictionary | `hamt_from`, `to_hamt` | `hamt_get hamt_set hamt_del hamt_has hamt_keys hamt_values hamt_merge` |
| Tree / BST | hierarchy / sorted tree | `build_tree`, `build_bst` | `tree_children tree_value tree_depth bst_insert bst_member` |
| Graph | nodes + labelled edges | `graph_empty` | see section 10 |
| Chunk | contiguous batch | `to_chunk`, `chunk n` | same as List |

```gatwd
length (set_elements (to_set [3, 1, 3, 2])) |> assert 3
set_contains 2 (to_set [1, 2, 3]) |> assert true
```

A dictionary (HAMT) keyed by strings:

```gatwd
let d = hamt_set "name" "ada" hamt_empty in hamt_get "name" d |> assert ada
```

Most functions put the collection **last**, so `hamt_get "name" d` also reads
`d |> hamt_get "name"`.

---

## 9. Maybe, Either and Pair — values that carry context

Instead of `null` or exceptions, gatwd has small wrapper values.

* **Maybe** — `just 5` ("there is a value") or `nothing` ("there is none").
* **Either** — a result that is `either_left` (an error) or `either_right` (success).
* **Pair** — two values glued together: `pair 1 2`, read with `fst` and `snd`.

```gatwd
fst (pair 1 2) |> assert 1
snd (pair 1 2) |> assert 2
is_just (just 5) |> assert true
is_nothing nothing |> assert true
maybe_default 0 nothing |> assert 0
```

Errors in a pipeline are themselves values: a failing step produces an `[ERROR] message` node
that the following steps skip, so one failure never silently turns into a wrong answer.

---

## 10. Graphs

A graph is a set of **nodes** joined by labelled, weighted **edges**. Add nodes **first**, then
edges between nodes that exist.

```gatwd
spawn S: x |> graph_empty |> graph_add_node N a |> graph_add_node N b |> graph_add_node N c |> graph_add_edge Owns 1 a b |> graph_add_edge Owns 1 b c |> graph_shortest_path a c |> collect_list |> assert [a b c]
```

`graph_add_node Color id`, `graph_add_edge Color weight from to`. Queries include
`graph_reachable`, `graph_scc` (strongly connected components), `graph_topo_sort`,
`graph_neighbors`, `graph_bfs`/`graph_dfs`, plus `graph_map_nodes`, `graph_filter_nodes` and
folds. Name nodes `Kind/scope/name` so two different things called "web" stay distinct.

---

## 11. Arrows — fanning out and merging

These combine several pipelines.

| Form | Meaning |
| :-- | :-- |
| `(f &&& g)` | send one value through both, producing a pair |
| `[f *** g]` | apply `f` to the first half of a pair and `g` to the second |
| `(a <+> b <+> c)` | run the value through every branch and concatenate the outputs |

```gatwd
type IntVal = Int
spawn IntVal: 10 |> (math_add 10 &&& math_add 20) |> format_tuple "%v,%v" |> assert 20,30
spawn IntVal: 5 |> (math_add 1 <+> math_add 10 <+> math_add 100) |> collect_list |> assert [6 15 105]
```

---

## 12. Strings, text, time and more (function forms)

Besides pipeline steps, gatwd has many *functions* you can call anywhere an expression is
allowed, including inside lambdas. All are curried and data-last. Highlights:

| Area | Functions |
| :-- | :-- |
| Strings | `str_split str_join str_trim str_upper str_lower str_replace str_contains str_starts_with str_slice str_lines str_words str_format` |
| Regex | `re_match re_find re_find_all re_replace re_split re_groups` |
| Chars | `char_upper char_is_digit char_code char_of_str` |
| Numbers | `abs min max pow sqrt floor ceil round gcd lcm divmod isqrt to_int to_float` |
| Ratios | `ratio numer denom ratio_parts to_ratio` |
| Time | `time_parse time_parse_with time_format time_add date_parse date_format dur_parse to_seconds` |
| Data formats | `json_parse json_encode json_pretty yaml_parse yaml_encode hex_encode base64_encode sha256 md5` |

```gatwd
str_split "," "a,b,c" |> assert [a b c]
str_join "-" ["a", "b"] |> assert a-b
str_upper "go" |> assert GO
re_match "^a+$" "aaa" |> assert true
```

Run `gatwd builtins` for the full, always-current list with parameter names.

---

## 13. Testing and failing loudly

* `assert value` — passes the stream on only if it equals `value`.
* `expect "WORD" "message"` — passes a node containing WORD, otherwise yields `[ERROR] message`.
* An assertion on an **empty** stream is an error: a pipeline that produces nothing is never
  mistaken for success.

```text
1 / 0                                  -- [ERROR] division by zero
[1, 2] |> map (\x -> nosuchfn x)       -- [ERROR] undefined identifier: nosuchfn
spawn IntVal: 1 |> filter "zzz" |> assert 1   -- [ERROR] assertion on an empty stream
```

Scripts receive their arguments as the list `ARGS`; any `def` is a task you can run with
`--task NAME`, and a `def main = \args -> ...` runs by default.

---

## 14. Keeping editor tooling in sync

Syntax highlighting, completion and snippets (Neovim, VS Code, tree-sitter) are generated, not
hand-written:

```text
go generate ./parser        # or: go run ./cmd/gatwd-gen
scripts/update_neovim_support.sh
```

* `spec/spec.go` lists keywords, types, operators and **pipeline steps** by hand.
* Every **function form** is read straight from the interpreter's registry
  (`cli.FunctionForms`), so new functions appear automatically.
* `cmd/gatwd-gen/main_test.go` fails when a step in `pkg/cli/cli.go` is missing from the spec,
  when a function form is missing, or when `docs/grammar.md` omits a keyword/operator or uses an
  undefined non-terminal.
* The EBNF itself lives only in [grammar.md](grammar.md); the parser is hand-written, so the
  grammar is kept honest by those tests rather than generated.
