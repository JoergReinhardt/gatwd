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
	MaybeVal  func(...Expression) (Expression, TyPattern, MaybeType)

	// ALTERNATETIVES TYPE (EITHER | OR)
	EitherOrType func(...Expression) Expression
	EitherOrVal  func(...Expression) (Expression, TyFnc, EitherOrType)

	// TODO: either poly-/ or option type (homolog)‥. also option value
	// should reference its type constructor and sibling types.

	//// POLYMORPHIC EXPRESSION (INSTANCE OF CASE-SWITCH)
	PolyType func(...Expression) (Expression, []ExprType, int)
	PolyVal  func(...Expression) (Expression, PolyType)
)

/// TRUTH TEST
//
// create a new test, scrutinizing its arguments and revealing true, or false
func NewTest(test func(...Expression) bool) TestType {
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
func NewTrinary(test func(...Expression) int) TrinaryType {
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
func NewComparator(comp func(...Expression) int) ComparatorType {
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
func NewCase(test Testable, expr Expression, argtype, retype d.Typed) CaseType {
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
func NewSwitch(cases ...CaseType) SwitchType {
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
				var result = current(args...)
				if result.Type().Match(None) {
					return result, remains
				}
				return result, cases
			}
			return NewNone(), cases
		}
		return pattern, cases
	}
}

func (t SwitchType) Reload() SwitchType { return NewSwitch(t.Cases()...) }
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
func NewMaybe(cas CaseType) MaybeType {
	var argtypes = make([]d.Typed, 0, len(cas.TypeArguments()))
	for _, arg := range cas.TypeArguments() {
		argtypes = append(argtypes, arg)
	}
	var (
		pattern = Def(Def(argtypes...), Def(Just|None), Def(cas.TypeReturn()))
	)
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			// pass arguments to case, check if result is none‥.
			if result := cas.Call(args...); !result.Type().Match(None) {
				// ‥.otherwise return a maybe just
				return MaybeVal(func(args ...Expression) (Expression, TyPattern, MaybeType) {
					if len(args) > 0 {
						// return result from passing
						// args to result of initial
						// call
						return result.Call(args...), Def(
							Def(argtypes...),
							Just,
							result.Type().TypeReturn(),
						), NewMaybe(cas)
					}
					return result, Def(
						Def(argtypes...),
						Just,
						result.Type().TypeReturn(),
					), NewMaybe(cas)
				})
			}
			// no matching arguments where passed, return none
			return MaybeVal(func(...Expression) (Expression, TyPattern, MaybeType) {
				return NewNone(), Def(None), NewMaybe(cas)
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

// maybe values methods
func (t MaybeVal) TypeFnc() TyFnc                     { return Maybe }
func (t MaybeVal) Call(args ...Expression) Expression { var result, _, _ = t(args...); return result }
func (t MaybeVal) String() string                     { var result, _, _ = t(); return result.String() }
func (t MaybeVal) Type() TyPattern                    { var _, pat, _ = t(); return pat }

//// OPTIONAL VALUE
///
// constructor takes two case expressions, first one expected to return the
// either result, second one expected to return the or result if the case
// matches. if none of the cases match, a none instance will be returned
func NewEitherOr(either, or CaseType) EitherOrType {
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
				Def(Or, orArgs),
			),
			Def(Either|Or),
			Def(
				Def(Either, either.TypeReturn()),
				Def(Or, or.TypeReturn()),
			))
	)

	return EitherOrType(func(args ...Expression) Expression {
		if len(args) > 0 {
			var result Expression
			if result = either.Call(args...); !result.Type().Match(None) {
				return EitherOrVal(func(args ...Expression) (Expression, TyFnc, EitherOrType) {
					if len(args) > 0 {
						return result.Call(args...), Either, NewEitherOr(either, or)
					}
					return result, Either, NewEitherOr(either, or)
				})
			}
			if result = or.Call(args...); !result.Type().Match(None) {
				return EitherOrVal(func(args ...Expression) (Expression, TyFnc, EitherOrType) {
					if len(args) > 0 {
						return result.Call(args...), Or, NewEitherOr(either, or)
					}
					return result, Or, NewEitherOr(either, or)
				})
			}
			return EitherOrVal(func(...Expression) (Expression, TyFnc, EitherOrType) {
				return NewNone(), None, NewEitherOr(either, or)
			})
		}
		return pattern
	})
}
func (o EitherOrType) TypeFnc() TyFnc                     { return Constructor }
func (o EitherOrType) Type() TyPattern                    { return o().Type() }
func (o EitherOrType) String() string                     { return o().String() }
func (o EitherOrType) Call(args ...Expression) Expression { return o(args...) }

//// ALTERNATIVE VALUE
///
func (o EitherOrVal) TypeFnc() TyFnc {
	var _, ft, _ = o()
	return ft
}
func (o EitherOrVal) AlternativeType() EitherOrType {
	var _, _, eo = o()
	return eo
}
func (o EitherOrVal) Value() Expression {
	var r, _, _ = o()
	return r
}
func (o EitherOrVal) ValType() TyPattern {
	return o.Value().Type()
}
func (o EitherOrVal) Type() TyPattern {
	return Def(o.ValType(), o.TypeFnc())
}
func (o EitherOrVal) String() string {
	return o.Value().String()
}
func (o EitherOrVal) Call(args ...Expression) Expression {
	return o.Value().Call(args...)
}

//// POLYMORPHIC TYPE
///
//
// declare new polymorphic named type from cases
func NewPolyType(name string, cases ...ExprType) PolyType {
	var (
		patterns = make([]d.Typed, 0, len(cases))
		pat      TyPattern
	)
	for _, c := range cases {
		patterns = append(patterns, c.Type())
	}
	pat = Def(DefSym(name), Def(patterns...))
	return createPolyType(pat, 0, cases...)
}

// type constructor to construct type instances holding execution state during
// recursion
func createPolyType(pat TyPattern, idx int, cases ...ExprType) PolyType {
	var length = len(cases)
	return func(args ...Expression) (Expression, []ExprType, int) {
		if len(args) > 0 { // arguments where passed
			if idx < length { // not all cases scrutinized yet
				// scrutinize arguments, retrieve expr, or none
				var expr = cases[idx](args...)
				// if none‥.
				if expr.Type().Match(None) {
					// either increment count, or reset to
					// zero, if all cases have been
					// scrutinized
					if idx == length-1 {
						idx = 0
					} else {
						idx += 1
					}
					// return poly type instance pointing
					// to next case for testing it's
					// arguments
					return createPolyType(pat, idx, cases...), cases, idx
				}
				// if arguments matched case, return result as
				// instance of polymorphic sub type instance
				return PolyVal(func(args ...Expression) (Expression, PolyType) {
					if len(args) > 0 {
						return expr.Call(args...), createPolyType(pat, idx, cases...)
					}
					return expr.Call(), createPolyType(pat, idx, cases...)
				}), cases, idx
			}
		}
		// return reset poly type instance, slice of cases and minus
		// one to indicate end of cases list
		return createPolyType(pat, 0, cases...), cases, 0
	}
}

// loops over all cases with a set of passed arguments and returns either
// result, or none
func (p PolyType) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var r, _, i = p(args...)
		for i > 0 {
			if !r.Type().Match(None) {
				return r
			}
			r, _, i = p(args...)
		}
	}
	return NewNone()
}

