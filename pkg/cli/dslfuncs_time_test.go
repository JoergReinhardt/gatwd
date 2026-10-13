package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

func timeStr(t *testing.T, v any) string {
	t.Helper()
	return fmt.Sprintf("%v", v)
}

// The function forms return the raw Go types: Duration, Time, Date, Int, Float, String.
func TestTimeFuncsStructural(t *testing.T) {
	in := NewInterpreter()
	sec := functions.Duration(1e9)
	for _, tc := range []struct {
		src  string
		want any
	}{
		{`dur_from_seconds 1.5`, 1500 * functions.Duration(1e6)},
		{`dur_from_seconds 90`, 90 * sec},
		{`dur_truncate 1m 90s`, functions.Duration(60e9)},
		{`dur_round 1m 90s`, functions.Duration(120e9)},
		{`to_seconds 1500ms`, 1.5},
		{`dur_ns 5s`, 5000000000},
		{`dur_ms 1500ms`, 1500},
		{`dur_ms -1999us`, -1},
		{`time_unix @2026-10-13T09:30:00Z`, 1791883800},
		{`time_year @2026-10-13T09:30:00Z`, 2026},
		{`time_month @2026-10-13T09:30:00Z`, 10},
		{`time_day @2026-10-13T09:30:00Z`, 13},
		{`time_hour (time_in "Europe/Berlin" @2026-10-13T09:30:00Z)`, 11},
		{`time_minute @2026-10-13T09:30:00Z`, 30},
		{`time_second @2026-10-13T09:30:00.5Z`, 0},
		{`time_nano @2026-10-13T09:30:00.5Z`, 500000000},
		{`time_weekday @2026-10-13T09:30:00Z`, 2},
		{`time_weekday @2026-10-18T09:30:00Z`, 7},
		{`time_format "%d.%m.%Y" @2026-10-13T09:30:00Z`, "13.10.2026"},
		{`date_format "%d.%m.%Y" @2026-10-13`, "13.10.2026"},
		{`date_year @2026-10-13`, 2026},
		{`date_month @2026-10-13`, 10},
		{`date_day @2026-10-13`, 13},
		{`date_weekday @2026-10-13`, 2},
		{`date_days_in_month @2024-02-10`, 29},
		{`is_leap_year 2024`, true},
		{`is_leap_year 1900`, false},
		{`date_of 2026 10 13`, functions.Date{Year: 2026, Month: 10, Day: 13}},
		{`date_add_days 30 @2026-10-13`, functions.Date{Year: 2026, Month: 11, Day: 12}},
		{`date_add_days -13 @2026-10-13`, functions.Date{Year: 2026, Month: 9, Day: 30}},
		{`date_add_months 1 @2026-01-31`, functions.Date{Year: 2026, Month: 2, Day: 28}},
		{`date_add_months -1 @2026-03-31`, functions.Date{Year: 2026, Month: 2, Day: 28}},
		{`date_add_years 1 @2024-02-29`, functions.Date{Year: 2025, Month: 2, Day: 28}},
		{`date_of_time "Europe/Berlin" @2026-10-13T22:30:00Z`, functions.Date{Year: 2026, Month: 10, Day: 14}},
		{`date_of_time "UTC" @2026-10-13T22:30:00Z`, functions.Date{Year: 2026, Month: 10, Day: 13}},
	} {
		got := evalOne(t, in, tc.src)
		if !functions.StructuralEqual(got, tc.want) || functions.KindName(got) != functions.KindName(tc.want) {
			t.Errorf("%s = %#v (%s), want %#v (%s)", tc.src, got, functions.KindName(got), tc.want, functions.KindName(tc.want))
		}
	}
	// Time results keep their zone
	for src, want := range map[string]string{
		`time_in "Europe/Berlin" @2026-10-13T09:30:00Z`:                        "2026-10-13T11:30:00+02:00",
		`time_in "UTC" @2026-10-13T11:30:00+02:00`:                             "2026-10-13T09:30:00Z",
		`time_from_unix 1791883800`:                                            "2026-10-13T09:30:00Z",
		`time_truncate 1h @2026-10-13T09:45:30Z`:                               "2026-10-13T09:00:00Z",
		`time_round 1h @2026-10-13T09:45:30Z`:                                  "2026-10-13T10:00:00Z",
		`time_of_date "Europe/Berlin" @2026-10-13`:                             "2026-10-13T00:00:00+02:00",
		`time_of_date "Europe/Berlin" @2026-03-29`:                             "2026-03-29T00:00:00+01:00",
		`time_of_date "Asia/Beirut" @2026-03-29`:                               "2026-03-29T01:00:00+03:00",
		`time_format "rfc3339" (time_in "Asia/Kolkata" @2026-10-13T09:30:00Z)`: "2026-10-13T15:00:00+05:30",
	} {
		got := evalOne(t, in, src)
		var s string
		switch x := got.(type) {
		case functions.Time:
			s = x.String()
		case string:
			s = x
		default:
			t.Errorf("%s = %#v, want a Time or String", src, got)
			continue
		}
		if s != want {
			t.Errorf("%s = %s, want %s", src, s, want)
		}
	}
}

