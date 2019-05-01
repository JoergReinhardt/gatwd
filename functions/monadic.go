package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	NoneVal    func()
	TruthFnc   func(...Callable) bool
	CaseFnc    func(...Callable) (Callable, bool)
	CaseSwitch func(...Callable) (Callable, Consumeable, bool)
)

///////////////////////////////////////////////////////////////////////////////
func NewNone() NoneVal {
	return func() {}
}
func (n NoneVal) Ident() Callable           { return n }
func (n NoneVal) Maybe() bool               { return false }
func (n NoneVal) Empty() bool               { return true }
func (n NoneVal) Eval(...d.Native) d.Native { return nil }
func (n NoneVal) Value() Callable           { return nil }
func (n NoneVal) Call(...Callable) Callable { return nil }
func (n NoneVal) String() string            { return "⊥" }
func (n NoneVal) Len() int                  { return 0 }
func (n NoneVal) TypeFnc() TyFnc            { return None }
func (n NoneVal) TypeNat() d.TyNat          { return d.Nil }
func (n NoneVal) Type() Typed               { return None }
func (n NoneVal) TypeName() string          { return n.String() }

///////////////////////////////////////////////////////////////////////////////
func NewTruth(
	truth func(...Callable) bool,
) TruthFnc {
	return truth
}
func (t TruthFnc) Call(args ...Callable) Callable {
	return NewFromData(d.BoolVal(t(args...)))
}
func (t TruthFnc) Eval(args ...d.Native) d.Native {
	return d.BoolVal(t(NatToFnc(args...)...))
}
func (t TruthFnc) Ident() Callable  { return t }
func (t TruthFnc) String() string   { return "Truth" }
func (t TruthFnc) TypeName() string { return t.String() }
func (t TruthFnc) Type() Typed      { return Truth }
func (t TruthFnc) TypeFnc() TyFnc   { return Truth }
func (t TruthFnc) TypeNat() d.TyNat { return d.Expression | d.Bool }

///////////////////////////////////////////////////////////////////////////////
//// CASE SWITCH
///
func NewCaseFnc(expr Callable, truth TruthFnc) CaseFnc {
	return func(args ...Callable) (Callable, bool) {
		if truth(args...) {
			return expr, true
		}
		return NewNone(), false
	}
}
func (c CaseFnc) Truth() Callable {
	var _, ok = c()
	return NewFromData(d.BoolVal(ok))
}
func (c CaseFnc) Expr() Callable {
	var expr, _ = c()
	return expr
}
func (c CaseFnc) Call(args ...Callable) Callable {
	var result, _ = c(args...)
	return result
}
func (c CaseFnc) Eval(args ...d.Native) d.Native {
	return c.Expr().Eval(args...)
}
func (c CaseFnc) String() string {
	return "Case " + c.Expr().String()
}
func (c CaseFnc) Ident() Callable  { return c }
func (c CaseFnc) Type() Typed      { return Switch }
func (c CaseFnc) TypeName() string { return c.Type().String() }
func (c CaseFnc) TypeFnc() TyFnc   { return Case }
func (c CaseFnc) TypeNat() d.TyNat { return d.Expression }

