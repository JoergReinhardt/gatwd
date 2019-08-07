package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// NONE VALUE CONSTRUCTOR
	NoneType func()

	// GENERIC EXPRESSION
	ConstVal func() Expression
	FuncVal  func(...Expression) Expression
	ExprVal  func(...Expression) (Expression, TyPattern)

	// TESTS AND COMPARE
	TestType       func(...Expression) bool
	TrinaryType    func(...Expression) int
	ComparatorType func(...Expression) int

	// CASE & SWITCH
	CaseType   func(...Expression) Expression
	SwitchType func(...Expression) (Expression, []CaseType)

	// MAYBE (JUST | NONE)
	MaybeType func(...Expression) ExprVal

	// OPTION (EITHER | OR)
	OptionType func(...Expression) ExprVal
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
func (n NoneType) TypeNat() d.TyNat                 { return d.Nil }
func (n NoneType) TypeElem() TyPattern              { return Def(None) }
func (n NoneType) Flag() d.BitFlag                  { return d.BitFlag(None) }
func (n NoneType) FlagType() d.Uint8Val             { return Flag_Function.U() }
func (n NoneType) Type() TyPattern                  { return Def(None) }
func (n NoneType) TypeName() string                 { return n.String() }
func (n NoneType) Slice() []Expression              { return []Expression{} }
func (n NoneType) Consume() (Expression, Consumeable) {
	return DeclareNone(), DeclareNone()
}

//// CONSTANT DECLARATION
///
// declares a constant value
func DecConstant(constant func() Expression) ConstVal { return constant }

func (c ConstVal) Type() TyPattern {
	return Def(
		None,
		Def(
			Constant,
			c().Type().TypeIdent(),
		),
		c().Type(),
	)
}
func (c ConstVal) TypeFnc() TyFnc                { return Constant }
func (c ConstVal) String() string                { return c().String() }
func (c ConstVal) Call(...Expression) Expression { return c() }

//// EXPRESSION DECLARATION
///
// declares an expression with defined argument-, return- and an optional identity type
func DecFuntion(
	fn func(...Expression) Expression,
	argtype, retype d.Typed,
	identypes ...d.Typed,
) FuncVal {

	var (
		ident TyPattern
	)

	if len(identypes) == 0 {
		ident = Def(Value)
	} else {
		ident = Def(identypes...)
	}

	var pattern = Def(
		argtype,
		ident,
		retype,
	)

	return func(args ...Expression) Expression {
		if len(args) > 0 {
			return fn(args...)
		}
		return pattern
	}
}

func (g FuncVal) TypeFnc() TyFnc                     { return Value }
func (g FuncVal) Type() TyPattern                    { return g().(TyPattern) }
func (g FuncVal) String() string                     { return g.Type().TypeName() }
func (g FuncVal) Call(args ...Expression) Expression { return g(args...) }

/// PARTIAL APPLYABLE EXPRESSION VALUE
//
// element values yield a subelements of optional, tuple, or enumerable
// expressions with sub-type pattern as second return value
func DecExpression(
	expr Expression,
	argtype, retype d.Typed,
	identypes ...d.Typed,
) ExprVal {
	var (
		arglen         int
		ident, pattern TyPattern
	)
	if len(identypes) == 0 {
		ident = Def(expr.TypeFnc())
	} else {
		ident = Def(identypes...)
	}
	if Flag_Pattern.Match(argtype.FlagType()) {
		arglen = len(argtype.(TyPattern).Patterns())
	} else {
		arglen = 1
	}
	pattern = Def(argtype, ident, retype)
	return func(args ...Expression) (Expression, TyPattern) {
		if len(args) > 0 {
			if pattern.TypeArguments().MatchArgs(args...) {
				var result Expression
				switch {
				case len(args) == arglen:
					result = expr.Call(args...)
					return result, result.Type()

				case len(args) < arglen:
					var argtypes = make(
						[]d.Typed, 0,
						len(pattern.TypeArguments()[len(args):]),
					)
					for _, atype := range pattern.TypeArguments()[len(args):] {
						argtypes = append(argtypes, atype)
					}
					var pattern = Def(Def(argtypes...), ident, retype)
					return DecExpression(FuncVal(
						func(lateargs ...Expression) Expression {
							if len(lateargs) > 0 {
								return expr.Call(append(
									args, lateargs...,
								)...)
							}
							return pattern
						}), Def(argtypes...), ident, retype), pattern

				case len(args) > arglen:
					var vector = NewVector()
					for len(args) > arglen {
						vector = vector.Con(
							expr.Call(args[:arglen]...))
						args = args[arglen:]
					}
					if len(args) > 0 {
						vector = vector.Con(DecExpression(
							expr, argtype, retype, identypes...,
						).Call(args...))
					}
					return vector, vector.Type()
				}
			}
		}
		return expr, pattern
	}
}
func (e ExprVal) Unbox() Expression                  { var expr, _ = e(); return expr }
func (e ExprVal) Type() TyPattern                    { var _, pat = e(); return pat }
func (e ExprVal) TypeArguments() TyPattern           { return e.Type().TypeArguments() }
func (e ExprVal) TypeReturn() TyPattern              { return e.Type().TypeReturn() }
func (e ExprVal) TypeFnc() TyFnc                     { return e.Unbox().TypeFnc() }
func (e ExprVal) String() string                     { return e.Unbox().String() }
func (e ExprVal) Call(args ...Expression) Expression { return e.Unbox().Call(args...) }

