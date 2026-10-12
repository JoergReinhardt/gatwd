# New Language Features Showcase

A single, self-asserting `.gatwd` script that exercises every language feature added
in the constructor-matching / first-class-lambdas / matrix-completeness work. Every
line ends in `|> assert <expected>`, so a clean run means every feature behaves.

```bash
# Build the interpreter
go build -o gatwd ./cmd/gatwd/...

# Run (prints each asserted value; any failure shows [ERROR] Assertion failed)
./gatwd examples/new_features/showcase.gatwd

# Statically lint without running it
./gatwd --check examples/new_features/showcase.gatwd
```

## What it demonstrates

| # | Feature | Example |
| :--- | :--- | :--- |
| 1 | **ADT constructor matching** | `match { Red -> … \| Green -> … }` routes each sum-type branch to its own arm |
| 2 | **First-class lambdas** | `def inc = \x -> x + 1` then `5 \|> inc`; `\|> (\x -> …)`; `let f = \x -> … in`; `map inc` |
| 3 | **`reject`** | negated filter over List, stream, and Tree |
| 4 | **`flatmap`** | monadic bind — flatten per-element collections |
| 5 | **Custom fold** | `fold (\acc x -> expr) seed` — any left fold |
| 6 | **`bimap`** | Bifunctor map over a 2-element pair / `Pair` / `Either` |
| 7 | **HAMT set-algebra** | `hamt_intersect`, `hamt_diff` (keyed) |
| 8 | **HAMT filter** | `filter` entries of a HAMT by value |

## Also available at the CLI

- **Loud errors.** An unknown step fails visibly instead of silently:
  `spawn IntVal: 5 |> frobnicate` → `[ERROR] unknown step or command: "frobnicate"`.
- **`--check`.** Structural validation (unbalanced delimiters, missing `=`, dangling `|>`)
  plus **non-exhaustive-`match` warnings** over closed sum types, e.g.
  `warning: non-exhaustive match on OS: missing Darwin`.

See [`docs/dsl_guide.md`](../../docs/dsl_guide.md) for the full language reference and
[`docs/language_roadmap.md`](../../docs/language_roadmap.md) for the design rationale.
</content>