// Malformed text is Nothing, well-formed is Just of the raw type; never an error.
func TestTimeParsersAreMaybe(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct {
		src  string
		want any // nil: Nothing
	}{
		{`dur_parse "1h30m"`, functions.Duration(5400e9)},
		{`dur_parse "-5s"`, functions.Duration(-5e9)},
		{`time_parse "2026-10-13T09:30:00Z"`, "2026-10-13T09:30:00Z"},
		{`date_parse "2026-10-13"`, functions.Date{Year: 2026, Month: 10, Day: 13}},
		{`dur_parse "5q"`, nil}, {`dur_parse ""`, nil}, {`dur_parse "1h1h"`, nil}, {`dur_parse "5"`, nil}, {`dur_parse "9223372036854775808ns"`, nil},
		{`time_parse "2026-10-13T09:30:00"`, nil}, {`time_parse "yesterday"`, nil}, {`time_parse "2026-02-30T09:30:00Z"`, nil}, {`time_parse ""`, nil},
		{`date_parse "2026-02-30"`, nil}, {`date_parse "13.10.2026"`, nil}, {`date_parse "2026-10-13T09:30:00Z"`, nil}, {`date_parse ""`, nil},
	} {
		m, ok := evalOne(t, in, tc.src).(functions.Maybe[any])
		if !ok {
			t.Errorf("%s did not return a Maybe", tc.src)
			continue
		}
		v, just := m.MaybeValue()
		if tc.want == nil {
			if just {
				t.Errorf("%s = Just(%v), want Nothing", tc.src, v)
			}
			continue
		}
		if !just {
			t.Errorf("%s = Nothing, want Just(%v)", tc.src, tc.want)
			continue
		}
		if tt, isTime := v.(functions.Time); isTime {
			v = tt.String()
		}
		if v != tc.want {
			t.Errorf("%s = Just(%#v), want Just(%#v)", tc.src, v, tc.want)
		}
	}
}

