package collections

import "iter"

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
