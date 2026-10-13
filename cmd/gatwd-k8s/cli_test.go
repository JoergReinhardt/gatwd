package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildEdition(t *testing.T) string {
	t.Helper()
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not on PATH")
	}
	bin := filepath.Join(t.TempDir(), "gatwd-k8s")
	if out, err := exec.Command(goBin, "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

// TestEditionCommandLine drives the real binary: the embedded prelude and the files it imports are listed and
// printable, --no-prelude removes it while the offline steps stay registered, and the
// generated completion and version commands work.
func TestEditionCommandLine(t *testing.T) {
	bin := buildEdition(t)
	cluster := filepath.Join("..", "..", "examples", "k8s_howto", "cluster")
	graph := "spawn StrVal: " + cluster + " |> manifests |> k8s_graph |> graph_order"

	out, err := exec.Command(bin, "prelude", "list").Output()
	// the entry comes first, then the files it imports (so they can be read with `prelude cat`)
	listed := strings.Fields(string(out))
	if err != nil || len(listed) < 2 || listed[0] != "lib/k8s/prelude.gatwd" || !strings.Contains(string(out), "lib/k8s/relations.gatwd") {
		t.Fatalf("prelude list: %v\n%s", err, out)
	}
	out, err = exec.Command(bin, "prelude", "cat", "lib/k8s/relations.gatwd").Output()
	if err != nil || !strings.Contains(string(out), "def rel_binds") {
		t.Fatalf("prelude cat of an imported file: %v\n%s", err, out)
	}
	out, err = exec.Command(bin, "prelude", "cat", "lib/k8s/prelude.gatwd").Output()
	if err != nil || !strings.Contains(string(out), "def k8s_graph") {
		t.Fatalf("prelude cat: %v\n%s", err, out)
	}

	out, err = exec.Command(bin, "eval", "-e", graph).Output()
	if err != nil || !strings.Contains(string(out), "∷ ") || strings.Contains(string(out), "[ERROR]") {
		t.Fatalf("eval with the prelude: %v\n%s", err, out)
	}
	out, err = exec.Command(bin, "eval", "--no-prelude", "-e", graph).Output()
	if err == nil || !strings.Contains(string(out), "[ERROR]") {
		t.Fatalf("--no-prelude must leave k8s_graph undefined (exit non-zero): %v\n%s", err, out)
	}
	out, err = exec.Command(bin, "eval", "--no-prelude", "-e", "spawn StrVal: "+cluster+" |> manifests |> collect_list").Output()
	if err != nil || strings.Contains(string(out), "[ERROR]") {
		t.Fatalf("--no-prelude must keep the manifests step: %v\n%s", err, out)
	}

	for _, sh := range []string{"bash", "zsh", "fish", "powershell"} {
		out, err = exec.Command(bin, "completion", sh).Output()
		if err != nil || len(out) < 100 {
			t.Errorf("completion %s: %v (%d bytes)", sh, err, len(out))
		}
	}
	if out, err = exec.Command(bin, "version").Output(); err != nil || !strings.HasPrefix(string(out), "gatwd-k8s ") {
		t.Errorf("version: %v\n%s", err, out)
	}
}

// TestEditionKustomize drives the real binary through the in-process kustomize path: no
// kustomize, kubectl or git on PATH, `manifests` renders an overlay on its own,
// --no-kustomize turns that off, and a broken overlay is a loud error naming the root.
func TestEditionKustomize(t *testing.T) {
	bin := buildEdition(t)
	root := filepath.Join("..", "..", "pkg", "k8sgraph", "testdata", "kustomize", "overlays")
	count := func(overlay string, flags ...string) (string, error) {
		argv := append(append([]string{}, flags...), "eval", "-e", "spawn StrVal: "+filepath.Join(root, overlay)+" |> manifests |> collect_list |> length")
		cmd := exec.Command(bin, argv...)
		cmd.Env = []string{"PATH=" + t.TempDir(), "HOME=" + t.TempDir()}
		out, err := cmd.Output()
		return strings.TrimSpace(string(out)), err
	}

	if out, err := count("prod"); err != nil || out != "∷ 4" {
		t.Fatalf("manifests on an overlay (auto-render): %v, %q; want 4 rendered resources", err, out)
	}
	if out, err := count("prod", "--no-kustomize"); err != nil || out != "∷ 3" {
		t.Fatalf("--no-kustomize: %v, %q; want the 3 plain files", err, out)
	}
	cmd := exec.Command(bin, "eval", "-e", "spawn StrVal: "+filepath.Join(root, "broken")+" |> manifests |> k8s_graph |> graph_edges")
	cmd.Env = []string{"PATH=" + t.TempDir(), "HOME=" + t.TempDir()}
	raw, err := cmd.Output()
	out := string(raw)
	if err == nil || !strings.Contains(out, "[ERROR]") || !strings.Contains(out, "overlays/broken") || !strings.Contains(out, "does-not-exist.yaml") {
		t.Fatalf("a broken overlay must fail loudly, naming the root and the file: %v, %q", err, out)
	}
	if help, err := exec.Command(bin, "--help").CombinedOutput(); err != nil || !strings.Contains(string(help), "--no-kustomize") {
		t.Errorf("--help must list --no-kustomize: %v\n%s", err, help)
	}
}

// TestEditionClientIsLazy: commands that need no cluster neither look for a kubeconfig nor
// print the old "no cluster access" notice; a crawl with no cluster is an error that names
// the cause (and exits non-zero), not a step that silently does not exist.
func TestEditionClientIsLazy(t *testing.T) {
	bin := buildEdition(t)
	noCluster := func(args ...string) *exec.Cmd {
		c := exec.Command(bin, args...)
		c.Env = append(os.Environ(), "KUBECONFIG="+filepath.Join(t.TempDir(), "absent"), "HOME="+t.TempDir(), "KUBERNETES_SERVICE_HOST=")
		return c
	}
	for _, args := range [][]string{{"version"}, {"completion", "bash"}, {"prelude", "list"}, {"eval", "-e", "1 + 1"}} {
		var stderr strings.Builder
		c := noCluster(args...)
		c.Stderr = &stderr
		if err := c.Run(); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if stderr.Len() != 0 {
			t.Errorf("%v wrote to stderr (the cluster client must be lazy): %q", args, stderr.String())
		}
	}

	var stderr strings.Builder
	c := noCluster("eval", "-e", "spawn StrVal: default |> k8s_pods")
	c.Stderr = &stderr
	out, err := c.Output()
	if err == nil || !strings.Contains(string(out), "no cluster access") {
		t.Fatalf("a crawl without a cluster must fail naming the cause: err=%v\nstdout: %s\nstderr: %s", err, out, stderr.String())
	}
}

// TestEditionWritesAreOptIn drives the real binary through the three layers that guard a
// write to the cluster. The first two need no cluster; the third (RBAC) is the cluster's.
//  1. the capability policy: k8s_apply is denied for the user's OWN script until
//     --allow-cluster-apply, and the rollout steps until --allow-cluster-rollout;
//  2. server dry run: with no --dry-run=false nothing is persisted (covered against a fake
//     cluster in pkg/k8sgraph; here, that the flag exists and the offline twin works);
//  3. the offline twin: --out DIR renders YAML files and needs no cluster.
func TestEditionWritesAreOptIn(t *testing.T) {
	bin := buildEdition(t)
	cluster := filepath.Join("..", "..", "examples", "k8s_howto", "cluster")
	apply := "spawn StrVal: " + cluster + " |> manifests |> k8s_apply"
	run := func(args ...string) (int, string) {
		c := exec.Command(bin, args...)
		c.Env = append(os.Environ(), "KUBECONFIG="+filepath.Join(t.TempDir(), "absent"), "HOME="+t.TempDir(), "KUBERNETES_SERVICE_HOST=")
		out, err := c.CombinedOutput()
		code := 0
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else if err != nil {
			t.Fatal(err)
		}
		return code, string(out)
	}

	out := filepath.Join(t.TempDir(), "rendered")
	rows := []struct {
		name string
		args []string
		ok   bool // exit status 0
		has  []string
		not  []string
	}{
		{"the user's own script is denied by default",
			[]string{"eval", "-e", apply}, false,
			[]string{"step 'k8s_apply' requires the cluster-apply capability (policy denies it)"}, []string{"written", "created"}},
		{"--out does not lift the denial",
			[]string{"--out", out, "eval", "-e", apply}, false,
			[]string{"requires the cluster-apply capability"}, []string{"written"}},
		{"the rollout capability does not grant apply",
			[]string{"--allow-cluster-rollout", "eval", "-e", apply}, false,
			[]string{"requires the cluster-apply capability"}, nil},
		{"a rollout step is denied by default",
			[]string{"eval", "-e", "spawn StrVal: Deployment/default/web |> k8s_scale 3"}, false,
			[]string{"step 'k8s_scale' requires the cluster-rollout capability (policy denies it)"}, nil},
		{"the apply capability does not grant rollout",
			[]string{"--allow-cluster-apply", "eval", "-e", "spawn StrVal: Deployment/default/web |> k8s_rollout_restart"}, false,
			[]string{"step 'k8s_rollout_restart' requires the cluster-rollout capability (policy denies it)"}, nil},
		{"rollout status is a read: no capability, so it reaches the (absent) cluster",
			[]string{"eval", "-e", "spawn StrVal: Deployment/default/web |> k8s_rollout_status"}, false,
			[]string{"no cluster access"}, []string{"requires the"}},
		{"allowed but with no cluster: a loud error, not a skipped write",
			[]string{"--allow-cluster-apply", "eval", "-e", apply}, false,
			[]string{"no cluster access"}, []string{"written", "created", "would-"}},
		{"the offline twin renders files with no cluster",
			[]string{"--allow-cluster-apply", "--out", out, "eval", "-e", apply}, true,
			[]string{"written"}, []string{"[ERROR]"}},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			code, o := run(r.args...)
			if (code == 0) != r.ok {
				t.Fatalf("exit %d, want ok=%v\n%s", code, r.ok, o)
			}
			for _, w := range r.has {
				if !strings.Contains(o, w) {
					t.Errorf("output lacks %q:\n%s", w, o)
				}
			}
			for _, b := range r.not {
				if strings.Contains(o, b) {
					t.Errorf("output must not contain %q:\n%s", b, o)
				}
			}
		})
	}
	if files, _ := filepath.Glob(filepath.Join(out, "*.yaml")); len(files) == 0 {
		t.Fatal("the offline twin wrote no files")
	}

	code, help := run("--help")
	if code != 0 {
		t.Fatalf("--help: %s", help)
	}
	for _, f := range []string{"--allow-cluster-apply", "--allow-cluster-rollout", "--dry-run", "--out", "--force-conflicts", "--field-manager"} {
		if !strings.Contains(help, f) {
			t.Errorf("--help does not list %s:\n%s", f, help)
		}
	}
	if !strings.Contains(help, "denied by default") || !strings.Contains(help, "--dry-run=false") {
		t.Errorf("--help must say the writes are denied by default and how to make them real:\n%s", help)
	}
	if strings.Contains(help, "--prune") || strings.Contains(help, "--delete") {
		t.Errorf("the first version has no delete or prune:\n%s", help)
	}
}
