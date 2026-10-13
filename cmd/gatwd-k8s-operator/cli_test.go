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
	idleJSON   = `{"status":"violations","count":2,"violations":[{"rule":"UnselectedPod","resource":"Pod/default/stray"},{"rule":"IdleDeployment","resource":"Deployment/default/lonely"}],"errors":[],"diagnostics":[]}` + "\n"
	brokenJSON = `{"status":"violations","count":5,"violations":[{"rule":"UnselectedPod","resource":"Pod/default/orphan"},{"rule":"UnusedService","resource":"Service/default/unused-svc"},{"rule":"DanglingIngress","resource":"Ingress/default/front"},{"rule":"MissingConfigMap","resource":"Pod/default/orphan"},{"rule":"DependencyCycle","resource":["Deployment/default/loop-a","Deployment/default/loop-b"]}],"errors":[],"diagnostics":[]}` + "\n"
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
		{"json output, compliant", []string{"check", "--manifests", "testdata/compliant.yaml", "--rules", absent, "--format", "json"}, 0, `{"status":"ok","count":0,"violations":[],"errors":[],"diagnostics":[]}` + "\n", nil},
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

// TestOperatorOverridePolicy drives the real binary through the execution policy (W7): an
// override is a script that did not come from the command line, so it may only compute. An
// OS command in it is an error unless the process is started with --allow-exec, and nothing
// inside the file can grant itself the capability.
func TestOperatorOverridePolicy(t *testing.T) {
	bin := buildOperator(t)
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	// a rule set whose entry point runs a program
	execRule := write("exec.gatwd", "def compliance = echo hello-from-override\n")
	// the same, after trying every spelling of "I may run programs"
	widening := write("widen.gatwd", "const AllowExec = true\nconst Policy = TrustedPolicy\ndef allow_exec = true\ndef compliance = echo hello-from-override\n")
	// a top-level statement: it would run while the override loads
	atLoad := write("atload.gatwd", "spawn StrVal: x |> echo ran-at-load\n")
	pure := write("pure.gatwd", "def probe = \\_ g -> rule_idle_deployment g\ndef compliance = k8s_graph |> fold probe 0\n")

	const denied = "[ERROR] step 'echo' requires the exec capability (policy denies it)"
	rows := []struct {
		name      string
		args      []string
		code      int
		stdoutHas []string
		stdoutNot []string
		stderrHas []string
	}{
		{"an override that runs a program is denied by default",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--rules", execRule}, 1,
			[]string{denied}, []string{"hello-from-override"}, nil},
		{"--allow-exec grants it",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--rules", execRule, "--allow-exec"}, 0,
			[]string{"hello-from-override"}, []string{"[ERROR]"}, nil},
		{"the other --allow flags do not grant exec",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--rules", execRule, "--allow-fs", "--allow-net", "--allow-env"}, 1,
			[]string{denied}, []string{"hello-from-override"}, nil},
		{"an override cannot widen the policy from inside the file",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--rules", widening}, 1,
			[]string{denied}, []string{"hello-from-override"}, nil},
		{"replace mode is no way around it",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--rules", execRule, "--rules-mode", "replace"}, 1,
			[]string{denied}, []string{"hello-from-override"}, nil},
		{"a top-level statement stops the start",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--rules", atLoad}, 1,
			nil, []string{"ran-at-load"}, []string{atLoad, "step 'echo' requires the exec capability (policy denies it)"}},
		{"a top-level statement runs once the process allows it",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--rules", atLoad, "--allow-exec"}, 0,
			[]string{"UnselectedPod"}, []string{"[ERROR]", "ran-at-load"}, nil},
		{"a pure override needs no capability and is unaffected",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--rules", pure}, 0,
			[]string{"[[IdleDeployment Deployment/default/lonely]]"}, []string{"[ERROR]"}, nil},
		{"the embedded rules over the offline source need no capability",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--rules", filepath.Join(dir, "absent.gatwd")}, 0,
			[]string{idleWant[:len(idleWant)-1]}, []string{"[ERROR]"}, nil},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			code, stdout, stderr := operator(t, bin, r.args...)
			if code != r.code {
				t.Fatalf("exit %d, want %d\nstdout: %s\nstderr: %s", code, r.code, stdout, stderr)
			}
			for _, want := range r.stdoutHas {
				if !strings.Contains(stdout, want) {
					t.Errorf("stdout lacks %q:\n%s", want, stdout)
				}
			}
			for _, bad := range r.stdoutNot {
				if strings.Contains(stdout, bad) {
					t.Errorf("stdout must not contain %q:\n%s", bad, stdout)
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

// TestOperatorHelpListsThePolicyFlags keeps the flags discoverable: each is off by default.
func TestOperatorHelpListsThePolicyFlags(t *testing.T) {
	bin := buildOperator(t)
	code, stdout, stderr := operator(t, bin, "check", "--help")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	for _, f := range []string{"--allow-exec", "--allow-fs", "--allow-net", "--allow-env"} {
		if !strings.Contains(stdout, f) || !strings.Contains(stdout, "denied by default") {
			t.Errorf("--help does not describe %s:\n%s", f, stdout)
		}
	}
}

// TestOperatorJSONDiagnostics: a denied capability reaches a log pipeline as a structured
// entry (kind, step, capability) next to the unchanged free-text `errors`.
func TestOperatorJSONDiagnostics(t *testing.T) {
	bin := buildOperator(t)
	override := filepath.Join(t.TempDir(), "exec.gatwd")
	if err := os.WriteFile(override, []byte("def compliance = echo hello-from-override\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := operator(t, bin, "check", "--manifests", "testdata/idle.yaml", "--rules", override, "--format", "json")
	if code != 1 {
		t.Fatalf("exit %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}
	for _, want := range []string{
		`"status":"error"`,
		`"errors":["step 'echo' requires the exec capability (policy denies it)"]`,
		`"diagnostics":[{"kind":"capability_denied","message":"step 'echo' requires the exec capability (policy denies it)","step":"echo","capability":"exec"}]`,
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout lacks %s:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "hello-from-override") {
		t.Errorf("the denied command ran:\n%s", stdout)
	}
}

// TestOperatorCheckLimits drives the real binary through --timeout and --max-output. The slow
// check is a real `sleep` run by an override (so --allow-exec); a check over a limit is a
// failed check (exit 1, never 3), loud in text and JSON, and the limits are off by default.
func TestOperatorCheckLimits(t *testing.T) {
	bin := buildOperator(t)
	slow := filepath.Join(t.TempDir(), "slow.gatwd")
	if err := os.WriteFile(slow, []byte("def compliance = sleep 20\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	const timeoutLine = "[ERROR] check timed out after 200ms"
	rows := []struct {
		name      string
		args      []string
		code      int
		stdout    string // exact
		stderrHas string
	}{
		{"a slow check is stopped and exits 1",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--rules", slow, "--allow-exec", "--timeout", "200ms"}, 1,
			timeoutLine + "\n", "the check failed"},
		{"under --fail-on-violation a timeout is still 1, not 3",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--rules", slow, "--allow-exec", "--timeout", "200ms", "--fail-on-violation"}, 1,
			timeoutLine + "\n", ""},
		{"json: errors and a structured timeout diagnostic",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--rules", slow, "--allow-exec", "--timeout", "200ms", "--format", "json"}, 1,
			`{"status":"error","count":0,"violations":[],"errors":["check timed out after 200ms"],"diagnostics":[{"kind":"timeout","message":"check timed out after 200ms","limit":"200ms"}]}` + "\n", ""},
		{"a fast check under a generous timeout is unaffected",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--timeout", "5m", "--max-output", "1048576"}, 0, idleWant, ""},
		{"output over the limit is an error, not a truncated report",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--max-output", "40"}, 1,
			"[ERROR] check output exceeded the limit of 40 bytes\n", "the check failed"},
		{"json: an output_limit diagnostic",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--max-output", "40", "--format", "json"}, 1,
			`{"status":"error","count":0,"violations":[],"errors":["check output exceeded the limit of 40 bytes"],"diagnostics":[{"kind":"output_limit","message":"check output exceeded the limit of 40 bytes","limit":"40"}]}` + "\n", ""},
		{"a limit as large as the report (79 bytes with its newline) passes",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--max-output", "79"}, 0, idleWant, ""},
		{"one byte less fails",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--max-output", "78"}, 1,
			"[ERROR] check output exceeded the limit of 78 bytes\n", "the check failed"},
		{"a negative timeout is a usage error",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--timeout", "-1s"}, 2, "", "--timeout must not be negative"},
		{"a negative output limit is a usage error",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--max-output=-1"}, 2, "", "--max-output must not be negative"},
		{"the limits are off by default, and 0 turns them off explicitly",
			[]string{"check", "--manifests", "testdata/idle.yaml", "--timeout", "0", "--max-output", "0"}, 0, idleWant, ""},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			code, stdout, stderr := operator(t, bin, r.args...)
			if code != r.code {
				t.Fatalf("exit %d, want %d\nstdout: %s\nstderr: %s", code, r.code, stdout, stderr)
			}
			if stdout != r.stdout {
				t.Errorf("stdout = %q\nwant     %q", stdout, r.stdout)
			}
			if !strings.Contains(stderr, r.stderrHas) {
				t.Errorf("stderr lacks %q:\n%s", r.stderrHas, stderr)
			}
		})
	}
}

// TestOperatorRunLoopSurvivesATimeout: every iteration of `run` gets its own timeout, a timed-out
// check is reported on stdout like any failing check, and the loop goes on with the next one.
func TestOperatorRunLoopSurvivesATimeout(t *testing.T) {
	bin := buildOperator(t)
	slow := filepath.Join(t.TempDir(), "slow.gatwd")
	if err := os.WriteFile(slow, []byte("def compliance = sleep 20\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := exec.Command(bin, "run", "--manifests", "testdata/idle.yaml", "--rules", slow, "--allow-exec", "--timeout", "200ms", "--interval", "50ms")
	c.Env = append(os.Environ(), "KUBECONFIG="+filepath.Join(t.TempDir(), "absent"), "HOME="+t.TempDir())
	stdout, err := c.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Process.Kill(); _ = c.Wait() }()

	lines := make(chan string, 8)
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()
	for i := 0; i < 3; i++ {
		select {
		case l, ok := <-lines:
			if !ok || l != "[ERROR] check timed out after 200ms" {
				t.Fatalf("check %d: line %q (open=%v); each iteration must report its own timeout and nothing else", i+1, l, ok)
			}
		case <-time.After(30 * time.Second):
			t.Fatalf("check %d: no output; the loop stopped after a timeout", i+1)
		}
	}
}

// TestOperatorHelpListsTheLimitFlags keeps --timeout and --max-output discoverable.
func TestOperatorHelpListsTheLimitFlags(t *testing.T) {
	bin := buildOperator(t)
	code, stdout, stderr := operator(t, bin, "check", "--help")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	for _, f := range []string{"--timeout DUR", "--max-output BYTES"} {
		if !strings.Contains(stdout, f) {
			t.Errorf("--help does not describe %s:\n%s", f, stdout)
		}
	}
}
