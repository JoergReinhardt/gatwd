package functions

import (
	d "github.com/JoergReinhardt/gatwd/data"
)

type (
	// GENERIC FUNCTION DEFINITION
	ConstFnc  func() Value
	UnaryFnc  func(Value) Value
	BinaryFnc func(a, b Value) Value
	NaryFnc   func(...Value) Value

	// OPTIONAL TYPES
	//
	// JUST/NONE
	NoneVal func()       // None and Just are a pair of optional types
	JustVal func() Value // implementing 'Option :: Maybe() bool'
	// EITHER/OR
	EitherVal func() Value // None and Just are a pair of optional types
	OrVal     func() Value // implementing 'Option :: Maybe() bool'
	// TRUE/FALSE
	TrueVal  func() Boolean // boolean constants true & false
	FalseVal func() Boolean // implementing 'Boolen :: Bool() bool'

	// OPTIONAL & CONDITIONAL
	PraedFnc    func(Value) bool // result impl. Bool() bool
	OptionalFnc func() Optional
)

// CONSTANT
//
// constant also conains immutable data that may be an instance of a type of
// the data package, or result of a function call guarantueed to allways return
// the same value.
func NewConstant(fnc func() Value) ConstFnc {
	return ConstFnc(func() Value { return fnc() })
}

func (c ConstFnc) Ident() Value                { return c() }
func (c ConstFnc) TypeFnc() TyFnc              { return Function }
func (c ConstFnc) TypeNat() d.TyNative         { return c().TypeNat() }
func (c ConstFnc) Eval(p ...d.Native) d.Native { return c().Eval() }
func (c ConstFnc) Call(d ...Value) Value       { return c() }

///// UNARY FUNCTION
func NewUnaryFnc(fnc func(f Value) Value) UnaryFnc {
	return UnaryFnc(func(f Value) Value { return fnc(f) })
}
func (u UnaryFnc) TypeNat() d.TyNative         { return d.Function.TypeNat() }
func (u UnaryFnc) TypeFnc() TyFnc              { return Function }
func (u UnaryFnc) Ident() Value                { return u }
func (u UnaryFnc) Eval(p ...d.Native) d.Native { return u }
func (u UnaryFnc) Call(d ...Value) Value {
	return u(d[0])
}

///// BINARY FUNCTION
func NewBinaryFnc(fnc func(a, b Value) Value) BinaryFnc {
	return BinaryFnc(func(a, b Value) Value { return fnc(a, b) })
}
func (b BinaryFnc) TypeNat() d.TyNative         { return d.Function.TypeNat() }
func (b BinaryFnc) TypeFnc() TyFnc              { return Function }
func (b BinaryFnc) Ident() Value                { return b }
func (b BinaryFnc) Eval(p ...d.Native) d.Native { return b }
func (b BinaryFnc) Call(d ...Value) Value       { return b(d[0], d[1]) }

///// NARY FUNCTION
func NewNaryFnc(fnc func(f ...Value) Value) NaryFnc {
	return NaryFnc(func(f ...Value) Value { return fnc(f...) })
}
func (n NaryFnc) TypeNat() d.TyNative         { return d.Function.TypeNat() }
func (n NaryFnc) TypeFnc() TyFnc              { return Function }
func (n NaryFnc) Ident() Value                { return n }
func (n NaryFnc) Eval(p ...d.Native) d.Native { return n }
func (n NaryFnc) Call(d ...Value) Value       { return n(d...) }

/////////////////////////////////////////////////////////////////////////////
///// OPTIONAL TYPE
////
/// RETURN TYPE PAIRS OF THE OPTIONAL TYPE
//
// EITHER
func NewEitherVal(v Value) EitherVal {
	return EitherVal(func() Value { return v })
}
func (e EitherVal) Ident() Value                { return e }
func (e EitherVal) Call(v ...Value) Value       { return e.Call(v...) }
func (e EitherVal) Eval(p ...d.Native) d.Native { return e().Eval(p...) }
func (e EitherVal) Maybe() bool                 { return true }
func (e EitherVal) Value() Value                { return e() }
func (e EitherVal) Nullable() d.Native          { return e.Eval() }
func (e EitherVal) TypeFnc() TyFnc              { return EitherOr | Either }
func (e EitherVal) TypeNat() d.TyNative         { return e().TypeNat() }
func (e EitherVal) String() string              { return e().String() }

