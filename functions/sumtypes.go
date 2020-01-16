/*

SUM TYPES
---------
sum types are defined as collection of zero or more elements of the same
type.

that includes signature types (flat function definitions), since their
arguments & return values are allways instances of the same type(s).
polymorphic & parametric functions on the other hand are to be
considered product types, since their return value type depends either
on the type(s) of argument(s) that are/is passed, or on argument values.

NONE values implement the group & collected interfaces (among others) to
be suitable as stand in for empty collections, depleted continuations,
functors, monads and the like.  none values mark the end of
continuations, and are returned whenever computational steps don't
return a proper value.  the call methods of continuations as well as any
enclosing main function, implement functional trampolin devices, by
iterating over the return values of calling each continuation and
re-assigning the resulting continuation as the next continuation to
call.  tail recursion optimization is implemented by passing on enclosed
state from call to call, while returning none values as results.  map,
fold, apply and bind omit none values.  the last call to a continuation
then consequently yields the final result, without having to pop
recursive stack frames while passing the result on from caller to
caller.

CONSTANT values don't have arguments and are defined by their return
types.

LAMBDA is a wrapper type to instanciate arbitrary functions that match
the expression signature as instance of the expression interface.  it
implements the bare minimum methods neccessary to do so, based solely on
the enclosed functions return values.  argument type of a lambda is
undefined, the enclosed expression has to be able to deal with every
possible argument types combination.

GENERATOR & ACCUMULATOR wrap functions with appropriate signature, to
implement the expression-, generator- and accumulator interfaces.

FUNCTION VALUES are type-safe function definitions with properly defined
return and argument(s) types.  function value identity is either derived from
enclosed expression, argument- and return types, or passed in as instance of
type-symbol.
*/
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// NONE, CONSTANT & LAMBDA
	NoneVal func()
	ConsVal func() Functor
	Lambda  func(...Functor) Functor

	//// GENERATOR | ACCUMULATOR
	GenVal func() (Functor, GenVal)
	AccVal func(...Functor) (Functor, AccVal)

	//// TYPE SAFE FUNCTION DEFINITION (SIGNATURE TYPE)
	FuncVal func(...Functor) Functor
)

///////////////////////////////////////////////////////////////////////////////
//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type.  implements countable,
// sliceable, consumeable, testable, compareable, key-, index- and generic pair
// interfaces to be able to stand in as return value for such expressions.
func NewNone() NoneVal { return func() {} }

func (n NoneVal) Head() Functor                     { return n }
func (n NoneVal) Tail() Applicative                 { return n }
func (n NoneVal) Cons(Functor) Applicative          { return n }
func (n NoneVal) ConsGroup(Applicative) Applicative { return n }
func (n NoneVal) Concat(Sequential) Applicative     { return n }
func (n NoneVal) Prepend(...Functor) Applicative    { return n }
func (n NoneVal) Append(...Functor) Applicative     { return n }
func (n NoneVal) Len() int                          { return 0 }
func (n NoneVal) Compare(...Functor) int            { return -1 }
func (n NoneVal) String() string                    { return "⊥" }
func (n NoneVal) Call(...Functor) Functor           { return nil }
func (n NoneVal) Key() Functor                      { return nil }
func (n NoneVal) Index() Functor                    { return nil }
func (n NoneVal) Left() Functor                     { return nil }
func (n NoneVal) Right() Functor                    { return nil }
func (n NoneVal) Both() Functor                     { return nil }
func (n NoneVal) Value() Functor                    { return nil }
func (n NoneVal) Empty() bool                       { return true }
func (n NoneVal) Test(...Functor) bool              { return false }
func (n NoneVal) TypeFnc() TyFnc                    { return None }
func (n NoneVal) TypeNat() d.TyNat                  { return d.Nil }
func (n NoneVal) Type() TyDef                       { return Def(None) }
func (n NoneVal) TypeElem() TyDef                   { return Def(None) }
func (n NoneVal) TypeName() string                  { return n.String() }
func (n NoneVal) Slice() []Functor                  { return []Functor{} }
func (n NoneVal) Flag() d.BitFlag                   { return d.BitFlag(None) }
func (n NoneVal) FlagType() d.Uint8Val              { return Kind_Fnc.U() }
func (n NoneVal) Continue() (Functor, Applicative)  { return NewNone(), NewNone() }
func (n NoneVal) Consume() (Functor, Applicative)   { return NewNone(), NewNone() }

///////////////////////////////////////////////////////////////////////////////
//// GENERIC CONSTANT DEFINITION
///
// declares a constant value
func NewConstant(constant func() Functor) ConsVal { return constant }

func (c ConsVal) Type() TyDef             { return Def(Constant, c().Type(), None) }
func (c ConsVal) TypeIdent() TyDef        { return c().Type().TypeId() }
func (c ConsVal) TypeRet() TyDef          { return c().Type().TypeRet() }
func (c ConsVal) TypeArgs() TyDef         { return Def(None) }
func (c ConsVal) TypeFnc() TyFnc          { return Constant }
func (c ConsVal) String() string          { return c().String() }
func (c ConsVal) Call(...Functor) Functor { return c() }

