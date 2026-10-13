package cli

import (
	"errors"
	"strconv"
	"strings"

	"github.com/joergreinhardt/gatwd/interactive"
)

// This file gives failures a machine-readable shape (hardening plan W7).
//
// Every failure in the DSL is an error VALUE that prints as free text (`[ERROR] step 'printf'
// requires the exec capability (policy denies it)`); that text is for people and stays as it
// is. A program that consumes the result — a log pipeline reading an operator's `--format
// json`, an API wrapping Run — should not have to parse prose to learn WHAT failed. A
// Failure is that classification: a stable Kind, the message, and the fields that name the
// culprit (the step, the capability).

// The kinds of Failure. A consumer may switch on them; new kinds may be added, so treat an
// unknown kind like FailError.
const (
	// FailCapabilityDenied: a step needed a capability the policy denies (Step, Capability).
	FailCapabilityDenied = "capability_denied"
	// FailUnknownStep: the word is neither a step nor a program on the PATH, usually a typo (Step).
	FailUnknownStep = "unknown_step"
	// FailStopped: the run was cut short by its context or output limit (Message names the cause).
	FailStopped = "stopped"
	// FailTimeout: a check ran longer than the Appliance's timeout (Limit is the duration, e.g. "30s").
	FailTimeout = "timeout"
	// FailOutputLimit: a check produced more than the Appliance's output limit (Limit is the byte count).
	FailOutputLimit = "output_limit"
	// FailNoResult: the pipeline yielded nothing at all, which a compliant rule set never does.
	FailNoResult = "no_result"
	// FailUnencodable: a finding is a value the report format cannot hold.
	FailUnencodable = "unencodable_value"
	// FailError: any other failure; Message is its text.
	FailError = "error"
)

// Failure is the structured form of one failure.
type Failure struct {
	// Kind classifies the failure; one of the Fail constants.
	Kind string
	// Message is the same free text the `[ERROR]` line carries.
	Message string
	// Step is the pipeline step involved, when the failure names one.
	Step string
	// Capability is the denied capability for FailCapabilityDenied.
	Capability string
	// Limit is the bound that was exceeded for FailTimeout (a duration such as "30s") and
	// FailOutputLimit (a number of bytes).
	Limit string
}

// FailureOf classifies err. The error may be wrapped anywhere in a chain (a step that
// reports `fold: …` around the real cause is still found); the first recognised cause wins.
func FailureOf(err error) Failure {
	d := Failure{Kind: FailError, Message: err.Error()}
	var capErr *CapabilityError
	var unknown *interactive.UnknownStepError
	var stopped *CancelError
	var timeout *TimeoutError
	var over *OutputLimitError
	switch {
	case errors.As(err, &timeout):
		d.Kind, d.Limit = FailTimeout, timeout.Limit.String()
	case errors.As(err, &over):
		d.Kind, d.Limit = FailOutputLimit, strconv.FormatInt(over.Limit, 10)
	case errors.As(err, &capErr):
		d.Kind, d.Step, d.Capability = FailCapabilityDenied, capErr.Step, string(capErr.Capability)
	case errors.As(err, &unknown):
		d.Kind, d.Step = FailUnknownStep, unknown.Name
	case errors.As(err, &stopped):
		d.Kind = FailStopped
	}
	return d
}

// json renders the diagnostic as one JSON object; step, capability and limit appear only when set.
func (d Failure) json() string {
	var b strings.Builder
	b.WriteString(`{"kind":` + quoteJSON(d.Kind) + `,"message":` + quoteJSON(d.Message))
	if d.Step != "" {
		b.WriteString(`,"step":` + quoteJSON(d.Step))
	}
	if d.Capability != "" {
		b.WriteString(`,"capability":` + quoteJSON(d.Capability))
	}
	if d.Limit != "" {
		b.WriteString(`,"limit":` + quoteJSON(d.Limit))
	}
	b.WriteByte('}')
	return b.String()
}
