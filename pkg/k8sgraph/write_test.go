package k8sgraph_test

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"os"
	"path/filepath"
	"regexp"
	"sigs.k8s.io/yaml"
	"slices"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"

	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph/k8sgraphtest"
)

const (
	webPath    = "testdata/write/web.yaml"
	scaledPath = "testdata/write/web-scaled.yaml"
	brokenPath = "testdata/write/broken.yaml"
	applyLine  = `spawn StrVal: ` + webPath + ` |> manifests |> k8s_apply`
)

var (
	deploymentsGVR = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	configMapsGVR  = schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}
)

// writeRig is an edition over an empty fake cluster that emulates server-side apply, the way
// the write steps are meant to be used from a trusted script that opted in.
type writeRig struct {
	client *fake.FakeDynamicClient
	interp *cli.Interpreter
	flags  *k8sgraph.WriteFlags
}

// newWriteRig builds the rig. The policy is the trusted script's PLUS the cluster-apply
// opt-in when grant is true; persist is --dry-run=false.
func newWriteRig(t *testing.T, grant, persist bool) *writeRig {
	t.Helper()
	client := k8sgraphtest.FakeClusterFromYAML(t, nil)
	k8sgraphtest.EmulateWrites(client)
	flags := k8sgraph.DefaultWriteFlags()
	flags.DryRun = !persist
	interp := k8sgraph.ApplyEdition(cli.NewEdition("write-test"), client, k8sgraph.WithWriteFlags(flags)).BuildInterpreter()
	if grant {
		interp.SetPolicy(interp.Policy().Grant(cli.CapClusterApply))
	}
	return &writeRig{client, interp, flags}
}

// records runs line and returns every value as [verb id outcome] strings, or "[ERROR] msg".
func (r *writeRig) records(t *testing.T, line string) []string {
	t.Helper()
	var out []string
	for _, v := range r.interp.EvaluateToValues(line) {
		if err, ok := v.(error); ok {
			out = append(out, "[ERROR] "+err.Error())
			continue
		}
		if str, isStr := v.(string); isStr {
			out = append(out, str) // a projection of a record (nth N) is a plain string
			continue
		}
		items, ok := functions.ToSliceAny(v)
		if !ok {
			t.Fatalf("%s: a record must be a list, got %T %v", line, v, v)
		}
		parts := make([]string, len(items))
		for i, it := range items {
			s, isStr := it.(string)
			if !isStr {
				t.Fatalf("%s: record element %d is %T, want string", line, i, it)
			}
			parts[i] = s
		}
		out = append(out, strings.Join(parts, " "))
	}
	if len(out) == 0 {
		t.Fatalf("%s yielded nothing: a write that reports nothing is indistinguishable from one that did nothing", line)
	}
	return out
}

// verbs is the multiset of API verbs the fake saw: the proof of what was sent to the server.
func (r *writeRig) verbs() []string {
	var vs []string
	for _, a := range r.client.Actions() {
		vs = append(vs, a.GetVerb())
	}
	return vs
}

func (r *writeRig) patches() []clienttesting.PatchActionImpl {
	var ps []clienttesting.PatchActionImpl
	for _, a := range r.client.Actions() {
		if p, ok := a.(clienttesting.PatchActionImpl); ok {
			ps = append(ps, p)
		}
	}
	return ps
}

func (r *writeRig) stored(gvr schema.GroupVersionResource, ns, name string) bool {
	_, err := r.client.Tracker().Get(gvr, ns, name)
	return err == nil
}

func wantRecords(t *testing.T, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("records\n got  %q\n want %q", got, want)
	}
}

// ---- layer 1: the capability policy ----

func TestApplyIsDeniedForTheTrustedScript(t *testing.T) {
	r := newWriteRig(t, false, true) // even a real-write setting does not matter: no capability
	if !r.interp.Policy().Allows(cli.CapExec) {
		t.Fatal("test premise: this is the trusted command-line interpreter")
	}
	got := r.records(t, applyLine)
	want := []string{"[ERROR] step 'k8s_apply' requires the cluster-apply capability (policy denies it)"}
	wantRecords(t, got, want)
	if acts := r.client.Actions(); len(acts) != 0 {
		t.Fatalf("a denied apply reached the cluster: %v", acts)
	}
	f := cli.FailureOf(r.interp.EvaluateToValues(applyLine)[0].(error))
	if f.Kind != cli.FailCapabilityDenied || f.Step != "k8s_apply" || f.Capability != "cluster-apply" {
		t.Fatalf("diagnostic = %+v", f)
	}
	// the rollout capability is a different switch
	r.interp.SetPolicy(r.interp.Policy().Grant(cli.CapClusterRollout))
	if got := r.records(t, applyLine); !strings.Contains(got[0], "requires the cluster-apply capability") {
		t.Fatalf("cluster-rollout must not grant apply: %q", got)
	}
}

