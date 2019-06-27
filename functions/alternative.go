package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// JUST VALUE CONSTRUCTOR
	OptionVal func(...Callable) Callable
	//// TRUTH VALUE CONSTRUCTOR
	TruthVal func(...Callable) TyFnc
	//// NONE VALUE CONSTRUCTOR
	NoneVal func()

	//// CASE & SWITCH TYPE CONSTRUCTORS
	CaseExpr   func(...Callable) (Callable, bool)
	CaseSwitch func(...Callable) (Callable, Callable, bool)

	//// MAYBE TYPE CONSTRUCTOR
	MaybeType func(...Callable) Callable

	//// EITHER TYPE CONSTRUCTOR
	OptionType func(...Callable) Callable

	//// ALTERNATIVE DECLARATION
)

//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type. implements consumeable,
// key-, index & generic pair interface to be returneable as such.
func NewNone() NoneVal { return func() {} }

func (n NoneVal) Ident() Callable                  { return n }
func (n NoneVal) Len() int                         { return 0 }
func (n NoneVal) TypeName() string                 { return "⊥" }
func (n NoneVal) Eval(args ...d.Native) d.Native   { return nil }
func (n NoneVal) Call(...Callable) Callable        { return nil }
func (n NoneVal) Key() Callable                    { return nil }
func (n NoneVal) Index() Callable                  { return nil }
func (n NoneVal) Left() Callable                   { return nil }
func (n NoneVal) Right() Callable                  { return nil }
func (n NoneVal) Both() Callable                   { return nil }
func (n NoneVal) Value() Callable                  { return nil }
func (n NoneVal) Empty() bool                      { return true }
func (n NoneVal) TypeNat() d.TyNat                 { return d.Nil }
func (n NoneVal) String() string                   { return n.String() }
func (n NoneVal) Head() Callable                   { return NewNone() }
func (n NoneVal) Tail() Consumeable                { return NewNone() }
func (n NoneVal) Consume() (Callable, Consumeable) { return NewNone(), NewNone() }
func (n NoneVal) TypeFnc() TyFnc                   { return None }
func (n NoneVal) Type() Typed {
	return Define(n.TypeName(), None)
}

//// TRUTH VALUE CONSTRUCTOR
func NewTruth(test func(...Callable) bool) TruthVal {
	return func(args ...Callable) TyFnc {
		if len(args) > 0 {
			if test(args...) {
				return True
			}
			return False
		}
		return Truth
	}
}

func NewTrinaryTruth(test func(...Callable) int) TruthVal {
	return func(args ...Callable) TyFnc {
		if len(args) > 0 {
			if test(args...) < 0 {
				return False
			}
			if test(args...) > 0 {
				return True
			}
			return Undecided
		}
		return Trinary
	}
}

func (t TruthVal) TypeNat() d.TyNat { return d.Function }
func (t TruthVal) String() string   { return t().TypeName() }
func (t TruthVal) TypeFnc() TyFnc {
	if t().Match(Trinary) {
		return Trinary
	}
	return Truth
}
func (t TruthVal) Call(args ...Callable) Callable { return t(args...) }
func (t TruthVal) Eval(args ...d.Native) d.Native {
	if t(NewNative(args...)).TypeFnc() == True {
		return d.BoolVal(true)
	}
	if t(NewNative(args...)).TypeFnc() == False {
		return d.BoolVal(false)
	}
	return d.NewNil()
}
func (t TruthVal) TypeName() string {
	if t() == Trinary {
		return "T → Truth → True | False | Undecided"
	}
	return "T → Truth → True | False"
}
func (t TruthVal) Type() Typed {
	if t() == Trinary {
		return TyDef(func() (string, Callable) {
			return t.TypeName(), NewVector(True, False, Undecided)
		})
	}
	return TyDef(func() (string, Callable) {
		return t.TypeName(), NewPair(True, False)
	})
}
func (t TruthVal) Bool(args ...Callable) bool {
	if t(args...) == True {
		return true
	}
	return false
}

