package adapters_test

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"

	"github.com/joergreinhardt/gatwd/adapters"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
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
