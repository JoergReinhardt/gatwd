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
	d "github.com/JoergReinhardt/godeep/data"
	f "github.com/JoergReinhardt/godeep/functions"
)

// object allocation with all parameters to pass
func allocateData(
	otype Otype,
	arity Arity,
	propertys Propertys,
	value f.Value,
	refs ...*object,
) object {
	return object{
		newInfo(
			Length(len(refs)),
			arity,
			propertys,
		),
		otype,
		value,
		refs,
	}
}

// make passed name a heap allocated primary value and pass the object it
// points to as it's sole reference
func declareGlobalSymbol(name string, obj object) *object {
	var decl = allocateData(
		Declaration,
		Arity(0),
		Default,
		f.NewPrimaryConstatnt(d.New(name)),
		&obj,
	)
	return &decl
}

// declares a named free variable in local scope
func declareLocalSymbol(name string, scope *object, obj *object) *object {
	// allocate a new object based on a functions/parameter with name as
	// accessor and the passed objects value, as as argument.
	var local = allocateData(
		Declaration,
		Arity(0),
		Default,
		f.NewPrimaryConstatnt(d.New(name)),
		scope,
		obj,
	)
	// append the named parameter to the reference slice of the object
	// it has lexicaly been defined in.
	return &local
}

// new anonymous localy scoped free variable
func declareAnonymous(scope *object, obj *object) *object {
	// allocate a new object based on a functions/parameter with name as
	// accessor and the passed objects value, as as argument.
	var local = allocateData(
		Declaration,
		Arity(0),
		Default,
		f.NewPrimaryConstatnt(d.NilVal{}),
		scope,
		obj,
	)
	// append the named parameter to the reference slice of the object
	// it has lexicaly been defined in.
	return &local
}

// a closure returning a primary data instance
func allocatePrimary(
	data ...d.Primary,
) *object {
	return allocatePrimaryData(data...)
}

// a primary pair
func allocatePrimaryPair(
	a, b d.Primary,
) *object {
	return allocateAtomicConstant(d.NewPair(a, b))
}

// default nary
func declareDefaultFunction(
	arity Arity,
	value f.Value,
	refs ...*object,
) *object {
	var obj = allocateData(FunctionClosure, arity, Default, value, refs...)
	return &obj
}

// eager nary
func declareEagerFunction(
	arity Arity,
	value f.Value,
	refs ...*object,
) *object {
	var obj = allocateData(FunctionClosure, arity, Eager, value, refs...)
	return &obj
}

// ConstFnc needs to be wrapped in an nary, to comply to call convention
func declareConstant(
	propertys Propertys,
	value f.ConstFnc,
) *object {
	var val = f.NewNaryFnc(func(...f.Value) f.Value { return value() })
	var obj = allocateData(FunctionClosure, Arity(0), Eager, val)
	return &obj
}

// UnaryFnc needs to be wrapped in an nary, to comply to call convention
func declareUnary(
	propertys Propertys,
	value f.UnaryFnc,
	ref *object,
) *object {
	var val = f.NewNaryFnc(func(...f.Value) f.Value { return value(ref) })
	var obj = allocateData(FunctionClosure, Arity(1), propertys, val, ref)
	return &obj
}

// BinaryFnc needs to be wrapped in an nary, to comply to call convention
func declareBinary(
	propertys Propertys,
	value f.BinaryFnc,
	refa *object,
	refb *object,
) *object {
	var val = f.NewNaryFnc(func(...f.Value) f.Value { return value(refa, refb) })
	var obj = allocateData(FunctionClosure, Arity(2), propertys, val, refa, refb)
	return &obj
}

// operators are binary and infix
func declareOperator(
	value f.BinaryFnc,
	left *object,
	right *object,
) *object {
	var val = f.NewNaryFnc(func(...f.Value) f.Value { return value(left, right) })
	var obj = allocateData(FunctionClosure, Arity(2), InFix, val, left, right)
	return &obj
}

