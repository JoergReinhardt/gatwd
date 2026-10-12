package collections

import (
	"encoding/json"
	"testing"
)

func TestSet(t *testing.T) {
	s := NewSet[string]("a", "b", "c")

	if !s.Contains("a") || !s.Contains("b") || !s.Contains("c") {
		t.Errorf("Expected set to contain 'a', 'b', 'c'")
	}
	if s.Contains("d") {
		t.Errorf("Expected set not to contain 'd'")
	}
	if s.Len() != 3 {
		t.Fatalf("expected len 3, got %d", s.Len())
	}

	// Immutability test
	s2 := s.Add("d")
	if !s2.Contains("d") {
		t.Errorf("Expected s2 to contain 'd'")
	}
	if s.Contains("d") {
		t.Errorf("Expected s to NOT contain 'd' after immutable Add")
	}
	if s2.Len() != 4 {
		t.Fatalf("expected len 4, got %d", s2.Len())
	}

	// Remove
	s3 := s2.Remove("a")
	if s3.Contains("a") {
		t.Errorf("Expected 'a' to be removed")
	}
	if s3.Len() != 3 {
		t.Fatalf("expected len 3 after remove, got %d", s3.Len())
	}

	elements := s.Elements()
	if len(elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(elements))
	}

	// JSON Marshal & Unmarshal
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	var sUnmarshaled Set[string]
	if err := json.Unmarshal(data, &sUnmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if sUnmarshaled.Len() != 3 {
		t.Fatalf("expected unmarshaled len 3, got %d", sUnmarshaled.Len())
	}

	// SQL Valuer and Scanner
	val, err := s.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}
	var sScanned Set[string]
	if err := sScanned.Scan(val); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if sScanned.Len() != 3 {
		t.Fatalf("expected scanned len 3, got %d", sScanned.Len())
	}
}

func TestSetAlgebraAndFP(t *testing.T) {
	s1 := NewSet[int](1, 2, 3)
	s2 := NewSet[int](3, 4, 5)

	// Union
	union := s1.Union(s2)
	if union.Len() != 5 || !union.Contains(1) || !union.Contains(5) {
		t.Fatalf("expected union of {1,2,3} and {3,4,5} to have 5 elements, got %d", union.Len())
	}

	// Intersect
	intersect := s1.Intersect(s2)
	if intersect.Len() != 1 || !intersect.Contains(3) {
		t.Fatalf("expected intersect to have {3}, got %v", intersect.Elements())
	}

	// Diff
	diff := s1.Diff(s2)
	if diff.Len() != 2 || !diff.Contains(1) || !diff.Contains(2) || diff.Contains(3) {
		t.Fatalf("expected diff to have {1, 2}, got %v", diff.Elements())
	}

	// Subset & Equal
	sSub := NewSet[int](1, 2)
	if !sSub.IsSubset(s1) {
		t.Errorf("expected sSub to be subset of s1")
	}
	if s1.IsSubset(sSub) {
		t.Errorf("expected s1 not to be subset of sSub")
	}
	sEqual := NewSet[int](3, 2, 1)
	if !s1.Equal(sEqual) {
		t.Errorf("expected s1 and sEqual to be Equal")
	}

	// Map
	mapped := s1.Map(func(x int) int { return x * 10 })
	if mapped.Len() != 3 || !mapped.Contains(10) || !mapped.Contains(20) || !mapped.Contains(30) {
		t.Errorf("expected mapped elements {10, 20, 30}, got %v", mapped.Elements())
	}

	// Fold
	sum := s1.Fold(0, func(acc any, x int) any { return acc.(int) + x })
	if sum != 6 {
		t.Errorf("expected sum 6, got %v", sum)
	}

	// Filter
	filtered := s1.Filter(func(x int) bool { return x > 1 })
	if filtered.Len() != 2 || filtered.Contains(1) {
		t.Errorf("expected filtered {2, 3}, got %v", filtered.Elements())
	}

	// ElementsAny
	elemsAny := s1.ElementsAny()
	if len(elemsAny) != 3 {
		t.Errorf("expected 3 ElementsAny, got %d", len(elemsAny))
	}
}