// OR
func NewOrVal(v Value) OrVal {
	return OrVal(func() Value { return v })
}
func (o OrVal) Ident() Value                { return o }
func (o OrVal) Call(v ...Value) Value       { return o.Call(v...) }
func (o OrVal) Eval(p ...d.Native) d.Native { return o().Eval(p...) }
func (o OrVal) Maybe() bool                 { return false }
func (o OrVal) Value() Value                { return o() }
func (o OrVal) Nullable() d.Native          { return o.Eval() }
func (o OrVal) TypeFnc() TyFnc              { return EitherOr | Or }
func (o OrVal) TypeNat() d.TyNative         { return o().TypeNat() }
func (o OrVal) String() string              { return o().String() }

// JUST
func NewJustVal(v Value) JustVal {
	return JustVal(func() Value { return v })
}
func (j JustVal) Ident() Value                { return j }
func (j JustVal) Call(v ...Value) Value       { return j.Call(v...) }
func (j JustVal) Eval(p ...d.Native) d.Native { return j().Eval(p...) }
func (j JustVal) Maybe() bool                 { return true }
func (j JustVal) Value() Value                { return j() }
func (j JustVal) Nullable() d.Native          { return j.Eval() }
func (j JustVal) TypeFnc() TyFnc              { return Option | Just }
func (j JustVal) TypeNat() d.TyNative         { return j().TypeNat() }
func (j JustVal) String() string              { return j().String() }

// NONE
func NewNone() NoneVal                      { return NoneVal(func() {}) }
func (n NoneVal) Ident() Value              { return n }
func (n NoneVal) Call(...Value) Value       { return n }
func (n NoneVal) Eval(...d.Native) d.Native { return d.NilVal{}.Eval() }
func (n NoneVal) Maybe() bool               { return false }
func (n NoneVal) Value() Value              { return NewNone() }
func (n NoneVal) Nullable() d.Native        { return d.NilVal{} }
func (n NoneVal) TypeFnc() TyFnc            { return Option | None }
func (n NoneVal) TypeNat() d.TyNative       { return d.Nil }
func (n NoneVal) String() string            { return "⊥" }

//// FUNCTIONAL TRUTH VALUES
///
// TRUE
func (t TrueVal) Call(...Value) Value {
	return New(d.BoolVal(true))
}
func (t TrueVal) Ident() Value              { return t }
func (t TrueVal) Eval(...d.Native) d.Native { return t }
func (t TrueVal) Bool() bool                { return true }
func (t TrueVal) Value() Value              { return t }
func (t TrueVal) TypeFnc() TyFnc            { return Truth | True }
func (t TrueVal) TypeNat() d.TyNative       { return d.Bool }
func (t TrueVal) String() string            { return "True" }

// FALSE
func (f FalseVal) Call(...Value) Value {
	return New(d.BoolVal(false))
}
func (f FalseVal) Iwdent() Value             { return f }
func (f FalseVal) Eval(...d.Native) d.Native { return f }
func (f FalseVal) Bool() bool                { return false }
func (f FalseVal) Value() Value              { return f }
func (f FalseVal) TypeFnc() TyFnc            { return Truth | False }
func (f FalseVal) TypeNat() d.TyNative       { return d.Bool }
func (f FalseVal) String() string            { return "False" }

/// PRAEDICATE
//
// encloses a test that expects some type of input value to test against and
// returns either true, or false.
func NewPraedicate(pred func(scrut Value) bool) PraedFnc { return PraedFnc(pred) }
func (p PraedFnc) TypeFnc() TyFnc                        { return Predicate }
func (p PraedFnc) TypeNat() d.TyNative                   { return d.Bool }
func (p PraedFnc) Ident() Value                          { return p }
func (p PraedFnc) String() string {
	return "T → λpredicate  → Bool"
}
func (p PraedFnc) Eval(dat ...d.Native) d.Native {
	if len(dat) > 0 {
		return d.BoolVal(p(New(dat[0])))
	}
	return d.NilVal{}
}
func (p PraedFnc) Call(v ...Value) Value {
	if len(v) > 0 {
		p(v[0])
	}
	return NewNone()
}
