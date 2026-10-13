package functions

import (
	"testing"
)

// A literal token is typed only when it is written canonically: the same discipline as
// ParseCanonicalNumber, so a typed token never rewrites what the user wrote.
func TestLiteralTokenTemporal(t *testing.T) {
	d := func(s string) Duration {
		v, err := ParseDuration(s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	date, _ := NewDate(2026, 10, 13)
	tm, _ := ParseTime("2026-10-13T10:00:00Z")
	tz, _ := ParseTime("2026-10-13T11:30:00+02:00")
	for text, want := range map[string]any{
		"5s": d("5s"), "1h30m": d("1h30m"), "30d": d("720h"), "90m": d("1h30m"), "-5s": d("-5s"),
		"1.5h": d("90m"), "0s": Duration(0), "250ms": d("250ms"), "2w": d("336h"),
		"@2026-10-13": date, "@2026-10-13T10:00:00Z": tm, "@2026-10-13T11:30:00+02:00": tz,
		// numbers are unchanged
		"10": 10, "1.5": 1.5,
	} {
		got := LiteralToken(text)
		if KindName(got) != KindName(want) || !StructuralEqual(got, want) {
			t.Errorf("LiteralToken(%q) = %#v (%s), want %#v (%s)", text, got, KindName(got), want, KindName(want))
		}
	}
	// the zone of a Time token survives
	if got := LiteralToken("@2026-10-13T11:30:00+02:00").(Time).String(); got != "2026-10-13T11:30:00+02:00" {
		t.Errorf("zone lost: %s", got)
	}
	// not canonical, not a literal, or not valid: stays the text
	for _, text := range []string{
		"007s", "1.50h", "+5s", "-0s", "0.0s", "00s", "1h1h", "5q", "s", "5 s", "", "-", "1h 30m", ".5s", "5S",
		"@2026-10-13T10:00:00+00:00", "@2026-10-13T10:00:00.250Z", "@2026-10-13T10:00:00", "@2026-02-30", "@2026-1-3",
		"@", "@x", "2026-10-13", "2026-10-13T10:00:00Z", "@2026-10-13T10:00:00z", "@2026-10-13 ",
		"007", "1.50", "alice",
	} {
		if got, ok := LiteralToken(text).(string); !ok || got != text {
			t.Errorf("LiteralToken(%q) = %#v, want the text", text, LiteralToken(text))
		}
	}
}

// TemporalLiteral prints a text that the lexer-facing parsers read back to the same value.
func TestTemporalLiteralRoundTrip(t *testing.T) {
	date, _ := NewDate(2026, 10, 13)
	tz, _ := ParseTime("2026-10-13T11:30:00.25+02:00")
	for _, v := range []any{Duration(5e9), Duration(-90 * 60e9), Duration(0), date, tz} {
		text, ok := TemporalLiteral(v)
		if !ok {
			t.Fatalf("TemporalLiteral(%#v) not ok", v)
		}
		if back := LiteralToken(text); !StructuralEqual(back, v) || KindName(back) != KindName(v) {
			t.Errorf("TemporalLiteral(%#v) = %q reads back as %#v", v, text, back)
		}
	}
	if _, ok := TemporalLiteral(5); ok {
		t.Error("an Int is not temporal")
	}
}
