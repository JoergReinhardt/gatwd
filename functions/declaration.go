package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	ValFunction func(...Expression) Expression
	ValConstant func() Expression
	ArgumentSet func() []d.Typed
	TypedExpr   func(...Expression) Expression
	CurryExpr   func(...Expression) Expression

	//// TYPESAFE COLLECTION
	CollectionType func(...Expression) (Expression, Consumeable)
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
func DefineArgumentSet(types ...d.Typed) ArgumentSet     { return func() []d.Typed { return types } }
func (a ArgumentSet) TypeFnc() TyFnc                     { return Argument }
func (a ArgumentSet) Type() TyPattern                    { return Def(a()...) }
func (a ArgumentSet) Len() int                           { return len(a()) }
func (a ArgumentSet) Head() Expression                   { return a.Type().Head() }
func (a ArgumentSet) Tail() Consumeable                  { return a.Type().Tail() }
func (a ArgumentSet) Consume() (Expression, Consumeable) { return a.Type().Consume() }
func (a ArgumentSet) String() string {
	var strs = make([]string, 0, a.Len())
	for _, t := range a() {
		strs = append(strs, t.String())
	}
	return strings.Join(strs, " → ")
}
func (a ArgumentSet) Call(args ...Expression) Expression {
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
func (a ArgumentSet) MatchArg(arg Expression) (ArgumentSet, bool) {
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
func (a ArgumentSet) MatchArgs(args ...Expression) bool {
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
func (e TypedExpr) ArgType() ArgumentSet { return e().(PairVal).Right().(ArgumentSet) }
func (e TypedExpr) Unbox() Expression    { return e().(PairVal).Left() }
func (e TypedExpr) Type() TyPattern      { return e.ArgType().Type() }
func (e TypedExpr) TypeFnc() TyFnc       { return Value }
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

//// TYPE SAFE COLLECTIONS
///
//
func DeclareCollection(col Consumeable, argtype d.Typed) CollectionType {
	return func(args ...Expression) (Expression, Consumeable) {
		if len(args) == 1 {
			var arg = args[0]
			if arg.Type().Match(argtype) {
				col = col.Append(arg)
			}
			if arg.TypeFnc().Match(Type) {
				if t, ok := arg.(TyFnc); ok {
					if t.Match(Type) {
						return NewNative(true), col
					}
				}
			}
		}
		if len(args) > 1 {
			var fas = make([]Expression, 0, len(args))
			for _, arg := range args {
				if argtype.Match(arg.Type()) {
					fas = append(fas, arg)
				}
			}
			col = col.Append(fas...)
		}
		var head, tail = col.Consume()
		return head, DeclareCollection(tail, argtype)
	}
}
func (c CollectionType) Consume() (Expression, Consumeable) { return c() }
func (c CollectionType) Head() Expression {
	var head, _ = c()
	return head
}
func (c CollectionType) Tail() Consumeable {
	var _, tail = c()
	return tail
}
func (c CollectionType) Unbox() Consumeable {
	var _, unboxed = c(Type)
	return unboxed
}
func (c CollectionType) TypeFnc() TyFnc    { return c.Unbox().TypeFnc() }
func (c CollectionType) Type() TyPattern   { return c.Unbox().Type() }
func (c CollectionType) TypeElem() d.Typed { return c.Unbox().TypeElem() }
func (c CollectionType) Len() int          { return c.Unbox().Len() }
func (c CollectionType) String() string    { return c.Unbox().String() }
func (c CollectionType) Append(args ...Expression) Consumeable {
	var _, col = c(args...)
	return col
}
func (c CollectionType) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var head, tail = c(args...)
		return NewPair(head, tail)
	}
	var head, tail = c()
	return NewPair(head, tail)
}
