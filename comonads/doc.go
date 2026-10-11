// Package comonads provides Comonadic data structures and spatial tree navigation.
//
// While monads model context injection and sequencing (value inside context), comonads
// model context extraction and spatial awareness (context focused on a current element).
//
// The central structure is the Rose Tree Zipper:
//
//   - Zipper: Focuses on a single node within a multi-way tree, keeping track of its ancestors,
//     left siblings, and right siblings.
//
//   - Spatial Navigation: Operations Left, Right, Down, and Up re-focus the zipper on neighboring
//     sub-branches in O(1) time without rebuilding the tree.
//
//   - Extract: The comonadic extract operation pulls the value currently in focus.
package comonads
