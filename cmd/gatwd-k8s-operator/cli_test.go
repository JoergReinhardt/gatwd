package main

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// buildOperator builds the real binary once per test.
func buildOperator(t *testing.T) string {
	t.Helper()
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not on PATH")
	}
	bin := filepath.Join(t.TempDir(), "gatwd-k8s-operator")
	if out, err := exec.Command(goBin, "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

// operator runs the binary with no reachable cluster and returns exit code, stdout, stderr.
func operator(t *testing.T, bin string, args ...string) (int, string, string) {
	t.Helper()
	c := exec.Command(bin, args...)
	c.Env = append(os.Environ(), "KUBECONFIG="+filepath.Join(t.TempDir(), "absent"), "HOME="+t.TempDir(), "KUBERNETES_SERVICE_HOST=")
	var out, errb bytes.Buffer
	c.Stdout, c.Stderr = &out, &errb
	err := c.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return code, out.String(), errb.String()
}

const (
	idleWant   = "[[UnselectedPod Pod/default/stray] [IdleDeployment Deployment/default/lonely]]\n"
	idleJSON   = `{"status":"violations","count":2,"violations":[{"rule":"UnselectedPod","resource":"Pod/default/stray"},{"rule":"IdleDeployment","resource":"Deployment/default/lonely"}],"errors":[]}` + "\n"
	brokenJSON = `{"status":"violations","count":5,"violations":[{"rule":"UnselectedPod","resource":"Pod/default/orphan"},{"rule":"UnusedService","resource":"Service/default/unused-svc"},{"rule":"DanglingIngress","resource":"Ingress/default/front"},{"rule":"MissingConfigMap","resource":"Pod/default/orphan"},{"rule":"DependencyCycle","resource":["Deployment/default/loop-a","Deployment/default/loop-b"]}],"errors":[]}` + "\n"
)

// TestOperatorCommandLine drives the real binary through every command line the how-to
// documents: the subcommands, the legacy single-dash spelling (with its notice), the layered
// override and its report, and the exit statuses 0, 1, 2 and 3.
func TestOperatorCommandLine(t *testing.T) {
	bin := buildOperator(t)
	dir := t.TempDir()
	layered := filepath.Join(dir, "layered.gatwd")
	_ = os.WriteFile(layered, []byte("def probe = \\_ g -> rule_idle_deployment g\ndef compliance = k8s_graph |> fold probe 0\n"), 0o644)
	broken := filepath.Join(dir, "broken.gatwd")
	_ = os.WriteFile(broken, []byte("def compliance = k8s_graph |> fold (\n"), 0o644)
	failing := filepath.Join(dir, "failing.gatwd")
	_ = os.WriteFile(failing, []byte("def compliance = k8s_graph |> fold (\\_ g -> no_such_rule g) 0\n"), 0o644)
	absent := filepath.Join(dir, "absent.gatwd")

	rows := []struct {
		name      string
		args      []string
		code      int
		stdout    string
		stderrHas []string
	}{
		{"check", []string{"check", "--manifests", "testdata/idle.yaml", "--rules", absent}, 0, idleWant, []string{"rules: embedded rules, no override"}},
		{"legacy form, long flags", []string{"--manifests", "testdata/idle.yaml", "--rules", absent}, 0, idleWant, nil},
		{"legacy single-dash flags (deprecated)", []string{"-manifests", "testdata/idle.yaml", "-rules", absent}, 0, idleWant,
			[]string{"-manifests is deprecated; use --manifests", "-rules is deprecated; use --rules"}},
		{"a layered override reuses an embedded rule", []string{"check", "--manifests", "testdata/idle.yaml", "--rules", layered}, 0, "[[IdleDeployment Deployment/default/lonely]]\n",
			[]string{"layered on the embedded rules", "overridden: compliance", "added: probe"}},
		{"replace mode cannot reach the embedded rule", []string{"check", "--manifests", "testdata/idle.yaml", "--rules", layered, "--rules-mode", "replace"}, 1, "",
			[]string{"REPLACES"}},
		{"a broken override fails the start", []string{"check", "--manifests", "testdata/idle.yaml", "--rules", broken}, 1, "", []string{broken, "unbalanced delimiter"}},
		{"violations with --fail-on-violation exit 3", []string{"check", "--manifests", "testdata/idle.yaml", "--rules", absent, "--fail-on-violation"}, 3, idleWant, []string{"2 violation(s) found"}},
		{"a compliant cluster exits 0 with --fail-on-violation", []string{"check", "--manifests", "testdata/compliant.yaml", "--rules", absent, "--fail-on-violation"}, 0, "[]\n", nil},
		{"a duplicate id is a failed check (1), not a violation (3)", []string{"check", "--manifests", "../../pkg/k8sgraph/testdata/duplicates.yaml", "--rules", absent, "--fail-on-violation"}, 1, "", []string{"the check failed"}},
		{"json output, violations", []string{"check", "--manifests", "testdata/idle.yaml", "--rules", absent, "--format", "json"}, 0, idleJSON, nil},
		{"json output, every rule violated", []string{"check", "--manifests", "testdata/broken.yaml", "--rules", absent, "--format", "json", "--fail-on-violation"}, 3, brokenJSON, []string{"5 violation(s) found"}},
		{"json output, compliant", []string{"check", "--manifests", "testdata/compliant.yaml", "--rules", absent, "--format", "json"}, 0, `{"status":"ok","count":0,"violations":[],"errors":[]}` + "\n", nil},
		{"a rule that errors at run time fails the check", []string{"check", "--manifests", "testdata/idle.yaml", "--rules", failing, "--format", "json"}, 1, "", []string{"the check failed"}},
		{"no source and no cluster", []string{"check", "--rules", absent}, 2, "", []string{"no --manifests given and no cluster reachable"}},
		{"unknown flag", []string{"check", "--manifests", "x", "--nope"}, 2, "", []string{"unknown flag"}},
		{"unknown format", []string{"check", "--manifests", "testdata/idle.yaml", "--format", "yaml"}, 2, "", []string{"text or json"}},
		{"unknown rules mode", []string{"check", "--manifests", "testdata/idle.yaml", "--rules-mode", "merge"}, 2, "", []string{"layer or replace"}},
		{"run without an interval", []string{"run", "--manifests", "testdata/idle.yaml"}, 2, "", []string{"--interval"}},
		{"rules cat", []string{"rules", "cat"}, 0, "", nil},
		{"version", []string{"version"}, 0, "", nil},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			code, stdout, stderr := operator(t, bin, r.args...)
			if code != r.code {
				t.Fatalf("exit %d, want %d\nstdout: %s\nstderr: %s", code, r.code, stdout, stderr)
			}
			switch r.name {
			case "rules cat":
				if !strings.Contains(stdout, "def compliance") || !strings.Contains(stdout, "rule_dependency_cycle") {
					t.Errorf("rules cat printed %q", stdout)
				}
			case "a rule that errors at run time fails the check":
				if !strings.Contains(stdout, `"status":"error"`) || !strings.Contains(stdout, "no_such_rule") {
					t.Errorf("stdout = %q, want an error status naming the failing rule", stdout)
				}
			case "a duplicate id is a failed check (1), not a violation (3)":
				for _, want := range []string{`duplicate node id "Pod/default/dup"`, "duplicates.yaml, document 1 (", "duplicates.yaml, document 2 ("} {
					if !strings.Contains(stdout, want) {
						t.Errorf("stdout = %q, want the duplicate-id error naming %q", stdout, want)
					}
				}
			case "version":
				if !strings.HasPrefix(stdout, "gatwd-k8s-operator ") {
					t.Errorf("version printed %q", stdout)
				}
			case "replace mode cannot reach the embedded rule":
				if !strings.Contains(stdout, "rule_idle_deployment") {
					t.Errorf("stdout = %q, want an error naming the missing embedded rule", stdout)
				}
			default:
				if stdout != r.stdout {
					t.Errorf("stdout = %q, want %q", stdout, r.stdout)
				}
			}
			for _, want := range r.stderrHas {
				if !strings.Contains(stderr, want) {
					t.Errorf("stderr lacks %q:\n%s", want, stderr)
				}
			}
		})
	}
}

