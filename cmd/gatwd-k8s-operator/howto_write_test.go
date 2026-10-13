package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

// The "Writing to the cluster" section of the how-to (§3.5) is pinned the same way as the rest
// (doc_test.go): every sample runs the REAL binaries, and its needles must be in the how-to text
// and in what the binary printed. The samples that need a cluster are pinned against a fake
// dynamic client in pkg/k8sgraph (write_test.go, rollout_test.go).

func TestHowToWriteDocSamples(t *testing.T) {
	edition, operatorBin := howtoBinaries(t)
	doc, err := os.ReadFile(howtoDoc)
	if err != nil {
		t.Fatal(err)
	}
	text := squash(string(doc))

	dir := t.TempDir()
	applyRule := filepath.Join(dir, "apply_rules.gatwd")
	mustWrite(t, applyRule, "def compliance = k8s_apply\n")
	out := filepath.Join(dir, "rendered")

	const prod = "pkg/k8sgraph/testdata/kustomize/overlays/prod"
	const cluster = "examples/k8s_howto/cluster"
	pipe := "spawn StrVal: " + prod + " |> manifests |> k8s_apply"
	rows := []docSample{
		{name: "k8s_apply is denied for your own script", which: "edition", code: 1,
			args: []string{"eval", "-e", pipe},
			docNeedles: []string{
				"gatwd-k8s eval -e 'spawn StrVal: " + prod + " |> manifests |> k8s_apply'",
				"∷ [ERROR] step 'k8s_apply' requires the cluster-apply capability (policy denies it)",
			},
			outNeedles: []string{"[ERROR] step 'k8s_apply' requires the cluster-apply capability (policy denies it)"}},
		{name: "the rollout steps are denied too, and status is not", which: "edition", code: 1,
			args:       []string{"eval", "-e", "spawn StrVal: Deployment/default/web |> k8s_scale 5"},
			docNeedles: []string{"∷ [ERROR] step 'k8s_scale' requires the cluster-rollout capability (policy denies it)"},
			outNeedles: []string{"[ERROR] step 'k8s_scale' requires the cluster-rollout capability (policy denies it)"}},
		{name: "rollout status is a read", which: "edition", code: 1,
			args:       []string{"eval", "-e", "spawn StrVal: Deployment/default/web |> k8s_rollout_status"},
			docNeedles: []string{"gatwd-k8s eval -e 'spawn StrVal: Deployment/default/web |> k8s_rollout_status'", "no flag: status only reads"},
			outNeedles: []string{"[ERROR] rollout status Deployment/default/web: no cluster access ("}},
		{name: "granted, no cluster, no --out: loud, per object", which: "edition", code: 1,
			args: []string{"--allow-cluster-apply", "eval", "-e", pipe},
			docNeedles: []string{
				"gatwd-k8s --allow-cluster-apply eval -e 'spawn StrVal: " + prod + " |> manifests |> k8s_apply'",
				"∷ [ERROR] cannot apply ConfigMap/prod/prod-web-config: reading the current object: no cluster access (",
			},
			outNeedles: []string{"∷ [ERROR] cannot apply ConfigMap/prod/prod-web-config: reading the current object: no cluster access ("}},
		{name: "the offline twin renders files in apply order", which: "edition",
			args: []string{"--allow-cluster-apply", "--out", out, "eval", "-e", "spawn StrVal: " + prod + " |> manifests |> k8s_graph |> k8s_apply_objects |> k8s_apply"},
			docNeedles: []string{
				"gatwd-k8s --allow-cluster-apply --out /tmp/rendered eval -e 'spawn StrVal: " + prod + " |> manifests |> k8s_graph |> k8s_apply_objects |> k8s_apply'",
				"∷ [apply ConfigMap/prod/prod-web-config written]",
				"∷ [apply Ingress/prod/prod-web-ing written]",
				"∷ [apply Service/prod/prod-web-svc written]",
				"∷ [apply Pod/prod/prod-web-pod written]",
			},
			outNeedles: []string{
				"∷ [apply ConfigMap/prod/prod-web-config written]\n∷ [apply Ingress/prod/prod-web-ing written]\n∷ [apply Service/prod/prod-web-svc written]\n∷ [apply Pod/prod/prod-web-pod written]",
			}},
		{name: "an override that applies is denied in the operator", which: "operator", code: 1,
			args: []string{"check", "--manifests", cluster, "--rules", applyRule},
			docNeedles: []string{
				"gatwd-k8s-operator check --manifests examples/k8s_howto/cluster --rules apply_rules.gatwd",
				"[ERROR] step 'k8s_apply' requires the cluster-apply capability (policy denies it)",
			},
			outNeedles: []string{"[ERROR] step 'k8s_apply' requires the cluster-apply capability (policy denies it)"}},
		{name: "the operator, with the capability, the fs capability and --out", which: "operator",
			args:       []string{"check", "--manifests", cluster, "--rules", applyRule, "--allow-cluster-apply", "--allow-fs", "--out", filepath.Join(dir, "operator-out")},
			docNeedles: []string{"--rules apply_rules.gatwd --allow-cluster-apply --allow-fs --out /tmp/rendered", "[apply ConfigMap/default/app-config written]"},
			outNeedles: []string{"[apply ConfigMap/default/app-config written]"}},
		{name: "the denial in the json report names the capability", which: "operator", code: 1,
			args:       []string{"check", "--manifests", cluster, "--rules", applyRule, "--format", "json"},
			docNeedles: []string{`"capability":"cluster-apply"`},
			outNeedles: []string{`"kind":"capability_denied","message":"step 'k8s_apply' requires the cluster-apply capability (policy denies it)","step":"k8s_apply","capability":"cluster-apply"`}},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			bin := edition
			if r.which == "operator" {
				bin = operatorBin
			}
			code, stdout, stderr := runIn(t, "../..", bin, "", r.args...)
			if code != r.code {
				t.Fatalf("exit %d, want %d\nstdout: %s\nstderr: %s", code, r.code, stdout, stderr)
			}
			got := stdout + stderr
			needles := r.outNeedles
			if len(needles) == 0 {
				needles = r.docNeedles
			}
			for _, want := range needles {
				if !strings.Contains(got, want) {
					t.Errorf("the binary no longer prints %q\nstdout: %s\nstderr: %s", want, stdout, stderr)
				}
			}
			for _, want := range r.docNeedles {
				if !strings.Contains(text, squash(want)) {
					t.Errorf("docs/k8s_howto.md no longer shows %q", want)
				}
			}
		})
	}
	// the rendered files are the ones the doc lists
	for _, f := range []string{"ConfigMap.prod.prod-web-config.yaml", "Ingress.prod.prod-web-ing.yaml", "Service.prod.prod-web-svc.yaml", "Pod.prod.prod-web-pod.yaml"} {
		if _, err := os.Stat(filepath.Join(out, f)); err != nil {
			t.Errorf("the offline twin did not write %s: %v", f, err)
		}
	}
	for _, claim := range []string{"deploy-write-role.yaml", "no step that deletes or prunes", "`--dry-run=false`", "`cluster-rollout`"} {
		if !strings.Contains(text, claim) {
			t.Errorf("the how-to's write section no longer says %q", claim)
		}
	}
}