func TestEditionFlagIsTheOptInForApply(t *testing.T) {
	client := k8sgraphtest.FakeClusterFromYAML(t, nil)
	k8sgraphtest.EmulateWrites(client)
	ed := k8sgraph.ApplyEdition(cli.NewEdition("write-flag-test"), client)
	run := func(argv ...string) string {
		var out, errb strings.Builder
		ed.Command().Execute(argv, strings.NewReader(""), &out, &errb)
		return out.String() + errb.String()
	}
	eval := []string{"eval", "-e", applyLine}
	if out := run(eval...); !strings.Contains(out, "requires the cluster-apply capability") {
		t.Fatalf("without --allow-cluster-apply: %q", out)
	}
	if got := len(client.Actions()); got != 0 {
		t.Fatalf("denied apply sent %d requests", got)
	}
	out := run(append([]string{"--allow-cluster-apply"}, eval...)...)
	if !strings.Contains(out, "would-create") || strings.Contains(out, "[ERROR]") {
		t.Fatalf("--allow-cluster-apply is a dry run by default: %q", out)
	}
	if r := (&writeRig{client: client}); r.stored(deploymentsGVR, "default", "web") {
		t.Fatal("the default run persisted an object")
	}
	out = run("--allow-cluster-apply", "--dry-run=false", "eval", "-e", applyLine)
	if !strings.Contains(out, "created") || strings.Contains(out, "would-") || strings.Contains(out, "[ERROR]") {
		t.Fatalf("--dry-run=false: %q", out)
	}
	if _, err := client.Tracker().Get(deploymentsGVR, "default", "web"); err != nil {
		t.Fatalf("the real apply stored nothing: %v", err)
	}
}

// ---- layer 2: dry run is the default ----

func TestDryRunIsTheDefaultAndPersistsNothing(t *testing.T) {
	r := newWriteRig(t, true, false)
	got := r.records(t, applyLine)
	wantRecords(t, got, []string{
		"apply ConfigMap/default/web-config would-create",
		"apply Deployment/default/web would-create",
		"apply Service/default/web would-create",
	})
	for _, ps := range r.patches() {
		if !slices.Equal(ps.PatchOptions.DryRun, []string{metav1.DryRunAll}) {
			t.Errorf("request for %s carries DryRun=%v, want [All]", ps.GetName(), ps.PatchOptions.DryRun)
		}
	}
	if n := len(r.patches()); n != 3 {
		t.Fatalf("want one apply request per object, got %d", n)
	}
	for _, name := range []string{"web-config"} {
		if r.stored(configMapsGVR, "default", name) {
			t.Fatalf("dry run persisted %s", name)
		}
	}
	if r.stored(deploymentsGVR, "default", "web") {
		t.Fatal("dry run persisted the Deployment")
	}
}

