package adapters

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"

	"github.com/joergreinhardt/gatwd/collections"
)

// This file is the WRITE half of the Kubernetes adapter: server-side apply on the dynamic
// client (hardening plan, "Writing to the cluster"). The read half (kubernetes_adapter.go)
// turns objects into DSL values; this half turns DSL values back into objects.
//
// # What it is allowed to do
//
// Create or update an object, with server-side apply under a named field manager; patch a
// workload (kubernetes_rollout.go). It can NOT delete: there is no delete, prune or
// delete-collection call anywhere in the adapters, and writableResource below is the narrow
// view of a resource the write code works through, so a delete is not even expressible. A
// guard test scans these sources for one.
//
// # Dry run is the default
//
// Every request carries the server dry-run option (DryRun: ["All"]) unless the caller's
// settings turn it off explicitly: the server validates and admits the object, runs the
// admission webhooks and returns what it WOULD have stored, and nothing is persisted. Real
// writes are the exception a caller must ask for.
//
// # Policy is not decided here
//
// Whether the script may write at all is the interpreter's capability policy
// (cli.CapClusterApply / cli.CapClusterRollout, checked by cli.GatedStep in the shim). This
// adapter is the mechanism; it never consults a policy and never reads a flag.

// DefaultFieldManager is the server-side apply field manager when the settings name none.
const DefaultFieldManager = "gatwd"

// WriteSettings are the knobs of a write, read when each object is written (so a flag parsed
// after the step was registered still applies). The zero value is a server dry run under
// DefaultFieldManager: it never persists anything.
type WriteSettings struct {
	// Persist turns the server dry run OFF: the write is real. False is the default.
	Persist bool
	// FieldManager owns the fields this write sets; "" means DefaultFieldManager.
	FieldManager string
	// ForceConflicts takes over fields another manager owns instead of failing. Off by default.
	ForceConflicts bool
	// Namespace is the namespace of a namespaced object that names none; "" means such an
	// object is an error (the manifest must say where it goes).
	Namespace string
}

func (s WriteSettings) manager() string {
	if s.FieldManager == "" {
		return DefaultFieldManager
	}
	return s.FieldManager
}

// dryRun is the DryRun option of a request: ["All"] unless the write is persisted.
func (s WriteSettings) dryRun() []string {
	if s.Persist {
		return nil
	}
	return []string{metav1.DryRunAll}
}

// WriteRecord is the result of writing ONE object, as a value the DSL can assert on: the
// stream carries one [Verb, ID, Outcome] list per object. ID is the qualified node id
// `Kind/namespace/name` (`_` for no namespace), the same id the graph prelude builds.
type WriteRecord struct {
	// Verb is the operation: apply, rollout-restart, rollout-undo, rollout-pause,
	// rollout-resume, scale.
	Verb string
	// ID is the qualified id of the object.
	ID string
	// Outcome says what happened (or, in a dry run, what would): created, configured,
	// unchanged, written, would-create, would-configure, ... see each operation.
	Outcome string
}

// ResolvedKind is where a kind lives: its resource and whether it is namespaced.
type ResolvedKind struct {
	GVR        schema.GroupVersionResource
	Namespaced bool
}

// ResourceResolver maps an object's apiVersion and kind to the resource that serves it.
type ResourceResolver interface {
	// Resolve returns the resource for kind in apiVersion (`apps/v1`, `v1`), or an error that
	// names both when the kind is unknown.
	Resolve(apiVersion, kind string) (ResolvedKind, error)
}

// ResolverFunc adapts a function to a ResourceResolver.
type ResolverFunc func(apiVersion, kind string) (ResolvedKind, error)

// Resolve implements ResourceResolver.
func (f ResolverFunc) Resolve(apiVersion, kind string) (ResolvedKind, error) {
	return f(apiVersion, kind)
}

// kindEntry is one row of the built-in kind table.
type kindEntry struct {
	apiVersion, kind, resource string
	namespaced                 bool
}

