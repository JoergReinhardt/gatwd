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

	//// NATIVE VALUE CONSTRUCTORS
	NativeConst func() d.Native
	NativeExpr  func(...Native) d.Native
	NativeCol   func(...Native) d.Sliceable
	NativeUbox  func(...Native) d.Sliceable
	NativePair  func(...Native) d.Paired
	NativeSet   func(...Native) d.Mapped

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
func (c ConstantExpr) Eval() d.Native                     { return native(c) }
func (c ConstantExpr) FlagType() d.Uint8Val               { return Flag_Functional.U() }
func (c ConstantExpr) TypeName() string                   { return c().TypeName() }
func (c ConstantExpr) Type() TyDef {
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

	var params = make([]Typed, 0, len(paratypes))
	for _, param := range paratypes {
		params = append(params, param.Type())
	}
	var typed = Define(name, retype.Type(), params...)

	return func(args ...Expression) Expression {
		if len(args) > 0 {
			return expr(args...)
		}
		return typed
	}
}

func (c GenericExpr) Ident() Expression                  { return c }
func (c GenericExpr) Type() TyDef                        { return c().(TyDef) }
func (c GenericExpr) String() string                     { return c().String() }
func (c GenericExpr) TypeName() string                   { return c().TypeName() }
func (c GenericExpr) FlagType() d.Uint8Val               { return Flag_Functional.U() }
func (c GenericExpr) TypeFnc() TyFnc                     { return c.Type().Return().TypeFnc() }
func (c GenericExpr) Call(args ...Expression) Expression { return c(args...) }
func (c GenericExpr) Eval() d.Native                     { return native(c) }

//// NATIVE EXPRESSION CONSTRUCTOR
///
// returns an expression with native return type implementing the callable
// interface
func New(inf ...interface{}) Expression {
	return NewData(d.New(inf...))
}

func NewData(args ...Native) Native {

	var nats = make([]d.Native, 0, len(args))
	for _, arg := range args {
		nats = append(nats, arg)
	}
	var nat = d.NewData(nats...)
	var match = nat.TypeNat().Match

	switch {
	case match(d.Slice):
		return NativeCol(func(args ...Native) d.Sliceable {
			return nat.(d.Sliceable)
		})
	case match(d.Unboxed):
		return NativeUbox(func(args ...Native) d.Sliceable {
			return nat.(d.Sliceable)
		})
	case match(d.Pair):
		return NativePair(func(args ...Native) d.Paired {
			return nat.(d.Paired)
		})
	case match(d.Map):
		return NativeSet(func(args ...Native) d.Mapped {
			return nat.(d.Mapped)
		})
	}
	return NativeConst(func() d.Native {
		return nat
	})
}

func native(args ...Expression) d.Native {
	var nat d.Native
	if len(args) == 0 {
		return d.NewNil()
	}
	var nats = make([]Native, 0, len(args))
	for _, arg := range args {
		nats = append(nats, NewData(arg))
	}
	return NewData(nats...).(Native)
	return nat
}

// ATOMIC NATIVE VALUE CONSTRUCTOR
func (n NativeConst) Call(...Expression) Expression { return n }
func (n NativeConst) Eval() d.Native                { return n() }
func (n NativeConst) TypeFnc() TyFnc                { return Data }
func (n NativeConst) TypeNat() d.TyNat              { return n().TypeNat() }
func (n NativeConst) FlagType() d.Uint8Val          { return Flag_Functional.U() }
func (n NativeConst) String() string                { return n().String() }
func (n NativeConst) TypeName() string              { return n().TypeName() }
func (n NativeConst) Type() TyDef {
	return Define(n().TypeNat().TypeName(), New(n().TypeNat()))
}

func (n NativeExpr) Call(args ...Expression) Expression {
	var nats = make([]Native, 0, len(args))
	for _, arg := range args {
		nats = append(nats, arg.(Native).Eval())
	}
	return NewData(n(nats...))
}
func (n NativeExpr) Eval() d.Native       { return n() }
func (n NativeExpr) TypeFnc() TyFnc       { return Data }
func (n NativeExpr) TypeNat() d.TyNat     { return n().TypeNat() }
func (n NativeExpr) FlagType() d.Uint8Val { return Flag_Functional.U() }
func (n NativeExpr) String() string       { return n().String() }
func (n NativeExpr) TypeName() string     { return n().TypeName() }
func (n NativeExpr) Type() TyDef {
	return Define(n().TypeNat().TypeName(), New(n().TypeNat()))
}