func TestApplyCreatesThenIsIdempotentThenConfigures(t *testing.T) {
	r := newWriteRig(t, true, true)
	wantRecords(t, r.records(t, applyLine), []string{
		"apply ConfigMap/default/web-config created",
		"apply Deployment/default/web created",
		"apply Service/default/web created",
	})
	for _, ps := range r.patches() {
		if ps.GetPatchType() != types.ApplyPatchType {
			t.Errorf("%s: patch type %s, want server-side apply", ps.GetName(), ps.GetPatchType())
		}
		if ps.PatchOptions.FieldManager != "gatwd" {
			t.Errorf("%s: field manager %q, want gatwd", ps.GetName(), ps.PatchOptions.FieldManager)
		}
		if len(ps.PatchOptions.DryRun) != 0 {
			t.Errorf("%s: a real write carries DryRun %v", ps.GetName(), ps.PatchOptions.DryRun)
		}
		if ps.PatchOptions.Force != nil && *ps.PatchOptions.Force {
			t.Errorf("%s: conflicts were forced without --force-conflicts", ps.GetName())
		}
	}

	// the same objects again: nothing changes
	wantRecords(t, r.records(t, applyLine), []string{
		"apply ConfigMap/default/web-config unchanged",
		"apply Deployment/default/web unchanged",
		"apply Service/default/web unchanged",
	})

	// a changed field: only that object is configured
	got := r.records(t, `spawn StrVal: `+scaledPath+` |> manifests |> k8s_apply`)
	wantRecords(t, got, []string{
		"apply ConfigMap/default/web-config unchanged",
		"apply Deployment/default/web configured",
		"apply Service/default/web unchanged",
	})
	obj, err := r.client.Tracker().Get(deploymentsGVR, "default", "web")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ReplaceAll(toJSON(t, obj), " ", ""), `"replicas":5`) {
		t.Fatalf("the Deployment was not reconfigured: %s", toJSON(t, obj))
	}

	// and in a dry run the same change is only announced
	r.flags.DryRun = true
	got = r.records(t, applyLine) // web.yaml again: replicas 2 differs from the stored 5
	if got[1] != "apply Deployment/default/web would-configure" {
		t.Fatalf("dry-run change = %q", got)
	}
	obj, _ = r.client.Tracker().Get(deploymentsGVR, "default", "web")
	if !strings.Contains(strings.ReplaceAll(toJSON(t, obj), " ", ""), `"replicas":5`) {
		t.Fatalf("a dry run changed the stored Deployment: %s", toJSON(t, obj))
	}
}

func TestForceConflictsAndFieldManagerAreExplicit(t *testing.T) {
	r := newWriteRig(t, true, true)
	r.flags.ForceConflicts = true
	r.flags.FieldManager = "ci-pipeline"
	r.records(t, applyLine)
	for _, ps := range r.patches() {
		if ps.PatchOptions.Force == nil || !*ps.PatchOptions.Force || ps.PatchOptions.FieldManager != "ci-pipeline" {
			t.Errorf("%s: options %+v", ps.GetName(), ps.PatchOptions)
		}
	}
}

// ---- fail loud ----

func TestWriteRefusalsAreLoudAndDoNotSwallowTheNeighbours(t *testing.T) {
	r := newWriteRig(t, true, true)
	got := r.records(t, `spawn StrVal: `+brokenPath+` |> manifests |> k8s_apply`)
	if len(got) != 5 {
		t.Fatalf("want one result per object (5), got %d: %q", len(got), got)
	}
	want := []string{
		"apply ConfigMap/default/first created",
		"[ERROR] cannot write a ConfigMap with no metadata.name",
		"[ERROR] cannot apply ConfigMap/_/no-namespace: ConfigMap is namespaced and the manifest names no metadata.namespace",
		"[ERROR] cannot apply Widget/default/w:",
		"apply ConfigMap/default/last created",
	}
	for i, w := range want {
		if !strings.HasPrefix(got[i], w) {
			t.Errorf("result %d = %q, want prefix %q", i, got[i], w)
		}
	}
	if !strings.Contains(got[3], "Widget") || !strings.Contains(got[3], "example.com/v1") {
		t.Errorf("the unknown-kind error must name kind and apiVersion: %q", got[3])
	}
	// only the two good objects reached the server as writes
	if n := len(r.patches()); n != 2 {
		t.Fatalf("%d apply requests for 2 valid objects", n)
	}
}

func TestApplyOfANonObjectIsAnError(t *testing.T) {
	r := newWriteRig(t, true, true)
	for line, want := range map[string]string{
		`spawn StrVal: x |> emit 5 |> k8s_apply`:     "cannot write",
		`spawn StrVal: x |> emit hello |> k8s_apply`: "cannot write string as a Kubernetes object",
		applyLine + ` extra`:                         "takes no argument",
	} {
		got := r.records(t, line)
		for _, g := range got {
			if !strings.HasPrefix(g, "[ERROR]") || !strings.Contains(g, want) {
				t.Errorf("%s: result %q, want an error containing %q (all results: %q)", line, g, want, got)
			}
		}
	}
	if len(r.client.Actions()) != 0 {
		t.Fatalf("a non-object reached the cluster: %v", r.verbs())
	}
}