///////////////////////////////////////////////////////////////////////////////
func NewCaseSwitch(caseExprs ...CaseFnc) CaseSwitch {
	var cs []Callable
	for _, c := range caseExprs {
		cs = append(cs, c)
	}
	return ConsCaseSwitch(NewList(cs...))
}
func ConsCaseSwitch(cases Consumeable) CaseSwitch {
	return func(args ...Callable) (Callable, Consumeable, bool) {
		if cas, cases := cases.(ListVal)(args...); cas != nil {
			if val, ok := cas.(CaseSwitch); ok {
				return val, cases, ok
			}
		}
		return NewNone(), cases, false
	}
}
func (c CaseSwitch) Expr() Callable {
	var expr, _, _ = c()
	return expr
}
func (c CaseSwitch) Cases() Consumeable {
	var _, cases, _ = c()
	return cases
}
func (c CaseSwitch) Truth() Callable {
	var _, _, ok = c()
	return NewFromData(d.BoolVal(ok))
}
func (c CaseSwitch) Call(args ...Callable) Callable {
	var expr Callable
	var cases Consumeable
	if expr, cases = c.Cases().(ListVal)(); expr != nil {
		return expr.Call(args...)
	}
	return ConsCaseSwitch(cases)
}
func (c CaseSwitch) Eval(args ...d.Native) d.Native {
	return c.Call(NatToFnc(args...)...)
}
func (c CaseSwitch) String() string {
	return "Switch " + c.Expr().String()
}
func (c CaseSwitch) Ident() Callable  { return c }
func (c CaseSwitch) Type() Typed      { return Switch }
func (c CaseSwitch) TypeName() string { return c.Type().String() }
func (c CaseSwitch) TypeFnc() TyFnc   { return Switch }
func (c CaseSwitch) TypeNat() d.TyNat { return d.Expression }

///////////////////////////////////////////////////////////////////////////////
type (
	ParamFlag  func() (string, d.TyNat, TyFnc)
	ParamValue func(...Callable) (Callable, ParamFlag)
	ParamType  func(...Callable) (Callable, ParamFlag)
)

///////////////////////////////////////////////////////////////////////////////
func NewParamType(
	cases CaseSwitch,
	nat d.TyNat,
	fnc TyFnc,
	name ...string,
) ParamType {

	var flag = NewParamFlag(nat, fnc, name...)
	var expr Callable

	return func(args ...Callable) (Callable, ParamFlag) {
		return expr, flag
	}
}

///////////////////////////////////////////////////////////////////////////////
func NewParamFlag(nat d.TyNat, fnc TyFnc, name ...string) ParamFlag {
	var str string

	if len(name) > 0 {
		for _, n := range name {
			str = str + n + "·"
		}
	}

	str = nat.String() + "·" + fnc.String()

	return func() (string, d.TyNat, TyFnc) {
		return str, nat, fnc
	}
}

func (f ParamFlag) Ident() Callable { return f }
func (f ParamFlag) Type() (d.TyNat, TyFnc) {
	return f.TypeNat(), HigherOrder | f.TypeFnc()
}
func (f ParamFlag) TypeName() string { return f.String() }
func (f ParamFlag) String() string {
	var name, _, _ = f()
	return name
}
func (f ParamFlag) TypeNat() d.TyNat {
	var _, nat, _ = f()
	return nat
}
func (f ParamFlag) TypeFnc() TyFnc {
	var _, _, fnc = f()
	return fnc
}
func (f ParamFlag) Eval(args ...d.Native) d.Native {
	return d.StrVal(f.String())
}
func (f ParamFlag) Call(args ...Callable) Callable {
	return NewFromData(f.Eval())
}

///////////////////////////////////////////////////////////////////////////////
func NewParamValue(
	expr func(...Callable) Callable,
	nat d.TyNat,
	fnc TyFnc,
	name ...string,
) ParamValue {
	return func(args ...Callable) (Callable, ParamFlag) {
		return expr(args...), NewParamFlag(nat, fnc, name...)
	}
}
func (v ParamValue) Ident() Callable { return v }
func (v ParamValue) Expr() Callable {
	var expr, _ = v()
	return expr
}
func (v ParamValue) TypeNat() d.TyNat {
	var _, typ = v()
	return typ.TypeNat()
}
func (v ParamValue) TypeFnc() TyFnc {
	var _, typ = v()
	return typ.TypeFnc()
}
func (v ParamValue) TypeName() string {
	var _, typ = v()
	return typ.TypeName()
}
func (v ParamValue) Type() (d.TyNat, TyFnc) {
	return v.TypeNat(), v.TypeFnc()
}
func (v ParamValue) String() string {
	return v.Expr().String()
}
func (v ParamValue) Eval(args ...d.Native) d.Native {
	return v.Expr().Eval(args...)
}
func (v ParamValue) Call(args ...Callable) Callable {
	return v.Expr().Call(args...)
}
