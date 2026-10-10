# Core Components Reference

This document breaks down the major types and functionalities provided by the `data` and `functions` packages.

## `data` Package

The `data` package implements native Go types wrapped in interfaces to allow them to participate in the BitFlag-based type system.

- **`BitFlag`**: An `uint` alias providing bitwise operations (`FlagLow`, `FlagHigh`, `FlagToggle`, `FlagMask`, `FlagMatch`). It acts as a lightweight, reflection-free type identifier.
- **Native Interfaces**:
  - `Numeral`, `Natural`, `Integer`, `Real`, `Rational`, `Imaginary`: Arithmetic bounds.
  - `Boolean`: Truth values.
  - `Text`, `Letter`, `BinaryMarshaler`: Textual and binary data bounds.
- **Collections**:
  - `Sliceable`, `Mutable`, `Mapped`: Interfaces for handling slices and maps in a type-safe way over `Native` types.
  - `SetVal`: A generic Set implementation that relies on BitFlag matches and Native identities.
- **Conversion Methods**: Extensive type conversion methods for translating between `Native` aliases and raw Go types (e.g., `IntVal` to `int`).

## `functions` Package

The `functions` package provides the actual functional programming toolkit. It builds upon `data` to provide dynamic and parametric function definitions.

- **`Expression`**: The central interface for functional types. Everything evaluates to an `Expression`.
- **Higher-Order Functions**:
  - `Curry(f, g)`: Takes two functions and returns a composited function.
  - `Map(cont, fn)`: Applies `fn` to every element in the continuation.
  - `Fold(acc, fn)`: Reduces a continuation into a single value using `fn`.
  - `Filter(cont, test)` / `Pass(cont, test)`: Filters a continuation based on a predicate.
  - `Zip(l, r, fn)`: Zips two continuations using a combining function.
  - `Bind(expr, func)`: Monadic bind operation.
- **Algebraic Data Types (ADTs)**:
  - `Tuple`: Fixed-size heterogeneous collection.
  - `Record`: Key-value pair collection with strong type boundaries.
  - `Enum`: Value constrained to a finite set of named states.
  - `Maybe` (`Just` / `None`): Represents a value that might be absent.
  - `Option` (`Either` / `Or`): Represents a value that can be one of two types.
- **Control Structures**:
  - `Switch` & `Case`: Functional conditionals evaluated against pattern matching predicates.
- **Continuations (Lazy Evaluation)**:
  - `Sequence`, `Vector`, `List`: Abstracted over the `Continuation` interface (`Head()` and `Tail()`). They evaluate elements only when pulled, allowing for infinite datasets (Generators, Accumulators).

## `prototypes` Package

A sandbox for experiments. Currently contains:
- `xml.go`: An experimental attempt to recursively unmarshal XML and map it into functional nodes (`XMLNode` and `walkXML`).

