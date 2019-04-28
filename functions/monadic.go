package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// PARAMETRIC TYPES
	///
	// parametric types are defineable recursively during runtime and
	// indicated by any bit flag that implements the typed interface. type
	// agnosticy is implemented by truth-/, case-/ and case-switch
	// expressions to decide which parametric type applys to the arguments.
	NoOp       func()
	TruthFnc   func(...Callable) bool
	CaseExpr   func(...Callable) (Callable, bool)
	CaseSwitch func(...Callable) (Callable, Consumeable, bool)

	//// PARAMETRIC FUNCTION
	///
	// parametric function returns the parametric types flag, when called
	// without arguments, or the result of applying the arguments to the
	// expression instead.
	ParamFnc func(...Callable) Callable

	//// PARAMETRIC VALUE
	///
	// returns enclosed expression applyed to the arguments, and a flag
	// implementing the typed interface.
	ParamVal func(...Callable) (Callable, Typed)

	//// PARAMETRIC CONSTRUCTOR
	///
	// parametric constructor generates values of a parametric type defined
	// at creation of the constructor
	ParamCon func(...Callable) (Parametric, bool)
)

///////////////////////////////////////////////////////////////////////////////
func NewNoOp() NoOp {
	return func() {}
}
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
func (n NoOp) Type() Typed               { return None }
func (n NoOp) TypeName() string          { return n.String() }

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
func (t TruthFnc) Ident() Callable     { return t }
func (t TruthFnc) String() string      { return "Truth" }
func (t TruthFnc) TypeName() string    { return t.String() }
func (t TruthFnc) Type() Typed         { return Truth }
func (t TruthFnc) TypeFnc() TyFnc      { return Truth }
func (t TruthFnc) TypeNat() d.TyNative { return d.Expression | d.Bool }

///////////////////////////////////////////////////////////////////////////////
func NewCaseExpr(truth TruthFnc) CaseExpr {

	return func(exprs ...Callable) (Callable, bool) {

		var expr = exprs[0]

		if len(exprs) > 0 {
			expr = NewVector(exprs...)
		}

		return expr, truth(exprs...)
	}
}
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
func (c CaseExpr) Ident() Callable     { return c }
func (c CaseExpr) TypeNat() d.TyNative { return d.Expression }
func (c CaseExpr) TypeFnc() TyFnc      { return Case }
func (c CaseExpr) Type() Typed         { return Case }
func (c CaseExpr) TypeName() string    { return Case.String() }

///////////////////////////////////////////////////////////////////////////////
func NewCaseSwitch(cases ...CaseExpr) CaseSwitch {

	var elems = []Callable{}

	for _, cas := range cases {
		elems = append(elems, cas)
	}

	var caselist = NewList(elems...)

	return func(args ...Callable) (Callable, Consumeable, bool) {

		var head, caselist = caselist()
		var result, ok = head.(CaseExpr)(args...)

		return result, caselist, ok
	}
}
func (c CaseSwitch) Truth() Callable {
	var _, _, ok = c()
	return NewFromData(d.BoolVal(ok))
}
func (c CaseSwitch) Cases() Consumeable {
	var _, cases, _ = c()
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
func (c CaseSwitch) Type() Typed         { return Switch }
func (c CaseSwitch) TypeName() string    { return c.Type().String() }
func (c CaseSwitch) Ident() Callable     { return c }
func (c CaseSwitch) TypeFnc() TyFnc      { return Switch }
func (c CaseSwitch) TypeNat() d.TyNative { return d.Expression }

///////////////////////////////////////////////////////////////////////////////
func NewParametricValue(
	expr func(...Callable) Callable,
	typ Typed,
) ParamVal {
	return func(args ...Callable) (Callable, Typed) {
		return expr(args...), typ
	}
}
func (t ParamVal) String() string {
	return t.TypeName() + " " + t.String()
}
func (t ParamVal) Type() Typed {
	var _, typ = t()
	return typ
}
func (t ParamVal) Expr() Callable {
	var result, _ = t()
	return result
}
func (t ParamVal) Call(args ...Callable) Callable {
	var result, _ = t(args...)
	return result
}
func (t ParamVal) Eval(args ...d.Native) d.Native {
	return t.Call(NatToFnc(args...)...)
}
func (t ParamVal) Ident() Callable     { return t }
func (t ParamVal) TypeFnc() TyFnc      { return Options }
func (t ParamVal) TypeNat() d.TyNative { return d.Expression }
func (t ParamVal) TypeName() string    { return t.Type().String() }

///////////////////////////////////////////////////////////////////////////////
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
func (t ParamFnc) Eval(args ...d.Native) d.Native {
	return t(NatToFnc(args...)...)
}
func (t ParamFnc) String() string {
	return t.TypeName() + " " + t.Expr().String()
}
func (t ParamFnc) Ident() Callable                { return t }
func (t ParamFnc) Expr() Callable                 { return t() }
func (t ParamFnc) Call(args ...Callable) Callable { return t(args...) }
func (t ParamFnc) Type() Typed                    { return t().(Typed) }
func (t ParamFnc) TypeName() string               { return t.Type().String() }
func (t ParamFnc) TypeNat() d.TyNative            { return d.Expression }
func (t ParamFnc) TypeFnc() TyFnc                 { return Options }

///////////////////////////////////////////////////////////////////////////////
func NewParamCon(
	base Typed,
	cases ListVal,
	args ...Callable,
) ParamCon {

	return ParamCon(

		func(args ...Callable) (Parametric, bool) {

			// return parametric type, when called without
			// parameters
			if len(args) > 0 {

				// decapitate cases list
				var head, cases = cases()

				// if decapitation yielded another case to
				// scrutinize‥.
				if head != nil {

					// apply arguments to yielded case‥.
					if param, ok := head.(Parametric); ok {

						// return parametric and true
						// if case evaluation succeded
						// in yielding value
						return param, ok
					}

					// construct case continuation if case
					// evaluation failed to yield a value,
					// pass on base type and arguments
					// passed with initial call.
					return NewParamCon(base,
						cases,
						args...,
					), false
				}
			}

			// no case matched the provided arguments, or no
			// arguments got passed by call in the first place‥.
			// return base type & 'false' as final result.
			return ParamVal(func(...Callable) (Callable, Typed) {
				return NewFromData(base), Type
			}), false
		})
}

func DefineParametricConstructor(
	base Typed,
	cons ...CaseExpr,
) ParamCon {
	var initials = []Callable{}
	for _, con := range cons {
		initials = append(initials, con)
	}
	var cases = NewList(initials...)
	return NewParamCon(base, cases)
}

func (p ParamCon) Call(args ...Callable) Callable {
	var result, _ = p(args...)
	return result
}
func (p ParamCon) Eval(args ...d.Native) d.Native {
	var result, _ = p(NatToFnc(args...)...)
	return result
}
func (p ParamCon) TypeFnc() TyFnc {
	var param, _ = p()
	return param.Call().(TyFnc)
}
func (p ParamCon) TypeName() string {
	var typ, _ = p()
	return typ.String()
}
func (p ParamCon) Type() Typed {
	var param, _ = p()
	return param.(Typed)
}
func (p ParamCon) TypeNat() d.TyNative { return d.Expression }
func (p ParamCon) String() string      { return p.TypeName() }

///////////////////////////////////////////////////////////////////////////////
