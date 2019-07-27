package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// NONE VALUE CONSTRUCTOR
	NoneVal  func()
	ConstVal func() Expression
	FuncVal  func(...Expression) Expression

	// TESTS AND COMPARE
	TestVal func(...Expression) bool
	TrinVal func(...Expression) int
	CompVal func(...Expression) int

	// CASE & SWITCH
	CaseVal   func(...Expression) Expression
	SwitchVal func(...Expression) (Expression, []CaseVal)

	// OPTION ELEMENT
	ElemVal func(...Expression) (Expression, TyPattern)

	// MAYBE (JUST | NONE)
	MaybeVal func(...Expression) ElemVal

	// OPTION (EITHER | OR)
	OptionVal func(...Expression) ElemVal

	// IF (THEN | ELSE)
	IfVal func(...Expression) ElemVal
)

func NewConstant(fn func() Expression) ConstVal  { return fn }
func (c ConstVal) TypeFnc() TyFnc                { return Constant }
func (c ConstVal) Type() TyPattern               { return c().(TyPattern) }
func (c ConstVal) String() string                { return c().String() }
func (c ConstVal) Call(...Expression) Expression { return c() }

func NewFunction(fn func(...Expression) Expression, pattern TyPattern) FuncVal {
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			return fn(args...)
		}
		return pattern
	}
}
func (g FuncVal) TypeFnc() TyFnc                     { return Value }
func (g FuncVal) Type() TyPattern                    { return g().(TyPattern) }
func (g FuncVal) String() string                     { return g().String() }
func (g FuncVal) Call(args ...Expression) Expression { return g(args...) }

//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type. implements countable,
// sliceable, consumeable, testable, compareable, key-, index- and generic pair
// interfaces to be able to stand in as return value for such expressions.
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
func (n NoneVal) TypeFnc() TyFnc                { return None }
func (n NoneVal) TypeElem() d.Typed             { return None }
func (n NoneVal) TypeNat() d.TyNat              { return d.Nil }
func (n NoneVal) Flag() d.BitFlag               { return d.BitFlag(None) }
func (n NoneVal) FlagType() d.Uint8Val          { return Flag_Function.U() }
func (n NoneVal) Type() TyPattern               { return Def(None) }
func (n NoneVal) TypeName() string              { return n.String() }
func (n NoneVal) Slice() []Expression           { return []Expression{} }
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
	return Def(Truth, Def(True, Lex_Pipe, False))
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
	return Def(Trinary, Def(True, Lex_Pipe, False, Lex_Pipe, Undecided))
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
	return Def(Compare, Def(Lesser, Lex_Pipe, Greater, Lex_Pipe, Equal))
}
func (t CompVal) String() string                     { return t.TypeFnc().TypeName() }
func (t CompVal) Test(args ...Expression) bool       { return t(args...) == 0 }
func (t CompVal) Less(args ...Expression) bool       { return t(args...) < 0 }
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
		if len(args) > 0 {
			if test.Test(args...) {
				return expr.Call(args...)
			}
			return NewNone()
		}
		return NewPair(test, expr)
	}
}

// return case constructor only takes a test as argument and returns a case,
// that returns its set of arguments, when matched by the enclosed test
func NewReturnCase(test Testable) CaseVal {
	return NewCase(test, NewFunction(
		func(args ...Expression) Expression {
			if len(args) > 0 {
				if len(args) > 1 {
					return NewVector(args...)
				}
				return args[0]
			}
			return NewNone()
		}, Def(Value)))
}

func (t CaseVal) TypeFnc() TyFnc { return Case }
func (t CaseVal) String() string { return t.TypeFnc().TypeName() }
func (t CaseVal) TypeReturn() TyPattern {
	var pair = t().(Paired)
	return pair.Right().Type()
}
func (t CaseVal) Type() TyPattern {
	var pair = t().(Paired)
	return Def(Case, Def(
		pair.Left().Type().Pattern(),
		pair.Right().Type().Pattern()))
}
func (t CaseVal) Test(args ...Expression) bool {
	var pair = t().(Paired)
	return pair.Left().(Testable).Test(args...)
}
func (t CaseVal) Unbox() (Testable, Expression) {
	var pair = t().(Paired)
	return pair.Left().(Testable), pair.Right()
}
func (t CaseVal) Call(args ...Expression) Expression { return t(args...) }

