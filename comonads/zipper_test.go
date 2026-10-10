package comonads_test

import (
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/comonads"
)

func TestZipper(t *testing.T) {
	// Build Tree: Root(1) -> Child(2) -> Grandchild(3)
	gc := collections.Leaf(3)
	child := collections.NewTree(2, collections.Cons(gc, collections.Nil[*collections.Tree[int]]()))
	root := collections.NewTree(1, collections.Cons(child, collections.Nil[*collections.Tree[int]]()))

	zipper := comonads.NewZipper(root)
	if zipper.Extract() != 1 {
		t.Fatalf("Expected root value 1")
	}

	// Move down to 2
	z2 := zipper.Down()
	if z2 == nil || z2.Extract() != 2 {
		t.Fatalf("Expected to move down to 2")
	}

	// Move down to 3
	z3 := z2.Down()
	if z3 == nil || z3.Extract() != 3 {
		t.Fatalf("Expected to move down to 3")
	}

	// Move back up to 2
	zUp2 := z3.Up()
	if zUp2 == nil || zUp2.Extract() != 2 {
		t.Fatalf("Expected to move up to 2, got %d", zUp2.Extract())
	}
}
