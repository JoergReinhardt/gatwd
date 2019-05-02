package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	NoneVal    func()
	TruthExpr  func(...Callable) bool
	CaseCheck  func(...Callable) (Callable, bool)
	CaseSwitch func(...Callable) (Callable, ListVal, bool)
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
//// TRUTH MODE
///
// truth mode flag determines if truth function will retun true only if all
// arguments evaluate true, or if any is af the arguments evaluates true.
type TruthMode bool

func (t TruthMode) String() string {
	if t {
		return "All"
	}
	return "Any"
}

const (
	All TruthMode = true
	Any TruthMode = false
)

// mode parameter is optional, default will return true if all passed arguments
// evaluate to be true.
func NewTruth(
	truth func(Callable) bool,
	mode ...TruthMode,
) TruthExpr {
	var all = All
	if len(mode) > 0 {
		all = mode[0]
	}
	if all {
		// All
		return func(args ...Callable) bool {
			for _, arg := range args {
				if !truth(arg) {
					return false
				}
			}
			return true
		}
	}
	// Any
	return func(args ...Callable) bool {
		for _, arg := range args {
			if truth(arg) {
				return true
			}
		}
		return false
	}
}
func (t TruthExpr) Call(args ...Callable) Callable {
	return NewFromData(d.BoolVal(t(args...)))
}
func (t TruthExpr) Eval(args ...d.Native) d.Native {
	return d.BoolVal(t(NatToFnc(args...)...))
}
func (t TruthExpr) Ident() Callable  { return t }
func (t TruthExpr) String() string   { return "Truth" }
func (t TruthExpr) TypeName() string { return t.String() }
func (t TruthExpr) Type() Typed      { return Truth }
func (t TruthExpr) TypeFnc() TyFnc   { return Truth }
func (t TruthExpr) TypeNat() d.TyNat { return d.Expression | d.Bool }

///////////////////////////////////////////////////////////////////////////////
//// CASE SWITCH
///
// case function represents a single case in a case switch and returns true and
// it's expression, if the passed arguments evaluate true, or false and a none
// instance otherwise.
func NewCaseFnc(expr Callable, truth TruthExpr) CaseCheck {
	return func(args ...Callable) (Callable, bool) {
		if truth(args...) {
			return expr, true
		}
		return NewNone(), false
	}
}
func (c CaseCheck) Truth() Callable {
	var _, ok = c()
	return NewFromData(d.BoolVal(ok))
}
func (c CaseCheck) Expr() Callable {
	var expr, _ = c()
	return expr
}
func (c CaseCheck) Call(args ...Callable) Callable {
	var result, _ = c(args...)
	return result
}
func (c CaseCheck) Eval(args ...d.Native) d.Native {
	return c.Expr().Eval(args...)
}
func (c CaseCheck) String() string {
	return "Case " + c.Expr().String()
}
func (c CaseCheck) Ident() Callable  { return c }
func (c CaseCheck) Type() Typed      { return Switch }
func (c CaseCheck) TypeName() string { return c.Type().String() }
func (c CaseCheck) TypeFnc() TyFnc   { return Case }
func (c CaseCheck) TypeNat() d.TyNat { return d.Expression }

///////////////////////////////////////////////////////////////////////////////
//// CASE SWITCH
///
// case-switch encloses case functions passed to it and evaluates one after
// another recursively and either returns the yielded value, an empty list of
// cases and 'true', or an instance of none, the list of remaining cases and
// 'false'.
func NewCaseSwitch(caseFncs ...CaseCheck) CaseSwitch {
	var list = NewList()
	for _, cf := range caseFncs {
		list = list.Cons(cf)
	}
	return ConsCaseSwitch(list)
}

