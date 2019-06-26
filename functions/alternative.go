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
	CaseSwitch func(...Callable) (Callable, Callable, bool)

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
func (p PredictArg) Eval(args ...d.Native) d.Native {
	return d.BoolVal(p(NewNative(args...)))
}
func (p PredictArg) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return NewNative(d.BoolVal(p(args[0])))
	}
	return NewNative(d.BoolVal(false))
}

// return single argument predicate as multi argument predicate
func (p PredictArg) Nargs() PredictNar {
	return func(args ...Callable) bool {
		if len(args) > 0 {
			return p(args[0])
		}
		return false
	}
}
func (p PredictArg) String() string   { return "T → Predicate → Truth" }
func (p PredictArg) TypeName() string { return p.String() }
func (p PredictArg) TypeFnc() TyFnc   { return Predicate }
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
func (p PredictNar) Eval(args ...d.Native) d.Native {
	return d.BoolVal(p(NewNative(args...)))
}
func (p PredictNar) String() string   { return "[T] →  Predicate → Truth" }
func (p PredictNar) TypeName() string { return p.String() }
func (p PredictNar) TypeFnc() TyFnc   { return Predicate }
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

func (p PredictAll) Eval(args ...d.Native) d.Native {
	return d.BoolVal(p(NewNative(args...)))
}
func (p PredictAll) TypeName() string { return "[T] → (All Predicate) → Truth" }
func (p PredictAll) String() string   { return p.TypeName() }
func (p PredictAll) TypeFnc() TyFnc   { return Predicate }
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
func (p PredictAny) Eval(args ...d.Native) d.Native {
	return d.BoolVal(p(NewNative(args...)))
}
func (p PredictAny) TypeName() string { return "[T] → (Any Predicate) → Truth" }
func (p PredictAny) String() string   { return p.TypeName() }
func (p PredictAny) TypeFnc() TyFnc   { return Predicate }
func (p PredictAny) TypeNat() d.TyNat { return d.Function }

//// CASE EXPRESSION & SWITCH
///
// eval converts its arguments to callable and evaluates the result to yield a
// return value of native type
func NewCase(predicate PredictNar, exprs ...Callable) CaseExpr {
	return func(args ...Callable) (Callable, bool) {
		if len(args) > 0 {
			if predicate(args...) {
				if len(exprs) > 0 {
					if len(exprs) > 1 {
						return Curry(exprs...).Call(args...), true
					}
					return exprs[0].Call(args...), true
				}
				if len(args) > 1 {
					return NewVector(args...), true
				}
				return args[0], true
			}
			if len(args) > 1 {
				return NewVector(args...), false
			}
			return args[0], false
		}
		// return predicate and case expressions, when called without
		// arguments
		return NewPair(predicate, NewVector(exprs...)), false
	}
}

