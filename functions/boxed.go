package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// FUNCTOR CONSTRUCTORS
	///
	// CONDITIONAL, CASE & OPTIONAL

	///// STATIC
	NoOp      func()
	OptVal    func() PairFnc
	EnumVal   func() PairFnc
	RecordVal func() PairFnc
	TrueFalse func() bool

	///// DYNAMIC
	TruthFnc func(...Callable) bool
	JustNone func(...Callable) OptVal
	EitherOr func(...Callable) OptVal
	CaseExpr func(...Callable) OptVal

	/// ERROR
)

// NULL NADA NONE NIENTE ZERO NAN
func NewNoOp() NoOp                      { return NoOp(func() {}) }
func (n NoOp) Maybe() bool               { return false }
func (n NoOp) Empty() bool               { return true }
func (n NoOp) String() string            { return "⊥" }
func (n NoOp) Len() int                  { return -2 }
func (n NoOp) Value() Callable           { return n }
func (n NoOp) Ident() Callable           { return n }
func (n NoOp) Call(...Callable) Callable { return n }
func (n NoOp) Eval(...d.Native) d.Native { return d.NilVal{} }
func (n NoOp) TypeNat() d.TyNative       { return d.Nil }
func (n NoOp) TypeFnc() TyFnc            { return None }

/// BASE TRUTH
func NewBaseTruth(truth bool) TrueFalse {
	if truth {
		return func() bool { return true }
	}
	return func() bool { return false }
}

func (t TrueFalse) Eval(...d.Native) d.Native { return d.BoolVal(t()) }
func (t TrueFalse) Call(...Callable) Callable { return t }
func (t TrueFalse) Ident() Callable           { return t }
func (t TrueFalse) TypeNat() d.TyNative       { return d.Bool }
func (t TrueFalse) TypeFnc() TyFnc {
	if t() {
		return True
	}
	return False
}

func (t TrueFalse) String() string {
	if t() {
		return "True"
	}
	return "False"
}

//// TRUTH PREDICATE FUNCTION
func NewTruthFunction(predicate Callable) TruthFnc {
	return func(args ...Callable) bool {
		if truth, ok := predicate.Call(args...).Eval().(d.BoolVal); ok {
			return bool(truth)
		}
		return false
	}
}

func (t TruthFnc) Ident() Callable                { return t }
func (t TruthFnc) String() string                 { return "Truth Function" }
func (t TruthFnc) TypeFnc() TyFnc                 { return Truth }
func (t TruthFnc) TypeNat() d.TyNative            { return d.Bool }
func (t TruthFnc) Call(args ...Callable) Callable { return New(t(args...)) }
func (t TruthFnc) Eval(args ...d.Native) d.Native { return t.Eval(args...) }

/// OPTIONAL VALUES
func NewOptVal(left, right Callable) OptVal {
	return OptVal(func() PairFnc {
		return NewPair(left, right)
	})
}

func (o OptVal) Ident() Callable     { return o }
func (o OptVal) Left() Callable      { return o().Left() }
func (o OptVal) Right() Callable     { return o().Right() }
func (o OptVal) TypeNat() d.TyNative { return o.Right().TypeNat() }
func (o OptVal) TypeFnc() TyFnc      { return o.Right().TypeFnc() | Option }
func (o OptVal) String() string {
	return "Optional " + o.Left().String() + " " + o.Right().String()
}

func (o OptVal) Call(args ...Callable) Callable { return o.Right().Call(args...) }

func (o OptVal) Eval(vars ...d.Native) d.Native {
	var args = []Callable{}
	for _, arg := range vars {
		args = append(args, NewFromData(arg))
	}
	return o.Call(args...)
}

/// ENUM VALUES
func NewEnumVal(key, val Callable) EnumVal {
	return EnumVal(
		func() PairFnc {
			return NewPair(key, val)
		})
}

func (o EnumVal) Ident() Callable     { return o }
func (o EnumVal) Left() Callable      { return o().Left() }
func (o EnumVal) Key() Callable       { return o.Left() }
func (o EnumVal) Right() Callable     { return o().Right() }
func (o EnumVal) Value() Callable     { return o.Right() }
func (o EnumVal) TypeNat() d.TyNative { return o.Right().TypeNat() }
func (o EnumVal) TypeFnc() TyFnc      { return o.Right().TypeFnc() | Enum }
func (o EnumVal) String() string {
	return o.Left().String() + "·" + o.Right().String()
}

func (o EnumVal) Call(args ...Callable) Callable { return o.Right().Call(args...) }

