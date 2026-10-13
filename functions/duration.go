package functions

import (
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"
)

// Duration is a signed length of elapsed time: the language's answer to "how long?".
//
// # The concept
//
// A length of time is a number with a unit. Three questions are easy to confuse and the
// language keeps them apart in three types: how long (Duration), which instant (Time) and
// which calendar day (Date). A Duration is the difference of two instants and the step by
// which an instant moves; it is a one-dimensional quantity (a "vector"), where an instant
// is a point. Point minus point is a vector, point plus vector is a point, vector plus
// vector is a vector, and point plus point is meaningless: the operator matrix in
// timeops.go is exactly that affine-space rule.
//
// # Representation
//
// Duration is `type Duration time.Duration` (int64 nanoseconds, about ±292 years). It is a
// named type so that a Go type switch never mistakes it for an Int: `5s + 1` is an error,
// not 5000000001. The length of every unit is fixed: a day (`d`) is exactly 24 hours and a
// week (`w`) exactly 7 days, which is why a Duration is not a calendar quantity; calendar
// months and years are functions on Date (date_add_months).
//
// # Literal and printed form
//
// `5s`, `250ms`, `1h30m`, `2d`, `1w`, `1.5h` are literals (see ParseDuration). A Duration
// prints in the canonical chain `1h30m`, `1s500ms`, `-5s`, `0s`, which lexes back to the
// same value. The largest unit printed is the hour, so 30 days prints as `720h`.
//
// # Example
//
//	d, _ := functions.ParseDuration("1h30m")
//	d.String()               // "1h30m"
//	functions.Duration(5e9)  // 5s
type Duration time.Duration

// Std is the standard library's time.Duration with the same length.
func (d Duration) Std() time.Duration { return time.Duration(d) }

// IsDuration reports whether v is a Duration.
func IsDuration(v any) bool { _, ok := v.(Duration); return ok }

// durationUnit is one suffix of the literal and its length in nanoseconds.
type durationUnit struct {
	name string
	ns   int64
}

// durationUnits lists the units in the order they are printed (largest first). The day and
// the week are accepted in literals only; printing stops at the hour so that a printed
// Duration never suggests a calendar length.
var durationUnits = []durationUnit{
	{"w", 7 * 24 * 3600e9}, {"d", 24 * 3600e9},
	{"h", 3600e9}, {"m", 60e9}, {"s", 1e9}, {"ms", 1e6}, {"us", 1e3}, {"ns", 1},
}

// DurationUnitNames lists the unit suffixes of a Duration literal, for error messages.
const DurationUnitNames = "ns us µs ms s m h d w"

func lookupDurationUnit(word string) (int64, bool) {
	switch word {
	case "µs", "μs": // U+00B5 (micro sign) and U+03BC (Greek mu), as Go's ParseDuration reads both
		return 1e3, true
	}
	for _, u := range durationUnits {
		if u.name == word {
			return u.ns, true
		}
	}
	return 0, false
}

// String prints the canonical chain: hours, minutes, seconds, milliseconds, microseconds,
// nanoseconds, each present only when non-zero (`1h30m`, `1s500ms`, `-5s`, `0s`). It is
// exact, never fractional, and the lexer reads it back to the same Duration.
func (d Duration) String() string {
	if d == 0 {
		return "0s"
	}
	var u uint64
	var sb strings.Builder
	if d < 0 {
		sb.WriteByte('-')
		u = uint64(-(int64(d) + 1)) + 1 // safe for the smallest int64
	} else {
		u = uint64(d)
	}
	for _, unit := range durationUnits[2:] { // from the hour down
		if q := u / uint64(unit.ns); q > 0 {
			fmt.Fprintf(&sb, "%d%s", q, unit.name)
			u -= q * uint64(unit.ns)
		}
	}
	return sb.String()
}

// maxDurationText bounds the text a Duration is parsed from.
const maxDurationText = 128