// builtinKinds are the kinds an apply can address without asking the API server; anything
// else (a CRD, a newer version) is resolved by discovery. Cluster-scoped kinds are the
// ones marked false.
var builtinKinds = []kindEntry{
	{"v1", "Namespace", "namespaces", false},
	{"v1", "Pod", "pods", true},
	{"v1", "Service", "services", true},
	{"v1", "ConfigMap", "configmaps", true},
	{"v1", "Secret", "secrets", true},
	{"v1", "Endpoints", "endpoints", true},
	{"v1", "PersistentVolumeClaim", "persistentvolumeclaims", true},
	{"v1", "PersistentVolume", "persistentvolumes", false},
	{"v1", "ServiceAccount", "serviceaccounts", true},
	{"v1", "ResourceQuota", "resourcequotas", true},
	{"v1", "LimitRange", "limitranges", true},
	{"apps/v1", "Deployment", "deployments", true},
	{"apps/v1", "ReplicaSet", "replicasets", true},
	{"apps/v1", "StatefulSet", "statefulsets", true},
	{"apps/v1", "DaemonSet", "daemonsets", true},
	{"batch/v1", "Job", "jobs", true},
	{"batch/v1", "CronJob", "cronjobs", true},
	{"networking.k8s.io/v1", "Ingress", "ingresses", true},
	{"networking.k8s.io/v1", "NetworkPolicy", "networkpolicies", true},
	{"rbac.authorization.k8s.io/v1", "Role", "roles", true},
	{"rbac.authorization.k8s.io/v1", "RoleBinding", "rolebindings", true},
	{"rbac.authorization.k8s.io/v1", "ClusterRole", "clusterroles", false},
	{"rbac.authorization.k8s.io/v1", "ClusterRoleBinding", "clusterrolebindings", false},
	{"autoscaling/v2", "HorizontalPodAutoscaler", "horizontalpodautoscalers", true},
	{"policy/v1", "PodDisruptionBudget", "poddisruptionbudgets", true},
	{"apiextensions.k8s.io/v1", "CustomResourceDefinition", "customresourcedefinitions", false},
}

// BuiltinKindResolver resolves the common built-in kinds from a table, without a network
// call. An unknown kind is an error that says so (and, in a chain, falls through to the next
// resolver).
func BuiltinKindResolver() ResourceResolver {
	return ResolverFunc(func(apiVersion, kind string) (ResolvedKind, error) {
		for _, e := range builtinKinds {
			if e.apiVersion == apiVersion && e.kind == kind {
				gv, err := schema.ParseGroupVersion(apiVersion)
				if err != nil {
					return ResolvedKind{}, err
				}
				return ResolvedKind{GVR: gv.WithResource(e.resource), Namespaced: e.namespaced}, nil
			}
		}
		return ResolvedKind{}, fmt.Errorf("kind %s in %s is not a built-in kind", kind, apiVersion)
	})
}

// DiscoveryResolver resolves any kind the API server serves (CRDs included) through the
// discovery client, which is obtained lazily by get on first use. A discovery failure is the
// error of the resolution, never "unknown kind" by silence.
func DiscoveryResolver(get func() (discovery.DiscoveryInterface, error)) ResourceResolver {
	return ResolverFunc(func(apiVersion, kind string) (ResolvedKind, error) {
		disco, err := get()
		if err != nil {
			return ResolvedKind{}, fmt.Errorf("cannot resolve kind %s in %s: %w", kind, apiVersion, err)
		}
		list, err := disco.ServerResourcesForGroupVersion(apiVersion)
		if err != nil {
			return ResolvedKind{}, fmt.Errorf("cannot resolve kind %s in %s: %w", kind, apiVersion, err)
		}
		gv, err := schema.ParseGroupVersion(apiVersion)
		if err != nil {
			return ResolvedKind{}, err
		}
		for _, r := range list.APIResources {
			if r.Kind == kind && !strings.Contains(r.Name, "/") {
				return ResolvedKind{GVR: gv.WithResource(r.Name), Namespaced: r.Namespaced}, nil
			}
		}
		return ResolvedKind{}, fmt.Errorf("the API server serves no kind %s in %s", kind, apiVersion)
	})
}

