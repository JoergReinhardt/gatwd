package collections

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/functions"
)

// InsertOrdered inserts the value v into the search tree b, ordering every pair of elements by
// functions.Compare, and returns the new tree (b is never changed). It is the dynamically typed
// face of BST.Insert: a *BST[any] holds the elements themselves, never their printed text, so
// 10 sorts after 9 and a list sorts as a list.
//
// Compare is a total order over numbers (an Int equals the Float of the same value, so 1 and
// 1.0 are one member), over Strings (as text), over Times and Durations, and over Lists and
// Pairs lexicographically. Two elements it cannot order (a String and an Int, a function)
// make InsertOrdered return an error that names both elements; the tree returned with the error
// is b, unchanged. A nil b is the empty tree.
func InsertOrdered(b *BST[any], v any) (*BST[any], error) {
	var err error
	out := b.Insert(v, orderedCmp(&err))
	if err != nil {
		return b, err
	}
	return out, nil
}

// ContainsOrdered reports whether v is a member of b under the same order InsertOrdered used.
// An element that cannot be ordered against one on the search path is an error naming both,
// never a quiet false. A nil b contains nothing.
func ContainsOrdered(b *BST[any], v any) (bool, error) {
	var err error
	ok := b.Contains(v, orderedCmp(&err))
	if err != nil {
		return false, err
	}
	return ok, nil
}

// orderedCmp is functions.Compare as the comparator a BST wants: a comparison that fails
// records its error (the first one, naming both elements) in *errp and reports "equal",
// which ends the descent.
func orderedCmp(errp *error) func(a, b any) int {
	return func(a, b any) int {
		if *errp != nil {
			return 0
		}
		c, err := functions.Compare(a, b)
		if err != nil {
			*errp = fmt.Errorf("cannot order %s and %s in a BST: %w", showOrdered(a), showOrdered(b), err)
			return 0
		}
		return c
	}
}

// showOrdered prints an element for an error message: a String quoted, anything else as is.
func showOrdered(v any) string {
	if s, ok := v.(string); ok {
		return fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("%v", v)
}
