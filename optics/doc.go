// Package optics provides pure functional Optics (Lenses and Prisms) for inspecting and mutating immutable data structures.
//
// In a purely functional system, records and AST nodes cannot be mutated in-place.
// Optics encapsulate bidirectional data access without manual boilerplate:
//   - Lens[S, A]: Functional getter and setter focusing on an invariant field A within S.
//   - Prism[S, A]: Functional optic focusing on a variant branch of an algebraic sum type.
//   - Traversal[S, A]: Optic focusing on zero-or-more targets at once (GetAll/Modify),
//     the generalisation of Lens (exactly one) and Prism (at most one).
//   - RecordLens: Ergonomic lens focusing on fields within dynamic Record (map[string]any) maps.
//   - MapLens: Strongly typed lens focusing on keys within generic Go maps.
//   - Graph optics (NodePayloadLens, ColorNodePrism, AllEdgeWeights, OutEdgeWeights):
//     focused immutable rewrites of a cyclic collections.Graph — edit one node's
//     payload, match a node of a given color, or bulk-adjust edge weights — each
//     returning a new graph with the HAMT adjacency's structural sharing preserved.
//   - Composition: ComposeLenses, ComposePrisms, and ComposeTraversals allow deep optic chaining.
//
// Optics form the mathematical backbone of the DSL's 'focus' and 'update' transducers.
package optics
