package cli

import (
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// Two BSTs are equal by their in-order content, never by tree shape; a BST is not a List.
func TestBSTEqualityIsContentNotShape(t *testing.T) {
	in := NewInterpreter()
	a, ok := evalOne(t, in, `bst_from [2, 1, 3]`).(*collections.BST[any])
	b, ok2 := evalOne(t, in, `bst_from [3, 2, 1]`).(*collections.BST[any])
	if !ok || !ok2 {
		t.Fatal("bst_from must yield a *BST[any]")
	}
	if !functions.StructuralEqual(a, b) {
		t.Error("equal content, different shape: must be equal")
	}
	c := evalOne(t, in, `bst_from [1, 2, 4]`)
	if functions.StructuralEqual(a, c) {
		t.Error("different content must not be equal")
	}
	if functions.StructuralEqual(a, evalOne(t, in, `[1, 2, 3]`)) {
		t.Error("a BST is not a List")
	}
}
