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

	// OPTIONAL & CONDITIONAL
	PraedFnc     func(Value) bool // result impl. Bool() bool
	OptionalFnc  func(Value) (Value, bool)
	ConditionFnc func() (a, b Optional)
	EitherFnc    func(Value) Value
	CaseCond     func(Value) CaseCondition
	IfCond       func(Value) Conditional
	ElseCond     func(Conditional, Value) Optional
	//
	NoneVal func()       // None and Just are a pair of optional types
	JustVal func() Value // implementing 'Option :: Maybe() bool'
	//
	TrueVal  func() Boolean // boolean constants true & false
	FalseVal func() Boolean // implementing 'Boolen :: Bool() bool'
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

//// RETURN TYPES OF THE OPTIONAL TYPE
///
// NONE
func NewNone() NoneVal                      { return NoneVal(func() {}) }
func (n NoneVal) Ident() Value              { return n }
func (n NoneVal) Call(...Value) Value       { return n }
func (n NoneVal) Eval(...d.Native) d.Native { return d.NilVal{}.Eval() }
func (n NoneVal) Maybe() bool               { return false }
func (n NoneVal) Nullable() d.Native        { return d.NilVal{} }
func (n NoneVal) TypeFnc() TyFnc            { return Option | None }
func (n NoneVal) TypeNat() d.TyNative       { return d.Nil }
func (n NoneVal) String() string            { return "⊥" }

// JUST
func NewJustVal(v Value) JustVal {
	return JustVal(func() Value { return v })
}
func (j JustVal) Ident() Value                { return j }
func (j JustVal) Call(...Value) Value         { return j }
func (j JustVal) Eval(p ...d.Native) d.Native { return j().Eval(p...) }
func (j JustVal) Maybe() bool                 { return true }
func (j JustVal) Nullable() d.Native          { return j.Eval() }
func (j JustVal) TypeFnc() TyFnc              { return Option | Just }
func (j JustVal) TypeNat() d.TyNative         { return j().TypeNat() }
func (j JustVal) String() string              { return j().String() }

// FUNCTIONAL TRUTH VALUES
func (t TrueVal) Call(...Value) Value {
	return New(d.BoolVal(true))
}
func (t TrueVal) Ident() Value              { return t }
func (t TrueVal) Eval(...d.Native) d.Native { return t }
func (t TrueVal) Bool() bool                { return true }
func (t TrueVal) TypeFnc() TyFnc            { return Truth | True }
func (t TrueVal) TypeNat() d.TyNative       { return d.Bool }
func (t TrueVal) String() string            { return "True" }

func (f FalseVal) Call(...Value) Value {
	return New(d.BoolVal(false))
}
func (f FalseVal) Iwdent() Value             { return f }
func (f FalseVal) Eval(...d.Native) d.Native { return f }
func (f FalseVal) Bool() bool                { return false }
func (f FalseVal) TypeFnc() TyFnc            { return Truth | False }
func (f FalseVal) TypeNat() d.TyNative       { return d.Bool }
func (f FalseVal) String() string            { return "False" }

// PRAEDICATE
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

// OPTIONAL
func NewOptionalVal(praed PraedFnc) OptionalFnc {
	return OptionalFnc(func(scrut Value) (Value, bool) {
		if praed(scrut) {
			return scrut, true
		}
		return NewNone(), false
	})
}
func (o OptionalFnc) Ident() Value            { return o }
func (o OptionalFnc) Maybe(scrut Value) bool  { _, ok := o(scrut); return ok }
func (o OptionalFnc) Value(scrut Value) Value { val, _ := o(scrut); return val }
func (o OptionalFnc) Nullable() d.Native      { return d.NilVal{} }
func (o OptionalFnc) TypeFnc() TyFnc          { return Option }
func (o OptionalFnc) TypeNat() d.TyNative     { return d.Booleans.TypeNat() }
func (o OptionalFnc) String() string {
	var str string
	return str
}
func (o OptionalFnc) Call(v ...Value) Value {
	if len(v) > 0 {
		val, ok := o(v[0])
		if ok {
			return val
		}
	}
	return NewNone()
}
func (o OptionalFnc) Eval(p ...d.Native) d.Native {
	if len(p) > 0 {
		val, ok := o(NewFromData(p[0]))
		if ok {
			return val.Eval()
		}
	}
	return d.NilVal{}
}

/// Conditional
func NewCondition(oa, ob Optional) ConditionFnc {
	return ConditionFnc(func() (a, b Optional) {
		return oa, ob
	})
}
func (e ConditionFnc) Ident() Value              { return e }
func (e ConditionFnc) Call(...Value) Value       { return e }
func (e ConditionFnc) Eval(...d.Native) d.Native { return d.NilVal{}.Eval() }
func (e ConditionFnc) Maybe(scrut Value) bool {
	var l, _ = e()
	var _, ok = l.(OptionalFnc)(scrut)
	return ok
}
func (e ConditionFnc) Nullable() d.Native  { return d.NilVal{} }
func (e ConditionFnc) TypeFnc() TyFnc      { return Option | Condition }
func (e ConditionFnc) TypeNat() d.TyNative { return d.Nil }
func (e ConditionFnc) String() string {
	var str string
	return str
}

// EITHER
func NewEither(oa, ob OptionalFnc) EitherFnc {
	return EitherFnc(func(scrut Value) Value {
		if val, ok := oa(scrut); ok {
			return val
		}
		if val, ok := ob(scrut); ok {
			return val
		}
		return NewNone()
	})
}
func (e EitherFnc) Ident() Value              { return e }
func (e EitherFnc) Call(...Value) Value       { return e }
func (e EitherFnc) Eval(...d.Native) d.Native { return d.NilVal{}.Eval() }
func (e EitherFnc) Maybe() bool               { return false }
func (e EitherFnc) Nullable() d.Native        { return d.NilVal{} }
func (e EitherFnc) TypeFnc() TyFnc            { return Option | Either }
func (e EitherFnc) TypeNat() d.TyNative       { return d.Nil }
func (e EitherFnc) String() string {
	var str string
	return str
}
