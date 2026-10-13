package functions_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/functions"
)

// A Duration is a length of time, a Time an instant, a Date a calendar day. Points and
// vectors: instant minus instant is a Duration, instant plus Duration is an instant.
func ExampleParseTime() {
	created, _ := functions.ParseTime("2026-08-01T00:00:00Z")
	now, _ := functions.ParseTime("2026-10-13T09:30:00Z")

	age, _ := functions.EvalBinaryOp(functions.OpSub, now, created) // Time - Time
	fmt.Println(age)

	limit, _ := functions.ParseDuration("30d")
	stale, _ := functions.EvalBinaryOp(functions.OpGt, age, limit)
	fmt.Println(stale)

	deadline, _ := functions.EvalBinaryOp(functions.OpAdd, now, limit) // Time + Duration
	fmt.Println(deadline)

	ratio, _ := functions.EvalBinaryOp(functions.OpDiv, age, limit) // Duration / Duration is exact
	fmt.Println(ratio)

	_, err := functions.EvalBinaryOp(functions.OpAdd, now, created) // Time + Time makes no sense
	fmt.Println(err)
	// Output:
	// 1761h30m
	// true
	// 2026-11-12T09:30:00Z
	// 3523r1440
	// operator '+' is not defined for Time and Time; two instants cannot be added; subtract them to get a Duration
}

// A Date moves by whole days; months clamp to the end of the month; a Date and a Time meet
// only through a named zone, and a day is not always 24 hours long.
func ExampleDate_AddMonths() {
	d, _ := functions.NewDate(2026, 1, 31)
	next, _ := d.AddMonths(1)
	fmt.Println(next) // 31 January + 1 month

	berlin, _ := functions.LoadZone("Europe/Berlin")
	spring, _ := functions.NewDate(2026, 3, 29) // the clocks go forward at 02:00
	start, _ := functions.TimeOfDate(berlin, spring)
	after, _ := spring.AddDays(1)
	end, _ := functions.TimeOfDate(berlin, after)
	length, _ := functions.EvalBinaryOp(functions.OpSub, end, start)
	fmt.Println(length) // the day is 23 hours long

	_, err := functions.EvalBinaryOp(functions.OpAdd, spring, functions.Duration(90*60e9))
	fmt.Println(err != nil) // Date + 90m: a Date moves by whole days only
	// Output:
	// 2026-02-28
	// 23h
	// true
}
