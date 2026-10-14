package functions

import (
	"fmt"
	"math/big"
	"strings"
	"time"
)

// This file holds the pure helpers behind the time function forms: the strftime-subset
// formatter, zone conversion, truncation and rounding, and the Date/Time bridge.

// ---- formatting ---------------------------------------------------------------------
//
// time_format and date_format take a layout in a small strftime subset. Go's own layouts
// are not exposed: they spell fields with the magic reference date (Mon Jan 2 15:04:05
// 2006), which nobody remembers and which fails silently (a wrong digit prints a literal).
//
//	%Y year (4 digits)        %m month (01-12)       %d day (01-31)
//	%H hour (00-23)           %M minute (00-59)      %S second (00-59)
//	%z offset +hhmm           %Z zone name or +hh:mm %% a percent sign
//
// and the names `rfc3339` (RFC 3339 with fractional seconds when present), `date` (%Y-%m-%d)
// and `clock` (%H:%M:%S). Every other character is copied; an unknown `%x` directive is an
// error. A Date has no time of day and no zone, so date_format accepts only %Y %m %d %%.

// FormatTime renders t with the layout (see the table above).
func FormatTime(layout string, t Time) (string, error) {
	switch layout {
	case "rfc3339":
		return t.String(), nil
	case "date":
		layout = "%Y-%m-%d"
	case "clock":
		layout = "%H:%M:%S"
	}
	std := t.t
	var sb strings.Builder
	for i := 0; i < len(layout); i++ {
		c := layout[i]
		if c != '%' {
			sb.WriteByte(c)
			continue
		}
		i++
		if i >= len(layout) {
			return "", fmt.Errorf("layout %q ends with a lone %%", layout)
		}
		switch layout[i] {
		case 'Y':
			fmt.Fprintf(&sb, "%04d", std.Year())
		case 'm':
			fmt.Fprintf(&sb, "%02d", int(std.Month()))
		case 'd':
			fmt.Fprintf(&sb, "%02d", std.Day())
		case 'H':
			fmt.Fprintf(&sb, "%02d", std.Hour())
		case 'M':
			fmt.Fprintf(&sb, "%02d", std.Minute())
		case 'S':
			fmt.Fprintf(&sb, "%02d", std.Second())
		case 'z':
			_, off := std.Zone()
			sb.WriteString(formatOffset(off, false))
		case 'Z':
			name, off := std.Zone()
			if name == "" {
				name = formatOffset(off, true)
			}
			sb.WriteString(name)
		case '%':
			sb.WriteByte('%')
		default:
			return "", fmt.Errorf("unknown layout directive %%%c in %q (directives: %%Y %%m %%d %%H %%M %%S %%z %%Z %%%%; names: rfc3339 date clock)", layout[i], layout)
		}
	}
	return sb.String(), nil
}

func formatOffset(off int, colon bool) string {
	sign := '+'
	if off < 0 {
		sign, off = '-', -off
	}
	if colon {
		return fmt.Sprintf("%c%02d:%02d", sign, off/3600, off%3600/60)
	}
	return fmt.Sprintf("%c%02d%02d", sign, off/3600, off%3600/60)
}

// FormatDate renders d with a layout of %Y %m %d %% (or the name `date`).
func FormatDate(layout string, d Date) (string, error) {
	if layout == "date" {
		return d.String(), nil
	}
	var sb strings.Builder
	for i := 0; i < len(layout); i++ {
		c := layout[i]
		if c != '%' {
			sb.WriteByte(c)
			continue
		}
		i++
		if i >= len(layout) {
			return "", fmt.Errorf("layout %q ends with a lone %%", layout)
		}
		switch layout[i] {
		case 'Y':
			fmt.Fprintf(&sb, "%04d", d.Year)
		case 'm':
			fmt.Fprintf(&sb, "%02d", d.Month)
		case 'd':
			fmt.Fprintf(&sb, "%02d", d.Day)
		case '%':
			sb.WriteByte('%')
		case 'H', 'M', 'S', 'z', 'Z':
			return "", fmt.Errorf("directive %%%c needs a time of day or a zone, which a Date does not have (a Date takes %%Y %%m %%d %%%%)", layout[i])
		default:
			return "", fmt.Errorf("unknown layout directive %%%c in %q (a Date takes %%Y %%m %%d %%%%, or the name date)", layout[i], layout)
		}
	}
	return sb.String(), nil
}

// ---- Duration helpers -----------------------------------------------------------------

