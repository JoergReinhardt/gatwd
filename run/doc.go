/*
  OBJECT CODE, STACK FRAMES & STATE

    struct to hold the info table flags common to all heap and stack onjects.
    thanks to golang memory layout, those flags are held perfectly tight (not a
    singele bit lost to alignment, or struct headers). the 'payload' of
    arguments, bytecode, or whatever (depends on frame type), is kept as value
    of type interface (which is a pointer) in a slice. all other info get's
    copyed to the stack. arguments are usually references, but may also be copys
    of (aliased, see data) values, to get them allocated on the 'real' stack
    during runtime. the fayout field marks, which arguments of type references
    with 'one'. frame-/ & object implementations, that utilze native type
    arguments are expected to embed a copy of the base info table, followed by
    additional struct fields, one field per argument, named according to the
    name of the corresponding layout constant.

    aiming for a fast implementation, the design trys to keep data from escaping
    to the heap, by using pointers as little as possible. gatwds internal heap
    implements a directed cyclic graph and can't be implemented in a more
    performant way (that i know of), than using pointers. all all data belonging
    to one object, is held thightly in a struct. contained values and arguments
    may be enforced to be arranged sequentially in a consequtive piece of memory,
    by serializing them into a byte array using encoding/gob.

    the closure that needs to be evaluated to yield the value, will allways
    involve some sort of name lookup and a function call at least, which is why
    using a call to an interface, can't hurt that much either‥. TODO: that needs
    of course to be validated later.

    gatwds internal stack, is a slice of fixed size structs, which should help
    keep things & stuff from ecaping to the heap as much as possible, while
    still allowing for a stack of arbitrary size. since frames have to be of a
    fixed size to make the runtime use arrays of values instead of pointers. to
    make that possible, arguments and free variables need to go some place
    else‥. they are accessable by dereferrencing the pointer to the value
    closure, which needs to be dereferenced and evaluated anyway, so accessing
    it's data should not add extra cpu cycles. for performance critical tasks,
    serialized versions of the values will be accessable via object and can be
    memcopyed to the argument stack. each frame has a length field, so that the
    runtime can access and pop the argument stack synchronously to the execution
    stack. the argument stack may be accessed using the encoding/gob interfaces,
    but that would involve dynamicly dispatched method calls. also gobs data
    structures are pretty well documented, so it's been suggested, to rather
    access those values directly based on length field and offset. in the end,
    that's of course left to the particular object implementation.

    we'll see how good this design will hold up aginst reality‥. otherwise some
    reimplementation involving fixed size arrays and/or use of the unsafe
    package will most likely occur. will be evaluated later)

    for argument sets > 32 args, an array (mem-copy, fixed size, no slice, to
    get it on the 'real' stack during computation) per argument type is expected

    TODO: come up with a naming convention for arrays of memcopyed values.

    TODO: see if argument reference/copy tagging can also be used to implement
    currying, partial application and discrimination between strong-/ and weak
    normal form (neccessary to implement full laziness, aka 'knowing when to
    stop', or 'suspended evaluation') →  compare number of pointer type
    arguments that are left to evaluate, with the functions arity and copy the
    results of argument expansion as values, deleting their tags, so that it
    could be evaluate if expression is saturated, evaluated and atomic in a
*/
package run
