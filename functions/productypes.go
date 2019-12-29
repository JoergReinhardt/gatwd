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
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// TESTS AND COMPARE
	TestFunc    func(Expression) bool
	TrinaryFunc func(Expression) int
	CompareFunc func(Expression) int

	// CASE & SWITCH
	CaseDef   func(...Expression) Expression // variadic to enable type overload
	SwitchDef func(...Expression) (Expression, []CaseDef)

	//// DECLARED EXPRESSION
	FuncDef func(...Expression) Expression

	//// POLYMORPHIC EXPRESSION (INSTANCE OF CASE-SWITCH)
	Polymorph func(...Expression) (Expression, []FuncDef, int)
	Variant   func(...Expression) (Expression, Polymorph)

	// TUPLE (TYPE[0]...TYPE[N])
	TupCon func(...Expression) TupVal
	TupVal []Expression

	//// RECORD (PAIR(KEY, VAL)[0]...PAIR(KEY, VAL)[N])
	RecCon func(...Expression) RecVal
	RecVal []KeyPair

	// MAYBE (JUST | NONE)
	MaybeDef func(...Expression) Expression
	JustVal  func(...Expression) Expression

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
// arguments to the expression, in case the test yielded true.  otherwise the
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

/// PARTIAL APPLYABLE EXPRESSION VALUE
//
// defines typesafe partialy applicable expression.  if the set of optional type
// argument(s) starts with a symbol, that will be assumed to be the types
// identity.  otherwise the identity is derived from the passed expression,
// types first field will be the return type, its second field the (set of)
// argument type(s), additional arguments are considered propertys.
func createFuncType(expr Expression, types ...d.Typed) TyComp {
	// if type arguments have been passed, build the type based on them‥.
	if len(types) > 0 {
		// if the first element in pattern is a symbol to be used as
		// ident, just define type from type arguments‥.
		if Kind_Sym.Match(types[0].Kind()) {
			return Def(types...)
		} else { // ‥.otherwise use the expressions ident type
			return Def(append([]d.Typed{expr.Type().TypeId()}, types...)...)
		}
	}
	// ‥.otherwise define by expressions identity entirely in terms of the
	// passed expression type
	return Def(expr.Type().TypeId(),
		expr.Type().TypeRet(),
		expr.Type().TypeArgs())

}

func Define(
	expr Expression,
	types ...d.Typed,
) FuncDef {
	var (
		ct     = createFuncType(expr, types...)
		arglen = ct.TypeArgs().Len()
	)
	// return partialy applicable function
	return func(args ...Expression) Expression {
		var length = len(args)
		if length > 0 {
			if ct.TypeArgs().MatchArgs(args...) {
				switch {
				// NUMBER OF PASSED ARGUMENTS MATCHES EXACTLY →
				case length == arglen:
					return expr.Call(args...)

				// NUMBER OF PASSED ARGUMENTS IS INSUFFICIENT →
				case length < arglen:
					// safe types of arguments remaining to be filled
					var (
						remains = ct.TypeArgs().Types()[length:]
						newpat  = Def(
							Def(Partial, ct.TypeId()),
							ct.TypeRet(),
							Def(remains...))
					)
					// define new function from remaining
					// set of argument types, enclosing the
					// current arguments & appending its
					// own aruments to them, when called.
					return Define(Lambda(func(lateargs ...Expression) Expression {
						// will return result, or
						// another partial, when called
						// with arguments
						if len(lateargs) > 0 {
							return expr.Call(append(
								args, lateargs...,
							)...)
						}
						// if no arguments where
						// passed, return the reduced
						// type ct
						return newpat
					}), newpat.Types()...)

				// NUMBER OF PASSED ARGUMENTS OVERSATISFYING →
				case length > arglen:
					// allocate vector to hold multiple instances
					var vector = NewVector()
					// iterate over arguments, allocate an instance per satisfying set
					for len(args) > arglen {
						vector = vector.Cons(
							expr.Call(args[:arglen]...)).(VecVal)
						args = args[arglen:]
					}
					if length > 0 { // number of leftover arguments is insufficient
						// add a partial expression as vectors last element
						vector = vector.Cons(Define(
							expr, ct.Types()...,
						).Call(args...)).(VecVal)
					}
					// return vector of instances
					return vector
				}
			}
			// passed argument(s) didn't match the expected type(s)
			return None
		}
		// no arguments where passed, return the expression type
		return ct
	}
}
func (e FuncDef) TypeFnc() TyFnc                     { return Constructor | Value }
func (e FuncDef) Type() TyComp                       { return e().(TyComp) }
func (e FuncDef) TypeId() TyComp                     { return e.Type().TypeId() }
func (e FuncDef) TypeArgs() TyComp                   { return e.Type().TypeArgs() }
func (e FuncDef) TypeRet() TyComp                    { return e.Type().TypeRet() }
func (e FuncDef) ArgCount() int                      { return e.Type().TypeArgs().Count() }
func (e FuncDef) String() string                     { return e().String() }
func (e FuncDef) Call(args ...Expression) Expression { return e(args...) }

