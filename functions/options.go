package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// NONE VALUE CONSTRUCTOR
	NoneVal func()
	//// TRUTH VALUE CONSTRUCTOR
	TestExpr func(...Expression) TyFnc
	// JUST VALUE CONSTRUCTOR
	OptionVal func(...Expression) Expression

	//// CASE & SWITCH TYPE CONSTRUCTORS
	CaseExpr   func(...Expression) (Expression, bool)
	CaseSwitch func(...Expression) (Expression, Expression, bool)

	//// EITHER TYPE CONSTRUCTOR
	OptionType func(...Expression) OptionVal

	//// ALTERNATIVE DECLARATION
)

//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type. implements consumeable,
// key-, index & generic pair interface to be returneable as such.
func NewNone() NoneVal { return func() {} }

func (n NoneVal) Ident() Expression              { return n }
func (n NoneVal) Head() Expression               { return n }
func (n NoneVal) Tail() Consumeable              { return n }
func (n NoneVal) Len() int                       { return 0 }
func (n NoneVal) String() string                 { return "⊥" }
func (n NoneVal) Eval(args ...d.Native) d.Native { return nil }
func (n NoneVal) Call(...Expression) Expression  { return nil }
func (n NoneVal) Key() Expression                { return nil }
func (n NoneVal) Index() Expression              { return nil }
func (n NoneVal) Left() Expression               { return nil }
func (n NoneVal) Right() Expression              { return nil }
func (n NoneVal) Both() Expression               { return nil }
func (n NoneVal) Value() Expression              { return nil }
func (n NoneVal) Empty() bool                    { return true }
func (n NoneVal) Flag() d.BitFlag                { return d.BitFlag(None) }
func (n NoneVal) TypeFnc() TyFnc                 { return None }
func (n NoneVal) TypeNat() d.TyNat               { return d.Nil }
func (n NoneVal) TypeName() string               { return n.String() }
func (n NoneVal) FlagType() d.Uint8Val           { return Flag_Functional.U() }
func (n NoneVal) Type() Typed {
	return Define(n.TypeName(), None)
}
func (n NoneVal) Consume() (Expression, Consumeable) {
	return NewNone(), NewNone()
}

//// TRUTH VALUE CONSTRUCTOR
func NewTruthTest(test func(...Expression) bool) TestExpr {
	return func(args ...Expression) TyFnc {
		if len(args) > 0 {
			if test(args...) {
				return True
			}
			return False
		}
		return Truth
	}
}

func NewTrinaryTest(test func(...Expression) int) TestExpr {
	return func(args ...Expression) TyFnc {
		if len(args) > 0 {
			if test(args...) > 0 {
				return True
			}
			if test(args...) < 0 {
				return False
			}
			return Undecided
		}
		return Trinary
	}
}

func NewCompareTest(test func(...Expression) int) TestExpr {
	return func(args ...Expression) TyFnc {
		if len(args) > 0 {
			if test(args...) > 0 {
				return Greater
			}
			if test(args...) < 0 {
				return Lesser
			}
			return Equal
		}
		return Compare
	}
}

func (t TestExpr) Call(args ...Expression) Expression { return t(args...) }
func (t TestExpr) String() string                     { return t().TypeName() }
func (t TestExpr) FlagType() d.Uint8Val               { return Flag_Functional.U() }
func (t TestExpr) TypeNat() d.TyNat                   { return d.Function }
func (t TestExpr) TypeFnc() TyFnc                     { return t() }
func (t TestExpr) Eval(args ...d.Native) d.Native {
	if t() == Compare {
		if t(NewNative(args...)).TypeFnc() == Lesser {
			return d.IntVal(-1)
		}
		if t(NewNative(args...)).TypeFnc() == Equal {
			return d.IntVal(0)
		}
		if t(NewNative(args...)).TypeFnc() == Greater {
			return d.IntVal(1)
		}
	}
	if t(NewNative(args...)).TypeFnc() == True {
		return d.BoolVal(true)
	}
	if t(NewNative(args...)).TypeFnc() == False {
		return d.BoolVal(false)
	}
	return d.NewNil()
}

func (t TestExpr) TypeName() string {
	if t() == Compare {
		return "Ord → Compare → Lesser | Greater | Equal"
	}
	if t() == Trinary {
		return "[T] → Trinary Truth → True | Undecided | False"
	}
	return "[T] → Truth → True | False"
}