// cons-case-switch takes a list of cases assumed to be case functions, pops
// the head applys the arguments and either returns the yielded value, an empty
// list of remaining cases and 'true', if the case evaluates true, or an
// instance of none the list of remaining cases and 'false'
func ConsCaseSwitch(cases ListVal) CaseSwitch {
	return func(args ...Callable) (Callable, ListVal, bool) {
		var head Callable
		if head, cases = cases(); head != nil {
			if check, ok := head.(CaseCheck); ok {
				if val, ok := check(args...); ok {
					return val, NewList(), true
				}
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
func (c CaseSwitch) Call(args ...Callable) Callable { return c.Expr().Call(args...) }
func (c CaseSwitch) Eval(args ...d.Native) d.Native { return c.Call(NatToFnc(args...)...) }
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
	TypeSignat func() (ParamFlag, []TypeSignat)
	ParamValue func(...Callable) (Callable, TypeSignat)
	TypeCheck  func(...ParamFlag) (ParamValue, bool)
	ParamType  func(...Callable) (TypeCheck, ListVal, bool)
)

///////////////////////////////////////////////////////////////////////////////
//// TYPE CHECK
///
// type check constructor expects a case-check expression as argument, that
// takes intances of type param-flags as it's arguments and also returns a
// value of that type.
// CaseCheck  func(...Callable) (Callable, bool)
func NewTypeCheck(check CaseCheck) TypeCheck {
	return func(flags ...ParamFlag) (ParamValue, bool) {
		var args = []Callable{}
		for _, arg := range flags {
			args = append(args, arg)
		}
		var value, isval = check(args...)
		if parm, ok := value.(ParamValue); ok {
			return parm, isval
		}

		return NewParamNone(),
			false
	}
}

func NewParamNone() ParamValue {
	var none = NewNone()
	return NewAtomicParamValue(
		none.Call,
		none.TypeNat(),
		none.TypeFnc(),
		none.String(),
	)
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
func NewTypeSignature(flag ParamFlag, signatures ...TypeSignat) TypeSignat {
	return func() (ParamFlag, []TypeSignat) {
		return flag, signatures
	}
}
func (f TypeSignat) Ident() Callable { return f }
func (f TypeSignat) Flag() ParamFlag {
	var flag, _ = f()
	return flag
}
func (f TypeSignat) SubFlags() []TypeSignat {
	var _, subs = f()
	return subs
}
func (f TypeSignat) Len() int {
	return len(f.SubFlags())
}
func (f TypeSignat) Atomic() bool {
	if f.Len() == 0 {
		return true
	}
	return false
}
func (f TypeSignat) Type() (d.TyNat, TyFnc) {
	return f.Flag().Type()
}
func (f TypeSignat) TypeName() string { return f.Flag().TypeName() }
func (f TypeSignat) TypeNat() d.TyNat {
	return f.Flag().TypeNat()
}
func (f TypeSignat) TypeFnc() TyFnc {
	return f.Flag().TypeFnc()
}
func (f TypeSignat) Eval(args ...d.Native) d.Native {
	return d.StrVal(f.String())
}
func (f TypeSignat) Call(args ...Callable) Callable {
	return NewFromData(f.Eval())
}
func (f TypeSignat) String() string {
	var str = f.Flag().TypeName()
	//str = str + strings.Join()
	return str
}

///////////////////////////////////////////////////////////////////////////////
func NewAtomicParamValue(
	expr func(...Callable) Callable,
	nat d.TyNat,
	fnc TyFnc,
	name ...string,
) ParamValue {
	return func(args ...Callable) (Callable, TypeSignat) {
		return expr(args...), NewTypeSignature(NewParamFlag(nat, fnc, name...))
	}
}
func NewNamedParamValue(
	expr func(...Callable) Callable,
	nat d.TyNat,
	fnc TyFnc,
	name string,
	sigs ...TypeSignat,
) ParamValue {
	return func(args ...Callable) (Callable, TypeSignat) {
		return expr(args...), NewTypeSignature(NewParamFlag(nat, fnc, name), sigs...)
	}
}
func NewParamValue(
	expr func(...Callable) Callable,
	nat d.TyNat,
	fnc TyFnc,
	sigs ...TypeSignat,
) ParamValue {
	return func(args ...Callable) (Callable, TypeSignat) {
		return expr(args...), NewTypeSignature(NewParamFlag(nat, fnc), sigs...)
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
