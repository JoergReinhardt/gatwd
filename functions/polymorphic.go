package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// FUNCTOR CONSTRUCTORS
	///
	// CONDITIONAL, CASE & OPTIONAL
	NoOp      func()
	OptVal    func() PairFnc
	TrueFalse func() d.BoolVal
	LGEqual   func() d.Int8Val
	JustNone  func(...Parametric) OptVal
	EitherOr  func(...Parametric) OptVal
	CaseExpr  func(...Parametric) OptVal
	EqualFnc  func(...Parametric) LGEqual
	TruthFnc  func(...Parametric) TrueFalse
	ErrorFnc  func(err ...error) []d.ErrorVal
)

func (e ErrorFnc) Error() string {
	var str string
	var l = len(e())
	for n, err := range e() {
		str = str + err.String()
		if n < l-1 {
			str = str + "\n"
		}
	}
	return str
}

func NewError(init d.ErrorVal) ErrorFnc {
	var errors = []d.ErrorVal{}
	return func(errs ...error) []d.ErrorVal {
		if len(errs) == 0 {
			return errors
		}
		for _, err := range errs {
			errors = append(errors, d.ErrorVal{err})
		}
		return errors
	}
}

func NewTruthConstant(truth bool) TrueFalse {
	return func() d.BoolVal { return d.BoolVal(truth) }
}
func NewTruthConstantFromNative(truth d.BoolVal) TrueFalse {
	return func() d.BoolVal { return truth }
}

func (t TrueFalse) Eval(...d.Native) d.Native { return t() }

func (t TrueFalse) Call(...Parametric) Parametric { return t }

func (t TrueFalse) Ident() Parametric { return t }

func (t TrueFalse) TypeNat() d.TyNative { return d.Bool }

func (t TrueFalse) TypeFnc() TyFnc {
	if t() {
		return True
	}
	return False
}

func (t TrueFalse) String() string {
	var str = "Truth|"
	var instval string
	if True.Flag().Match(t.TypeFnc()) {
		instval = "True"
	} else {
		instval = "False"
	}
	return str + instval
}

func NewTruthFunction(predicate Parametric) TruthFnc {
	return func(args ...Parametric) TrueFalse {
		return NewTruthConstantFromNative(
			predicate.Call(args...).Eval().(d.BoolVal))
	}
}

func (t TruthFnc) Ident() Parametric { return t }

func (t TruthFnc) String() string { return "Predicate Function" }

func (t TruthFnc) TypeFnc() TyFnc { return Truth }

func (t TruthFnc) TypeNat() d.TyNative { return d.Bool }

func (t TruthFnc) Call(args ...Parametric) Parametric { return t(args...) }

func (t TruthFnc) Eval(args ...d.Native) d.Native { return t.Eval(args...) }

/// OPTIONAL VALUES
// based on a paired with extra bells'n whistles
func NewOptVal(left, right Parametric) OptVal {
	return OptVal(func() PairFnc {
		return NewPair(left, right)
	})
}

func (o OptVal) Ident() Parametric { return o }

func (o OptVal) Left() Parametric { return o().Left() }

func (o OptVal) Right() Parametric { return o().Right() }

func (o OptVal) TypeNat() d.TyNative { return o.Right().TypeNat() }

func (o OptVal) TypeFnc() TyFnc { return o.Right().TypeFnc() | Option }

func (o OptVal) String() string {
	return "optional: " + o.Left().String() + "|" + o.Right().String()
}

func (o OptVal) Call(args ...Parametric) Parametric { return o.Right().Call(args...) }

func (o OptVal) Eval(vars ...d.Native) d.Native {
	var args = []Parametric{}
	for _, arg := range vars {
		args = append(args, NewFromData(arg))
	}
	return o.Call(args...)
}

/// EQUALITY VALUE
func NewLGEqual(lge int) LGEqual {
	return func() d.Int8Val { return d.Int8Val(int8(lge)) }
}
func (l LGEqual) Init() Parametric { return l }

func (l LGEqual) Eval(args ...d.Native) d.Native { return l().Eval(args...) }

func (l LGEqual) Call(...Parametric) Parametric { return l }

func (l LGEqual) TypeNat() d.TyNative { return d.Equals }

func (l LGEqual) TypeFnc() TyFnc { return Equality }

func (l LGEqual) String() string {
	if l() < 0 {
		return "lesser"
	}
	if l() > 0 {
		return "greater"
	}
	return "equal"
}

/// EQUALITY FUNCTION
func NewEqualFnc(eqfnc Parametric) EqualFnc {
	return func(args ...Parametric) LGEqual {
		var num, ok = eqfnc.Call(args...).(d.Integer)
		if ok {
			return NewLGEqual(num.Int())
		}
		// eqfnc did not yild a data integer instance
		return NewLGEqual(-2)
	}
}

