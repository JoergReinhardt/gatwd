/*
OBJECT CODE GENERATION

  in an [https://xkcd.com/224/](ideal functional programming based world),
  everything would be functions calling functions on a per name base‥. but
  unfortunately, we can't have nice things. literature and pop culture alike
  suggest, that the framerate of a universe defined in such terms would
  terribly suffer in case said universe is run on common day hardware.

  literature also suggests, that there is a common solution, known as 'object
  code', that allows to strore and process the huge acyclic graph that program
  syntax tend's to form in highly efficient manners. which further suggests,
  that literature is the main problem here, since as may as it bring's the
  solution, it's ignorance erroding nature made me aware of that problem (and
  the whole world of abstractions it plays a role in) in the first place so,
  darn you literature!

  object code consists of objects of previously known size, combining an
  instructions with the data (or references there of) it is supposed to be
  applyed up on and a return address to write the result to. the data fields in
  such objects need to be machine values, or addresses there of and the
  operation needs to be a machine operation at runtime‥. but that's left to go
  compiler and linker to figure out. internaly those fields take the form of
  names, that translate to index positions on the internal heap, which is
  represented by a slice (or maybe even array) holding native value instances.

  during program creation, compilation, or interpretation. the operation is a
  golang native function implementing certain interfaces, to make it applyable
  to the arguments, during runtime. the function can be defined as such in
  godeep, or upstream and be compiled in staticly, or be a composition of parts
  predefined by this library.

  to address instances of function, or data values (buildtin, or user defined.
  static, or dynamic), internal dynamic references need to exist, which is
  beeing accomplished by having an internal representation of a heap and
  exexution stack. a pre known object size allows for performant internal
  representation, storage and accessability of those objects.

  let's do something clever here‥., or even better, blalantly steal from
  haskell's implementation. objects will consist of an info table, having a
  fixed size for all object types a flag to mark the particular object type and
  a payload of arguments of variable length, that may contain a mixture of
  references to the heap (or golang ptr, not shure yet), as well as copys of
  unboxed values. the info tables layout field encodes length and type of the
  variable payload of arguments, the 'operation' points to native golang code
  (staticly linked) in form of a closure, that evaluates the up on the pauload
  arguments. payload fields can be of type address, as well as native values.
  to accompany for that, the layout has info about the number of payload
  fields, and if they are reference, or value.  (references go first, than
  values). values can of course reference to other functions, therefore objects
  recursively and such implicitly form the AST.

 for evaluation, heap values are converted to corresponding stack frame
 instances, allocated and evaluated on the stack by a closure of type StateFn.
 those closures close over the entire application state, including stack and
 heap to access directly and consequently don't get any parameters passed.
 state functions progress the state enclosed over and return another state
 function enclosin that new state, to be called in a loop until program
 completion.

 while godeep thrieves to be closures all the way down, during compilation &
 interpretation, all those closure declarations and calls would turn into a
 performance bottleneck. usually virtual machines at this stage don't deal
 with types any more and are only concerned with addresses, bytecode and
 length fields.

 since godeep is currently in the phase of being bootstrapped, a solution
 somewhere inbetween has been chosen. info tables tend to get passed around a
 lot and therefore are implemented as structs containing exclusively fields
 of fixed size. golangs struct allignment magic makes that pack tightly.
 payloads on the other hand currently contain typed values, which can vary in
 sice. the heap, as well as the stack, therefore are currently slices,
 containing pointers to real values, or closures there over. in a later state
 info table and payload may also have distinct implementations, depending on
 if use as stack frame, or rather as heap object. currently those share
 implementation.

 TODO: change this to be a real byte code representation, using minimum
 required space, arrays, sync.Pool reallocation pooling‥. as soon as parser
 is bootstrapped far enough for it, to make sense. serialization and
 marshalling should be rather trivial to implement using golang interfaces
 and methods. (mark my words!)

 while object code ommits names, interpreter and parser need those user
 defined identifyers to do their thing. the current stage provides an
 external reversed index of names mapped to objects.
*/
package parse
