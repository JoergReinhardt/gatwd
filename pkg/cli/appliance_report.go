package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/joergreinhardt/gatwd/functions"
)

// This file turns the values one evaluation of an Appliance pipeline produced into the report
// an operator prints, in one of two formats, and decides what counts as a failure.
//
// A rule set conventionally answers with one list of findings, each a [rule, resource] pair,
// and `[]` means compliant:
//
//	[[UnselectedPod Pod/default/stray] [IdleDeployment Deployment/default/lonely]]
//
// The report reads that convention without knowing any domain:
//
//   - a value that is an error is a failure of the check itself (a rule that could not run);
//   - a pipeline that yields no value at all is a failure too — a rule set that returns `[]`
//     is compliant, one that returns nothing is broken (Mandate 6);
//   - every other value is findings: a list contributes its elements, anything else counts as
//     one finding.

// Report formats of an Appliance command (--format).
const (
	FormatText = "text"
	FormatJSON = "json"
)

// checkReport is the classified outcome of one evaluation.
type checkReport struct {
	values   []any     // everything the pipeline yielded, as given
	findings []any     // the non-error values, lists flattened
	errs     []string  // failures of the check itself
	diags    []Failure // the same failures, classified; diags[i] describes errs[i]
}

// classify sorts the values of one evaluation into findings and errors.
func classify(values []any) checkReport {
	r := checkReport{values: values}
	if len(values) == 0 {
		r.addError(Failure{Kind: FailNoResult, Message: "the pipeline yielded no result (a rule set that is compliant answers [], not nothing)"})
		return r
	}
	for _, v := range values {
		if err, ok := v.(error); ok {
			r.addError(FailureOf(err))
			continue
		}
		if items, ok := functions.ToSliceAny(v); ok && !isMapLike(v) {
			r.findings = append(r.findings, items...)
			continue
		}
		r.findings = append(r.findings, v)
	}
	return r
}

// addError records one failure of the check itself, both as the text `errors` carries and as
// its classified form.
func (r *checkReport) addError(d Failure) {
	r.errs = append(r.errs, d.Message)
	r.diags = append(r.diags, d)
}

// writeText prints each value on its own line, as the pipeline produced it; an error value is
// prefixed `[ERROR]` like everywhere else in the DSL.
func (r checkReport) writeText(w io.Writer) {
	for _, v := range r.values {
		if err, ok := v.(error); ok {
			fmt.Fprintf(w, "[ERROR] %v\n", err)
			continue
		}
		fmt.Fprintln(w, v)
	}
	if len(r.values) == 0 {
		fmt.Fprintf(w, "[ERROR] %s\n", r.errs[0])
	}
}

// writeJSON prints the report as one JSON object on one line (JSON Lines), so a log pipeline
// can ingest each check as an event:
//
//	{"status":"violations","count":2,"violations":[{"rule":"UnselectedPod","resource":"Pod/default/stray"},…],"errors":[],"diagnostics":[]}
//
// status is ok, violations or error; count is the number of findings; every key is always
// present. errors lists the failure messages as plain strings; diagnostics describes the same
// failures, one object each and in the same order, with a stable "kind" (see the Fail
// constants) and, where the failure names them, the "step" and the "capability" — so a
// consumer need not parse the prose:
//
//	"diagnostics":[{"kind":"capability_denied","message":"step 'echo' requires the exec capability (policy denies it)","step":"echo","capability":"exec"}]
//
// A finding that is a [rule, resource] pair with a String rule becomes
// {"rule":…,"resource":…}; any other finding is {"value":…}. A value JSON cannot hold (a Set, a
// function) is reported in errors, never dropped. It returns the number of errors reported.
func (r checkReport) writeJSON(w io.Writer) int {
	errs := append([]string(nil), r.errs...)
	diags := append([]Failure(nil), r.diags...)
	var b strings.Builder
	b.WriteString(`{"status":`)
	// the status is computed after encoding, because an unencodable finding is an error
	var items []string
	for _, f := range r.findings {
		item, err := findingJSON(f)
		if err != nil {
			errs = append(errs, err.Error())
			diags = append(diags, Failure{Kind: FailUnencodable, Message: err.Error()})
			continue
		}
		items = append(items, item)
	}
	status := "ok"
	switch {
	case len(errs) > 0:
		status = "error"
	case len(items) > 0:
		status = "violations"
	}
	b.WriteString(quoteJSON(status))
	fmt.Fprintf(&b, `,"count":%d,"violations":[%s],"errors":[`, len(r.findings), strings.Join(items, ","))
	for i, e := range errs {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(quoteJSON(e))
	}
	b.WriteString(`],"diagnostics":[`)
	for i, d := range diags {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(d.json())
	}
	b.WriteString("]}")
	fmt.Fprintln(w, b.String())
	return len(errs)
}

// findingJSON encodes one finding.
func findingJSON(f any) (string, error) {
	var value strings.Builder
	if parts, ok := functions.ToSliceAny(f); ok && len(parts) == 2 {
		if rule, isString := parts[0].(string); isString {
			if err := encodeJSON(&value, parts[1], "$.resource"); err != nil {
				return "", fmt.Errorf("finding %s: %w", rule, err)
			}
			return `{"rule":` + quoteJSON(rule) + `,"resource":` + value.String() + `}`, nil
		}
	}
	if err := encodeJSON(&value, f, "$.value"); err != nil {
		return "", fmt.Errorf("finding %v: %w", f, err)
	}
	return `{"value":` + value.String() + `}`, nil
}

func quoteJSON(s string) string {
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	return strings.TrimSuffix(b.String(), "\n")
}
