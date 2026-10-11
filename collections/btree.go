package collections

// BTree is a simplified immutable B-Tree node.
type BTree[T any] struct {
	keys     []T
	children []*BTree[T]
}

// NewBTree creates a new empty BTree.
func NewBTree[T any]() *BTree[T] {
	return &BTree[T]{keys: []T{}, children: []*BTree[T]{}}
}

// NewBTreeNode creates a new BTree node manually.
func NewBTreeNode[T any](keys []T, children []*BTree[T]) *BTree[T] {
	return &BTree[T]{keys: keys, children: children}
}

// Graph implements Graphable for the Visualizer
func (b *BTree[T]) Graph() any {
	if b == nil {
		return "<nil BTree>"
	}
	if len(b.children) == 0 {
		return struct{ Keys []T }{Keys: b.keys}
	}
	return struct {
		Keys     []T
		Children []*BTree[T]
	}{Keys: b.keys, Children: b.children}
}