//// POLYMORPHIC TYPE
///
//
// declare new polymorphic named type from cases
func NewPolyType(name string, defs ...FuncDef) Polymorph {
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
func createPolyType(pattern TyComp, idx int, defs ...FuncDef) Polymorph {
	var length = len(defs)
	return func(args ...Expression) (Expression, []FuncDef, int) {
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
func (p Polymorph) Cases() []FuncDef {
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

//// TUPLE TYPE
///
// tuple type constructor expects a slice of field types and possibly a symbol
// type flag, to define the types name, otherwise 'tuple' is the type name and
// the sequence of field types is shown instead
func NewTupleType(types ...d.Typed) TupCon {
	return func(args ...Expression) TupVal {
		var tup = make(TupVal, 0, len(args))
		if Def(types...).MatchArgs(args...) {
			for _, arg := range args {
				tup = append(tup, arg)
			}
		}
		if len(tup) == 0 {
			for _, t := range types {
				if Kind_Comp.Match(t.Kind()) {
					tup = append(tup, t.(TyComp))
				}
				tup = append(tup, Def(t))
			}
		}
		return tup
	}
}

func (t TupCon) Call(args ...Expression) Expression { return t(args...) }
func (t TupCon) TypeFnc() TyFnc                     { return Tuple | Constructor }
func (t TupCon) String() string                     { return t.Type().String() }
func (t TupCon) Type() TyComp {
	var types = make([]d.Typed, 0, len(t()))
	for _, c := range t() {
		if Kind_Comp.Match(c.Type().Kind()) {
			types = append(types, c.(TyComp))
			continue
		}
		types = append(types, c.Type())
	}
	return Def(Tuple, Def(types...))
}

/// TUPLE VALUE
// tuple value is a slice of expressions, constructed by a tuple type
// constructor validated according to its type pattern.
func (t TupVal) Len() int { return len(t) }
func (t TupVal) String() string {
	var strs = make([]string, 0, t.Len())
	for _, val := range t {
		strs = append(strs, val.String())
	}
	return "[" + strings.Join(strs, ", ") + "]"
}
func (t TupVal) Get(idx int) Expression {
	if idx < t.Len() {
		return t[idx]
	}
	return NewNone()
}
func (t TupVal) TypeFnc() TyFnc                     { return Tuple }
func (t TupVal) Call(args ...Expression) Expression { return NewVector(append(t, args...)...) }
func (t TupVal) Type() TyComp {
	var types = make([]d.Typed, 0, len(t))
	for _, tup := range t {
		types = append(types, tup.Type())
	}
	return Def(Tuple, Def(types...))
}

//// RECORD TYPE
///
//
func NewRecordType(fields ...KeyPair) RecCon {
	return func(args ...Expression) RecVal {
		var rec = make(RecVal, 0, len(args))
		if len(args) > 0 {
			for n, arg := range args {
				if len(fields) > n && arg.Type().Match(Key|Pair) {
					if kp, ok := arg.(KeyPair); ok {
						if strings.Compare(
							string(kp.KeyStr()),
							string(fields[n].KeyStr()),
						) == 0 &&
							fields[n].Value().Type().Match(
								kp.Value().Type(),
							) {
							rec = append(rec, kp)
						}
					}
				}
			}
		}
		if len(rec) == 0 {
			return fields
		}
		return rec
	}
}

func (t RecCon) Call(args ...Expression) Expression { return t(args...) }
func (t RecCon) TypeFnc() TyFnc                     { return Record | Constructor }
func (t RecCon) Type() TyComp {
	var types = make([]d.Typed, 0, len(t()))
	for _, field := range t() {
		types = append(types, Def(
			DefSym(field.KeyStr()),
			Def(field.Value().Type()),
		))
	}
	return Def(Record, Def(types...))
}
func (t RecCon) String() string { return t.Type().String() }

/// RECORD VALUE
// tuple value is a slice of expressions, constructed by a tuple type
// constructor validated according to its type pattern.
func (t RecVal) TypeFnc() TyFnc { return Record }
func (t RecVal) Call(args ...Expression) Expression {
	var fields = make([]Expression, 0, len(t)+len(args))
	for _, field := range t {
		fields = append(fields, field)
	}
	for _, arg := range args {
		if arg.Type().Match(Pair | Key) {
			if kp, ok := arg.(KeyPair); ok {
				fields = append(fields, kp)
			}
		}
	}
	return NewVector(fields...)
}
func (t RecVal) Type() TyComp {
	var types = make([]d.Typed, 0, len(t))
	for _, tup := range t {
		types = append(types, tup.Type())
	}
	return Def(Record, Def(types...))
}
func (t RecVal) Len() int { return len(t) }
func (t RecVal) String() string {
	var strs = make([]string, 0, t.Len())
	for _, field := range t {
		strs = append(strs,
			`"`+field.Key().String()+`"`+" ∷ "+field.Value().String())
	}
	return "{" + strings.Join(strs, " ") + "}"
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
