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

	if res := run("trace_to_pod", cluster); !reflect.DeepEqual(res.Lines, []string{"[nginx-svc app-pod]"}) {
		t.Errorf("trace_to_pod = %v", res.Lines)
	}
	if res := run("who_configures_what", cluster); len(res.Lines) != 1 || !strings.Contains(res.Lines[0], "app-config-Configures->worker") {
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
	for _, want := range []string{"app-config-Configures->worker", "[nginx-svc app-pod]", "1"} {
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
	if err != nil || !strings.Contains(string(out), "∷ [nginx-svc app-pod]") {
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
		src + " |> graph_edges_of UsesDatabase |> collect_list":                       "[orders-db-credentials-UsesDatabase->orders]",
		src + " |> graph_shortest_path orders-db-credentials pg-prod |> collect_list": "[orders-db-credentials orders pg-prod]",
		src + " |> graph_edges_of Mounts |> collect_list":                             "[app-data-Mounts->app-pod]",
	}
	for line, want := range cases {
		if got := one(t, i, line); got != want {
			t.Errorf("%s\n got  %s\n want %s", line, got, want)
		}
	}
}

// TestHowToRulesOverride: the shipped override is LAYERED on the embedded rules (decision D5): it
// holds only the additions, calls the embedded rule_* closures, and the operator reports what it
// overrode and added. The same additions behind a full copy of the embedded rules give the same
// answer in replace mode, and a layered override that reuses one embedded rule works while the
// replace mode (where the embedded rules are gone) says which rule is missing.
func TestHowToRulesOverride(t *testing.T) {
	override := filepath.Join(howtoDir, "rules.gatwd")
	cluster := filepath.Join(howtoDir, "cluster")
	want := "[[IdleDeployment worker] [UnusedService backend-svc] [MissingConfigMap worker] [DatabaseWithoutCredentials audit]]"

	app, err := newAppliance(nil, override, cli.OverrideLayer)
	if err != nil {
		t.Fatal(err)
	}
	if got := violations(t, app.Execute(checkPipeline(cluster))); !reflect.DeepEqual(got, []string{want}) {
		t.Fatalf("layered violations = %v\nwant              %s", got, want)
	}
	rep := app.Report()
	if !reflect.DeepEqual(rep.Overridden, []string{"compliance"}) {
		t.Errorf("Overridden = %v, want only compliance (the six embedded rules are not copied)", rep.Overridden)
	}
	for _, name := range []string{"rule_database_without_credentials", "infra_graph", "rel_uses_database"} {
		found := false
		for _, a := range rep.Added {
			found = found || a == name
		}
		if !found {
			t.Errorf("Added = %v lacks %s", rep.Added, name)
		}
	}

	// Replace mode: the override must carry the whole rule set, so give it a copy of the
	// embedded rules in front of the same additions.
	embedded, err := rulesFS.ReadFile("rules/compliance.gatwd")
	if err != nil {
		t.Fatal(err)
	}
	additions, err := os.ReadFile(override)
	if err != nil {
		t.Fatal(err)
	}
	full := filepath.Join(t.TempDir(), "full.gatwd")
	if err := os.WriteFile(full, append(append(embedded, '\n'), additions...), 0o644); err != nil {
		t.Fatal(err)
	}
	app3, err := newAppliance(nil, full, cli.OverrideReplace)
	if err != nil {
		t.Fatal(err)
	}
	if got := violations(t, app3.Execute(checkPipeline(cluster))); !reflect.DeepEqual(got, []string{want}) {
		t.Fatalf("replace-mode violations = %v\nwant                  %s", got, want)
	}

	// An override that only REUSES an embedded rule: fine layered, a named error in replace mode.
	p := filepath.Join(t.TempDir(), "reuse.gatwd")
	_ = os.WriteFile(p, []byte(`def probe = \_ g -> rule_idle_deployment g
def compliance = k8s_graph |> fold probe 0
`), 0o644)
	app2, err := newAppliance(nil, p, cli.OverrideLayer)
	if err != nil {
		t.Fatal(err)
	}
	if got := violations(t, app2.Execute(checkPipeline(cluster))); !reflect.DeepEqual(got, []string{"[[IdleDeployment worker]]"}) {
		t.Fatalf("layered reuse of rule_idle_deployment = %v", got)
	}
	app4, err := newAppliance(nil, p, cli.OverrideReplace)
	if err != nil {
		t.Fatal(err)
	}
	if out := fmt.Sprint(app4.Execute(checkPipeline(cluster))); !strings.Contains(out, "rule_idle_deployment") {
		t.Fatalf("replace mode: want an error naming the missing embedded rule, got %s", out)
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
		if obj.GetKind() == "Deployment" {
			// The container's args must be a valid command line of the operator's own tree
			// (long flags, a subcommand), so the sample cannot drift from the binary.
			containers, _, _ := unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
			rawArgs, _, _ := unstructured.NestedStringSlice(containers[0].(map[string]any), "args")
			cmd, rest, err := operatorTree().Root().Find(rawArgs)
			if err != nil || cmd.Name() != "run" {
				t.Fatalf("deploy.yaml args %v: command %v, err %v; want the run subcommand", rawArgs, cmd.Name(), err)
			}
			if err := cmd.ParseFlags(rest); err != nil {
				t.Fatalf("deploy.yaml args %v are not valid operator flags: %v", rawArgs, err)
			}
		}
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
