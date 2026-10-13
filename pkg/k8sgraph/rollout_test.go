package k8sgraph_test

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph/k8sgraphtest"
)

const rolloutFixture = "testdata/write/rollout.yaml"

// newRolloutRig is a trusted script on the rollout fixture cluster; grant adds the
// cluster-rollout opt-in, persist is --dry-run=false.
func newRolloutRig(t *testing.T, grant, persist bool) *writeRig {
	t.Helper()
	raw, err := os.ReadFile(rolloutFixture)
	if err != nil {
		t.Fatal(err)
	}
	client := k8sgraphtest.FakeClusterFromYAML(t, raw)
	k8sgraphtest.EmulateWrites(client)
	flags := k8sgraph.DefaultWriteFlags()
	flags.DryRun = !persist
	interp := k8sgraph.ApplyEdition(cli.NewEdition("rollout-test"), client, k8sgraph.WithWriteFlags(flags)).BuildInterpreter()
	if grant {
		interp.SetPolicy(interp.Policy().Grant(cli.CapClusterRollout))
	}
	return &writeRig{client, interp, flags}
}

func on(id, step string) string { return `spawn StrVal: ` + id + ` |> ` + step }

func (r *writeRig) deployment(t *testing.T, name string) *unstructured.Unstructured {
	t.Helper()
	obj, err := r.client.Tracker().Get(deploymentsGVR, "default", name)
	if err != nil {
		t.Fatal(err)
	}
	return obj.(*unstructured.Unstructured)
}

func (r *writeRig) image(t *testing.T, name string) string {
	t.Helper()
	cs, _, _ := unstructured.NestedSlice(r.deployment(t, name).Object, "spec", "template", "spec", "containers")
	return cs[0].(map[string]any)["image"].(string)
}

// onlyReadsAndPatches asserts the cluster saw nothing but gets, lists and patches: never a
// create, update or delete.
func (r *writeRig) onlyReadsAndPatches(t *testing.T) {
	t.Helper()
	for _, v := range r.verbs() {
		if !slices.Contains([]string{"get", "list", "patch"}, v) {
			t.Fatalf("the rollout family sent a %q request; it may only get, list and patch (verbs: %v)", v, r.verbs())
		}
	}
}

// ---- status: read-only, no capability ----

func TestRolloutStatusReadsOnlyAndNeedsNoCapability(t *testing.T) {
	r := newRolloutRig(t, false, true) // trusted script, NO rollout capability, real-write setting
	for id, want := range map[string]string{
		"Deployment/default/web":    "rollout-status Deployment/default/web complete",
		"Deployment/default/slow":   "rollout-status Deployment/default/slow progressing: 1 out of 3 new replicas have been updated",
		"Deployment/default/halted": "rollout-status Deployment/default/halted paused",
	} {
		got := r.records(t, on(id, "k8s_rollout_status"))
		wantRecords(t, got, []string{want})
	}
	for _, v := range r.verbs() {
		if v != "get" {
			t.Fatalf("rollout status sent a %q request: it must only read (%v)", v, r.verbs())
		}
	}
	// through a payload from the cluster, not just an id
	got := r.records(t, `spawn StrVal: default |> k8s_deployments |> hamt_get metadata |> hamt_get name`)
	if len(got) != 3 {
		t.Fatalf("fixture crawl: %q", got)
	}
	// a kind with no status, a missing object, a malformed id: errors, never "complete"
	for line, want := range map[string]string{
		on("StatefulSet/default/db", "k8s_rollout_status"):   "does not support StatefulSet",
		on("Deployment/default/ghost", "k8s_rollout_status"): "not found",
		on("web", "k8s_rollout_status"):                      "is not a workload id",
		on("Service/default/web", "k8s_rollout_status"):      "is not a workload kind",
		on("Deployment/default/web", "k8s_rollout_status 3"): "takes no argument",
	} {
		g := r.records(t, line)
		if len(g) != 1 || !strings.HasPrefix(g[0], "[ERROR]") || !strings.Contains(g[0], want) {
			t.Errorf("%s = %q, want an error containing %q", line, g, want)
		}
	}
}

// ---- layer 1: every mutating step is denied for the trusted script ----

