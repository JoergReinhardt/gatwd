package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// treeRun executes the command tree of a plain interpreter and returns the exit code and
// both streams.
func treeRun(t *testing.T, stdin string, argv ...string) (code int, stdout, stderr string) {
	t.Helper()
	tree := NewCommandTree(CommandSpec{
		Name:  "gatwd",
		Build: func(bool) *Interpreter { return NewInterpreter() },
	})
	var out, errb bytes.Buffer
	code = tree.Execute(argv, strings.NewReader(stdin), &out, &errb)
	return code, out.String(), errb.String()
}

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

const compatScript = "def main = \\args -> length args\n1 + 2\n"

// TestCommandTreeBackwardCompatible pins every invocation the hand-rolled parser accepted
// (and rejected) before the cobra tree, with the exit code and the observable output.
func TestCommandTreeBackwardCompatible(t *testing.T) {
	script := writeFile(t, "s.gatwd", compatScript)
	badCheck := writeFile(t, "chk.gatwd", "def inc : Int -> Int = \\x -> x + 1\n(1 +\n")

	cases := []struct {
		name       string
		argv       []string
		stdin      string
		code       int
		outHas     []string // all must appear in stdout, in order
		outLacks   []string
		stderrHas  string
		stderrNone bool
	}{
		{name: "script", argv: []string{script}, code: 0,
			outHas: []string{"Running Batch Mode: " + script, "∷ 3", "∷ 0"}},
		{name: "script with args", argv: []string{script, "--", "a", "b"}, code: 0,
			outHas: []string{"∷ 3", "∷ 2"}},
		{name: "args keep dashes", argv: []string{script, "--", "-x", "--y", "z"}, code: 0,
			outHas: []string{"∷ 3", "∷ 3"}},
		{name: "task before script", argv: []string{"--task", "main", script, "--", "a"}, code: 0,
			outHas: []string{"∷ 3", "∷ 1"}},
		{name: "task after script", argv: []string{script, "--task", "main", "--", "x", "y", "z"}, code: 0,
			outHas: []string{"∷ 3", "∷ 3"}},
		{name: "task equals", argv: []string{"--task=main", script}, code: 0,
			outHas: []string{"∷ 3", "∷ 0"}},
		{name: "unknown task", argv: []string{"--task=nope", script}, code: 1,
			stderrHas: `unknown task "nope"`, outLacks: []string{"No output or node unhandled"}},
		{name: "interactive loads script then reads stdin", argv: []string{"-i", script}, stdin: "1 + 1\nexit\n", code: 0,
			outHas: []string{"∷ 3", "Running Interactive Mode", "∷ 2"}, outLacks: []string{"Running Batch Mode"}},
		{name: "no arguments is the REPL", argv: nil, stdin: "exit\n", code: 0,
			outHas: []string{"Running Interactive Mode"}},
		{name: "REPL ends at end of input", argv: nil, stdin: "", code: 0,
			outHas: []string{"Running Interactive Mode"}},
		{name: "unreadable script", argv: []string{"/no/such/file.gatwd"}, code: 1,
			stderrHas: "cannot read script"},
		{name: "unknown flag", argv: []string{"--bogus"}, code: 2, stderrHas: "bogus"},
		{name: "second positional", argv: []string{script, script}, code: 2,
			stderrHas: "script arguments go after --"},
		{name: "task needs a name", argv: []string{"--task"}, code: 2, stderrHas: "task"},
		{name: "check clean", argv: []string{"--check", script}, code: 0, stderrNone: true,
			outLacks: []string{"Running Batch Mode"}},
		{name: "check short flag", argv: []string{"-c", script}, code: 0, stderrNone: true},
		{name: "check structural error", argv: []string{"--check", badCheck}, code: 1,
			stderrHas: badCheck + ":2:1: error: unbalanced delimiter"},
		{name: "check without a file", argv: []string{"--check"}, code: 2, stderrHas: "check"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, out, errOut := treeRun(t, c.stdin, c.argv...)
			if code != c.code {
				t.Fatalf("exit %d, want %d\nstdout:\n%s\nstderr:\n%s", code, c.code, out, errOut)
			}
			pos := 0
			for _, want := range c.outHas {
				idx := strings.Index(out[pos:], want)
				if idx < 0 {
					t.Fatalf("stdout lacks %q (after offset %d):\n%s", want, pos, out)
				}
				pos += idx + len(want)
			}
			for _, bad := range c.outLacks {
				if strings.Contains(out, bad) {
					t.Errorf("stdout must not contain %q:\n%s", bad, out)
				}
			}
			if c.stderrHas != "" && !strings.Contains(errOut, c.stderrHas) {
				t.Errorf("stderr lacks %q:\n%s", c.stderrHas, errOut)
			}
			if c.stderrNone && errOut != "" {
				t.Errorf("stderr must be empty, got:\n%s", errOut)
			}
		})
	}
}

