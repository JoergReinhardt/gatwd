package functions_test

import (
	"strings"
	"testing"
	"time"

	"github.com/joergreinhardt/gatwd/functions"
)

func tm(t *testing.T, text string) functions.Time {
	t.Helper()
	v, err := functions.ParseTime(text)
	if err != nil {
		t.Fatalf("ParseTime(%q): %v", text, err)
	}
	return v
}

func TestParseTime(t *testing.T) {
	for _, tc := range []struct {
		text string
		unix int64
		nano int
		zone string // the display zone's offset in seconds
		off  int
	}{
		{"2026-10-13T09:30:00Z", 1791883800, 0, "UTC", 0},
		{"2026-10-13T11:30:00+02:00", 1791883800, 0, "", 7200},
		{"2026-10-13T04:30:00-05:00", 1791883800, 0, "", -18000},
		{"2026-10-13T09:30:00+00:00", 1791883800, 0, "UTC", 0}, // a zero offset is UTC
		{"2026-10-13T09:30:00.25Z", 1791883800, 250000000, "UTC", 0},
		{"1970-01-01T00:00:00Z", 0, 0, "UTC", 0},
		{"0001-01-01T00:00:00Z", -62135596800, 0, "UTC", 0},
		{"9999-12-31T23:59:59.999999999Z", 253402300799, 999999999, "UTC", 0},
	} {
		got := tm(t, tc.text)
		std := got.Std()
		if std.Unix() != tc.unix || std.Nanosecond() != tc.nano {
			t.Errorf("ParseTime(%q) = %d.%09d, want %d.%09d", tc.text, std.Unix(), std.Nanosecond(), tc.unix, tc.nano)
		}
		if name, off := std.Zone(); off != tc.off {
			t.Errorf("ParseTime(%q) zone offset = %d (%s), want %d", tc.text, off, name, tc.off)
		}
		if std.Location().String() != tc.zone {
			t.Errorf("ParseTime(%q) location = %q, want %q (never the host's Local)", tc.text, std.Location(), tc.zone)
		}
	}
}

func TestParseTimeErrors(t *testing.T) {
	for _, tc := range []struct{ text, want string }{
		{"", "malformed time"},
		{"2026-10-13T09:30:00", "no zone"},
		{"2026-10-13T09:30:00.5", "no zone"},
		{"2026-10-13", "malformed time"},
		{"2026-10-13 09:30:00Z", "malformed time"},
		{"2026-10-13T09:30Z", "malformed time"},
		{"2026-02-30T09:30:00Z", "day out of range"},
		{"2026-13-01T09:30:00Z", "month out of range"},
		{"2026-10-13T25:30:00Z", "hour out of range"},
		{"2026-10-13T09:60:00Z", "minute out of range"},
		{"2026-10-13T09:30:60Z", "second out of range"},
		{"2026-10-13T09:30:00Zjunk", "malformed time"},
		{" 2026-10-13T09:30:00Z", "malformed time"},
		{"2026-10-13t09:30:00z", "malformed time"},
		{"0000-12-31T23:59:59Z", "out of range"},
		{strings.Repeat("2", 100), "longer than"},
	} {
		_, err := functions.ParseTime(tc.text)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("ParseTime(%q) error = %v, want one containing %q", tc.text, err, tc.want)
		}
	}
}

// Property: what a Time prints re-parses to an equal instant (and prints the same).
func TestTimePrintParseRoundTrip(t *testing.T) {
	for _, text := range []string{
		"2026-10-13T09:30:00Z", "2026-10-13T11:30:00+02:00", "2026-10-13T04:30:00-05:00",
		"2026-10-13T09:30:00.25Z", "2026-10-13T09:30:00.000000001Z", "0001-01-01T00:00:00Z",
		"9999-12-31T23:59:59.999999999Z", "2026-10-13T09:30:00+05:45",
	} {
		a := tm(t, text)
		if a.String() != text {
			t.Errorf("String of %q = %q", text, a.String())
		}
		b := tm(t, a.String())
		if a.Compare(b) != 0 || a.String() != b.String() {
			t.Errorf("round trip of %q changed the value: %s", text, b)
		}
	}
}

