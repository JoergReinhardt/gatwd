package collections_test

import (
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
)

func TestBTreeNodeLabelAndChildren(t *testing.T) {
	leaf1 := collections.NewBTreeNode([]int{10, 20}, nil)
	leaf2 := collections.NewBTreeNode([]int{30, 40}, nil)
	root := collections.NewBTreeNode([]int{25}, []*collections.BTree[int]{leaf1, leaf2})

	if got, want := root.NodeLabel(), "[25]"; got != want {
		t.Fatalf("root.NodeLabel() = %v, want %v", got, want)
	}

	children := root.NodeChildren()
	if len(children) != 2 {
		t.Fatalf("root.NodeChildren() = %d children, want 2", len(children))
	}
	gotLeaf1, ok := children[0].(*collections.BTree[int])
	if !ok || gotLeaf1 != leaf1 {
		t.Fatalf("root.NodeChildren()[0] = %v, want the original leaf1 pointer", children[0])
	}
	if got, want := gotLeaf1.NodeLabel(), "[10 20]"; got != want {
		t.Fatalf("leaf1.NodeLabel() = %v, want %v", got, want)
	}

	// A leaf must report no children, not a non-nil empty slice.
	if got := leaf1.NodeChildren(); got != nil {
		t.Fatalf("leaf1.NodeChildren() = %v, want nil for a leaf", got)
	}
}

func TestBTreeNodeLabelNilReceiver(t *testing.T) {
	var b *collections.BTree[int]
	if got, want := b.NodeLabel(), "<nil BTree>"; got != want {
		t.Fatalf("nil.NodeLabel() = %v, want %v", got, want)
	}
	if got := b.NodeChildren(); got != nil {
		t.Fatalf("nil.NodeChildren() = %v, want nil", got)
	}
}
