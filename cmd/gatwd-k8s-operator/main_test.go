package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/pkg/cli"
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
	app, err := newAppliance(nil, filepath.Join(t.TempDir(), "absent.gatwd"), cli.OverrideLayer)
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"testdata/idle.yaml":                     "[[UnselectedPod Pod/default/stray no Service selects this Pod, so it receives no traffic] [IdleDeployment Deployment/default/lonely this Deployment owns no ReplicaSet, so it manages no Pods]]",
		"../../pkg/k8sgraph/testdata/stack.yaml": "[[UnselectedPod Pod/default/other-pod no Service selects this Pod, so it receives no traffic]]",
	}
	for path, want := range cases {
		got := violations(t, app.Execute(checkPipeline(path)))
		if !reflect.DeepEqual(got, []string{want}) {
			t.Errorf("%s: violations = %v, want %s", path, got, want)
		}
	}
}

const brokenWant = "[[UnselectedPod Pod/default/orphan no Service selects this Pod, so it receives no traffic] [UnusedService Service/default/unused-svc the selector of this Service matches no Pod] [DanglingIngress Ingress/default/front this Ingress routes to a Service that does not exist] [MissingConfigMap Pod/default/orphan this workload references a ConfigMap that does not exist] [DependencyCycle [Deployment/default/loop-a Deployment/default/loop-b] these resources own or depend on each other in a cycle]]"

// TestOperatorAllRules: one fixture violating every embedded rule yields exactly the
// expected violations, in rule order — and the compliant parts (a selector-less
// ExternalName Service, the missing-config Pod's other fields) are not flagged.
func TestOperatorAllRules(t *testing.T) {
	app, err := newAppliance(nil, filepath.Join(t.TempDir(), "absent.gatwd"), cli.OverrideLayer)
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
	app, err := newAppliance(k8sgraphtest.FakeClusterFromFile(t, "testdata/broken.yaml"), filepath.Join(t.TempDir(), "absent.gatwd"), cli.OverrideLayer)
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
	app, err := newAppliance(nil, override, cli.OverrideLayer)
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
	if got := violations(t, app.Execute(checkPipeline("testdata/idle.yaml"))); !reflect.DeepEqual(got, []string{"[[UnselectedPod Pod/default/stray no Service selects this Pod, so it receives no traffic] [IdleDeployment Deployment/default/lonely this Deployment owns no ReplicaSet, so it manages no Pods]]"}) {
		t.Fatalf("after override removal, violations = %v", got)
	}
}

// TestOperatorBrokenOverride: a broken ConfigMap fails the start loudly and, on a hot reload,
// leaves the last good rules serving — it never silently falls back to the embedded rules.
func TestOperatorBrokenOverride(t *testing.T) {
	override := filepath.Join(t.TempDir(), "rules.gatwd")
	good := "def compliance = k8s_graph |> fold (\\_ g -> [ [\"AnyPod\", (hamt_path \"metadata.name\" p)] | p <- (graph_payloads_of \"Pod\" g) ]) 0\n"
	if err := os.WriteFile(override, []byte(good), 0o644); err != nil {
		t.Fatal(err)
	}
	app, err := newAppliance(nil, override, cli.OverrideLayer)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(override, []byte("def compliance = k8s_graph |> fold (\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var le *cli.LoadError
	if err := app.Reload(); !errors.As(err, &le) {
		t.Fatalf("Reload of a broken override: %v, want a *cli.LoadError", err)
	}
	if got := violations(t, app.Execute(checkPipeline("testdata/idle.yaml"))); !reflect.DeepEqual(got, []string{"[[AnyPod stray]]"}) {
		t.Fatalf("after a failed reload the last good rules must serve, got %v", got)
	}
	if _, err := newAppliance(nil, override, cli.OverrideLayer); !errors.As(err, &le) {
		t.Fatalf("a broken override at start: %v, want a *cli.LoadError", err)
	}

	// replace mode: an override without the entry point is a load error, not a runtime surprise.
	if err := os.WriteFile(override, []byte("def rule_x = \\g -> []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := newAppliance(nil, override, cli.OverrideReplace); err == nil || !strings.Contains(err.Error(), "compliance") {
		t.Fatalf("replace mode without compliance: %v", err)
	}
}

// TestOperatorReportsDuplicateIds is the negative control of qualified node ids: two
// documents with the same kind, namespace and name are not "no violations" (which prints
// []) but an error that names the id — the operator never rules on a graph in which one of
// the two resources has silently replaced the other.
func TestOperatorReportsDuplicateIds(t *testing.T) {
	app, err := newAppliance(nil, filepath.Join(t.TempDir(), "absent.gatwd"), cli.OverrideLayer)
	if err != nil {
		t.Fatal(err)
	}
	got := violations(t, app.Execute(checkPipeline("../../pkg/k8sgraph/testdata/duplicates.yaml")))
	if len(got) != 1 || !strings.Contains(got[0], `duplicate node id "Pod/default/dup"`) {
		t.Fatalf("violations = %v, want one error naming Pod/default/dup", got)
	}
}
