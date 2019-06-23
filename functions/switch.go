package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// PREDICATE
	PredictArg func(Callable) bool
	PredictNar func(...Callable) bool
	PredictAll func(...Callable) bool
	PredictAny func(...Callable) bool

	//// CASE EXPRESSION & SWITCH
	CaseExpr   func(...Callable) (Callable, bool)
	CaseSwitch func(...Callable) (Callable, bool)

	//// MAYBE | JUST | NONE
	MaybeVal func(...Callable) Callable
	JustVal  func(...Callable) Callable

	//// EITHER | LEFT | RIGHT
	EitherVal func(...Callable) Callable
	LeftVal   func(...Callable) Callable
	RightVal  func(...Callable) Callable
)

//// PREDICATES
///
// predict one is an expression that returns either true, or false depending on
// first passed arguement passed. succeeding arguements are ignored
func NewPredictArg(pred func(Callable) bool) PredictArg {
	return func(arg Callable) bool { return pred(arg) }
}
func (p PredictArg) Eval() d.Native { return d.BoolVal(false) }
func (p PredictArg) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return NewNative(d.BoolVal(p(args[0])))
	}
	return NewNative(d.BoolVal(false))
}

// return single argument predicate as multi argument predicate
func (p PredictArg) ToPredictNarg() PredictNar {
	return func(args ...Callable) bool {
		if len(args) > 0 {
			return p(args[0])
		}
		return false
	}
}
func (p PredictArg) String() string   { return "Argument Predicate" }
func (p PredictArg) TypeFnc() TyFnc   { return Predicate | Truth }
func (p PredictArg) TypeNat() d.TyNat { return d.Function }

///////////////////////////////////////////////////////////////////////////////
// predict many returns true, or false depending on all arguments that have
// been passed calling it
func NewPredictNarg(pred func(...Callable) bool) PredictNar {
	return func(args ...Callable) bool {
		return pred(args...)
	}
}
func (p PredictNar) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return NewNative(d.BoolVal(p(args...)))
	}
	return NewNative(d.BoolVal(false))
}
func (p PredictNar) Eval() d.Native   { return d.BoolVal(false) }
func (p PredictNar) String() string   { return "Nary Predicate" }
func (p PredictNar) TypeFnc() TyFnc   { return Predicate | Truth }
func (p PredictNar) TypeNat() d.TyNat { return d.Function }

///////////////////////////////////////////////////////////////////////////////
// all-predicate returns true, if all arguments passed yield true, when applyed
// to predicate one after another
func NewPredictAll(pred func(Callable) bool) PredictNar {
	return func(args ...Callable) bool {
		var result = true
		for _, arg := range args {
			if !pred(arg) {
				return false
			}
		}
		return result
	}
}

// call passes arguments on to the enclosed all-predicate
func (p PredictAll) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return NewNative(d.BoolVal(p(args...)))
	}
	return NewNative(d.BoolVal(false))
}

func (p PredictAll) Eval() d.Native { return d.BoolVal(false) }

func (p PredictAll) String() string   { return "All Predicate" }
func (p PredictAll) TypeFnc() TyFnc   { return Predicate | Truth }
func (p PredictAll) TypeNat() d.TyNat { return d.Function }

///////////////////////////////////////////////////////////////////////////////
// will return true, if any of the passed arguments yield true, when applyed to
// predicate one after another
func NewPredictAny(pred func(Callable) bool) PredictNar {
	return func(args ...Callable) bool {
		var result = false
		for _, arg := range args {
			if pred(arg) {
				return true
			}
		}
		return result
	}
}
func (p PredictAny) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return NewNative(d.BoolVal(p(args...)))
	}
	return NewNative(d.BoolVal(false))
}
func (p PredictAny) Eval() d.Native   { return d.BoolVal(false) }
func (p PredictAny) String() string   { return "Any Predicate" }
func (p PredictAny) TypeFnc() TyFnc   { return Predicate | Truth }
func (p PredictAny) TypeNat() d.TyNat { return d.Function }

//// CASE EXPRESSION & SWITCH
///
// eval converts its arguments to callable and evaluates the result to yield a
// return value of native type
func NewCase(predicate PredictNar, exprs ...Callable) CaseExpr {
	return func(args ...Callable) (Callable, bool) {
		if predicate(args...) {
			if len(exprs) > 0 {
				if len(exprs) > 1 {
					return Curry(exprs...).Call(args...), true
				}
				return exprs[0].Call(args...), true
			}
			if len(args) > 0 {
				return NewVector(args...), true
			}
			return args[0], true
		}
		return NewNone(), false
	}
}

func (s CaseExpr) Call(args ...Callable) Callable {
	val, ok := s(args...)
	return NewPair(val, New(ok))
}
func (s CaseExpr) Eval() d.Native   { return d.NewNil() }
func (s CaseExpr) String() string   { return "Case" }
func (s CaseExpr) TypeName() string { return "Case" }
func (s CaseExpr) TypeFnc() TyFnc   { return Case | Truth }
func (s CaseExpr) TypeNat() d.TyNat { return d.Function }