// DurationFromSeconds is the Duration of x seconds (an Int, Ratio or finite Float),
// truncated toward zero to whole nanoseconds; an out-of-range or non-finite value is an error.
func DurationFromSeconds(x any) (Duration, error) {
	f, err := scaleFactor(x)
	if err != nil {
		return 0, fmt.Errorf("want a number of seconds: %w", err)
	}
	return durationFromRat(new(big.Rat).Mul(f, big.NewRat(1e9, 1)), "dur_from_seconds")
}

// Seconds is the Duration as a Float number of seconds.
func (d Duration) Seconds() float64 { return time.Duration(d).Seconds() }

func checkGranularity(unit Duration) error {
	if unit <= 0 {
		return fmt.Errorf("the granularity must be a positive Duration, got %s", unit)
	}
	return nil
}

// DurationTruncate rounds d toward zero to a multiple of unit.
func DurationTruncate(unit, d Duration) (Duration, error) {
	if err := checkGranularity(unit); err != nil {
		return 0, err
	}
	return d - d%unit, nil
}

// DurationRound rounds d to the nearest multiple of unit, halves away from zero; a result
// beyond the Duration range is an error.
func DurationRound(unit, d Duration) (Duration, error) {
	if err := checkGranularity(unit); err != nil {
		return 0, err
	}
	rem := d % unit
	base := d - rem
	r := rem
	if r < 0 {
		r = -r
	}
	if r >= unit-r { // at least half: away from zero
		step := unit
		if d < 0 {
			step = -unit
		}
		res := base + step
		if (step > 0 && res < base) || (step < 0 && res > base) {
			return 0, fmt.Errorf("Duration overflow rounding %s to a multiple of %s", d, unit)
		}
		return res, nil
	}
	return base, nil
}

// ---- Time helpers -----------------------------------------------------------------------

// TimeUnix is the whole seconds since the Unix epoch (floor, so it is monotone).
func TimeUnix(t Time) int64 { return t.t.Unix() }

// TimeFromUnix is the instant n seconds after the Unix epoch (displayed in UTC).
func TimeFromUnix(n int64) (Time, error) {
	if n < -62135596800 || n > 253402300799 {
		return Time{}, fmt.Errorf("time out of range: %d seconds is outside the years %d to %d", n, minYear, maxYear)
	}
	return NewTime(time.Unix(n, 0).UTC())
}

// TimeField reads one calendar field of t in its display zone: year, month, day, hour,
// minute, second, nano or weekday (ISO: 1 Monday .. 7 Sunday).
func TimeField(field string, t Time) int64 {
	s := t.t
	switch field {
	case "year":
		return int64(s.Year())
	case "month":
		return int64(s.Month())
	case "day":
		return int64(s.Day())
	case "hour":
		return int64(s.Hour())
	case "minute":
		return int64(s.Minute())
	case "second":
		return int64(s.Second())
	case "nano":
		return int64(s.Nanosecond())
	case "weekday":
		if w := int64(s.Weekday()); w != 0 {
			return w
		}
		return 7
	}
	panic("functions.TimeField: unknown field " + field)
}

// TimeTruncate rounds t down to a multiple of d counted from the zero instant of the
// standard library (0001-01-01T00:00:00Z); the display zone is kept. Because the count
// starts at a UTC instant, `1d` truncates to UTC midnight whatever the display zone, and
// `1h` to a whole hour of UTC (a zone with a 30 or 45 minute offset sees a half hour).
func TimeTruncate(d Duration, t Time) (Time, error) {
	if err := checkGranularity(d); err != nil {
		return Time{}, err
	}
	return NewTime(t.t.Truncate(time.Duration(d)))
}

// TimeRound is TimeTruncate to the nearest multiple of d (halves round up).
func TimeRound(d Duration, t Time) (Time, error) {
	if err := checkGranularity(d); err != nil {
		return Time{}, err
	}
	return NewTime(t.t.Round(time.Duration(d)))
}

// ---- the Date / Time bridge ---------------------------------------------------------

// DateOfTime is the calendar day of t in the zone loc.
func DateOfTime(loc *time.Location, t Time) (Date, error) {
	return dateOfStd(t.t.In(loc))
}

// TimeOfDate is the first instant of the day d in the zone loc (its midnight), displayed
// in loc. When a daylight-saving gap swallows midnight (the clock jumps from 23:59 to
// 01:00) it is the first valid instant after the gap; when the clock repeats midnight it
// is the earlier of the two.
func TimeOfDate(loc *time.Location, d Date) (Time, error) {
	wall := d.std().Unix() // the wall clock 00:00 as if it were UTC
	cand := time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, loc)
	// the offsets in force around that day: the candidate's and the neighbours' (a
	// transition near midnight changes it)
	offsets := map[int]bool{}
	for _, at := range []time.Time{cand, cand.Add(-24 * time.Hour), cand.Add(24 * time.Hour)} {
		_, off := at.Zone()
		offsets[off] = true
	}
	var best time.Time
	found := false
	for off := range offsets {
		inst := time.Unix(wall-int64(off), 0).In(loc)
		_, got := inst.Zone()
		if inst.Unix()+int64(got) < wall {
			continue // the wall clock there is before midnight: not the day's first instant
		}
		if !found || inst.Before(best) {
			best, found = inst, true
		}
	}
	if !found {
		return Time{}, fmt.Errorf("no midnight for %s in %s", d, loc)
	}
	return NewTime(best)
}

