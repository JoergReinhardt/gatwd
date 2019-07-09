/*
  FUNCTIONAL CONTAINERS

  containers implement enumeration of functional types, aka lists, vectors
  sets, pairs, tuples‥.
*/
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// CONSTANT & GENERIC EXPRESSIONS
	ConstantExpr func() Expression
	GenericExpr  func(...Expression) Expression

	//// EXPRESSION VALUE CONSTRUCTOR
	PartialExpr func(...Expression) Expression
)

//// CONSTANT VALUE CONSTRUCTOR
///
// constant expression constructor takes a generic function returning a value
// of expression type and takes its methods from that value.
func NewConstant(constant func() Expression) ConstantExpr { return constant }

func (c ConstantExpr) Ident() Expression                  { return c }
func (c ConstantExpr) Call(args ...Expression) Expression { return c() }
func (c ConstantExpr) Arity() Arity                       { return Arity(0) }
func (c ConstantExpr) TypeFnc() TyFnc                     { return Constant }
func (c ConstantExpr) String() string                     { return c().String() }
func (c ConstantExpr) FlagType() d.Uint8Val               { return Flag_Function.U() }
func (c ConstantExpr) TypeName() string                   { return c().TypeName() }
func (c ConstantExpr) Type() Typed {
	return Define("ϝ → "+c().TypeName(), c())
}

//// GENERIC EXPRESSION VALUE CONSTRUCTOR ////
///
// generic expression constructor takes an expression, name, returntype and
// parameter types, creates a type definition and returns a wrapper returning
// the type definition, when no arguments are passed
func NewGeneric(
	expr func(...Expression) Expression,
	name string,
	retype Expression,
	paratypes ...Expression,
) GenericExpr {

	var params = make([]Expression, 0, len(paratypes))
	for _, param := range paratypes {
		params = append(params, param)
	}
	var typed = Define(name, retype.TypeFnc(), params...)

	return func(args ...Expression) Expression {
		if len(args) > 0 {
			return expr(args...)
		}
		return typed
	}
}

func (c GenericExpr) Ident() Expression                  { return c }
func (c GenericExpr) Type() Typed                        { return c().(Typed) }
func (c GenericExpr) String() string                     { return c().String() }
func (c GenericExpr) TypeName() string                   { return c().TypeName() }
func (c GenericExpr) FlagType() d.Uint8Val               { return Flag_Function.U() }
func (c GenericExpr) TypeFnc() TyFnc                     { return c.Type().(TyDef).Return().TypeFnc() }
func (c GenericExpr) Call(args ...Expression) Expression { return c(args...) }

//// EXPRESSION TYPE CONSTRUCTOR
///
// TODO: make nary type safe by deriving type switch from signature and
// exposing it in a match method
//
// expression type definition takes an optional name, an expression and a
// number of expressions, or typed definitions to declare the expression
// signature. last signature expression is assumed to be the return type. all
// signature arguments before that are assumed to be the arguments types.
//
// if no signature is passed, return type is derived from expression. if no
// signature, or only return type are passed, argument types are assumed to be
// parametric matching all types.
//
// defined expressions can are enumerated and partialy applyable.
func DefinePartial(
	name string,
	expr Expression,
	retype Expression,
	paratypes ...Expression,
) PartialExpr {

	var arity = len(paratypes)

	var params = make([]Expression, 0, arity)
	for _, param := range paratypes {
		params = append(params, param.TypeFnc())
	}
	var typed = Define(name, retype, params...)

	// create and return nary expression
	return func(args ...Expression) Expression {

		var parmlen = len(args) // count arguments

		if parmlen > 0 { // if arguments where passed
			// argument number SATISFIES expression arity EXACTLY
			if parmlen == arity {
				return expr.Call(args...)
			}
			// argument number UNDERSATISFIES expression arity
			if parmlen < arity {
				return DefinePartial(name, PartialExpr(
					func(lateargs ...Expression) Expression {
						return expr.Call(append(lateargs,
							args...)...)
					}), retype, paratypes[parmlen:]...)
			}
			// argument number OVERSATISFIES expressions arity
			if parmlen > arity {
				var remain []Expression
				args, remain = args[:arity], args[arity:]
				var vec = NewVector(expr.Call(args...))
				for len(remain) > arity {
					args, remain = remain[:arity], remain[arity:]
					vec = vec.Append(expr.Call(args...))
				}
				return vec.Append(expr.Call(remain...))
			}
		}
		// if no arguments are passed, return definition
		return typed
	}
}

// returns the value returned when calling itself directly, passing arguments
func (n PartialExpr) Ident() Expression     { return n }
func (n PartialExpr) Type() Typed           { return n().(Typed) }
func (n PartialExpr) String() string        { return n.TypeName() }
func (n PartialExpr) TypeName() string      { return n.Type().(TyDef).Name() }
func (n PartialExpr) FlagType() d.Uint8Val  { return Flag_DataCons.U() }
func (n PartialExpr) Arity() Arity          { return n.Type().(TyDef).Arity() }
func (n PartialExpr) Return() Expression    { return n.Type().(TyDef).Return() }
func (n PartialExpr) Pattern() []Expression { return n.Type().(TyDef).Pattern() }
func (n PartialExpr) TypeFnc() TyFnc {
	return n.Return().(Expression).TypeFnc()
}
func (n PartialExpr) Call(args ...Expression) Expression { return n(args...) }
