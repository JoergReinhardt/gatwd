# Adapter Roadmap — Growing Domain Coverage Without Bloating Editions

> Status: **DRAFT.** Written 2026-10-15. Names which domains are planned next (generic
> devops substrate, not specific to any one project using gatwd) and records a real
> prerequisite gap found while planning them — see §2, not yet fixed.

---

## 1. Why this doc exists

Mandate 5 (`AGENTS.md` §3) already says third-party domain dependencies belong in
`adapters/`/`pkg/plugins/`, minted into Editions, so that a binary that doesn't use a domain
doesn't pay for it. `docs/development_playbook.md` §5c sharpens *when* to build a new one
generically: a committed future consumer (named, on an actual backlog) is enough, not just
an existing one. This doc applies that test to a concrete set of planned domains, and writes
down a gap in the "doesn't pay for it" promise found while doing so.

---

## 2. Prerequisite: `adapters/` needs per-domain package isolation (found 2026-10-15, not yet fixed)

**Verified, not assumed:** `go list -deps ./cmd/gatwd/...` — the plain core CLI, no domain
Edition wired in at all — pulls in roughly 160 `k8s.io`/`sigs.k8s.io` packages, including real
`k8s.io/client-go` packages (not just shared type definitions). The cause: `adapters/
kubernetes_adapter.go`, `kubernetes_rollout.go` and `kubernetes_write.go` share the single
flat `adapters` package with the generic, domain-free adapters (`fs.go`, `net.go`,
`http_adapter.go`, `scanner.go`, …) that `pkg/cli` imports directly. Go's import graph is
granular at the **package** level, not the symbol level for this purpose — one shared package
for every domain's adapter code means every binary that imports any of it imports all of it,
even though `internal/archtest`'s Clean Core check still passes (`adapters/` is Layer 4,
explicitly where domain-specific I/O is *allowed* to live; the gap is that it isn't split
*enough*, not that it's in the wrong place).

This contradicts, in practice, the "an Edition that doesn't use a domain pays nothing for it"
promise `AGENTS.md`/`STYLEGUIDE.md` both state. It is a real prerequisite to fix — move each
domain's adapter code into its own importable package (its existing shim package, e.g.
`pkg/k8sgraph`, or a new `adapters/<domain>` subpackage) — **before** adding more domain
adapters to the same shared package repeats the mistake for each one. Verify any fix with the
same `go list -deps`/dependency-count check this finding used, for every domain, not just the
one that prompted it.

---

## 3. Planned domains

A deliberately generic list (these are widely-used, non-proprietary systems — naming them is
no different from gatwd's docs already naming Kubernetes, Redis, Kafka or AWS as example
third-party domains, Mandate 5), sequenced by expected reuse per the committed-backlog test
(`development_playbook.md` §5c):

1. **Version-control-hosted designated state** (a GitLab-shaped API: read files/trees out of a
   hosted repository as a `manifests`-equivalent offline source). Generalizes beyond one VCS
   host if a second is ever needed; GitLab is the first concrete target.
2. **A relational-database graph sink/source** (PostgreSQL via `pgx`): durable graph storage,
   and relational-schema introspection as a source in its own right.
3. **A network/infrastructure-inventory system adapter** (NetBox-shaped): read authoritative
   device/IPAM data as graph nodes — and, a genuinely different shape from a read-only source,
   write corrections back through the gated-capability pattern (`development_playbook.md`
   §5c), the same shape `k8s_apply`/`CapClusterApply` already established.

Each is a candidate for the same shim pattern `pkg/k8sgraph` already proved: an isolated I/O
adapter (its own package, per §2) → a thin, zero-logic shim → an embedded `.gatwd` prelude →
a pure-assembly Edition `main.go`.

---

## 4. Documentation stays split per domain, on purpose

Already the existing convention; worth stating as a standing rule instead of leaving it
implicit, so it survives as more domains land. Verified today: `spec/spec.go` — and
everything `cmd/gatwd-gen` generates from it (editor syntax files, the tree-sitter grammar) —
contains **zero** domain-specific entries (`grep -c k8s_ spec/spec.go` is 0); `cli.FunctionForms()`
(what the generator introspects) only ever walks `pkg/cli`'s own static builtin map, never a
domain package's runtime-loaded prelude defs. `docs/dsl_guide.md` documents the core language
only; `docs/k8s_howto.md` is the existing example of a domain getting its own howto doc rather
than being folded into the core guide. As each domain in §3 lands, give it the same treatment
— its own howto doc, its own package `doc.go` — never add its vocabulary to `dsl_guide.md` or
`spec/spec.go`. This is what keeps the core-facing documentation and generated editor tooling
free of keyword creep as the adapter roster grows, and it requires no new mechanism — only
continuing to do, deliberately, what already happened once.
