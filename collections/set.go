package collections

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"iter"
)

// Set is a fully persistent, immutable set backed by a 32-way branching HAMT.
// It provides O(log32 N) (~O(1)) membership checks and additions with structural sharing.
type Set[T comparable] struct {
	hamt *HAMT[T, struct{}]
}

// NewSet creates a new persistent Set containing optional initial items.
func NewSet[T comparable](items ...T) *Set[T] {
	h := NewHAMT[T, struct{}]()
	for _, item := range items {
		h = h.Set(item, struct{}{})
	}
	return &Set[T]{hamt: h}
}

// Contains checks if the item is present in the Set.
func (s *Set[T]) Contains(item T) bool {
	if s == nil || s.hamt == nil {
		return false
	}
	_, exists := s.hamt.Get(item)
	return exists
}

// Elements returns a slice of all items in the Set.
func (s *Set[T]) Elements() []T {
	if s == nil || s.hamt == nil {
		return nil
	}
	var res []T
	for k := range s.hamt.Keys() {
		res = append(res, k)
	}
	return res
}

// Len returns the number of items in the Set.
func (s *Set[T]) Len() int {
	if s == nil || s.hamt == nil {
		return 0
	}
	return s.hamt.Len()
}

// Seq returns a Go 1.23 standard iter.Seq over all items.
func (s *Set[T]) Seq() iter.Seq[T] {
	if s == nil || s.hamt == nil {
		return func(yield func(T) bool) {}
	}
	return s.hamt.Keys()
}

// Add returns a new Set containing the item, with structural sharing.
func (s *Set[T]) Add(item T) *Set[T] {
	if s == nil || s.hamt == nil {
		return &Set[T]{hamt: NewHAMT[T, struct{}]().Set(item, struct{}{})}
	}
	return &Set[T]{hamt: s.hamt.Set(item, struct{}{})}
}

// Remove returns a new Set without the item, with structural sharing.
func (s *Set[T]) Remove(item T) *Set[T] {
	if s == nil || s.hamt == nil {
		return s
	}
	return &Set[T]{hamt: s.hamt.Delete(item)}
}

// String implements fmt.Stringer.
func (s *Set[T]) String() string {
	return "Set"
}

// MarshalJSON implements json.Marshaler.
func (s *Set[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Elements())
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Set[T]) UnmarshalJSON(data []byte) error {
	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	s.hamt = NewSet[T](items...).hamt
	return nil
}

// Value implements driver.Valuer.
func (s *Set[T]) Value() (driver.Value, error) {
	b, err := s.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements sql.Scanner.
func (s *Set[T]) Scan(src any) error {
	var source []byte
	switch v := src.(type) {
	case string:
		source = []byte(v)
	case []byte:
		source = v
	case nil:
		source = []byte("[]")
	default:
		return fmt.Errorf("incompatible type for Set Scan: %T", src)
	}
	return s.UnmarshalJSON(source)
}

// Graph exposes the Set's elements as a slice for visualizer tree/DOT formatting.
func (s *Set[T]) Graph() any {
	if s == nil || s.hamt == nil || s.Len() == 0 {
		return "<empty Set>"
	}
	var items []T
	for item := range s.Seq() {
		items = append(items, item)
	}
	return items
}
