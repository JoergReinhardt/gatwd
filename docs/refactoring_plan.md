# Refactoring Plan: Modernizing `gatwd`

This document maps out a concrete refactoring plan to fulfill the original vision of the project—building an ML-like DSL interpreter with dynamic "quasi types" and infinite loops—by marrying the original `BitFlag` runtime registry with modern Go Generics and Iterators.

## The Core Paradigm: Host Engine vs. Guest DSL

The refactoring revolves around strictly separating the **Host Language Mechanics** from the **Guest DSL Runtime**:
1. **The Host Engine (Go):** Uses Go 1.18 Generics and Go 1.23 Iterators (`iter.Seq`) for memory safety, generic AST nodes, and high-performance trampolining (loops). It eliminates the need to box every single variable into an `Expression` interface.
2. **The Guest DSL Runtime:** Preserves the `BitFlag` type system. Go Generics cannot evaluate dynamically parsed user types at runtime, so the BitFlag system acts as the runtime type registry for the interpreted language, allowing Algebraic Data Types (Sum/Product types) to be evaluated on the fly.

---

## Phase 1: Modernizing the Host Mechanics (The Trampoline)

The first step is to replace the custom continuation logic (`tail.Continue()`) with modern Go Iterators.

1. **Replace `Continuation` and `Group`:**
   - Deprecate the complex interface hierarchies in `functions/interfaces.go`.
   - Map infinite lazy sequences directly to `iter.Seq[T]`. This acts as the modern "trampoline," keeping recursive logic on the heap/loop rather than blowing up the call stack.
2. **Generic Functional Primitives:**
   - Rewrite `Map`, `Fold`, `Filter`, and `Bind` (`FlatMap`) as strongly-typed generic functions that operate on `iter.Seq[T]`.
   - *Example:* `func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U]`

## Phase 2: Isolating the DSL Runtime Type Registry

The `BitFlag` system is a brilliant mathematical representation of an Algebraic Type System. It needs to be preserved but isolated from the Go-native host wrappers.

1. **Extract `BitFlag`:**
   - Move the bitflag arithmetic (`FlagLow`, `FlagHigh`, `FlagToggle`, `FlagMatch`) from the `data` package into a dedicated `dsltypes` package.
2. **Preserve Algebraic Constructors:**
   - Keep the logic that allows combining flags via bitwise OR (`A | B`) to represent **Sum Types** (like `Trinary = Truth | Undecided`).
   - Keep the `TyComp` logic for composing **Product Types** (Tuples, Records).
   - *Crucially:* These will no longer wrap Go integers or strings for Go's sake; they will exist purely as tags assigned to DSL variables.

## Phase 3: Building the Typed AST Engine

With the host engine optimized and the type registry isolated, we build the bridge: the AST Nodes.

1. **Generic AST Nodes:**
   - Define a generic node structure that pairs raw data with the DSL's runtime type.
   - *Example:* 
     ```go
     type ASTNode[T any] struct {
         Value    T
         DSLType  dsltypes.BitFlag
     }
     ```
2. **The Evaluator Loop:**
   - Create an interpreter that evaluates an `iter.Seq[ASTNode[any]]`.
   - The loop will use `FlagMatch(node.DSLType, ExpectedType)` to perform pattern matching and enforce the DSL's dynamic typing, exactly as originally intended.

## Phase 4: Implementing the ML-Like DSL

With the foundation solid, the final phase realizes the roadmap.

1. **The Parser:**
   - Implement the ML-syntax parser (or jsonnet/yaml configuration parser).
2. **Dynamic Quasi-Types:**
   - When the parser encounters a user defining a new type (e.g., `type Config = Server | Client`), the engine dynamically allocates a new composed `BitFlag` in the `dsltypes` registry.
3. **Execution:**
   - The parsed script is converted into a lazy stream of `ASTNode`s and fed into the iterator-based evaluator, running functional side-effects and infinite state machines efficiently without stack overflows.

## Phase 5: Domain Application - Interactive Graph Traversal

Once the engine is stable, implement the killer feature: a functional Infrastructure-as-Code graph crawler capable of on-the-fly adapter registration.

1. **The Adapter Registry (Type Classes)**:
   - Build a global registry mapping `dsltypes.BitFlag`s to Adapters (functions taking an `ASTNode` and returning an `iter.Seq[ASTNode]`).
   - Adapters act as Type Classes. If a node has the `HasSSH` BitFlag, it inherently gains access to the SSH adapter logic.
2. **The IO Monad (Tool Agnosticity)**:
   - Enforce that all Adapters return an `iter.Seq[ASTNode]` (the IO Monad format for the evaluator).
   - This allows an adapter to be a native Go package (like `golang.org/x/crypto/ssh`) or an agnostic CLI wrapper executing `exec.Command("kubectl ...")` and mapping its `stdout` back into `ASTNode`s. The engine does not care which is used.
3. **The Interactive REPL (Pause & Resume)**:
   - When the graph crawler evaluator encounters a node type it lacks an adapter for, it does not throw a fatal error.
   - It yields a "Need Input" pause state to the trampoline.
   - The user is prompted for a CLI command or API call for the new node.
   - The engine generates a new `BitFlag`, registers a generic `exec` Adapter for it based on the user's input, and resumes the `iter.Seq` trampoline to continue exploring the graph.

## Phase 6: Parallel Execution (Functional Concurrency)

Because the architecture strictly isolates side-effects (IO Monad) and uses immutable AST Nodes and BitFlags, parallel execution can be introduced trivially without the standard pitfalls of race conditions or deadlocks.

1. **Host-Level `ParMap` (Parallel Map)**:
   - Implement a generic `ParMap` and `ParFlatMap` function alongside the sequential versions.
   - Under the hood, these functions take an `iter.Seq[T]`, spin up a worker pool of Go goroutines to process the nodes concurrently, and safely yield the results back into a single `iter.Seq[U]` stream.
2. **DSL-Level Syntax**:
   - Expose parallel execution to the DSL user syntax (e.g., distinguishing between a sequential map `|>` and a parallel map `||>`).
3. **Graph Crawler Performance**:
   - This allows the Infrastructure Graph Crawler to query 1,000 servers or parse 10,000 YAML files simultaneously while remaining mathematically pure and thread-safe.
