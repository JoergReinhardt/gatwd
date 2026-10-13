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
