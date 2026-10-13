package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// This file is the ROLLOUT family of the write adapter: the `kubectl rollout` and `kubectl
// scale` equivalents on the dynamic client. Everything here is a patch of an existing
// workload (never a create, never a delete); Status only reads. Like Apply, every request
// carries the server dry-run option unless the settings persist the write.
//
//	Status   rollout status   read-only: Get
//	Restart  rollout restart  merge patch: pod template annotation kubectl.kubernetes.io/restartedAt
//	Undo     rollout undo     JSON patch: spec.template := the template of an earlier ReplicaSet
//	Pause    rollout pause    merge patch: spec.paused = true
//	Resume   rollout resume   merge patch: spec.paused = false
//	Scale    scale            merge patch: spec.replicas
//
// Scale patches spec.replicas of the workload itself rather than its `scale` subresource, so
// the RBAC it needs is `patch` on the workload kind and nothing more.

// RestartedAtAnnotation is the pod-template annotation `kubectl rollout restart` sets; a new
// value changes the template, which makes the controller roll the pods.
const RestartedAtAnnotation = "kubectl.kubernetes.io/restartedAt"

const revisionAnnotation = "deployment.kubernetes.io/revision"

// workload is where a rollout kind lives and which operations it supports.
type workload struct {
	gvr                                 schema.GroupVersionResource
	restart, pause, undo, status, scale bool
}

var workloads = map[string]workload{
	"Deployment":  {schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, true, true, true, true, true},
	"StatefulSet": {schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}, true, false, false, false, true},
	"DaemonSet":   {schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}, true, false, false, false, false},
	"ReplicaSet":  {schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "replicasets"}, false, false, false, false, true},
}

// target is a workload named by the DSL: an object payload or the qualified id Kind/ns/name.
type target struct {
	kind, namespace, name string
	w                     workload
}

func (t target) id() string { return t.kind + "/" + t.namespace + "/" + t.name }

// parseTarget reads the DSL value naming a workload: an object payload (what `manifests` and
// the crawl produce) or a qualified id string `Deployment/default/web`. op is the operation,
// for the error message; ok says whether the kind supports it.
func parseTarget(v any, op string, ok func(workload) bool) (target, error) {
	var kind, ns, name string
	if s, isStr := v.(string); isStr {
		parts := strings.Split(s, "/")
		if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[1] == "_" || parts[2] == "" {
			return target{}, fmt.Errorf("%s: %q is not a workload id; want Kind/namespace/name, e.g. Deployment/default/web", op, s)
		}
		kind, ns, name = parts[0], parts[1], parts[2]
	} else {
		obj, err := ObjectFromPayload(v)
		if err != nil {
			return target{}, fmt.Errorf("%s: %w", op, err)
		}
		if obj.GetNamespace() == "" {
			return target{}, fmt.Errorf("%s: %s names no metadata.namespace", op, ObjectID(obj))
		}
		kind, ns, name = obj.GetKind(), obj.GetNamespace(), obj.GetName()
	}
	w, known := workloads[kind]
	if !known {
		return target{}, fmt.Errorf("%s: %s/%s/%s is not a workload kind (want Deployment, StatefulSet, DaemonSet or ReplicaSet)", op, kind, ns, name)
	}
	if !ok(w) {
		return target{}, fmt.Errorf("%s does not support %s (%s/%s/%s)", op, kind, kind, ns, name)
	}
	return target{kind: kind, namespace: ns, name: name, w: w}, nil
}

func (a *Applier) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

