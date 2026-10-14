package cli

import (
	"bytes"
	"strings"
	"testing"
)

// TestTaskCommandPromotesATaskToItsOwnSubcommand proves the generic, domain-free
// bridge: a plain `def` in an Edition's own prelude becomes a first-class
// subcommand (`NAME shout -- ...`), not just reachable via `NAME task shout -- ...`.
func TestTaskCommandPromotesATaskToItsOwnSubcommand(t *testing.T) {
	edition := NewEdition("verbtest").
		WithPrelude("def shout = map (\\a -> a ++ \"!\")\n")
	edition = edition.WithCommand(TaskCommand(edition, "shout", "shout the arguments", "shout"))

	tree := edition.Command()
	run := func(argv ...string) (int, string, string) {
		var o, e bytes.Buffer
		return tree.Execute(argv, strings.NewReader(""), &o, &e), o.String(), e.String()
	}

	code, out, errOut := run("shout", "--", "x", "y")
	if code != 0 || !strings.Contains(out, "x!") || !strings.Contains(out, "y!") {
		t.Fatalf("shout -- x y: exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errOut)
	}
	// The original, generic `task` subcommand must still reach the same def —
	// TaskCommand ADDS a verb, it does not take the task's name away from `task`.
	code, out, errOut = run("task", "shout", "--", "z")
	if code != 0 || !strings.Contains(out, "z!") {
		t.Fatalf("task shout -- z: exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errOut)
	}
}

// TestTaskCommandOwnDefineFlag proves a verb command's own -D/--define reaches
// the task, independent of the root's: a task that reads a named constant (the
// gp_roots `-D OVERLAY`/`-D PATTERN` convention) still works when promoted to a
// verb, not only via `task NAME -D X=Y -- ...`.
func TestTaskCommandOwnDefineFlag(t *testing.T) {
	edition := NewEdition("verbtest").
		WithPrelude("def greet = map (\\_ -> GREETING)\n")
	edition = edition.WithCommand(TaskCommand(edition, "greet", "greet", "greet"))

	tree := edition.Command()
	var o, e bytes.Buffer
	code := tree.Execute([]string{"greet", "-D", "GREETING=hi", "--", "x"}, strings.NewReader(""), &o, &e)
	if code != 0 || !strings.Contains(o.String(), "hi") {
		t.Fatalf("greet -D GREETING=hi -- x: exit %d\nstdout:\n%s\nstderr:\n%s", code, o.String(), e.String())
	}
}

// A task that produces no output at all must say so, not exit silently — the
// same "(No output or node unhandled)" notice the root script runner gives,
// kept consistent so a verb command cannot look like a passing no-op.
func TestTaskCommandReportsNoOutput(t *testing.T) {
	edition := NewEdition("verbtest").
		WithPrelude("def nothing = filter (\\_ -> false)\n")
	edition = edition.WithCommand(TaskCommand(edition, "nothing", "produce nothing", "nothing"))

	tree := edition.Command()
	var o, e bytes.Buffer
	code := tree.Execute([]string{"nothing", "--", "x"}, strings.NewReader(""), &o, &e)
	if code != 0 || !strings.Contains(o.String(), "No output or node unhandled") {
		t.Fatalf("nothing -- x: exit %d\nstdout:\n%s\nstderr:\n%s", code, o.String(), e.String())
	}
}
