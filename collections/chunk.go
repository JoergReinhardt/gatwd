package collections

import (
	"fmt"
	"iter"
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