// partial application object keeps partially applyed function calls generated
// by an undersaturated call to a function closure
func declarePartial(
	obj *object,
	refs ...*object,
) *object {
	var o object
	// store and count argument references, reduce function arity by the
	// number of passed arguments (is expected to be smaller than arity).
	if obj.info.Arity-Arity(len(refs)) > 0 {
		o = allocateData(
			PartialApplication,
			obj.info.Arity-Arity(len(refs)),
			obj.info.Propertys,
			obj.Value,
			refs...,
		)
		return &o
	}
	return obj
}

// call continuation takes a list of arguments, keeps them in a slice for
// substitution, when called to evaluate the return value of the call that
// generated the continuation.
func callContinuation(
	refs ...*object,
) *object {
	var closure f.Callable
	switch len(refs) {
	case 0:
		closure = f.NewNaryFnc(func(...f.Value) f.Value {
			return f.NewPrimaryConstatnt(d.NilVal{})
		})
	case 1:
		closure = f.NewNaryFnc(func(...f.Value) f.Value {
			return refs[0].Value
		})
	case 2:
		closure = f.NewNaryFnc(func(...f.Value) f.Value {
			return f.NewPair(refs[0].Value, refs[1].Value)
		})
	default:
		var vals = []f.Value{}
		for _, ref := range refs {
			vals = append(vals, ref.Value)
		}
		closure = f.NewNaryFnc(func(...f.Value) f.Value {
			return f.NewVector(vals...)
		})
	}
	var obj = allocateData(
		CallContinuation,
		Arity(len(refs)),
		Data,
		closure,
	)
	return &obj
}

// case is very similar to a call continuation. return value type of the
// evaluated expressions, if the scrutenee failed to met the praedicate another
// case expression, until a value is returned. case expressions have to be
// fully satisfied in the sum of  their cases‥. aka, the last return _has_ to
// be a value, taken from a default case all passed case  expressions failed to
// evaluate.
func evaluateCase(
	refs ...*object,
) *object {
	var length = len(refs)
	var con = []*object{}
	var closure f.Callable
	switch length {
	case 0:
		closure = f.NewNaryFnc(func(...f.Value) f.Value {
			return f.NewPrimaryConstatnt(d.BoolVal(true))
		})
	case 1:
		if call, ok := refs[0].Value.(f.Callable); ok {
			closure = f.NewNaryFnc(func(...f.Value) f.Value {
				return call.Call()
			})
		}
	case 2:
		if call, ok := refs[0].Value.(f.Callable); ok {
			closure = f.NewNaryFnc(func(...f.Value) f.Value {
				return call.Call(refs[1].Value)
			})
		}
	default:
		if c, ok := refs[0].Value.(f.Callable); ok {
			closure = f.NewNaryFnc(func(...f.Value) f.Value { return c.Call(refs[1]) })
			con = refs[2:]
		}
	}
	var obj = allocateData(
		CaseContinuation,
		Arity(length),
		Default,
		closure,
		con...,
	)
	return &obj
}

// an object indirection with value pointing to references entry code & painter
// reference as single reference.
func referTo(
	ref *object,
) *object {
	var ind = allocateData(
		Indirection,
		Arity(0),
		ref.info.Propertys,
		ref.Value,
		ref,
	)
	return &ind
}

// blackhole is an indirection that keeps a thunk from being evaluated, while
// it's allready been evaluated. keeps evaluation of recursive thunks lazy.
func blackHole(
	ref *object,
) *object {
	var ind = allocateData(
		BlackHole,
		Arity(0),
		ref.info.Propertys,
		ref.Value,
		ref,
	)
	return &ind
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
	var val, args = extractThunkExpression(refs...)
	var obj = allocateData(
		Thunk,
		Arity(0),
		propertys,
		val,
		args...,
	)
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
		if c, ok := args[0].Value.(f.Callable); ok {
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
					vals = append(vals, obj.Value)
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
