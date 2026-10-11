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
