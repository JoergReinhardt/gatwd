package comonads

import "github.com/joergreinhardt/gatwd/collections"

// TreeContext stores the breadcrumbs (the path back to the root) 
// for the Zipper, allowing it to move Up, Left, and Right.
type TreeContext[T any] struct {
	Parent *TreeZipper[T]
	// Using slices for sibling storage for simplicity of O(1) random access,
	// though purely functional versions might use Lists.
	LeftSiblings  []*collections.Tree[T]
	RightSiblings []*collections.Tree[T]
}

// TreeZipper is a Comonad acting as a cursor over a Tree.
//
// It holds the currently focused Node (Extract) and the Context 
// required to mathematically reconstruct the rest of the Tree.
// This allows O(1) local modification and navigation in an immutable structure.
type TreeZipper[T any] struct {
	Focus   *collections.Tree[T]
	Context *TreeContext[T]
}

// NewZipper creates a new cursor starting at the root of a tree.
func NewZipper[T any](root *collections.Tree[T]) *TreeZipper[T] {
	return &TreeZipper[T]{
		Focus:   root,
		Context: nil, // Root has no parent or siblings
	}
}

// Extract fulfills the Comonad interface (W[A] -> A).
// It returns the value currently focused by the Zipper.
func (z *TreeZipper[T]) Extract() T {
	return z.Focus.Value()
}

// Down moves the cursor to the first child of the current focus.
// Returns nil if there are no children.
func (z *TreeZipper[T]) Down() *TreeZipper[T] {
	children := z.Focus.Children()
	if children.IsEmpty() {
		return nil
	}

	head, _ := children.Head()
	tail, _ := children.Tail()
	
	// Collect right siblings
	var rights []*collections.Tree[T]
	for curr := tail; !curr.IsEmpty(); {
		h, _ := curr.Head()
		rights = append(rights, h)
		curr, _ = curr.Tail()
	}

	return &TreeZipper[T]{
		Focus: head,
		Context: &TreeContext[T]{
			Parent:        z,
			LeftSiblings:  []*collections.Tree[T]{},
			RightSiblings: rights,
		},
	}
}

// Up moves the cursor to the parent node.
// Returns nil if already at the root.
func (z *TreeZipper[T]) Up() *TreeZipper[T] {
	if z.Context == nil || z.Context.Parent == nil {
		return nil
	}

	// Reconstruct the children list of the parent
	var newChildren *collections.List[*collections.Tree[T]] = collections.Nil[*collections.Tree[T]]()
	
	// Prepend right siblings (iterating backwards to preserve order in Cons)
	for i := len(z.Context.RightSiblings) - 1; i >= 0; i-- {
		newChildren = collections.Cons(z.Context.RightSiblings[i], newChildren)
	}
	
	// Prepend the focus
	newChildren = collections.Cons(z.Focus, newChildren)
	
	// Prepend left siblings
	for i := len(z.Context.LeftSiblings) - 1; i >= 0; i-- {
		newChildren = collections.Cons(z.Context.LeftSiblings[i], newChildren)
	}

	// Create a new parent node with the reconstructed children
	newParentNode := collections.NewTree(z.Context.Parent.Focus.Value(), newChildren)

	return &TreeZipper[T]{
		Focus:   newParentNode,
		Context: z.Context.Parent.Context, // Inherit the parent's context
	}
}