func TestApplyWithNoClusterFailsLoudly(t *testing.T) {
	interp := k8sgraph.ApplyEdition(cli.NewEdition("nocluster"), nil, k8sgraph.WithWriteFlags(k8sgraph.DefaultWriteFlags())).BuildInterpreter()
	interp.SetPolicy(interp.Policy().Grant(cli.CapClusterApply))
	vals := interp.EvaluateToValues(applyLine)
	if len(vals) != 3 {
		t.Fatalf("want one error per object, got %v", vals)
	}
	for _, v := range vals {
		err, ok := v.(error)
		if !ok || !strings.Contains(err.Error(), "no cluster is configured") {
			t.Fatalf("got %v, want a 'no cluster is configured' error", v)
		}
	}
}

func TestApplyRefusesToWriteToAFailedClient(t *testing.T) {
	// the lazy client of the edition binary with no kubeconfig: an error per object, never silence
	boom := k8sgraph.LazyClient(func() (dynamicIface, error) { return nil, os.ErrNotExist })
	interp := k8sgraph.ApplyEdition(cli.NewEdition("lazyfail"), boom, k8sgraph.WithWriteFlags(k8sgraph.DefaultWriteFlags())).BuildInterpreter()
	interp.SetPolicy(interp.Policy().Grant(cli.CapClusterApply))
	for _, v := range interp.EvaluateToValues(applyLine) {
		err, ok := v.(error)
		if !ok || !strings.Contains(err.Error(), "no cluster access") {
			t.Fatalf("got %v, want a 'no cluster access' error", v)
		}
	}
}

// ---- ids and the DSL ----

func TestRecordIdsAreTheGraphNodeIds(t *testing.T) {
	r := newWriteRig(t, true, false)
	ids := r.records(t, `spawn StrVal: `+webPath+` |> manifests |> k8s_apply |> nth 1`)
	nodes := edgesOf(t, r.interp, `spawn StrVal: `+webPath+` |> manifests |> k8s_graph |> graph_nodes |> collect_list`)
	slices.Sort(ids)
	if len(ids) != 3 || !slices.Equal(ids, nodes) {
		t.Fatalf("write ids %q differ from the graph's node ids %q", ids, nodes)
	}
}

func TestApplyAppliesAListInTheOrderTheDSLComputed(t *testing.T) {
	// k8s_apply_objects puts the payloads in k8s_apply_order; k8s_apply takes the list as a batch
	r := newWriteRig(t, true, true)
	got := r.records(t, `spawn StrVal: `+webPath+` |> manifests |> k8s_graph |> k8s_apply_objects |> k8s_apply |> nth 1`)
	want := edgesOf(t, r.interp, `spawn StrVal: `+webPath+` |> manifests |> k8s_graph |> k8s_apply_order |> collect_list`)
	slices.Sort(got)
	slices.Sort(want)
	if len(got) != 3 || !slices.Equal(got, want) {
		t.Fatalf("applied %q, apply order %q", got, want)
	}
	// the order itself: the requests reach the server in the order of the list
	var sent []string
	for _, p := range r.patches() {
		sent = append(sent, p.GetName())
	}
	order := r.interp.EvaluateToValues(`spawn StrVal: ` + webPath + ` |> manifests |> k8s_graph |> k8s_apply_order |> collect_list`)
	items, _ := functions.ToSliceAny(order[0])
	if len(sent) != len(items) {
		t.Fatalf("sent %v for order %v", sent, items)
	}
	for i, id := range items {
		if !strings.HasSuffix(id.(string), "/"+sent[i]) {
			t.Errorf("request %d was for %s, the DSL order says %v", i, sent[i], id)
		}
	}
	// and a list that contains a non-object is an error for that element only
	r2 := newWriteRig(t, true, false)
	mixed := r2.records(t, `spawn StrVal: x |> emit [1, 2] |> k8s_apply`)
	for _, m := range mixed {
		if !strings.HasPrefix(m, "[ERROR]") {
			t.Errorf("a non-object in a batch must be an error, got %q", m)
		}
	}
}

// ---- the offline twin: --out DIR ----

func offlineRig(t *testing.T, out string, policy cli.Policy) *cli.Interpreter {
	t.Helper()
	flags := k8sgraph.DefaultWriteFlags()
	flags.Out = out
	interp := k8sgraph.ApplyEdition(cli.NewEdition("offline-write"), nil, k8sgraph.WithWriteFlags(flags)).BuildInterpreter() // no cluster at all
	interp.SetPolicy(policy)
	return interp
}

