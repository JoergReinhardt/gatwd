---
name: phase-delivery
description: "Deliver larger work as gated phases: grounded design doc, generic prerequisites before domain code, ask-and-record decision points, written deferrals, one commit per sub-step with all gates green, performance protocol, handoffs, and safe behavior in a shared working tree."
---

# Phase Delivery Skill

Use for any change bigger than a bug fix, and whenever a project spans contexts. Long form:
`docs/development_playbook.md` §§2, 3, 7, 9, 10. Rule: `AGENTS.md` Mandate 8.

## Procedure
1. **Ground the design in the code.** Open the grammar, engine types and the machinery you will
   extend; put exact signatures and file anchors in an appendix of `docs/<topic>_design.md`.
2. **Split generic from domain** and make generic work numbered *prerequisite phases* (P1, P2…).
3. **Plan phases small enough to gate**: each ends with the full verification gates (`AGENTS.md`
   §5) green and one commit `feat(scope): summary (<project> Phase N)` whose body lists what was
   added, decisions, tests, and "all gates green".
4. **Decision points** — when the choice is plausible-either-way and hard to reverse (dependency
   vs external binary; stream-merge strategy; data representation): present options, give a
   recommendation, **ask**, and record the decision and its date. If the user has delegated, take
   the recommended option and say so.
5. **Deferrals are written** in the phase entry ("deferred, not silent", with reason).
6. **Verify like a skeptic**: pinning tests first; negative controls for new guards; the script
   regression diff for interpreter/parser/evaluator changes (baseline binary from a `git worktree`;
   run baseline several times — some scripts are nondeterministic on their own); `-race` for
   parallel code; `gatwd --check` silent over all scripts.
7. **Performance**: measure → profile per stage → pin semantics → fix at the owning layer →
   re-measure → record (including negative results).
8. **Close the phase**: design-doc entry DONE (what changed, tests, remaining ideas), Appendix B
   ticked, roadmap/status blocks updated, superseded handoffs bannered or deleted.
9. **Hand off** with `.agents/prompts/phase-handoff.md` when a context ends mid-project; state what
   is uncommitted/unpushed.

## Shared working tree etiquette
Other sessions or people may edit the same tree (e.g. library-extraction work, docs). Therefore:
re-read a file immediately before editing; prefer additive edits; never revert a change you did not
make; stage by explicit path (never `git add -A`; `.claude/`, scratch docs and `output.txt` are
not yours); do not commit or push unless asked; if you find unexpected files, investigate before
touching them.

## Anti-patterns
Stacking unverified work · deciding an irreversible design silently · a handoff with line numbers
and HEAD hashes · optimising the stage you guessed · leaving two "current" handoffs · declaring a
phase done with its Appendix B checklist unticked.