// patchOne sends one patch of the given type to the target, under the settings' dry-run
// option, and says whether the server accepted it.
func (a *Applier) patchOne(ctx context.Context, t target, pt types.PatchType, body any) error {
	if a.Client == nil {
		return fmt.Errorf("cannot change %s: no cluster is configured", t.id())
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	st := a.settings()
	_, err = resourceFor(a.Client, t.w.gvr, t.namespace).Patch(ctx, t.name, pt, raw,
		metav1.PatchOptions{FieldManager: st.manager(), DryRun: st.dryRun()})
	return err
}

func (a *Applier) get(ctx context.Context, t target, op string) (*unstructured.Unstructured, error) {
	if a.Client == nil {
		return nil, fmt.Errorf("%s %s: no cluster is configured", op, t.id())
	}
	obj, err := resourceFor(a.Client, t.w.gvr, t.namespace).Get(ctx, t.name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", op, t.id(), err)
	}
	return obj, nil
}

// outcome picks the spelling of a result: real when the write is persisted, would in a dry run.
func (a *Applier) outcome(real, would string) string {
	if a.settings().Persist {
		return real
	}
	return would
}

// Restart annotates the pod template with the current time (kubectl.kubernetes.io/restartedAt),
// which makes the controller replace the pods. Outcome restarted, or would-restart in a dry run.
func (a *Applier) Restart(ctx context.Context, v any) (WriteRecord, error) {
	t, err := parseTarget(v, "rollout restart", func(w workload) bool { return w.restart })
	if err != nil {
		return WriteRecord{}, err
	}
	body := map[string]any{"spec": map[string]any{"template": map[string]any{"metadata": map[string]any{
		"annotations": map[string]any{RestartedAtAnnotation: a.now().UTC().Format(time.RFC3339)}}}}}
	if err := a.patchOne(ctx, t, types.MergePatchType, body); err != nil {
		return WriteRecord{}, fmt.Errorf("rollout restart %s: %w", t.id(), err)
	}
	return WriteRecord{Verb: "rollout-restart", ID: t.id(), Outcome: a.outcome("restarted", "would-restart")}, nil
}

// Pause sets spec.paused on a Deployment: outcome paused / would-pause, or unchanged when it
// already is paused (no request is sent).
func (a *Applier) Pause(ctx context.Context, v any) (WriteRecord, error) {
	return a.setPaused(ctx, v, true)
}

// Resume clears spec.paused on a Deployment: outcome resumed / would-resume, or unchanged
// when it is not paused.
func (a *Applier) Resume(ctx context.Context, v any) (WriteRecord, error) {
	return a.setPaused(ctx, v, false)
}

func (a *Applier) setPaused(ctx context.Context, v any, pause bool) (WriteRecord, error) {
	op, verb, done, would := "rollout resume", "rollout-resume", "resumed", "would-resume"
	if pause {
		op, verb, done, would = "rollout pause", "rollout-pause", "paused", "would-pause"
	}
	t, err := parseTarget(v, op, func(w workload) bool { return w.pause })
	if err != nil {
		return WriteRecord{}, err
	}
	cur, err := a.get(ctx, t, op)
	if err != nil {
		return WriteRecord{}, err
	}
	if paused, _, _ := unstructured.NestedBool(cur.Object, "spec", "paused"); paused == pause {
		return WriteRecord{Verb: verb, ID: t.id(), Outcome: "unchanged"}, nil
	}
	if err := a.patchOne(ctx, t, types.MergePatchType, map[string]any{"spec": map[string]any{"paused": pause}}); err != nil {
		return WriteRecord{}, fmt.Errorf("%s %s: %w", op, t.id(), err)
	}
	return WriteRecord{Verb: verb, ID: t.id(), Outcome: a.outcome(done, would)}, nil
}

// Scale sets spec.replicas: outcome scaled:N / would-scale:N, or unchanged when the workload
// already has N replicas. A negative N is an error.
func (a *Applier) Scale(ctx context.Context, v any, replicas int64) (WriteRecord, error) {
	t, err := parseTarget(v, "scale", func(w workload) bool { return w.scale })
	if err != nil {
		return WriteRecord{}, err
	}
	if replicas < 0 {
		return WriteRecord{}, fmt.Errorf("scale %s: replicas must not be negative, got %d", t.id(), replicas)
	}
	cur, err := a.get(ctx, t, "scale")
	if err != nil {
		return WriteRecord{}, err
	}
	if n, found, _ := nestedInt(cur.Object, "spec", "replicas"); found && n == replicas {
		return WriteRecord{Verb: "scale", ID: t.id(), Outcome: "unchanged"}, nil
	}
	if err := a.patchOne(ctx, t, types.MergePatchType, map[string]any{"spec": map[string]any{"replicas": replicas}}); err != nil {
		return WriteRecord{}, fmt.Errorf("scale %s: %w", t.id(), err)
	}
	n := strconv.FormatInt(replicas, 10)
	return WriteRecord{Verb: "scale", ID: t.id(), Outcome: a.outcome("scaled:"+n, "would-scale:"+n)}, nil
}

// Undo rolls a Deployment back to an earlier revision from its ReplicaSet history: revision 0
// means the one before the current, any other value that revision. It replaces spec.template
// with that ReplicaSet's template (less the controller's pod-template-hash label), as `kubectl
// rollout undo` does. Outcome rolled-back:REV / would-roll-back:REV, or unchanged when the
// Deployment already has that template. A paused Deployment, a missing revision and an
// absent history are errors.
func (a *Applier) Undo(ctx context.Context, v any, revision int64) (WriteRecord, error) {
	t, err := parseTarget(v, "rollout undo", func(w workload) bool { return w.undo })
	if err != nil {
		return WriteRecord{}, err
	}
	dep, err := a.get(ctx, t, "rollout undo")
	if err != nil {
		return WriteRecord{}, err
	}
	if paused, _, _ := unstructured.NestedBool(dep.Object, "spec", "paused"); paused {
		return WriteRecord{}, fmt.Errorf("rollout undo %s: the deployment is paused; resume it first", t.id())
	}
	history, err := a.history(ctx, t, dep)
	if err != nil {
		return WriteRecord{}, err
	}
	current := annotationInt(dep, revisionAnnotation)
	var chosen *revisionRS
	switch {
	case revision < 0:
		return WriteRecord{}, fmt.Errorf("rollout undo %s: revision must not be negative, got %d", t.id(), revision)
	case revision == 0:
		for i := range history { // newest first
			if history[i].revision < current {
				chosen = &history[i]
				break
			}
		}
		if chosen == nil {
			return WriteRecord{}, fmt.Errorf("rollout undo %s: no previous revision (current is %d, history %v)", t.id(), current, revisionsOf(history))
		}
	default:
		for i := range history {
			if history[i].revision == revision {
				chosen = &history[i]
			}
		}
		if chosen == nil {
			return WriteRecord{}, fmt.Errorf("rollout undo %s: revision %d not found (history %v)", t.id(), revision, revisionsOf(history))
		}
	}
	tmpl, found, err := unstructured.NestedMap(chosen.obj.Object, "spec", "template")
	if err != nil || !found {
		return WriteRecord{}, fmt.Errorf("rollout undo %s: revision %d has no pod template", t.id(), chosen.revision)
	}
	if labels, ok, _ := unstructured.NestedMap(tmpl, "metadata", "labels"); ok {
		delete(labels, "pod-template-hash")
		_ = unstructured.SetNestedMap(tmpl, labels, "metadata", "labels")
	}
	rev := strconv.FormatInt(chosen.revision, 10)
	if curTmpl, _, _ := unstructured.NestedMap(dep.Object, "spec", "template"); sameJSON(curTmpl, tmpl) {
		return WriteRecord{Verb: "rollout-undo", ID: t.id(), Outcome: "unchanged"}, nil
	}
	patch := []map[string]any{{"op": "replace", "path": "/spec/template", "value": tmpl}}
	if err := a.patchOne(ctx, t, types.JSONPatchType, patch); err != nil {
		return WriteRecord{}, fmt.Errorf("rollout undo %s: %w", t.id(), err)
	}
	return WriteRecord{Verb: "rollout-undo", ID: t.id(), Outcome: a.outcome("rolled-back:"+rev, "would-roll-back:"+rev)}, nil
}

type revisionRS struct {
	revision int64
	obj      *unstructured.Unstructured
}

func revisionsOf(h []revisionRS) []int64 {
	out := make([]int64, len(h))
	for i, r := range h {
		out[i] = r.revision
	}
	return out
}

// history lists the ReplicaSets the Deployment owns, newest revision first.
func (a *Applier) history(ctx context.Context, t target, dep *unstructured.Unstructured) ([]revisionRS, error) {
	rsGVR := workloads["ReplicaSet"].gvr
	opts := metav1.ListOptions{}
	if sel, found, _ := unstructured.NestedStringMap(dep.Object, "spec", "selector", "matchLabels"); found && len(sel) > 0 {
		opts.LabelSelector = labelSelector(sel)
	}
	list, err := resourceFor(a.Client, rsGVR, t.namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("rollout undo %s: listing its ReplicaSets: %w", t.id(), err)
	}
	var out []revisionRS
	for i := range list.Items {
		rs := &list.Items[i]
		if !ownedBy(rs, "Deployment", t.name, string(dep.GetUID())) {
			continue
		}
		if rev := annotationInt(rs, revisionAnnotation); rev > 0 {
			out = append(out, revisionRS{revision: rev, obj: rs})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].revision > out[j].revision })
	return out, nil
}

