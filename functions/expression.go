package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// GENERIC EXPRESSIONS
	NoneVal   func()
	ConstVal  func() Expression
	ExprVal   func(...Expression) Expression
	ExprType  func(...Expression) Expression
	ParamType func(...Expression) (Expression, []ExprType)
)

//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type. implements countable,
// sliceable, consumeable, testable, compareable, key-, index- and generic pair
// interfaces to be able to stand in as return value for such expressions.
func NewNone() NoneVal { return func() {} }

func (n NoneVal) Head() Expression                   { return n }
func (n NoneVal) Tail() Consumeable                  { return n }
func (n NoneVal) Append(...Expression) Consumeable   { return n }
func (n NoneVal) Len() int                           { return 0 }
func (n NoneVal) Compare(...Expression) int          { return -1 }
func (n NoneVal) String() string                     { return "⊥" }
func (n NoneVal) Call(...Expression) Expression      { return nil }
func (n NoneVal) Key() Expression                    { return nil }
func (n NoneVal) Index() Expression                  { return nil }
func (n NoneVal) Left() Expression                   { return nil }
func (n NoneVal) Right() Expression                  { return nil }
func (n NoneVal) Both() Expression                   { return nil }
func (n NoneVal) Value() Expression                  { return nil }
func (n NoneVal) Empty() d.BoolVal                   { return true }
func (n NoneVal) Test(...Expression) bool            { return false }
func (n NoneVal) TypeFnc() TyFnc                     { return None }
func (n NoneVal) TypeNat() d.TyNat                   { return d.Nil }
func (n NoneVal) Type() TyPattern                    { return Def(None) }
func (n NoneVal) TypeElem() TyPattern                { return Def(None) }
func (n NoneVal) TypeName() string                   { return n.String() }
func (n NoneVal) Slice() []Expression                { return []Expression{} }
func (n NoneVal) Flag() d.BitFlag                    { return d.BitFlag(None) }
func (n NoneVal) FlagType() d.Uint8Val               { return Flag_Function.U() }
func (n NoneVal) Consume() (Expression, Consumeable) { return NewNone(), NewNone() }

//// CONSTANT DECLARATION
///
// declares a constant value
func DecConstant(constant func() Expression) ConstVal { return constant }

func (c ConstVal) Type() TyPattern {
	return Def(None, Def(Constant, c().Type().TypeIdent()), c().Type())
}
func (c ConstVal) TypeFnc() TyFnc                { return Constant }
func (c ConstVal) String() string                { return c().String() }
func (c ConstVal) Call(...Expression) Expression { return c() }

//// EXPRESSION DECLARATION
///
// declares an expression with defined argument-, return- and an optional identity type
func DecFuntion(fn func(...Expression) Expression) ExprVal { return fn }

func (g ExprVal) TypeFnc() TyFnc                     { return Value }
func (g ExprVal) Type() TyPattern                    { return g().Type() }
func (g ExprVal) String() string                     { return g().String() }
func (g ExprVal) Call(args ...Expression) Expression { return g(args...) }

/// PARTIAL APPLYABLE EXPRESSION VALUE
//
// element values yield a subelements of optional, tuple, or enumerable
// expressions with sub-type pattern as second return value
func ConstructExpressionType(
	expr Expression,
	argtype, retype d.Typed,
	identypes ...d.Typed,
) ExprType {

	var (
		arglen         int
		ident, pattern TyPattern
	)
	if len(identypes) == 0 {
		ident = Def(expr.TypeFnc())
	} else {
		ident = Def(identypes...)
	}
	if Flag_Pattern.Match(argtype.FlagType()) {
		arglen = len(argtype.(TyPattern).Patterns())
	} else {
		arglen = 1
	}
	pattern = Def(argtype, ident, retype)

	return func(args ...Expression) Expression {
		var length = len(args)
		if length > 0 {
			if pattern.TypeArguments().MatchArgs(args...) {
				var result Expression
				switch {
				case length == arglen:
					result = expr.Call(args...)
					return result

				case length < arglen:
					var argtypes = make(
						[]d.Typed, 0,
						len(pattern.TypeArguments()[length:]),
					)
					for _, atype := range pattern.TypeArguments()[length:] {
						argtypes = append(argtypes, atype)
					}
					var pattern = Def(Def(argtypes...), ident, retype)
					return ConstructExpressionType(ExprVal(
						func(lateargs ...Expression) Expression {
							if len(lateargs) > 0 {
								return expr.Call(append(
									args, lateargs...,
								)...)
							}
							return pattern
						}), Def(argtypes...), ident, retype)

				case length > arglen:
					var vector = NewVector()
					for len(args) > arglen {
						vector = vector.Con(
							expr.Call(args[:arglen]...))
						args = args[arglen:]
					}
					if length > 0 {
						vector = vector.Con(ConstructExpressionType(
							expr, argtype, retype, identypes...,
						).Call(args...))
					}
					return vector
				}
			}
			return NewNone()
		}
		return pattern
	}
}
func (e ExprType) TypeFnc() TyFnc                     { return Value }
func (e ExprType) Type() TyPattern                    { return e().(TyPattern) }
func (e ExprType) String() string                     { return e().String() }
func (e ExprType) Call(args ...Expression) Expression { return e(args...) }

//// PARAMETRIC VALUE
///
// declare parametric value from set of declared expressions.
func ConstructParametricType(name string, exprs ...ExprType) ParamType {
	var (
		length     = len(exprs)
		cases      = make([]CaseType, 0, length)
		types      = make([]d.Typed, 0, length)
		symbol     TySymbol
		pattern    TyPattern
		caseswitch SwitchType
	)
	for _, expr := range exprs {
		cases = append(cases, CaseType(expr))
	}
	for _, expr := range exprs {
		types = append(types, expr.Type())
	}
	if name == "" {
		var strs = make([]string, 0, length+length-1)
		for n, expr := range exprs {
			strs = append(strs, expr.Type().String())
			if n < length-1 {
				strs = append(strs, " | ")
			}
		}
	}
	symbol = DefSym(name)
	pattern = Def(symbol, Def(types...))
	caseswitch = DecSwitch(cases...)
	return func(args ...Expression) (Expression, []ExprType) {
		if len(args) > 0 {
			var result, cases = caseswitch(args...)
			return result, exprs[length-len(cases):]
		}
		return pattern, exprs
	}
}
func (p ParamType) Call(args ...Expression) Expression { return p.Call(args...) }
func (p ParamType) String() string                     { return p.Type().String() }
func (p ParamType) TypeFnc() TyFnc                     { return Parametric }
func (p ParamType) Type() TyPattern {
	var pattern, _ = p()
	return pattern.(TyPattern)
}