func TestRolloutWritesAreDeniedForTheTrustedScript(t *testing.T) {
	r := newRolloutRig(t, false, true)
	if r.interp.Policy().Allows(cli.CapClusterRollout) {
		t.Fatal("test premise broken")
	}
	for _, step := range []string{"k8s_rollout_restart", "k8s_rollout_pause", "k8s_rollout_resume", "k8s_rollout_undo", "k8s_rollout_undo 1", "k8s_scale 5"} {
		name, _, _ := strings.Cut(step, " ")
		got := r.records(t, on("Deployment/default/web", step))
		want := "[ERROR] step '" + name + "' requires the cluster-rollout capability (policy denies it)"
		if len(got) != 1 || got[0] != want {
			t.Errorf("%s = %q, want [%q]", step, got, want)
		}
	}
	if acts := r.client.Actions(); len(acts) != 0 {
		t.Fatalf("a denied rollout step reached the cluster: %v", r.verbs())
	}
	// the apply capability is another switch
	r.interp.SetPolicy(r.interp.Policy().Grant(cli.CapClusterApply))
	if got := r.records(t, on("Deployment/default/web", "k8s_scale 5")); !strings.Contains(got[0], "cluster-rollout capability") {
		t.Fatalf("cluster-apply must not grant the rollout family: %q", got)
	}
}

func TestRolloutEditionFlagIsTheOptIn(t *testing.T) {
	raw, _ := os.ReadFile(rolloutFixture)
	client := k8sgraphtest.FakeClusterFromYAML(t, raw)
	k8sgraphtest.EmulateWrites(client)
	ed := k8sgraph.ApplyEdition(cli.NewEdition("rollout-flag-test"), client)
	run := func(argv ...string) string {
		var out, errb strings.Builder
		ed.Command().Execute(argv, strings.NewReader(""), &out, &errb)
		return out.String() + errb.String()
	}
	line := on("Deployment/default/web", "k8s_scale 7")
	if out := run("eval", "-e", line); !strings.Contains(out, "requires the cluster-rollout capability") {
		t.Fatalf("without the flag: %q", out)
	}
	if out := run("--allow-cluster-apply", "eval", "-e", line); !strings.Contains(out, "requires the cluster-rollout capability") {
		t.Fatalf("--allow-cluster-apply must not grant rollout: %q", out)
	}
	if out := run("--allow-cluster-rollout", "eval", "-e", line); !strings.Contains(out, "would-scale:7") {
		t.Fatalf("--allow-cluster-rollout is a dry run by default: %q", out)
	}
	if out := run("--allow-cluster-rollout", "--dry-run=false", "eval", "-e", line); !strings.Contains(out, "scaled:7") || strings.Contains(out, "would-") {
		t.Fatalf("--dry-run=false: %q", out)
	}
	obj, _ := client.Tracker().Get(deploymentsGVR, "default", "web")
	if n, _, _ := unstructured.NestedInt64(obj.(*unstructured.Unstructured).Object, "spec", "replicas"); n != 7 {
		t.Fatalf("replicas = %d after the real scale", n)
	}
}

// ---- layer 2: dry run is the default ----

func TestRolloutDryRunIsTheDefaultAndChangesNothing(t *testing.T) {
	r := newRolloutRig(t, true, false)
	before := toJSON(t, r.deployment(t, "web"))
	for step, want := range map[string]string{
		"k8s_rollout_restart": "rollout-restart Deployment/default/web would-restart",
		"k8s_rollout_pause":   "rollout-pause Deployment/default/web would-pause",
		"k8s_rollout_undo":    "rollout-undo Deployment/default/web would-roll-back:2",
		"k8s_scale 5":         "scale Deployment/default/web would-scale:5",
	} {
		wantRecords(t, r.records(t, on("Deployment/default/web", step)), []string{want})
	}
	if after := toJSON(t, r.deployment(t, "web")); after != before {
		t.Fatalf("a dry run changed the Deployment:\n before %s\n after  %s", before, after)
	}
	n := 0
	for _, p := range r.patches() {
		n++
		if !slices.Equal(p.PatchOptions.DryRun, []string{metav1.DryRunAll}) {
			t.Errorf("patch %d of %s carries DryRun=%v", n, p.GetName(), p.PatchOptions.DryRun)
		}
	}
	if n != 4 {
		t.Fatalf("%d patch requests for 4 steps", n)
	}
	r.onlyReadsAndPatches(t)
}

