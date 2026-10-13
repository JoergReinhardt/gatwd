# How-To: the Kubernetes Edition and Operator

This guide covers the two binaries built on the Kubernetes graph:

| Binary | What it is | Use it to |
| :-- | :-- | :-- |
| `gatwd-k8s` | an **Edition**: the full interpreter plus the Kubernetes steps and graph prelude | run your own scripts and ad-hoc queries |
| `gatwd-k8s-operator` | an **Appliance**: the edition plus embedded compliance rules, hot-reloadable | check a cluster (or manifests) against rules, once or on a schedule |

Both read a **live cluster** (kubeconfig or in-cluster service account) or **offline manifests**
(files, directories, `kustomize build` output) and lift the resources into one typed graph
(`k8s_graph`). The cluster client is built only when a live crawl step runs, so `version`, `completion`, `check` and offline work never look for a kubeconfig; a crawl with no reachable cluster is an `[ERROR] no cluster access (…)` result that names the cause.

Design background: [`k8s_graph_design.md`](k8s_graph_design.md). Language reference:
[`dsl_guide.md`](dsl_guide.md). Everything below was run against the files in
[`examples/k8s_howto/`](../examples/k8s_howto/) and is covered by tests
(`cmd/gatwd-k8s-operator/howto_test.go`). Commands assume the repository root.

```bash
go build -o /tmp/gatwd-k8s ./cmd/gatwd-k8s
go build -o /tmp/gatwd-k8s-operator ./cmd/gatwd-k8s-operator

/tmp/gatwd-k8s-operator check --manifests examples/k8s_howto/cluster   # offline compliance check
echo 'spawn StrVal: examples/k8s_howto/cluster |> manifests |> k8s_graph |> graph_order' | /tmp/gatwd-k8s
```

---

## 1. Defining tasks with gatwd scripts

A **task** is a named `def` in a script. Two kinds, for two jobs:

```haskell
-- examples/k8s_howto/tasks.gatwd
def graph = manifests |> k8s_graph                                    -- a stream fragment (its input is the argument)
def trace_to_pod = graph |> graph_shortest_path nginx-svc app-pod |> collect_list
def count_kind = \kind g -> length (graph_payloads_of kind g)         -- a value function (takes arguments)
def main = (who_configures_what <+> trace_to_pod <+> pod_count)       -- the default task
```

- A **pipeline-bodied** `def` is a reusable fragment you drop into a pipeline (`... |> trace_to_pod`) or run
  as a task: the command-line arguments arrive as its input stream.
- A **lambda-bodied** `def` is a curried function value: `count_kind "Pod" g`, or partially applied
  (`count_kind "Pod"`). As a task it is called with the `ARGS` list.
- `main` is the task that runs when none is selected; keep invocations out of the top level.

### Passing arguments and picking one task

The script path is the first argument; everything after `--` is handed to the script as `ARGS`, and
`--task NAME` runs one `def` instead of the script's `main`. `tasks.gatwd` takes the manifest path from its
arguments, so the same script serves any directory:

```bash
gatwd-k8s examples/k8s_howto/tasks.gatwd -- examples/k8s_howto/cluster                       # runs `main`: all three tasks
gatwd-k8s examples/k8s_howto/tasks.gatwd --task trace_to_pod -- examples/k8s_howto/cluster   # ONE task
gatwd-k8s check examples/k8s_howto/tasks.gatwd                                                # validate without running (`--check` also works)
gatwd-k8s task trace_to_pod -s examples/k8s_howto/tasks.gatwd -- examples/k8s_howto/cluster   # the same single task, as a subcommand
gatwd-k8s -i examples/k8s_howto/tasks.gatwd -- examples/k8s_howto/cluster                    # run, then a prompt with the defs loaded
```

A pipeline-bodied task is a **fragment**: the arguments arrive as its input stream, one `String` node each,
so `manifests` receives the path. A lambda-bodied task is called with the `ARGS` list. With no arguments a
fragment gets an empty stream and prints `[]`. `ARGS` and `main` are specified in
[`dsl_guide.md`](dsl_guide.md) §23. A task name that does not exist exits non-zero and lists the defined ones.

For the interactive route, load the definitions and type an invocation at the prompt (this also works piped,
handy in shell scripts and CI):

```bash
printf 'spawn StrVal: examples/k8s_howto/cluster |> trace_to_pod\nexit\n' | gatwd-k8s -i examples/k8s_howto/tasks.gatwd
```

`:load other.gatwd` at the prompt loads another script. To compose several script files, make a small
wrapper script that `import`s them (imports resolve relative to the importing file).