/// TEST
//
// create a new test, scrutinizing its arguments and revealing true, or false
func DecTest(test func(...Expression) bool) TestType {
	return func(args ...Expression) bool { return test(args...) }
}
func (t TestType) TypeFnc() TyFnc               { return Truth }
func (t TestType) Type() TyPattern              { return Def(True | False) }
func (t TestType) String() string               { return t.TypeFnc().TypeName() }
func (t TestType) Test(args ...Expression) bool { return t(args...) }
func (t TestType) Compare(args ...Expression) int {
	if t(args...) {
		return 0
	}
	return -1
}
func (t TestType) Call(args ...Expression) Expression {
	return DecData(d.BoolVal(t(args...)))
}

/// TRINARY TEST
//
// create a trinary test, that can yield true, false, or undecided, computed by
// scrutinizing its arguments
func DecTrinary(test func(...Expression) int) TrinaryType {
	return func(args ...Expression) int {
		return test(args...)
	}
}
func (t TrinaryType) TypeFnc() TyFnc { return Trinary }
func (t TrinaryType) Type() TyPattern {
	return Def(True | False | Undecided)
}
func (t TrinaryType) String() string                     { return t.TypeFnc().TypeName() }
func (t TrinaryType) Test(args ...Expression) bool       { return t(args...) == 0 }
func (t TrinaryType) Compare(args ...Expression) int     { return t(args...) }
func (t TrinaryType) Call(args ...Expression) Expression { return DecData(d.IntVal(t(args...))) }

/// COMPARE
//
// create a comparator expression that yields minus one in case the argument is
// lesser, zero in case its equal and plus one in case it is greater than the
// enclosed value to compare against.
func DecComparator(comp func(...Expression) int) ComparatorType {
	return func(args ...Expression) int {
		return comp(args...)
	}
}
func (t ComparatorType) Type() TyPattern {
	return Def(Lesser | Greater | Equal)
}
func (t ComparatorType) TypeFnc() TyFnc                     { return Compare }
func (t ComparatorType) String() string                     { return t.Type().TypeName() }
func (t ComparatorType) Test(args ...Expression) bool       { return t(args...) == 0 }
func (t ComparatorType) Less(args ...Expression) bool       { return t(args...) < 0 }
func (t ComparatorType) Compare(args ...Expression) int     { return t(args...) }
func (t ComparatorType) Call(args ...Expression) Expression { return DecData(d.IntVal(t(args...))) }

/// CASE
//
// case constructor takes a test and an expression, in order for the resulting
// case instance to test its arguments and yield the result of applying those
// arguments to the expression, in case the test yielded true. otherwise the
// case will yield none.
func DecCase(test Testable, expr Expression, argtype, retype d.Typed) CaseType {

	var pattern = Def(argtype, Def(Case, test.Type()), retype)

	return func(args ...Expression) Expression {
		if len(args) > 0 {
			if test.Test(args...) {
				return expr.Call(args...)
			}
			return DeclareNone()
		}
		return pattern
	}
}

