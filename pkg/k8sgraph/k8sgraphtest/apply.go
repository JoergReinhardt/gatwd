package k8sgraphtest

import (
	"encoding/json"
	"slices"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"
)

// EmulateWrites teaches a fake dynamic client the two things the write adapter relies on and
// client-go's fake lacks:
//
//   - server-side apply (a patch of type ApplyPatchType) CREATES an object that does not exist
//     (the stock fake fails with NotFound) and otherwise merges the applied fields into the
//     stored object, maps recursively and everything else replaced;
//   - a request with the server dry-run option (DryRun: ["All"]) PERSISTS NOTHING, whatever
//     its patch type: an apply returns the object it would have stored, any other patch returns
//     the stored object unchanged.
//
// Every request is still recorded by the fake (client.Actions()), with its options, so a test
// asserts on exactly what the adapter sent. Install it once, right after building the client.
func EmulateWrites(c *fake.FakeDynamicClient) {
	c.PrependReactor("patch", "*", func(action clienttesting.Action) (bool, runtime.Object, error) {
		patch, ok := action.(clienttesting.PatchActionImpl)
		if !ok {
			return false, nil, nil
		}
		dry := slices.Contains(patch.PatchOptions.DryRun, metav1.DryRunAll)
		gvr, ns, name := patch.GetResource(), patch.GetNamespace(), patch.GetName()
		existing, getErr := c.Tracker().Get(gvr, ns, name)
		if patch.GetPatchType() != types.ApplyPatchType {
			if dry && getErr == nil {
				return true, existing, nil // dry run: the stored object, untouched
			}
			return false, nil, nil // a real patch: the stock reaction
		}
		applied := &unstructured.Unstructured{Object: map[string]any{}}
		if err := json.Unmarshal(patch.GetPatch(), &applied.Object); err != nil {
			return true, nil, err
		}
		applied.SetName(name)
		if ns != "" {
			applied.SetNamespace(ns)
		}
		switch {
		case apierrors.IsNotFound(getErr):
			if !dry {
				if err := c.Tracker().Create(gvr, applied, ns); err != nil {
					return true, nil, err
				}
			}
			return true, applied, nil
		case getErr != nil:
			return true, nil, getErr
		}
		cur, ok := existing.(*unstructured.Unstructured)
		if !ok {
			return true, nil, apierrors.NewBadRequest("the fake holds a non-unstructured object")
		}
		merged := &unstructured.Unstructured{Object: mergeMaps(cur.DeepCopy().Object, applied.Object)}
		if !dry {
			if err := c.Tracker().Update(gvr, merged, ns); err != nil {
				return true, nil, err
			}
		}
		return true, merged, nil
	})
}

// mergeMaps overlays src on dst: nested maps merge, any other value (lists included) is replaced.
func mergeMaps(dst, src map[string]any) map[string]any {
	for k, v := range src {
		sm, sok := v.(map[string]any)
		dm, dok := dst[k].(map[string]any)
		if sok && dok {
			dst[k] = mergeMaps(dm, sm)
			continue
		}
		dst[k] = v
	}
	return dst
}
