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
// # Quick start (Edition assembly)
//
//	client, _ := k8sgraph.DynamicClientFromKubeconfig()
//	cli.NewEdition("gatwd-k8s").
//	    WithStep(k8sgraph.CrawlStep("k8s_pods", client, k8sgraph.PodsGVR)).
//	    Run()
//
// Then in the DSL:  spawn StrVal: default |> k8s_pods |> hamt_path spec.selector
package k8sgraph