// ---- the operations, with exactly the requests they send ----

func TestRolloutRestart(t *testing.T) {
	r := newRolloutRig(t, true, true)
	wantRecords(t, r.records(t, on("Deployment/default/web", "k8s_rollout_restart")), []string{"rollout-restart Deployment/default/web restarted"})
	ps := r.patches()
	if len(ps) != 1 || ps[0].GetPatchType() != types.MergePatchType || ps[0].GetName() != "web" || len(ps[0].PatchOptions.DryRun) != 0 {
		t.Fatalf("patches = %+v", ps)
	}
	var body map[string]any
	if err := json.Unmarshal(ps[0].GetPatch(), &body); err != nil {
		t.Fatal(err)
	}
	stamp, found, _ := unstructured.NestedString(body, "spec", "template", "metadata", "annotations", "kubectl.kubernetes.io/restartedAt")
	if !found || !strings.Contains(stamp, "T") || !strings.HasSuffix(stamp, "Z") {
		t.Fatalf("patch body %s lacks an RFC 3339 restartedAt", ps[0].GetPatch())
	}
	stored, _, _ := unstructured.NestedString(r.deployment(t, "web").Object, "spec", "template", "metadata", "annotations", "kubectl.kubernetes.io/restartedAt")
	if stored != stamp {
		t.Fatalf("stored restartedAt %q, sent %q", stored, stamp)
	}
	// the other pod-template fields are untouched
	if r.image(t, "web") != "nginx:1.27" {
		t.Fatal("restart changed the image")
	}
	// StatefulSet and DaemonSet restart too; a ReplicaSet does not
	wantRecords(t, r.records(t, on("StatefulSet/default/db", "k8s_rollout_restart")), []string{"rollout-restart StatefulSet/default/db restarted"})
	wantRecords(t, r.records(t, on("DaemonSet/default/agent", "k8s_rollout_restart")), []string{"rollout-restart DaemonSet/default/agent restarted"})
	if g := r.records(t, on("ReplicaSet/default/web-ccc", "k8s_rollout_restart")); !strings.Contains(g[0], "does not support ReplicaSet") {
		t.Fatalf("got %q", g)
	}
	r.onlyReadsAndPatches(t)
}

func TestRolloutPauseAndResume(t *testing.T) {
	r := newRolloutRig(t, true, true)
	id := "Deployment/default/web"
	wantRecords(t, r.records(t, on(id, "k8s_rollout_pause")), []string{"rollout-pause " + id + " paused"})
	if p, _, _ := unstructured.NestedBool(r.deployment(t, "web").Object, "spec", "paused"); !p {
		t.Fatal("pause did not pause")
	}
	sent := len(r.patches())
	wantRecords(t, r.records(t, on(id, "k8s_rollout_pause")), []string{"rollout-pause " + id + " unchanged"})
	if len(r.patches()) != sent {
		t.Fatal("pausing a paused deployment sent a request")
	}
	wantRecords(t, r.records(t, on(id, "k8s_rollout_resume")), []string{"rollout-resume " + id + " resumed"})
	if p, found, _ := unstructured.NestedBool(r.deployment(t, "web").Object, "spec", "paused"); p || !found {
		t.Fatalf("resume left paused=%v (found %v)", p, found)
	}
	wantRecords(t, r.records(t, on(id, "k8s_rollout_resume")), []string{"rollout-resume " + id + " unchanged"})
	for _, p := range r.patches() {
		if p.GetPatchType() != types.MergePatchType {
			t.Errorf("patch type %s", p.GetPatchType())
		}
	}
	// the halted deployment resumes
	wantRecords(t, r.records(t, on("Deployment/default/halted", "k8s_rollout_resume")), []string{"rollout-resume Deployment/default/halted resumed"})
	// only a Deployment pauses
	if g := r.records(t, on("StatefulSet/default/db", "k8s_rollout_pause")); !strings.Contains(g[0], "does not support StatefulSet") {
		t.Fatalf("got %q", g)
	}
	r.onlyReadsAndPatches(t)
}

