package collections_test

import (
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
)

func TestList(t *testing.T) {
	// Build a list: [1, 2, 3] -> Cons(1, Cons(2, Cons(3, Nil)))
	list := collections.Cons(1, collections.Cons(2, collections.Cons(3, collections.Nil[int]())))

	if list.IsEmpty() {
		t.Fatal("Expected list to not be empty")
	}

	head, ok := list.Head()
	if !ok || head != 1 {
		t.Fatalf("Expected head 1, got %d", head)
	}

	// Test Iterator
	sum := 0
	for val := range list.Seq() {
		sum += val
	}

	if sum != 6 {
		t.Fatalf("Expected sum 6, got %d", sum)
	}
}

func TestTree(t *testing.T) {
	// Build a Tree:
	//       1
	//     /   \
	//    2     3
	//   / \
	//  4   5

	leaf4 := collections.Leaf(4)
	leaf5 := collections.Leaf(5)

	childrenOf2 := collections.Cons(leaf4, collections.Cons(leaf5, collections.Nil[*collections.Tree[int]]()))
	node2 := collections.NewTree(2, childrenOf2)

	node3 := collections.Leaf(3)

	childrenOf1 := collections.Cons(node2, collections.Cons(node3, collections.Nil[*collections.Tree[int]]()))
	root := collections.NewTree(1, childrenOf1)

	// Depth-first traversal should yield: 1, 2, 4, 5, 3
	expected := []int{1, 2, 4, 5, 3}
	var result []int

	for val := range root.WalkDepthFirst() {
		result = append(result, val)
	}

	if len(result) != len(expected) {
		t.Fatalf("Expected %d elements, got %d", len(expected), len(result))
	}

	for i, v := range expected {
		if result[i] != v {
			t.Errorf("At index %d expected %d, got %d", i, v, result[i])
		}
	}
}