//// GENERIC FUNCTION DEFINITION
///
// declares a constant value
func NewLambda(fnc func(...Functor) Functor) Lambda {
	return func(args ...Functor) Functor {
		if len(args) > 0 {
			return fnc(args...)
		}
		return fnc()
	}
}

func (c Lambda) Call(args ...Functor) Functor {
	if len(args) > 0 {
		return c(args...)
	}
	return c()
}
func (c Lambda) String() string       { return c().String() }
func (c Lambda) TypeFnc() TyFnc       { return c().TypeFnc() }
func (c Lambda) Type() TyDef          { return c().Type() }
func (c Lambda) TypeIdent() TyDef     { return c().Type().TypeId() }
func (c Lambda) TypeReturn() TyDef    { return c().Type().TypeRet() }
func (c Lambda) TypeArguments() TyDef { return c().Type().TypeArgs() }

///////////////////////////////////////////////////////////////////////////////
//// GENERATOR
///
// expects an expression that returns an unboxed value, when called empty and
// some notion of 'next' value, relative to its arguments, if arguments where
// passed.
func NewGenerator(init, generate Functor) GenVal {
	return func() (Functor, GenVal) {
		var next = generate.Call(init)
		return init, NewGenerator(next, generate)
	}
}
func (g GenVal) Cons(Functor) Applicative          { return g }
func (g GenVal) ConsGroup(Applicative) Applicative { return g }
func (g GenVal) Expr() Functor {
	var expr, _ = g()
	return expr
}
func (g GenVal) Generator() GenVal {
	var _, gen = g()
	return gen
}
func (g GenVal) Call(args ...Functor) Functor {
	if len(args) > 0 {
		return NewPair(g.Expr().Call(args...), g.Generator())
	}
	return NewPair(g.Expr(), g.Generator())
}
func (g GenVal) TypeFnc() TyFnc  { return Generator }
func (g GenVal) Type() TyDef     { return Def(Generator, g.Head().Type()) }
func (g GenVal) TypeElem() TyDef { return g.Head().Type() }
func (g GenVal) String() string  { return g.Head().String() }
func (g GenVal) Empty() bool {
	if g.Head().Type().Match(None) {
		return true
	}
	return false
}
func (g GenVal) Concat(grp Sequential) Applicative { return NewListFromGroup(g).Concat(grp) }
func (g GenVal) Continue() (Functor, Applicative)  { return g() }
func (g GenVal) Head() Functor                     { return g.Expr() }
func (g GenVal) Tail() Applicative                 { return g.Generator() }

///////////////////////////////////////////////////////////////////////////////
//// ACCUMULATOR
///
// accumulator expects an expression as input, that returns itself unboxed,
// when called empty and returns a new accumulator accumulating its value and
// arguments to create a new accumulator, if arguments where passed.
func NewAccumulator(
	acc Functor,
	fnc func(acc Functor, args ...Functor) Functor,
) AccVal {
	return AccVal(func(args ...Functor) (Functor, AccVal) {
		if len(args) > 0 {
			acc = fnc(acc, args...)
			return acc, NewAccumulator(acc, fnc)
		}
		acc = fnc(acc)
		return acc, NewAccumulator(acc, fnc)
	})
}

func (g AccVal) Concat(grp Sequential) Applicative {
	return NewListFromGroup(g).Concat(grp)
}
func (g AccVal) ConsGroup(con Applicative) Applicative {
	var args = []Functor{}
	for head, con := con.Continue(); !con.Empty(); {
		args = append(args, head)
	}
	return NewList(g(args...))
}
func (g AccVal) Cons(arg Functor) Applicative { return NewList(g(arg)) }
func (g AccVal) Result() Functor {
	var res, _ = g()
	return res
}
func (g AccVal) Accumulator() AccVal {
	var _, acc = g()
	return acc
}
func (g AccVal) Call(args ...Functor) Functor {
	if len(args) > 0 {
		var res, acc = g(args...)
		return NewPair(res, acc)
	}
	return g.Result()
}
func (g AccVal) TypeFnc() TyFnc { return Accumulator }
func (g AccVal) Type() TyDef {
	return Def(
		Accumulator,
		g.Head().Type().TypeRet(),
		g.Head().Type().TypeArgs(),
	)
}
func (g AccVal) String() string { return g.Head().String() }

func (a AccVal) Empty() bool {
	if a.Head().Type().Match(None) {
		return true
	}
	return false
}
func (g AccVal) Head() Functor                    { return g.Result() }
func (g AccVal) TypeElem() TyDef                  { return g.Head().Type() }
func (g AccVal) Tail() Applicative                { return g.Accumulator() }
func (g AccVal) Continue() (Functor, Applicative) { return g() }

