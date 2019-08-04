package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	ValFunction func(...Expression) Expression
	ValConstant func() Expression
	ArumentSet  func() []d.Typed
	TypedExpr   func(...Expression) Expression
	CurryExpr   func(...Expression) Expression

	//// COLLECTION TYPES
	ListType  func(...Expression) (Expression, ColList)
	Vec       func(...Expression) []Expression
	TypedPair func(...Expression) (l, r Expression)

	ColPairL func(...Paired) (Paired, ColPairL)
	ColPairV func(...Paired) []Paired

	ColVal func(...Expression) (Expression, Consumeable)
)

//// FUNCTION DECLARATION
///
// declares an expression from some generic functions, with a signature
// indicating that it takes expressions as arguments and returns an expression
func DeclareFunction(fn func(...Expression) Expression, reType TyPattern) ValFunction {
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			return fn(args...)
		}
		return reType
	}
}
func (g ValFunction) TypeFnc() TyFnc                     { return Value }
func (g ValFunction) Type() TyPattern                    { return g().(TyPattern) }
func (g ValFunction) String() string                     { return g().String() }
func (g ValFunction) Call(args ...Expression) Expression { return g(args...) }

//// CONSTANT DECLARATION
///
// declares an expression from a constant function, that returns an expression
func DeclareConstant(fn func() Expression) ValConstant { return fn }
func (c ValConstant) TypeFnc() TyFnc                   { return Constant | c().TypeFnc() }
func (c ValConstant) Type() TyPattern                  { return c().Type() }
func (c ValConstant) String() string                   { return c().String() }
func (c ValConstant) Call(...Expression) Expression    { return c() }

//// ARGUMENT SET
///
// define a set of arguments as a sequence of argument types.
func DefineArgumentSet(types ...d.Typed) ArumentSet     { return func() []d.Typed { return types } }
func (a ArumentSet) TypeFnc() TyFnc                     { return Argument }
func (a ArumentSet) Type() TyPattern                    { return Def(a()...) }
func (a ArumentSet) Len() int                           { return len(a()) }
func (a ArumentSet) Head() Expression                   { return a.Type().Head() }
func (a ArumentSet) Tail() Consumeable                  { return a.Type().Tail() }
func (a ArumentSet) Consume() (Expression, Consumeable) { return a.Type().Consume() }
func (a ArumentSet) String() string {
	var strs = make([]string, 0, a.Len())
	for _, t := range a() {
		strs = append(strs, t.String())
	}
	return strings.Join(strs, " → ")
}
func (a ArumentSet) Call(args ...Expression) Expression {
	if len(args) > 0 {
		if a.MatchArgs(args...) {
			if len(args) > 1 {
				return NewVector(args...)
			}
			return args[0]
		}
	}
	return NewNone()
}
func (a ArumentSet) MatchArg(arg Expression) (ArumentSet, bool) {
	var (
		types   = a()
		current d.Typed
	)
	if len(types) > 0 {
		current = types[0]
	}
	if len(types) > 1 {
		types = types[1:]
	} else {
		types = []d.Typed{}
	}
	return DefineArgumentSet(types...),
		current.Match(arg.Type())
}
func (a ArumentSet) MatchArgs(args ...Expression) bool {
	var (
		at      = a
		ok      bool
		current Expression
	)
	for len(args) > 0 {
		if len(args) > 0 {
			current = args[0]
		}
		if len(args) > 1 {
			args = args[1:]
		} else {
			args = []Expression{}
		}
		if at, ok = at.MatchArg(current); !ok {
			return ok // ← will be false
		}
	}
	return ok // ← will be true
}

