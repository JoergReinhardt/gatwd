package cli

import (
	"os/exec"
	"strings"
	"testing"
)

// TestZipPairsAndKeepsItsArguments pins the `zip CMD` step (hardening Track Q). It used to
// yield only CMD's output (the tour's `echo Primary: %s |> zip date` printed just the date)
// and to cut CMD at its first non-letter (`zip echo hello world` ran a bare `echo`).
func TestZipPairsAndKeepsItsArguments(t *testing.T) {
	if _, err := exec.LookPath("echo"); err != nil {
		t.Skip("no echo binary")
	}
	in := NewInterpreter()

	// the incoming value is the first of the pair, CMD's output the second, whole arguments kept
	got := in.evaluatePipelineStr(`spawn User: joerg |> echo Primary: %s |> zip echo hello world`)
	if len(got) != 1 || got[0] != "[Primary: joerg, hello world]" {
		t.Fatalf("zip echo hello world = %q, want [[Primary: joerg, hello world]]", got)
	}

	// %s in CMD is the incoming value, and the pair is what format_tuple consumes
	got = in.evaluatePipelineStr(`spawn User: joerg |> zip echo hi-%s |> format_tuple "%v and %v"`)
	if len(got) != 1 || got[0] != "joerg and hi-joerg" {
		t.Fatalf("zip echo hi-%%s |> format_tuple = %q, want [joerg and hi-joerg]", got)
	}
}

func TestZipFailsLoud(t *testing.T) {
	if _, err := exec.LookPath("true"); err != nil {
		t.Skip("no true binary")
	}
	in := NewInterpreter()
	for _, row := range []struct{ name, statement, want string }{
		{"no command", `spawn User: joerg |> zip`, "expected a command"},
		{"a command that produces nothing", `spawn User: joerg |> zip true`, "nothing to pair with joerg"},
		{"a failing command", `spawn User: joerg |> zip sh -c "exit 3"`, "exit status 3"},
		{"zip is a whole word", `spawn User: joerg |> zipx foo`, "zipx"},
	} {
		got := in.evaluatePipelineStr(row.statement)
		joined := strings.Join(got, "\n")
		if len(got) != 1 || !strings.Contains(joined, "[ERROR]") || !strings.Contains(joined, row.want) {
			t.Errorf("%s: %s = %q, want one [ERROR] containing %q", row.name, row.statement, got, row.want)
		}
	}
}

// TestFanOutDoesNotPairAFailure: `(cat /missing &&& date)` used to print [<nil>, <date>].
func TestFanOutDoesNotPairAFailure(t *testing.T) {
	if _, err := exec.LookPath("cat"); err != nil {
		t.Skip("no cat binary")
	}
	in := NewInterpreter()
	got := in.evaluatePipelineStr(`spawn User: joerg |> (cat /nonexistent_gatwd_file &&& echo ok)`)
	joined := strings.Join(got, "\n")
	if len(got) != 1 || !strings.Contains(joined, "[ERROR]") || strings.Contains(joined, "<nil>") {
		t.Fatalf("a failing branch must be the error alone, got %q", got)
	}
}
