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
	DeclaredExpr func(...Expression) Expression
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
func (c ConstantExpr) FlagType() d.Uint8Val               { return Flag_Function.U() }
func (c ConstantExpr) String() string                     { return c().String() }
func (c ConstantExpr) ElemType() Typed                    { return c().Type() }
func (c ConstantExpr) TypeName() string                   { return c().TypeName() }
func (c ConstantExpr) Type() Typed {
	return Define("ϝ → "+c().TypeName(), c().Type())
}

//// GENERIC EXPRESSION VALUE CONSTRUCTOR ////
///
// generic expression constructor takes an expression, name, returntype and
// parameter types, creates a type definition and returns a wrapper returning
// the type definition, when no arguments are passed
func NewGeneric(
	expr func(...Expression) Expression,
	name string,
	retype Typed,
	paratypes ...Typed,
) GenericExpr {

	var typed = Define(name, retype, paratypes...)

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
func (c GenericExpr) TypeFnc() TyFnc                     { return c().TypeFnc() }
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
func Declare(
	expr Expression,
	name string,
	retype Typed,
	paratypes ...Typed,
) DeclaredExpr {
	var typed = Define(name, retype, paratypes...)
	return func(args ...Expression) Expression {
		var arglen = Arity(len(args))
		if arglen == typed.Arity() {
		}
		return NewPair(typed, expr)
	}
}

// returns the value returned when calling itself directly, passing arguments
func (n DeclaredExpr) Ident() Expression    { return n }
func (n DeclaredExpr) String() string       { return n().String() }
func (n DeclaredExpr) TypeName() string     { return n.TypeDef().TypeName() }
func (n DeclaredExpr) FlagType() d.Uint8Val { return Flag_DataCons.U() }
func (n DeclaredExpr) Expr() Expression {
	return n().(Paired).Right().(Expression)
}
func (n DeclaredExpr) TypeDef() TyDef                     { return n().(Paired).Left().(TyDef) }
func (n DeclaredExpr) Type() Typed                        { return n.TypeDef() }
func (n DeclaredExpr) TypeFnc() TyFnc                     { return n.TypeDef().TypeFnc() }
func (n DeclaredExpr) TypeNat() d.TyNat                   { return n.TypeDef().TypeNat() }
func (n DeclaredExpr) Arity() Arity                       { return n.TypeDef().Arity() }
func (n DeclaredExpr) Return() Typed                      { return n.TypeDef().Return() }
func (n DeclaredExpr) Pattern() []Typed                   { return n.TypeDef().Arguments() }
func (n DeclaredExpr) Call(args ...Expression) Expression { return n(args...) }
func (n DeclaredExpr) Eval(args ...d.Native) d.Native {
	if n.TypeFnc().Match(Data) {
		if data, ok := n.Expr().(Native); ok {
			if len(args) > 0 {
				return data.Eval(args...)
			}
			return data.Eval()
		}
	}
	return d.NewNil()
}