func labelSelector(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + "=" + m[k]
	}
	return strings.Join(parts, ",")
}

func ownedBy(obj *unstructured.Unstructured, kind, name, uid string) bool {
	for _, o := range obj.GetOwnerReferences() {
		if o.Kind == kind && o.Name == name && (uid == "" || o.UID == "" || string(o.UID) == uid) {
			return true
		}
	}
	return false
}

func annotationInt(obj *unstructured.Unstructured, key string) int64 {
	n, _ := strconv.ParseInt(obj.GetAnnotations()[key], 10, 64)
	return n
}

// nestedInt reads an integer field however the object was decoded: the dynamic client yields
// int64, a JSON decoder float64.
func nestedInt(obj map[string]any, fields ...string) (int64, bool, error) {
	v, found, err := unstructured.NestedFieldNoCopy(obj, fields...)
	if err != nil || !found {
		return 0, false, err
	}
	switch n := v.(type) {
	case int64:
		return n, true, nil
	case int:
		return int64(n), true, nil
	case float64:
		return int64(n), true, nil
	}
	return 0, false, fmt.Errorf("%s is %T, not a number", strings.Join(fields, "."), v)
}

func sameJSON(a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return string(x) == string(y)
}

// Status reports how a Deployment's rollout is going, like `kubectl rollout status` without
// the wait: [rollout-status, id, complete] or [rollout-status, id, progressing: …] (also
// "paused" and "waiting: …" for a spec the controller has not seen). It only reads the
// Deployment, so it needs only the `get` permission and no write capability. An unreadable or
// unknown workload is an error; an unfinished rollout is a result, not an error.
func (a *Applier) Status(ctx context.Context, v any) (WriteRecord, error) {
	t, err := parseTarget(v, "rollout status", func(w workload) bool { return w.status })
	if err != nil {
		return WriteRecord{}, err
	}
	dep, err := a.get(ctx, t, "rollout status")
	if err != nil {
		return WriteRecord{}, err
	}
	return WriteRecord{Verb: "rollout-status", ID: t.id(), Outcome: deploymentStatus(dep)}, nil
}

func deploymentStatus(d *unstructured.Unstructured) string {
	gen, _, _ := nestedInt(d.Object, "metadata", "generation")
	observed, _, _ := nestedInt(d.Object, "status", "observedGeneration")
	if gen > observed {
		return "waiting: the deployment spec has not been observed yet"
	}
	if paused, _, _ := unstructured.NestedBool(d.Object, "spec", "paused"); paused {
		return "paused"
	}
	want, found, _ := nestedInt(d.Object, "spec", "replicas")
	if !found {
		want = 1
	}
	updated, _, _ := nestedInt(d.Object, "status", "updatedReplicas")
	total, _, _ := nestedInt(d.Object, "status", "replicas")
	available, _, _ := nestedInt(d.Object, "status", "availableReplicas")
	switch {
	case updated < want:
		return fmt.Sprintf("progressing: %d out of %d new replicas have been updated", updated, want)
	case total > updated:
		return fmt.Sprintf("progressing: %d old replicas are pending termination", total-updated)
	case available < updated:
		return fmt.Sprintf("progressing: %d of %d updated replicas are available", available, updated)
	}
	return "complete"
}
