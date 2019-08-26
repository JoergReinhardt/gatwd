package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// TESTS AND COMPARE
	TestType       func(...Expression) bool
	TrinaryType    func(...Expression) int
	ComparatorType func(...Expression) int

	// CASE & SWITCH
	CaseType   func(...Expression) Expression
	SwitchType func(...Expression) (Expression, []CaseType)

	//// POLYMORPHIC EXPRESSION
	PolyVal SwitchType

	// MAYBE (JUST | NONE)
	MaybeType func(...Expression) Expression
	MaybeVal  func(...Expression) (Expression, TyPattern)

	// OPTION (EITHER | OR)
	VariantType func(...Expression) Expression
	VariantVal  func(...Expression) (Expression, TyPattern)

	//// ENUMERABLE
	EnumType TupleVal
	EnumVal  func(...Expression) TupleVal
)

/// TRUTH TEST
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
	return func(args ...Expression) int { return test(args...) }
}
func (t TrinaryType) TypeFnc() TyFnc                     { return Trinary }
func (t TrinaryType) Type() TyPattern                    { return Def(True | False | Undecided) }
func (t TrinaryType) Call(args ...Expression) Expression { return DecData(d.IntVal(t(args...))) }
func (t TrinaryType) String() string                     { return t.TypeFnc().TypeName() }
func (t TrinaryType) Test(args ...Expression) bool       { return t(args...) == 0 }
func (t TrinaryType) Compare(args ...Expression) int     { return t(args...) }

/// COMPARATOR
//
// create a comparator expression that yields minus one in case the argument is
// lesser, zero in case its equal and plus one in case it is greater than the
// enclosed value to compare against.
func DecComparator(comp func(...Expression) int) ComparatorType {
	return func(args ...Expression) int { return comp(args...) }
}
func (t ComparatorType) TypeFnc() TyFnc                     { return Compare }
func (t ComparatorType) Type() TyPattern                    { return Def(Lesser | Greater | Equal) }
func (t ComparatorType) Call(args ...Expression) Expression { return DecData(d.IntVal(t(args...))) }
func (t ComparatorType) String() string                     { return t.Type().TypeName() }
func (t ComparatorType) Test(args ...Expression) bool       { return t(args...) == 0 }
func (t ComparatorType) Less(args ...Expression) bool       { return t(args...) < 0 }
func (t ComparatorType) Compare(args ...Expression) int     { return t(args...) }

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
			return NewNone()
		}
		return NewPair(pattern, test)
	}
}

func (t CaseType) TypeFnc() TyFnc                     { return Case }
func (t CaseType) Type() TyPattern                    { return t().(Paired).Left().(TyPattern) }
func (t CaseType) Test() TestType                     { return t().(Paired).Right().(TestType) }
func (t CaseType) TypeReturn() TyPattern              { return t.Type().Pattern()[2] }
func (t CaseType) TypeIdent() TyPattern               { return t.Type().Pattern()[1] }
func (t CaseType) TypeArguments() []TyPattern         { return t.Type().Pattern()[0].Pattern() }
func (t CaseType) String() string                     { return t.TypeFnc().TypeName() }
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
					remains = remains[:0]
				}
				return current(args...), remains
			}
			return NewNone(), remains
		}
		return pattern, cases
	}
}

func (t SwitchType) Reload() SwitchType { return DecSwitch(t.Cases()...) }
func (t SwitchType) String() string     { return t.Type().TypeName() }
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
	for len(remains) > 0 {
		result, remains = t(args...)
		if !result.TypeFnc().Match(None) {
			return result
		}
	}
	return NewNone()
}