//// CASE EXPRESSION & SWITCH
///
// eval converts its arguments to callable and evaluates the result to yield a
// return value of native type
func NewCase(test TruthVal, exprs ...Callable) CaseExpr {
	return func(args ...Callable) (Callable, bool) {
		if len(args) > 0 {
			if test.Bool(args...) {
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
		return NewPair(test, NewVector(exprs...)), false
	}
}

func (s CaseExpr) Test() Callable    { pred, _ := s.Decompose(); return pred }
func (s CaseExpr) Exprs() []Callable { _, exps := s.Decompose(); return exps }
func (s CaseExpr) TypeFnc() TyFnc    { return Case }
func (s CaseExpr) TypeNat() d.TyNat  { return d.Function }
func (s CaseExpr) Decompose() (Callable, []Callable) {
	pair, _ := s()
	return pair.(PairVal).Left().(Callable),
		pair.(PairVal).Right().(VecCol)()
}
func (s CaseExpr) Eval(args ...d.Native) d.Native {
	var pair, _ = s()
	return pair.(Paired).Left().(Evaluable).Eval(args...)
}
func (s CaseExpr) Call(args ...Callable) Callable {
	if len(args) > 0 {
		_, ok := s(args...)
		return New(ok)
	}
	var pair, _ = s()
	return pair
}
func (s CaseExpr) String() string { return s.TypeName() }
func (s CaseExpr) TypeName() string {
	var name string
	var test = s.Test().TypeName()
	var exprs = s.Exprs()
	for n, expr := range exprs {
		name = name + " | " + expr.TypeName()
		if n < len(exprs)-1 {
			name = "\n" + name
		}
	}
	return "Case " + test + " ⇒ " + name
}
func (s CaseExpr) Type() Typed {

	var types = []Callable{}

	for _, expr := range s.Exprs() {
		types = append(types, expr.Type().(TyDef))
	}

	return TyDef(func() (string, Callable) {
		return s.TypeName(), NewVector(types...)
	})
}

// applys passed arguments to all enclosed cases in the order passed to the
// switch constructor
func NewSwitch(inis ...CaseExpr) CaseSwitch {

	var num = len(inis)

	var tempt, tempc = make([]Callable, 0, num),
		make([]Callable, 0, num)
	num = 0

	for _, arg := range inis {
		tempt = append(tempt, arg.Type().(TyDef))
		tempc = append(tempc, arg)
	}

	var cases = NewVector(tempc...)
	var types = NewVector(tempt...)

	// create and return case switch enclosing list of predicates
	return func(args ...Callable) (Callable, Callable, bool) {

		var reargs = NewVector(args...)

		// if call yielded any case
		if len(args) > 0 {
			// scrutinize argument(s) by applying the case
			if cases.Len() > 0 {
				num += num
				var current, cases = cases.Consume()
				if result, ok := current.(CaseExpr)(args...); ok {
					return result, NewIndexPair(num, current), true
				}
				// return set of arguments and false indicator. don't
				// replenish cases of the parially applyed.
				return reargs, cases, false
			}
		}
		// when no arguments where passed, return list of
		// remaining cases cases & false
		return types, cases, false
	}
}

// call method iterates over all cases until either boolean indicates
// scrutinized arguments to return, or cases depletion
func (s CaseSwitch) MatchNext(args ...Callable) (
	result Callable,
	tests Callable,
	ok bool,
) {
	if len(args) > 0 {
		// call switch instance passing the arguments
		result, tests, ok = s(args...)
		// while call not yields none
		for !result.TypeFnc().Match(None) {
			// if boolean indicates success
			if ok {
				// return set of arguments
				return result, tests, ok
			}
			// otherwise call switch again to scrutinize next case
			result, tests, ok = s(args...)
		}
	}
	return s()
}

func (s CaseSwitch) MatchAll(args ...Callable) (Callable, Callable, bool) {

	// safe count, to eventually return
	var types, cases, ok = s()
	var count = cases.(VecCol).Len()

	// if arguments have been passed
	if len(args) > 0 {

		var current, result Callable

		// if there are further cases left
		if count > 0 {
			// while call not yields none
			current, cases = cases.(VecCol).Consume()

			if current != nil {
				// if boolean indicates success
				if result, ok = current.(CaseExpr)(); ok {
					// return set of arguments
					return result, NewPair(New(count), current), ok
				}
			}
			return NewVector(args...), cases, ok
		}
	}
	// return none if all cases are scrutinized, or no arguments where
	// passed
	return NewVector(args...), types, ok
}

func (s CaseSwitch) Call(args ...Callable) Callable {
	result, _, _ := s.MatchAll(args...)
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
func NewMaybe(cas CaseExpr, types ...Typed) MaybeType {

	// no arguments where passed, return type and/or expected
	// argument signature
	var typeVec VecCol

	if len(types) > 0 {

		var ctype Callable
		var vals = []Callable{}

		for _, typ := range types {
			if t, ok := typ.(TyDef); ok {
				ctype = t
			} else {
				ctype = typ.(TyFnc)
			}
			vals = append(vals, ctype)
		}
		typeVec = NewVector(vals...)
	}

	return func(args ...Callable) Callable {

		if len(args) > 0 {
			if val, ok := cas(args...); ok {
				return NewJust(val)
			}
			return NewNone()
		}
		return typeVec
	}
}

func (m MaybeType) Eval(args ...d.Native) d.Native { return m().Eval(args...) }
func (m MaybeType) Expr() Callable                 { return m() }
func (m MaybeType) TypeNat() d.TyNat               { return m().TypeNat() }
func (m MaybeType) TypeFnc() TyFnc                 { return Maybe }
func (m MaybeType) ElemType() Typed                { return m() }
func (m MaybeType) String() string                 { return m().String() }
func (m MaybeType) TypeName() string {

	var l = m().(VecCol).Len()

	if l > 0 {

		var head Callable
		var cons = m().(Consumeable)

		head, cons = cons.Consume()
		if l > 0 {

			var jname, argn string
			head, cons = cons.Consume()

			if l > 2 {

				head, cons = cons.Consume()

				for head != nil {
					argn = head.TypeName() + " → "
					head, cons = cons.Consume()
				}

			} else {
				argn = head.TypeName() + " → "
			}
			return argn + " Maybe → Just " + jname + " | None"
		}
		return "[T] → Maybe → Just " +
			cons.Head().TypeName() +
			" | None"
	}
	return "[T] → Maybe → Just | None"
}
func (m MaybeType) Type() Typed {
	return TyDef(func() (string, Callable) {
		return m().TypeName(), m.ElemType().(TyDef)
	})
}

/// JUST
func NewJust(expr Callable) OptionVal {
	return func(args ...Callable) Callable {
		return expr.Call(args...)
	}
}
func (j OptionVal) Call(args ...Callable) Callable { return j(args...) }
func (j OptionVal) Eval(args ...d.Native) d.Native { return j().Eval(args...) }
func (j OptionVal) TypeNat() d.TyNat               { return j().TypeNat() }
func (j OptionVal) ElemType() TyFnc                { return j().TypeFnc() }
func (j OptionVal) TypeFnc() TyFnc                 { return Just }
func (j OptionVal) String() string                 { return j().String() }
func (j OptionVal) TypeName() string               { return "Just " + j().TypeName() }
func (j OptionVal) Type() Typed {
	return TyDef(func() (string, Callable) { return j().TypeName(), j.ElemType() })
}

//// OPTION TYPE CONSTRUCTOR
///
func NewOptionType(test CaseExpr, either, or Callable) OptionType {
	return func(args ...Callable) Callable {
		var option Callable
		return option
	}
}
func (o OptionType) Expr() Callable                 { return o() }
func (o OptionType) TypeFnc() TyFnc                 { return Option }
func (o OptionType) TypeNat() d.TyNat               { return d.Function }
func (o OptionType) String() string                 { return o.TypeName() }
func (o OptionType) Eval(args ...d.Native) d.Native { return o().Eval(args...) }
func (o OptionType) Call(args ...Callable) Callable { return o(args...) }
func (o OptionType) TypeName() string {
	var str string
	return str
}
func (o OptionType) Type() Typed {
	var opt OptionType
	return opt
}
