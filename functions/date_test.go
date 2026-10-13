package functions_test

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

func dt(t *testing.T, text string) functions.Date {
	t.Helper()
	d, err := functions.ParseDate(text)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", text, err)
	}
	return d
}

func TestNewDateAndParseDate(t *testing.T) {
	for _, tc := range []struct {
		text string
		want functions.Date
	}{
		{"2026-10-13", functions.Date{Year: 2026, Month: 10, Day: 13}},
		{"0001-01-01", functions.Date{Year: 1, Month: 1, Day: 1}},
		{"9999-12-31", functions.Date{Year: 9999, Month: 12, Day: 31}},
		{"2024-02-29", functions.Date{Year: 2024, Month: 2, Day: 29}},
	} {
		got := dt(t, tc.text)
		if got != tc.want || got.String() != tc.text {
			t.Errorf("ParseDate(%q) = %#v (%s)", tc.text, got, got)
		}
	}
	for _, tc := range []struct{ text, want string }{
		{"", "malformed date"},
		{"2026-02-30", "day out of range"},
		{"2026-02-29", "day out of range"},
		{"2100-02-29", "day out of range"}, // 2100 is not a leap year
		{"2026-13-01", "month out of range"},
		{"2026-00-10", "month out of range"},
		{"2026-10-00", "day out of range"},
		{"2026-1-3", "malformed date"},
		{"26-10-13", "malformed date"},
		{"2026/10/13", "malformed date"},
		{"2026-10-13T00:00:00Z", "malformed date"},
		{" 2026-10-13", "malformed date"},
		{"0000-01-01", "out of range"},
	} {
		if _, err := functions.ParseDate(tc.text); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("ParseDate(%q) error = %v, want one containing %q", tc.text, err, tc.want)
		}
	}
	// NewDate names the field
	for _, tc := range []struct {
		y, m, d int
		want    string
	}{
		{2026, 13, 1, "month 13"}, {2026, 0, 1, "month 0"}, {2026, 2, 30, "day 30"}, {2026, 4, 31, "day 31"},
		{0, 1, 1, "year 0"}, {10000, 1, 1, "year 10000"}, {2026, 1, 0, "day 0"},
	} {
		if _, err := functions.NewDate(tc.y, tc.m, tc.d); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("NewDate(%d,%d,%d) error = %v, want one containing %q", tc.y, tc.m, tc.d, err, tc.want)
		}
	}
}

func TestLeapYearsAndMonthLengths(t *testing.T) {
	for y, want := range map[int]bool{2000: true, 1900: false, 2024: true, 2026: false, 2100: false, 400: true, 4: true, 1: false} {
		if functions.IsLeapYear(y) != want {
			t.Errorf("IsLeapYear(%d) = %v", y, !want)
		}
	}
	for m, want := range map[int]int{1: 31, 2: 28, 3: 31, 4: 30, 5: 31, 6: 30, 7: 31, 8: 31, 9: 30, 10: 31, 11: 30, 12: 31, 0: 0, 13: 0} {
		if got := functions.DaysInMonth(2026, m); got != want {
			t.Errorf("DaysInMonth(2026, %d) = %d, want %d", m, got, want)
		}
	}
	if functions.DaysInMonth(2024, 2) != 29 {
		t.Error("February 2024 has 29 days")
	}
}

