// Package comonads provides Comonadic data structures and spatial tree navigation.
//
// While monads model context injection and sequencing (value inside context), comonads
// model context extraction and spatial awareness (context focused on a current element).
//
// The central structures are:
//
//   - Rose Tree Zipper (Zipper): Focuses on a single node within a multi-way tree, keeping track
//     of its ancestors, left siblings, and right siblings. Operations Left, Right, Down, and Up
//     re-focus the zipper on neighboring sub-branches in O(1) time without rebuilding the tree.
//
//   - 2D Toroidal Grid Comonad (Grid, GridZipper): Models discrete 2-dimensional toroidal surfaces.
//     Satisfies the comonad laws via Extract (W[A] -> A), Duplicate (W[A] -> W[W[A]]), and
//     Extend/Cobind ((W[A] -> B) -> W[A] -> W[B]). Enables simultaneous neighborhood transitions,
//     spatial stencils (Moore, Von Neumann), and 2D image/grid convolution filters.
//
//   - Graph Cursor Comonad (GraphCursor): Focuses a single node of a cyclic
//     collections.Graph together with its in/out neighbourhood. Extract reads the
//     focused node; ExtendGraph/DuplicateGraph recompute every node's payload from a
//     cursor focused on it ((W[A] -> B) -> W[A] -> W[B]). This is the comonadic way
//     to express neighbourhood propagation over an infrastructure graph — taint
//     spread, trust/score diffusion, compliance fixpoints — without a hand-rolled
//     traversal.
package comonads


