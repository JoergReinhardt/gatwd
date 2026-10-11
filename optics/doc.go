// Package optics provides pure functional Optics (Lenses and Prisms) for inspecting and mutating immutable data structures.
//
// In a purely functional system, records and AST nodes cannot be mutated in-place.
// Optics encapsulate bidirectional data access without manual boilerplate:
//   - Lens[S, A]: Functional getter and setter focusing on an invariant field A within S.
//   - Prism[S, A]: Functional optic focusing on a variant branch of an algebraic sum type.
//   - RecordLens: Ergonomic lens focusing on fields within dynamic Record (map[string]any) maps.
//   - MapLens: Strongly typed lens focusing on keys within generic Go maps.
//   - Composition: ComposeLenses and ComposePrisms allow deep optic chaining.
//
// Optics form the mathematical backbone of the DSL's 'focus' and 'update' transducers.
package optics
