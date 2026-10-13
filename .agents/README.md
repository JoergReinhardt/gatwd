# Repository Agent Skills & Configurations

This directory contains repository-specific agent skills, task prompts and rules for `gatwd`. Any AI coding assistant (Antigravity, Claude Code, Cursor, Windsurf, Copilot, etc.) operating in this repository can read and execute them. It is the single source of truth: tool-specific entry points (`CLAUDE.md`, `.claude/skills/`) only import or symlink to it.

## Where to start

1. [`AGENTS.md`](../AGENTS.md) — layers, the eight mandates, verification gates, prevention rules.
2. [`docs/development_playbook.md`](../docs/development_playbook.md) — how work is done here and why (delivery protocol, gates, Edition/Appliance pattern, DSL traps, performance and handoff protocols, the k8s-graph case study).
3. The skill or prompt below that matches your task.

## Available Skills (`skills/`)

| Skill | Path | Description |
| :--- | :--- | :--- |
| **`recursive-dogfood`** | [`skills/recursive-dogfood/SKILL.md`](skills/recursive-dogfood/SKILL.md) | Eat your own dogfood recursively: audit older code, enforce pure DSL-first implementations over bare-minimum Go mockups, and ensure exhaustive operator-type matrix completeness (step and function form). |
| **`all-layers-sync`** | [`skills/all-layers-sync/SKILL.md`](skills/all-layers-sync/SKILL.md) | Wire lower-level Go abstractions all the way through to the DSL as builtins — step form and curried function form — with exhaustive, non-vacuous tests in `tests/dsl_exhaustive_test.gatwd`. |
| **`project-doc-sync`** | [`skills/project-doc-sync/SKILL.md`](skills/project-doc-sync/SKILL.md) | Synchronize godoc, `doc.go`, `Example*`, grammar/DSL guide/components docs; rot check for stale counts, false "not yet" claims, broken links and superseded handoffs. |
| **`dsl-authoring`** | [`skills/dsl-authoring/SKILL.md`](skills/dsl-authoring/SKILL.md) | Write pure-DSL programs, relations and rules in the project's idioms and avoid the known traps. |
| **`edition-appliance`** | [`skills/edition-appliance/SKILL.md`](skills/edition-appliance/SKILL.md) | Mint an Edition/Appliance: adapter → zero-logic shim → embedded prelude → pure-assembly main, with parity and negative-fixture tests. |
| **`phase-delivery`** | [`skills/phase-delivery/SKILL.md`](skills/phase-delivery/SKILL.md) | Run larger work as gated phases: grounded design doc, decision points, deferrals, perf protocol, handoffs, shared-tree etiquette. |

## Task Prompts (`prompts/`)

Ready-to-fill prompts with a definition of done — see [`prompts/README.md`](prompts/README.md): add a language feature · fill a matrix cell · add a DSL relation or rule · mint an Edition/Appliance · investigate performance · write a phase handoff · sync documentation · review against the mandates.

## Machine-enforced rules

Rules that are cheaper to enforce than to remember live in tests/scripts, not prose:

- `internal/archtest` — library-layer import boundaries, the Clean Core rule (no third-party import, no domain identifier in core), and documentation links (no broken relative links, no machine-specific absolute paths).
- `run_dsl_tests.sh` — fails on `[ERROR]`, on statements that produce no output (vacuous asserts), and on `gatwd --check` structural errors.
- `gatwd --check` — structural errors plus advisory typing/exhaustiveness warnings; must stay silent on committed scripts.

For detailed guidelines and architectural invariants, refer to [`AGENTS.md`](../AGENTS.md), [`PROMPTS.md`](../PROMPTS.md) and [`STYLEGUIDE.md`](../STYLEGUIDE.md).
