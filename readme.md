# Gophers All The Way Down

Gophers all the way down is a library embracing functional programming, written in Go.

## Purpose

This is a pre-alpha version and should under no circumstances be used yet! 

The main purpose of this library is educational. I felt to strong urge to get a basic understanding on the principles of functional programming. While the Haskell community goes great length to explain those principles, it seems to be in the nature of things, that once understood, one immediately looses the capability to explain those principles. The same is true it seems, for the principles of higher mathematics. Regarding the latter, I found a way to overcome this 'law of nature' in the past, by implementing algorithms I had trouble understanding in code. Having code compile and tests go green rules out misconception and provides the confidence necessary, not to accidently overestimating the personal level of understanding. Having made this experience a couple of times in the past, I figured, why not write a functional library?

Former self educational projects had no other purpose but self education. There was always some standard library implementing, whatever I was working on in far better and more professional ways and consequently those projects never where much of a use. With this exercise that might be different. While Java has Scale, C# has F# and JavaScript has Closure, I couldn't find any functional programming library written in Go, that was particularly useful, let alone providing an alternative ML like syntax to write the code in.

Should this library reach the necessary stability and set of features, it should be possible to write a functional parser library on top, implementing ML like Syntax. Generating Domain specific languages could be one example of a useful application. Building a decent version of a jsonnet/ksonnet interpreter, able to deal with sets of multiple interconnected json/yaml declarations, without taking forever, or spilling the stack would be another. Last but not least, it provides 'kind of generics', at least for everyone willing, to write code in FP style. A far fetched goal would be the implementation of a compiler to strip all the layers of indirection, boxing and interface lookups, introduced by the library and generating strongly typed go code, relying on the go compiler for typesafety instead.

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
