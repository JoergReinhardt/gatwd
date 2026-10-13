# Copilot instructions for gatwd

This file is a pointer only. The single source of truth is vendor-neutral and lives in the
repository; read these before working (do not duplicate their content here):

- `AGENTS.md` — layers, the 8 mandates, verification gates (§5), prevention rules (§5a), Go tooling (§6)
- `PROMPTS.md` — refactoring/extension rules and documentation mandate
- `STYLEGUIDE.md` — Go idioms, Category Theory naming, test and DSL style
- `.agents/README.md` — index of skills and prompts
- `.agents/skills/*/SKILL.md` — `recursive-dogfood`, `all-layers-sync`, `project-doc-sync`,
  `dsl-authoring`, `edition-appliance`, `phase-delivery` (also exposed via `.github/skills` and `.claude/skills` symlinks)
- `.agents/prompts/` — ready-made task prompts (language feature, matrix cell, relation/rule, edition, perf, handoff, docs sync, review)
- `docs/development_playbook.md` — how work is done here and why

Before every commit run: `gofmt -l`, `go vet ./...`, `go test ./...`, `./run_dsl_tests.sh`
(plus `go test -race ./pkg/... ./evaluator ./parser` for interpreter/parallel changes).
Do not commit or push unless asked; stage files by name; the working tree may be shared.
