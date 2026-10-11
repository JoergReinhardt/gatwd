package collections

import (
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

	// Immutability test
	s2 := s.Add("d")
	if !s2.Contains("d") {
		t.Errorf("Expected s2 to contain 'd'")
	}
	if s.Contains("d") {
		t.Errorf("Expected s to NOT contain 'd' after immutable Add")
	}

	elements := s.Elements()
	if len(elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(elements))
	}
}