func (t TestExpr) Type() Typed { return Define(t().TypeName(), t()) }

func (t TestExpr) Test(args ...Expression) bool {
	if t() == Compare {
		if t(args...) == Lesser || t(args...) == Greater {
			return false
		}
		if t(args...) == Equal {
			return true
		}
	}
	if t() == Trinary {
		if t(args...) == False || t(args...) == Undecided {
			return false
		}
		if t(args...) == True {
			return true
		}
	}
	if t(args...) != True {
		return false
	}
	return true
}

func (t TestExpr) Compare(args ...Expression) int {
	if t() == Compare {
		if t(args...) == Lesser {
			return -1
		}
		if t(args...) == Equal {
			return 0
		}
		if t(args...) == Greater {
			return 1
		}
	}
	if t() == Trinary {
		if t(args...) == False {
			return -1
		}
		if t(args...) == Undecided {
			return 0
		}
		if t(args...) == True {
			return 1
		}
	}
	if t(args...) != True {
		return -1
	}
	return 0
}

//// CASE EXPRESSION & SWITCH
///
// eval converts its arguments to callable and evaluates the result to yield a
// return value of native type
func NewCase(test TestExpr, exprs ...Expression) CaseExpr {
	var expr Expression
	if len(exprs) > 0 {
		expr = Curry(exprs...)
	}
	return func(args ...Expression) (Expression, bool) {
		if len(args) > 0 {
			if test.Test(args...) {
				if expr != nil {
					return expr.Call(args...), true
				}
				return NewVector(args...), true
			}
			if expr != nil {
				return NewVector(args...), false
			}
			return NewVector(args...), false
		}
		if expr != nil {
			return NewPair(test, expr), false
		}
		return NewPair(test, NewNone()), false
	}
}

func (s CaseExpr) TypeFnc() TyFnc       { return Case }
func (s CaseExpr) TypeNat() d.TyNat     { return d.Function }
func (s CaseExpr) FlagType() d.Uint8Val { return Flag_Functional.U() }
func (s CaseExpr) String() string       { return s.TypeName() }
func (s CaseExpr) Test() TestExpr {
	var pair, _ = s()
	return pair.(Paired).Left().(TestExpr)
}
func (s CaseExpr) Expr() Expression {
	var pair, _ = s()
	return pair.(Paired).Right()
}
func (s CaseExpr) Eval(nats ...d.Native) d.Native {
	if len(nats) > 0 {
		var args = make([]Expression, 0, len(nats))
		for _, nat := range nats {
			args = append(args, NewNative(nat))
		}
		if result, ok := s(args...); ok {
			return result
		}
	}
	return d.NewNil()
}
func (s CaseExpr) Call(args ...Expression) Expression {
	if result, ok := s(args...); ok {
		return result
	}
	return NewNone()
}
func (s CaseExpr) TypeName() string {
	var expr = s.Expr()
	var test = s.Test()
	var str = "Case [T] → (" + test.Type().TypeName() + ") → True ⇒ "
	if !expr.TypeFnc().Match(None) {
		if expr.FlagType() == Flag_Def.U() {
			str = str + expr.TypeName()
		} else {
			str = str +
				"[T] → " +
				expr.TypeFnc().TypeName() +
				" → T"
		}
	} else {
		str = str + " [T]"
	}
	return str + "\n"
}
func (s CaseExpr) Type() Typed {
	var pair, _ = s()
	return Define(s.TypeName(), pair.(Paired))
}

// applys passed arguments to all enclosed cases in the order passed to the
// switch constructor
func NewSwitch(cases ...CaseExpr) CaseSwitch {
	var exprs = make([]Expression, 0, len(cases))
	for _, cas := range cases {
		exprs = append(exprs, cas)
	}
	return conSwitch(exprs...)
}