func TestRolloutScale(t *testing.T) {
	r := newRolloutRig(t, true, true)
	wantRecords(t, r.records(t, on("Deployment/default/web", "k8s_scale 5")), []string{"scale Deployment/default/web scaled:5"})
	if n, _, _ := unstructured.NestedInt64(r.deployment(t, "web").Object, "spec", "replicas"); n != 5 {
		t.Fatalf("replicas = %d", n)
	}
	var body map[string]any
	ps := r.patches()
	if len(ps) != 1 || ps[0].GetPatchType() != types.MergePatchType || ps[0].GetSubresource() != "" {
		t.Fatalf("scale must be one merge patch of the workload: %+v", ps)
	}
	json.Unmarshal(ps[0].GetPatch(), &body)
	if n, _, _ := unstructured.NestedFloat64(body, "spec", "replicas"); n != 5 || len(body) != 1 {
		t.Fatalf("patch body %s", ps[0].GetPatch())
	}
	wantRecords(t, r.records(t, on("Deployment/default/web", "k8s_scale 5")), []string{"scale Deployment/default/web unchanged"})
	if len(r.patches()) != 1 {
		t.Fatal("scaling to the current size sent a request")
	}
	wantRecords(t, r.records(t, on("Deployment/default/web", "k8s_scale 0")), []string{"scale Deployment/default/web scaled:0"})
	wantRecords(t, r.records(t, on("StatefulSet/default/db", "k8s_scale 3")), []string{"scale StatefulSet/default/db scaled:3"})
	wantRecords(t, r.records(t, on("ReplicaSet/default/web-ccc", "k8s_scale 1")), []string{"scale ReplicaSet/default/web-ccc scaled:1"})
	for line, want := range map[string]string{
		on("Deployment/default/web", "k8s_scale"):      "needs the replica count",
		on("Deployment/default/web", "k8s_scale many"): "needs the replica count",
		on("Deployment/default/web", "k8s_scale -2"):   "must not be negative",
		on("DaemonSet/default/agent", "k8s_scale 2"):   "does not support DaemonSet",
		on("Deployment/default/ghost", "k8s_scale 2"):  "not found",
	} {
		g := r.records(t, line)
		if len(g) != 1 || !strings.HasPrefix(g[0], "[ERROR]") || !strings.Contains(g[0], want) {
			t.Errorf("%s = %q, want an error containing %q", line, g, want)
		}
	}
	r.onlyReadsAndPatches(t)
}

func TestRolloutUndo(t *testing.T) {
	r := newRolloutRig(t, true, true)
	id := "Deployment/default/web"
	// previous revision: 2 (nginx:1.26), found among the ReplicaSets this Deployment owns
	wantRecords(t, r.records(t, on(id, "k8s_rollout_undo")), []string{"rollout-undo " + id + " rolled-back:2"})
	if got := r.image(t, "web"); got != "nginx:1.26" {
		t.Fatalf("image after undo = %s", got)
	}
	tmpl, _, _ := unstructured.NestedMap(r.deployment(t, "web").Object, "spec", "template")
	if labels, _, _ := unstructured.NestedMap(tmpl, "metadata", "labels"); labels["pod-template-hash"] != nil || labels["app"] != "web" {
		t.Fatalf("the rolled-back template keeps the controller's hash label: %v", labels)
	}
	ps := r.patches()
	if len(ps) != 1 || ps[0].GetPatchType() != types.JSONPatchType {
		t.Fatalf("undo must be one JSON patch: %+v", ps)
	}
	var ops []map[string]any
	json.Unmarshal(ps[0].GetPatch(), &ops)
	if len(ops) != 1 || ops[0]["op"] != "replace" || ops[0]["path"] != "/spec/template" {
		t.Fatalf("patch %s", ps[0].GetPatch())
	}
	// the template already matches: nothing to do
	wantRecords(t, r.records(t, on(id, "k8s_rollout_undo 2")), []string{"rollout-undo " + id + " unchanged"})
	if len(r.patches()) != 1 {
		t.Fatal("an undo to the current template sent a request")
	}
	// an explicit older revision
	wantRecords(t, r.records(t, on(id, "k8s_rollout_undo 1")), []string{"rollout-undo " + id + " rolled-back:1"})
	if got := r.image(t, "web"); got != "nginx:1.25" {
		t.Fatalf("image after undo 1 = %s", got)
	}
	// loud failures
	for line, want := range map[string]string{
		on(id, "k8s_rollout_undo 9"):                        "revision 9 not found", // the revision of a ReplicaSet this Deployment does not own
		on(id, "k8s_rollout_undo 99"):                       "revision 99 not found",
		on(id, "k8s_rollout_undo abc"):                      "a revision number",
		on(id, "k8s_rollout_undo -1"):                       "must not be negative",
		on("Deployment/default/slow", "k8s_rollout_undo"):   "no previous revision",
		on("Deployment/default/halted", "k8s_rollout_undo"): "is paused",
		on("StatefulSet/default/db", "k8s_rollout_undo"):    "does not support StatefulSet",
	} {
		g := r.records(t, line)
		if len(g) != 1 || !strings.HasPrefix(g[0], "[ERROR]") || !strings.Contains(g[0], want) {
			t.Errorf("%s = %q, want an error containing %q", line, g, want)
		}
	}
	// a ReplicaSet of another Deployment (revision 9, image evil:9) is never a candidate
	if r.image(t, "web") == "evil:9" {
		t.Fatal("undo adopted a ReplicaSet the Deployment does not own")
	}
	r.onlyReadsAndPatches(t)
}

