# How-To: the Kubernetes Edition and Operator

This guide covers the two binaries built on the Kubernetes graph:

| Binary | What it is | Use it to |
| :-- | :-- | :-- |
| `gatwd-k8s` | an **Edition**: the full interpreter plus the Kubernetes steps and graph prelude | run your own scripts and ad-hoc queries |
| `gatwd-k8s-operator` | an **Appliance**: the edition plus embedded compliance rules, hot-reloadable | check a cluster (or manifests) against rules, once or on a schedule |

Both read a **live cluster** (kubeconfig or in-cluster service account) or **offline manifests**
(files, directories, kustomize overlays rendered in process) and lift the resources into one typed graph
(`k8s_graph`). The cluster client is built only when a live crawl step runs, so `version`, `completion`, `check` and offline work never look for a kubeconfig; a crawl with no reachable cluster is an `[ERROR] no cluster access (…)` result that names the cause.

Design background: [`k8s_graph_design.md`](k8s_graph_design.md). Language reference:
[`dsl_guide.md`](dsl_guide.md). Everything below was run against the files in
[`examples/k8s_howto/`](../examples/k8s_howto/) and is covered by tests
(`cmd/gatwd-k8s-operator/howto_test.go` runs the examples; `doc_test.go` runs the real binaries and fails when a command or
output shown here no longer matches what they print). Commands assume the repository root.

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
def trace_to_pod = graph |> graph_shortest_path Service/default/nginx-svc Pod/default/app-pod |> collect_list
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
gatwd-k8s prelude list              # the embedded library: the entry file and the files beside it
gatwd-k8s prelude cat lib/k8s/relations.gatwd   # read one; copy, edit, load a copy with --prelude (read after the library, so it redefines what it names)
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
- A comprehension needs a generator: `[ e | guard ]` alone yields `[]` even when the guard holds. Bind the value first
  (`[ e | q <- [r], guard ]`, as `k8s_scope` does).

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
# [[IdleDeployment Deployment/default/worker] [UnusedService Service/default/backend-svc] [MissingConfigMap Deployment/default/worker] [DatabaseWithoutCredentials Database/default/audit]]
#   (stderr)  gatwd-k8s-operator: rules: override …/rules.gatwd layered on the embedded rules (mode layer)
#   (stderr)  gatwd-k8s-operator: rules: overridden: compliance
#   (stderr)  gatwd-k8s-operator: rules: added: infra_graph, rel_hosted_on, rel_uses_database, rule_database_without_credentials, secret_target
```

Imports inside an override resolve against **the override file's directory**, not the working directory, so
`import "infra_relations.gatwd"` finds a sibling key of the same mounted ConfigMap.

**An override written for the old ids fails loudly too.** One that still calls `k8s_names` (removed: `k8s_ref` replaced it)
or builds ids from `metadata.name` alone is not silently wrong: the former is `[ERROR] fold: undefined identifier: k8s_names`
(exit 1), the latter finds no node (§2.3, `k8s_ref`). Migrate it to `k8s_ref kind ns name g` and `k8s_id_of`.

**A broken override fails loudly.** An unreadable file, a structural error (`unbalanced delimiter`), a statement
that prints `[ERROR]`, an import that cannot be read, or a `compliance` left undefined is a load error naming the
file and the problems: at start-up the operator exits 1; on a hot reload it prints the error on stderr and keeps
checking with the **last good rules** — it never silently falls back to the embedded ones. (A rule that is
syntactically fine but fails when it runs is reported by the check itself as `[ERROR] …`.)

**Replace mode.** Where the platform team wants to own every rule, the override can replace the embedded
rules instead (`--rules-mode replace`, see §3.2). The file then must be a complete rule set — a copy of the embedded
file plus changes — and it cannot call `rule_*` names it does not define.

**What an override may do (security).** An override is a script that did **not** come from the command line: it is
a file somebody mounted, often from a ConfigMap that other people can edit. The language can do more than compute — a
word that is not a step is run as an operating-system program, there are file and network steps, and `import` reads
a file — so the operator starts the rules, the override above all, under a **deny-all policy** (decision D8):

| Capability | What it is | Gated steps | Flag that grants it |
| :-- | :-- | :-- | :-- |
| `exec` | running OS commands | an unknown step word (`printf …`, `sh -c …`), `memoize CMD`, `zip CMD` | `--allow-exec` |
| `fs` | touching files | `scan_file`, `scan_words`, `walk_fs`, `write_file`, `append_file`, `import "file"` | `--allow-fs` |
| `net` | opening sockets | `dial_tcp`, `listen_tcp` | `--allow-net` |
| `env` | reading the environment | (no builtin reads it yet; the switch for the one that will) | `--allow-env` |

A denied step does not run; it is an error value, never an empty result — in text `[ERROR] step 'printf' requires the
exec capability (policy denies it)`, in `--format json` an entry in `errors` and a structured one in `diagnostics`
(`"kind":"capability_denied","step":"printf","capability":"exec"`). A top-level statement that needs a capability
fails the **load** (exit 1, the file and the step are named); a rule that needs one fails the **check**.

- **The override cannot widen its own policy.** The policy is set by the process's flags, in Go; no statement of the
  language names it, so `const AllowExec = true` or a `def` that looks like a switch changes nothing (a test drives the
  real binary through those attempts). A ConfigMap edit can never grant what the Deployment's `args:` do not.
- **What stays allowed:** everything the operator wires itself — `manifests`, the cluster crawl (`k8s_resources`,
  `k8s_*`) and the embedded rules — works with no flag; a pure override (graph rules, comprehensions) needs none.
- **Imports:** an override's `import "sibling.gatwd"` reads a file, so it needs `--allow-fs`. Without it the start fails
  with `import "…" requires the fs capability`; either inline the sibling file or grant the capability. (Imports of the
  embedded rules are not file access and are always fine.)
- **Which scripts run OS commands?** Only a script that is trusted: the one you name on the command line of `gatwd` or
  `gatwd-k8s` (and `-e`, preludes, the REPL), or one you grant `--allow-exec` for. Scripts from anywhere else —
  an override, a `RunOptions.Source` — do not.
- **Time and size.** A granted capability should come with a bound: `--timeout DUR` kills a stuck step and the OS command
  it started, and `--max-output BYTES` fails a check that prints too much (§3.2). Neither is on by default.

### 2.3 The graph model and its extension points

Every resource becomes a node **coloured by its `kind`**, with the whole resource (a nested map) as
payload and a **qualified id `Kind/namespace/name`** (`Pod/default/web`, `Service/prod/web`). A resource
without a namespace — a cluster-scoped kind, or a manifest that names none — gets the scope `_`
(`Role/_/admin`). Edges are `["Color", weight, "from", "to"]` over those ids; a relation is a pass that
adds them. So a new node kind needs **no model change at all** — anything with `kind` and `metadata.name`
is a node — and a new relation is a DSL pass. Declared vocabulary (`type NodeColor = …`) is for readers
and exhaustiveness checks.

What the qualified id buys, and what it asks of you:

- **Equal names no longer collide.** A Service and a Pod called `web`, or `web` in `default` and in
  `prod`, are different nodes (`pkg/k8sgraph/testdata/collisions.yaml`). A resource defined **twice**
  (same kind, namespace and name — two files of one kustomization, a pasted document) is an error that
  names both declarations — **and the file each came from** — never one silently replacing the other:
  `[ERROR] … duplicate node id "Pod/default/dup": declared as Pod in rendered/a.yaml ({…"app":"first"…}) and again as Pod in rendered/more/b.yaml, document 2 ({…"app":"second"…}) — the two declarations are different`
  (`pkg/k8sgraph/testdata/duplicates.yaml` for two documents of one file, `duplicates-files/` for two files; the
  operator prints that error instead of a violation list and exits 1). The position (`document N`) appears only
  when the file holds several documents; a rendered kustomization root reads `kustomize render of DIR, document N`.
  Offline manifests carry this provenance beside the resource (invisible to rules, equality and JSON); a live
  cluster has none, so a live duplicate names the payloads only.
- **References resolve in the referrer's namespace**, because Kubernetes references are namespace-local:
  an Ingress backend, a ConfigMap or claim a pod uses, a Service's selector, an `ownerReference` (which
  names the owner's kind — an owner reference without `kind` is invalid Kubernetes and an error here, `key 'kind' not found`, not a guess by name), a `roleRef` (kind and name) and a RoleBinding subject (which may carry its own
  `namespace`). `k8s_ref kind ns name g` returns the id that exists — `Kind/ns/name`, else `Kind/_/name`
  (a cluster-scoped target or a manifest without namespaces), else `Kind/ns/name` again so a missing
  target still reads as the id it was looked for under. Use it in your own relations instead of building
  ids by hand; `k8s_id_of resource` is the id of a payload and `k8s_short id` its bare name (the helpers are
  in `pkg/k8sgraph/lib/k8s/ids.gatwd`).
- **Queries use the qualified ids:** `graph_shortest_path Ingress/default/web Pod/default/nginx-pod`,
  `graph_reachable Deployment/default/nginx-deploy`. Rule results print them too, so a violation says
  which namespace it is in.
- **Live and offline must agree on the namespace.** A live crawl always reports `metadata.namespace`;
  a manifest that omits it is `Kind/_/name`. Compare like with like: give offline manifests the
  namespace they will have in the cluster (`kubectl kustomize` output already does), or the ids differ.
- Edges only join nodes that exist (`graph_has_node` guards), so a reference to something missing yields
  no edge, and an edge to an id that is no node is never followed. Find dangling references with a rule
  (see `rule_dangling_ingress`).

### 2.4 Example: Secret → database → DBMS, with a SQL (or Redis, …) adapter

Goal: a Kubernetes Secret holds credentials for a database, the database lives in a DBMS, and you want to
ask "which DBMS does this workload ultimately depend on?" across both worlds.

**Step 1 — put the data in the graph.** Your adapter yields one value per database and per DBMS, in the
same shape every resource has: nested maps with `kind`, `metadata.name` (and `metadata.namespace`, if the
source has a scope) and `spec`. The ids follow from that: `DBMS/default/pg-prod`, `Database/default/orders`.

```yaml
kind: DBMS                                     kind: Database
metadata: {name: pg-prod, namespace: default}  metadata: {name: orders, namespace: default}
spec: {engine: postgres, …}                    spec: {dbms: pg-prod}
```

You can try the whole model first with **no Go**: `examples/k8s_howto/cluster/databases.yaml` is those
records as a manifest, so `manifests` ingests them.

**Step 2 — add the relations in DSL** (`examples/k8s_howto/infra_relations.gatwd`):

```haskell
def secret_target = \s -> hamt_get "gatwd.io/database" (hamt_path "metadata.annotations" s)
def rel_uses_database = fold (\_ g -> graph_add_edges [ ["UsesDatabase", 1, (k8s_id_of s), did] | s <- (graph_payloads_of "Secret" g), …, did <- [ (k8s_ref "Database" (k8s_scope s) (secret_target s) g) ], graph_has_node did g ] g) graph_empty
def rel_hosted_on     = fold (\_ g -> graph_add_edges [ ["HostedOn", 1, (k8s_id_of d), mid] | d <- (graph_payloads_of "Database" g), …, mid <- [ (k8s_ref "DBMS" (k8s_scope d) (hamt_path "spec.dbms" d) g) ], graph_has_node mid g ] g) graph_empty
def infra_graph = k8s_graph |> rel_uses_database |> rel_hosted_on
```

```bash
gatwd-k8s examples/k8s_howto/infra.gatwd
# [Secret/default/orders-db-credentials-UsesDatabase->Database/default/orders]
# [Database/default/audit-HostedOn->DBMS/default/pg-prod Database/default/orders-HostedOn->DBMS/default/pg-prod]
# [Secret/default/orders-db-credentials Database/default/orders DBMS/default/pg-prod]   ← the path Secret → Database → DBMS
```

**Why an annotation instead of reading the credentials?** A Secret's `data` is base64, and the DSL can decode it
(`base64_decode`, `hex_decode`, `sha256`; §2.5), so the relation *could* read the database name out of the credentials.
It does not, on purpose: an annotation (`gatwd.io/database: orders`) is plaintext, visible from a live cluster with
`kubectl`, and keeps credentials out of the graph logic and out of every rule's output. Matching by *exact equality* of
a value to the name of a node in the same namespace also works (that is how the inferred `Calls` relation works:
`k8s_ref` turns the value into the Service's id). **The security call, decided and documented:** a rule may *inspect*
a decoded value inside its own comprehension (is the password empty? is it the default?) and may report a
**fingerprint** (`sha256` of the bytes, hex-encoded) or the *fact* ("weak password"), but never the decoded value: a
violation line goes to logs. Note that the standard crawl still holds the Secret's payload, `data` included, in the
graph in memory (see *Security* below); keeping it out of the graph altogether is an ingest-time redaction that the
edition does not do yet (an open question in the hardening plan).

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

Then `spawn StrVal: default |> infra_resources |> infra_graph |> graph_shortest_path Secret/default/orders-db-credentials DBMS/default/pg-prod`.
This assembly (extra adapter step + extra embedded prelude on top of `ApplyEdition`) was built and run for
this guide. Rules for the adapter: do real I/O only, yield **nested maps/lists, not strings** (with `kind`, `metadata.name`
and a `metadata.namespace` or other scope, from which the prelude builds the qualified id), turn errors
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

(put it in your override/prelude; pinned against a fake cluster by `TestLiveCrawlWithoutSecrets`: with it the crawl yields no Secret, without it both). The sample `deploy.yaml` grants **no**
access to Secrets by default.

**Design decision (2026-10-13): the tool does not redact Secret `data` at ingest.** Rules may need it (a rule that checks
a credential's shape or age cannot work on a redacted copy), and what the tool can see is already decided one layer
down: by the kubeconfig or service account it runs as, and by the RBAC role you bind to it. Restrict access there
(drop `secrets` from the role, as the sample `deploy.yaml` does) and, where an installation needs it, drop `k8s_secrets`
from the source as above; a rule result prints resource ids, never payloads.

### 2.5 Structured data inside resources: YAML, JSON, base64, fingerprints

A ConfigMap often carries a whole file as one string, and a Secret's `data` is base64. The language reads both
without a Go change (all of it generic, in core — nothing Kubernetes-specific):

| You have | Use | You get |
| :-- | :-- | :-- |
| a YAML or JSON document in a string | `yaml_parse s` · `json_parse s` | a **`Maybe`** of the same nested shape a resource has, so `hamt_path "a.b" d` navigates it; malformed text is `Nothing`, never an error |
| a value to write out | `yaml_encode v` · `json_encode v` | text (sorted keys, deterministic) |
| a base64 string (a Secret value) | `base64_decode s` | a `Maybe Bytes`; `utf8_decode b` is the `Maybe String` |
| bytes to compare or show | `sha256 b` · `hex_encode b` · `bytes_length b` | a hex fingerprint · hex text · the length |
| text to match (an image tag, a label value) | `re_match pat s` · `re_groups pat s` | a `Bool` · a `Maybe` of the groups (RE2: linear time, so an untrusted pattern cannot hang a check) |
| a quantity (`500m`, `2Gi`) | `str_to_int`, `ratio n d` | an exact `Ratio` (`ratio 500 1000` is `1r2`); sums stay exact |

`examples/k8s_howto/data.gatwd` puts them to work over `examples/k8s_howto/data/data.yaml` (a ConfigMap holding a YAML
entry and a JSON entry, and two Secrets):

```haskell
-- examples/k8s_howto/data.gatwd
def field_of = \parse key field cm -> maybe_fold nothing (\d -> hamt_get field d) (parse (hamt_path ("data." ++ key) cm))
def cores : String -> Num = \q -> maybe_fold 0 (\m -> ratio m 1000) (str_to_int (str_slice 0 ((length q) - 1) q))
def password_bytes = \s -> base64_decode (hamt_path "data.password" s)
def rule_weak_password = \g -> [ ["WeakPassword", (k8s_id_of s)] | s <- (graph_payloads_of "Secret" g), maybe_fold true (\b -> (bytes_length b) < 8) (password_bytes s) ]
def fingerprint = \s -> maybe_fold "undecodable" (\b -> str_slice 0 12 (sha256 b)) (password_bytes s)
```

```bash
gatwd-k8s examples/k8s_howto/data.gatwd -- examples/k8s_howto/data
# ∷ [3 2Gi 3r4 [[WeakPassword Secret/default/weak-secret]] [622493693890 f52fbd32b2b3]]
#     replicas (YAML) · memory (JSON) · 500m + 250m, exact · the weak Secret · both fingerprints
```

The annotations (`def cores : String -> Num`) are advisory and state intent: `gatwd --check` reports only a definite contradiction (`cores 5`), and stays silent on these files.

Things to know. **Each decoder says what happens on `Nothing`** — `maybe_fold default f m` — and here an undecodable
password counts as weak (a rule must not go quiet about what it cannot read). A dotted key (`app.yaml`) cannot be
reached with `hamt_path`, which splits on dots; use `hamt_get "app.yaml"` on the `data` map. **Decoded YAML dates stay
text** (`2026-10-13` is a `String`) until you ask: `time_parse` / `date_parse` (§2.6). **Credentials:** the rule reports
the resource id and a fingerprint, never the decoded value (the rule above, §2.4 and the security note there).

### 2.6 Ages and expiry: `Duration`, `Time` and the injectable clock

`Duration` (`30d`, `90m`, `1h30m`), `Time` (`@2026-10-13T09:30:00Z`) and `Date` (`@2026-10-13`) are separate types with
their own arithmetic: `Time - Time` is a `Duration`, `Time + Duration` a `Time`, `Duration` compares with `Duration`
(`dsl_guide.md` §15 has the full table). `time_parse s` turns `metadata.creationTimestamp` (an RFC 3339 string) into a
`Maybe Time`. A rule for "Pods older than 30 days" is then one comparison — the only question is where "now" comes from.

`now` reads the clock, a side effect, so it is a **step** and not a function: take **one reading at the edge** and pass it to
the rule as an ordinary value, which keeps the rule pure and testable. `examples/k8s_howto/age_rules.gatwd` does that,
as an operator override:

```haskell
-- examples/k8s_howto/age_rules.gatwd
let max_age = 30d
let clock = (now |> head)
def born = \r -> head ([ (time_parse (hamt_path "metadata.creationTimestamp" q)) | q <- [r], hamt_has_path "metadata.creationTimestamp" q ] ++ [ nothing ])
def older_than : Time -> Duration -> Any -> Bool = \now limit r -> maybe_fold false (\t -> (now - t) > limit) (born r)
def rule_stale_pod = \now limit g -> [ ["StalePod", (k8s_id_of p)] | p <- (graph_payloads_of "Pod" g), (older_than now limit p) ] ++ [ ["UnknownAge", (k8s_id_of p)] | p <- (graph_payloads_of "Pod" g), is_nothing (born p) ]
```

```bash
gatwd-k8s-operator check --manifests examples/k8s_howto/aged --rules examples/k8s_howto/age_rules.gatwd
gatwd-k8s examples/k8s_howto/age.gatwd -- examples/k8s_howto/aged     # the age rule alone, as a script
# ∷ [[StalePod Pod/default/old-pod] [UnknownAge Pod/default/bad-stamp-pod] [UnknownAge Pod/default/no-stamp-pod]]
```

(The output above is for a run in October 2026; `fresh-pod` was created on 2026-10-10 and becomes a `StalePod` on
2026-11-09. That is the reason the tests inject the clock rather than read it.)

- **A Pod whose age cannot be known is reported** (`UnknownAge`: no `creationTimestamp`, as in most hand-written manifests, or
  one that does not parse), never treated as fresh. A live cluster always has the stamp.
- **`let clock = (now |> head)` at the top of an override is read again before every check**: the operator evaluates the
  override on each `check` and each `run --interval` round, so a long-running operator does not judge by the clock of its start.
  `now` needs no `--allow-*` flag (it is not an OS command, file or socket).
- **Tests inject the clock** and never wait or read the wall clock: `interp := edition.BuildInterpreter(); interp.Clock = func() time.Time { … }`
  (`TestHowToAgeRule` runs the rule at three different instants, including one before the oldest Pod).
- **Limits:** a difference beyond about ±292 years is an error (`Duration overflow`), so a stamp in the year 3000 makes the
  whole check fail loudly instead of wrapping. `-D MAX_AGE=30d` and `const MAX_AGE = 30d` bind a real
  `Duration` (a canonical Duration, Time or Date token is typed like a number; `007s` or a quoted `"30d"` stay text), so
  `age > MAX_AGE` works inside a rule. Certificate expiry works the same way
  (`(expires - now) < 14d`) once the date has been read out of the Secret.

---

## 3. Building, installing and deploying

### 3.1 Build

The module needs **Go 1.27** (`go.mod` says `go 1.27.0`; the in-process kustomize renderer requires it). An older `go` command fetches
the toolchain by itself (`GOTOOLCHAIN=auto`, the default); a container build stage must use a `golang:1.27` image or newer. The Kubernetes client libraries (`client-go`, `apimachinery`) are at `v0.38.0-alpha.2`, a pre-release; revisit when v0.38.0 is tagged.

```bash
go build -o bin/gatwd-k8s ./cmd/gatwd-k8s
go build -o bin/gatwd-k8s-operator ./cmd/gatwd-k8s-operator