func (s CaseExpr) Decompose() (Callable, []Callable) {
	pair, _ := s()
	return pair.(PairVal).Left().(Callable),
		pair.(PairVal).Right().(VecCol)()
}
func (s CaseExpr) Eval(args ...d.Native) d.Native {
	var pair, _ = s()
	return pair.(Paired).Left().(Callable).Eval(args...)
}
func (s CaseExpr) Call(args ...Callable) Callable {
	if len(args) > 0 {
		_, ok := s(args...)
		return New(ok)
	}
	var pair, _ = s()
	return pair
}
func (s CaseExpr) TypeName() string {
	var name string
	var pred = s.Predicate().TypeName()
	var exprs = s.Expressions()
	for n, expr := range exprs {
		name = "(" + expr.TypeName() + ")" + name
		if n < len(exprs)-1 {
			name = " → " + name
		}
	}
	return "Case " + pred + " ⇒ [T] → " + name
}
func (s CaseExpr) String() string          { return s.TypeName() }
func (s CaseExpr) Predicate() Callable     { pred, _ := s.Decompose(); return pred }
func (s CaseExpr) Expressions() []Callable { _, exps := s.Decompose(); return exps }
func (s CaseExpr) TypeNat() d.TyNat        { return d.Function }
func (s CaseExpr) TypeFnc() TyFnc          { return Case }

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
	return func(args ...Callable) (Callable, Callable, bool) {
		// call consumeable to yield current case and reassign list of
		// remaining cases
		current, cases = cases()

		var argvec = NewVector(args...)
		var idx = 0

		// if call yielded any case
		if current != nil {
			// scrutinize argument(s) by applying the case
			if len(args) > 0 {
				if result, ok := current.(CaseExpr)(args...); ok {
					// replenish cases before returning
					// successfully scrutinized arguments
					idx = 0
					cases = NewList(caseslice...)
					// return result of case application,
					// index position of matching case and
					// the case that matched
					return result, NewPair(New(idx), current), true
				}
				// update case index
				idx += 1
				// return set of arguments and false indicator. don't
				// replenish cases of the parially applyed.
				return argvec, cases, false
			}
			// when no arguments where passed, return list of
			// remaining cases cases & false
			return cases, current, false
		}
		// all case are depleted not scrutinizeing the arguments →
		// replenish list of cases befor returning the final result
		idx = 0
		cases = NewList(caseslice...)
		// return final none and false indicator
		return NewNone(), cases, false
	}
}

// call method iterates over all cases until either boolean indicates
// scrutinized arguments to return, or cases depletion
func (s CaseSwitch) MatchNext(args ...Callable) (Callable, Callable) {
	// call switch instance passing the arguments
	var result, matcher, ok = s(args...)
	// while call not yields none
	for !result.TypeFnc().Match(None) {
		// if boolean indicates success
		if ok {
			// return set of arguments
			return result, matcher
		}
		// otherwise call switch again to scrutinize next case
		result, matcher, ok = s(args...)
	}
	return NewNone(), matcher
}

func (s CaseSwitch) MatchAll(args ...Callable) (Callable, Callable) {

	var result, matcher Callable
	var ok bool

	// if arguments have been passed
	if len(args) > 0 {
		// call switch instance passing the arguments
		result, matcher, ok = s(args...)
		// while call not yields none
		for !result.TypeFnc().Match(None) {
			// if boolean indicates success
			if ok {
				// return set of arguments
				return result, matcher
			}
			// otherwise call switch again to scrutinize next case
			result, matcher, ok = s(args...)
		}
	}
	// return none if all cases are scrutinized, or no arguments where
	// passed
	return NewNone(), matcher
}

func (s CaseSwitch) Call(args ...Callable) Callable {
	var result, _ = s.MatchAll(args...)
	return result
}

// eval converts its arguments to callable and evaluates the result to yield a
// return value of native type
func (s CaseSwitch) Eval(args ...d.Native) d.Native { return d.NewNil() }
func (s CaseSwitch) TypeFnc() TyFnc                 { return Switch }
func (s CaseSwitch) TypeNat() d.TyNat               { return d.Function }
func (s CaseSwitch) String() string {
	return "[T] → (Case Switch) → (T, [T]) "
}

/// MAYBE
func NewMaybe(cas CaseExpr) MaybeVal {
	return func(args ...Callable) Callable {
		if len(args) > 0 {
			if val, ok := cas(args...); ok {
				return NewJust(val)
			}
			return NewNone()
		}
		var exprs = cas.Expressions()
		if len(exprs) > 1 {
			return NewVector(exprs...)
		}
		return exprs[0]
	}
}

func (m MaybeVal) Call(args ...Callable) Callable { return m(args...) }
func (m MaybeVal) Eval(args ...d.Native) d.Native { return m().Eval(args...) }
func (m MaybeVal) TypeNat() d.TyNat               { return m().TypeNat() }
func (m MaybeVal) ElemType() TyFnc                { return m().TypeFnc() }
func (m MaybeVal) TypeFnc() TyFnc                 { return Maybe }
func (m MaybeVal) String() string                 { return m().String() }
func (m MaybeVal) TypeName() string               { return "Maybe " + m().TypeName() }

