package cli

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// Duration, Time and Date as members of the collection rows: structural equality inside
// List/Pair/Maybe, Set membership and HAMT keys agree with `==` (a Time is its instant).
func TestTimeTypesInCollections(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct {
		src  string
		want any
	}{
		{`[1s, 2s] == [1000ms, 2000ms]`, true},
		{`[1s, 2s] == [1s, 3s]`, false},
		{`[@2026-10-13T09:30:00Z] == [@2026-10-13T11:30:00+02:00]`, true},
		{`[@2026-10-13] == [@2026-10-13]`, true},
		{`[@2026-10-13] == [@2026-10-14]`, false},
		{`(pair 1s @2026-10-13) == (pair 1000ms @2026-10-13)`, true},
		{`(just 5s) == (just 5000ms)`, true},
		{`(just 5s) == (just 6s)`, false},
		{`(just 5s) == (just 5)`, false},
		{`member @2026-10-13T11:30:00+02:00 (to_set [@2026-10-13T09:30:00Z])`, true},
		{`member 1000ms (to_set [1s, 2s])`, true},
		{`member @2026-10-13 (to_set [@2026-10-13, @2026-10-14])`, true},
		{`member @2026-10-15 (to_set [@2026-10-13, @2026-10-14])`, false},
		{`member 1s (to_set [@2026-10-13])`, false},
		{`member 1 (to_set [1s])`, false},
		{`set_contains @2026-10-13T11:30:00+02:00 (set_add @2026-10-13T09:30:00Z set_empty)`, true},
		{`length (to_set [@2026-10-13T09:30:00Z, @2026-10-13T11:30:00+02:00, @2026-10-13T10:30:00+01:00])`, 1},
		{`length (to_set [1s, 1000ms, 2s])`, 2},
		{`length (to_set [@2026-10-13, @2026-10-13, @2026-10-14])`, 2},
		{`length (set_remove @2026-10-13T11:30:00+02:00 (to_set [@2026-10-13T09:30:00Z]))`, 0},
		{`hamt_get @2026-10-13T11:30:00+02:00 (hamt_from [[@2026-10-13T09:30:00Z, "a"]])`, "a"},
		{`hamt_get 1000ms (hamt_from [[1s, "a"]])`, "a"},
		{`hamt_get @2026-10-14 (hamt_from [[@2026-10-13, "a"], [@2026-10-14, "b"]])`, "b"},
		// a Date key is a Date, not the text it prints as (String is not the printed form of a value)
		{`hamt_has_path "2026-10-13" (hamt_from [[@2026-10-13, "a"]])`, false},
		{`hamt_has_path "2026-10-13" (hamt_from [["2026-10-13", "a"]])`, true},
	} {
		got := evalOne(t, in, tc.src)
		if !functions.StructuralEqual(got, tc.want) {
			t.Errorf("%s = %#v, want %#v", tc.src, got, tc.want)
		}
	}
	// the raw memory type: a Set of canonical UTC Times
	s, ok := evalOne(t, in, `to_set [@2026-10-13T11:30:00+02:00]`).(*collections.Setoid[any])
	if !ok {
		t.Fatal("to_set must yield a *Setoid[any]")
	}
	if el := s.ElementsAny(); len(el) != 1 || el[0].(functions.Time).String() != "2026-10-13T09:30:00Z" {
		t.Errorf("a Set member is the canonical UTC Time, got %v", el)
	}
}

// JSON and YAML encode the three types as Strings (RFC 3339, ISO 8601, the Duration
// literal), never silently as numbers; decoding leaves text as text.
func TestTimeTypesEncode(t *testing.T) {
	in := NewInterpreter()
	for src, want := range map[string]string{
		`json_encode @2026-10-13T09:30:00Z`:      `"2026-10-13T09:30:00Z"`,
		`json_encode @2026-10-13T11:30:00+02:00`: `"2026-10-13T11:30:00+02:00"`,
		`json_encode @2026-10-13T09:30:00.25Z`:   `"2026-10-13T09:30:00.25Z"`,
		`json_encode @2026-10-13`:                `"2026-10-13"`,
		`json_encode 1h30m`:                      `"1h30m"`,
		`json_encode -5s`:                        `"-5s"`,
		`json_encode [@2026-10-13, 90m]`:         `["2026-10-13","1h30m"]`,
		`json_encode (hamt_from [["at", @2026-10-13T09:30:00Z], ["for", 90m]])`: `{"at":"2026-10-13T09:30:00Z","for":"1h30m"}`,
	} {
		if got := evalOne(t, in, src); got != want {
			t.Errorf("%s = %#v, want %s", src, got, want)
		}
	}
	// YAML quotes them, so reading the document back gives text, not a YAML timestamp
	y := evalOne(t, in, `yaml_encode [@2026-10-13, @2026-10-13T09:30:00Z]`).(string)
	if !strings.Contains(y, `"2026-10-13"`) || !strings.Contains(y, `"2026-10-13T09:30:00Z"`) {
		t.Errorf("yaml_encode must quote date-like strings: %q", y)
	}
	// decode: text stays text; the parsers revive it explicitly
	if got := evalOne(t, in, `maybe_default 0 (map (\d -> date_parse (hamt_path "day" d)) (json_parse (json_encode (hamt_from [["day", @2026-10-13]]))))`); got == nil {
		t.Error("nil")
	} else if m, ok := got.(functions.Maybe[any]); !ok || !m.IsJust() {
		t.Errorf("round trip through JSON and date_parse = %#v", got)
	}
	if got := evalOne(t, in, `maybe_default "?" (map (\d -> hamt_path "day" d) (json_parse (json_encode (hamt_from [["day", @2026-10-13]]))))`); got != "2026-10-13" {
		t.Errorf("decoded text = %#v (a String, not a Date)", got)
	}
}
