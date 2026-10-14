package collections

import (
	"fmt"
	"strings"
)

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

// NodeLabel implements visualizer.TreeNode: a B-tree node holds several keys rather than one
// value, so its label is all of them bracket-joined ("[k1 k2 k3]", the same convention
// List.String() uses) — one line per node instead of a synthetic "Keys"/"Children" struct layer.
func (b *BTree[T]) NodeLabel() any {
	if b == nil {
		return "<nil BTree>"
	}
	parts := make([]string, len(b.keys))
	for i, k := range b.keys {
		parts[i] = fmt.Sprintf("%v", k)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

// NodeChildren implements visualizer.TreeNode: this node's children, in order.
func (b *BTree[T]) NodeChildren() []any {
	if b == nil || len(b.children) == 0 {
		return nil
	}
	out := make([]any, len(b.children))
	for i, c := range b.children {
		out[i] = c
	}
	return out
}
