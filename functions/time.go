package functions

import (
	"fmt"
	"math/big"
	"time"
	_ "time/tzdata" // the IANA zone database travels with the binary: zone names never depend on the host
)

// Time is a point on the timeline: the language's answer to "which instant?".
//
// # The concept
//
// An instant is a position, not a length: the moment the build finished, the second a
// certificate expires. Two instants can be ordered and subtracted (the difference is a
// Duration), and an instant can be moved by a Duration, but two instants cannot be added:
// "Tuesday plus Wednesday" has no meaning. That is the points-and-vectors rule of an
// affine space, and timeops.go implements exactly it.
//
// # Representation: an instant that remembers a display zone
//
// A Time wraps Go's time.Time (the standard library's instant). It is an instant plus the
// zone it is DISPLAYED in, nothing more: `time_in "Europe/Berlin" t` is the same instant
// rendered in another zone, `==` and the orderings compare instants and ignore the zone,
// and the calendar fields (time_hour, date_of_time) read the zone. There is deliberately
// no zone-less "wall clock" type (no DateTime): a date and time of day with no zone is not
// a point on the timeline, and mixing it with one is a classic source of bugs. Local time
// is never implicit: a parsed time with no zone is an error, and the default zone is UTC.
//
// Time is a struct around time.Time (not time.Time itself) for two reasons: fmt prints a
// bare time.Time as a Go debugging string (`2026-10-13 09:30:00 +0000 UTC`), where a Time
// prints as the RFC 3339 literal the lexer reads back; and the wrapper is the one place
// that enforces the range (years 1 to 9999) and strips the monotonic clock reading, so a
// Time never carries state that would make two equal instants compare unequal.
//
// # Literal and printed form
//
// `@2026-10-13T09:30:00Z` and `@2026-10-13T11:30:00+02:00` are literals (RFC 3339); a Time
// prints without the `@` (`2026-10-13T09:30:00Z`), which time_parse reads back.
//
// # Example
//
//	t, _ := functions.ParseTime("2026-10-13T11:30:00+02:00")
//	t.String() // "2026-10-13T11:30:00+02:00"
//	t.UTC()    // 2026-10-13T09:30:00Z, the same instant
type Time struct{ t time.Time }

// Years outside this range are an error for a Time and a Date: RFC 3339 writes a four-digit
// year, so these are exactly the instants whose printed form reads back.
const (
	minYear = 1
	maxYear = 9999
)

// NewTime wraps a standard library instant as a Time. It strips the monotonic clock
// reading and reports an error for an instant whose year (in UTC or in its own zone) is
// outside 1 to 9999.
func NewTime(t time.Time) (Time, error) {
	t = t.Round(0) // drop the monotonic reading: equal instants must stay equal
	if y := t.UTC().Year(); y < minYear || y > maxYear {
		return Time{}, fmt.Errorf("time out of range: year %d is outside %d to %d", y, minYear, maxYear)
	}
	if y := t.Year(); y < minYear || y > maxYear {
		return Time{}, fmt.Errorf("time out of range: year %d in zone %s is outside %d to %d", y, t.Location(), minYear, maxYear)
	}
	return Time{t: t}, nil
}

// Std is the standard library's time.Time for the same instant and zone.
func (t Time) Std() time.Time { return t.t }

// IsTime reports whether v is a Time.
func IsTime(v any) bool { _, ok := v.(Time); return ok }

// String renders RFC 3339 with the fractional seconds that are present: `2026-10-13T09:30:00Z`,
// `2026-10-13T11:30:00.25+02:00`. The lexer and time_parse read it back to the same instant.
func (t Time) String() string { return t.t.Format(time.RFC3339Nano) }

// UTC is the same instant displayed in UTC.
func (t Time) UTC() Time { return Time{t: t.t.UTC()} }

// In is the same instant displayed in the zone loc. The result may be out of the range of
// years for an instant close to the edge of it; that is an error.
func (t Time) In(loc *time.Location) (Time, error) { return NewTime(t.t.In(loc)) }

