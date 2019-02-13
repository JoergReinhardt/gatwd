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
func declareGlobalSymbol(name string, obj *Object) Object {
	return Object{
		newInfo(
			obj.Info.Length,
			obj.Info.Arity,
			obj.Info.Propertys,
		),
		Declaration,
		f.NewNaryFnc(func(...f.Value) f.Value {
			return f.NewPrimaryConstatnt(d.New(name))
		}),
		[]*Object{obj},
	}
}

// declares a named free variable in local scope
func declareLocalSymbol(name string, scope *Object, obj *Object) Object {
	return Object{
		newInfo(
			obj.Info.Length,
			obj.Info.Arity,
			obj.Info.Propertys,
		),
		Declaration,
		f.NewNaryFnc(func(...f.Value) f.Value {
			return f.NewPrimaryConstatnt(d.New(name))
		}),
		[]*Object{scope, obj},
	}
}

// new anonymous localy scoped free variable
func declareAnonymous(scope *Object, obj *Object) Object {
	return Object{
		newInfo(
			obj.Info.Length,
			obj.Info.Arity,
			obj.Info.Propertys,
		),
		Declaration,
		obj.Expr,
		[]*Object{scope, obj},
	}
}

////////////////////////////
//// DATA ALLOCATION //////
///		     /////
// (see also run/dataObjects.go)
// a closure returning a primary data instance
//
// PRIMARYS (COLLECTIONS INCLUDED)
func allocatePrimary(data ...d.Primary) Object {
	return allocatePrimaryData(data...)
}

// PRIMARY PAIR
func allocatePrimaryPair(a, b d.Primary) Object {
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
) Object {
	return Object{
		newInfo(
			length,
			arity,
			props,
		),
		FunctionClosure,
		expr,
		refs,
	}
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
) Object {
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
) Object {
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
) Object {
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
) Object {
	return instanciateBinary(expr, InFix, left, right)
}

// NARY
func instanciateNary(
	expr f.NaryFnc,
	props Propertys,
	args ...*Object,
) Object {
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
	expr f.Callable,
) Object {
	return instanciateFunction(
		Length(0),
		arity,
		props,
		expr,
	)
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
) Object {
	return instanciateFunction(
		Length(len(args)),
		expr.Arity,
		expr.Propertys,
		expr.Expr.(f.Callable),
		args...,
	)
}

func caseContinuation(
	scrutenee f.Value,
	cases ...*Object,
) Object {
	var o Object
	return o
}

// an object indirection with value pointing to referenced entry code & pointer
// and copy of it's info table.  reference as single reference.
func referTo(
	ref *Object,
) Object {
	return Object{
		ref.Info,
		Indirection,
		ref.Otype,
		[]*Object{ref},
	}
}

// blackhole is an indirection that keeps a thunk from being evaluated, while
// it's allready been evaluated. keeps evaluation of recursive thunks lazy.
func blackHole(
	ref *Object,
) Object {
	return Object{
		ref.Info,
		BlackHole,
		ref.Otype,
		[]*Object{ref},
	}
}

// thunk object is a list of expressions that dont take values, but enclose
// over their free Variables instead. for evaluation, a thunk takes itself as
// parameter, trys to extract a callable expressions to evaluate from the set
// arguments & evaluates it. the return value replaces an update frame, that is
// expected to have been pushed on the stack, before evaluation occurs.
// depending on the return value type, a thunk may be overwritten by a new
// object created from the update. thunk evaluation is recursive, until thunk
// is replaced by some final atomic result of the evaluation. thunk evaluation
// may also go on infinite, if thunk happens to represent an infinite list.
func evaluateThunk(
	propertys Propertys,
	refs ...*Object,
) *Object {
	// extract closure and remaining free variables from thunk
	var obj Object
	return &obj
}

// extracts expressions and arguments from thunk object, parametrizizes
// function call with passed object references as free variables. all passed
// objects are assumed to be atomic! if recursive thunk evaluation is intendet,
// it is been dealt with by the extracted value expression, that get's called.
// the closure is in control over evaluations lazyness, fixity, arity, return
// value, pushing of frames, blackhole-shadowing & updating the heap object, etc‥.
func extractThunkExpression(refs ...*Object) (val f.Value, args []*Object) {

	if len(args) > 0 {
		// check first parameter for callability
		if c, ok := args[0].Expr.(f.Callable); ok {
			// the first parameter is the closure to call‥. no
			// further parameters got passed
			if len(args) == 1 {
				val = c.Call()
			}
			// an additional free value got passed
			if len(args) == 2 {
				args = []*Object{refs[1]}
				val = c.Call(args[1])
			}
			// call with multiple free values
			if len(args) > 2 {
				args = []*Object{}
				var vals = []f.Value{}
				for _, obj := range refs[1:] {
					args = append(args, obj)
					vals = append(vals, obj.Expr)
				}
				val = c.Call(vals...)
			}
		}
	} else {
		// no arguments where passed. TODO: make up your mind regarding error handling!!!
		val = f.NewNaryFnc(
			func(...f.Value) f.Value {
				return f.NewPrimaryConstatnt(d.NilVal{})
			})
	}
	return val, args
}