//// TYPE SAFE EXPRESSION
///
// declare a type-safe expression. argument types will be matched with the
// types of passed arguments. declared expression can be applyed partialy. for
// multi parameter function, there a three possible sorts of legal calls:
//
// - a call can be undersatisfied by not passing all arguments. in that case a
//   new UeclaredExpr is returned, with an argument set reduced by the the
//   arguments passed and enclosing those.
//
// - a call can pass the exact right number and types of arguments, in which
//    case they will be applyed to the enclosed expression to yield the result.
//
// - a call can pass a sequence of multiple argument sets in which case a
//   vector of results, the last of which might be a partialy applyed
//   expression, will be returned,
func DeclareExpression(expr Expression, types ...d.Typed) TypedExpr {
	var tlen = len(types)
	return func(args ...Expression) Expression {
		var alen = len(args)
		if alen > 0 {
			switch {

			// satisfied
			case alen == tlen:
				var matcher = DefineArgumentSet(types...)
				if matcher.MatchArgs(args...) {
					return expr.Call(args...)
				}

			// undersatisfied
			case alen < tlen:
				var (
					currenTypes = types[:alen]
					remainTypes = types[alen:]
					matcher     = DefineArgumentSet(currenTypes...)
				)
				if matcher.MatchArgs(args...) {
					return DeclareExpression(DeclareFunction(
						func(lateargs ...Expression) Expression {
							return expr.Call(
								append(
									args,
									lateargs...,
								)...)
						}, expr.Type()), remainTypes...)
				}

			// oversatisfied
			case alen > tlen:
				var (
					currenArgs = args[:tlen]
					remainArgs = args[tlen:]
					matcher    = DefineArgumentSet(types...)
					vec        = NewVector()
				)
				if matcher.MatchArgs(currenArgs...) {
					for len(remainArgs) > 0 {
						if len(remainArgs) >= tlen {
							currenArgs = remainArgs[:tlen]
							remainArgs = remainArgs[tlen:]
						} else {
							currenArgs = remainArgs
							remainArgs = []Expression{}
						}
						vec = vec.Con(
							DeclareExpression(
								expr, types...,
							)(currenArgs...))
					}
					return vec
				}
			}
			return NewNone()
		}
		return NewPair(expr, DefineArgumentSet(types...))
	}
}
func (e TypedExpr) ArgType() ArumentSet { return e().(PairVal).Right().(ArumentSet) }
func (e TypedExpr) Unbox() Expression   { return e().(PairVal).Left() }
func (e TypedExpr) Type() TyPattern     { return e.ArgType().Type() }
func (e TypedExpr) TypeFnc() TyFnc      { return Value }
func (e TypedExpr) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return e(args...)
	}
	return e.Unbox()
}
func (e TypedExpr) String() string {
	return strings.Join(append(
		make(
			[]string, 0,
			e.ArgType().Len(),
		),
		e.ArgType().String(),
		e.Unbox().Type().String(),
		e.Unbox().Type().String()),
		" → ",
	)
}

//// CURRY UNTYPED & TYPESAFE
///
// curry instances of simple expression type
func UntypedCurry(fns ...Expression) Expression {
	return CurryExpr(func(args ...Expression) Expression {
		if len(fns) > 0 {
			if len(fns) > 1 {
				if len(args) > 0 {
					return fns[0].Call(UntypedCurry(
						append(fns[1:], args...)...))
				}
				return fns[0].Call(UntypedCurry(fns[1:]...))
			}
			if len(args) > 0 {
				return fns[0].Call(args...)
			}
			return fns[0].Call()
		}
		return NewNone()
	})
}

// curry instances of typesafe expressions
func Curry(fns ...TypedExpr) Expression {
	var args = make([]Expression, 0, len(fns))
	for _, arg := range fns {
		args = append(args, arg)
	}
	return UntypedCurry(args...)
}

// method set to implement curryed expressions
func (c CurryExpr) String() string                     { return c().String() }
func (c CurryExpr) Call(args ...Expression) Expression { return c(args...) }
func (c CurryExpr) Type() TyPattern                    { return c().Type() }
func (c CurryExpr) TypeFnc() TyFnc                     { return c().TypeFnc() }
