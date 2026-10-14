# gatwd-easybell-gitops

A user-space Gatwd Edition for a kustomize-based Kubernetes GitOps repository. It renders
overlays offline, builds the resource graph with the generic k8s prelude and adds the
**Ingress -> Service -> Deployment/StatefulSet** view (hosts, IP, port names).

All logic is DSL in `lib/*.gatwd`; `main.go` is pure assembly. The directory imports core only
through `pkg/cli` and `pkg/k8sgraph`, so it can move to its own repository.

## Decision (2026-10-14)

User-space editions live in `userland/`, one directory per edition, self-contained.

## Use

```sh
go build -o gatwd-easybell-gitops ./userland/easybell-gitops
cd <gitops>/manifests/easybell

# one overlay: graph, tree, DOT
# (`manifests` reads its path from the spawned value, not an inline argument)
./gatwd-easybell-gitops eval -e 'spawn StrVal: ai-library/overlays/production |> manifests |> gitops_graph |> gp_trees |> visualize'
./gatwd-easybell-gitops eval -e 'spawn StrVal: ai-library/overlays/production |> manifests |> gitops_graph |> visualize --dot'

# every production overlay of one project (-D OVERLAY picks the overlay name; -D PATTERN is a
# glob over the FULL discovered root path — "**" matches everything, see gp_roots below)
./gatwd-easybell-gitops -D OVERLAY=production -D PATTERN=** eval -e 'spawn StrVal: x |> walk_fs ai-library |> gp_roots |> manifests |> gitops_graph |> gp_trees |> visualize'

# a wide scan over EVERY project's production overlay: some roots may be broken
# (a dangling component symlink, a bad patch, ...); skip_errors reports each on
# stderr and drops it instead of losing every project's resources over one bad root
./gatwd-easybell-gitops -D OVERLAY=production -D PATTERN=** eval -e 'spawn StrVal: x |> walk_fs . |> gp_roots |> manifests |> skip_errors |> gitops_graph |> gp_trees |> visualize'
```

