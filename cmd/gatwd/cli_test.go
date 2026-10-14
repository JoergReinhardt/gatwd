package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestCommandLine drives the real binary through the invocations that must keep working:
// a script with arguments and `main`, --task, --check exit codes, and a malformed command
// line (exit 2).
func TestCommandLine(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not on PATH")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "gatwd")
	if out, err := exec.Command(goBin, "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	script := filepath.Join(dir, "s.gatwd")
	bad := filepath.Join(dir, "bad.gatwd")
	if err := os.WriteFile(script, []byte("def main = \\args -> length args\ndef n = \\args -> 42\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bad, []byte("(1 +\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	exit := func(args ...string) (int, string) {
		cmd := exec.Command(bin, args...)
		cmd.Stdin = strings.NewReader("")
		out, err := cmd.CombinedOutput()
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode(), string(out)
		}
		if err != nil {
			t.Fatal(err)
		}
		return 0, string(out)
	}
	for _, c := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{script, "--", "a", "b"}, 0, "∷ 2"},
		{[]string{script, "--task", "n"}, 0, "∷ 42"},
		{[]string{"--check", script}, 0, ""},
		{[]string{"--check", bad}, 1, "unbalanced delimiter"},
		{[]string{"check", script, bad}, 1, "unbalanced delimiter"},
		{[]string{"--nonsense"}, 2, "nonsense"},
		{[]string{script, script}, 2, "script arguments go after --"},
		{[]string{"eval", "-e", "1 + 2"}, 0, "∷ 3"},
		{[]string{"completion", "bash"}, 0, "bash completion"},
		{[]string{"--help"}, 0, "Usage:"},
	} {
		code, out := exit(c.args...)
		if code != c.code || !strings.Contains(out, c.want) {
			t.Errorf("gatwd %q: exit %d (want %d), output lacks %q:\n%s", c.args, code, c.code, c.want, out)
		}
	}
}

// TestSignalStopsSpinningScript sends SIGINT and SIGTERM to the real binary running an
// endless script: it must stop promptly with 130 / 143.
func TestSignalStopsSpinningScript(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil || runtime.GOOS == "windows" {
		t.Skip("needs the go toolchain and POSIX signals")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "gatwd")
	if out, err := exec.Command(goBin, "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	script := filepath.Join(dir, "spin.gatwd")
	if err := os.WriteFile(script, []byte("def spin = \\n -> spin (n + 1)\nspin 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for sig, want := range map[syscall.Signal]int{syscall.SIGINT: 130, syscall.SIGTERM: 143} {
		cmd := exec.Command(bin, script)
		cmd.Stdin = strings.NewReader("")
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		_ = cmd.Process.Signal(sig)
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case err := <-done:
			ee, ok := err.(*exec.ExitError)
			if !ok || ee.ExitCode() != want {
				t.Errorf("%v: exit %v, want %d", sig, err, want)
			}
		case <-time.After(10 * time.Second):
			_ = cmd.Process.Kill()
			t.Errorf("%v: binary did not stop", sig)
		}
	}
}
