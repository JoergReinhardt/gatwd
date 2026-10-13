package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
)

// The how-to (docs/k8s_howto.md) ships its examples as real files under
// examples/k8s_howto. These tests execute them, so the guide cannot drift from the code.

const howtoDir = "../../examples/k8s_howto"

func howtoEdition() *cli.Interpreter {
	return k8sgraph.ApplyEdition(cli.NewEdition("howto"), nil).BuildInterpreter()
}

func one(t *testing.T, i *cli.Interpreter, line string) string {
	t.Helper()
	vals := i.EvaluateToValues(line)
	if len(vals) != 1 {
		t.Fatalf("%s: want one value, got %v", line, vals)
	}
	return fmt.Sprint(vals[0])
}

// TestHowToTasks: tasks.gatwd takes the manifest path from its arguments; `main` runs every
// task and --task selects one — through the same Run entry the command line uses, with no
// text surgery on the script.
func TestHowToTasks(t *testing.T) {
	script := filepath.Join(howtoDir, "tasks.gatwd")
	cluster := filepath.Join(howtoDir, "cluster")
	run := func(task string, args ...string) cli.RunResult {
		t.Helper()
		res, err := howtoEdition().Run(cli.RunOptions{Script: script, Task: task, Args: args})
		if err != nil {
			t.Fatalf("task %q: %v", task, err)
		}
		return res
	}

	if res := run("trace_to_pod", cluster); !reflect.DeepEqual(res.Lines, []string{"[Service/default/nginx-svc Pod/default/app-pod]"}) {
		t.Errorf("trace_to_pod = %v", res.Lines)
	}
	if res := run("who_configures_what", cluster); len(res.Lines) != 1 || !strings.Contains(res.Lines[0], "ConfigMap/default/app-config-Configures->Deployment/default/worker") {
		t.Errorf("who_configures_what = %v", res.Lines)
	}
	if res := run("pod_count", cluster); !reflect.DeepEqual(res.Lines, []string{"1"}) {
		t.Errorf("pod_count = %v", res.Lines)
	}

	// No --task: `main` merges the three tasks, so all three results appear.
	res := run("", cluster)
	if res.Task != "main" || len(res.Lines) != 3 {
		t.Fatalf("main = %+v, want task main with 3 lines", res)
	}
	all := strings.Join(res.Lines, "\n")
	for _, want := range []string{"ConfigMap/default/app-config-Configures->Deployment/default/worker", "[Service/default/nginx-svc Pod/default/app-pod]", "1"} {
		if !strings.Contains(all, want) {
			t.Errorf("main output lacks %q:\n%s", want, all)
		}
	}

	// A task name that does not exist is an error naming the defined ones.
	if _, err := howtoEdition().Run(cli.RunOptions{Script: script, Task: "nope", Args: []string{cluster}}); err == nil ||
		!strings.Contains(err.Error(), "trace_to_pod") {
		t.Errorf("unknown task: error = %v, want one listing the defined tasks", err)
	}
}

// TestHowToTasksCommandLine builds the k8s edition and runs tasks.gatwd exactly as the
// how-to shows: script, --task, and arguments after `--`.
func TestHowToTasksCommandLine(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not on PATH")
	}
	bin := filepath.Join(t.TempDir(), "gatwd-k8s")
	build := exec.Command(goBin, "build", "-o", bin, "../gatwd-k8s")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	script := filepath.Join(howtoDir, "tasks.gatwd")
	cluster := filepath.Join(howtoDir, "cluster")

	out, err := exec.Command(bin, script, "--task", "trace_to_pod", "--", cluster).Output()
	if err != nil || !strings.Contains(string(out), "∷ [Service/default/nginx-svc Pod/default/app-pod]") {
		t.Fatalf("--task trace_to_pod: %v\n%s", err, out)
	}
	out, err = exec.Command(bin, script, "--", cluster).Output()
	if err != nil || strings.Count(string(out), "∷ ") != 3 {
		t.Fatalf("main: %v\n%s", err, out)
	}
	if err := exec.Command(bin, script, "--task", "nope").Run(); err == nil {
		t.Fatal("an unknown task must exit non-zero")
	}
}

