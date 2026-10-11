package collections

import (
	"encoding/json"
	"testing"
)

func TestMap(t *testing.T) {
	m := NewMap[string, int]()
	if m == nil {
		t.Fatal("Expected map to not be nil")
	}

	m1 := m.Set("foo", 10).Set("bar", 20)
	if m1.Len() != 2 {
		t.Fatalf("expected len 2, got %d", m1.Len())
	}

	v, ok := m1.Get("foo")
	if !ok || v != 10 {
		t.Fatalf("expected foo=10, got %d", v)
	}

	// Immutability
	if m.Len() != 0 {
		t.Fatalf("original map mutated! len=%d", m.Len())
	}

	// Delete
	m2 := m1.Delete("foo")
	if m2.Len() != 1 {
		t.Fatalf("expected len 1 after delete, got %d", m2.Len())
	}
	if _, ok := m2.Get("foo"); ok {
		t.Fatalf("key 'foo' should be deleted")
	}

	// JSON Marshal & Unmarshal
	data, err := json.Marshal(m1)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	var mUnmarshaled Map[string, int]
	if err := json.Unmarshal(data, &mUnmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if mUnmarshaled.Len() != 2 {
		t.Fatalf("expected unmarshaled len 2, got %d", mUnmarshaled.Len())
	}

	// SQL Valuer and Scanner
	val, err := m1.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}
	var mScanned Map[string, int]
	if err := mScanned.Scan(val); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if mScanned.Len() != 2 {
		t.Fatalf("expected scanned len 2, got %d", mScanned.Len())
	}
}
