package collections_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
)

// A search tree of values is ordered by comparison, not by printed text: 10 comes after 9.
// Elements that cannot be ordered against each other are an error naming both.
func ExampleInsertOrdered() {
	var tree *collections.BST[any]
	for _, v := range []any{3, 1, 10, 9, 2} {
		tree, _ = collections.InsertOrdered(tree, v)
	}
	var in []any
	for v := range tree.Seq() {
		in = append(in, v)
	}
	fmt.Println(in)

	_, err := collections.InsertOrdered(tree, "ten")
	fmt.Println(err != nil)
	ok, _ := collections.ContainsOrdered(tree, 10)
	fmt.Println(ok)
	// Output:
	// [1 2 3 9 10]
	// true
	// true
}
