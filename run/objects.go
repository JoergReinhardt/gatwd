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
//// SYMBOL DECLARATION
///
// define passed string as symbol
// & let it point to primary value
// allocated on heap.
func declareGlobalSymbol(name string, obj *object) object {
	return object{
		newInfo(
			obj.info.Length,
			obj.info.Arity,
			obj.info.Propertys,
		),
		Declaration,
		f.NewNaryFnc(func(...f.Value) f.Value {
			return f.NewPrimaryConstatnt(d.New(name))
		}),
		[]*object{obj},
	}
}

// declares a named free variable in local scope
func declareLocalSymbol(name string, scope *object, obj *object) object {
	return object{
		newInfo(
			obj.info.Length,
			obj.info.Arity,
			obj.info.Propertys,
		),
		Declaration,
		f.NewNaryFnc(func(...f.Value) f.Value {
			return f.NewPrimaryConstatnt(d.New(name))
		}),
		[]*object{scope, obj},
	}
}

// new anonymous localy scoped free variable
func declareAnonymous(scope *object, obj *object) object {
	return object{
		newInfo(
			obj.info.Length,
			obj.info.Arity,
			obj.info.Propertys,
		),
		Declaration,
		obj.Expr,
		[]*object{scope, obj},
	}
}

////////////////////////////////////////
//// DATA CONSTRUCTOR ALLOCATION //////
///				 /////
// (see also run/dataObjects.go) ////
//
// object constructors allways have
// all their arguments satisfied &
// allocate the constructed resut
// on the heap
func refer(obj ...*object) []*object { return obj }

// a closure returning a primary data instance
func allocatePrimary(
	data ...d.Primary,
) object {
	return allocatePrimaryData(data...)
}

