package collections

type Color bool
const (
	Red   Color = true
	Black Color = false
)

type RBTree[T any] struct {
	color Color
	value T
	left  *RBTree[T]
	right *RBTree[T]
}

func LeafRB[T any](v T) *RBTree[T] {
	return &RBTree[T]{color: Red, value: v} // new nodes are Red
}

func balance[T any](c Color, v T, l, r *RBTree[T]) *RBTree[T] {
	if c == Black {
		if l != nil && l.color == Red {
			if l.left != nil && l.left.color == Red {
				return &RBTree[T]{color: Red, value: l.value,
					left:  &RBTree[T]{color: Black, value: l.left.value, left: l.left.left, right: l.left.right},
					right: &RBTree[T]{color: Black, value: v, left: l.right, right: r}}
			}
			if l.right != nil && l.right.color == Red {
				return &RBTree[T]{color: Red, value: l.right.value,
					left:  &RBTree[T]{color: Black, value: l.value, left: l.left, right: l.right.left},
					right: &RBTree[T]{color: Black, value: v, left: l.right.right, right: r}}
			}
		}
		if r != nil && r.color == Red {
			if r.left != nil && r.left.color == Red {
				return &RBTree[T]{color: Red, value: r.left.value,
					left:  &RBTree[T]{color: Black, value: v, left: l, right: r.left.left},
					right: &RBTree[T]{color: Black, value: r.value, left: r.left.right, right: r.right}}
			}
			if r.right != nil && r.right.color == Red {
				return &RBTree[T]{color: Red, value: r.value,
					left:  &RBTree[T]{color: Black, value: v, left: l, right: r.left},
					right: &RBTree[T]{color: Black, value: r.right.value, left: r.right.left, right: r.right.right}}
			}
		}
	}
	return &RBTree[T]{color: c, value: v, left: l, right: r}
}

func (t *RBTree[T]) insert(v T, cmp func(a, b T) int) *RBTree[T] {
	if t == nil { return LeafRB(v) }
	
	if cmp(v, t.value) < 0 {
		return balance(t.color, t.value, t.left.insert(v, cmp), t.right)
	} else if cmp(v, t.value) > 0 {
		return balance(t.color, t.value, t.left, t.right.insert(v, cmp))
	}
	return t
}

func (t *RBTree[T]) Insert(v T, cmp func(a, b T) int) *RBTree[T] {
	root := t.insert(v, cmp)
	root.color = Black // Root is always Black
	return root
}

func (t *RBTree[T]) Graph() any {
	if t == nil { return "<nil RBTree>" }
	c := "white"
	if t.color == Red { c = "red" }
	return RBNodeGraph[T]{ Value: t.value, Left: t.left, Right: t.right, NodeColor: c }
}

type RBNodeGraph[T any] struct {
	Value     T
	Left      *RBTree[T]
	Right     *RBTree[T]
	NodeColor string
}