func TestRolloutStepsAcceptAWorkloadPayload(t *testing.T) {
	r := newRolloutRig(t, true, true)
	got := r.records(t, `spawn StrVal: default |> k8s_deployments |> k8s_scale 4`)
	slices.Sort(got)
	wantRecords(t, got, []string{
		"scale Deployment/default/halted scaled:4",
		"scale Deployment/default/slow scaled:4",
		"scale Deployment/default/web scaled:4",
	})
}

func TestRolloutWithNoClusterFailsLoudly(t *testing.T) {
	interp := k8sgraph.ApplyEdition(cli.NewEdition("nocluster-rollout"), nil, k8sgraph.WithWriteFlags(k8sgraph.DefaultWriteFlags())).BuildInterpreter()
	interp.SetPolicy(interp.Policy().Grant(cli.CapClusterRollout))
	for _, step := range []string{"k8s_rollout_status", "k8s_scale 2", "k8s_rollout_restart"} {
		g := values(t, interp, on("Deployment/default/web", step))
		if len(g) != 1 || !strings.HasPrefix(g[0], "[ERROR]") || !strings.Contains(g[0], "no cluster is configured") {
			t.Errorf("%s = %q", step, g)
		}
	}
}

// TestHowToRolloutSamplesMatchTheFakeCluster pins the rollout samples of docs/k8s_howto.md §3.5
// (which need a cluster, so the binary tests cannot run them): the record each one shows is
// exactly what the step answers on the fixture cluster, in a dry run, and the doc shows it.
func TestHowToRolloutSamplesMatchTheFakeCluster(t *testing.T) {
	raw, err := os.ReadFile("../../docs/k8s_howto.md")
	if err != nil {
		t.Fatal(err)
	}
	doc := strings.Join(strings.Fields(string(raw)), " ")
	for _, row := range []struct {
		step, record string
		persist      bool
		real         string // the persisted outcome the doc names, "" when none
	}{
		{"k8s_rollout_restart", "rollout-restart Deployment/default/web would-restart", false, "restarted"},
		{"k8s_rollout_undo", "rollout-undo Deployment/default/web would-roll-back:2", false, ""},
		{"k8s_scale 5", "scale Deployment/default/web would-scale:5", false, ""},
		{"k8s_rollout_status", "rollout-status Deployment/default/web complete", false, ""},
	} {
		r := newRolloutRig(t, true, row.persist)
		got := r.records(t, on("Deployment/default/web", row.step))
		wantRecords(t, got, []string{row.record})
		if !strings.Contains(doc, "["+row.record+"]") {
			t.Errorf("docs/k8s_howto.md no longer shows [%s]", row.record)
		}
		if row.real != "" {
			if !strings.Contains(doc, row.real) {
				t.Errorf("docs/k8s_howto.md no longer names the real outcome %q", row.real)
			}
			real := newRolloutRig(t, true, true)
			if g := real.records(t, on("Deployment/default/web", row.step)); !strings.HasSuffix(g[0], " "+row.real) {
				t.Errorf("%s persisted = %q, want outcome %q", row.step, g, row.real)
			}
		}
	}
}
