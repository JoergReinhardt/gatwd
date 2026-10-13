package collections

import "iter"

// BST is an immutable functional Binary Search Tree.
type BST[T any] struct {
	value T
	left  *BST[T]
	right *BST[T]
}

// LeafBST creates a new Leaf node.
func LeafBST[T any](v T) *BST[T] {
	return &BST[T]{value: v}
}

// NodeBST creates a new branch node.
func NodeBST[T any](v T, l *BST[T], r *BST[T]) *BST[T] {
	return &BST[T]{value: v, left: l, right: r}
}

// Insert immutably inserts a value into the BST using the provided comparator.
func (b *BST[T]) Insert(v T, cmp func(a, b T) int) *BST[T] {
	if b == nil {
		return LeafBST(v)
	}

	if cmp(v, b.value) < 0 {
		return NodeBST(b.value, b.left.Insert(v, cmp), b.right)
	} else if cmp(v, b.value) > 0 {
		return NodeBST(b.value, b.left, b.right.Insert(v, cmp))
	}
	return b // Ignore duplicates
}

// Graph implements Graphable for the Visualizer
func (b *BST[T]) Graph() any {
	if b == nil {
		return "<nil BST>"
	}
	return struct {
		Value T
		Left  *BST[T]
		Right *BST[T]
	}{Value: b.value, Left: b.left, Right: b.right}
}

// Contains reports whether v is in the tree under the comparator cmp (the same total
// order Insert used). A nil tree is empty and contains nothing. It runs in O(depth)
// by following a single root-to-leaf path.
func (b *BST[T]) Contains(v T, cmp func(a, b T) int) bool {
	for b != nil {
		switch c := cmp(v, b.value); {
		case c < 0:
			b = b.left
		case c > 0:
			b = b.right
		default:
			return true
		}
	}
	return false
}

// Size returns the number of values stored (duplicates were ignored on Insert). A
// nil tree has size 0.
func (b *BST[T]) Size() int {
	if b == nil {
		return 0
	}
	return 1 + b.left.Size() + b.right.Size()
}

// Seq yields the values in ascending (in-order) sequence, lazily, so a BST is a
// Foldable: a search tree's catamorphism is the sorted sequence. A nil tree yields
// nothing.
func (b *BST[T]) Seq() iter.Seq[T] {
	return func(yield func(T) bool) {
		b.walk(yield)
	}
}

func (b *BST[T]) walk(yield func(T) bool) bool {
	if b == nil {
		return true
	}
	return b.left.walk(yield) && yield(b.value) && b.right.walk(yield)
}

// IsBST marks a BST for structural equality in package functions, which sits below
// collections and so cannot name the type.
func (b *BST[T]) IsBST() bool { return true }

// BSTElements returns the values in ascending order as a slice of any: the sequence that
// equality compares, so two search trees holding the same values are equal whatever
// insertion order gave them their shape.
func (b *BST[T]) BSTElements() []any {
	var out []any
	for v := range b.Seq() {
		out = append(out, v)
	}
	return out
}
