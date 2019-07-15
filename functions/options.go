package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// NONE VALUE CONSTRUCTOR
	NoneVal func()

	// TESTS AND COMPARE
	TestVal func(...Expression) bool
	TrinVal func(...Expression) int
	CompVal func(...Expression) int

	// CASE & SWITCH
	CaseVal   func(...Expression) Expression
	SwitchVal func(...Expression) (Expression, SwitchVal)

	// MAYBE (JUST | NONE)
	JustVal  func(...Expression) Expression
	MaybeVal func(...Expression) Expression

	// OPTION (EITHER | OR)
	OrVal     func(...Expression) Expression
	EitherVal func(...Expression) Expression
	OptionVal func(...Expression) Expression
)

//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type. implements consumeable,
// key-, index & generic pair interface to be returneable as such.
func NewNone() NoneVal { return func() {} }

func (n NoneVal) Head() Expression              { return n }
func (n NoneVal) Tail() Consumeable             { return n }
func (n NoneVal) Len() d.IntVal                 { return 0 }
func (n NoneVal) String() string                { return "⊥" }
func (n NoneVal) Call(...Expression) Expression { return nil }
func (n NoneVal) Key() Expression               { return nil }
func (n NoneVal) Index() Expression             { return nil }
func (n NoneVal) Left() Expression              { return nil }
func (n NoneVal) Right() Expression             { return nil }
func (n NoneVal) Both() Expression              { return nil }
func (n NoneVal) Value() Expression             { return nil }
func (n NoneVal) Empty() d.BoolVal              { return true }
func (n NoneVal) Test(...Expression) bool       { return false }
func (n NoneVal) Compare(...Expression) int     { return -1 }
func (n NoneVal) Flag() d.BitFlag               { return d.BitFlag(None) }
func (n NoneVal) TypeFnc() TyFnc                { return None }
func (n NoneVal) TypeNat() d.TyNat              { return d.Nil }
func (n NoneVal) TypeElem() d.Typed             { return None }
func (n NoneVal) TypeName() string              { return n.String() }
func (n NoneVal) FlagType() d.Uint8Val          { return Flag_Function.U() }
func (n NoneVal) Type() TyPattern               { return ConPattern(None) }
func (n NoneVal) Consume() (Expression, Consumeable) {
	return NewNone(), NewNone()
}

/// TEST
func NewTest(test func(...Expression) bool) TestVal {
	return func(args ...Expression) bool {
		return test(args...)
	}
}
func (t TestVal) TypeFnc() TyFnc { return Truth }
func (t TestVal) Type() TyPattern {
	return ConPattern(t.TypeFnc(), ConPattern(True, False))
}
func (t TestVal) String() string               { return t.TypeFnc().TypeName() }
func (t TestVal) Test(args ...Expression) bool { return t(args...) }
func (t TestVal) Compare(args ...Expression) int {
	if t(args...) {
		return 0
	}
	return -1
}
func (t TestVal) Call(args ...Expression) Expression {
	return NewData(d.BoolVal(t(args...)))
}

/// TRINARY TEST
func NewTrinary(test func(...Expression) int) TrinVal {
	return func(args ...Expression) int {
		return test(args...)
	}
}
func (t TrinVal) TypeFnc() TyFnc { return Trinary }
func (t TrinVal) Type() TyPattern {
	return ConPattern(t.TypeFnc(), ConPattern(True, False, Undecided))
}
func (t TrinVal) String() string                     { return t.TypeFnc().TypeName() }
func (t TrinVal) Test(args ...Expression) bool       { return t(args...) == 0 }
func (t TrinVal) Compare(args ...Expression) int     { return t(args...) }
func (t TrinVal) Call(args ...Expression) Expression { return NewData(d.IntVal(t(args...))) }

/// COMPARE
func NewCompare(comp func(...Expression) int) CompVal {
	return func(args ...Expression) int {
		return comp(args...)
	}
}
func (t CompVal) TypeFnc() TyFnc { return Compare }
func (t CompVal) Type() TyPattern {
	return ConPattern(t.TypeFnc(), ConPattern(Lesser, Greater, Equal))
}
func (t CompVal) String() string                     { return t.TypeFnc().TypeName() }
func (t CompVal) Test(args ...Expression) bool       { return t(args...) == 0 }
func (t CompVal) Compare(args ...Expression) int     { return t(args...) }
func (t CompVal) Call(args ...Expression) Expression { return NewData(d.IntVal(t(args...))) }

/// CASE
func NewCase(test Testable, expr Expression) CaseVal {
	return func(args ...Expression) Expression {
		if len(args) == 0 {
			return NewCase(test, expr)
		}
		if test.Test(args...) {
			return expr.Call(args...)
		}
		return NewNone()
	}
}
func (t CaseVal) TypeFnc() TyFnc  { return Case }
func (t CaseVal) Type() TyPattern { return ConPattern(t.TypeFnc()) }
func (t CaseVal) String() string  { return t.TypeFnc().TypeName() }
func (t CaseVal) Test(args ...Expression) bool {
	if t(args...).Type().Match(None) {
		return false
	}
	return true
}
func (t CaseVal) Call(args ...Expression) Expression { return t(args...) }

/// SWITCH
func NewSwitch(cases ...CaseVal) SwitchVal {
	return func(args ...Expression) (Expression, SwitchVal) {
		if len(args) == 0 {
			return NewNone(),
				NewSwitch(cases...)
		}
		var current CaseVal
		if len(cases) > 0 {
			if len(cases) > 1 {
				current, cases = cases[0], cases[1:]
			}
			current, cases = cases[0], []CaseVal{}
		}
		if current != nil {
			return current.Call(args...), NewSwitch(cases...)
		}
		return NewNone(), nil
	}
}
func (t SwitchVal) TypeFnc() TyFnc  { return Switch }
func (t SwitchVal) Type() TyPattern { return ConPattern(t.TypeFnc()) }
func (t SwitchVal) String() string  { return t.TypeFnc().TypeName() }
func (t SwitchVal) Call(args ...Expression) Expression {
	var expr, swi = t(args...)
	for swi != nil {
		if !expr.TypeFnc().Match(None) {
			return expr
		}
		expr, swi = t(args...)
	}
	return NewNone()
}