Parameters: pass them after `--` and read them from `ARGS`; for a fixed default, a `const` still works
(`const CLUSTER = path/to/manifests`).

---

## 2. Extending it

### 2.1 Checking that DSL changes work without recompiling

Everything written in `.gatwd` is read at run time, so the loop is: edit, re-run, look.

```bash
gatwd-k8s check my.gatwd            # structural errors (exit 1) and advisory warnings
gatwd-k8s prelude cat lib/k8s/prelude.gatwd   # read the embedded relations; copy, edit, load with --prelude
gatwd-k8s --no-prelude my.gatwd     # run without the embedded prelude (the steps stay registered)
gatwd-k8s my.gatwd                  # run it
gatwd-k8s-operator check --manifests examples/k8s_howto/cluster --rules my_rules.gatwd   # try operator rules
gatwd-k8s-operator run --manifests … --rules my_rules.gatwd --interval 5s               # and edit while it runs
```

With `run --interval`, the operator re-reads the rules file before each check, so saving an edit changes the
very next report — no restart. Only Go changes (adapters, the embedded prelude, the operator itself) need
a rebuild.

**Debugging.** Evaluate small pieces in the prompt first. Two traps:

- A lambda that fails inside `map`, `filter`, `flatmap` or `fold` — inline or as a `def` — is reported as
  `[ERROR] …` naming the cause (`fold: undefined identifier: …`, `map: element 2: …`). If a rule
  unexpectedly yields `[]`, the source list or a guard is empty, not a swallowed error.
- Inside a plain list literal, parenthesise calls: `[(f x), (g y)]`, not `[f x, g y]`.

More traps and their causes: [`development_playbook.md`](development_playbook.md) §6.

### 2.2 Changing the operator's rules

The operator embeds `cmd/gatwd-k8s-operator/rules/compliance.gatwd`. A file given with `--rules` (default
`/etc/gatwd/rules.gatwd`, i.e. a mounted ConfigMap) is **layered on top of** the embedded rules (decision D5):
the embedded file loads first, then the override in the same interpreter, so the override names only what it
changes and can call every embedded `rule_*` closure. Redefining a name overrides it; a new name extends.

> To add a rule: define `rule_mine` and redefine `compliance` (the embedded one lists the six embedded rules
> explicitly) to include it. The operator **reports** what the override did when it loads and on every reload:
> `rules: overridden: compliance` / `rules: added: rule_mine`. A def that is an identical copy of the embedded
> one is not reported. If the file is missing or empty only the embedded rules run.

`examples/k8s_howto/rules.gatwd` is exactly that: the infra relations (inlined — a ConfigMap key is one file),
one extra rule, and a `compliance` that runs the embedded rules plus the new one over `infra_graph`:

```bash
gatwd-k8s-operator check --manifests examples/k8s_howto/cluster --rules examples/k8s_howto/rules.gatwd
# [[IdleDeployment worker] [UnusedService backend-svc] [MissingConfigMap worker] [DatabaseWithoutCredentials audit]]
#   (stderr)  gatwd-k8s-operator: rules: override …/rules.gatwd layered on the embedded rules (mode layer)
#   (stderr)  gatwd-k8s-operator: rules: overridden: compliance
#   (stderr)  gatwd-k8s-operator: rules: added: infra_graph, rel_hosted_on, rel_uses_database, rule_database_without_credentials, secret_target
```

Imports inside an override resolve against **the override file's directory**, not the working directory, so
`import "infra_relations.gatwd"` finds a sibling key of the same mounted ConfigMap.

**A broken override fails loudly.** An unreadable file, a structural error (`unbalanced delimiter`), a statement
that prints `[ERROR]`, an import that cannot be read, or a `compliance` left undefined is a load error naming the
file and the problems: at start-up the operator exits 1; on a hot reload it prints the error on stderr and keeps
checking with the **last good rules** — it never silently falls back to the embedded ones. (A rule that is
syntactically fine but fails when it runs is reported by the check itself as `[ERROR] …`.)

**Replace mode.** Where the platform team wants to own every rule, the override can replace the embedded
rules instead (`--rules-mode replace`, see §3.2). The file then must be a complete rule set — a copy of the embedded
file plus changes — and it cannot call `rule_*` names it does not define.

### 2.3 The graph model and its extension points

