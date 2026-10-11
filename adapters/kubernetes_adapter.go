package adapters

import (
	"context"
	"iter"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

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
