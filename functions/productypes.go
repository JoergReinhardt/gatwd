/*

PRODUCT TYPES
-------------

productypes are parametric types, that generate a subtype for every possible
combination of its arguments. product types are defined as the union of unions
of its elements types. different elements of a product type product may be of
various type.

examples for productypes
  - records & tuples since their elements can vary in type
  - maybe → just|none
  - either|or
  - the function signature type
    - a signatures type is always function
    - varying argument & return types define different sub types of the
      function type
    - a function types argument & return types are analog to record, or tuples
      element types
  - polymorphic function type
    - polymorph types are a set of possibly different signature types
    - every distinct set of signature types is a unique sub type of the
      polymorph type

*/
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (

	// TESTS AND COMPARE
	TestFunc    func(Expression) bool
	TrinaryFunc func(Expression) int
	CompareFunc func(Expression) int

	// CASE & SWITCH
	// needs to be variadic in orderto enable type overload
	CaseDef   func(...Expression) Expression
	SwitchDef func(...Expression) (Expression, []CaseDef)

	// MAYBE (JUST | NONE)
	MaybeType func(...Expression) Expression
	JustVal   func(...Expression) Expression

	// ALTERNATETIVES TYPE (EITHER | OR)
	AlternateDef func(...Expression) Expression
	EitherVal    func(...Expression) Expression
	OrVal        func(...Expression) Expression
)

/// TRUTH TEST
//
// create a new test, scrutinizing its arguments and revealing true, or false
func NewTest(test func(Expression) bool) TestFunc {
	return func(arg Expression) bool { return test(arg) }
}
func (t TestFunc) TypeFnc() TyFnc {
	return Truth
}
func (t TestFunc) Type() TyComp {
	return Def(True | False)
}
func (t TestFunc) String() string {
	return t.TypeFnc().TypeName()
}
func (t TestFunc) Test(arg Expression) bool {
	return t(arg)
}
func (t TestFunc) Compare(arg Expression) int {
	if t(arg) {
		return 0
	}
	return -1
}
func (t TestFunc) Call(args ...Expression) Expression {
	if len(args) == 1 {
		return Box(d.BoolVal(t(args[0])))
	}
	if len(args) > 1 {
		return Box(d.BoolVal(t(NewVector(args...))))
	}
	return Box(d.BoolVal(false))
}

/// TRINARY TEST
//
// create a trinary test, that can yield true, false, or undecided, computed by
// scrutinizing its arguments
func NewTrinary(test func(Expression) int) TrinaryFunc {
	return func(arg Expression) int { return test(arg) }
}
func (t TrinaryFunc) TypeFnc() TyFnc {
	return Trinary
}
func (t TrinaryFunc) Type() TyComp {
	return Def(True | False | Undecided)
}
func (t TrinaryFunc) Call(arg Expression) Expression {
	return Box(d.IntVal(t(arg)))
}
func (t TrinaryFunc) String() string {
	return t.TypeFnc().TypeName()
}
func (t TrinaryFunc) Test(arg Expression) bool {
	return t(arg) == 0
}
func (t TrinaryFunc) Compare(arg Expression) int {
	return t(arg)
}

/// COMPARATOR
//
// create a comparator expression that yields minus one in case the argument is
// lesser, zero in case its equal and plus one in case it is greater than the
// enclosed value to compare against.
func NewComparator(comp func(Expression) int) CompareFunc {
	return func(arg Expression) int { return comp(arg) }
}
func (t CompareFunc) TypeFnc() TyFnc {
	return Compare
}
func (t CompareFunc) Type() TyComp {
	return Def(Lesser | Greater | Equal)
}
func (t CompareFunc) Call(arg Expression) Expression {
	return Box(d.IntVal(t(arg)))
}
func (t CompareFunc) String() string {
	return t.Type().TypeName()
}
func (t CompareFunc) Test(arg Expression) bool {
	return t(arg) == 0
}
func (t CompareFunc) Less(arg Expression) bool {
	return t(arg) < 0
}
func (t CompareFunc) Compare(arg Expression) int {
	return t(arg)
}

/// CASE
//
// case constructor takes a test and an expression, in order for the resulting
// case instance to test its arguments and yield the result of applying those
// arguments to the expression, in case the test yielded true.  otherwise the
// case will yield none.
func NewCase(
	test Testable,
	expr Expression,
	argtype, retype d.Typed,
) CaseDef {

	var pattern = Def(Def(Case, test.Type()), retype, argtype)

	return func(args ...Expression) Expression {
		if len(args) > 0 {
			if len(args) > 1 {
				if test.Test(
					NewVector(args...),
				) {
					return expr.Call(NewVector(
						args...))
				}
			}
			if test.Test(args[0]) {
				return expr.Call(args[0])
			}
			return NewNone()
		}
		return NewPair(pattern, test)
	}
}

func (t CaseDef) TypeFnc() TyFnc { return Case }
func (t CaseDef) Type() TyComp {
	return t().(Paired).Left().(TyComp)
}
func (t CaseDef) Test() TestFunc {
	return t().(Paired).Right().(TestFunc)
}
func (t CaseDef) TypeIdent() TyComp {
	return t.Type().Pattern()[0]
}
func (t CaseDef) TypeReturn() TyComp {
	return t.Type().Pattern()[1]
}
func (t CaseDef) TypeArguments() TyComp {
	return t.Type().Pattern()[2]
}
func (t CaseDef) String() string {
	return t.TypeFnc().TypeName()
}
func (t CaseDef) Call(args ...Expression) Expression {
	return t(args...)
}

