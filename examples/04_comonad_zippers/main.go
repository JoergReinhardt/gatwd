package main

import (
	"fmt"
	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/comonads"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func main() {
	opts := visualizer.ParseOptions("--curved --border --border-color=yellow --alternate-bg=blue,magenta")

	headerOpts := visualizer.ParseOptions("--heavy --border --border-color=cyan --color=white")
	fmt.Println(visualizer.Format("Operator: Comonadic Zippers", headerOpts))

	// 1. Build an immutable Rose Tree representing an AST
	child3 := collections.Leaf("Pod-3")
	child2 := collections.Leaf("Pod-2")
	child1 := collections.Leaf("Pod-1")

	// Create a list of children
	children := collections.Cons(child1, collections.Cons(child2, collections.Cons(child3, collections.Nil[*collections.Tree[string]]())))
	root := collections.NewTree("Kubernetes-Cluster", children)

	fmt.Println(visualizer.Format(map[string]any{
		"RoseTree": root,
		"Children": []string{"Pod-1", "Pod-2", "Pod-3"},
	}, opts))

	// 2. Wrap it in a Zipper for O(1) Cursor Navigation
	zipper := comonads.NewZipper(root)
	fmt.Println(visualizer.Format(map[string]string{
		"Step":   "Initial Cursor",
		"Focus":  zipper.Extract(),
		"Status": "At Root",
	}, opts))

	// 3. Navigate the AST using purely functional steps
	zipper = zipper.Down()
	fmt.Println(visualizer.Format(map[string]string{
		"Step":   "Moved Down",
		"Focus":  zipper.Extract(),
		"Status": "Navigated to Child",
	}, opts))

	zipper = zipper.Up()
	fmt.Println(visualizer.Format(map[string]string{
		"Step":   "Moved Up",
		"Focus":  zipper.Extract(),
		"Status": "Restored to Root",
	}, opts))

	fmt.Println("Notice how we navigated the immutable AST purely mechanically!")
}
