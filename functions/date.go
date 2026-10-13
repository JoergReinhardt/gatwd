package functions

import (
	"fmt"
	"time"
)

// Date is a calendar day: the language's answer to "which day?".
//
// # The concept
//
// A calendar day has no time of day and no zone: it is "the 13th of October 2026" wherever
// you stand. That is not a point on the timeline (the same day starts at different instants
// in Berlin and in Tokyo), which is why a Date is its own type and never converts to a Time
// implicitly: date_of_time and time_of_date take the zone as an argument and the language
// has no operator that mixes the two. A due date, a birthday or a release day is a Date; the
// moment a certificate expires is a Time; how long until it is a Duration.
//
// # Representation
//
// Date is a plain comparable struct of year, month and day in the proleptic Gregorian
// calendar, valid for the years 1 to 9999 (the years an ISO 8601 date writes with four
// digits). Because it is comparable, a Date is a valid key of a HAMT and a member of a Set,
// and `==` is plain structural equality. It prints as ISO 8601, `2026-10-13`, which the
// lexer reads back as the literal `@2026-10-13`.
//
// # Arithmetic with a Duration (D17)
//
// A day is a whole, zone-free step, so `date + 7d` and `deadline - today` are exact and
// readable: Date ± Duration is allowed when the Duration is a whole number of days (`2d`,
// `-1w`, `0s`), and `Date - Date` is a Duration that is a multiple of `1d`. Anything that is
// not whole days (`date + 90m`) is an error. Calendar months and years are not fixed
// lengths and always go through functions (date_add_months, date_add_years).
//
// # Example
//
//	d, _ := functions.NewDate(2026, 10, 13)
//	d.String() // "2026-10-13"
type Date struct {
	Year  int
	Month int
	Day   int
}

// NewDate returns the Date for year, month (1 to 12) and day, or an error naming the field
// that is out of range (an impossible day such as 2026-02-30 included).
func NewDate(year, month, day int) (Date, error) {
	if year < minYear || year > maxYear {
		return Date{}, fmt.Errorf("date out of range: year %d is outside %d to %d", year, minYear, maxYear)
	}
	if month < 1 || month > 12 {
		return Date{}, fmt.Errorf("invalid date: month %d is outside 1 to 12", month)
	}
	if dim := DaysInMonth(year, month); day < 1 || day > dim {
		return Date{}, fmt.Errorf("invalid date: day %d is outside 1 to %d for %04d-%02d", day, dim, year, month)
	}
	return Date{Year: year, Month: month, Day: day}, nil
}

// IsDate reports whether v is a Date.
func IsDate(v any) bool { _, ok := v.(Date); return ok }

// String renders ISO 8601, `2026-10-13`.
func (d Date) String() string { return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day) }

// IsLeapYear reports whether year is a leap year of the Gregorian calendar.
func IsLeapYear(year int) bool { return year%4 == 0 && (year%100 != 0 || year%400 == 0) }

// DaysInMonth is the number of days of month (1 to 12) in year; 0 for an invalid month.
func DaysInMonth(year, month int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	case 2:
		if IsLeapYear(year) {
			return 29
		}
		return 28
	}
	return 0
}

// std is the midnight UTC of the day, for calendar arithmetic only.
func (d Date) std() time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
}

// dateOfStd reads back the calendar day of an instant, in the zone it carries.
func dateOfStd(t time.Time) (Date, error) {
	return NewDate(t.Year(), int(t.Month()), t.Day())
}

// Compare orders two dates chronologically (-1, 0 or 1).
func (d Date) Compare(e Date) int {
	switch {
	case d.Year != e.Year:
		return cmpInt(d.Year, e.Year)
	case d.Month != e.Month:
		return cmpInt(d.Month, e.Month)
	}
	return cmpInt(d.Day, e.Day)
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// ParseDate reads ISO 8601 calendar text `2026-10-13` (four-digit year, two-digit month and
// day). Anything else, including a time of day, and an impossible day are errors.
func ParseDate(text string) (Date, error) {
	t, err := time.Parse(time.DateOnly, text)
	if err != nil {
		return Date{}, fmt.Errorf("malformed date %q (want YYYY-MM-DD, such as 2026-10-13): %s", text, trimParseError(err))
	}
	return dateOfStd(t)
}

// AddDays moves the date by n days (negative goes back); the result must stay within the
// years 1 to 9999.
func (d Date) AddDays(n int64) (Date, error) {
	if n > 4_000_000 || n < -4_000_000 {
		return Date{}, fmt.Errorf("date out of range: %d days is beyond the years %d to %d", n, minYear, maxYear)
	}
	return dateOfStd(d.std().AddDate(0, 0, int(n)))
}

// AddMonths moves the date by n calendar months. The day clamps to the end of the target
// month (31 January plus one month is 28 or 29 February, as in Java's plusMonths), so the
// result is always a valid date and a month is never skipped.
func (d Date) AddMonths(n int64) (Date, error) {
	if n > 120_000 || n < -120_000 {
		return Date{}, fmt.Errorf("date out of range: %d months is beyond the years %d to %d", n, minYear, maxYear)
	}
	total := int64(d.Year)*12 + int64(d.Month-1) + n
	year, month := int(total/12), int(total%12)
	if total < 0 {
		return Date{}, fmt.Errorf("date out of range: before year %d", minYear)
	}
	month++
	day := d.Day
	if dim := DaysInMonth(year, month); year >= minYear && year <= maxYear && day > dim {
		day = dim
	}
	return NewDate(year, month, day)
}

// AddYears moves the date by n calendar years, clamping 29 February to 28 February in a
// year that is not a leap year.
func (d Date) AddYears(n int64) (Date, error) {
	if n > 20_000 || n < -20_000 {
		return Date{}, fmt.Errorf("date out of range: %d years is beyond the years %d to %d", n, minYear, maxYear)
	}
	return d.AddMonths(n * 12)
}

// Weekday is the ISO 8601 weekday number: 1 for Monday up to 7 for Sunday.
func (d Date) Weekday() int {
	w := int(d.std().Weekday())
	if w == 0 {
		return 7
	}
	return w
}

// daysBetween is the number of days from e to d (negative when d is earlier).
func (d Date) daysBetween(e Date) int64 {
	return (d.std().Unix() - e.std().Unix()) / 86400
}

// dateSub is a - b as a Duration of whole days, or an error beyond about ±292 years.
func dateSub(a, b Date) (Duration, error) {
	days := a.daysBetween(b)
	const maxDays = 106751 // the whole days within ±(2^63 - 1) nanoseconds
	if days > maxDays || days < -maxDays {
		return 0, fmt.Errorf("Duration overflow: the difference between %s and %s is beyond about ±292 years", a, b)
	}
	return Duration(days * 24 * int64(time.Hour)), nil
}

// wholeDays reports the number of days a Duration spans when it is a whole number of days.
func wholeDays(d Duration) (int64, bool) {
	const day = 24 * int64(time.Hour)
	return int64(d) / day, int64(d)%day == 0
}