/// SWITCH
//
// switch takes a slice of cases and evaluates them against its arguments to
// yield either a none value, or the result of the case application and a
// switch enclosing the remaining cases.  id all cases are depleted, a none
// instance will be returned as result and nil will be yielded instead of the
// switch value
//
// when called, a switch evaluates all it's cases until it yields either
// results from applying the first case that matched the arguments, or none.
func NewSwitch(cases ...CaseDef) SwitchDef {
	var types = make([]d.Typed, 0, len(cases))
	for _, c := range cases {
		types = append(types, c.Type())
	}
	var (
		current CaseDef
		remains = cases
		pattern = Def(Switch, Def(types...))
	)
	return func(args ...Expression) (Expression, []CaseDef) {
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
				remains = cases
				return result, cases
			}
			remains = cases
			return NewNone(), cases
		}
		return pattern, cases
	}
}
func (t SwitchDef) Cases() []CaseDef {
	var _, cases = t()
	return cases
}
func (t SwitchDef) Type() TyComp {
	var pat, _ = t()
	return pat.(TyComp)
}
func (t SwitchDef) String() string {
	return t.Type().TypeName()
}
func (t SwitchDef) TypeFnc() TyFnc {
	return Switch
}
func (t SwitchDef) Call(args ...Expression) Expression {
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
func NewMaybe(cas CaseDef) MaybeType {
	var argtypes = make([]d.Typed, 0, len(cas.TypeArguments()))
	for _, arg := range cas.TypeArguments() {
		argtypes = append(argtypes, arg)
	}
	var (
		pattern = Def(Maybe, Def(Def(
			Just, cas.TypeReturn()),
			None), Def(argtypes...))
	)
	return MaybeType(func(args ...Expression) Expression {
		if len(args) > 0 {
			// pass arguments to case, check if result is none‥.
			if result := cas.Call(args...); !result.
				Type().Match(None) {
				// ‥.otherwise return a maybe just
				return JustVal(func(args ...Expression) Expression {
					if len(args) > 0 {
						return result.Call(args...)
					}
					return result.Call()
				})
			}
			// no matching arguments where passed, return none
			return NewNone()
		}
		return pattern
	})
}

func (t MaybeType) TypeFnc() TyFnc                     { return Maybe }
func (t MaybeType) Type() TyComp                       { return t().(TyComp) }
func (t MaybeType) TypeArguments() TyComp              { return t().Type().TypeArgs() }
func (t MaybeType) TypeReturn() TyComp                 { return t().Type().TypeRet() }
func (t MaybeType) String() string                     { return t().String() }
func (t MaybeType) Call(args ...Expression) Expression { return t.Call(args...) }

// maybe values methods
func (t JustVal) Call(args ...Expression) Expression { return t(args...) }
func (t JustVal) String() string                     { return t().String() }
func (t JustVal) Type() TyComp                       { return t().Type() }
func (t JustVal) TypeFnc() TyFnc                     { return Just | t().TypeFnc() }

//// OPTIONAL VALUE
///
// constructor takes two case expressions, first one expected to return the
// either result, second one expected to return the or result if the case
// matches.  if none of the cases match, a none instance will be returned
func NewEitherOr(test Testable, either, or Expression) AlternateDef {
	var pattern = Def(
		Def(
			Def(Either, either.Type().TypeId()),
			Def(Or, or.Type().TypeId()),
		),
		Def(
			Def(Either, either.Type().TypeRet()),
			Def(Or, or.Type().TypeRet()),
		),
		Def(
			Def(Either, either.Type().TypeArgs()),
			Def(Or, or.Type().TypeArgs()),
		),
	)

	return AlternateDef(func(args ...Expression) Expression {
		if len(args) > 0 {
			if len(args) > 1 {
				if test.Test(NewVector(args...)) {
					return EitherVal(either.Call)
				}
			}
			if test.Test(args[0]) {
				return EitherVal(either.Call)
			}
			return OrVal(or.Call)
		}
		return pattern
	})
}
func (o AlternateDef) TypeFnc() TyFnc                     { return Option }
func (o AlternateDef) Type() TyComp                       { return o().Type() }
func (o AlternateDef) String() string                     { return o().String() }
func (o AlternateDef) Call(args ...Expression) Expression { return o(args...) }

/// EITHER VALUE
func (o EitherVal) TypeFnc() TyFnc                     { return Either }
func (o EitherVal) Type() TyComp                       { return o().Type() }
func (o EitherVal) String() string                     { return o().String() }
func (o EitherVal) Call(args ...Expression) Expression { return o.Call(args...) }

/// OR VALUE
func (o OrVal) TypeFnc() TyFnc                     { return Or }
func (o OrVal) Type() TyComp                       { return o().Type() }
func (o OrVal) String() string                     { return o().String() }
func (o OrVal) Call(args ...Expression) Expression { return o.Call(args...) }