# static, reproducible, smaller — what you want for a container (no cgo needed)
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/gatwd-k8s-operator ./cmd/gatwd-k8s-operator
```

The static operator is a single ELF binary of roughly 33 MB, 31 MiB (the in-process kustomize library accounts for about 3 MB of that). To install into `$GOBIN` (default
`~/go/bin`) from a checkout: `go install ./cmd/gatwd-k8s ./cmd/gatwd-k8s-operator`.

### 3.2 Run it

| | `gatwd-k8s` | `gatwd-k8s-operator` |
| :-- | :-- | :-- |
| Input | one script file, or stdin at a prompt | flags below |
| Cluster access | in-cluster service account, else `$KUBECONFIG`, else `~/.kube/config`; built lazily on the first crawl — with none reachable a crawl step yields `[ERROR] no cluster access (…)` (offline steps are unaffected) | the same; with `--manifests` it never needs one |
| Flags | `[script] [-- args]`, `-i`, `--task NAME`, `-e STMT`, `-D NAME=VALUE`, `--prelude FILE`, `--no-prelude`, `--check file`; subcommands `run eval check task repl prelude builtins version completion` (see the DSL guide §23) | subcommands `check` (once), `run --interval DUR` (repeat), `rules cat` (print the embedded rules), `version`, `completion`; flags `--manifests PATH` (omit for live), `--namespace NS` (default `default`; `""` = all namespaces), `--rules FILE` (default `/etc/gatwd/rules.gatwd`), `--rules-mode layer\|replace` (default `layer`, §2.2), `--interval DUR`, `--format text\|json`, `--no-kustomize` (read a kustomization root as plain files, §3.2), `--fail-on-violation` (exit 3, §3.2), `--timeout DUR` and `--max-output BYTES` (per-check limits, both off by default, §3.2), and the execution policy `--allow-exec`, `--allow-fs`, `--allow-net`, `--allow-env` (all off, §2.2). Without a subcommand the operator behaves as before: `--interval` makes it loop, otherwise it checks once. The single-dash spellings of earlier releases (`-manifests`, `-rules`, `-interval`, …) still work for one release and print a deprecation notice on stderr |
| Output | each result on stdout | one result per check: `[]` when compliant, otherwise `[[Rule resource] …]` |

**Exit codes.** `0` the check ran; `1` the command failed — an override that does not load, a rule that errors
(`[ERROR] …`, including a duplicate node id), a pipeline that yields nothing, a check that went over `--timeout` or
`--max-output` (below); `2` a malformed command line
(unknown flag or `--format`, no `--manifests` and no reachable cluster); `3` the check ran and found violations,
**only with `--fail-on-violation`**. Without that flag the operator reports violations but exits 0, so a
scheduled `run` is not restarted for finding something. A check that both errored and found violations exits `1`.
For a CI gate that tells a policy finding from a broken run:

```bash
gatwd-k8s-operator check --manifests ./rendered --fail-on-violation
case $? in
  0) echo "compliant" ;;
  3) echo "policy violations"; exit 1 ;;
  *) echo "the check itself failed"; exit 2 ;;