func (t CaseType) TypeFnc() TyFnc        { return Case }
func (t CaseType) String() string        { return t.TypeFnc().TypeName() }
func (t CaseType) Type() TyPattern       { return t().(TyPattern) }
func (t CaseType) TypeIdent() TyPattern  { return t().(TyPattern).Patterns()[1] }
func (t CaseType) TypeReturn() TyPattern { return t().(TyPattern).Patterns()[2] }
func (t CaseType) TypeArguments() []TyPattern {
	return t().(TyPattern).Patterns()[0].Patterns()
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
func DecSwitch(cases ...CaseType) SwitchType {
	var types = make([]d.Typed, 0, len(cases))
	for _, c := range cases {
		types = append(types, c.Type())
	}

	var (
		current CaseType
		remains = cases
		pattern = Def(Switch, Def(types...))
	)

	return func(args ...Expression) (Expression, []CaseType) {
		if len(args) > 0 {
			if remains != nil {
				current = remains[0]
				if len(remains) > 1 {
					remains = remains[1:]
				} else {
					remains = nil
				}
				return current(args...), remains
			}
			return DeclareNone(), remains
		}
		return pattern, cases
	}
}
func (t SwitchType) String() string     { return t.Type().TypeName() }
func (t SwitchType) Reload() SwitchType { return DecSwitch(t.Cases()...) }
func (t SwitchType) TypeFnc() TyFnc     { return Switch }
func (t SwitchType) Type() TyPattern {
	var pat, _ = t()
	return pat.(TyPattern)
}
func (t SwitchType) Cases() []CaseType {
	var _, cases = t()
	return cases
}
func (t SwitchType) Call(args ...Expression) Expression {
	var (
		remains = t.Cases()
		result  Expression
	)
	for remains != nil {
		result, remains = t(args...)
		if !result.TypeFnc().Match(None) {
			return result
		}
	}
	return DeclareNone()
}

/// MAYBE VALUE
//
// the constructor takes a case expression, expected to return a result, if the
// case matches the arguments and either returns the resulting none instance,
// or creates a just instance enclosing the resulting value.
func DecMaybe(test CaseType) MaybeType {

	var (
		result  Expression
		ttype   = test.Type()
		pattern = Def(
			ttype.TypeReturn().TypeArguments(),
			Def(
				Maybe,
				ttype.TypeReturn().TypeIdent(),
			),
			ttype.TypeReturn().TypeReturn(),
		)
	)

	return func(args ...Expression) ExprVal {
		if len(args) > 0 {
			if result = test(args...); !result.TypeFnc().Match(None) {
				return DecExpression(
					result,
					pattern.TypeArguments(),
					pattern.TypeReturn(),
					Def(Just, result.Type()),
				)
			}
			return DecExpression(result, None, None) // ← will be None
		}
		return DecExpression(pattern, None, Pattern)
	}
}
func (t MaybeType) Call(args ...Expression) Expression { return t(args...) }
func (t MaybeType) String() string                     { return t.Type().TypeName() }
func (t MaybeType) TypeFnc() TyFnc                     { return Maybe }
func (t MaybeType) Type() TyPattern {
	return t().Unbox().(TyPattern).Type()
}
func (t MaybeType) TypeArguments() TyPattern {
	return t().Unbox().(TyPattern).TypeArguments()
}
func (t MaybeType) TypeReturn() TyPattern {
	return t().Unbox().(TyPattern).TypeReturn()
}

/// OPTIONAL VALUE
//
// constructor takes two case expressions, first one expected to return the
// either result, second one expected to return the or result if the case
// matches. if none of the cases match, a none instance will be returned
func DecOption(either, or CaseType) OptionType {

	var (
		ets = make([]d.Typed, 0, len(either.TypeArguments()))
		ots = make([]d.Typed, 0, len(or.TypeArguments()))
	)
	for _, t := range either.TypeArguments() {
		ets = append(ets, t)
	}
	for _, t := range or.TypeArguments() {
		ots = append(ots, t)
	}

	var (
		result     Expression
		eitherargs = Def(ets...)
		orargs     = Def(ots...)
		eithertype = Def(
			eitherargs,
			Def(
				Either,
				either.TypeReturn(),
			),
			either.TypeReturn(),
		)
		ortype = Def(
			orargs,
			Def(
				Or,
				or.TypeReturn(),
			),
			or.TypeReturn(),
		)
		pattern = Def(
			Def(eitherargs, Lex_Pipe, orargs),
			Def(
				Option,
				Def(
					eithertype.TypeIdent(),
					ortype.TypeIdent(),
				)),
			Def(eithertype, Lex_Pipe, ortype),
		)
	)

	return func(args ...Expression) ExprVal {
		if len(args) > 0 {
			if result = either(args...); !result.TypeFnc().Match(None) {
				return DecExpression(
					result,
					eitherargs,
					either.TypeReturn(),
				)
			}
			if result = or(args...); !result.TypeFnc().Match(None) {
				return DecExpression(
					result,
					orargs,
					or.TypeReturn(),
				)
			}
			return DecExpression(result, None, None) // ← result will be None
		}
		return DecExpression(pattern, None, Pattern)
	}
}
func (t OptionType) TypeFnc() TyFnc                     { return Option }
func (t OptionType) Call(args ...Expression) Expression { return t(args...) }
func (t OptionType) String() string                     { return t.Type().TypeName() }
func (t OptionType) Type() TyPattern                    { return t().Unbox().(TyPattern) }