// applys passed arguments to all enclosed cases in the order passed to the
// switch constructor
func NewSwitch(exprs ...CaseExpr) CaseSwitch {

	// cast predicates as slice of callables
	var caseslice = []Callable{}
	// range over predicates
	for _, exprs := range exprs {
		// append predicate to slice of predicates
		caseslice = append(caseslice, exprs)
	}
	// create cases from predicate slice
	var cases = NewList(caseslice...)
	// allocate value to assign current case to
	var current Callable

	// create and return case switch enclosing list of predicates
	return func(args ...Callable) (Callable, bool) {
		// call consumeable to yield current case and reassign list of
		// remaining cases
		current, cases = cases()
		// if call yielded any case
		if current != nil {
			// scrutinize argument(s) by applying the case
			if expr, ok := current.(CaseExpr)(args...); ok {
				// replenish cases before returning
				// successfully scrutinized arguments
				cases = NewList(caseslice...)
				// return set of arguments and true
				return expr, true
			}
			// return set of arguments and false indicator. don't
			// replenish cases of the parially applyed.
			return NewVector(args...), false
		}
		// all case are depleted not scrutinizeing the arguments →
		// replenish list of cases befor returning the final result
		cases = NewList(caseslice...)
		// return final none and false indicator
		return NewNone(), false
	}
}

// call method iterates over all cases until either boolean indicates
// scrutinized arguments to return, or cases depletion
func (s CaseSwitch) Call(args ...Callable) Callable {
	// if arguments have been passed
	if len(args) > 0 {
		// call switch instance passing the arguments
		var vec, ok = s(args...)
		// if call did not yield an instance of none
		for !vec.TypeFnc().Match(None) {
			// if boolean indicates success
			if ok {
				// return set of arguments
				return vec
			}
			// otherwise call switch to scrutinize next case
			vec, ok = s(args...)
		}
	}
	// return none if all cases are scrutinized, or no arguments where
	// passed
	return NewNone()
}

// eval converts its arguments to callable and evaluates the result to yield a
// return value of native type
func (s CaseSwitch) Eval() d.Native   { return d.NewNil() }
func (s CaseSwitch) String() string   { return "Switch Case" }
func (s CaseSwitch) TypeFnc() TyFnc   { return Switch | Case }
func (s CaseSwitch) TypeNat() d.TyNat { return d.Function }

/// MAYBE
func NewMaybe(cas CaseExpr, equ Callable) MaybeVal {
	return func(args ...Callable) Callable {
		if len(args) > 0 {
			if val, ok := cas(args...); ok {
				return NewJust(val)
			}
			return NewNone()
		}
		// no args, wrap justs types in pair
		return equ
	}
}
func (m MaybeVal) Call(args ...Callable) Callable { return m(args...) }
func (m MaybeVal) Eval() d.Native                 { return m().TypeNat() }
func (m MaybeVal) TypeNat() d.TyNat               { return m().TypeNat() }
func (m MaybeVal) TypeFnc() TyFnc                 { return m().TypeFnc() }
func (m MaybeVal) String() string                 { return m().String() }
func (m MaybeVal) TypeName() string               { return "Maybe " + m().TypeName() }

/// JUST
func NewJust(expr Callable) JustVal {
	return func(args ...Callable) Callable {
		return expr.Call(args...)
	}
}
func (j JustVal) Call(args ...Callable) Callable { return j(args...) }
func (j JustVal) Eval() d.Native                 { return j().Eval() }
func (j JustVal) TypeNat() d.TyNat               { return j().TypeNat() }
func (j JustVal) TypeFnc() TyFnc                 { return Just | j().TypeFnc() }
func (j JustVal) String() string                 { return j().String() }
func (j JustVal) TypeName() string               { return "Just " + j().TypeName() }

/// EITHER
func NewEither(cas CaseExpr, left, right Callable) EitherVal {
	return func(args ...Callable) Callable {
		var val Callable
		var ok bool
		if len(args) > 0 {
			if val, ok = cas(args...); ok {
				return NewLeft(val)
			}
			return NewRight(val)
		}
		// no arguments, wrap both types in a pair of pairs
		return NewPair(left, right)
	}
}
func (e EitherVal) Call(args ...Callable) Callable { return e(args...) }
func (e EitherVal) String() string                 { return e().String() }
func (e EitherVal) Eval() d.Native                 { return e().TypeNat() }
func (e EitherVal) TypeNat() d.TyNat               { return d.Function }
func (e EitherVal) TypeFnc() TyFnc                 { return Either }
func (e EitherVal) TypeName() string {
	return "Either " + e().(Paired).Left().TypeName() +
		"Or " + e().(Paired).Right().TypeName()
}
func (e EitherVal) LeftTypeNat() d.TyNat {
	return e().(PairVal).Left().TypeNat()
}
func (e EitherVal) RightTypeNat() d.TyNat {
	return e().(PairVal).Right().TypeNat()
}
func (e EitherVal) LeftTypeFnc() TyFnc {
	return e().(PairVal).Left().TypeFnc()
}
func (e EitherVal) RightTypeFnc() TyFnc {
	return e().(PairVal).Right().TypeFnc()
}

/// LEFT
func NewLeft(expr Callable) LeftVal {
	return func(args ...Callable) Callable {
		return expr.Call(args...)
	}
}
func (j LeftVal) String() string                 { return j().String() }
func (j LeftVal) TypeName() string               { return "Left " + j().TypeName() }
func (j LeftVal) Call(args ...Callable) Callable { return j(args...) }
func (j LeftVal) Eval() d.Native                 { return j().Eval() }
func (j LeftVal) TypeNat() d.TyNat               { return d.Function }
func (j LeftVal) TypeFnc() TyFnc                 { return Left }

/// RIGHT
func NewRight(expr Callable) RightVal {
	return func(args ...Callable) Callable {
		return expr.Call(args...)
	}
}
func (j RightVal) String() string                 { return j().String() }
func (j RightVal) TypeName() string               { return "Right " + j().TypeName() }
func (j RightVal) Call(args ...Callable) Callable { return j(args...) }
func (j RightVal) Eval() d.Native                 { return j().Eval() }
func (j RightVal) TypeNat() d.TyNat               { return d.Function }
func (j RightVal) TypeFnc() TyFnc                 { return Right }
