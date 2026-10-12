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

// Union returns the set-theoretic union of s and other (s ∪ other).
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	if s == nil || s.hamt == nil {
		if other == nil {
			return NewSet[T]()
		}
		return other
	}
	if other == nil || other.hamt == nil {
		return s
	}
	base, toAdd := s, other
	if other.Len() > s.Len() {
		base, toAdd = other, s
	}
	h := base.hamt
	for item := range toAdd.Seq() {
		h = h.Set(item, struct{}{})
	}
	return &Set[T]{hamt: h}
}

// Intersect returns the set-theoretic intersection of s and other (s ∩ other).
func (s *Set[T]) Intersect(other *Set[T]) *Set[T] {
	res := NewSet[T]()
	if s == nil || s.hamt == nil || other == nil || other.hamt == nil {
		return res
	}
	smaller, larger := s, other
	if other.Len() < s.Len() {
		smaller, larger = other, s
	}
	for item := range smaller.Seq() {
		if larger.Contains(item) {
			res = res.Add(item)
		}
	}
	return res
}

// Diff returns the set-theoretic relative complement/difference of other from s (s \ other).
func (s *Set[T]) Diff(other *Set[T]) *Set[T] {
	if s == nil || s.hamt == nil {
		return NewSet[T]()
	}
	if other == nil || other.hamt == nil {
		return s
	}
	res := s
	for item := range other.Seq() {
		res = res.Remove(item)
	}
	return res
}

// IsSubset returns true if all elements of s are members of other (s ⊆ other).
func (s *Set[T]) IsSubset(other *Set[T]) bool {
	if s == nil || s.hamt == nil || s.Len() == 0 {
		return true
	}
	if other == nil || other.hamt == nil || other.Len() < s.Len() {
		return false
	}
	for item := range s.Seq() {
		if !other.Contains(item) {
			return false
		}
	}
	return true
}

// Equal returns true if s and other contain exactly the same elements.
func (s *Set[T]) Equal(other *Set[T]) bool {
	if s == other {
		return true
	}
	sLen := 0
	if s != nil {
		sLen = s.Len()
	}
	oLen := 0
	if other != nil {
		oLen = other.Len()
	}
	if sLen != oLen {
		return false
	}
	if sLen == 0 {
		return true
	}
	return s.IsSubset(other)
}

// Map returns a new Set containing the results of applying f to each element.
func (s *Set[T]) Map(f func(T) T) *Set[T] {
	res := NewSet[T]()
	if s == nil || s.hamt == nil {
		return res
	}
	for item := range s.Seq() {
		res = res.Add(f(item))
	}
	return res
}

// Fold reduces all elements of the set using the accumulator function f.
func (s *Set[T]) Fold(initial any, f func(acc any, item T) any) any {
	acc := initial
	if s == nil || s.hamt == nil {
		return acc
	}
	for item := range s.Seq() {
		acc = f(acc, item)
	}
	return acc
}

// Filter returns a new Set containing only the elements satisfying predicate.
func (s *Set[T]) Filter(predicate func(T) bool) *Set[T] {
	res := NewSet[T]()
	if s == nil || s.hamt == nil {
		return res
	}
	for item := range s.Seq() {
		if predicate(item) {
			res = res.Add(item)
		}
	}
	return res
}

// Reject returns a new Set containing the elements for which predicate is false
// — the complement of Filter.
func (s *Set[T]) Reject(predicate func(T) bool) *Set[T] {
	return s.Filter(func(x T) bool { return !predicate(x) })
}

// Bind (monadic flatMap) applies f to each element — where f itself yields a
// Set — and unions the results into a single Set (duplicates collapse, as the
// Set monad is idempotent).
func (s *Set[T]) Bind(f func(T) *Set[T]) *Set[T] {
	res := NewSet[T]()
	if s == nil || s.hamt == nil {
		return res
	}
	for item := range s.Seq() {
		sub := f(item)
		if sub == nil {
			continue
		}
		for v := range sub.Seq() {
			res = res.Add(v)
		}
	}
	return res
}

// ElementsAny returns all elements in the set as a slice of any.
func (s *Set[T]) ElementsAny() []any {
	if s == nil || s.hamt == nil {
		return nil
	}
	var res []any
	for item := range s.Seq() {
		res = append(res, item)
	}
	return res
}

// IsSet implements a marker method for type introspection.
func (s *Set[T]) IsSet() bool {
	return true
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
