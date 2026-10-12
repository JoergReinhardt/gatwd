package collections

import (
	"fmt"
	"iter"
	"reflect"
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

// Len returns the number of elements in the list.
func (l *List[T]) Len() int {
	count := 0
	curr := l
	for curr != nil && !curr.IsEmpty() {
		count++
		curr = curr.tail
	}
	return count
}

// Reverse returns a new list with elements in reverse order.
func (l *List[T]) Reverse() *List[T] {
	rev := Nil[T]()
	curr := l
	for curr != nil && !curr.IsEmpty() {
		rev = Cons(curr.head, rev)
		curr = curr.tail
	}
	return rev
}

// Concat returns the concatenation of l and other (l ++ other).
func (l *List[T]) Concat(other *List[T]) *List[T] {
	if l == nil || l.IsEmpty() {
		if other == nil {
			return Nil[T]()
		}
		return other
	}
	if other == nil || other.IsEmpty() {
		return l
	}
	rev := l.Reverse()
	res := other
	curr := rev
	for curr != nil && !curr.IsEmpty() {
		res = Cons(curr.head, res)
		curr = curr.tail
	}
	return res
}

// Map applies f to every element in the list and returns a new List.
func (l *List[T]) Map(f func(T) T) *List[T] {
	if l == nil || l.IsEmpty() {
		return Nil[T]()
	}
	var items []T
	for v := range l.Seq() {
		items = append(items, f(v))
	}
	res := Nil[T]()
	for i := len(items) - 1; i >= 0; i-- {
		res = Cons(items[i], res)
	}
	return res
}

// FoldLeft reduces the list from left to right: f(...f(f(initial, x1), x2)...).
func (l *List[T]) FoldLeft(initial any, f func(acc any, item T) any) any {
	acc := initial
	curr := l
	for curr != nil && !curr.IsEmpty() {
		acc = f(acc, curr.head)
		curr = curr.tail
	}
	return acc
}

// FoldRight reduces the list from right to left: f(x1, f(x2, ...f(xn, initial)...)).
func (l *List[T]) FoldRight(initial any, f func(item T, acc any) any) any {
	if l == nil || l.IsEmpty() {
		return initial
	}
	rev := l.Reverse()
	acc := initial
	curr := rev
	for curr != nil && !curr.IsEmpty() {
		acc = f(curr.head, acc)
		curr = curr.tail
	}
	return acc
}

// Filter returns a new list containing only elements that satisfy predicate.
func (l *List[T]) Filter(predicate func(T) bool) *List[T] {
	var items []T
	for v := range l.Seq() {
		if predicate(v) {
			items = append(items, v)
		}
	}
	res := Nil[T]()
	for i := len(items) - 1; i >= 0; i-- {
		res = Cons(items[i], res)
	}
	return res
}

// Reject returns a new List containing the elements for which predicate is false
// — the complement of Filter.
func (l *List[T]) Reject(predicate func(T) bool) *List[T] {
	return l.Filter(func(x T) bool { return !predicate(x) })
}

// Equal returns true if l and other contain identical elements in the same order.
func (l *List[T]) Equal(other *List[T]) bool {
	if l == other {
		return true
	}
	c1, c2 := l, other
	for {
		e1 := c1 == nil || c1.IsEmpty()
		e2 := c2 == nil || c2.IsEmpty()
		if e1 && e2 {
			return true
		}
		if e1 != e2 {
			return false
		}
		if !reflect.DeepEqual(c1.head, c2.head) {
			return false
		}
		c1 = c1.tail
		c2 = c2.tail
	}
}

// IsList implements a marker method for type introspection.
func (l *List[T]) IsList() bool {
	return true
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