// TestCommandTreeSubcommands covers the commands the old command line did not have.
func TestCommandTreeSubcommands(t *testing.T) {
	script := writeFile(t, "s.gatwd", compatScript)
	tasks := writeFile(t, "tasks.gatwd", "def double = \\args -> length args * 2\ndef shout = emit hi\n")

	cases := []struct {
		name      string
		argv      []string
		code      int
		outHas    string
		stderrHas string
	}{
		{"run", []string{"run", script, "--", "q"}, 0, "∷ 1", ""},
		{"run needs a script", []string{"run"}, 2, "", "script"},
		{"run rejects a second script", []string{"run", script, script}, 2, "", "script arguments go after --"},
		{"eval", []string{"eval", "-e", "1 + 2"}, 0, "∷ 3", ""},
		{"eval repeats -e in order", []string{"eval", "-e", "1 + 1", "-e", "5 + 5"}, 0, "∷ 10", ""},
		{"eval needs -e", []string{"eval"}, 2, "", "-e"},
		{"eval error exits 1", []string{"eval", "-e", "1 + undefined_zzz"}, 1, "[ERROR]", "failed or printed nothing"},
		{"eval silent statement exits 1", []string{"eval", "-e", "emit 1 |> filter (\\x -> x > 5)"}, 1, "No output", "failed or printed nothing"},
		{"root -e after a script", []string{script, "-e", "7 + 1"}, 0, "∷ 8", ""},
		{"check many files", []string{"check", script, script}, 0, "", ""},
		{"check needs a file", []string{"check"}, 2, "", "file"},
		{"check unreadable", []string{"check", "/no/such.gatwd"}, 1, "", "cannot read file"},
		{"task from script", []string{"task", "double", "--script", tasks, "--", "a", "b"}, 0, "∷ 4", ""},
		{"task needs a name", []string{"task"}, 2, "", "task"},
		{"task unknown", []string{"task", "nope", "--script", tasks}, 1, "", `unknown task "nope"`},
		{"define", []string{"eval", "-D", "GREETING=hello", "-e", "GREETING"}, 0, "hello", ""},
		{"define needs =", []string{"eval", "-D", "GREETING", "-e", "1"}, 2, "", "NAME=VALUE"},
		{"define cannot shadow ARGS", []string{"eval", "-D", "ARGS=x", "-e", "1"}, 2, "", "ARGS"},
		{"prelude file", []string{"eval", "--prelude", writeFile(t, "p.gatwd", "def ten = \\u -> 10\n"), "-e", "ten 0"}, 0, "∷ 10", ""},
		{"prelude missing", []string{"eval", "--prelude", "/no/such.gatwd", "-e", "1"}, 1, "", "cannot read prelude"},
		{"version", []string{"version"}, 0, "gatwd", ""},
		{"builtins", []string{"builtins"}, 0, "length", ""},
		{"prelude list without an edition", []string{"prelude", "list"}, 0, "", ""},
		{"prelude cat unknown", []string{"prelude", "cat", "nope"}, 1, "", `no prelude "nope"`},
		{"completion bash", []string{"completion", "bash"}, 0, "bash completion", ""},
		{"completion zsh", []string{"completion", "zsh"}, 0, "compdef", ""},
		{"help", []string{"--help"}, 0, "Usage:", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, out, errOut := treeRun(t, "", c.argv...)
			if code != c.code {
				t.Fatalf("exit %d, want %d\nstdout:\n%s\nstderr:\n%s", code, c.code, out, errOut)
			}
			if c.outHas != "" && !strings.Contains(out, c.outHas) {
				t.Errorf("stdout lacks %q:\n%s", c.outHas, out)
			}
			if c.stderrHas != "" && !strings.Contains(errOut, c.stderrHas) {
				t.Errorf("stderr lacks %q:\n%s", c.stderrHas, errOut)
			}
		})
	}
}

