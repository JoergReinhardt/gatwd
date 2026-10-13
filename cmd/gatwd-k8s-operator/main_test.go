package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
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
		"testdata/idle.yaml":                     "[[UnselectedPod stray] [IdleDeployment lonely]]",
		"../../pkg/k8sgraph/testdata/stack.yaml": "[[UnselectedPod other-pod]]",
	}
	for path, want := range cases {
		got := violations(t, app.Execute(checkPipeline(path)))
		if !reflect.DeepEqual(got, []string{want}) {
			t.Errorf("%s: violations = %v, want %s", path, got, want)
		}
	}
}

func TestOperatorHotReloadsRules(t *testing.T) {
	override := filepath.Join(t.TempDir(), "rules.gatwd")
	app, err := newAppliance(nil, override)
	if err != nil {
		t.Fatal(err)
	}

	strict := "def compliance = k8s_graph |> fold (\\_ g -> [ [\"AnyPod\", (hamt_path \"metadata.name\" p)] | p <- (graph_payloads_of \"Pod\" g) ]) 0\n"
	if err := os.WriteFile(override, []byte(strict), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := app.Reload(); err != nil {
		t.Fatal(err)
	}
	if got := violations(t, app.Execute(checkPipeline("testdata/idle.yaml"))); !reflect.DeepEqual(got, []string{"[[AnyPod stray]]"}) {
		t.Fatalf("after reload, violations = %v", got)
	}

	if err := os.Remove(override); err != nil {
		t.Fatal(err)
	}
	if err := app.Reload(); err != nil {
		t.Fatal(err)
	}
	if got := violations(t, app.Execute(checkPipeline("testdata/idle.yaml"))); !reflect.DeepEqual(got, []string{"[[UnselectedPod stray] [IdleDeployment lonely]]"}) {
		t.Fatalf("after override removal, violations = %v", got)
	}
}
