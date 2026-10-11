package collections

import (
	"testing"
)

func TestMap(t *testing.T) {
	m := NewMap[string, int]()
	
	m2 := m.Set("a", 1).Set("b", 2)

	if val, ok := m2.Get("a"); !ok || val != 1 {
		t.Errorf("Expected m2['a'] == 1, got %v", val)
	}

	// Immutability test
	if _, ok := m.Get("a"); ok {
		t.Errorf("Expected m to be empty after immutable Set")
	}

	m3 := m2.Set("a", 10)
	if val, ok := m3.Get("a"); !ok || val != 10 {
		t.Errorf("Expected m3['a'] == 10, got %v", val)
	}
	if val, ok := m2.Get("a"); !ok || val != 1 {
		t.Errorf("Expected m2['a'] == 1 after m3 mutation, got %v", val)
	}
}
