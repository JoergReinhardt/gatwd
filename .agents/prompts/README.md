# Task Prompts

Self-contained prompts for the recurring kinds of work in `gatwd`. Each one names the
files to read first, the steps, and a **definition of done**. They distil how the
language and the Kubernetes-graph tool were actually built (see
[`docs/development_playbook.md`](../../docs/development_playbook.md)); the binding rules
are in [`AGENTS.md`](../../AGENTS.md).

Usage (human or agent): copy the matching prompt, fill the `<…>` blanks, and keep the
definition of done as the acceptance test. Always finish with the verification gates
(`AGENTS.md` §5).

| Prompt | Use it when |
| :-- | :-- |
| [`add-language-feature.md`](add-language-feature.md) | new syntax, value kind, step, or builtin in the DSL |
| [`fill-matrix-cell.md`](fill-matrix-cell.md) | an operation exists for some types but not others |
| [`add-dsl-relation-or-rule.md`](add-dsl-relation-or-rule.md) | a new graph relation, compliance rule or ingest in a domain prelude |
| [`mint-edition-appliance.md`](mint-edition-appliance.md) | a new domain tool: adapter → shim → prelude → Edition/Appliance |
| [`investigate-performance.md`](investigate-performance.md) | something is slow or scales badly |
| [`maintain-language-spec-and-editors.md`](maintain-language-spec-and-editors.md) | syntax, keyword, step, builtin or snippet changed: spec → generator → editors/tokens |
| [`phase-handoff.md`](phase-handoff.md) | a context is ending mid-project and a fresh one must continue |
| [`sync-documentation.md`](sync-documentation.md) | after a feature lands, or before a release |
| [`review-against-mandates.md`](review-against-mandates.md) | reviewing a diff (your own or someone else's) |

Skills with the full procedures: [`../skills/`](../skills/) — `recursive-dogfood`,
`all-layers-sync`, `project-doc-sync`, `dsl-authoring`, `edition-appliance`, `phase-delivery`.
