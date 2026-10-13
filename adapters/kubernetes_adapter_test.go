package adapters_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic/fake"

	"github.com/joergreinhardt/gatwd/adapters"
	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

func TestNativeKubernetesAdapter(t *testing.T) {
	// 1. Define the GVR we are looking for (e.g., Pods)
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	// 2. Setup a Fake Dynamic Client with some mock data
	mockPod1 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "nginx-pod",
				"namespace": "default",
			},
		},
	}
	mockPod2 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "redis-pod",
				"namespace": "default",
			},
		},
	}

	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClient(scheme, mockPod1, mockPod2)

	const IsKubePod = dsltypes.BitFlag(1 << 9)
	const IsNamespace = dsltypes.BitFlag(1 << 10)

	// 3. Initialize the Native Adapter
	adapter := adapters.NativeKubernetesAdapter(fakeClient, gvr, IsKubePod)

	// 4. Register it with the Engine
	engine := evaluator.NewEngine[string]()
	engine.Register(IsNamespace, adapter)

	// 5. Create a stream with the root namespace "default"
	initialStream := func(yield func(dsltypes.ASTNode[string]) bool) {
		yield(dsltypes.NewASTNode("default", IsNamespace))
	}

	// 6. Evaluate!
	resultStream := engine.Evaluate(initialStream)

	var results []string
	for node := range resultStream {
		if !node.Match(IsKubePod) {
			t.Fatalf("Expected output to have IsKubePod flag")
		}
		results = append(results, node.Value)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 pods discovered, got %d", len(results))
	}
	if results[0] != "nginx-pod" || results[1] != "redis-pod" {
		t.Fatalf("Unexpected execution output: %v", results)
	}
}

// hamtGet navigates one level of a nested-HAMT resource payload in tests.
func hamtGet(t *testing.T, v any, key string) any {
	t.Helper()
	h, ok := v.(*collections.HAMT[string, any])
	if !ok {
		t.Fatalf("expected *HAMT at key %q, got %T", key, v)
	}
	res, found := h.Get(key)
	if !found {
		t.Fatalf("key %q not found in HAMT", key)
	}
	return res
}

func TestNativeKubernetesResourceAdapter(t *testing.T) {
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	// A resource with the full nested shape the graph builder needs: nested
	// metadata, labels, a one-element ownerReferences array, and a spec.
	mockPod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "nginx-pod",
				"namespace": "default",
				"labels": map[string]interface{}{
					"app": "nginx",
				},
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"kind": "ReplicaSet",
						"name": "nginx-rs",
					},
				},
			},
			"spec": map[string]interface{}{
				"serviceAccountName": "nginx-sa",
			},
		},
	}

	scheme := runtime.NewScheme()
	fakeClient := fake.NewSimpleDynamicClient(scheme, mockPod)

	adapter := adapters.NativeKubernetesResourceAdapter(fakeClient, gvr)

	var payloads []any
	for node := range adapter(dsltypes.NewASTNode[any]("default", dsltypes.BitFlag(0))) {
		payloads = append(payloads, node.Value)
	}

	if len(payloads) != 1 {
		t.Fatalf("expected 1 resource payload, got %d", len(payloads))
	}

	root := payloads[0]
	// Top level must be a HAMT, not a bare name string.
	if _, ok := root.(*collections.HAMT[string, any]); !ok {
		t.Fatalf("expected top-level *HAMT payload, got %T", root)
	}

	// Scalar leaf carried through unchanged.
	if kind := hamtGet(t, root, "kind"); kind != "Pod" {
		t.Fatalf("kind = %v, want Pod", kind)
	}

	// Nested object -> nested HAMT, navigable with repeated Get.
	meta := hamtGet(t, root, "metadata")
	if name := hamtGet(t, meta, "name"); name != "nginx-pod" {
		t.Fatalf("metadata.name = %v, want nginx-pod", name)
	}
	labels := hamtGet(t, meta, "labels")
	if app := hamtGet(t, labels, "app"); app != "nginx" {
		t.Fatalf("metadata.labels.app = %v, want nginx", app)
	}

	// Array -> List, order-preserving, elements recursively lifted.
	owners, ok := hamtGet(t, meta, "ownerReferences").(*collections.List[any])
	if !ok {
		t.Fatalf("ownerReferences should be a *List")
	}
	if owners.Len() != 1 {
		t.Fatalf("ownerReferences len = %d, want 1", owners.Len())
	}
	head, _ := owners.Head()
	if ownerName := hamtGet(t, head, "name"); ownerName != "nginx-rs" {
		t.Fatalf("ownerReferences[0].name = %v, want nginx-rs", ownerName)
	}

	spec := hamtGet(t, root, "spec")
	if sa := hamtGet(t, spec, "serviceAccountName"); sa != "nginx-sa" {
		t.Fatalf("spec.serviceAccountName = %v, want nginx-sa", sa)
	}
}