func values(t *testing.T, interp *cli.Interpreter, line string) []string {
	t.Helper()
	r := &writeRig{interp: interp}
	return r.records(t, line)
}

func TestOutRendersOneYamlFilePerObjectWithNoCluster(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "rendered") // not yet existing: the step creates it
	interp := offlineRig(t, dir, cli.TrustedPolicy().Grant(cli.CapClusterApply))
	wantRecords(t, values(t, interp, applyLine), []string{
		"apply ConfigMap/default/web-config written",
		"apply Deployment/default/web written",
		"apply Service/default/web written",
	})
	files, _ := filepath.Glob(filepath.Join(dir, "*.yaml"))
	var names []string
	for _, f := range files {
		names = append(names, filepath.Base(f))
	}
	wantNames := []string{"ConfigMap.default.web-config.yaml", "Deployment.default.web.yaml", "Service.default.web.yaml"}
	if !slices.Equal(names, wantNames) {
		t.Fatalf("files %v, want %v", names, wantNames)
	}
	// a second render is idempotent, like a second apply
	wantRecords(t, values(t, interp, applyLine), []string{
		"apply ConfigMap/default/web-config unchanged",
		"apply Deployment/default/web unchanged",
		"apply Service/default/web unchanged",
	})
	// a changed object rewrites only its file
	scaled := values(t, interp, `spawn StrVal: `+scaledPath+` |> manifests |> k8s_apply`)
	if scaled[1] != "apply Deployment/default/web written" || scaled[0] != "apply ConfigMap/default/web-config unchanged" {
		t.Fatalf("re-render of a changed object: %q", scaled)
	}

	// round trip: the rendered directory is read back by `manifests` as the same graph
	reader := k8sgraph.ApplyEdition(cli.NewEdition("reader"), nil).BuildInterpreter()
	for _, q := range []string{"graph_nodes", "graph_edges"} {
		want := edgesOf(t, reader, `spawn StrVal: `+scaledPath+` |> manifests |> k8s_graph |> `+q+` |> collect_list`)
		got := edgesOf(t, reader, `spawn StrVal: `+dir+` |> manifests |> k8s_graph |> `+q+` |> collect_list`)
		if len(want) == 0 && q == "graph_nodes" || !slices.Equal(got, want) {
			t.Fatalf("%s: rendered %v, source %v", q, got, want)
		}
	}
}

func TestOutParityWithTheLiveApply(t *testing.T) {
	// the SAME objects, applied to the fake cluster and rendered to files, are the same content
	rig := newWriteRig(t, true, true)
	rig.records(t, applyLine)
	dir := t.TempDir()
	interp := offlineRig(t, dir, cli.TrustedPolicy().Grant(cli.CapClusterApply))
	values(t, interp, applyLine)
	for _, c := range []struct {
		gvr        schema.GroupVersionResource
		file, name string
	}{
		{configMapsGVR, "ConfigMap.default.web-config.yaml", "web-config"},
		{deploymentsGVR, "Deployment.default.web.yaml", "web"},
	} {
		stored, err := rig.client.Tracker().Get(c.gvr, "default", c.name)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(filepath.Join(dir, c.file))
		if err != nil {
			t.Fatal(err)
		}
		var fromFile map[string]any
		if err := yaml.Unmarshal(raw, &fromFile); err != nil {
			t.Fatal(err)
		}
		a, _ := json.Marshal(fromFile)
		b, _ := json.Marshal(stored.(*unstructured.Unstructured).Object)
		if string(a) != string(b) {
			t.Errorf("%s: file\n  %s\n live\n  %s", c.name, a, b)
		}
	}
}

func TestOutNeedsTheFileSystemCapabilityAndTheApplyCapability(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "must-not-exist")
	// apply without fs: the render is denied, naming fs
	interp := offlineRig(t, dir, cli.DenyAll().Grant(cli.CapClusterApply))
	for _, g := range values(t, interp, applyLine) {
		if g != "[ERROR] step 'k8s_apply' requires the fs capability (policy denies it)" {
			t.Errorf("got %q", g)
		}
	}
	// fs without apply: denied by the apply gate (the trusted script has fs)
	interp = offlineRig(t, dir, cli.TrustedPolicy())
	if g := values(t, interp, applyLine); len(g) != 1 || !strings.Contains(g[0], "requires the cluster-apply capability") {
		t.Errorf("got %q", g)
	}
	if _, err := os.Stat(dir); err == nil {
		t.Fatal("a denied render created the directory")
	}
}

