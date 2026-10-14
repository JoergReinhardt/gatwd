# Adapter Roadmap — "Batteries Included"

> Status: **DRAFT.** Written 2026-10-15, refactor plan added same day. Names which domains
> are planned next (generic devops substrate, not specific to any one project using gatwd),
> plans the fix for the one place the "Batteries Included" promise currently fails (§2), and
> points at `docs/development_playbook.md` §5 for the precise per-domain package split every
> phase below follows.

---

## 1. Why this doc exists, and what "Batteries Included" means

Mandate 5 (`AGENTS.md` §3) already says third-party domain dependencies belong in their own
adapter package, minted into Editions, so that a binary that doesn't use a domain doesn't pay
for it. **"Batteries Included"** is this project's name for the intended end state of
following that rule across many domains: a wide, ever-expanding shelf of ready-made domain
adapters and shims living in this one repository (Kubernetes today; §3 below for what's
planned next) — so a new gatwd-based tool picks off the shelf instead of writing I/O glue
from scratch, **and no Edition pays, in DSL keyword space or compiled size, for a battery it
didn't pick up.** `docs/development_playbook.md` §5c sharpens *when* to add a new one
generically (a committed future consumer is enough, not just an existing one); §5's table
(same doc) is the precise split this roadmap's phases implement.

The second half of the name — "no Edition pays for what it didn't pick up" — is not a
slogan. §2 records the one place it's currently false, found by checking rather than assuming.

---

## 2. Refactor: give `adapters/kubernetes_*.go` its own package (Phase 0, blocks §3)

### 2.1 The problem, verified

`go list -deps ./cmd/gatwd/...` — the plain core CLI, no domain Edition wired in at all —
pulls in roughly 160 `k8s.io`/`sigs.k8s.io` packages, including real `k8s.io/client-go`
packages, not just shared type definitions. Cause: `adapters/kubernetes_adapter.go`,
`kubernetes_rollout.go` and `kubernetes_write.go` share the single flat `adapters` package
with the generic, domain-free adapters (`fs.go`, `net.go`, `http_adapter.go`, `scanner.go`,
`clock.go`, …) that `pkg/cli` imports directly. Go's import graph is granular at the
**package** level: one shared package for every domain's adapter code means every binary that
imports any of it imports all of it. This does not fail `internal/archtest`'s Clean Core check
(`adapters/` is Layer 4, explicitly where domain-specific I/O is allowed to live — the gap is
that it isn't split *enough*, not that it's in the wrong place), but it does contradict the
"Batteries Included" promise in practice, and it is the template every future domain adapter
would otherwise copy.

### 2.2 The fix

Per `docs/development_playbook.md` §5's table: move the three Kubernetes-specific files (and
their tests) into their own package, `adapters/kubernetes/`, importable only by `pkg/k8sgraph`
— never by `pkg/cli` or any other core package. This is a pure move (file contents don't need
to change beyond the package clause and import paths), which keeps the risk low and the gate
cheap to run.

### 2.3 Phased, gated steps

1. **Pin the regression test first.** Add (or confirm) a test asserting `go list -deps
   ./cmd/gatwd/...` contains **zero** `k8s.io`/`sigs.k8s.io`/`k8s.io/client-go` entries — run
   it now, confirm it **fails** against the current tree (the negative control proving the
   test can fail), before moving anything.
2. **Create `adapters/kubernetes/`** and move `kubernetes_adapter.go`, `kubernetes_rollout.go`,
   `kubernetes_write.go` and their `_test.go` files into it; update the package clause
   (`package kubernetes`) and fix import paths in the moved files.
3. **Update `pkg/k8sgraph`'s imports** from `adapters` to `adapters/kubernetes` wherever it
   references `NativeKubernetesResourceAdapter` or the rollout/write helpers; re-run `pkg/
   k8sgraph`'s own test suite unchanged (a pure import-path move should not change any
   behavior — if it does, something besides the package move happened).
4. **Re-run the Phase 1 regression test** — it must now pass (zero k8s.io entries in the plain
   CLI's dependency list), proving the fix, not just that something moved.
5. **Full gates**: `gofmt -l`, `go vet ./...`, `go test ./...`, `./run_dsl_tests.sh`,
   `go test -race ./pkg/... ./evaluator ./parser` (this touches package structure under
   `pkg/k8sgraph`'s dependency, treat it with the same care as an interpreter-adjacent change
   even though no DSL-visible behavior changes).
6. **One commit**, per the usual phase-delivery discipline (`AGENTS.md` §3 Mandate 8) — this
   is a self-contained, independently valuable fix, not bundled with any new domain's work.

This phase is a **prerequisite**, not a nice-to-have: adding §3's domains into the same
unsplit `adapters` package would repeat the mistake three more times, compounding it instead
of fixing it once.

---

## 3. Planned domains (after §2)

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

Each lands directly in its own `adapters/<domain>/` package from the start (§2's fix is what
makes that the obvious, unremarkable default instead of a special case) — an isolated I/O
adapter → a thin, zero-logic shim in `pkg/<domain>/` → an embedded `.gatwd` prelude → a
pure-assembly Edition `main.go`, exactly the split `docs/development_playbook.md` §5's table
specifies.

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
