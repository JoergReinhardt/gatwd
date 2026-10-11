package comonads

import (
	"testing"
	"github.com/joergreinhardt/gatwd/collections"
)

func TestTreeZipper_LeftRight(t *testing.T) {
	// Build a simple tree: Root -> [Child1, Child2, Child3]
	child1 := collections.Leaf("Child1")
	child2 := collections.Leaf("Child2")
	child3 := collections.Leaf("Child3")
	
	// Construct List: Cons(Child1, Cons(Child2, Cons(Child3, Nil)))
	l3 := collections.Cons(child3, collections.Nil[*collections.Tree[string]]())
	l2 := collections.Cons(child2, l3)
	l1 := collections.Cons(child1, l2)
	
	root := collections.NewTree("Root", l1)

	// Start at root
	z := NewZipper(root)
	if z.Extract() != "Root" {
		t.Fatalf("Expected 'Root', got %v", z.Extract())
	}

	// Move down to first child (Child1)
	zDown := z.Down()
	if zDown == nil || zDown.Extract() != "Child1" {
		t.Fatalf("Expected 'Child1', got %v", zDown.Extract())
	}

	// Move right to sibling (Child2)
	zRight1 := zDown.Right()
	if zRight1 == nil || zRight1.Extract() != "Child2" {
		t.Fatalf("Expected 'Child2', got %v", zRight1.Extract())
	}

	// Move right again (Child3)
	zRight2 := zRight1.Right()
	if zRight2 == nil || zRight2.Extract() != "Child3" {
		t.Fatalf("Expected 'Child3', got %v", zRight2.Extract())
	}
	
	// Move right again (Should be nil, end of siblings)
	zRight3 := zRight2.Right()
	if zRight3 != nil {
		t.Fatalf("Expected nil, got %v", zRight3.Extract())
	}

	// Move left back to (Child2)
	zLeft1 := zRight2.Left()
	if zLeft1 == nil || zLeft1.Extract() != "Child2" {
		t.Fatalf("Expected 'Child2', got %v", zLeft1.Extract())
	}

	// Move left back to (Child1)
	zLeft2 := zLeft1.Left()
	if zLeft2 == nil || zLeft2.Extract() != "Child1" {
		t.Fatalf("Expected 'Child1', got %v", zLeft2.Extract())
	}
	
	// Move left again (Should be nil, start of siblings)
	zLeft3 := zLeft2.Left()
	if zLeft3 != nil {
		t.Fatalf("Expected nil, got %v", zLeft3.Extract())
	}

	// Move Up from Child2 should yield Root again, with structure preserved
	zUp := zRight1.Up()
	if zUp == nil || zUp.Extract() != "Root" {
		t.Fatalf("Expected 'Root', got %v", zUp.Extract())
	}
}
