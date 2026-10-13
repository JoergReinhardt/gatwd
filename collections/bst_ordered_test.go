package collections

import (
	"math/big"
	"slices"
	"strings"
	"testing"
)

func bstOf(t *testing.T, items ...any) *BST[any] {
	t.Helper()
	var tree *BST[any]
	for _, it := range items {
		var err error
		if tree, err = InsertOrdered(tree, it); err != nil {
			t.Fatalf("InsertOrdered(%v): %v", it, err)
		}
	}
	return tree
}

// A BST[any] holds the elements themselves, ordered by value: the Ints 10 and 9 are in numeric
// order (as text "10" < "9"), the raw element types survive, and equal values (1 and 1.0) are one member.
func TestInsertOrderedKeepsValues(t *testing.T) {
	got := slices.Collect(bstOf(t, 3, 1, 2, 10, 9).Seq())
	if !slices.Equal(got, []any{1, 2, 3, 9, 10}) {
		t.Errorf("in order: %#v, want the Ints 1 2 3 9 10", got)
	}
	for _, it := range got {
		if _, ok := it.(int); !ok {
			t.Errorf("element %#v was turned into %T", it, it)
		}
	}
	if got := slices.Collect(bstOf(t, "b", "a", "c", "10", "9").Seq()); !slices.Equal(got, []any{"10", "9", "a", "b", "c"}) {
		t.Errorf("Strings order as text: %#v", got)
	}
	if got := slices.Collect(bstOf(t, 2.5, 1, 2).Seq()); !slices.Equal(got, []any{1, 2, 2.5}) {
		t.Errorf("an Int and a Float order by value: %#v", got)
	}
	if got := slices.Collect(bstOf(t, big.NewInt(3), 2, 10).Seq()); len(got) != 3 || got[0] != 2 {
		t.Errorf("mixed Int and big integer: %#v", got)
	}
	if n := bstOf(t, 1, 1.0, 1).Size(); n != 1 {
		t.Errorf("1 and 1.0 are one member, size %d", n)
	}
	// persistence: the receiver is unchanged
	base := bstOf(t, 1, 2)
	if _, err := InsertOrdered(base, 3); err != nil || base.Size() != 2 {
		t.Errorf("InsertOrdered mutated its receiver (size %d, err %v)", base.Size(), err)
	}
	if nilTree, err := InsertOrdered(nil, "x"); err != nil || nilTree.Size() != 1 {
		t.Errorf("the nil tree is the empty tree: %v %v", nilTree, err)
	}
}

// Elements that cannot be ordered are a loud error naming both of them, and the tree is unchanged.
func TestInsertOrderedIncomparableIsAnError(t *testing.T) {
	base := bstOf(t, 1, 2)
	got, err := InsertOrdered(base, "a")
	if err == nil || !strings.Contains(err.Error(), `"a"`) || !strings.Contains(err.Error(), "BST") {
		t.Fatalf("err = %v, want one naming the String element", err)
	}
	if got != base {
		t.Error("the tree returned with an error must be the unchanged receiver")
	}
	if _, err := ContainsOrdered(base, "a"); err == nil || !strings.Contains(err.Error(), `"a"`) {
		t.Errorf("ContainsOrdered with a String against Ints: err = %v", err)
	}
	if ok, err := ContainsOrdered(base, 2); !ok || err != nil {
		t.Errorf("ContainsOrdered 2 = %v, %v", ok, err)
	}
	if ok, err := ContainsOrdered(base, 3); ok || err != nil {
		t.Errorf("ContainsOrdered 3 = %v, %v, want false", ok, err)
	}
	if ok, err := ContainsOrdered(nil, 1); ok || err != nil {
		t.Errorf("the empty tree contains nothing: %v %v", ok, err)
	}
}
