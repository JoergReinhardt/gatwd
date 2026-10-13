package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
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

// TestHowToTasks: the task definitions in tasks.gatwd (everything before its first
// invocation) work when evaluated from a script, and select-one-task-at-a-time works by
// running an invocation against the loaded definitions.
func TestHowToTasks(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(howtoDir, "tasks.gatwd"))
	if err != nil {
		t.Fatal(err)
	}
	defs, _, _ := strings.Cut(string(raw), "\nspawn ")
	i := howtoEdition()
	i.EvaluateScript(defs)
	cluster := filepath.Join(howtoDir, "cluster")

	if got := one(t, i, "spawn StrVal: "+cluster+" |> trace_to_pod"); got != "[nginx-svc app-pod]" {
		t.Errorf("trace_to_pod = %s", got)
	}
	if got := one(t, i, "spawn StrVal: "+cluster+" |> who_configures_what"); !strings.Contains(got, "app-config-Configures->worker") {
		t.Errorf("who_configures_what = %s", got)
	}
	if got := one(t, i, "spawn StrVal: "+cluster+` |> graph |> fold (\_ g -> count_kind "Pod" g) 0`); got != "1" {
		t.Errorf("count_kind Pod = %s", got)
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

// TestHowToRulesOverride: an override file replaces the embedded rules wholesale; the
// shipped example keeps all six embedded rules and adds one.
func TestHowToRulesOverride(t *testing.T) {
	app, err := newAppliance(nil, filepath.Join(howtoDir, "rules.gatwd"))
	if err != nil {
		t.Fatal(err)
	}
	got := violations(t, app.Execute(checkPipeline(filepath.Join(howtoDir, "cluster"))))
	want := "[[IdleDeployment worker] [UnusedService backend-svc] [MissingConfigMap worker] [DatabaseWithoutCredentials audit]]"
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
