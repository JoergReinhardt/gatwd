package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// NONE VALUE CONSTRUCTOR
	NoneType func()

	// TESTS AND COMPARE
	TestType       func(...Expression) bool
	TrinaryType    func(...Expression) int
	ComparatorType func(...Expression) int

	// CASE & SWITCH
	CaseType   func(...Expression) Expression
	SwitchType func(...Expression) (Expression, []CaseType)

	// OPTION ELEMENT
	ElemType func(...Expression) (Expression, TyPattern)

	// MAYBE (JUST | NONE)
	MaybeType func(...Expression) ElemType

	// OPTION (EITHER | OR)
	OptionType func(...Expression) ElemType

	// IF (THEN | ELSE)
	BranchType func(...Expression) ElemType
)

//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type. implements countable,
// sliceable, consumeable, testable, compareable, key-, index- and generic pair
// interfaces to be able to stand in as return value for such expressions.
func DeclareNone() NoneType { return func() {} }

func (n NoneType) Head() Expression                 { return n }
func (n NoneType) Tail() Consumeable                { return n }
func (n NoneType) Append(...Expression) Consumeable { return n }
func (n NoneType) Len() int                         { return 0 }
func (n NoneType) String() string                   { return "⊥" }
func (n NoneType) Call(...Expression) Expression    { return nil }
func (n NoneType) Key() Expression                  { return nil }
func (n NoneType) Index() Expression                { return nil }
func (n NoneType) Left() Expression                 { return nil }
func (n NoneType) Right() Expression                { return nil }
func (n NoneType) Both() Expression                 { return nil }
func (n NoneType) Value() Expression                { return nil }
func (n NoneType) Empty() d.BoolVal                 { return true }
func (n NoneType) Test(...Expression) bool          { return false }
func (n NoneType) Compare(...Expression) int        { return -1 }
func (n NoneType) TypeFnc() TyFnc                   { return None }
func (n NoneType) TypeElem() d.Typed                { return None }
func (n NoneType) TypeNat() d.TyNat                 { return d.Nil }
func (n NoneType) Flag() d.BitFlag                  { return d.BitFlag(None) }
func (n NoneType) FlagType() d.Uint8Val             { return Flag_Function.U() }
func (n NoneType) Type() TyPattern                  { return Def(None) }
func (n NoneType) TypeName() string                 { return n.String() }
func (n NoneType) Slice() []Expression              { return []Expression{} }
func (n NoneType) Consume() (Expression, Consumeable) {
	return DeclareNone(), DeclareNone()
}

/// TEST
//
// create a new test, scrutinizing its arguments and revealing true, or false
func DeclareTest(test func(...Expression) bool) TestType {
	return func(args ...Expression) bool {
		return test(args...)
	}
}
func (t TestType) TypeFnc() TyFnc               { return Truth }
func (t TestType) Type() TyPattern              { return Def(True, Lex_Pipe, False) }
func (t TestType) String() string               { return t.TypeFnc().TypeName() }
func (t TestType) Test(args ...Expression) bool { return t(args...) }
func (t TestType) Compare(args ...Expression) int {
	if t(args...) {
		return 0
	}
	return -1
}
func (t TestType) Call(args ...Expression) Expression {
	return DeclareData(d.BoolVal(t(args...)))
}

/// TRINARY TEST
//
// create a trinary test, that can yield true, false, or undecided, computed by
// scrutinizing its arguments
func DeclareTrinary(test func(...Expression) int) TrinaryType {
	return func(args ...Expression) int {
		return test(args...)
	}
}
func (t TrinaryType) TypeFnc() TyFnc { return Trinary }
func (t TrinaryType) Type() TyPattern {
	return Def(Trinary, Lex_Pipe, False, Lex_Pipe, Undecided)
}
func (t TrinaryType) String() string                     { return t.TypeFnc().TypeName() }
func (t TrinaryType) Test(args ...Expression) bool       { return t(args...) == 0 }
func (t TrinaryType) Compare(args ...Expression) int     { return t(args...) }
func (t TrinaryType) Call(args ...Expression) Expression { return DeclareData(d.IntVal(t(args...))) }

/// COMPARE
//
// create a comparator expression that yields minus one in case the argument is
// lesser, zero in case its equal and plus one in case it is greater than the
// enclosed value to compare against.
func DeclareComparator(comp func(...Expression) int) ComparatorType {
	return func(args ...Expression) int {
		return comp(args...)
	}
}
func (t ComparatorType) TypeFnc() TyFnc { return Compare }
func (t ComparatorType) Type() TyPattern {
	return Def(Lesser, Lex_Pipe, Greater, Lex_Pipe, Equal)
}
func (t ComparatorType) String() string                     { return t.TypeFnc().TypeName() }
func (t ComparatorType) Test(args ...Expression) bool       { return t(args...) == 0 }
func (t ComparatorType) Less(args ...Expression) bool       { return t(args...) < 0 }
func (t ComparatorType) Compare(args ...Expression) int     { return t(args...) }
func (t ComparatorType) Call(args ...Expression) Expression { return DeclareData(d.IntVal(t(args...))) }