Every resource becomes a node **coloured by its `kind`**, with id `metadata.name` and the whole resource
(a nested map) as payload. Edges are `["Color", weight, "from", "to"]`; a relation is a pass that adds
them. So a new node kind needs **no model change at all** — anything with `kind` and `metadata.name`
is a node — and a new relation is a DSL pass. Declared vocabulary (`type NodeColor = …`) is for readers
and exhaustiveness checks.

Two caveats that matter when you add kinds:

- **Ids are `metadata.name` only.** A Pod and a Service with the same name collide, and so would a
  Database named like a Service. Give new kinds a naming convention that cannot collide.
- Edges only join nodes that exist (`graph_has_node` guards), so a reference to something missing yields
  no edge. Find dangling references with a rule (see `rule_dangling_ingress`).

### 2.4 Example: Secret → database → DBMS, with a SQL (or Redis, …) adapter

Goal: a Kubernetes Secret holds credentials for a database, the database lives in a DBMS, and you want to
ask "which DBMS does this workload ultimately depend on?" across both worlds.

**Step 1 — put the data in the graph.** Your adapter yields one value per database and per DBMS, in the
same shape every resource has: nested maps with `kind`, `metadata.name` and `spec`.

```yaml
kind: DBMS                       kind: Database
metadata: {name: pg-prod}        metadata: {name: orders}
spec: {engine: postgres, …}      spec: {dbms: pg-prod}
```

You can try the whole model first with **no Go**: `examples/k8s_howto/cluster/databases.yaml` is those
records as a manifest, so `manifests` ingests them.

**Step 2 — add the relations in DSL** (`examples/k8s_howto/infra_relations.gatwd`):

```haskell
def secret_target = \s -> hamt_get "gatwd.io/database" (hamt_path "metadata.annotations" s)
def rel_uses_database = fold (\_ g -> graph_add_edges [ ["UsesDatabase", 1, (hamt_path "metadata.name" s), (secret_target s)] | s <- (graph_payloads_of "Secret" g), … graph_has_node (secret_target s) g ] g) graph_empty
def rel_hosted_on     = fold (\_ g -> graph_add_edges [ ["HostedOn", 1, (hamt_path "metadata.name" d), (hamt_path "spec.dbms" d)] | d <- (graph_payloads_of "Database" g), … ] g) graph_empty
def infra_graph = k8s_graph |> rel_uses_database |> rel_hosted_on
```

```bash
gatwd-k8s examples/k8s_howto/infra.gatwd
# [orders-db-credentials-UsesDatabase->orders]
# [audit-HostedOn->pg-prod orders-HostedOn->pg-prod]
# [orders-db-credentials orders pg-prod]            ← the path Secret → Database → DBMS
```

**Why an annotation instead of reading the credentials?** The DSL has string functions (`str_*`) but no base64 yet
(the encodings tier of the hardening plan adds it), and a Secret's `data` is base64. An annotation (`gatwd.io/database: orders`) is plaintext, visible from a live
cluster, and keeps credentials out of the graph logic. Matching by *exact equality* of a value to a node
name also works (that is how the inferred `Calls` relation works). If you genuinely need decoded values,
add a **generic** `base64_decode` builtin to core first (the Go-standard-library rule: wire `encoding/base64`
through every layer with a non-domain DSL test), never a Secret-specific hack in the shim.

**Step 3 — plug in the real adapter (Go).** You said the SQL/Redis adapters already exist and are
integrated. All that is needed is to expose each as a step, merge the streams, and ship the relations as an
embedded prelude. A new edition `main.go` is pure assembly:

```go
//go:embed lib/infra/*.gatwd          // relations.gatwd (Step 2) + prelude.gatwd below
var infraFS embed.FS

func main() {
    k8sClient, _ := k8sgraph.DynamicClientFromKubeconfig() // nil ⇒ offline-only edition
    k8sgraph.ApplyEdition(cli.NewEdition("gatwd-infra"), k8sClient).
        WithStep(cli.AdapterStep("sql_databases", sqlgraph.Databases(db))). // your adapter: evaluator.Adapter[any]
        WithStep(cli.AdapterStep("redis_instances", redisgraph.Instances(rdb))).
        WithPreludeFS(infraFS, "lib/infra/prelude.gatwd").
        Run()
}
```

```haskell
-- lib/infra/prelude.gatwd
import "relations.gatwd"
def infra_resources = (k8s_resources <+> sql_databases <+> redis_instances)   -- live: merge the streams
def infra_offline   = (manifests <+> sql_databases)                           -- offline k8s + live DB
```

