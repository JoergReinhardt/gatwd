package evaluator

import (
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
)

// TestIsSubsetValue covers the generic subset predicate structurally (STYLEGUIDE §4):
// Record/HAMT (key,value) containment and Set membership, the empty-subset base case,
// and the type-mismatch error — asserting against raw collection values, not strings.
func TestIsSubsetValue(t *testing.T) {
	// HAMT ⊆ HAMT by (key, value) — the selector ⊆ labels shape.
	labels := collections.NewHAMT[string, any]().Set("app", "nginx").Set("tier", "web")
	selOK := collections.NewHAMT[string, any]().Set("app", "nginx")
	selBadVal := collections.NewHAMT[string, any]().Set("app", "other")
	selBadKey := collections.NewHAMT[string, any]().Set("zone", "a")

	cases := []struct {
		name       string
		sub, super any
		want       bool
		wantErr    bool
	}{
		{"hamt subset holds", selOK, labels, true, false},
		{"hamt value mismatch", selBadVal, labels, false, false},
		{"hamt missing key", selBadKey, labels, false, false},
		{"empty hamt is subset", collections.NewHAMT[string, any](), labels, true, false},
		{"set subset holds", setOf(1, 2), setOf(1, 2, 3), true, false},
		{"set not subset", setOf(1, 4), setOf(1, 2, 3), false, false},
		{"empty set is subset", collections.NewSet[any](), setOf(1, 2), true, false},
		{"type mismatch map vs set", selOK, setOf(1, 2), false, true},
	}
	for _, c := range cases {
		got, err := IsSubsetValue(c.sub, c.super)
		if (err != nil) != c.wantErr {
			t.Errorf("%s: err=%v, wantErr=%v", c.name, err, c.wantErr)
			continue
		}
		if err == nil && got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func setOf(xs ...any) *collections.Set[any] {
	s := collections.NewSet[any]()
	for _, x := range xs {
		s = s.Add(x)
	}
	return s
}
