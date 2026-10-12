package collections

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"iter"
	"reflect"
	"strings"
)

// Chunk represents an eager, contiguous bounded array for CPU cache efficiency.
// It wraps a standard Go slice, providing a fixed-capacity buffer to batch
// data for cache-friendly bulk processing.
type Chunk[T any] struct {
	items []T
	size  int
}

// NewChunk creates a new Chunk with the given capacity.
func NewChunk[T any](capacity int) *Chunk[T] {
	return &Chunk[T]{
		items: make([]T, 0, capacity),
		size:  capacity,
	}
}

// Append adds an element to the chunk. Returns true if added, false if chunk is full.
func (c *Chunk[T]) Append(item T) bool {
	if len(c.items) >= c.size {
		return false
	}
	c.items = append(c.items, item)
	return true
}

// Items returns the underlying slice of items.
func (c *Chunk[T]) Items() []T {
	return c.items
}

// Elements returns all elements in the chunk as a slice of any.
func (c *Chunk[T]) Elements() []any {
	var elems []any
	for _, item := range c.items {
		elems = append(elems, item)
	}
	return elems
}

// IsEmpty returns true if the chunk has no elements.
func (c *Chunk[T]) IsEmpty() bool {
	return len(c.items) == 0
}

// Get returns the element at index i, or false if out of bounds.
func (c *Chunk[T]) Get(i int) (T, bool) {
	if i < 0 || i >= len(c.items) {
		var zero T
		return zero, false
	}
	return c.items[i], true
}

// IsFull returns true if the chunk has reached its capacity.
func (c *Chunk[T]) IsFull() bool {
	return len(c.items) >= c.size
}

// Len returns the current number of elements in the chunk.
func (c *Chunk[T]) Len() int {
	return len(c.items)
}

// Capacity returns the maximum capacity of the chunk.
func (c *Chunk[T]) Capacity() int {
	return c.size
}

// Seq returns a lazy Go 1.23 iterator over the chunk's elements.
func (c *Chunk[T]) Seq() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, item := range c.items {
			if !yield(item) {
				return
			}
		}
	}
}

// Clear removes all elements from the chunk, reusing the underlying array.
func (c *Chunk[T]) Clear() {
	c.items = c.items[:0]
}

// Concat returns a new Chunk concatenating the items of c and other.
func (c *Chunk[T]) Concat(other *Chunk[T]) *Chunk[T] {
	cLen := 0
	if c != nil {
		cLen = len(c.items)
	}
	oLen := 0
	if other != nil {
		oLen = len(other.items)
	}
	totalCap := cLen + oLen
	if c != nil && c.size > totalCap {
		totalCap = c.size
	}
	res := NewChunk[T](totalCap)
	if c != nil {
		for _, item := range c.items {
			res.Append(item)
		}
	}
	if other != nil {
		for _, item := range other.items {
			res.Append(item)
		}
	}
	return res
}

// Map applies f to each element of the chunk and returns a new Chunk.
func (c *Chunk[T]) Map(f func(T) T) *Chunk[T] {
	if c == nil {
		return NewChunk[T](0)
	}
	res := NewChunk[T](c.size)
	for _, item := range c.items {
		res.Append(f(item))
	}
	return res
}

// Fold reduces all elements of the chunk using accumulator function f.
func (c *Chunk[T]) Fold(initial any, f func(acc any, item T) any) any {
	acc := initial
	if c == nil {
		return acc
	}
	for _, item := range c.items {
		acc = f(acc, item)
	}
	return acc
}

// Filter returns a new Chunk containing only elements satisfying predicate.
func (c *Chunk[T]) Filter(predicate func(T) bool) *Chunk[T] {
	if c == nil {
		return NewChunk[T](0)
	}
	res := NewChunk[T](c.size)
	for _, item := range c.items {
		if predicate(item) {
			res.Append(item)
		}
	}
	return res
}

// Equal returns true if c and other have identical elements in the same order.
func (c *Chunk[T]) Equal(other *Chunk[T]) bool {
	if c == other {
		return true
	}
	cLen := 0
	if c != nil {
		cLen = len(c.items)
	}
	oLen := 0
	if other != nil {
		oLen = len(other.items)
	}
	if cLen != oLen {
		return false
	}
	if cLen == 0 {
		return true
	}
	for i := 0; i < cLen; i++ {
		if !reflect.DeepEqual(c.items[i], other.items[i]) {
			return false
		}
	}
	return true
}

// IsChunk implements a marker method for type introspection.
func (c *Chunk[T]) IsChunk() bool {
	return true
}

func (c *Chunk[T]) String() string {
	if c.IsEmpty() {
		return "[]"
	}
	var elements []string
	for _, item := range c.items {
		elements = append(elements, fmt.Sprintf("%v", item))
	}
	return "[" + strings.Join(elements, " ") + "]"
}

// Reverse returns a new Chunk with elements in reverse order.
func (c *Chunk[T]) Reverse() *Chunk[T] {
	if c == nil || len(c.items) == 0 {
		return NewChunk[T](0)
	}
	res := NewChunk[T](c.size)
	for i := len(c.items) - 1; i >= 0; i-- {
		res.Append(c.items[i])
	}
	return res
}

// Slice returns a sub-chunk from start (inclusive) to end (exclusive), bounds-checked.
func (c *Chunk[T]) Slice(start, end int) *Chunk[T] {
	if c == nil || start >= len(c.items) || start >= end || end <= 0 {
		return NewChunk[T](0)
	}
	if start < 0 {
		start = 0
	}
	if end > len(c.items) {
		end = len(c.items)
	}
	subItems := c.items[start:end]
	res := NewChunk[T](len(subItems))
	for _, it := range subItems {
		res.Append(it)
	}
	return res
}

// MarshalJSON implements json.Marshaler.
func (c *Chunk[T]) MarshalJSON() ([]byte, error) {
	if c == nil || c.IsEmpty() {
		return []byte("[]"), nil
	}
	return json.Marshal(c.items)
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Chunk[T]) UnmarshalJSON(data []byte) error {
	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	c.items = items
	c.size = len(items)
	return nil
}

// Value implements driver.Valuer.
func (c *Chunk[T]) Value() (driver.Value, error) {
	b, err := c.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements sql.Scanner.
func (c *Chunk[T]) Scan(src any) error {
	var source []byte
	switch v := src.(type) {
	case string:
		source = []byte(v)
	case []byte:
		source = v
	case nil:
		source = []byte("[]")
	default:
		return fmt.Errorf("incompatible type for Chunk Scan: %T", src)
	}
	return c.UnmarshalJSON(source)
}

// Graph exposes the underlying slice for visualizer graph and tree formatting.
func (c *Chunk[T]) Graph() any {
	if c == nil || c.IsEmpty() {
		return "<empty Chunk>"
	}
	return c.items
}
