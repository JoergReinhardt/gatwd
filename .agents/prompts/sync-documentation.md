# Prompt: synchronise documentation after a change

**Task:** bring docs in line with `<feature/change set>`.

**Read first:** `.agents/skills/project-doc-sync/SKILL.md`, `docs/development_playbook.md` §8.

**Checklist**
1. godoc on every new exported identifier (starts with its name; explains concept → why → how →
   example); package `doc.go` updated; `Example*` with `// Output:`.
2. `docs/grammar.md` for any syntax; `docs/dsl_guide.md` for behavior/builtins (remove any
   "not yet supported" that is now false — grep for it); `docs/components.md` for types/methods;
   `docs/architecture.md` (additive edits only — other sessions edit it); `README.md` features.
3. Design doc: phase entry → DONE with what changed; Appendix B ticked; remaining ideas listed.
   Roadmap/status blocks updated.
4. Handoffs: mark superseded ones; never leave two "current" guides.
5. **Rot check** — grep docs for hard-coded counts ("34-step", "N tutorials"), `legacy/` references,
   stale package lists, "planned"/"not yet" for things that now exist; verify every relative link
   resolves.
6. Dated snapshots (`assessment.md`, `report.md`) carry a "historical" banner instead of edits.

**Definition of done:** every claim in the changed docs was checked against the tree; links
resolve; no hard-coded counts introduced.
