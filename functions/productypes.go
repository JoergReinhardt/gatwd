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
	//// BOOLEAN ALGEBRA
	///
	// BOOL TYPE
	Bool bool

	// MAYBE (JUST | NONE)
	Maybe    func(...Functor) JustNone
	JustNone Def

	// TESTS AND COMPARE
	Test    func(args ...Functor) Functor
	Compare func(args ...Functor) Functor

	// CASE & SWITCH
	// needs to be variadic in orderto enable type overload
	Switch func(...Functor) (Functor, []Case)
	Case   func(...Functor) Functor

	// ALTERNATETIVES TYPE (EITHER | OR)
	EitherOr  func(...Functor) Functor
	EitherBox func(...Functor) Functor
	OrBox     func(...Functor) Functor

	//// Parametric
	/// funtions return type depends on argument type(s)
	PolyDef Switch
)

// method set of bool alias type
func (b Bool) TypeFnc() TyFnc {
	if b {
		return True
	}
	return False
}
func (b Bool) Type() Decl {
	return Declare(b.TypeFnc())
}
func (b Bool) Call(...Functor) Functor {
	return Box(d.BoolVal(b))
}
func (b Bool) String() string {
	if b {
		return "True"
	}
	return "False"
}

//// MAYBE VALUE
///
// wraps a function definition to cast the result of applying arguments to it
// as instance of just/none.  the empty call is overloaded to return the
// unboxed definition.
func NewMaybe(def Def) Maybe {
	return func(args ...Functor) JustNone {
		if len(args) > 0 {
			return JustNone(def.Call(args...).(Def))
		}
		return JustNone(def)
	}
}

func (t Maybe) Call(args ...Functor) Functor { return t(args...) }
func (t Maybe) Unbox() Def                   { return Def(t()) }
func (t Maybe) String() string               { return t.Unbox().String() }
func (t Maybe) TypeArguments() Decl          { return t.Unbox().Type().TypeArgs() }
func (t Maybe) TypeReturn() Decl             { return t.Unbox().Type().TypeRet() }
func (t Maybe) TypeFnc() TyFnc               { return Option }
func (t Maybe) Type() Decl {
	return Declare(
		Option, Declare(Declare(
			Boxed, t.Unbox().Type(),
		), None))
}

// maybe values methods
func (t JustNone) Call(args ...Functor) Functor { return t.Call(args...) }
func (t JustNone) String() string               { return t().String() }
func (t JustNone) Type() Decl                   { return t().Type() }
func (t JustNone) TypeFnc() TyFnc               { return T | t().TypeFnc() }

/// TRUTH TEST
//
// create a new test, scrutinizing its arguments and revealing true, or false
func NewTest(
	argtype d.Typed,
	test func(a, b Functor) bool,
) Test {
	return Test(Define(Lambda(func(args ...Functor) Functor {
		if len(args) == 0 { // return argument type, when called empty
			if Kind_Decl.Match(argtype.Kind()) {
				return argtype.(Decl)
			}
			return Declare(argtype)
		}
		// type & bounds check provided by function definition
		return Box(d.BoolVal(test(args[0], args[1])))
	}), Truth, Boolean, argtype, argtype))
}
func (t Test) TypeFnc() TyFnc {
	return Truth
}
func (t Test) Type() Decl {
	return Declare(True | False)
}
func (t Test) String() string {
	return t.TypeFnc().TypeName()
}
func (t Test) Test(a, b Functor) bool {
	var result = t(a, b)
	if result.Type().Match(d.Bool) {
		return bool(result.(Atom)().(d.BoolVal))
	}
	return false
}
func (t Test) Compare(a, b Functor) int {
	if t.Test(a, b) {
		return 0
	}
	return -1
}
func (t Test) Call(args ...Functor) Functor { return t(args...) }

/// COMPARATOR
//
// create a comparator expression that yields minus one in case the argument is
// lesser, zero in case its equal and plus one in case it is greater than the
// enclosed value to compare against.
func NewComparator(
	argtype d.Typed,
	comp func(a, b Functor) int,
) Compare {
	return Compare(Define(Lambda(func(args ...Functor) Functor {
		if len(args) == 0 { // return argument type, when called empty
			if Kind_Decl.Match(argtype.Kind()) {
				return argtype.(Decl)
			}
			return Declare(argtype)
		}
		return Box(d.IntVal(comp(args[0], args[1])))
	}), Comparator, Integer, argtype))
}
func (t Compare) TypeFnc() TyFnc {
	return Comparator
}
func (t Compare) Type() Decl {
	return Declare(Lesser | Greater | Equal)
}
func (t Compare) Call(a, b Functor) Functor { return t(a, b) }
func (t Compare) String() string {
	return t.Type().TypeName()
}
func (t Compare) Compare(a, b Functor) int {
	var result = t(a, b)
	if !result.Type().Match(None) {
		return int(result.(Atom)().(d.IntVal))
	}
	return -2
}
func (t Compare) Test(a, b Functor) bool {
	return t.Compare(a, b) == 0
}
func (t Compare) Less(a, b Functor) bool {
	return t.Compare(a, b) < 0
}

/// CASE
//
// case constructor takes a test and an expression, in order for the resulting
// case instance to test its arguments and yield the result of applying those
// arguments to the expression, in case the test yielded true.  otherwise the
// case will yield none.
func NewCase(
	test Testable,
	expr Functor,
	argtype, retype d.Typed,
) Case {

	var pattern = Declare(Declare(Option, test.Type()), retype, argtype)

	return func(args ...Functor) Functor {
		if len(args) > 0 {
			if len(args) > 1 {
				if test.Test(
					args[0], args[1],
				) {
					return expr.Call(args...)
				}
			}
			return NewNone()
		}
		return NewPair(pattern, test)
	}
}