// TestNativeKubernetesResourceAdapterThroughEitherT verifies the pure
// Adapter[any] lifts cleanly into the DSL's Either[error, any] engine payload.
func TestNativeKubernetesResourceAdapterThroughEitherT(t *testing.T) {
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	cm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata":   map[string]interface{}{"name": "app-config", "namespace": "default"},
		},
	}
	fakeClient := fake.NewSimpleDynamicClient(runtime.NewScheme(), cm)

	lifted := evaluator.EitherT(adapters.NativeKubernetesResourceAdapter(fakeClient, gvr))

	seed := functions.Right[error, any](any("default"))
	count := 0
	for node := range lifted(dsltypes.NewASTNode(seed, dsltypes.BitFlag(0))) {
		if node.Value.IsLeft() {
			err, _ := node.Value.UnwrapLeft()
			t.Fatalf("unexpected Left: %v", err)
		}
		val, _ := node.Value.UnwrapRight()
		if kind := hamtGet(t, val, "kind"); kind != "ConfigMap" {
			t.Fatalf("kind = %v, want ConfigMap", kind)
		}
		count++
	}
	if count != 1 {
		t.Fatalf("expected 1 Right payload, got %d", count)
	}
}

// stubDiscovery is a minimal discovery.DiscoveryInterface that returns a fixed
// preferred-resource list, so EnumerateResources' filtering can be asserted
// without a live API server (the client-go fake returns nil here).
type stubDiscovery struct {
	discovery.DiscoveryInterface
	resources []*metav1.APIResourceList
}

func (s *stubDiscovery) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return s.resources, nil
}

func TestEnumerateResources(t *testing.T) {
	disco := &stubDiscovery{
		resources: []*metav1.APIResourceList{
			{
				GroupVersion: "v1",
				APIResources: []metav1.APIResource{
					{Name: "pods", Namespaced: true, Kind: "Pod", Verbs: metav1.Verbs{"list", "get"}},
					{Name: "pods/log", Namespaced: true, Kind: "Pod", Verbs: metav1.Verbs{"get"}}, // subresource -> skip
					{Name: "bindings", Namespaced: true, Kind: "Binding", Verbs: metav1.Verbs{"create"}}, // not listable -> skip
				},
			},
			{
				// A CRD group must appear just like core kinds.
				GroupVersion: "example.com/v1",
				APIResources: []metav1.APIResource{
					{Name: "widgets", Namespaced: true, Kind: "Widget", Verbs: metav1.Verbs{"list"}},
				},
			},
		},
	}

	gvrs, err := adapters.EnumerateResources(disco)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := map[schema.GroupVersionResource]bool{}
	for _, g := range gvrs {
		got[g] = true
	}

	wantPods := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	wantWidgets := schema.GroupVersionResource{Group: "example.com", Version: "v1", Resource: "widgets"}
	if !got[wantPods] {
		t.Errorf("expected pods GVR in result, got %v", gvrs)
	}
	if !got[wantWidgets] {
		t.Errorf("expected CRD widgets GVR in result, got %v", gvrs)
	}
	if len(gvrs) != 2 {
		t.Fatalf("expected exactly 2 listable, non-subresource GVRs, got %d: %v", len(gvrs), gvrs)
	}
}