// Compare orders two instants (-1, 0 or 1); the display zone plays no part.
func (t Time) Compare(u Time) int { return t.t.Compare(u.t) }

// ParseTime reads RFC 3339 text: `2026-10-13T09:30:00Z`, `2026-10-13T11:30:00+02:00`,
// `2026-10-13T09:30:00.25Z`. The zone (`Z` or an offset) is required: a time without one is
// an error, because local time is never implicit. An invalid calendar value (month 13,
// February 30), a missing part and trailing text are errors that say so. The zone of the
// result is UTC for `Z` and for a zero offset, and a fixed zone of the given offset
// otherwise (never the host's local zone, so the result does not depend on the machine).
func ParseTime(text string) (Time, error) {
	if len(text) > 64 {
		return Time{}, fmt.Errorf("time text is longer than 64 characters")
	}
	t, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		if _, e2 := time.Parse("2006-01-02T15:04:05.999999999", text); e2 == nil {
			return Time{}, fmt.Errorf("time %q has no zone (end it with Z or an offset such as +02:00; local time is never implicit)", text)
		}
		return Time{}, fmt.Errorf("malformed time %q (want RFC 3339, such as 2026-10-13T09:30:00Z): %s", text, trimParseError(err))
	}
	_, off := t.Zone()
	if off == 0 {
		t = t.UTC()
	} else {
		t = t.In(time.FixedZone("", off))
	}
	return NewTime(t)
}

// trimParseError drops the verbose layout echo from a time.Parse error and keeps the cause.
func trimParseError(err error) string {
	if pe, ok := err.(*time.ParseError); ok && len(pe.Message) > 2 {
		return pe.Message[len(": "):]
	}
	return "does not parse"
}

// LoadZone returns the IANA zone called name (`UTC`, `Europe/Berlin`). The empty name and
// `Local` are errors: the host's zone is never implicit. An unknown name is an error.
func LoadZone(name string) (*time.Location, error) {
	if name == "" || name == "Local" {
		return nil, fmt.Errorf("zone %q: local time is never implicit; name an IANA zone such as UTC or Europe/Berlin", name)
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("unknown time zone %q (use an IANA name such as Europe/Berlin)", name)
	}
	return loc, nil
}

// timeSub is a - b as a Duration, exact, and an error when the difference does not fit
// the ±292 years of a Duration (Go's own Sub saturates silently).
func timeSub(a, b Time) (Duration, error) {
	ns := new(big.Int).Sub(big.NewInt(a.t.Unix()), big.NewInt(b.t.Unix()))
	ns.Mul(ns, big.NewInt(1e9))
	ns.Add(ns, big.NewInt(int64(a.t.Nanosecond()-b.t.Nanosecond())))
	if !ns.IsInt64() {
		return 0, fmt.Errorf("Duration overflow: the difference between %s and %s is beyond about ±292 years", a, b)
	}
	return Duration(ns.Int64()), nil
}

// timeAdd is t + d in the same zone, an error when the result leaves the supported years.
func timeAdd(t Time, d Duration) (Time, error) {
	res, err := NewTime(t.t.Add(time.Duration(d)))
	if err != nil {
		return Time{}, fmt.Errorf("%s + %s: %w", t, d, err)
	}
	return res, nil
}

// CanonicalKey is the form of v under which equal values are the same Set member or HAMT
// key. A Time is equal to another when it is the same instant, whatever zone shows it, so a
// Time becomes its UTC form (the display zone is dropped); every other value is itself. The
// collections' key sites (a Set element, a HAMT key, the distinct/frequencies/to_set steps)
// go through it so that `==` and membership agree.
func CanonicalKey(v any) any {
	if t, ok := v.(Time); ok {
		return t.UTC()
	}
	return v
}

// KeyText is the printed form of CanonicalKey(v): the text a HAMT key or a distinct/frequency
// bucket is made of.
func KeyText(v any) string { return fmt.Sprintf("%v", CanonicalKey(v)) }