// ChainResolvers tries each resolver in order and returns the first success; when every one
// fails the error lists all the causes.
func ChainResolvers(rs ...ResourceResolver) ResourceResolver {
	return ResolverFunc(func(apiVersion, kind string) (ResolvedKind, error) {
		var errs []error
		for _, r := range rs {
			rk, err := r.Resolve(apiVersion, kind)
			if err == nil {
				return rk, nil
			}
			errs = append(errs, err)
		}
		return ResolvedKind{}, errors.Join(errs...)
	})
}

// writableResource is the narrow view of one resource the write code works through. It has
// no Delete, DeleteCollection, Create, Update or Watch: a write is a server-side apply or a
// patch, and what it may read is a Get and a List. A delete cannot be written against this
// interface, and kubernetes_write_guard_test.go scans the adapters for the calls anyway.
type writableResource interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error)
	List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error)
	Apply(ctx context.Context, name string, obj *unstructured.Unstructured, opts metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error)
}

// resourceFor narrows the dynamic client to one resource, in a namespace when the kind is
// namespaced.
func resourceFor(client dynamic.Interface, gvr schema.GroupVersionResource, namespace string) writableResource {
	if namespace == "" {
		return client.Resource(gvr)
	}
	return client.Resource(gvr).Namespace(namespace)
}

// Applier writes DSL objects to a cluster by server-side apply, and renders them to a
// directory (RenderTo) when there is no cluster.
type Applier struct {
	// Client is the cluster. nil means no cluster is configured: Apply fails loudly and only
	// RenderTo works.
	Client dynamic.Interface
	// Resolver maps an object's kind to its resource. Required when Client is set.
	Resolver ResourceResolver
	// Settings is read for every object; nil means the zero WriteSettings (a dry run).
	Settings func() WriteSettings
}

func (a *Applier) settings() WriteSettings {
	if a.Settings == nil {
		return WriteSettings{}
	}
	return a.Settings()
}