Then `spawn StrVal: default |> infra_resources |> infra_graph |> graph_shortest_path orders-db-credentials pg-prod`.
This assembly (extra adapter step + extra embedded prelude on top of `ApplyEdition`) was built and run for
this guide. Rules for the adapter: do real I/O only, yield **nested maps/lists, not strings**, turn errors
into error nodes, keep third-party drivers inside the adapter file, and give it an offline twin (files) with
the same shape. Full recipe: [`.agents/skills/edition-appliance/SKILL.md`](../.agents/skills/edition-appliance/SKILL.md).

**Step 4 — tests that keep it honest.** A fixture with a Secret, a Database and a DBMS and the expected
edges; a **negative** fixture (a Database with no Secret) that must trip your rule; and, if a source has a
live and an offline form, a parity test that runs the same prelude over both (see
`pkg/k8sgraph/live_parity_test.go`).

**Security.** The standard crawl includes Secrets, whose payload (including `data`) is held in memory and
in the graph. If you do not need them, drop `k8s_secrets` from the source and from the RBAC role:

```haskell
def k8s_resources = (k8s_pods <+> k8s_services <+> k8s_configmaps <+> k8s_endpoints <+> k8s_deployments <+> k8s_replicasets <+> k8s_ingresses <+> k8s_pvcs <+> k8s_serviceaccounts <+> k8s_roles <+> k8s_rolebindings)
```

(put it in your override/prelude; verified against a fake cluster). The sample `deploy.yaml` grants **no**
access to Secrets by default.

---

## 3. Building, installing and deploying

### 3.1 Build

```bash
go build -o bin/gatwd-k8s ./cmd/gatwd-k8s
go build -o bin/gatwd-k8s-operator ./cmd/gatwd-k8s-operator

# static, reproducible, smaller — what you want for a container (no cgo needed)
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/gatwd-k8s-operator ./cmd/gatwd-k8s-operator
```

The static operator is a single ELF binary of roughly 29 MB. To install into `$GOBIN` (default
`~/go/bin`) from a checkout: `go install ./cmd/gatwd-k8s ./cmd/gatwd-k8s-operator`.

### 3.2 Run it

| | `gatwd-k8s` | `gatwd-k8s-operator` |
| :-- | :-- | :-- |
| Input | one script file, or stdin at a prompt | flags below |
| Cluster access | in-cluster service account, else `$KUBECONFIG`, else `~/.kube/config`; built lazily on the first crawl — with none reachable a crawl step yields `[ERROR] no cluster access (…)` (offline steps are unaffected) | the same; with `--manifests` it never needs one |
| Flags | `[script] [-- args]`, `-i`, `--task NAME`, `-e STMT`, `-D NAME=VALUE`, `--prelude FILE`, `--no-prelude`, `--check file`; subcommands `run eval check task repl prelude builtins version completion` (see the DSL guide §23) | subcommands `check` (once), `run --interval DUR` (repeat), `rules cat` (print the embedded rules), `version`, `completion`; flags `--manifests PATH` (omit for live), `--namespace NS` (default `default`; `""` = all namespaces), `--rules FILE` (default `/etc/gatwd/rules.gatwd`), `--rules-mode layer\|replace` (default `layer`, §2.2), `--interval DUR`, `--format text\|json`, `--fail-on-violation` (see below). Without a subcommand the operator behaves as before: `--interval` makes it loop, otherwise it checks once. The single-dash spellings of earlier releases (`-manifests`, `-rules`, `-interval`, …) still work for one release and print a deprecation notice on stderr |
| Output | each result on stdout | one result per check: `[]` when compliant, otherwise `[[Rule resource] …]` |

**Exit codes.** `0` the check ran; `1` the command failed — an override that does not load, a rule that errors
(`[ERROR] …`), a pipeline that yields nothing, or, with `--fail-on-violation`, a non-empty result; `2` a
malformed command line (unknown flag or `--format`, no `--manifests` and no reachable cluster). Without
`--fail-on-violation` the operator reports violations but exits 0, so a scheduled `run` is not restarted for
finding something. For a CI gate:

```bash
gatwd-k8s-operator check --manifests ./rendered --fail-on-violation || { echo "policy violations"; exit 1; }
```

`run --interval` is a daemon: a failing check is printed and the loop continues; only a start-up failure ends it
(and `--fail-on-violation` is rejected there).

**Output formats.** `--format text` (default) prints the result as before: `[]` when compliant, otherwise
`[[Rule resource] …]`; an error prints as `[ERROR] …`. `--format json` prints **one JSON object per check, on one
line**, for log pipelines (every key always present):

