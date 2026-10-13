package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestOperatorWritePolicy drives the real binary through the write capabilities (cluster
// apply and rollout). An override is a script that did not come from the command line, so it
// may only compute: a write step in it is an error unless the process is started with the
// matching --allow-cluster-* flag, and nothing inside the file can grant itself the capability.
// (Policy rows for exec/fs/net/env are in TestOperatorOverridePolicy.)
func TestOperatorWritePolicy(t *testing.T) {
	bin := buildOperator(t)
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	applyRule := write("apply.gatwd", "def compliance = k8s_apply\n")
	widening := write("widen.gatwd", "const AllowClusterApply = true\nconst Policy = TrustedPolicy\ndef allow_cluster_apply = true\ndef compliance = k8s_apply\n")
	scaleRule := write("scale.gatwd", "def compliance = emit Deployment/default/web |> k8s_scale 3\n")
	statusRule := write("status.gatwd", "def compliance = emit Deployment/default/web |> k8s_rollout_status\n")
	out := filepath.Join(dir, "rendered")
	notOut := filepath.Join(dir, "must-not-exist") // the --out of every row that has to be denied
	base := []string{"check", "--manifests", "testdata/idle.yaml"}
	with := func(rules string, extra ...string) []string {
		return append(append(append([]string{}, base...), "--rules", rules), extra...)
	}

	const deniedApply = "[ERROR] step 'k8s_apply' requires the cluster-apply capability (policy denies it)"
	rows := []struct {
		name      string
		args      []string
		code      int
		stdoutHas []string
		stdoutNot []string
	}{
		{"an override that applies is denied by default",
			with(applyRule), 1, []string{deniedApply}, []string{"written", "created"}},
		{"--out does not lift the denial",
			with(applyRule, "--out", notOut), 1, []string{deniedApply}, []string{"written"}},
		{"the other --allow flags do not grant apply",
			with(applyRule, "--allow-exec", "--allow-fs", "--allow-net", "--allow-env", "--allow-cluster-rollout"), 1, []string{deniedApply}, nil},
		{"an override cannot widen the policy from inside the file",
			with(widening), 1, []string{deniedApply}, nil},
		{"replace mode is no way around it",
			with(applyRule, "--rules-mode", "replace"), 1, []string{deniedApply}, nil},
		{"granted, but the offline source has no cluster: a loud error, never a silent skip",
			with(applyRule, "--allow-cluster-apply"), 1, []string{"no cluster is configured"}, []string{"written", "created", "would-"}},
		{"the offline twin needs the fs capability as well",
			with(applyRule, "--allow-cluster-apply", "--out", notOut), 1,
			[]string{"[ERROR] step 'k8s_apply' requires the fs capability (policy denies it)"}, []string{"written"}},
		{"with both capabilities the offline twin renders files",
			with(applyRule, "--allow-cluster-apply", "--allow-fs", "--out", out), 0,
			[]string{"written"}, []string{"[ERROR]"}},
		{"a rollout step is denied by default",
			with(scaleRule), 1, []string{"[ERROR] step 'k8s_scale' requires the cluster-rollout capability (policy denies it)"}, nil},
		{"the apply capability does not grant rollout",
			with(scaleRule, "--allow-cluster-apply"), 1, []string{"requires the cluster-rollout capability"}, nil},
		{"granted rollout reaches the step (no cluster to talk to)",
			with(scaleRule, "--allow-cluster-rollout"), 1, []string{"no cluster is configured"}, []string{"requires the cluster-rollout"}},
		{"rollout status is a read: no capability needed",
			with(statusRule), 1, []string{"no cluster is configured"}, []string{"requires the"}},
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
		})
	}
	files, _ := filepath.Glob(filepath.Join(out, "*.yaml"))
	if len(files) == 0 {
		t.Fatal("the permitted offline render wrote no files")
	}
	if _, err := os.Stat(notOut); err == nil {
		t.Fatal("a denied run created its --out directory")
	}
}

// TestOperatorHelpListsTheWriteFlags keeps the flags discoverable and the defaults visible.
func TestOperatorHelpListsTheWriteFlags(t *testing.T) {
	bin := buildOperator(t)
	code, stdout, stderr := operator(t, bin, "check", "--help")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	for _, f := range []string{"--allow-cluster-apply", "--allow-cluster-rollout", "--dry-run", "--out", "--force-conflicts", "--field-manager"} {
		if !strings.Contains(stdout, f) {
			t.Errorf("--help does not list %s:\n%s", f, stdout)
		}
	}
	if !strings.Contains(stdout, "denied by default") || !strings.Contains(stdout, "--dry-run=false") {
		t.Errorf("--help must say the writes are denied by default and how to make them real:\n%s", stdout)
	}
}
