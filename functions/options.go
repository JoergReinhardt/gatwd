package functions

import (
	"strings"

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

	// MAYBE (JUST | NONE)
	MaybeType func(...Expression) Expression

	// OPTION (EITHER | OR)
	OptionType func(...Expression) Expression

	// TUPLE (ELEM | ‥. | ELEM)
	TupleVal  []Expression
	TupleType func(...Expression) TupleVal
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
		return pattern
	}
}

func (t CaseType) TypeFnc() TyFnc                     { return Case }
func (t CaseType) Type() TyPattern                    { return t().(TyPattern) }
func (t CaseType) TypeReturn() TyPattern              { return t().(TyPattern).Patterns()[2] }
func (t CaseType) TypeIdent() TyPattern               { return t().(TyPattern).Patterns()[1] }
func (t CaseType) TypeArguments() []TyPattern         { return t().(TyPattern).Patterns()[0].Patterns() }
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
					remains = nil
				}
				return current(args...), remains
			}
			return NewNone(), remains
		}
		return pattern, cases
	}
}

func (t SwitchType) Reload() SwitchType { return DecSwitch(t.Cases()...) }
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
	return NewNone()
}

/// MAYBE VALUE
//
// the constructor takes a case expression, expected to return a result, if the
// case matches the arguments and either returns the resulting none instance,
// or creates a just instance enclosing the resulting value.
func DecMaybe(test CaseType) MaybeType {
	var (
		result Expression
		ttype  = test.Type()
	)
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			if result = test(args...); !result.TypeFnc().Match(None) {
				return DecExpression(
					result,
					None,
					Def(Just, result.Type().TypeReturn()),
					result.Type(),
				)
			}
			return DecExpression(result, None, None) // ← will be None
		}
		return DecExpression(
			test,
			ttype.TypeArguments(),
			ttype.TypeReturn(),
			Def(Maybe),
		)
	}
}
func (t MaybeType) TypeFnc() TyFnc                     { return Maybe }
func (t MaybeType) Type() TyPattern                    { return t().Type() }
func (t MaybeType) TypeArguments() TyPattern           { return t().Type().TypeArguments() }
func (t MaybeType) TypeReturn() TyPattern              { return t().Type().TypeReturn() }
func (t MaybeType) String() string                     { return t.Type().TypeName() }
func (t MaybeType) Call(args ...Expression) Expression { return t(args...) }

//// OPTIONAL VALUE
///
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
		orargs     = Def(ots...)
		eitherargs = Def(ets...)
		ortype     = Def(orargs, Def(Or, or.TypeReturn()), or.TypeReturn())
		eithertype = Def(eitherargs, Def(
			Either, either.TypeReturn()), either.TypeReturn())
		pattern = Def(Def(eitherargs, Lex_Pipe, orargs),
			Def(Option, Def(eithertype.TypeIdent(), ortype.TypeIdent())),
			Def(eithertype, Lex_Pipe, ortype))
	)
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			if result = either(args...); !result.TypeFnc().Match(None) {
				return DecExpression(result, eitherargs, either.TypeReturn())
			}
			if result = or(args...); !result.TypeFnc().Match(None) {
				return DecExpression(result, orargs, or.TypeReturn())
			}
			return DecExpression(result, None, None) // ← result will be None
		}
		return pattern
	}
}
func (o OptionType) TypeFnc() TyFnc                     { return Option }
func (o OptionType) Call(args ...Expression) Expression { return o(args...) }
func (o OptionType) String() string                     { return o.Type().TypeName() }
func (o OptionType) Type() TyPattern                    { return o().(TyPattern) }

//// TUPLE TYPE
///
//
func (t TupleVal) TypeFnc() TyFnc { return Tuple }
func (t TupleVal) Type() TyPattern {
	var types = make([]d.Typed, 0, len(t))
	for _, field := range t {
		types = append(types, field.Type())
	}
	return Def(Tuple, Def(types...))
}
func (t TupleVal) String() string {
	var strs = make([]string, 0, len(t))
	for _, field := range t {
		strs = append(strs, field.String())
	}
	return "[" + strings.Join(strs, " ") + "]"
}
func (t TupleVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var elems = make([]Expression, 0, len(args))
		for _, arg := range args {
			if arg.TypeFnc().Match(Data) {
				if data, ok := arg.(Native); ok {
					if data.Eval().Type().Match(d.Int) {
						if i, ok := data.Eval().(d.IntVal); ok {
							if len(args) == 1 {
								return t[i.Int()]
							}
							elems = append(elems, t[i.Int()])
						}
					}
				}
			}
		}
	}
	return t
}

func DefineTuple(examples ...Expression) TupleType {
	var (
		length  = len(examples)
		types   = make([]Expression, length, length)
		members = make([]Expression, length, length)
	)
	for n, e := range examples {
		if Flag_Pattern.Match(e.Type().FlagType()) {
			types[n] = e.(TyPattern)
			members[n] = NewNone()
			continue
		}
		if Flag_Symbol.Match(e.Type().FlagType()) {
			types[n] = e.(TySymbol)
			members[n] = NewNone()
			continue
		}
		types[n] = e.Type()
		members[n] = e
	}
	return func(args ...Expression) TupleVal {
		if len(args) > 0 {
			for n, arg := range args {
				if arg.TypeFnc().Match(Index) {
					if ipair, ok := arg.(IndexPairType); ok {
						if types[ipair.Index()].Type().Match(arg.Type()) &&
							members[ipair.Index()].Type().Match(None) {
							members[ipair.Index()] = ipair.Value()
						}
						continue
					}
				}
				if types[n].Type().Match(arg.Type()) {
					members[n] = arg
				}
			}
		}
		for _, elem := range members {
			if elem.Type().Match(None) {
				return types
			}
		}
		return members
	}
}
func (t TupleType) TypeFnc() TyFnc { return Type | Tuple }
func (t TupleType) Type() TyPattern {
	var (
		elements = t()
		length   = len(elements)
		types    = make([]d.Typed, 0, length)
	)
	for _, elem := range elements {
		types = append(types, elem.Type())
	}
	return Def(Tuple, Def(types...))
}
func (t TupleType) String() string {
	var (
		types = t()
		strs  = make([]string, 0, len(types))
	)
	for _, typ := range types {
		strs = append(strs, typ.Type().TypeName())
	}
	return strings.Join(strs, " ")
}
func (t TupleType) Call(args ...Expression) Expression {
	var tuple = t(args...)
	for _, elem := range tuple {
		if elem.Type().Match(None) {
			return NewNone()
		}
	}
	return tuple
}
func (t TupleType) Declare(args ...Expression) TupleVal {
	var tuple = t.Call(args...)
	if tuple.Type().Match(None) {
		return []Expression{}
	}
	return tuple.(TupleVal)
}
