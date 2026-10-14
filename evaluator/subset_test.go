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

	// Nested cases (the manifest ⊆ live shape, docs/manifest_live_parity_design.md
	// §2.2/§3): a matched value that is itself map-like or collection-like must
	// only CONTAIN its counterpart's declared fields, not equal it — this is what
	// used to be flat-equal-compared (the bug) and is now recursive.
	nestedSuper := collections.NewHAMT[string, any]().
		Set("spec", collections.NewHAMT[string, any]().Set("replicas", 3).Set("selector", "app=web"))
	nestedSubHolds := collections.NewHAMT[string, any]().
		Set("spec", collections.NewHAMT[string, any]().Set("replicas", 3)) // super's spec has MORE (selector) — still a subset
	nestedSubMismatch := collections.NewHAMT[string, any]().
		Set("spec", collections.NewHAMT[string, any]().Set("replicas", 5)) // a nested VALUE differs — still caught
	listSuper := collections.NewHAMT[string, any]().
		Set("containers", []any{collections.NewHAMT[string, any]().Set("name", "app").Set("image", "nginx")})
	listSubHolds := collections.NewHAMT[string, any]().
		Set("containers", []any{collections.NewHAMT[string, any]().Set("name", "app")}) // super's element has MORE (image)

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
		{"nested hamt-in-hamt subset holds (super has more)", nestedSubHolds, nestedSuper, true, false},
		{"nested hamt-in-hamt value mismatch still caught", nestedSubMismatch, nestedSuper, false, false},
		{"nested list-in-hamt subset holds (super's element has more)", listSubHolds, listSuper, true, false},
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