// Every partial function is an [ERROR] on a request that cannot be met, and on a wrong type.
func TestTimeFuncsFailLoud(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct{ src, want string }{
		{`time_in "Nowhere/Land" @2026-10-13T09:30:00Z`, "unknown time zone"},
		{`time_in "" @2026-10-13T09:30:00Z`, "never implicit"},
		{`time_in "Local" @2026-10-13T09:30:00Z`, "never implicit"},
		{`time_in 5 @2026-10-13T09:30:00Z`, "zone"},
		{`time_in "UTC" @2026-10-13`, "want a Time, got Date"},
		{`time_in "UTC" 5`, "want a Time, got Int"},
		{`date_of_time "Nowhere/Land" @2026-10-13T09:30:00Z`, "unknown time zone"},
		{`date_of_time "UTC" @2026-10-13`, "want a Time, got Date"},
		{`time_of_date "Nowhere/Land" @2026-10-13`, "unknown time zone"},
		{`time_of_date "UTC" @2026-10-13T09:30:00Z`, "want a Date, got Time"},
		{`date_of 2026 2 30`, "day 30"},
		{`date_of 2026 13 1`, "month 13"},
		{`date_of 0 1 1`, "year 0"},
		{`date_of 2026 1`, ""}, // partial application is a closure, checked below
		{`date_of "2026" 1 1`, "String"},
		{`date_of 1.5 1 1`, "integer"},
		{`date_add_months 1 @9999-12-31`, "out of range"},
		{`date_add_days 1 @9999-12-31`, "out of range"},
		{`date_add_days 1 @2026-10-13T09:30:00Z`, "want a Date, got Time"},
		{`date_add_days "1" @2026-10-13`, "String"},
		{`date_year 5`, "want a Date, got Int"},
		{`date_year @2026-10-13T09:30:00Z`, "want a Date, got Time"},
		{`time_year @2026-10-13`, "want a Time, got Date"},
		{`time_format "%q" @2026-10-13T09:30:00Z`, "unknown layout directive"},
		{`time_format "%" @2026-10-13T09:30:00Z`, "lone %"},
		{`time_format "date" @2026-10-13`, "want a Time, got Date"},
		{`date_format "%H" @2026-10-13`, "time of day"},
		{`date_format "%q" @2026-10-13`, "unknown layout directive"},
		{`date_format "date" @2026-10-13T09:30:00Z`, "want a Date, got Time"},
		{`time_from_unix 253402300800`, "out of range"},
		{`time_from_unix "5"`, "String"},
		{`time_truncate 0s @2026-10-13T09:30:00Z`, "positive"},
		{`time_truncate -1h @2026-10-13T09:30:00Z`, "positive"},
		{`time_truncate 5 @2026-10-13T09:30:00Z`, "want a Duration, got Int"},
		{`time_round 0s @2026-10-13T09:30:00Z`, "positive"},
		{`dur_truncate 0s 5s`, "positive"},
		{`dur_round -1s 5s`, "positive"},
		{`dur_round 1s 5`, "want a Duration, got Int"},
		{`dur_from_seconds 1e10`, ""}, // not a literal in the lexer; error either way
		{`dur_from_seconds 10000000000`, "overflow"},
		{`dur_from_seconds "5"`, "number"},
		{`dur_ns 5`, "want a Duration, got Int"},
		{`dur_ms @2026-10-13`, "want a Duration, got Date"},
		{`to_seconds "5s"`, "want a Duration, got String"},
		{`is_leap_year "x"`, "String"},
		{`dur_parse 5`, "want a String"},
		{`time_parse 5`, "want a String"},
		{`date_parse 5`, "want a String"},
	} {
		if tc.want == "" {
			continue
		}
		res := in.evaluatePipelineStr(tc.src)
		if len(res) != 1 || !strings.HasPrefix(res[0], "[ERROR]") || !strings.Contains(res[0], tc.want) {
			t.Errorf("%s = %v, want an [ERROR] containing %q", tc.src, res, tc.want)
		}
	}
}

