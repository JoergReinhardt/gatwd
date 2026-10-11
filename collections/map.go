package collections

import (
	"database/sql/driver"
	"iter"
)

// Map is a fully persistent, immutable dictionary backed by a 32-way branching HAMT.
// It provides O(log32 N) (~O(1)) lookups, insertions, and deletions with structural sharing.
type Map[K comparable, V any] struct {
	hamt *HAMT[K, V]
}

// NewMap creates a new persistent Map backed by a HAMT.
func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{hamt: NewHAMT[K, V]()}
}

// Set associates a key with a value, returning a new Map with structural sharing.
func (m *Map[K, V]) Set(k K, v V) *Map[K, V] {
	if m == nil || m.hamt == nil {
		return &Map[K, V]{hamt: NewHAMT[K, V]().Set(k, v)}
	}
	return &Map[K, V]{hamt: m.hamt.Set(k, v)}
}

// Get retrieves the value associated with the key.
func (m *Map[K, V]) Get(k K) (V, bool) {
	if m == nil || m.hamt == nil {
		var zero V
		return zero, false
	}
	return m.hamt.Get(k)
}

// Delete removes a key, returning a new Map with structural sharing.
func (m *Map[K, V]) Delete(k K) *Map[K, V] {
	if m == nil || m.hamt == nil {
		return m
	}
	return &Map[K, V]{hamt: m.hamt.Delete(k)}
}

// Len returns the number of key-value pairs in the Map.
func (m *Map[K, V]) Len() int {
	if m == nil || m.hamt == nil {
		return 0
	}
	return m.hamt.Len()
}

// Seq returns a Go 1.23 standard iter.Seq2 of (key, value) pairs.
func (m *Map[K, V]) Seq() iter.Seq2[K, V] {
	if m == nil || m.hamt == nil {
		return func(yield func(K, V) bool) {}
	}
	return m.hamt.Seq()
}

// String implements fmt.Stringer.
func (m *Map[K, V]) String() string {
	return "Map"
}

// MarshalJSON implements json.Marshaler.
func (m *Map[K, V]) MarshalJSON() ([]byte, error) {
	if m == nil || m.hamt == nil {
		return []byte("{}"), nil
	}
	return m.hamt.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *Map[K, V]) UnmarshalJSON(data []byte) error {
	h := NewHAMT[K, V]()
	if err := h.UnmarshalJSON(data); err != nil {
		return err
	}
	m.hamt = h
	return nil
}

// Value implements driver.Valuer.
func (m *Map[K, V]) Value() (driver.Value, error) {
	if m == nil || m.hamt == nil {
		return "{}", nil
	}
	return m.hamt.Value()
}

// Scan implements sql.Scanner.
func (m *Map[K, V]) Scan(src any) error {
	h := NewHAMT[K, V]()
	if err := h.Scan(src); err != nil {
		return err
	}
	m.hamt = h
	return nil
}

// Graph exposes the persistent Map's contents for visualizer tree/DOT formatting.
func (m *Map[K, V]) Graph() any {
	if m == nil || m.hamt == nil {
		return "<empty Map>"
	}
	return m.hamt.Graph()
}
