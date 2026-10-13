# Prompt: add a language feature (syntax, value kind, step or builtin)

**Task:** `<describe the capability, and the program that needs it>`

**Read first:** `AGENTS.md` (Mandates 1, 2, 3, 6), `docs/grammar.md`, `parser/infix.go`
(tokens, `Expr` nodes, precedence, the `appDisabled` list-literal guard), `pkg/cli/cli.go`
(`parsePipelineSteps` step dispatch, `EvaluateLine`, `graphLookup`/`exprFallback`),
`pkg/cli/closures.go`, `pkg/cli/dslfuncs*.go`, `pkg/cli/typecheck.go`.

**Steps**
1. **State the missing primitive.** If an application motivated this, say what the DSL could
   not express. Do *not* work around it in Go (Mandate 1).
2. **Pick the right layer** for each piece: a value kind or ADT → `functions/`; a collection →
   `collections/`; a combinator → `arrows/`/`optics/`; a stream step → `evaluator/`; syntax →
   `parser/infix.go`; wiring → `pkg/cli`. Keep L3 parser-free (pass lambdas as Go closures).
3. **Write the pinning tests first**: a structural Go test on the raw type, a parser test if
   syntax changes (`parser/lambda_test.go` is the model), then the DSL assertions.
4. **Implement it generically** — nothing domain-shaped (Mandate 5; `internal/archtest` will fail).
5. **Wire it through**: step form in `cli.go` *and* function form via `defn` in `dslfuncs*.go`
   (curried, data-last, type-preserving, errors as values). Update name resolution so it works in
   every expression context (`exprFallback`), not just one.
6. **Check collateral damage**: a new token/operator can change how *existing* text parses (the
   `->` token, juxtaposition, the infix-step detector). Run the script regression diff.
7. **Teach the checker**: `pkg/cli/typecheck.go` must not warn on valid new syntax; extend it if
   the feature has a type story. `gatwd --check` stays silent on all repo scripts.
8. **Fail loud**: bad use returns an `[ERROR]`, never an empty result.
9. **Docs**: `docs/grammar.md`, `docs/dsl_guide.md`, `README.md` features, `Example*`, godoc.

**Definition of done:** structural test + parser test + DSL §N assertions (output verified
non-empty) + step *and* function form + `Example` + docs + all gates green + regression diff
clean + `gatwd --check` silent.

**Don't:** add a Go loop that fakes it · add a domain word to core · stringify values · leave a
type cell empty without a recorded mathematical reason.
