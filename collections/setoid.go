package collections

import (
	"encoding/json"
	"fmt"
	"iter"
)

// Setoid is a persistent set whose notion of "the same element" is supplied by the
// caller, not by Go's == on the element type.
//
// What is the concept? In mathematics a set needs an equivalence relation: which
// elements count as one. A set together with that relation is a setoid (the word
// constructive mathematics and type theory use). Set[T] borrows the relation from Go:
// two elements are the same when `==` says so on the Go values, which is the wrong
// relation for a dynamic language. A list is a pointer, so two lists with the same
// elements are two members; the Int 1 and the Float 1.0 are different Go values; a
// slice cannot be hashed at all. A Setoid asks the caller for the relation instead.
//
// Why is it needed? The DSL's Set must follow the language's own `==` (`[1, 2]` and
// `[1, 2]` are one member, `1` and `1.0` are one, `1` and "1" are two, a Time seen in two
// zones is one). Printing the element and keeping the text loses that (a list of Ints and
// a String that looks like it collapse), so the elements are kept as themselves and
// matched through a canonical string key.
//
// How is it implemented? The key function maps an element to a canonical string such that
// two elements are the same exactly when their keys are equal (the DSL passes
// evaluator.EncodeKey, the key design its HAMT already uses). The members live in a
// HAMT[string, T] from key to the element, so membership is O(log32 N), the structure is
// shared between versions, and an element is read back as the value that was stored, never
// decoded from text. The key function may fail (a function value has no equality); the
// operations that need it then return the error instead of guessing.
//
// The first element to arrive for a key stays: adding a member that is already present
// is a no-op, and Union keeps the receiver's elements for the keys both sets hold. Two
// Setoids that are combined (Union, Intersect, Diff, IsSubset, Equal) must have been built
// with the same key function; the members' stored keys are compared and the key function
// is not consulted for them.
//
// Example:
//
//	s, _ := NewSetoid(key)       // key: canonical string of an element
//	s, _ = s.Add([]any{1, 2})
//	s, _ = s.Add([]any{1, 2})    // the same member: Len() stays 1
//
// T is not constrained to be comparable: the key function is the whole of the relation.
type Setoid[T any] struct {
	members *HAMT[string, T]
	key     func(T) (string, error)
}