/// MAYBE VALUE
//
// the constructor takes a case expression, expected to return a result, if the
// case matches the arguments and either returns the resulting none instance,
// or creates a just instance enclosing the resulting value.
func DecMaybe(cas CaseType) MaybeType {
	var argtypes = make([]d.Typed, 0, len(cas.TypeArguments()))
	for _, arg := range cas.TypeArguments() {
		argtypes = append(argtypes, arg)
	}
	var (
		pattern = Def(Def(argtypes...), Def(Just|None), Def(cas.TypeReturn()))
	)
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			if result := cas.Call(args...); !result.Type().Match(None) {
				return MaybeVal(func(args ...Expression) (Expression, TyPattern) {
					if len(args) > 0 {
						return result.Call(args...),
							Def(
								Def(argtypes...),
								Just,
								result.Type().TypeReturn(),
							)
					}
					return result, Def(
						Def(argtypes...),
						Just,
						result.Type().TypeReturn(),
					)
				})
			}
			return MaybeVal(func(...Expression) (Expression, TyPattern) {
				return NewNone(), Def(None)
			})
		}
		return pattern
	}
}

func (t MaybeType) TypeFnc() TyFnc                     { return Constructor }
func (t MaybeType) Type() TyPattern                    { return t().(TyPattern) }
func (t MaybeType) TypeArguments() TyPattern           { return t().Type().TypeArguments() }
func (t MaybeType) TypeReturn() TyPattern              { return t().Type().TypeReturn() }
func (t MaybeType) String() string                     { return t().String() }
func (t MaybeType) Call(args ...Expression) Expression { return t.Call(args...) }

func (t MaybeVal) TypeFnc() TyFnc                     { return Maybe }
func (t MaybeVal) Call(args ...Expression) Expression { var result, _ = t(args...); return result }
func (t MaybeVal) String() string                     { var result, _ = t(); return result.String() }
func (t MaybeVal) Type() TyPattern                    { var _, pat = t(); return pat }

//// OPTIONAL VALUE
///
// constructor takes two case expressions, first one expected to return the
// either result, second one expected to return the or result if the case
// matches. if none of the cases match, a none instance will be returned
func DecVariant(either, or CaseType) VariantType {
	var (
		typesEither = make([]d.Typed, 0, len(either.TypeArguments()))
		typesOr     = make([]d.Typed, 0, len(or.TypeArguments()))
	)
	for _, arg := range either.TypeArguments() {
		typesEither = append(typesEither, arg)
	}
	for _, arg := range or.TypeArguments() {
		typesOr = append(typesOr, arg)
	}
	var (
		eitherArgs, orArgs = Def(typesEither...), Def(typesOr...)

		pattern = Def(
			Def(
				Def(Either, eitherArgs),
				Lex_Pipe,
				Def(Or, orArgs),
			),
			Def(Either|Or),
			Def(
				Def(Either, either.TypeReturn()),
				Lex_Pipe,
				Def(Or, or.TypeReturn()),
			))
	)

	return VariantType(func(args ...Expression) Expression {
		if len(args) > 0 {
			var result Expression
			if result = either.Call(args...); !result.Type().Match(None) {
				return VariantVal(func(args ...Expression) (Expression, TyPattern) {
					if len(args) > 0 {
						return result.Call(args...),
							Def(
								eitherArgs,
								Def(Either),
								result.Type().TypeReturn(),
							)
					}
					return result, Def(
						eitherArgs,
						Def(Either),
						result.Type().TypeReturn(),
					)
				})
			}
			if result = or.Call(args...); !result.Type().Match(None) {
				return VariantVal(func(args ...Expression) (Expression, TyPattern) {
					if len(args) > 0 {
						return result.Call(args...),
							Def(
								orArgs,
								Def(Or),
								result.Type().TypeReturn(),
							)
					}
					return result, Def(
						orArgs,
						Def(Or),
						result.Type().TypeReturn(),
					)
				})
			}
			return VariantVal(func(...Expression) (Expression, TyPattern) {
				return NewNone(), Def(None)
			})
		}
		return pattern
	})
}
func (o VariantType) TypeFnc() TyFnc                     { return Constructor }
func (o VariantType) Type() TyPattern                    { return o().Type() }
func (o VariantType) String() string                     { return o().String() }
func (o VariantType) Call(args ...Expression) Expression { return o(args...) }

func (o VariantVal) TypeFnc() TyFnc { return Variant }
func (o VariantVal) Call(args ...Expression) Expression {
	var result, _ = o(args...)
	return result
}
func (o VariantVal) String() string {
	var result, _ = o()
	return result.String()
}
func (o VariantVal) Type() TyPattern { var _, pat = o(); return pat }
