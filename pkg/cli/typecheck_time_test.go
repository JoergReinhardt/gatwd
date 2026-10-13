package cli

import "testing"

// A valid program over Duration, Time and Date stays silent (zero false positives), and the
// result types come from the runtime operator matrix.
func TestTimeTypesAreCheckedSilentOnValid(t *testing.T) {
	valid := `
def age : Time -> Time -> Duration = \now created -> now - created
def stale : Time -> Time -> Bool = \now created -> (now - created) > 30d
def soon : Time -> Time -> Bool = \now exp -> exp - now < 14d
def later : Time -> Duration -> Time = \t d -> t + d
def earlier : Time -> Duration -> Time = \t d -> t - d
def twice : Duration -> Duration = \d -> d * 2
def half : Duration -> Duration = \d -> d / 2
def share : Duration -> Duration -> Ratio = \a b -> a / b
def total : Duration -> Duration -> Duration = \a b -> a + b
def week_on : Date -> Date = \d -> d + 7d
def until : Date -> Date -> Duration = \a b -> a - b
def tomorrow : Date -> Date = \d -> date_add_days 1 d
def lower : Comparable -> Comparable -> Bool = \a b -> a < b
def sooner : Time -> Time -> Bool = \a b -> a < b
lower 1s 2s
lower @2026-10-13 @2026-10-14
lower @2026-10-13T09:30:00Z @2026-10-13T10:30:00Z
stale @2026-10-13T09:30:00Z @2026-01-01T00:00:00Z
twice 5s
share 90m 1h
week_on @2026-10-13
until @2026-10-14 @2026-10-13
now |> map (\t -> soon t @2026-10-20T00:00:00Z)
now |> map (\t -> t - @2026-10-13T00:00:00Z)
[1s, 2s] |> map (\d -> d * 3)
[@2026-10-13] |> map (\d -> d - 1w)
`
	if ws := warnings(valid); len(ws) != 0 {
		t.Fatalf("a valid time program must be silent, got %v", ws)
	}
}

// Each definite contradiction among the time types is reported, naming both types.
func TestTimeTypesAreCheckedReportsContradictions(t *testing.T) {
	for src, want := range map[string]string{
		"5s + 1":  "operator '+' is not defined for Duration and Int",
		"1 + 5s":  "operator '+' is not defined for Int and Duration",
		"5s * 5s": "operator '*' is not defined for Duration and Duration",
		"@2026-10-13T09:30:00Z + @2026-10-13T09:30:00Z": "operator '+' is not defined for Time and Time",
		"@2026-10-13 + @2026-10-13":                     "operator '+' is not defined for Date and Date",
		"@2026-10-13 < @2026-10-13T00:00:00Z":           "operator '<' is not defined for Date and Time",
		"@2026-10-13T09:30:00Z * 2":                     "operator '*' is not defined for Time and Int",
		"5s < 5":                                        "operator '<' is not defined for Duration and Int",
		"5 / 5s":                                        "operator '/' is not defined for Int and Duration",
		"def f : Duration -> Int = \\d -> d\nf 5s":      "annotation says Int",
		"def f : Time -> Duration = \\t -> t\nf @2026-10-13T09:30:00Z":               "annotation says Duration",
		"def f : Time -> Time = \\t -> t\nf 5s":                                      "expected Time",
		"def f : Date -> Date = \\d -> d\nf @2026-10-13T09:30:00Z":                   "expected Date",
		"def f : Num -> Num = \\x -> x\nf 5s":                                        "expected Num",
		"def f : Duration -> Duration = \\d -> d\nf 5":                               "expected Duration",
		"def f : Time -> Duration = \\t -> t - @2026-10-13\nf @2026-10-13T09:30:00Z": "operator '-' is not defined for Time and Date",
	} {
		if ws := warnings(src); !hasWarning(ws, want) {
			t.Errorf("%q: want a warning containing %q, got %v", src, want, ws)
		}
	}
	if ty := tyOfValue(mustEvalOne(t, "5s")); ty.Name != "Duration" || ty.isNumeric() {
		t.Errorf("a Duration is typed Duration and is not numeric, got %v", ty)
	}
	if ty := tyOfValue(mustEvalOne(t, "@2026-10-13T09:30:00Z")); ty.Name != "Time" {
		t.Errorf("got %v", ty)
	}
	if ty := tyOfValue(mustEvalOne(t, "@2026-10-13")); ty.Name != "Date" {
		t.Errorf("got %v", ty)
	}
	if t2, _ := binaryType("/", Ty{Name: "Duration"}, Ty{Name: "Duration"}); t2.Name != "Ratio" {
		t.Errorf("Duration / Duration is typed %v, want Ratio", t2)
	}
	if t2, msg := binaryType("-", Ty{Name: "Time"}, Ty{Name: "Time"}); t2.Name != "Duration" || msg != "" {
		t.Errorf("Time - Time is typed %v (%s)", t2, msg)
	}
	if t2, _ := binaryType("+", Ty{Name: "Date"}, Ty{Name: "Duration"}); t2.Name != "Date" {
		t.Errorf("Date + Duration is typed %v", t2)
	}
}

func mustEvalOne(t *testing.T, src string) any {
	t.Helper()
	got := NewInterpreter().EvaluateToValues(src)
	if len(got) != 1 {
		t.Fatalf("%s = %v", src, got)
	}
	return got[0]
}

// Comparable holds for all three, Num for none, Appendable for none.
func TestTimeTypesAndConstraints(t *testing.T) {
	for _, name := range []string{"Duration", "Time", "Date"} {
		ty := Ty{Name: name}
		if !satisfies(Ty{Name: "Comparable"}, ty) {
			t.Errorf("%s must be Comparable", name)
		}
		if satisfies(Ty{Name: "Num"}, ty) || satisfies(Ty{Name: "Appendable"}, ty) || ty.isNumeric() {
			t.Errorf("%s must be neither Num nor Appendable", name)
		}
	}
}
