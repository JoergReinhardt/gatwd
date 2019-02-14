/*
  HEAP OBJECT CONSTRUCTORS

    this file contains implementations of constructors for heap objects. they
    use the constructors from functions/constructors.go &
    functions/functions.go as closure to be evaluated to form the constructors
    for inbuildt static types of closures. describes and parametrizes them by
    creating appropriate info tables and defining and instanciating types to
    hold info and data associated with the particular kind of heap-object they
    construct.
*/
package run

import (
	d "github.com/JoergReinhardt/gatwd/data"
	f "github.com/JoergReinhardt/gatwd/functions"
)

//////////////////////////
////// HEAP OBJECTS /////
/////
//// DECLARATIONS
///
// define passed string as symbol
// & let it point to primary value
// allocated on heap.
func declareGlobalSymbol(
	name string,
	obj *Object,
) *Object {
	var object = allocateObject()
	(*object).Info.Length = obj.Info.Length
	(*object).Info.Arity = obj.Info.Arity
	(*object).Info.Propertys = obj.Info.Propertys
	(*object).Otype = Declaration
	(*object).Value = f.NewNaryFnc(func(...f.Value) f.Value {
		return f.NewPrimaryConstatnt(d.New(name))
	})
	(*object).Refs[0] = obj
	return object
}

// declares a named free variable in local scope
func declareLocalSymbol(
	name string,
	scope *Object,
	obj *Object,
) *Object {
	var object = allocateObject()
	(*object).Info.Length = obj.Info.Length
	(*object).Info.Arity = obj.Info.Arity
	(*object).Info.Propertys = obj.Info.Propertys
	(*object).Otype = Declaration
	(*object).Value = f.NewNaryFnc(func(...f.Value) f.Value {
		return f.NewPrimaryConstatnt(d.New(name))
	})
	(*object).Refs = append(object.Refs, scope, obj)
	return object
}

// new anonymous localy scoped free variable
func declareAnonymous(
	scope *Object,
	obj *Object,
) *Object {
	var object = allocateObject()
	(*object).Info.Length = obj.Info.Length
	(*object).Info.Arity = obj.Info.Arity
	(*object).Info.Propertys = obj.Info.Propertys
	(*object).Otype = Declaration
	(*object).Value = obj.Value
	(*object).Refs = append(object.Refs, scope, obj)
	return object
}

////////////////////////////
//// DATA ALLOCATION //////
///		     /////
// (see also run/dataObjects.go)
// a closure returning a primary data instance
//
// PRIMARYS (COLLECTIONS INCLUDED)
func allocatePrimary(data ...d.Primary) *Object {
	return allocatePrimaryData(data...)
}

// PRIMARY PAIR
func allocatePrimaryPair(a, b d.Primary) *Object {
	return allocateAtomicConstant(d.NewPair(a, b))
}

//////////////////////////////////////////
////// FUNCTION VALUE HEAP OBJECTS //////
/////
//// function application expects all arguments to be in head normal form,
/// atomic, free and unbound to be finally evaluated and memoized for further
//  use‥. this is rarely the case when program execution is demanded. to reach
//  head normal form, expressions generate heap objects to evaluate
//  subexpressions. upate frames get pushed on the stack. those evaluations
//  trigger further evaluations and generate more heap objects and stack frames,
//  until atomic values, or saturated function applications in head normal form
//  are encountered. at that point actual evaluation starts, update frames get
//  updated, overwrite their return addresses with results accordingly, which in
//  turn triggers further evaluation of thunk evaluations previously suspendet
//  due to non evaluated argumens, and so on‥. until heat death of universe
//  spoils all the fun.
//
//   ‥.now for real:
//   - all args are expected to be atomic
//   - all funcs are in hnf ¬ free variables  ⇒ all refs are args
//   - if arity < ref →  call under saturated ⇒ partial application
//   - if arity ≡ ref →  call saturated       ⇒ function application
//   - if arity > ref →  call over saturated  ⇒ call continuation
//
// generic constructor exposing all parameters, to be used by specific constructors
func instanciateFunction(
	length Length,
	arity Arity,
	props Propertys,
	expr f.Callable,
	refs ...*Object,
) *Object {
	var object = allocateObject()
	(*object).Info.Length = length
	(*object).Info.Arity = arity
	(*object).Info.Propertys = props
	(*object).Otype = FunctionCall
	(*object).Value = expr
	(*object).Refs = append(object.Refs, refs...)
	return object
}

