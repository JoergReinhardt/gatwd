package functions_test

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

func TestFormatTime(t *testing.T) {
	berlin, _ := functions.LoadZone("Europe/Berlin")
	u := tm(t, "2026-10-13T09:05:07Z")
	z, err := u.In(berlin)
	if err != nil {
		t.Fatal(err)
	}
	fixed := tm(t, "2026-10-13T09:05:07+05:45")
	for _, tc := range []struct {
		layout string
		t      functions.Time
		want   string
	}{
		{"%Y-%m-%d %H:%M:%S", u, "2026-10-13 09:05:07"},
		{"%Y-%m-%d %H:%M:%S", z, "2026-10-13 11:05:07"},
		{"%z %Z", u, "+0000 UTC"},
		{"%z %Z", z, "+0200 CEST"},
		{"%z %Z", fixed, "+0545 +05:45"},
		{"100%% at %H", u, "100% at 09"},
		{"rfc3339", z, "2026-10-13T11:05:07+02:00"},
		{"date", z, "2026-10-13"},
		{"clock", z, "11:05:07"},
		{"", u, ""},
		{"no directives", u, "no directives"},
		{"%Y%m%d", tm(t, "0001-02-03T00:00:00Z"), "00010203"},
	} {
		got, err := functions.FormatTime(tc.layout, tc.t)
		if err != nil || got != tc.want {
			t.Errorf("FormatTime(%q, %s) = %q, %v; want %q", tc.layout, tc.t, got, err, tc.want)
		}
	}
	for _, bad := range []string{"%", "%q", "%y", "%j", "abc%"} {
		if _, err := functions.FormatTime(bad, u); err == nil {
			t.Errorf("FormatTime(%q) must be an error", bad)
		}
	}
}

func TestFormatDate(t *testing.T) {
	d := dt(t, "2026-03-04")
	for layout, want := range map[string]string{"%Y-%m-%d": "2026-03-04", "%d.%m.%Y": "04.03.2026", "date": "2026-03-04", "%%%Y": "%2026", "": ""} {
		if got, err := functions.FormatDate(layout, d); err != nil || got != want {
			t.Errorf("FormatDate(%q) = %q, %v; want %q", layout, got, err, want)
		}
	}
	for _, bad := range []string{"%H", "%M", "%S", "%z", "%Z", "%q", "%", "clock", "rfc3339"} {
		got, err := functions.FormatDate(bad, d)
		if bad == "clock" || bad == "rfc3339" {
			if err != nil || got != bad {
				t.Errorf("FormatDate(%q) is a literal here: %q, %v", bad, got, err)
			}
			continue
		}
		if err == nil {
			t.Errorf("FormatDate(%q) must be an error", bad)
		}
	}
}

// Property: formatting then parsing returns the same value.
func TestFormatParseRoundTrips(t *testing.T) {
	for _, text := range []string{"2026-10-13", "0001-01-01", "9999-12-31", "2024-02-29"} {
		d := dt(t, text)
		s, err := functions.FormatDate("%Y-%m-%d", d)
		if err != nil || dt(t, s) != d {
			t.Errorf("date round trip of %s via %q: %v", text, s, err)
		}
	}
	for _, text := range []string{"2026-10-13T09:30:00Z", "2026-10-13T11:30:00+02:00", "2026-10-13T09:30:00.123456789Z"} {
		u := tm(t, text)
		s, _ := functions.FormatTime("rfc3339", u)
		if back := tm(t, s); back.Compare(u) != 0 || back.String() != u.String() {
			t.Errorf("time round trip of %s via %q = %s", text, s, back)
		}
	}
}

func TestDurationHelpers(t *testing.T) {
	for _, tc := range []struct {
		x    any
		want functions.Duration
	}{{1.5, 1500e6}, {90, 90 * sec}, {-2, -2 * sec}, {0.0000000019, 1}, {-0.0000000019, -1}} {
		got, err := functions.DurationFromSeconds(tc.x)
		if err != nil || got != tc.want {
			t.Errorf("DurationFromSeconds(%v) = %v, %v; want %v", tc.x, got, err, tc.want)
		}
	}
	for _, bad := range []any{"5", 1e30, 1e10, complex(1, 1)} {
		if _, err := functions.DurationFromSeconds(bad); err == nil {
			t.Errorf("DurationFromSeconds(%v) must be an error", bad)
		}
	}
	for _, tc := range []struct {
		unit, d, trunc, round functions.Duration
	}{
		{min_, 90 * sec, min_, 2 * min_},
		{min_, -90 * sec, -min_, -2 * min_},
		{min_, 89 * sec, min_, min_},
		{min_, 29 * sec, 0, 0},
		{min_, -29 * sec, 0, 0},
		{min_, 120 * sec, 2 * min_, 2 * min_},
		{hour, 3 * hour, 3 * hour, 3 * hour},
		{sec, 1500e6, sec, 2 * sec},
	} {
		if got, err := functions.DurationTruncate(tc.unit, tc.d); err != nil || got != tc.trunc {
			t.Errorf("truncate %v to %v = %v, %v; want %v", tc.d, tc.unit, got, err, tc.trunc)
		}
		if got, err := functions.DurationRound(tc.unit, tc.d); err != nil || got != tc.round {
			t.Errorf("round %v to %v = %v, %v; want %v", tc.d, tc.unit, got, err, tc.round)
		}
	}
	for _, unit := range []functions.Duration{0, -sec} {
		if _, err := functions.DurationTruncate(unit, sec); err == nil {
			t.Errorf("truncate to %v must be an error", unit)
		}
		if _, err := functions.DurationRound(unit, sec); err == nil {
			t.Errorf("round to %v must be an error", unit)
		}
	}
	if _, err := functions.DurationRound(functions.Duration(1<<62), functions.Duration(1<<62+1<<61)); err == nil || !strings.Contains(err.Error(), "overflow") {
		t.Errorf("rounding past the range: %v", err)
	}
}

