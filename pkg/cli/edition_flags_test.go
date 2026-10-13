package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// TestEditionWithFlags pins the generic flag hook: a flag an edition declares is parsed
// (before any step runs), is visible to the steps through the variable it binds, shows up
// in --help, applies to subcommands, and a flag that was not declared stays a usage error.
func TestEditionWithFlags(t *testing.T) {
	var seen bool
	ed := NewEdition("gatwd-flagtest").
		WithPrelude("type StrVal = String").
		WithFlags(func(f *pflag.FlagSet) { f.BoolVar(&seen, "loud", false, "make steps loud") }).
		WithStep(MapStep("probe", func(v any) any {
			if seen {
				return "loud"
			}
			return "quiet"
		}))

	run := func(argv ...string) (int, string, string) {
		var out, errb bytes.Buffer
		code := ed.Command().Execute(argv, strings.NewReader(""), &out, &errb)
		return code, out.String(), errb.String()
	}

	cases := []struct {
		name string
		argv []string
		want string
	}{
		{"default", []string{"eval", "-e", "spawn StrVal: hello |> probe"}, "quiet"},
		{"flag before the subcommand", []string{"--loud", "eval", "-e", "spawn StrVal: hello |> probe"}, "loud"},
		{"flag after the subcommand", []string{"eval", "--loud", "-e", "spawn StrVal: hello |> probe"}, "loud"},
		{"state does not leak into the next tree", []string{"eval", "-e", "spawn StrVal: hello |> probe"}, "quiet"},
	}
	for _, c := range cases {
		code, out, errb := run(c.argv...)
		if code != 0 || !strings.Contains(out, c.want) || (c.want == "quiet" && strings.Contains(out, "loud")) {
			t.Errorf("%s: %v -> exit %d, stdout %q, stderr %q; want %q", c.name, c.argv, code, out, errb, c.want)
		}
	}

	if code, out, _ := run("--help"); code != 0 || !strings.Contains(out, "--loud") || !strings.Contains(out, "make steps loud") {
		t.Errorf("--help must list the declared flag, got exit %d:\n%s", code, out)
	}
	if code, _, errb := run("--quiet", "eval", "-e", "spawn StrVal: hello |> probe"); code != 2 || !strings.Contains(errb, "unknown flag") {
		t.Errorf("an undeclared flag must stay a usage error (exit 2), got %d: %s", code, errb)
	}
}
