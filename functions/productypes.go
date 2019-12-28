package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// TESTS AND COMPARE
	TestFunc    func(Expression) bool
	TrinaryFunc func(Expression) int
	CompareFunc func(Expression) int

	// CASE & SWITCH
	CaseDef   func(...Expression) Expression // needs to be variadic since overloaded
	SwitchDef func(...Expression) (Expression, []CaseDef)

	// MAYBE (JUST | NONE)
	MaybeDef func(...Expression) Expression
	JustVal  func(...Expression) Expression

	// ALTERNATETIVES TYPE (EITHER | OR)
	EitherOrDef func(...Expression) Expression
	EitherVal   func(...Expression) Expression
	OrVal       func(...Expression) Expression

	// TODO: either poly-/ or option type (homolog)‥. also option value
	// should reference its type constructor and sibling types.

	//// POLYMORPHIC EXPRESSION (INSTANCE OF CASE-SWITCH)
	Polymorph func(...Expression) (Expression, []FuncVal, int)
	Variant   func(...Expression) (Expression, Polymorph)

	//// ENUMERABLE
	EnumDef func(d.Numeral) EnumVal
	EnumVal func(...Expression) (Expression, d.Numeral, EnumDef)
)

/// TRUTH TEST
//
// create a new test, scrutinizing its arguments and revealing true, or false
func NewTest(test func(Expression) bool) TestFunc {
	return func(arg Expression) bool { return test(arg) }
}
func (t TestFunc) TypeFnc() TyFnc           { return Truth }
func (t TestFunc) Type() TyComp             { return Def(True | False) }
func (t TestFunc) String() string           { return t.TypeFnc().TypeName() }
func (t TestFunc) Test(arg Expression) bool { return t(arg) }
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
func (t TrinaryFunc) TypeFnc() TyFnc                 { return Trinary }
func (t TrinaryFunc) Type() TyComp                   { return Def(True | False | Undecided) }
func (t TrinaryFunc) Call(arg Expression) Expression { return Box(d.IntVal(t(arg))) }
func (t TrinaryFunc) String() string                 { return t.TypeFnc().TypeName() }
func (t TrinaryFunc) Test(arg Expression) bool       { return t(arg) == 0 }
func (t TrinaryFunc) Compare(arg Expression) int     { return t(arg) }

/// COMPARATOR
//
// create a comparator expression that yields minus one in case the argument is
// lesser, zero in case its equal and plus one in case it is greater than the
// enclosed value to compare against.
func NewComparator(comp func(Expression) int) CompareFunc {
	return func(arg Expression) int { return comp(arg) }
}
func (t CompareFunc) TypeFnc() TyFnc                 { return Compare }
func (t CompareFunc) Type() TyComp                   { return Def(Lesser | Greater | Equal) }
func (t CompareFunc) Call(arg Expression) Expression { return Box(d.IntVal(t(arg))) }
func (t CompareFunc) String() string                 { return t.Type().TypeName() }
func (t CompareFunc) Test(arg Expression) bool       { return t(arg) == 0 }
func (t CompareFunc) Less(arg Expression) bool       { return t(arg) < 0 }
func (t CompareFunc) Compare(arg Expression) int     { return t(arg) }

