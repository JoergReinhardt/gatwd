/*
 */
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	NoOp       func()
	TruthFnc   func(...Callable) bool
	CaseExpr   func(...Callable) (Callable, bool)
	CaseSwitch func(...Callable) (Callable, bool, Consumeable)

	///// PARAMETRIC TYPES
	////
	/// PARAMETRIC VALUE
	//
	// returns enclosed expression applyed to the arguments, and a flag
	// implementing the typed interface.
	ParamVal func(...Callable) (Callable, Typed)

	/// PARAMETRIC FUNCTION
	//
	// parametric function returns the parametric types flag, when called
	// without arguments, or the result of applying the arguments to the
	// expression instead.
	ParamFnc func(...Callable) Callable

	/// PARAMETRIC CONSTRUCTOR
	//
	// parametric constructor applys its arguments to the first of a list
	// of cases, to either yield an expressions of one of the parametric
	// types defined by the data constructor, a boolean to indicate if the
	// computation yielded a value, or returned the arguments that got
	// passed instead, and the list of remaining case expressions.
	ParamCon func(...Callable) (Callable, bool, Consumeable)
)

/////////////////////////////////////////////////////////////////////////////

/////////////////////////////////////////////////////////////////////////////
func NewNoOp() NoOp                      { return func() {} }
func (n NoOp) Ident() Callable           { return n }
func (n NoOp) Maybe() bool               { return false }
func (n NoOp) Empty() bool               { return true }
func (n NoOp) Eval(...d.Native) d.Native { return nil }
func (n NoOp) Value() Callable           { return nil }
func (n NoOp) Call(...Callable) Callable { return nil }
func (n NoOp) String() string            { return "⊥" }
func (n NoOp) Len() int                  { return 0 }
func (n NoOp) TypeFnc() TyFnc            { return None }
func (n NoOp) TypeNat() d.TyNative       { return d.Nil }
func (n NoOp) Type() TyFnc               { return None }
func (n NoOp) TypeName() string          { return n.String() }

/////////////////////////////////////////////////////////////////////////////
func NewTruth(
	truth func(...Callable) bool,
) TruthFnc {
	return truth
}
func (t TruthFnc) Ident() Callable     { return t }
func (t TruthFnc) String() string      { return "Truth" }
func (t TruthFnc) TypeFnc() TyFnc      { return Truth }
func (t TruthFnc) TypeNat() d.TyNative { return d.Expression | d.Bool }
func (t TruthFnc) Call(args ...Callable) Callable {
	return NewFromData(d.BoolVal(t(args...)))
}
func (t TruthFnc) Eval(args ...d.Native) d.Native {
	return d.BoolVal(t(NatToFnc(args...)...))
}

/////////////////////////////////////////////////////////////////////////////
func NewCaseExpr(truth TruthFnc) CaseExpr {

	return func(exprs ...Callable) (Callable, bool) {

		var expr = exprs[0]

		if len(exprs) > 0 {
			expr = NewVector(exprs...)
		}

		return expr, truth(exprs...)
	}
}
func (c CaseExpr) TypeFnc() TyFnc      { return Case }
func (c CaseExpr) TypeNat() d.TyNative { return d.Expression }
func (c CaseExpr) Truth() Callable {
	var _, truth = c()
	return NewFromData(d.BoolVal(truth))
}
func (c CaseExpr) Expr() Callable {
	var expr, _ = c()
	return expr
}
func (c CaseExpr) String() string {
	return "Case " + c.Expr().String()
}
func (c CaseExpr) Call(args ...Callable) Callable {
	var result, _ = c(args...)
	return result
}
func (c CaseExpr) Eval(args ...d.Native) d.Native {
	return c.Expr().Eval(args...)
}

/////////////////////////////////////////////////////////////////////////////
func NewCaseSwitch(cases ...CaseExpr) CaseSwitch {

	var elems = []Callable{}

	for _, cas := range cases {
		elems = append(elems, cas)
	}

	var caselist = NewList(elems...)

	return func(args ...Callable) (Callable, bool, Consumeable) {

		var head, caselist = caselist()
		var result, truth = head.(CaseExpr)(args...)

		return result, truth, caselist
	}
}
func (c CaseSwitch) TypeFnc() TyFnc      { return Switch }
func (c CaseSwitch) TypeNat() d.TyNative { return d.Expression }
func (c CaseSwitch) Truth() Callable {
	var _, truth, _ = c()
	return NewFromData(d.BoolVal(truth))
}
func (c CaseSwitch) Cases() Consumeable {
	var _, _, cases = c()
	return cases
}
func (c CaseSwitch) Expr() Callable {
	var expr, _, _ = c()
	return expr
}
func (c CaseSwitch) Call(args ...Callable) Callable {
	var result, _, _ = c(args...)
	return result
}
func (c CaseSwitch) Eval(args ...d.Native) d.Native {
	return c.Expr().Eval(args...)
}
func (c CaseSwitch) String() string {
	return "Switch " + c.Expr().String()
}

/////////////////////////////////////////////////////////////////////////////
func NewParametricValue(
	expr func(...Callable) Callable,
	typ Typed,
) ParamVal {
	return func(args ...Callable) (Callable, Typed) {
		return expr(args...), typ
	}
}
func (t ParamVal) Ident() Callable     { return t }
func (t ParamVal) TypeFnc() TyFnc      { return Options }
func (t ParamVal) TypeNat() d.TyNative { return d.Expression }
func (t ParamVal) TypeName() string    { return t.Type().String() }
func (t ParamVal) String() string {
	return t.TypeName() + " " + t.String()
}
func (t ParamVal) Type() Typed {
	var _, typ = t()
	return typ
}
func (t ParamVal) Call(args ...Callable) Callable {
	var result, _ = t(args...)
	return result
}
func (t ParamVal) Eval(args ...d.Native) d.Native {
	return t.Call(NatToFnc(args...)...)
}

/////////////////////////////////////////////////////////////////////////////
func NewParametricFunction(
	expr func(...Callable) Callable,
	typ Typed,
) ParamFnc {
	return func(args ...Callable) Callable {
		if len(args) == 0 {
			return NewFromData(typ)
		}
		return expr(args...)
	}
}
func (t ParamFnc) Ident() Callable                { return t }
func (t ParamFnc) Expr(args ...Callable) Callable { return t(args...) }
func (t ParamFnc) Type() Typed                    { return t().(Typed) }
func (t ParamFnc) TypeFnc() TyFnc                 { return Options }
func (t ParamFnc) TypeNat() d.TyNative            { return d.Expression }
func (t ParamFnc) TypeName() string               { return t.Type().String() }
func (t ParamFnc) String() string {
	return t.TypeName() + " " + t.Expr().String()
}
func (t ParamFnc) Call(args ...Callable) Callable {
	return t(args...)
}
func (t ParamFnc) Eval(args ...d.Native) d.Native {
	return t(NatToFnc(args...)...)
}

/////////////////////////////////////////////////////////////////////////////
