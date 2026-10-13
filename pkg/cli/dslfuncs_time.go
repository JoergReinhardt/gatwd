package cli

import (
	"fmt"
	"math/big"
	"time"

	"github.com/joergreinhardt/gatwd/functions"
)

// registerTimeFuncs adds the function forms of Duration, Time and Date (hardening §5.6),
// the generic helpers of functions/duration.go, time.go, date.go and timefmt.go. All are
// data-last and curried. Text parsers answer with a Maybe (malformed text is Nothing, never
// an error); impossible requests (an unknown zone, 2026-02-30, a layout directive that does
// not exist) are errors. A Date and a Time never convert implicitly: date_of_time and
// time_of_date take the zone as an argument.
//
//	Duration  dur_ns d · dur_ms d · to_seconds d · dur_from_seconds x · dur_parse s (Maybe)
//	          dur_truncate unit d · dur_round unit d       (unit: a positive Duration, `1m`)
//	Time      time_parse s (Maybe) · time_format layout t · time_unix t · time_from_unix n
//	          time_year/month/day/hour/minute/second/nano t · time_weekday t (1 Monday .. 7)
//	          time_in zone t · time_truncate d t · time_round d t
//	Date      date_of y m d · date_parse s (Maybe) · date_format layout d
//	          date_year/month/day d · date_weekday d · date_add_days n d
//	          date_add_months n d · date_add_years n d · date_days_in_month d · is_leap_year y
//	Bridge    date_of_time zone t · time_of_date zone d
func registerTimeFuncs(m map[string]any) {
	asDur := func(v any) (functions.Duration, error) {
		if d, ok := v.(functions.Duration); ok {
			return d, nil
		}
		return 0, fmt.Errorf("want a Duration, got %s", functions.KindName(v))
	}
	asTime := func(v any) (functions.Time, error) {
		if t, ok := v.(functions.Time); ok {
			return t, nil
		}
		return functions.Time{}, fmt.Errorf("want a Time, got %s", functions.KindName(v))
	}
	asDate := func(v any) (functions.Date, error) {
		if d, ok := v.(functions.Date); ok {
			return d, nil
		}
		return functions.Date{}, fmt.Errorf("want a Date, got %s", functions.KindName(v))
	}
	asZone := func(v any) (*time.Location, error) {
		s, err := asString(v)
		if err != nil {
			return nil, fmt.Errorf("a zone is an IANA name String: %w", err)
		}
		return functions.LoadZone(s)
	}
	i64 := func(n int64) any { return functions.CanonInt(big.NewInt(n)) }
	maybe := func(v any, err error) (any, error) {
		if err != nil {
			return functions.Nothing[any](), nil
		}
		return functions.Just[any](v), nil
	}

	// ---- Duration
	defn(m, "dur_ns", "d", func(a []any) (any, error) {
		d, err := asDur(a[0])
		return i64(int64(d)), err
	})
	defn(m, "dur_ms", "d", func(a []any) (any, error) {
		d, err := asDur(a[0])
		return i64(int64(d) / 1e6), err // truncated toward zero
	})
	defn(m, "to_seconds", "d", func(a []any) (any, error) {
		d, err := asDur(a[0])
		return d.Seconds(), err
	})
	defn(m, "dur_from_seconds", "x", func(a []any) (any, error) { return functions.DurationFromSeconds(a[0]) })
	defn(m, "dur_parse", "s", func(a []any) (any, error) {
		s, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		return maybe(functions.ParseDuration(s))
	})
	defn(m, "dur_truncate", "unit d", func(a []any) (any, error) {
		u, err := asDur(a[0])
		if err != nil {
			return nil, err
		}
		d, err := asDur(a[1])
		if err != nil {
			return nil, err
		}
		return functions.DurationTruncate(u, d)
	})
	defn(m, "dur_round", "unit d", func(a []any) (any, error) {
		u, err := asDur(a[0])
		if err != nil {
			return nil, err
		}
		d, err := asDur(a[1])
		if err != nil {
			return nil, err
		}
		return functions.DurationRound(u, d)
	})

	// ---- Time
	defn(m, "time_parse", "s", func(a []any) (any, error) {
		s, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		return maybe(functions.ParseTime(s))
	})
	defn(m, "time_format", "layout t", func(a []any) (any, error) {
		layout, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		t, err := asTime(a[1])
		if err != nil {
			return nil, err
		}
		return functions.FormatTime(layout, t)
	})
	defn(m, "time_unix", "t", func(a []any) (any, error) {
		t, err := asTime(a[0])
		return i64(functions.TimeUnix(t)), err
	})
	defn(m, "time_from_unix", "n", func(a []any) (any, error) {
		n, err := asInt(a[0])
		if err != nil {
			return nil, err
		}
		return functions.TimeFromUnix(int64(n))
	})
	for _, field := range []string{"year", "month", "day", "hour", "minute", "second", "nano", "weekday"} {
		defn(m, "time_"+field, "t", func(a []any) (any, error) {
			t, err := asTime(a[0])
			if err != nil {
				return nil, err
			}
			return i64(functions.TimeField(field, t)), nil
		})
	}
	defn(m, "time_in", "zone t", func(a []any) (any, error) {
		loc, err := asZone(a[0])
		if err != nil {
			return nil, err
		}
		t, err := asTime(a[1])
		if err != nil {
			return nil, err
		}
		return t.In(loc)
	})
	defn(m, "time_truncate", "d t", func(a []any) (any, error) {
		d, err := asDur(a[0])
		if err != nil {
			return nil, err
		}
		t, err := asTime(a[1])
		if err != nil {
			return nil, err
		}
		return functions.TimeTruncate(d, t)
	})
	defn(m, "time_round", "d t", func(a []any) (any, error) {
		d, err := asDur(a[0])
		if err != nil {
			return nil, err
		}
		t, err := asTime(a[1])
		if err != nil {
			return nil, err
		}
		return functions.TimeRound(d, t)
	})

	// ---- Date
	defn(m, "date_of", "y mo d", func(a []any) (any, error) {
		y, err := asInt(a[0])
		if err != nil {
			return nil, err
		}
		mo, err := asInt(a[1])
		if err != nil {
			return nil, err
		}
		d, err := asInt(a[2])
		if err != nil {
			return nil, err
		}
		return functions.NewDate(y, mo, d)
	})
	defn(m, "date_parse", "s", func(a []any) (any, error) {
		s, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		return maybe(functions.ParseDate(s))
	})
	defn(m, "date_format", "layout d", func(a []any) (any, error) {
		layout, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		d, err := asDate(a[1])
		if err != nil {
			return nil, err
		}
		return functions.FormatDate(layout, d)
	})
	defn(m, "date_year", "d", func(a []any) (any, error) {
		d, err := asDate(a[0])
		return i64(int64(d.Year)), err
	})
	defn(m, "date_month", "d", func(a []any) (any, error) {
		d, err := asDate(a[0])
		return i64(int64(d.Month)), err
	})
	defn(m, "date_day", "d", func(a []any) (any, error) {
		d, err := asDate(a[0])
		return i64(int64(d.Day)), err
	})
	defn(m, "date_weekday", "d", func(a []any) (any, error) {
		d, err := asDate(a[0])
		return i64(int64(d.Weekday())), err
	})
	defn(m, "date_days_in_month", "d", func(a []any) (any, error) {
		d, err := asDate(a[0])
		return i64(int64(functions.DaysInMonth(d.Year, d.Month))), err
	})
	addN := func(name string, f func(functions.Date, int64) (functions.Date, error)) {
		defn(m, name, "n d", func(a []any) (any, error) {
			n, err := asInt(a[0])
			if err != nil {
				return nil, err
			}
			d, err := asDate(a[1])
			if err != nil {
				return nil, err
			}
			return f(d, int64(n))
		})
	}
	addN("date_add_days", functions.Date.AddDays)
	addN("date_add_months", functions.Date.AddMonths)
	addN("date_add_years", functions.Date.AddYears)
	defn(m, "is_leap_year", "y", func(a []any) (any, error) {
		y, err := asInt(a[0])
		return functions.IsLeapYear(y), err
	})

	// ---- the Date / Time bridge
	defn(m, "date_of_time", "zone t", func(a []any) (any, error) {
		loc, err := asZone(a[0])
		if err != nil {
			return nil, err
		}
		t, err := asTime(a[1])
		if err != nil {
			return nil, err
		}
		return functions.DateOfTime(loc, t)
	})
	defn(m, "time_of_date", "zone d", func(a []any) (any, error) {
		loc, err := asZone(a[0])
		if err != nil {
			return nil, err
		}
		d, err := asDate(a[1])
		if err != nil {
			return nil, err
		}
		return functions.TimeOfDate(loc, d)
	})
}