func conSwitch(cases ...Expression) CaseSwitch {

	var casevec = NewVector(cases...)

	return func(args ...Expression) (Expression, Expression, bool) {

		var depleted = NewVector()
		var head Expression
		var current CaseExpr
		var argvec VecCol

		if len(args) > 0 {
			argvec = NewVector(args...)
			if casevec.Len() > 0 {
				head, casevec = casevec.ConsumeVec()
				current = head.(CaseExpr)
				if result, ok := current(args...); ok {
					casevec = NewVector(cases...)
					return result, NewPair(current, argvec),
						true
				}
				depleted = depleted.Append(NewPair(current, argvec))
				return NewPair(current, argvec),
					conSwitch(casevec()...),
					false
			}
			casevec = NewVector(cases...)
			return nil, NewPair(casevec, argvec), false
		}
		return nil, NewPair(casevec, depleted), false
	}
}

func (s CaseSwitch) TypeFnc() TyFnc       { return Switch }
func (s CaseSwitch) TypeNat() d.TyNat     { return d.Function }
func (s CaseSwitch) FlagType() d.Uint8Val { return Flag_Functional.U() }
func (s CaseSwitch) String() string       { return s.TypeName() }

func (s CaseSwitch) TestAll(args ...Expression) (Expression, Expression) {
	var result, caseargs Expression
	if len(args) > 0 {
		var ok bool
		result, caseargs, ok = s(args...)
		for result != nil {
			if ok {
				return result, caseargs
			}
			result, caseargs, ok = caseargs.(CaseSwitch)(args...)
		}
		return nil, caseargs
	}
	result, caseargs, _ = s()
	return result, caseargs
}

func (s CaseSwitch) Eval(nats ...d.Native) d.Native {
	if len(nats) > 0 {
		var args = make([]Expression, 0, len(nats))
		for _, nat := range nats {
			args = append(args, NewNative(nat))
		}
		var result, _ = s.TestAll(args...)
		if result != nil {
			return result
		}
	}
	return d.NewNil()
}

func (s CaseSwitch) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var result, _ = s.TestAll(args...)
		if result != nil {
			return result
		}
	}
	return NewNone()
}

func (s CaseSwitch) TypeName() string {
	return "[T] → (Case Switch) → (T, [T]) "
}
func (s CaseSwitch) Type() Typed {
	return Define(s.TypeName(), s.TypeFnc())
}

///////////////////////////////////////////////////////////////////////////////
/// OPTION TYPE CONSTRUCTOR
func NewOptionType(test CaseSwitch, types ...Expression) OptionType {
	return func(args ...Expression) OptionVal {
		return NewOptionVal(test, NewVector(types...))
	}
}

func (o OptionType) Call(args ...Expression) Expression { return o().Call(args...) }
func (o OptionType) Eval(args ...d.Native) d.Native     { return o().Eval(args...) }
func (o OptionType) Expr() Expression                   { return o() }
func (o OptionType) FlagType() d.Uint8Val               { return Flag_Def.U() }
func (o OptionType) TypeNat() d.TyNat                   { return o().TypeNat() }
func (o OptionType) TypeFnc() TyFnc                     { return Option }
func (o OptionType) ElemType() Typed                    { return o() }
func (o OptionType) String() string                     { return o().String() }
func (o OptionType) Type() Typed {
	return TyDef(func() (string, Expression) {
		return o().TypeName(), o.ElemType().(TyDef)
	})
}
func (o OptionType) TypeName() string {
	var name string
	return name
}

/// OPTION VALUE CONSTRUCTOR
func NewOptionVal(test CaseSwitch, exprs ...Expression) OptionVal {
	return func(args ...Expression) Expression {
		var expr, index = test.TestAll(args...)
		if !expr.TypeFnc().Match(None) {
			var idx = int(index.(NativeExpr)().(d.IntVal))
			var result = exprs[idx]
			if len(args) > 0 {
				return result.Call(args...)
			}
			return result
		}
		return expr
	}
}
func (o OptionVal) Call(args ...Expression) Expression { return o().Call(args...) }
func (o OptionVal) Eval(args ...d.Native) d.Native     { return o().Eval(args...) }
func (o OptionVal) TypeNat() d.TyNat                   { return o().TypeNat() }
func (o OptionVal) TypeFnc() TyFnc                     { return o(HigherOrder).TypeFnc() }
func (o OptionVal) FlagType() d.Uint8Val               { return Flag_Functional.U() }
func (o OptionVal) String() string                     { return o(HigherOrder).String() }
func (o OptionVal) TypeName() string                   { return o(HigherOrder).TypeName() }
func (o OptionVal) Type() Typed                        { return Define(o().TypeName(), o()) }
