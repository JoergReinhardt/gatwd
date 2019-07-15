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
// testable, compareable, key-, index- and generic pair interfaces to be
// returneable standing in for such expressions as return value.
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
//
// create a new test, scrutinizing its arguments and revealing true, or false
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
//
// create a trinary test, that can yield true, false, or undecided, computed by
// scrutinizing its arguments
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
//
// create a comparator expression that yields minus one in case the argument is
// lesser, zero in case its equal and plus one in case it is greater than the
// enclosed value to compare against.
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
//
// case constructor takes a test and an expression, in order for the resulting
// case instance to test its arguments and yield the result of applying those
// arguments to the expression, in case the test yielded true. otherwise the
// case will yield none.
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
//
// switch takes a slice of cases and evaluates them against its arguments to
// yield either a none value, or the result of the case application and a
// switch enclosing the remaining cases. id all cases are depleted, a none
// instance will be returned as result and nil will be yielded instead of the
// switch value
//
// when called, a switch evaluates all it's cases until it yields either
// results from applying the first case that matched the arguments, or none.
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
