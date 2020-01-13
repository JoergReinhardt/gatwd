/*

PRODUCT TYPES
-------------
*/
package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (

	// TESTS AND COMPARE
	Praedicate   func(Expression) bool
	PraedTrinary func(Expression) int
	PraedCompare func(Expression) int

	// CASE & SWITCH
	// needs to be variadic in orderto enable type overload
	SwitchDef func(...Expression) (Expression, []CaseDef)
	CaseDef   func(...Expression) Expression

	// MAYBE (JUST | NONE)
	OptionalDef func(...Expression) Expression
	JustVal     func(...Expression) Expression

	// ALTERNATETIVES TYPE (EITHER | OR)
	AlternateDef func(...Expression) Expression
	EitherVal    func(...Expression) Expression
	OrVal        func(...Expression) Expression

	//// Parametric
	/// funtions return type depends on argument type(s)
	PolyDef SwitchDef
)

/// TRUTH TEST
//
// create a new test, scrutinizing its arguments and revealing true, or false
func NewTest(test func(Expression) bool) Praedicate {
	return func(arg Expression) bool { return test(arg) }
}
func (t Praedicate) TypeFnc() TyFnc {
	return Truth
}
func (t Praedicate) Type() TyDef {
	return Def(True | False)
}
func (t Praedicate) String() string {
	return t.TypeFnc().TypeName()
}
func (t Praedicate) Test(arg Expression) bool {
	return t(arg)
}
func (t Praedicate) Compare(arg Expression) int {
	if t(arg) {
		return 0
	}
	return -1
}
func (t Praedicate) Call(args ...Expression) Expression {
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
func NewTrinary(test func(Expression) int) PraedTrinary {
	return func(arg Expression) int { return test(arg) }
}
func (t PraedTrinary) TypeFnc() TyFnc {
	return Trinary
}
func (t PraedTrinary) Type() TyDef {
	return Def(True | False | Undecided)
}
func (t PraedTrinary) Call(arg Expression) Expression {
	return Box(d.IntVal(t(arg)))
}
func (t PraedTrinary) String() string {
	return t.TypeFnc().TypeName()
}
func (t PraedTrinary) Test(arg Expression) bool {
	return t(arg) == 0
}
func (t PraedTrinary) Compare(arg Expression) int {
	return t(arg)
}

/// COMPARATOR
//
// create a comparator expression that yields minus one in case the argument is
// lesser, zero in case its equal and plus one in case it is greater than the
// enclosed value to compare against.
func NewComparator(comp func(Expression) int) PraedCompare {
	return func(arg Expression) int { return comp(arg) }
}
func (t PraedCompare) TypeFnc() TyFnc {
	return Compare
}
func (t PraedCompare) Type() TyDef {
	return Def(Lesser | Greater | Equal)
}
func (t PraedCompare) Call(arg Expression) Expression {
	return Box(d.IntVal(t(arg)))
}
func (t PraedCompare) String() string {
	return t.Type().TypeName()
}
func (t PraedCompare) Test(arg Expression) bool {
	return t(arg) == 0
}
func (t PraedCompare) Less(arg Expression) bool {
	return t(arg) < 0
}
func (t PraedCompare) Compare(arg Expression) int {
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
func (t CaseDef) Type() TyDef {
	return t().(Paired).Left().(TyDef)
}
func (t CaseDef) Test() Praedicate {
	return t().(Paired).Right().(Praedicate)
}
func (t CaseDef) TypeId() TyDef {
	return t.Type().Pattern()[0]
}
func (t CaseDef) TypeRet() TyDef {
	return t.Type().Pattern()[1]
}
func (t CaseDef) TypeArgs() TyDef {
	return t.Type().Pattern()[2]
}
func (t CaseDef) TypeName() string {
	return t.Type().TypeName()
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
func (t SwitchDef) Type() TyDef {
	var pat, _ = t()
	return pat.(TyDef)
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
func NewMaybe(cas CaseDef) OptionalDef {
	var argtypes = make([]d.Typed, 0, len(cas.TypeArgs()))
	for _, arg := range cas.TypeArgs() {
		argtypes = append(argtypes, arg)
	}
	var (
		pattern = Def(Optionals, Def(Def(
			Just, cas.TypeRet()),
			None), Def(argtypes...))
	)
	return OptionalDef(func(args ...Expression) Expression {
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

func (t OptionalDef) TypeFnc() TyFnc                     { return Optionals }
func (t OptionalDef) Type() TyDef                        { return t().(TyDef) }
func (t OptionalDef) TypeArguments() TyDef               { return t().Type().TypeArgs() }
func (t OptionalDef) TypeReturn() TyDef                  { return t().Type().TypeRet() }
func (t OptionalDef) String() string                     { return t().String() }
func (t OptionalDef) Call(args ...Expression) Expression { return t.Call(args...) }

// maybe values methods
func (t JustVal) Call(args ...Expression) Expression { return t(args...) }
func (t JustVal) String() string                     { return t().String() }
func (t JustVal) Type() TyDef                        { return t().Type() }
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
func (o AlternateDef) TypeFnc() TyFnc                     { return Alternatives }
func (o AlternateDef) Type() TyDef                        { return o().Type() }
func (o AlternateDef) String() string                     { return o().String() }
func (o AlternateDef) Call(args ...Expression) Expression { return o(args...) }

/// EITHER VALUE
func (o EitherVal) TypeFnc() TyFnc                     { return Either }
func (o EitherVal) Type() TyDef                        { return o().Type() }
func (o EitherVal) String() string                     { return o().String() }
func (o EitherVal) Call(args ...Expression) Expression { return o.Call(args...) }

/// OR VALUE
func (o OrVal) TypeFnc() TyFnc                     { return Or }
func (o OrVal) Type() TyDef                        { return o().Type() }
func (o OrVal) String() string                     { return o().String() }
func (o OrVal) Call(args ...Expression) Expression { return o.Call(args...) }

///////////////////////////////////////////////////////////////////////////////
/// POLYMORPHIC FUNCTION VALUE
//
// polymorphic functions returns values of different type depending on
// !_ARGUMENT_TYPE_!  function definitions behave just like case definitions
// and cast as such, which makes polymorphic values a special case of generics
func NewPolymorph(variants ...FuncVal) PolyDef {
	var cases = make([]CaseDef, 0, len(variants))
	for _, v := range variants {
		cases = append(cases, CaseDef(v))
	}
	return NewParametric(cases...)
}

/// GENERIC FUNCTION VALUE
//   generic functions return values of different types depending on
//   !_ARGUMENT_VALUE_!
func NewParametric(cases ...CaseDef) PolyDef {
	return PolyDef(NewSwitch(cases...))
}
func (p PolyDef) Call(args ...Expression) Expression {
	return SwitchDef(p).Call(args...)
}
func (p PolyDef) Len() int         { return len(p.Cases()) }
func (p PolyDef) Cases() []CaseDef { return SwitchDef(p).Cases() }
func (p PolyDef) TypeFnc() TyFnc   { return Polymorph }
func (p PolyDef) String() string   { return p.TypeName() }
func (p PolyDef) TypeArgs() TyDef  { return p.Type().TypeArgs() }
func (p PolyDef) TypeRet() TyDef   { return p.Type().TypeRet() }
func (p PolyDef) TypeId() TyDef    { return p.Type().TypeId() }
func (p PolyDef) Type() TyDef {
	var (
		args = make([]d.Typed, 0, p.Len())
		rets = make([]d.Typed, 0, p.Len())
		ids  = make([]d.Typed, 0, p.Len())
	)
	for _, c := range p.Cases() {
		args = append(args, c.TypeArgs())
		rets = append(rets, DefAlt(c.TypeRet()))
		ids = append(ids, c.TypeId())
	}
	return Def(
		Def(Polymorph, DefAlt(ids...)),
		DefAlt(rets...),
		DefAlt(args...),
	)
}

// ["argtypes...( → seperated) = retypes ...( | seperated)"] for each case
func (p PolyDef) Signatures() []string {

	var strs = make([]string, 0, p.Len())

	for _, c := range p.Cases() {
		var (
			args = make([]string, 0, c.TypeArgs().Len())
			rets = make([]string, 0, c.TypeRet().Len())
		)
		for _, arg := range c.TypeArgs() {
			args = append(args, arg.TypeName())
		}
		for _, ret := range c.TypeRet() {
			rets = append(rets, ret.TypeName())
		}
		strs = append(strs, strings.Join([]string{
			strings.Join(args, " → "),
			strings.Join(rets, " | "),
		}, " ＝ "))
	}
	return strs
}

// returns generic type definition as first line and case identity follwed by
// signature for each case in a consequtive line
func (p PolyDef) TypeName() string {
	var str = "Τ :: * → *" // generic parametric type
	// range over case type identitys
	for n, id := range p.TypeId()[1].(TyDef) {
		// concat case type ident with corresponding signature
		str = str + "\n" + id.TypeName() +
			" " + p.Signatures()[n]
	}
	return str
}
