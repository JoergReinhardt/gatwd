package adapters

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func obj(m map[string]any) *unstructured.Unstructured { return &unstructured.Unstructured{Object: m} }

// Every branch of the workload rollout state machines, one case each, so that disabling any
// single condition makes a case fail.
func TestRolloutStateBranches(t *testing.T) {
	meta := map[string]any{"generation": int64(2)}
	for _, c := range []struct {
		name, kind string
		o          map[string]any
		want       string
	}{
		{"sts unobserved", "StatefulSet", map[string]any{"metadata": meta, "status": map[string]any{"observedGeneration": int64(1)}}, "not been observed"},
		{"sts not ready", "StatefulSet", map[string]any{"metadata": meta, "spec": map[string]any{"replicas": int64(3)}, "status": map[string]any{"observedGeneration": int64(2), "readyReplicas": int64(1)}}, "1 of 3 replicas are ready"},
		{"sts not updated", "StatefulSet", map[string]any{"metadata": meta, "spec": map[string]any{"replicas": int64(3)}, "status": map[string]any{"observedGeneration": int64(2), "readyReplicas": int64(3), "updatedReplicas": int64(2), "currentRevision": "a", "updateRevision": "a"}}, "2 of 3 replicas have been updated"},
		{"sts revision differs", "StatefulSet", map[string]any{"metadata": meta, "spec": map[string]any{"replicas": int64(1)}, "status": map[string]any{"observedGeneration": int64(2), "readyReplicas": int64(1), "updatedReplicas": int64(1), "currentRevision": "a", "updateRevision": "b"}}, "have been updated"},
		{"sts complete", "StatefulSet", map[string]any{"metadata": meta, "spec": map[string]any{"replicas": int64(1)}, "status": map[string]any{"observedGeneration": int64(2), "readyReplicas": int64(1), "updatedReplicas": int64(1), "currentRevision": "a", "updateRevision": "a"}}, "complete"},
		{"ds unobserved", "DaemonSet", map[string]any{"metadata": meta, "status": map[string]any{"observedGeneration": int64(1)}}, "not been observed"},
		{"ds not updated", "DaemonSet", map[string]any{"metadata": meta, "status": map[string]any{"observedGeneration": int64(2), "desiredNumberScheduled": int64(3), "updatedNumberScheduled": int64(1), "numberAvailable": int64(3)}}, "1 out of 3 new pods"},
		{"ds not available", "DaemonSet", map[string]any{"metadata": meta, "status": map[string]any{"observedGeneration": int64(2), "desiredNumberScheduled": int64(3), "updatedNumberScheduled": int64(3), "numberAvailable": int64(2)}}, "2 of 3 updated pods are available"},
		{"ds complete", "DaemonSet", map[string]any{"metadata": meta, "status": map[string]any{"observedGeneration": int64(2), "desiredNumberScheduled": int64(3), "updatedNumberScheduled": int64(3), "numberAvailable": int64(3)}}, "complete"},
	} {
		if got := rolloutState(c.kind, obj(c.o)); !strings.Contains(got, c.want) {
			t.Errorf("%s: state %q lacks %q", c.name, got, c.want)
		}
	}
}
