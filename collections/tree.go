package collections

import (
	"encoding/json"
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

// Filter returns a new Tree that keeps the root and, recursively, only those
// child subtrees whose root value satisfies the predicate. A child whose value
// fails the predicate is pruned together with all of its descendants. The root
// value is always retained, because a Tree must have a root — this completes the
// Filter row across every collection type.
func (t *Tree[T]) Filter(predicate func(T) bool) *Tree[T] {
	if t == nil {
		return nil
	}
	var kept []*Tree[T]
	if t.children != nil {
		for child := range t.children.Seq() {
			if predicate(child.value) {
				kept = append(kept, child.Filter(predicate))
			}
		}
	}
	newChildren := Nil[*Tree[T]]()
	for i := len(kept) - 1; i >= 0; i-- {
		newChildren = Cons(kept[i], newChildren)
	}
	return NewTree[T](t.value, newChildren)
}

// Reject returns a new Tree pruning the child subtrees whose root value satisfies
// the predicate — the complement of Filter.
func (t *Tree[T]) Reject(predicate func(T) bool) *Tree[T] {
	return t.Filter(func(x T) bool { return !predicate(x) })
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

// AddChild returns a new Tree with child appended to its list of children.
func (t *Tree[T]) AddChild(child *Tree[T]) *Tree[T] {
	if t == nil {
		return child
	}
	if child == nil {
		return t
	}
	var newChildren []*Tree[T]
	if t.children != nil {
		for c := range t.children.Seq() {
			newChildren = append(newChildren, c)
		}
	}
	newChildren = append(newChildren, child)
	var cl *List[*Tree[T]] = Nil[*Tree[T]]()
	for i := len(newChildren) - 1; i >= 0; i-- {
		cl = Cons(newChildren[i], cl)
	}
	return NewTree[T](t.value, cl)
}

// Size returns the total count of nodes in the tree (this node plus all descendants).
func (t *Tree[T]) Size() int {
	if t == nil {
		return 0
	}
	count := 1
	if t.children != nil {
		for c := range t.children.Seq() {
			count += c.Size()
		}
	}
	return count
}

// Depth returns the maximum height/depth of the tree (1 for leaf).
func (t *Tree[T]) Depth() int {
	if t == nil {
		return 0
	}
	maxChildDepth := 0
	if t.children != nil {
		for c := range t.children.Seq() {
			cd := c.Depth()
			if cd > maxChildDepth {
				maxChildDepth = cd
			}
		}
	}
	return 1 + maxChildDepth
}

// Flatten returns a persistent List of all tree node values in depth-first order.
func (t *Tree[T]) Flatten() *List[T] {
	if t == nil {
		return Nil[T]()
	}
	var items []T
	for v := range t.WalkDepthFirst() {
		items = append(items, v)
	}
	res := Nil[T]()
	for i := len(items) - 1; i >= 0; i-- {
		res = Cons(items[i], res)
	}
	return res
}

// FlattenChunk returns a bounded contiguous Chunk of all tree node values in depth-first order.
func (t *Tree[T]) FlattenChunk() *Chunk[T] {
	if t == nil {
		return NewChunk[T](0)
	}
	sz := t.Size()
	c := NewChunk[T](sz)
	for v := range t.WalkDepthFirst() {
		c.Append(v)
	}
	return c
}

type treeJSON[T any] struct {
	Value    T          `json:"value"`
	Children []*Tree[T] `json:"children,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (t *Tree[T]) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("null"), nil
	}
	var chList []*Tree[T]
	if t.children != nil {
		for c := range t.children.Seq() {
			chList = append(chList, c)
		}
	}
	return json.Marshal(treeJSON[T]{Value: t.value, Children: chList})
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Tree[T]) UnmarshalJSON(data []byte) error {
	var tj treeJSON[T]
	if err := json.Unmarshal(data, &tj); err != nil {
		return err
	}
	var cl *List[*Tree[T]] = Nil[*Tree[T]]()
	for i := len(tj.Children) - 1; i >= 0; i-- {
		cl = Cons(tj.Children[i], cl)
	}
	t.value = tj.Value
	t.children = cl
	return nil
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