func TestTimeArithmetic(t *testing.T) {
	base := tm(t, "2026-10-13T09:30:00Z")
	// Time ± Duration is a Time (structural: the raw Go type is functions.Time)
	got := mustEval(t, functions.OpAdd, base, functions.Duration(90*min_))
	if _, ok := got.(functions.Time); !ok || got.(functions.Time).String() != "2026-10-13T11:00:00Z" {
		t.Errorf("t + 90m = %#v", got)
	}
	got = mustEval(t, functions.OpSub, base, functions.Duration(30*day))
	if got.(functions.Time).String() != "2026-09-13T09:30:00Z" {
		t.Errorf("t - 30d = %v", got)
	}
	// Time - Time is a Duration
	other := tm(t, "2026-10-13T08:00:00Z")
	if d := mustEval(t, functions.OpSub, base, other); d != functions.Duration(90*min_) {
		t.Errorf("t - u = %#v, want 1h30m", d)
	}
	if d := mustEval(t, functions.OpSub, other, base); d != functions.Duration(-90*min_) {
		t.Errorf("u - t = %#v, want -1h30m", d)
	}
	// (t + d) - t == d
	for _, d := range []functions.Duration{0, 1, -1, sec, -hour, 30 * day, 100000 * day, -100000 * day} {
		plus := mustEval(t, functions.OpAdd, base, d)
		if back := mustEval(t, functions.OpSub, plus, base); back != d {
			t.Errorf("(t + %v) - t = %v", d, back)
		}
	}
	// the display zone is kept by the arithmetic
	z := tm(t, "2026-10-13T11:30:00+02:00")
	if s := mustEval(t, functions.OpAdd, z, functions.Duration(hour)).(functions.Time).String(); s != "2026-10-13T12:30:00+02:00" {
		t.Errorf("zone lost: %s", s)
	}
	// sub-second parts
	frac := tm(t, "2026-10-13T09:30:00.25Z")
	if d := mustEval(t, functions.OpSub, base, frac); d != functions.Duration(-250e6) {
		t.Errorf("fraction difference = %v", d)
	}
}

func TestTimeOrderingIgnoresZone(t *testing.T) {
	z := tm(t, "2026-10-13T11:30:00+02:00")
	u := tm(t, "2026-10-13T09:30:00Z")
	if !functions.StructuralEqual(z, u) {
		t.Error("the same instant in two zones must be equal")
	}
	if mustEval(t, functions.OpEq, z, u) != true || mustEval(t, functions.OpNeq, z, u) != false {
		t.Error("== / != must compare instants")
	}
	later := tm(t, "2026-10-13T09:30:01Z")
	for _, tc := range []struct {
		op   functions.OpKind
		l, r functions.Time
		want bool
	}{
		{functions.OpLt, u, later, true}, {functions.OpLt, later, u, false}, {functions.OpLt, z, u, false},
		{functions.OpLte, z, u, true}, {functions.OpGt, later, z, true}, {functions.OpGte, z, later, false},
	} {
		if got := mustEval(t, tc.op, tc.l, tc.r); got != tc.want {
			t.Errorf("%v %s %v = %v, want %v", tc.l, tc.op, tc.r, got, tc.want)
		}
	}
	if c, err := functions.Compare(z, later); err != nil || c != -1 {
		t.Errorf("Compare = %d, %v", c, err)
	}
	if functions.StructuralEqual(u, "2026-10-13T09:30:00Z") || functions.StructuralEqual(u, 1791883800) {
		t.Error("a Time never equals its text or its Unix seconds")
	}
	if functions.StructuralEqual(u, functions.Duration(1)) {
		t.Error("a Time never equals a Duration")
	}
}

