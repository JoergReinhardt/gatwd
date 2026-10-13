package adapters

import (
	"context"
	"iter"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
)

// NativeKubernetesAdapter leverages the k8s.io/client-go dynamic client.
//
// Because gatwd parses DSL types dynamically at runtime, we cannot use
// strongly-typed Kubernetes structs (like corev1.Pod). Instead, we use
// the DynamicClient and GroupVersionResource (GVR). This allows the Graph
// Crawler to query ANY resource in Kubernetes (including Custom Resource
// Definitions / CRDs) completely on the fly!
//
// The `node.Value` passed to this adapter is treated as the Namespace to search.
// If it is empty, it searches across all namespaces.
func NativeKubernetesAdapter(client dynamic.Interface, gvr schema.GroupVersionResource, targetFlag dsltypes.BitFlag) evaluator.Adapter[string] {
	return func(node dsltypes.ASTNode[string]) iter.Seq[dsltypes.ASTNode[string]] {
		return func(yield func(dsltypes.ASTNode[string]) bool) {

			namespace := node.Value

			// Use the dynamic client to fetch ANY Kubernetes resource
			list, err := client.Resource(gvr).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return // Silently halt this branch on RBAC or network error
			}

			// Lazily yield the discovered resources back into the Trampoline
			for _, item := range list.Items {
				// We yield the name of the resource. We could also yield the raw JSON
				// or the unstructured.Unstructured object itself if we changed `T any`
				// in the engine to support complex objects.
				childNode := dsltypes.NewASTNode(item.GetName(), targetFlag)

				if !yield(childNode) {
					return
				}
			}
		}
	}
}

// unstructuredToValue recursively lifts a decoded Kubernetes resource
// (the generic Go shape produced by k8s.io/apimachinery's JSON/YAML decoder —
// `map[string]any` objects, `[]any` arrays, and scalar leaves) into gatwd's
// own persistent collections, so the DSL can navigate it with the existing
// HAMT/List builtins (`hamt_get`, `hamt_values`, list comprehensions, …).
//
// The mapping is the obvious structure-preserving one:
//
//   - a JSON object `map[string]any` becomes an immutable `*collections.HAMT[string, any]`
//     (keys are always strings in Kubernetes manifests), so `… |> hamt_get spec
//     |> hamt_get selector` walks the nesting;
//   - a JSON array `[]any` becomes an immutable `*collections.List[any]` (preserving
//     order), so `ownerReferences`, `containers`, `ports`, … iterate in the DSL;
//   - every scalar leaf (`string`, `float64`/`int64`, `bool`, `nil`) is carried
//     through unchanged.
//
// Because both target collections are persistent (structurally shared), an
// operator reacting to a watch event can add/remove a few resources without
// deep-copying the whole graph (design doc §6 #6).
func unstructuredToValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		h := collections.NewHAMT[string, any]()
		for k, val := range t {
			h = h.Set(k, unstructuredToValue(val))
		}
		return h
	case []any:
		// Build the cons-list back-to-front so the original order is preserved.
		lst := collections.Nil[any]()
		for i := len(t) - 1; i >= 0; i-- {
			lst = collections.Cons(unstructuredToValue(t[i]), lst)
		}
		return lst
	default:
		return v
	}
}

// NativeKubernetesResourceAdapter is the structured (graph-building) successor to
// NativeKubernetesAdapter: instead of yielding a resource's bare *name* as a
// `string`, it yields the resource's **full content** as a nested
// `*collections.HAMT[string, any]` payload — metadata, labels, ownerReferences,
// spec and status all navigable from the DSL via the existing `hamt_get` /
// `hamt_values` builtins.
//
// Why a separate adapter (not a flag on the old one)? Graph construction needs
// the whole resource to discover typed, directional edges — a Service's
// `spec.selector`, a child's `metadata.ownerReferences`, an Ingress backend —
// none of which survive in a name-only projection. The name-only
// NativeKubernetesAdapter is kept for back-compat with simple crawls.
//
// The returned adapter is a pure `Adapter[any]`; the DSL engine runs over
// `Adapter[functions.Either[error, any]]`, so callers lift it with
// evaluator.EitherT (see pkg/k8sgraph.CrawlStep / cli.AdapterStep). A List error
// is surfaced as a node whose value is the `error`, which EitherT converts into a
// Left so the failure propagates loudly instead of silently halting the branch.
//
// The input node's value is treated as the namespace to list; an empty / non-string
// value lists across all namespaces (`Namespace("")`). The GVR may name any
// built-in resource *or any CRD*, since the dynamic client is schema-agnostic;
// EnumerateResources discovers the full GVR set (incl. CRDs) at runtime.
func NativeKubernetesResourceAdapter(client dynamic.Interface, gvr schema.GroupVersionResource) evaluator.Adapter[any] {
	return func(node dsltypes.ASTNode[any]) iter.Seq[dsltypes.ASTNode[any]] {
		return func(yield func(dsltypes.ASTNode[any]) bool) {

			namespace := ""
			if ns, ok := node.Value.(string); ok {
				namespace = ns
			}

			list, err := client.Resource(gvr).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
			if err != nil {
				// Surface the failure as an error-valued node; EitherT lifts it to Left.
				yield(dsltypes.NewASTNode[any](err, node.DSLType))
				return
			}

			for _, item := range list.Items {
				payload := unstructuredToValue(item.Object)
				if !yield(dsltypes.NewASTNode[any](payload, node.DSLType)) {
					return
				}
			}
		}
	}
}

// EnumerateResources uses a discovery client to list every listable resource the
// API server exposes — built-in kinds **and** CRDs — as a slice of
// GroupVersionResource values ready to feed NativeKubernetesResourceAdapter.
//
// This is what makes "lift every resource in the cluster into the graph" hold
// without a hard-coded GVR table: CRDs installed on the cluster appear in
// discovery just like core kinds. Subresources (names containing "/", e.g.
// "pods/log") and non-listable resources are skipped, since they are not graph
// nodes. Partial-discovery errors (a single unavailable API group) are tolerated:
// whatever groups *did* resolve are still returned alongside the error, matching
// client-go's ServerPreferredResources contract.
func EnumerateResources(disco discovery.DiscoveryInterface) ([]schema.GroupVersionResource, error) {
	lists, err := disco.ServerPreferredResources()

	var gvrs []schema.GroupVersionResource
	for _, rl := range lists {
		if rl == nil {
			continue
		}
		gv, parseErr := schema.ParseGroupVersion(rl.GroupVersion)
		if parseErr != nil {
			continue
		}
		for _, r := range rl.APIResources {
			if strings.Contains(r.Name, "/") {
				continue // subresource, not a standalone node
			}
			if !containsVerb(r.Verbs, "list") {
				continue // cannot be crawled as a collection
			}
			gvrs = append(gvrs, gv.WithResource(r.Name))
		}
	}
	return gvrs, err
}

// containsVerb reports whether the API resource advertises the given verb
// (e.g. "list") in its discovery metadata.
func containsVerb(verbs metav1.Verbs, want string) bool {
	for _, v := range verbs {
		if v == want {
			return true
		}
	}
	return false
}