// all type case returns its set of arguments, if the types of all passed
// arguments are matched by the test comparing it to the type passed to the
// constructor
func NewTypeCase(typ d.Typed) CaseVal {
	var pattern TyPattern
	if Flag_Pattern.Match(typ.FlagType()) {
		pattern = typ.(TyPattern)
	} else {
		pattern = Def(typ)
	}
	return NewCase(NewTest(
		func(args ...Expression) bool {
			for _, arg := range args {
				if !arg.Type().Match(typ) {
					return false
				}
			}
			return true
		}),
		NewFunction(func(args ...Expression) Expression {
			if len(args) > 0 {
				if len(args) > 1 {
					return NewVector(args...)
				}
				return args[0]
			}
			return pattern
		}, pattern))
}

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
	var (
		all     = cases
		current CaseVal
		types   = make([]d.Typed, 0, len(cases))
	)
	for _, c := range cases {
		types = append(types, c.Type().Pattern())
	}

	return func(args ...Expression) (Expression, []CaseVal) {
		if len(args) > 0 {
			if len(cases) > 0 {
				current = cases[0]
				if len(cases) > 1 {
					cases = cases[1:]
				} else {
					cases = []CaseVal{}
				}
			}
			return current(args...), cases
		}
		return NewNone(), all
	}
}
func (t SwitchVal) TypeFnc() TyFnc { return Switch }
func (t SwitchVal) String() string { return t.Type().TypeName() }
func (t SwitchVal) Reload() SwitchVal {
	var _, cases = t()
	return NewSwitch(cases...)
}
func (t SwitchVal) Type() TyPattern {
	var (
		_, cases = t()
		types    = make([]d.Typed, 0, len(cases)+1)
	)
	types = append(types, Switch)
	for _, c := range cases {
		types = append(types, c.Type())
	}
	return Def(types...)
}
func (t SwitchVal) Call(args ...Expression) Expression {
	var result, cases = t(args...)
	for len(cases) > 0 {
		if !result.TypeFnc().Match(None) {
			break
		}
		result, cases = NewSwitch(cases...)(args...)
	}
	t = t.Reload()
	return result
}

/// TYPE SWITCHES
//
// type switch constructors take a list of types, generate a case for every
// type argument and return a switch over all enclosed cases
//
// type all switches return set of arguments passed, if the types of all
// arguments passed to are matched by one of the enclosed type case
func NewTypeSwitch(types ...d.Typed) SwitchVal {
	var cases = make([]CaseVal, 0, len(types))
	for _, t := range types {
		cases = append(cases, NewTypeCase(t))
	}
	return NewSwitch(cases...)
}

/// ELEMENT VALUE
//
// element values yield a subelements of optional, tuple, or enumerable
// expressions with sub-type pattern as second return value
func NewElement(expr Expression, typed d.Typed) ElemVal {
	var pattern = Def(typed, expr.Type())
	return func(args ...Expression) (Expression, TyPattern) {
		if len(args) > 0 {
			return expr.Call(args...), pattern
		}
		return expr, pattern
	}
}
func (e ElemVal) Type() TyPattern                    { var _, p = e(); return p }
func (e ElemVal) Unbox() Expression                  { var r, _ = e(); return r }
func (e ElemVal) TypeReturn() TyPattern              { return e.Unbox().Type() }
func (e ElemVal) TypeFnc() TyFnc                     { return e.Unbox().TypeFnc() }
func (e ElemVal) Call(args ...Expression) Expression { return e.Unbox().Call(args...) }
func (e ElemVal) String() string                     { return e.Unbox().String() }

