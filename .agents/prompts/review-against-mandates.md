# Prompt: review a diff against the mandates

Review `<diff/branch>`. For each item answer **pass / fail / n.a.** with a file:line.

1. **M1 Anti-Mockup** — any Go loop, hard-coded sample or string hack standing in for a DSL
   capability? Is the missing primitive implemented generically instead?
2. **M2 Matrix** — new/changed operation: every valid type present? step *and* function form?
   function form data-last, curried, type-preserving, errors as values? skipped cells justified?
3. **M3 All-Layers** — reachable from the DSL? asserted in `tests/dsl_exhaustive_test.gatwd`,
   including a *non-domain* assertion for any core primitive?
4. **M4 Dogfood** — older code that should now use it? duplicated lookup/loop code introduced?
5. **M5 Clean Core** — domain word/import in core? (`go test ./internal/archtest`) shim has logic?
6. **M6 Fail Loud** — any path that yields nothing/`nil`/unchanged input on failure? DSL tests
   with possibly-empty output? "no violations" test without a negative fixture? guard
   negative-controlled?
7. **M7 Pure Assembly** — `main.go` free of domain logic? relations/rules in `.gatwd`? parity test?
8. **M8 Delivery** — decision recorded? deferrals written? one logical commit? handoff/design
   doc updated? nothing unrelated staged?
9. **Quality** — godoc/Example/doc.go; docs updated (no stale "not yet"); gofmt; gates green;
   race-clean if parallel; script regression diff clean; `gatwd --check` silent.

Report blockers first, then smaller findings. Quote evidence; do not speculate.
