package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	NoneVal    func()
	TruthExpr  func(...Callable) bool
	CaseExpr   func(...Callable) (Callable, bool)
	SwitchExpr func(...Callable) (Callable, ListVal, bool)
)

///////////////////////////////////////////////////////////////////////////////
func (n NoneVal) Ident() Callable           { return n }
func (n NoneVal) Len() int                  { return 0 }
func (n NoneVal) String() string            { return "⊥" }
func (n NoneVal) Eval(...d.Native) d.Native { return nil }
func (n NoneVal) Value() Callable           { return nil }
func (n NoneVal) Call(...Callable) Callable { return nil }
func (n NoneVal) Empty() bool               { return true }
func (n NoneVal) TypeFnc() TyFnc            { return None }
func (n NoneVal) TypeNat() d.TyNat          { return d.Nil }
func (n NoneVal) TypeName() string          { return n.String() }
func NewNone() NoneVal                      { return func() {} }

///////////////////////////////////////////////////////////////////////////////
//// TRUTH MODE
///
// truth mode flag determines if truth function will retun true only if all
// arguments evaluate true, or if any is af the arguments evaluates true.
type TruthMode bool

func (t TruthMode) String() string {
	if t {
		return "All True"
	}
	return "Any True"
}

const (
	All TruthMode = true
	Any TruthMode = false
)

// truth function iterates over passed arguments and behaves in one of two
// modes, according to the mode parameter.  mode parameter is optional, default
// will return true if all passed arguments evaluate to be true. setting the
// parameter to 'Any', will change truth behaviour to yield false on first
// value that evaluates true
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
func NewCase(expr Callable, truth TruthExpr) CaseExpr {
	return func(args ...Callable) (Callable, bool) {
		if truth(args...) {
			return expr, true
		}
		return NewNone(), false
	}
}
func (c CaseExpr) Truth() Callable {
	var _, ok = c()
	return NewFromData(d.BoolVal(ok))
}
func (c CaseExpr) Expr() Callable {
	var expr, _ = c()
	return expr
}
func (c CaseExpr) Call(args ...Callable) Callable {
	var result, _ = c(args...)
	return result
}
func (c CaseExpr) Eval(args ...d.Native) d.Native {
	return c.Expr().Eval(args...)
}
func (c CaseExpr) String() string {
	return "Case " + c.Expr().String()
}
func (c CaseExpr) Ident() Callable  { return c }
func (c CaseExpr) Type() Typed      { return Switch }
func (c CaseExpr) TypeName() string { return c.Type().String() }
func (c CaseExpr) TypeFnc() TyFnc   { return Case }
func (c CaseExpr) TypeNat() d.TyNat { return d.Expression }

///////////////////////////////////////////////////////////////////////////////
//// CASE SWITCH
///
// case-switch encloses case functions passed to it and evaluates one after
// another recursively and either returns the yielded value, an empty list of
// cases and 'true', or an instance of none, the list of remaining cases and
// 'false'.
func NewSwitch(caseFncs ...CaseExpr) SwitchExpr {
	var args = []Callable{}
	for _, arg := range caseFncs {
		args = append(args, arg)
	}
	var list = NewList(args...)
	return ConsCaseSwitch(list)
}

// cons-case-switch takes a list of cases assumed to be case functions, pops
// the head applys the arguments and either returns the yielded value, an empty
// list of remaining cases and 'true', if the case evaluates true, or an
// instance of none the list of remaining cases and 'false'
func ConsCaseSwitch(cases ListVal) SwitchExpr {
	return func(args ...Callable) (Callable, ListVal, bool) {
		var head Callable
		if head, cases = cases(); head != nil {
			if check, ok := head.(CaseExpr); ok {
				if val, ok := check(args...); ok {
					return val, NewList(), true
				}
			}
		}
		return NewNone(), cases, false
	}
}
func (c SwitchExpr) Expr() Callable {
	var expr, _, _ = c()
	return expr
}
func (c SwitchExpr) Cases() Consumeable {
	var _, cases, _ = c()
	return cases
}
func (c SwitchExpr) Truth() Callable {
	var _, _, ok = c()
	return NewFromData(d.BoolVal(ok))
}
func (c SwitchExpr) Call(args ...Callable) Callable { return c.Expr().Call(args...) }
func (c SwitchExpr) Eval(args ...d.Native) d.Native { return c.Call(NatToFnc(args...)...) }
func (c SwitchExpr) String() string                 { return c.TypeFnc().String() + c.Expr().String() }
func (c SwitchExpr) TypeName() string               { return c.TypeFnc().String() }
func (c SwitchExpr) TypeNat() d.TyNat               { return d.Expression }
func (c SwitchExpr) TypeFnc() TyFnc                 { return Switch }
func (c SwitchExpr) Ident() Callable                { return c }
