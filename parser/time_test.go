package parser_test

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/parser"
)

func TestTimeTokens(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want []string // "T:" marks a TokTime
	}{
		{"@2026-10-13T09:30:00Z", []string{"T:2026-10-13T09:30:00Z"}},
		{"@2026-10-13T11:30:00+02:00", []string{"T:2026-10-13T11:30:00+02:00"}},
		{"@2026-10-13T11:30:00-05:30", []string{"T:2026-10-13T11:30:00-05:30"}},
		{"@2026-10-13T09:30:00.250Z", []string{"T:2026-10-13T09:30:00.250Z"}},
		{"(@2026-10-13T09:30:00Z)", []string{"(", "T:2026-10-13T09:30:00Z", ")"}},
		{"@2026-10-13T09:30:00Z+1h", []string{"T:2026-10-13T09:30:00Z", "+", "D:1h"}},
		{"@2026-10-13T09:30:00Z-1h", []string{"T:2026-10-13T09:30:00Z", "-", "D:1h"}},
		{"@2026-10-13T09:30:00Z - @2026-10-13T08:00:00Z", []string{"T:2026-10-13T09:30:00Z", "-", "T:2026-10-13T08:00:00Z"}},
		{"[@2026-10-13T09:30:00Z,@2026-10-13T09:30:01Z]", []string{"[", "T:2026-10-13T09:30:00Z", ",", "T:2026-10-13T09:30:01Z", "]"}},
	} {
		toks, err := parser.TokenizeInfix(tc.src)
		if err != nil {
			t.Errorf("%q: %v", tc.src, err)
			continue
		}
		var got []string
		for _, tk := range toks[:len(toks)-1] {
			switch tk.Type {
			case parser.TokTime:
				got = append(got, "T:"+tk.Text)
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

func TestTimeLiteralValues(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want string // printed form
	}{
		{"@2026-10-13T09:30:00Z", "2026-10-13T09:30:00Z"},
		{"@2026-10-13T11:30:00+02:00", "2026-10-13T11:30:00+02:00"},
		{"@2026-10-13T09:30:00Z + 1h30m", "2026-10-13T11:00:00Z"},
		{"@2026-10-13T09:30:00Z - 2d", "2026-10-11T09:30:00Z"},
		{"@2026-10-13T11:30:00+02:00 - 30m", "2026-10-13T11:00:00+02:00"},
	} {
		val, err := parser.EvalInfixString(tc.src, nil)
		if err != nil {
			t.Errorf("%s: %v", tc.src, err)
			continue
		}
		if got, ok := val.(functions.Time); !ok || got.String() != tc.want {
			t.Errorf("%s = %#v (%T), want the Time %s", tc.src, val, val, tc.want)
		}
	}
	val, err := parser.EvalInfixString("@2026-10-13T09:30:00Z - @2026-10-13T08:00:00Z", nil)
	if err != nil || val != functions.Duration(90*60*1e9) {
		t.Errorf("t - u = %#v, %v", val, err)
	}
	val, err = parser.EvalInfixString("@2026-10-13T11:30:00+02:00 == @2026-10-13T09:30:00Z", nil)
	if err != nil || val != true {
		t.Errorf("same instant, two zones: %v, %v", val, err)
	}
	// the motivating rule shape, with the clock passed in as an ordinary value
	lookup := func(n string) (any, bool) {
		switch n {
		case "now":
			tm, _ := functions.ParseTime("2026-10-13T09:30:00Z")
			return tm, true
		case "expires":
			tm, _ := functions.ParseTime("2026-10-20T09:30:00Z")
			return tm, true
		}
		return nil, false
	}
	val, err = parser.EvalInfixString("expires - now < 14d", lookup)
	if err != nil || val != true {
		t.Errorf("expires - now < 14d = %v, %v", val, err)
	}
	val, err = parser.EvalInfixString("expires - now < 3d", lookup)
	if err != nil || val != false {
		t.Errorf("expires - now < 3d = %v, %v", val, err)
	}
}

func TestTimeLiteralErrors(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"@2026-10-13T09:30:00", "no zone"},
		{"@2026-10-13T09:30:00.5", "no zone"},
		{"@2026-02-30T09:30:00Z", "day out of range"},
		{"@2026-13-13T09:30:00Z", "month out of range"},
		{"@2026-10-13T24:30:00Z", "hour out of range"},
		{"@2026-10-13T09:30:00+2:00", "offset"},
		{"@2026-10-13T09:30:00+0200", "offset"},
		{"@2026-10-13T09:30Z", "hh:mm:ss"},
		{"@2026-10-13T09:30:00.Z", "decimal point"},
		{"@2026-10-13T09:30:00Zx", "unexpected"},
		{"@2026-10-13T09:30:00Z9", "unexpected"},
		{"@", "malformed time literal"},
		{"@foo", "malformed time literal"},
		{"@2026-1-3T09:30:00Z", "malformed time literal"},
		{"@0000-01-01T00:00:00Z", "out of range"},
		{"@9999-12-31T23:59:59Z + 1s", "out of range"},
	} {
		_, err := parser.EvalInfixString(tc.src, nil)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error = %v, want one containing %q", tc.src, err, tc.want)
		}
	}
}