func TestOutDoesNotTouchTheCluster(t *testing.T) {
	r := newWriteRig(t, true, true)
	r.flags.Out = t.TempDir()
	r.records(t, applyLine)
	if acts := r.client.Actions(); len(acts) != 0 {
		t.Fatalf("--out talked to the cluster: %v", r.verbs())
	}
}

func TestOutRefusesAnInvalidObjectAndKeepsTheRest(t *testing.T) {
	dir := t.TempDir()
	interp := offlineRig(t, dir, cli.TrustedPolicy().Grant(cli.CapClusterApply))
	got := values(t, interp, `spawn StrVal: `+brokenPath+` |> manifests |> k8s_apply`)
	if len(got) != 5 || !strings.HasPrefix(got[1], "[ERROR] cannot write a ConfigMap with no metadata.name") {
		t.Fatalf("got %q", got)
	}
	for i, w := range map[int]string{0: "apply ConfigMap/default/first written", 2: "apply ConfigMap/_/no-namespace written", 3: "apply Widget/default/w written", 4: "apply ConfigMap/default/last written"} {
		if got[i] != w {
			t.Errorf("result %d = %q, want %q", i, got[i], w)
		}
	}
	if files, _ := filepath.Glob(filepath.Join(dir, "*.yaml")); len(files) != 4 {
		t.Fatalf("%d files for 4 valid objects: %v", len(files), files)
	}
}

func TestOutFlagOnTheCommandLine(t *testing.T) {
	dir := t.TempDir()
	ed := k8sgraph.ApplyEdition(cli.NewEdition("out-flag-test"), nil)
	var out, errb strings.Builder
	ed.Command().Execute([]string{"--allow-cluster-apply", "--out", dir, "eval", "-e", applyLine}, strings.NewReader(""), &out, &errb)
	if !strings.Contains(out.String(), "written") || strings.Contains(out.String(), "[ERROR]") {
		t.Fatalf("stdout %q stderr %q", out.String(), errb.String())
	}
	if files, _ := filepath.Glob(filepath.Join(dir, "*.yaml")); len(files) != 3 {
		t.Fatalf("files %v", files)
	}
}

// ---- guards ----

// TestNoCallCanDeleteAnything: the first version writes by apply and patch only. No
// non-test source of the adapters or the shim may call a delete.
func TestNoCallCanDeleteAnything(t *testing.T) {
	deleteCall := regexp.MustCompile(`\.(Delete|DeleteCollection|DeleteWithResult)\(`)
	var scanned int
	for _, dir := range []string{".", "../../adapters"} {
		files, _ := filepath.Glob(filepath.Join(dir, "*.go"))
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			scanned++
			raw, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			for n, line := range strings.Split(string(raw), "\n") {
				if t := strings.TrimSpace(line); strings.HasPrefix(t, "//") {
					continue
				}
				if deleteCall.MatchString(line) {
					t.Errorf("%s:%d calls a delete: the write steps never delete or prune\n  %s", f, n+1, strings.TrimSpace(line))
				}
			}
		}
	}
	if scanned < 10 {
		t.Fatalf("scanned only %d files: the guard is looking in the wrong place", scanned)
	}
}

// TestEveryWriteStepIsGated fails the build when a write step is wired without the capability
// gate: a recordStep(...) constructor call must sit inside cli.GatedStep(...).
func TestEveryWriteStepIsGated(t *testing.T) {
	files, _ := filepath.Glob("*.go")
	var sites int
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(string(raw), "\n")
		for n, line := range lines {
			trim := strings.TrimSpace(line)
			if strings.HasPrefix(trim, "//") || !strings.Contains(line, "recordStep(") || strings.HasPrefix(trim, "func recordStep") {
				continue
			}
			sites++
			if !strings.Contains(line, "cli.GatedStep(cli.Cap") {
				t.Errorf("%s:%d wires a write step without cli.GatedStep(cli.Cap…, …):\n  %s", f, n+1, trim)
			}
		}
	}
	if sites == 0 {
		t.Fatal("no write step found: the guard is looking for the wrong constructor")
	}
}

func toJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := jsonMarshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

type dynamicIface = dynamic.Interface

var jsonMarshal = json.Marshal
