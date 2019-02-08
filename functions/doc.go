/*
FUNCTIONS

  interfaces of higher order functional types.

  the functions package imports the data package, to provide primary data
  types, as base to define the functional data types up on. all function values
  implement data.Primary. functional types additionally have to implement the
  Call(...Value) Value method. functional values are implemented as (closure)
  function types.

  all functional types implement data.Primary, as well as data.Evaluable.
  function types that return instances of data.Primary are of the primary type
  of the return value concatenated to 'data.Function' constant. the functional
  types that don't return an instances of a primary, use the additional
  constants and and sets of constants defined in the data package, not directly
  associated to the types defined there (HigherOrder, Collection, Set...). that
  way every instance of every type can be determined as being an instance of
  one particular primary type, (which is one of the premises needed to
  implement arithmetric, parametric types) and guarantuees each type to
  implement the golang interfaces that primary type is expected to implement.

  instances of type value can either be defined as function type (here, or
  upstream), or just by remapping a data.Primary instances Eval() method to be
  the Call(...Value) method of some closure (of a function type implementing
  Call). the passed Value is not needed, or expected and therefore ignored,
  when called.
*/
package functions