func TestTimeFieldsAndUnix(t *testing.T) {
	u := tm(t, "2026-10-18T23:59:58.5Z") // a Sunday
	for field, want := range map[string]int64{"year": 2026, "month": 10, "day": 18, "hour": 23, "minute": 59, "second": 58, "nano": 5e8, "weekday": 7} {
		if got := functions.TimeField(field, u); got != want {
			t.Errorf("TimeField(%s) = %d, want %d", field, got, want)
		}
	}
	berlin, _ := functions.LoadZone("Europe/Berlin")
	z, _ := u.In(berlin) // already Monday in Berlin
	if functions.TimeField("weekday", z) != 1 || functions.TimeField("day", z) != 19 {
		t.Error("calendar fields read the display zone")
	}
	if functions.TimeUnix(u) != functions.TimeUnix(z) {
		t.Error("the Unix time is zone-independent")
	}
	for _, n := range []int64{0, 1791883800, -1, -62135596800, 253402300799} {
		back, err := functions.TimeFromUnix(n)
		if err != nil || functions.TimeUnix(back) != n {
			t.Errorf("unix round trip of %d = %v, %v", n, back, err)
		}
	}
	for _, n := range []int64{-62135596801, 253402300800, 1 << 60} {
		if _, err := functions.TimeFromUnix(n); err == nil {
			t.Errorf("TimeFromUnix(%d) must be out of range", n)
		}
	}
	// negative Unix seconds floor: 0.5 s before the epoch is second -1
	if tt, _ := functions.ParseTime("1969-12-31T23:59:59.5Z"); functions.TimeUnix(tt) != -1 {
		t.Error("TimeUnix floors")
	}
}

func TestTimeTruncateRound(t *testing.T) {
	u := tm(t, "2026-10-13T09:45:30.5Z")
	for _, tc := range []struct {
		d            functions.Duration
		trunc, round string
	}{
		{hour, "2026-10-13T09:00:00Z", "2026-10-13T10:00:00Z"},
		{min_, "2026-10-13T09:45:00Z", "2026-10-13T09:46:00Z"},
		{sec, "2026-10-13T09:45:30Z", "2026-10-13T09:45:31Z"},
		{day, "2026-10-13T00:00:00Z", "2026-10-13T00:00:00Z"},
	} {
		if got, err := functions.TimeTruncate(tc.d, u); err != nil || got.String() != tc.trunc {
			t.Errorf("truncate %v = %v, %v; want %s", tc.d, got, err, tc.trunc)
		}
		if got, err := functions.TimeRound(tc.d, u); err != nil || got.String() != tc.round {
			t.Errorf("round %v = %v, %v; want %s", tc.d, got, err, tc.round)
		}
	}
	// 1d is UTC midnight whatever the display zone (documented)
	berlin, _ := functions.LoadZone("Europe/Berlin")
	z, _ := u.In(berlin)
	if got, _ := functions.TimeTruncate(day, z); got.String() != "2026-10-13T02:00:00+02:00" {
		t.Errorf("truncate 1d in Berlin = %s, want UTC midnight shown as 02:00+02:00", got)
	}
	for _, d := range []functions.Duration{0, -sec} {
		if _, err := functions.TimeTruncate(d, u); err == nil {
			t.Errorf("TimeTruncate(%v) must be an error", d)
		}
		if _, err := functions.TimeRound(d, u); err == nil {
			t.Errorf("TimeRound(%v) must be an error", d)
		}
	}
	// the first instant: truncating cannot leave the range; rounding the last can
	if _, err := functions.TimeRound(day, tm(t, "9999-12-31T23:59:59Z")); err == nil {
		t.Error("rounding the last second up to year 10000 must be an error")
	}
}

