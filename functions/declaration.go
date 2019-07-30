package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	ConstVal     func() Expression
	FuncVal      func(...Expression) Expression
	ArgType      func() []d.Typed
	DeclaredExpr func(...Expression) Expression
)

func NewConstant(fn func() Expression) ConstVal  { return fn }
func (c ConstVal) TypeFnc() TyFnc                { return Constant | c().TypeFnc() }
func (c ConstVal) Type() TyPattern               { return c().Type() }
func (c ConstVal) String() string                { return c().String() }
func (c ConstVal) Call(...Expression) Expression { return c() }

func NewFunction(fn func(...Expression) Expression, pattern TyPattern) FuncVal {
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			return fn(args...)
		}
		return NewPair(fn(), pattern)
	}
}
func (g FuncVal) TypeFnc() TyFnc                     { return Value }
func (g FuncVal) Type() TyPattern                    { return g().(Paired).Right().(TyPattern) }
func (g FuncVal) Unbox() Expression                  { return g().(Paired).Left() }
func (g FuncVal) String() string                     { return g.Unbox().String() }
func (g FuncVal) Call(args ...Expression) Expression { return g(args...) }

// declare types of set of arguments
func DeclareArguments(types ...d.Typed) ArgType      { return func() []d.Typed { return types } }
func (a ArgType) TypeFnc() TyFnc                     { return Argument }
func (a ArgType) Type() TyPattern                    { return Def(a()...) }
func (a ArgType) Len() int                           { return len(a()) }
func (a ArgType) Head() Expression                   { return a.Type().Head() }
func (a ArgType) Tail() Consumeable                  { return a.Type().Tail() }
func (a ArgType) Consume() (Expression, Consumeable) { return a.Type().Consume() }
func (a ArgType) String() string {
	var strs = make([]string, 0, a.Len())
	for _, t := range a() {
		strs = append(strs, t.String())
	}
	return strings.Join(strs, " → ")
}
func (a ArgType) Call(args ...Expression) Expression {
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
func (a ArgType) MatchArg(arg Expression) (ArgType, bool) {
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
	return DeclareArguments(types...),
		current.Match(arg.Type())
}
func (a ArgType) MatchArgs(args ...Expression) bool {
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

// declare type-safe expression that can be partialy applyed
func DeclareExpression(expr Expression, types ...d.Typed) DeclaredExpr {
	var tlen = len(types)
	return func(args ...Expression) Expression {
		var alen = len(args)
		if alen > 0 {
			switch {
			case alen == tlen:
				var matcher = DeclareArguments(types...)
				if matcher.MatchArgs(args...) {
					return expr.Call(args...)
				}

			case alen < tlen:
				var (
					currenTypes = types[:alen]
					remainTypes = types[alen:]
					matcher     = DeclareArguments(currenTypes...)
				)
				if matcher.MatchArgs(args...) {
					return DeclareExpression(NewFunction(
						func(lateargs ...Expression) Expression {
							return expr.Call(
								append(
									args,
									lateargs...,
								)...)
						}, expr.Type()), remainTypes...)
				}

			case alen > tlen:
				var (
					currenArgs = args[:tlen]
					remainArgs = args[tlen:]
					matcher    = DeclareArguments(types...)
				)
				if matcher.MatchArgs(currenArgs...) {
					var list = NewList(expr.Call(currenArgs...))
					for len(remainArgs) > 0 {
						if len(remainArgs) >= tlen {
							currenArgs = remainArgs[:tlen]
							remainArgs = remainArgs[tlen:]
						} else {
							currenArgs = remainArgs
							remainArgs = []Expression{}
						}
						list = list.Con(
							DeclareExpression(
								expr,
								types...,
							)(currenArgs...))
					}
					return list
				}
			}
			return NewNone()
		}
		return NewPair(expr, DeclareArguments(types...))
	}
}
func (e DeclaredExpr) ArgType() ArgType  { return e().(PairVal).Right().(ArgType) }
func (e DeclaredExpr) Unbox() Expression { return e().(PairVal).Left() }
func (e DeclaredExpr) Type() TyPattern   { return e.ArgType().Type() }
func (e DeclaredExpr) TypeFnc() TyFnc    { return Value }
func (e DeclaredExpr) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return e(args...)
	}
	return e.Unbox()
}
func (e DeclaredExpr) String() string {
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