///////////////////////////////////
//// FUNCTIONS OF KNOWN ARITY ////
///
// functions application for functions of known arity, takes exactly as much
// parameters as the function definition demands, so both can be inferred. in
// case more, or less parameters are passed, a call continuation, or partial
// application will be allocated and pushed instead.
//
// CONSTANT
func instanciateConstant(
	props Propertys,
	expr f.ConstFnc,
) *Object {
	return instanciateFunction(
		Length(0),
		Arity(0),
		props,
		expr,
	)
}

// UNARY
func instanciateUnary(
	props Propertys,
	expr f.UnaryFnc,
	argument *Object,
) *Object {
	return instanciateFunction(
		Length(1),
		Arity(1),
		props,
		expr,
		argument,
	)
}

// BINARY
func instanciateBinary(
	expr f.BinaryFnc,
	props Propertys,
	first *Object,
	second *Object,
) *Object {
	return instanciateFunction(
		Length(2),
		Arity(2),
		props,
		expr,
		first, second,
	)
}

// OPERATOR (BINARY IN-FIX)
func instanciateOperator(
	expr f.BinaryFnc,
	left *Object,
	right *Object,
) *Object {
	return instanciateBinary(expr, InFix, left, right)
}

// NARY
func instanciateNary(
	expr f.NaryFnc,
	props Propertys,
	args ...*Object,
) *Object {
	var arglen = len(args)
	return instanciateFunction(
		Length(arglen),
		Arity(arglen),
		props,
		expr,
		args...,
	)
}

// partial application object keeps the result of an undersaturated call to a
// function as a closure over the part of the arguments list that has been
// passed in that call.  the partial application object has the arity of the
// original funcrtion reduced by the number of arguments that got passed.  and
// expects further arguments to be passed in consequtive calls until a call
// completes the list of arguments, in which case the function will return the
// result.
//
// the argument list of this object type is empty, since if further references
// to arguments where known at the time of object creation, they would have
// been applyed. substitution of the missing parameters happens at another call
// site after this object has been passed, or called by name. at that point,
// another call object of type partial application, value, or indirection  will
// be created as result and overwrite this instance.
//
// PARTIAL APPLICATION
func partialApplication(
	arity Arity,
	props Propertys,
	call f.Callable,
) *Object {
	var object = allocateObject()
	(*object).Info.Length = Length(0)
	(*object).Info.Arity = arity
	(*object).Info.Propertys = props
	(*object).Otype = PartialApplication
	(*object).Value = call
	return object
}

// call continuation pushes arguments passed, but not consumed by the
// preceeding call. on to the stack until the preceeding call returns a
// value (callable function).
//
// the expression reference has not been evaluated to a function, at the time
// of object creation, but will be updated to point to a callable function,
// when the control flow returns to this frame, at which point, it will be
// applyed tp the arguments, stored here.
func callContinuation(
	expr *Object,
	args ...*Object,
) *Object {
	var object = allocateObject()
	(*object).Info.Length = Length(len(args))
	(*object).Info.Arity = expr.Arity
	(*object).Info.Propertys = expr.Propertys
	(*object).Otype = CallContinuation
	(*object).Value = expr.Value.(f.Callable)
	(*object).Refs = args
	return object
}

func caseContinuation(
	scrutenee f.Value,
	cases ...*Object,
) *Object {
	var object = allocateObject()
	(*object).Info.Length = Length(len(cases))
	(*object).Info.Arity = Arity(1)
	(*object).Info.Propertys = cases[0].Propertys
	(*object).Otype = CaseContinuation
	(*object).Value = cases[0].Value.(f.Callable)
	(*object).Refs = cases
	return object
}

// an object indirection with value pointing to referenced entry code & pointer
// and copy of it's info table.  reference as single reference.
func referTo(ref *Object) *Object {
	var object = allocateObject()
	(*object).Info.Length = ref.Length
	(*object).Info.Arity = ref.Arity
	(*object).Info.Propertys = ref.Propertys
	(*object).Otype = Indirection
	(*object).Value = ref.Value.(f.Callable)
	(*object).Refs = append(object.Refs, ref)
	return object
}

