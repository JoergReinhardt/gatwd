package main

import (
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

// TestEditionCommandLine drives the real binary: the embedded prelude is listed and
// printable, --no-prelude removes it while the offline steps stay registered, and the
// generated completion and version commands work.
func TestEditionCommandLine(t *testing.T) {
	bin := buildEdition(t)
	cluster := filepath.Join("..", "..", "examples", "k8s_howto", "cluster")
	graph := "spawn StrVal: " + cluster + " |> manifests |> k8s_graph |> graph_order"

	out, err := exec.Command(bin, "prelude", "list").Output()
	if err != nil || strings.TrimSpace(string(out)) != "lib/k8s/prelude.gatwd" {
		t.Fatalf("prelude list: %v\n%s", err, out)
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
	// (no collect_list here: it drops error nodes, see docs/hardening_plan.md §8)
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
