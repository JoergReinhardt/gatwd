package main

import (
	"fmt"
	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/comonads"
)

func main() {
	fmt.Println("=== Operator: Comonadic Zippers ===")

	// 1. Build an immutable Rose Tree representing an AST
	child3 := collections.Leaf("Pod-3")
	child2 := collections.Leaf("Pod-2")
	child1 := collections.Leaf("Pod-1")
	
	// Create a list of children
	children := collections.Cons(child1, collections.Cons(child2, collections.Cons(child3, collections.Nil[*collections.Tree[string]]())))
	root := collections.NewTree("Kubernetes-Cluster", children)

	// 2. Wrap it in a Zipper for O(1) Cursor Navigation
	zipper := comonads.NewZipper(root)
	fmt.Printf("Initial Cursor: %s\n", zipper.Extract())

	// 3. Navigate the AST using purely functional steps
	zipper = zipper.Down()
	fmt.Printf("Moved Down -> Cursor: %s\n", zipper.Extract())

	// Right is not implemented in our Zipper context (we didn't finish left/right in Phase 8).
	// We will just demonstrate Up!
	zipper = zipper.Up()
	fmt.Printf("Moved Up -> Cursor: %s\n", zipper.Extract())
	
	fmt.Println("Notice how we navigated the immutable AST purely mechanically!")
}