func TestDateArithmeticWithDurations(t *testing.T) {
	d := dt(t, "2026-10-13")
	for _, tc := range []struct {
		op   functions.OpKind
		r    functions.Duration
		want string
	}{
		{functions.OpAdd, 7 * day, "2026-10-20"},
		{functions.OpAdd, 0, "2026-10-13"},
		{functions.OpAdd, -7 * day, "2026-10-06"},
		{functions.OpSub, 7 * day, "2026-10-06"},
		{functions.OpSub, -7 * day, "2026-10-20"},
		{functions.OpAdd, 7 * 7 * day, "2026-12-01"}, // 7w
		{functions.OpAdd, 19 * day, "2026-11-01"},
		{functions.OpAdd, 366 * day, "2027-10-14"},
		{functions.OpSub, 285 * day, "2026-01-01"},
	} {
		got := mustEval(t, tc.op, d, tc.r)
		if gd, ok := got.(functions.Date); !ok || gd.String() != tc.want {
			t.Errorf("%v %s %v = %#v, want the Date %s", d, tc.op, tc.r, got, tc.want)
		}
	}
	// month and year ends
	for _, tc := range []struct {
		start string
		n     int64
		want  string
	}{
		{"2026-02-28", 1, "2026-03-01"}, {"2024-02-28", 1, "2024-02-29"}, {"2024-02-29", 1, "2024-03-01"},
		{"2026-12-31", 1, "2027-01-01"}, {"2027-01-01", -1, "2026-12-31"}, {"0001-01-02", -1, "0001-01-01"},
	} {
		got := mustEval(t, functions.OpAdd, dt(t, tc.start), functions.Duration(tc.n)*day).(functions.Date)
		if got.String() != tc.want {
			t.Errorf("%s + %dd = %s, want %s", tc.start, tc.n, got, tc.want)
		}
	}
	// Date - Date is a Duration, a multiple of 1d
	if got := mustEval(t, functions.OpSub, dt(t, "2026-10-13"), dt(t, "2026-10-06")); got != functions.Duration(7*day) {
		t.Errorf("d - e = %#v, want 7d", got)
	}
	if got := mustEval(t, functions.OpSub, dt(t, "2026-10-06"), dt(t, "2026-10-13")); got != functions.Duration(-7*day) {
		t.Errorf("e - d = %#v, want -7d", got)
	}
	if got := mustEval(t, functions.OpSub, d, d); got != functions.Duration(0) {
		t.Errorf("d - d = %#v", got)
	}
	// (d + n days) - d == n days
	for _, n := range []int64{0, 1, -1, 28, 365, -365, 1000, -1000, 10000} {
		plus := mustEval(t, functions.OpAdd, d, functions.Duration(n)*day)
		if back := mustEval(t, functions.OpSub, plus, d); back != functions.Duration(n)*day {
			t.Errorf("(d + %dd) - d = %v", n, back)
		}
	}
}

func TestDateNegatives(t *testing.T) {
	d := dt(t, "2026-10-13")
	u := tm(t, "2026-10-13T09:30:00Z")
	first, last := dt(t, "0001-01-01"), dt(t, "9999-12-31")
	for _, tc := range []struct {
		name string
		op   functions.OpKind
		l, r any
		want string
	}{
		{"Date + 90m", functions.OpAdd, d, functions.Duration(90 * min_), "whole days"},
		{"Date - 36h", functions.OpSub, d, functions.Duration(36 * hour), "whole days"},
		{"Date + 1ns", functions.OpAdd, d, functions.Duration(1), "whole days"},
		{"Date + Date", functions.OpAdd, d, d, "two days cannot be added"},
		{"Date * Int", functions.OpMul, d, 2, "not defined for Date and Int"},
		{"Date + Int", functions.OpAdd, d, 1, "not defined for Date and Int"},
		{"Int + Date", functions.OpAdd, 1, d, "not defined for Int and Date"},
		{"Duration + Date", functions.OpAdd, functions.Duration(day), d, "write the Date first"},
		{"Duration - Date", functions.OpSub, functions.Duration(day), d, "write the Date first"},
		{"Date < Time", functions.OpLt, d, u, "convert with date_of_time or time_of_date"},
		{"Time < Date", functions.OpLt, u, d, "convert with date_of_time or time_of_date"},
		{"Date == Time", functions.OpEq, d, u, "no zone"},
		{"Date != Time", functions.OpNeq, d, u, "no zone"},
		{"Date - Time", functions.OpSub, d, u, "not defined for Date and Time"},
		{"Time - Date", functions.OpSub, u, d, "not defined for Time and Date"},
		{"Date < Int", functions.OpLt, d, 5, "not defined for Date and Int"},
		{"Date < String", functions.OpLt, d, "2026-10-14", "not defined for Date and String"},
		{"Date / Int", functions.OpDiv, d, 2, "not defined"},
		{"Date % Duration", functions.OpMod, d, functions.Duration(day), "not defined"},
		{"end of range", functions.OpAdd, last, functions.Duration(day), "out of range"},
		{"start of range", functions.OpSub, first, functions.Duration(day), "out of range"},
		{"Date - Date overflow", functions.OpSub, last, first, "Duration overflow"},
	} {
		got, err := eval(t, tc.op, tc.l, tc.r)
		if err == nil {
			t.Errorf("%s: %v %s %v = %#v, want an error containing %q", tc.name, tc.l, tc.op, tc.r, got, tc.want)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %q does not contain %q", tc.name, err, tc.want)
		}
	}
	if _, err := functions.Negate(d); err == nil {
		t.Error("negating a Date must be an error")
	}
	if _, err := functions.Abs(d); err == nil {
		t.Error("abs of a Date must be an error")
	}
}