// NATIVE SLICE VALUE CONSTRUCTOR
func (n NativeCol) Call(args ...Expression) Expression {
	return NewData(n(native(args...)))
}
func (n NativeCol) Len() int                   { return n().Len() }
func (n NativeCol) TypeFnc() TyFnc             { return Data }
func (n NativeCol) Eval() d.Native             { return n() }
func (n NativeCol) Sequential() d.Sequential   { return n().(d.DataSlice) }
func (n NativeCol) Head() d.Native             { return n.Sequential().Head() }
func (n NativeCol) Tail() d.Sequential         { return n.Sequential().Tail() }
func (n NativeCol) Shift() d.Sequential        { return n.Sequential().Shift() }
func (n NativeCol) SliceNat() []d.Native       { return n().Slice() }
func (n NativeCol) Get(key d.Native) d.Native  { return n().Get(key) }
func (n NativeCol) GetInt(idx int) d.Native    { return n().GetInt(idx) }
func (n NativeCol) Range(s, e int) d.Sliceable { return n().Range(s, e) }
func (n NativeCol) Empty() bool                { return n().Empty() }
func (n NativeCol) Copy() d.Native             { return n().Copy() }
func (n NativeCol) TypeNat() d.TyNat           { return n().TypeNat() }
func (n NativeCol) ElemType() d.TyNat          { return n().ElemType() }
func (n NativeCol) String() string             { return n().String() }
func (n NativeCol) TypeName() string           { return n().TypeName() }
func (n NativeCol) FlagType() d.Uint8Val       { return Flag_Functional.U() }
func (n NativeCol) Slice() []d.Native          { return n().Slice() }
func (n NativeCol) Type() TyDef                { return Define(n().TypeName(), NewData(n.TypeNat())) }
func (n NativeCol) SliceExpr() []Expression {
	var slice = make([]Expression, 0, n.Len())
	for _, nat := range n.Slice() {
		slice = append(slice, NewData(nat))
	}
	return slice
}

func (n NativeUbox) Call(args ...Expression) Expression {
	return NewData(n(native(args...)))
}
func (n NativeUbox) TypeFnc() TyFnc             { return Data }
func (n NativeUbox) Eval() d.Native             { return n() }
func (n NativeUbox) Len() int                   { return n().Len() }
func (n NativeUbox) Get(key d.Native) d.Native  { return n().Get(key) }
func (n NativeUbox) GetInt(idx int) d.Native    { return n().GetInt(idx) }
func (n NativeUbox) Range(s, e int) d.Sliceable { return n().Range(s, e) }
func (n NativeUbox) Copy() d.Native             { return n().Copy() }
func (n NativeUbox) Empty() bool                { return n().Empty() }
func (n NativeUbox) Slice() []d.Native          { return n().Slice() }
func (n NativeUbox) TypeNat() d.TyNat           { return n().TypeNat() }
func (n NativeUbox) ElemType() d.TyNat          { return n().ElemType() }
func (n NativeUbox) TypeName() string           { return n().TypeName() }
func (n NativeUbox) FlagType() d.Uint8Val       { return Flag_Functional.U() }
func (n NativeUbox) String() string             { return n().String() }
func (n NativeUbox) Type() TyDef                { return Define(n.Eval().TypeName(), NewData(n.TypeNat())) }
func (n NativeUbox) SliceExpr() []Expression {
	var slice = make([]Expression, 0, n.Len())
	for _, nat := range n.Slice() {
		slice = append(slice, NewData(nat))
	}
	return slice
}

// NATIVE PAIR VALUE CONSTRUCTOR
func (n NativePair) Call(args ...Expression) Expression {
	return NewData(n(native(args...)))
}
func (n NativePair) TypeFnc() TyFnc        { return Data }
func (n NativePair) TypeNat() d.TyNat      { return n().TypeNat() }
func (n NativePair) Eval() d.Native        { return n() }
func (n NativePair) Left() d.Native        { return n().Left() }
func (n NativePair) Right() d.Native       { return n().Right() }
func (n NativePair) Both() (l, r d.Native) { return n().Both() }
func (n NativePair) LeftType() d.TyNat     { return n().LeftType() }
func (n NativePair) RightType() d.TyNat    { return n().RightType() }
func (n NativePair) SubType() d.Typed      { return n().TypeNat() }
func (n NativePair) TypeName() string      { return n().TypeName() }
func (n NativePair) FlagType() d.Uint8Val  { return Flag_Functional.U() }
func (n NativePair) String() string        { return n().String() }
func (n NativePair) Type() TyDef {
	return Define(n().TypeName(), NewData(n().TypeNat()))
}
func (n NativePair) Pair() Paired {
	return NewPair(
		NewData(n().Left()),
		NewData(n().Right()))
}
func (n NativePair) LeftExpr() Expression  { return NewData(n().Left()) }
func (n NativePair) RightExpr() Expression { return NewData(n().Right()) }
func (n NativePair) BothExpr() (l, r Expression) {
	return NewData(n().Left()),
		NewData(n().Right())
}