// yamlDocs decodes every document of a manifest file.
func yamlDocs(t *testing.T, path string) []*unstructured.Unstructured {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	dec := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(raw), 4096)
	var out []*unstructured.Unstructured
	for {
		obj := &unstructured.Unstructured{}
		if err := dec.Decode(&obj.Object); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if len(obj.Object) > 0 {
			out = append(out, obj)
		}
	}
	return out
}

func verbsOf(t *testing.T, role *unstructured.Unstructured) []string {
	t.Helper()
	var all []string
	rules, _, _ := unstructured.NestedSlice(role.Object, "rules")
	for _, r := range rules {
		vs, _, _ := unstructured.NestedStringSlice(r.(map[string]any), "verbs")
		all = append(all, vs...)
	}
	return all
}

// TestWriteRoleIsSeparateAndNarrow: the shipped RBAC stays read-only, and the opt-in write
// example grants only what the write steps use: no delete, no update, no wildcard.
func TestWriteRoleIsSeparateAndNarrow(t *testing.T) {
	// the shipped manifest: get and list, nothing else
	for _, obj := range yamlDocs(t, filepath.Join(howtoDir, "deploy.yaml")) {
		if obj.GetKind() != "ClusterRole" && obj.GetKind() != "Role" {
			continue
		}
		for _, v := range verbsOf(t, obj) {
			if v != "get" && v != "list" {
				t.Errorf("the shipped %s %s grants %q: it must stay read-only", obj.GetKind(), obj.GetName(), v)
			}
		}
	}

	docs := yamlDocs(t, filepath.Join(howtoDir, "deploy-write-role.yaml"))
	var roles, bindings int
	for _, obj := range docs {
		switch obj.GetKind() {
		case "Role":
			roles++
			if obj.GetNamespace() == "" {
				t.Errorf("Role %s must be namespaced: prefer one namespace to a ClusterRole", obj.GetName())
			}
			got := verbsOf(t, obj)
			if len(got) == 0 {
				t.Errorf("Role %s grants nothing", obj.GetName())
			}
			for _, v := range got {
				if !slices.Contains([]string{"get", "list", "create", "patch"}, v) {
					t.Errorf("Role %s grants %q; the write steps need only get, list, create and patch", obj.GetName(), v)
				}
			}
			rules, _, _ := unstructured.NestedSlice(obj.Object, "rules")
			for _, r := range rules {
				res, _, _ := unstructured.NestedStringSlice(r.(map[string]any), "resources")
				for _, name := range res {
					if name == "secrets" || name == "*" {
						t.Errorf("Role %s names resource %q", obj.GetName(), name)
					}
				}
			}
		case "RoleBinding":
			bindings++
		case "ClusterRole", "ClusterRoleBinding":
			t.Errorf("the write example must not be cluster-wide: %s %s", obj.GetKind(), obj.GetName())
		}
	}
	if roles != 2 || bindings != 2 {
		t.Fatalf("roles %d, bindings %d; want the apply and rollout Roles and their bindings", roles, bindings)
	}
	// it is opt-in: nothing in the shipped Deployment grants a write flag
	for _, obj := range yamlDocs(t, filepath.Join(howtoDir, "deploy.yaml")) {
		if obj.GetKind() != "Deployment" {
			continue
		}
		containers, _, _ := unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
		args, _, _ := unstructured.NestedStringSlice(containers[0].(map[string]any), "args")
		for _, a := range args {
			if strings.HasPrefix(a, "--allow-cluster") || strings.HasPrefix(a, "--dry-run") {
				t.Errorf("the shipped Deployment passes %q: writing must stay opt-in", a)
			}
		}
	}
}