/// CASE
//
// case constructor takes a test and an expression, in order for the resulting
// case instance to test its arguments and yield the result of applying those
// arguments to the expression, in case the test yielded true. otherwise the
// case will yield none.
func NewCase(test Testable, expr Expression, argtype, retype d.Typed) CaseDef {
	var pattern = Def(Def(Case, test.Type()), retype, argtype)
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			if len(args) > 1 {
				if test.Test(NewVector(args...)) {
					return expr.Call(NewVector(args...))
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

func (t CaseDef) TypeFnc() TyFnc                     { return Case }
func (t CaseDef) Type() TyComp                       { return t().(Paired).Left().(TyComp) }
func (t CaseDef) Test() TestFunc                     { return t().(Paired).Right().(TestFunc) }
func (t CaseDef) TypeIdent() TyComp                  { return t.Type().Pattern()[0] }
func (t CaseDef) TypeReturn() TyComp                 { return t.Type().Pattern()[1] }
func (t CaseDef) TypeArguments() TyComp              { return t.Type().Pattern()[2] }
func (t CaseDef) String() string                     { return t.TypeFnc().TypeName() }
func (t CaseDef) Call(args ...Expression) Expression { return t(args...) }

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
func (t SwitchDef) reload() SwitchDef { return NewSwitch(t.Cases()...) }
func (t SwitchDef) String() string    { return t.Type().TypeName() }
func (t SwitchDef) TypeFnc() TyFnc    { return Switch }
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
func NewMaybe(cas CaseDef) MaybeDef {
	var argtypes = make([]d.Typed, 0, len(cas.TypeArguments()))
	for _, arg := range cas.TypeArguments() {
		argtypes = append(argtypes, arg)
	}
	var (
		pattern = Def(Def(Just|None), Def(cas.TypeReturn()), Def(argtypes...))
	)
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			// pass arguments to case, check if result is none‥.
			if result := cas.Call(args...); !result.Type().Match(None) {
				// ‥.otherwise return a maybe just
				return JustVal(func(args ...Expression) Expression {
					if len(args) > 0 {
						// return result from passing
						// args to result of initial
						// call
						return result.Call(args...)
					}
					return result.Call()
				})
			}
			// no matching arguments where passed, return none
			return NewNone()
		}
		return pattern
	}
}

func (t MaybeDef) TypeFnc() TyFnc                     { return Maybe }
func (t MaybeDef) Type() TyComp                       { return t().(TyComp) }
func (t MaybeDef) TypeArguments() TyComp              { return t().Type().TypeArgs() }
func (t MaybeDef) TypeReturn() TyComp                 { return t().Type().TypeRet() }
func (t MaybeDef) String() string                     { return t().String() }
func (t MaybeDef) Call(args ...Expression) Expression { return t.Call(args...) }

// maybe values methods
func (t JustVal) TypeFnc() TyFnc                     { return Just }
func (t JustVal) Call(args ...Expression) Expression { return t(args...) }
func (t JustVal) String() string                     { return t().String() }
func (t JustVal) Type() TyComp                       { return t().Type() }

//// OPTIONAL VALUE
///
// constructor takes two case expressions, first one expected to return the
// either result, second one expected to return the or result if the case
// matches. if none of the cases match, a none instance will be returned
func NewEitherOr(test Testable, either, or Expression) EitherOrDef {
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

	return EitherOrDef(func(args ...Expression) Expression {
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
func (o EitherOrDef) TypeFnc() TyFnc                     { return Option }
func (o EitherOrDef) Type() TyComp                       { return o().Type() }
func (o EitherOrDef) String() string                     { return o().String() }
func (o EitherOrDef) Call(args ...Expression) Expression { return o(args...) }

//// ALTERNATIVE VALUE
///
func (o EitherVal) TypeFnc() TyFnc                     { return Either }
func (o EitherVal) Type() TyComp                       { return o().Type() }
func (o EitherVal) String() string                     { return o().String() }
func (o EitherVal) Call(args ...Expression) Expression { return o.Call(args...) }

///
func (o OrVal) TypeFnc() TyFnc                     { return Or }
func (o OrVal) Type() TyComp                       { return o().Type() }
func (o OrVal) String() string                     { return o().String() }
func (o OrVal) Call(args ...Expression) Expression { return o.Call(args...) }

//// POLYMORPHIC TYPE
///
//
// declare new polymorphic named type from cases
func NewPolyType(name string, defs ...FuncVal) Polymorph {
	var (
		types   = make([]d.Typed, 0, len(defs))
		pattern TyComp
	)
	for _, def := range defs {
		types = append(types, def.Type())
	}
	pattern = Def(DefSym(name), Def(types...))
	return createPolyType(pattern, 0, defs...)
}

// type constructor to construct type instances holding execution state during
// recursion
func createPolyType(pattern TyComp, idx int, defs ...FuncVal) Polymorph {
	var length = len(defs)
	return func(args ...Expression) (Expression, []FuncVal, int) {
		if len(args) > 0 { // arguments where passed
			if idx < length { // not all cases scrutinized yet
				// scrutinize arguments, retrieve fnc, or none
				var fnc = defs[idx](args...)
				// if none‥.
				if fnc.Type().Match(None) {
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
					return createPolyType(
							pattern, idx, defs...),
						defs, idx
				}
				// argument is not none if it matched case,
				// return result as variant of polymorphic type
				return Variant(func(args ...Expression) (Expression, Polymorph) {
					if len(args) > 0 {
						return fnc.Call(args...), createPolyType(
							pattern, idx, defs...)
					}
					return fnc.Call(), createPolyType(
						pattern, idx, defs...)
				}), defs, idx
			}
		}
		// return poly type instance with index set to zero
		return createPolyType(pattern, 0, defs...), defs, 0
	}
}

// call loops over all cases with a passed set of arguments and returns either
// result, or none
func (p Polymorph) Call(args ...Expression) Expression {
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
func (p Polymorph) TypeFnc() TyFnc { return Parametric }

// type is the sum of all argument set and return value types, identity is
// defined by passed name
func (p Polymorph) Type() TyComp {
	var (
		t, _, _  = p()
		pat      = t.(TyComp)
		identype = pat.Pattern()[0]
		argtypes = make([]d.Typed, 0, len(pat.Pattern()))
		retypes  = make([]d.Typed, 0, len(pat.Pattern()))
	)
	for _, pat := range pat.Pattern()[1:] {
		argtypes = append(argtypes, Def(pat.TypeArgs()...))
		retypes = append(retypes, pat.TypeRet())
	}
	return Def(identype, Def(retypes...), Def(argtypes...))
}

// returns set of all sub-type defining cases
func (p Polymorph) Cases() []FuncVal {
	var _, c, _ = p()
	return c
}

// returns set index of last evaluated case
func (p Polymorph) Index() int {
	var _, _, i = p()
	return i
}
func (p Polymorph) String() string {
	var (
		cases              = p.Cases()
		length             = len(cases)
		arguments, returns = make([]string, 0, length), make([]string, 0, length)
	)
	for _, c := range cases {
		var (
			args   = c.Type().TypeArgs()
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
		returns = append(returns, c.Type().TypeRet().TypeName())
	}
	return "(" + strings.Join(arguments, " | ") + ")" +
		" → " + p.Type().Pattern()[0].TypeName() +
		" → " + "(" + strings.Join(returns, " | ") + ")"
}

//// POLYMORPHIC SUBTYPE INSTANCE VALUE
///
//
func (p Variant) Expr() Expression {
	var e, _ = p()
	return e
}
func (p Variant) PolyType() Polymorph {
	var _, t = p()
	return t
}
func (p Variant) String() string { return p.Expr().String() }
func (p Variant) TypeFnc() TyFnc { return Parametric }
func (p Variant) Type() TyComp {
	return Def(Def(
		Parametric,
		DefValNat(d.IntVal(p.PolyType().Index())),
	),
		p.Expr().Type(),
	)
}
func (p Variant) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return p.Expr().Call(args...)
	}
	return p.Expr()
}

//// ENUM TYPE
///
// declares an enumerable type returning instances from the set of enumerables
// defined by the passed function
func NewEnumType(fnc func(d.Numeral) Expression) EnumDef {
	return func(idx d.Numeral) EnumVal {
		return func(args ...Expression) (Expression, d.Numeral, EnumDef) {
			if len(args) > 0 {
				return fnc(idx).Call(args...), idx, NewEnumType(fnc)
			}
			return fnc(idx), idx, NewEnumType(fnc)
		}
	}
}
func (e EnumDef) Expr() Expression            { return e(d.IntVal(0)) }
func (e EnumDef) Alloc(idx d.Numeral) EnumVal { return e(idx) }
func (e EnumDef) Type() TyComp {
	return Def(Enum, e.Expr().Type().TypeRet())
}
func (e EnumDef) TypeFnc() TyFnc { return Enum }
func (e EnumDef) String() string { return e.Type().TypeName() }
func (e EnumDef) Call(args ...Expression) Expression {
	if len(args) > 0 {
		if len(args) > 1 {
			var vec = NewVector()
			for _, arg := range args {
				vec = vec.Cons(e.Call(arg)).(VecVal)
			}
			return vec
		}
		var arg = args[0]
		if arg.Type().Match(Data) {
			if nat, ok := arg.(NatEval); ok {
				if i, ok := nat.Eval().(d.Numeral); ok {
					return e(i)
				}
			}
		}
	}
	return e
}

//// ENUM VALUE
///
//
func (e EnumVal) Expr() Expression {
	var expr, _, _ = e()
	return expr
}
func (e EnumVal) Index() d.Numeral {
	var _, idx, _ = e()
	return idx
}
func (e EnumVal) EnumType() EnumDef {
	var _, _, et = e()
	return et
}
func (e EnumVal) Alloc(idx d.Numeral) EnumVal { return e.EnumType().Alloc(idx) }
func (e EnumVal) Next() EnumVal {
	var result = e.EnumType()(e.Index().Int() + d.IntVal(1))
	return result
}
func (e EnumVal) Previous() EnumVal {
	var result = e.EnumType()(e.Index().Int() - d.IntVal(1))
	return result
}
func (e EnumVal) String() string { return e.Expr().String() }
func (e EnumVal) Type() TyComp {
	var (
		nat d.Native
		idx = e.Index()
	)
	if idx.Type().Match(d.BigInt) {
		nat = idx.BigInt()
	} else {
		nat = idx.Int()
	}
	return Def(Def(Enum, DefValNat(nat)), e.Expr().Type())
}
func (e EnumVal) TypeFnc() TyFnc { return Enum | e.Expr().TypeFnc() }
func (e EnumVal) Call(args ...Expression) Expression {
	var r, _, _ = e(args...)
	return r
}