///////////////////////////////////////////////////////////////////////////////
//// TAGGED, TYPE-SAFE, PARTIAL APPLYABLE FUNCTION
///
// helper function to create function type definition:
//  - identity:
//    - first typearg if its a type-symbol
//    - first typeargs type-id, if instance of comp-type
//    - first typeargs type-name, if instance of d.typed
//    - create from expressions type signature, (argtypes → idtype →
//      returntype), if no typeargs where passed
//  - return & argument types are carried over from expression
func createFuncType(expr Functor, types ...d.Typed) TyDef {
	if len(types) > 0 {
		// if first types argument is a symbol, use it as identity
		if Kind_Sym.Match(types[0].Kind()) {
			return Def(types...)
		} else {
			var (
				name  string
				ident TySym
			)
			// if first types argument is a composed type
			if Kind_Comp.Match(types[0].Kind()) {
				// return composed types identity
				name = types[0].(TyDef).
					TypeId().TypeName()
			} else {
				// return flat types name
				name = types[0].TypeName()
			}
			// create type identity from name
			ident = DefSym(name)
			// if more types arguments where passed‥.
			if len(types) > 1 {
				// return ident followed by remaining types
				// arguments forreturn and argument types
				return Def(append(
					[]d.Typed{ident},
					types[1:]...)...)
			}
		}
	}
	// no types arguments where passed, compose type identity from
	// expressions identity, return & argument types
	var (
		name = "(" +
			expr.Type().TypeArgs().TypeName() +
			" → " +
			expr.Type().TypeId().TypeName() +
			" → " +
			expr.Type().TypeRet().TypeName() +
			")"
		ident = DefSym(name)
	)
	// return type identity, take return & argument types from expression.
	return Def(ident,
		expr.Type().TypeRet(),
		expr.Type().TypeArgs())

}

// define creates and returns a tagged and type-safe data constructor, or
// function definition & signature (function type).
func Define(
	expr Functor,
	types ...d.Typed,
) FuncVal {

	// create the function type definition and take the number of expexted
	// arguments
	var (
		ft     = createFuncType(expr, types...)
		arglen = ft.TypeArgs().Len()
	)

	// return partialy applable function
	return func(args ...Functor) Functor {

		// take number of passed arguments
		var length = len(args)

		/////////////////////////
		// NO ARGUMENTS PASSED →
		if length == 0 {
			return ft
		}

		// test if arguments passed match argument types
		if ft.TypeArgs().MatchArgs(args...) {

			// switch based on number of passed arguments relative
			// to number of arguments expected by the type
			// definition
			switch {
			////////////////////////////////////////////////
			// NUMBER OF PASSED ARGUMENTS MATCHES EXACTLY →
			case length == arglen:
				// return result of calling the enclosed
				// expression passing the arguments
				return expr.Call(args...)

			/////////////////////////////////////////////
			// NUMBER OF PASSED ARGUMENTS INSUFFICIENT →
			case length < arglen:

				// safe types of arguments remaining to be
				// filled
				var (
					remains = ft.TypeArgs().Types()[length:]
					newpat  = Def(Def(Partial, ft.TypeId()),
						ft.TypeRet(),
						Def(remains...))
				)

				// define partial function from remaining set
				// of argument types, that encloses all
				// arguments that have been passed and
				// validated so far, appends arguments passed
				// in later calls and returns the result from
				// applying them to the enclosed function.
				return Define(Lambda(func(lateargs ...Functor) Functor {
					// will return result, or yet another
					// partial, depending on arguments
					if len(lateargs) > 0 {
						return expr.Call(append(
							args, lateargs...,
						)...)
					}
					// if no arguments where passed, return
					// the redefined partial type
					return newpat
				}), newpat.Types()...)

			//////////////////////////////////////////////
			// NUMBER OF PASSED ARGUMENTS OVERSATURATED →
			case length > arglen:

				// allocate a vector to hold multiple instances
				// of return type
				var vector = NewVector()

				// iterate over passed arguments, allocate one
				// instance of defined type per satisfying set
				// of arguments
				for len(args) > arglen {
					vector = vector.Cons(expr.Call(
						args[:arglen]...)).(VecVal)
					args = args[arglen:]
				}

				// check for leftover arguments that don't
				// satisfy the definition, and possibly create
				// and return a partial as last element in the
				// slice of return values, when such exist.
				if length > 0 {
					// add a partial expression as vectors
					// last element
					vector = vector.Cons(Define(
						expr, ft.Types()...,
					).Call(args...)).(VecVal)
				}

				// return vector of instances
				return vector
			}
		}
		////////////////////////////////////////////
		// ARGUMENT TYPES DO NOT MATCH DEFINITION →
		return None
	}
}

func (e FuncVal) Call(args ...Functor) Functor { return e(args...) }

func (e FuncVal) TypeFnc() TyFnc   { return Value }
func (e FuncVal) Type() TyDef      { return e().(TyDef) }
func (e FuncVal) TypeId() TyDef    { return e.Type().TypeId() }
func (e FuncVal) TypeArgs() TyDef  { return e.Type().TypeArgs() }
func (e FuncVal) TypeRet() TyDef   { return e.Type().TypeRet() }
func (e FuncVal) TypeName() string { return e.Type().TypeName() }
func (e FuncVal) ArgCount() int    { return e.Type().TypeArgs().Count() }
func (e FuncVal) String() string   { return e().String() }
