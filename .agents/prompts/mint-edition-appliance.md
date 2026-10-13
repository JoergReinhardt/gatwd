# Prompt: mint a new Edition / Appliance for a domain

**Task:** `<domain>` tool that `<reads X, derives Y, reports/serves Z>`; Edition (user scripts) or
Appliance (single purpose, embedded rules)?

**Read first:** `.agents/skills/edition-appliance/SKILL.md`; `AGENTS.md` Mandates 5 and 7;
`docs/k8s_graph_design.md` §0, §4, Appendix B; `pkg/k8sgraph/`; `cmd/gatwd-k8s*/`.

**Steps** (each is a commit; gate each)
1. **Design doc** `docs/<domain>_design.md`: use cases, what core already offers, gaps, decisions
   (with options and a recommendation — **ask the user** about irreversible choices), phase plan.
2. **Generic prerequisites in core** (only what a second domain could use). Wire through all layers.
3. **Adapter** under `adapters/`: real I/O only; emit structured payloads (nested `HAMT`/`List`,
   not strings); errors as `Left`; confine third-party SDK imports to named files; offline twin
   (files/fixtures) producing the *same* payload shape.
4. **Thin shim** `pkg/<domain>/`: `<Domain>Step(keyword, client, …)` → `cli.AdapterStep`;
   `ApplyEdition(e, client)` (nil client ⇒ offline-only); `//go:embed lib/<domain>/*.gatwd`.
5. **Prelude** `lib/<domain>/*.gatwd`: `types`, `ids` (qualified `Kind/scope/name`), `ingest`
   (`graph_insert_node`: a duplicate id is an error), `relations`, `queries`, `sources`
   (merge producers with `(a <+> b <+> …)`), one `prelude.gatwd` entry point.
6. **`cmd/gatwd-<domain>`** (Edition) / **`cmd/gatwd-<domain>-<role>`** (Appliance): pure assembly
   `main.go`; embedded `rules/*.gatwd`; ConfigMap override; testable `newAppliance`/pipeline helpers.
7. **Tests**: fixtures; fake client from the same fixtures (`<domain>test` helper package);
   **parity test** live-vs-offline; **negative fixture** violating every rule; an identity fixture
   (same name, two kinds/scopes) and a duplicate-id fixture; hot-reload test;
   benchmark on a synthetic workload.
8. **Docs**: `doc.go`, README feature line, design-doc phase entries, handoff if mid-project.

**Definition of done:** `internal/archtest` green (no domain in core); `main.go` has no domain
logic; parity + negative-fixture tests pass; a benchmark exists; docs updated.

**Don't:** import the SDK in core · author relations in Go · test against a real service.