/// MAYBE VALUE
//
// the constructor takes a case expression, expected to return a result, if the
// case matches the arguments and either returns the resulting none instance,
// or creates a just instance enclosing the resulting value.
func NewMaybe(test CaseVal) MaybeVal {
	var result Expression
	return func(args ...Expression) ElemVal {
		if len(args) > 0 {
			if result = test(args...); !result.TypeFnc().Match(None) {
				return NewElement(result, Just)
			}
			return NewElement(result, None) // ← will be None
		}
		return NewElement(test, Truth)
	}
}
func (t MaybeVal) TypeFnc() TyFnc                     { return Maybe }
func (t MaybeVal) Call(args ...Expression) Expression { return t(args...) }
func (t MaybeVal) Unbox() CaseVal                     { return t().Unbox().(CaseVal) }
func (t MaybeVal) String() string                     { return t.Type().TypeName() }
func (t MaybeVal) Type() TyPattern                    { return Def(Maybe, t.Unbox().TypeReturn()) }

/// OPTIONAL VALUE
//
// constructor takes two case expressions, first one expected to return the
// either result, second one expected to return the or result if the case
// matches. if none of the cases match, a none instance will be returned
func NewOption(either, or CaseVal) OptionVal {
	var result Expression
	return func(args ...Expression) ElemVal {
		if len(args) > 0 {
			if result = either(args...); !result.TypeFnc().Match(None) {
				return NewElement(result, Either)
			}
			if result = or(args...); !result.TypeFnc().Match(None) {
				return NewElement(result, Or)
			}
			return NewElement(result, None) // ← will be None
		}
		return NewElement(NewPair(either, or), Pair)
	}
}
func (t OptionVal) TypeFnc() TyFnc                     { return Option }
func (t OptionVal) Call(args ...Expression) Expression { return t(args...) }
func (t OptionVal) String() string                     { return t.Type().TypeName() }
func (t OptionVal) Type() TyPattern {
	return Def(Option, Def(
		Def(Either, t.EitherCase().TypeReturn()),
		Lex_Pipe,
		Def(Or, t.OrCase().TypeReturn())))
}
func (t OptionVal) Unbox() (CaseVal, CaseVal) {
	var either, or = t().Unbox().(Paired).Both()
	return either.(CaseVal), or.(CaseVal)
}
func (t OptionVal) EitherCase() CaseVal {
	var pair = t().Unbox().(Paired)
	return pair.Left().(CaseVal)
}
func (t OptionVal) OrCase() CaseVal {
	var pair = t().Unbox().(Paired)
	return pair.Right().(CaseVal)
}

/// IF THEN ELSE CONDITION
//
// if statement is a slight variation of an optional.
func NewIf(then, els CaseVal) IfVal {
	var result Expression
	return func(args ...Expression) ElemVal {
		if len(args) > 0 {
			if result = then(args...); !result.TypeFnc().Match(None) {
				return NewElement(result, Then)
			}
			if result = els(args...); !result.TypeFnc().Match(None) {
				return NewElement(result, Else)
			}
			return NewElement(result, None) // ← will be None
		}
		return NewElement(NewPair(then, els), Pair)
	}
}
func (t IfVal) TypeFnc() TyFnc                     { return If }
func (t IfVal) Call(args ...Expression) Expression { return t(args...) }
func (t IfVal) String() string                     { return t.Type().TypeName() }
func (t IfVal) Type() TyPattern {
	return Def(If, Def(
		Def(Then, t.Then().Type()),
		Lex_Pipe,
		Def(Else, t.Else().Type())))
}
func (t IfVal) Unbox() (CaseVal, CaseVal) {
	var then, els = t().Unbox().(Paired).Both()
	return then.(CaseVal), els.(CaseVal)
}
func (t IfVal) Then() CaseVal {
	var pair = t().Unbox().(Paired)
	return pair.Left().(CaseVal)
}
func (t IfVal) Else() CaseVal {
	var pair = t().Unbox().(Paired)
	return pair.Right().(CaseVal)
}
