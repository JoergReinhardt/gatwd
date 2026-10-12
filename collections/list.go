package collections

import (
	"fmt"
	"iter"
	"strings"
)

// List represents an immutable, singly-linked list (a Cons list).
//
// In functional programming, lists are persistent. Adding an element
// (Cons) doesn't modify the existing list; it creates a new head that
// points to the existing list as its tail. This guarantees thread safety
// and eliminates side effects.
type List[T any] struct {
	head  T
	tail  *List[T]
	empty bool
}

// Nil returns an empty list.
func Nil[T any]() *List[T] {
	return &List[T]{empty: true}
}

// Cons (Construct) prepends an element to the front of a list, returning a new List.
func Cons[T any](head T, tail *List[T]) *List[T] {
	return &List[T]{
		head:  head,
		tail:  tail,
		empty: false,
	}
}

// IsEmpty returns true if the list has no elements.
func (l *List[T]) IsEmpty() bool {
	if l == nil {
		return true
	}
	return l.empty
}

// Head returns the first element of the list.
func (l *List[T]) Head() (T, bool) {
	if l.IsEmpty() {
		var zero T
		return zero, false
	}
	return l.head, true
}

// Elements returns a slice of all elements as any.
func (l *List[T]) Elements() []any {
	var elems []any
	for v := range l.Seq() {
		elems = append(elems, v)
	}
	return elems
}

// Tail returns the rest of the list after the head.
func (l *List[T]) Tail() (*List[T], bool) {
	if l.IsEmpty() {
		return nil, false
	}
	return l.tail, true
}

// Seq returns a lazy Go 1.23 iterator over the list.
//
// This bridges the gap between pure persistent functional structures
// and our modern iterator-based evaluator engine.
func (l *List[T]) Seq() iter.Seq[T] {
	return func(yield func(T) bool) {
		curr := l
		for !curr.IsEmpty() {
			if !yield(curr.head) {
				return
			}
			curr = curr.tail
		}
	}
}

// Graph exposes the internal structure for generic UTF-8 Tree visualizations
func (l *List[T]) Graph() any {
	if l == nil || l.empty {
		return "<empty List>"
	}
	return struct {
		Head T
		Tail *List[T]
	}{
		Head: l.head,
		Tail: l.tail,
	}
}

func (l *List[T]) String() string {
	if l.IsEmpty() {
		return "[]"
	}
	var elements []string
	current := l
	for !current.IsEmpty() {
		head, _ := current.Head()
		elements = append(elements, fmt.Sprintf("%v", head))
		current, _ = current.Tail()
	}
	return "[" + strings.Join(elements, " ") + "]"
}