```bash
gatwd-k8s-operator check --manifests examples/k8s_howto/cluster --format json
# {"status":"violations","count":3,"violations":[{"rule":"IdleDeployment","resource":"worker"},…],"errors":[]}
```

`status` is `ok`, `violations` or `error`; a resource that is a list (a dependency cycle) stays a JSON array.
Diagnostics — the rules report, reload failures — always go to **stderr**, so stdout stays machine-readable.

The operator reads plain manifests (`manifests` step). For kustomize output, render first
(`kubectl kustomize overlay > rendered/all.yaml`, then `--manifests rendered`). In `gatwd-k8s` scripts the
separate `kustomize` step renders a kustomization root itself — `spawn StrVal: ./overlay |> kustomize |> k8s_graph` —
but it shells out to the `kustomize` binary, which must be installed (it is not in this guide's tested set).

### 3.3 Container image

A static binary needs nothing else in the image. This minimal image was built and run for this guide
(`distroless/static:nonroot` is the usual production base and works the same way):

```dockerfile
FROM scratch
COPY gatwd-k8s-operator /gatwd-k8s-operator
USER 65532:65532
ENTRYPOINT ["/gatwd-k8s-operator"]
```

```bash
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o gatwd-k8s-operator ./cmd/gatwd-k8s-operator
docker build -t registry.example.com/gatwd-k8s-operator:0.1.0 .
docker run --rm -v "$PWD/examples/k8s_howto/cluster:/data:ro" registry.example.com/gatwd-k8s-operator:0.1.0 check --manifests /data
```

The image needs nothing else for live checks or plain manifests; the `kustomize` step would also need that binary in the image.

### 3.4 Deploy to a cluster

[`examples/k8s_howto/deploy.yaml`](../examples/k8s_howto/deploy.yaml) is a complete, minimal deployment: a
namespace, a service account, a **read-only** `ClusterRole` (`get`/`list` only — the operator lists, it does
not watch or write), the binding, and the Deployment (non-root, read-only root filesystem, no capabilities).

```bash
kubectl apply -f examples/k8s_howto/deploy.yaml
kubectl -n gatwd create configmap gatwd-rules --from-file=rules.gatwd=examples/k8s_howto/rules.gatwd   # optional override
kubectl -n gatwd logs deploy/gatwd-k8s-operator -f
```

- **Where the results go:** stdout, so `kubectl logs` — one line per check (`run --interval 5m` in the sample).
  Ship them with your normal log pipeline and alert on lines that are not `[]`.
- **Changing rules live:** `kubectl create configmap … --dry-run=client -o yaml | kubectl apply -f -`. The
  kubelet refreshes the mounted file (typically within about a minute) and the operator re-reads it before
  the next check. **Do not mount with `subPath`** — Kubernetes does not update `subPath` mounts.
- The ConfigMap volume is `optional: true`, so the operator starts with its embedded rules if the ConfigMap
  does not exist yet.
- **One replica is enough**; there is no leader election because it only reads and prints.
- **Scope:** the sample uses `--namespace ""` (all namespaces) with a `ClusterRole`. For one namespace use
  `--namespace NAME` and a `Role` + `RoleBinding` instead.
- **Add a kind** (a CRD): add a crawl step in `pkg/k8sgraph` (`CoreCrawls`), list it in
  `lib/k8s/sources.gatwd`, and add `get`/`list` for it to the `ClusterRole`.

### 3.5 Troubleshooting

| Symptom | Likely cause |
| :-- | :-- |
| `[ERROR] no cluster access (…)` from a `k8s_*` crawl step | kubeconfig auth plugin not supported or no cluster; use `--manifests` / the `manifests` step or fix the kubeconfig |
| `no --manifests given and no cluster reachable` (exit 2) | the operator has no source |
| Operator prints `[ERROR] …` after you added a rule | a name in the rule does not resolve, or an operator rejected its operands; the message names it (§2.1) |
| Rule output is empty but should not be | a node kind is missing from the source list or RBAC, or ids collide (§2.3) |
| Edge you expect is missing | an endpoint node does not exist (`graph_has_node` guard), or the relation guards a field the resource lacks |
| Duplicate edges in a listing | harmless: a relation can be derived twice (e.g. `Configures` via a volume *and* an env reference) |

---

## 4. Known limitations (as of this writing)

- The DSL has no base64 builtin yet (§2.4); the string library `str_*` exists.
- Node ids are `metadata.name` only (§2.3).
- `||>` runs the *following per-element step* across the stream on a worker pool (order is not
  deterministic); a whole-stream step such as `fold` or `sort` after `||>` would see single nodes.
