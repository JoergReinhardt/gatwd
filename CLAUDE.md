# CLAUDE.md

Claude Code configuration for **gatwd** — an industrial-strength, mathematically
rigorous, pure functional programming language, execution engine, and dynamic
infrastructure graph processor written in Go 1.23+.

This file is the Claude Code entry point. The authoritative agent instructions,
prompts, and skills live in the repository's **vendor-neutral** agent resources
under `.agents/` (readable by any AI assistant — Antigravity, Claude Code,
Cursor, Windsurf, Copilot). They are imported below rather than duplicated, so
there is a single source of truth.

## Imported agent resources (single source of truth)

- @AGENTS.md — architecture manual, layered hierarchy, the 8 mandates, prevention rules
- @PROMPTS.md — refactoring/extension instructions and the documentation mandate
- @STYLEGUIDE.md — Go-idiomatic + Category-Theory naming, doc, test and DSL-style conventions
- @.agents/README.md — index of the in-repository agent skills and task prompts

Read on demand (not imported, to keep context small):

- `docs/development_playbook.md` — how work is done here and why: delivery protocol,
  gates, the Edition/Appliance/shim pattern, DSL traps, perf and handoff protocols,
  and the k8s-graph case study.
- `.agents/prompts/` — ready-made task prompts (language feature, matrix cell,
  relation/rule, Edition/Appliance, performance, handoff, docs sync).

## Skills

The repository skills are available to Claude Code under `.claude/skills/`
(symlinked to their canonical home in `.agents/skills/`):

- **recursive-dogfood** — audit older code for newer abstractions, enforce pure
  DSL-first (anti-mockup) implementations, verify full operator-type matrix
  completeness.
- **all-layers-sync** — wire lower-level Go abstractions all the way through to
  the DSL (step form *and* curried function form) with exhaustive tests in
  `tests/dsl_exhaustive_test.gatwd`.
- **project-doc-sync** — maintain godoc on exported symbols, package `doc.go`
  overviews, and testable `Example*` functions; eliminate legacy drift, stale
  counts and superseded handoffs.
- **dsl-authoring** — write pure-DSL programs/relations/rules in the project's
  idioms and avoid the known traps.
- **edition-appliance** — mint an Edition/Appliance: adapter → thin shim →
  embedded prelude → pure-assembly main, with parity and negative-fixture tests.
- **phase-delivery** — run larger work as gated phases with decision points,
  recorded deferrals, perf protocol and handoffs.

## The 8 mandates (quick reference — see AGENTS.md for detail)

1. **Anti-Mockup / Hard Work First** — never fake a Gatwd solution with Go loops;
   implement missing DSL/engine primitives first, then solve purely in the DSL.
2. **Exhaustive Operator-Type Matrix** — every operator/combinator covers the full
   Cartesian product of types where valid, in step form *and* as a curried,
   data-last, type-preserving function form.
3. **All-Layers-Through** — wire new capabilities through to Layer 5 as builtins;
   test them in the DSL, with a non-domain assertion for every core primitive.
4. **Recursive Dogfooding** — revisit older code to adopt newer abstractions.
5. **Clean Core Prevention** — core imports only stdlib + this module and names no
   domain; enforced by `internal/archtest`. Generic first, domain in a shim.
6. **Fail Loud** — no silent failures; an assert on an empty stream passes
   vacuously, so pair rules with negative fixtures and negative-control guards.
7. **Pure Assembly** — Editions/Appliances/shims wire; relations and rules live
   in `.gatwd`; prove live/offline parity with fakes.
8. **Phase-Gated, Decision-Recorded Delivery** — small gated phases, ask at real
   decision points and record them, write deferrals down, hand off in writing,
   and treat the working tree as shared (re-read before editing, never revert
   others' changes, commit/push only when asked).

## Layered hierarchy (import only downward: Layer N ← 0..N-1)

| Layer | Packages |
| :--- | :--- |
| 0 | `functions/`, `dsltypes/` |
| 1 | `collections/` |
| 2 | `arrows/`, `comonads/`, `optics/`, `transducers/`, `transformers/` |
| front end | `parser/` (imports only L0–1; consumed by L3 and L5) |
| 3 | `evaluator/` |
| 4 | `adapters/`, `interactive/`, `visualizer/`, `pkg/plugins/` |
| 5 | `pkg/cli/`, `pkg/<domain>/` (e.g. `pkg/k8sgraph`), `cmd/` |

## Verification (run before every commit)

```bash
gofmt -l <changed dirs>  # must list nothing for files you touched
go vet ./...             # compile/vet + no circular imports
go test ./...            # unit, structural, Example*, and internal/archtest (layers, Clean Core, doc links)
./run_dsl_tests.sh       # DSL suite: fails on [ERROR], on silent statements, on --check errors
go test -race ./pkg/... ./evaluator ./parser   # when touching interpreter/parallel code
```

Interpreter, parser or evaluator changes also need the script regression diff and a
silent `gatwd --check` over every `*.gatwd` (see the playbook §3).

## Working notes for Claude Code

- Prefer `go doc` / `gopls` over grep for navigation and renames (AGENTS.md §6).
- Write scratch DSL scripts with `cat > f <<'EOF'` (zsh `echo` mangles `\x`), and
  read every output line of a new DSL assertion — empty output passes `assert`.
- This tree is often shared with another session (e.g. library-extraction work):
  re-read a file right before editing it and keep edits additive.
- Do not commit or push unless asked; stage by name (`.claude/`, scratch docs and
  `output.txt` are not yours to add).
