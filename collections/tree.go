package collections

import (
	"iter"
	"reflect"
)

// Tree represents an immutable, persistent Rose Tree.
//
// A Rose Tree is a tree where each node has a value and an ordered
// collection of children (which can be any number, including zero).
// It is the foundational data structure for parsing ASTs, representing
// hierarchical scopes, or modeling infrastructure graphs.
//
// Because the children are stored in an immutable `List`, the entire
// Tree is thread-safe and can be safely passed across goroutines.
type Tree[T any] struct {
	value    T
	children *List[*Tree[T]]
}

// NewTree constructs a new Tree node with a value and a list of children.
func NewTree[T any](value T, children *List[*Tree[T]]) *Tree[T] {
	if children == nil {
		children = Nil[*Tree[T]]()
	}
	return &Tree[T]{
		value:    value,
		children: children,
	}
}

// Leaf constructs a Tree node with a value and no children.
func Leaf[T any](value T) *Tree[T] {
	return &Tree[T]{
		value:    value,
		children: Nil[*Tree[T]](),
	}
}

// Value returns the value stored at this tree node.
func (t *Tree[T]) Value() T {
	return t.value
}

// Children returns the immutable list of child trees.
func (t *Tree[T]) Children() *List[*Tree[T]] {
	return t.children
}

// WalkDepthFirst returns a lazy iterator that traverses the tree in a depth-first manner.
//
// This is incredibly powerful for evaluating ASTs or crawling graphs.
// It flattens the multidimensional tree into a lazy 1D stream that our
// evaluator loop can process using `Bind`.
func (t *Tree[T]) WalkDepthFirst() iter.Seq[T] {
	return func(yield func(T) bool) {
		// Yield current node
		if !yield(t.value) {
			return
		}

		// Recursively yield children
		for child := range t.children.Seq() {
			for v := range child.WalkDepthFirst() {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Map applies f to the value of this node and recursively to all children.
func (t *Tree[T]) Map(f func(T) T) *Tree[T] {
	if t == nil {
		return nil
	}
	var newChildrenList *List[*Tree[T]] = Nil[*Tree[T]]()
	if t.children != nil && !t.children.IsEmpty() {
		var mappedChildren []*Tree[T]
		for child := range t.children.Seq() {
			mappedChildren = append(mappedChildren, child.Map(f))
		}
		for i := len(mappedChildren) - 1; i >= 0; i-- {
			newChildrenList = Cons(mappedChildren[i], newChildrenList)
		}
	}
	return NewTree[T](f(t.value), newChildrenList)
}

// Fold evaluates the tree bottom-up via catamorphism: each node combines its value
// and the folded results of all its children.
func (t *Tree[T]) Fold(f func(val T, childResults []any) any) any {
	if t == nil {
		return nil
	}
	var childResults []any
	if t.children != nil {
		for child := range t.children.Seq() {
			childResults = append(childResults, child.Fold(f))
		}
	}
	return f(t.value, childResults)
}

// Equal returns true if t and other are structurally equal (matching values and children).
func (t *Tree[T]) Equal(other *Tree[T]) bool {
	if t == other {
		return true
	}
	if t == nil || other == nil {
		return false
	}
	if !reflect.DeepEqual(t.value, other.value) {
		return false
	}
	tChildLen := 0
	if t.children != nil {
		tChildLen = t.children.Len()
	}
	oChildLen := 0
	if other.children != nil {
		oChildLen = other.children.Len()
	}
	if tChildLen != oChildLen {
		return false
	}
	if tChildLen == 0 {
		return true
	}
	c1 := t.children
	c2 := other.children
	for c1 != nil && !c1.IsEmpty() && c2 != nil && !c2.IsEmpty() {
		if !c1.head.Equal(c2.head) {
			return false
		}
		c1 = c1.tail
		c2 = c2.tail
	}
	return true
}

// IsTree implements a marker method for type introspection.
func (t *Tree[T]) IsTree() bool {
	return true
}

// Graph exposes the internal structure for generic UTF-8 Tree visualizations
func (t *Tree[T]) Graph() any {
	if t == nil {
		return "<nil Tree>"
	}
	return struct {
		Value    T
		Children *List[*Tree[T]]
	}{
		Value:    t.value,
		Children: t.children,
	}
}