func (t Case) TypeFnc() TyFnc { return Option }
func (t Case) Type() Decl {
	return t().(Paired).Left().(Decl)
}
func (t Case) Test() Test {
	return t().(Paired).Right().(Test)
}
func (t Case) TypeId() Decl {
	return t.Type().Pattern()[0]
}
func (t Case) TypeRet() Decl {
	return t.Type().Pattern()[1]
}
func (t Case) TypeArgs() Decl {
	return t.Type().Pattern()[2]
}
func (t Case) TypeName() string {
	return t.Type().TypeName()
}
func (t Case) String() string {
	return t.TypeFnc().TypeName()
}
func (t Case) Call(args ...Functor) Functor {
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
func NewSwitch(cases ...Case) Switch {
	var types = make([]d.Typed, 0, len(cases))
	for _, c := range cases {
		types = append(types, c.Type())
	}
	var (
		current Case
		remains = cases
		pattern = Declare(Choice, Declare(types...))
	)
	return func(args ...Functor) (Functor, []Case) {
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
func (t Switch) Cases() []Case {
	var _, cases = t()
	return cases
}
func (t Switch) Type() Decl {
	var pat, _ = t()
	return pat.(Decl)
}
func (t Switch) String() string {
	return t.Type().TypeName()
}
func (t Switch) TypeFnc() TyFnc {
	return Choice
}
func (t Switch) Call(args ...Functor) Functor {
	var (
		cases   = t.Cases()
		current Case
		result  Functor
	)
	for len(cases) > 0 {
		current = cases[0]
		if len(cases) > 1 {
			cases = cases[1:]
		} else {
			cases = cases[:0]
		}
		result = current(args...)
		if !result.TypeFnc().Match(None) {
			return result
		}
	}
	return NewNone()
}

//// OPTIONAL VALUE
///
// constructor takes two case expressions, first one expected to return the
// either result, second one expected to return the or result if the case
// matches.  if none of the cases match, a none instance will be returned
func NewEitherOr(test Testable, either, or Functor) EitherOr {
	var pattern = Declare(
		Declare(
			Declare(Either, either.Type().TypeId()),
			Declare(Or, or.Type().TypeId()),
		),
		Declare(
			Declare(Either, either.Type().TypeRet()),
			Declare(Or, or.Type().TypeRet()),
		),
		Declare(
			Declare(Either, either.Type().TypeArgs()),
			Declare(Or, or.Type().TypeArgs()),
		),
	)

	return EitherOr(func(args ...Functor) Functor {
		if len(args) > 0 {
			if len(args) > 1 {
				if test.Test(args[0], args[1]) {
					return EitherBox(either.Call)
				}
			}
			return OrBox(or.Call)
		}
		return pattern
	})
}
func (o EitherOr) TypeFnc() TyFnc               { return Alternative }
func (o EitherOr) Type() Decl                   { return o().Type() }
func (o EitherOr) String() string               { return o().String() }
func (o EitherOr) Call(args ...Functor) Functor { return o(args...) }

/// EITHER VALUE
func (o EitherBox) TypeFnc() TyFnc               { return Either }
func (o EitherBox) Type() Decl                   { return o().Type() }
func (o EitherBox) String() string               { return o().String() }
func (o EitherBox) Call(args ...Functor) Functor { return o.Call(args...) }

/// OR VALUE
func (o OrBox) TypeFnc() TyFnc               { return Or }
func (o OrBox) Type() Decl                   { return o().Type() }
func (o OrBox) String() string               { return o().String() }
func (o OrBox) Call(args ...Functor) Functor { return o.Call(args...) }

///////////////////////////////////////////////////////////////////////////////
/// POLYMORPHIC FUNCTION VALUE
//
// polymorphic functions returns values of different type depending on
// !_ARGUMENT_TYPE_!  function definitions behave just like case definitions
// and cast as such, which makes polymorphic values a special case of generics
func NewPolymorph(variants ...Def) PolyDef {
	var cases = make([]Case, 0, len(variants))
	for _, v := range variants {
		cases = append(cases, Case(v))
	}
	return NewParametric(cases...)
}

/// GENERIC FUNCTION VALUE
//   generic functions return values of different types depending on
//   !_ARGUMENT_VALUE_!
func NewParametric(cases ...Case) PolyDef {
	return PolyDef(NewSwitch(cases...))
}
func (p PolyDef) Call(args ...Functor) Functor {
	return Switch(p).Call(args...)
}
func (p PolyDef) Len() int       { return len(p.Cases()) }
func (p PolyDef) Cases() []Case  { return Switch(p).Cases() }
func (p PolyDef) TypeFnc() TyFnc { return Polymorph }
func (p PolyDef) String() string { return p.TypeName() }
func (p PolyDef) TypeArgs() Decl { return p.Type().TypeArgs() }
func (p PolyDef) TypeRet() Decl  { return p.Type().TypeRet() }
func (p PolyDef) TypeId() Decl   { return p.Type().TypeId() }
func (p PolyDef) Type() Decl {
	var (
		args = make([]d.Typed, 0, p.Len())
		rets = make([]d.Typed, 0, p.Len())
		ids  = make([]d.Typed, 0, p.Len())
	)
	for _, c := range p.Cases() {
		args = append(args, c.TypeArgs())
		rets = append(rets, DecAlt(c.TypeRet()))
		ids = append(ids, c.TypeId())
	}
	return Declare(
		Declare(Polymorph, DecAlt(ids...)),
		DecAlt(rets...),
		DecAlt(args...),
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
	for n, id := range p.TypeId()[1].(Decl) {
		// concat case type ident with corresponding signature
		str = str + "\n" + id.TypeName() +
			" " + p.Signatures()[n]
	}
	return str
}
