package collections

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
