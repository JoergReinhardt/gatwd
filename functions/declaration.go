package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	ArgType      func() []d.Typed
	DeclaredExpr func(...Expression) Expression
)

// declare types of set of arguments
func DeclareArguments(types ...d.Typed) ArgType { return func() []d.Typed { return types } }
func (a ArgType) TypeFnc() TyFnc                { return Argument }
func (a ArgType) Type() TyPattern               { return Def(a()...) }
func (a ArgType) Len() int                      { return len(a()) }
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
	var tlength = len(types)
	return func(args ...Expression) Expression {
		var alength = len(args)
		if alength > 0 {
			switch {
			case alength == tlength:
				var matcher = DeclareArguments(types...)
				if matcher.MatchArgs(args...) {
					return expr.Call(args...)
				}
			case alength < tlength:
				var (
					currenTypes = types[:alength]
					remainTypes = types[alength:]
					matcher     = DeclareArguments(currenTypes...)
				)
				if matcher.MatchArgs(args...) {
					return DeclareExpression(NewFunction(
						func(lateargs ...Expression) Expression {
							return DeclareExpression(
								expr, remainTypes...,
							)(append(args, lateargs...)...)
						}, expr.Type()), remainTypes...)
				}
			case alength > tlength:
				var (
					currenArgs = args[:tlength]
					remainArgs = args[tlength:]
					matcher    = DeclareArguments(types...)
				)
				if matcher.MatchArgs(currenArgs...) {
					return NewList(
						expr.Call(currenArgs...),
					).Con(DeclareExpression(
						expr, types...,
					)(remainArgs...))
				}
			}
		}
		return NewPair(expr, DeclareArguments(types...))
	}
}
func (e DeclaredExpr) TypeFnc() TyFnc                     { return Value }
func (e DeclaredExpr) Type() TyPattern                    { return e.Expr().Type() }
func (e DeclaredExpr) ArgType() ArgType                   { return e().(PairVal).Right().(ArgType) }
func (e DeclaredExpr) Expr() Expression                   { return e().(PairVal).Left() }
func (e DeclaredExpr) Call(args ...Expression) Expression { return e(args...) }
func (e DeclaredExpr) String() string {
	return strings.Join(append(
		make(
			[]string, 0,
			e.ArgType().Len(),
		),
		e.ArgType().String(),
		e.Expr().Type().String(),
		e.Expr().Type().String()),
		" → ",
	)
}
