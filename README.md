# Gophers All The Way Down

![Gphers All The Way Down Logo](https://github.com/JoergReinhardt/gatwd/blob/master/doc/pic/gopher-logo.png) Gophers all the way down is an experimental library embracing functional programming, written in Go.

> [!IMPORTANT]
> **V2 Modernization Complete (Go 1.23)**
> 
> The library has recently undergone a massive architectural overhaul to leverage modern Go Generics (Go 1.18+) and Iterators (Go 1.23+). The heavy empty-interface boxing of the old `data` package has been moved to a `legacy/` directory, while the core mathematical brilliance (the `BitFlag` type registry) has been extracted to power a blazing-fast, purely functional **Infrastructure Graph Crawler DSL**.
>
> 📚 **New Documentation & Guides:**
> - [DSL Guide & Tutorial](docs/dsl_guide.md): Learn the syntax of the language, how the interactive REPL works, and see examples of crawling Kubernetes clusters!
> - [Native Adapters & Extendability](docs/native_adapters.md): Learn the "Closure Pattern" for building high-performance, native Go extensions.
> - [Error Handling & Auditing](docs/error_handling.md): Learn how the EitherT Monad Transformer turns the crawler into a Compliance Engine.
> - [Architecture Overview](docs/architecture.md): Understand the "Host Engine vs Guest DSL" paradigm driving the interpreter.
> - [Refactoring Roadmap](docs/refactoring_plan.md): Read the step-by-step history of how the academic Category Theory concepts (Optics, Comonadic Zippers, Transducers) were mapped into modern Go.
> - [Styleguide & Contribution](STYLEGUIDE.md): The golden rules of writing functional code and educational documentation in this repository.

## Purpose

This is a pre-alpha version and should under no circumstances be used yet! 

The main purpose of this library is educational. I felt to strong urge to get a basic understanding on the principles of functional programming. While the Haskell community goes great length to explain those principles, it seems to be in the nature of things, that once understood, one immediately looses the capability to explain those principles. The same is true it seems, for the principles of higher mathematics. Regarding the latter, I found a way to overcome this 'law of nature' in the past, by implementing algorithms I had trouble understanding in code. Having code compile and tests go green rules out misconception and provides the confidence necessary, not to accidently overestimating the personal level of understanding. Having made this experience a couple of times in the past, I figured, why not write a functional library?

Former self educational projects had no other purpose but self education. There was always some standard library implementing, whatever I was working on in far better and more professional ways and consequently those projects never where much of a use. With this exercise that might be different. While Java has Scale, C# has F# and JavaScript has Closure, I couldn't find any functional programming library written in Go, that was particularly useful, let alone providing an alternative ML like syntax to write the code in.

Should this library reach the necessary stability and set of features, it should be possible to write a functional parser library on top, implementing ML like Syntax. Generating Domain specific languages could be one example of a useful application. Building a decent version of a jsonnet/ksonnet interpreter, able to deal with sets of multiple interconnected json/yaml declarations, without taking forever, or spilling the stack would be another. Last but not least, it provides 'kind of generics', at least for everyone willing, to write code in FP style. A far fetched goal would be the implementation of a compiler to strip all the layers of indirection, boxing and interface lookups, introduced by the library and generating strongly typed go code, relying on the go compiler for typesafety instead.

## Advanced Functional Mechanics (Modern Implementation)
* **Trampolines**: Infinite laziness safely achieved via `iter.Seq`.
* **Zippers (Comonads)**: O(1) cursor navigation through immutable ASTs via `TreeZipper`.
* **Optics**: Deep structural immutability through generic `Lens` and `Prism`.
* **Transducers**: Fused `Map` and `Filter` pipelines for high-performance iteration.
* **Kleisli Arrows**: Combinatorial logic (`***`, `&&&`) for IO Monad Adapters.
* **Monad Transformers**: Stacking Monadic contexts safely (e.g., `EitherT`).

---

> [!WARNING]
> **Legacy / Obsolete Documentation**
> 
> The sections below document the original Go 1.15 prototype. The implementation described here relied heavily on variadic arguments, empty interfaces, and heavy boxing (`IntVal`, `StrVal`). 
> 
> With the V2 modernization using Go Generics, **this approach is obsolete**. However, the original code remains fully intact and passes all tests inside the `legacy/` directory for historical reference and backward compatibility.

## Features:

* lazy execution
* strongly typed
* pattern matching
* user defined types
* function composition
* parametric polymorphism
* partial application & curry
* Hindley–Milner(esque) type system
* monadic encapsulation of side effects 
* generic/parametric collections (list, vector, set)
* type classes providing methods, that can be derived from
* fast (bitflag based) typechecking independent from reflect
* lots of behaviour sharing interfaces, no 'empty interface' instances
* unboxed alias types for all Go Natives and a couple of other useful native types

## Implementation:

* relies on variadic arguments and multiple return values
* no empty interfaces
* the data package provides:
  * type aliases for all native types 
  * alias types for slices of all native types
  * alias types for sets of (interfaced) natives, for the most useful key types
  * type aliases for math/big Int, Float and Rat
  * the 'Native' interface shared by all alias types. 
  * lots of other interfaces for groups of types that share behavior such as.
    * numerals
    * textuals (string, rune, byte, byte slice)
    * binarys (byte, byte slice)

  * a generic slice of (interfaced) natives, featuring helpful methods on slices
  * type flags that make the type detectable without relying on reflection.
  * poor mans implementation of 'type classes', by checking for membership of type flags in or concatenated sets of flags
  * lots of type conversions implemented as methods on alias types.

* the function package provides
  * more interfaces for expressions that share common behaviour
  * pattern matching

  * user defined types:
    * type constructor
    * parametric, strongly typed expression
    * type classes providing default methods to derive from
    * parametric types that construct values of sub-types, based on argument type(s)

  * type pattern with elements of types:
    * fixed type flag 
    * user defined symbol
    * type variable
    * lexical token

  * generic implementations of:
    * comparator
    * test
    * case
    * case-switch
    * sub type element
    * conditional (if|then|else)
    * maybe (just|none)
    * option (either|or)
    * enum
    * tuple
    * pair
    * indexed pair (accessor of type 'int')
    * keyed pair (accessor of type 'string')
    * list
    * list of pairs
    * vector
    * vector of pairs
    * set
    * operator
    * generator
    * accumulator
    * number
    * letter
    * text
    * binary data
    * functor
    * applicative
    * monad

## Roadmap:

* completing the base features (see above)

  * current state of implementation:
    * data package entirely as described
    * generic collections, pairs, maybe and optional 
    * function composition
    * pattern matching

  * work in progress:
    * type & value constructor for user defined types
    * type checked and pattern matched expressions and types
    * enum & tuple
    * type classes
    * functor
    * applicative 
    * monad

* function composition based parser library
* ML syntax parser
* ML interpreter

* data structure generation from:
  * json/yaml
  * ini format
  * markdown
  * ‥.

* compiler to generate go native typed code, stripped of as much indirection as possible

## Modernized Roadmap (Go 1.18+)

With the advent of Go Generics (Go 1.18) and Go Iterators (Go 1.23), the architectural goals of `gatwd` can be modernized while preserving its core ambition: building a runtime engine for an ML-like DSL with dynamic type definition and infinite lazy loops.

* **Preserve the Runtime Type Engine:** 
  The custom `BitFlag` type system remains strictly necessary for the DSL interpreter. Go Generics cannot generate new types dynamically at runtime. The BitFlag system allows the DSL to parse user definitions (e.g., `type Quasi = A | B`) and enforce them purely at runtime.
* **Optimize AST Nodes via Generics:** 
  Instead of genericizing the DSL types, use Go Generics to type-safe the AST nodes in the Go engine (e.g., `type Node[T any] struct { Value T; TypeInfo BitFlag }`). This removes the massive overhead of boxing everything into the `Expression` interface and type assertions.
* **Modernize the Trampoline via Iterators:** 
  The manual `for head, tail = tail.Continue()` trampoline logic (which allowed infinite recursion depth without stack limits) can be gracefully replaced by Go 1.23 native `iter.Seq`. This provides the exact same pull-based lazy evaluation (trampolining) for infinite sequences but in an idiomatic, high-performance way.
* **Pivot Focus:** 
  Shift away from building a generic functional standard library for Go (which is largely solved natively now) and double down on building the AST parser, ML syntax interpreter, and jsonnet-style configuration engine using the modernized architecture.