func TestDateEqualityAndOrder(t *testing.T) {
	a, b := dt(t, "2026-10-13"), dt(t, "2026-10-14")
	if !functions.StructuralEqual(a, dt(t, "2026-10-13")) || functions.StructuralEqual(a, b) {
		t.Error("Date equality is structural")
	}
	if functions.StructuralEqual(a, "2026-10-13") || functions.StructuralEqual(a, 20261013) {
		t.Error("a Date never equals its text or a number")
	}
	for _, tc := range []struct {
		op   functions.OpKind
		l, r functions.Date
		want bool
	}{
		{functions.OpLt, a, b, true}, {functions.OpLt, b, a, false}, {functions.OpLt, a, a, false},
		{functions.OpLte, a, a, true}, {functions.OpGt, b, a, true}, {functions.OpGte, a, b, false},
		{functions.OpEq, a, a, true}, {functions.OpNeq, a, b, true},
		{functions.OpLt, dt(t, "2025-12-31"), a, true}, // year decides before month
		{functions.OpLt, dt(t, "2026-09-30"), a, true}, // month decides before day
		{functions.OpGt, dt(t, "2026-11-01"), a, true},
	} {
		if got := mustEval(t, tc.op, tc.l, tc.r); got != tc.want {
			t.Errorf("%v %s %v = %v, want %v", tc.l, tc.op, tc.r, got, tc.want)
		}
	}
	if c, err := functions.Compare(a, b); err != nil || c != -1 {
		t.Errorf("Compare = %d, %v", c, err)
	}
	if _, err := functions.Compare(a, 1); err == nil {
		t.Error("Compare(Date, Int) must be an error")
	}
	if _, err := functions.Compare(a, tm(t, "2026-10-13T00:00:00Z")); err == nil {
		t.Error("Compare(Date, Time) must be an error")
	}
	// Date is comparable, so it keys maps directly
	m := map[functions.Date]int{a: 1, b: 2}
	if m[dt(t, "2026-10-14")] != 2 {
		t.Error("Date must be a usable map key")
	}
}