// thunk is a, possibly composed, expression. a thunk may contain other thunks,
// local free variables, references to global variables, and might enclose over
// other parameters. thunk evaluation updates the thunk object, which may
// result in another thunk object, or a value, constant, primary‥.  that can't
// be further evaluated.
//
// every piece of source code that forms a valid expression is a thunk, since
// all including definitions of locals and/or global values and types,
// declarations of function- and primary values as literals, or referenced by
// name, function- and base operator applications‥. neccessary to evaluate such
// an expression are included, either referenced by name or as literal value,
// so that it doesn't take any further arguments to evaluate all contained sub
// expressions recursively. the return address is the declared receiver of the
// expressions return value (which allways exists, since in functional
// languages everything is an expression & has to return a value).
//
// another type of thunks are composed data constructors, generators & io
// related functions, since they enclose over, or access data that
// co-determines the evaluation result (along with arguments that may, or may
// not get passed)
//
// for thunk evaluation, an update frame get's pushed on to the stack, the
// thunk expression get's evaluated with it's references as agruments and the
// return value (which may be another thunk value) overwrites the thunk object
// on the heap.
//
// all thunks are eventually reduced by repeated evaluation, whenever the value
// they return is needed. each succecutive evaluation generates a new instance
// to express the altered (usually reduced) state of the expression, until head
// normal form is reached. for data structures that's achieved on depletion,
// when all data has been consumed. for recursive expressions, when all
// subexpressions have been evaluated and reduced to generate a single
// expression in srong head normal form.
// a program completes, when all
//
// expressions neccessary to generate the demanded output, have successfully
// been reduced to head normal form, which includes recursive evaluation of all
// thunks expressing sub expressions that yield values, necessary to do that.
//
// expressions may generate infinite lists, in which case some other exit
// condition has to be defined to eventually reach program completion.
func thunk(
	expr f.Callable,
	props Propertys,
	refs ...*Object,
) *Object {
	var object = allocateObject()
	(*object).Info.Length = Length(len(refs))
	(*object).Info.Arity = Arity(0)
	(*object).Info.Propertys = props
	(*object).Otype = Thunk
	(*object).Value = expr
	(*object).Refs = append(object.Refs, refs...)
	return object
}

// blackhole is an indirection that keeps a thunk from being evaluated, while
// it's allready been evaluated. keeps evaluation of recursive thunks lazy.
func blackHole(ref *Object) *Object {
	var object = allocateObject()
	(*object).Info.Length = Length(0)
	(*object).Info.Arity = Arity(0)
	(*object).Info.Propertys = ref.Propertys
	(*object).Otype = BlackHole
	(*object).Value = f.NewNone()
	(*object).Refs = append(object.Refs, ref)
	return object
}

// byte code contains a piece of source code, the start position at which that
// piece sourcecode appears, (miss-) uses the info.Length field, to save the
// length of that piece in byte. and a reference to the heap object, that got
// created and allocated based on this piece of the source code.
func byteCode(
	pos int,
	text string,
	ref *Object,
) *Object {
	var object = allocateObject()
	(*object).Info.Length = Length(len(text))
	(*object).Info.Arity = Arity(0)
	(*object).Info.Propertys = Default
	(*object).Otype = ByteCode
	(*object).Value = f.NewPair(f.NewPrimaryConstatnt(d.IntVal(pos)),
		f.NewPrimaryConstatnt(d.StrVal(text)))
	(*object).Refs = append(object.Refs, ref)
	return object
}

// sys call keeps references to all io & other sys call related objects, like
// command line flags, os process control signals, byte code of the running
// program, buffers, streams, etc‥.  each thread has exactly one sys call
// reference which may contain further (sub-) threads, that may reference it's
// parent, or some other instances of a sys call object. the list of references
// may be mutated at any given time by adding, or removing objects, objects can
// be mutated, or replaced,under the condition that the objects reference
// address and object type remain constant over the entire objects lifetime
// (since it might be referenced by other objects, or enclosed by closures yet
// to be evaluated).
//
// if an io sys call has been allocated, the returned channel is drained once
// every time the state function is called and the expression it yields is
// evaluated. the expression is expected to perform all side effects, that are
// sheduled to be performed by evaluating expressions referenced in the list of
// io sys calls references & return a new expression to replace the current one
// to be evaluated the next time state function is been called.
//
// layout of the io sys call is entirely left to the program. some sort of
// sheduling may be implemented to return no-ops, in case none of the system
// tasks needs imediate evaluation.
func system(
	expr f.Callable,
	refs ...*Object,
) *Object {
	var object = allocateObject()
	(*object).Info.Length = Length(len(refs))
	(*object).Info.Arity = Arity(0)
	(*object).Info.Propertys = Eager | SideEffect | Mutable
	(*object).Otype = ByteCode
	(*object).Value = expr
	(*object).Refs = append(object.Refs, refs...)
	return object
}

