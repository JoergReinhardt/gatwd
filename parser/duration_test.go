package parser_test

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/parser"
)

func TestDurationTokens(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want []string // Text of each token (without EOF), "D:" marks a TokDuration
	}{
		{"5s", []string{"D:5s"}},
		{"1h30m", []string{"D:1h30m"}},
		{"1h30m15s", []string{"D:1h30m15s"}},
		{"1.5h", []string{"D:1.5h"}},
		{"250ms", []string{"D:250ms"}},
		{"3µs", []string{"D:3µs"}},
		{"5s+1m", []string{"D:5s", "+", "D:1m"}},
		{"5s*2", []string{"D:5s", "*", "2"}},
		{"(5s)", []string{"(", "D:5s", ")"}},
		{"[1s,2s]", []string{"[", "D:1s", ",", "D:2s", "]"}},
		{"x-5s", []string{"x", "-", "D:5s"}},
		{"5s-1s", []string{"D:5s", "-", "D:1s"}}, // after a Duration a minus is a subtraction
		{"5 s", []string{"5", "s"}},              // a space makes it an application again
		{"5 s 6", []string{"5", "s", "6"}},
		{"2i", []string{"2i"}},       // complex stays complex
		{"3r4", []string{"3r4"}},     // ratio stays ratio
		{"0.75r", []string{"0.75r"}}, // exact decimal stays ratio
		{"5r", nil},                  // a bare ratio suffix is still an error (see below)
	} {
		toks, err := parser.TokenizeInfix(tc.src)
		if tc.want == nil {
			if err == nil {
				t.Errorf("%q: want an error", tc.src)
			}
			continue
		}
		if err != nil {
			t.Errorf("%q: %v", tc.src, err)
			continue
		}
		var got []string
		for _, tk := range toks[:len(toks)-1] {
			if tk.Type == parser.TokDuration {
				got = append(got, "D:"+tk.Text)
			} else {
				got = append(got, tk.Text)
			}
		}
		if strings.Join(got, "|") != strings.Join(tc.want, "|") {
			t.Errorf("%q tokens = %v, want %v", tc.src, got, tc.want)
		}
	}
}

func TestDurationLiteralValues(t *testing.T) {
	const sec = functions.Duration(1e9)
	for _, tc := range []struct {
		src  string
		want any
	}{
		{"5s", 5 * sec},
		{"1h30m", 5400 * sec},
		{"2d", 172800 * sec},
		{"1w", 604800 * sec},
		{"-5s", -5 * sec},
		{"- 5s", -5 * sec},
		{"+5s", 5 * sec},
		{"5s + 1m", 65 * sec},
		{"1h - 30m", 1800 * sec},
		{"5s * 3", 15 * sec},
		{"3 * 5s", 15 * sec},
		{"1m / 4", 15 * sec},
		{"1m * 2 + 5s", 125 * sec},
		{"5s < 1m", true},
		{"1m <= 60s", true},
		{"5s == 5000ms", true},
		{"5s != 5000ms", false},
		{"1h / 1h", 1},
		{"-(5s)", -5 * sec},
	} {
		val, err := parser.EvalInfixString(tc.src, nil)
		if err != nil {
			t.Errorf("%s: %v", tc.src, err)
			continue
		}
		if !functions.StructuralEqual(val, tc.want) || functions.IsDuration(val) != functions.IsDuration(tc.want) {
			t.Errorf("%s = %#v (%T), want %#v", tc.src, val, val, tc.want)
		}
	}
}

// Durations inside lists, comprehensions and applications.
func TestDurationInExpressions(t *testing.T) {
	val, err := parser.EvalInfixString("age > 30d", func(n string) (any, bool) {
		if n == "age" {
			return functions.Duration(40 * 24 * 3600e9), true
		}
		return nil, false
	})
	if err != nil || val != true {
		t.Errorf("age > 30d = %v, %v", val, err)
	}
	val, err = parser.EvalInfixString("age > 30d", func(n string) (any, bool) {
		if n == "age" {
			return functions.Duration(10 * 24 * 3600e9), true
		}
		return nil, false
	})
	if err != nil || val != false {
		t.Errorf("10d age > 30d = %v, %v", val, err)
	}
	val, err = parser.EvalInfixString("[x * 2 | x <- [1s, 2s, 3s]]", nil)
	if err != nil {
		t.Fatal(err)
	}
	items, _ := functions.ToSliceAny(val)
	if len(items) != 3 || items[2] != functions.Duration(6e9) {
		t.Errorf("comprehension over durations = %#v", items)
	}
}

// A malformed literal is a lex error that names the problem.
func TestDurationLiteralErrors(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"5q", "unknown duration unit"},
		{"5min", "unknown duration unit"},
		{"5secs", "unknown duration unit"},
		{"1h1h", "used twice"},
		{"1h30", "no unit"},
		{"1h0.5m", "only the first segment"},
		{"0.1ns", "not a whole number of nanoseconds"},
		{"9223372036854775808ns", "out of range"},
		{"106752d", "out of range"},
		{"5_s", "unknown duration unit"},
	} {
		_, err := parser.EvalInfixString(tc.src, nil)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error = %v, want one containing %q", tc.src, err, tc.want)
		}
	}
}

// Pins: what a space or an operator between number and word still means.
func TestDurationDoesNotChangeOtherForms(t *testing.T) {
	// `5 s` is still an application of 5 (here: s is undefined, 5 is not callable).
	if _, err := parser.EvalInfixString("5 s", func(n string) (any, bool) { return nil, false }); err == nil {
		t.Error("`5 s` must still be an application, not a Duration")
	}
	// Other literals are untouched.
	for src, want := range map[string]any{"2 + 3": 5, "7 / 2": 3.5, "2i * 2i": complex(-4, 0)} {
		got, err := parser.EvalInfixString(src, nil)
		if err != nil || !functions.StructuralEqual(got, want) {
			t.Errorf("%s = %v, %v; want %v", src, got, err, want)
		}
	}
}
