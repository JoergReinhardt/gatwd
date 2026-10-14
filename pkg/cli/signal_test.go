package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestCommandTreeStopSignal: a signal cancels a spinning script, the run ends with the
// Unix status for that signal, and a second signal calls exit at once.
func TestCommandTreeStopSignal(t *testing.T) {
	for sig, want := range map[os.Signal]int{os.Interrupt: 130, syscall.SIGTERM: 143} {
		ch := make(chan os.Signal, 2)
		tree := NewCommandTree(CommandSpec{Name: "gatwd", Build: func(bool) *Interpreter { return NewInterpreter() }})
		tree.signals = func() (<-chan os.Signal, func()) { return ch, func() {} }
		tree.exit = func(int) { t.Error("exit must not be called on a single signal") }
		script := writeFile(t, "spin.gatwd", "def spin = \\n -> spin (n + 1)\nspin 0\n")
		ch <- sig
		var out, errb bytes.Buffer
		code := tree.Execute([]string{script}, strings.NewReader(""), &out, &errb)
		if code != want {
			t.Errorf("%v: exit %d, want %d (stderr %q)", sig, code, want, errb.String())
		}
		if !strings.Contains(errb.String(), "stopped") {
			t.Errorf("%v: no one-line report: %q", sig, errb.String())
		}
	}
}

func TestStopOnSignalSecondSignalExits(t *testing.T) {
	ch := make(chan os.Signal, 2)
	exited := make(chan int, 1)
	ctx, release := stopOnSignal("x", func() (<-chan os.Signal, func()) { return ch, func() {} },
		func(c int) { exited <- c }, t.Context(), &bytes.Buffer{})
	defer release()
	ch <- os.Interrupt
	<-ctx.Done()
	ch <- os.Interrupt
	if c := <-exited; c != ExitFailed {
		t.Fatalf("exit %d", c)
	}
}

// TestREPLInterruptReturnsToPrompt: SIGINT stops a spinning statement and the REPL carries on
// with the next line.
func TestREPLInterruptReturnsToPrompt(t *testing.T) {
	in := NewInterpreter()
	var out bytes.Buffer
	in.Out = &out
	pr, pw := io.Pipe()
	done := make(chan struct{})
	go func() { in.REPL(pr); close(done) }()
	io.WriteString(pw, "def spin = \\n -> spin (n + 1)\nspin 0\n")
	time.Sleep(500 * time.Millisecond)
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	time.Sleep(300 * time.Millisecond)
	io.WriteString(pw, "1 + 41\nexit\n")
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("REPL did not come back")
	}
	if !strings.Contains(out.String(), "42") {
		t.Fatalf("statement after the interrupt did not run:\n%s", out.String())
	}
}