// ---- parsing with a layout ----------------------------------------------------------

// ParseTimeWith reads text with the strftime-subset layout FormatTime writes: %Y %m %d %H
// %M %S, %z (an offset such as +0200, +02:00 or Z) and %%; every other layout character must
// match literally. The layout must carry %Y %m %d and %z: a missing date field or zone is
// an error, never a default (local time is never implicit).
func ParseTimeWith(layout, text string) (Time, error) {
	if len(text) > 64 || len(layout) > 64 {
		return Time{}, fmt.Errorf("time text and layout are limited to 64 characters")
	}
	fields := map[byte]int{}
	zoneOff, haveZone := 0, false
	pos := 0
	digits := func(n int) (int, error) {
		if pos+n > len(text) {
			return 0, fmt.Errorf("text %q ends early (want %d digits at offset %d)", text, n, pos)
		}
		v := 0
		for _, c := range []byte(text[pos : pos+n]) {
			if c < '0' || c > '9' {
				return 0, fmt.Errorf("text %q: want %d digits at offset %d", text, n, pos)
			}
			v = v*10 + int(c-'0')
		}
		pos += n
		return v, nil
	}
	for i := 0; i < len(layout); i++ {
		c := layout[i]
		if c != '%' {
			if pos >= len(text) || text[pos] != c {
				return Time{}, fmt.Errorf("text %q does not match layout %q at offset %d (want %q)", text, layout, pos, string(c))
			}
			pos++
			continue
		}
		i++
		if i >= len(layout) {
			return Time{}, fmt.Errorf("layout %q ends with a lone %%", layout)
		}
		d := layout[i]
		var err error
		switch d {
		case 'Y':
			fields[d], err = digits(4)
		case 'm', 'd', 'H', 'M', 'S':
			fields[d], err = digits(2)
		case 'z':
			rest := text[pos:]
			switch {
			case strings.HasPrefix(rest, "Z"):
				pos++
			case len(rest) >= 5 && (rest[0] == '+' || rest[0] == '-'):
				sign := 1
				if rest[0] == '-' {
					sign = -1
				}
				pos++
				var hh, mm int
				if hh, err = digits(2); err == nil {
					if pos < len(text) && text[pos] == ':' {
						pos++
					}
					if mm, err = digits(2); err == nil {
						zoneOff = sign * (hh*3600 + mm*60)
					}
				}
			default:
				err = fmt.Errorf("text %q: want an offset (Z, +hhmm or +hh:mm) at offset %d", text, pos)
			}
			haveZone = true
		case '%':
			if pos >= len(text) || text[pos] != '%' {
				err = fmt.Errorf("text %q does not match layout %q at offset %d (want %%)", text, layout, pos)
			}
			pos++
		default:
			return Time{}, fmt.Errorf("unknown layout directive %%%c in %q (directives: %%Y %%m %%d %%H %%M %%S %%z %%%%)", d, layout)
		}
		if err != nil {
			return Time{}, err
		}
	}
	if pos != len(text) {
		return Time{}, fmt.Errorf("text %q has trailing characters after layout %q", text, layout)
	}
	for _, need := range []byte{'Y', 'm', 'd'} {
		if _, ok := fields[need]; !ok {
			return Time{}, fmt.Errorf("layout %q has no %%%c: a date field is never defaulted", layout, need)
		}
	}
	if !haveZone {
		return Time{}, fmt.Errorf("layout %q has no %%z: a time without a zone is never implicit", layout)
	}
	if fields['m'] < 1 || fields['m'] > 12 || fields['d'] < 1 || fields['H'] > 23 || fields['M'] > 59 || fields['S'] > 59 {
		return Time{}, fmt.Errorf("time %q has a field out of range", text)
	}
	loc := time.UTC
	if zoneOff != 0 {
		loc = time.FixedZone("", zoneOff)
	}
	t := time.Date(fields['Y'], time.Month(fields['m']), fields['d'], fields['H'], fields['M'], fields['S'], 0, loc)
	if t.Day() != fields['d'] {
		return Time{}, fmt.Errorf("time %q is not a real calendar day", text)
	}
	return NewTime(t)
}
