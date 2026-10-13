package k8sgraph

import (
	"context"
	"fmt"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

// LazyDynamicClient is DynamicClientFromKubeconfig deferred to first use.
//
// An Edition registers its live crawl steps when the process starts, long before it knows
// whether the command line will touch a cluster at all (`version`, `completion`, `check`
// and a script that only reads manifests never do). Building the client eagerly made those
// commands load a kubeconfig and print a "no cluster access" notice. The lazy client builds
// the real client the first time a crawl asks for a resource, once, and reports a failure
// where it matters: a crawl step then yields an `[ERROR] no cluster access (…)` node that
// names the cause, instead of the step silently not existing.
func LazyDynamicClient() dynamic.Interface {
	return LazyClient(DynamicClientFromKubeconfig)
}

// LazyClient defers build until the first Resource call and remembers its outcome, success
// or failure, so the client is constructed at most once per process.
func LazyClient(build func() (dynamic.Interface, error)) dynamic.Interface {
	return &lazyClient{build: build}
}

type lazyClient struct {
	build func() (dynamic.Interface, error)
	once  sync.Once
	c     dynamic.Interface
	err   error
}

func (l *lazyClient) get() (dynamic.Interface, error) {
	l.once.Do(func() {
		l.c, l.err = l.build()
		if l.err != nil {
			l.err = fmt.Errorf("no cluster access (%w); use manifests for offline work", l.err)
		}
	})
	return l.c, l.err
}

// Resource implements dynamic.Interface. When the client cannot be built it returns a
// resource whose every operation fails with that error.
func (l *lazyClient) Resource(gvr schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	c, err := l.get()
	if err != nil {
		return failedResource{err: err}
	}
	return c.Resource(gvr)
}

// failedResource is the resource of a client that could not be built: every call returns
// the construction error, so nothing downstream can mistake it for an empty cluster.
type failedResource struct{ err error }

func (f failedResource) Namespace(string) dynamic.ResourceInterface { return f }

func (f failedResource) Create(context.Context, *unstructured.Unstructured, metav1.CreateOptions, ...string) (*unstructured.Unstructured, error) {
	return nil, f.err
}

func (f failedResource) Update(context.Context, *unstructured.Unstructured, metav1.UpdateOptions, ...string) (*unstructured.Unstructured, error) {
	return nil, f.err
}

func (f failedResource) UpdateStatus(context.Context, *unstructured.Unstructured, metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, f.err
}

func (f failedResource) Delete(context.Context, string, metav1.DeleteOptions, ...string) error {
	return f.err
}

func (f failedResource) DeleteCollection(context.Context, metav1.DeleteOptions, metav1.ListOptions) error {
	return f.err
}

func (f failedResource) Get(context.Context, string, metav1.GetOptions, ...string) (*unstructured.Unstructured, error) {
	return nil, f.err
}

func (f failedResource) List(context.Context, metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return nil, f.err
}

func (f failedResource) Watch(context.Context, metav1.ListOptions) (watch.Interface, error) {
	return nil, f.err
}

func (f failedResource) Patch(context.Context, string, types.PatchType, []byte, metav1.PatchOptions, ...string) (*unstructured.Unstructured, error) {
	return nil, f.err
}

func (f failedResource) Apply(context.Context, string, *unstructured.Unstructured, metav1.ApplyOptions, ...string) (*unstructured.Unstructured, error) {
	return nil, f.err
}

func (f failedResource) ApplyStatus(context.Context, string, *unstructured.Unstructured, metav1.ApplyOptions) (*unstructured.Unstructured, error) {
	return nil, f.err
}
