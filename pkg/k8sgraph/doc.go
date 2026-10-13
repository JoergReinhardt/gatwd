// Package k8sgraph is the thin, domain-specific Go shim that bridges a live
// Kubernetes cluster to the generic gatwd graph DSL. It is deliberately small:
// its only jobs are I/O plumbing (building dynamic/discovery clients from a
// kubeconfig or in-cluster service account) and registration (handing configured
// resource adapters to the generic cli.AdapterStep as named pipeline steps).
//
// # Where the logic lives (Mandate 5, Clean Core)
//
// This package contains NO graph algorithms and NO Kubernetes relation rules.
//
//   - The graph abstraction (nodes, edges, walks, folds, optics) is generic core
//     under collections/, comonads/, optics/, evaluator/ and the graph_* DSL
//     builtins — it never mentions Kubernetes.
//   - The resource I/O (dynamic client → nested-HAMT payload, discovery → GVR set)
//     is generic-enough-to-reuse adapter code under adapters/ (client-go is
//     confined there and here).
//   - The relation rules themselves (Owns, Selects, RoutesTo, …) are authored in
//     pure gatwd DSL in the embedded prelude (lib/k8s/*.gatwd), NOT Go loops.
//
// So this shim is pure assembly: it wires core + adapters + the DSL prelude
// together into an Edition/Appliance. See docs/k8s_graph_design.md §4.
//
// # The prelude (lib/k8s/*.gatwd)
//
//   - types.gatwd      node colors (resource kinds) and edge colors (relations)
//   - ingest.gatwd     k8s_ingest: a resource stream folded into a typed graph
//   - relations.gatwd  Owns, Selects (label-indexed), RoutesTo, Configures, Mounts,
//     Binds and the INFERRED Calls, plus the helper closures they share
//   - sources.gatwd    k8s_resources: every live crawl stream merged with <+>
//   - queries.gatwd    k8s_apply_order, k8s_cycles
//   - prelude.gatwd    the embed entry point; defines k8s_graph
//
// One resource stream feeds one pipeline, offline or live — the downstream is
// source-agnostic because both produce the same nested-HAMT payloads:
//
//	spawn StrVal: ./manifests |> manifests |> k8s_graph       (offline)
//	spawn StrVal: default |> k8s_resources |> k8s_graph       (live cluster)
//
// # Kustomize, in process
//
// The `manifests` step renders a kustomization root on its own: the
// sigs.k8s.io/kustomize library runs inside the process (InProcessRenderer; no
// kustomize, kubectl or git binary), and the rendered resources enter the graph like
// any other manifest. Files and directories that are not kustomization roots are
// read as before. The edition's `--no-kustomize` flag (ApplyEdition's option
// WithoutKustomize) switches that off; the explicit `kustomize` step always renders.
// Exec plugins, Helm and remote bases are refused, and a render error is an [ERROR]
// naming the root. The library is imported here and nowhere else (Clean Core).
//
//	spawn StrVal: ./overlays/prod |> manifests |> k8s_graph
//
// # Quick start (Edition assembly)
//
//	client := k8sgraph.LazyDynamicClient() // built on the first crawl, never for `version` or offline work
//	cli.NewEdition("gatwd-k8s").
//	    WithStep(k8sgraph.CrawlStep("k8s_pods", client, k8sgraph.PodsGVR)).
//	    Run()
//
// Then in the DSL:  spawn StrVal: default |> k8s_pods |> hamt_path spec.selector
package k8sgraph