/// CASE
//
// case constructor takes a test and an expression, in order for the resulting
// case instance to test its arguments and yield the result of applying those
// arguments to the expression, in case the test yielded true. otherwise the
// case will yield none.
func DeclareCase(test Testable, expr Expression) CaseType {
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			if test.Test(args...) {
				return expr.Call(args...)
			}
			return DeclareNone()
		}
		return NewPair(test, expr)
	}
}

// return case constructor only takes a test as argument and returns a case,
// that returns its set of arguments, when matched by the enclosed test
func DeclareYield(test Testable) CaseType {
	return DeclareCase(test, DeclareFunction(
		func(args ...Expression) Expression {
			if len(args) > 0 {
				if len(args) > 1 {
					return NewVector(args...)
				}
				return args[0]
			}
			return DeclareNone()
		}, Def(Value)))
}

// declare a case that yields the passed argument, when it's type matches the
// declared type.
func DeclareTypeCase(typ d.Typed) CaseType {
	return DeclareYield(DeclareTest(
		func(args ...Expression) bool {
			// if passed type is a type-pattern, match-args
			if Flag_Pattern.Match(typ.FlagType()) {
				if pat, ok := typ.(TyPattern); ok {
					return pat.MatchArgs(args...)
				}
			}
			// if the passed type is a flat type, all arguments
			// must match that type.
			for _, arg := range args {
				if !arg.Type().Match(typ) {
					return false
				}
			}
			return true
		}))
}