func TestTimeNegatives(t *testing.T) {
	u := tm(t, "2026-10-13T09:30:00Z")
	last := tm(t, "9999-12-31T23:59:59Z")
	first := tm(t, "0001-01-01T00:00:00Z")
	for _, tc := range []struct {
		name string
		op   functions.OpKind
		l, r any
		want string
	}{
		{"Time + Time", functions.OpAdd, u, u, "two instants cannot be added"},
		{"Time * Int", functions.OpMul, u, 2, "not defined for Time and Int"},
		{"Time / Duration", functions.OpDiv, u, functions.Duration(1), "not defined for Time and Duration"},
		{"Time + Int", functions.OpAdd, u, 1, "not defined for Time and Int"},
		{"Time - Int", functions.OpSub, u, 1, "not defined for Time and Int"},
		{"Int + Time", functions.OpAdd, 1, u, "not defined for Int and Time"},
		{"Duration + Time", functions.OpAdd, functions.Duration(1), u, "write the Time first"},
		{"Duration - Time", functions.OpSub, functions.Duration(1), u, "has no meaning"},
		{"Time + String", functions.OpAdd, u, "x", "not defined"},
		{"Time ** Int", functions.OpPow, u, 2, "not defined"},
		{"Time % Time", functions.OpMod, u, u, "not defined for Time and Time"},
		{"Time < Duration", functions.OpLt, u, functions.Duration(1), "not defined for Time and Duration"},
		{"Time < Int", functions.OpLt, u, 5, "not defined for Time and Int"},
		{"String > Time", functions.OpGt, "x", u, "not defined for String and Time"},
		{"Time == Duration is an error too (two time types)", functions.OpEq, u, functions.Duration(1), "not defined for Time and Duration"},
		{"end of range", functions.OpAdd, last, functions.Duration(sec), "out of range"},
		{"start of range", functions.OpSub, first, functions.Duration(sec), "out of range"},
		{"Time - Time overflow", functions.OpSub, last, first, "Duration overflow"},
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
	// a Time is a point, not a vector: unary minus and abs are errors.
	if _, err := functions.Negate(u); err == nil {
		t.Error("negating a Time must be an error")
	}
	if _, err := functions.Abs(u); err == nil {
		t.Error("abs of a Time must be an error")
	}
	if _, err := functions.NewTime(time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Error("year 10000 must be out of range")
	}
	if _, err := functions.NewTime(time.Date(0, 12, 31, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Error("year 0 must be out of range")
	}
}

// The monotonic clock reading Go attaches to time.Now() must not survive: two instants
// that are equal must be equal as Times.
func TestNewTimeStripsMonotonic(t *testing.T) {
	now := time.Now()
	a, err := functions.NewTime(now)
	if err != nil {
		t.Fatal(err)
	}
	if a.Std() != now.Round(0) {
		t.Error("the monotonic reading must be stripped")
	}
	if !strings.HasPrefix(a.String(), now.UTC().Format("2006")) {
		t.Errorf("String() = %q", a)
	}
}

func TestTimeIsItsOwnType(t *testing.T) {
	var v any = tm(t, "2026-10-13T09:30:00Z")
	if functions.IsNumber(v) || functions.IsNumeric(v) || functions.IsDuration(v) {
		t.Error("a Time is neither a number nor a Duration")
	}
	if !functions.IsTime(v) || !functions.IsTemporal(v) || functions.KindName(v) != "Time" {
		t.Error("IsTime/IsTemporal/KindName")
	}
	if _, ok := v.(time.Time); ok {
		t.Error("a Time is the wrapper, never a bare time.Time, whose printed form is a debugging string")
	}
}

// Daylight saving: an instant plus 24h is not "tomorrow at the same wall-clock time".
func TestTimeDaylightSaving(t *testing.T) {
	berlin, err := functions.LoadZone("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	inBerlin := func(text string) functions.Time {
		v, err := tm(t, text).In(berlin)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	for _, tc := range []struct {
		name, start, want string
		d                 functions.Duration
	}{
		// spring forward 2026-03-29 02:00 CET -> 03:00 CEST: a day is 23 hours long
		{"spring: +24h lands an hour later on the clock", "2026-03-28T12:00:00Z", "2026-03-29T14:00:00+02:00", 24 * hour},
		{"spring: noon + 23h is noon the next day", "2026-03-28T11:00:00Z", "2026-03-29T12:00:00+02:00", 23 * hour},
		// fall back 2026-10-25 03:00 CEST -> 02:00 CET: a day is 25 hours long
		{"fall: +24h lands an hour earlier on the clock", "2026-10-24T10:00:00Z", "2026-10-25T11:00:00+01:00", 24 * hour},
		{"fall: noon + 25h is noon the next day", "2026-10-24T10:00:00Z", "2026-10-25T12:00:00+01:00", 25 * hour},
	} {
		start := inBerlin(tc.start)
		got := mustEval(t, functions.OpAdd, start, tc.d).(functions.Time)
		if got.String() != tc.want {
			t.Errorf("%s: %s + %v = %s, want %s", tc.name, start, tc.d, got, tc.want)
		}
		// the instant arithmetic is exact either way
		if back := mustEval(t, functions.OpSub, got, start); back != tc.d {
			t.Errorf("%s: difference %v, want %v", tc.name, back, tc.d)
		}
	}
	// the same instant, two zones
	a := inBerlin("2026-10-25T00:30:00Z") // 02:30 CEST
	b := inBerlin("2026-10-25T01:30:00Z") // 02:30 CET: the ambiguous wall-clock time happens twice
	if a.String() != "2026-10-25T02:30:00+02:00" || b.String() != "2026-10-25T02:30:00+01:00" {
		t.Errorf("the repeated hour: %s, %s", a, b)
	}
	if a.Compare(b) != -1 {
		t.Error("the two 02:30 instants are an hour apart")
	}
}

func TestLoadZone(t *testing.T) {
	for _, name := range []string{"UTC", "Europe/Berlin", "America/New_York", "Asia/Kolkata", "Australia/Lord_Howe"} {
		if _, err := functions.LoadZone(name); err != nil {
			t.Errorf("LoadZone(%q): %v", name, err)
		}
	}
	for _, tc := range []struct{ name, want string }{
		{"", "never implicit"},
		{"Local", "never implicit"},
		{"Nowhere/Land", "unknown time zone"},
		{"../etc/passwd", "unknown time zone"},
		{"europe/berlin", "unknown time zone"},
	} {
		if _, err := functions.LoadZone(tc.name); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("LoadZone(%q) error = %v, want one containing %q", tc.name, err, tc.want)
		}
	}
}

// The range is checked on the instant (UTC) AND on the displayed year: either one out of
// range is an error, so a printed Time always reads back.
func TestTimeRangeChecksBothYears(t *testing.T) {
	// the instant is in year 10000 although its own zone still shows 9999
	if _, err := functions.ParseTime("9999-12-31T23:30:00-05:00"); err == nil || !strings.Contains(err.Error(), "out of range") {
		t.Errorf("an instant in UTC year 10000: %v", err)
	}
	// the instant is in year 9999 but its zone shows year 10000
	last := tm(t, "9999-12-31T23:30:00Z")
	if _, err := last.In(time.FixedZone("", 2*3600)); err == nil || !strings.Contains(err.Error(), "out of range") {
		t.Errorf("a display year of 10000: %v", err)
	}
	// and the symmetric case at the other end
	if _, err := functions.ParseTime("0001-01-01T00:30:00+02:00"); err == nil {
		t.Error("an instant in UTC year 0 must be out of range")
	}
	first := tm(t, "0001-01-01T00:30:00Z")
	if _, err := first.In(time.FixedZone("", -2*3600)); err == nil {
		t.Error("a display year of 0 must be out of range")
	}
}
