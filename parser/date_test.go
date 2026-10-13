package parser_test

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/parser"
)

func TestDateTokens(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want []string // "A:" marks a TokDate
	}{
		{"@2026-10-13", []string{"A:2026-10-13"}},
		{"@2026-10-13+7d", []string{"A:2026-10-13", "+", "D:7d"}},
		{"@2026-10-13 - @2026-01-01", []string{"A:2026-10-13", "-", "A:2026-01-01"}},
		{"[@2026-10-13,@2026-10-14]", []string{"[", "A:2026-10-13", ",", "A:2026-10-14", "]"}},
		{"(@2026-10-13)", []string{"(", "A:2026-10-13", ")"}},
	} {
		toks, err := parser.TokenizeInfix(tc.src)
		if err != nil {
			t.Errorf("%q: %v", tc.src, err)
			continue
		}
		var got []string
		for _, tk := range toks[:len(toks)-1] {
			switch tk.Type {
			case parser.TokDate:
				got = append(got, "A:"+tk.Text)
			case parser.TokDuration:
				got = append(got, "D:"+tk.Text)
			default:
				got = append(got, tk.Text)
			}
		}
		if strings.Join(got, "|") != strings.Join(tc.want, "|") {
			t.Errorf("%q tokens = %v, want %v", tc.src, got, tc.want)
		}
	}
}

func TestDateLiteralValues(t *testing.T) {
	for src, want := range map[string]string{
		"@2026-10-13":        "2026-10-13",
		"@2026-10-13 + 7d":   "2026-10-20",
		"@2026-10-13 - 1w":   "2026-10-06",
		"@2024-03-01 - 1d":   "2024-02-29",
		"@2026-12-31 + 1d":   "2027-01-01",
		"@2026-10-13 + 0s":   "2026-10-13",
		"@2026-10-13 + 168h": "2026-10-20",
	} {
		val, err := parser.EvalInfixString(src, nil)
		if err != nil {
			t.Errorf("%s: %v", src, err)
			continue
		}
		if d, ok := val.(functions.Date); !ok || d.String() != want {
			t.Errorf("%s = %#v, want the Date %s", src, val, want)
		}
	}
	val, err := parser.EvalInfixString("@2026-10-13 - @2026-10-06", nil)
	if err != nil || val != functions.Duration(7*24*3600e9) {
		t.Errorf("d - e = %#v, %v", val, err)
	}
	val, err = parser.EvalInfixString("@2026-10-13 < @2026-10-14", nil)
	if err != nil || val != true {
		t.Errorf("d < e = %v, %v", val, err)
	}
}

func TestDateLiteralErrors(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"@2026-02-30", "day out of range"},
		{"@2026-13-01", "month out of range"},
		{"@2026-1-3", "malformed time literal"},
		{"@2026-10-13x", "unexpected"},
		{"@2026-10-139", "unexpected"},
		{"@0000-01-01", "out of range"},
		{"@2026-10-13 + 90m", "whole days"},
		{"@2026-10-13 + @2026-10-13", "two days cannot be added"},
		{"@2026-10-13 < @2026-10-13T00:00:00Z", "no zone"},
	} {
		_, err := parser.EvalInfixString(tc.src, nil)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error = %v, want one containing %q", tc.src, err, tc.want)
		}
	}
}