// NATIVE SET VALUE CONSTRUCTOR

func (n NativeSet) Call(args ...Expression) Expression {
	return NewData(n(native(args...)))
}
func (n NativeSet) Ident() Expression                    { return n }
func (n NativeSet) Eval() d.Native                       { return n() }
func (n NativeSet) TypeFnc() TyFnc                       { return Data }
func (n NativeSet) TypeNat() d.TyNat                     { return n().TypeNat() }
func (n NativeSet) Len() int                             { return n().Len() }
func (n NativeSet) Slice() []d.Native                    { return n().Slice() }
func (n NativeSet) GetNat(acc d.Native) (d.Native, bool) { return n().Get(acc) }
func (n NativeSet) SetNat(acc, val d.Native) d.Mapped    { return n().Set(acc, val) }
func (n NativeSet) Delete(acc d.Native) bool             { return n().Delete(acc) }
func (n NativeSet) Get(acc d.Native) (d.Native, bool)    { return n().Get(acc) }
func (n NativeSet) Set(acc, val d.Native) d.Mapped       { return n().Set(acc, val) }
func (n NativeSet) Keys() []d.Native                     { return n().Keys() }
func (n NativeSet) Data() []d.Native                     { return n().Data() }
func (n NativeSet) Fields() []d.Paired                   { return n().Fields() }
func (n NativeSet) KeyType() d.TyNat                     { return n().KeyType() }
func (n NativeSet) ValType() d.TyNat                     { return n().ValType() }
func (n NativeSet) SubType() d.Typed                     { return n().TypeNat() }
func (n NativeSet) TypeName() string                     { return n().TypeName() }
func (n NativeSet) FlagType() d.Uint8Val                 { return Flag_Functional.U() }
func (n NativeSet) String() string                       { return n().String() }
func (n NativeSet) Type() TyDef {
	return Define(n().TypeName(), NewData(n()))
}
func (n NativeSet) KeysExpr() []Expression {
	var exprs = make([]Expression, 0, n.Len())
	for _, key := range n().Keys() {
		exprs = append(exprs, NewData(key))
	}
	return exprs
}
func (n NativeSet) DataExpr() []Expression {
	var exprs = make([]Expression, 0, n.Len())
	for _, val := range n().Data() {
		exprs = append(exprs, NewData(val))
	}
	return exprs
}
func (n NativeSet) SliceExpr() []Expression {
	var slice = make([]Expression, 0, n.Len())
	for _, nat := range n.Fields() {
		slice = append(slice, NewData(nat))
	}
	return slice
}
func (n NativeSet) Pairs() []Paired {
	var pairs = []Paired{}
	for _, field := range n.Fields() {
		pairs = append(
			pairs, NewPair(
				NewData(field.Left()),
				NewData(field.Right())))
	}
	return pairs
}

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

	var params = make([]Typed, 0, arity)
	for _, param := range paratypes {
		params = append(params, param.Type())
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
func (n PartialExpr) Ident() Expression                  { return n }
func (n PartialExpr) Type() TyDef                        { return n().(TyDef) }
func (n PartialExpr) String() string                     { return n.TypeName() }
func (n PartialExpr) TypeName() string                   { return n.Type().Name() }
func (n PartialExpr) FlagType() d.Uint8Val               { return Flag_DataCons.U() }
func (n PartialExpr) Arity() Arity                       { return n.Type().Arity() }
func (n PartialExpr) Return() Typed                      { return n.Type().Return() }
func (n PartialExpr) Pattern() []Typed                   { return n.Type().Pattern() }
func (n PartialExpr) TypeFnc() TyFnc                     { return n.Return().TypeFnc() }
func (n PartialExpr) Call(args ...Expression) Expression { return n(args...) }
