# Gophers All The Way Down (gatwd) - Architecture

The `gatwd` repository is an experimental library that attempts to bring powerful functional programming concepts to Go. It predates Go generics (introduced in Go 1.18) and uses high-level workarounds based on closures, interface composition, and bitflag-based type checking to achieve features like Algebraic Data Types (ADTs), laziness, currying, and Hindley-Milner-esque type systems.

## 1. The Core Paradigm: Host Engine vs. Guest DSL

To understand the architecture, one must strictly separate the **Host Language Mechanics** (how the Go engine runs) from the **Guest DSL Runtime** (how the interpreted language behaves).

### The Host Engine (Modern Go)
The internal plumbing of the interpreter—memory management, AST node structures, lazy evaluation loops, and functional primitives (`Map`, `Fold`, `Bind`)—is powered by modern Go 1.18+ Generics and Go 1.23+ Iterators (`iter.Seq`). This allows the host Go code to run blazingly fast and safely without relying on massive interface boxing or reflection. The use of `iter.Seq` acts as a native "trampoline," allowing mathematically infinite sequences and recursion without blowing up the Go call stack.

### The Guest DSL Runtime (BitFlag Registry)
Because Go is statically typed, its Generics cannot evaluate dynamically parsed user types at runtime. To support an ML-like DSL where users can define new "quasi types" on the fly, the system maintains a custom **BitFlag-based runtime type registry**.

- **BitFlag Identifiers**: Instead of using Go's `reflect`, the DSL tags every AST node with a unique bitflag (`1 << iota`).
- **Dynamic Algebraic Types**: By combining these bitflags (using bitwise OR `A | B` for Sum types, or `TyComp` structs for Product types), the engine can parse a DSL definition like `type MyUnion = A | B` and immediately enforce it during evaluation.
- **Pattern Matching**: The interpreter enforces dynamic typing simply by applying fast bitwise arithmetic (`FlagMatch`) on the AST nodes as they stream through the evaluator.

## 2. Package Structure

### `data` Package
Provides type-safe wrappers around all native Go primitives. 
- Implements `Native` interface for numbers, texts, binaries.
- Defines BitFlag aliases (`TyNat`) to represent these types efficiently.
- Groups types into "Type Classes" (e.g., `Integers = Int | Int8 | Int16 | ...`) using bitwise OR operations.
- Implements generic collections (Sets, Maps, Slices) operating on `Native` interfaces.

### `functions` Package
Builds the functional programming abstractions on top of the `data` package.
- **Data Constructors**: Functions to lift `Native` values into the functional `Expression` context.
- **Type Constructors**: Higher-order functions used to construct complex types (Tuples, Records, Vectors, Monads).
- **Control Flow**: Implements functional control flow like `Switch`, `Case`, `Maybe`, `Option`, `Either`.
- **Continuations & Monads**: `Continuation`, `Group`, `Monadic`, `Functorial` interfaces providing `Map`, `Fold`, `Bind`, `Apply` and infinite list processing through lazy evaluation.

### `prototypes` Package
Contains unfinished experiments (e.g., recursive XML parsing).

## 3. Key Concepts

### Bitflag Type Checking
Instead of checking types at runtime via `v.(type)` or reflection, every type has a unique bit flag (`1 << iota`). Composite types (Sum types / Unions) are represented by ORing the flags (`A | B`). Checking if a value is of a certain type is a fast bitwise operation (`value.Flag() & target_flag != 0`).

### Algebraic Data Types (ADTs)
The `functions` package provides robust support for ADTs:
- **Product Types**: Tuples, Records. Combines multiple types into one.
- **Sum Types**: Enums, Sets. Represents a value that can be one of several different types.
- **Options**: `Maybe` (`Just` / `None`) and `Option` (`Either` / `Or`) provide null-safety and error handling functionally.

### Laziness & Continuations
Lists and sequences are implemented as `Continuation`s. Instead of holding all elements in memory, they return a `Head()` (the current element) and a `Tail()` (a closure representing the rest of the computation), allowing for lazy execution and infinite sequences.

### Pattern Matching
Functions can define structural type bounds and use pattern matching to execute different cases based on the `BitFlag` signature of the arguments passed, allowing polymorphic behavior without empty interface checking.