//// JUST NONE
///
// expression is applyed to arguments passed at runtime. result of calling the
// expression is applyed to predex. if the predicate matches, result is
// returned as 'just' value, otherwise NoOp is returned
func NewJustNone(predex TruthFnc, expr Parametric) JustNone {
	return JustNone(func(args ...Parametric) OptVal {
		var result = expr.Call(args...)
		if predex(result)() {
			return NewOptVal(New(true), result)
		}
		return NewOptVal(New(false), NewNoOp())
	})
}

func (j JustNone) Ident() Parametric { return j }

func (j JustNone) String() string { return "just-none" }

func (j JustNone) TypeFnc() TyFnc { return Option | Just | None }

func (j JustNone) TypeNat() d.TyNative { return d.Function }

func (j JustNone) Return(args ...Parametric) OptVal { return j(args...) }

func (j JustNone) Call(args ...Parametric) Parametric { return j(args...) }

func (j JustNone) Eval(vars ...d.Native) d.Native {
	var args = []Parametric{}
	for _, v := range vars {
		args = append(args, NewFromData(v))
	}
	return j(args...)
}

//// EITHER OR
///
// left pair value indicates: 0 = 'either', 1 = 'or', -1 = 'no value yielded'
func NewEitherOr(predex TruthFnc, either, or Parametric) EitherOr {
	return EitherOr(func(args ...Parametric) OptVal {
		var val Parametric

		val = either.Call(args...)

		if predex(val)() {
			return NewOptVal(New(0), val)
		}

		val = or.Call(args...)

		if predex(val)() {
			return NewOptVal(New(1), val)
		}

		return NewOptVal(New(-1), NewNoOp())
	})
}

func (e EitherOr) Ident() Parametric { return e }

func (e EitherOr) String() string { return "either-or" }

func (e EitherOr) TypeFnc() TyFnc { return Option | Either | Or }

func (e EitherOr) TypeNat() d.TyNative { return d.Function }

func (e EitherOr) Return(args ...Parametric) OptVal { return e(args...) }

func (e EitherOr) Call(args ...Parametric) Parametric { return e(args...) }

func (e EitherOr) Eval(vars ...d.Native) d.Native {
	var args = []Parametric{}
	for _, v := range vars {
		args = append(args, NewFromData(v))
	}
	return e.Call(args...)
}

//// SWITCH CASE
///
// switch case applies predicate to arguments passed at runtime & either
// returns either the enclosed expression, the next case, or no-op in case an
// error occured.
func NewSwitchCase(predex TruthFnc, value Parametric, nextcase ...CaseExpr) CaseExpr {
	// if runtime arguments applyed to predicate expression yields true, value
	// will be returned, or otherwise the next case will be the return value.
	return CaseExpr(func(args ...Parametric) OptVal {
		if predex(args...)() { // return value if runtime args match predicate
			return OptVal(func() PairFnc {
				return NewPair(New(0), value)
			})
		} // return next switch case to test against, if at least one
		// more case was passed
		if len(nextcase) > 0 {
			return OptVal(func() PairFnc {
				return NewPair(New(1), nextcase[0])
			})
		}
		// no case matched
		return OptVal(func() PairFnc {
			return NewPair(New(-1), NewNoOp())
		})
	})
}

func (s CaseExpr) Ident() Parametric { return s }

func (s CaseExpr) String() string { return "switch-case" }

func (s CaseExpr) TypeFnc() TyFnc { return Option | Case | Switch }

func (s CaseExpr) TypeNat() d.TyNative { return d.Function }

func (s CaseExpr) Return(args ...Parametric) OptVal { return s(args...) }

func (s CaseExpr) Call(args ...Parametric) Parametric { return s(args...) }

func (s CaseExpr) Eval(vars ...d.Native) d.Native {
	var args = []Parametric{}
	for _, v := range vars {
		args = append(args, NewFromData(v))
	}
	return s.Call(args...)
}

// NONE
func NewNoOp() NoOp { return NoOp(func() {}) }

func (n NoOp) Maybe() bool { return false }

func (n NoOp) Empty() bool { return true }

func (n NoOp) String() string { return "⊥" }

func (n NoOp) Len() int { return -1 }

func (n NoOp) Value() Parametric { return n }

func (n NoOp) Ident() Parametric { return n }

func (n NoOp) Call(...Parametric) Parametric { return n }

func (n NoOp) Eval(...d.Native) d.Native { return d.NilVal{} }

func (n NoOp) TypeNat() d.TyNative { return d.Nil }

func (n NoOp) TypeFnc() TyFnc { return None }
