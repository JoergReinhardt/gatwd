package collections

import (
	"cmp"
	"slices"
	"testing"
)

// TestBSTContainsSizeSeq checks the search, cardinality and in-order fold of the
// persistent BST, including the nil (empty) tree and duplicate insertion.
func TestBSTContainsSizeSeq(t *testing.T) {
	var tree *BST[int]
	if tree.Size() != 0 || tree.Contains(1, cmp.Compare[int]) {
		t.Fatal("nil tree must be empty")
	}
	for _, v := range []int{5, 2, 8, 1, 9, 5, 3} {
		tree = tree.Insert(v, cmp.Compare[int])
	}
	if tree.Size() != 6 {
		t.Errorf("size %d, want 6 (duplicate 5 ignored)", tree.Size())
	}
	got := slices.Collect(tree.Seq())
	if !slices.Equal(got, []int{1, 2, 3, 5, 8, 9}) {
		t.Errorf("in-order %v", got)
	}
	for _, v := range got {
		if !tree.Contains(v, cmp.Compare[int]) {
			t.Errorf("missing %d", v)
		}
	}
	if tree.Contains(4, cmp.Compare[int]) {
		t.Error("4 was never inserted")
	}
	// persistence: inserting does not mutate the receiver
	before := tree.Size()
	_ = tree.Insert(100, cmp.Compare[int])
	if tree.Size() != before {
		t.Error("Insert mutated the original tree")
	}
}
