package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/pkg/k8sgraph/k8sgraphtest"
)

func violations(t *testing.T, vals []any) []string {
	t.Helper()
	var out []string
	for _, v := range vals {
		out = append(out, fmt.Sprint(v))
	}
	return out
}

func TestOperatorEmbeddedRules(t *testing.T) {
	app, err := newAppliance(nil, filepath.Join(t.TempDir(), "absent.gatwd"))
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"testdata/idle.yaml":                     "[[UnselectedPod Pod/default/stray] [IdleDeployment Deployment/default/lonely]]",
		"../../pkg/k8sgraph/testdata/stack.yaml": "[[UnselectedPod Pod/default/other-pod]]",
	}
	for path, want := range cases {
		got := violations(t, app.Execute(checkPipeline(path)))
		if !reflect.DeepEqual(got, []string{want}) {
			t.Errorf("%s: violations = %v, want %s", path, got, want)
		}
	}
}

const brokenWant = "[[UnselectedPod Pod/default/orphan] [UnusedService Service/default/unused-svc] [DanglingIngress Ingress/default/front] [MissingConfigMap Pod/default/orphan] [DependencyCycle [Deployment/default/loop-a Deployment/default/loop-b]]]"

// TestOperatorAllRules: one fixture violating every embedded rule yields exactly the
// expected violations, in rule order — and the compliant parts (a selector-less
// ExternalName Service, the missing-config Pod's other fields) are not flagged.
func TestOperatorAllRules(t *testing.T) {
	app, err := newAppliance(nil, filepath.Join(t.TempDir(), "absent.gatwd"))
	if err != nil {
		t.Fatal(err)
	}
	got := violations(t, app.Execute(checkPipeline("testdata/broken.yaml")))
	if !reflect.DeepEqual(got, []string{brokenWant}) {
		t.Fatalf("violations = %v\nwant        %s", got, brokenWant)
	}
}

// TestOperatorLiveSource: the operator, given a (fake) cluster instead of manifests,
// reports exactly what it reports for the same objects offline.
func TestOperatorLiveSource(t *testing.T) {
	app, err := newAppliance(k8sgraphtest.FakeClusterFromFile(t, "testdata/broken.yaml"), filepath.Join(t.TempDir(), "absent.gatwd"))
	if err != nil {
		t.Fatal(err)
	}
	got := violations(t, app.Execute(livePipeline("default")))
	if !reflect.DeepEqual(got, []string{brokenWant}) {
		t.Fatalf("live violations = %v\nwant              %s", got, brokenWant)
	}
}

func TestOperatorHotReloadsRules(t *testing.T) {
	override := filepath.Join(t.TempDir(), "rules.gatwd")
	app, err := newAppliance(nil, override)
	if err != nil {
		t.Fatal(err)
	}

	strict := "def compliance = k8s_graph |> fold (\\_ g -> [ [\"AnyPod\", (k8s_id_of p)] | p <- (graph_payloads_of \"Pod\" g) ]) 0\n"
	if err := os.WriteFile(override, []byte(strict), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := app.Reload(); err != nil {
		t.Fatal(err)
	}
	if got := violations(t, app.Execute(checkPipeline("testdata/idle.yaml"))); !reflect.DeepEqual(got, []string{"[[AnyPod Pod/default/stray]]"}) {
		t.Fatalf("after reload, violations = %v", got)
	}

	if err := os.Remove(override); err != nil {
		t.Fatal(err)
	}
	if err := app.Reload(); err != nil {
		t.Fatal(err)
	}
	if got := violations(t, app.Execute(checkPipeline("testdata/idle.yaml"))); !reflect.DeepEqual(got, []string{"[[UnselectedPod Pod/default/stray] [IdleDeployment Deployment/default/lonely]]"}) {
		t.Fatalf("after override removal, violations = %v", got)
	}
}

// TestOperatorReportsDuplicateIds is the negative control of qualified node ids: two
// documents with the same kind, namespace and name are not "no violations" (which prints
// []) but an error that names the id — the operator never rules on a graph in which one of
// the two resources has silently replaced the other.
func TestOperatorReportsDuplicateIds(t *testing.T) {
	app, err := newAppliance(nil, filepath.Join(t.TempDir(), "absent.gatwd"))
	if err != nil {
		t.Fatal(err)
	}
	got := violations(t, app.Execute(checkPipeline("../../pkg/k8sgraph/testdata/duplicates.yaml")))
	if len(got) != 1 || !strings.Contains(got[0], `duplicate node id "Pod/default/dup"`) {
		t.Fatalf("violations = %v, want one error naming Pod/default/dup", got)
	}
}
