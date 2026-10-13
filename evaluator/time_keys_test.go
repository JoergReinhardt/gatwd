package evaluator

import (
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

func timeAt(t *testing.T, text string) functions.Time {
	t.Helper()
	v, err := functions.ParseTime(text)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// The deduplicating steps bucket a Time by its instant: the same instant shown in two
// zones is one element, in step form exactly as in the function forms. Two different
// instants stay two. (Raw memory types, not formatted text.)
func TestSetLikeStepsKeyTimeByInstant(t *testing.T) {
	utc := timeAt(t, "2026-10-13T09:30:00Z")
	berlin := timeAt(t, "2026-10-13T11:30:00+02:00")
	other := timeAt(t, "2026-10-13T09:30:01Z")
	sec, thousand := functions.Duration(1e9), functions.Duration(1e9)

	emit := StepEmitValues(utc, berlin, other, sec, thousand)
	count := func(steps ...PipelineStep[functions.Either[error, any]]) []any {
		return runPipelineToValues(append([]PipelineStep[functions.Either[error, any]]{emit}, steps...))
	}
	if got := count(StepDistinct()); len(got) != 3 { // utc==berlin, other, 1s==1s
		t.Errorf("distinct kept %v, want 3 elements", got)
	}
	set := count(StepToSet())
	if len(set) != 1 || set[0].(*collections.Setoid[any]).Len() != 3 {
		t.Errorf("to_set = %#v, want one Set of 3", set)
	}
	freq := count(StepFrequencies())
	h, ok := freq[0].(*collections.HAMT[string, any])
	if !ok || h.Len() != 3 {
		t.Fatalf("frequencies = %#v", freq)
	}
	utcKey, err := EncodeKey(utc)
	if err != nil {
		t.Fatal(err)
	}
	if berlinKey, _ := EncodeKey(berlin); berlinKey != utcKey {
		t.Errorf("one instant, one key: %q vs %q", utcKey, berlinKey)
	}
	if n, _ := h.Get(utcKey); n != 2 {
		t.Errorf("the instant seen in two zones was counted %v times, want 2", n)
	}
	hamt := count(StepToHAMT())
	if len(hamt) != 1 {
		t.Fatalf("to_hamt = %#v", hamt)
	}
	if v, ok := hamt[0].(*collections.HAMT[string, any]).Get(utcKey); !ok {
		t.Errorf("the HAMT is keyed by the encoded UTC instant, got %v", v)
	}
}

// DecodeKey inverts EncodeKey for the three time types, and a Time decodes to its UTC instant.
func TestTimeKeysRoundTrip(t *testing.T) {
	berlin := timeAt(t, "2026-10-13T11:30:00.250+02:00")
	date, err := functions.ParseDate("2026-10-13")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []any{berlin, date, functions.Duration(-90e9)} {
		key, err := EncodeKey(v)
		if err != nil {
			t.Fatalf("%v: %v", v, err)
		}
		got := DecodeKey(key)
		if !functions.StructuralEqual(got, v) {
			t.Errorf("DecodeKey(EncodeKey(%v)) = %#v", v, got)
		}
	}
	dk, _ := EncodeKey(date)
	sk, _ := EncodeKey("2026-10-13")
	if dk == sk {
		t.Error("a Date key must differ from the String it prints as")
	}
}