// TestHowToExtendModelFromScript: the database extension needs no Go and no rebuild — a
// script that defines relations over the same payload shape extends the model, and the
// original relations are untouched.
func TestHowToExtendModelFromScript(t *testing.T) {
	i := howtoEdition()
	if err := i.LoadScriptFS(os.DirFS(howtoDir), "infra_relations.gatwd"); err != nil {
		t.Fatal(err)
	}
	src := "spawn StrVal: " + filepath.Join(howtoDir, "cluster") + " |> manifests |> infra_graph"
	cases := map[string]string{
		src + " |> graph_edges_of UsesDatabase |> collect_list":                                                   "[Secret/default/orders-db-credentials-UsesDatabase->Database/default/orders]",
		src + " |> graph_shortest_path Secret/default/orders-db-credentials DBMS/default/pg-prod |> collect_list": "[Secret/default/orders-db-credentials Database/default/orders DBMS/default/pg-prod]",
		src + " |> graph_edges_of Mounts |> collect_list":                                                         "[PersistentVolumeClaim/default/app-data-Mounts->Pod/default/app-pod]",
	}
	for line, want := range cases {
		if got := one(t, i, line); got != want {
			t.Errorf("%s\n got  %s\n want %s", line, got, want)
		}
	}
}

// TestHowToRulesOverride: an override file replaces the embedded rules wholesale; the
// shipped example keeps all six embedded rules and adds one.
func TestHowToRulesOverride(t *testing.T) {
	app, err := newAppliance(nil, filepath.Join(howtoDir, "rules.gatwd"))
	if err != nil {
		t.Fatal(err)
	}
	got := violations(t, app.Execute(checkPipeline(filepath.Join(howtoDir, "cluster"))))
	want := "[[IdleDeployment Deployment/default/worker] [UnusedService Service/default/backend-svc] [MissingConfigMap Deployment/default/worker] [DatabaseWithoutCredentials Database/default/audit]]"
	if !reflect.DeepEqual(got, []string{want}) {
		t.Fatalf("violations = %v\nwant        %s", got, want)
	}

	// A minimal override that only REUSES an embedded rule must not silently look like success:
	// the embedded rules are gone, so the pipeline cannot find them.
	p := filepath.Join(t.TempDir(), "reuse.gatwd")
	_ = os.WriteFile(p, []byte(`def compliance = k8s_graph |> fold probe 0
def probe = \_ g -> rule_idle_deployment g
`), 0o644)
	app2, _ := newAppliance(nil, p)
	out := fmt.Sprint(app2.Execute(checkPipeline(filepath.Join(howtoDir, "cluster"))))
	if !strings.Contains(out, "rule_idle_deployment") {
		t.Fatalf("expected an error naming the missing embedded rule (override replaces, not layers), got %s", out)
	}
}

// TestHowToDeployManifest: deploy.yaml decodes, and by default the operator is NOT granted
// access to Secrets.
func TestHowToDeployManifest(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(howtoDir, "deploy.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	dec := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(raw), 4096)
	var kinds []string
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
		kinds = append(kinds, obj.GetKind())
		if obj.GetKind() == "ClusterRole" {
			rules, _, _ := unstructured.NestedSlice(obj.Object, "rules")
			for _, r := range rules {
				res, _, _ := unstructured.NestedStringSlice(r.(map[string]any), "resources")
				for _, name := range res {
					if name == "secrets" {
						t.Errorf("the default ClusterRole must not grant access to secrets")
					}
				}
			}
		}
	}
	want := []string{"Namespace", "ServiceAccount", "ClusterRole", "ClusterRoleBinding", "Deployment"}
	if !reflect.DeepEqual(kinds, want) {
		t.Fatalf("kinds = %v, want %v", kinds, want)
	}
}