// ObjectFromPayload turns the value the DSL carries for an object — the nested
// `*collections.HAMT[string, any]` / `*collections.List[any]` structure `manifests` and the
// crawl steps produce (the inverse of unstructuredToValue), or plain Go maps and slices —
// into an unstructured object, and validates what every write needs: apiVersion, kind and
// metadata.name. A value JSON cannot hold, or a missing field, is an error that names it.
func ObjectFromPayload(v any) (*unstructured.Unstructured, error) {
	raw, err := valueToUnstructured(v, "$")
	if err != nil {
		return nil, err
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("cannot write %T as a Kubernetes object: want a mapping with apiVersion, kind and metadata", v)
	}
	obj := &unstructured.Unstructured{Object: m}
	if obj.GetAPIVersion() == "" {
		return nil, fmt.Errorf("cannot write an object with no apiVersion (kind %q, name %q)", obj.GetKind(), obj.GetName())
	}
	if obj.GetKind() == "" {
		return nil, fmt.Errorf("cannot write an object with no kind (apiVersion %s, name %q)", obj.GetAPIVersion(), obj.GetName())
	}
	if obj.GetName() == "" {
		return nil, fmt.Errorf("cannot write a %s with no metadata.name", obj.GetKind())
	}
	for _, part := range []string{obj.GetKind(), obj.GetName(), obj.GetNamespace()} {
		if strings.ContainsAny(part, `/\`) || part == ".." {
			return nil, fmt.Errorf("cannot write %s/%s: %q is not a valid name part", obj.GetKind(), obj.GetName(), part)
		}
	}
	return obj, nil
}

// valueToUnstructured lifts a DSL value back to the generic JSON shape client-go encodes.
func valueToUnstructured(v any, path string) (any, error) {
	switch t := v.(type) {
	case nil, string, bool, int64, float64:
		return t, nil
	case int:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case uint32:
		return int64(t), nil
	case float32:
		return float64(t), nil
	case *collections.HAMT[string, any]:
		out := make(map[string]any, t.Len())
		for k, val := range t.Seq() {
			x, err := valueToUnstructured(val, path+"."+k)
			if err != nil {
				return nil, err
			}
			out[k] = x
		}
		return out, nil
	case *collections.List[any]:
		out := make([]any, 0, t.Len())
		i := 0
		for val := range t.Seq() {
			x, err := valueToUnstructured(val, fmt.Sprintf("%s[%d]", path, i))
			if err != nil {
				return nil, err
			}
			out = append(out, x)
			i++
		}
		return out, nil
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			x, err := valueToUnstructured(val, path+"."+k)
			if err != nil {
				return nil, err
			}
			out[k] = x
		}
		return out, nil
	case []any:
		out := make([]any, 0, len(t))
		for i, val := range t {
			x, err := valueToUnstructured(val, fmt.Sprintf("%s[%d]", path, i))
			if err != nil {
				return nil, err
			}
			out = append(out, x)
		}
		return out, nil
	}
	return nil, fmt.Errorf("cannot write %T at %s as part of a Kubernetes object", v, path)
}

// ObjectID is the qualified node id `Kind/namespace/name` of obj, with `_` for no namespace:
// the id scheme of the graph prelude (pkg/k8sgraph/lib/k8s/ids.gatwd), so a write record and
// a graph node name one object the same way.
func ObjectID(obj *unstructured.Unstructured) string {
	ns := obj.GetNamespace()
	if ns == "" {
		ns = "_"
	}
	return obj.GetKind() + "/" + ns + "/" + obj.GetName()
}

// Apply server-side applies one object (the DSL payload) and reports what happened as
// [apply, id, outcome]. Outcomes: created, configured, unchanged; in a dry run (the default)
// would-create, would-configure, unchanged. An error — a payload that is no object, a kind
// the cluster does not serve, a conflict with another field manager, an API refusal — is
// returned, never swallowed; nothing is applied for that object.
func (a *Applier) Apply(ctx context.Context, payload any) (WriteRecord, error) {
	obj, err := ObjectFromPayload(payload)
	if err != nil {
		return WriteRecord{}, err
	}
	if a.Client == nil {
		return WriteRecord{}, fmt.Errorf("cannot apply %s: no cluster is configured (give --out DIR to render the objects to files instead)", ObjectID(obj))
	}
	if a.Resolver == nil {
		return WriteRecord{}, errors.New("cannot apply: the write adapter has no kind resolver")
	}
	st := a.settings()
	rk, err := a.Resolver.Resolve(obj.GetAPIVersion(), obj.GetKind())
	if err != nil {
		return WriteRecord{}, fmt.Errorf("cannot apply %s: %w", ObjectID(obj), err)
	}
	if err := placeInNamespace(obj, rk, st); err != nil {
		return WriteRecord{}, err
	}
	id := ObjectID(obj)
	// Server-side apply rejects server-owned bookkeeping in the body.
	unstructured.RemoveNestedField(obj.Object, "metadata", "managedFields")
	unstructured.RemoveNestedField(obj.Object, "metadata", "resourceVersion")

	res := resourceFor(a.Client, rk.GVR, obj.GetNamespace())
	before, err := res.Get(ctx, obj.GetName(), metav1.GetOptions{})
	exists := err == nil
	if err != nil && !apierrors.IsNotFound(err) {
		return WriteRecord{}, fmt.Errorf("cannot apply %s: reading the current object: %w", id, err)
	}
	after, err := res.Apply(ctx, obj.GetName(), obj, metav1.ApplyOptions{
		FieldManager: st.manager(),
		Force:        st.ForceConflicts,
		DryRun:       st.dryRun(),
	})
	if err != nil {
		return WriteRecord{}, fmt.Errorf("cannot apply %s: %w", id, err)
	}
	return WriteRecord{Verb: "apply", ID: id, Outcome: applyOutcome(exists, exists && sameContent(before, after), st.Persist)}, nil
}

// applyOutcome names what an apply did: created, configured or unchanged; in a dry run the
// two changes read would-create and would-configure, and a no-op is still unchanged.
func applyOutcome(existed, same, persisted bool) string {
	switch {
	case existed && same:
		return "unchanged"
	case existed && persisted:
		return "configured"
	case existed:
		return "would-configure"
	case persisted:
		return "created"
	}
	return "would-create"
}

// placeInNamespace settles the namespace of obj against its kind: a namespaced kind needs
// one (from the manifest or the settings), a cluster-scoped kind must not name one.
func placeInNamespace(obj *unstructured.Unstructured, rk ResolvedKind, st WriteSettings) error {
	switch {
	case rk.Namespaced && obj.GetNamespace() == "":
		if st.Namespace == "" {
			return fmt.Errorf("cannot apply %s: %s is namespaced and the manifest names no metadata.namespace (set it, or give a default namespace)", ObjectID(obj), obj.GetKind())
		}
		obj.SetNamespace(st.Namespace)
	case !rk.Namespaced && obj.GetNamespace() != "":
		return fmt.Errorf("cannot apply %s: %s is cluster-scoped and must not set metadata.namespace", ObjectID(obj), obj.GetKind())
	}
	return nil
}

// sameContent reports whether two versions of an object differ only in server bookkeeping.
func sameContent(a, b *unstructured.Unstructured) bool {
	strip := func(o *unstructured.Unstructured) map[string]any {
		c := o.DeepCopy()
		for _, f := range []string{"managedFields", "resourceVersion", "generation", "uid", "creationTimestamp"} {
			unstructured.RemoveNestedField(c.Object, "metadata", f)
		}
		return c.Object
	}
	return reflect.DeepEqual(strip(a), strip(b))
}

// RenderTo is the offline twin of Apply: it writes the object to dir as YAML, one file per
// object named `Kind.namespace.name.yaml` (`_` for no namespace), and reports
// [apply, id, written] — or [apply, id, unchanged] when the file already holds exactly that
// content, so a second render is as idempotent as a second apply. It needs no cluster and no
// resolver, and applies the same validation as Apply (a payload that is no object, or has no
// kind or name, is an error); the namespace default of the settings is NOT applied, the file
// holds the object as the manifest wrote it. Whether the caller may write files at all is
// the caller's policy (the shim requires the fs capability).
func (a *Applier) RenderTo(dir string, payload any) (WriteRecord, error) {
	obj, err := ObjectFromPayload(payload)
	if err != nil {
		return WriteRecord{}, err
	}
	id := ObjectID(obj)
	unstructured.RemoveNestedField(obj.Object, "metadata", "managedFields")
	body, err := yaml.Marshal(obj.Object)
	if err != nil {
		return WriteRecord{}, fmt.Errorf("cannot render %s: %w", id, err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return WriteRecord{}, fmt.Errorf("cannot render %s: %w", id, err)
	}
	path := filepath.Join(dir, strings.ReplaceAll(id, "/", ".")+".yaml")
	if old, err := os.ReadFile(path); err == nil && bytes.Equal(old, body) {
		return WriteRecord{Verb: "apply", ID: id, Outcome: "unchanged"}, nil
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return WriteRecord{}, fmt.Errorf("cannot render %s: %w", id, err)
	}
	return WriteRecord{Verb: "apply", ID: id, Outcome: "written"}, nil
}
