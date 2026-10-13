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
  (`cli.NewAppliance(WithEdition, WithEmbeddedFS, WithConfigMapOverride, WithOverrideMode, WithRequiredDefs)`).
  The override is **layered on top of** the embedded rules by default (it may call them and names
  only what it changes); `OverrideReplace` is the explicit alternative. A broken override is a
  `*cli.LoadError`, never a silent fall-back, and a failed reload keeps the last good rules.
  Its command line is `cli.NewApplianceTree` (`check`, `run --interval`, `rules cat`, `--format
  text|json`, `--fail-on-violation`; exit 0 ok / 1 failed / 2 malformed): the domain supplies only
  `Bind` (its source flags and a setup that builds the Appliance and the pipeline).
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
   `(a <+> b <+> …)`. Use the DSL skill (`dsl-authoring`).
6. **Main**: assembly only. No domain loop, relation or type. Helpers for flags/pipeline strings
   are fine and should be unit-tested (`newAppliance`, `checkPipeline`, `livePipeline`).
7. **Tests**
   - fixtures under `testdata/`; a *full* fixture touching every relation;
   - `TestLiveOfflineParity`: same prelude over fake client vs manifests, comparing nodes, declared
     and inferred edges, and asserting the offline result is non-empty (no vacuous equality);
   - a **negative fixture** violating every rule, with an exact expected violation list;
   - hot-reload test (override file → `Reload` → behavior changes → file removed → default again),
     plus a **broken override** (start fails loudly; reload keeps the last good rules) and a layered
     override that reuses an embedded rule;
   - one real-binary row per command line and exit status (0, 1 for violations/errors, 2 for a
     malformed line), a JSON-format row, and "commands that need no cluster build no client"
     (`k8sgraph.LazyDynamicClient`);
   - tests that need an external binary skip when it is absent; never contact a real service;
   - `internal/archtest` must stay green (no domain/SDK in core).
8. **Benchmark** on a synthetic, scalable workload before declaring done; profile per pass.
9. **Docs**: `doc.go` for shim and cmds, README feature, design-doc phase entries, handoff.

## Anti-patterns
- SDK or domain identifier imported into core · relations written as Go loops ·
  `main.go` containing logic · a shim that grew an algorithm · parity "verified by eye" ·
  only the compliant fixture tested · tests that need a cluster.