/// JUST
func NewJust(expr Callable) JustVal {
	return func(args ...Callable) Callable {
		return expr.Call(args...)
	}
}
func (j JustVal) Call(args ...Callable) Callable { return j(args...) }
func (j JustVal) Eval(args ...d.Native) d.Native { return j().Eval(args...) }
func (j JustVal) TypeNat() d.TyNat               { return j().TypeNat() }
func (j JustVal) ElemType() TyFnc                { return j().TypeFnc() }
func (j JustVal) TypeFnc() TyFnc                 { return Just }
func (j JustVal) String() string                 { return j().String() }
func (j JustVal) TypeName() string               { return "Just " + j().TypeName() }

//// EITHER
///
// either takes a case expression passes its arguments & applys the result to
// either the left, or right constructor to yield the final value. the value
// wrapping might be handled by the case expression allready, in which case nil
// can be passed to NewEither for left, right, or both constructor and the
// return from the case switch will be passed directly the left, or right
// value constructor directly.
func NewEither(cas CaseExpr, left, right Callable) EitherVal {

	return func(args ...Callable) Callable {

		var val Callable
		var ok bool

		if len(args) > 0 {
			if val, ok = cas(args...); ok {
				if left == nil {
					return NewLeft(val)
				}
				return NewLeft(left.Call(val))
			}
			if right == nil {
				return NewRight(val)
			}
			return NewRight(right.Call(args...))
		}
		// no arguments, wrap both types in a pair of pairs if
		// expression to apply on result before either left, or right
		// values are created is omitted, it needs to be set by the
		// expression(s) provided in case, or replaced by a Unary that
		// just returns it's argument
		if left == nil {

			var exprs = cas.Expressions()

			switch {
			case len(exprs) > 1:
				left = Curry(exprs...)
			case len(exprs) == 1:
				left = exprs[0]
			default:
				left = NewUnary(
					func(arg Callable) Callable {
						return arg
					})
			}
		}
		if right == nil {

			var exprs = cas.Expressions()

			switch {
			case len(exprs) > 1:
				right = Curry(exprs...)
			case len(exprs) == 1:
				right = exprs[0]
			default:
				right = NewUnary(
					func(arg Callable) Callable {
						return arg
					})
			}
		}
		return NewPair(left, right)
	}
}
func (e EitherVal) Call(args ...Callable) Callable { return e(args...) }
func (e EitherVal) String() string                 { return e().String() }
func (e EitherVal) Eval(args ...d.Native) d.Native { return e().Eval(args...) }
func (e EitherVal) TypeNat() d.TyNat               { return e().TypeNat() }
func (e EitherVal) TypeFnc() TyFnc                 { return Either }
func (e EitherVal) TypeName() string {
	return "Either " + e.LeftType().TypeName() +
		" Or " + e.RightType().TypeName()
}
func (e EitherVal) LeftType() TyFnc  { return e().(Paired).Left().TypeFnc() }
func (e EitherVal) RightType() TyFnc { return e().(Paired).Right().TypeFnc() }
func (e EitherVal) LeftTypeNat() d.TyNat {
	return e().(PairVal).Left().TypeNat()
}
func (e EitherVal) RightTypeNat() d.TyNat {
	return e().(PairVal).Right().TypeNat()
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
func (j LeftVal) Eval(args ...d.Native) d.Native { return j().Eval(args...) }
func (j LeftVal) TypeNat() d.TyNat               { return j().TypeNat() }
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
func (j RightVal) Eval(args ...d.Native) d.Native { return j().Eval(args...) }
func (j RightVal) TypeNat() d.TyNat               { return j().TypeNat() }
func (j RightVal) TypeFnc() TyFnc                 { return Right }