esac
```

For a log pipeline or a CI artifact add `--format json` (below); the exit status is the same, so
`check … --fail-on-violation --format json > report.json` both stores the report and gates the job.

**Kustomize.** The `manifests` step renders a *kustomization root* — a directory holding `kustomization.yaml`,
`kustomization.yml` or `Kustomization` — by itself, **in process**: the kustomize library is linked into the
binary, so no `kustomize`, `kubectl` or `git` is needed, not even in a scratch container. Every other path (a
file, a directory without such a file) is read exactly as before. The operator's `-manifests overlay/` renders
the same way.

```bash
gatwd-k8s eval -e 'spawn StrVal: pkg/k8sgraph/testdata/kustomize/overlays/prod |> manifests |> k8s_graph |> graph_order'      # ∷ 4: rendered, prefix and namespace applied
gatwd-k8s --no-kustomize eval -e 'spawn StrVal: pkg/k8sgraph/testdata/kustomize/overlays/prod |> manifests |> collect_list |> length'   # ∷ 3: the plain files
```

- **`--no-kustomize`** (a persistent flag of `gatwd-k8s` and of `gatwd-k8s-operator`; the operator's flag is the
  edition's `k8sgraph.WithoutKustomize()` option) makes
  `manifests` read a root as the plain files it contains, as it did before auto-render. The root's own
  `kustomization.yaml` is then one of the documents and usually has no `kind`, so building a graph from it fails with
  `key 'kind' not found`: the switch is for a tree of already-rendered YAML that happens to hold such a file. The
  operator renders by default too: `gatwd-k8s-operator check --manifests DIR --no-kustomize` reads the root as plain
  files. The flag belongs to the `--manifests` source; with the live cluster (no `--manifests`) it is a usage error
  (exit 2) rather than a flag that silently does nothing.
- The separate **`kustomize` step** (`… |> kustomize |> k8s_graph`) always renders, whatever the flag says. Use
  `kubectl kustomize overlay > rendered/all.yaml` only if you want an artifact on disk.
- Only the path you give is tested for being a root. A directory *tree* that contains roots (`overlays/dev`,
  `overlays/prod`) is read file by file, so point `manifests` at the overlay you mean.
- The output is what `kustomize build` prints (the tests compare it with `kubectl kustomize` when that is
  installed). **Refused, loudly:** exec plugins, KRM functions and Helm charts; remote bases and files (`https://…`,
  `git@…`, `github.com/…`), which are rejected *before* anything is fetched; files outside the kustomization root
  (kustomize's load restriction). A broken overlay is `[ERROR] kustomize: rendering <root> failed: <kustomize's
  message>` and a non-zero exit, never an empty graph.
- Cost: the kustomize library adds about 3 MB (11%) to the stripped static binaries (§3.1).

`run --interval` is a daemon: a failing check is printed and the loop continues; only a start-up failure ends it
(and `--fail-on-violation` is rejected there).

**Limits per check: `--timeout` and `--max-output`.** Rules (an override above all) can be slow or can produce more than
anyone wants to read, so one check can be bounded in time and in size. Both are **off by default** (`0` is no limit);
a negative value is a usage error (exit 2). Each `check`, and each round of `run --interval`, gets its own fresh limit —
a slow round does not eat the next one's time.

| Flag | Meaning |
| :-- | :-- |
| `--timeout DUR` | stop a check that runs longer than `DUR` (`30s`, `2m`): every pipeline step stops at its next node and an OS command a step started (`--allow-exec`) is killed |
| `--max-output BYTES` | fail a check whose results are longer than `BYTES` as the text report prints them (each result and its newline); a stream that never ends is stopped as soon as it passes the limit |

A check that goes over a limit is a **failed check**: exit `1` (not `3`, even under `--fail-on-violation`), and the
report is that one error — the results it had produced up to then are dropped, never mixed in and never silently cut
short:

```bash
gatwd-k8s-operator check --manifests examples/k8s_howto/cluster --max-output 40
# [ERROR] check output exceeded the limit of 40 bytes            (exit 1)
gatwd-k8s-operator check --manifests DIR --rules slow.gatwd --allow-exec --timeout 30s
# [ERROR] check timed out after 30s                              (exit 1)
```

In `--format json` the same failure is an entry in `errors` and a structured one in `diagnostics`:
`{"kind":"timeout","message":"check timed out after 30s","limit":"30s"}` or
`{"kind":"output_limit","message":"check output exceeded the limit of 40 bytes","limit":"40"}`. `run --interval`
reports the failed round like any other failing check and carries on with the next one. Sensible starting values for a
scheduled `run --interval 5m`: `--timeout 2m` (shorter than the interval, so rounds never overlap) and `--max-output
1048576` (1 MiB of report per round); the limits bound what a check *emits*, the timeout is what bounds the memory and
time it spends computing. Loading the rules (start-up, hot reload) is not part of a check and is not bounded.

**Output formats.** `--format text` (default) prints the result as before: `[]` when compliant, otherwise
`[[Rule resource] …]`; an error prints as `[ERROR] …`. `--format json` prints **one JSON object per check, on one
line**, for log pipelines (every key always present):

```bash
gatwd-k8s-operator check --manifests examples/k8s_howto/cluster --format json
# {"status":"violations","count":3,"violations":[{"rule":"IdleDeployment","resource":"Deployment/default/worker"},…],"errors":[],"diagnostics":[]}
```

`status` is `ok`, `violations` or `error`; a resource that is a list (a dependency cycle) stays a JSON array.
`errors` holds the failure messages as plain strings and `diagnostics` the same failures as objects with a `kind`
(`capability_denied`, `unknown_step`, `stopped`, `no_result`, `unencodable_value`, `error`) and, where they apply,
the `step` and `capability` — so a log pipeline can alert on "an override tried to run a program" without parsing text.
Diagnostics — the rules report, reload failures — always go to **stderr**, so stdout stays machine-readable.

The operator renders kustomization roots the same way `manifests` does (§3.2); the separate `kustomize` step in `gatwd-k8s` scripts always renders, in process.

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

The image needs nothing else for live checks, plain manifests or kustomize overlays: rendering happens inside the binary.

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
- **Bounding a round:** the sample's `args:` carry `--timeout 2m` and `--max-output 1048576` as a commented-out line; with
  an override that can run a program (`--allow-exec`) a timeout is what keeps a stuck command from stalling the loop (§3.2).
- **What the override may do:** the Deployment's `args:` are the only place a capability is granted (§2.2). A mounted
  `rules.gatwd` that `import`s a sibling key needs `--allow-fs` in `args:`; one that runs a program needs `--allow-exec`;
  a pure rule set (the sample's) needs nothing. Editing the ConfigMap can never add a flag.
- **Add a kind** (a CRD): add a crawl step in `pkg/k8sgraph` (`CoreCrawls`), list it in
  `lib/k8s/sources.gatwd`, and add `get`/`list` for it to the `ClusterRole`.

### 3.5 Troubleshooting

| Symptom | Likely cause |
| :-- | :-- |
| `[ERROR] no cluster access (…)` from a `k8s_*` crawl step | kubeconfig auth plugin not supported or no cluster; use `--manifests` / the `manifests` step or fix the kubeconfig |
| `no --manifests given and no cluster reachable` (exit 2) | the operator has no source |
| Operator prints `[ERROR] …` after you added a rule | a name in the rule does not resolve, or an operator rejected its operands; the message names it (§2.1) |
| Rule output is empty but should not be | a node kind is missing from the source list or RBAC (§2.3) |
| `[ERROR] … duplicate node id "Kind/ns/name"` | the same resource is defined twice (the message quotes both); remove one, or give them different names or namespaces (§2.3) |
| A reference you expect has no edge, or live and offline disagree | the two sides carry different namespaces (`Pod/_/x` vs `Pod/default/x`), or the target lives in another namespace than the referrer (§2.3) |
| `[ERROR] step 'x' requires the exec capability (policy denies it)` (or `fs`, `net`) from an override | the override runs a program, reads a file or `import`s one; grant it on the command line (`--allow-exec`, `--allow-fs`, …) or remove it (§2.2) |
| `[ERROR] check timed out after 30s` (exit 1) | the check ran longer than `--timeout` (a stuck command in an override, or a cluster crawl that is too slow); raise the limit or fix the rule (§3.2) |
| `[ERROR] check output exceeded the limit of N bytes` (exit 1) | the report is larger than `--max-output`; raise it or narrow the rules (§3.2) |
| `[ERROR] … Duration overflow: the difference between … is beyond about ±292 years` from an age rule | a `creationTimestamp` (or expiry) centuries away from now; fix the data (§2.6) |
| `key 'kind' not found` after `--no-kustomize`, or from an `ownerReference` | the root's `kustomization.yaml` was read as a document (§3.2), or the owner reference names no `kind` (§2.3) |
| Edge you expect is missing | an endpoint node does not exist (`graph_has_node` guard), or the relation guards a field the resource lacks |

---

## 4. Known limitations (as of this writing)

- Node ids are `Kind/namespace/name` (§2.3). The duplicate error names the file (and document) of both
  declarations for offline manifests; for a live crawl, which has no files, it quotes the payloads only.
- `--no-kustomize` exists on both `gatwd-k8s` and `gatwd-k8s-operator` (§3.2); on the operator it applies to `--manifests` only.
- The operator has `--timeout` and `--max-output` per check (§3.2), but no signal handling: `run --interval` ends only
  when the process is killed. A single closure that never returns cannot be interrupted at all (the timeout is checked
  between nodes, and kills the OS commands a step started), and `--max-output` bounds what a check emits, not the
  memory it uses while it computes.
- The policy's `env` capability gates nothing yet: no builtin reads the environment (§2.2).
- Rules cannot read the clock inside a lambda: `now` is a step, so an age rule takes the reading as a parameter from a
  top-level `let clock = (now |> head)` (§2.6). `-D X=30d` and `const X = 30d` are a `Duration`.
- The standard crawl holds Secrets, `data` included, in the graph in memory (§2.4, *Security*); there is no ingest-time
  redaction.
- `||>` runs the *following per-element step* across the stream on a worker pool (order is not
  deterministic); a whole-stream step such as `fold` or `sort` after `||>` would see single nodes.
