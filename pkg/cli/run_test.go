package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseRunArgs(t *testing.T) {
	cases := []struct {
		name string
		argv []string
		want runArgs
		err  string
	}{
		{"none", nil, runArgs{}, ""},
		{"script", []string{"s.gatwd"}, runArgs{script: "s.gatwd"}, ""},
		{"interactive", []string{"-i", "s.gatwd"}, runArgs{script: "s.gatwd", interactive: true}, ""},
		{"args after dashes", []string{"s.gatwd", "--", "a", "-b", "--c"}, runArgs{script: "s.gatwd", args: []string{"a", "-b", "--c"}}, ""},
		{"task separate", []string{"--task", "t", "s.gatwd"}, runArgs{script: "s.gatwd", task: "t"}, ""},
		{"task equals", []string{"s.gatwd", "--task=t", "--", "x"}, runArgs{script: "s.gatwd", task: "t", args: []string{"x"}}, ""},
		{"task without script", []string{"--task", "t"}, runArgs{task: "t"}, ""},
		{"empty args", []string{"s.gatwd", "--"}, runArgs{script: "s.gatwd", args: []string{}}, ""},
		{"task needs name", []string{"--task"}, runArgs{}, "--task needs a task name"},
		{"unknown flag", []string{"--bogus"}, runArgs{}, `unknown flag "--bogus"`},
		{"second positional", []string{"a.gatwd", "b.gatwd"}, runArgs{script: "a.gatwd"}, `unexpected argument "b.gatwd"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := parseRunArgs(c.argv)
			if c.err != "" {
				if err == nil || !strings.Contains(err.Error(), c.err) {
					t.Fatalf("error = %v, want one containing %q", err, c.err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(got.args) == 0 && len(c.want.args) == 0 {
				got.args, c.want.args = nil, nil
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v, want %+v", got, c.want)
			}
		})
	}
}

func writeScript(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "s.gatwd")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRunMainFunctionGetsArgs(t *testing.T) {
	script := writeScript(t, `def main = \args -> map (\a -> a ++ "!") args`+"\n")
	res, err := NewInterpreter().Run(RunOptions{Script: script, Args: []string{"x", "a |> b"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Task != "main" || !reflect.DeepEqual(res.Lines, []string{"[x! a |> b!]"}) {
		t.Fatalf("got %+v", res)
	}
}

func TestRunWithoutMainRunsNoTask(t *testing.T) {
	script := writeScript(t, "def helper = \\x -> x\n")
	res, err := NewInterpreter().Run(RunOptions{Script: script})
	if err != nil || res.Task != "" || len(res.Lines) != 0 {
		t.Fatalf("got %+v, %v; want no task and no error", res, err)
	}
}

func TestRunSelectedTaskReplacesMain(t *testing.T) {
	script := writeScript(t, "def main = \\args -> 1\ndef other = \\args -> 2\n")
	res, err := NewInterpreter().Run(RunOptions{Script: script, Task: "other"})
	if err != nil || res.Task != "other" || !reflect.DeepEqual(res.Lines, []string{"2"}) {
		t.Fatalf("got %+v, %v", res, err)
	}
}

func TestRunFragmentTaskStreamsArgs(t *testing.T) {
	script := writeScript(t, "def twice = map (\\x -> x ++ x)\n")
	res, err := NewInterpreter().Run(RunOptions{Script: script, Task: "twice", Args: []string{"a", "b c"}})
	if err != nil || !reflect.DeepEqual(res.Lines, []string{"aa", "b cb c"}) {
		t.Fatalf("got %+v, %v; want one result per argument", res, err)
	}
}

func TestRunBindsArgsAndArg0ForTopLevelStatements(t *testing.T) {
	script := writeScript(t, "def probe = \\args -> [ARG0, (nth 0 args)]\n")
	i := NewInterpreter()
	res, err := i.Run(RunOptions{Script: script, Task: "probe", Args: []string{"v"}})
	if err != nil || !reflect.DeepEqual(res.Lines, []string{"[" + script + " v]"}) {
		t.Fatalf("got %+v, %v", res, err)
	}
}

func TestArgsDefaultToEmptyList(t *testing.T) {
	i := NewInterpreter()
	got := i.evaluatePipelineStr("ARGS")
	if !reflect.DeepEqual(got, []string{"[]"}) {
		t.Fatalf("ARGS = %v, want [[]]", got)
	}
	if got := i.evaluatePipelineStr("ARG0"); !reflect.DeepEqual(got, []string{""}) {
		t.Fatalf("ARG0 = %v, want an empty string", got)
	}
}

func TestRunFailsLoud(t *testing.T) {
	script := writeScript(t, "def main = \\args -> 1\ndef two = \\a b -> a\ndef bad = nosuchstepxyz\n")
	cases := []struct {
		name string
		o    RunOptions
		want string
	}{
		{"unknown task", RunOptions{Script: script, Task: "nope"}, `unknown task "nope" (defined: bad, main, two)`},
		{"wrong arity", RunOptions{Script: script, Task: "two"}, "takes 2 arguments"},
		{"unreadable script", RunOptions{Script: filepath.Join(t.TempDir(), "missing.gatwd")}, "cannot read script"},
		{"fragment error", RunOptions{Script: script, Task: "bad", Args: []string{"x"}}, "reported 1 error"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := NewInterpreter().Run(c.o)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("error = %v, want one containing %q", err, c.want)
			}
		})
	}
}

// A bare constant used as a whole statement resolves to its value; its text is not a
// statement any parser accepts (it used to print nothing).
func TestBareConstantStatement(t *testing.T) {
	i := NewInterpreter()
	i.EvaluateLine("const N = 5")
	if got := i.evaluatePipelineStr("N"); !reflect.DeepEqual(got, []string{"5"}) {
		t.Fatalf("N = %v, want [5]", got)
	}
}

// ARGS is a list value: an argument that looks like pipeline syntax or contains spaces must
// survive a pipeline that starts from ARGS (its `%v` text would be re-split and re-parsed).
func TestArgsAreNotSplicedAsText(t *testing.T) {
	i := NewInterpreter()
	i.setArgs("s.gatwd", []string{"a |> b", "c d"})
	got := i.evaluatePipelineStr(`ARGS |> map (\a -> a ++ "!")`)
	if !reflect.DeepEqual(got, []string{"[a |> b! c d!]"}) {
		t.Fatalf("got %v", got)
	}
	// ARGS as an operand inside a longer head: two arguments, not the four words of its text.
	if got := i.evaluatePipelineStr("length ARGS"); !reflect.DeepEqual(got, []string{"2"}) {
		t.Fatalf("length ARGS = %v, want [2]", got)
	}
}
