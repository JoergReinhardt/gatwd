package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// GENERIC EXPRESSIONS
	NoneVal   func()
	ConstVal  func() Expression
	FuncVal   func(...Expression) Expression
	ExprType  func(...Expression) ExprVal
	ExprVal   Expression
	TupleType func(...Expression) TupleVal
	TupleVal  []Expression
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
	return Def(Constant|c().Type().TypeFnc(), c().Type())
}
func (c ConstVal) TypeFnc() TyFnc                { return Constant }
func (c ConstVal) String() string                { return c().String() }
func (c ConstVal) Call(...Expression) Expression { return c() }

//// FUNCTION DECLARATION
///
// declares a constant value
func DecFunction(fnc func(...Expression) Expression) FuncVal { return fnc }

func (c FuncVal) Type() TyPattern {
	return Def(Value|c().Type().TypeFnc(), c().Type())
}
func (c FuncVal) TypeFnc() TyFnc                { return Value }
func (c FuncVal) String() string                { return c().String() }
func (c FuncVal) Call(...Expression) Expression { return c() }

/// PARTIAL APPLYABLE EXPRESSION VALUE
//
// element values yield a subelements of optional, tuple, or enumerable
// expressions with sub-type pattern as second return value
func Declare(
	expr Expression,
	argtype, retype d.Typed,
	identypes ...d.Typed,
) ExprType {

	if !Flag_Pattern.Match(argtype.FlagType()) {
		argtype = Def(argtype)
	}

	var (
		ident    d.Typed
		pattern  TyPattern
		arglen   = argtype.(TyPattern).Count()
		argtypes = argtype.(TyPattern).Types()
	)

	if len(identypes) == 0 {
		ident = expr.TypeFnc()
	} else {
		if len(identypes) == 1 {
			ident = identypes[0]
		}
		ident = Def(identypes...)
	}

	pattern = Def(argtype, ident, retype)

	return func(args ...Expression) ExprVal {
		var length = len(args)
		if length > 0 {
			if pattern.TypeArguments().MatchArgs(args...) {
				switch {
				case length == arglen:
					return expr.Call(args...)

				case length < arglen:
					argtypes = argtypes[length:]
					return Declare(ExprType(
						func(lateargs ...Expression) ExprVal {
							if len(lateargs) > 0 {
								return expr.Call(append(
									args, lateargs...,
								)...)
							}
							return Def(Def(argtypes...), ident, retype)
						}), Def(argtypes...), ident, retype)

				case length > arglen:
					var vector = NewVector()
					for len(args) > arglen {
						vector = vector.Con(
							expr.Call(args[:arglen]...))
						args = args[arglen:]
					}
					if length > 0 {
						vector = vector.Con(Declare(
							expr, argtype, retype, identypes...,
						).Call(args...))
					}
					return vector
				}
			}
			return None
		}
		return pattern
	}
}
func (e ExprType) TypeFnc() TyFnc                     { return Constructor | Value }
func (e ExprType) Type() TyPattern                    { return e().Call().(TyPattern) }
func (e ExprType) ArgCount() int                      { return e.Type().TypeArguments().Count() }
func (e ExprType) String() string                     { return e().String() }
func (e ExprType) Call(args ...Expression) Expression { return e(args...) }

//// TUPLE DECLARATION
///
//
func DecTuple(types ...d.Typed) ExprType {
	var (
		pattern = make([]Expression, 0, len(types))
		symbol  d.Typed
	)
	if len(types) > 1 {
		if Flag_Symbol.Match(types[0].FlagType()) {
			symbol = types[0]
			types = types[1:]
		}
	}
	if symbol == nil {
		symbol = Tuple
	}
	for _, typ := range types {
		if Flag_Pattern.Match(typ.FlagType()) {
			pattern = append(pattern, typ.(TyPattern))
		} else {
			pattern = append(pattern, Def(typ))
		}
	}
	return Declare(
		TupleType(func(args ...Expression) TupleVal {
			if len(args) > 0 {
				return args
			}
			return pattern
		}),
		Def(types...), Def(symbol, Def(types...)))
}
func (t TupleType) Call(args ...Expression) Expression { return t(args...) }
func (t TupleType) String() string                     { return t.Type().String() }
func (t TupleType) TypeFnc() TyFnc                     { return Tuple | Constructor }
func (t TupleType) Type() TyPattern {
	var (
		elems = t()
		count = len(elems)
		types = make([]d.Typed, 0, count)
	)
	for _, elem := range elems {
		types = append(types, elem.Type())
	}
	return Def(Tuple, Def(types...))
}

func (t TupleVal) Count() int                    { return len(t) }
func (t TupleVal) TypeFnc() TyFnc                { return Tuple }
func (t TupleVal) Call(...Expression) Expression { return t }
func (t TupleVal) Type() TyPattern {
	var types = make([]d.Typed, 0, t.Count())
	for _, elem := range t {
		types = append(types, elem.Type())
	}
	return Def(Tuple, Def(types...))
}
func (t TupleVal) String() string {
	var strs = make([]string, 0, t.Count())
	for _, elem := range t {
		strs = append(strs, elem.String())
	}
	return "[" + strings.Join(strs, ", ") + "]"
}
