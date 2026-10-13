// Package k8sgraphtest is test support for the Kubernetes graph edition: it turns an
// offline manifest fixture into a fake client-go dynamic client, i.e. the same
// cluster a live crawl would see. It exists so the live/offline parity test, the
// operator's live-source test and benchmarks share one fixture loader. Production
// code must not import it (it depends on client-go's fake package).
package k8sgraphtest

import (
	"bytes"
	"io"
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic/fake"

	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
)

// listKinds maps each crawled resource to the list kind the fake client needs
// registered before it can List that resource.
var listKinds = map[string]string{
	"pods": "PodList", "services": "ServiceList", "configmaps": "ConfigMapList", "secrets": "SecretList",
	"endpoints": "EndpointsList", "deployments": "DeploymentList", "replicasets": "ReplicaSetList",
	"ingresses": "IngressList", "persistentvolumeclaims": "PersistentVolumeClaimList",
	"serviceaccounts": "ServiceAccountList", "roles": "RoleList", "rolebindings": "RoleBindingList",
}

// FakeClusterFromYAML decodes a multi-document manifest (YAML or JSON) and returns a
// fake dynamic client holding every object in the namespace "default". Every GVR in
// k8sgraph.CoreCrawls is registered so kinds with no objects still list cleanly.
func FakeClusterFromYAML(t testing.TB, raw []byte) *fake.FakeDynamicClient {
	t.Helper()
	dec := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(raw), 4096)
	var objs []runtime.Object
	for {
		obj := &unstructured.Unstructured{}
		if err := dec.Decode(&obj.Object); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if len(obj.Object) == 0 {
			continue
		}
		obj.SetNamespace("default")
		objs = append(objs, obj)
	}
	kinds := map[schema.GroupVersionResource]string{}
	for _, c := range k8sgraph.CoreCrawls {
		kinds[c.GVR] = listKinds[c.GVR.Resource]
	}
	return fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), kinds, objs...)
}

// FakeClusterFromFile is FakeClusterFromYAML over a manifest file.
func FakeClusterFromFile(t testing.TB, path string) *fake.FakeDynamicClient {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return FakeClusterFromYAML(t, raw)
}