// function type is polymorph
func (p PolyType) TypeFnc() TyFnc { return Polymorph }

// type is the sum of all argument set and return value types, identity is
// defined by passed name
func (p PolyType) Type() TyPattern {
	var (
		t, _, _  = p()
		pat      = t.(TyPattern)
		identype = pat.Pattern()[0]
		argtypes = make([]d.Typed, 0, len(pat.Pattern()))
		retypes  = make([]d.Typed, 0, len(pat.Pattern()))
	)
	for _, pat := range pat.Pattern()[1:] {
		argtypes = append(argtypes, Def(pat.TypeArguments()...))
		retypes = append(retypes, pat.TypeReturn())
	}
	return Def(Def(argtypes...), identype, Def(retypes...))
}

// returns set of all sub-type defining cases
func (p PolyType) Cases() []ExprType {
	var _, c, _ = p()
	return c
}

// returns set index of last evaluated case
func (p PolyType) Index() int {
	var _, _, i = p()
	return i
}
func (p PolyType) String() string {
	var (
		cases              = p.Cases()
		length             = len(cases)
		arguments, returns = make([]string, 0, length), make([]string, 0, length)
	)
	for _, c := range cases {
		var (
			args   = c.Type().TypeArguments()
			argstr string
		)
		if len(args) > 0 {
			var argstrs = make([]string, 0, len(args))
			for _, arg := range args {
				argstrs = append(argstrs, arg.TypeName())
			}
			argstr = strings.Join(argstrs, " → ")
		} else {
			argstr = args[0].TypeName()
		}
		arguments = append(arguments, argstr)
		returns = append(returns, c.Type().TypeReturn().TypeName())
	}
	return "(" + strings.Join(arguments, " | ") + ")" +
		" → " + p.Type().Pattern()[0].TypeName() +
		" → " + "(" + strings.Join(returns, " | ") + ")"
}

//// POLYMORPHIC SUBTYPE INSTANCE VALUE
///
//
func (p PolyVal) Expr() Expression {
	var e, _ = p()
	return e
}
func (p PolyVal) PolyType() PolyType {
	var _, t = p()
	return t
}
func (p PolyVal) String() string { return p.Expr().String() }
func (p PolyVal) TypeFnc() TyFnc { return Polymorph }
func (p PolyVal) Type() TyPattern {
	return Def(Def(
		Polymorph,
		DefValNative(d.IntVal(p.PolyType().Index())),
	),
		p.Expr().Type(),
	)
}
func (p PolyVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return p.Expr().Call(args...)
	}
	return p.Expr()
}