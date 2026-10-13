package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/interactive"
)

func TestFailureOfClassifiesByCause(t *testing.T) {
	cap := &CapabilityError{Capability: CapNet, Step: "dial_tcp"}
	rows := []struct {
		name string
		err  error
		want Failure
	}{
		{"capability denied", cap,
			Failure{Kind: FailCapabilityDenied, Message: cap.Error(), Step: "dial_tcp", Capability: "net"}},
		{"capability denied, wrapped by another step", fmt.Errorf("fold: %w", cap),
			Failure{Kind: FailCapabilityDenied, Message: "fold: " + cap.Error(), Step: "dial_tcp", Capability: "net"}},
		{"unknown step", &interactive.UnknownStepError{Name: "frobnicate"},
			Failure{Kind: FailUnknownStep, Message: `unknown step or command: "frobnicate"`, Step: "frobnicate"}},
		{"stopped by the deadline", &CancelError{Cause: context.DeadlineExceeded},
			Failure{Kind: FailStopped, Message: "the run was stopped: context deadline exceeded"}},
		{"stopped by the output limit", &CancelError{Cause: ErrOutputLimit},
			Failure{Kind: FailStopped, Message: "the run was stopped: output limit exceeded"}},
		{"anything else", errors.New("undefined identifier: no_such_rule"),
			Failure{Kind: FailError, Message: "undefined identifier: no_such_rule"}},
		{"a wrapped plain error", fmt.Errorf("map: %w", errors.New("boom")),
			Failure{Kind: FailError, Message: "map: boom"}},
	}
	for _, r := range rows {
		if got := FailureOf(r.err); got != r.want {
			t.Errorf("%s: FailureOf = %+v, want %+v", r.name, got, r.want)
		}
	}
}

// The classification must match what a real evaluation produces, not only hand-built errors.
func TestEvaluationErrorsCarryTheirKind(t *testing.T) {
	in, _ := quietInterpreter()
	vals := in.EvaluateToValues("spawn StrVal: s |> frobnicate-nothing")
	if len(vals) != 1 {
		t.Fatalf("vals = %v", vals)
	}
	if f := FailureOf(vals[0].(error)); f.Kind != FailUnknownStep || f.Step != "frobnicate-nothing" {
		t.Errorf("unknown step: %+v", f)
	}
	in.SetPolicy(DenyAll())
	vals = in.EvaluateToValues("spawn StrVal: s |> echo hi")
	if len(vals) != 1 {
		t.Fatalf("vals = %v", vals)
	}
	if f := FailureOf(vals[0].(error)); f.Kind != FailCapabilityDenied || f.Step != "echo" || f.Capability != "exec" {
		t.Errorf("denied step: %+v", f)
	}
}

func TestJSONReportCarriesDiagnostics(t *testing.T) {
	denied := &CapabilityError{Capability: CapExec, Step: "echo"}
	rep := classify([]any{denied, errors.New("plain failure")})

	var buf bytes.Buffer
	n := rep.writeJSON(&buf)
	want := `{"status":"error","count":0,"violations":[],"errors":["step 'echo' requires the exec capability (policy denies it)","plain failure"],` +
		`"diagnostics":[{"kind":"capability_denied","message":"step 'echo' requires the exec capability (policy denies it)","step":"echo","capability":"exec"},` +
		`{"kind":"error","message":"plain failure"}]}` + "\n"
	if buf.String() != want || n != 2 {
		t.Fatalf("JSON report:\n got %q (%d errors)\nwant %q", buf.String(), n, want)
	}

	// the text form is the free text it always was
	var text bytes.Buffer
	rep.writeText(&text)
	if want := "[ERROR] step 'echo' requires the exec capability (policy denies it)\n[ERROR] plain failure\n"; text.String() != want {
		t.Fatalf("text report = %q, want %q", text.String(), want)
	}
}

func TestJSONReportDiagnosticsAreAlignedWithErrors(t *testing.T) {
	// errors[i] and diagnostics[i] describe the same failure, whatever mixes with findings
	rep := classify([]any{errors.New("first"), []any{"RuleA", "res"}, &CapabilityError{Capability: CapFS, Step: "scan_file"}})
	var buf bytes.Buffer
	rep.writeJSON(&buf)
	line := buf.String()
	for _, want := range []string{
		`"errors":["first","step 'scan_file' requires the fs capability (policy denies it)"]`,
		`"diagnostics":[{"kind":"error","message":"first"},{"kind":"capability_denied","message":"step 'scan_file' requires the fs capability (policy denies it)","step":"scan_file","capability":"fs"}]`,
	} {
		if !strings.Contains(line, want) {
			t.Errorf("report lacks %s:\n%s", want, line)
		}
	}
}

func TestJSONReportEmptyAndHealthyHaveEmptyDiagnostics(t *testing.T) {
	var buf bytes.Buffer
	classify([]any{[]any{}}).writeJSON(&buf)
	if want := `{"status":"ok","count":0,"violations":[],"errors":[],"diagnostics":[]}` + "\n"; buf.String() != want {
		t.Fatalf("healthy report = %q, want %q", buf.String(), want)
	}
	// nothing at all is a failure with its own kind (Mandate 6)
	buf.Reset()
	classify(nil).writeJSON(&buf)
	if !strings.Contains(buf.String(), `"diagnostics":[{"kind":"no_result"`) || !strings.Contains(buf.String(), `"status":"error"`) {
		t.Fatalf("an empty result must be a no_result diagnostic: %s", buf.String())
	}
}

func TestJSONReportUnencodableFindingIsADiagnostic(t *testing.T) {
	var buf bytes.Buffer
	classify([]any{[]any{"Rule", make(chan int)}}).writeJSON(&buf)
	line := buf.String()
	if !strings.Contains(line, `"status":"error"`) || !strings.Contains(line, `"kind":"unencodable_value"`) {
		t.Fatalf("an unencodable finding must be reported, not dropped: %s", line)
	}
}