// DST (Europe/Berlin 2026-03-29 spring forward at 02:00, 2026-10-25 fall back at 03:00):
// date_of_time, time_of_date and the 24h versus whole-day difference.
func TestDateTimeBridgeAcrossDaylightSaving(t *testing.T) {
	berlin, _ := functions.LoadZone("Europe/Berlin")
	for _, tc := range []struct{ date, want string }{
		{"2026-03-28", "2026-03-28T00:00:00+01:00"},
		{"2026-03-29", "2026-03-29T00:00:00+01:00"}, // the gap is at 02:00, midnight exists
		{"2026-03-30", "2026-03-30T00:00:00+02:00"},
		{"2026-10-24", "2026-10-24T00:00:00+02:00"},
		{"2026-10-25", "2026-10-25T00:00:00+02:00"},
		{"2026-10-26", "2026-10-26T00:00:00+01:00"},
	} {
		got, err := functions.TimeOfDate(berlin, dt(t, tc.date))
		if err != nil || got.String() != tc.want {
			t.Errorf("TimeOfDate(Berlin, %s) = %v, %v; want %s", tc.date, got, err, tc.want)
		}
	}
	// the day lengths: 23h at the spring change, 25h at the autumn change, 24h otherwise
	for _, tc := range []struct {
		date string
		want functions.Duration
	}{{"2026-03-28", 24 * hour}, {"2026-03-29", 23 * hour}, {"2026-10-24", 24 * hour}, {"2026-10-25", 25 * hour}} {
		a, _ := functions.TimeOfDate(berlin, dt(t, tc.date))
		next, _ := dt(t, tc.date).AddDays(1)
		b, _ := functions.TimeOfDate(berlin, next)
		if got := mustEval(t, functions.OpSub, b, a); got != tc.want {
			t.Errorf("the day %s lasts %v, want %v", tc.date, got, tc.want)
		}
	}
	// date_add_days 1 is "tomorrow"; + 24h is not, across the change
	start, _ := functions.TimeOfDate(berlin, dt(t, "2026-03-29"))
	plus24 := mustEval(t, functions.OpAdd, start, functions.Duration(24*hour)).(functions.Time)
	if d, _ := functions.DateOfTime(berlin, plus24); d.String() != "2026-03-30" {
		t.Errorf("midnight + 24h is on %s", d)
	}
	if plus24.String() != "2026-03-30T01:00:00+02:00" {
		t.Errorf("midnight + 24h = %s: not midnight tomorrow", plus24)
	}
	// date_of_time reads the zone: 23:30 UTC on the 24th is already the 25th in Berlin
	late := tm(t, "2026-10-24T23:30:00Z")
	if d, _ := functions.DateOfTime(berlin, late); d.String() != "2026-10-25" {
		t.Errorf("DateOfTime = %s", d)
	}
	utc, _ := functions.LoadZone("UTC")
	if d, _ := functions.DateOfTime(utc, late); d.String() != "2026-10-24" {
		t.Errorf("DateOfTime(UTC) = %s", d)
	}
	// the repeated hour: 02:30 happens twice on 2026-10-25, an hour apart, same date
	first := tm(t, "2026-10-25T00:30:00Z")
	second := tm(t, "2026-10-25T01:30:00Z")
	d1, _ := functions.DateOfTime(berlin, first)
	d2, _ := functions.DateOfTime(berlin, second)
	if d1 != d2 || d1.String() != "2026-10-25" {
		t.Errorf("the repeated hour: %s %s", d1, d2)
	}
}

// A midnight swallowed by a gap resolves to the first instant after it; a midnight that
// repeats resolves to the earlier one.
func TestTimeOfDateMidnightGapAndOverlap(t *testing.T) {
	beirut, err := functions.LoadZone("Asia/Beirut") // 2026-03-29 00:00 jumps to 01:00
	if err != nil {
		t.Fatal(err)
	}
	got, err := functions.TimeOfDate(beirut, dt(t, "2026-03-29"))
	if err != nil || got.String() != "2026-03-29T01:00:00+03:00" {
		t.Errorf("gap at midnight: %v, %v; want the first valid instant 01:00+03:00", got, err)
	}
	havana, err := functions.LoadZone("America/Havana") // 2026-11-01 01:00 CDT falls back to 00:00 CST
	if err != nil {
		t.Fatal(err)
	}
	got, err = functions.TimeOfDate(havana, dt(t, "2026-11-01"))
	if err != nil || got.String() != "2026-11-01T00:00:00-04:00" {
		t.Errorf("repeated midnight: %v, %v; want the earlier 00:00-04:00", got, err)
	}
	// Property: for ordinary zones and many days the result's own date is the date asked for
	// and it is the first instant of that day.
	for _, zone := range []string{"UTC", "Europe/Berlin", "America/New_York", "Asia/Kolkata", "Australia/Lord_Howe", "Pacific/Apia"} {
		loc, _ := functions.LoadZone(zone)
		d := dt(t, "2026-01-01")
		for i := 0; i < 400; i++ {
			at, err := functions.TimeOfDate(loc, d)
			if err != nil {
				t.Fatalf("%s %s: %v", zone, d, err)
			}
			if back, _ := functions.DateOfTime(loc, at); back != d {
				t.Errorf("%s %s: TimeOfDate lands on %s", zone, d, back)
			}
			prev := mustEval(t, functions.OpSub, at, functions.Duration(1)).(functions.Time)
			if back, _ := functions.DateOfTime(loc, prev); back == d {
				t.Errorf("%s %s: %s is not the first instant of the day (1ns earlier is still the day)", zone, d, at)
			}
			d, _ = d.AddDays(1)
		}
	}
}