func (t CaseType) TypeFnc() TyFnc { return Case }
func (t CaseType) String() string { return t.TypeFnc().TypeName() }
func (t CaseType) TypeReturn() TyPattern {
	var pair = t().(Paired)
	return pair.Right().Type()
}
func (t CaseType) Type() TyPattern {
	var pair = t().(Paired)
	return Def(Case, Def(
		pair.Left().Type().Pattern(),
		pair.Right().Type().Pattern()))
}
func (t CaseType) Test(args ...Expression) bool {
	var pair = t().(Paired)
	return pair.Left().(Testable).Test(args...)
}
func (t CaseType) Unbox() (Testable, Expression) {
	var pair = t().(Paired)
	return pair.Left().(Testable), pair.Right()
}
func (t CaseType) Call(args ...Expression) Expression { return t(args...) }

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
func DeclareSwitch(cases ...CaseType) SwitchType {
	var (
		all     = cases
		current CaseType
		types   = make([]d.Typed, 0, len(cases))
	)
	for _, c := range cases {
		types = append(types, c.Type().Pattern())
	}

	return func(args ...Expression) (Expression, []CaseType) {
		if len(args) > 0 {
			if len(cases) > 0 {
				current = cases[0]
				if len(cases) > 1 {
					cases = cases[1:]
				} else {
					cases = []CaseType{}
				}
			}
			return current(args...), cases
		}
		return DeclareNone(), all
	}
}
func (t SwitchType) TypeFnc() TyFnc { return Switch }
func (t SwitchType) String() string { return t.Type().TypeName() }
func (t SwitchType) Cases() []CaseType {
	var _, cases = t()
	return cases
}
func (t SwitchType) Reload() SwitchType {
	var _, cases = t()
	return DeclareSwitch(cases...)
}
func (t SwitchType) Type() TyPattern {
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
func (t SwitchType) Call(args ...Expression) Expression {
	var result, cases = t(args...)
	for len(cases) > 0 {
		if !result.TypeFnc().Match(None) {
			break
		}
		result, cases = DeclareSwitch(cases...)(args...)
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
func NewTypeSwitch(patterns ...TyPattern) SwitchType {
	var cases = make([]CaseType, 0, len(patterns))
	for _, pattern := range patterns {
		cases = append(
			cases, DeclareYield(DeclareTest(func(args ...Expression) bool {
				return pattern.MatchArgs(args...)
			})))
	}
	return DeclareSwitch(cases...)
}

/// ELEMENT VALUE
//
// element values yield a subelements of optional, tuple, or enumerable
// expressions with sub-type pattern as second return value
func DeclareElementType(expr Expression, typed d.Typed) ElemType {
	var pattern = Def(typed, expr.Type())
	return func(args ...Expression) (Expression, TyPattern) {
		if len(args) > 0 {
			return expr.Call(args...), pattern
		}
		return expr, pattern
	}
}
func (e ElemType) Type() TyPattern                    { var _, p = e(); return p }
func (e ElemType) Unbox() Expression                  { var r, _ = e(); return r }
func (e ElemType) TypeReturn() TyPattern              { return e.Unbox().Type() }
func (e ElemType) TypeFnc() TyFnc                     { return e.Unbox().TypeFnc() }
func (e ElemType) Call(args ...Expression) Expression { return e.Unbox().Call(args...) }
func (e ElemType) String() string                     { return e.Unbox().String() }

/// MAYBE VALUE
//
// the constructor takes a case expression, expected to return a result, if the
// case matches the arguments and either returns the resulting none instance,
// or creates a just instance enclosing the resulting value.
func DeclareMaybe(test CaseType) MaybeType {
	var result Expression
	return func(args ...Expression) ElemType {
		if len(args) > 0 {
			if result = test(args...); !result.TypeFnc().Match(None) {
				return DeclareElementType(result, Just)
			}
			return DeclareElementType(result, None) // ← will be None
		}
		return DeclareElementType(test, Truth)
	}
}
func (t MaybeType) TypeFnc() TyFnc                     { return Maybe }
func (t MaybeType) Call(args ...Expression) Expression { return t(args...) }
func (t MaybeType) Unbox() CaseType                    { return t().Unbox().(CaseType) }
func (t MaybeType) String() string                     { return t.Type().TypeName() }
func (t MaybeType) Type() TyPattern                    { return t.Unbox().TypeReturn() }

/// OPTIONAL VALUE
//
// constructor takes two case expressions, first one expected to return the
// either result, second one expected to return the or result if the case
// matches. if none of the cases match, a none instance will be returned
func DeclareOption(either, or CaseType) OptionType {
	var result Expression
	return func(args ...Expression) ElemType {
		if len(args) > 0 {
			if result = either(args...); !result.TypeFnc().Match(None) {
				return DeclareElementType(result, Either)
			}
			if result = or(args...); !result.TypeFnc().Match(None) {
				return DeclareElementType(result, Or)
			}
			return DeclareElementType(result, None) // ← will be None
		}
		return DeclareElementType(NewPair(either, or), Pair)
	}
}
func (t OptionType) TypeFnc() TyFnc                     { return Option }
func (t OptionType) Call(args ...Expression) Expression { return t(args...) }
func (t OptionType) String() string                     { return t.Type().TypeName() }
func (t OptionType) Type() TyPattern {
	return Def(Def(Either, t.EitherCase().TypeReturn()),
		Lex_Pipe, Def(Or, t.OrCase().TypeReturn()))
}
func (t OptionType) Unbox() (CaseType, CaseType) {
	var either, or = t().Unbox().(Paired).Both()
	return either.(CaseType), or.(CaseType)
}
func (t OptionType) EitherCase() CaseType {
	var pair = t().Unbox().(Paired)
	return pair.Left().(CaseType)
}
func (t OptionType) OrCase() CaseType {
	var pair = t().Unbox().(Paired)
	return pair.Right().(CaseType)
}

/// IF THEN ELSE CONDITION
//
// if statement is a slight variation of an optional.
func DeclareBranch(then, els CaseType) BranchType {
	var result Expression
	return func(args ...Expression) ElemType {
		if len(args) > 0 {
			if result = then(args...); !result.TypeFnc().Match(None) {
				return DeclareElementType(result, Then)
			}
			if result = els(args...); !result.TypeFnc().Match(None) {
				return DeclareElementType(result, Else)
			}
			return DeclareElementType(result, None) // ← will be None
		}
		return DeclareElementType(NewPair(then, els), Pair)
	}
}
func (t BranchType) TypeFnc() TyFnc                     { return If }
func (t BranchType) Call(args ...Expression) Expression { return t(args...) }
func (t BranchType) String() string                     { return t.Type().TypeName() }
func (t BranchType) Type() TyPattern {
	return Def(Def(Then, t.Then().Type()),
		Lex_Pipe, Def(Else, t.Else().Type()))
}
func (t BranchType) Unbox() (CaseType, CaseType) {
	var then, els = t().Unbox().(Paired).Both()
	return then.(CaseType), els.(CaseType)
}
func (t BranchType) Then() CaseType {
	var pair = t().Unbox().(Paired)
	return pair.Left().(CaseType)
}
func (t BranchType) Else() CaseType {
	var pair = t().Unbox().(Paired)
	return pair.Right().(CaseType)
}