func (o EnumVal) Eval(vars ...d.Native) d.Native {
	var args = []Callable{}
	for _, arg := range vars {
		args = append(args, NewFromData(arg))
	}
	return o.Call(args...)
}

/// RECORD VALUES
func NewRecordVal(key, val Callable) RecordVal {
	return RecordVal(func() PairFnc {
		return NewPair(key, val)
	})
}

func (o RecordVal) Ident() Callable     { return o }
func (o RecordVal) Left() Callable      { return o().Left() }
func (o RecordVal) Key() Callable       { return o.Left() }
func (o RecordVal) Right() Callable     { return o().Right() }
func (o RecordVal) Value() Callable     { return o.Right() }
func (o RecordVal) TypeNat() d.TyNative { return o.Right().TypeNat() }
func (o RecordVal) TypeFnc() TyFnc      { return o.Right().TypeFnc() | Record }
func (o RecordVal) String() string {
	return o.Left().String() + "∷ " + o.Right().String()
}

func (o RecordVal) Call(args ...Callable) Callable { return o.Right().Call(args...) }

func (o RecordVal) Eval(vars ...d.Native) d.Native {
	var args = []Callable{}
	for _, arg := range vars {
		args = append(args, NewFromData(arg))
	}
	return o.Call(args...)
}

//// JUST NONE
///
// expression is applyed to arguments passed at runtime. result of calling the
// expression is applyed to predex. if the predicate matches, result is
// returned as 'just' value, otherwise NoOp is returned
func NewJustNone(test TruthFnc, expr Callable) JustNone {
	return JustNone(func(args ...Callable) OptVal {
		var result = expr.Call(args...)
		if test(result) {
			return NewOptVal(New(true), result)
		}
		return NewOptVal(New(false), NewNoOp())
	})
}

func (j JustNone) Ident() Callable                { return j }
func (j JustNone) String() string                 { return "Just None" }
func (j JustNone) TypeFnc() TyFnc                 { return Option | Just | None }
func (j JustNone) TypeNat() d.TyNative            { return d.Function }
func (j JustNone) Return(args ...Callable) OptVal { return j(args...) }
func (j JustNone) Call(args ...Callable) Callable { return j(args...) }
func (j JustNone) Eval(vars ...d.Native) d.Native {
	var args = []Callable{}
	for _, v := range vars {
		args = append(args, NewFromData(v))
	}
	return j(args...)
}

//// EITHER OR
///
// left pair value indicates: 0 = 'either', 1 = 'or', -1 = 'no value yielded'
func NewEitherOr(predex TruthFnc, either, or Callable) EitherOr {
	return EitherOr(func(args ...Callable) OptVal {
		var val Callable

		val = either.Call(args...)

		if predex(val) {
			return NewOptVal(New(0), val)
		}

		val = or.Call(args...)

		if predex(val) {
			return NewOptVal(New(1), val)
		}

		return NewOptVal(New(-1), NewNoOp())
	})
}

func (e EitherOr) Ident() Callable                { return e }
func (e EitherOr) String() string                 { return "either-or" }
func (e EitherOr) TypeFnc() TyFnc                 { return Option | Either | Or }
func (e EitherOr) TypeNat() d.TyNative            { return d.Function }
func (e EitherOr) Return(args ...Callable) OptVal { return e(args...) }
func (e EitherOr) Call(args ...Callable) Callable { return e(args...) }
func (e EitherOr) Eval(vars ...d.Native) d.Native {
	var args = []Callable{}
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
func NewSwitchCase(predex TruthFnc, value Callable, nextcase ...CaseExpr) CaseExpr {
	// if runtime arguments applyed to predicate expression yields true, value
	// will be returned, or otherwise the next case will be the return value.
	return CaseExpr(func(args ...Callable) OptVal {
		if predex(args...) { // return value if runtime args match predicate
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

func (s CaseExpr) Ident() Callable                { return s }
func (s CaseExpr) String() string                 { return "switch-case" }
func (s CaseExpr) TypeFnc() TyFnc                 { return Option | Case | Switch }
func (s CaseExpr) TypeNat() d.TyNative            { return d.Function }
func (s CaseExpr) Return(args ...Callable) OptVal { return s(args...) }
func (s CaseExpr) Call(args ...Callable) Callable { return s(args...) }
func (s CaseExpr) Eval(vars ...d.Native) d.Native {
	var args = []Callable{}
	for _, v := range vars {
		args = append(args, NewFromData(v))
	}
	return s.Call(args...)
}
