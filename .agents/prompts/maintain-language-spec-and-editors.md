# Maintain the language spec and the editor-support pipeline

Use when you add or change syntax, a keyword, a pipeline step, a builtin / function form, a
type name, or a snippet. One source feeds every editor and the parser's token tables, so the
spec is edited by hand and everything else is generated.

## The pipeline

```
spec/spec.go  (hand-edited: keywords, steps, types, operators, snippets)
      + cli.FunctionForms()  (the registered expression functions, read at generation time)
   -> go run ./cmd/gatwd-gen .
        editors/neovim/{syntax,ftdetect,lua/plugins,snippets}/…
        editors/vscode/syntaxes/gatwd.tmLanguage.json
        editors/treesitter/grammar.js
        parser/tokens_gen.go          (Go constants used by the parser)
```

Never edit a generated file: its header says `DO NOT EDIT`; the next run overwrites it.

## Steps

1. Read `spec/spec.go` and `cmd/gatwd-gen/main.go` (the `gen*` functions) and `docs/grammar.md`.
2. Change `spec/spec.go` (or register the builtin in `pkg/cli`; function forms are picked up
   automatically, steps and keywords must be listed in the spec).
3. If the syntax changed, update `docs/grammar.md` (EBNF; every non-terminal must be defined)
   and the DSL guide / `docs/reference.md` (its ```gatwd blocks run as tests).
4. Regenerate: `go run ./cmd/gatwd-gen .` and review `git diff` of the generated files; they
   should change only in the way you intended.
5. Snippets (`Snippets` in the spec) must be valid current gatwd: write them in the most
   readable form the language offers (clause defs, guards), and check each with
   `gatwd --check`.
6. Gates: `go test ./cmd/gatwd-gen ./parser ./pkg/cli ./internal/archtest` plus the standard
   gates. `cmd/gatwd-gen/main_test.go` fails when the spec misses a registered function form or
   pipeline step, or when `docs/grammar.md` and the spec disagree.
7. Do not hand-write editor files in `editors/` that the generator owns; add a new generator
   function instead, and a test next to it.

## Definition of done

- `go run ./cmd/gatwd-gen .` leaves a clean diff (idempotent) and nothing generated was edited by hand.
- The spec, `docs/grammar.md`, and every generated artifact agree (the gen tests pass).
- A negative control was run for any new coverage guard (omit a spec entry, watch it fail).
- A new keyword/step/type appears in highlighting, snippets (if useful), tree-sitter and the
  parser token constants.
