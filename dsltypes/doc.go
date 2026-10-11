// Package dsltypes provides the dynamic type system and AST node bridge between the host Go engine and the guest DSL.
//
// Because Go generics are resolved strictly at compile time, evaluating dynamic user-defined
// types (e.g. type OS = Linux | Darwin) requires a runtime type representation.
//
// Core Concepts:
//
//   - BitFlag: A 64-bit unsigned integer acting as a lightweight, reflection-free type identifier.
//     Sum types are formed by bitwise OR; pattern matching is evaluated via fast bitmask operations.
//
//   - TyComp: Represents composed Product Types (Tuples, Records) and function signatures.
//
//   - ASTNode[T]: The boundary container pairing a raw Go value (generically typed for the host)
//     with its BitFlag quasi-type tag. Implements fmt.Stringer for clean output rendering.
package dsltypes