// a primary pair
func allocatePrimaryPair(
	a, b d.Primary,
) object {
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
func instanciateFunction(
	length Length,
	arity Arity,
	props Propertys,
	expr f.Callable,
	refs ...*object,
) object {
	return object{
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
// functions of known arity, take only as much parameters as their  allows
// for. in case more, or less parameters are passed call continuation, or
// partial application will allocated and pushed.
//
// constant nuroriously passes on taking constants
func instanciateConstant(
	props Propertys,
	expr f.ConstFnc,
) object {
	return instanciateFunction(
		Length(0),
		Arity(0),
		props,
		expr,
	)
}

// unary takes a reference to it's argument
func instanciateUnary(
	props Propertys,
	expr f.UnaryFnc,
	argument *object,
) object {
	return instanciateFunction(
		Length(1),
		Arity(1),
		props,
		expr,
		argument,
	)
}

// binary takes two argument references
func instanciateBinary(
	expr f.BinaryFnc,
	props Propertys,
	first *object,
	second *object,
) object {
	return instanciateFunction(
		Length(2),
		Arity(2),
		props,
		expr,
		first, second,
	)
}

// operators are binary functions, usually defined in infix notation
func instanciateOperator(
	expr f.BinaryFnc,
	left *object,
	right *object,
) object {
	return instanciateBinary(expr, InFix, left, right)
}

// generic nary fnc. length and arity are determined by the number of passed
// arguments, under- and oversaturated calls generate other heap object types.
func instanciateNary(
	expr f.NaryFnc,
	props Propertys,
	args ...*object,
) object {
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
// function as a closure over the part of the arguments list that got passed in
// that call.  the partial application object has the arity of the original
// funcrtion reduced by the number of arguments that got passed.  and expects
// further arguments to be passed in consequtive calls until a call completes
// the list of arguments, in which case the function will return the result.
//
// the argument list of the object type is empty, since if further references
// to arguments where known at the time of object allocation, they would have
// been applyed. consequtive calls passing further arguments will overwrite
// this object with, either another partiail application, or a result of type
// value/indirection, or constant.
func partialApplication(
	arity Arity,
	props Propertys,
	expr f.Callable,
) object {
	return instanciateFunction(
		Length(0),
		arity,
		props,
		expr,
	)
}

// call continuation pushes overflow arguments on stack until oversatuated call
// returns a (function) value, that will then be applyed to these arguments.
//
// expression (not yet evaluated to a function) and arguments both get passed
// as references. the expression is an update frame expected to be overwritten
// by the result of the preceeding call. and will then be applyed to the
// arguments stored here. when suspension is evaluated. the call continuation
// therefore needs to be pushed on stack first.
func callContinuation(
	expr *object,
	args ...*object,
) object {
	return instanciateFunction(
		Length(len(args)),
		expr.Arity,
		expr.Propertys,
		expr.Expr.(f.Callable),
		args...,
	)
}

// case is very similar to a thunk, in as it contains seceral case expressions,
// that themseves possibly encloses free variables, and/or other arguments
// besides the scrutinee. the scrutinee is the mandatory expression, the case
// tests against, again a full fleged expression possibly including further
// variables and parameters. the case expression is expexted to either return
// the evaluation result of the scrutinee, in case it matched the case, or the
// next case expression to apply the scrutinee to, with the scrutenees result
// as it's argument allready enclosed. at some point a value instead of another
// case statement will be returned and taken as the value to update (overwrite,
// or redirect) the return address. the return value has to match the type of
// the return address value, which results in one case _having_ to return a
// value, or the use of an an 'Either T = T | Just' type.
//
// the objects of type evaluate-case, coordinate the evaluation of the composed
// case statement passed during initialization, by folding the list of
// references to case expressions to test against, over the scrutinee as
// initial element of the fold expecting the result to contain the evaluated
// pick that got chosen by the case-compositions evaluation.
func caseContinuation(
	scrut f.Value,
	cases ...*object,
) object {
	for _, cref := range cases {
		// each case expression in the list expects the scrutinee as
		// it's single argument. the case expression is expected to
		// know, if the scrutinee is just a value, or an expression, in
		// which case it substitutes it's arguments and evaluates it.
		// other free variables need to be enclosed by the unary case
		// expression, in which case the object would reference a
		// partioal application.
		//
		// ‥.if it's a case continuation
		if cref.Otype.Match(CaseContinuation) {
			// ‥.cast as unary function
			if cxpr, ok := cref.Expr.(f.UnaryFnc); ok {
				// ‥.apply scrutinee to case
				var result = cxpr(scrut)
				// another case to test against
				if result.TypeHO().Flag().Match(f.Case) {
				}
				// every other type of value will be returned
			}
		}
	}
	var o object
	return o
}

// an object indirection with value pointing to referenced entry code & pointer
// and copy of it's info table.  reference as single reference.
func referTo(
	ref *object,
) object {
	return object{
		ref.info,
		Indirection,
		ref.Otype,
		[]*object{ref},
	}
}

// blackhole is an indirection that keeps a thunk from being evaluated, while
// it's allready been evaluated. keeps evaluation of recursive thunks lazy.
func blackHole(
	ref *object,
) object {
	return object{
		ref.info,
		BlackHole,
		ref.Otype,
		[]*object{ref},
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
	refs ...*object,
) *object {
	// extract closure and remaining free variables from thunk
	var obj object
	return &obj
}

// extracts expressions and arguments from thunk object, parametrizizes
// function call with passed object references as free variables. all passed
// objects are assumed to be atomic! if recursive thunk evaluation is intendet,
// it is been dealt with by the extracted value expression, that get's called.
// the closure is in control over evaluations lazyness, fixity, arity, return
// value, pushing of frames, blackhole-shadowing & updating the heap object, etc‥.
func extractThunkExpression(refs ...*object) (val f.Value, args []*object) {

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
				args = []*object{refs[1]}
				val = c.Call(args[1])
			}
			// call with multiple free values
			if len(args) > 2 {
				args = []*object{}
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
