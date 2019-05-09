/*
  runtime is a monad, representing the evar changing state of an ongoing
  computation on a stream of commands and/or data, resulting in some kind of
  transformation and return of that transformed state, encapsulated in a new
  instance of the runctime type.

  the contained state is never exposed outside the monads scope. to function
  composition of runtime state, a dynamic set of methods will be returned
  instead. a reference to the (shared) state is passed to methods as first
  argument, when called. the method set may be mutated at any time during
  existence of the runtime instance.

  input is provided by a shared buffer, that can be read from, or written to by
  state function, as well as other instances, writing to the buffer, by means
  of methods from the method set. output is yielded into a queue, again shared,
  with possible consumers of that output. the heap is another thred safe slice
  of values, that may, or may not be shared with functions outside runtimes
  scope, and/or more likely asynchronous subprocesses of this runtime instance.

  the stack is not intendet to be shared and orovides serialization of the
  computation. the set of symbols maps names to references to their definitions
  and/or instances on the heap. a signal channel provides a way to interfer
  with, or trigger the evaluation, when runtime state is run in a close loop. a
  value updated by a side effect, may for instance signal it's avaiability, to
  trigger previously suspended evaluation, depending on that value state fnc
  may also return all by it self, when state evaluation comes to a halt, by
  returning nil, instead of the next runtime continuation.
*/
package functions

import ()
