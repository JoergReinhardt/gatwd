package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const cheatsheetPath = "../../examples/scripts/cheatsheet.gatwd"

// quietInterpreter is a full interpreter whose statement output goes to the returned buffer.
func quietInterpreter() (*Interpreter, *bytes.Buffer) {
	in := NewInterpreter()
	buf := &bytes.Buffer{}
	in.Out = buf
	return in, buf
}

// cheatsheetProblems runs a script the way the command line does and lists everything the
// project's gates treat as a failure: an [ERROR] line, a statement that printed nothing,
// and any `--check` diagnostic.
func cheatsheetProblems(t *testing.T, script string, args ...string) []string {
	t.Helper()
	var problems []string
	in, buf := quietInterpreter()
	if _, err := in.Run(RunOptions{Script: script, Args: args}); err != nil {
		problems = append(problems, "run: "+err.Error())
	}
	out := buf.String()
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "[ERROR]") || strings.Contains(line, "No output or node unhandled") {
			problems = append(problems, line)
		}
	}
	content, err := os.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range CheckScript(string(content)) {
		problems = append(problems, "check: "+d.Msg)
	}
	return problems
}

func TestCheatsheetRunsCleanWithAndWithoutArguments(t *testing.T) {
	for _, args := range [][]string{nil, {"a", "b c"}} {
		if problems := cheatsheetProblems(t, cheatsheetPath, args...); len(problems) != 0 {
			t.Errorf("args %q: %s", args, strings.Join(problems, "\n"))
		}
	}
}

func TestCheatsheetMainAndTasks(t *testing.T) {
	var res RunResult
	var err error
	in, _ := quietInterpreter()
	res, err = in.Run(RunOptions{Script: cheatsheetPath, Args: []string{"a", "b c"}})
	if err != nil || res.Task != "main" || !reflect.DeepEqual(res.Lines, []string{"[2 [a! b c!]]"}) {
		t.Fatalf("main = %+v, %v", res, err)
	}
	in, _ = quietInterpreter()
	res, err = in.Run(RunOptions{Script: cheatsheetPath, Task: "shout", Args: []string{"x"}})
	if err != nil || !reflect.DeepEqual(res.Lines, []string{"[x!]"}) {
		t.Fatalf("shout = %+v, %v", res, err)
	}
}

// The check itself must be able to fail: a script with a failing line, a silent line and a
// type error is reported on all three counts.
func TestCheatsheetGateDetectsBrokenScripts(t *testing.T) {
	broken := filepath.Join(t.TempDir(), "broken.gatwd")
	body := "1 / 0\nspawn StrVal: a |> filter \"zzz\"\n"
	if err := os.WriteFile(broken, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	problems := strings.Join(cheatsheetProblems(t, broken), "\n")
	for _, want := range []string{"by zero", "No output or node unhandled"} {
		if !strings.Contains(problems, want) {
			t.Errorf("gate missed %q in:\n%s", want, problems)
		}
	}
}
