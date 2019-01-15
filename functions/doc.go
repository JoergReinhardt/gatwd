/*
FUNCTION GENERALIZATION

  ambda calculus states, that all functions can be expressed as functions
  taking one argument, by currying in additional data and behaviour. all
  computation can then be expressed in those terms‥. and while that's the base
  of all that's done here, and generally considered to be a great thing, it
  also turns out to be a pain in the behind, when applyed to a strongly typed
  language on real world problems.

  to get things done anyway, data types and function signatures need to be
  generalized over, in a more reasonable way. data types of arguments and
  return values already get generalized by the data package using type aliasing
  and adding the flag method.

  functions can be further discriminated by means of arity (number & type of
  input arguments) and fixity (syntactical side, on which they expect to bind
  there parameter(s)). golangs capability of returning multiple values, is of
  no relevance in terms of functional programming, but very usefull in
  imlementing a type system on top of it. so is the ability to define methods
  on function types. functions in the terms of godeep are closures, closing
  over arbitrary functions together with there arguments and return values,
  btw. placeholders there of and an id/signature poir for typesystem and
  runtime, to handle (partial} application and evaluation.

  to deal with golang index operators and conrol structures, a couple of
  internal function signatures, containing non aliased types (namely bool, int
  & string) will also be made avaiable for enclosure.
*/
package functions