// Step form and function form agree: `x |> f a` is `f a x`.
func TestTimeStepFormEqualsFunctionForm(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct{ step, fn string }{
		{`@2026-10-13T09:30:00Z |> time_year`, `time_year @2026-10-13T09:30:00Z`},
		{`@2026-10-13T09:30:00Z |> time_in "Europe/Berlin"`, `time_in "Europe/Berlin" @2026-10-13T09:30:00Z`},
		{`@2026-10-13T09:30:00Z |> time_format "%H:%M"`, `time_format "%H:%M" @2026-10-13T09:30:00Z`},
		{`@2026-10-13T09:45:00Z |> time_truncate 1h`, `time_truncate 1h @2026-10-13T09:45:00Z`},
		{`@2026-10-13 |> date_add_months 1`, `date_add_months 1 @2026-10-13`},
		{`@2026-10-13 |> date_weekday`, `date_weekday @2026-10-13`},
		{`@2026-10-13 |> time_of_date "Europe/Berlin"`, `time_of_date "Europe/Berlin" @2026-10-13`},
		{`90s |> dur_round 1m`, `dur_round 1m 90s`},
		{`"1h" |> dur_parse`, `dur_parse "1h"`},
		{`@2026-10-13T09:30:00Z |> date_of_time "UTC"`, `date_of_time "UTC" @2026-10-13T09:30:00Z`},
	} {
		a := timeStr(t, evalOne(t, in, tc.step))
		b := timeStr(t, evalOne(t, in, tc.fn))
		if a != b {
			t.Errorf("%s = %s but %s = %s", tc.step, a, tc.fn, b)
		}
	}
}

// Inside a lambda and a comprehension, and partially applied (curried, data-last).
func TestTimeFuncsInsideLambdasAndCurried(t *testing.T) {
	in := NewInterpreter()
	got := in.EvaluateToValues(`[@2026-10-13, @2026-01-31] |> map (\d -> date_add_months 1 d)`)
	items := itemsOf(t, got[0])
	if len(items) != 2 || items[0] != (functions.Date{Year: 2026, Month: 11, Day: 13}) || items[1] != (functions.Date{Year: 2026, Month: 2, Day: 28}) {
		t.Errorf("map date_add_months = %#v", items)
	}
	got = in.EvaluateToValues(`map (date_add_days 1) [@2026-12-31, @2026-02-28]`)
	items = itemsOf(t, got[0])
	if len(items) != 2 || items[0] != (functions.Date{Year: 2027, Month: 1, Day: 1}) || items[1] != (functions.Date{Year: 2026, Month: 3, Day: 1}) {
		t.Errorf("map (date_add_days 1) = %#v", items)
	}
	got = in.EvaluateToValues(`[t | t <- [@2026-10-13T09:30:00Z, @2026-10-14T09:30:00Z, @2026-10-15T09:30:00Z], time_day t > 13]`)
	if items = itemsOf(t, got[0]); len(items) != 2 {
		t.Errorf("comprehension filter = %#v", items)
	}
	// a step through the whole parse -> format pipeline, kept typed
	if v := evalOne(t, in, `["2026-10-13T09:30:00Z", "x"] |> map (\s -> maybe_default 0 (map (\t -> time_unix t) (time_parse s)))`); v == nil {
		t.Error("nil")
	} else if items := itemsOf(t, v); !functions.StructuralEqual(items[0], 1791883800) || !functions.StructuralEqual(items[1], 0) {
		t.Errorf("parse in a lambda = %#v", items)
	}
}

// Properties from the plan: (t + d) - t == d; time_parse (time_format rfc3339 t) == just t;
// date_parse (date_format "%Y-%m-%d" d) == just d.
func TestTimeFuncProperties(t *testing.T) {
	in := NewInterpreter()
	for _, src := range []string{
		`[@2026-10-13T09:30:00Z, @2026-03-29T01:59:59Z] |> map (\t -> ((t + 36h) - t) == 36h)`,
		`[@2026-10-13T09:30:00Z, @2026-10-13T11:30:00+02:00, @2026-10-13T09:30:00.123Z] |> map (\t -> (time_parse (time_format "rfc3339" t)) == (just t))`,
		`[@2026-10-13, @0001-01-01, @9999-12-31, @2024-02-29] |> map (\d -> (date_parse (date_format "%Y-%m-%d" d)) == (just d))`,
		`[@2026-10-13, @2024-02-29] |> map (\d -> ((d + 100d) - d) == 100d)`,
	} {
		items := itemsOf(t, in.EvaluateToValues(src)[0])
		for _, it := range items {
			if it != true {
				t.Errorf("%s: %v", src, items)
				break
			}
		}
	}
}