// ParseDuration reads a Duration from text: an optional sign and one or more segments, each
// a number immediately followed by a unit (`1h30m`, `250ms`, `-5s`, `1.5h`, `2d`, `1w`).
//
// The units are ns us µs ms s m h d w, of fixed length (a `d` is 24 hours). The segments
// may come in any order but a unit may appear only once, only the first segment may carry a
// fractional part, and the total must be a whole number of nanoseconds that fits an int64.
// Whitespace, an empty text, a missing or unknown unit (`5q`, `5`), a repeated unit
// (`1h1h`) and an overflow are errors that say which. The expression lexer uses this for
// its literals and dur_parse uses it with a Maybe around it.
func ParseDuration(text string) (Duration, error) {
	if text == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if len(text) > maxDurationText {
		return 0, fmt.Errorf("duration text is longer than %d characters", maxDurationText)
	}
	s := text
	neg := false
	switch s[0] {
	case '-':
		neg = true
		s = s[1:]
	case '+':
		s = s[1:]
	}
	if s == "" {
		return 0, fmt.Errorf("malformed duration %q: a sign without a number", text)
	}
	total := new(big.Rat)
	seen := map[int64]bool{}
	for first := true; s != ""; first = false {
		// the number: digits with an optional fractional part
		end := 0
		for end < len(s) && s[end] >= '0' && s[end] <= '9' {
			end++
		}
		if end == 0 {
			return 0, fmt.Errorf("malformed duration %q: expected a number before %q", text, s)
		}
		fractional := false
		if end < len(s) && s[end] == '.' {
			fractional = true
			end++
			digits := 0
			for end < len(s) && s[end] >= '0' && s[end] <= '9' {
				end++
				digits++
			}
			if digits == 0 {
				return 0, fmt.Errorf("malformed duration %q: nothing after the decimal point", text)
			}
		}
		if fractional && !first {
			return 0, fmt.Errorf("malformed duration %q: only the first segment may have a fractional part (write 1h30m, not 1h0.5h)", text)
		}
		num, ok := new(big.Rat).SetString(s[:end])
		if !ok {
			return 0, fmt.Errorf("malformed duration %q: bad number %q", text, s[:end])
		}
		s = s[end:]
		// the unit: everything up to the next digit
		end = 0
		for end < len(s) && (s[end] < '0' || s[end] > '9') && s[end] != '.' {
			end++
		}
		word := s[:end]
		s = s[end:]
		if word == "" {
			return 0, fmt.Errorf("malformed duration %q: the number %s has no unit (units: %s)", text, num.RatString(), DurationUnitNames)
		}
		ns, ok := lookupDurationUnit(word)
		if !ok {
			return 0, fmt.Errorf("unknown duration unit %q in %q (units: %s)", word, text, DurationUnitNames)
		}
		if seen[ns] {
			return 0, fmt.Errorf("malformed duration %q: the unit %q is used twice", text, word)
		}
		seen[ns] = true
		total.Add(total, num.Mul(num, new(big.Rat).SetInt64(ns)))
	}
	if !total.IsInt() {
		return 0, fmt.Errorf("duration %q is not a whole number of nanoseconds", text)
	}
	n := total.Num()
	if neg {
		n = new(big.Int).Neg(n)
	}
	if !n.IsInt64() {
		return 0, fmt.Errorf("duration %q is out of range (about ±292 years)", text)
	}
	return Duration(n.Int64()), nil
}

// durationFromRat is the whole-nanosecond Duration nearest zero to r (truncation toward
// zero, as Go's time.Duration arithmetic does); what names the operation for the error
// when the result does not fit an int64 of nanoseconds.
func durationFromRat(r *big.Rat, what string) (Duration, error) {
	q := new(big.Int).Quo(r.Num(), r.Denom()) // Quo truncates toward zero
	if !q.IsInt64() {
		return 0, fmt.Errorf("Duration overflow in %s (the range is about ±292 years)", what)
	}
	return Duration(q.Int64()), nil
}

// scaleFactor reads the number a Duration is multiplied or divided by as an exact
// fraction: an Int, a Ratio, or a Float (taken at its exact binary value; NaN and the
// infinities are errors).
func scaleFactor(v any) (*big.Rat, error) {
	if r, ok := ratOperand(v); ok {
		return r, nil
	}
	if f, ok := floatOperand(v); ok && (math.IsNaN(f) || math.IsInf(f, 0)) {
		return nil, fmt.Errorf("cannot scale a Duration by %v", f)
	}
	return nil, fmt.Errorf("cannot scale a Duration by %s; use a real number", KindName(v))
}
