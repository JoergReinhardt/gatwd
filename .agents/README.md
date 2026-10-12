# Repository Agent Skills & Configurations

This directory contains repository-specific agent skills and rules for `gatwd`. Any AI coding assistant (Antigravity, Claude Code, Cursor, Windsurf, Copilot, etc.) operating in this repository can read and execute these skills.

## Available Skills

| Skill | Path | Description |
| :--- | :--- | :--- |
| **`recursive-dogfood`** | [`skills/recursive-dogfood/SKILL.md`](skills/recursive-dogfood/SKILL.md) | Eat your own dogfood recursively: audit older code, enforce pure DSL-first implementations over bare-minimum Go mockups, and ensure exhaustive operator-type matrix completeness. |
| **`all-layers-sync`** | [`skills/all-layers-sync/SKILL.md`](skills/all-layers-sync/SKILL.md) | Audit lower-level Go abstractions and wire them all the way through to the DSL as builtins with exhaustive tests in `tests/dsl_exhaustive_test.gatwd`. |
| **`project-doc-sync`** | [`skills/project-doc-sync/SKILL.md`](skills/project-doc-sync/SKILL.md) | Audit and synchronize godoc documentation, package `doc.go` overviews, and testable `Example*` functions. |

For detailed guidelines and architectural invariants, refer to [`AGENTS.md`](../AGENTS.md) and [`PROMPTS.md`](../PROMPTS.md).