func TestDateAddMonthsAndYears(t *testing.T) {
	for _, tc := range []struct {
		start string
		n     int64
		want  string
	}{
		{"2026-01-31", 1, "2026-02-28"},  // clamps to the end of the month
		{"2024-01-31", 1, "2024-02-29"},  // leap year
		{"2026-01-31", 2, "2026-03-31"},  // does not clamp when the day exists
		{"2026-03-31", -1, "2026-02-28"}, // backwards
		{"2026-10-13", 3, "2027-01-13"},  // across a year
		{"2026-10-13", -10, "2025-12-13"},
		{"2026-10-13", 0, "2026-10-13"},
		{"2026-10-13", 12, "2027-10-13"},
		{"2026-10-13", -12, "2025-10-13"},
		{"2026-12-31", 1, "2027-01-31"},
		{"2026-11-30", 3, "2027-02-28"},
		{"0001-02-01", -1, "0001-01-01"},
	} {
		got, err := dt(t, tc.start).AddMonths(tc.n)
		if err != nil || got.String() != tc.want {
			t.Errorf("%s add_months %d = %v, %v; want %s", tc.start, tc.n, got, err, tc.want)
		}
	}
	// property: adding months never skips a month, and the day is the clamped original
	for m := int64(-30); m <= 30; m++ {
		for _, start := range []string{"2026-01-31", "2024-02-29", "2026-05-15", "2026-12-31", "2000-03-31"} {
			s := dt(t, start)
			got, err := s.AddMonths(m)
			if err != nil {
				t.Fatalf("%s + %d months: %v", start, m, err)
			}
			wantTotal := int64(s.Year)*12 + int64(s.Month-1) + m
			gotTotal := int64(got.Year)*12 + int64(got.Month-1)
			if gotTotal != wantTotal {
				t.Errorf("%s + %d months = %s: skipped or repeated a month", start, m, got)
			}
			if want := min(s.Day, functions.DaysInMonth(got.Year, got.Month)); got.Day != want {
				t.Errorf("%s + %d months = %s: day should be %d", start, m, got, want)
			}
		}
	}
	for _, tc := range []struct {
		start string
		n     int64
		want  string
	}{
		{"2024-02-29", 1, "2025-02-28"}, {"2024-02-29", 4, "2028-02-29"}, {"2026-10-13", -26, "2000-10-13"}, {"2026-10-13", 0, "2026-10-13"},
	} {
		got, err := dt(t, tc.start).AddYears(tc.n)
		if err != nil || got.String() != tc.want {
			t.Errorf("%s add_years %d = %v, %v; want %s", tc.start, tc.n, got, err, tc.want)
		}
	}
	for _, tc := range []struct {
		name string
		f    func() (functions.Date, error)
	}{
		{"before year 1", func() (functions.Date, error) { return dt(t, "0001-01-15").AddMonths(-1) }},
		{"after year 9999", func() (functions.Date, error) { return dt(t, "9999-12-15").AddMonths(1) }},
		{"months near int64 max", func() (functions.Date, error) { return dt(t, "2026-01-01").AddMonths(1<<63 - 1) }},
		{"huge months", func() (functions.Date, error) { return dt(t, "2026-01-01").AddMonths(1 << 40) }},
		{"huge years", func() (functions.Date, error) { return dt(t, "2026-01-01").AddYears(-(1 << 40)) }},
		{"huge days", func() (functions.Date, error) { return dt(t, "2026-01-01").AddDays(1 << 50) }},
		{"year 10000", func() (functions.Date, error) { return dt(t, "9999-01-01").AddYears(1) }},
	} {
		if got, err := tc.f(); err == nil || !strings.Contains(err.Error(), "out of range") {
			t.Errorf("%s: got %v, %v; want an out-of-range error", tc.name, got, err)
		}
	}
}

func TestDateWeekday(t *testing.T) {
	for text, want := range map[string]int{
		"2026-10-12": 1, "2026-10-13": 2, "2026-10-14": 3, "2026-10-15": 4, "2026-10-16": 5,
		"2026-10-17": 6, "2026-10-18": 7, "2000-01-01": 6, "1970-01-01": 4, "0001-01-01": 1,
	} {
		if got := dt(t, text).Weekday(); got != want {
			t.Errorf("Weekday(%s) = %d, want %d", text, got, want)
		}
	}
}

func TestDateIsItsOwnType(t *testing.T) {
	var v any = dt(t, "2026-10-13")
	if functions.IsNumber(v) || functions.IsDuration(v) || functions.IsTime(v) {
		t.Error("a Date is neither a number, a Duration nor a Time")
	}
	if !functions.IsDate(v) || !functions.IsTemporal(v) || functions.KindName(v) != "Date" {
		t.Error("IsDate/IsTemporal/KindName")
	}
}