// NewSetoid creates a Setoid with the given key function and optional initial items. The
// key function must map equal elements to equal keys and unequal elements to different
// keys; it must not be nil. It reports the key function's error for the first item it
// cannot key.
func NewSetoid[T any](key func(T) (string, error), items ...T) (*Setoid[T], error) {
	if key == nil {
		return nil, fmt.Errorf("a Setoid needs a key function")
	}
	s := &Setoid[T]{members: NewHAMT[string, T](), key: key}
	for _, it := range items {
		var err error
		if s, err = s.Add(it); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Add returns a Setoid that also holds item. A member with the same key already present is
// kept (the receiver is returned unchanged).
func (s *Setoid[T]) Add(item T) (*Setoid[T], error) {
	k, err := s.key(item)
	if err != nil {
		return s, err
	}
	if _, present := s.members.Get(k); present {
		return s, nil
	}
	return &Setoid[T]{members: s.members.Set(k, item), key: s.key}, nil
}

// Remove returns a Setoid without the member equal to item (a missing one is not an error).
func (s *Setoid[T]) Remove(item T) (*Setoid[T], error) {
	k, err := s.key(item)
	if err != nil {
		return s, err
	}
	if _, present := s.members.Get(k); !present {
		return s, nil
	}
	return &Setoid[T]{members: s.members.Delete(k), key: s.key}, nil
}

// Contains reports whether a member equal to item is present. It is an error when item
// has no key.
func (s *Setoid[T]) Contains(item T) (bool, error) {
	k, err := s.key(item)
	if err != nil {
		return false, err
	}
	_, ok := s.members.Get(k)
	return ok, nil
}

// Len returns the number of members.
func (s *Setoid[T]) Len() int {
	if s == nil || s.members == nil {
		return 0
	}
	return s.members.Len()
}

// Seq iterates the members in no particular order.
func (s *Setoid[T]) Seq() iter.Seq[T] {
	return func(yield func(T) bool) {
		if s == nil || s.members == nil {
			return
		}
		for _, v := range s.members.Seq() {
			if !yield(v) {
				return
			}
		}
	}
}

// Elements returns the members as a slice, in no particular order.
func (s *Setoid[T]) Elements() []T {
	var out []T
	for v := range s.Seq() {
		out = append(out, v)
	}
	return out
}

// ElementsAny returns the members as a slice of any (the duck-typed reader the DSL's
// collection functions use).
func (s *Setoid[T]) ElementsAny() []any {
	var out []any
	for v := range s.Seq() {
		out = append(out, v)
	}
	return out
}

// IsSet marks the Setoid as a Set for type introspection (the DSL's `Set` kind).
func (s *Setoid[T]) IsSet() bool { return true }

// String implements fmt.Stringer; like Set it prints as its kind.
func (s *Setoid[T]) String() string { return "Set" }

// Graph exposes the members as a slice for the visualizer's tree/DOT output.
func (s *Setoid[T]) Graph() any {
	if s.Len() == 0 {
		return "<empty Set>"
	}
	return s.Elements()
}

// MarshalJSON implements json.Marshaler: the members as a JSON array.
func (s *Setoid[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Elements())
}

// Union returns the members of both sets; for a key both hold, the receiver's element stays.
func (s *Setoid[T]) Union(other *Setoid[T]) *Setoid[T] {
	if other.Len() == 0 {
		return s
	}
	if s.Len() == 0 {
		return other
	}
	h := s.members
	for k, v := range other.members.Seq() {
		if _, present := h.Get(k); !present {
			h = h.Set(k, v)
		}
	}
	return &Setoid[T]{members: h, key: s.key}
}

// Intersect returns the receiver's members that other also holds.
func (s *Setoid[T]) Intersect(other *Setoid[T]) *Setoid[T] {
	h := NewHAMT[string, T]()
	if s.Len() > 0 && other.Len() > 0 {
		for k, v := range s.members.Seq() {
			if _, ok := other.members.Get(k); ok {
				h = h.Set(k, v)
			}
		}
	}
	return &Setoid[T]{members: h, key: s.key}
}

// Diff returns the receiver's members that other does not hold (s \ other).
func (s *Setoid[T]) Diff(other *Setoid[T]) *Setoid[T] {
	if other.Len() == 0 || s.Len() == 0 {
		return s
	}
	h := s.members
	for k := range other.members.Keys() {
		h = h.Delete(k)
	}
	return &Setoid[T]{members: h, key: s.key}
}

// IsSubset reports whether every member of s is a member of other.
func (s *Setoid[T]) IsSubset(other *Setoid[T]) bool {
	if s.Len() == 0 {
		return true
	}
	if other.Len() < s.Len() {
		return false
	}
	for k := range s.members.Keys() {
		if _, ok := other.members.Get(k); !ok {
			return false
		}
	}
	return true
}

// Equal reports whether s and other hold the same members.
func (s *Setoid[T]) Equal(other *Setoid[T]) bool {
	return s.Len() == other.Len() && s.IsSubset(other)
}

// Filter returns the members for which predicate holds.
func (s *Setoid[T]) Filter(predicate func(T) bool) *Setoid[T] {
	return &Setoid[T]{members: s.members.Filter(func(_ string, v T) bool { return predicate(v) }), key: s.key}
}

// Reject returns the members for which predicate does not hold.
func (s *Setoid[T]) Reject(predicate func(T) bool) *Setoid[T] {
	return s.Filter(func(v T) bool { return !predicate(v) })
}

// Map returns the Setoid of f applied to every member; results that are the same element
// collapse to one, and a result without a key is an error.
func (s *Setoid[T]) Map(f func(T) T) (*Setoid[T], error) {
	out := &Setoid[T]{members: NewHAMT[string, T](), key: s.key}
	for v := range s.Seq() {
		var err error
		if out, err = out.Add(f(v)); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// Fold reduces the members (in no particular order) to one value, starting from initial.
func (s *Setoid[T]) Fold(initial any, f func(acc any, item T) any) any {
	acc := initial
	for v := range s.Seq() {
		acc = f(acc, v)
	}
	return acc
}
