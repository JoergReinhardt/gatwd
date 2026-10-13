---
name: edition-appliance
description: "Mint a Gatwd Edition or Appliance for a domain the right way: generic prerequisites in core, a real-I/O adapter with an offline twin, a zero-logic shim, an embedded DSL prelude, a pure-assembly main, and tests for live/offline parity, negative fixtures and hot reload."
---

# Edition / Appliance Skill

Reference implementation: `adapters/kubernetes_adapter.go`, `adapters/manifest_adapter.go`,
`pkg/k8sgraph/`, `cmd/gatwd-k8s/`, `cmd/gatwd-k8s-operator/`. Rules: `AGENTS.md` Mandates 5 and 7.
Prompt: `.agents/prompts/mint-edition-appliance.md`.

## Vocabulary
- **Edition** — a binary distribution users run arbitrary scripts against
  (`cli.NewEdition(name)…Run()`).
- **Appliance** — a single-purpose release with embedded rules and hot reload
  (`cli.NewAppliance(WithEdition, WithEmbeddedFS, WithConfigMapOverride)`).
- **Shim** — `pkg/<domain>`: wiring only.
- **Prelude** — `lib/<domain>/*.gatwd` embedded with `//go:embed`: the domain's vocabulary and
  logic, loaded through `Edition.WithPreludeFS`.

## Layout
```
adapters/<domain>_adapter.go        real I/O; SDK imports confined here; structured payloads; errors → Left
adapters/<domain>_manifest*.go      offline twin producing the SAME payload shape
pkg/<domain>/steps.go               <Domain>Step(keyword, client, …) = cli.AdapterStep(adapter)
pkg/<domain>/edition.go             ApplyEdition(e, client): steps + prelude; nil client ⇒ offline-only
pkg/<domain>/lib/<domain>/*.gatwd   types, ingest, relations, queries, sources, prelude (entry)
pkg/<domain>/<domain>test/          fake client built from the same fixtures (test support)
cmd/gatwd-<domain>/main.go          Edition: pure assembly
cmd/gatwd-<domain>-<role>/          Appliance: edition + rules/*.gatwd + flags; testable helpers
```

## Procedure
1. **Design doc first**, with decisions recorded (ask the user about irreversible ones).
2. **Generic prerequisites** land in core before any domain code and are asserted with *non-domain*
   DSL tests (`<+>` stream merge, `AdapterStep`, `WithPreludeFS`, `hamt_path`, `subset_of`, …).
3. **Adapter**: yield nested `HAMT`/`List` payloads, not names; wrap list/IO errors as error nodes;
   a pluggable seam (`WithRenderer`) beats a heavy dependency for external tools.
4. **Shim**: no graph or relation logic. Keep a `CoreCrawls`-style table (keyword → resource) so
   `ApplyEdition`, the prelude's `sources.gatwd` and the parity test cannot drift.
5. **Prelude**: relations are graph→graph `fold` passes; rules are closures; merge producers with
   `(a <+> b <+> …)`. Use the DSL skill (`dsl-authoring`). **Node ids are qualified**
   (`Kind/scope/name`, `_` for no scope — see `lib/k8s/ids.gatwd`), ingest uses
   `graph_insert_node` so a duplicate id is an error, and references resolve in the referrer's
   scope (`k8s_ref`) and only join existing nodes.
6. **Main**: assembly only. No domain loop, relation or type. Helpers for flags/pipeline strings
   are fine and should be unit-tested (`newAppliance`, `checkPipeline`, `livePipeline`).
7. **Tests**
   - fixtures under `testdata/`; a *full* fixture touching every relation;
   - `TestLiveOfflineParity`: same prelude over fake client vs manifests, comparing nodes, declared
     and inferred edges, and asserting the offline result is non-empty (no vacuous equality);
   - a **negative fixture** violating every rule, with an exact expected violation list;
   - an **identity fixture**: the same local name under two kinds and two scopes (they must stay
     separate nodes, with scope-local references) and a **duplicate** fixture (the same id twice
     must be an error naming both) — see `pkg/k8sgraph/testdata/{collisions,duplicates}.yaml`;
   - hot-reload test (override file → `Reload` → behavior changes → file removed → default again);
   - tests that need an external binary skip when it is absent; never contact a real service;
   - `internal/archtest` must stay green (no domain/SDK in core).
8. **Benchmark** on a synthetic, scalable workload before declaring done; profile per pass.
9. **Docs**: `doc.go` for shim and cmds, README feature, design-doc phase entries, handoff.

## Anti-patterns
- Ids that are a bare local name (two kinds or scopes collide) · ingest with `graph_add_node`
  (a duplicate overwrites silently) · SDK or domain identifier imported into core · relations written as Go loops ·
  `main.go` containing logic · a shim that grew an algorithm · parity "verified by eye" ·
  only the compliant fixture tested · tests that need a cluster.
