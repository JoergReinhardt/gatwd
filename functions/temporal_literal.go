package functions

import "strings"

// ParseCanonicalDuration reads a Duration written as a literal token (`5s`, `1h30m`, `30d`,
// `-2w`, `1.5h`): the text ParseDuration accepts, written canonically.
//
// The discipline is the one ParseCanonicalNumber applies to numbers: a token is typed only
// when reading it loses nothing about how it was written. For a number that means "prints
// back unchanged". A Duration cannot meet that test as stated, because a Duration is a
// quantity with many spellings (`90m`, `1h30m` and `5400s` are one value, and `30d` prints
// as `720h`), and the whole point of typing `const MAX_AGE = 30d` is that people write
// durations in the unit they think in. So the test is the one that is about *spelling*, not
// about unit choice: every numeric segment must itself be canonical, with no `+` sign, no
// leading zero and no trailing zero in a fraction. Text such as `007s`, `1.50h`, `+5s`,
// `-0s` and `0.0s` is therefore not a typed token and stays a String, exactly as `007` and
// `1.50` do. The expression lexer reads every one of these as a Duration, so a typed token
// never contradicts the lexer; it only declines to guess for the spellings it would
// normalise.
//
// ok is false when text is not a Duration literal at all (`5`, `alice`, `5q`) or is written
// non-canonically.
func ParseCanonicalDuration(text string) (Duration, bool) {
	if text == "" || len(text) > maxDurationText {
		return 0, false
	}
	body := strings.TrimPrefix(text, "-")
	if body == "" || body[0] < '0' || body[0] > '9' {
		return 0, false
	}
	d, err := ParseDuration(text)
	if err != nil {
		return 0, false
	}
	if d == 0 && body != text {
		return 0, false // -0s
	}
	// Check each numeric segment: the digits (with an optional fraction) before each unit.
	for s := body; s != ""; {
		end := 0
		for end < len(s) && s[end] >= '0' && s[end] <= '9' {
			end++
		}
		whole := s[:end]
		if len(whole) > 1 && whole[0] == '0' {
			return 0, false // leading zero
		}
		if end < len(s) && s[end] == '.' {
			end++
			from := end
			for end < len(s) && s[end] >= '0' && s[end] <= '9' {
				end++
			}
			if s[end-1] == '0' && end > from {
				return 0, false // trailing zero of the fraction
			}
		}
		for end < len(s) && (s[end] < '0' || s[end] > '9') && s[end] != '.' {
			end++ // the unit
		}
		s = s[end:]
	}
	return d, true
}

// ParseCanonicalTemporal reads a Time or Date written as a literal token, with its leading
// `@`: `@2026-10-13` (a Date), `@2026-10-13T09:30:00Z` or `@2026-10-13T11:30:00+02:00` (a
// Time). It is ParseCanonicalNumber's discipline for the calendar types: the token is typed
// only when it prints back unchanged, so `@2026-10-13T10:00:00+00:00` (which prints with a
// `Z`) and `@2026-10-13T10:00:00.250Z` (which prints `.25Z`) stay Strings. ok is false for
// anything else, including a Time without a zone and an impossible calendar day.
func ParseCanonicalTemporal(text string) (any, bool) {
	if len(text) < 2 || text[0] != '@' {
		return nil, false
	}
	body := text[1:]
	if strings.IndexByte(body, 'T') >= 0 {
		t, err := ParseTime(body)
		if err != nil || t.String() != body {
			return nil, false
		}
		return t, true
	}
	d, err := ParseDate(body)
	if err != nil || d.String() != body {
		return nil, false
	}
	return d, true
}

// TemporalLiteral is the source text that lexes back to the Duration, Time or Date v:
// `1h30m`, `@2026-10-13T09:30:00Z`, `@2026-10-13`. A Time or Date prints without its `@`,
// so splicing that printed form into script text would read as a subtraction; this is the
// form to splice. ok is false when v is not one of the three types.
func TemporalLiteral(v any) (string, bool) {
	switch x := v.(type) {
	case Duration:
		return x.String(), true
	case Time:
		return "@" + x.String(), true
	case Date:
		return "@" + x.String(), true
	}
	return "", false
}

// parseTemporalToken reads a Duration, Time or Date literal as the expression lexer does,
// without insisting on the canonical spelling (`007s` is 7 seconds there, so it is here).
// It is for operands that are code, such as the `5s` of `|> + 5s`; LiteralToken, for text
// that is data, uses the canonical readers instead.
func parseTemporalToken(text string) (any, bool) {
	if text == "" {
		return nil, false
	}
	if text[0] == '@' {
		body := text[1:]
		if strings.IndexByte(body, 'T') >= 0 {
			if t, err := ParseTime(body); err == nil {
				return t, true
			}
			return nil, false
		}
		if d, err := ParseDate(body); err == nil {
			return d, true
		}
		return nil, false
	}
	first := strings.TrimLeft(text, "+-")
	if len(text)-len(first) > 1 || first == "" || first[0] < '0' || first[0] > '9' {
		return nil, false
	}
	if d, err := ParseDuration(strings.TrimPrefix(text, "+")); err == nil {
		return d, true
	}
	return nil, false
}
