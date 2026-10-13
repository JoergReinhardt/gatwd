package k8sgraphtest

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
)

const twoNamespaces = `
apiVersion: v1
kind: Pod
metadata: {name: web, namespace: prod}
---
apiVersion: v1
kind: Pod
metadata: {name: web, namespace: staging}
---
apiVersion: v1
kind: Pod
metadata: {name: bare}
`

// TestFakeClusterKeepsNamespaces: the fake cluster puts each object where its
// manifest says, and only a namespace-less object lands in "default". Before, every
// object was forced into "default", so two pods named alike in different namespaces
// could not be expressed at all.
func TestFakeClusterKeepsNamespaces(t *testing.T) {
	client := FakeClusterFromYAML(t, []byte(twoNamespaces))
	count := func(ns string) int {
		list, err := client.Resource(k8sgraph.PodsGVR).Namespace(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		return len(list.Items)
	}
	for ns, want := range map[string]int{"prod": 1, "staging": 1, "default": 1, "": 3} {
		if got := count(ns); got != want {
			t.Errorf("pods in namespace %q = %d, want %d", ns, got, want)
		}
	}
}