- `gitops_graph` = `k8s_graph` plus `Exposes` (Service -> Deployment or StatefulSet: the Service
  selector is a subset of the pod template labels, same namespace — both workload kinds carry
  the same PodTemplateSpec shape, and a Service's selector cannot tell them apart) plus
  `DependsOn` (see below).
- `DependsOn`/`Hairpin`: Deployment/StatefulSet -> Service or Ingress, derived from a URL-shaped
  value in a ConfigMap the workload mounts (core's generic `Configures` edge already says which
  workload that is). This cluster's internal service DNS is **not** Kubernetes' default
  `<svc>.<ns>.svc.cluster.local` — CoreDNS here answers the bare `<svc>.<ns>.svc` (see
  `sipwise-broker/base/configmap.yaml`'s `CTA_URL`/`ORDER_GATEWAY_URL`/... in the real repo), so a
  value shaped like `http://app.cta-backend-staging.svc` resolves straight to that Service via a
  **`DependsOn`** edge. A value shaped like `https://<host>.dev-k8s.easybell.de/...` or
  `...k8s.easybell.de/...` (an external-dns hostname this cluster answers for) resolves instead to
  the Ingress declaring that exact host, via a **`Hairpin`** edge — the external load-balancer hop
  in between is not a Kubernetes resource and is not represented, so the edge goes straight from
  the workload to that Ingress, which is already linked onward by `RoutesTo`/`Exposes`. The two are
  distinct colors (not both `DependsOn`) precisely so a query can tell them apart on its own —
  `graph_edges_of Hairpin g` finds every round-trip dependency without re-deriving anything. A
  ConfigMap is never part of the dependency path itself — only `Configures` touches it, so it
  shows up in the graph attached to the side of the workload it configures, not inline between it
  and what it depends on. The same ConfigMap or Ingress can anchor edges from many different
  consumers; that is ordinary graph fan-in, not something flattened away.
- `gp_trees` = one tree per Ingress; labels carry `hosts=`, `ip=` (external-dns target),
  `via=` (backend port) and the port names. (`gp_trees` does not currently show `DependsOn`/
  `Hairpin`/`Configures` — query them directly with `graph_edges_of <Color>`, or render the raw
  graph with the `graph`/`wide_graph` tasks below, or add a dedicated view on top.)
- `gp_roots` lists the kustomization roots of the chosen overlay (`-D OVERLAY=...`) whose full
  path ALSO matches the constant `PATTERN` (`-D PATTERN=...`, required — `**` matches every
  project). See "Scoping a render down" below.
- `skip_errors` (a generic core step, `evaluator.StepSkipErrors`) reports each Left on stderr
  and drops it; without it, one root that fails to render loses every other root's resources
  too, because `gitops_graph`'s ingest fold withholds its whole aggregate over a failed stream
  (Mandate 6). Only insert it when scanning many roots at once that may legitimately fail; a
  single overlay (the first two examples above) should stay strict so a render failure is seen
  immediately.

## `tasks.gatwd`: named views instead of typing the pipeline out

`tasks.gatwd` wraps the three pipelines above as named tasks, so the command line only needs a
path:

```sh
./gatwd-easybell-gitops task main  -s tasks.gatwd -- ai-library/overlays/production           # tree
./gatwd-easybell-gitops task dot   -s tasks.gatwd -- ai-library/overlays/production            # DOT
./gatwd-easybell-gitops -D OVERLAY=production -D PATTERN=** task wide  -s tasks.gatwd -- /path/to/repo   # whole repo, tree
./gatwd-easybell-gitops task graph -s tasks.gatwd -- ai-library/overlays/production            # raw graph DOT
./gatwd-easybell-gitops -D OVERLAY=production -D PATTERN=** task wide_graph -s tasks.gatwd -- /path/to/repo  # raw graph, whole repo
```

`tasks.gatwd` names **no source** — every task reads its path from the `--` argument, so the same
script applies to any overlay or repo; swap sources by rerunning with a different argument, never
by editing the script (design pattern: `AGENTS.md` §5a, `.agents/skills/dsl-authoring` §6).
`task` (the subcommand, not the bare `[script]` positional form) also skips the interpreter's
startup banner, so its output can be piped straight into `dot` without stripping a line first.
`main` and `dot` are pipeline-bodied defs (fragments): run as a task, the arguments after `--`
arrive as their input stream, one String node each, which is what `manifests` consumes. `wide`/
`wide_graph` use bare `walk_fs` (no inline path): it reads the repo root from that same stream
instead of a literal, the same way `manifests` reads its path — that is what makes them reusable
across repos with nothing to edit here. `graph`/`wide_graph` render the raw graph (every node,
every relation) instead of the curated Ingress-rooted tree `main`/`dot`/`wide` show.

## Scoping a render down

The whole-fleet `wide_graph` is wide enough that it needs a browser (not a raster viewer like
`eog`) and too wide to debug anything in. Two ways to narrow it:

1. **`-D PATTERN=<glob>`** narrows which projects `gp_roots` includes, by matching the FULL
   discovered root path. `-D PATTERN=**` matches everything (required — there is no silent
   default, same as `-D OVERLAY`). A literal path with no wildcard matches itself and everything
   under it; `*`/`?`/`[...]` match within one `/`-delimited segment; `**` (its own segment)
   matches zero or more whole segments, so it stands for "any number of directories in between":

   ```sh
   # only the project(s) matching a name pattern, from the whole repo
   -D PATTERN=/home/j/.../manifests/easybell/sip*/** task wide_graph -s tasks.gatwd -- /home/j/.../manifests/easybell
   ```

   `glob_match`/`glob_root` (the primitives behind this) are generic core builtins
   (`pkg/cli/dslfuncs_glob.go`), not specific to this edition — usable in any script for any
   path-matching need.

2. **Filtering by relation kind is a plain graph query — it never needs a prelude change.**
   "Show me every root with a hairpin dependency" doesn't need a new relation or a new view; it's
   `graph_edges_of Hairpin g` (or `graph_filter_edges`/`graph_filter_nodes` for a filtered subgraph
   to visualize) over the graph that's already built. A short ad hoc query, with no edit to
   `tasks.gatwd`:

   ```sh
   ./gatwd-easybell-gitops -D OVERLAY=staging -D PATTERN=** eval -e \
     'spawn StrVal: /path/to/repo |> walk_fs |> gp_roots |> manifests |> skip_errors |> gitops_graph |> graph_edges_of Hairpin |> collect_list'
   ```

   `task` has no `-e`/`--eval` flag of its own (only the root `[script]`/`run` form does); `-D`
   already gives named-constant filtering the same short, no-prelude-edit shape `-e` would, so
   `gp_roots`'s `PATTERN` uses it rather than adding a new flag for one more case.

## Tests

`go test ./userland/...` runs the edition over `testdata/stack.yaml`: a complete chain, a
Service no Deployment backs (`orphan`, must get no `Exposes`), a missing path that must surface
as an error, a ConfigMap-derived `DependsOn`/`Hairpin` pair plus a real-but-unrelated URL that
gets neither, and (over `testdata/roots/`, a tiny two-project fixture) `gp_roots`'s `PATTERN`
narrowing a scan down to one project, or to none.