// TestOperatorRunLoop: `run --interval` keeps checking and exits only when stopped.
func TestOperatorRunLoop(t *testing.T) {
	bin := buildOperator(t)
	c := exec.Command(bin, "run", "--manifests", "testdata/idle.yaml", "--rules", filepath.Join(t.TempDir(), "absent"), "--interval", "50ms")
	c.Env = append(os.Environ(), "KUBECONFIG="+filepath.Join(t.TempDir(), "absent"), "HOME="+t.TempDir())
	stdout, err := c.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Process.Kill(); _ = c.Wait() }()

	lines := make(chan string, 4)
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()
	for i := 0; i < 2; i++ {
		select {
		case l, ok := <-lines:
			if !ok || l != strings.TrimSpace(idleWant) {
				t.Fatalf("check %d: line %q (open=%v)", i+1, l, ok)
			}
		case <-time.After(30 * time.Second):
			t.Fatalf("check %d: no output; the loop did not run again", i+1)
		}
	}
}

// TestOperatorNoKustomize drives the real binary through the --no-kustomize switch. An
// override that prints the node ids shows what was read: the rendered overlay (the name
// prefix applied) or, with the flag, the plain files of the kustomization root.
func TestOperatorNoKustomize(t *testing.T) {
	bin := buildOperator(t)
	ids := filepath.Join(t.TempDir(), "ids.gatwd")
	_ = os.WriteFile(ids, []byte("def compliance = k8s_graph |> k8s_apply_order |> collect_list\n"), 0o644)
	root := "testdata/kustomize-root"

	rows := []struct {
		name      string
		args      []string
		code      int
		stdout    string
		stderrHas string
	}{
		{"a kustomization root is rendered by default", []string{"check", "--manifests", root, "--rules", ids}, 0, "[Pod/default/x-web]\n", ""},
		{"--no-kustomize reads the plain files", []string{"check", "--manifests", root, "--rules", ids, "--no-kustomize"}, 0, "[Kustomization/_/k Pod/default/web]\n", ""},
		{"--no-kustomize=false renders", []string{"check", "--manifests", root, "--rules", ids, "--no-kustomize=false"}, 0, "[Pod/default/x-web]\n", ""},
		{"the flag works after the subcommand and in the legacy form", []string{"--manifests", root, "--rules", ids, "--no-kustomize"}, 0, "[Kustomization/_/k Pod/default/web]\n", ""},
		{"with the live source the flag would do nothing, so it is refused", []string{"check", "--no-kustomize", "--rules", ids}, 2, "", "--no-kustomize applies to --manifests"},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			code, stdout, stderr := operator(t, bin, r.args...)
			if code != r.code || stdout != r.stdout || !strings.Contains(stderr, r.stderrHas) {
				t.Fatalf("exit %d (want %d)\nstdout: %q (want %q)\nstderr: %s (want %q)", code, r.code, stdout, r.stdout, stderr, r.stderrHas)
			}
		})
	}

	if _, out, _ := operator(t, bin, "check", "--help"); !strings.Contains(out, "--no-kustomize") {
		t.Errorf("--help must document --no-kustomize:\n%s", out)
	}
}
