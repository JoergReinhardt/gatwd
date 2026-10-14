package cli_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

// runWithDeadline runs script under a 200ms context and fails if it does not return
// promptly, which is exactly what a missing interrupt check looks like.
func runWithDeadline(t *testing.T, script string) string {
	t.Helper()
	interp := cli.NewInterpreter()
	var out strings.Builder
	interp.Out = &out
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	interp.SetContext(ctx)
	done := make(chan string, 1)
	go func() {
		interp.EvaluateScript(script)
		done <- out.String()
	}()
	select {
	case s := <-done:
		return s
	case <-time.After(5 * time.Second):
		t.Fatalf("runaway closure was not interrupted: %s", script)
		return ""
	}
}

func TestInterruptSpinningClosure(t *testing.T) {
	cases := map[string]string{
		"toplevel":      "def spin = \\n -> spin (n + 1)\nspin 0",
		"map":           "def spin = \\n -> spin (n + 1)\n[1, 2] |> map (\\x -> spin x)",
		"comprehension": "def spin = \\n -> spin (n + 1)\n[ spin x | x <- [1, 2] ]",
		"fold":          "def spin = \\n -> spin (n + 1)\n[1, 2] |> fold (\\a x -> spin x) 0",
	}
	for name, script := range cases {
		t.Run(name, func(t *testing.T) {
			got := runWithDeadline(t, script)
			if !strings.Contains(strings.ToLower(got), "cancel") && !strings.Contains(strings.ToLower(got), "deadline") && !strings.Contains(strings.ToLower(got), "timed out") {
				t.Fatalf("expected a cancellation error, got %q", got)
			}
		})
	}
}

func TestInterruptLeavesFiniteRecursionAlone(t *testing.T) {
	got := runWithDeadline(t, "def count 0 = 0\ndef count n = 1 + count (n - 1)\ncount 8")
	if !strings.Contains(got, "8") || strings.Contains(got, "ERROR") {
		t.Fatalf("finite recursion disturbed: %q", got)
	}
}
