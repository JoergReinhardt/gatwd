package collections

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"
)

// foldKey is a deliberately non-Go equality: case-insensitive text. "A" and "a" are one
// member although Go's == tells them apart; "" has no key (the error path).
func foldKey(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("the empty text has no key")
	}
	return strings.ToLower(s), nil
}

func sorted(s *Setoid[string]) []string {
	out := s.Elements()
	sort.Strings(out)
	return out
}

func mustSetoid(t *testing.T, items ...string) *Setoid[string] {
	t.Helper()
	s, err := NewSetoid(foldKey, items...)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSetoidIdentityIsTheKeyFunctionNotGoEquality(t *testing.T) {
	s := mustSetoid(t, "A", "a", "b")
	if s.Len() != 2 {
		t.Fatalf("Len = %d, want 2 (A and a are one member)", s.Len())
	}
	if got := sorted(s); !slices.Equal(got, []string{"A", "b"}) {
		t.Errorf("members = %v, want the first to arrive to stay: [A b]", got)
	}
	if ok, err := s.Contains("B"); err != nil || !ok {
		t.Errorf("Contains(B) = %v, %v; want true", ok, err)
	}
	if ok, _ := s.Contains("c"); ok {
		t.Error("Contains(c) = true, want false")
	}
}

func TestSetoidIsPersistent(t *testing.T) {
	base := mustSetoid(t, "a")
	more, _ := base.Add("b")
	less, _ := more.Remove("A")
	if base.Len() != 1 || more.Len() != 2 || less.Len() != 1 {
		t.Errorf("lens = %d, %d, %d; want 1, 2, 1", base.Len(), more.Len(), less.Len())
	}
	if ok, _ := base.Contains("b"); ok {
		t.Error("Add mutated the receiver")
	}
	if ok, _ := more.Contains("a"); !ok {
		t.Error("Remove mutated the receiver")
	}
}

func TestSetoidKeyErrorsAreReturnedNotSwallowed(t *testing.T) {
	if _, err := NewSetoid(foldKey, "a", ""); err == nil {
		t.Error("NewSetoid with an unkeyable item = nil error")
	}
	s := mustSetoid(t, "a")
	if _, err := s.Add(""); err == nil {
		t.Error("Add of an unkeyable item = nil error")
	}
	if _, err := s.Remove(""); err == nil {
		t.Error("Remove of an unkeyable item = nil error")
	}
	if _, err := s.Contains(""); err == nil {
		t.Error("Contains of an unkeyable item = nil error")
	}
	if _, err := s.Map(func(string) string { return "" }); err == nil {
		t.Error("Map to an unkeyable result = nil error")
	}
	if _, err := NewSetoid[string](nil); err == nil {
		t.Error("NewSetoid(nil) = nil error")
	}
}

func TestSetoidAlgebra(t *testing.T) {
	a := mustSetoid(t, "A", "b", "c")
	b := mustSetoid(t, "B", "C", "d")
	cases := []struct {
		name string
		got  *Setoid[string]
		want []string
	}{
		{"union keeps the receiver's element", a.Union(b), []string{"A", "b", "c", "d"}},
		{"intersect", a.Intersect(b), []string{"b", "c"}},
		{"diff", a.Diff(b), []string{"A"}},
		{"filter", a.Filter(func(s string) bool { return s != "b" }), []string{"A", "c"}},
		{"reject", a.Reject(func(s string) bool { return s != "b" }), []string{"b"}},
	}
	for _, c := range cases {
		if got := sorted(c.got); !slices.Equal(got, c.want) {
			t.Errorf("%s = %v, want %v", c.name, got, c.want)
		}
	}
	if !mustSetoid(t, "x", "Y").Equal(mustSetoid(t, "y", "X")) {
		t.Error("sets with the same members under the key function are not Equal")
	}
	if mustSetoid(t, "x").Equal(mustSetoid(t, "x", "y")) {
		t.Error("sets of different size are Equal")
	}
	if !mustSetoid(t, "x").IsSubset(mustSetoid(t, "X", "y")) || mustSetoid(t, "z").IsSubset(a) {
		t.Error("IsSubset wrong")
	}
	if !mustSetoid(t).IsSubset(a) {
		t.Error("the empty set is a subset of every set")
	}
}

func TestSetoidMapFoldAndReaders(t *testing.T) {
	s := mustSetoid(t, "a", "B")
	m, err := s.Map(strings.ToUpper)
	if err != nil || m.Len() != 2 {
		t.Fatalf("Map = %v, %v", m, err)
	}
	collapsed, err := s.Map(func(string) string { return "z" })
	if err != nil || collapsed.Len() != 1 {
		t.Errorf("Map collapsing every member: Len = %d, %v; want 1", collapsed.Len(), err)
	}
	n := s.Fold(0, func(acc any, _ string) any { return acc.(int) + 1 })
	if n != 2 {
		t.Errorf("Fold = %v, want 2", n)
	}
	if !s.IsSet() || s.String() != "Set" || len(s.ElementsAny()) != 2 {
		t.Error("the duck-typed set readers are wrong")
	}
	if b, err := s.MarshalJSON(); err != nil || len(b) < 2 || b[0] != '[' {
		t.Errorf("MarshalJSON = %s, %v", b, err)
	}
	var empty *Setoid[string]
	if empty.Len() != 0 || len(empty.Elements()) != 0 {
		t.Error("a nil Setoid is not empty")
	}
}