// TestCommandTreeFragmentTask keeps the D4 model on the new command line: a
// pipeline-bodied task receives the arguments as its input stream.
func TestCommandTreeFragmentTask(t *testing.T) {
	p := writeFile(t, "f.gatwd", "def echo_args = map (\\a -> a ++ \"!\")\n")
	code, out, errOut := treeRun(t, "", "task", "echo_args", "--script", p, "--", "x", "y")
	if code != 0 || !strings.Contains(out, "x!") || !strings.Contains(out, "y!") {
		t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errOut)
	}
}

func TestCommandTreeEditionCommandsAndNormalizer(t *testing.T) {
	hello := &cobra.Command{
		Use:  "hello",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Println("hello from an edition command")
			return nil
		},
	}
	tree := NewCommandTree(CommandSpec{
		Name:     "ed",
		Build:    func(bool) *Interpreter { return NewInterpreter() },
		Commands: []*cobra.Command{hello},
		Preludes: []Prelude{{Name: "base.gatwd", Source: func() (string, error) { return "def one = \\u -> 1\n", nil }}},
		NormalizeArgs: func(argv []string) []string {
			out := make([]string, len(argv))
			for i, a := range argv {
				if a == "-greet" {
					a = "hello"
				}
				out[i] = a
			}
			return out
		},
	})
	run := func(argv ...string) (int, string, string) {
		var o, e bytes.Buffer
		return tree.Execute(argv, strings.NewReader(""), &o, &e), o.String(), e.String()
	}
	if code, out, _ := run("hello"); code != 0 || !strings.Contains(out, "hello from an edition command") {
		t.Fatalf("edition command: exit %d, out %q", code, out)
	}
	if code, out, _ := run("-greet"); code != 0 || !strings.Contains(out, "hello from an edition command") {
		t.Fatalf("normalizer: exit %d, out %q", code, out)
	}
	if code, out, _ := run("prelude", "list"); code != 0 || !strings.Contains(out, "base.gatwd") {
		t.Fatalf("prelude list: exit %d, out %q", code, out)
	}
	if code, out, _ := run("prelude", "cat", "base.gatwd"); code != 0 || !strings.Contains(out, "def one") {
		t.Fatalf("prelude cat: exit %d, out %q", code, out)
	}
}

// TestCommandTreeNoPreludeIsHonoured proves --no-prelude reaches the interpreter builder.
func TestCommandTreeNoPreludeIsHonoured(t *testing.T) {
	var seen []bool
	tree := NewCommandTree(CommandSpec{
		Name: "gatwd",
		Build: func(noPrelude bool) *Interpreter {
			seen = append(seen, noPrelude)
			return NewInterpreter()
		},
	})
	var o, e bytes.Buffer
	if code := tree.Execute([]string{"eval", "-e", "1 + 1"}, strings.NewReader(""), &o, &e); code != 0 {
		t.Fatal(e.String())
	}
	if code := tree.Execute([]string{"eval", "--no-prelude", "-e", "1 + 1"}, strings.NewReader(""), &o, &e); code != 0 {
		t.Fatal(e.String())
	}
	if len(seen) != 2 || seen[0] || !seen[1] {
		t.Fatalf("Build saw noPrelude=%v, want [false true]", seen)
	}
}

// TestInterpreterOutWriter pins decision (a): top-level statements print to Interpreter.Out.
func TestInterpreterOutWriter(t *testing.T) {
	var buf bytes.Buffer
	in := NewInterpreter()
	in.Out = &buf
	in.EvaluateLine("1 + 2")
	if got := buf.String(); got != "∷ 3\n" {
		t.Fatalf("Out got %q", got)
	}
}