// a thread object contains another state function enclosing a state struct
// which references a heap. that can be an independent instance of a graph of
// linked heap objects, that graph can be connected by references through cross
// references defined in objects referencing thread safe data structures, or a
// thread safe reference to some shared heap. each thread also has its own
// instance of a stack. the structure of the thread object, if and how it is
// connected to other running threads is entirely defined by the program.
// synchronization between threads is performed by sharing thread safe data
// structures from the list of syscall objects. threads can be sheduled by an
// enclosing control thread, but may as well share control on a mutual base, be
// arranged in form of a ring and evaluated round robin‥.
//
// if a sys-call object exists in a thread, the thread may list a reference to
// itself among it's syscall object references to expose process control and or
// data to other threads, a cli-/, gui-/, or browser event loop, rpc message bus
// thigy‥.
func thread(
	sf StateFnc,
	refs ...*Object,
) *Object {
	var object = allocateObject()
	(*object).Info.Length = Length(len(refs))
	(*object).Info.Arity = Arity(0)
	(*object).Info.Propertys = Eager | Mutable
	(*object).Otype = System
	(*object).Value = sf
	(*object).Refs = append(object.Refs, refs...)
	return object
}

// synchronous accessable io, other side effect, or shared data blocks call
// untilil data has been successfully read, or written, or an appropriate
// option has been returned. the latter implements non-blocking calls to a
// blocking device.
//
// referenced closure may, or may not be readable, writeable, or both, which is
// indicated with -1 for writable, 0 for readable and 1 if both is the case.
// shareability (thread safety) and eagernes (eager/lazy) must be indicated by
// setting the corresponding flags true.
func synchronous(
	rw f.Callable,
	writeable int,
	shared bool,
	eager bool,
	refs ...*Object,
) *Object {
	var otype = Sync
	var props = SideEffect
	switch {
	case eager:
		props = props | Eager
	case shared:
	case writeable < 0:
		props = props | Mutable
	case writeable == 0:
		props = props | Data
	case writeable > 0:
		props = props | Mutable | Data
	}
	var object = allocateObject()
	(*object).Info.Length = Length(len(refs))
	(*object).Info.Arity = Arity(1)
	(*object).Info.Propertys = props
	(*object).Otype = otype
	(*object).Value = rw
	(*object).Refs = append(object.Refs, refs...)
	return object
}

// asynchronous accessable io, other side effect, or shared data. behaves like
// a queue to place call continuations in, that take the yielded value as
// parameter. all continuations in the queue will be updated and pushed on the
// call stack for evaluation, whenever a new value is yielded. yielded value
// may be a function demanding parameters,implementing a wtite to an
// asynchronous resource, a (parameterless) closure that returns values,
// implementing asynchronous read or a closure expecting mandatory, optional,
// plain, or variadic parameters and returns a value. the object can be thread
// safe or not, all, of which must be indicated to the runtime by setting the
// writeable, shared and eager agruments accordingly.
func asynchronous(
	queue f.Callable,
	writeable int,
	shared bool,
	eager bool,
	refs ...*Object,
) *Object {
	var otype = Sync
	var props = SideEffect
	switch {
	case eager:
		props = props | Eager
	case shared:
		otype = otype | Shared
	case writeable < 0:
		props = props | Mutable
	case writeable == 0:
		props = props | Data
	case writeable > 0:
		props = props | Mutable | Data
	}
	var object = allocateObject()
	(*object).Info.Length = Length(len(refs))
	(*object).Info.Arity = Arity(1)
	(*object).Info.Propertys = props
	(*object).Otype = otype
	(*object).Value = queue
	(*object).Refs = append(object.Refs, refs...)
	return object
}
